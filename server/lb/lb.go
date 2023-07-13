package lb

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/skirrund/gcloud/bootstrap"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/registry"
	"github.com/skirrund/gcloud/server"
	"github.com/skirrund/gcloud/server/http/client"
	"github.com/skirrund/gcloud/server/request"
	"github.com/skirrund/gcloud/utils"
)

const (
	default_timeout                 = 10 * time.Second
	ConnectionTimeout               = "server.http.client.timeout"
	RetryOnConnectionFailure        = "server.http.retry.onConnectionFailure"
	RetryEnabled                    = "server.http.retry.enabled"
	RetryOnAllOperations            = "server.http.retry.allOperations"
	MaxRetriesOnNextServiceInstance = "server.http.retry.maxRetriesOnNextServiceInstance"
	RetryableStatusCodes            = "server.http.retry.retryableStatusCodes"
	ProtocolHttp                    = "http://"
	ProtocolHttps                   = "https://"
)

var once sync.Once
var defaultTransport *http.Transport

type ServerPool struct {
	Services sync.Map
	client   client.HttpClient
}

var defaultClient NetHttpClient

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
		sp = &ServerPool{}
		sp.client = defaultClient
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

func (s *ServerPool) GetService(name string) *service {
	v, ok := s.Services.Load(name)
	if ok && v != nil {
		logger.Info("[LB] load from cache")
		return v.(*service)
	} else {
		if bootstrap.MthApplication != nil && bootstrap.MthApplication.Registry != nil {
			ins, err := bootstrap.MthApplication.Registry.SelectInstances(name)
			logger.Info("[LB] load from registry")
			if err != nil {
				return nil
			}
			err = bootstrap.MthApplication.Registry.Subscribe(name)
			if err != nil {
				logger.Error("[LB] ", err.Error())
			}
			return s.setService(name, ins)
		} else {
			logger.Warn("[LB] registry not found")
			return nil
		}
	}
}

func (s *service) NextIndex() int {
	// 通过原子操作递增 current 的值，并通过对 slice 的长度取模来获得当前索引值。所以，返回值总是介于 0 和 slice 的长度之间，毕竟我们想要的是索引值，而不是总的计数值
	return int(atomic.AddInt64(&s.Current, 1) % int64(len(s.Instances)))
}

// get next instance
func (s *service) GetNextPeer() *registry.Instance {
	next := s.NextIndex()
	length := len(s.Instances)
	if length == 0 {
		return nil
	}
	idx := next % length
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

// lb对接收到的请求 进行负载均衡
func (s *ServerPool) Run(req *request.Request) (int, error) {
	logger.Info("[LB] >>>>>>LbOptions", req.LbOptions)
	if len(req.ServiceName) == 0 {
		return s.client.Exec(req)
	}
	srv := s.GetService(req.ServiceName)
	if srv == nil {
		req.Url = s.GetUrl(req.ServiceName, req.Path)
		logger.Warn("no available service for " + req.ServiceName)
		return s.client.Exec(req)
	}

	lbo := req.LbOptions
	if lbo == nil {
		lbo = request.NewDefaultLbOptions()
	}
	retrys := lbo.Retrys
	if retrys >= len(srv.Instances) {
		logger.Info("[LB] retry all instances:", req.Url, ",instances num:", len(srv.Instances), ",retrys:", retrys)
		return lbo.CurrentStatuCode, lbo.CurrentError
	}
	if retrys > lbo.MaxRetriesOnNextServiceInstance {
		logger.Info("[LB] Max retry reached:", req.Url, ",", len(srv.Instances), ",", retrys)
		return lbo.CurrentStatuCode, lbo.CurrentError
	}

	instance := srv.GetNextPeer()
	logger.Info("[LB] get instance", instance)
	if instance == nil {
		return 0, errors.New("no available service" + req.ServiceName)
	}
	req.Url = instance.GetUrl() + req.Path
	if len(req.Url) == 0 {
		return 0, errors.New("request url  is empty")
	}
	status, err := s.client.Exec(req)
	if checkRetry(err, status) {
		logger.Info("[LB] retry next:", req.ServiceName)
		retrys += 1
		lbo.Retrys = retrys
		lbo.CurrentStatuCode = status
		lbo.CurrentError = err
		req.LbOptions = lbo
		return s.Run(req)
	}
	return status, err

}

func checkRetry(err error, status int) bool {
	if err != nil {
		ue, ok := err.(*url.Error)
		logger.Info("[LB] checkRetry error *url.Error:", ok)
		if ok {
			if ue.Err != nil {
				no, ok := ue.Err.(*net.OpError)
				if ok && no.Op == "dial" {
					return true
				}
			}
		} else {
			no, ok := err.(*net.OpError)
			if ok && no.Op == "dial" {
				return true
			}
		}
		if status == 404 || status == 502 || status == 504 {
			return true
		}
	}
	return false
}

func requestEnd(url string, start time.Time) {
	logger.Info("[http] request url method return :", url, " elapsed:", time.Since(start).Milliseconds())
}
