package cache

import (
	"container/heap"
	"fmt"
	"log"
	"sync"
	"time"
)

//ErrInvalidTTL indicates the registry got a bad duration as a TTL
type ErrInvalidTTL time.Duration

func (e ErrInvalidTTL) Error() string {
	return fmt.Sprintf("Couldn't use %s as a ttl", time.Duration(e))
}

//ErrTTLNotFound indicates the registry could not find a TTL for the provided key.
type ErrTTLNotFound string

func (e ErrTTLNotFound) Error() string {
	return fmt.Sprintf("Couldn't find a ttl for key %v", string(e))
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
	nextTTLExpire *time.Timer
	sync.RWMutex
}

func newTTLRegistry(table HashTable) *ttlRegistry {
	reg := &ttlRegistry{
		ttlByKey: make(map[string]*ttlInfo),
		queue:    make(ttlQueue, 0),
		table:    table,
	}

	heap.Init(&reg.queue)
	return reg
}

func (reg *ttlRegistry) RegisterTTL(key string, created time.Time, ttl time.Duration) error {
	if ttl <= 0 {
		return ErrInvalidTTL(ttl)
	}

	reg.Lock()
	defer reg.Unlock()
	var ti *ttlInfo
	var exists bool

	if ti, exists = reg.ttlByKey[key]; !exists {
		ti = &ttlInfo{
			key:    key,
			expire: created.Add(ttl).UTC(),
		}
	} else if exists {
		ti.expire = created.Add(ttl).UTC()
	}

	// peek the next ttl, if it's after the one we're adding reset the timer to our newly added ttl
	for reg.queue.Len() > 0 {
		next := reg.queue[0]
		if next.expire.After(ti.expire) {
			if reg.nextTTLExpire != nil {
				reg.nextTTLExpire.Stop()
			}
			reg.nextTTLExpire = time.AfterFunc(ti.expire.Sub(time.Now().UTC()), reg.expireKeys)
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

func (reg *ttlRegistry) GetTTL(key string) (time.Duration, error) {
	reg.RLock()
	defer reg.RUnlock()
	ti, ok := reg.ttlByKey[key]
	if !ok {
		return time.Duration(0), ErrTTLNotFound(key)
	}

	return ti.expire.Sub(time.Now().UTC()), nil
}

func (reg *ttlRegistry) UnregisterTTL(key string) error {
	reg.Lock()
	defer reg.Unlock()

	ti, exists := reg.ttlByKey[key]
	if !exists {
		return ErrKeyNotFound(key)
	}

	ti.expire = time.Time{}
	heap.Remove(&reg.queue, ti.ix)
	delete(reg.ttlByKey, key)
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
			heap.Pop(&reg.queue)
			continue
		}

		if next.expire.After(now) {
			if reg.nextTTLExpire != nil {
				reg.nextTTLExpire.Stop()
			}
			reg.nextTTLExpire = time.AfterFunc(next.expire.Sub(now), reg.expireKeys)
			return
		}

		r := reg.table.Unset(next.key)
		if _, ok := r.Err.(ErrKeyNotFound); r.Err != nil && !ok {
			log.Printf("Couldn't unset key while expiring key %v: %+v", next.key, r.Err)
		}

		heap.Pop(&reg.queue)
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
