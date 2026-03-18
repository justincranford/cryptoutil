// Copyright (c) 2025 Justin Cranford
//
// Tests for RunResponseFormatContracts.
package contract

import "testing"

func TestRunResponseFormatContracts(t *testing.T) {
	t.Parallel()

	RunResponseFormatContracts(t, testContractServer)
}
