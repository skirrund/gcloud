package response

import "github.com/skirrund/gcloud/server/http/cookie"

type Response struct {
	Body        []byte
	ContentType string
	Cookies     map[string]*cookie.Cookie
	Headers     map[string][]string
	StatusCode  int
	Protocol    string
}
