package request

import (
	"context"
	"io"
	"os"
	"time"
)

type File struct {
	//FieldName string
	//if len(FileBytes)>0 will use FileBytes as filestream
	File      *os.File
	FileName  string
	FileBytes []byte
}

type Request struct {
	Url         string
	ServiceName string
	Path        string
	Headers     map[string]string
	Params      io.Reader
	IsJson      bool
	RespResult  interface{}
	TimeOut     time.Duration
	Method      string
	LbOptions   *LbOptions
	HasFile     bool
	Context     context.Context
}

type LbOptions struct {
	MaxRetriesOnNextServiceInstance int
	RetryableStatusCodes            []int
	Enabled                         bool
	Retrys                          int
	CurrentStatuCode                int
	CurrentError                    error
}

func NewDefaultLbOptions() *LbOptions {
	return &LbOptions{
		MaxRetriesOnNextServiceInstance: 1,
		RetryableStatusCodes:            []int{404, 502, 503},
		Enabled:                         true,
		Retrys:                          0,
	}
}

func (r *Request) WithContext(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	r.Context = ctx
}
