package parser

import (
	"github.com/skirrund/gcloud/parser/javaproperties"
	"github.com/spf13/viper"
)

func NewDefaultParser() *viper.Viper {
	cr := NewDefaultCodecRegistry()
	return viper.NewWithOptions(viper.WithCodecRegistry(cr))
}
func NewDefaultCodecRegistry() *viper.DefaultCodecRegistry {
	cr := viper.NewCodecRegistry()
	cr.RegisterCodec("properties", &javaproperties.Codec{})
	return cr
}
func NewParserWithOptions(options ...viper.Option) {
	cr := NewDefaultCodecRegistry()
	options = append(options, viper.WithCodecRegistry(cr))
	viper.NewWithOptions(options...)
}
