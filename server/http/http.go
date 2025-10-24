package http

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/skirrund/gcloud/bootstrap/env"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/server/lb"
	"github.com/skirrund/gcloud/server/request"
	"github.com/skirrund/gcloud/server/response"
	"github.com/skirrund/gcloud/tracer"
	"github.com/skirrund/gcloud/utils"
	"github.com/skirrund/gcloud/utils/decimal"
)

const (
	// PROTOCOL_HTTP                   = "http://"
	// PROTOCOL_HTTPS                  = "https://"
	default_timeout     = 10 * time.Second
	HTTP_LOG_ENABLE_KEY = "server.http.log.params"
	// WRITE_TIMEOUT_PROPERTIES        = "server.http.writeTimeout"
	// READ_TIMEOUT_PROPERTIES         = "server.http.readTimeout"
	// RetryOnConnectionFailure        = "server.http.retry.onConnectionFailure"
	// RetryEnabled                    = "server.http.retry.enabled"
	// RetryOnAllOperations            = "server.http.retry.allOperations"
	// maxRetriesOnNextServiceInstance = "server.http.retry.maxRetriesOnNextServiceInstance"
	// RetryableStatusCodes            = "server.http.retry.retryableStatusCodes"
	// RetryTimes                      = "server.http.retry.times"
	ContentTypeJson               = "application/json;charset=utf-8"
	ContentTypeXWWWFormUrlencoded = "application/x-www-form-urlencoded;charset=utf-8"
	ContentTypeText               = "text/plain;charset=utf-8"
	ContentTypehtml               = "text/html;charset=utf-8"
)

type GHttp struct {
	ctx context.Context
	H2C bool
}

var DefaultClient GHttp

var DefaultH2CClient = GHttp{H2C: true}

func (h GHttp) WithTracerContext(ctx context.Context) GHttp {
	if ctx != nil {
		ctx = tracer.WithTraceID(ctx)
	} else {
		ctx = tracer.NewTraceIDContext()
	}
	return GHttp{ctx: ctx, H2C: h.H2C}
}

func getRequest(url string, method string, headers map[string]string, params []byte, isJson bool, timeOut time.Duration) *request.Request {
	return &request.Request{
		Url:       url,
		Method:    method,
		Headers:   headers,
		Params:    params,
		IsJson:    isJson,
		TimeOut:   timeOut,
		LbOptions: request.NewDefaultLbOptions(),
	}
}

func getRequestLb(serviceName string, path string, method string, headers map[string]string, params []byte, isJson bool, timeOut time.Duration) *request.Request {
	return &request.Request{
		ServiceName: serviceName,
		Path:        path,
		Method:      method,
		Headers:     headers,
		Params:      params,
		IsJson:      isJson,
		TimeOut:     timeOut,
		LbOptions:   request.NewDefaultLbOptions(),
	}
}

func getJSONData(params any) []byte {
	var reader []byte
	if p, ok := params.(string); ok {
		reader = []byte(p)
	} else if b, ok := params.([]byte); ok {
		reader = b
	} else {
		body, _ := utils.Marshal(params)
		if env.GetInstance().GetBool(HTTP_LOG_ENABLE_KEY) {
			logger.Info("[http] getJSONData:", logger.GetLogStr(string(body)))
		}
		reader = body
	}
	return reader
}

func getFormData(params map[string]any) []byte {
	var values url.Values = make(map[string][]string)
	log := env.GetInstance().GetBool(HTTP_LOG_ENABLE_KEY)
	for k, v := range params {
		t := reflect.TypeOf(v)
		val := reflect.ValueOf(v)
		kind := t.Kind()
		switch kind {
		case reflect.Array:
		case reflect.Slice:
			l := val.Len()
			for i := 0; i < l; i++ {
				value := val.Index(i)
				kind1 := value.Kind()
				switch kind1 {
				case reflect.String:
					values.Add(k, value.String())
				case reflect.Bool:
					values.Add(k, strconv.FormatBool(value.Bool()))
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					values.Add(k, strconv.FormatInt(value.Int(), 10))
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					values.Add(k, strconv.FormatUint(value.Uint(), 10))
				case reflect.Float32, reflect.Float64:
					values.Add(k, decimal.NewFromFloat(value.Float()).String())
				default:
					values.Add(k, value.String())
				}
			}
		case reflect.String:
			values.Add(k, val.String())
		default:
			s, err := utils.MarshalToString(v)
			if err == nil {
				values.Add(k, s)
			}
		}
	}
	valuesStr := values.Encode()
	if log {
		logger.Info("[http] getFormData string:", valuesStr)
	}
	return []byte(valuesStr)
}

func getMultipartFormData(params map[string]any, files map[string]*request.File) (reader []byte, contentType string) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	var err error
	log := env.GetInstance().GetBool(HTTP_LOG_ENABLE_KEY)
	if len(files) > 0 {
		var reader io.Reader
		for k, v := range files {
			if v == nil {
				continue
			}
			fn := v.FileName
			if len(fn) == 0 {
				if v.File != nil {
					fn = filepath.Base(v.File.Name())
				}
			}
			if len(v.FileBytes) > 0 {
				reader = bytes.NewReader(v.FileBytes)
			} else {
				reader = v.File
			}
			w, err := bodyWriter.CreateFormFile(k, fn)
			if err != nil {
				logger.Error("[http] getMultipartFormData error:", err)
				continue
			}
			if wr, err := io.Copy(w, reader); err == nil {
				if log {
					logger.Info("[http] getMultipartFormData:", wr/1024)
				}
			} else {
				logger.Error("[http] getMultipartFormData err:", err)
			}
		}
	}

	for k, v := range params {
		if v == nil {
			continue
		}
		if val, ok := v.(string); ok {
			if len(val) > 0 {
				err = bodyWriter.WriteField(k, val)
				if log {
					logger.Info("[http] getMultipartFormData:", k, ":", logger.GetLogStr(string(val)))
				}
			}
		} else if val, ok := v.(*string); ok {
			if val != nil && len(*val) > 0 {
				err = bodyWriter.WriteField(k, *val)
			}
		} else {
			str, err := utils.MarshalToString(v)
			if err != nil {
				logger.Error("[http] getMultipartFormData error:", err)
				continue
			}
			if len(str) == 0 || str == "[]" || str == "{}" {
				continue
			}
			err = bodyWriter.WriteField(k, str)
			if err != nil {
				logger.Error("[http] getMultipartFormData error:", err)
				continue
			}
			if log {
				logger.Info("[http] getMultipartFormData:", k, ":", logger.GetLogStr(string(str)))
			}
		}
		if err != nil {
			logger.Error("[http] getMultipartFormData error:", err)
		}
	}
	bodyWriter.Close()
	return bodyBuf.Bytes(), bodyWriter.FormDataContentType()
}

func getUrlWithParams(urlStr string, params map[string]any) string {
	if env.GetInstance().GetBool(HTTP_LOG_ENABLE_KEY) {
		logger.Info("[http] getUrlWithParams:", params)
	}
	url1, err := url.Parse(urlStr)
	if err != nil {
		logger.Error("[http] getUrlWithParams error", err.Error())
		return urlStr
	}
	vals := url1.Query()
	if len(params) > 0 {
		for k, v := range params {
			if s, ok := v.(string); ok {
				vals.Add(k, s)
			} else {
				s, err := utils.MarshalToString(v)
				if err == nil {
					vals.Add(k, s)
				}
			}
		}
	}
	url1.RawQuery = vals.Encode()
	return url1.String()
}

func getUrlWithParams2(urlStr string, params url.Values) string {
	if env.GetInstance().GetBool(HTTP_LOG_ENABLE_KEY) {
		logger.Info("[http] getUrlWithParams:", params)
	}
	url1, err := url.Parse(urlStr)
	if err != nil {
		logger.Error("[http] getUrlWithParams error", err.Error())
		return urlStr
	}
	vals := url1.Query()
	if len(params) > 0 {
		for k, v := range params {
			if len(v) > 0 {
				vals.Add(k, v[0])
			}
		}
	}
	url1.RawQuery = vals.Encode()
	return url1.String()
}

func (h GHttp) GetUrl(url string, headers map[string]string, params map[string]any, result any) (*response.Response, error) {
	return h.GetUrlWithTimeout(url, headers, params, result, default_timeout)
}
func (h GHttp) GetUrlWithTimeout(url string, headers map[string]string, params map[string]any, result any, timeout time.Duration) (*response.Response, error) {
	req := getRequest(getUrlWithParams(url, params), http.MethodGet, headers, nil, false, timeout)
	req.WithContext(h.ctx)
	req.H2C = h.H2C
	return lb.GetInstance().Run(req, result)
}

func (h GHttp) Get(serviceName string, path string, headers map[string]string, params map[string]any, result any) (*response.Response, error) {
	return h.GetWithTimeout(serviceName, path, headers, params, result, default_timeout)
}

func (h GHttp) GetWithTimeout(serviceName string, path string, headers map[string]string, params map[string]any, result any, timeout time.Duration) (*response.Response, error) {
	req := getRequestLb(serviceName, getUrlWithParams(path, params), http.MethodGet, headers, nil, false, timeout)
	req.WithContext(h.ctx)
	req.H2C = h.H2C
	return lb.GetInstance().Run(req, result)
}

// 同PostUrlWithTimeout
func (h GHttp) PostUrl(url string, headers map[string]string, params map[string]any, result any) (*response.Response, error) {
	return h.PostUrlWithTimeout(url, headers, params, result, default_timeout)
}

// 通用请求方法
func (h GHttp) DoUrl(urlStr, method, contentType string, headers map[string]string, queryParams url.Values, body []byte, result any, timeout time.Duration) (*response.Response, error) {
	urlStr = getUrlWithParams2(urlStr, queryParams)
	if headers == nil {
		headers = make(map[string]string)
	}
	if len(contentType) > 0 {
		headers["Content-Type"] = contentType
	}
	req := getRequest(urlStr, method, headers, body, false, timeout)
	req.WithContext(h.ctx)
	req.H2C = h.H2C
	return lb.GetInstance().Run(req, result)
}

func (h GHttp) PostUrlWithTimeout(url string, headers map[string]string, params map[string]any, result any, timeout time.Duration) (*response.Response, error) {
	req := getRequest(url, http.MethodPost, headers, getFormData(params), false, timeout)
	req.WithContext(h.ctx)
	req.H2C = h.H2C
	return lb.GetInstance().Run(req, result)
}
func (h GHttp) PostFormDataUrl(url string, headers map[string]string, params url.Values, result any) (*response.Response, error) {
	return h.PostFormDataUrlWithTimeout(url, headers, params, result, default_timeout)
}
func (h GHttp) PostFormDataUrlWithTimeout(url string, headers map[string]string, params url.Values, result any, timeout time.Duration) (*response.Response, error) {
	//reader := strings.NewReader(params.Encode())
	req := getRequest(url, http.MethodPost, headers, []byte(params.Encode()), false, timeout)
	req.WithContext(h.ctx)
	req.H2C = h.H2C
	return lb.GetInstance().Run(req, result)
}
func (h GHttp) PostFile(url string, headers map[string]string, params map[string]any, files map[string]*request.File, result any) (*response.Response, error) {
	return h.PostFileWithTimeout(url, headers, params, files, result, default_timeout)
}

func (h GHttp) PostFileWithTimeout(url string, headers map[string]string, params map[string]any, files map[string]*request.File, result any, timeout time.Duration) (*response.Response, error) {
	reader, ct := getMultipartFormData(params, files)
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = ct
	req := getRequest(url, http.MethodPost, headers, reader, false, timeout)
	req.HasFile = true
	req.WithContext(h.ctx)
	req.H2C = h.H2C
	return lb.GetInstance().Run(req, result)
}

func (h GHttp) Post(serviceName string, path string, headers map[string]string, params map[string]any, result any) (*response.Response, error) {
	return h.PostWithTimeout(serviceName, path, headers, params, result, default_timeout)
}

func (h GHttp) PostWithTimeout(serviceName string, path string, headers map[string]string, params map[string]any, result any, timeout time.Duration) (*response.Response, error) {
	req := getRequestLb(serviceName, path, http.MethodPost, headers, getFormData(params), false, timeout)
	req.WithContext(h.ctx)
	req.H2C = h.H2C
	return lb.GetInstance().Run(req, result)
}

func (h GHttp) PostFormData(serviceName string, path string, headers map[string]string, params url.Values, result any) (*response.Response, error) {
	return h.PostFormDataWithTimeout(serviceName, path, headers, params, result, default_timeout)
}

func (h GHttp) PostFormDataWithTimeout(serviceName string, path string, headers map[string]string, params url.Values, result any, timeout time.Duration) (*response.Response, error) {
	// reader := strings.NewReader(params.Encode())
	req := getRequestLb(serviceName, path, http.MethodPost, headers, []byte(params.Encode()), false, timeout)
	req.WithContext(h.ctx)
	req.H2C = h.H2C
	return lb.GetInstance().Run(req, result)
}

func (h GHttp) PostJSONUrl(url string, headers map[string]string, params any, result any) (*response.Response, error) {
	return h.PostJSONUrlWithTimeout(url, headers, params, result, default_timeout)
}

func (h GHttp) PostJSONUrlWithTimeout(url string, headers map[string]string, params any, result any, timeout time.Duration) (*response.Response, error) {
	reader := getJSONData(params)
	req := getRequest(url, http.MethodPost, headers, reader, true, timeout)
	req.WithContext(h.ctx)
	req.H2C = h.H2C
	return lb.GetInstance().Run(req, result)
}

func (h GHttp) PostJSON(serviceName string, path string, headers map[string]string, params any, result any) (*response.Response, error) {
	return h.PostJSONWithTimeout(serviceName, path, headers, params, result, default_timeout)
}

func (h GHttp) PostJSONWithTimeout(serviceName string, path string, headers map[string]string, params any, result any, timeout time.Duration) (*response.Response, error) {
	reader := getJSONData(params)
	req := getRequestLb(serviceName, path, http.MethodPost, headers, reader, true, timeout)
	req.WithContext(h.ctx)
	req.H2C = h.H2C
	return lb.GetInstance().Run(req, result)
}
