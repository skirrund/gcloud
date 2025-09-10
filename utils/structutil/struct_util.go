package structutil

import "reflect"

// 复制对象，dst必须为指针
func StructCopy(src, dst any) {
	if src == nil || dst == nil {
		return
	}
	ptrObj := reflect.ValueOf(src)
	if ptrObj.Kind() == reflect.Pointer {
		// 获取指针指向的值
		src = ptrObj.Elem().Interface()
	}
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst).Elem()
	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		dstField := dstVal.FieldByName(srcVal.Type().Field(i).Name)
		if dstField.IsValid() && dstField.CanSet() {
			dstField.Set(srcField)
		}
	}
}
