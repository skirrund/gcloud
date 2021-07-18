package lb

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

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
	DEFAULT_TIMEOUT                 = 30
	RetryOnConnectionFailure        = "server.http.retry.onConnectionFailure"
	RetryEnabled                    = "server.http.retry.enabled"
	RetryOnAllOperations            = "server.http.retry.allOperations"
	MaxRetriesOnNextServiceInstance = "server.http.retry.maxRetriesOnNextServiceInstance"
	RetryableStatusCodes            = "server.http.retry.retryableStatusCodes"
)

var once sync.Once
var once1 sync.Once

var defaultTransport *http.Transport

type ServerPool struct {
	Services sync.Map
}

type service struct {
	Instances []*registry.Instance
	Current   int64
}

var sp *ServerPool

var httpClient *http.Client

func init() {

}

func GetClient() *http.Client {
	if httpClient != nil {
		return httpClient
	}
	once1.Do(func() {
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
			MaxIdleConnsPerHost:   2,
		}
		httpClient = &http.Client{
			Timeout:   DEFAULT_TIMEOUT * time.Second,
			Transport: defaultTransport,
		}
	})
	return httpClient
}

func GetInstance() *ServerPool {
	if sp != nil {
		return sp
	}
	once.Do(func() {
		sp = &ServerPool{}
		server.RegisterEventHook(server.RegistryChangeEvent, server.EventHook(sp.regChange))
	})
	return sp
}

func (s *ServerPool) regChange(eventType server.EventName, eventInfo interface{}) error {
	if eventType == server.RegistryChangeEvent {
		logger.Info("[LB] registry change:", eventInfo)
		if info, ok := eventInfo.(map[string][]*registry.Instance); ok {
			for k, v := range info {
				s.SetService(k, v)
			}
		}
	}
	return nil
}

func (s *ServerPool) SetService(name string, instances []*registry.Instance) *service {
	srv := &service{
		Instances: instances,
		Current:   -1,
	}
	s.Services.Store(name, srv)
	return srv
}

func (s *ServerPool) getService(name string) *service {
	v, ok := s.Services.Load(name)
	if ok && v != nil {
		logger.Info("[LB] load from cache")
		return v.(*service)
	} else {
		ins, err := bootstrap.MthApplication.Registry.SelectInstances(name)
		logger.Info("[LB] load from nacos")
		if err != nil {
			return nil
		}
		bootstrap.MthApplication.Registry.Subscribe(name)
		return s.SetService(name, ins)
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

// lb对接收到的请求 进行负载均衡
func (s *ServerPool) Run(req *request.Request) (int, error) {
	logger.Info("[LB] >>>>>>LbOptions", req.LbOptions)
	if len(req.ServiceName) == 0 {
		return do(req)
	}
	srv := s.getService(req.ServiceName)
	if srv == nil {
		return 0, errors.New("no available service for " + req.ServiceName)
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
	var request *http.Request
	var response *http.Response
	url := req.Url
	if len(url) == 0 {
		return 0, errors.New("[http] request url  is empty")
	}
	params := req.Params
	timeOut := req.TimeOut
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
		request, err = http.NewRequest(http.MethodPost, url, params)
		if err != nil {
			logger.Error("[http] NewRequest error:", err, ",", url)
			return statusCode, err
		}
		if isJson {
			request.Header.Set("Content-Type", "application/json;charset=utf-8")
		} else if req.HasFile {

		} else {
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
		}

	} else {
		request, err = http.NewRequest(http.MethodGet, url, nil)
	}
	if err != nil {
		logger.Error("[http] NewRequest error:", err, ",", url)
		return statusCode, err
	}

	if timeOut == 0 {
		timeOut = DEFAULT_TIMEOUT * time.Second
	}
	setHeader(request.Header, headers)

	start := time.Now()

	defer requestEnd(url, start)
	span, err := zipkin.WrapHttp(request, req.ServiceName)
	if err == nil {
		defer span.Finish()
	}
	response, err = GetClient().Do(request)

	if err != nil {
		logger.Error("[http] client.Do error:", err.Error(), ",", url, ",")
		return 0, err
	}
	defer response.Body.Close()
	sc := response.StatusCode
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Error("[http] response body read error:", url)
		return sc, err
	}
	if sc != http.StatusOK {
		logger.Error("[http] StatusCode error:", sc, ",", url, ",", string(b))
		return sc, errors.New("http code error:" + strconv.FormatInt(int64(sc), 10))
	}

	ct := response.Header.Get("Content-Type")
	logger.Info("[http] response content-type:", ct)
	d, err := decoder.GetDecoder(ct).DecoderObj(b, respResult)
	_, ok := d.(decoder.StringDecoder)
	_, ok1 := d.(decoder.JSONDecoder)
	if ok || ok1 {
		str := string(b)
		if len(str) > 1000 {
			str = utils.SubStr(str, 0, 1000)
		}
		logger.Info("[http] response:", str)
	} else {
		logger.Info("[http] response:strean not log")
	}

	return sc, nil
}

func requestEnd(url string, start time.Time) {
	logger.Info("[http] request url method return :", url, " elapsed:", time.Since(start).Milliseconds())
}
