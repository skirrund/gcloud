package env

import (
	bytes2 "bytes"
	"io"
	"os"
	"strings"

	"github.com/skirrund/gcloud/server"

	"github.com/skirrund/gcloud/logger"
	"github.com/spf13/viper"
)

type env struct {
	config *viper.Viper
	base   map[string]interface{}
}

const (
	SERVER_ADDRESS_KEY    = "server.address"
	SERVER_PORT_KEY       = "server.port"
	SERVER_PROFILE_KEY    = "server.profile"
	SERVER_CONFIGFILE_KEY = "server.config.file"
	SERVER_SERVERNAME_KEY = "server.name"
	LOGGER_DIR_KEY        = "logger.dir"
	LOGGER_MAXAGE_KEY     = "logger.maxAge"
	LOGGER_CONSOLE        = "logger.console"
	LOGGER_JSON           = "logger.json"
	ZIPKIN_URL_KEY        = "zipkin.url"
)

var e *env

func init() {
	e = &env{
		config: viper.New(),
		base:   make(map[string]interface{}),
	}
	server.RegisterEventHook(server.ConfigChangeEvent, e.MergeConfig)
}

func GetInstance() *env {
	return e
}

func (e *env) LoadProfileBaseConfig(profile string, configType string) {
	cfgPath := e.GetString(SERVER_CONFIGFILE_KEY)
	path, _ := os.Getwd()
	if len(cfgPath) == 0 {
		if len(profile) > 0 {
			cfgPath = path + "/conf/bootstrap-" + profile + "." + configType
			logger.Info(cfgPath)
		}
	}
	logger.Info("[ENV] load config file profile:", cfgPath)
	if len(cfgPath) > 0 {
		_, err := os.Stat(cfgPath)
		if err == nil {
			logger.Info("path>>>>" + path)
			contents, err := os.ReadFile(cfgPath)
			if err == nil {
				pcfg := viper.New()
				ct := cfgPath[strings.LastIndex(cfgPath, ".")+1:]
				pcfg.SetConfigType(ct)
				err = pcfg.ReadConfig(bytes2.NewReader(contents))
				if err == nil {
					settings := pcfg.AllSettings()
					e.config.MergeConfigMap(settings)
					e.base = e.config.AllSettings()
				}
			}
		} else {
			logger.Error("[ENV] load config file profile error:", err.Error())
		}
	}
}

func (e *env) SetBaseConfig(reader io.Reader, configType string) error {
	cfg := e.config
	cfg.SetConfigType(configType)
	cfg.SetConfigName("base")
	err := cfg.ReadConfig(reader)
	if err != nil {
		logger.Error("[ENV] SetBaseConfig error", err.Error())
		return err
	}
	e.base = cfg.AllSettings()
	return nil
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

func (e *env) Shutdown() error {
	return nil
}
func (e *env) Read() error {
	return nil
}
func (e *env) Watch() error {
	return nil
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
func (nc *env) UnmarshalKey(key string, objPtr interface{}) error {
	return nc.config.UnmarshalKey(key, objPtr)
}
