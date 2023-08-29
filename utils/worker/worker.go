package worker

import (
	"runtime/debug"

	"github.com/skirrund/gcloud/logger"
)

type worker struct {
	quene chan struct{}
	Limit uint64
}

var DefaultWorker worker

const (
	DefaultLimit uint64 = 512
)

func init() {
	DefaultWorker = worker{
		quene: make(chan struct{}, DefaultLimit),
		Limit: DefaultLimit,
	}
}

func Init(limit uint64) worker {
	return worker{
		quene: make(chan struct{}, limit),
		Limit: limit,
	}
}
func (w worker) Execute(f func()) {
	w.quene <- struct{}{}
	go func() {
		defer func() {
			<-w.quene
		}()
		execute(f)
	}()
}

func execute(f func()) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[worker] error recover :", err, "\n", string(debug.Stack()))
		}
	}()
	f()
}

func AsyncExecute(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("[worker] error recover :", err, "\n", string(debug.Stack()))
			}
		}()
		f()
	}()
}
