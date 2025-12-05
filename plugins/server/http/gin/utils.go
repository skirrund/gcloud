package gin

import (
	"errors"
	"reflect"

	"github.com/skirrund/gcloud/response"
	"github.com/skirrund/gcloud/utils/validator"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func CheckQueryParamsWithErrorMsg(name string, v *string, errorMsg string, ctx *gin.Context) bool {
	str := ctx.Query(name)
	return CheckParamsWithErrorMsg(name, str, v, errorMsg, ctx)
}

func CheckHeaderParamsWithErrorMsg(name string, v *string, errorMsg string, ctx *gin.Context) bool {
	str := ctx.GetHeader(name)
	return CheckParamsWithErrorMsg(name, str, v, errorMsg, ctx)
}

func CheckParamsWithErrorMsg(name string, str string, v *string, errorMsg string, ctx *gin.Context) bool {
	*v = str
	if len(str) == 0 {
		if len(errorMsg) == 0 {
			ctx.JSON(200, response.ValidateError[any](name+"不能为空"))
		} else {
			ctx.JSON(200, response.ValidateError[any](errorMsg))
		}

		return false
	}
	return true
}

func CheckPostFormParamsWithErrorMsg(name string, v *string, errorMsg string, ctx *gin.Context) bool {
	str := ctx.PostForm(name)
	if len(str) == 0 {
		str = ctx.Query(name)
	}
	return CheckParamsWithErrorMsg(name, str, v, errorMsg, ctx)
}

func CheckQueryParams(name string, v *string, ctx *gin.Context) bool {
	return CheckQueryParamsWithErrorMsg(name, v, "", ctx)
}

func CheckPostFormParams(name string, v *string, ctx *gin.Context) bool {
	return CheckPostFormParamsWithErrorMsg(name, v, "", ctx)
}

func CheckHeaderParams(name string, v *string, ctx *gin.Context) bool {
	return CheckHeaderParamsWithErrorMsg(name, v, "", ctx)
}

func SendJSON(ctx *gin.Context, data any) {
	ctx.JSON(200, data)
}

// 优先级
// path > form > query > cookie > header > json > raw_body
func ShouldBind(ctx *gin.Context, obj any) (err error) {
	if obj == nil {
		return nil
	}
	t := reflect.TypeOf(obj)
	kind := t.Kind()
	if kind == reflect.Pointer {
		t = t.Elem()
	} else {
		return nil
	}
	kind = t.Kind()
	var tagHeader, tagQuery, tagPath, tagForm bool
	if kind == reflect.Struct {
		nfs := t.NumField()
		for i := range nfs {
			sf := t.Field(i)
			if len(sf.Tag.Get("header")) > 0 {
				tagHeader = true
			}
			if len(sf.Tag.Get("query")) > 0 {
				tagQuery = true
			}
			if len(sf.Tag.Get("form")) > 0 {
				tagForm = true
			}
			if len(sf.Tag.Get("path")) > 0 {
				tagPath = true
			}
		}
	}
	if err = ctx.ShouldBind(obj); err != nil {
		return err
	}
	if tagHeader {
		ctx.ShouldBindHeader(obj)
	}
	if tagQuery || tagForm {
		ctx.ShouldBindWith(obj, binding.Query)
		ctx.ShouldBindWith(obj, binding.Form)
	}
	if tagPath {
		ctx.ShouldBindUri(obj)
	}
	return nil
}

func ShouldBindAndValidate(ctx *gin.Context, data any) error {
	err := ShouldBind(ctx, data)
	if err != nil {
		return err
	}
	err = validator.ValidateStruct(data)
	if err != nil {
		return errors.New(validator.ErrResp(err))
	}
	return nil

}

func QueryArray(ctx *gin.Context, name string) []string {
	array := ctx.QueryArray(name)
	var params []string
	if len(array) > 0 {
		for _, a := range array {
			if len(a) == 0 {
				continue
			}
			if strings.Contains(a, ",") {
				tmp := strings.Split(a, ",")
				params = append(params, tmp...)
			} else {
				params = append(params, a)
			}
		}
	}
	return params
}
func PostFormArray(ctx *gin.Context, name string) []string {
	array := ctx.PostFormArray(name)
	var params []string
	if len(array) > 0 {
		for _, a := range array {
			if len(a) == 0 {
				continue
			}
			if strings.Contains(a, ",") {
				tmp := strings.Split(a, ",")
				params = append(params, tmp...)
			} else {
				params = append(params, a)
			}
		}
	}
	if len(params) > 0 {
		return params
	} else {
		return QueryArray(ctx, name)
	}
}
