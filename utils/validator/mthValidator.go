package validator

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"reflect"
	"strings"
)

// 定义一个全局翻译器
var trans ut.Translator

// InitTrans 初始化翻译器
func InitTrans(locale string) (err error) {
	//修改gin框架中的Validator属性，实现自定制
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册一个获取json tag的自定义方法
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("tip"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		zhT := zh.New() //中文翻译器
		enT := en.New() //英文翻译器

		// 第一个参数是备用（fallback）的语言环境
		// 后面的参数是应该支持的语言环境（支持多个）
		// uni := ut.New(zhT, zhT) 也是可以的
		uni := ut.New(enT, zhT, enT)

		// locale 通常取决于 http 请求头的 'Accept-Language'
		var ok bool
		// 也可以使用 uni.FindTranslator(...) 传入多个locale进行查找
		trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		// 添加额外翻译
		_ = v.RegisterTranslation("required", trans, func(ut ut.Translator) error {
			return ut.Add("required", "{0}不能为空", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("required", fe.Field())
			return t
		})
		_ = v.RegisterTranslation("gte", trans, func(ut ut.Translator) error {
			return ut.Add("gte", "{0}必须大于等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("gte", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("gt", trans, func(ut ut.Translator) error {
			return ut.Add("gt", "{0}必须大于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("gt", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("lte", trans, func(ut ut.Translator) error {
			return ut.Add("lte", "{0}必须小于等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("lte", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("lt", trans, func(ut ut.Translator) error {
			return ut.Add("lt", "{0}必须小于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("lt", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("len", trans, func(ut ut.Translator) error {
			return ut.Add("len", "{0}长度必须是{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("len", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("min", trans, func(ut ut.Translator) error {
			return ut.Add("min", "{0}最小值为{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("min", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("max", trans, func(ut ut.Translator) error {
			return ut.Add("max", "{0}最大值为{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("max", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("eq", trans, func(ut ut.Translator) error {
			return ut.Add("eq", "{0}不等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("eq", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("ne", trans, func(ut ut.Translator) error {
			return ut.Add("ne", "{0}不能等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("ne", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("eqfield", trans, func(ut ut.Translator) error {
			return ut.Add("eqfield", "{0}必须等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("eqfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("eqcsfield", trans, func(ut ut.Translator) error {
			return ut.Add("eqcsfield", "{0}必须等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("eqcsfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("necsfield", trans, func(ut ut.Translator) error {
			return ut.Add("necsfield", "{0}不能等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("necsfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("gtcsfield", trans, func(ut ut.Translator) error {
			return ut.Add("gtcsfield", "{0}必须大于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("gtcsfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("gtecsfield", trans, func(ut ut.Translator) error {
			return ut.Add("gtecsfield", "{0}必须大于或等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("gtecsfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("ltcsfield", trans, func(ut ut.Translator) error {
			return ut.Add("ltcsfield", "{0}必须小于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("ltcsfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("ltecsfield", trans, func(ut ut.Translator) error {
			return ut.Add("ltecsfield", "{0}必须小于或等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("ltecsfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("gtecsfield", trans, func(ut ut.Translator) error {
			return ut.Add("gtecsfield", "{0}必须大于或等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("gtecsfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("ltcsfield", trans, func(ut ut.Translator) error {
			return ut.Add("ltcsfield", "{0}必须小于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("ltcsfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("ltecsfield", trans, func(ut ut.Translator) error {
			return ut.Add("ltecsfield", "{0}必须小于或等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("ltecsfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("nefield", trans, func(ut ut.Translator) error {
			return ut.Add("nefield", "{0}不能等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("nefield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("gtfield", trans, func(ut ut.Translator) error {
			return ut.Add("gtfield", "{0}必须大于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("gtfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("gtefield", trans, func(ut ut.Translator) error {
			return ut.Add("gtefield", "{0}必须大于或等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("gtefield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("ltfield", trans, func(ut ut.Translator) error {
			return ut.Add("ltfield", "{0}必须小于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("ltfield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("ltefield", trans, func(ut ut.Translator) error {
			return ut.Add("ltefield", "{0}必须小于或等于{1}", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("ltefield", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("alpha", trans, func(ut ut.Translator) error {
			return ut.Add("alpha", "{0}只能包含字母", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("alpha", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("alphanum", trans, func(ut ut.Translator) error {
			return ut.Add("alphanum", "{0}只能包含字母和数字", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("alphanum", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("numeric", trans, func(ut ut.Translator) error {
			return ut.Add("numeric", "{0}必须是一个有效的数值", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("numeric", fe.Field(), fe.Param())
			return t
		})
		_ = v.RegisterTranslation("number", trans, func(ut ut.Translator) error {
			return ut.Add("number", "{0}必须是一个有效的数字", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("number", fe.Field(), fe.Param())
			return t
		})
		//_ = v.RegisterTranslation("required_without", trans, func(ut ut.Translator) error {
		//	return ut.Add("required_without", "{0} 不能为空", true)
		//}, func(ut ut.Translator, fe validator.FieldError) string {
		//	t, _ := ut.T("required_without", fe.Field())
		//	return t
		//})
		//_ = v.RegisterTranslation("required_without_all", trans, func(ut ut.Translator) error {
		//	return ut.Add("required_without_all", "{0} 不能为空", true)
		//}, func(ut ut.Translator, fe validator.FieldError) string {
		//	t, _ := ut.T("required_without_all", fe.Field())
		//	return t
		//})

		// 注册翻译器
		switch locale {
		case "en":
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		case "zh":
			err = zhTranslations.RegisterDefaultTranslations(v, trans)
		default:
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		}
		return
	}
	return
}

func addValueToMap(fields map[string]string) map[string]interface{} {
	res := make(map[string]interface{})
	for field, err := range fields {
		fieldArr := strings.SplitN(field, ".", 2)
		if len(fieldArr) > 1 {
			NewFields := map[string]string{fieldArr[1]: err}
			returnMap := addValueToMap(NewFields)
			if res[fieldArr[0]] != nil {
				for k, v := range returnMap {
					res[fieldArr[0]].(map[string]interface{})[k] = v
				}
			} else {
				res[fieldArr[0]] = returnMap
			}
			continue
		} else {
			res[field] = err
			continue
		}
	}
	return res
}

// 去掉结构体名称前缀
func removeTopStruct(fields map[string]string) map[string]interface{} {
	lowerMap := map[string]string{}
	for field, err := range fields {
		fieldArr := strings.SplitN(field, ".", 2)
		lowerMap[fieldArr[1]] = err
	}
	res := addValueToMap(lowerMap)
	return res
}

//handler中调用的错误翻译方法
func ErrResp(err error) string {
	errs, ok := err.(validator.ValidationErrors)
	fmt.Println(reflect.TypeOf(err))
	if !ok {
		return err.Error()
	}
	errs = errs[0:1]
	errMap := removeTopStruct(errs.Translate(trans))
	if len(errMap) > 0 {
		return handleErrMap("", errMap)
	}
	return ""
}

func handleErrMap(keyPrefix string, errRes interface{}) string {
	if v, ok := errRes.(map[string]interface{}); ok {
		for key, value := range v {
			if v, ok := value.(map[string]interface{}); ok {
				return handleErrMap(unionPrefix(keyPrefix, key), v)
			} else {
				return unionPrefix(keyPrefix, fmt.Sprintf("%v", value))
			}
		}
	} else {
		return unionPrefix(keyPrefix, fmt.Sprintf("%v", v))
	}
	return unionPrefix(keyPrefix, fmt.Sprintf("%v", errRes))
}

func unionPrefix(prefix string, v string) string {
	if prefix != "" {
		return prefix + "." + v
	} else {
		return v
	}
}
