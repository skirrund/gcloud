package lb

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/skirrund/gcloud/bootstrap/env"

	"github.com/skirrund/gcloud/plugins/zipkin"

	"github.com/skirrund/gcloud/bootstrap"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/registry"
	"github.com/skirrund/gcloud/server"
	"github.com/skirrund/gcloud/server/decoder"
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
}

type service struct {
	Instances []*registry.Instance
	Current   int64
}

var sp *ServerPool

type clientMap struct {
	Clients map[time.Duration]*http.Client
	Mu      sync.Mutex
}

var clients clientMap

func init() {
	clients = clientMap{Clients: make(map[time.Duration]*http.Client)}
	defaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //不校验服务端证书
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxConnsPerHost:       0,
		MaxIdleConnsPerHost:   20,
	}
	GetClient(default_timeout)
}
func GetClient(timeout time.Duration) *http.Client {
	if c, ok := clients.Clients[timeout]; ok {
		return c
	}
	if timeout <= 0 {
		cfgTimeout := env.GetInstance().GetInt64(ConnectionTimeout)
		if cfgTimeout > 0 {
			timeout = time.Duration(cfgTimeout) * time.Second
		}
	}
	clients.Mu.Lock()
	defer clients.Mu.Unlock()
	if c, ok := clients.Clients[timeout]; ok {
		return c
	}
	hc := &http.Client{
		Timeout:   timeout,
		Transport: defaultTransport,
	}
	clients.Clients[timeout] = hc
	return hc
}

func GetInstance() *ServerPool {
	if sp != nil {
		return sp
	}
	once.Do(func() {
		sp = &ServerPool{}
		err := server.RegisterEventHook(server.RegistryChangeEvent, server.EventHook(sp.regChange))
		if err != nil {
			logger.Error("[LB]:", err)
		}
	})
	return sp
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
		return do(req)
	}
	srv := s.GetService(req.ServiceName)
	if srv == nil {
		req.Url = s.GetUrl(req.ServiceName, req.Path)
		logger.Warn("no available service for " + req.ServiceName)
		return do(req)
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
	status, err := do(req)
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
		if ok && ue.Timeout() {
			if ue.Err != nil {
				no, ok := ue.Err.(*net.OpError)
				if ok && no.Op == "dial" {
					return true
				}
			}
		}
		if status == 404 || status == 502 || status == 504 {
			return true
		}
	}

	return false
}

func setHeader(header http.Header, headers map[string]string) {
	if headers == nil {
		return
	}
	for k, v := range headers {
		header.Set(k, v)
	}
}

func do(req *request.Request) (statusCode int, err error) {
	var doRequest *http.Request
	var response *http.Response
	reqUrl := req.Url
	if len(reqUrl) == 0 {
		return 0, errors.New("[http] request url  is empty")
	}
	params := req.Params
	headers := req.Headers
	isJson := req.IsJson
	respResult := req.RespResult
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[[http]] recover :", err)
		}
	}()
	if req.Method == "POST" {
		if params == nil {
			logger.Warn("[http] NewRequest with body nil")
		}
		doRequest, err = http.NewRequest(http.MethodPost, reqUrl, params)
		if err != nil {
			logger.Error("[http] NewRequest error:", err, ",", reqUrl)
			return statusCode, err
		}
		if isJson {
			doRequest.Header.Set("Content-Type", "application/json;charset=utf-8")
		} else if req.HasFile {

		} else {
			doRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
		}

	} else {
		doRequest, err = http.NewRequest(http.MethodGet, reqUrl, nil)
	}
	if err != nil {
		logger.Error("[http] NewRequest error:", err, ",", reqUrl)
		return statusCode, err
	}
	setHeader(doRequest.Header, headers)
	start := time.Now()
	defer requestEnd(reqUrl, start)
	span, err := zipkin.WrapHttp(doRequest, req.ServiceName)
	if err == nil {
		defer span.Finish()
	}
	timeOut := req.TimeOut
	if timeOut == 0 {
		timeOut = default_timeout
	}
	httpC := GetClient(timeOut)
	response, err = httpC.Do(doRequest)
	if err != nil {
		logger.Error("[http] client.Do error:", err.Error(), ",", reqUrl, ",")
		return 0, err
	}
	defer response.Body.Close()
	sc := response.StatusCode
	b, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Error("[http] response body read error:", reqUrl)
		return sc, err
	}
	if sc != http.StatusOK {
		logger.Error("[http] StatusCode error:", sc, ",", reqUrl, ",", string(b))
		return sc, errors.New("http code error:" + strconv.FormatInt(int64(sc), 10))
	}

	ct := response.Header.Get("Content-Type")
	logger.Info("[http] response content-type:", ct)
	d, err := decoder.GetDecoder(ct).DecoderObj(b, respResult)
	_, ok := d.(decoder.StreamDecoder)
	if !ok {
		str := string(b)
		if len(str) > 1000 {
			str = utils.SubStr(str, 0, 1000)
		}
		logger.Info("[http] response:", str)
	} else {
		logger.Info("[http] response:stream not log")
	}

	return sc, nil
}

func requestEnd(url string, start time.Time) {
	logger.Info("[http] request url method return :", url, " elapsed:", time.Since(start).Milliseconds())
}
