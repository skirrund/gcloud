package registry

import (
	"strconv"
)

var RegistryCenter IRegistry

type IRegistry interface {
	RegisterInstance() error
	Shutdown()
	Subscribe(serviceName string) error
	GetInstance(serviceName string) *Instance
	SelectInstances(serviceName string) ([]*Instance, error)
}

type Instance struct {
	Ip       string
	Port     uint64
	Metadata map[string]string
}

const (
	PROTOCOL_HTTP  = "http://"
	PROTOCOL_HTTPS = "https://"
)

func (i *Instance) GetHost() string {
	return i.Ip + ":" + strconv.FormatUint(i.Port, 10)
}

func (i *Instance) GetUrl() string {
	return PROTOCOL_HTTP + i.GetHost()
}
