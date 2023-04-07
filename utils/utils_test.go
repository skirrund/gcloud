package utils

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bytedance/sonic"
)

type TestJson struct {
	Id   int64     `json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type"`
	T    *DateTime `json:"t"`
}

func TestXxx(t *testing.T) {
	s := `{"id":1,"name":"test","type":"1","t":""}`
	tj := &TestJson{}
	err := sonic.ConfigStd.UnmarshalFromString(s, tj)
	err = json.Unmarshal([]byte(s), tj)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tj)
}
