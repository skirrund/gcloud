package sha256

import (
	"testing"
)

func TestSha256(t *testing.T) {
	payload := "{\"IdCard\":\"" + "610125198501257136" + "\",\"Name\":\"" + "史哲理" + "\"}"
	t.Log(Sha256Hex([]byte(payload)))
}
