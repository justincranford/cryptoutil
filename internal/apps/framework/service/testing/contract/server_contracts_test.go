// Copyright (c) 2025 Justin Cranford
//
// Tests for RunServerContracts.
package contract

import "testing"

func TestRunServerContracts(t *testing.T) {
	t.Parallel()

	RunServerContracts(t, testContractServer)
}
