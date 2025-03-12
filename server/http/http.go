package http

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/skirrund/gcloud/bootstrap/env"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/server/lb"
	"github.com/skirrund/gcloud/server/request"
	"github.com/skirrund/gcloud/server/response"
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
)

func getRequest(url string, method string, headers map[string]string, params io.Reader, isJson bool, timeOut time.Duration) *request.Request {
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

func getRequestLb(serviceName string, path string, method string, headers map[string]string, params io.Reader, isJson bool, timeOut time.Duration) *request.Request {
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

func getJSONData(params any) io.Reader {
	var reader io.Reader
	if p, ok := params.(string); ok {
		reader = strings.NewReader(p)
	} else if b, ok := params.([]byte); ok {
		reader = bytes.NewReader(b)
	} else {
		body, _ := utils.Marshal(params)
		if env.GetInstance().GetBool(HTTP_LOG_ENABLE_KEY) {
			logger.Info("[http] getJSONData:", logger.GetLogStr(string(body)))
		}
		reader = bytes.NewReader(body)
	}
	return reader
}

func getFormData(params map[string]interface{}) io.Reader {
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
	return strings.NewReader(valuesStr)
}

func getMultipartFormData(params map[string]any, files map[string]*request.File) (reader io.Reader, contentType string) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	var err error
	defer bodyWriter.Close()
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
	return bodyBuf, bodyWriter.FormDataContentType()
}

func getUrlWithParams(urlStr string, params map[string]interface{}) string {
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

func GetUrl(url string, headers map[string]string, params map[string]any, result any) (*response.Response, error) {
	return GetUrlWithTimeout(url, headers, params, result, default_timeout)
}
func GetUrlWithTimeout(url string, headers map[string]string, params map[string]any, result any, timeout time.Duration) (*response.Response, error) {
	req := getRequest(getUrlWithParams(url, params), http.MethodGet, headers, nil, false, timeout)
	return lb.GetInstance().Run(req, result)
}

func Get(serviceName string, path string, headers map[string]string, params map[string]any, result any) (*response.Response, error) {
	return GetWithTimeout(serviceName, path, headers, params, result, default_timeout)
}

func GetWithTimeout(serviceName string, path string, headers map[string]string, params map[string]any, result any, timeout time.Duration) (*response.Response, error) {
	req := getRequestLb(serviceName, getUrlWithParams(path, params), http.MethodGet, headers, nil, false, timeout)
	return lb.GetInstance().Run(req, result)
}

func PostUrl(url string, headers map[string]string, params map[string]any, result any) (*response.Response, error) {
	return PostUrlWithTimeout(url, headers, params, result, default_timeout)
}

func PostUrlWithTimeout(url string, headers map[string]string, params map[string]any, result any, timeout time.Duration) (*response.Response, error) {
	req := getRequest(url, http.MethodPost, headers, getFormData(params), false, timeout)
	return lb.GetInstance().Run(req, result)
}
func PostFormDataUrl(url string, headers map[string]string, params url.Values, result any) (*response.Response, error) {
	return PostFormDataUrlWithTimeout(url, headers, params, result, default_timeout)
}
func PostFormDataUrlWithTimeout(url string, headers map[string]string, params url.Values, result any, timeout time.Duration) (*response.Response, error) {
	reader := strings.NewReader(params.Encode())
	req := getRequest(url, http.MethodPost, headers, reader, false, timeout)
	return lb.GetInstance().Run(req, result)
}
func PostFile(url string, headers map[string]string, params map[string]any, files map[string]*request.File, result any) (*response.Response, error) {
	return PostFileWithTimeout(url, headers, params, files, result, default_timeout)
}

func PostFileWithTimeout(url string, headers map[string]string, params map[string]any, files map[string]*request.File, result any, timeout time.Duration) (*response.Response, error) {
	reader, ct := getMultipartFormData(params, files)
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = ct
	req := getRequest(url, http.MethodPost, headers, reader, false, timeout)
	req.HasFile = true
	return lb.GetInstance().Run(req, result)
}

func Post(serviceName string, path string, headers map[string]string, params map[string]any, result any) (*response.Response, error) {
	return PostWithTimeout(serviceName, path, headers, params, result, default_timeout)
}

func PostWithTimeout(serviceName string, path string, headers map[string]string, params map[string]any, result any, timeout time.Duration) (*response.Response, error) {
	req := getRequestLb(serviceName, path, http.MethodPost, headers, getFormData(params), false, timeout)
	return lb.GetInstance().Run(req, result)
}

func PostFormData(serviceName string, path string, headers map[string]string, params url.Values, result any) (*response.Response, error) {
	return PostFormDataWithTimeout(serviceName, path, headers, params, result, default_timeout)
}

func PostFormDataWithTimeout(serviceName string, path string, headers map[string]string, params url.Values, result any, timeout time.Duration) (*response.Response, error) {
	reader := strings.NewReader(params.Encode())
	req := getRequestLb(serviceName, path, http.MethodPost, headers, reader, false, timeout)
	return lb.GetInstance().Run(req, result)
}

func PostJSONUrl(url string, headers map[string]string, params any, result any) (*response.Response, error) {
	return PostJSONUrlWithTimeout(url, headers, params, result, default_timeout)
}

func PostJSONUrlWithTimeout(url string, headers map[string]string, params any, result any, timeout time.Duration) (*response.Response, error) {
	reader := getJSONData(params)
	req := getRequest(url, http.MethodPost, headers, reader, true, timeout)
	return lb.GetInstance().Run(req, result)
}

func PostJSON(serviceName string, path string, headers map[string]string, params any, result any) (*response.Response, error) {
	return PostJSONWithTimeout(serviceName, path, headers, params, result, default_timeout)
}

func PostJSONWithTimeout(serviceName string, path string, headers map[string]string, params any, result any, timeout time.Duration) (*response.Response, error) {
	reader := getJSONData(params)
	req := getRequestLb(serviceName, path, http.MethodPost, headers, reader, true, timeout)
	return lb.GetInstance().Run(req, result)
}
