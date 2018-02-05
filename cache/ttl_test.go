package cache

import (
	"container/heap"
	"testing"
	"time"
)

func getTestRegistry() *ttlRegistry {
	table := newTable()
	return newTtlRegistry(table)
}

func TestTtlRegistrationAndHeapTest(t *testing.T) {
	testCases := []struct {
		Key string
		Ttl time.Duration
	}{
		{
			Key: "Test",
			Ttl: 5 * time.Second,
		}, {
			Key: "Test2",
			Ttl: 6 * time.Second,
		}, {
			Key: "Test3",
			Ttl: 4 * time.Second,
		},
	}

	reg := getTestRegistry()
	now := time.Now().UTC()

	if err := reg.RegisterTtl(testCases[0].Key, now, testCases[0].Ttl); err != nil {
		t.Fatalf("Experienced error when registering ttl: %+v", err)
	}

	if err := reg.RegisterTtl(testCases[1].Key, now, testCases[1].Ttl); err != nil {
		t.Fatalf("Experienced error when registering ttl: %+v", err)
	}

	// Check the registry's hash and map to ensure ordering is correct
	ti1, ok := reg.ttlByKey[testCases[0].Key]
	if !ok {
		t.Fatalf("Couldn't find ttl for key %v in ttlByKey map", testCases[0].Key)
	}

	ti2, ok := reg.ttlByKey[testCases[1].Key]
	if !ok {
		t.Fatalf("Couldn't find ttl for key %v in ttlByKey map", testCases[0].Key)
	}

	if ti1.expire.After(ti2.expire) {
		t.Fatalf("Test case 1's expire time was after test case 2's expire time!")
	}

	// Check the queue to make sure it's in the order I expect. I expect the first key
	if reg.queue[0].key != testCases[0].Key {
		t.Fatalf("Queue has %v as first key, expected %v", reg.queue[0].key, testCases[0].Key)
	}

	// Add the third key with an earlier time to ensure queue gets adjusted
	if err := reg.RegisterTtl(testCases[2].Key, now, testCases[2].Ttl); err != nil {
		t.Fatalf("Experienced error when registering ttl: %+v", err)
	}

	// Now the queue should have testCase[2] as the first element
	if reg.queue[0].key != testCases[2].Key {
		t.Fatalf("Queue has %v as first key, expected %v", reg.queue[0].key, testCases[2].Key)
	}

	// Now let's ensure Pop gives us the right entry back out
	popped := heap.Pop(&reg.queue).(*ttlInfo)
	if popped.key != testCases[2].Key {
		t.Fatalf("Got ttl info with key %v after popping but expected %v", popped.key, testCases[2].Key)
	}

	// Next popped should be test case 0
	popped = heap.Pop(&reg.queue).(*ttlInfo)
	if popped.key != testCases[0].Key {
		t.Fatalf("Got ttl info with key %v after popping but expected %v", popped.key, testCases[0].Key)
	}

	// Lastly we should pop test case 1
	popped = heap.Pop(&reg.queue).(*ttlInfo)
	if popped.key != testCases[1].Key {
		t.Fatalf("Got ttl info with key %v after popping but expected %v", popped.key, testCases[1].Key)
	}
}

func TestExpiration(t *testing.T) {
	testCases := []*struct {
		Key    string
		Ttl    time.Duration
		Expire time.Time
	}{
		{
			Key: "Test",
			Ttl: 1 * time.Second,
		}, {
			Key: "Test2",
			Ttl: 2 * time.Second,
		}, {
			Key: "Test3",
			Ttl: 500 * time.Millisecond,
		},
	}

	reg := getTestRegistry()
	now := time.Now().UTC()

	for _, ti := range testCases {
		t.Run(ti.Key, func(t *testing.T) {
			reg.table.Set(ti.Key, "Garbage")
			if err := reg.RegisterTtl(ti.Key, now, ti.Ttl); err != nil {
				t.Fatalf("Experiences error when registering ttl: %+v", err)
			}
			ti.Expire = now.Add(ti.Ttl)
		})
	}

	// wait for the first one to expire
	time.Sleep(testCases[2].Ttl + 250*time.Millisecond)

	ti, ok := reg.ttlByKey[testCases[2].Key]
	if ok {
		t.Fatalf("Found ttl info for key that should have been expired: %v", ti.key)
	}

	if reg.queue[0].key == testCases[2].Key {
		t.Fatalf("Found ttl infor in queue for key that should have been expired: %v", ti.key)
	}

	r := reg.table.Get(testCases[2].Key)
	if _, ok := r.Err.(ErrKeyNotFound); r.Err == nil || !ok {
		t.Fatalf("Expected to not find key in table for key that should have expired: Key: %v; Err: %+v", testCases[2].Key, r)
	}

	time.Sleep(1 * time.Second)

	ti, ok = reg.ttlByKey[testCases[0].Key]
	if ok {
		t.Fatalf("Found ttl info for key that should have been expired: %v", ti.key)
	}

	if reg.queue[0].key == testCases[0].Key {
		t.Fatalf("Found ttl infor in queue for key that should have been expired: %v", ti.key)
	}

	r = reg.table.Get(testCases[0].Key)
	if _, ok := r.Err.(ErrKeyNotFound); r.Err == nil || !ok {
		t.Fatalf("Expected to not find key in table for key that should have expired: Key: %v; Err: %+v", testCases[0].Key, r)
	}

	time.Sleep(1 * time.Second)

	ti, ok = reg.ttlByKey[testCases[1].Key]
	if ok {
		t.Fatalf("Found ttl info for key that should have been expired: %v", ti.key)
	}

	if reg.queue.Len() > 0 && reg.queue[0].key == testCases[1].Key {
		t.Fatalf("Found ttl infor in queue for key that should have been expired: %v", ti.key)
	}

	r = reg.table.Get(testCases[1].Key)
	if _, ok := r.Err.(ErrKeyNotFound); r.Err == nil || !ok {
		t.Fatalf("Expected to not find key in table for key that should have expired: Key: %v; Err: %+v", testCases[1].Key, r)
	}

}
