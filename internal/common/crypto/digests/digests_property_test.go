// Copyright (c) 2025 Justin Cranford

package digests

import (
	"bytes"
	"errors"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestHKDFInvariants verifies HKDF key derivation properties using property-based testing.
func TestHKDFInvariants(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	// Property 1: HKDF determinism - same inputs always produce same outputs
	properties.Property("HKDF is deterministic", prop.ForAll(
		func(ikm, salt, info []byte, length uint) bool {
			if len(ikm) == 0 || length == 0 || length > 64 {
				return true // Skip invalid inputs (empty IKM, zero/large length)
			}

			output1, err1 := HKDFwithSHA256(ikm, salt, info, int(length))
			output2, err2 := HKDFwithSHA256(ikm, salt, info, int(length))

			// Both should succeed or fail consistently
			if err1 != nil || err2 != nil {
				return errors.Is(err1, err2) || (err1 != nil && err2 != nil)
			}

			// Outputs must be identical
			return bytes.Equal(output1, output2)
		},
		gen.SliceOf(gen.UInt8()), // IKM
		gen.SliceOf(gen.UInt8()), // Salt
		gen.SliceOf(gen.UInt8()), // Info
		gen.UIntRange(1, 64),     // Output length 1-64 bytes
	))

	// Property 2: HKDF output length correctness
	properties.Property("HKDF produces correct output length", prop.ForAll(
		func(ikm, salt, info []byte, length uint) bool {
			if len(ikm) == 0 || length == 0 || length > 64 {
				return true // Skip invalid inputs
			}

			output, err := HKDFwithSHA256(ikm, salt, info, int(length))
			if err != nil {
				return false // Should succeed for valid inputs
			}

			return len(output) == int(length)
		},
		gen.SliceOf(gen.UInt8()),
		gen.SliceOf(gen.UInt8()),
		gen.SliceOf(gen.UInt8()),
		gen.UIntRange(1, 64),
	))

	// Property 3: HKDF avalanche effect - different IKMs produce different outputs
	properties.Property("HKDF has avalanche effect", prop.ForAll(
		func(ikm1, ikm2, salt, info []byte, length uint) bool {
			if len(ikm1) == 0 || len(ikm2) == 0 || length == 0 || length > 64 {
				return true // Skip invalid inputs
			}

			if bytes.Equal(ikm1, ikm2) {
				return true // Skip identical IKMs
			}

			output1, err1 := HKDFwithSHA256(ikm1, salt, info, int(length))
			output2, err2 := HKDFwithSHA256(ikm2, salt, info, int(length))

			if err1 != nil || err2 != nil {
				return true // Skip errors
			}

			// Different inputs should produce different outputs
			return !bytes.Equal(output1, output2)
		},
		gen.SliceOf(gen.UInt8()),
		gen.SliceOf(gen.UInt8()),
		gen.SliceOf(gen.UInt8()),
		gen.SliceOf(gen.UInt8()),
		gen.UIntRange(1, 64),
	))

	properties.TestingRun(t)
}

// TestSHA256Invariants verifies SHA-256 hashing properties using property-based testing.
func TestSHA256Invariants(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	// Property 1: SHA-256 determinism - same input always produces same hash
	properties.Property("SHA-256 is deterministic", prop.ForAll(
		func(input []byte) bool {
			hash1 := SHA256(input)
			hash2 := SHA256(input)

			return bytes.Equal(hash1, hash2)
		},
		gen.SliceOf(gen.UInt8()),
	))

	// Property 2: SHA-256 fixed output length - always 32 bytes (256 bits)
	properties.Property("SHA-256 produces 32-byte output", prop.ForAll(
		func(input []byte) bool {
			hash := SHA256(input)

			return len(hash) == 32
		},
		gen.SliceOf(gen.UInt8()),
	))

	// Property 3: SHA-256 avalanche effect - different inputs produce different hashes
	properties.Property("SHA-256 has avalanche effect", prop.ForAll(
		func(input1, input2 []byte) bool {
			if bytes.Equal(input1, input2) {
				return true // Skip identical inputs
			}

			hash1 := SHA256(input1)
			hash2 := SHA256(input2)

			// Different inputs should produce different hashes
			return !bytes.Equal(hash1, hash2)
		},
		gen.SliceOf(gen.UInt8()),
		gen.SliceOf(gen.UInt8()),
	))

	properties.TestingRun(t)
}
