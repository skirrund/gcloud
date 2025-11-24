package gin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/skirrund/gcloud/logger"
	gm "github.com/skirrund/gcloud/plugins/server/http/gin/middleware"
	"github.com/skirrund/gcloud/plugins/server/http/gin/prometheus"
	"github.com/skirrund/gcloud/response"
	"github.com/skirrund/gcloud/server"
	"github.com/skirrund/gcloud/tracer"
	uValidator "github.com/skirrund/gcloud/utils/validator"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Server struct {
	Srv     *gin.Engine
	Options server.Options
}

const DefaultMaxRequestBodySize = 104857600 // 100MB

func NewServer(options server.Options, routerProvider func(engine *gin.Engine), middleware ...gin.HandlerFunc) server.Server {
	srv := &Server{}
	srv.Options = options
	gin.SetMode(gin.ReleaseMode)
	s := gin.New()
	// if options.H2C {
	// 	s.UseH2C = true
	// }
	s.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("[GIN] recover:", recovered, "\n", string(debug.Stack()))

		c.JSON(200, response.Fail[any](fmt.Sprintf("%v", recovered)))
		//		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	//s.Use(cors)
	//s.Use(zipkinMiddleware)
	s.Use(gm.TraceMiddleware, gm.LoggingMiddleware)
	//zipkin.InitZipkinTracer(s)
	gp := prometheus.New(s)
	s.Use(gp.Middleware())
	if len(middleware) > 0 {
		s.Use(middleware...)
	}
	// metrics采样
	s.GET("/metrics", gin.WrapH(promhttp.Handler()))
	//s.Use(sentinelMiddleware)
	//initSwagger(s)

	pprof.Register(s)
	routerProvider(s)
	srv.Srv = s
	return srv
}

func (server *Server) Shutdown() {
}

func (server *Server) GetServeServer() any {
	return server.Srv
}

func (server *Server) Run(graceful ...func()) {
	srv := &http.Server{
		Addr:         server.Options.Address,
		Handler:      server.Srv.Handler(),
		ReadTimeout:  4 * time.Minute,
		WriteTimeout: 4 * time.Minute,
	}
	if server.Options.H2C {
		h2s := &http2.Server{}
		if server.Options.MaxConcurrentStreams > 0 {
			h2s.MaxConcurrentStreams = server.Options.MaxConcurrentStreams
		} else {
			h2s.MaxConcurrentStreams = 256
		}
		h2s.IdleTimeout = 15 * time.Second
		srv.Handler = h2c.NewHandler(server.Srv, h2s)
	}

	bodySize := server.Options.MaxRequestBodySize
	if bodySize > 0 {
		srv.MaxHeaderBytes = bodySize
	} else {
		srv.MaxHeaderBytes = DefaultMaxRequestBodySize
	}
	go func() {
		logger.Info("[GIN] server starting on:", server.Options.Address, " h2c:", server.Options.H2C)
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
		grace(server, graceful...)
		logger.Panic("[GIN]Server forced to shutdown:", err)
		return
	}
	grace(server, graceful...)
	logger.Info("[GIN]server has been shutdown")
}

func grace(server *Server, g ...func()) {
	server.Shutdown()
	if len(g) > 0 {
		var wg sync.WaitGroup
		for _, f := range g {
			wg.Go(f)
		}
		wg.Wait()
	}

}

// InitTrans 初始化翻译器
func InitTrans(locale string, validate binding.StructValidator) (err error) {
	//修改gin框架中的Validator属性，实现自定制
	if v, ok := validate.Engine().(*validator.Validate); ok {
		return uValidator.InitValidator(locale, v)
	}
	return
}

func GetTraceContext(ctx *gin.Context) context.Context {
	id := ctx.GetString(tracer.TraceIDKey)
	return tracer.NewContextFromTraceId(id)
}
