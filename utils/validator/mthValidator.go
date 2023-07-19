package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// InitTrans 初始化翻译器
func InitTrans(locale string, validate binding.StructValidator) (err error) {
	//修改gin框架中的Validator属性，实现自定制
	if v, ok := validate.Engine().(*validator.Validate); ok {
		return InitValidator(locale, v)
	}
	return
}
