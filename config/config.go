package config

import (
	"github.com/skirrund/gcloud/server"

	"github.com/spf13/viper"
)

type IConfig interface {
	Get(key string) interface{}
	GetString(key string) string
	GetStringWithDefault(key string, defaultString string) string
	GetStringSlice(key string) []string
	GetInt64(key string) int64
	GetInt64WithDefault(key string, defaultInt64 int64) int64
	GetInt(key string) int
	GetIntWithDefault(key string, defaultInt int) int
	GetUint(key string) uint
	GetUintWithDefault(key string, defaultUint uint) uint
	GetUint64(key string) uint64
	GetUint64WithDefault(key string, defaultUint64 uint64) uint64
	GetBool(key string) bool
	GetFloat64(key string) float64
	Shutdown()
	Read() error
	Watch()
	MergeConfig(eventType server.EventName, eventInfo interface{}) error
	SetBaseConfig(cfg *viper.Viper)
	GetStringMapString(key string) map[string]string
}
