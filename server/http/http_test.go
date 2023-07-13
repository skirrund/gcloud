package http

import (
	"testing"

	"github.com/skirrund/gcloud/server/lb"
)

func TestGet(t *testing.T) {
	var resp []byte
	lb.GetInstance().SetHttpClient(lb.FastHttpClient{})
	GetUrl("http://127.0.0.1:32766/v1/article/test?t1=t1&t2=t2&t3=t3", nil, nil, &resp)
}
