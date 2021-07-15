package validator

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/skirrund/gcloud/logger"
)

func Validate(obj interface{}) error {
	t := reflect.TypeOf(obj)
	kind := t.Kind()
	var bean reflect.Value
	if kind == reflect.Ptr {
		t = t.Elem()
		bean = reflect.ValueOf(obj).Elem()
	}
	if kind == reflect.Struct {
		logger.Error("[common] NewOptions: check type error not Struct")
		return nil
	}

	num := t.NumField()
	for i := 0; i != num; i++ {
		f := t.Field(i)
		v := bean.Field(i)
		stag := f.Tag
		tag, ok := stag.Lookup("blank")
		value := v.Interface()
		if ok {
			if isBlank(value) {
				return errors.New(tag)
			}
		}
		tag, ok = stag.Lookup("length")
		if ok {
			if strings.Contains(tag, "-") {
				strs := strings.Split(tag, "-")
				if len(strs) == 1 {
					min, err := strconv.ParseInt(strs[0], 10, 64)
					if err != nil && !checkLength(value, min, 0) {
						return errors.New("length error[" + f.Name + "],min:" + strs[0])
					}
				}
			}

		}

	}
	return nil
}

func checkLength(obj interface{}, min int64, max int64) bool {
	check := true
	if str, ok := obj.(string); ok {
		l := int64(utf8.RuneCountInString(str))
		if min > 0 {
			check = (l >= min)
		}
		if max > 0 {
			check = (check && l <= max)
		}
	}
	return check

}

func isBlank(obj interface{}) bool {
	if obj == nil {
		return true
	}
	if str, ok := obj.(string); ok {
		return len(strings.TrimSpace(str)) == 0
	}
	return false
}
