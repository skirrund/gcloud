package gfiber

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/skirrund/gcloud/server"
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
			vals := QueryArray(context, "a")
			fmt.Println(vals)

			// d := &Test{}
			// if err := ShouldBindBody(context, d); err != nil {
			// 	context.JSON(response.Fail(validator.ErrResp(err)))
			// 	return nil
			// }
			context.JSON("test")
			return server.Error{Code: "2000000", Msg: "123"}
		})
	})
	srv.Run(func() {
		fmt.Println("shut down")
	})
}
