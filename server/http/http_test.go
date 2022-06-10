package http

import (
	"testing"

	"github.com/skirrund/gcloud/utils"
)

func TestGet(t *testing.T) {
	s, _ := utils.MarshalToString(1)
	t.Log(s)
}
