package mq

import (
	"context"
	"time"
)

type IClient interface {
	//sync
	Send(msg *Message) error
	SendAsync(msg *Message) error
	Subscribes(options ...ConsumerOptions)
	//async
	Subscribe(options ConsumerOptions)
	SubscribeSync(options ConsumerOptions)
	Close()
}

type ACKMode uint32

const (
	ACK_AUTO ACKMode = iota
	ACK_MANUAL
)

type SubscriptionType int

const (

	// Exclusive there can be only 1 consumer on the same topic with the same subscription name
	Exclusive SubscriptionType = iota

	// Shared subscription mode, multiple consumer will be able to use the same subscription name
	// and the messages will be dispatched according to
	// a round-robin rotation between the connected consumers
	Shared

	// Failover subscription mode, multiple consumer will be able to use the same subscription name
	// but only 1 consumer will receive the messages.
	// If that consumer disconnects, one of the other connected consumers will start receiving messages.
	Failover

	// KeyShared subscription mode, multiple consumer will be able to use the same
	// subscription and all messages with the same key will be dispatched to only one consumer
	KeyShared
)

type NatsOpts struct {
	Stream        string
	PullBatchSize int
}

type SubOpts struct {
	Name string
}

type ConsumerOptions struct {
	Topic                 string
	SubscriptionName      string
	SubscriptionType      SubscriptionType
	MessageListener       func(ctx context.Context, message *Message) error
	ACKMode               ACKMode
	RetryTimes            uint64
	MaxMessageChannelSize uint64
	NatsOpts              NatsOpts
}

type Message struct {
	Topic           string
	Header          map[string]string
	Payload         []byte
	RedeliveryCount uint64
	DeliverAfter    time.Duration
	DeliverAt       time.Time
	NatsOpts        NatsOpts
	SubOpts         SubOpts
}
