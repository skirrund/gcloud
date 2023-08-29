package worker

import (
	"runtime/debug"

	"github.com/skirrund/gcloud/logger"
)

type Worker struct {
	quene chan struct{}
	Limit uint64
}

var DefaultWorker Worker

const (
	DefaultLimit uint64 = 512
)

func init() {
	DefaultWorker = Worker{
		quene: make(chan struct{}, DefaultLimit),
		Limit: DefaultLimit,
	}
}

func Init(limit uint64) Worker {
	return Worker{
		quene: make(chan struct{}, limit),
		Limit: limit,
	}
}
func (w Worker) Execute(f func()) {
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
