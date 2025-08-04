package tracer

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type CtxTraceId struct {
}

const (
	TraceIDKey = "gcloud-x-trace-id"
)

func GetCtxKey() CtxTraceId {
	return CtxTraceId{}
}

func NewTraceIDContext() context.Context {
	return context.WithValue(context.Background(), GetCtxKey(), GenerateId())
}
func NewContextFromTraceId(traceId string) context.Context {
	return WithContext(context.Background(), traceId)
}

func WithContext(ctx context.Context, traceId string) context.Context {
	if len(traceId) == 0 {
		traceId = GenerateId()
	}
	if ctx == nil {
		ctx = context.Background()
		return context.WithValue(ctx, GetCtxKey(), traceId)
	} else {
		if tid := GetTraceID(ctx); tid == nil {
			return context.WithValue(ctx, GetCtxKey(), traceId)
		}
	}
	return context.WithValue(ctx, GetCtxKey(), traceId)
}

func WithTraceID(ctx context.Context) context.Context {
	return WithContext(ctx, "")
}

func GetTraceID(ctx context.Context) (traceId any) {
	return ctx.Value(GetCtxKey())
}

func GenerateId() string {
	uidG, _ := uuid.NewV7()
	return strings.ReplaceAll(uidG.String(), "-", "")
}
