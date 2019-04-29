package tsm1

import "testing"

func TestCacheKeyIterator_Single(t *testing.T) {
	v0 := NewValue(1, 1.0)

	writes := map[string][]Value{
		"cpu,host=A#!~#value": {v0},
	}

	c := NewCache(0)

	for k, v := range writes {
		if err := c.Write([]byte(k), v); err != nil {
			t.Fatalf("failed to write key foo to cache: %s", err.Error())
		}
	}

	iter := newCacheKeyIterator(c, 1, nil)
	var readValues bool
	for iter.Next() {
		key, _, _, block, err := iter.Read()
		if err != nil {
			t.Fatalf("unexpected error read: %v", err)
		}

		values, err := DecodeBlock(block, nil)
		if err != nil {
			t.Fatalf("unexpected error decode: %v", err)
		}

		if got, exp := string(key), "cpu,host=A#!~#value"; got != exp {
			t.Fatalf("key mismatch: got %v, exp %v", got, exp)
		}

		if got, exp := len(values), len(writes); got != exp {
			t.Fatalf("values length mismatch: got %v, exp %v", got, exp)
		}

		for _, v := range values {
			readValues = true
			assertValueEqual(t, v, v0)
		}
	}

	if !readValues {
		t.Fatalf("failed to read any values")
	}
}

func TestCacheKeyIterator_Chunked(t *testing.T) {
	v0 := NewValue(1, 1.0)
	v1 := NewValue(2, 2.0)

	writes := map[string][]Value{
		"cpu,host=A#!~#value": {v0, v1},
	}

	c := NewCache(0)

	for k, v := range writes {
		if err := c.Write([]byte(k), v); err != nil {
			t.Fatalf("failed to write key foo to cache: %s", err.Error())
		}
	}

	iter := newCacheKeyIterator(c, 1, nil)
	var readValues bool
	var chunk int
	for iter.Next() {
		key, _, _, block, err := iter.Read()
		if err != nil {
			t.Fatalf("unexpected error read: %v", err)
		}

		values, err := DecodeBlock(block, nil)
		if err != nil {
			t.Fatalf("unexpected error decode: %v", err)
		}

		if got, exp := string(key), "cpu,host=A#!~#value"; got != exp {
			t.Fatalf("key mismatch: got %v, exp %v", got, exp)
		}

		if got, exp := len(values), 1; got != exp {
			t.Fatalf("values length mismatch: got %v, exp %v", got, exp)
		}

		for _, v := range values {
			readValues = true
			assertValueEqual(t, v, writes["cpu,host=A#!~#value"][chunk])
		}
		chunk++
	}

	if !readValues {
		t.Fatalf("failed to read any values")
	}
}

// Tests that the CacheKeyIterator will abort if the interrupt channel is closed
func TestCacheKeyIterator_Abort(t *testing.T) {
	v0 := NewValue(1, 1.0)

	writes := map[string][]Value{
		"cpu,host=A#!~#value": {v0},
	}

	c := NewCache(0)

	for k, v := range writes {
		if err := c.Write([]byte(k), v); err != nil {
			t.Fatalf("failed to write key foo to cache: %s", err.Error())
		}
	}

	intC := make(chan struct{})

	iter := newCacheKeyIterator(c, 1, intC)

	var aborted bool
	for iter.Next() {
		//Abort
		close(intC)

		_, _, _, _, err := iter.Read()
		if err == nil {
			t.Fatalf("unexpected error read: %v", err)
		}
		aborted = err != nil
	}

	if !aborted {
		t.Fatalf("iteration not aborted")
	}
}

func assertValueEqual(t *testing.T, a, b Value) {
	if got, exp := a.UnixNano(), b.UnixNano(); got != exp {
		t.Fatalf("time mismatch: got %v, exp %v", got, exp)
	}
	if got, exp := a.Value(), b.Value(); got != exp {
		t.Fatalf("value mismatch: got %v, exp %v", got, exp)
	}
}
