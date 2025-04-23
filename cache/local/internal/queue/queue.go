package queue

import (
	"sync"

	"github.com/skirrund/gcloud/cache/local/internal/xmath"
)

type Growable[T any] struct {
	mutex    sync.Mutex
	notEmpty sync.Cond
	notFull  sync.Cond
	buf      []T
	head     int
	tail     int
	count    int
	minCap   int
	maxCap   int
}

func NewGrowable[T any](minCap, maxCap uint32) *Growable[T] {
	minCap = xmath.RoundUpPowerOf2(minCap)
	maxCap = xmath.RoundUpPowerOf2(maxCap)

	g := &Growable[T]{
		buf:    make([]T, minCap),
		minCap: int(minCap),
		maxCap: int(maxCap),
	}

	g.notEmpty = *sync.NewCond(&g.mutex)
	g.notFull = *sync.NewCond(&g.mutex)

	return g
}

func (g *Growable[T]) Push(item T) {
	g.mutex.Lock()
	for g.count == g.maxCap {
		g.notFull.Wait()
	}
	g.push(item)
	g.mutex.Unlock()
}

func (g *Growable[T]) push(item T) {
	g.grow()
	g.buf[g.tail] = item
	g.tail = g.next(g.tail)
	g.count++
	g.notEmpty.Signal()
}

func (g *Growable[T]) Pop() T {
	g.mutex.Lock()
	for g.count == 0 {
		g.notEmpty.Wait()
	}
	item := g.pop()
	g.mutex.Unlock()
	return item
}

func (g *Growable[T]) TryPop() (T, bool) {
	var zero T
	g.mutex.Lock()
	if g.count == 0 {
		g.mutex.Unlock()
		return zero, false
	}
	item := g.pop()
	g.mutex.Unlock()
	return item, true
}

func (g *Growable[T]) pop() T {
	var zero T

	item := g.buf[g.head]
	g.buf[g.head] = zero

	g.head = g.next(g.head)
	g.count--

	g.notFull.Signal()

	return item
}

func (g *Growable[T]) Clear() {
	g.mutex.Lock()
	for g.count > 0 {
		g.pop()
	}
	g.mutex.Unlock()
}

func (g *Growable[T]) grow() {
	if g.count != len(g.buf) {
		return
	}
	g.resize()
}

func (g *Growable[T]) resize() {
	newBuf := make([]T, g.count<<1)
	if g.tail > g.head {
		copy(newBuf, g.buf[g.head:g.tail])
	} else {
		n := copy(newBuf, g.buf[g.head:])
		copy(newBuf[n:], g.buf[:g.tail])
	}

	g.head = 0
	g.tail = g.count
	g.buf = newBuf
}

func (g *Growable[T]) next(i int) int {
	return (i + 1) & (len(g.buf) - 1)
}
