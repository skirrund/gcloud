package gfiber

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/skirrund/gcloud/server"
)

type Test struct {
	Id   int64
	Id2  int64
	Code string
}

type WxAuthReq struct {
	Signature    string `form:"signature" query:"signature" binding:"required,min=1" validate:"required,min=1"`
	Timestamp    string `form:"timestamp" query:"timestamp" binding:"required,min=1" validate:"required,min=1"`
	Nonce        string `form:"nonce" query:"nonce" binding:"required,min=1" validate:"required,min=1"`
	EncryptType  string `form:"encrypt_type" query:"encrypt_type"`
	MsgSignature string `form:"msg_signature" query:"msg_signature"`
}

func TestFiberServer(t *testing.T) {
	options := server.Options{
		ServerName: "fiber_test",
		Address:    ":8080",
	}
	srv := NewServer(options, func(engine *fiber.App) {
		engine.Post("/test", func(context *fiber.Ctx) error {
			req := WxAuthReq{}
			ShouldBindQuery(context, &req)
			fmt.Printf("%+v", req)

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
