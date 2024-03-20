package local

import (
	"fmt"
	"testing"
)

type TestSetStruct struct {
	Id   int64
	Name string
}

func TestSet(t *testing.T) {
	Set("test", &TestSetStruct{Name: "test"}, 10)
	obj := Get("test").(TestSetStruct)
	fmt.Println(obj.Name)

}
