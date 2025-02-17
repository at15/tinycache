package cache

import (
	"fmt"
	"net/http"
	"time"
)

type EvictionPolicy int

const (
	EvictionPolicyNone EvictionPolicy = iota
	EvictionPolicyOldest
	EvictionPolicyNewest
	EvictionPolicyLRU
	EvictionPolicyMRU
)

type Options struct {
	TTL            time.Duration
	EvictionPolicy EvictionPolicy
}

func ParseFromRequest(r *http.Request) (Options, error) {
	ttl := r.URL.Query().Get("ttl")
	if ttl == "" {
		ttl = "0"
	}
	// "300ms", "-1.5h" or "2h45m".
	ttlDuration, err := time.ParseDuration(ttl)
	if err != nil {
		return Options{}, err
	}
	if ttlDuration < 0 {
		return Options{}, fmt.Errorf("ttl cannot be negative: %s", ttl)
	}

	policy := r.URL.Query().Get("policy")
	if policy == "" {
		policy = "lru"
	}

	var evictionPolicy EvictionPolicy
	switch policy {
	case "lru":
		evictionPolicy = EvictionPolicyLRU
	case "mru":
		evictionPolicy = EvictionPolicyMRU
	case "oldest":
		evictionPolicy = EvictionPolicyOldest
	case "newest":
		evictionPolicy = EvictionPolicyNewest
	default:
		return Options{}, fmt.Errorf("invalid policy: %s", policy)
	}

	return Options{
		TTL:            ttlDuration,
		EvictionPolicy: evictionPolicy,
	}, nil
}

// TODO: the interface seems to be wrong, should be func(o *Options) error or return a new Options to allow modifying the default option
// type Option func(o Options) error
// type Option func(*Options) error

// Cache interface that only has one implementation ... [LRUCache]
type Cache interface {
	// TODO: options interface provided is likely wrong
	// TODO: different eviction policy for each bucket + key? shouldn't entire cache have the same eviction policy?
	// Set(bucket string, key string, value []byte, opts ...Option) error
	// Get(bucket string, key string, opts ...Option) ([]byte, error)
	// Delete(bucket string, key string, opts ...Option) error

	Set(bucket string, key string, value []byte, opts Options) error
	Get(bucket string, key string, opts Options) ([]byte, error)
	Delete(bucket string, key string) error
}
