package utils

type Error struct {
	Err     error
	Msginfo *Msginfo
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

func NewError(err error, msginfo *Msginfo) *Error {
	return &Error{
		Err:     err,
		Msginfo: msginfo,
	}
}
