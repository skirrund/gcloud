package decimal

import (
	"fmt"
	"testing"

	"github.com/skirrund/gcloud/utils"
)

type TestObj struct {
	TestD Decimal `json:"testD"`
}

func TestXxx(t *testing.T) {
	a := 1100.1
	b := a * 100
	fmt.Println(b)
	c, _ := NewFromString("1100.123456")
	d := c.Mul(NewFromInt(100)).RoundFloor(2)
	fmt.Println(d)
	obj := &TestObj{}
	err := utils.UnmarshalFromString(`{"testD":""}`, obj)
	fmt.Println(err, obj.TestD)
}
