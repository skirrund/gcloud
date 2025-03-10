package utils

import (
	"fmt"
	"net/url"
	"testing"

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
	fmt.Println(err, testStruct)
	l, err := url.Parse("https://h5.mediext.com/health/?productName=ls2025&cardTypeCode=425&pbm_project_id=PR000001&_t_1740624715288")
	fmt.Println(err, l)
}
