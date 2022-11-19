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
	ShutdownEvent        EventName = "shutdown"
	CertRenewEvent                 = "certrenew"
	InstanceStartupEvent           = "instancestartup"
	InstanceRestartEvent           = "instancerestart"
	RegistryChangeEvent            = "registryChange"
	//ConfigChangeEvent should use the all config data as info
	ConfigChangeEvent = "configChange"
)

type EventHook func(eventType EventName, eventInfo interface{}) error

var eventHooks = struct {
	sync.Mutex
	EventHook map[EventName][]EventHook
}{EventHook: make(map[EventName][]EventHook)}

func RegisterEventHook(name string, hook ...EventHook) error {
	if name == "" {
		logger.Error("[server] event hook must have a name")
		return errors.New("[server] event hook must have a name")
	}
	logger.Info("[server] RegisterEventHook:"+name, hook)
	en := EventName(name)
	v, ok := eventHooks.EventHook[en]
	var list []EventHook
	if ok {
		v = append(v, hook...)
	} else {
		v = list
	}
	eventHooks.Lock()
	eventHooks.EventHook[en] = v
	defer eventHooks.Unlock()
	return nil
}

// EmitEvent executes the different hooks passing the EventType as an
// argument. This is a blocking function. Hook developers should
// use 'go' keyword if they don't want to block Caddy.
func EmitEvent(event EventName, info interface{}) {
	funcs, ok := eventHooks.EventHook[event]
	if ok {
		logger.Info("[server] EmitEvent exec ", event)
		for i := range funcs {
			f := funcs[i]
			err := f(event, info)
			if err != nil {
				logger.Infof("[server] error on '%s' hook: %v", event, err)
			}
		}
	}
}
