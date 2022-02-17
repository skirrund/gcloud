package nacos_registry

import (
	"github.com/skirrund/gcloud/registry"
	"testing"
)

func TestRegistry(t *testing.T) {
	ops := registry.Options{
		ServerAddrs: []string{"nacos1:8848"},
		ClientOptions: registry.ClientOptions{
			AppName: "test",
		},
		RegistryOptions: registry.RegistryOptions{
			ServiceName: "test",
			ServicePort: 8899,
			Version:     "0.1",
		},
	}
	reg := NewRegistry(ops)
	reg.RegisterInstance()
	for {

	}
}
