package mq

import "context"

type Consumer interface {
	OnMessage(ctx context.Context, message Message) error
}
