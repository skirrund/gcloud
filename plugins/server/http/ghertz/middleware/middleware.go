package middleware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/response"
	"github.com/skirrund/gcloud/utils/worker"
)

const MAX_PRINT_BODY_LEN = 2048

var reg = regexp.MustCompile(`.*\.(js|css|png|jpg|jpeg|gif|svg|webp|bmp|html|htm).*$`)

func Cors(ctx context.Context, c *app.RequestContext) {
	request := c.GetRequest()
	method := string(request.Method())
	origin := string(request.Header.Peek("Origin"))
	//接收客户端发送的origin （重要！）

	request.Header.Set("Access-Control-Allow-Origin", origin)
	request.Header.Set("Access-Control-Allow-Methods", "*")
	request.Header.Set("Access-Control-Allow-Headers", "*")
	request.Header.Set("Access-Control-Allow-Credentials", "true")
	if method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
		return
	} else {
		c.Next(ctx)
	}
}

func getLogBodyStr(bb []byte) string {
	if len(bb) > MAX_PRINT_BODY_LEN {
		bb = bb[:(MAX_PRINT_BODY_LEN - 1)]
	}
	return string(bb)
}

func MyRecoveryHandler(c context.Context, ctx *app.RequestContext, err interface{}, stack []byte) {
	logger.Error("[Hertz] recover:", err, "\n", string(stack))
	ctx.Response.Reset()
	ctx.JSON(200, response.Fail[any](fmt.Sprintf("%v", err)))
}

func LoggingMiddleware(c context.Context, ctx *app.RequestContext) {
	start := time.Now()
	reqBody := getLogBodyStr(ctx.Request.Body())
	req := ctx.GetRequest()
	uri := string(req.URI().FullURI())
	sUri := string(req.URI().RequestURI())
	ct := string(req.Header.ContentType())
	method := string(req.Header.Method())
	ctx.Next(c)
	if strings.HasPrefix(sUri, "/metrics") {
		return
	}
	if strings.HasPrefix(sUri, "/swagger") {
		return
	}
	if reg.MatchString(sUri) {
		return
	}
	rResp := ctx.GetResponse()
	bb := rResp.Body()
	respStatus := rResp.StatusCode()
	respBody := getLogBodyStr(bb)
	worker.AsyncExecute(func() {
		requestEnd(uri, ct, method, reqBody, respBody, strconv.FormatInt(int64(respStatus), 10), start)
	})
}

func requestEnd(uri, ct, method, reqBody, respBody, respStatus string, start time.Time) {
	logger.Info("\n [Hertz] uri:", uri,
		"\n [Hertz] content-type:", ct,
		"\n [Hertz] method:", method,
		"\n [Hertz] body:"+reqBody,
		"\n [Hertz] status:"+respStatus,
		"\n [Hertz] response:"+respBody,
		"\n [Hertz] cost:"+strconv.FormatInt(time.Since(start).Milliseconds(), 10)+"ms")
}
