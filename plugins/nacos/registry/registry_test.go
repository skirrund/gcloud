package nacos_registry

import (
	"testing"

	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/registry"
)

func TestRegistry(t *testing.T) {
	N203 := "mse-32343d30-p.nacos-ans.mse.aliyuncs.com:8848"
	N201 := "mse-6be4cfe0-p.nacos-ans.mse.aliyuncs.com:8848"
	logger.Info(N201, N203)
	ops := registry.Options{
		ServerAddrs: []string{N203},
		ClientOptions: registry.ClientOptions{
			AppName: "test",
			LogDir:  "/Users/jerry.shi/logs/nacos/go",
		},
		RegistryOptions: registry.RegistryOptions{
			ServiceName: "test",
			ServicePort: 8899,
			Version:     "0.1",
		},
	}
	reg := NewRegistry(ops)
	reg.RegisterInstance()
}
