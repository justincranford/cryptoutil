// Copyright (c) 2025 Justin Cranford
//
// Tests for RunContractTests - verifies the full contract suite against the test server.
package contract

import "testing"

func TestRunContractTests(t *testing.T) {
	t.Parallel()

	RunContractTests(t, testContractServer)
}
