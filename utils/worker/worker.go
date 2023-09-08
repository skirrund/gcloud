package worker

import (
	"math"
	"runtime/debug"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/skirrund/gcloud/logger"
)

type worker struct {
	p     *ants.Pool
	Limit int
}

var DefaultWorker worker

const (
	DefaultLimit int = math.MaxUint16
)

func init() {
	p, err := ants.NewPool(DefaultLimit, ants.WithExpiryDuration(10*time.Second))
	if err != nil {
		panic(err)
	}
	DefaultWorker = worker{
		p:     p,
		Limit: DefaultLimit,
	}
}

func Init(limit int) worker {
	p, _ := ants.NewPool(DefaultLimit, ants.WithExpiryDuration(10*time.Second))
	return worker{
		p:     p,
		Limit: limit,
	}
}
func (w worker) Release() {
	w.p.Release()
}

func (w worker) Execute(f func()) error {
	return w.p.Submit(f)
}

func AsyncExecute(f func()) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("[worker] error recover :", err, "\n", string(debug.Stack()))
		}
	}()
	return DefaultWorker.p.Submit(f)
}
