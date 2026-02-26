// Copyright (c) 2025 Justin Cranford

package random

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testHello = "hello"
	testWorld = "world"
)

// TestStringPointersToBytes tests converting string pointers to byte slices.
func TestStringPointersToBytes(t *testing.T) {
	t.Parallel()

	t.Run("all valid pointers", func(t *testing.T) {
		t.Parallel()

		str1 := testHello
		str2 := testWorld
		str3 := "test"

		result := StringPointersToBytes(&str1, &str2, &str3)
		require.Len(t, result, 3, "Should have 3 byte slices")
		require.Equal(t, []byte(testHello), result[0])
		require.Equal(t, []byte(testWorld), result[1])
		require.Equal(t, []byte("test"), result[2])
	})

	t.Run("with nil pointers", func(t *testing.T) {
		t.Parallel()

		str1 := testHello
		str2 := testWorld

		result := StringPointersToBytes(&str1, nil, &str2, nil)
		require.Len(t, result, 2, "Should skip nil pointers")
		require.Equal(t, []byte(testHello), result[0])
		require.Equal(t, []byte(testWorld), result[1])
	})

	t.Run("all nil pointers", func(t *testing.T) {
		t.Parallel()

		result := StringPointersToBytes(nil, nil, nil)
		require.Len(t, result, 0, "Should return empty slice")
	})

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()

		result := StringPointersToBytes()
		require.Len(t, result, 0, "Should return empty slice")
	})
}

// TestUint64ToBytes tests converting uint64 to bytes.
func TestUint64ToBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    uint64
		expected []byte
	}{
		{
			name:     "zero",
			value:    0,
			expected: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "one",
			value:    1,
			expected: []byte{0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			name:     "max uint64",
			value:    0xFFFFFFFFFFFFFFFF,
			expected: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "specific value",
			value:    0x0102030405060708,
			expected: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Uint64ToBytes(tt.value)
			require.Equal(t, tt.expected, result)
			require.Len(t, result, cryptoutilSharedMagic.IMMinPasswordLength, "Should always be 8 bytes")

			// Verify round-trip conversion.
			decoded := binary.BigEndian.Uint64(result)
			require.Equal(t, tt.value, decoded, "Should decode back to original value")
		})
	}
}

// TestUint32ToBytes tests converting uint32 to bytes.
func TestUint32ToBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    uint32
		expected []byte
	}{
		{
			name:     "zero",
			value:    0,
			expected: []byte{0, 0, 0, 0},
		},
		{
			name:     "one",
			value:    1,
			expected: []byte{0, 0, 0, 1},
		},
		{
			name:     "max uint32",
			value:    0xFFFFFFFF,
			expected: []byte{0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "specific value",
			value:    0x01020304,
			expected: []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Uint32ToBytes(tt.value)
			require.Equal(t, tt.expected, result)
			require.Len(t, result, 4, "Should always be 4 bytes")

			// Verify round-trip conversion.
			decoded := binary.BigEndian.Uint32(result)
			require.Equal(t, tt.value, decoded, "Should decode back to original value")
		})
	}
}

// TestUint16ToBytes tests converting uint16 to bytes.
func TestUint16ToBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    uint16
		expected []byte
	}{
		{
			name:     "zero",
			value:    0,
			expected: []byte{0, 0},
		},
		{
			name:     "one",
			value:    1,
			expected: []byte{0, 1},
		},
		{
			name:     "max uint16",
			value:    0xFFFF,
			expected: []byte{0xFF, 0xFF},
		},
		{
			name:     "specific value",
			value:    0x0102,
			expected: []byte{0x01, 0x02},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Uint16ToBytes(tt.value)
			require.Equal(t, tt.expected, result)
			require.Len(t, result, 2, "Should always be 2 bytes")

			// Verify round-trip conversion.
			decoded := binary.BigEndian.Uint16(result)
			require.Equal(t, tt.value, decoded, "Should decode back to original value")
		})
	}
}

// TestInt64ToBytes tests converting int64 to bytes.
func TestInt64ToBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    int64
		expected []byte
	}{
		{
			name:     "zero",
			value:    0,
			expected: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "positive one",
			value:    1,
			expected: []byte{0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			name:     "negative one",
			value:    -1,
			expected: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "max int64",
			value:    0x7FFFFFFFFFFFFFFF,
			expected: []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "min int64",
			value:    -0x8000000000000000,
			expected: []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Int64ToBytes(tt.value)
			require.Equal(t, tt.expected, result)
			require.Len(t, result, cryptoutilSharedMagic.IMMinPasswordLength, "Should always be 8 bytes")
		})
	}
}

// TestInt32ToBytes tests converting int32 to bytes.
func TestInt32ToBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    int32
		expected []byte
	}{
		{
			name:     "zero",
			value:    0,
			expected: []byte{0, 0, 0, 0},
		},
		{
			name:     "positive one",
			value:    1,
			expected: []byte{0, 0, 0, 1},
		},
		{
			name:     "negative one",
			value:    -1,
			expected: []byte{0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "max int32",
			value:    0x7FFFFFFF,
			expected: []byte{0x7F, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "min int32",
			value:    -0x80000000,
			expected: []byte{0x80, 0x00, 0x00, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Int32ToBytes(tt.value)
			require.Equal(t, tt.expected, result)
			require.Len(t, result, 4, "Should always be 4 bytes")
		})
	}
}

// TestInt16ToBytes tests converting int16 to bytes.
func TestInt16ToBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    int16
		expected []byte
	}{
		{
			name:     "zero",
			value:    0,
			expected: []byte{0, 0},
		},
		{
			name:     "positive one",
			value:    1,
			expected: []byte{0, 1},
		},
		{
			name:     "negative one",
			value:    -1,
			expected: []byte{0xFF, 0xFF},
		},
		{
			name:     "max int16",
			value:    0x7FFF,
			expected: []byte{0x7F, 0xFF},
		},
		{
			name:     "min int16",
			value:    -0x8000,
			expected: []byte{0x80, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Int16ToBytes(tt.value)
			require.Equal(t, tt.expected, result)
			require.Len(t, result, 2, "Should always be 2 bytes")
		})
	}
}

// TestSafeIntConversions tests the internal safe conversion functions.
func TestSafeIntConversions(t *testing.T) {
	t.Parallel()

	t.Run("safeIntToUint64", func(t *testing.T) {
		t.Parallel()

		// Positive value.
		require.Equal(t, uint64(1), safeIntToUint64(1))

		// Zero.
		require.Equal(t, uint64(0), safeIntToUint64(0))

		// Negative value (two's complement).
		require.Equal(t, uint64(0xFFFFFFFFFFFFFFFF), safeIntToUint64(-1))
	})

	t.Run("safeIntToUint32", func(t *testing.T) {
		t.Parallel()

		// Positive value.
		require.Equal(t, uint32(1), safeIntToUint32(1))

		// Zero.
		require.Equal(t, uint32(0), safeIntToUint32(0))

		// Negative value (two's complement).
		require.Equal(t, uint32(0xFFFFFFFF), safeIntToUint32(-1))
	})

	t.Run("safeIntToUint16", func(t *testing.T) {
		t.Parallel()

		// Positive value.
		require.Equal(t, uint16(1), safeIntToUint16(1))

		// Zero.
		require.Equal(t, uint16(0), safeIntToUint16(0))

		// Negative value (two's complement).
		require.Equal(t, uint16(0xFFFF), safeIntToUint16(-1))
	})
}
