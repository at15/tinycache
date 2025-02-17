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
	capacity         int
	ttlCheckInterval time.Duration
	stop             chan struct{}

	// mu locaks all the buckets and the order list.
	// We don't use a RWMutex because even read operation
	// can do updates due to evict and updating usage order.
	mu sync.Mutex
	// buckets maps to key values where value is a pointer to an element in the order list.
	// The actual value is stored in the [list.Element] Value field.
	buckets map[string]map[string]*list.Element
	// order is by default the insertion order
	// If user use [EvictionPolicyLRU] or [EvictionPolicyMRU], then the order is also updated
	// during Get and Set.
	order *list.List
}

type cacheEntry struct {
	// Save the bucket key to use in eviction
	bucket     string
	key        string
	value      []byte
	expiration time.Time
}

func NewLRUCache(capacity int, ttlCheckInterval time.Duration) *LRUCache {
	c := &LRUCache{
		capacity:         capacity,
		ttlCheckInterval: ttlCheckInterval,
		stop:             make(chan struct{}),

		buckets: make(map[string]map[string]*list.Element),
		order:   list.New(),
	}
	c.startTTLCheck()
	return c
}

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
	entry := cacheEntry{bucket: bucket, key: key, value: value, expiration: expiration}

	// Check if the key already exists
	e, ok := b[key]
	if ok {
		e.Value = entry
		if opts.EvictionPolicy == EvictionPolicyLRU || opts.EvictionPolicy == EvictionPolicyMRU {
			c.order.MoveToBack(e)
		}
		return nil
	}

	// Evict before inserting new key
	size := c.order.Len()
	if size >= c.capacity {
		c.evict(opts.EvictionPolicy)
	}

	// Add new key to the bucket
	e = c.order.PushBack(entry)
	b[key] = e

	return nil
}

func (c *LRUCache) Get(bucket string, key string, opts Options) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Per requirement, use Oldest eviction policy on Get.
	// TODO: this requirement is quite confusing, why not use
	// the policy provided in the options?
	size := c.order.Len()
	if size >= c.capacity {
		c.evict(EvictionPolicyNone)
	}

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
	// Lazy TTL
	if !entry.expiration.IsZero() && entry.expiration.Before(time.Now()) {
		c.del(e)
		return nil, fmt.Errorf("key %s expired", key)
	}

	// Update order for LRU and MRU
	if opts.EvictionPolicy == EvictionPolicyLRU || opts.EvictionPolicy == EvictionPolicyMRU {
		c.order.MoveToBack(e)
	}

	return entry.value, nil
}

// Delete key from the cache, empty bucket is also removed.
func (c *LRUCache) Delete(bucket string, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Error if not exists
	b, ok := c.buckets[bucket]
	if !ok {
		return fmt.Errorf("bucket %s not found", bucket)
	}

	_, ok = b[key]
	if !ok {
		return fmt.Errorf("key %s not found", key)
	}

	// Delete if exists
	c.del(b[key])
	return nil
}

// Stop the background TTL check (if any).
// NOTE: Even if you stop the check in the background
// [Get] still checks the TTL.
func (c *LRUCache) Stop() {
	close(c.stop)
}

// Called by Set and Get when capacity is reached
func (c *LRUCache) evict(policy EvictionPolicy) {
	// No need to lock, caller already holds the lock

	var e *list.Element
	switch policy {
	case EvictionPolicyMRU, EvictionPolicyNewest:
		e = c.order.Back()
	default:
		// LRU, None, Oldest
		e = c.order.Front()
	}

	c.del(e)
}

// Shared by evict and Delete.
// NOTE: caller must hold the write lock.
func (c *LRUCache) del(e *list.Element) {
	c.order.Remove(e)

	entry := e.Value.(cacheEntry)
	b := c.buckets[entry.bucket]
	delete(b, entry.key)
	if len(b) == 0 {
		delete(c.buckets, entry.bucket)
	}
}

func (c *LRUCache) startTTLCheck() {
	if c.ttlCheckInterval <= 0 {
		return
	}

	ticker := time.NewTicker(c.ttlCheckInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.checkExpired()
			case <-c.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *LRUCache) checkExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, b := range c.buckets {
		for _, e := range b {
			entry := e.Value.(cacheEntry)
			if !entry.expiration.IsZero() && entry.expiration.Before(time.Now()) {
				c.del(e)
			}
		}
	}
}
