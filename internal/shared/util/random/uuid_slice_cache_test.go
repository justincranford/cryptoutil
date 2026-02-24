// Copyright (c) 2025 Justin Cranford

package random

import (
	"errors"
	"sync"
	"testing"

	cryptoutilSharedUtil "cryptoutil/internal/shared/util"
	cryptoutilSharedUtilCache "cryptoutil/internal/shared/util/cache"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	testStringValue = "value"
)

// TestGenerateUUIDv7 tests UUID v7 generation.
func TestGenerateUUIDv7(t *testing.T) {
	t.Parallel()

	// Generate UUID.
	uuid, err := GenerateUUIDv7()
	require.NoError(t, err, "Failed to generate UUIDv7")
	require.NotNil(t, uuid, "Generated UUID should not be nil")
	require.NotEqual(t, googleUuid.Nil, *uuid, "Generated UUID should not be zero")
	require.NotEqual(t, googleUuid.Max, *uuid, "Generated UUID should not be max")

	// Generate another UUID and verify they're different.
	uuid2, err := GenerateUUIDv7()
	require.NoError(t, err, "Failed to generate second UUIDv7")
	require.NotEqual(t, *uuid, *uuid2, "UUIDs should be unique")
}

// TestGenerateUUIDv7Function tests the UUID generation function factory.
func TestGenerateUUIDv7Function(t *testing.T) {
	t.Parallel()

	// Get the generator function.
	generator := GenerateUUIDv7Function()
	require.NotNil(t, generator, "Generator function should not be nil")

	// Call the generator function.
	uuid, err := generator()
	require.NoError(t, err, "Failed to generate UUID from function")
	require.NotNil(t, uuid, "Generated UUID should not be nil")
	require.NotEqual(t, googleUuid.Nil, *uuid, "Generated UUID should not be zero")
}

// TestValidateUUID tests UUID validation with parameterized test cases.
func TestValidateUUID(t *testing.T) {
	t.Parallel()

	msg := "test UUID"

	tests := []struct {
		name        string
		uuid        *googleUuid.UUID
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid UUID",
			uuid: func() *googleUuid.UUID {
				u := googleUuid.Must(googleUuid.NewV7())

				return &u
			}(),
			expectError: false,
		},
		{
			name:        "nil UUID",
			uuid:        nil,
			expectError: true,
			errorMsg:    "UUID cannot be nil",
		},
		{
			name: "zero UUID",
			uuid: func() *googleUuid.UUID {
				u := googleUuid.Nil

				return &u
			}(),
			expectError: true,
			errorMsg:    "UUID cannot be zero",
		},
		{
			name: "max UUID",
			uuid: func() *googleUuid.UUID {
				u := googleUuid.Max

				return &u
			}(),
			expectError: true,
			errorMsg:    "UUID cannot be max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

err := ValidateUUID(tt.uuid, msg)

			if tt.expectError {
				require.Error(t, err, "Expected validation error")
				require.Contains(t, err.Error(), "test UUID", "Error should contain message")
			} else {
				require.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

// TestValidateUUIDs tests UUID slice validation.
func TestValidateUUIDs(t *testing.T) {
	t.Parallel()

	msg := "test UUIDs"
	validUUID1 := googleUuid.Must(googleUuid.NewV7())
	validUUID2 := googleUuid.Must(googleUuid.NewV7())

	tests := []struct {
		name        string
		uuids       []googleUuid.UUID
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid UUIDs",
			uuids:       []googleUuid.UUID{validUUID1, validUUID2},
			expectError: false,
		},
		{
			name:        "nil slice",
			uuids:       nil,
			expectError: true,
			errorMsg:    "UUIDs cannot be nil",
		},
		{
			name:        "empty slice",
			uuids:       []googleUuid.UUID{},
			expectError: true,
			errorMsg:    "UUIDs cannot be empty",
		},
		{
			name:        "contains zero UUID",
			uuids:       []googleUuid.UUID{validUUID1, googleUuid.Nil},
			expectError: true,
			errorMsg:    "offset 1",
		},
		{
			name:        "contains max UUID",
			uuids:       []googleUuid.UUID{googleUuid.Max, validUUID1},
			expectError: true,
			errorMsg:    "offset 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

err := ValidateUUIDs(tt.uuids, msg)

			if tt.expectError {
				require.Error(t, err, "Expected validation error")
				require.Contains(t, err.Error(), "test UUIDs", "Error should contain message")
			} else {
				require.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

// TestContains tests the generic Contains function.
func TestContains(t *testing.T) {
	t.Parallel()

	// Test with strings.
	str1 := "apple"
	str2 := "banana"
	str3 := "cherry"
	slice := []*string{&str1, &str2}

	require.True(t, cryptoutilSharedUtil.Contains(slice, &str1), "Should find 'apple' in slice")
	require.True(t, cryptoutilSharedUtil.Contains(slice, &str2), "Should find 'banana' in slice")
	require.False(t, cryptoutilSharedUtil.Contains(slice, &str3), "Should not find 'cherry' in slice")

	// Test with integers.
	int1 := 1
	int2 := 2
	int3 := 3
	intSlice := []*int{&int1, &int2}

	require.True(t, cryptoutilSharedUtil.Contains(intSlice, &int1), "Should find 1 in slice")
	require.False(t, cryptoutilSharedUtil.Contains(intSlice, &int3), "Should not find 3 in slice")
}

// TestGetCached tests the caching behavior of GetCached.
func TestGetCached(t *testing.T) {
	t.Parallel()

	// Test with cached=true.
	t.Run("cached mode", func(t *testing.T) {
		t.Parallel()

		callCount := 0
		getter := func() any {
			callCount++

			return testStringValue
		}
		syncOnce := &sync.Once{}

		// First call should execute getter and return value.
		result1 := cryptoutilSharedUtilCache.GetCached(true, syncOnce, getter)
		require.Equal(t, testStringValue, result1, "First call should return value")
		require.Equal(t, 1, callCount, "Getter should be called once")

		// Second call with same sync.Once should not call getter again,
		// but returns nil because value variable is local to each call.
		result2 := cryptoutilSharedUtilCache.GetCached(true, syncOnce, getter)
		require.Nil(t, result2, "Second call returns nil (sync.Once blocks re-execution)")
		require.Equal(t, 1, callCount, "Getter should not be called again")
	})

	// Test with cached=false (same behavior due to sync.Once).
	t.Run("non-cached mode", func(t *testing.T) {
		t.Parallel()

		callCount := 0
		getter := func() any {
			callCount++

			return testStringValue
		}
		syncOnce := &sync.Once{}

		// First call executes getter.
		result1 := cryptoutilSharedUtilCache.GetCached(false, syncOnce, getter)
		require.Equal(t, testStringValue, result1, "First call should return value")
		require.Equal(t, 1, callCount, "Getter should be called once")

		// Second call doesn't execute getter (sync.Once).
		result2 := cryptoutilSharedUtilCache.GetCached(false, syncOnce, getter)
		require.Nil(t, result2, "Second call returns nil")
		require.Equal(t, 1, callCount, "Getter not called again")
	})
}

// TestGetCachedWithError tests the caching behavior with error handling.
func TestGetCachedWithError(t *testing.T) {
	t.Parallel()

	t.Run("success path", func(t *testing.T) {
		t.Parallel()

		callCount := 0
		getter := func() (any, error) {
			callCount++

			return testStringValue, nil
		}

		syncOnce := &sync.Once{}

		// First call executes getter and returns value.
		result, err := cryptoutilSharedUtilCache.GetCachedWithError(true, syncOnce, getter)
		require.NoError(t, err, "Should not return error")
		require.Equal(t, testStringValue, result, "Should return value")
		require.Equal(t, 1, callCount, "Getter should be called once")

		// Second call doesn't execute getter, returns nil.
		result2, err2 := cryptoutilSharedUtilCache.GetCachedWithError(true, syncOnce, getter)
		require.NoError(t, err2, "Should not return error")
		require.Nil(t, result2, "Second call returns nil")
		require.Equal(t, 1, callCount, "Getter not called again")
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()

		testErr := errors.New("test error")
		callCount := 0
		getter := func() (any, error) {
			callCount++

			return nil, testErr
		}

		syncOnce := &sync.Once{}

		// First call executes getter and returns error.
		result, err := cryptoutilSharedUtilCache.GetCachedWithError(false, syncOnce, getter)
		require.Error(t, err, "Should return error")
		require.ErrorIs(t, err, testErr, "Should return test error")
		require.Nil(t, result, "Should return nil value on error")
		require.Equal(t, 1, callCount, "Getter should be called")
	})
}
