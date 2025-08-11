package utils

import (
	"context"
	"fmt"
	"testing"

	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/utils/decimal"
)

type TestJson struct {
	Id   decimal.Decimal `json:"id"`
	Name string          `json:"name"`
	Type string          `json:"type"`
	T    DateTime        `json:"t"`
}

func TestXxx(t *testing.T) {
	str := `{"id":12.0}`
	testStruct := &TestJson{}
	err := UnmarshalFromString(str, &testStruct)
	testStruct.Id.IntPart()
	logger.InfoContext(context.TODO(), err)
	fmt.Println(err, testStruct)
}
