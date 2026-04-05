// Copyright (c) 2025 Justin Cranford
//

package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// RunServiceContracts verifies service infrastructure accessor contracts.
// Tests 3 contracts:
//  1. JWKGen() returns a non-nil JWK generation service
//  2. Telemetry() returns a non-nil telemetry service
//  3. Barrier() returns a non-nil barrier (encryption-at-rest) service
func RunServiceContracts(t *testing.T, server ServiceServer) {
	t.Helper()

	t.Run("jwkgen_is_non_nil", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, server.JWKGen(), "JWKGen() must return a non-nil JWK generation service")
	})

	t.Run("telemetry_is_non_nil", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, server.Telemetry(), "Telemetry() must return a non-nil telemetry service")
	})

	t.Run("barrier_is_non_nil", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, server.Barrier(), "Barrier() must return a non-nil barrier service")
	})
}
