package mq

import (
	"time"

	"github.com/skirrund/gcloud/mq/consumer"
)

type IClient interface {
	Send(topic string, msg string) error
	SendDelay(topic string, msg string, delay time.Duration) error
	SendDelayAt(topic string, msg string, delayAt time.Time) error
	SendAsync(topic string, msg string) error
	SendDelayAsync(topic string, msg string, delay time.Duration) error
	Subscribes(options ...ConsumerOptions)
	Subscribe(options ConsumerOptions)
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

type ConsumerOptions struct {
	Topic                 string
	SubscriptionName      string
	SubscriptionType      SubscriptionType
	MessageListener       consumer.Consumer
	ACKMode               ACKMode
	RetryTimes            uint32
	MaxMessageChannelSize uint32
}
