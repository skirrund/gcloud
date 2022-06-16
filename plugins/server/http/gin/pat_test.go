package gin

import (
	"regexp"
	"testing"
)

func TestXxx(t *testing.T) {
	reg := regexp.MustCompile(`.*\.(js|css|png|jpg|jpeg|gif).*$`)
	t.Log(">>>>>>", reg.MatchString("http://123.com/js.png"))
}
