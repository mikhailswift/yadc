package cache

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

const (
	NoTtl = time.Duration(0)
)

type ErrInvalidTtl time.Duration

func (e ErrInvalidTtl) Error() string {
	return fmt.Sprintf("Couldn't use %s as a ttl", e)
}

type ErrTtlNotFound string

func (e ErrTtlNotFound) Error() string {
	return fmt.Sprintf("Couldn't find a ttl for key %v", e)
}

type ttlInfo struct {
	key    string
	expire time.Time
	ix     int
}

type ttlRegistry struct {
	ttlByKey      map[string]*ttlInfo
	queue         ttlQueue
	table         HashTable
	nextTtlExpire *time.Timer
	sync.RWMutex
}

func newTtlRegistry(table HashTable) *ttlRegistry {
	reg := &ttlRegistry{
		ttlByKey: make(map[string]*ttlInfo),
		queue:    make(ttlQueue, 0),
		table:    table,
	}

	heap.Init(&reg.queue)
	return reg
}

func (reg *ttlRegistry) RegisterTtl(key string, created time.Time, ttl time.Duration) error {
	reg.Lock()
	defer reg.Unlock()

	var ti *ttlInfo
	var exists bool

	if ti, exists = reg.ttlByKey[key]; !exists {
		if ttl <= 0 {
			return ErrInvalidTtl(ttl)
		}

		ti = &ttlInfo{
			key:    key,
			expire: created.Add(ttl).UTC(),
		}
	} else if !exists && ttl <= 0 {
		// a value of zero or below will erase the ttl
		ti.expire = time.Time{}
	}

	// peek the next ttl, if it's after the one we're adding reset the timer to our newly added ttl
	for reg.queue.Len() > 0 {
		next := reg.queue[0]
		// pop any ttls that are no longer valid
		if next.expire.IsZero() {
			reg.queue.Pop()
			continue
		}

		if next.expire.After(ti.expire) {
			reg.nextTtlExpire.Stop()
			reg.nextTtlExpire = time.AfterFunc(ti.expire.Sub(time.Now().UTC()), reg.expireKeys)
		}

		break
	}

	if !exists {
		heap.Push(&reg.queue, ti)
		reg.ttlByKey[key] = ti
	} else {
		heap.Fix(&reg.queue, ti.ix)
	}

	return nil
}

func (reg *ttlRegistry) GetTtl(key string) (time.Duration, error) {
	reg.RLock()
	defer reg.RUnlock()
	ti, ok := reg.ttlByKey[key]
	if !ok {
		return time.Duration(0), ErrTtlNotFound(key)
	}

	return ti.expire.Sub(time.Now().UTC()), nil
}

func (reg *ttlRegistry) UnregisterTtl(key string) error {
	reg.Lock()
	defer reg.Unlock()

	ti, exists := reg.ttlByKey[key]
	if !exists {
		return ErrKeyNotFound(key)
	}

	ti.expire = time.Time{}
	return nil
}

func (reg *ttlRegistry) expireKeys() {
	reg.Lock()
	defer reg.Unlock()
	now := time.Now().UTC()

	for reg.queue.Len() > 0 {
		// peek the next to make sure we should expire
		next := reg.queue[0]

		// disregard ttls with zero time
		if next.expire.IsZero() {
			reg.queue.Pop()
			continue
		}

		if next.expire.After(now) {
			reg.nextTtlExpire.Stop()
			reg.nextTtlExpire = time.AfterFunc(next.expire.Sub(now), reg.expireKeys)
			return
		}

		reg.queue.Pop()
		delete(reg.ttlByKey, next.key)
	}
}

type ttlQueue []*ttlInfo

func (q ttlQueue) Len() int { return len(q) }

func (q ttlQueue) Less(i, j int) bool {
	// Pop needs to send the next ttl that's about to expire, so this needs return the ttl that happens first
	return q[i].expire.Before(q[j].expire)
}

func (q ttlQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].ix = i
	q[j].ix = j
}

func (q *ttlQueue) Push(x interface{}) {
	ti := x.(*ttlInfo)
	ti.ix = len(*q)
	*q = append(*q, ti)
}

func (q *ttlQueue) Pop() interface{} {
	old := *q
	n := len(old)
	next := old[n-1]
	*q = old[:n-1]
	next.ix = -1
	return next
}
