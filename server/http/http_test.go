package http

import (
	"fmt"
	"testing"

	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/server/decoder"
	"github.com/skirrund/gcloud/tracer"
)

func TestXxx(t *testing.T) {
	var r []byte
	logCtx := tracer.NewTraceIDContext()
	logger.InfoContext(logCtx, "123123123123")
	client := DefaultH2CClient.WithTracerContext(logCtx)
	_, err := client.GetUrl("http://127.0.0.1:8080/test", nil, nil, &r)
	fmt.Println(string(r), err)
}

func TestDecoder(t *testing.T) {
	de := decoder.StringDecoder{}
	resp := []byte("哈喽a")
	var b []byte
	de.DecoderObj(resp, &b)
	fmt.Println(string(b))
}
