package cache

import (
	"time"
)

// Cacher defines the functionality a Cache needs to implement
type Cacher interface {
	Set(key, value string, ttl time.Duration) Result
	Unset(key string) Result
	Get(key string) Result
	SetTTL(key string, ttl time.Duration) Result
	GetTTL(key string) (time.Duration, error)
}

type memCache struct {
	table       HashTable
	ttlRegistry *ttlRegistry
}

// NewCache returns a newly instantiated Cache that's ready to use
func NewCache() Cacher {
	table := newTable()
	ttlReg := newTTLRegistry(table)
	return &memCache{
		table:       table,
		ttlRegistry: ttlReg,
	}
}

//Set will attempt to set a key and value with a specified TTL.  If TTL is less than or equal to zero it will not set the TTL.
func (c *memCache) Set(key, value string, ttl time.Duration) Result {
	r := c.table.Set(key, value)
	if r.Err != nil {
		return r
	}

	if ttl > 0 {
		err := c.ttlRegistry.RegisterTTL(key, r.n.created, ttl)
		if err != nil {
			c.table.Unset(key)
			return Result{
				Action: Failed,
				Err:    err,
			}
		}
	}

	return r
}

//Unset will unset the provided key from the cache.
func (c *memCache) Unset(key string) Result {
	r := c.table.Unset(key)
	if r.Err != nil {
		return r
	}

	err := c.ttlRegistry.UnregisterTTL(key)
	if _, ok := err.(ErrKeyNotFound); !ok && err != nil {
		return Result{
			Err:    err,
			Action: Failed,
		}
	}

	return r
}

//Get will attempt to retrieve a specified key from the cache.
func (c *memCache) Get(key string) Result {
	return c.table.Get(key)
}

//SetTTL will set the TTL for a provided key.
func (c *memCache) SetTTL(key string, ttl time.Duration) Result {
	r := c.table.Get(key)
	if r.Err != nil {
		return r
	}

	err := c.ttlRegistry.RegisterTTL(key, r.n.created, ttl)
	if err != nil {
		return Result{
			Action: Failed,
			Err:    err,
		}
	}

	return Result{
		Action: Updated,
		n:      r.n,
	}
}

//GetTTL will return the TTL for a provided key.
func (c *memCache) GetTTL(key string) (time.Duration, error) {
	return c.ttlRegistry.GetTTL(key)
}
