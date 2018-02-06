package cache

import (
	"fmt"
	"sync"
	"time"
)

//HashTable is the interface that defines what a table needs to do to be used as a hash table in the Cacher interface
type HashTable interface {
	Set(key, value string) result
	Unset(key string) result
	Get(key string) result
}

//ErrKeyNotFound is returned when a requested key could not be found in the table
type ErrKeyNotFound string

func (e ErrKeyNotFound) Error() string {
	return fmt.Sprintf("Could not find key: %+v", e)
}

type mapHashTable struct {
	m map[string]*node
	sync.RWMutex
}

func newTable() HashTable {
	return &mapHashTable{
		m: make(map[string]*node),
	}
}

func (t *mapHashTable) Set(key, value string) result {
	t.Lock()
	defer t.Unlock()

	n, ok := t.m[key]
	if ok {
		n.value = value
		n.created = time.Now().UTC()
		return result{
			Node:   *n,
			Action: Updated,
		}
	}

	n = newRecord(key, value)
	t.m[key] = n
	return result{
		Node:   *n,
		Action: Created,
	}
}

func (t *mapHashTable) Unset(key string) result {
	t.Lock()
	defer t.Unlock()
	n, exists := t.m[key]
	if !exists {
		return result{
			Action: Failed,
			Err:    ErrKeyNotFound(key),
		}
	}

	delete(t.m, key)
	return result{
		Action: Deleted,
		Node:   *n,
	}
}

func (t *mapHashTable) Get(key string) result {
	t.Lock()
	defer t.Unlock()

	n, exists := t.m[key]
	if !exists {
		return result{
			Action: Failed,
			Err:    ErrKeyNotFound(key),
		}
	}

	return result{
		Action: Retrieved,
		Node:   *n,
	}
}
