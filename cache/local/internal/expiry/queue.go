package expiry

import "github.com/skirrund/gcloud/cache/local/internal/generated/node"

type queue[K comparable, V any] struct {
	head node.Node[K, V]
	tail node.Node[K, V]
	len  int
}

func newQueue[K comparable, V any]() *queue[K, V] {
	return &queue[K, V]{}
}

func (q *queue[K, V]) length() int {
	return q.len
}

func (q *queue[K, V]) isEmpty() bool {
	return q.length() == 0
}

func (q *queue[K, V]) push(n node.Node[K, V]) {
	if q.isEmpty() {
		q.head = n
		q.tail = n
	} else {
		n.SetPrevExp(q.tail)
		q.tail.SetNextExp(n)
		q.tail = n
	}

	q.len++
}

func (q *queue[K, V]) pop() node.Node[K, V] {
	if q.isEmpty() {
		return nil
	}

	result := q.head
	q.delete(result)
	return result
}

func (q *queue[K, V]) delete(n node.Node[K, V]) {
	next := n.NextExp()
	prev := n.PrevExp()

	if node.Equals(prev, nil) {
		if node.Equals(next, nil) && !node.Equals(q.head, n) {
			return
		}

		q.head = next
	} else {
		prev.SetNextExp(next)
		n.SetPrevExp(nil)
	}

	if node.Equals(next, nil) {
		q.tail = prev
	} else {
		next.SetPrevExp(prev)
		n.SetNextExp(nil)
	}

	q.len--
}

func (q *queue[K, V]) clear() {
	for !q.isEmpty() {
		q.pop()
	}
}
