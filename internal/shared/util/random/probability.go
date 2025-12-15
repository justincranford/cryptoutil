// Copyright (c) 2025 Justin Cranford
//
//

package random

import (
	"crypto/rand"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	bytesPerUint32 = 4
	bitsPerByte    = 8
	SkipMessage    = "Skipped by probability sampling"
	float32_0      = float32(0.0)
	float32_1      = float32(1.0)
)

// SkipByProbability skips the test based on the given probability.
// prob should be between 0.0 and 1.0, where 1.0 means always run, 0.0 means never run.
func SkipByProbability(t *testing.T, prob float32) {
	t.Helper()

	require.NoError(t, validateProbability(prob))

	// Use injectable random float generator to allow deterministic testing.
	skip := randFloat32(t) > prob
	if skip {
		t.Skip(SkipMessage)
	}
}

// validateProbability returns an error if prob not in [0,1].
func validateProbability(prob float32) error {
	if prob < float32_0 {
		return fmt.Errorf("probability %v is less than %v", prob, float32_0)
	}

	if prob > float32_1 {
		return fmt.Errorf("probability %v is greater than %v", prob, float32_1)
	}

	return nil
}

// normalizedRandomFloat32 generates a cryptographically secure random float32 in [0,1).
func normalizedRandomFloat32(t *testing.T) float32 {
	var b [bytesPerUint32]byte

	_, err := rand.Read(b[:])
	require.NoError(t, err)

	randomUint32 := uint32(0)
	for i, v := range b {
		// shift by bits (8 bits per byte) not by number of bytes
		randomUint32 |= uint32(v) << (i * bitsPerByte)
	}

	return float32(randomUint32) / float32(math.MaxUint32)
}

// randFloat32 is a variable so tests can inject deterministic behavior.
var randFloat32 = func(t *testing.T) float32 { return normalizedRandomFloat32(t) }
