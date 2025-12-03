# Passthru3: Requirements Checklist

**Purpose**: Complete traceability matrix - every requirement with manual verification steps
**Updated**: 2025-12-01

---

## How to Use This Checklist

1. Each requirement has a **Verification Command** - run it
2. Each requirement has **Expected Result** - verify you see it
3. Mark checkbox only AFTER running command and seeing expected result
4. Add date and initials when verified

---

## R1: Demo CLI Requirements

### R1.1: KMS Demo CLI

| ID | Requirement | Verification Command | Expected Result | Verified |
|----|-------------|---------------------|-----------------|----------|
| R1.1.1 | KMS demo starts server | `go run ./cmd/demo kms` | "Started KMS server" message | [ ] |
| R1.1.2 | KMS demo health checks pass | `go run ./cmd/demo kms` | "Health checks passed" message | [ ] |
| R1.1.3 | KMS demo operations work | `go run ./cmd/demo kms` | "KMS operations demonstrated" message | [ ] |
| R1.1.4 | KMS demo exits successfully | `go run ./cmd/demo kms` | 4/4 passed, exit code 0 | [ ] |

### R1.2: Identity Demo CLI

| ID | Requirement | Verification Command | Expected Result | Verified |
|----|-------------|---------------------|-----------------|----------|
| R1.2.1 | Identity demo parses config | `go run ./cmd/demo identity` | "Parsed configuration" message | [ ] |
| R1.2.2 | Identity demo starts AuthZ server | `go run ./cmd/demo identity` | "Started Identity AuthZ server" message | [ ] |
| R1.2.3 | Identity demo health checks pass | `go run ./cmd/demo identity` | "Health checks passed" message | [ ] |
| R1.2.4 | Identity demo verifies OpenID config | `go run ./cmd/demo identity` | "OpenID configuration verified" message | [ ] |
| R1.2.5 | Identity demo gets token | `go run ./cmd/demo identity` | "OAuth 2.1 client_credentials flow demonstrated" message | [ ] |
| R1.2.6 | Identity demo exits successfully | `go run ./cmd/demo identity` | 5/5 passed, exit code 0 | [ ] |

### R1.3: Integration Demo CLI

| ID | Requirement | Verification Command | Expected Result | Verified |
|----|-------------|---------------------|-----------------|----------|
| R1.3.1 | Integration demo starts Identity | `go run ./cmd/demo all` | "Started Identity server" message | [ ] |
| R1.3.2 | Integration demo starts KMS | `go run ./cmd/demo all` | "Started KMS server" message | [ ] |
| R1.3.3 | Integration demo waits for services | `go run ./cmd/demo all` | "Service health checks passed" message | [ ] |
| R1.3.4 | Integration demo gets token | `go run ./cmd/demo all` | "Obtained access token" message | [ ] |
| R1.3.5 | Integration demo validates token | `go run ./cmd/demo all` | "Token validated by KMS" message | [ ] |
| R1.3.6 | Integration demo performs KMS operation | `go run ./cmd/demo all` | "Authenticated KMS operation completed" message | [ ] |
| R1.3.7 | Integration demo verifies audit | `go run ./cmd/demo all` | "Audit log verified" message | [ ] |
| R1.3.8 | Integration demo exits successfully | `go run ./cmd/demo all` | 7/7 passed, exit code 0 | [ ] |

---

## R2: Docker Compose Requirements

### R2.1: KMS Docker Compose

| ID | Requirement | Verification Command | Expected Result | Verified |
|----|-------------|---------------------|-----------------|----------|
| R2.1.1 | KMS compose config validates | `docker compose -f deployments/kms/compose.demo.yml --profile demo config` | YAML output, no errors | [ ] |
| R2.1.2 | KMS compose builds | `docker compose -f deployments/kms/compose.demo.yml --profile demo build` | Build completes | [ ] |
| R2.1.3 | KMS compose starts | `docker compose -f deployments/kms/compose.demo.yml --profile demo up -d` | All containers start | [ ] |
| R2.1.4 | KMS services healthy | `docker compose -f deployments/kms/compose.demo.yml --profile demo ps` | All services "healthy" or "running" | [ ] |
| R2.1.5 | KMS Swagger UI accessible | Open `https://localhost:8080/ui/swagger/doc.json` in browser | OpenAPI JSON returned | [ ] |
| R2.1.6 | KMS compose stops cleanly | `docker compose -f deployments/kms/compose.demo.yml --profile demo down -v` | All containers removed | [ ] |

### R2.2: Identity Docker Compose

| ID | Requirement | Verification Command | Expected Result | Verified |
|----|-------------|---------------------|-----------------|----------|
| R2.2.1 | Identity compose config validates | `docker compose -f deployments/identity/compose.simple.yml --profile demo config` | YAML output, no errors | [ ] |
| R2.2.2 | Identity compose builds | `docker compose -f deployments/identity/compose.simple.yml --profile demo build` | Build completes | [ ] |
| R2.2.3 | Identity compose starts | `docker compose -f deployments/identity/compose.simple.yml --profile demo up -d` | All containers start | [ ] |
| R2.2.4 | Identity services healthy | `docker compose -f deployments/identity/compose.simple.yml --profile demo ps` | All services "healthy" or "running" | [ ] |
| R2.2.5 | Identity OpenID config accessible | Open `https://localhost:8082/.well-known/openid-configuration` | OpenID JSON returned | [ ] |
| R2.2.6 | Identity compose stops cleanly | `docker compose -f deployments/identity/compose.simple.yml --profile demo down -v` | All containers removed | [ ] |

### R2.3: Telemetry Docker Compose

| ID | Requirement | Verification Command | Expected Result | Verified |
|----|-------------|---------------------|-----------------|----------|
| R2.3.1 | No port conflicts on Windows | `docker compose -f deployments/telemetry/compose.yml config` | Port 15679 (not 55679) for zPages | [ ] |
| R2.3.2 | OTEL collector starts | Part of compose up | Collector container running | [ ] |
| R2.3.3 | Grafana accessible | Open `http://localhost:3000` | Grafana login page | [ ] |

---

## R3: Code Quality Requirements

| ID | Requirement | Verification Command | Expected Result | Verified |
|----|-------------|---------------------|-----------------|----------|
| R3.1 | All code builds | `go build ./...` | No errors | [ ] |
| R3.2 | All lint passes | `golangci-lint run ./...` | Zero errors | [ ] |
| R3.3 | Demo package lint passes | `golangci-lint run ./internal/cmd/demo/...` | Zero errors | [ ] |
| R3.4 | No TODOs in integration.go | `grep -c "TODO" internal/cmd/demo/integration.go` | Count = 0 | [ ] |

---

## R4: OAuth 2.1 Requirements

| ID | Requirement | Verification Command | Expected Result | Verified |
|----|-------------|---------------------|-----------------|----------|
| R4.1 | Discovery endpoint works | `curl -k https://127.0.0.1:18080/.well-known/openid-configuration` | Valid JSON with issuer, token_endpoint, jwks_uri | [ ] |
| R4.2 | JWKS endpoint works | `curl -k https://127.0.0.1:18080/.well-known/jwks.json` OR `/oauth2/v1/jwks` | Valid JWKS JSON | [ ] |
| R4.3 | Token endpoint works | Use demo-client/demo-secret with client_credentials | 200 OK with access_token | [ ] |
| R4.4 | Token is valid JWT | Decode returned access_token | Valid JWT with claims | [ ] |

---

## R5: Security Requirements

| ID | Requirement | Verification Command | Expected Result | Verified |
|----|-------------|---------------------|-----------------|----------|
| R5.1 | TLS enabled | All endpoints use HTTPS | Connection established | [ ] |
| R5.2 | Demo client uses secrets | Demo-client has client_secret not in URL | Secret in header/body only | [ ] |
| R5.3 | PKCE available | Check discovery endpoint | code_challenge_methods_supported includes S256 | [ ] |

---

## R6: Network Requirements

| ID | Requirement | Verification Command | Expected Result | Verified |
|----|-------------|---------------------|-----------------|----------|
| R6.1 | Services on same Docker network | Check compose config | telemetry-network shared | [ ] |
| R6.2 | OTEL collector reachable | Identity can send telemetry | No connection errors in logs | [ ] |
| R6.3 | Grafana receives data | Check Grafana datasources | Data visible in Grafana | [ ] |

---

## Verification Log

| Date | Verifier | Requirements Verified | Notes |
|------|----------|----------------------|-------|
| | | | |

---

## Sign-Off

- [ ] All R1 requirements verified
- [ ] All R2 requirements verified
- [ ] All R3 requirements verified
- [ ] All R4 requirements verified
- [ ] All R5 requirements verified
- [ ] All R6 requirements verified

**Final Sign-Off Date**: ____________
**Verified By**: ____________
