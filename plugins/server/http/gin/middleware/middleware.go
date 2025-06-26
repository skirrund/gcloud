package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/utils/worker"
)

const MAX_PRINT_BODY_LEN = 2048

var reg = regexp.MustCompile(`.*\.(js|css|png|jpg|jpeg|gif|svg|webp|bmp|html|htm).*$`)

func Cors(c *gin.Context) {
	method := c.Request.Method
	origin := c.Request.Header.Get("Origin")
	//接收客户端发送的origin （重要！）
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Access-Control-Allow-Credentials", "true")
	if method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
	} else {
		c.Next()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	bodyBuf *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	//memory copy here!
	w.bodyBuf.Write(b)
	return w.ResponseWriter.Write(b)
}

func requestEnd(uri string, contentType string, method string, start time.Time, strBody string, reqBody, status, respCt string) {
	if strings.HasPrefix(uri, "/metrics") {
		strBody = "ignore..."
	}
	if strings.HasPrefix(uri, "/swagger") {
		return
	}
	if reg.MatchString(uri) {
		return
	}
	logger.Info("\n [GIN] uri:", uri,
		"\n [GIN] content-type:", contentType,
		"\n [GIN] method:", method,
		"\n [GIN] body:"+reqBody,
		"\n [GIN] status:"+status,
		"\n [GIN] response-content-type:"+respCt,
		"\n [GIN] response:"+strBody,
		"\n [GIN] cost:"+strconv.FormatInt(time.Since(start).Milliseconds(), 10)+"ms")
}

func LoggingMiddleware(ctx *gin.Context) {
	start := time.Now()
	blw := bodyLogWriter{bodyBuf: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
	ctx.Writer = blw
	bb, err := io.ReadAll(ctx.Request.Body)
	if err == nil {
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bb))
	}
	ctx.Next()
	strBody := strings.Trim(blw.bodyBuf.String(), "\n")
	if utf8.RuneCountInString(strBody) > MAX_PRINT_BODY_LEN {
		strBody = strBody[:(MAX_PRINT_BODY_LEN - 1)]
	}
	if len(bb) > MAX_PRINT_BODY_LEN {
		bb = bb[:(MAX_PRINT_BODY_LEN - 1)]
	}
	req := ctx.Request
	uri := req.RequestURI
	uri1, _ := url.QueryUnescape(uri)
	ct := req.Header.Get("content-type")
	method := req.Method
	status := ctx.Writer.Status()
	respCt := ctx.Writer.Header().Get("Content-Type")
	worker.AsyncExecute(func() {
		requestEnd(uri1, ct, method, start, strBody, string(bb), strconv.FormatInt(int64(status), 10), respCt)
	})
}
