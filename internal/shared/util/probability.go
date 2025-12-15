// Copyright (c) 2025 Justin Cranford
//
//

package util

import (
	"math/rand"
	"testing"
)

// SkipByProbability skips the test based on the given probability.
// prob should be between 0.0 and 1.0, where 1.0 means always run, 0.0 means never run.
func SkipByProbability(t *testing.T, prob float64) {
	if rand.Float64() > prob { //nolint:gosec // math/rand is acceptable for test probability sampling
		t.Skip("Skipped by probability sampling")
	}
}
