package http

import (
	"fmt"
	"testing"

	"github.com/skirrund/gcloud/server/decoder"
)

func TestXxx(t *testing.T) {
	var r []byte
	resp, err := PostJSON("test", "http://127.0.0.1:8080/test", nil, nil, &r)
	fmt.Println(resp, err)
}

func TestDecoder(t *testing.T) {
	de := decoder.StringDecoder{}
	resp := []byte("哈喽a")
	var b []byte
	de.DecoderObj(resp, &b)
	fmt.Println(string(b))
}
