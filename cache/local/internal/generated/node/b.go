package node

import (
	"sync/atomic"
	"unsafe"
)

// B is a cache entry that provide the following features:
//
// 1. Base
type B[K comparable, V any] struct {
	key       K
	value     V
	prev      *B[K, V]
	next      *B[K, V]
	state     uint32
	frequency uint8
	queueType uint8
}

// NewB creates a new B.
func NewB[K comparable, V any](key K, value V, expiration, cost uint32) Node[K, V] {
	return &B[K, V]{
		key:   key,
		value: value,
		state: aliveState,
	}
}

// CastPointerToB casts a pointer to B.
func CastPointerToB[K comparable, V any](ptr unsafe.Pointer) Node[K, V] {
	return (*B[K, V])(ptr)
}

func (n *B[K, V]) Key() K {
	return n.key
}

func (n *B[K, V]) Value() V {
	return n.value
}

func (n *B[K, V]) AsPointer() unsafe.Pointer {
	return unsafe.Pointer(n)
}

func (n *B[K, V]) Prev() Node[K, V] {
	return n.prev
}

func (n *B[K, V]) SetPrev(v Node[K, V]) {
	if v == nil {
		n.prev = nil
		return
	}
	n.prev = (*B[K, V])(v.AsPointer())
}

func (n *B[K, V]) Next() Node[K, V] {
	return n.next
}

func (n *B[K, V]) SetNext(v Node[K, V]) {
	if v == nil {
		n.next = nil
		return
	}
	n.next = (*B[K, V])(v.AsPointer())
}

func (n *B[K, V]) PrevExp() Node[K, V] {
	panic("not implemented")
}

func (n *B[K, V]) SetPrevExp(v Node[K, V]) {
	panic("not implemented")
}

func (n *B[K, V]) NextExp() Node[K, V] {
	panic("not implemented")
}

func (n *B[K, V]) SetNextExp(v Node[K, V]) {
	panic("not implemented")
}

func (n *B[K, V]) HasExpired() bool {
	return false
}

func (n *B[K, V]) Expiration() uint32 {
	panic("not implemented")
}

func (n *B[K, V]) Cost() uint32 {
	return 1
}

func (n *B[K, V]) IsAlive() bool {
	return atomic.LoadUint32(&n.state) == aliveState
}

func (n *B[K, V]) Die() {
	atomic.StoreUint32(&n.state, deadState)
}

func (n *B[K, V]) Frequency() uint8 {
	return n.frequency
}

func (n *B[K, V]) IncrementFrequency() {
	n.frequency = minUint8(n.frequency+1, maxFrequency)
}

func (n *B[K, V]) DecrementFrequency() {
	n.frequency--
}

func (n *B[K, V]) ResetFrequency() {
	n.frequency = 0
}

func (n *B[K, V]) MarkSmall() {
	n.queueType = smallQueueType
}

func (n *B[K, V]) IsSmall() bool {
	return n.queueType == smallQueueType
}

func (n *B[K, V]) MarkMain() {
	n.queueType = mainQueueType
}

func (n *B[K, V]) IsMain() bool {
	return n.queueType == mainQueueType
}

func (n *B[K, V]) Unmark() {
	n.queueType = unknownQueueType
}
