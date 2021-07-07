package server

import (
	"errors"
	"sync"

	"github.com/skirrund/gcloud/logger"
)

type Server interface {
	Run(graceful func())
	Shutdown()
}

type Options struct {
	ServerName string
	Address    string
	//Container  Server
	//	Registry   registry.IRegistry
	//	Config     config.IConfig
	//Mq         *mq.Client
	//Redis      *redis.RedisClient
	//IdWorker   *common.Worker
}

type EventName string

const (
	StartupEvent         EventName = "startup"
	ShutdownEvent                  = "shutdown"
	CertRenewEvent                 = "certrenew"
	InstanceStartupEvent           = "instancestartup"
	InstanceRestartEvent           = "instancerestart"
	RegistryChangeEvent            = "registryChange"
	//ConfigChangeEvent should use the all config data as info
	ConfigChangeEvent = "configChange"
)

type EventHook func(eventType EventName, eventInfo interface{}) error

var eventHooks = &sync.Map{}

func RegisterEventHook(name string, hook EventHook) error {
	if name == "" {
		logger.Error("[server] event hook must have a name")
		return errors.New("[server] event hook must have a name")
	}
	logger.Info("[server] RegisterEventHook:"+name, hook)
	_, dup := eventHooks.LoadOrStore(name, hook)
	if dup {
		logger.Error("[server] hook named " + name + " already registered")
	}
	return nil
}

// EmitEvent executes the different hooks passing the EventType as an
// argument. This is a blocking function. Hook developers should
// use 'go' keyword if they don't want to block Caddy.
func EmitEvent(event EventName, info interface{}) {
	eventHooks.Range(func(k, v interface{}) bool {
		logger.Info("[server] EmitEvent exec ", k.(string))
		if e, _ := k.(string); e == string(event) {
			err := v.(EventHook)(event, info)
			if err != nil {
				logger.Infof("[server] error on '%s' hook: %v", k.(string), err)
			}
		}

		return true
	})
}
