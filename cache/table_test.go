package cache

import (
	"fmt"
	"testing"
)

func TestSetKey(t *testing.T) {
	testKey := "Test"
	expectVal := "Value"
	table := newTable()

	r := table.Set(testKey, expectVal)
	if r.Err != nil {
		t.Fatalf("Failed to set key to table: %+v", r.Err)
	}
}

func TestGetKey(t *testing.T) {
	testKey := "Test"
	expectedVal := "Value 123"
	table := newTable()

	r := table.Set(testKey, expectedVal)
	if r.Err != nil {
		t.Fatalf("Failed to set key: %+v", r.Err)
	}

	r = table.Get(testKey)
	if r.Err != nil {
		t.Fatalf("Failed to get key: %+v", r.Err)
	}

	if r.Node.value != expectedVal {
		t.Fatalf("Got unexpected value back from table for key %v: got %v, expected %v", r.Node.key, r.Node.value, expectedVal)
	}

	failKey := "Test 2"
	r = table.Get(failKey)
	if _, ok := r.Err.(ErrKeyNotFound); !ok {
		t.Fatalf("Got unexpected error when trying to get key that doesn't exist: %+v", r.Err)
	}
}

func TestUnsetKey(t *testing.T) {
	testKey := "Test"
	testVal := "Val"
	table := newTable()

	r := table.Set(testKey, testVal)
	if r.Err != nil {
		t.Fatalf("Failed to set key %v to table: %+v", testKey, r.Err)
	}

	r = table.Unset(testKey)
	if r.Err != nil {
		t.Fatalf("Failed to unset key %v: %+v", testKey, r.Err)
	}

	r = table.Get(testKey)
	if _, ok := r.Err.(ErrKeyNotFound); r.Err == nil || !ok {
		t.Fatalf("Expected to get ErrKeyNotFound, got: %+v", r)
	}
}

func BenchmarkSet(b *testing.B) {
	table := newTable()
	fmt.Println("benching")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := fmt.Sprintf("%v", i)
		r := table.Set(v, v)
		if r.Err != nil {
			b.Fatalf("Failed to set %v: %+v", v, r.Err)
		}
	}
}
