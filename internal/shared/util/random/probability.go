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
	bytesPerUint32 = 4
	SkipMessage    = "Skipped by probability sampling"
)

// SkipByProbability skips the test based on the given probability.
// prob should be between 0.0 and 1.0, where 1.0 means always run, 0.0 means never run.
func SkipByProbability(t *testing.T, prob float32) {
	t.Helper()

	require.GreaterOrEqual(t, prob, 0.0)
	require.LessOrEqual(t, prob, 1.0)

	skip := normalizedRandomFloat32(t) > prob
	if skip {
		t.Skip(SkipMessage)
	}
}

// normalizedRandomFloat32 generates a cryptographically secure random float32 in [0,1).
func normalizedRandomFloat32(t *testing.T) float32 {
	var b [bytesPerUint32]byte

	_, err := rand.Read(b[:])
	require.NoError(t, err)

	randomUint32 := uint32(0)
	for i, v := range b {
		randomUint32 |= uint32(v) << (i * bytesPerUint32)
	}

	return float32(randomUint32) / float32(math.MaxUint32)
}
