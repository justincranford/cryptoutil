// Copyright (c) 2025 Justin Cranford
//
//

package digests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// IMPORTANT: All Fuzz* test function names MUST be unique and MUST NOT be substrings of any other fuzz test names.
// This ensures cross-platform compatibility with the `-fuzz` parameter (no quotes or regex needed).
// Example: "FuzzHKDF" conflicts with "FuzzHKDFwithSHA256", so we use "FuzzHKDFAllVariants" instead.

// FuzzHKDFAllVariants tests HKDF function with various inputs to ensure it doesn't crash.
func FuzzHKDFAllVariants(f *testing.F) {
	// Add seed corpus with valid inputs
	f.Add("SHA256", []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFSHA256OutputLength)
	f.Add("SHA384", []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFSHA384OutputLength)
	f.Add("SHA512", []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFSHA512OutputLength)
	f.Add("SHA224", []byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFSHA224OutputLength)

	f.Fuzz(func(t *testing.T, digestName string, secret, salt, info []byte, outputLength int) {
		// Skip invalid digest names
		if digestName != "SHA224" && digestName != "SHA256" && digestName != "SHA384" && digestName != "SHA512" {
			t.Skip("Skipping invalid digest name for fuzzing")
		}

		// Skip invalid output lengths
		if outputLength <= 0 || outputLength > cryptoutilSharedMagic.HKDFSHA512MaxLength { // Max for SHA512
			t.Skip("Skipping invalid output length for fuzzing")
		}

		// Skip nil or empty secrets
		if len(secret) == 0 {
			t.Skip("Skipping empty secret for fuzzing")
		}

		result, err := HKDF(digestName, secret, salt, info, outputLength)

		// For valid inputs, we should get a result or a specific error
		if err == nil {
			require.Len(t, result, outputLength, "HKDF should return result of correct length")
		} else {
			// Check that error is one of the expected errors
			require.True(t, errors.Is(err, ErrInvalidNilDigestFunction) ||
				errors.Is(err, ErrInvalidNilSecret) ||
				errors.Is(err, ErrInvalidEmptySecret) ||
				errors.Is(err, ErrInvalidOutputBytesLengthNegative) ||
				errors.Is(err, ErrInvalidOutputBytesLengthZero) ||
				errors.Is(err, ErrInvalidOutputBytesLengthTooBig),
				"HKDF should return expected error, got: %v", err)
		}
	})
}

// FuzzHKDFwithSHA256 tests HKDF-SHA256 with various inputs.
func FuzzHKDFwithSHA256(f *testing.F) {
	f.Add([]byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFSHA256OutputLength)

	f.Fuzz(func(t *testing.T, secret, salt, info []byte, outputLength int) {
		// Skip invalid output lengths for SHA256 (max 255*32 = 8160)
		if outputLength <= 0 || outputLength > cryptoutilSharedMagic.HKDFSHA256MaxLength {
			t.Skip("Skipping invalid output length for fuzzing")
		}

		// Skip empty secrets
		if len(secret) == 0 {
			t.Skip("Skipping empty secret for fuzzing")
		}

		result, err := HKDFwithSHA256(secret, salt, info, outputLength)
		if err == nil {
			require.Len(t, result, outputLength, "HKDFwithSHA256 should return result of correct length")
		} else {
			// Should only get parameter validation errors
			require.True(t, errors.Is(err, ErrInvalidNilSecret) ||
				errors.Is(err, ErrInvalidEmptySecret) ||
				errors.Is(err, ErrInvalidOutputBytesLengthNegative) ||
				errors.Is(err, ErrInvalidOutputBytesLengthZero) ||
				errors.Is(err, ErrInvalidOutputBytesLengthTooBig),
				"HKDFwithSHA256 should return expected error, got: %v", err)
		}
	})
}

// FuzzHKDFwithSHA384 tests HKDF-SHA384 with various inputs.
func FuzzHKDFwithSHA384(f *testing.F) {
	f.Add([]byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFSHA384OutputLength)

	f.Fuzz(func(t *testing.T, secret, salt, info []byte, outputLength int) {
		// Skip invalid output lengths for SHA384 (max 255*48 = 12240)
		if outputLength <= 0 || outputLength > cryptoutilSharedMagic.HKDFSHA384MaxLength {
			t.Skip("Skipping invalid output length for fuzzing")
		}

		// Skip empty secrets
		if len(secret) == 0 {
			t.Skip("Skipping empty secret for fuzzing")
		}

		result, err := HKDFwithSHA384(secret, salt, info, outputLength)
		if err == nil {
			require.Len(t, result, outputLength, "HKDFwithSHA384 should return result of correct length")
		} else {
			// Should only get parameter validation errors
			require.True(t, errors.Is(err, ErrInvalidNilSecret) ||
				errors.Is(err, ErrInvalidEmptySecret) ||
				errors.Is(err, ErrInvalidOutputBytesLengthNegative) ||
				errors.Is(err, ErrInvalidOutputBytesLengthZero) ||
				errors.Is(err, ErrInvalidOutputBytesLengthTooBig),
				"HKDFwithSHA384 should return expected error, got: %v", err)
		}
	})
}

// FuzzHKDFwithSHA512 tests HKDF-SHA512 with various inputs.
func FuzzHKDFwithSHA512(f *testing.F) {
	f.Add([]byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFSHA512OutputLength)

	f.Fuzz(func(t *testing.T, secret, salt, info []byte, outputLength int) {
		// Skip invalid output lengths for SHA512 (max 255*64 = 16320)
		if outputLength <= 0 || outputLength > cryptoutilSharedMagic.HKDFSHA512MaxLength {
			t.Skip("Skipping invalid output length for fuzzing")
		}

		// Skip empty secrets
		if len(secret) == 0 {
			t.Skip("Skipping empty secret for fuzzing")
		}

		result, err := HKDFwithSHA512(secret, salt, info, outputLength)
		if err == nil {
			require.Len(t, result, outputLength, "HKDFwithSHA512 should return result of correct length")
		} else {
			// Should only get parameter validation errors
			require.True(t, errors.Is(err, ErrInvalidNilSecret) ||
				errors.Is(err, ErrInvalidEmptySecret) ||
				errors.Is(err, ErrInvalidOutputBytesLengthNegative) ||
				errors.Is(err, ErrInvalidOutputBytesLengthZero) ||
				errors.Is(err, ErrInvalidOutputBytesLengthTooBig),
				"HKDFwithSHA512 should return expected error, got: %v", err)
		}
	})
}

// FuzzHKDFwithSHA224 tests HKDF-SHA224 with various inputs.
func FuzzHKDFwithSHA224(f *testing.F) {
	f.Add([]byte("secret"), []byte("salt"), []byte("info"), cryptoutilSharedMagic.HKDFSHA224OutputLength)

	f.Fuzz(func(t *testing.T, secret, salt, info []byte, outputLength int) {
		// Skip invalid output lengths for SHA224 (max 255*28 = 7140)
		if outputLength <= 0 || outputLength > cryptoutilSharedMagic.HKDFSHA224MaxLength {
			t.Skip("Skipping invalid output length for fuzzing")
		}

		// Skip empty secrets
		if len(secret) == 0 {
			t.Skip("Skipping empty secret for fuzzing")
		}

		result, err := HKDFwithSHA224(secret, salt, info, outputLength)
		if err == nil {
			require.Len(t, result, outputLength, "HKDFwithSHA224 should return result of correct length")
		} else {
			// Should only get parameter validation errors
			require.True(t, errors.Is(err, ErrInvalidNilSecret) ||
				errors.Is(err, ErrInvalidEmptySecret) ||
				errors.Is(err, ErrInvalidOutputBytesLengthNegative) ||
				errors.Is(err, ErrInvalidOutputBytesLengthZero) ||
				errors.Is(err, ErrInvalidOutputBytesLengthTooBig),
				"HKDFwithSHA224 should return expected error, got: %v", err)
		}
	})
}
