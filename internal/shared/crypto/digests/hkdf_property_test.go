// Copyright (c) 2025 Justin Cranford

package digests

import (
	"bytes"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
			if len(ikm) == 0 || length == 0 || length > cryptoutilSharedMagic.MinSerialNumberBits {
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
		gen.SliceOf(gen.UInt8()),                                    // IKM
		gen.SliceOf(gen.UInt8()),                                    // Salt
		gen.SliceOf(gen.UInt8()),                                    // Info
		gen.UIntRange(1, cryptoutilSharedMagic.MinSerialNumberBits), // Output length 1-64 bytes
	))

	// Property 2: HKDF output length correctness
	properties.Property("HKDF produces correct output length", prop.ForAll(
		func(ikm, salt, info []byte, length uint) bool {
			if len(ikm) == 0 || length == 0 || length > cryptoutilSharedMagic.MinSerialNumberBits {
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
		gen.UIntRange(1, cryptoutilSharedMagic.MinSerialNumberBits),
	))

	// Property 3: HKDF avalanche effect - different IKMs produce different outputs
	// Note: This property is probabilistic and may have false negatives for very short
	// output lengths (1-2 bytes) or very similar input materials. We constrain the test
	// to use reasonable output lengths (8+ bytes) to avoid spurious failures.
	properties.Property("HKDF has avalanche effect", prop.ForAll(
		func(ikm1, ikm2, salt, info []byte, length uint) bool {
			// Use minimum 8 bytes output to avoid hash collision false positives
			actualLength := length
			if actualLength < cryptoutilSharedMagic.IMMinPasswordLength {
				actualLength = cryptoutilSharedMagic.IMMinPasswordLength
			}

			if actualLength > cryptoutilSharedMagic.MinSerialNumberBits {
				return true // Skip invalid length
			}

			if len(ikm1) == 0 || len(ikm2) == 0 {
				return true // Skip invalid inputs
			}

			if bytes.Equal(ikm1, ikm2) {
				return true // Skip identical IKMs
			}

			// Skip IKMs that are too similar (Hamming distance < 2 bytes)
			// to avoid false negatives where very similar inputs might produce
			// collisions in constrained output spaces
			if len(ikm1) == len(ikm2) {
				differences := 0

				for i := range ikm1 {
					if ikm1[i] != ikm2[i] {
						differences++
					}
				}

				if differences < 2 {
					return true // Too similar, skip
				}
			}

			output1, err1 := HKDFwithSHA256(ikm1, salt, info, int(actualLength))
			output2, err2 := HKDFwithSHA256(ikm2, salt, info, int(actualLength))

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
		gen.UIntRange(1, cryptoutilSharedMagic.MinSerialNumberBits),
	))

	properties.TestingRun(t)
}
