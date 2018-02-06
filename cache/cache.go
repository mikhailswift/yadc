package cache

import (
	"time"
)

// Cacher defines the functionality a Cache needs to implement
type Cacher interface {
	Set(key, value string, ttl time.Duration) result
	Unset(key string) result
	Get(key string) result
	SetTTL(key string, ttl time.Duration) result
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

func (c *memCache) Set(key, value string, ttl time.Duration) result {
	r := c.table.Set(key, value)
	if r.Err != nil {
		return r
	}

	if ttl > 0 {
		err := c.ttlRegistry.RegisterTTL(key, r.Node.created, ttl)
		if err != nil {
			c.table.Unset(key)
			return result{
				Action: Failed,
				Err:    err,
			}
		}
	}

	return r
}

func (c *memCache) Unset(key string) result {
	r := c.table.Unset(key)
	if r.Err != nil {
		return r
	}

	err := c.ttlRegistry.UnregisterTTL(key)
	if _, ok := err.(ErrKeyNotFound); !ok && err != nil {
		return result{
			Err:    err,
			Action: Failed,
		}
	}

	return r
}

func (c *memCache) Get(key string) result {
	return c.table.Get(key)
}

func (c *memCache) SetTTL(key string, ttl time.Duration) result {
	r := c.table.Get(key)
	if r.Err != nil {
		return r
	}

	err := c.ttlRegistry.RegisterTTL(key, r.Node.created, ttl)
	if err != nil {
		return result{
			Action: Failed,
			Err:    err,
		}
	}

	return result{
		Action: Updated,
		Node:   r.Node,
	}
}

func (c *memCache) GetTTL(key string) (time.Duration, error) {
	return c.ttlRegistry.GetTTL(key)
}
