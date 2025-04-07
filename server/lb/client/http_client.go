package client

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"maps"

	"github.com/skirrund/gcloud/bootstrap/env"
	"github.com/skirrund/gcloud/logger"
	gCookie "github.com/skirrund/gcloud/server/http/cookie"
	"github.com/skirrund/gcloud/server/request"
	gResp "github.com/skirrund/gcloud/server/response"
)

const (
	default_timeout   = 30 * time.Second
	ConnectionTimeout = "server.http.client.timeout"
)

var defaultTransport *http.Transport

type NetHttpClient struct {
}

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
			KeepAlive: 10 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       10 * time.Second,
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

func (NetHttpClient) Exec(req *request.Request) (r *gResp.Response, err error) {
	var doRequest *http.Request
	var response *http.Response
	reqUrl := req.Url
	r = &gResp.Response{
		Cookies: make(map[string]*gCookie.Cookie),
		Headers: make(map[string][]string),
	}
	if len(reqUrl) == 0 {
		return r, errors.New("[lb-http] request url  is empty")
	}
	params := bytes.NewReader(req.Params)
	headers := req.Headers
	isJson := req.IsJson
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[lb-http] recover :", err)
		}
	}()
	if req.Method == "POST" {
		if params == nil {
			logger.Warn("[lb-http] NewRequest with body nil")
		}
		doRequest, err = http.NewRequest(http.MethodPost, reqUrl, params)
		if err != nil {
			logger.Error("[lb-http] NewRequest error:", err, ",", reqUrl)
			return r, err
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
		return r, err
	}
	setHeader(doRequest.Header, headers)
	timeOut := req.TimeOut
	if timeOut == 0 {
		timeOut = default_timeout
	}
	httpC := GetClient(timeOut)

	response, err = httpC.Do(doRequest)
	if err != nil {
		logger.Error("[lb-http] client.Do error:", err.Error(), ",", reqUrl, ",")
		return r, err
	}
	defer response.Body.Close()
	sc := response.StatusCode
	r.StatusCode = sc
	ct := response.Header.Get("Content-Type")
	r.ContentType = ct
	b, err := io.ReadAll(response.Body)
	logger.Info("[lb-http] response statusCode:", sc, " content-type:", ct)
	r.Body = b
	if err != nil {
		logger.Error("[lb-http] response body read error:", reqUrl)
		return r, err
	}
	cks := response.Cookies()
	for _, c := range cks {
		val, _ := url.QueryUnescape(c.Value)
		r.Cookies[c.Name] = &gCookie.Cookie{
			Key:      c.Name,
			Value:    val,
			Expire:   c.Expires,
			MaxAge:   c.MaxAge,
			Domain:   c.Domain,
			Path:     c.Path,
			HttpOnly: c.HttpOnly,
			Secure:   c.Secure,
			SameSite: getSameSite(c.SameSite),
		}
	}
	respHeaders := response.Header
	maps.Copy(r.Headers, respHeaders)
	if sc != http.StatusOK {
		logger.Error("[lb-http] StatusCode error:", sc, ",", reqUrl, ",", string(b))
		return r, errors.New("lb-http code error:" + strconv.FormatInt(int64(sc), 10))
	}
	return r, nil
}

func getSameSite(sameSite http.SameSite) (s gCookie.CookieSameSite) {
	switch sameSite {
	case http.SameSiteDefaultMode:
		return
	case http.SameSiteLaxMode:
		s = gCookie.CookieSameSiteLaxMode
	case http.SameSiteStrictMode:
		s = gCookie.CookieSameSiteStrictMode
	case http.SameSiteNoneMode:
		s = gCookie.CookieSameSiteNoneMode
	}
	return s
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
		logger.Info("[lb-http] checkRetry error *url.Error:", ok)
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
