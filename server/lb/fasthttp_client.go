package lb

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/server/decoder"
	"github.com/skirrund/gcloud/server/request"
	"github.com/skirrund/gcloud/utils"
	"github.com/valyala/fasthttp"
)

type FastHttpClient struct{}

var fastClient *fasthttp.Client

const (
	DefaultTimeout = 30 * time.Second
)

func init() {
	fastClient = &fasthttp.Client{
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, 3*time.Second)
		},
		MaxConnsPerHost:     2000,
		MaxIdleConnDuration: DefaultTimeout,
		MaxConnDuration:     DefaultTimeout,
		ReadTimeout:         5 * time.Minute,
		WriteTimeout:        5 * time.Minute,
		MaxConnWaitTimeout:  5 * time.Second,
	}
}

func (FastHttpClient) Exec(req *request.Request) (statusCode int, err error) {
	doRequest := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(doRequest)
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)
	reqUrl := req.Url
	if len(reqUrl) == 0 {
		return 0, errors.New("[lb-fasthttp] request url  is empty")
	}
	params := req.Params
	headers := req.Headers
	isJson := req.IsJson
	respResult := req.RespResult
	doRequest.Header.SetMethod(req.Method)
	doRequest.SetRequestURI(reqUrl)
	reqHeader := &doRequest.Header
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[lb-fasthttp]] recover :", err)
		}
	}()
	if req.Method != http.MethodGet && req.Method != http.MethodHead && params != nil {
		bodyBytes, _ := io.ReadAll(params)
		doRequest.SetBody(bodyBytes)
		//doRequest.SetBodyStream(params, -1)
		if isJson {
			reqHeader.SetContentType("application/json;charset=utf-8")
		} else if req.HasFile {

		} else {
			reqHeader.SetContentType("application/x-www-form-urlencoded;charset=utf-8")
		}
	}
	setFasthttpHeader(reqHeader, headers)
	start := time.Now()
	defer requestEnd(reqUrl, start)
	timeOut := req.TimeOut
	if timeOut == 0 {
		timeOut = default_timeout
	}
	err = fastClient.DoTimeout(doRequest, response, timeOut)
	if err != nil {
		logger.Error("[lb-fasthttp] fasthttp.Do error:", err.Error(), ",", reqUrl, ",")
		return 0, err
	}
	sc := response.StatusCode()
	ct := string(response.Header.ContentType())
	logger.Info("[lb-fasthttp] response statusCode:", sc, " content-type:", ct)
	if sc >= http.StatusMultipleChoices && sc <= http.StatusPermanentRedirect {
		location := string(response.Header.Peek("Location"))
		logger.Warn("[lb-fasthttp] DoRedirects{ statusCode:", sc, ",location:", location, "}")
		if len(location) > 0 {
			response.Reset()
			doRequest.SetRequestURI(location)
			err = fastClient.DoTimeout(doRequest, response, timeOut)
			if err != nil {
				logger.Error("[lb-fasthttp] DoRedirects error:", err.Error(), ",", reqUrl, ",")
				return 0, err
			}
			sc = response.StatusCode()
			ct = string(response.Header.ContentType())
			logger.Info("[lb-fasthttp] DoRedirects response statusCode:", sc, " content-type:", ct)
		}
	}
	b := response.Body()
	if sc != http.StatusOK {
		logger.Error("[lb-fasthttp] StatusCode error:", sc, ",", reqUrl, ",", string(b))
		return sc, errors.New("fasthttp code error:" + strconv.FormatInt(int64(sc), 10))
	}
	d, err := decoder.GetDecoder(ct).DecoderObj(b, respResult)
	_, ok := d.(decoder.StreamDecoder)
	if !ok {
		str := string(b)
		if len(str) > 1000 {
			str = utils.SubStr(str, 0, 1000)
		}
		logger.Info("[lb-fasthttp] response:", str)
	} else {
		logger.Info("[lb-fasthttp] response:stream not log")
	}

	return sc, nil
}

func setFasthttpHeader(header *fasthttp.RequestHeader, headers map[string]string) {
	if headers == nil {
		return
	}
	for k, v := range headers {
		header.Set(k, v)
	}
}

func (FastHttpClient) CheckRetry(err error, status int) bool {
	if err != nil {
		if err == fasthttp.ErrDialTimeout {
			return true
		}
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
