package unixtime

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	// We need this package because time.Now() is slower, allocates memory,
	// and we don't need a more precise time for the expiry time (and most other operations).
	now       uint32
	startTime int64

	mutex         sync.Mutex
	countInstance int
	done          chan struct{}
)

func startTimer() {
	done = make(chan struct{})
	atomic.StoreInt64(&startTime, time.Now().Unix())
	atomic.StoreUint32(&now, uint32(0))

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				//nolint:gosec // there will never be an overflow
				atomic.StoreUint32(&now, uint32(t.Unix()-StartTime()))
			case <-done:
				return
			}
		}
	}()
}

// Start should be called when the cache instance is created to initialize the timer.
func Start() {
	mutex.Lock()
	defer mutex.Unlock()

	if countInstance == 0 {
		startTimer()
	}

	countInstance++
}

// Stop should be called when closing and stopping the cache instance to stop the timer.
func Stop() {
	mutex.Lock()
	defer mutex.Unlock()

	countInstance--
	if countInstance == 0 {
		done <- struct{}{}
		close(done)
	}
}

// Now returns time as a Unix time, the number of seconds elapsed since program start.
func Now() uint32 {
	return atomic.LoadUint32(&now)
}

// SetNow sets the current time.
//
// NOTE: use only for testing and debugging.
func SetNow(t uint32) {
	atomic.StoreUint32(&now, t)
}

// StartTime returns the start time of the program.
func StartTime() int64 {
	return atomic.LoadInt64(&startTime)
}
