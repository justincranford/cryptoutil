// Copyright (c) 2025 Justin Cranford

package digests

import (
	"bytes"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

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

			return len(hash) == cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes
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
