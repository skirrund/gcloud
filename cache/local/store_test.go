package local

import (
	"fmt"
	"testing"
	"time"
)

type TestSetStruct struct {
	Id   int64
	Name string
}

func TestSet(t *testing.T) {
	testInt := 123
	Set("test", TestSetStruct{Name: "test"}, 1)
	Set("test", testInt, 1)
	time.Sleep(3 * time.Second)
	obj := Get("test")
	fmt.Println(obj)

}
