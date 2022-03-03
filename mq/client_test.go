package mq

import (
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/mq/consumer"
	"strconv"
	"testing"
	"time"
)

type Test struct {
}

func (Test) OnMessage(message consumer.Message) error {
	logger.Info("onmsg:", message.Value)
	time.Sleep(1 * time.Second)
	return nil
}
func TestInitClient(t *testing.T) {
	client := InitClient("pulsar://pulsar1:6650,pulsar2:6650,pulsar3:6650", 0, 0, "test")
	for i := 0; i != 10000; i++ {
		go client.Send("test1", "test1-"+strconv.FormatInt(int64(i), 10))
	}
	client.Subscribe(ConsumerOptions{
		Topic:            "test1",
		SubscriptionName: "test1",
		SubscriptionType: Shared,
		MessageListener:  Test{},
	})

	select {}
}
