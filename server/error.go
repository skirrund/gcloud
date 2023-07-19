package server

import "github.com/skirrund/gcloud/response"

type Error struct {
	Code   string
	Msg    string
	SubMsg string
	Result bool
}

func (e Error) Error() string {
	if len(e.SubMsg) > 0 {
		return e.SubMsg
	} else {
		return e.Msg
	}
}
func NewError(msg string) *Error {
	return &Error{
		Code: response.ERROR,
		Msg:  msg,
	}
}
func NewErrorSubMsg(subMsg string) *Error {
	return &Error{
		Code:   response.ERROR,
		SubMsg: subMsg,
	}
}
func NewErrorAllMsg(msg, subMsg string) *Error {
	return &Error{
		Code:   response.ERROR,
		Msg:    msg,
		SubMsg: subMsg,
	}
}
