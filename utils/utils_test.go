package utils

import (
	"encoding/json"
	"fmt"
	"testing"
)

type TestJson struct {
	Id   int64           `json:"id"`
	Name string          `json:"name"`
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"raw"`
	T    int64           `json:"t"`
}

func TestXxx(t *testing.T) {
	s := "{\"id\":1,\"name\":\"test\",\"type\":\"1\",\"raw\":{\"t\":1}}"
	tj := &TestJson{}
	err := UnmarshalFromString(s, tj)
	if err != nil {
		fmt.Println(err)
		return
	}
	bytes, _ := tj.Raw.MarshalJSON()
	fmt.Println(string(bytes))
	err = Unmarshal(bytes, tj)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tj.T)

}
