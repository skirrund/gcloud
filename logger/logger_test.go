package logger

import (
	"context"
	"testing"
	"time"

	"github.com/skirrund/gcloud/tracer"
)

func TestLogger(t *testing.T) {
	initLog("logger", "test", "111", true, false, 1*time.Hour)
	ctx := context.Background()
	InfofContext(tracer.NewTraceIDContext(), "info.......%s%s%s", "1", "-", "2")
	Info("warn1.....")
	WarnContext(tracer.WithTraceID(ctx), "info.......", "1", "-", "2")
	Info("warn3.....")
	WarnfContext(tracer.WithTraceID(ctx), "info.......%s", "2")
	Warnf("info.......%s", "2")
}
