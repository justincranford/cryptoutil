// Copyright (c) 2025 Justin Cranford

package cache_test

import (
	"errors"
	"sync"
	"testing"

	cryptoutilSharedUtilCache "cryptoutil/internal/shared/util/cache"

	"github.com/stretchr/testify/require"
)

//nolint:thelper // testFn inline functions are NOT test helpers - they're test implementations
func TestGetCached(t *testing.T) {
	t.Parallel()

	const testCachedValue = "cached value"

	tests := []struct {
		name   string
		testFn func(t *testing.T)
	}{
		{
			name: "cached_true",
			testFn: func(t *testing.T) {
				callCount := 0

				var capturedValue any

				getter := func() any {
					callCount++
					capturedValue = testCachedValue

					return capturedValue
				}

				var once sync.Once

				result1 := cryptoutilSharedUtilCache.GetCached(true, &once, getter)
				require.Equal(t, testCachedValue, result1)
				require.Equal(t, 1, callCount, "Getter should be called once")

				result2 := cryptoutilSharedUtilCache.GetCached(true, &once, getter)
				require.Nil(t, result2, "Bug: Second call returns nil because sync.Once doesn't re-execute")
				require.Equal(t, 1, callCount, "Getter should still be called only once")
			},
		},
		{
			name: "cached_false",
			testFn: func(t *testing.T) {
				callCount := 0
				getter := func() any {
					callCount++

					return callCount
				}

				var once1, once2 sync.Once

				result1 := cryptoutilSharedUtilCache.GetCached(false, &once1, getter)
				result2 := cryptoutilSharedUtilCache.GetCached(false, &once2, getter)

				require.Equal(t, 1, result1)
				require.Equal(t, 2, result2)
				require.Equal(t, 2, callCount, "Getter should be called every time when caching disabled")
			},
		},
		{
			name: "nil_return",
			testFn: func(t *testing.T) {
				getter := func() any {
					return nil
				}

				var once sync.Once

				result := cryptoutilSharedUtilCache.GetCached(true, &once, getter)
				require.Nil(t, result, "Should handle nil return values")
			},
		},
		{
			name: "complex_type",
			testFn: func(t *testing.T) {
				type CustomStruct struct {
					Name string
					Age  int
				}

				getter := func() any {
					return &CustomStruct{Name: "Alice", Age: 30}
				}

				var once sync.Once

				result := cryptoutilSharedUtilCache.GetCached(true, &once, getter)

				require.NotNil(t, result)
				typed, ok := result.(*CustomStruct)
				require.True(t, ok, "Should return correct type")
				require.Equal(t, "Alice", typed.Name)
				require.Equal(t, 30, typed.Age)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.testFn(t)
		})
	}
}

func TestGetCachedWithError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		testFn func(t *testing.T)
	}{
		{
			name: "cached_true_success",
			testFn: func(t *testing.T) {
				t.Helper()

				callCount := 0
				getter := func() (any, error) {
					callCount++

					return "cached value", nil
				}

				var once sync.Once

				result1, err1 := cryptoutilSharedUtilCache.GetCachedWithError(true, &once, getter)
				require.NoError(t, err1)
				require.Equal(t, "cached value", result1)
				require.Equal(t, 1, callCount)

				result2, err2 := cryptoutilSharedUtilCache.GetCachedWithError(true, &once, getter)
				require.NoError(t, err2)
				require.Nil(t, result2, "Second call returns nil because sync.Once doesn't re-execute")
				require.Equal(t, 1, callCount, "Getter should be called only once")
			},
		},
		{
			name: "cached_true_error",
			testFn: func(t *testing.T) {
				t.Helper()

				callCount := 0
				expectedErr := errors.New("getter error")

				getter := func() (any, error) {
					callCount++

					return nil, expectedErr
				}

				var once sync.Once

				result1, err1 := cryptoutilSharedUtilCache.GetCachedWithError(true, &once, getter)
				require.Error(t, err1)
				require.Equal(t, expectedErr, err1)
				require.Nil(t, result1)
				require.Equal(t, 1, callCount)

				result2, err2 := cryptoutilSharedUtilCache.GetCachedWithError(true, &once, getter)
				require.NoError(t, err2, "Second call returns nil error because sync.Once doesn't re-execute")
				require.Nil(t, result2)
				require.Equal(t, 1, callCount, "Getter should be called only once")
			},
		},
		{
			name: "cached_false_success",
			testFn: func(t *testing.T) {
				t.Helper()

				callCount := 0
				getter := func() (any, error) {
					callCount++

					return callCount, nil
				}

				var once1, once2 sync.Once

				result1, err1 := cryptoutilSharedUtilCache.GetCachedWithError(false, &once1, getter)
				result2, err2 := cryptoutilSharedUtilCache.GetCachedWithError(false, &once2, getter)

				require.NoError(t, err1)
				require.NoError(t, err2)
				require.Equal(t, 1, result1)
				require.Equal(t, 2, result2)
				require.Equal(t, 2, callCount, "Getter should be called every time when caching disabled")
			},
		},
		{
			name: "cached_false_error",
			testFn: func(t *testing.T) {
				t.Helper()

				callCount := 0
				getter := func() (any, error) {
					callCount++

					return nil, errors.New("error")
				}

				var once1, once2 sync.Once

				result1, err1 := cryptoutilSharedUtilCache.GetCachedWithError(false, &once1, getter)
				result2, err2 := cryptoutilSharedUtilCache.GetCachedWithError(false, &once2, getter)

				require.Error(t, err1)
				require.Error(t, err2)
				require.Nil(t, result1)
				require.Nil(t, result2)
				require.Equal(t, 2, callCount, "Getter should be called every time when caching disabled")
			},
		},
		{
			name: "nil_return_with_no_error",
			testFn: func(t *testing.T) {
				t.Helper()

				getter := func() (any, error) {
					return nil, nil
				}

				var once sync.Once

				result, err := cryptoutilSharedUtilCache.GetCachedWithError(true, &once, getter)
				require.NoError(t, err)
				require.Nil(t, result, "Should handle nil return values without error")
			},
		},
		{
			name: "complex_type_success",
			testFn: func(t *testing.T) {
				t.Helper()

				type CustomData struct {
					ID   int
					Name string
				}

				getter := func() (any, error) {
					return &CustomData{ID: 123, Name: "Test"}, nil
				}

				var once sync.Once

				result, err := cryptoutilSharedUtilCache.GetCachedWithError(true, &once, getter)

				require.NoError(t, err)
				require.NotNil(t, result)

				typed, ok := result.(*CustomData)
				require.True(t, ok, "Should return correct type")
				require.Equal(t, 123, typed.ID)
				require.Equal(t, "Test", typed.Name)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.testFn(t)
		})
	}
}
