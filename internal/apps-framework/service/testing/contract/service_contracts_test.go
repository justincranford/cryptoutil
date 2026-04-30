// Copyright (c) 2025-2026 Justin Cranford.
//
// Tests for RunServiceContracts.
package contract

import "testing"

func TestRunServiceContracts(t *testing.T) {
	t.Parallel()

	RunServiceContracts(t, testContractServer)
}
