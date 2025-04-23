package expiry

import "github.com/skirrund/gcloud/cache/local/internal/generated/node"

type Disabled[K comparable, V any] struct{}

func NewDisabled[K comparable, V any]() *Disabled[K, V] {
	return &Disabled[K, V]{}
}

func (d *Disabled[K, V]) Add(n node.Node[K, V]) {
}

func (d *Disabled[K, V]) Delete(n node.Node[K, V]) {
}

func (d *Disabled[K, V]) DeleteExpired() {
}

func (d *Disabled[K, V]) Clear() {
}
