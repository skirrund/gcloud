package lb

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/skirrund/gcloud/bootstrap/env"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/plugins/zipkin"
	"github.com/skirrund/gcloud/server/decoder"
	"github.com/skirrund/gcloud/server/request"
	"github.com/skirrund/gcloud/utils"
)

type NetHttpClient struct{}

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

func (NetHttpClient) Exec(req *request.Request) (statusCode int, err error) {
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

func setHeader(header http.Header, headers map[string]string) {
	if headers == nil {
		return
	}
	for k, v := range headers {
		header.Set(k, v)
	}
}

func (NetHttpClient) CheckRetry(err error, status int) bool {
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
