# Phase 2: Complete Deferred I2 Features Implementation Guide

**Duration**: Days 5-6 (6-8 hours)
**Prerequisites**: Phase 1 complete (all CI/CD workflows passing)
**Status**: ❌ Not Started

## Overview

Phase 2 completes 4 mandatory deferred features from Iteration 2 (I2) that were documented but not fully implemented. These features bring the JOSE and CA APIs to production readiness.

**Task Breakdown**:

- P2.1: JOSE E2E Test Suite (3-4h) - Integration testing for all JOSE endpoints
- P2.2: CA OCSP Responder (2h) - Certificate status checking protocol
- P2.3: JOSE Docker Integration (1-2h) - Multi-instance deployment
- P2.4-P2.7: Already Complete ✅ (EST cacerts, simpleenroll, simplereenroll, TSA timestamp)
- P2.8: EST serverkeygen (OPTIONAL - BLOCKED on PKCS#7 library)

**Note**: Tasks P2.4-P2.7 are already complete, leaving only 3 mandatory tasks: P2.1, P2.2, P2.3

## Task Details

---

### P2.1: JOSE E2E Test Suite ⭐ CRITICAL

**Priority**: HIGH
**Effort**: 3-4 hours
**Status**: ❌ Not Started

**Objective**: Create comprehensive integration tests for all 10 JOSE API endpoints to validate end-to-end functionality.

**Current State**:

- Unit tests exist for individual JOSE components
- No integration tests exercising full API workflow
- No Docker Compose integration for JOSE server testing

**Implementation Strategy**:

```bash
# Step 1: Review existing JOSE API endpoints
# Per api/jose/openapi_spec_paths.yaml:
# - POST /jwk/v1/generate
# - POST /jws/v1/sign
# - POST /jws/v1/verify
# - POST /jwe/v1/encrypt
# - POST /jwe/v1/decrypt
# - POST /jwt/v1/generate
# - POST /jwt/v1/validate
# And additional utility endpoints

# Step 2: Create integration test structure
mkdir -p internal/jose/server
touch internal/jose/server/jwk_integration_test.go
touch internal/jose/server/jws_integration_test.go
touch internal/jose/server/jwe_integration_test.go
touch internal/jose/server/jwt_integration_test.go
```

**Files to Create**:

- `internal/jose/server/jwk_integration_test.go`
- `internal/jose/server/jws_integration_test.go`
- `internal/jose/server/jwe_integration_test.go`
- `internal/jose/server/jwt_integration_test.go`

**Integration Test Pattern**:

```go
// File: internal/jose/server/jwk_integration_test.go
//go:build integration

package server_test

import (
    "testing"
    "github.com/stretchr/testify/require"
    googleUuid "github.com/google/uuid"
)

var (
    testJOSEServer *joseApp.Server
    testBaseURL    string
)

func TestMain(m *testing.M) {
    // Start JOSE server ONCE with in-memory SQLite
    config := &joseConfig.Config{
        BindAddress: "127.0.0.1",
        Port:        0,  // Dynamic port allocation
        Database: &joseConfig.DatabaseConfig{
            Type: "sqlite",
            DSN:  "file::memory:?cache=shared",
        },
    }

    var err error
    testJOSEServer, err = joseApp.NewServer(config)
    if err != nil {
        panic(err)
    }

    go testJOSEServer.Start()

    // Get actual port
    testBaseURL = fmt.Sprintf("https://127.0.0.1:%d", testJOSEServer.ActualPort())

    exitCode := m.Run()

    _ = testJOSEServer.Shutdown()
    os.Exit(exitCode)
}

func TestJWKGeneration_Integration(t *testing.T) {
    t.Parallel()

    // Create HTTP client
    client := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        },
    }

    // Test JWK generation endpoint
    keyID := googleUuid.NewV7()
    reqBody := map[string]interface{}{
        "key_id":    keyID.String(),
        "algorithm": "RS256",
        "key_size":  2048,
    }

    jsonBody, err := json.Marshal(reqBody)
    require.NoError(t, err)

    resp, err := client.Post(
        testBaseURL+"/jwk/v1/generate",
        "application/json",
        bytes.NewReader(jsonBody),
    )
    require.NoError(t, err)
    defer resp.Body.Close()

    require.Equal(t, http.StatusOK, resp.StatusCode)

    // Validate response contains valid JWK
    var jwkResp map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&jwkResp)
    require.NoError(t, err)
    require.NotEmpty(t, jwkResp["kty"])  // Key type
    require.NotEmpty(t, jwkResp["kid"])  // Key ID
}
```

**Acceptance Criteria**:

- ✅ All 10 JOSE API endpoints have integration tests
- ✅ Tests use TestMain pattern with shared server instance
- ✅ Each test uses unique UUIDv7 for data isolation
- ✅ Integration tests complete in <2 minutes
- ✅ Coverage for JOSE server package >95%
- ✅ Tests tagged with `//go:build integration`

**Validation Commands**:

```bash
# Run integration tests
go test -tags=integration ./internal/jose/server -v

# Verify coverage
go test -tags=integration -coverprofile=test-output/coverage_jose_integration.out ./internal/jose/server
go tool cover -func=test-output/coverage_jose_integration.out
```

**Testing Workflow**:

1. JWK endpoint tests: generate RSA, EC, Ed25519 keys
2. JWS endpoint tests: sign and verify with various algorithms
3. JWE endpoint tests: encrypt and decrypt with different key types
4. JWT endpoint tests: generate and validate tokens
5. Error path tests: invalid inputs, missing parameters

---

### P2.2: CA OCSP Responder ⭐ CRITICAL

**Priority**: HIGH
**Effort**: 2 hours
**Status**: ❌ Not Started

**Objective**: Implement RFC 6960 OCSP (Online Certificate Status Protocol) responder for real-time certificate revocation checking.

**Current State**:

- CRL generation implemented (`/ca/v1/crl`)
- No OCSP responder endpoint
- Certificate status tracking exists in database

**Implementation Strategy**:

```bash
# Step 1: Review RFC 6960 OCSP protocol
# - Request: POST /ca/v1/ocsp with DER-encoded OCSP request
# - Response: DER-encoded OCSP response with certificate status

# Step 2: Create OCSP handler
touch internal/ca/handler/ocsp.go
```

**Files to Create/Modify**:

- `internal/ca/handler/ocsp.go` (CREATE)
- `internal/ca/server/routes.go` (MODIFY - add route)
- `internal/ca/handler/ocsp_test.go` (CREATE)

**OCSP Handler Implementation Pattern**:

```go
// File: internal/ca/handler/ocsp.go
package handler

import (
    "crypto/x509"
    "encoding/base64"
    "io"
    "net/http"

    "golang.org/x/crypto/ocsp"
)

type OCSPHandler struct {
    ca *CAService  // Certificate authority service
}

func NewOCSPHandler(ca *CAService) *OCSPHandler {
    return &OCSPHandler{ca: ca}
}

// HandleOCSP processes OCSP requests per RFC 6960
func (h *OCSPHandler) HandleOCSP(w http.ResponseWriter, r *http.Request) {
    // Parse OCSP request from POST body or GET parameter
    var ocspReq *ocsp.Request
    var err error

    if r.Method == http.MethodPost {
        body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "failed to read request", http.StatusBadRequest)
            return
        }
        ocspReq, err = ocsp.ParseRequest(body)
    } else if r.Method == http.MethodGet {
        // GET request with base64-encoded OCSP request in path
        encoded := r.URL.Path[len("/ca/v1/ocsp/"):]
        body, err := base64.StdEncoding.DecodeString(encoded)
        if err != nil {
            http.Error(w, "invalid base64", http.StatusBadRequest)
            return
        }
        ocspReq, err = ocsp.ParseRequest(body)
    }

    if err != nil {
        http.Error(w, "invalid OCSP request", http.StatusBadRequest)
        return
    }

    // Query certificate status from database
    status, revokedAt, reason, err := h.ca.GetCertificateStatus(r.Context(), ocspReq.SerialNumber)
    if err != nil {
        http.Error(w, "failed to check status", http.StatusInternalServerError)
        return
    }

    // Build OCSP response
    resp := ocsp.Response{
        Status:       status,  // ocsp.Good, ocsp.Revoked, or ocsp.Unknown
        SerialNumber: ocspReq.SerialNumber,
        ThisUpdate:   time.Now(),
        NextUpdate:   time.Now().Add(24 * time.Hour),
    }

    if status == ocsp.Revoked {
        resp.RevokedAt = revokedAt
        resp.RevocationReason = reason
    }

    // Sign OCSP response
    respBytes, err := ocsp.CreateResponse(h.ca.IssuerCert, h.ca.ResponderCert, resp, h.ca.ResponderKey)
    if err != nil {
        http.Error(w, "failed to create response", http.StatusInternalServerError)
        return
    }

    // Return DER-encoded OCSP response
    w.Header().Set("Content-Type", "application/ocsp-response")
    w.WriteHeader(http.StatusOK)
    w.Write(respBytes)
}
```

**Acceptance Criteria**:

- ✅ `/ca/v1/ocsp` endpoint accepts POST requests with OCSP requests
- ✅ Returns RFC 6960 compliant OCSP responses
- ✅ Certificate status: good, revoked, unknown
- ✅ Integration with existing revocation database
- ✅ GET method support for base64-encoded requests (optional but recommended)
- ✅ Unit tests cover all status types
- ✅ Coverage ≥95%

**Validation Commands**:

```bash
# Run handler tests
go test ./internal/ca/handler -v -run=TestOCSP

# Test OCSP endpoint manually
openssl ocsp -issuer ca-cert.pem -cert test-cert.pem -url http://127.0.0.1:8080/ca/v1/ocsp -resp_text
```

**Testing Requirements**:

1. Test good certificate status
2. Test revoked certificate status
3. Test unknown certificate status
4. Test invalid OCSP request format
5. Test GET method with base64 encoding

---

### P2.3: JOSE Docker Integration ⭐ CRITICAL

**Priority**: HIGH
**Effort**: 1-2 hours
**Status**: ❌ Not Started

**Objective**: Add JOSE server instances to Docker Compose deployment for multi-instance testing and production-like environment.

**Current State**:

- Docker Compose has PostgreSQL, otel-collector, grafana-otel-lgtm services
- No JOSE server instances configured
- JOSE configs exist in `configs/jose/`

**Implementation Strategy**:

```bash
# Step 1: Review existing compose structure
# Per deployments/compose/compose.yml:
# - cryptoutil-sqlite (port 8080, SQLite backend)
# - cryptoutil-postgres-1 (port 8081, PostgreSQL backend)
# - cryptoutil-postgres-2 (port 8082, PostgreSQL backend)

# Step 2: Add JOSE services following same pattern
# - jose-sqlite (port 8080, SQLite backend)
# - jose-postgres-1 (port 8081, PostgreSQL backend)
# - jose-postgres-2 (port 8082, PostgreSQL backend)
```

**Files to Modify**:

- `deployments/compose/compose.yml` (ADD jose services)
- `deployments/jose/docker-compose.yml` (CREATE - JOSE-specific override)
- `configs/jose/*.yml` (VERIFY configs exist)

**Docker Compose Service Pattern**:

```yaml
# File: deployments/compose/compose.yml (additions)

  jose-sqlite:
    image: cryptoutil/jose:latest
    container_name: jose-sqlite
    build:
      context: ../..
      dockerfile: deployments/jose/Dockerfile
      args:
        GO_VERSION: ${GO_VERSION}
        VCS_REF: ${VCS_REF:-unknown}
        BUILD_DATE: ${BUILD_DATE:-unknown}
    ports:
      - "8080:8080"  # Public API
      - "9090:9090"  # Admin API
    secrets:
      - database_url_secret
      - unseal_secret_1
      - unseal_secret_2
      - unseal_secret_3
    command:
      - "jose-server"
      - "--config=/app/configs/jose/jose-sqlite.yml"
      - "--database-url=file:///run/secrets/database_url_secret"
      - "--dev"
    depends_on:
      postgres:
        condition: service_healthy
      otel-collector:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/livez"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 20s
    networks:
      - cryptoutil-network

  jose-postgres-1:
    image: cryptoutil/jose:latest
    container_name: jose-postgres-1
    build:
      context: ../..
      dockerfile: deployments/jose/Dockerfile
    ports:
      - "8081:8080"
      - "9091:9090"
    secrets:
      - database_url_secret
      - unseal_secret_1
      - unseal_secret_2
      - unseal_secret_3
    command:
      - "jose-server"
      - "--config=/app/configs/jose/jose-postgresql-1.yml"
      - "--database-url=file:///run/secrets/database_url_secret"
    depends_on:
      postgres:
        condition: service_healthy
      otel-collector:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/livez"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 20s
    networks:
      - cryptoutil-network

  jose-postgres-2:
    image: cryptoutil/jose:latest
    container_name: jose-postgres-2
    build:
      context: ../..
      dockerfile: deployments/jose/Dockerfile
    ports:
      - "8082:8080"
      - "9092:9090"
    secrets:
      - database_url_secret
      - unseal_secret_1
      - unseal_secret_2
      - unseal_secret_3
    command:
      - "jose-server"
      - "--config=/app/configs/jose/jose-postgresql-2.yml"
      - "--database-url=file:///run/secrets/database_url_secret"
    depends_on:
      postgres:
        condition: service_healthy
      otel-collector:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "--no-check-certificate", "-q", "-O", "/dev/null", "https://127.0.0.1:9090/livez"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 20s
    networks:
      - cryptoutil-network
```

**Configuration Files to Create/Verify**:

```bash
# Verify config files exist
ls configs/jose/jose-sqlite.yml
ls configs/jose/jose-postgresql-1.yml
ls configs/jose/jose-postgresql-2.yml

# If missing, create from template
cp configs/jose/jose-common.yml configs/jose/jose-sqlite.yml
# Edit jose-sqlite.yml to set database type, CORS origins, etc.
```

**Acceptance Criteria**:

- ✅ Three JOSE services added: jose-sqlite, jose-postgres-1, jose-postgres-2
- ✅ Services use ports 8080-8082 for public API
- ✅ Services use ports 9090-9092 for admin API
- ✅ Health checks use `wget` with HTTPS loopback (127.0.0.1)
- ✅ All services start successfully
- ✅ All services pass health checks within 30 seconds
- ✅ Services can be queried via Swagger UI

**Validation Commands**:

```bash
# Start all services
docker compose -f ./deployments/compose/compose.yml up -d

# Check service status
docker compose -f ./deployments/compose/compose.yml ps

# Verify health checks
docker compose -f ./deployments/compose/compose.yml logs jose-sqlite
docker compose -f ./deployments/compose/compose.yml logs jose-postgres-1
docker compose -f ./deployments/compose/compose.yml logs jose-postgres-2

# Test Swagger endpoints (Windows PowerShell)
Invoke-WebRequest -Uri https://localhost:8080/ui/swagger/doc.json -SkipCertificateCheck
Invoke-WebRequest -Uri https://localhost:8081/ui/swagger/doc.json -SkipCertificateCheck
Invoke-WebRequest -Uri https://localhost:8082/ui/swagger/doc.json -SkipCertificateCheck

# Shutdown
docker compose -f ./deployments/compose/compose.yml down -v
```

**Docker Network Reminders**:

- Per 02-02.docker.instructions.md: ALWAYS use `127.0.0.1` in containers (not `localhost`)
- Use `wget` for health checks (available in Alpine), not `curl`
- Health checks must use `--no-check-certificate` for self-signed TLS certs

---

### P2.4-P2.7: Already Complete ✅

These tasks were marked as deferred in I2 but have since been completed:

| Task | Feature | Implementation | Status |
|------|---------|----------------|--------|
| P2.4 | EST cacerts | `internal/ca/handler/est_cacerts.go` | ✅ Complete |
| P2.5 | EST simpleenroll | `internal/ca/handler/est_simpleenroll.go` | ✅ Complete |
| P2.6 | EST simplereenroll | `internal/ca/handler/est_simplereenroll.go` | ✅ Complete |
| P2.7 | TSA timestamp | `internal/ca/handler/tsa_timestamp.go` | ✅ Complete |

**Validation**:

```bash
# Verify implementations exist
ls internal/ca/handler/est_*.go
ls internal/ca/handler/tsa_*.go
```

No action required for P2.4-P2.7.

---

### P2.8: EST serverkeygen (OPTIONAL - BLOCKED)

**Priority**: LOW
**Effort**: 3-4 hours (if PKCS#7 library issue resolved)
**Status**: ⚠️ BLOCKED on PKCS#7 library compatibility

**Blocker Details**:

The EST serverkeygen feature requires PKCS#7/CMS envelope support. The go.mozilla.org/pkcs7 library has compatibility issues with modern Go versions and may not support all required PKCS#7 operations for RFC 7030 compliance.

**DECISION**: Mark as OPTIONAL and DEFER until:

1. Suitable PKCS#7 library identified, OR
2. Custom PKCS#7 implementation created, OR
3. Feature explicitly requested by stakeholder

**No action required for P2.8 in this iteration.**

---

## Progress Tracking

After completing each task, update `PROGRESS.md`:

```bash
# Edit PROGRESS.md to mark task complete
# Update executive summary percentages
# Commit and push
git add specs/001-cryptoutil/PROGRESS.md
git commit -m "docs(speckit): mark P2.X complete"
git push
```

## Validation Checklist

Before marking Phase 2 complete, verify:

- [ ] P2.1: JOSE integration tests passing, coverage >95%
- [ ] P2.2: OCSP endpoint responding to requests, all status types tested
- [ ] P2.3: All 3 JOSE services starting and healthy in Docker Compose
- [ ] P2.4-P2.7: Verified as already complete
- [ ] P2.8: Acknowledged as OPTIONAL/BLOCKED
- [ ] PROGRESS.md updated with all P2.1-P2.3 marked complete
- [ ] All CI/CD workflows still passing (regression check)

## Next Phase

After Phase 2 complete:

- Proceed to Phase 3: Achieve Coverage Targets
- Use PHASE3-IMPLEMENTATION.md guide
- Update PROGRESS.md executive summary
