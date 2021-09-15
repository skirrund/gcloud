package response

type Response struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	SubMessage string      `json:"subMessage"`
	Result     interface{} `json:"result"`
	Success    bool        `json:"success"`
}

type Msginfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	ERROR   string = "500000"
	SUCCESS string = "200000"
	MSG     string = "消息处理成功"
)

var SUCCESS_MSG = Msginfo{Code: SUCCESS, Message: MSG}
var VALIDATE_ERROR = Msginfo{Code: "400000", Message: "数据校验不合法"}
var ACCESS_PERM_DENIED = Msginfo{Code: "400001", Message: "对不起,您没有访问权限"}
var CAPTCHA_VALIDATE_FAIL = Msginfo{Code: "400002", Message: "图片验证码验证失败"}
var CAPTCHA_EXPIRE = Msginfo{Code: "400003", Message: "图片验证码已失效"}
var PHARMACY_NOT_BIND = Msginfo{Code: "400004", Message: "药店信息未绑定"}
var PHARMACY_NOT_EXIST = Msginfo{Code: "400005", Message: "药店信息不存在"}
var PHONE_OR_PASSWORD_EMPTY = Msginfo{Code: "400006", Message: "手机号或密码不能为空"}
var PHONE_OR_PASSWORD_ERROR = Msginfo{Code: "400007", Message: "手机号或密码不正确"}
var PHONE_ERROR_FORMAT = Msginfo{Code: "400008", Message: "手机号格式不正确"}
var VALIDATE_API_ERROR = Msginfo{Code: "400010", Message: "数据校验不合法"}
var REQUEST_FREQUENTLY_ERROR = Msginfo{Code: "400011", Message: "请求过于频繁"}
var EXCEPTION = Msginfo{Code: ERROR, Message: "系统繁忙,请稍后重试"}
var COMMON_EXCEPTION = Msginfo{Code: "500001", Message: "消息处理异常"}
var DB_INSERT_EXCEPTION = Msginfo{Code: "500002", Message: "数据插入异常"}
var DB_UPDATE_EXCEPTION = Msginfo{Code: "500003", Message: "数据更新异常"}
var DB_SELECT_EXCEPTION = Msginfo{Code: "500004", Message: "数据查询异常"}
var DB_KEY_DUPLICATE = Msginfo{Code: "500011", Message: "主键或唯一性约束冲突"}
var SMS_TEMPLATE_ERROR = Msginfo{Code: "4B2001", Message: "无效的短信模板"}

var FLOW_EXCEPTION = Msginfo{Code: "503000", Message: "请求过于拥挤，请稍候重试"}
var DEGRADE_EXCEPTION = Msginfo{Code: "503001", Message: "请求被降级，请稍候重试"}
var PARAM_FLOW_EXCEPTION = Msginfo{Code: "503002", Message: "请求过于拥挤，请稍候重试"}
var SYSTEM_BLOCK_EXCEPTION = Msginfo{Code: "503003", Message: "系统被保护，请稍候重试"}
var AUTHORITY_EXCEPTION = Msginfo{Code: "503004", Message: "访问被限制，请稍候重试"}

func NewMsgInfo(code string, msg string) *Msginfo {
	return &Msginfo{
		Code:    code,
		Message: msg,
	}
}

func (mi Msginfo) String() string {
	return mi.Code + ":" + mi.Message
}

func (resp *Response) IsSuccess() bool {
	return resp.Code == SUCCESS
}

func CreateMsgInfo(msgInfo Msginfo, subMsg string) Response {
	return Response{
		Code:       msgInfo.Code,
		Message:    msgInfo.Message,
		SubMessage: subMsg,
		Success:    false,
	}
}

func ValidateError(subMsg string) Response {
	return Response{
		Code:       VALIDATE_API_ERROR.Code,
		Message:    VALIDATE_API_ERROR.Message,
		SubMessage: subMsg,
	}
}

func CreateMsgInfoResult(msgInfo Msginfo, result interface{}) Response {
	return Response{
		Code:    msgInfo.Code,
		Message: msgInfo.Message,
		Result:  result,
	}
}

func Fail(msg string) Response {
	return Response{
		Code:    EXCEPTION.Code,
		Message: msg,
		Success: false,
	}
}

func FailSubMsg(subMsg string) Response {
	return Response{
		Code:       EXCEPTION.Code,
		Message:    EXCEPTION.Message,
		Success:    false,
		SubMessage: subMsg,
	}
}

func DefaultFailSubMsg(subMsg string) Response {
	return Response{
		Code:       EXCEPTION.Code,
		Message:    EXCEPTION.Message,
		Success:    false,
		SubMessage: subMsg,
	}
}

func FailSubMsgResult(msg string, subMsg string, result interface{}) Response {
	return Response{
		Code:       EXCEPTION.Code,
		Message:    msg,
		Success:    false,
		SubMessage: subMsg,
		Result:     result,
	}
}
func Success(data interface{}) Response {
	return Response{
		Code:    SUCCESS_MSG.Code,
		Result:  data,
		Message: SUCCESS_MSG.Message,
		Success: true,
	}
}

func Create(code string, msg string, subMsg string, data interface{}) Response {
	return Response{
		Code:       code,
		Result:     data,
		Message:    msg,
		SubMessage: subMsg,
	}
}
