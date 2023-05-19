package http

import (
	"sync"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	var resp []byte
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			PostFormDataUrlWithTimeout("https://www.baidu.com", nil, nil, &resp, 15*time.Second)
		}()
	}
	wg.Wait()
}
