package ghertz

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	hertzServer "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/skirrund/gcloud/server"
)

type Test struct {
	Id   int64
	Id2  int64
	Code string
}

func TestHertzServer(t *testing.T) {
	options := server.Options{
		ServerName: "hertz_test",
		Address:    ":8080",
	}
	srv := NewServer(options, func(engine *hertzServer.Hertz) {
		engine.POST("/test", func(c context.Context, ctx *app.RequestContext) {
			//t := &Test{}
			strs := QueryArray(ctx, "t")
			fmt.Println(strs)
			strs = PostFormArray(ctx, "t")
			fmt.Println(strs)
			ctx.JSON(http.StatusOK, strs)
		})
	})
	srv.Run(func() {
		fmt.Println("shut down")
	})
}
