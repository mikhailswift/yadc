package cache

import (
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	testCases := []struct {
		Key            string
		Value          string
		TTL            time.Duration
		ExpectedAction action
	}{
		{"Test Key 1", "Test Value 1", 5 * time.Minute, Created},
		{"Test Key 2", "Test Value 2", 1 * time.Second, Created},
		{"Test Key 3", "Test Value 3", 0 * time.Second, Created},
		{"Test Key 1", "Test Value 4", 5 * time.Minute, Updated},
	}

	c := NewCache()
	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			if r := c.Set(tc.Key, tc.Value, tc.TTL); r.Action != tc.ExpectedAction {
				t.Fatalf("Got unexpected result back for key %v: %v", tc.Key, r)
			}
		})
	}
}

func TestUnset(t *testing.T) {
	testCases := []struct {
		Key            string
		Value          string
		TTL            time.Duration
		ExpectedAction action
	}{
		{"Test Key 1", "Test Value 1", 5 * time.Minute, Created},
		{"Test Key 2", "Test Value 2", 1 * time.Second, Created},
		{"Test Key 3", "Test Value 3", 0 * time.Second, Created},
	}

	c := NewCache()
	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			if r := c.Set(tc.Key, tc.Value, tc.TTL); r.Action != tc.ExpectedAction && r.Err != nil {
				t.Fatalf("Got unexpected result back for key %v: %v", tc.Key, r)
			}
		})
	}

	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			if r := c.Unset(tc.Key); r.Action != Deleted && r.Err != nil {
				t.Fatalf("Got unexpected result back for key %v: %v", tc.Key, r)
			}
		})
	}

	r := c.Unset("Garbage Key")
	if _, ok := r.Err.(ErrKeyNotFound); r.Action != Failed || !ok || r.Err == nil {
		t.Fatalf("Expected to get an action of Failed and a ErrKeyNotFound when unsetting a garbage key but got %v", r)
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		Key            string
		Value          string
		TTL            time.Duration
		ExpectedAction action
	}{
		{"Test Key 1", "Test Value 1", 5 * time.Minute, Created},
		{"Test Key 2", "Test Value 2", 1 * time.Second, Created},
		{"Test Key 3", "Test Value 3", 0 * time.Second, Created},
	}

	c := NewCache()
	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			if r := c.Set(tc.Key, tc.Value, tc.TTL); r.Action != tc.ExpectedAction {
				t.Fatalf("Got unexpected result back for key %v: %v", tc.Key, r)
			}
		})
	}

	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			r := c.Get(tc.Key)
			if r.Action != Retrieved || r.Err != nil {
				t.Fatalf("Got an unexpected action or error when getting key %v: %v", tc.Key, r)
			}

			if tc.Value != r.n.value {
				t.Fatalf("Got an unexpected value for key %v: Actual: %v Expected: %v", tc.Key, r.n.value, tc.Value)
			}
		})
	}
}

func TestSetTTL(t *testing.T) {
	testCases := []struct {
		Key            string
		Value          string
		TTL            time.Duration
		ExpectedAction action
	}{
		{"Test Key 1", "Test Value 1", 5 * time.Minute, Created},
		{"Test Key 2", "Test Value 2", 1 * time.Second, Created},
		{"Test Key 3", "Test Value 3", 0 * time.Second, Created},
	}

	c := NewCache()
	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			if r := c.Set(tc.Key, tc.Value, tc.TTL); r.Action != tc.ExpectedAction {
				t.Fatalf("Got unexpected result back for key %v: %v", tc.Key, r)
			}
		})
	}

	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			r := c.SetTTL(tc.Key, tc.TTL+1*time.Second)
			if r.Action != Updated || r.Err != nil {
				t.Fatalf("Couldn't set TTL for key %v: %v", tc.Key, r)
			}
		})
	}

	r := c.SetTTL("Garbage Key", 5*time.Minute)
	if _, ok := r.Err.(ErrKeyNotFound); r.Action != Failed || r.Err == nil || !ok {
		t.Fatalf("Didn't get ErrKeyNotFound or Failed action when setting TTL for garbage key: %v", r)
	}

	r = c.SetTTL(testCases[0].Key, -5*time.Minute)
	if _, ok := r.Err.(ErrInvalidTTL); r.Action != Failed || r.Err == nil || !ok {
		t.Fatalf("didn't get ErrInvalidTTL when setting negative TTL: %v", r)
	}
}

func TestCacheGetTTL(t *testing.T) {
	testCases := []struct {
		Key            string
		Value          string
		TTL            time.Duration
		ExpectedAction action
	}{
		{"Test Key 1", "Test Value 1", 5 * time.Minute, Created},
		{"Test Key 2", "Test Value 2", 1 * time.Second, Created},
		{"Test Key 3", "Test Value 3", 0 * time.Second, Created},
	}

	c := NewCache()
	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			if r := c.Set(tc.Key, tc.Value, tc.TTL); r.Action != tc.ExpectedAction {
				t.Fatalf("Got unexpected result back for key %v: %v", tc.Key, r)
			}
		})
	}

	for _, tc := range testCases {
		t.Run(tc.Key, func(t *testing.T) {
			ttl, err := c.GetTTL(tc.Key)
			if _, ok := err.(ErrTTLNotFound); tc.TTL == 0*time.Second && (err == nil || !ok) {
				t.Fatalf("Didn't get ErrTTLNotFound for key %v which had 0 TTL: Err: %+v, TTL: %s", tc.Key, err, ttl)
			} else if tc.TTL > 0*time.Second && (err != nil || ttl < tc.TTL-1000*time.Microsecond) {
				t.Fatalf("Got unexpected err or ttl for key %v: Err: %+v TTL: %s", tc.Key, err, ttl)
			}
		})
	}

	newTTL := testCases[0].TTL + 1*time.Hour
	r := c.SetTTL(testCases[0].Key, newTTL)
	if r.Action != Updated || r.Err != nil {
		t.Fatalf("Failed to update TTL for key %v: %v", testCases[0].Key, r)
	}

	ttl, err := c.GetTTL(testCases[0].Key)
	if err != nil || ttl < newTTL-1000*time.Microsecond {
		t.Fatalf("Got unexpected err or ttl for key %v: Err: %+v TTL: %s", testCases[0].Key, err, ttl)
	}
}
