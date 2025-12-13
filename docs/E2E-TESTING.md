# End-to-End (E2E) Testing Guide

**Last Updated**: December 12, 2025

## Overview

The cryptoutil project uses comprehensive E2E testing to validate complete workflows across all services (KMS, CA, JOSE, Identity). Tests use real service deployments with Docker Compose, real databases (PostgreSQL, SQLite), and real telemetry infrastructure.

## Test Architecture

### Test Infrastructure (`internal/test/e2e/`)

```
internal/test/e2e/
├── e2e_test.go              # Main test entry point, suite infrastructure
├── fixtures.go              # Test fixtures and shared setup
├── assertions.go            # Custom assertions for service validation
├── http_utils.go            # HTTP client utilities
├── docker_*.go              # Docker health checks and container management
├── log_utils.go             # Logging infrastructure
├── infrastructure.go        # Infrastructure manager
├── test_suite.go            # Base test suite with summary reporting
│
├── oauth_workflow_test.go   # OAuth 2.1 authorization flows (P4.1)
├── kms_workflow_test.go     # KMS encrypt/decrypt/sign/verify workflows (P4.2)
├── ca_workflow_test.go      # CA certificate lifecycle workflows (P4.3)
└── jose_workflow_test.go    # JOSE JWT/JWK/JWE workflows (P4.4)
```

### Test Patterns

All E2E tests follow these patterns:

1. **Suite-based testing** with `testify/suite`
2. **Build tag** `//go:build e2e` to isolate from unit tests
3. **Fixtures and assertions** for shared infrastructure
4. **Real services** via Docker Compose (no mocks for happy path)
5. **Summary reporting** with pass/fail/skip tracking

## Workflow Test Files

### OAuth 2.1 Workflows (`oauth_workflow_test.go`)

**Suite**: `OAuthWorkflowSuite`

**Tests**:

- `TestAuthorizationCodeFlowWithPKCE` - OAuth 2.1 authorization code + PKCE flow
  1. Register OAuth client via AuthZ admin API
  2. Generate PKCE code verifier and challenge (S256)
  3. Build authorization URL with PKCE challenge
  4. Simulate user consent
  5. Exchange authorization code with PKCE verifier for tokens
  6. Validate access token (JWT signature and claims)
  7. Refresh token flow
  8. Revoke tokens

- `TestClientCredentialsFlow` - OAuth 2.1 client credentials grant
  1. Register client with client_secret
  2. Request token using client credentials
  3. Validate access token
  4. Introspect token

**Dependencies**: Identity services (AuthZ, IdP) deployed

**Reference**: `internal/identity/test/e2e/identity_e2e_test.go`

### KMS Workflows (`kms_workflow_test.go`)

**Suite**: `KMSWorkflowSuite`

**Tests**:

- `TestEncryptDecryptWorkflow` - Complete encrypt/decrypt cycle
  1. Create elastic key pool
  2. Generate material key (AES-256-GCM)
  3. Encrypt plaintext data
  4. Decrypt ciphertext
  5. Verify decrypted plaintext matches original
  6. Test with multiple key versions (rotation)
  7. Delete material key

- `TestSignVerifyWorkflow` - Complete sign/verify cycle
  1. Create elastic key pool
  2. Generate material key (ECDSA P-384)
  3. Sign payload
  4. Verify signature
  5. Test signature verification with rotated keys
  6. Test invalid signature detection

- `TestKeyRotationWorkflow` - Key rotation and version management
  1. Create elastic key with initial material key
  2. Encrypt data with version 1
  3. Rotate key (create version 2)
  4. Encrypt new data with version 2
  5. Decrypt old data with version 1 (historical lookup)
  6. Decrypt new data with version 2 (latest)
  7. Verify both decryptions succeed

**Dependencies**: KMS services (cryptoutil-sqlite, cryptoutil-postgres-*)

### CA Workflows (`ca_workflow_test.go`)

**Suite**: `CAWorkflowSuite`

**Tests**:

- `TestCertificateLifecycleWorkflow` - Complete certificate lifecycle
  1. Generate private key and CSR (crypto/x509)
  2. Submit CSR to CA API
  3. Receive issued TLS server certificate
  4. Parse certificate (verify subject, validity, extensions)
  5. Verify certificate signature chain
  6. Revoke certificate via CA API
  7. Fetch CRL from distribution point
  8. Verify revoked certificate appears in CRL

- `TestOCSPWorkflow` - OCSP responder functionality
  1. Issue certificate
  2. Build OCSP request for certificate serial
  3. Query OCSP responder endpoint
  4. Verify OCSP response status (good)
  5. Revoke certificate
  6. Query OCSP responder again
  7. Verify OCSP response status (revoked)
  8. Verify OCSP response signature

- `TestCRLDistributionWorkflow` - CRL generation and distribution
  1. Issue multiple certificates
  2. Revoke subset of certificates
  3. Fetch CRL from distribution point URL
  4. Parse CRL (crypto/x509)
  5. Verify CRL signature
  6. Verify revoked certificates in CRL
  7. Verify non-revoked certificates NOT in CRL
  8. Test CRL update after new revocation

- `TestCertificateProfilesWorkflow` - Different certificate profiles
  1. Issue TLS server certificate (serverAuth EKU)
  2. Issue TLS client certificate (clientAuth EKU)
  3. Issue code signing certificate (codeSigning EKU)
  4. Verify each certificate has correct EKU extensions
  5. Verify key usage extensions match profile
  6. Verify validity periods match profile constraints

**Dependencies**: CA services (ca-sqlite, ca-postgres-*)

### JOSE Workflows (`jose_workflow_test.go`)

**Suite**: `JOSEWorkflowSuite`

**Tests**:

- `TestJWTSignVerifyWorkflow` - JWT signing and verification
  1. Generate JWK (ES384) via JOSE API
  2. Create JWT with standard claims (sub, iss, exp, iat)
  3. Add custom claims (roles, permissions)
  4. Sign JWT using JWK
  5. Verify JWT signature
  6. Validate standard claims (expiration, issuer)
  7. Validate custom claims
  8. Test expired token rejection

- `TestJWKSEndpointWorkflow` - JWKS discovery endpoint
  1. Generate multiple JWKs (ES384, RS256)
  2. Fetch JWKS from `/.well-known/jwks.json`
  3. Verify public keys published correctly
  4. Verify key IDs (kid) match
  5. Verify private keys NOT exposed in JWKS
  6. Use JWKS public keys to verify JWTs

- `TestJWKRotationWorkflow` - JWK rotation and backward compatibility
  1. Generate JWK version 1
  2. Sign JWT with version 1
  3. Rotate to JWK version 2
  4. Sign new JWT with version 2
  5. Verify both JWTs with JWKS endpoint
  6. Verify old JWT still validates (backward compatibility)
  7. Verify new JWTs use version 2 kid
  8. Test JWKS contains both versions during rotation

- `TestJWEEncryptionWorkflow` - JWE encryption and decryption
  1. Generate encryption key (A256GCM)
  2. Create plaintext payload
  3. Encrypt payload as JWE
  4. Verify JWE structure (header, encrypted_key, iv, ciphertext, tag)
  5. Decrypt JWE
  6. Verify decrypted plaintext matches original
  7. Test with different encryption algorithms (A128GCM, A256GCM)

**Dependencies**: JOSE services (jose-server)

## Running E2E Tests

### Prerequisites

1. **Docker and Docker Compose** installed and running
2. **Go 1.25.5+** installed
3. **Services deployed** via Docker Compose

### Local Execution

**Deploy all services**:

```powershell
# Deploy KMS services
docker compose -f ./deployments/compose/compose.yml up -d

# Deploy CA services
docker compose -f ./deployments/ca/compose.yml up -d

# Deploy JOSE services
docker compose -f ./deployments/jose/compose.yml up -d

# Deploy Identity services (for OAuth tests)
docker compose -f ./deployments/identity/compose.yml up -d
```

**Run all E2E tests**:

```powershell
go test -tags=e2e -v -timeout=30m ./internal/test/e2e/
```

**Run specific workflow tests**:

```powershell
# OAuth workflows only
go test -tags=e2e -v -run TestOAuthWorkflow ./internal/test/e2e/

# KMS workflows only
go test -tags=e2e -v -run TestKMSWorkflow ./internal/test/e2e/

# CA workflows only
go test -tags=e2e -v -run TestCAWorkflow ./internal/test/e2e/

# JOSE workflows only
go test -tags=e2e -v -run TestJOSEWorkflow ./internal/test/e2e/
```

**Cleanup**:

```powershell
# Stop all services
docker compose -f ./deployments/compose/compose.yml down -v
docker compose -f ./deployments/ca/compose.yml down -v
docker compose -f ./deployments/jose/compose.yml down -v
docker compose -f ./deployments/identity/compose.yml down -v
```

### CI/CD Execution

E2E tests run automatically in `.github/workflows/ci-e2e.yml` on:

- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`
- Manual trigger via `workflow_dispatch`

**Workflow Steps**:

1. Build Docker images
2. Deploy services (KMS, CA, JOSE, telemetry)
3. Verify service health
4. Run E2E tests with `-tags=e2e`
5. Collect service logs
6. Upload artifacts (logs, reports)
7. Cleanup (stop all services)

## Test Output and Artifacts

### Log Files

E2E tests create timestamped log files:

```
./workflow-reports/e2e/e2e-test-YYYY-MM-DD_HH-MM-SS.log
```

### Service Logs (CI/CD)

CI/CD workflow collects service logs:

```
./workflow-reports/e2e/
├── cryptoutil-sqlite.log       # KMS SQLite instance
├── cryptoutil-postgres-1.log   # KMS PostgreSQL instance 1
├── cryptoutil-postgres-2.log   # KMS PostgreSQL instance 2
├── ca-sqlite.log               # CA SQLite instance
├── jose-server.log             # JOSE server
├── postgres.log                # PostgreSQL database
├── otel-collector.log          # OpenTelemetry Collector
└── grafana-lgtm.log            # Grafana LGTM stack
```

### Summary Reports

Tests generate summary reports with:

- Total steps executed
- Pass/fail/skip counts
- Success rate percentage
- Duration metrics
- Detailed step breakdown with status emoji

## Troubleshooting

### Services Not Ready

**Symptom**: Tests fail with connection refused errors

**Solution**:

1. Verify services are running: `docker compose ps`
2. Check health status: All services should show "healthy"
3. Increase wait times in Docker Compose files if needed
4. Check service logs: `docker compose logs <service-name>`

### Certificate Validation Errors

**Symptom**: TLS handshake failures, certificate verification errors

**Solution**:

1. Use `-k` flag with `curl` for self-signed certificates
2. Tests use `InsecureSkipVerify` for self-signed certificates in dev
3. Verify CA certificates are properly configured

### Test Timeouts

**Symptom**: Tests exceed 30-minute timeout

**Solution**:

1. Run tests with increased timeout: `-timeout=60m`
2. Run specific test files individually
3. Check for deadlocks or blocking operations
4. Verify database connections are not exhausted

### Port Conflicts

**Symptom**: Services fail to start due to port already in use

**Solution**:

1. Check for existing services: `netstat -ano | findstr :8080`
2. Stop conflicting services
3. Modify port mappings in Docker Compose files

## Development Guidelines

### Adding New E2E Tests

1. **Create test file** with `//go:build e2e` tag
2. **Define test suite** extending `suite.Suite`
3. **Use fixtures** from `NewTestFixture(t)`
4. **Use assertions** from `NewServiceAssertions(t, logger)`
5. **Document steps** in TODO comments before implementation
6. **Test locally** before committing
7. **Update this documentation** with new test descriptions

### Test Data Management

- **Use UUIDv7** for unique test data (thread-safe, process-safe)
- **Use magic constants** from `internal/common/magic/` package
- **Isolate data** per test case (no shared state)
- **Clean up** test data in `TearDownSuite()` if needed

### CI/CD Integration

- E2E tests run automatically in `ci-e2e.yml`
- No manual intervention required
- Test failures block merge to main/develop
- Check GitHub Actions logs for detailed failure analysis

## Related Documentation

- **Testing Instructions**: `.github/instructions/01-04.testing.instructions.md`
- **Docker Instructions**: `.github/instructions/02-02.docker.instructions.md`
- **Observability Instructions**: `.github/instructions/02-03.observability.instructions.md`
- **CI/CD Workflow**: `.github/workflows/ci-e2e.yml`
- **Identity E2E Tests**: `internal/identity/test/e2e/` (reference implementation)

## Future Enhancements

- [ ] Add Identity E2E tests to unified suite
- [ ] Add load testing with Gatling (Browser API)
- [ ] Add performance benchmarks for E2E workflows
- [ ] Add chaos engineering tests (service failures, network partitions)
- [ ] Add multi-region deployment testing
- [ ] Add security scanning of E2E traffic (mTLS, certificate validation)

---

For questions or issues, refer to the main [README](../README.md) or check the [test infrastructure code](../internal/test/e2e/).
