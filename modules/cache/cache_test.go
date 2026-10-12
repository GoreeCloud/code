// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package cache

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"

	"code.forgejo.org/go-chi/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestCache(t *testing.T) {
	var err error
	var testCache cache.Cache

	testRedisHost := os.Getenv("TEST_REDIS_SERVER")
	if testRedisHost == "" {
		testCache, err = newCache(setting.Cache{
			Adapter: "memory",
			TTL:     time.Minute,
		})
	} else {
		testCache, err = newCache(setting.Cache{
			Adapter: "redis",
			Conn:    fmt.Sprintf("redis://%s", testRedisHost),
			TTL:     time.Minute,
		})
	}
	require.NoError(t, err)
	require.NotNil(t, testCache)

	t.Cleanup(test.MockVariableValue(&conn, testCache))
	t.Cleanup(test.MockVariableValue(&setting.CacheService.TTL, 24*time.Hour))
}

func TestNewContext(t *testing.T) {
	require.NoError(t, Init())

	setting.CacheService.Cache = setting.Cache{Adapter: "redis", Conn: "some random string"}
	con, err := newCache(setting.Cache{
		Adapter:  "rand",
		Conn:     "false conf",
		Interval: 100,
	})
	require.Error(t, err)
	assert.Nil(t, con)
}

func TestGetCache(t *testing.T) {
	createTestCache(t)

	assert.NotNil(t, GetCache())
}

func TestGetString(t *testing.T) {
	createTestCache(t)

	data, err := GetString("key", func() (string, error) {
		return "", errors.New("some error")
	})
	require.Error(t, err)
	assert.Empty(t, data)

	data, err = GetString("key", func() (string, error) {
		return "", nil
	})
	require.NoError(t, err)
	assert.Empty(t, data)

	data, err = GetString("key", func() (string, error) {
		return "some data", nil
	})
	require.NoError(t, err)
	assert.Empty(t, data)
	Remove("key")

	data, err = GetString("key", func() (string, error) {
		return "some data", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "some data", data)

	data, err = GetString("key", func() (string, error) {
		return "", errors.New("some error")
	})
	require.NoError(t, err)
	assert.Equal(t, "some data", data)
	Remove("key")
}

func TestGetInt(t *testing.T) {
	createTestCache(t)

	data, err := GetInt("key", func() (int, error) {
		return 0, errors.New("some error")
	})
	require.Error(t, err)
	assert.Equal(t, 0, data)

	data, err = GetInt("key", func() (int, error) {
		return 0, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 0, data)

	data, err = GetInt("key", func() (int, error) {
		return 100, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 0, data)
	Remove("key")

	data, err = GetInt("key", func() (int, error) {
		return 100, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 100, data)

	data, err = GetInt("key", func() (int, error) {
		return 0, errors.New("some error")
	})
	require.NoError(t, err)
	assert.Equal(t, 100, data)
	Remove("key")
}

func TestGetInt64(t *testing.T) {
	createTestCache(t)

	data, err := GetInt64("key", func() (int64, error) {
		return 0, errors.New("some error")
	})
	require.Error(t, err)
	assert.EqualValues(t, 0, data)

	data, err = GetInt64("key", func() (int64, error) {
		return 0, nil
	})
	require.NoError(t, err)
	assert.EqualValues(t, 0, data)

	data, err = GetInt64("key", func() (int64, error) {
		return 100, nil
	})
	require.NoError(t, err)
	assert.EqualValues(t, 0, data)
	Remove("key")

	data, err = GetInt64("key", func() (int64, error) {
		return 100, nil
	})
	require.NoError(t, err)
	assert.EqualValues(t, 100, data)

	data, err = GetInt64("key", func() (int64, error) {
		return 0, errors.New("some error")
	})
	require.NoError(t, err)
	assert.EqualValues(t, 100, data)
	Remove("key")
}

func TestCacheConcurrencySafety(t *testing.T) {
	createTestCache(t)

	testRedisHost := os.Getenv("TEST_REDIS_SERVER")
	if testRedisHost == "" {
		t.Skip("only relevant for a remote redis host")
	}

	numTests := 20
	numIncrements := 1000
	for testCount := range numTests {
		t.Run(fmt.Sprintf("attempt:%d", testCount), func(t *testing.T) {
			var counter atomic.Int64
			var wg sync.WaitGroup
			var firstError atomic.Value

			getFunc := func() (int64, error) {
				lastValue := counter.Load()
				time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
				return lastValue, nil
			}

			for range numIncrements {
				wg.Go(func() {
					counterValue := counter.Add(1)
					Remove(t.Name())
					cachedValue, err := GetInt64(t.Name(), getFunc)
					if err != nil {
						firstError.CompareAndSwap(nil, fmt.Sprintf("cache error: %v", err))
					} else if cachedValue < counterValue {
						firstError.CompareAndSwap(nil, fmt.Sprintf("incremented to value %d but retrieved value %d from cache", counterValue, cachedValue))
					}
				})
			}

			wg.Wait()
			require.EqualValues(t, numIncrements, counter.Load())
			if err := firstError.Load(); err != nil {
				t.Fatal(err)
			}

			// Without invalidating the cache, check what was last stored in it.
			value, err := GetInt64(t.Name(), func() (int64, error) {
				t.Fatal("getFunc should not be invoked")
				return 0, errors.New("getFunc should not be invoked")
			})

			require.NoError(t, err)
			assert.EqualValues(t, numIncrements, value)
		})
	}
}
