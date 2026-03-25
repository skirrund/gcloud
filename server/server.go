package server

import (
	"errors"
	"sync"
	"time"

	"github.com/skirrund/gcloud/logger"
)

type Server interface {
	Run(graceful ...func())
	Shutdown()
	GetServeServer() any
}

type Options struct {
	ServerName           string
	Address              string
	Concurrency          int
	IdleTimeout          time.Duration
	MaxRequestBodySize   int
	H2C                  bool
	H2                   bool
	MaxConcurrentStreams uint32
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
	CertRenewEvent       EventName = "certrenew"
	InstanceStartupEvent EventName = "instancestartup"
	InstanceRestartEvent EventName = "instancerestart"
	RegistryChangeEvent  EventName = "registryChange"
	//ConfigChangeEvent should use the all config data as info
	ConfigChangeEvent EventName = "configChange"
)

type EventHook func(eventType EventName, eventInfo any) error

var eventHooks = struct {
	sync.Mutex
	EventHook map[EventName][]EventHook
}{EventHook: make(map[EventName][]EventHook)}

func RegisterEventHookFirst(name EventName, hook ...EventHook) error {
	return doRegisterEventHook(name, true, hook...)
}

func doRegisterEventHook(name EventName, isFirst bool, hook ...EventHook) error {
	if name == "" {
		logger.Error("[server] event hook must have a name")
		return errors.New("[server] event hook must have a name")
	}
	logger.Info("[server] RegisterEventHook:"+name, hook)
	v, ok := eventHooks.EventHook[name]
	var list []EventHook
	list = append(list, hook...)
	if ok {
		if isFirst {
			list = append(list, v...)
		} else {
			list = append(v, list...)
		}
	}
	v = list
	eventHooks.Lock()
	defer eventHooks.Unlock()
	eventHooks.EventHook[name] = v
	return nil
}

func RegisterEventHook(name EventName, hook ...EventHook) error {
	return doRegisterEventHook(name, false, hook...)
}

// EmitEvent executes the different hooks passing the EventType as an
// argument. This is a blocking function. Hook developers should
// use 'go' keyword if they don't want to block Caddy.
func EmitEvent(event EventName, info any) {
	funcs, ok := eventHooks.EventHook[event]
	if ok {
		logger.Info("[server] EmitEvent exec ", event)
		for i := range funcs {
			f := funcs[i]
			go func() {
				err := f(event, info)
				if err != nil {
					logger.Infof("[server] error on '%s' hook: %v", event, err)
				}
			}()
		}
	}
}
