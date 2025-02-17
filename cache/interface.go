package cache

import "time"

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

// TODO: the interface seems to be wrong, should be func(o *Options) error or return a new Options to allow modifying the default option
// type Option func(o Options) error
// type Option func(*Options) error

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
