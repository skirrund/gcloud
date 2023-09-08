package client

import (
	"github.com/skirrund/gcloud/server/request"
	"github.com/skirrund/gcloud/server/response"
)

type HttpClient interface {
	Exec(req *request.Request) (resp *response.Response, err error)
	CheckRetry(err error, status int) bool
}
