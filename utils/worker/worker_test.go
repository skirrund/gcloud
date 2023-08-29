package worker

import (
	"fmt"
	"testing"
)

func TestAsyncExecute(t *testing.T) {
	for i := 0; i < 1000; i++ {
		i1 := i
		fmt.Println("send:", i1)
		DefaultWorker.Execute(func() {
			fmt.Println(i1)

		})
	}
	println("done")
}
