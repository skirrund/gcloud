package node

import (
	"sync/atomic"
	"unsafe"
)

// BC is a cache entry that provide the following features:
//
// 1. Base
//
// 2. Cost
type BC[K comparable, V any] struct {
	key       K
	value     V
	prev      *BC[K, V]
	next      *BC[K, V]
	cost      uint32
	state     uint32
	frequency uint8
	queueType uint8
}

// NewBC creates a new BC.
func NewBC[K comparable, V any](key K, value V, expiration, cost uint32) Node[K, V] {
	return &BC[K, V]{
		key:   key,
		value: value,
		cost:  cost,
		state: aliveState,
	}
}

// CastPointerToBC casts a pointer to BC.
func CastPointerToBC[K comparable, V any](ptr unsafe.Pointer) Node[K, V] {
	return (*BC[K, V])(ptr)
}

func (n *BC[K, V]) Key() K {
	return n.key
}

func (n *BC[K, V]) Value() V {
	return n.value
}

func (n *BC[K, V]) AsPointer() unsafe.Pointer {
	return unsafe.Pointer(n)
}

func (n *BC[K, V]) Prev() Node[K, V] {
	return n.prev
}

func (n *BC[K, V]) SetPrev(v Node[K, V]) {
	if v == nil {
		n.prev = nil
		return
	}
	n.prev = (*BC[K, V])(v.AsPointer())
}

func (n *BC[K, V]) Next() Node[K, V] {
	return n.next
}

func (n *BC[K, V]) SetNext(v Node[K, V]) {
	if v == nil {
		n.next = nil
		return
	}
	n.next = (*BC[K, V])(v.AsPointer())
}

func (n *BC[K, V]) PrevExp() Node[K, V] {
	panic("not implemented")
}

func (n *BC[K, V]) SetPrevExp(v Node[K, V]) {
	panic("not implemented")
}

func (n *BC[K, V]) NextExp() Node[K, V] {
	panic("not implemented")
}

func (n *BC[K, V]) SetNextExp(v Node[K, V]) {
	panic("not implemented")
}

func (n *BC[K, V]) HasExpired() bool {
	return false
}

func (n *BC[K, V]) Expiration() uint32 {
	panic("not implemented")
}

func (n *BC[K, V]) Cost() uint32 {
	return n.cost
}

func (n *BC[K, V]) IsAlive() bool {
	return atomic.LoadUint32(&n.state) == aliveState
}

func (n *BC[K, V]) Die() {
	atomic.StoreUint32(&n.state, deadState)
}

func (n *BC[K, V]) Frequency() uint8 {
	return n.frequency
}

func (n *BC[K, V]) IncrementFrequency() {
	n.frequency = minUint8(n.frequency+1, maxFrequency)
}

func (n *BC[K, V]) DecrementFrequency() {
	n.frequency--
}

func (n *BC[K, V]) ResetFrequency() {
	n.frequency = 0
}

func (n *BC[K, V]) MarkSmall() {
	n.queueType = smallQueueType
}

func (n *BC[K, V]) IsSmall() bool {
	return n.queueType == smallQueueType
}

func (n *BC[K, V]) MarkMain() {
	n.queueType = mainQueueType
}

func (n *BC[K, V]) IsMain() bool {
	return n.queueType == mainQueueType
}

func (n *BC[K, V]) Unmark() {
	n.queueType = unknownQueueType
}
