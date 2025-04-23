package xruntime

import (
	"math/rand/v2"
)

func Fastrand() uint32 {
	//nolint:gosec // we don't need a cryptographically secure random number generator
	return rand.Uint32()
}
