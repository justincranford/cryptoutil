// Copyright (c) 2025 Justin Cranford
//
//

package digests

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchmarkHKDFwithSHA256 benchmarks HKDF-SHA256.
func BenchmarkHKDFwithSHA256(b *testing.B) {
	secret := []byte("benchmark secret")
	salt := []byte("benchmark salt")
	info := []byte("benchmark info")
	outputLength := cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := HKDFwithSHA256(secret, salt, info, outputLength)
		require.NoError(b, err, "HKDFwithSHA256 should not fail")
	}
}

// BenchmarkHKDFwithSHA384 benchmarks HKDF-SHA384.
func BenchmarkHKDFwithSHA384(b *testing.B) {
	secret := []byte("benchmark secret")
	salt := []byte("benchmark salt")
	info := []byte("benchmark info")
	outputLength := cryptoutilSharedMagic.HMACSHA384KeySize

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := HKDFwithSHA384(secret, salt, info, outputLength)
		require.NoError(b, err, "HKDFwithSHA384 should not fail")
	}
}

// BenchmarkHKDFwithSHA512 benchmarks HKDF-SHA512.
func BenchmarkHKDFwithSHA512(b *testing.B) {
	secret := []byte("benchmark secret")
	salt := []byte("benchmark salt")
	info := []byte("benchmark info")
	outputLength := cryptoutilSharedMagic.MinSerialNumberBits

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := HKDFwithSHA512(secret, salt, info, outputLength)
		require.NoError(b, err, "HKDFwithSHA512 should not fail")
	}
}

// BenchmarkHKDFwithSHA224 benchmarks HKDF-SHA224.
func BenchmarkHKDFwithSHA224(b *testing.B) {
	secret := []byte("benchmark secret")
	salt := []byte("benchmark salt")
	info := []byte("benchmark info")
	outputLength := cryptoutilSharedMagic.HKDFSHA224OutputLength

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := HKDFwithSHA224(secret, salt, info, outputLength)
		require.NoError(b, err, "HKDFwithSHA224 should not fail")
	}
}
