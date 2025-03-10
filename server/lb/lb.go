package lb

import (
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
	Current   int64
}

var sp *ServerPool

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

func (s *ServerPool) regChange(eventType server.EventName, eventInfo interface{}) error {
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
	s.Services.Store(name, srv)
	return srv
}

func (s *ServerPool) GetService(name string, currentIdx int64) *service {
	// v, ok := s.Services.Load(name)
	// if ok && v != nil {
	// 	logger.Info("[LB] load from cache:", name)
	// 	return v.(*service)
	// } else {
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
		srv := &service{
			Instances: ins,
			Current:   currentIdx,
		}
		return srv
	} else {
		logger.Warn("[LB] registry not found:", name)
		return nil
	}
	// }
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
func (s *service) GetNextPeer() (*registry.Instance, int64) {
	length := int64(len(s.Instances))
	if length == 0 {
		return nil, -1
	}
	if length == 1 {
		return s.Instances[0], 0
	}
	next := s.NextIndex()
	idx := next % length
	if idx > length-1 {
		return s.Instances[0], 0
	}
	return s.Instances[idx], idx

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

func unmarshal(resp *response.Response, respResult any) error {
	if resp == nil {
		resp = &response.Response{}
		return nil
	}
	ct := resp.ContentType
	body := resp.Body
	if len(body) > 0 {
		d, err := decoder.GetDecoder(ct).DecoderObj(body, respResult)
		if err != nil {
			logger.Error("[LB] response error:", err.Error())
			return err
		}
		_, ok := d.(decoder.StreamDecoder)
		if !ok {
			str := string(body)
			worker.DefaultWorker.Execute(func() {
				if len(str) > 1000 {
					str = utils.SubStr(str, 0, 1000)
				}
				logger.Info("[LB] response:", str)
			})
		} else {
			logger.Info("[http] response:stream not log")
		}
	}
	return nil
}

// lb对接收到的请求 进行负载均衡
func (s *ServerPool) Run(req *request.Request, respResult any) (*response.Response, error) {
	logger.Info("[LB] >>>>>>LbOptions", req.LbOptions)
	start := time.Now()
	if len(req.ServiceName) == 0 {
		defer requestEnd(req.Url, start)
		resp, err := s.client.Exec(req)
		unmarshal(resp, respResult)
		return resp, err
	}
	lbo := req.LbOptions
	if lbo == nil {
		lbo = request.NewDefaultLbOptions()
	}
	retrys := lbo.Retrys
	currentServiceInstanceIdx := int64(-1)
	if retrys > 0 {
		currentServiceInstanceIdx = lbo.CurrentServiceInstanceIdx
	}
	srv := s.GetService(req.ServiceName, currentServiceInstanceIdx)
	if srv == nil {
		req.Url = s.GetUrl(req.ServiceName, req.Path)
		defer requestEnd(req.Url, start)
		logger.Warn("no available service for " + req.ServiceName)
		resp, err := s.client.Exec(req)
		unmarshal(resp, respResult)
		return resp, err
	}

	if retrys >= len(srv.Instances) {
		resp := &response.Response{}
		resp.StatusCode = lbo.CurrentStatuCode
		logger.Info("[LB] retry all instances:", req.Url, ",instances num:", len(srv.Instances), ",retrys:", retrys)
		return resp, lbo.CurrentError
	}
	if retrys > lbo.MaxRetriesOnNextServiceInstance {
		resp := &response.Response{}
		resp.StatusCode = lbo.CurrentStatuCode
		logger.Info("[LB] Max retry reached:", req.Url, ",", len(srv.Instances), ",", retrys)
		return resp, lbo.CurrentError
	}

	instance, idx := srv.GetNextPeer()
	logger.Info("[LB] get instance", instance)
	if instance == nil {
		return &response.Response{}, errors.New("no available service" + req.ServiceName)
	}
	req.Url = instance.GetUrl() + req.Path
	if len(req.Url) == 0 {
		return &response.Response{}, errors.New("request url  is empty")
	}
	defer requestEnd(req.Url, start)
	resp, err := s.client.Exec(req)
	if s.client.CheckRetry(err, resp.StatusCode) {
		logger.Info("[LB] retry next:", req.ServiceName)
		retrys += 1
		lbo.Retrys = retrys
		lbo.CurrentStatuCode = resp.StatusCode
		lbo.CurrentError = err
		lbo.CurrentServiceInstanceIdx = idx
		req.LbOptions = lbo
		return s.Run(req, respResult)
	} else {
		unmarshal(resp, respResult)
	}
	return resp, err

}

func requestEnd(url string, start time.Time) {
	worker.AsyncExecute(func() {
		logger.Infof("[http] request url method return :%s,elapsed:%dms", url, time.Since(start).Milliseconds())
	})
}
