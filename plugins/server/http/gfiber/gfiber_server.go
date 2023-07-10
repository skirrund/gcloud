package gfiber

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/response"
	"github.com/skirrund/gcloud/server"
	"github.com/skirrund/gcloud/utils/validator"
	"github.com/skirrund/gcloud/utils/worker"
)

type Server struct {
	Srv     *fiber.App
	Options server.Options
}

const MAX_PRINT_BODY_LEN = 2048

var reg = regexp.MustCompile(`.*\.(js|css|png|jpg|jpeg|gif|svg|webp|bmp|html|htm).*$`)

func NewServer(options server.Options, routerProvider func(engine *fiber.App), middleware ...any) server.Server {
	srv := &Server{}
	srv.Options = options
	cfg := fiber.Config{
		BodyLimit:    100 * 1024 * 1024,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*fiber.Error); ok {
				if e.Code == fiber.StatusInternalServerError {
					c.JSON(response.Fail(err.Error()))
				} else {
					c.SendStatus(e.Code)
				}
			} else {
				logger.Error("[Fiber] error:", err, "\n", string(debug.Stack()))
				c.JSON(response.Fail(err.Error()))
			}
			return nil
		},
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	}
	s := fiber.New(cfg)
	hanlders := getCfg()
	s.Use(hanlders...)
	if len(middleware) > 0 {
		s.Use(middleware...)
	}
	routerProvider(s)
	s.Use("/", func(c *fiber.Ctx) error {
		c.SendStatus(fiber.StatusNotFound)
		return nil
	})
	srv.Srv = s
	return srv
}
func getCfg() []any {
	var handlers []any
	recoverCfg := recover.Config{
		Next: func(c *fiber.Ctx) bool {
			return false
		},
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			logger.Error("[Fiber] recover:", e, "\n", string(debug.Stack()))
			c.JSON(response.Fail(fmt.Sprintf("%v", e)))
		},
	}
	handlers = append(handlers, recover.New(recoverCfg), loggingMiddleware)
	return handlers
}

func loggingMiddleware(ctx *fiber.Ctx) error {
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

func getLogBodyStr(bb []byte) string {
	if len(bb) > MAX_PRINT_BODY_LEN {
		bb = bb[:(MAX_PRINT_BODY_LEN - 1)]
	}
	return string(bb)
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

func (server *Server) Shutdown() {
}

func (server *Server) Run(graceful func()) {
	// srv := &http.Server{
	// 	Addr:         server.Options.Address,
	// 	Handler:      server.Srv,
	// 	ReadTimeout:  60 * time.Second,
	// 	WriteTimeout: 60 * time.Second,
	// }
	go func() {
		logger.Info("[Fiber] server starting on:", server.Options.Address)
		if err := server.Srv.Listen(server.Options.Address); err != nil && err != http.ErrServerClosed {
			logger.Panic("[Fiber] listen:", err.Error())
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("[Fiber]Shutting down server...")
	if err := server.Srv.Shutdown(); err != nil {
		grace(server, graceful)
		logger.Panic("[Fiber]Server forced to shutdown:", err)
	}
	grace(server, graceful)
	logger.Info("[Fiber]server has been shutdown")
}

func grace(server *Server, g func()) {
	server.Shutdown()
	g()
}

// ShouldBindBody binds the request body to a struct.
// It supports decoding the following content types based on the Content-Type header:
// application/json, application/xml, application/x-www-form-urlencoded, multipart/form-data
// If none of the content types above are matched, it will return a ErrUnprocessableEntity error
func ShouldBindBody(ctx *fiber.Ctx, obj any) error {
	err := ctx.BodyParser(obj)
	if err != nil {
		return err
	}
	return validator.ValidateStruct(obj)
}

func ShouldBindParams(ctx *fiber.Ctx, obj any) error {
	err := ctx.ParamsParser(obj)
	if err != nil {
		return err
	}
	return validator.ValidateStruct(obj)
}

func ShouldBindQuery(ctx *fiber.Ctx, obj any) error {
	err := ctx.QueryParser(obj)
	if err != nil {
		return err
	}
	return validator.ValidateStruct(obj)
}
func ShouldBindHeader(ctx *fiber.Ctx, obj any) error {
	err := ctx.ReqHeaderParser(obj)
	if err != nil {
		return err
	}
	return validator.ValidateStruct(obj)
}
