package feign

import (
	"github.com/skirrund/gcloud/server/http"
	"github.com/skirrund/gcloud/server/response"
)

type Client struct {
	ServiceName string
	Url         string
}

func (c *Client) Get(path string, headers map[string]string, params map[string]interface{}, result interface{}) (*response.Response, error) {
	if len(c.Url) > 0 {
		return http.GetUrl(c.Url+path, headers, params, result)
	}
	return http.Get(c.ServiceName, path, headers, params, result)
}
func (c *Client) PostJSON(path string, headers map[string]string, params interface{}, result interface{}) (*response.Response, error) {
	if len(c.Url) > 0 {
		return http.PostJSONUrl(c.Url+path, headers, params, result)
	}
	return http.PostJSON(c.ServiceName, path, headers, params, result)
}
func (c *Client) Post(path string, headers map[string]string, params map[string]interface{}, result interface{}) (*response.Response, error) {
	if len(c.Url) > 0 {
		return http.PostUrl(c.Url+path, headers, params, result)
	}
	return http.Post(c.ServiceName, path, headers, params, result)
}
