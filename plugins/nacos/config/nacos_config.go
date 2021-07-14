package nacos_config

import (
	"fmt"
	"io"
	"strings"

	commonConfig "github.com/skirrund/gcloud/config"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/plugins/nacos"

	"github.com/skirrund/gcloud/server"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"
)

const (
	NACOS_CONFIG_PREFIX_KEY         = "nacos.config.prefix"
	NACOS_CONFIG_FILE_EXTENSION_KEY = "nacos.config.file-extension"
	NACOS_CONFIG_SERVER_ADDR_KEY    = "nacos.config.server-addr"
	NACOS_CONFIG_GROUP_KEY          = "nacos.config.group"
	NACOS_CONFIG_NAMESPACE_KEY      = "nacos.config.namespace"
)

type nacosConfigCenter struct {
	opts   commonConfig.Options
	client config_client.IConfigClient
}

var config *viper.Viper
var nc *nacosConfigCenter

func configure(n *nacosConfigCenter, opts commonConfig.Options) error {
	client, err := nacos.CreateConfigClient(opts)
	if err != nil {
		return err
	}
	n.client = client
	return nil
}

func CreateInstance(opts commonConfig.Options) *nacosConfigCenter {
	nc = &nacosConfigCenter{}
	nc.opts = opts
	configure(nc, opts)
	config = viper.New()
	err := nc.Read()
	if err != nil {
		logger.Panic(err)
		return nc
		//panic(err)
	}
	logger.Info("[nacos] CreateInstance EmitEvent ConfigChangeEvent")
	server.EmitEvent(server.ConfigChangeEvent, config)
	nc.Watch()
	return nc
}

func (nc *nacosConfigCenter) Set(key string, value interface{}) {
	config.Set(key, value)
}

func (nc *nacosConfigCenter) Get(key string) interface{} {
	return config.Get(key)
}

func (nc *nacosConfigCenter) GetStringWithDefault(key string, defaultString string) string {
	v := nc.GetString(key)
	if len(v) == 0 {
		return defaultString
	}
	return v
}

func (nc *nacosConfigCenter) GetInt(key string) int {
	return config.GetInt(key)
}

func (nc *nacosConfigCenter) GetIntWithDefault(key string, defaultInt int) int {
	v := nc.GetInt(key)
	if v == 0 {
		return defaultInt
	}
	return v
}

func (nc *nacosConfigCenter) GetInt64WithDefault(key string, defaultInt64 int64) int64 {
	v := nc.GetInt64(key)
	if v == 0 {
		return defaultInt64
	}
	return v
}
func (nc *nacosConfigCenter) GetInt64(key string) int64 {
	return config.GetInt64(key)
}
func (nc *nacosConfigCenter) GetString(key string) string {
	return config.GetString(key)
}
func (nc *nacosConfigCenter) GetStringSlice(key string) []string {
	return config.GetStringSlice(key)
}

func (nc *nacosConfigCenter) GetUint64(key string) uint64 {
	return config.GetUint64(key)
}
func (nc *nacosConfigCenter) GetUint64WithDefault(key string, defaultUint64 uint64) uint64 {
	v := nc.GetUint64(key)
	if v == 0 {
		return defaultUint64
	}
	return v
}
func (nc *nacosConfigCenter) GetUint(key string) uint {
	return config.GetUint(key)
}
func (nc *nacosConfigCenter) GetUintWithDefault(key string, defaultUint uint) uint {
	v := nc.GetUint(key)
	if v == 0 {
		return defaultUint
	}
	return v
}
func (nc *nacosConfigCenter) GetBool(key string) bool {
	return config.GetBool(key)
}
func (nc *nacosConfigCenter) GetFloat64(key string) float64 {
	return config.GetFloat64(key)
}

func (nc *nacosConfigCenter) GetStringMapString(key string) map[string]string {
	return config.GetStringMapString(key)
}

func (nc *nacosConfigCenter) MergeConfig(eventType server.EventName, eventInfo interface{}) error {
	return nil
}

func (nc *nacosConfigCenter) SetBaseConfig(reader io.Reader, configType string) {
	baseCfg := viper.New()
	baseCfg.SetConfigName("base")
	baseCfg.SetConfigType(configType)
	baseCfg.ReadConfig(reader)
	config.MergeConfigMap(baseCfg.AllSettings())
}

func (c *nacosConfigCenter) Read() error {
	dataId := c.dataId()
	content, err := c.client.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  c.opts.ConfigOptions.Group,
	})
	if err != nil {
		return fmt.Errorf("[nacos] error reading data from nacos: %s【%v】", dataId, err)
	}
	if len(content) == 0 {
		return fmt.Errorf("[nacos] error reading data from nacos: %s【文件内容为空】", dataId)
	}
	reader := strings.NewReader(content)
	config.SetConfigType(c.opts.ConfigOptions.FileExtension)
	config.ReadConfig(reader)
	return nil
}

func (c *nacosConfigCenter) String() string {
	return "nacos"
}

func (c *nacosConfigCenter) Watch() {
	newConfigWatcher(c)
}

func (nc *nacosConfigCenter) dataId() string {
	opts := nc.opts.ConfigOptions
	return opts.Prefix + "-" + opts.Env + "." + opts.FileExtension
}

func newConfigWatcher(nc *nacosConfigCenter) {
	logger.Info("[nacos] ListenConfig DataId:" + nc.dataId() + ",Group:" + nc.opts.ConfigOptions.Group)

	err := nc.client.ListenConfig(vo.ConfigParam{
		DataId: nc.dataId(),
		Group:  nc.opts.ConfigOptions.Group,
		OnChange: func(namespace, group, dataId, data string) {
			logger.Info("[nacos] config changed dataId:" + dataId + ",ns:" + namespace + ",group:" + group)
			reader := strings.NewReader(data)
			config.SetConfigType(nc.opts.ConfigOptions.FileExtension)
			err := config.ReadConfig(reader)
			if err != nil {
				logger.Error("[nacos] watch error:", err.Error())
				return
			}
			logger.Info("[nacos] watch EmitEvent ConfigChangeEvent")
			server.EmitEvent(server.ConfigChangeEvent, config)
		},
	})
	if err != nil {
		logger.Error("[nacos] newConfigWatcher error:" + err.Error())
	}
}

func (nc *nacosConfigCenter) Shutdown() {
	p := vo.ConfigParam{
		DataId: nc.dataId(),
		Group:  nc.opts.ConfigOptions.Group,
	}
	logger.Info("[nacos] CancelListenConfig:[dataId=" + p.DataId + "],[group=" + p.Group + "]")

	err := nc.client.CancelListenConfig(p)
	if err != nil {
		logger.Error("[nacos] CancelListenConfig error: ", err.Error())
	}
}
