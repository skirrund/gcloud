package http

import (
	"testing"
)

func TestGet(t *testing.T) {
	var r string
	code, err := GetUrl("https://www.baidu.com", nil, nil, &r)
	t.Log(code, "::", err)
}
