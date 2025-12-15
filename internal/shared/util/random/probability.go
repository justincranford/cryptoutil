// Copyright (c) 2025 Justin Cranford
//
//

package random

import (
	"crypto/rand"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// bytesPerUint64 is the number of bytes in a uint64.
	bytesPerUint64 = 8
)

// SkipByProbability skips the test based on the given probability.
// prob should be between 0.0 and 1.0, where 1.0 means always run, 0.0 means never run.
func SkipByProbability(t *testing.T, prob float32) {
	t.Helper()

	require.GreaterOrEqual(t, prob, 0.0)
	require.LessOrEqual(t, prob, 1.0)

	if normalizedRandomFloat64(t) > prob {
		t.Skip("Skipped by probability sampling")
	}
}

// normalizedRandomFloat64 generates a cryptographically secure random float64 in [0,1).
func normalizedRandomFloat64(t *testing.T) float32 {
	var b [bytesPerUint64]byte

	_, err := rand.Read(b[:])
	require.NoError(t, err)

	randomUint64 := uint64(0)
	for i, v := range b {
		randomUint64 |= uint64(v) << (i * bytesPerUint64)
	}

	return float32(randomUint64) / float32(math.MaxUint64)
}
