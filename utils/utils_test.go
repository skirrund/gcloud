package utils

import (
	"fmt"
	"testing"

	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/utils/mth_code"
)

type TestJson struct {
	Id   int64     `json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type"`
	T    *DateTime `json:"t"`
}

func TestXxx(t *testing.T) {
	str, err := mth_code.MthDesDecrypt("_mpCipyuv1XroqO1Xsda2GKPVeMa7RUl")
	str1, _ := mth_code.MthDesEncrypt("422429197108230024")
	fmt.Println(str, err, str1)
	logger.Info("123123123")
	t.Log("222")
}
