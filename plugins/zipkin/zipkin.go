package zipkin

import (
	"errors"
	"sync"

	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/utils"

	"github.com/skirrund/gcloud/bootstrap/env"

	"github.com/opentracing/opentracing-go"
	zkOt "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

var zkTracer opentracing.Tracer
var zkReporter reporter.Reporter
var once sync.Once

func GetTracer() opentracing.Tracer {
	once.Do(func() {
		initZipkinTracer()
	})
	return zkTracer
}

func initZipkinTracer() error {
	cfg := env.GetInstance()
	url := cfg.GetString(env.ZIPKIN_URL_KEY)
	if len(url) == 0 {
		logger.Fatal("[zipkin] unable to create local reporter:url is empty")
		return errors.New("url is empty")
	}
	zkReporter = zipkinhttp.NewReporter(url)
	serviceName := cfg.GetString(env.SERVER_SERVERNAME_KEY)
	port := cfg.GetString(env.SERVER_PORT_KEY)
	addr := utils.LocalIP() + ":" + port
	endpoint, err := zipkin.NewEndpoint(serviceName, addr)
	if err != nil {
		logger.Fatal("[zipkin] unable to create local endpoint:", err, ",serviceName:", serviceName, ",", addr)
		return err
	}
	nativeTracer, err := zipkin.NewTracer(zkReporter, zipkin.WithTraceID128Bit(true), zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		logger.Fatal("[zipkin] unable to create tracer: ", err)
		return err
	}
	zkTracer = zkOt.Wrap(nativeTracer)
	opentracing.SetGlobalTracer(zkTracer)
	return nil
}

func Close() {
	if zkReporter != nil {
		zkReporter.Close()
	}
}
