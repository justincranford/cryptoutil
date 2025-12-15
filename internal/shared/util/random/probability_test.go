package random

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizedRandomFloat32_Range(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
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
		{"AlwaysRun", 1.0, 0.0, false},
		{"NeverRun", 0.0, 1.0, true},
		{"HalfRun_Skip", 0.5, 0.6, true},
		{"HalfRun_Run", 0.5, 0.4, false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("Case/"+tc.name, func(t *testing.T) {
			t.Parallel()

			// inject deterministic rand and restore after test
			orig := randFloat32
			t.Cleanup(func() { randFloat32 = orig })
			randFloat32 = func(t *testing.T) float32 { return tc.randValue }

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
		{"NegativeProb", -0.1, true},
		{"GreaterThanOne", 1.1, true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("Case/"+tc.name, func(t *testing.T) {
			t.Parallel()

			// validateProbability returns an error for invalid inputs
			err := validateProbability(tc.prob)
			require.Error(t, err)
		})
	}
}
