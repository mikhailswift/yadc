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
		Key         string
		Value       string
		TTL         time.Duration
		Expire      time.Time
		Expired     bool
		NeverExpire bool
	}{
		{
			Key:         "Test",
			Value:       "Garbage",
			TTL:         1 * time.Second,
			Expired:     false,
			NeverExpire: false,
		}, {
			Key:         "Test2",
			Value:       "Garbage",
			TTL:         2 * time.Second,
			Expired:     false,
			NeverExpire: false,
		}, {
			Key:         "Test3",
			Value:       "Garbage",
			TTL:         500 * time.Millisecond,
			Expired:     false,
			NeverExpire: false,
		}, {
			Key:         "Unexpired Test",
			Value:       "Garbage",
			TTL:         1 * time.Minute,
			Expired:     false,
			NeverExpire: true,
		},
	}

	reg := getTestRegistry()
	now := time.Now().UTC()

	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			reg.table.Set(tc.Key, "Garbage")
			if err := reg.RegisterTTL(tc.Key, now, tc.TTL); err != nil {
				t.Fatalf("Experiences error when registering ttl: %+v", err)
			}
			tc.Expire = now.Add(tc.TTL)

			if tc.NeverExpire {
				if err := reg.UnregisterTTL(tc.Key); err != nil {
					t.Fatalf("Failed to unregister key that shouldn't expire: Key: %v Err: %+v", tc.Key, err)
				}
			}
		})
	}

	for _ = range time.Tick(750 * time.Millisecond) {
		finished := true
		now := time.Now().UTC()
		for _, tc := range testCases {
			reg.RLock()
			r := reg.table.Get(tc.Key)
			_, isKeyNotFoundErr := r.Err.(ErrKeyNotFound)
			ti, tiExists := reg.ttlByKey[tc.Key]

			// if this key shouldn't expire it shouldn't be in the ttlByKey since it'd have been unregistered
			if tc.NeverExpire && (tiExists || r.Action != Retrieved || r.GetValue() != tc.Value || r.GetKey() != tc.Key) {
				t.Fatalf("Key that shouldn't have expired had something bad happen: Key: %v ; Result: %v ; ttl info: %v", tc.Key, r, ti)
			}

			// if this key has expired in the last iteration mark it as such
			if tc.Expire.Before(now) {
				tc.Expired = true
			}

			if tc.Expired && (r.Action != Failed || !isKeyNotFoundErr || r.Err == nil || tiExists) {
				t.Fatalf("Key that should have expired was still found: Key %v ; Result: %v; ttl info: %v", tc.Key, r, ti)
			}

			// if we still have a case that hasn't expired yet set finished to false
			if !tc.NeverExpire && !tc.Expired {
				finished = false
			}
			reg.RUnlock()
		}

		if finished {
			break
		}
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
