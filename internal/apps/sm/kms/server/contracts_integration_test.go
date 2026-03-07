//go:build integration

// Copyright (c) 2025 Justin Cranford
//

package server

import (
"testing"

cryptoutilContract "cryptoutil/internal/apps/template/service/testing/contract"
)

// TestKMSServer_ContractCompliance verifies sm-kms implements all service template contracts.
// Ensures behavioral consistency between sm-kms and other cryptoutil services.
func TestKMSServer_ContractCompliance(t *testing.T) {
t.Parallel()
cryptoutilContract.RunContractTests(t, testIntegrationServer)
}