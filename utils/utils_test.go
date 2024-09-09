package utils

import (
	"fmt"
	"testing"
)

type TestJson struct {
	Id   int64    `json:"id"`
	Name string   `json:"name"`
	Type string   `json:"type"`
	T    DateTime `json:"t"`
}

func TestXxx(t *testing.T) {
	str := `{"t":null}`
	testStruct := &TestJson{}
	err := UnmarshalFromString(str, &testStruct)
	fmt.Println(err, testStruct)
}
