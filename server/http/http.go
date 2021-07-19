package http

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/skirrund/gcloud/bootstrap/env"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/server/lb"
	"github.com/skirrund/gcloud/server/request"
	"github.com/skirrund/gcloud/utils"
)

const (
	// PROTOCOL_HTTP                   = "http://"
	// PROTOCOL_HTTPS                  = "https://"
	DEFAULT_TIMEOUT     = 10
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

func getRequest(url string, method string, headers map[string]string, params io.Reader, isJson bool, respResult interface{}, timeOut time.Duration) *request.Request {
	return &request.Request{
		Url:        url,
		Method:     method,
		Headers:    headers,
		Params:     params,
		IsJson:     isJson,
		RespResult: respResult,
		TimeOut:    timeOut,
		LbOptions:  request.NewDefaultLbOptions(),
	}
}

func getRequestLb(serviceName string, path string, method string, headers map[string]string, params io.Reader, isJson bool, respResult interface{}, timeOut time.Duration) *request.Request {
	return &request.Request{
		ServiceName: serviceName,
		Path:        path,
		Method:      method,
		Headers:     headers,
		Params:      params,
		IsJson:      isJson,
		RespResult:  respResult,
		TimeOut:     timeOut,
		LbOptions:   request.NewDefaultLbOptions(),
	}
}

func getJSONData(params interface{}) io.Reader {
	body, _ := utils.Marshal(params)
	if env.GetInstance().GetBool(HTTP_LOG_ENABLE_KEY) {
		logger.Info("[http] getJSONData:", logger.GetLogStr(string(body)))
	}
	return bytes.NewReader(body)
}

func getFormData(params map[string]interface{}) io.Reader {
	var values url.Values = make(map[string][]string)
	log := env.GetInstance().GetBool(HTTP_LOG_ENABLE_KEY)
	for k, v := range params {
		if s, ok := v.(string); ok {
			values.Add(k, s)
			if log {
				logger.Info("[http] getFormData string:", k, ":", logger.GetLogStr(string(s)))
			}
		} else {
			s, err := utils.MarshalToString(s)
			if err == nil {
				values.Add(k, s)
			}
			if log {
				logger.Info("[http] getFormData:", k, ":", logger.GetLogStr(string(s)))
			}
		}
	}
	return strings.NewReader(values.Encode())
}

func getMultipartFormData(params map[string]interface{}, files map[string]*request.File) (reader io.Reader, contentType string) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	var err error
	defer bodyWriter.Close()
	log := env.GetInstance().GetBool(HTTP_LOG_ENABLE_KEY)
	if files != nil && len(files) > 0 {
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
	var p string
	if params != nil {
		for k, v := range params {
			if s, ok := v.(string); ok {
				p = p + "&" + k + "=" + url.QueryEscape(s)
			} else {
				s, err := utils.MarshalToString(v)
				if err == nil {
					p = p + "&" + k + "=" + url.QueryEscape(s)
				}
			}

		}
	}

	if strings.HasPrefix(p, "&") {
		p = p[1:]
	}
	if strings.Index(urlStr, "?") != -1 {
		urlStr += p
	} else {
		urlStr = urlStr + "?" + p
	}
	return urlStr
}

func GetUrl(url string, params map[string]interface{}, result interface{}) (int, error) {
	req := getRequest(getUrlWithParams(url, params), http.MethodGet, nil, nil, false, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}

func Get(serviceName string, path string, params map[string]interface{}, result interface{}) (int, error) {
	req := getRequestLb(serviceName, getUrlWithParams(path, params), http.MethodGet, nil, nil, false, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}

func PostUrl(url string, params map[string]interface{}, result interface{}) (int, error) {
	req := getRequest(url, http.MethodPost, nil, getFormData(params), false, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}
func PostFile(url string, params map[string]interface{}, files map[string]*request.File, result interface{}) (int, error) {
	reader, ct := getMultipartFormData(params, files)
	headers := map[string]string{"Content-Type": ct}
	req := getRequest(url, http.MethodPost, headers, reader, false, result, DEFAULT_TIMEOUT*time.Second)
	req.HasFile = true
	return lb.GetInstance().Run(req)
}

func Post(serviceName string, path string, params map[string]interface{}, result interface{}) (int, error) {
	req := getRequestLb(serviceName, path, http.MethodPost, nil, getFormData(params), false, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}

func PostWithHeaderUrl(url string, headers map[string]string, params map[string]interface{}, result interface{}) (int, error) {
	req := getRequest(url, http.MethodPost, headers, getFormData(params), false, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}

func PostWithHeader(serviceName string, path string, headers map[string]string, params map[string]interface{}, result interface{}) (int, error) {
	req := getRequestLb(serviceName, path, http.MethodPost, headers, getFormData(params), false, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}

func PostJSONUrl(url string, params interface{}, result interface{}) (int, error) {
	req := getRequest(url, http.MethodPost, nil, getJSONData(params), true, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}

func PostJSONStringUrl(url string, params string, result interface{}) (int, error) {
	req := getRequest(url, http.MethodPost, nil, strings.NewReader(params), true, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}

func PostJSONString(serviceName string, path string, params string, result interface{}) (int, error) {
	req := getRequestLb(serviceName, path, http.MethodPost, nil, strings.NewReader(params), true, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}

func PostJSON(serviceName string, path string, params interface{}, result interface{}) (int, error) {
	req := getRequestLb(serviceName, path, http.MethodPost, nil, getJSONData(params), true, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}

func PostJSONWithHeaderUrl(url string, headers map[string]string, params interface{}, result interface{}) (int, error) {
	req := getRequest(url, http.MethodPost, headers, getJSONData(params), true, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}

func PostJSONWithHeader(serviceName string, path string, headers map[string]string, params interface{}, result interface{}) (int, error) {
	req := getRequestLb(serviceName, path, http.MethodPost, headers, getJSONData(params), true, result, DEFAULT_TIMEOUT*time.Second)
	return lb.GetInstance().Run(req)
}
