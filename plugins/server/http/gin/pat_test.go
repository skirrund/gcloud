package gin

import (
	"context"
	"fmt"
	"testing"

	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/server/http"
)

func TestXxx(t *testing.T) {
	urlStr := "https://127.0.0.1:8080/test"
	//urlStr = "http://www.baidu.com"
	resp, err := http.DefaultH2CClient.GetUrl(urlStr, nil, nil, nil)
	fmt.Println(resp.Protocol, ":", err)
}

func TestGinServer(t *testing.T) {
}

func DemoTrace(ctx context.Context) {
	logger.InfoContext(ctx, "demo-----")
}
