package client

import "github.com/skirrund/gcloud/server/request"

type HttpClient interface {
	Exec(req *request.Request) (statusCode int, err error)
}
