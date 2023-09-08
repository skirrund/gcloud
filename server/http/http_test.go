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
	params := map[string]interface{}{"locationUrl": "https%3A%2F%2Ftest-h5.mediext.com%2Fhealth%2F%3FcardTypeCode%3D160%23%2Fhealth-rainbow-activation"}
	r := response.Response[any]{}
	_, err := PostJSONUrl("https://test-h5.mediext.com/gateway/wx/v1/check", nil, params, &r)
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
