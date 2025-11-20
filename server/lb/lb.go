package lb

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/skirrund/gcloud/bootstrap"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/registry"
	"github.com/skirrund/gcloud/server"
	"github.com/skirrund/gcloud/server/decoder"
	"github.com/skirrund/gcloud/server/http/client"
	lbClient "github.com/skirrund/gcloud/server/lb/client"
	"github.com/skirrund/gcloud/server/request"
	"github.com/skirrund/gcloud/server/response"
	"github.com/skirrund/gcloud/tracer"
	"github.com/skirrund/gcloud/utils"
	"github.com/skirrund/gcloud/utils/worker"
)

const (
	ConnectionTimeout               = "server.http.client.timeout"
	RetryOnConnectionFailure        = "server.http.retry.onConnectionFailure"
	RetryEnabled                    = "server.http.retry.enabled"
	RetryOnAllOperations            = "server.http.retry.allOperations"
	MaxRetriesOnNextServiceInstance = "server.http.retry.maxRetriesOnNextServiceInstance"
	RetryableStatusCodes            = "server.http.retry.retryableStatusCodes"
	ProtocolHttp                    = "http://"
	ProtocolHttps                   = "https://"
	maxRoundRobin                   = 100000000
)

var once sync.Once

type ServerPool struct {
	Services sync.Map
	client   client.HttpClient
}

type service struct {
	Instances []*registry.Instance
	H2C       bool
	Current   int64
}

var sp *ServerPool

const (
	HTTP2C = "h2c"
)

func GetInstance() *ServerPool {
	if sp != nil {
		return sp
	}
	once.Do(func() {
		sp = &ServerPool{
			client: lbClient.NetHttpClient{},
		}
		err := server.RegisterEventHook(server.RegistryChangeEvent, server.EventHook(sp.regChange))
		if err != nil {
			logger.Error("[LB]:", err)
		}
	})
	return sp
}

func (s *ServerPool) SetHttpClient(client client.HttpClient) {
	s.client = client
}

func (s *ServerPool) GetHttpClient() client.HttpClient {
	return s.client
}

func (s *ServerPool) regChange(eventType server.EventName, eventInfo any) error {
	if eventType == server.RegistryChangeEvent {
		logger.Info("[LB] registry change:", eventInfo)
		if info, ok := eventInfo.(map[string][]*registry.Instance); ok {
			for k, v := range info {
				s.setService(k, v)
			}
		}
	}
	return nil
}

func (s *ServerPool) setService(name string, instances []*registry.Instance) *service {
	srv := &service{
		Instances: instances,
		Current:   -1,
	}
	if len(instances) > 0 {
		h2c := instances[0].Metadata[HTTP2C]
		if h2c == "true" {
			srv.H2C = true
		} else {
			srv.H2C = false
		}
	}
	s.Services.Store(name, srv)
	return srv
}

func (s *ServerPool) GetService(name string) *service {
	v, ok := s.Services.Load(name)
	if ok && v != nil {
		logger.Info("[LB] load from cache:", name)
		return v.(*service)
	} else {
		if bootstrap.MthApplication != nil && bootstrap.MthApplication.Registry != nil {
			ins, err := bootstrap.MthApplication.Registry.SelectInstances(name)
			logger.Info("[LB] load from registry:", name)
			if err != nil {
				return nil
			}
			err = bootstrap.MthApplication.Registry.Subscribe(name)
			if err != nil {
				logger.Error("[LB] registry rubscribe error:", name, "=>", err.Error())
			}
			return s.setService(name, ins)
		} else {
			logger.Warn("[LB] registry not found:", name)
			return nil
		}
	}
}

func (s *service) NextIndex() int64 {
	// 通过原子操作递增 current 的值，并通过对 slice 的长度取模来获得当前索引值。所以，返回值总是介于 0 和 slice 的长度之间，毕竟我们想要的是索引值，而不是总的计数值
	n := atomic.AddInt64(&s.Current, 1)
	idx := n % int64(len(s.Instances))
	if n > maxRoundRobin {
		s.Current = -1
	}
	return idx
}

// get next instance
func (s *service) GetNextPeer() *registry.Instance {
	length := int64(len(s.Instances))
	if length == 0 {
		return nil
	}
	if length == 1 {
		return s.Instances[0]
	}
	next := s.NextIndex()
	idx := next % length
	if idx > length-1 {
		return s.Instances[0]
	}
	return s.Instances[idx]

}

func (s *ServerPool) GetUrl(serviceName string, path string) string {
	if !strings.HasPrefix(serviceName, ProtocolHttp) && !strings.HasPrefix(serviceName, ProtocolHttps) {
		serviceName = ProtocolHttp + serviceName
	}
	if strings.HasSuffix(serviceName, "/") {
		serviceName = utils.SubStr(serviceName, 0, len(serviceName)-1)
	}
	if strings.HasPrefix(path, "/") {
		return serviceName + path
	} else {
		return serviceName + "/" + path
	}
}

func unmarshal(ctx context.Context, resp *response.Response, respResult any) error {
	if resp == nil {
		resp = &response.Response{}
		return nil
	}
	ct := resp.ContentType
	body := resp.Body
	if len(body) > 0 {
		d, err := decoder.GetDecoder(ct).DecoderObj(body, respResult)
		if err != nil {
			logger.ErrorContext(ctx, "[LB] response error:", err.Error())
			return err
		}
		_, ok := d.(decoder.StreamDecoder)
		if !ok {
			str := string(body)
			worker.DefaultWorker.Execute(func() {
				if len(str) > 1000 {
					str = utils.SubStr(str, 0, 1000)
				}
				logger.InfoContext(ctx, "[LB] response:", str)
			})
		} else {
			logger.InfoContext(ctx, "[http] response:stream not log")
		}
	}
	return nil
}

// lb对接收到的请求 进行负载均衡
func (s *ServerPool) Run(req *request.Request, respResult any) (*response.Response, error) {
	loggerCtx := req.Context
	if loggerCtx == nil {
		loggerCtx = tracer.NewTraceIDContext()
	} else {
		loggerCtx = tracer.WithTraceID(loggerCtx)
	}
	req.Context = loggerCtx
	logger.InfoContext(loggerCtx, "[LB] >>>>>>LbOptions", req.LbOptions)
	start := time.Now()
	if len(req.ServiceName) == 0 {
		defer requestEnd(loggerCtx, req.Url, start)
		resp, err := s.client.Exec(req)
		unmarshal(loggerCtx, resp, respResult)
		return resp, err
	}
	srv := s.GetService(req.ServiceName)
	if srv == nil {
		req.Url = s.GetUrl(req.ServiceName, req.Path)
		defer requestEnd(loggerCtx, req.Url, start)
		logger.WarnContext(loggerCtx, "no available service for "+req.ServiceName)
		resp, err := s.client.Exec(req)
		unmarshal(loggerCtx, resp, respResult)
		return resp, err
	}
	lbo := req.LbOptions
	if lbo == nil {
		lbo = request.NewDefaultLbOptions()
	}
	retrys := lbo.Retrys
	if retrys >= len(srv.Instances) {
		resp := &response.Response{}
		resp.StatusCode = lbo.CurrentStatuCode
		logger.InfoContext(loggerCtx, "[LB] retry all instances:", req.Url, ",instances num:", len(srv.Instances), ",retrys:", retrys)
		return resp, lbo.CurrentError
	}
	if retrys > lbo.MaxRetriesOnNextServiceInstance {
		resp := &response.Response{}
		resp.StatusCode = lbo.CurrentStatuCode
		logger.InfoContext(loggerCtx, "[LB] Max retry reached:", req.ServiceName, "=>", req.Url, ",", len(srv.Instances), ",", retrys)
		return resp, lbo.CurrentError
	}

	instance := srv.GetNextPeer()
	logger.InfoContext(loggerCtx, "[LB] get instance:", req.ServiceName, "=>", instance)
	if instance == nil {
		return &response.Response{}, errors.New("no available service" + req.ServiceName)
	}
	if !strings.HasPrefix(req.Path, "/") {
		req.Path = "/" + req.Path
	}
	req.Url = instance.GetUrl() + req.Path
	if len(req.Url) == 0 {
		return &response.Response{}, errors.New("request url  is empty")
	}
	req.H2C = srv.H2C
	defer requestEnd(loggerCtx, req.Url, start)
	resp, err := s.client.Exec(req)
	if err != nil {
		if s.client.CheckRetry(err, resp.StatusCode) {
			logger.InfoContext(loggerCtx, "[LB] retry next:", req.ServiceName)
			retrys += 1
			lbo.Retrys = retrys
			lbo.CurrentStatuCode = resp.StatusCode
			lbo.CurrentError = err
			req.LbOptions = lbo
			return s.Run(req, respResult)
		} else {
			unmarshal(loggerCtx, resp, respResult)
		}
	} else {
		unmarshal(loggerCtx, resp, respResult)
	}

	return resp, err

}

func requestEnd(ctx context.Context, url string, start time.Time) {
	worker.AsyncExecute(func() {
		logger.InfofContext(ctx, "[http] request url method return :%s,elapsed:%dms", url, time.Since(start).Milliseconds())
	})
}
