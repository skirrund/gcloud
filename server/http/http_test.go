package http

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	var p = make(map[string]interface{})
	p["p"] = "1"
	p["p2"] = 2
	p["array"] = []float64{1.12345, 2.1234567}
	r := getFormData(p)
	fmt.Println(r)
}
