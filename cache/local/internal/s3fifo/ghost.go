package s3fifo

import (
	"github.com/skirrund/gcloud/cache/local/internal/deque"
	"github.com/skirrund/gcloud/cache/local/internal/generated/node"
	"github.com/skirrund/gcloud/cache/local/internal/hashtable"
)

type ghost[K comparable, V any] struct {
	q         *deque.Deque[uint64]
	m         map[uint64]struct{}
	main      *main[K, V]
	small     *small[K, V]
	hasher    hashtable.Hasher[K]
	evictNode func(node.Node[K, V])
}

func newGhost[K comparable, V any](main *main[K, V], evictNode func(node.Node[K, V])) *ghost[K, V] {
	return &ghost[K, V]{
		q:         &deque.Deque[uint64]{},
		m:         make(map[uint64]struct{}),
		main:      main,
		hasher:    hashtable.NewHasher[K](),
		evictNode: evictNode,
	}
}

func (g *ghost[K, V]) isGhost(n node.Node[K, V]) bool {
	h := g.hasher.Hash(n.Key())
	_, ok := g.m[h]
	return ok
}

func (g *ghost[K, V]) insert(n node.Node[K, V]) {
	g.evictNode(n)

	h := g.hasher.Hash(n.Key())

	if _, ok := g.m[h]; ok {
		return
	}

	maxLength := g.small.length() + g.main.length()
	if maxLength == 0 {
		return
	}

	for g.q.Len() >= maxLength {
		v := g.q.PopFront()
		delete(g.m, v)
	}

	g.q.PushBack(h)
	g.m[h] = struct{}{}
}

func (g *ghost[K, V]) clear() {
	g.q.Clear()
	for k := range g.m {
		delete(g.m, k)
	}
}
