package lrucache

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

func BenchmarkCacheTTL_Add(b *testing.B) {
	cache, cancel := NewWithTTL(1000, time.Second)
	defer cancel()

	for i := 0; i < b.N; i++ {
		cache.Add(strconv.Itoa(i), i)
	}
}

func BenchmarkCacheTTL_Get(b *testing.B) {
	cache, cancel := NewWithTTL(1000, time.Second)
	defer cancel()

	for i := 0; i < 1000; i++ {
		cache.Add(strconv.Itoa(i), i)
	}

	for i := 0; i < b.N; i++ {
		cache.Get(strconv.Itoa(i))
	}
}

func Test_NewTTL(t *testing.T) {
	capacity := 5
	cache, cancel := NewWithTTL(capacity, time.Second)
	defer cancel()

	require.NotNil(t, cache)
	assert.NotNil(t, cache.data)
	assert.NotNil(t, cache.queue)
	assert.Equal(t, capacity, cache.cap)
}

func Test_NewTTL_AddWithExpiration(t *testing.T) {
	capacity := 4
	cache, cancel := NewWithTTL(capacity, time.Second)
	defer cancel()

	cache.Add("first", "value")
	cache.AddWithTTL("first", "another value", time.Second*3)
	cache.AddWithTTL("second", "another value", time.Second*1)
	cache.AddWithTTL("third", "another value", time.Second*6)
	cache.Add("forth", 4)

	assert.Equal(t, 4, cache.Len())
	assert.Equal(t, 4, cache.queue.Len())
	assert.Equal(t, 3, cache.expQueue.Len())

	time.Sleep(time.Second * 4)

	assert.Equal(t, 2, cache.Len())
	assert.Equal(t, 2, cache.queue.Len())
	assert.Equal(t, 1, cache.expQueue.Len())
}

func Test_CacheTTL_Cap(t *testing.T) {
	capacity := 5
	cache, cancel := NewWithTTL(capacity, time.Second)
	defer cancel()

	assert.Equal(t, capacity, cache.Cap())
}

func Test_CacheTTL_Len(t *testing.T) {
	capacity := 3
	cache, cancel := NewWithTTL(capacity, time.Second)
	defer cancel()

	cases := []struct {
		name        string
		key         string
		value       any
		expectedLen int
	}{
		{
			name:        "free space",
			key:         "first",
			value:       1,
			expectedLen: 1,
		},
		{
			name:        "free space",
			key:         "second",
			value:       2,
			expectedLen: 2,
		},
		{
			name:        "free space",
			key:         "third",
			value:       3,
			expectedLen: 3,
		},
		{
			name:        "no free space",
			key:         "forth",
			value:       4,
			expectedLen: 3,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache.Add(c.key, c.value)
			assert.Equal(t, c.expectedLen, cache.Len())
		})
	}
}

func Test_CacheTTL_Add(t *testing.T) {
	cache, cancel := NewWithTTL(3, time.Second)
	defer cancel()

	cases := []struct {
		name        string
		key         string
		value       any
		expectedLen int
	}{
		{
			name:        "free space",
			key:         "first",
			value:       1,
			expectedLen: 1,
		},
		{
			name:        "free space",
			key:         "second",
			value:       2,
			expectedLen: 2,
		},
		{
			name:        "free space",
			key:         "third",
			value:       3,
			expectedLen: 3,
		},
		{
			name:        "no free space",
			key:         "forth",
			value:       4,
			expectedLen: 3,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache.Add(c.key, c.value)
			assert.Equal(t, cache.queue.Len(), c.expectedLen)
			assert.Equal(t, len(cache.data), c.expectedLen)
		})
	}
}

func Test_CacheTTL_Update(t *testing.T) {
	cache, cancel := NewWithTTL(3, time.Second)
	defer cancel()

	cache.Add("key", "value")
	cache.Add("key", "another value")

	value, ok := cache.Get("key")
	assert.NotNil(t, value)
	assert.Equal(t, "another value", value)
	assert.Equal(t, true, ok)
	assert.Equal(t, 1, len(cache.data))
	assert.Equal(t, 1, cache.queue.Len())
}

func Test_CacheTTL_Clear(t *testing.T) {
	capacity := 5
	cache, cancel := NewWithTTL(capacity, time.Second)
	defer cancel()

	cache.Add("key", "value")

	cache.Clear()

	assert.NotNil(t, cache.data)
	assert.NotNil(t, cache.queue)
	assert.Equal(t, 0, len(cache.data))
	assert.Equal(t, 0, cache.queue.Len())
	assert.Equal(t, 0, cache.expQueue.Len())
}

func Test_CacheTTL_Get(t *testing.T) {
	capacity := 3
	cache, cancel := NewWithTTL(capacity, time.Second)
	defer cancel()

	cache.Add("first", 1)
	cache.Add("second", struct{ n int }{2})
	cache.Add("third", "three")
	cache.Add("forth", 4)

	cases := []struct {
		name          string
		key           string
		expectedValue any
		expectedOk    bool
	}{
		{
			name:          "value doesn't exist",
			key:           "random key",
			expectedValue: nil,
			expectedOk:    false,
		},
		{
			name:          "value doesn't exist",
			key:           "first",
			expectedValue: nil,
			expectedOk:    false,
		},
		{
			name:          "value exists",
			key:           "second",
			expectedValue: struct{ n int }{2},
			expectedOk:    true,
		},
		{
			name:          "value exists",
			key:           "third",
			expectedValue: "three",
			expectedOk:    true,
		},
		{
			name:          "value exists",
			key:           "forth",
			expectedValue: 4,
			expectedOk:    true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			value, ok := cache.Get(c.key)

			assert.Equal(t, c.expectedValue, value)
			assert.Equal(t, c.expectedOk, ok)
		})
	}
}

func Test_CacheTTL_Remove(t *testing.T) {
	capacity := 3
	cache, cancel := NewWithTTL(capacity, time.Second)
	defer cancel()

	cache.Add("first", 1)
	cache.Add("second", 2)
	cache.Add("third", 3)

	cases := []struct {
		name        string
		key         string
		expectedLen int
	}{
		{
			name:        "remove value",
			key:         "first",
			expectedLen: 2,
		},
		{
			name:        "remove value",
			key:         "second",
			expectedLen: 1,
		},
		{
			name:        "remove value",
			key:         "third",
			expectedLen: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache.Remove(c.key)

			assert.Equal(t, c.expectedLen, len(cache.data))
			assert.Equal(t, c.expectedLen, cache.queue.Len())
		})
	}
}
