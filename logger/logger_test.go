package logger

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	initLog("logger", "test", "111", true, false, 1*time.Hour)
	Info("info.......")
	Warn("warn.....")
}
