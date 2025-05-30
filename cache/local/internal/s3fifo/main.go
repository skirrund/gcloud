package s3fifo

import (
	"github.com/skirrund/gcloud/cache/local/internal/generated/node"
)

const maxReinsertions = 20

type main[K comparable, V any] struct {
	q         *queue[K, V]
	cost      int
	maxCost   int
	evictNode func(node.Node[K, V])
}

func newMain[K comparable, V any](maxCost int, evictNode func(node.Node[K, V])) *main[K, V] {
	return &main[K, V]{
		q:         newQueue[K, V](),
		maxCost:   maxCost,
		evictNode: evictNode,
	}
}

func (m *main[K, V]) insert(n node.Node[K, V]) {
	m.q.push(n)
	n.MarkMain()
	m.cost += int(n.Cost())
}

func (m *main[K, V]) evict() {
	reinsertions := 0
	for m.cost > 0 {
		n := m.q.pop()

		if !n.IsAlive() || n.HasExpired() || n.Frequency() == 0 {
			n.Unmark()
			m.cost -= int(n.Cost())
			m.evictNode(n)
			return
		}

		// to avoid the worst case O(n), we remove the 20th reinserted consecutive element.
		reinsertions++
		if reinsertions >= maxReinsertions {
			n.Unmark()
			m.cost -= int(n.Cost())
			m.evictNode(n)
			return
		}

		m.q.push(n)
		n.DecrementFrequency()
	}
}

func (m *main[K, V]) delete(n node.Node[K, V]) {
	m.cost -= int(n.Cost())
	n.Unmark()
	m.q.delete(n)
}

func (m *main[K, V]) length() int {
	return m.q.length()
}

func (m *main[K, V]) clear() {
	m.q.clear()
	m.cost = 0
}

func (m *main[K, V]) isFull() bool {
	return m.cost >= m.maxCost
}
