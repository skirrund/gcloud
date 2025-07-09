package tracer

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type CtxTraceId struct {
}

const (
	TraceIDKey = "x-trace-id"
)

func GetCtxKey() CtxTraceId {
	return CtxTraceId{}
}

func NewTraceIDContext() context.Context {
	return context.WithValue(context.Background(), GetCtxKey(), GenerateId())
}
func NewContextFromTraceId(traceId string) context.Context {
	return context.WithValue(context.Background(), GetCtxKey(), traceId)
}

func WithContext(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, GetCtxKey(), traceId)
}

func WithTraceID(ctx context.Context) context.Context {
	return context.WithValue(ctx, GetCtxKey(), GenerateId())
}

func GetTraceID(ctx context.Context) (traceId any) {
	return ctx.Value(GetCtxKey())
}

func GenerateId() string {
	uidG, _ := uuid.NewV7()
	return strings.ReplaceAll(uidG.String(), "-", "")
}
