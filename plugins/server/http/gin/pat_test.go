package gin

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/server"
)

func TestXxx(t *testing.T) {
	reg := regexp.MustCompile(`.*\.(js|css|png|jpg|jpeg|gif).*$`)
	t.Log(">>>>>>", reg.MatchString("http://123.com/js.png"))
}

func TestGinServer(t *testing.T) {
	options := server.Options{
		ServerName: "gin_test",
		Address:    ":8080",
	}
	srv := NewServer(options, func(engine *gin.Engine) {
		engine.POST("/test", func(context *gin.Context) {
			//v := context.QueryArray("a")
			// fmt.Println(v)
			DemoTrace(GetTraceContext(context))
			context.JSON(200, "test")
		})
	})
	srv.Run(func() {
		fmt.Println("shut down")
	})
}

func DemoTrace(ctx context.Context) {
	logger.InfoContext(ctx, "demo-----")
}
