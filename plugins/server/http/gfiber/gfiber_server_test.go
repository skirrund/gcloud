package gfiber

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/skirrund/gcloud/response"
	"github.com/skirrund/gcloud/server"
	"github.com/skirrund/gcloud/utils/validator"
)

type Test struct {
	Id  int64 `validate:"required,gte=6" `
	Id2 int64 `validate:"required,gte=2" `
}

func TestFiberServer(t *testing.T) {
	options := server.Options{
		ServerName: "fiber_test",
		Address:    ":8080",
	}
	srv := NewServer(options, func(engine *fiber.App) {
		engine.Post("/test", func(context *fiber.Ctx) error {
			d := &Test{}
			if err := ShouldBindBody(context, d); err != nil {
				context.JSON(response.Fail(validator.ErrResp(err)))
				return nil
			}
			context.JSON("test")
			return nil
		})
	})
	srv.Run(func() {
		fmt.Println("shut down")
	})
}
