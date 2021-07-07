package mth_pulsar

import (
	"runtime"
	"sync"
	"time"

	"github.com/skirrund/gcloud/logger"

	"github.com/apache/pulsar-client-go/pulsar"
)

type PulsarClient struct {
	client  pulsar.Client
	appName string
	mt      sync.Mutex
}

const (
	defaultConnectionTimeout = 5 * time.Second
	defaultOperationTimeout  = 30 * time.Second
)

var once sync.Once

func NewClient(url string, connectionTimeoutSecond int64, operationTimeoutSecond int64, appName string) *PulsarClient {
	if pulsarClient != nil {
		return pulsarClient
	}
	once.Do(func() {
		pulsarClient = &PulsarClient{
			client:  createClient(url, connectionTimeoutSecond, operationTimeoutSecond),
			appName: appName,
		}
	})
	return pulsarClient

}

func createClient(url string, connectionTimeoutSecond int64, operationTimeoutSecond int64) pulsar.Client {
	var cts time.Duration
	var ots time.Duration
	if connectionTimeoutSecond > 0 {
		cts = time.Duration(connectionTimeoutSecond) * time.Second
	} else {
		cts = defaultConnectionTimeout
	}
	if operationTimeoutSecond > 0 {
		ots = time.Duration(operationTimeoutSecond) * time.Second
	} else {
		ots = defaultOperationTimeout
	}
	logger.Infof("[pulsar]start init pulsar-client:" + url)
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     url,
		ConnectionTimeout:       cts,
		OperationTimeout:        ots,
		MaxConnectionsPerBroker: runtime.NumCPU(),
	})
	if err != nil {
		panic(err)
	}
	logger.Infof("[pulsar]finished init pulsar-client")
	return client
}

func (pc *PulsarClient) Close() {
	pc.client.Close()
}
