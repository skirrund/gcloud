package feign

import (
	"time"

	"github.com/skirrund/gcloud/server/http"
	"github.com/skirrund/gcloud/server/response"
)

type Client struct {
	ServiceName string
	Url         string
}

const (
	defaultTimeOut = 10 * time.Second
)

func (c *Client) Get(path string, headers map[string]string, params map[string]any, result any) (*response.Response, error) {
	return c.GetWithTimeout(path, headers, params, result, defaultTimeOut)
}
func (c *Client) GetWithTimeout(path string, headers map[string]string, params map[string]any, result any, timeOut time.Duration) (*response.Response, error) {
	if len(c.Url) > 0 {
		return http.DefaultClient.GetUrlWithTimeout(c.Url+path, headers, params, result, timeOut)
	}
	return http.DefaultClient.GetWithTimeout(c.ServiceName, path, headers, params, result, timeOut)
}
func (c *Client) PostJSON(path string, headers map[string]string, params any, result any) (*response.Response, error) {
	return c.PostJSONWithTimeout(path, headers, params, result, defaultTimeOut)
}

func (c *Client) PostJSONWithTimeout(path string, headers map[string]string, params any, result any, timeOut time.Duration) (*response.Response, error) {
	if len(c.Url) > 0 {
		return http.DefaultClient.PostJSONUrlWithTimeout(c.Url+path, headers, params, result, timeOut)
	}
	return http.DefaultClient.PostJSONWithTimeout(c.ServiceName, path, headers, params, result, timeOut)
}
func (c *Client) Post(path string, headers map[string]string, params map[string]any, result any) (*response.Response, error) {
	return c.PostWithTimeout(path, headers, params, result, defaultTimeOut)
}

func (c *Client) PostWithTimeout(path string, headers map[string]string, params map[string]any, result any, timeOut time.Duration) (*response.Response, error) {
	if len(c.Url) > 0 {
		return http.DefaultClient.PostUrlWithTimeout(c.Url+path, headers, params, result, timeOut)
	}
	return http.DefaultClient.PostWithTimeout(c.ServiceName, path, headers, params, result, timeOut)
}
