package cache

import (
	"sync"
	"time"
)

type HashTable interface {
	Set(key, value string) result
	Unset(key string) result
	Get(key string) result
}

type mapHashTable struct {
	m map[string]*node
	sync.RWMutex
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
