// Copyright (c) 2025 Justin Cranford

package random

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestNormalizedRandomFloat32_Range(t *testing.T) {
	t.Parallel()

	for i := 0; i < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; i++ {
		v := normalizedRandomFloat32(t)
		require.GreaterOrEqual(t, v, float32_0)
		require.Less(t, v, float32_1)
	}
}

func TestSkipByProbability_HappyPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		prob       float32
		randValue  float32
		shouldSkip bool
	}{
		{"AlwaysRun", cryptoutilSharedMagic.TestProbAlways, cryptoutilSharedMagic.BaselineContributionZero, false},
		{"NeverRun", cryptoutilSharedMagic.BaselineContributionZero, cryptoutilSharedMagic.TestProbAlways, true},
		{"HalfRun_Skip", cryptoutilSharedMagic.Tolerance50Percent, cryptoutilSharedMagic.RiskScoreHigh, true},
		{"HalfRun_Run", cryptoutilSharedMagic.Tolerance50Percent, cryptoutilSharedMagic.RiskScoreMedium, false},
	}

	for _, tc := range tests {
		t.Run("Case/"+tc.name, func(t *testing.T) {
			// CANNOT use t.Parallel() - parallel tests cause race condition with global randFloat32
			// even with mutex protection, because t.Cleanup runs after subtests complete
			// inject deterministic rand and restore after test (thread-safe with mutex)
			randFloat32Mutex.Lock()

			orig := randFloat32
			randFloat32 = func(_ *testing.T) float32 { return tc.randValue }

			randFloat32Mutex.Unlock()

			defer func() {
				randFloat32Mutex.Lock()

				randFloat32 = orig

				randFloat32Mutex.Unlock()
			}()

			// Run inner subtest to observe whether SkipByProbability calls t.Skip
			didSkip := false
			ok := t.Run("inner", func(t *testing.T) {
				defer func() { didSkip = t.Skipped() }()

				SkipByProbability(t, tc.prob)
			})

			// t.Run returns false if the subtest failed
			require.True(t, ok, "inner subtest failed unexpectedly")
			require.Equal(t, tc.shouldSkip, didSkip)
		})
	}
}

func TestSkipByProbability_SadPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		prob         float32
		expectsPanic bool
	}{
		{"NegativeProb", -cryptoutilSharedMagic.Tolerance10Percent, true},
		{"GreaterThanOne", 1.1, true},
	}

	for _, tc := range tests {
		t.Run("Case/"+tc.name, func(t *testing.T) {
			t.Parallel()

			// validateProbability returns an error for invalid inputs
			err := validateProbability(tc.prob)
			require.Error(t, err)
		})
	}
}

func TestSamplingBool_HappyPaths(t *testing.T) {
	t.Parallel()

	t.Run("Case/RateZero_AlwaysFalse", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; i++ {
			result, err := SamplingBool(cryptoutilSharedMagic.BaselineContributionZero)
			require.NoError(t, err)
			require.False(t, result, "rate=0.0 must always return false")
		}
	})

	t.Run("Case/RateOne_AlwaysTrue", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < cryptoutilSharedMagic.JoseJADefaultMaxMaterials; i++ {
			result, err := SamplingBool(cryptoutilSharedMagic.TestProbAlways)
			require.NoError(t, err)
			require.True(t, result, "rate=1.0 must always return true")
		}
	})

	t.Run("Case/RateHalf_StatisticalDistribution", func(t *testing.T) {
		t.Parallel()

		const iterations = 1000

		trueCount := 0

		for i := 0; i < iterations; i++ {
			result, err := SamplingBool(cryptoutilSharedMagic.ConfidenceWeightFactors)
			require.NoError(t, err)

			if result {
				trueCount++
			}
		}

		// Expect ~50% true; allow ±20% tolerance for statistical variance.
		ratio := float64(trueCount) / iterations
		require.InDelta(t, cryptoutilSharedMagic.ConfidenceWeightFactors, ratio, 0.2, "SamplingBool(0.5) should return true ~50%% of the time")
	})
}
