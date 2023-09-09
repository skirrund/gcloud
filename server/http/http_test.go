package http

import (
	"fmt"
	"testing"
	"time"

	"github.com/skirrund/gcloud/response"
	"github.com/skirrund/gcloud/server/decoder"
)

func TestXxx(t *testing.T) {
	var b []byte
	params := map[string]interface{}{"locationUrl": ""}
	r := response.Response[any]{}
	_, err := PostJSONUrl("https://www.baidu.com", nil, params, &r)
	fmt.Println(string(b), err)
	time.Sleep(3 * time.Second)
}

func TestDecoder(t *testing.T) {
	de := decoder.StringDecoder{}
	resp := []byte("哈喽a")
	var b []byte
	de.DecoderObj(resp, &b)
	fmt.Println(string(b))
}
