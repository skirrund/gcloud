package expiry

import "github.com/skirrund/gcloud/cache/local/internal/generated/node"

type Fixed[K comparable, V any] struct {
	q          *queue[K, V]
	deleteNode func(node.Node[K, V])
}

func NewFixed[K comparable, V any](deleteNode func(node.Node[K, V])) *Fixed[K, V] {
	return &Fixed[K, V]{
		q:          newQueue[K, V](),
		deleteNode: deleteNode,
	}
}

func (f *Fixed[K, V]) Add(n node.Node[K, V]) {
	f.q.push(n)
}

func (f *Fixed[K, V]) Delete(n node.Node[K, V]) {
	f.q.delete(n)
}

func (f *Fixed[K, V]) DeleteExpired() {
	for !f.q.isEmpty() && f.q.head.HasExpired() {
		f.deleteNode(f.q.pop())
	}
}

func (f *Fixed[K, V]) Clear() {
	f.q.clear()
}
