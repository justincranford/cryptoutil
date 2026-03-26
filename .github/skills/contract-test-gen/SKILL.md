---
name: contract-test-gen
description: "Generate cross-service contract compliance tests for cryptoutil services. Use when adding a new service or integration test suite to verify it conforms to all shared framework behavioral contracts (health endpoints, dual-server isolation, response format)."
argument-hint: "[service name or package path]"
---

Generate cross-service contract compliance tests for cryptoutil services.

## Purpose

Every cryptoutil service MUST call RunContractTests to verify it conforms to the shared framework behavioral contracts. Use this skill when:

- Creating a new service from skeleton-template
- Adding integration tests to an existing service
- Verifying a service after migration to the service builder pattern

## What Contracts Are Tested

RunContractTests verifies 8 contracts per service:
1. GET /admin/api/v1/livez returns HTTP 200 with {"status":"alive"}
2. GET /admin/api/v1/readyz (when ready) returns HTTP 200 with {"status":"ready"}
3. GET /browser/api/v1/health returns HTTP 200 with {"status":"healthy"}
4. GET /service/api/v1/health returns HTTP 200 with {"status":"healthy"}
5. Public and admin servers are isolated to different ports
6. Admin server only accessible via 127.0.0.1 (never 0.0.0.0 externally)
7. Response bodies are valid JSON
8. All responses include consistent status field

## Required: Service Interface

The service MUST implement the full server.ServiceServer interface (see internal/apps/framework/service/server/contract.go):

`go
// Add to your service server type declaration to enforce at compile time:
var _ cryptoutilTemplateServiceServer.ServiceServer = (*YourServiceServer)(nil)
`

The key contract methods used by RunContractTests:
- PublicBaseURL() string - returns "<https://127.0.0.1:\><port\>"
- AdminBaseURL() string - returns "<https://127.0.0.1:\><port\>"
- SetReady(bool) - used by RunReadyzNotReadyContract

## TestMain Template (Manual)

`go
package yourservice_test

import (
"context"
"fmt"
"os"
"testing"

cryptoutilContract "cryptoutil/internal/apps/framework/service/testing/contract"
cryptoutilYourServer "cryptoutil/internal/apps/your-ps-id/server"
cryptoutilYourConfig "cryptoutil/internal/apps/your-ps-id/server/config"
cryptoutilE2eHelpers "cryptoutil/internal/apps/framework/service/testing/e2e_helpers"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var testServer *cryptoutilYourServer.YourServiceServer

func TestMain(m *testing.M) {
ctx := context.Background()

cfg := cryptoutilYourConfig.DefaultTestConfig()
var err error

testServer, err = cryptoutilYourServer.NewFromConfig(ctx, cfg)
if err != nil {
panic(fmt.Sprintf("TestMain: failed to create test server: %v", err))
}

cryptoutilE2eHelpers.MustStartAndWaitForDualPorts(testServer, func() error {
return testServer.Start(ctx)
})

testServer.SetReady(true) // MANDATORY: called explicitly here

exitCode := m.Run()

shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
defer cancel()

_ = testServer.Shutdown(shutdownCtx)

os.Exit(exitCode)
}
`

## Contract Test Function

`go
func TestYourService_ContractCompliance(t *testing.T) {
t.Parallel()
cryptoutilContract.RunContractTests(t, testServer)
}
`

## Critical Notes

- **SetReady(true)**: MANDATORY after MustStartAndWaitForDualPorts returns. Without it, readyz returns 503 and the readyz contract fails.
- **DisableKeepAlives: true**: The contract package's built-in HTTP client already sets this. For custom HTTP clients in other integration tests, set it manually.
- **DefaultDataServerShutdownTimeout**: A `time.Duration` constant — NEVER multiply by `time.Second`. Use directly.
- RunReadyzNotReadyContract is NOT included in RunContractTests; call it separately in a sequential (non-parallel) test if needed.

## References

Read [ARCHITECTURE.md Section 10.3.5](../../../docs/ARCHITECTURE.md#1035-cross-service-contract-test-pattern) for full contract pattern documentation.
Read [ARCHITECTURE.md Section 10.3.4](../../../docs/ARCHITECTURE.md#1034-test-http-client-patterns) for DisableKeepAlives requirement.
Read [ARCHITECTURE.md Section 10.3.1](../../../docs/ARCHITECTURE.md#1031-testmain-pattern) for TestMain integration pattern.
