package utils

import (
	"github.com/skirrund/gcloud/response"
)

type Error struct {
	Err     error
	Msginfo *response.Msginfo
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	if e.Msginfo != nil {
		return e.Msginfo.Message
	}
	return ""
}

func NewError(err error, msginfo *response.Msginfo) *Error {
	return &Error{
		Err:     err,
		Msginfo: msginfo,
	}
}
