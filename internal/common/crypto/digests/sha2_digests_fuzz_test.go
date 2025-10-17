package digests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// FuzzSHA512 tests SHA512 digest with various inputs.
func FuzzSHA512(f *testing.F) {
	// Add seed corpus
	f.Add([]byte("hello world"))
	f.Add([]byte(""))
	f.Add([]byte("a"))
	f.Add(make([]byte, 1000)) // Large input

	f.Fuzz(func(t *testing.T, input []byte) {
		result := SHA512(input)

		// SHA512 always produces 64 bytes
		require.Len(t, result, 64, "SHA512 should produce 64 bytes")

		// Same input should always produce same output
		result2 := SHA512(input)
		require.Len(t, result2, 64, "SHA512 should produce consistent length")
		require.Equal(t, result, result2, "SHA512 should produce consistent output")
	})
}

// FuzzSHA384 tests SHA384 digest with various inputs.
func FuzzSHA384(f *testing.F) {
	// Add seed corpus
	f.Add([]byte("hello world"))
	f.Add([]byte(""))
	f.Add([]byte("a"))
	f.Add(make([]byte, 1000)) // Large input

	f.Fuzz(func(t *testing.T, input []byte) {
		result := SHA384(input)

		// SHA384 always produces 48 bytes
		require.Len(t, result, 48, "SHA384 should produce 48 bytes")

		// Same input should always produce same output
		result2 := SHA384(input)
		require.Len(t, result2, 48, "SHA384 should produce consistent length")
		require.Equal(t, result, result2, "SHA384 should produce consistent output")
	})
}

// FuzzSHA256 tests SHA256 digest with various inputs.
func FuzzSHA256(f *testing.F) {
	// Add seed corpus
	f.Add([]byte("hello world"))
	f.Add([]byte(""))
	f.Add([]byte("a"))
	f.Add(make([]byte, 1000)) // Large input

	f.Fuzz(func(t *testing.T, input []byte) {
		result := SHA256(input)

		// SHA256 always produces 32 bytes
		require.Len(t, result, 32, "SHA256 should produce 32 bytes")

		// Same input should always produce same output
		result2 := SHA256(input)
		require.Len(t, result2, 32, "SHA256 should produce consistent length")
		require.Equal(t, result, result2, "SHA256 should produce consistent output")
	})
}

// FuzzSHA224 tests SHA224 digest with various inputs.
func FuzzSHA224(f *testing.F) {
	// Add seed corpus
	f.Add([]byte("hello world"))
	f.Add([]byte(""))
	f.Add([]byte("a"))
	f.Add(make([]byte, 1000)) // Large input

	f.Fuzz(func(t *testing.T, input []byte) {
		result := SHA224(input)

		// SHA224 always produces 28 bytes
		require.Len(t, result, 28, "SHA224 should produce 28 bytes")

		// Same input should always produce same output
		result2 := SHA224(input)
		require.Len(t, result2, 28, "SHA224 should produce consistent length")
		require.Equal(t, result, result2, "SHA224 should produce consistent output")
	})
}
