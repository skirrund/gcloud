package node

import (
	"sync/atomic"
	"unsafe"

	"github.com/skirrund/gcloud/cache/local/internal/unixtime"
)

// BE is a cache entry that provide the following features:
//
// 1. Base
//
// 2. Expiration
type BE[K comparable, V any] struct {
	key        K
	value      V
	prev       *BE[K, V]
	next       *BE[K, V]
	prevExp    *BE[K, V]
	nextExp    *BE[K, V]
	expiration uint32
	state      uint32
	frequency  uint8
	queueType  uint8
}

// NewBE creates a new BE.
func NewBE[K comparable, V any](key K, value V, expiration, cost uint32) Node[K, V] {
	return &BE[K, V]{
		key:        key,
		value:      value,
		expiration: expiration,
		state:      aliveState,
	}
}

// CastPointerToBE casts a pointer to BE.
func CastPointerToBE[K comparable, V any](ptr unsafe.Pointer) Node[K, V] {
	return (*BE[K, V])(ptr)
}

func (n *BE[K, V]) Key() K {
	return n.key
}

func (n *BE[K, V]) Value() V {
	return n.value
}

func (n *BE[K, V]) AsPointer() unsafe.Pointer {
	return unsafe.Pointer(n)
}

func (n *BE[K, V]) Prev() Node[K, V] {
	return n.prev
}

func (n *BE[K, V]) SetPrev(v Node[K, V]) {
	if v == nil {
		n.prev = nil
		return
	}
	n.prev = (*BE[K, V])(v.AsPointer())
}

func (n *BE[K, V]) Next() Node[K, V] {
	return n.next
}

func (n *BE[K, V]) SetNext(v Node[K, V]) {
	if v == nil {
		n.next = nil
		return
	}
	n.next = (*BE[K, V])(v.AsPointer())
}

func (n *BE[K, V]) PrevExp() Node[K, V] {
	return n.prevExp
}

func (n *BE[K, V]) SetPrevExp(v Node[K, V]) {
	if v == nil {
		n.prevExp = nil
		return
	}
	n.prevExp = (*BE[K, V])(v.AsPointer())
}

func (n *BE[K, V]) NextExp() Node[K, V] {
	return n.nextExp
}

func (n *BE[K, V]) SetNextExp(v Node[K, V]) {
	if v == nil {
		n.nextExp = nil
		return
	}
	n.nextExp = (*BE[K, V])(v.AsPointer())
}

func (n *BE[K, V]) HasExpired() bool {
	return n.expiration <= unixtime.Now()
}

func (n *BE[K, V]) Expiration() uint32 {
	return n.expiration
}

func (n *BE[K, V]) Cost() uint32 {
	return 1
}

func (n *BE[K, V]) IsAlive() bool {
	return atomic.LoadUint32(&n.state) == aliveState
}

func (n *BE[K, V]) Die() {
	atomic.StoreUint32(&n.state, deadState)
}

func (n *BE[K, V]) Frequency() uint8 {
	return n.frequency
}

func (n *BE[K, V]) IncrementFrequency() {
	n.frequency = minUint8(n.frequency+1, maxFrequency)
}

func (n *BE[K, V]) DecrementFrequency() {
	n.frequency--
}

func (n *BE[K, V]) ResetFrequency() {
	n.frequency = 0
}

func (n *BE[K, V]) MarkSmall() {
	n.queueType = smallQueueType
}

func (n *BE[K, V]) IsSmall() bool {
	return n.queueType == smallQueueType
}

func (n *BE[K, V]) MarkMain() {
	n.queueType = mainQueueType
}

func (n *BE[K, V]) IsMain() bool {
	return n.queueType == mainQueueType
}

func (n *BE[K, V]) Unmark() {
	n.queueType = unknownQueueType
}
