package s3fifo

import (
	"github.com/skirrund/gcloud/cache/local/internal/generated/node"
)

type small[K comparable, V any] struct {
	q         *queue[K, V]
	main      *main[K, V]
	ghost     *ghost[K, V]
	cost      int
	maxCost   int
	evictNode func(node.Node[K, V])
}

func newSmall[K comparable, V any](
	maxCost int,
	main *main[K, V],
	ghost *ghost[K, V],
	evictNode func(node.Node[K, V]),
) *small[K, V] {
	return &small[K, V]{
		q:         newQueue[K, V](),
		main:      main,
		ghost:     ghost,
		maxCost:   maxCost,
		evictNode: evictNode,
	}
}

func (s *small[K, V]) insert(n node.Node[K, V]) {
	s.q.push(n)
	n.MarkSmall()
	s.cost += int(n.Cost())
}

func (s *small[K, V]) evict() {
	if s.cost == 0 {
		return
	}

	n := s.q.pop()
	s.cost -= int(n.Cost())
	n.Unmark()
	if !n.IsAlive() || n.HasExpired() {
		s.evictNode(n)
		return
	}

	if n.Frequency() > 1 {
		s.main.insert(n)
		for s.main.isFull() {
			s.main.evict()
		}
		n.ResetFrequency()
		return
	}

	s.ghost.insert(n)
}

func (s *small[K, V]) delete(n node.Node[K, V]) {
	s.cost -= int(n.Cost())
	n.Unmark()
	s.q.delete(n)
}

func (s *small[K, V]) length() int {
	return s.q.length()
}

func (s *small[K, V]) clear() {
	s.q.clear()
	s.cost = 0
}
