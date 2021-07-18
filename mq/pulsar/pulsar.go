package mth_pulsar

import (
	"errors"
	"sync"
	"time"

	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/utils"

	baseConsumer "github.com/skirrund/gcloud/mq/consumer"

	"github.com/apache/pulsar-client-go/pulsar"
)

type Message struct {
	Msg pulsar.Message
}

type ConsumerOptions struct {
	Topic            string
	SubscriptionName string
	SubscriptionType pulsar.SubscriptionType
	MessageListener  func(baseConsumer.Message) error
	ACKMode          uint32
	RetryTimes       uint32
}

var consumers sync.Map

const (
	MAX_RETRY_TIMES = 50
)

var pulsarClient *PulsarClient

func InitClient(url string, connectionTimeoutSecond int64, operationTimeoutSecond int64, name string) *PulsarClient {
	return NewClient(url, connectionTimeoutSecond, operationTimeoutSecond, name)
}

func getAppName(appName string) string {
	name := utils.Uuid()
	if len(appName) == 0 {
		return name
	} else {
		return appName + "-" + name
	}
}

func (pc *PulsarClient) Send(topic string, msg string) error {
	return pc.doSend(topic, msg, 0)
}

func (pc *PulsarClient) SendDelay(topic string, msg string, delay time.Duration) error {
	return pc.doSend(topic, msg, delay)
}
func (pc *PulsarClient) SendDelayAt(topic string, msg string, delayAt time.Time) error {
	return pc.doSendDelayAt(topic, msg, delayAt)
}
func (pc *PulsarClient) SendAsync(topic string, msg string) error {
	return pc.doSendAsync(topic, msg, 0)
}

func (pc *PulsarClient) SendDelayAsync(topic string, msg string, delay time.Duration) error {
	return pc.doSendAsync(topic, msg, delay)
}
func (pc *PulsarClient) SendDelayAtAsync(topic string, msg string, delayAt time.Time) error {
	return pc.doSendDelayAtAsync(topic, msg, delayAt)
}

func (pc *PulsarClient) Subscribe(opts ConsumerOptions) {
	go pc.doSubscribe(opts)
}

func (pc *PulsarClient) doSubscribe(opts ConsumerOptions) error {
	subscriptionName := opts.SubscriptionName
	topic := opts.Topic
	logger.Infof("[pulsar]ConsumerOptions:%v", opts)
	options := pulsar.ConsumerOptions{
		Topic:               topic,
		SubscriptionName:    subscriptionName,
		Type:                opts.SubscriptionType,
		Name:                getAppName(pc.appName),
		NackRedeliveryDelay: 15 * time.Second,
		//Schema:              pulsar.NewStringSchema(nil),
	}
	if opts.RetryTimes == 0 {
		opts.RetryTimes = MAX_RETRY_TIMES
	}
	schema := pulsar.NewJSONSchema(`"string"`, nil)

	channel := make(chan pulsar.ConsumerMessage, 100)
	options.MessageChannel = channel
	consumer, err := pc.client.Subscribe(options)
	if err != nil {
		logger.Error(errors.New("[pulsar] Subscribe error:" + err.Error()))
		panic("[pulsar] Subscribe error:" + err.Error())
		//return err
	}
	consumers.Store(topic+":"+subscriptionName, opts)

	logger.Infof("[pulsar]store consumerOptions:"+topic+":"+subscriptionName, ",", opts)

	for cm := range channel {
		msg := cm.Message
		//		var s = ""
		msgStr := ""
		logger.Infof("[pulsar] consumer info=>subName:%s,msgId:%v,reDeliveryCount:%d,publishTime:%v,producerName:%s", cm.Subscription(), msg.ID(), msg.RedeliveryCount(), msg.PublishTime(), msg.ProducerName())
		err := schema.Decode(msg.Payload(), &msgStr)
		if err != nil {
			logger.Info("[pulsar] consumer msg:", err.Error())
		} else {
			logger.Info("[pulsar] consumer msg:", msgStr)
		}
		err = opts.MessageListener(baseConsumer.Message{
			Value:           msgStr,
			RedeliveryCount: msg.RedeliveryCount(),
		})
		if err == nil {
			consumer.Ack(msg)
		} else {
			logger.Error("[pulsar] consumer error:" + err.Error())
			copts, ok := consumers.Load(topic + ":" + cm.Subscription())
			retryTimes := uint32(0)
			ACKMode := uint32(0)
			if ok {
				if opt, o := copts.(ConsumerOptions); o {
					retryTimes = opt.RetryTimes
					if retryTimes > MAX_RETRY_TIMES {
						retryTimes = MAX_RETRY_TIMES
					}
					ACKMode = opt.ACKMode
				} else {
					logger.Errorf("[pulsar] consumerOptions type error=> options:%v", copts)
				}
			} else {
				logger.Error("[pulsar] can not find ConsumerOptions=>subName:" + topic + ":" + cm.Subscription())
			}

			rt := msg.RedeliveryCount()
			if ACKMode == 1 && rt < retryTimes {
				logger.Infof("[pulsar]consummer error and retry=> subscriptionName:"+cm.Subscription()+",initRetryTimes:%d,retryTimes:%d,ack:%d", retryTimes, rt, ACKMode)
				consumer.Nack(msg)
			} else {
				logger.Infof("[pulsar]consummer error and can not retry=> subscriptionName:"+cm.Subscription()+",initRetryTimes:%d,retryTimes:%d,ack:%d", retryTimes, rt, ACKMode)
				consumer.Ack(msg)
			}

		}
	}
	return nil
}
