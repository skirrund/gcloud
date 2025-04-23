package local

import (
	"github.com/skirrund/gcloud/cache/local/internal/core"
	"github.com/skirrund/gcloud/cache/local/internal/generated/node"
	"github.com/skirrund/gcloud/cache/local/internal/unixtime"
)

func zeroValue[V any]() V {
	var zero V
	return zero
}

// Extension is an access point for inspecting and performing low-level operations based on the cache's runtime
// characteristics. These operations are optional and dependent on how the cache was constructed
// and what abilities the implementation exposes.
type Extension[K comparable, V any] struct {
	cache *core.Cache[K, V]
}

func newExtension[K comparable, V any](cache *core.Cache[K, V]) Extension[K, V] {
	return Extension[K, V]{
		cache: cache,
	}
}

func (e Extension[K, V]) createEntry(n node.Node[K, V]) Entry[K, V] {
	var expiration int64
	if e.cache.WithExpiration() {
		expiration = unixtime.StartTime() + int64(n.Expiration())
	}

	return Entry[K, V]{
		key:        n.Key(),
		value:      n.Value(),
		expiration: expiration,
		cost:       n.Cost(),
	}
}

// GetQuietly returns the value associated with the key in this cache.
//
// Unlike Get in the cache, this function does not produce any side effects
// such as updating statistics or the eviction policy.
func (e Extension[K, V]) GetQuietly(key K) (V, bool) {
	n, ok := e.cache.GetNodeQuietly(key)
	if !ok {
		return zeroValue[V](), false
	}

	return n.Value(), true
}

// GetEntry returns the cache entry associated with the key in this cache.
func (e Extension[K, V]) GetEntry(key K) (Entry[K, V], bool) {
	n, ok := e.cache.GetNode(key)
	if !ok {
		return Entry[K, V]{}, false
	}

	return e.createEntry(n), true
}

// GetEntryQuietly returns the cache entry associated with the key in this cache.
//
// Unlike GetEntry, this function does not produce any side effects
// such as updating statistics or the eviction policy.
func (e Extension[K, V]) GetEntryQuietly(key K) (Entry[K, V], bool) {
	n, ok := e.cache.GetNodeQuietly(key)
	if !ok {
		return Entry[K, V]{}, false
	}

	return e.createEntry(n), true
}
