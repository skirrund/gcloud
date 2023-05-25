package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/skirrund/gcloud/server"
	"regexp"
	"testing"
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
		engine.GET("/test", func(context *gin.Context) {
			panic("test")
			//		context.JSON(200, "test")
		})
	})
	srv.Run(func() {
		fmt.Println("shut down")
	})
}
