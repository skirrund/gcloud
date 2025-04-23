package xruntime

import (
	"runtime"
)

const (
	// CacheLineSize is useful for preventing false sharing.
	CacheLineSize = 64
)

// Parallelism returns the maximum possible number of concurrently running goroutines.
func Parallelism() uint32 {
	//nolint:gosec // there will never be an overflow
	maxProcs := uint32(runtime.GOMAXPROCS(0))
	//nolint:gosec // there will never be an overflow
	numCPU := uint32(runtime.NumCPU())
	if maxProcs < numCPU {
		return maxProcs
	}
	return numCPU
}
