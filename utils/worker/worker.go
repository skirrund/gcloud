package worker

import (
	"github.com/skirrund/gcloud/logger"
	"runtime/debug"
)

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
