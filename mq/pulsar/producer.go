package mth_pulsar

import (
	"context"
	"errors"

	//	"sync"
	"time"

	"github.com/skirrund/gcloud/logger"

	"github.com/apache/pulsar-client-go/pulsar"
)

var producers = make(map[string]pulsar.Producer)

func (pc *PulsarClient) getProducer(topic string) (pulsar.Producer, error) {
	p, ok := producers[topic]
	if ok && p != nil {
		logger.Info("[pulsar]load producer fromcache1:", topic, ",", true)
		return p.(pulsar.Producer), nil
	}
	pc.mt.Lock()
	defer pc.mt.Unlock()
	p, ok = producers[topic]
	var err error
	if !ok || p == nil {
		p, err = createProducer(topic)
		producers[topic] = p
	} else {
		logger.Info("[pulsar]load producer fromcache2:", topic, ",", true)
	}

	return p.(pulsar.Producer), err
}

func createProducer(topic string) (pulsar.Producer, error) {
	logger.Info("[pulsar]start create pulsar.Producer:", topic)
	pp := pulsar.ProducerOptions{
		Topic:  topic,
		Name:   getAppName(pulsarClient.appName),
		Schema: pulsar.NewJSONSchema(`"string"`, nil),
	}

	producer, err := pulsarClient.client.CreateProducer(pp)
	if err != nil {
		logger.Error("[pulsar]error create pulsar.Producer:", err)
	} else {
		logger.Info("[pulsar]finished create pulsar.Producer:", topic)
	}
	return producer, err
}

func createMsg(msg string, deliverAfter time.Duration) *pulsar.ProducerMessage {
	message := &pulsar.ProducerMessage{
		//Payload:      msg,
		Value:        msg,
		DeliverAfter: deliverAfter,
	}
	return message
}
func createMsgDeliverAt(msg string, deliverAt time.Time) *pulsar.ProducerMessage {
	message := &pulsar.ProducerMessage{
		Value:     msg,
		DeliverAt: deliverAt,
	}
	return message
}

func (pc *PulsarClient) doSend(topic string, msg string, deliverAfter time.Duration) error {
	if len(topic) == 0 {
		return errors.New("[pulsar] topic is empty")
	}
	logger.Info("[pulsar] send msg =>topic:" + topic + ":" + string(msg))
	message := createMsg(msg, deliverAfter)
	producer, err := pc.getProducer(topic)
	if err != nil {
		return err
	}
	msgId, err := producer.Send(context.Background(), message)
	if err != nil {
		logger.Error("[pulsar]发送消息失败: ", err)
		return err
	}
	if msgId == nil {
		return errors.New("[pulsar]发送消息失败[messageId为空]:" + topic)
	}
	return nil
}

func (pc *PulsarClient) doSendAsync(topic string, msg string, deliverAfter time.Duration) error {
	var err error
	if len(topic) == 0 {
		err = errors.New("[pulsar] topic is empty")
		logger.Error(err.Error())
		return err
	}
	message := createMsg(msg, deliverAfter)
	p, err := pc.getProducer(topic)
	if err != nil {
		return err
	}
	p.SendAsync(context.Background(), message, func(msgId pulsar.MessageID, msg *pulsar.ProducerMessage, err error) {
		if err != nil {
			logger.Error("[pulsar]发送doSendAsync消息失败:", err)
		} else {
			logger.Info("[pulsar] doSendAsync finish:", msgId)
		}
	})
	return nil
}

func (pc *PulsarClient) doSendDelayAt(topic string, msg string, deliverAt time.Time) error {
	if len(topic) == 0 {
		return errors.New("[pulsar] topic is empty")
	}
	message := createMsgDeliverAt(msg, deliverAt)
	p, err := pc.getProducer(topic)
	if err != nil {
		return err
	}
	msgId, err := p.Send(context.Background(), message)
	if err != nil {
		return err
	}
	if msgId == nil {
		return errors.New("[pulsar]发送消息失败[messageId为空]:" + topic)
	}
	logger.Info("[pulsar] doSendDelayAt finish: ", msg)
	return nil
}

func (pc *PulsarClient) doSendDelayAtAsync(topic string, msg string, deliverAt time.Time) error {
	if len(topic) == 0 {
		return errors.New("[pulsar] topic is not empty")
	}
	message := createMsgDeliverAt(msg, deliverAt)
	p, err := pc.getProducer(topic)
	if err != nil {
		return err
	}
	p.SendAsync(context.Background(), message, func(msgId pulsar.MessageID, msg *pulsar.ProducerMessage, err error) {
		if err != nil {
			logger.Error("[pulsar]发送doSendDelayAtAsync消息失败:", err)
		} else {
			logger.Info("[pulsar] doSendDelayAtAsync finish:", msg.Value)
		}
	})
	return nil
}
