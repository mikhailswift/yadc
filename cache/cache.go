package cache

import (
	"sync"
	"time"
)

// Cacher defines the functionality a Cache needs to implement
type Cacher interface {
	Set(key, value string, ttl time.Duration) Result
	Unset(key string) Result
	Get(key string) Result
	SetTTL(key string, ttl time.Duration) Result
	GetTTL(key string) Result
	UnsetAll() Result
}

type memCache struct {
	table       HashTable
	ttlRegistry *ttlRegistry
	sync.RWMutex
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
	c.Lock()
	defer c.Unlock()
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
	c.Lock()
	defer c.Unlock()
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
	c.RLock()
	defer c.RUnlock()
	return c.table.Get(key)
}

//SetTTL will set the TTL for a provided key.
func (c *memCache) SetTTL(key string, ttl time.Duration) Result {
	c.Lock()
	defer c.Unlock()
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
func (c *memCache) GetTTL(key string) Result {
	c.RLock()
	defer c.RUnlock()
	ttl, err := c.ttlRegistry.GetTTL(key)
	if err != nil {
		return Result{
			Action: Failed,
			Err:    err,
		}
	}

	return Result{
		Action: RetrievedTTL,
		ttl:    ttl,
	}
}

//UnsetAll will unset all keys in the cache
func (c *memCache) UnsetAll() Result {
	c.Lock()
	defer c.Unlock()
	c.table.Clear()
	c.ttlRegistry.Reset()
	return Result{
		Action: Cleared,
	}
}
