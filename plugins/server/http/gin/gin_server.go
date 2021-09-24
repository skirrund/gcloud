package gin

import (
	"bytes"
	"context"
	"fmt"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/plugins/server/http/gin/prometheus"
	"github.com/skirrund/gcloud/plugins/zipkin"
	"github.com/skirrund/gcloud/response"
	"github.com/skirrund/gcloud/server"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	Srv     *gin.Engine
	Options server.Options
}

const MAX_PRINT_BODY_LEN = 1024

type bodyLogWriter struct {
	gin.ResponseWriter
	bodyBuf *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	//memory copy here!
	w.bodyBuf.Write(b)
	return w.ResponseWriter.Write(b)
}

func NewServer(options server.Options, routerProvider func(engine *gin.Engine)) server.Server {
	srv := &Server{}
	srv.Options = options
	gin.SetMode(gin.ReleaseMode)
	s := gin.New()
	s.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("[GIN] recover:", recovered)

		c.JSON(200, response.Fail(fmt.Sprintf("%v", recovered)))
		return
		//		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	s.Use(cors)
	s.Use(zipkinMiddleware)
	s.Use(loggingMiddleware)
	//zipkin.InitZipkinTracer(s)
	gp := prometheus.New(s)
	s.Use(gp.Middleware())
	// metrics采样
	s.GET("/metrics", gin.WrapH(promhttp.Handler()))
	s.Use(sentinelMiddleware)
	initSwagger(s)

	pprof.Register(s)
	routerProvider(s)
	srv.Srv = s
	return srv
}

func initSwagger(e *gin.Engine)  {
	e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func cors(c *gin.Context) {
	method := c.Request.Method
	//origin := c.Request.Header.Get("Origin")
	// //接收客户端发送的origin （重要！）
	// header := c.Writer.Header()
	// header.Set("Access-Control-Allow-Origin", "*")
	// //服务器支持的所有跨域请求的方法
	// header.Set("Access-Control-Allow-Methods", "*")
	// //允许跨域设置可以返回其他子段，可以自定义字段
	// c.Header("Access-Control-Allow-Headers", "*")
	// // 允许浏览器（客户端）可以解析的头部 （重要）
	// //c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
	// //设置缓存时间
	// //	c.Header("Access-Control-Max-Age", "172800")
	// //允许客户端传递校验信息比如 cookie (重要)
	// c.Header("Access-Control-Allow-Credentials", "true")
	if method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
	} else {
		c.Next()
	}
}

func sentinelMiddleware(c *gin.Context) {
	var args []interface{}
	rawQuery := c.Request.URL.RawQuery
	if len(rawQuery) > 0 {
		params := strings.Split(rawQuery, "&")
		for _, param := range params {
			kv := strings.Split(param, "=")
			if len(kv) > 1 && len(kv[1]) > 0 {
				args = append(args, kv[1])
			}
		}
	}
	if c.Request.Method == "POST" {
		c.Request.ParseForm()
		for _, v := range c.Request.PostForm {
			args = append(args, v)
		}
	}
	requestUri := c.Request.RequestURI
	if strings.Contains(requestUri, "?") {
		requestUri = requestUri[0:strings.Index(requestUri, "?")]
	}
	entry, b :=  sentinel.Entry(requestUri, sentinel.WithTrafficType(base.Inbound), sentinel.WithArgs(args...))
	if b != nil {
		c.Abort()
		switch b.BlockType() {
		case base.BlockTypeCircuitBreaking:
			c.JSON(200, response.DEGRADE_EXCEPTION)
			return
		case base.BlockTypeFlow:
			c.JSON(200, response.FLOW_EXCEPTION)
			return
		case base.BlockTypeHotSpotParamFlow:
			c.JSON(200, response.PARAM_FLOW_EXCEPTION)
			return
		case base.BlockTypeSystemFlow:
			c.JSON(200, response.SYSTEM_BLOCK_EXCEPTION)
			return
		case base.BlockTypeIsolation:
			c.JSON(200, response.AUTHORITY_EXCEPTION)
			return
		}
	}
	c.Next()
	entry.Exit()
}

func loggingMiddleware(ctx *gin.Context) {
	start := time.Now()
	blw := bodyLogWriter{bodyBuf: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
	ctx.Writer = blw
	bb, err := io.ReadAll(ctx.Request.Body)
	if err == nil {
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bb))
	}
	ctx.Next()
	strBody := strings.Trim(blw.bodyBuf.String(), "\n")
	if len(strBody) > MAX_PRINT_BODY_LEN {
		strBody = strBody[:(MAX_PRINT_BODY_LEN - 1)]
	}
	if len(bb) > MAX_PRINT_BODY_LEN {
		bb = bb[:(MAX_PRINT_BODY_LEN - 1)]
	}
	defer requestEnd(ctx, start, strBody, string(bb))
}

func zipkinMiddleware(c *gin.Context) {
	tracer := zipkin.GetTracer()
	if tracer != nil {
		carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
		clientContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
		// 将tracer注入到gin的中间件中
		var serverSpan opentracing.Span
		if err == nil {
			serverSpan = tracer.StartSpan(
				c.Request.Method+" "+c.FullPath(), opentracing.FollowsFrom(clientContext))
		} else {
			serverSpan = tracer.StartSpan(c.Request.Method + " " + c.FullPath())
		}
		defer serverSpan.Finish()
	}
	c.Next()
}

func requestEnd(ctx *gin.Context, start time.Time, strBody string, reqBody string) {
	req := ctx.Request
	uri, _ := url.QueryUnescape(req.RequestURI)
	if strings.HasPrefix(uri, "/metrics") {
		strBody = "ignore..."
	}
	if strings.HasPrefix(uri, "/swagger") {
		return
	}
	logger.Info("\n [GIN] uri:" + uri + "\n [GIN] method:" + req.Method + "\n [GIN] body:" + reqBody + "\n [GIN] response:" + strBody + "\n [GIN] cost:" + strconv.FormatInt(time.Since(start).Milliseconds(), 10) + "ms")
}

func (server *Server) Shutdown() {
	defer zipkin.Close()
}

func (server *Server) Run(graceful func()) {
	srv := &http.Server{
		Addr:         server.Options.Address,
		Handler:      server.Srv,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	go func() {
		logger.Info("[GIN] server starting on:", server.Options.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Panic("[GIN] listen:", err.Error())
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("[GIN]Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		grace(server, graceful)
		logger.Panic("[GIN]Server forced to shutdown:", err)
	}
	grace(server, graceful)
	logger.Info("[GIN]server has been shutdown")
}

func grace(server *Server, g func()) {
	server.Shutdown()
	g()
}
