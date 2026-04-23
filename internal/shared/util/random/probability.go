// Copyright (c) 2025 Justin Cranford
//
//

// Package random provides cryptographically secure random number generation utilities.
package random

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"sync"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	// SkipMessage is the message used when a test is skipped by probability sampling.
	SkipMessage = "Skipped by probability sampling"
)

var (
	// randFloat32 is a variable so tests can inject deterministic behavior.
	randFloat32 = func(t *testing.T) float32 { return normalizedRandomFloat32(t) }

	// randFloat32Mutex protects randFloat32 modifications for thread-safe injection.
	randFloat32Mutex sync.RWMutex
)

// SkipByProbability skips the test based on the given probability.
// prob should be between 0.0 and 1.0, where 1.0 means always run, 0.0 means never run.
func SkipByProbability(t *testing.T, prob float32) {
	t.Helper()

	require.NoError(t, validateProbability(prob))

	// Use injectable random float generator to allow deterministic testing.
	randFloat32Mutex.RLock()

	randFn := randFloat32

	randFloat32Mutex.RUnlock()

	skip := randFn(t) > prob
	if skip {
		t.Skip(SkipMessage)
	}
}

// validateProbability returns an error if prob not in [0,1].
func validateProbability(prob float32) error {
	if prob < cryptoutilSharedMagic.Float32Zero {
		return fmt.Errorf("probability %v is less than %v", prob, cryptoutilSharedMagic.Float32Zero)
	}

	if prob > cryptoutilSharedMagic.Float32One {
		return fmt.Errorf("probability %v is greater than %v", prob, cryptoutilSharedMagic.Float32One)
	}

	return nil
}

// normalizedRandomFloat32 generates a cryptographically secure random float32 in [0,1).
func normalizedRandomFloat32(t *testing.T) float32 {
	var b [cryptoutilSharedMagic.BytesPerUint32]byte

	_, err := crand.Read(b[:])
	require.NoError(t, err)

	randomUint32 := uint32(0)
	for i, v := range b {
		// shift by bits (8 bits per byte) not by number of bytes.
		randomUint32 |= uint32(v) << (i * cryptoutilSharedMagic.BitsToBytes)
	}

	return float32(randomUint32) / float32(math.MaxUint32)
}

// SamplingBool returns true with probability rate in [0.0, 1.0) using crypto/rand.
// Suitable for production probabilistic decisions such as audit sampling.
// Returns an error if random byte generation fails.
func SamplingBool(rate float64) (bool, error) {
	var b [cryptoutilSharedMagic.BytesPerUint32]byte

	if _, err := crand.Read(b[:]); err != nil {
		return false, fmt.Errorf("failed to generate sampling random: %w", err)
	}

	randomUint32 := uint32(0)
	for i, v := range b {
		randomUint32 |= uint32(v) << (i * cryptoutilSharedMagic.BitsToBytes)
	}

	return float64(randomUint32)/float64(math.MaxUint32) < rate, nil
}
