package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/utils/worker"
)

const MAX_PRINT_BODY_LEN = 2048

var reg = regexp.MustCompile(`.*\.(js|css|png|jpg|jpeg|gif|svg|webp|bmp|html|htm).*$`)

func Cors(c *fiber.Ctx) error {
	request := c.Request()
	method := string(request.Header.Method())
	origin := string(request.Header.Peek("Origin"))
	//接收客户端发送的origin （重要！）

	request.Header.Set("Access-Control-Allow-Origin", origin)
	request.Header.Set("Access-Control-Allow-Methods", "*")
	request.Header.Set("Access-Control-Allow-Headers", "*")
	request.Header.Set("Access-Control-Allow-Credentials", "true")
	if method == "OPTIONS" {
		return c.SendStatus(http.StatusOK)
	} else {
		return c.Next()
	}
}

func getLogBodyStr(bb []byte) string {
	if len(bb) > MAX_PRINT_BODY_LEN {
		bb = bb[:(MAX_PRINT_BODY_LEN - 1)]
	}
	return string(bb)
}

func LoggingMiddleware(ctx *fiber.Ctx) error {
	start := time.Now()
	reqBody := getLogBodyStr(ctx.Body())
	req := ctx.Request()
	uri := string(req.URI().FullURI())
	sUri := string(req.URI().RequestURI())
	ct := string(req.Header.ContentType())
	method := string(req.Header.Method())
	err := ctx.Next()
	if strings.HasPrefix(sUri, "/metrics") {
		return err
	}
	if strings.HasPrefix(sUri, "/swagger") {
		return err
	}
	if reg.MatchString(sUri) {
		return err
	}
	bb := ctx.Response().Body()
	respStatus := ctx.Response().StatusCode()
	respBody := getLogBodyStr(bb)
	worker.AsyncExecute(func() {
		requestEnd(uri, ct, method, reqBody, respBody, strconv.FormatInt(int64(respStatus), 10), start)
	})
	return err
}

func requestEnd(uri, ct, method, reqBody, respBody, respStatus string, start time.Time) {
	logger.Info("\n [Fiber] uri:", uri,
		"\n [Fiber] content-type:", ct,
		"\n [Fiber] method:", method,
		"\n [Fiber] body:"+reqBody,
		"\n [Fiber] status:"+respStatus,
		"\n [Fiber] response:"+respBody,
		"\n [Fiber] cost:"+strconv.FormatInt(time.Since(start).Milliseconds(), 10)+"ms")
}
