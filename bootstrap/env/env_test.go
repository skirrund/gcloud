package env

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/skirrund/gcloud/parser"
)

func TestParseJavaproperties(t *testing.T) {
	pcfg := parser.NewDefaultParser()
	pcfg.SetConfigType("properties")
	err := pcfg.ReadConfig(bytes.NewReader([]byte(`
		# 这是一个注释
		# 这是一个注释
		# 这是一个注释
		pbm.common.web.logging.level=INFO
		datasource.queryFields=true false
		`)))
	fmt.Println(err)
	str := pcfg.GetString("pbm.common.web.logging.level")
	fmt.Println(str)
	str = pcfg.GetString("datasource.dsn")
	fmt.Println(str)
	s := pcfg.GetStringSlice("datasource.queryFields")
	fmt.Println(s)
}
