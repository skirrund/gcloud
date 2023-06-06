package worker

import (
	"fmt"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func TestAsyncExecute(t *testing.T) {
	a, b, c, d := 1, 2, 3, 4
	ids := []*int{&a, &b, &c, &d}
	for i := range ids {
		wg.Add(1)
		//wg.Add(1)
		x := ids[i]
		AsyncExecute(func() {
			demo(*x)
		})
	}
	wg.Wait()
	println("done")
}

func demo(x int) {
	wg.Done()
	fmt.Println(x)
}
