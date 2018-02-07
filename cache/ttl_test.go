package cache

import (
	"container/heap"
	"testing"
	"time"
)

func getTestRegistry() *ttlRegistry {
	table := newTable()
	return newTTLRegistry(table)
}

func TestTTLRegistrationAndHeapTest(t *testing.T) {
	testCases := []struct {
		Key string
		TTL time.Duration
	}{
		{
			Key: "Test",
			TTL: 5 * time.Second,
		}, {
			Key: "Test2",
			TTL: 6 * time.Second,
		}, {
			Key: "Test3",
			TTL: 4 * time.Second,
		},
	}

	reg := getTestRegistry()
	now := time.Now().UTC()

	if err := reg.RegisterTTL(testCases[0].Key, now, testCases[0].TTL); err != nil {
		t.Fatalf("Experienced error when registering ttl: %+v", err)
	}

	if err := reg.RegisterTTL(testCases[1].Key, now, testCases[1].TTL); err != nil {
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
	if err := reg.RegisterTTL(testCases[2].Key, now, testCases[2].TTL); err != nil {
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
		Value  string
		TTL    time.Duration
		Expire time.Time
	}{
		{
			Key:   "Test",
			Value: "Garbage",
			TTL:   1 * time.Second,
		}, {
			Key:   "Test2",
			Value: "Garbage",
			TTL:   2 * time.Second,
		}, {
			Key:   "Test3",
			Value: "Garbage",
			TTL:   500 * time.Millisecond,
		}, {
			Key:   "Unexpired Test",
			Value: "Garbage",
			TTL:   1 * time.Minute,
		},
	}

	reg := getTestRegistry()
	now := time.Now().UTC()

	for _, ti := range testCases {
		t.Run(ti.Key, func(t *testing.T) {
			reg.table.Set(ti.Key, "Garbage")
			if err := reg.RegisterTTL(ti.Key, now, ti.TTL); err != nil {
				t.Fatalf("Experiences error when registering ttl: %+v", err)
			}
			ti.Expire = now.Add(ti.TTL)
		})
	}

	// wait for the first one to expire
	time.Sleep(testCases[2].TTL + 250*time.Millisecond)

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

	// Expire test case 4
	if err := reg.UnregisterTTL(testCases[3].Key); err != nil {
		t.Fatalf("Couldn't unregister key %v: %+v", testCases[3].Key, err)
	}

	ti, ok = reg.ttlByKey[testCases[3].Key]
	if ok {
		t.Fatalf("Found ttl info for key that should have been expired: %v", ti.key)
	}

	for _, ti := range reg.queue {
		if ti.key == testCases[3].Key {
			t.Fatalf("Found unregistered key in ttl queue: %+v", ti.key)
		}
	}

	time.Sleep(1 * time.Second)

	if r := reg.table.Get(testCases[3].Key); r.Action != Retrieved || r.GetValue() != testCases[3].Value {
		t.Fatalf("Couldn't retrieve value from table for key who had TTL unregistered: %v", r)
	}

	ti, ok = reg.ttlByKey[testCases[0].Key]
	if ok {
		t.Fatalf("Found ttl info for key that should have been expired: %v", ti.key)
	}

	if reg.queue[0].key == testCases[0].Key {
		t.Fatalf("Found ttl info in queue for key that should have been expired: %v", ti.key)
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

func TestGetTTL(t *testing.T) {
	testCases := []struct {
		Key string
		TTL time.Duration
	}{
		{
			Key: "Test",
			TTL: 5 * time.Second,
		}, {
			Key: "Test2",
			TTL: 6 * time.Second,
		}, {
			Key: "Test3",
			TTL: 4 * time.Second,
		},
	}

	reg := getTestRegistry()
	now := time.Now().UTC()
	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			if err := reg.RegisterTTL(tc.Key, now, tc.TTL); err != nil {
				t.Fatalf("Couldn't set TTL for key %v: %+v", tc.Key, err)
			}
		})
	}

	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			ttl, err := reg.GetTTL(tc.Key)
			if err != nil {
				t.Fatalf("Couldn't get TTL for key %v: %+v", tc.Key, err)
			}

			// some time may have dropped off the TTL since we added it (should be close to negligable)
			if ttl > tc.TTL {
				t.Fatalf("Got an unexpected TTL for key %v, got %s, expected %s", tc.Key, ttl, tc.TTL)
			}
		})
	}

	// Ensure we get an error when finding a TTL for a key that isn't registered
	ttl, err := reg.GetTTL("Garbage Key")
	if _, ok := err.(ErrTTLNotFound); err == nil || !ok {
		t.Fatalf("Got an unexpected error when requesting a TTL for a key that shouldn't exist: TTL: %s Err: %+v", ttl, err)
	}
}

func TestUnregisterTTL(t *testing.T) {
	testCases := []struct {
		Key string
		TTL time.Duration
	}{
		{
			Key: "Test",
			TTL: 5 * time.Second,
		}, {
			Key: "Test2",
			TTL: 6 * time.Second,
		}, {
			Key: "Test3",
			TTL: 4 * time.Second,
		},
	}

	reg := getTestRegistry()
	now := time.Now().UTC()
	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			if err := reg.RegisterTTL(tc.Key, now, tc.TTL); err != nil {
				t.Fatalf("Couldn't register TTL for key %v: %+v", tc.Key, err)
			}
		})
	}

	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			if err := reg.UnregisterTTL(tc.Key); err != nil {
				t.Fatalf("Couldn't unregister TTL for key %v: %+v", tc.Key, err)
			}

			ttl, err := reg.GetTTL(tc.Key)
			if _, ok := err.(ErrTTLNotFound); err == nil || !ok {
				t.Fatalf("Didn't get ErrTTLNotFound after unregistering TTL for key %v: TTL: %v, Err: %+v", tc.Key, ttl, err)
			}
		})
	}

	// ttlByKey should be empty after all ttls have been unregistered
	if len(reg.ttlByKey) != 0 {
		t.Fatalf("Found unexpected TTLs in ttlByKey map: %v", reg.ttlByKey)
	}

	// All ttls in the queue should be set to zero time after unregister
	for _, ti := range reg.queue {
		if !ti.expire.IsZero() {
			t.Fatalf("Found TTLs with unexpected values in TTL registry queue: Key: %v; Expire: %s", ti.key, ti.expire)
		}
	}
}
