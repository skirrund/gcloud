package worker

import (
	"fmt"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func TestAsyncExecute(t *testing.T) {
	for i := 0; i <= 10; i++ {
		wg.Add(1)
		x := i
		AsyncExecute(func() {
			demo(x)
		})
	}
	wg.Wait()
	println("done")
}

func demo(x int) {
	wg.Done()
	fmt.Println(x)
}
