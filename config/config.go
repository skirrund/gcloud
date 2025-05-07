package config

type IConfig interface {
	Get(key string) any
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
	Set(key string, value any)
	GetFloat64(key string) float64
	Shutdown() error
	Read() error
	Watch() error
	GetStringMapString(key string) map[string]string
}
