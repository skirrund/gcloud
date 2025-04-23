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
	SetWithTtl("test", TestSetStruct{Name: "test"}, 2*time.Second)
	SetWithTtl("test", testInt, 2*time.Second)
	time.Sleep(1 * time.Second)
	obj := Get("test")
	fmt.Println(obj)

}
