Generate cross-service contract compliance tests for framework behavioral contracts.

**Full Copilot original**: [.github/skills/contract-test-gen/SKILL.md](.github/skills/contract-test-gen/SKILL.md)

Provide the PS-ID to generate contract tests for (e.g., `sm-kms`).

## 8 Contracts Verified

1. **Dual HTTPS endpoints** — public `:8080` and admin `:9090` both start
2. **Health endpoint** — `GET /service/api/v1/health` returns `200 OK`
3. **Livez endpoint** — `GET /admin/api/v1/livez` returns `200 OK`
4. **Readyz endpoint** — `GET /admin/api/v1/readyz` returns `200 OK`
5. **Shutdown endpoint** — `POST /admin/api/v1/shutdown` returns `200 OK`
6. **Server isolation** — admin port only accessible from `127.0.0.1`
7. **Response format consistency** — error responses use shared/apperr Error format
8. **ServiceServer interface** — service implements `ServiceServer` interface

## TestMain Template

```go
// internal/apps/{ps-id}/server/contract_test.go
package server_test

import (
    "os"
    "testing"
)

var (
    sharedServer  *application.Application
    publicBaseURL string
    adminBaseURL  string
)

func TestMain(m *testing.M) {
    ctx := context.Background()

    cfg := config.MustLoadTestConfig()
    sharedServer = application.MustStart(ctx, cfg)

    // CRITICAL: SetReady MUST be called explicitly after ports are up
    // Framework does NOT call SetReady automatically
    sharedServer.MustStartAndWaitForDualPorts(ctx)
    sharedServer.SetReady(true)

    publicBaseURL = fmt.Sprintf("https://127.0.0.1:%d", cfg.PublicPort)
    adminBaseURL = fmt.Sprintf("https://127.0.0.1:%d", cfg.AdminPort)

    code := m.Run()

    sharedServer.Shutdown(ctx)
    os.Exit(code)
}
```

## Contract Test Template

```go
func TestContractHealthEndpoint(t *testing.T) {
    t.Parallel()

    resp, err := sharedHTTPClient.Get(publicBaseURL + "/service/api/v1/health")
    require.NoError(t, err)
    defer resp.Body.Close()

    require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestContractAdminIsolation(t *testing.T) {
    t.Parallel()

    // Admin must be inaccessible from non-loopback
    // (test that it only binds to 127.0.0.1)
    listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", adminPort))
    require.Error(t, err, "admin port must not be accessible from 0.0.0.0")
}
```
