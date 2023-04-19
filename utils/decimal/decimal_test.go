package decimal

import (
	"fmt"
	"testing"
)

func TestXxx(t *testing.T) {
	a := 1100.1
	b := a * 100
	fmt.Println(b)
	c, _ := NewFromString("1100.123456")
	d := c.Mul(NewFromInt(100)).RoundFloor(2)
	fmt.Println(d)
}
