// Copyright (c) 2025 Justin Cranford
//
//

package digests

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
)

// BenchmarkSHA512 benchmarks SHA512 digest.
func BenchmarkSHA512(b *testing.B) {
	data := []byte("benchmark data for SHA512 hashing performance test")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SHA512(data)
	}
}

// BenchmarkSHA384 benchmarks SHA384 digest.
func BenchmarkSHA384(b *testing.B) {
	data := []byte("benchmark data for SHA384 hashing performance test")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SHA384(data)
	}
}

// BenchmarkSHA256 benchmarks SHA256 digest.
func BenchmarkSHA256(b *testing.B) {
	data := []byte("benchmark data for SHA256 hashing performance test")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SHA256(data)
	}
}

// BenchmarkSHA224 benchmarks SHA224 digest.
func BenchmarkSHA224(b *testing.B) {
	data := []byte("benchmark data for SHA224 hashing performance test")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SHA224(data)
	}
}

// BenchmarkSHA512Large benchmarks SHA512 with large data.
func BenchmarkSHA512Large(b *testing.B) {
	data := make([]byte, cryptoutilSharedMagic.DefaultLogsBatchSize*cryptoutilSharedMagic.DefaultLogsBatchSize) // 1MB of data

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SHA512(data)
	}
}

// BenchmarkSHA256Large benchmarks SHA256 with large data.
func BenchmarkSHA256Large(b *testing.B) {
	data := make([]byte, cryptoutilSharedMagic.DefaultLogsBatchSize*cryptoutilSharedMagic.DefaultLogsBatchSize) // 1MB of data

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SHA256(data)
	}
}
