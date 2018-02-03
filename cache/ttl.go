package cache

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type ErrInvalidTtl time.Duration

func (e ErrInvalidTtl) Error() string {
	return fmt.Sprintf("Couldn't use %s as a ttl", e)
}

type ttlInfo struct {
	key    string
	expire time.Time
}

type ttlRegistry struct {
	ttlByKey      map[string]*ttlInfo
	queue         ttlQueue
	table         HashTable
	nextTtlExpire *time.Timer
	sync.RWMutex
}

func NewRegistry(table HashTable) *ttlRegistry {
	reg := &ttlRegistry{
		ttlByKey: make(map[string]*ttlInfo),
		queue:    make(ttlQueue, 0),
		table:    table,
	}

	heap.Init(&reg.queue)
	return reg
}

func (reg *ttlRegistry) RegisterTtl(key string, created time.Time, ttl time.Duration) error {
	if ttl <= 0 {
		return ErrInvalidTtl(ttl)
	}

	new := &ttlInfo{
		key:    key,
		expire: created.Add(ttl).UTC(),
	}

	reg.Lock()
	defer reg.Unlock()

	// peek the next ttl, if it's after the one we're adding reset the timer to our newly added ttl
	if reg.queue.Len() > 0 {
		next := reg.queue[0]
		if next.expire.After(new.expire) {
			reg.nextTtlExpire.Stop()
			reg.nextTtlExpire = time.AfterFunc(new.expire.Sub(time.Now().UTC()), reg.expireKeys)
		}
	}

	heap.Push(&reg.queue, new)
	reg.ttlByKey[key] = new
	return nil
}

func (reg *ttlRegistry) expireKeys() {
	reg.Lock()
	defer reg.Unlock()
	now := time.Now().UTC()

	for reg.queue.Len() > 0 {
		// peek the next to make sure we should expire
		next := reg.queue[0]
		if next.expire.After(now) {
			reg.nextTtlExpire.Stop()
			reg.nextTtlExpire = time.AfterFunc(next.expire.Sub(now), reg.expireKeys)
			return
		}

		reg.queue.Pop()
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
}

func (q *ttlQueue) Push(x interface{}) {
	*q = append(*q, x.(*ttlInfo))
}

func (q *ttlQueue) Pop() interface{} {
	old := *q
	n := len(old)
	next := old[n-1]
	*q = old[0 : n-1]
	return next
}
