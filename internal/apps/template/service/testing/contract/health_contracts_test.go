// Copyright (c) 2025 Justin Cranford
//
// Tests for RunHealthContracts and RunReadyzNotReadyContract.
package contract

import "testing"

func TestRunHealthContracts(t *testing.T) {
	t.Parallel()

	RunHealthContracts(t, testContractServer)
}

// TestRunReadyzNotReadyContract tests that readyz returns 503 when server is not ready.
// Not parallel: temporarily modifies server ready state.
func TestRunReadyzNotReadyContract(t *testing.T) {
	RunReadyzNotReadyContract(t, testContractServer)
}
