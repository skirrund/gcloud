package env

import (
	"io"
	"sync"

	"github.com/skirrund/gcloud/config"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/server"

	"github.com/spf13/viper"
)

type env struct {
	config *viper.Viper
}

const (
	SERVER_ADDRESS_KEY    = "server.address"
	SERVER_PORT_KEY       = "server.port"
	SERVER_PROFILE_KEY    = "server.profile"
	SERVER_SERVERNAME_KEY = "server.name"
	LOGGER_DIR_KEY        = "logger.dir"
	LOGGER_MAXAGE_KEY     = "logger.maxAge"
	ZIPKIN_URL_KEY        = "zipkin.url"
)

var e *env
var once sync.Once

func init() {
	e = &env{
		config: viper.New(),
	}
	server.RegisterEventHook(server.ConfigChangeEvent, e.MergeConfig)
}

func GetInstance() config.IConfig {
	return e
}

func (e *env) SetBaseConfig(reader io.Reader, configType string) {
	once.Do(func() {
		bc := e.config
		bc.SetConfigType(configType)
		bc.SetConfigName("bootstrap")
		bc.ReadConfig(reader)
	})
}

func (e *env) MergeConfig(eventType server.EventName, eventInfo interface{}) (err error) {
	logger.Info("[ENV] config changed")
	if cfg, ok := eventInfo.(*viper.Viper); ok {
		logger.Info("[ENV] config changed type viper.Viper")
		err = e.config.MergeConfigMap(cfg.AllSettings())
	}
	if cfg, ok := eventInfo.(map[string]interface{}); ok {
		err = e.config.MergeConfigMap(cfg)
	}
	if err != nil {
		logger.Error("[ENV] config changed Error:" + err.Error())
	}
	return
}

func (e *env) Shutdown() {

}
func (e *env) Read() error {
	return nil
}
func (e *env) Watch() {
}

func (nc *env) Get(key string) interface{} {
	return nc.config.Get(key)
}

func (nc *env) Set(key string, value interface{}) {
	nc.config.Set(key, value)
}

func (e *env) GetStringWithDefault(key string, defaultString string) string {
	v := e.GetString(key)
	if len(v) == 0 {
		return defaultString
	}
	return v
}

func (e *env) GetStringMapString(key string) map[string]string {
	return e.config.GetStringMapString(key)
}

func (nc *env) GetInt(key string) int {
	return nc.config.GetInt(key)
}

func (nc *env) GetIntWithDefault(key string, defaultInt int) int {
	v := nc.GetInt(key)
	if v == 0 {
		return defaultInt
	}
	return v
}

func (nc *env) GetInt64WithDefault(key string, defaultInt64 int64) int64 {
	v := nc.GetInt64(key)
	if v == 0 {
		return defaultInt64
	}
	return v
}
func (nc *env) GetInt64(key string) int64 {
	return nc.config.GetInt64(key)
}
func (nc *env) GetString(key string) string {
	return nc.config.GetString(key)
}
func (nc *env) GetStringSlice(key string) []string {
	return nc.config.GetStringSlice(key)
}

func (nc *env) GetUint64(key string) uint64 {
	return nc.config.GetUint64(key)
}
func (nc *env) GetUint64WithDefault(key string, defaultUint64 uint64) uint64 {
	v := nc.GetUint64(key)
	if v == 0 {
		return defaultUint64
	}
	return v
}
func (nc *env) GetUint(key string) uint {
	return nc.config.GetUint(key)
}
func (nc *env) GetUintWithDefault(key string, defaultUint uint) uint {
	v := nc.GetUint(key)
	if v == 0 {
		return defaultUint
	}
	return v
}
func (nc *env) GetBool(key string) bool {
	return nc.config.GetBool(key)
}
func (nc *env) GetFloat64(key string) float64 {
	return nc.config.GetFloat64(key)
}
