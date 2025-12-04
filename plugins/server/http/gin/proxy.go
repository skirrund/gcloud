package gin

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/server/lb"
	"github.com/skirrund/gcloud/server/request"
)

func ProxyService(serviceName, path string, ctx *gin.Context, timeout time.Duration) error {
	logger.Info("[startProxy-gin]:", serviceName, ",path:", path)
	req := ctx.Request
	gRequest := &request.Request{
		ServiceName: serviceName,
		Path:        path,
		Method:      req.Method,
		TimeOut:     timeout,
		IsProxy:     true,
	}
	if bodyBytes, err := io.ReadAll(req.Body); err != nil {
		return err
	} else {
		gRequest.Params = bodyBytes
	}

	ctxHeader := make(map[string]string)
	h := req.Header
	for k, vals := range h {
		if len(vals) > 0 {
			ctxHeader[k] = vals[0]
		}
	}
	gRequest.Headers = ctxHeader
	gRequest.LbOptions = request.NewDefaultLbOptions()
	gresp, err := lb.GetInstance().Run(gRequest, nil)
	if err != nil {
		return err
	}
	proxyResp := req.Response
	if len(gresp.Headers) > 0 {
		for k, v := range gresp.Headers {
			if len(v) == 0 {
				proxyResp.Header.Set(k, v[0])
			} else {
				for _, vv := range v {
					proxyResp.Header.Add(k, vv)
				}
			}
		}
	}
	sc := gresp.StatusCode
	ctx.Status(sc)
	w := ctx.Writer
	w.Write(gresp.Body)
	w.Flush()
	return nil
}
