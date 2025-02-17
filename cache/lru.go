package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// LRUCache implements a [Cache] that supports different [EvictionPolicy].
// LRU instead of Lru https://google.github.io/styleguide/go/decisions.html#initialisms
type LRUCache struct {
	size int

	// mu locaks all the buckets and the order list.
	mu sync.RWMutex
	// buckets maps to key values where value is a pointer to an element in the order list.
	// The actual value is stored in the [list.Element] Value field.
	buckets map[string]map[string]*list.Element
	// order is by default the insertion order
	// If user use [EvictionPolicyLRU] or [EvictionPolicyMRU], then the order is also updated
	// during Get and Set.
	order *list.List
}

type cacheEntry struct {
	value      []byte
	expiration time.Time
}

func NewLRUCache(size int) *LRUCache {
	return &LRUCache{
		size:    size,
		buckets: make(map[string]map[string]*list.Element),
		order:   list.New(),
	}
}

// TODO: evict, update policy etc.
func (c *LRUCache) Set(bucket string, key string, value []byte, opts Options) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create bucket if not exists
	b, ok := c.buckets[bucket]
	if !ok {
		b = make(map[string]*list.Element)
		c.buckets[bucket] = b
	}

	// Create new entry
	expiration := time.Time{}
	if opts.TTL > 0 {
		expiration = time.Now().Add(opts.TTL)
	}
	entry := cacheEntry{value: value, expiration: expiration}

	// Check if the key already exists
	e, ok := b[key]
	if ok {
		e.Value = entry
		return nil
	}

	// Add new key to the bucket
	e = c.order.PushBack(entry)
	b[key] = e

	return nil
}

func (c *LRUCache) Get(bucket string, key string, opts Options) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// TODO: define error for not found
	b, ok := c.buckets[bucket]
	if !ok {
		return nil, fmt.Errorf("bucket %s not found", bucket)
	}

	e, ok := b[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	entry := e.Value.(cacheEntry)
	if !entry.expiration.IsZero() && entry.expiration.Before(time.Now()) {
		return nil, fmt.Errorf("key %s expired", key)
	}

	// TODO: update order

	return entry.value, nil
}
