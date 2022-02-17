package nacos_config

import (
	"os"
	"testing"

	commonCfg "github.com/skirrund/gcloud/config"
)

func TestConfig(t *testing.T) {
	opts := commonCfg.Options{
		ServerAddrs: []string{"nacos1:8848"},
		ClientOptions: commonCfg.ClientOptions{
			NamespaceId: "PBM-Service",
			LogDir:      ".",
			TimeoutMs:   5000,
			AppName:     "test-nacos",
		},
		ConfigOptions: commonCfg.ConfigOptions{
			Prefix:        "pbm-common-service",
			FileExtension: "yaml",
			Env:           "test",
			Group:         "DEFAULT_GROUP",
		},
	}
	t.Log(">>>>>")
	nacos := CreateInstance(opts)
	t.Log(nacos.GetString("datasource.dsn"))
	t.Log("end")
	var bytes []byte
	os.Stdin.Read(bytes)
}
