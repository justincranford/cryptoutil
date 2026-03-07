// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
"testing"

cryptoutilContract "cryptoutil/internal/apps/template/service/testing/contract"
)

// TestSmIMServer_ContractCompliance verifies sm-im implements all service template contracts.
// Ensures behavioral consistency between sm-im and other cryptoutil services.
func TestSmIMServer_ContractCompliance(t *testing.T) {
t.Parallel()
cryptoutilContract.RunContractTests(t, testSmIMServer)
}
