// Copyright (c) 2025 Justin Cranford
//
// Tests for RunServiceContracts.
package contract

import "testing"

func TestRunServiceContracts(t *testing.T) {
	t.Parallel()

	RunServiceContracts(t, testContractServer)
}
