package mq

import (
	"time"

	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/mq/consumer"
	"github.com/skirrund/gcloud/mq/pulsar"

	"github.com/apache/pulsar-client-go/pulsar"
)

type Client struct {
	MqClient interface{}
	AppName  string
}

type ConsumerOptions struct {
	Topic            string
	SubscriptionName string
	SubscriptionType subscriptionType
	MessageListener  consumer.Consumer
	ACKMode          ACKMode
	RetryTimes       uint32
}

type ACKMode uint32

const (
	ACK_AUTO ACKMode = iota
	ACK_MANUAL
)

type subscriptionType int

const (

	// Exclusive there can be only 1 consumer on the same topic with the same subscription name
	Exclusive subscriptionType = iota

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

const (
	SERVER_URL_KEY         = "mq.service-url"
	CONNECTION_TIMEOUT_KEY = "mq.connectionTimeout"
	OPERATION_TIMEOUT_KEY  = "mq.operationTimeout"
)

var MqClient *Client

func InitClient(url string, connectionTimeoutSecond int64, operationTimeoutSecond int64, appName string) *Client {
	logger.Infof("[MQ] start init mq-client:%s,connTimeout:%d,operTimeout:%d,appNmae:%s", url, connectionTimeoutSecond, operationTimeoutSecond, appName)
	MqClient = &Client{
		MqClient: mth_pulsar.NewClient(url, connectionTimeoutSecond, operationTimeoutSecond, appName),
		AppName:  appName,
	}
	return MqClient
}

func DefaultSubscriptionType() subscriptionType {
	return Shared
}

func (c *Client) Send(topic string, msg string) error {
	return c.doSendDelay(topic, msg, 0, false)
}

func (c *Client) doSendDelay(topic string, msg string, delay time.Duration, async bool) error {
	logger.Info("[MQ] send message:", topic, ",", string(msg), ",", delay, ",", async)
	if mqc, ok := c.MqClient.(*mth_pulsar.PulsarClient); ok {
		if async {
			return mqc.SendDelayAsync(topic, msg, delay)
		} else {
			return mqc.SendDelay(topic, msg, delay)
		}
	} else {
		logger.Error("[MQ] no available MQClient")
	}
	return nil
}

func (c *Client) SendDelay(topic string, msg string, delay time.Duration) error {
	return c.doSendDelay(topic, msg, delay, false)
}
func (c *Client) SendDelayAt(topic string, msg string, delayAt time.Time) error {
	if mqc, ok := c.MqClient.(*mth_pulsar.PulsarClient); ok {
		return mqc.SendDelayAt(topic, msg, delayAt)
	}
	return nil
}
func (c *Client) SendAsync(topic string, msg string) error {
	return c.doSendDelay(topic, msg, 0, true)
}

func (c *Client) SendDelayAsync(topic string, msg string, delay time.Duration) error {
	return c.doSendDelay(topic, msg, delay, true)
}
func (c *Client) SendDelayAtAsync(topic string, msg string, delayAt time.Time) error {
	if mqc, ok := c.MqClient.(*mth_pulsar.PulsarClient); ok {
		return mqc.SendDelayAtAsync(topic, msg, delayAt)
	}
	return nil
}

func (c *Client) Subscribes(options ...ConsumerOptions) {
	for _, op := range options {
		c.Subscribe(op)
	}
}

func (c *Client) Subscribe(options ConsumerOptions) {
	if mqc, ok := c.MqClient.(*mth_pulsar.PulsarClient); ok {
		opts := mth_pulsar.ConsumerOptions{
			Topic:            options.Topic,
			SubscriptionName: options.SubscriptionName,
			SubscriptionType: pulsar.SubscriptionType(options.SubscriptionType),
			ACKMode:          uint32(options.ACKMode),
			RetryTimes:       uint32(options.RetryTimes),
			MessageListener: func(msg consumer.Message) error {
				err := options.MessageListener.OnMessage(msg)
				return err
			},
		}
		mqc.Subscribe(opts)
	}

}

func (c *Client) Close() {
	if mqc, ok := c.MqClient.(*mth_pulsar.PulsarClient); ok {
		mqc.Close()
	} else {
		logger.Error("[MQ]no available MQClient")
	}
}
