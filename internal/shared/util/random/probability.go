// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package random provides cryptographically secure random number generation utilities.
package random

import (
	crand "crypto/rand"
	"fmt"
	"io"
	"math"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	// SkipMessage is the message used when a test is skipped by probability sampling.
	SkipMessage = "Skipped by probability sampling"
)

// SkipByProbability skips the test based on the given probability.
// prob should be between 0.0 and 1.0, where 1.0 means always run, 0.0 means never run.
func SkipByProbability(t *testing.T, prob float32) {
	t.Helper()

	skipByProbability(t, prob, func() (float64, error) {
		return normalizedRandomFloat64(crand.Reader)
	})
}

func skipByProbability(t *testing.T, prob float32, randomFloatFn func() (float64, error)) {
	t.Helper()

	skip, err := shouldSkipByProbability(prob, randomFloatFn)
	require.NoError(t, err)

	if skip {
		t.Skip(SkipMessage)
	}
}

func shouldSkipByProbability(prob float32, randomFloatFn func() (float64, error)) (bool, error) {
	if err := validateProbability(prob); err != nil {
		return false, err
	}

	if randomFloatFn == nil {
		return false, fmt.Errorf("random float function is nil")
	}

	randomFloat, err := randomFloatFn()
	if err != nil {
		return false, fmt.Errorf("failed to generate probability sample: %w", err)
	}

	if math.IsNaN(randomFloat) || randomFloat < float64(cryptoutilSharedMagic.Float32Zero) || randomFloat >= float64(cryptoutilSharedMagic.Float32One) {
		return false, fmt.Errorf("random float %v is outside [0,1)", randomFloat)
	}

	return randomFloat >= float64(prob), nil
}

// validateProbability returns an error if prob is not in [0,1].
func validateProbability(prob float32) error {
	switch {
	case math.IsNaN(float64(prob)):
		return fmt.Errorf("probability is NaN")
	case prob < cryptoutilSharedMagic.Float32Zero:
		return fmt.Errorf("probability %v is less than %v", prob, cryptoutilSharedMagic.Float32Zero)
	case prob > cryptoutilSharedMagic.Float32One:
		return fmt.Errorf("probability %v is greater than %v", prob, cryptoutilSharedMagic.Float32One)
	default:
		return nil
	}
}

// normalizedRandomFloat64 generates a cryptographically secure random float64 in [0,1).
func normalizedRandomFloat64(randomReader io.Reader) (float64, error) {
	randomUint32, err := readRandomUint32(randomReader)
	if err != nil {
		return 0, err
	}

	denominator := math.Ldexp(float64(cryptoutilSharedMagic.Float32One), cryptoutilSharedMagic.BytesPerUint32*cryptoutilSharedMagic.BitsToBytes)

	return float64(randomUint32) / denominator, nil
}

func readRandomUint32(randomReader io.Reader) (uint32, error) {
	if randomReader == nil {
		return 0, fmt.Errorf("random reader is nil")
	}

	var b [cryptoutilSharedMagic.BytesPerUint32]byte

	if _, err := io.ReadFull(randomReader, b[:]); err != nil {
		return 0, fmt.Errorf("failed to read random uint32: %w", err)
	}

	randomUint32 := uint32(0)
	for i, v := range b {
		randomUint32 |= uint32(v) << (i * cryptoutilSharedMagic.BitsToBytes)
	}

	return randomUint32, nil
}

// SamplingBool returns true with probability rate in [0.0, 1.0] using crypto/rand.
// Suitable for production probabilistic decisions such as audit sampling.
// Returns an error if random byte generation fails.
func SamplingBool(rate float64) (bool, error) {
	return samplingBool(rate, crand.Reader)
}

func samplingBool(rate float64, randomReader io.Reader) (bool, error) {
	if math.IsNaN(rate) || rate < cryptoutilSharedMagic.BaselineContributionZero || rate > cryptoutilSharedMagic.TestProbAlways {
		return false, fmt.Errorf("sampling rate %v is outside [0,1]", rate)
	}

	randomFloat, err := normalizedRandomFloat64(randomReader)
	if err != nil {
		return false, fmt.Errorf("failed to generate sampling random: %w", err)
	}

	return randomFloat < rate, nil
}
