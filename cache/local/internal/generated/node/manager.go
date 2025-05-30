package node

import (
	"strings"
	"unsafe"
)

const (
	unknownQueueType uint8 = iota
	smallQueueType
	mainQueueType

	maxFrequency uint8 = 3
)

const (
	aliveState uint32 = iota
	deadState
)

// Node is a cache entry.
type Node[K comparable, V any] interface {
	// Key returns the key.
	Key() K
	// Value returns the value.
	Value() V
	// AsPointer returns the node as a pointer.
	AsPointer() unsafe.Pointer
	// Prev returns the previous node in the eviction policy.
	Prev() Node[K, V]
	// SetPrev sets the previous node in the eviction policy.
	SetPrev(v Node[K, V])
	// Next returns the next node in the eviction policy.
	Next() Node[K, V]
	// SetNext sets the next node in the eviction policy.
	SetNext(v Node[K, V])
	// PrevExp returns the previous node in the expiration policy.
	PrevExp() Node[K, V]
	// SetPrevExp sets the previous node in the expiration policy.
	SetPrevExp(v Node[K, V])
	// NextExp returns the next node in the expiration policy.
	NextExp() Node[K, V]
	// SetNextExp sets the next node in the expiration policy.
	SetNextExp(v Node[K, V])
	// HasExpired returns true if node has expired.
	HasExpired() bool
	// Expiration returns the expiration time.
	Expiration() uint32
	// Cost returns the cost of the node.
	Cost() uint32
	// IsAlive returns true if the entry is available in the hash-table.
	IsAlive() bool
	// Die sets the node to the dead state.
	Die()
	// Frequency returns the frequency of the node.
	Frequency() uint8
	// IncrementFrequency increments the frequency of the node.
	IncrementFrequency()
	// DecrementFrequency decrements the frequency of the node.
	DecrementFrequency()
	// ResetFrequency resets the frequency.
	ResetFrequency()
	// MarkSmall sets the status to the small queue.
	MarkSmall()
	// IsSmall returns true if node is in the small queue.
	IsSmall() bool
	// MarkMain sets the status to the main queue.
	MarkMain()
	// IsMain returns true if node is in the main queue.
	IsMain() bool
	// Unmark sets the status to unknown.
	Unmark()
}

func Equals[K comparable, V any](a, b Node[K, V]) bool {
	if a == nil {
		return b == nil || b.AsPointer() == nil
	}
	if b == nil {
		return a.AsPointer() == nil
	}
	return a.AsPointer() == b.AsPointer()
}

type Config struct {
	WithExpiration bool
	WithCost       bool
}

type Manager[K comparable, V any] struct {
	create      func(key K, value V, expiration, cost uint32) Node[K, V]
	fromPointer func(ptr unsafe.Pointer) Node[K, V]
}

func NewManager[K comparable, V any](c Config) *Manager[K, V] {
	var sb strings.Builder
	sb.WriteString("b")
	if c.WithExpiration {
		sb.WriteString("e")
	}
	if c.WithCost {
		sb.WriteString("c")
	}
	nodeType := sb.String()
	m := &Manager[K, V]{}

	switch nodeType {
	case "bec":
		m.create = NewBEC[K, V]
		m.fromPointer = CastPointerToBEC[K, V]
	case "bc":
		m.create = NewBC[K, V]
		m.fromPointer = CastPointerToBC[K, V]
	case "be":
		m.create = NewBE[K, V]
		m.fromPointer = CastPointerToBE[K, V]
	case "b":
		m.create = NewB[K, V]
		m.fromPointer = CastPointerToB[K, V]
	default:
		panic("not valid nodeType")
	}
	return m
}

func (m *Manager[K, V]) Create(key K, value V, expiration, cost uint32) Node[K, V] {
	return m.create(key, value, expiration, cost)
}

func (m *Manager[K, V]) FromPointer(ptr unsafe.Pointer) Node[K, V] {
	return m.fromPointer(ptr)
}

func minUint8(a, b uint8) uint8 {
	if a < b {
		return a
	}

	return b
}
