package node

import (
	"sync/atomic"
	"unsafe"

	"github.com/skirrund/gcloud/cache/local/internal/unixtime"
)

// BEC is a cache entry that provide the following features:
//
// 1. Base
//
// 2. Expiration
//
// 3. Cost
type BEC[K comparable, V any] struct {
	key        K
	value      V
	prev       *BEC[K, V]
	next       *BEC[K, V]
	prevExp    *BEC[K, V]
	nextExp    *BEC[K, V]
	expiration uint32
	cost       uint32
	state      uint32
	frequency  uint8
	queueType  uint8
}

// NewBEC creates a new BEC.
func NewBEC[K comparable, V any](key K, value V, expiration, cost uint32) Node[K, V] {
	return &BEC[K, V]{
		key:        key,
		value:      value,
		expiration: expiration,
		cost:       cost,
		state:      aliveState,
	}
}

// CastPointerToBEC casts a pointer to BEC.
func CastPointerToBEC[K comparable, V any](ptr unsafe.Pointer) Node[K, V] {
	return (*BEC[K, V])(ptr)
}

func (n *BEC[K, V]) Key() K {
	return n.key
}

func (n *BEC[K, V]) Value() V {
	return n.value
}

func (n *BEC[K, V]) AsPointer() unsafe.Pointer {
	return unsafe.Pointer(n)
}

func (n *BEC[K, V]) Prev() Node[K, V] {
	return n.prev
}

func (n *BEC[K, V]) SetPrev(v Node[K, V]) {
	if v == nil {
		n.prev = nil
		return
	}
	n.prev = (*BEC[K, V])(v.AsPointer())
}

func (n *BEC[K, V]) Next() Node[K, V] {
	return n.next
}

func (n *BEC[K, V]) SetNext(v Node[K, V]) {
	if v == nil {
		n.next = nil
		return
	}
	n.next = (*BEC[K, V])(v.AsPointer())
}

func (n *BEC[K, V]) PrevExp() Node[K, V] {
	return n.prevExp
}

func (n *BEC[K, V]) SetPrevExp(v Node[K, V]) {
	if v == nil {
		n.prevExp = nil
		return
	}
	n.prevExp = (*BEC[K, V])(v.AsPointer())
}

func (n *BEC[K, V]) NextExp() Node[K, V] {
	return n.nextExp
}

func (n *BEC[K, V]) SetNextExp(v Node[K, V]) {
	if v == nil {
		n.nextExp = nil
		return
	}
	n.nextExp = (*BEC[K, V])(v.AsPointer())
}

func (n *BEC[K, V]) HasExpired() bool {
	return n.expiration <= unixtime.Now()
}

func (n *BEC[K, V]) Expiration() uint32 {
	return n.expiration
}

func (n *BEC[K, V]) Cost() uint32 {
	return n.cost
}

func (n *BEC[K, V]) IsAlive() bool {
	return atomic.LoadUint32(&n.state) == aliveState
}

func (n *BEC[K, V]) Die() {
	atomic.StoreUint32(&n.state, deadState)
}

func (n *BEC[K, V]) Frequency() uint8 {
	return n.frequency
}

func (n *BEC[K, V]) IncrementFrequency() {
	n.frequency = minUint8(n.frequency+1, maxFrequency)
}

func (n *BEC[K, V]) DecrementFrequency() {
	n.frequency--
}

func (n *BEC[K, V]) ResetFrequency() {
	n.frequency = 0
}

func (n *BEC[K, V]) MarkSmall() {
	n.queueType = smallQueueType
}

func (n *BEC[K, V]) IsSmall() bool {
	return n.queueType == smallQueueType
}

func (n *BEC[K, V]) MarkMain() {
	n.queueType = mainQueueType
}

func (n *BEC[K, V]) IsMain() bool {
	return n.queueType == mainQueueType
}

func (n *BEC[K, V]) Unmark() {
	n.queueType = unknownQueueType
}
