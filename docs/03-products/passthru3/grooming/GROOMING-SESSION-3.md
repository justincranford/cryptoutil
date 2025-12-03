# Grooming Session 3: Testing & Validation Strategy

**Purpose**: Define complete testing approach BEFORE implementation
**Date**: 2025-12-01

---

## Topic 1: Unit Testing Requirements

### Q1.1: What unit tests are required for integration.go?

**Decision**: Test helper functions, not the full demo flow

**Required Tests**:

| Function | Test Cases |
|----------|-----------|
| `newDemoHTTPClient()` | Returns valid client with TLS config |
| `waitForHealth()` | Success on healthy endpoint, timeout on unhealthy |
| `getClientCredentialsToken()` | Valid response parsing, error handling |
| `validateJWT()` | Valid token, expired token, invalid signature |

**Test File**: `internal/cmd/demo/integration_test.go`

### Q1.2: What mocking approach?

**Decision**: Use httptest.Server for HTTP endpoints

**Pattern**:

```go
func TestWaitForHealth(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name       string
        handler    http.HandlerFunc
        timeout    time.Duration
        wantErr    bool
    }{
        {
            name: "healthy endpoint returns immediately",
            handler: func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
            },
            timeout: 5 * time.Second,
            wantErr: false,
        },
        {
            name: "unhealthy endpoint times out",
            handler: func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusServiceUnavailable)
            },
            timeout: 100 * time.Millisecond,
            wantErr: true,
        },
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            srv := httptest.NewServer(tc.handler)
            defer srv.Close()

            ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
            defer cancel()

            err := waitForHealth(ctx, http.DefaultClient, srv.URL)
            if tc.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

---

## Topic 2: Manual Verification Commands

### Q2.1: Complete manual verification checklist

**Pre-Implementation Verification**:

```bash
# Verify existing demos still work
go run ./cmd/demo kms
go run ./cmd/demo identity
```

**Post-Implementation Verification**:

```bash
# 1. Build verification
go build ./...

# 2. Lint verification
golangci-lint run ./internal/cmd/demo/...

# 3. Unit test verification
go test -v ./internal/cmd/demo/...

# 4. Integration demo verification
go run ./cmd/demo all

# 5. Full demo suite
go run ./cmd/demo  # Should show all available demos
```

### Q2.2: Docker Compose verification commands

**Identity Docker Compose**:

```bash
# Start
docker compose -f deployments/identity/compose.demo.yml --profile demo up -d

# Verify
docker compose -f deployments/identity/compose.demo.yml --profile demo ps
curl -k https://localhost:8082/.well-known/openid-configuration

# Stop
docker compose -f deployments/identity/compose.demo.yml --profile demo down -v
```

**KMS Docker Compose**:

```bash
# Start
docker compose -f deployments/kms/compose.demo.yml --profile demo up -d

# Verify
docker compose -f deployments/kms/compose.demo.yml --profile demo ps
curl -k https://localhost:8080/ui/swagger/doc.json

# Stop
docker compose -f deployments/kms/compose.demo.yml --profile demo down -v
```

---

## Topic 3: Acceptance Criteria

### Q3.1: What defines "complete" for integration demo?

**Acceptance Criteria**:

| Criterion | Verification |
|-----------|--------------|
| Starts Identity server | Output shows "Started Identity server" |
| Starts KMS server | Output shows "Started KMS server" |
| Health checks pass | Output shows "Service health checks passed" |
| Token obtained | Output shows "Obtained access token" |
| Token validated | Output shows "Token validated" |
| KMS operation works | Output shows "Authenticated KMS operation completed" |
| Audit verified | Output shows "Audit log verified" |
| Exit code 0 | `echo $?` returns 0 |
| No TODOs | `grep -c "TODO" integration.go` returns 0 |
| Lint passes | `golangci-lint run ./internal/cmd/demo/...` returns 0 |

### Q3.2: Sample successful output

```
╔══════════════════════════════════════════════════════════════════╗
║                    Integration Demo - cryptoutil                 ║
╠══════════════════════════════════════════════════════════════════╣
║ This demo shows KMS and Identity server integration              ║
╚══════════════════════════════════════════════════════════════════╝

Step 1/7: Start Identity server... ✅ PASS
  • AuthZ server running on https://127.0.0.1:18080

Step 2/7: Start KMS server... ✅ PASS
  • KMS server running on https://127.0.0.1:18081
  • Token validation configured for https://127.0.0.1:18080

Step 3/7: Service health checks... ✅ PASS
  • Identity: https://127.0.0.1:18080/health - OK
  • KMS: https://127.0.0.1:18081/health - OK

Step 4/7: Obtain access token... ✅ PASS
  • Token endpoint: https://127.0.0.1:18080/oauth2/v1/token
  • Client: demo-client
  • Scopes: demo:all
  • Token type: Bearer
  • Expires in: 3600s

Step 5/7: Validate token... ✅ PASS
  • JWKS endpoint: https://127.0.0.1:18080/oauth2/v1/jwks
  • Signature: Valid
  • Claims: Valid
  • Expiration: Valid

Step 6/7: Perform authenticated KMS operation... ✅ PASS
  • Operation: List keys
  • Authorization: Bearer token
  • Response: 200 OK

Step 7/7: Verify audit log... ✅ PASS
  • Audit entry found for authenticated operation
  • User: demo-client
  • Operation: list_keys

══════════════════════════════════════════════════════════════════
                         Results: 7/7 PASSED
══════════════════════════════════════════════════════════════════
```

---

## Topic 4: Regression Testing

### Q4.1: What existing tests must continue passing?

**Required Passing Tests**:

```bash
# Demo package tests
go test -v ./internal/cmd/demo/...

# Full test suite (may timeout - acceptable)
go test -v -timeout 15m ./...
```

### Q4.2: Coverage requirements

**Target Coverage**:

| Package | Minimum Coverage |
|---------|-----------------|
| `internal/cmd/demo` | 60% |

**Verification**:

```bash
go test -coverprofile=coverage.out ./internal/cmd/demo/...
go tool cover -func=coverage.out | grep total
```

---

## Topic 5: Failure Scenarios

### Q5.1: Expected failure behaviors

| Scenario | Expected Behavior |
|----------|------------------|
| Identity server fails to start | Clear error message, exit code 1 |
| KMS server fails to start | Clear error message, cleanup Identity, exit code 1 |
| Health check timeout | Clear error with URL, exit code 1 |
| Token request fails | Clear error with HTTP status, exit code 1 |
| Token validation fails | Clear error with reason, exit code 1 |
| KMS operation fails | Clear error with HTTP status, exit code 1 |

### Q5.2: Error message format

```
Step X/7: <step name>... ❌ FAIL

Error: <brief description>
  Detail: <specific error>
  URL: <if applicable>
  Status: <if HTTP error>

Suggestion: <how to fix>
```

---

## Sign-Off

**All testing strategies in this document are LOCKED**

- [ ] Reviewed and approved
- [ ] No open questions remain
- [ ] Ready for implementation

**Date**: ____________
**Approved By**: ____________
