package feign

import (
	"github.com/skirrund/gcloud/server/http"
)

type Client struct {
	ServiceName string
}

func (c *Client) Get(path string, params map[string]interface{}, result interface{}) (int, error) {
	return http.Get(c.ServiceName, path, params, result)
}
func (c *Client) PostJSON(path string, params interface{}, result interface{}) (int, error) {
	return http.PostJSON(c.ServiceName, path, params, result)
}
func (c *Client) PostJSONWithHeader(path string, headers map[string]string, params interface{}, result interface{}) (int, error) {
	return http.PostJSONWithHeader(c.ServiceName, path, headers, params, result)
}
func (c *Client) Post(path string, params map[string]interface{}, result interface{}) (int, error) {
	return http.Post(c.ServiceName, path, params, result)
}
func (c *Client) PostWithHeader(path string, headers map[string]string, params map[string]interface{}, result interface{}) (int, error) {
	return http.PostWithHeader(c.ServiceName, path, headers, params, result)
}
