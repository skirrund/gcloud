package gin

import (
	"context"
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/server"
	"github.com/skirrund/gcloud/server/http"
)

func TestXxx(t *testing.T) {
	urlStr := "https://127.0.0.1:8080/test"
	//urlStr = "http://www.baidu.com"
	resp, err := http.DefaultH2CClient.GetUrl(urlStr, nil, nil, nil)
	fmt.Println(resp.Protocol, ":", err)
}

func TestGinServer(t *testing.T) {
	options := server.Options{
		ServerName: "gin_test",
		Address:    ":8080",
		H2C:        true,
	}
	srv := NewServer(options, func(engine *gin.Engine) {
		engine.GET("/test", func(context *gin.Context) {
			//v := context.QueryArray("a")
			// fmt.Println(v)
			context.String(200, "123")
		})
	})
	srv.Run(func() {
		fmt.Println("shut down")
	})
}

func DemoTrace(ctx context.Context) {
	logger.InfoContext(ctx, "demo-----")
}
