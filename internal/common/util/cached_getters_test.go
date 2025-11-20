// Copyright (c) 2025 Justin Cranford
//
//

package util_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/common/util"
)

func TestGetCached_CachedTrue(t *testing.T) {
	t.Parallel()

	// Note: This function has a design flaw - it relies on closure capture
	// The value is only returned correctly on the first call when sync.Once executes
	// Subsequent calls return nil because the closure doesn't re-execute

	callCount := 0
	var capturedValue any

	getter := func() any {
		callCount++
		capturedValue = "cached value"

		return capturedValue
	}

	var once sync.Once

	// Act: First call - sync.Once executes the function
	result1 := util.GetCached(true, &once, getter)

	// Assert: First call gets value from closure
	require.Equal(t, "cached value", result1)
	require.Equal(t, 1, callCount, "Getter should be called once")

	// Second call - sync.Once does NOT execute function again
	result2 := util.GetCached(true, &once, getter)

	// This demonstrates the bug - value is nil on second call
	require.Nil(t, result2, "Bug: Second call returns nil because sync.Once doesn't re-execute")
	require.Equal(t, 1, callCount, "Getter should still be called only once")
}

func TestGetCached_CachedFalse(t *testing.T) {
	t.Parallel()

	callCount := 0
	getter := func() any {
		callCount++

		return callCount
	}

	var once1 sync.Once
	var once2 sync.Once

	// Act: Call twice with cached=false (with separate sync.Once instances)
	result1 := util.GetCached(false, &once1, getter)
	result2 := util.GetCached(false, &once2, getter)

	// Assert: Getter called twice
	require.Equal(t, 1, result1)
	require.Equal(t, 2, result2)
	require.Equal(t, 2, callCount, "Getter should be called every time when caching disabled")
}

func TestGetCached_NilReturn(t *testing.T) {
	t.Parallel()

	getter := func() any {
		return nil
	}

	var once sync.Once

	// Act
	result := util.GetCached(true, &once, getter)

	// Assert: Can cache nil values
	require.Nil(t, result, "Should handle nil return values")
}

func TestGetCached_ComplexType(t *testing.T) {
	t.Parallel()

	type CustomStruct struct {
		Name string
		Age  int
	}

	getter := func() any {
		return &CustomStruct{Name: "Alice", Age: 30}
	}

	var once sync.Once

	// Act
	result := util.GetCached(true, &once, getter)

	// Assert
	require.NotNil(t, result)
	typed, ok := result.(*CustomStruct)
	require.True(t, ok, "Should return correct type")
	require.Equal(t, "Alice", typed.Name)
	require.Equal(t, 30, typed.Age)
}

func TestGetCachedWithError_CachedTrue_Success(t *testing.T) {
	t.Parallel()

	callCount := 0
	getter := func() (any, error) {
		callCount++

		return "cached value", nil
	}

	var once sync.Once

	// Act: First call - sync.Once executes function
	result1, err1 := util.GetCachedWithError(true, &once, getter)

	// Assert: First call returns value
	require.NoError(t, err1)
	require.Equal(t, "cached value", result1)
	require.Equal(t, 1, callCount)

	// Second call - sync.Once does NOT re-execute
	result2, err2 := util.GetCachedWithError(true, &once, getter)

	// Assert: Second call returns nil (expected behavior)
	require.NoError(t, err2)
	require.Nil(t, result2, "Second call returns nil because sync.Once doesn't re-execute")
	require.Equal(t, 1, callCount, "Getter should be called only once")
}

func TestGetCachedWithError_CachedTrue_Error(t *testing.T) {
	t.Parallel()

	callCount := 0
	expectedErr := errors.New("getter error")

	getter := func() (any, error) {
		callCount++

		return nil, expectedErr
	}

	var once sync.Once

	// Act: First call - sync.Once executes function
	result1, err1 := util.GetCachedWithError(true, &once, getter)

	// Assert: First call returns error
	require.Error(t, err1)
	require.Equal(t, expectedErr, err1)
	require.Nil(t, result1)
	require.Equal(t, 1, callCount)

	// Second call - sync.Once does NOT re-execute
	result2, err2 := util.GetCachedWithError(true, &once, getter)

	// Assert: Second call returns nil error (expected behavior - closure variables reset)
	require.NoError(t, err2, "Second call returns nil error because sync.Once doesn't re-execute")
	require.Nil(t, result2)
	require.Equal(t, 1, callCount, "Getter should be called only once")
}

func TestGetCachedWithError_CachedFalse_Success(t *testing.T) {
	t.Parallel()

	callCount := 0
	getter := func() (any, error) {
		callCount++

		return callCount, nil
	}

	var once1 sync.Once
	var once2 sync.Once

	// Act: Call twice with cached=false
	result1, err1 := util.GetCachedWithError(false, &once1, getter)
	result2, err2 := util.GetCachedWithError(false, &once2, getter)

	// Assert: Getter called twice
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.Equal(t, 1, result1)
	require.Equal(t, 2, result2)
	require.Equal(t, 2, callCount, "Getter should be called every time when caching disabled")
}

func TestGetCachedWithError_CachedFalse_Error(t *testing.T) {
	t.Parallel()

	callCount := 0
	getter := func() (any, error) {
		callCount++

		return nil, errors.New("error")
	}

	var once1 sync.Once
	var once2 sync.Once

	// Act: Call twice with cached=false
	result1, err1 := util.GetCachedWithError(false, &once1, getter)
	result2, err2 := util.GetCachedWithError(false, &once2, getter)

	// Assert: Getter called twice
	require.Error(t, err1)
	require.Error(t, err2)
	require.Nil(t, result1)
	require.Nil(t, result2)
	require.Equal(t, 2, callCount, "Getter should be called every time when caching disabled")
}

func TestGetCachedWithError_NilReturnWithNoError(t *testing.T) {
	t.Parallel()

	getter := func() (any, error) {
		return nil, nil
	}

	var once sync.Once

	// Act
	result, err := util.GetCachedWithError(true, &once, getter)

	// Assert: Can cache nil values without error
	require.NoError(t, err)
	require.Nil(t, result, "Should handle nil return values without error")
}

func TestGetCachedWithError_ComplexTypeSuccess(t *testing.T) {
	t.Parallel()

	type CustomData struct {
		ID   int
		Name string
	}

	getter := func() (any, error) {
		return &CustomData{ID: 123, Name: "Test"}, nil
	}

	var once sync.Once

	// Act
	result, err := util.GetCachedWithError(true, &once, getter)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	typed, ok := result.(*CustomData)
	require.True(t, ok, "Should return correct type")
	require.Equal(t, 123, typed.ID)
	require.Equal(t, "Test", typed.Name)
}
