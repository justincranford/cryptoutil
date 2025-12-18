# Product Architecture Refactoring - Task List

**Date**: 2025-12-17
**Objective**: Refactor project to align with 4-product architecture (Identity, SM, PKI, JOSE)

---

## Architecture Overview

### Current State Analysis

**Current Structure** (misaligned):

- `internal/kms/` - Should be `internal/sm/kms/`
- `internal/jose/` - Should be `internal/jose/ja/` (JOSE Appliance service)
- `internal/identity/` - Already aligned but services not separated
- `internal/ca/` - Should be `internal/pki/ca/`
- `cmd/cryptoutil/` - Suite-level executable exists
- `cmd/*-server/` - Product/service-level executables missing structure

**Target Structure** (per DIR-STRUCTURE-NOTES.txt):

- **Suite**: cryptoutil (1 executable for all 4 products)
- **Products** (4): identity, sm, pki, jose
- **Services** (8): identity-authz, identity-idp, identity-rs, identity-rp, identity-spa, sm-kms, pki-ca, jose-ja

### Key Architectural Gaps

1. **Product/Service Hierarchy**: Current code doesn't reflect product→service nesting
2. **Private PKI for Inter-Service Auth**: No implementation of suite→product→service CA hierarchy
3. **Executable Organization**: Missing product-level and service-level executables
4. **Service Naming**: Inconsistent (kms vs sm-kms, ca vs pki-ca, jose vs jose-ja)
5. **Deployment Structure**: Docker Compose doesn't reflect product grouping
6. **API Structure**: OpenAPI specs don't reflect product grouping

---

## Phase 1: Foundation - Directory Structure (Steps 1-20)

### Step 1-5: Create Target Directory Structure

- [ ] **Step 1**: Create `internal/app/` directory structure
  - [ ] `internal/app/main.go` (registry of 4 products)
  - [ ] `internal/app/identity/main.go` (registry of 5 services)
  - [ ] `internal/app/sm/main.go` (registry of 1 service)
  - [ ] `internal/app/pki/main.go` (registry of 1 service)
  - [ ] `internal/app/jose/main.go` (registry of 1 service)

- [ ] **Step 2**: Create service directories under products
  - [ ] `internal/app/identity/authz/main.go`
  - [ ] `internal/app/identity/idp/main.go`
  - [ ] `internal/app/identity/rs/main.go`
  - [ ] `internal/app/identity/rp/main.go`
  - [ ] `internal/app/identity/spa/main.go`
  - [ ] `internal/app/sm/kms/main.go`
  - [ ] `internal/app/pki/ca/main.go`
  - [ ] `internal/app/jose/ja/main.go`

- [ ] **Step 3**: Create cmd/ structure per DIR-STRUCTURE-NOTES.txt
  - [ ] `cmd/dev/cicd/main.go` (move from cmd/cicd)
  - [ ] `cmd/dev/demo/main.go` (move from cmd/demo)
  - [ ] `cmd/dev/workflow/main.go` (move from cmd/workflow)
  - [ ] `cmd/dev/e2e/main.go` (create new)

- [ ] **Step 4**: Create product-level executables
  - [ ] `cmd/product/identity/main.go`
  - [ ] `cmd/product/sm/main.go`
  - [ ] `cmd/product/pki/main.go`
  - [ ] `cmd/product/jose/main.go`

- [ ] **Step 5**: Create service-level executables
  - [ ] `cmd/service/identity-authz/main.go`
  - [ ] `cmd/service/identity-idp/main.go`
  - [ ] `cmd/service/identity-rs/main.go`
  - [ ] `cmd/service/identity-rp/main.go`
  - [ ] `cmd/service/identity-spa/main.go`
  - [ ] `cmd/service/sm-kms/main.go`
  - [ ] `cmd/service/pki-ca/main.go`
  - [ ] `cmd/service/jose-ja/main.go`

### Step 6-10: Runtime Infrastructure

- [ ] **Step 6**: Create `internal/runtime/` package
  - [ ] `internal/runtime/command/start/` (service startup)
  - [ ] `internal/runtime/command/stop/` (service shutdown via admin API)
  - [ ] `internal/runtime/command/client/` (service client invocation)
  - [ ] `internal/runtime/command/readyz/` (deep health check)
  - [ ] `internal/runtime/command/livez/` (lite health check)

- [ ] **Step 7**: Create transport layer
  - [ ] `internal/runtime/transport/public_https.go` (0.0.0.0:8080 in container, configurable externally)
  - [ ] `internal/runtime/transport/admin_https.go` (127.0.0.1:9090 admin-only)
  - [ ] `internal/runtime/transport/middleware/` (shared middleware)

- [ ] **Step 8**: Create execution context
  - [ ] `internal/runtime/execution/context.go` (service execution context)
  - [ ] `internal/runtime/execution/runner.go` (service runner orchestration)

- [ ] **Step 9**: Create deployment subcommands
  - [ ] `internal/runtime/command/deployment/initsuiteca.go` (suite root CA)
  - [ ] `internal/runtime/command/deployment/initproductca.go` (product sub CA)
  - [ ] `internal/runtime/command/deployment/addserviceca.go` (service sub CA)

- [ ] **Step 10**: Service registry infrastructure
  - [ ] `internal/app/registry.go` (product/service registry interface)
  - [ ] Product registry implementations in each `internal/app/<product>/main.go`

### Step 11-15: Private PKI Infrastructure

- [ ] **Step 11**: Suite-level CA management
  - [ ] `internal/shared/crypto/pki/suite_ca.go` (suite root CA creation/storage)
  - [ ] `internal/shared/crypto/pki/suite_ca_store.go` (file/secret storage backend)

- [ ] **Step 12**: Product-level CA management
  - [ ] `internal/shared/crypto/pki/product_ca.go` (product sub CA issuance)
  - [ ] `internal/shared/crypto/pki/product_ca_store.go` (product CA storage)

- [ ] **Step 13**: Service-level CA management
  - [ ] `internal/shared/crypto/pki/service_ca.go` (service sub CA issuance)
  - [ ] `internal/shared/crypto/pki/service_ca_store.go` (service CA storage)

- [ ] **Step 14**: TLS server cert management
  - [ ] `internal/shared/crypto/pki/server_cert.go` (TLS server cert issuance from service CA)
  - [ ] Integrate with existing `internal/shared/crypto/certificate/` utilities

- [ ] **Step 15**: TLS client cert management
  - [ ] `internal/shared/crypto/pki/client_cert.go` (TLS client cert issuance from service CA)
  - [ ] Client cert validation in service middleware

### Step 16-20: Configuration & API Structure

- [ ] **Step 16**: Create product-specific config structures
  - [ ] `configs/product/identity/` (identity product configs)
  - [ ] `configs/product/sm/` (sm product configs)
  - [ ] `configs/product/pki/` (pki product configs)
  - [ ] `configs/product/jose/` (jose product configs)

- [ ] **Step 17**: Create service-specific config structures
  - [ ] `configs/service/identity-authz/`
  - [ ] `configs/service/identity-idp/`
  - [ ] `configs/service/identity-rs/`
  - [ ] `configs/service/identity-rp/`
  - [ ] `configs/service/identity-spa/`
  - [ ] `configs/service/sm-kms/`
  - [ ] `configs/service/pki-ca/`
  - [ ] `configs/service/jose-ja/`

- [ ] **Step 18**: Refactor API structure
  - [ ] `api/product/identity/` (identity product OpenAPI specs)
  - [ ] `api/product/sm/` (sm product OpenAPI specs)
  - [ ] `api/product/pki/` (pki product OpenAPI specs)
  - [ ] `api/product/jose/` (jose product OpenAPI specs)

- [ ] **Step 19**: Create deployment structure
  - [ ] `deployments/product/identity/compose.yml` (identity product services)
  - [ ] `deployments/product/sm/compose.yml` (sm product services)
  - [ ] `deployments/product/pki/compose.yml` (pki product services)
  - [ ] `deployments/product/jose/compose.yml` (jose product services)

- [ ] **Step 20**: Suite-level deployment orchestration
  - [ ] `deployments/suite/compose.yml` (all 4 products, includes suite CA init)
  - [ ] `deployments/suite/docker-compose.override.yml` (local dev overrides)

---

## Phase 2: Code Migration - SM (Secret Management) Product (Steps 21-30)

### Step 21-25: Migrate KMS to SM Product

- [ ] **Step 21**: Create SM product structure
  - [ ] Copy `internal/kms/` → `internal/sm/kms/`
  - [ ] Update package names: `package kms` → `package kms` (keep same, parent is `sm`)
  - [ ] Update import paths: `cryptoutil/internal/kms` → `cryptoutil/internal/sm/kms`

- [ ] **Step 22**: Update KMS service imports project-wide
  - [ ] Find all `import "cryptoutil/internal/kms"` references
  - [ ] Replace with `import cryptoutilSmKms "cryptoutil/internal/sm/kms"`
  - [ ] Update all call sites to use `cryptoutilSmKms` alias

- [ ] **Step 23**: Create SM product main.go
  - [ ] `internal/app/sm/main.go` with service registry for KMS
  - [ ] Implement product-level CLI: `sm deployment <subcommand>`, `sm kms <subcommand>`

- [ ] **Step 24**: Create SM KMS service main.go
  - [ ] `internal/app/sm/kms/main.go` with service-specific logic
  - [ ] Wire up to runtime commands (up, down, live, ready, client)

- [ ] **Step 25**: Update SM KMS executables
  - [ ] `cmd/product/sm/main.go` (SM product executable)
  - [ ] `cmd/service/sm-kms/main.go` (KMS service executable)
  - [ ] Update `cmd/cryptoutil/main.go` to register SM product

### Step 26-30: SM Product API & Deployment

- [ ] **Step 26**: Refactor SM KMS API specs
  - [ ] Move `api/kms/` → `api/product/sm/kms/`
  - [ ] Update OpenAPI spec with product context
  - [ ] Regenerate API client/server code

- [ ] **Step 27**: Create SM product configs
  - [ ] `configs/product/sm/product.yml` (product-level settings)
  - [ ] `configs/service/sm-kms/service.yml` (KMS service settings)
  - [ ] Merge with existing `configs/kms/` settings

- [ ] **Step 28**: Create SM product deployment
  - [ ] `deployments/product/sm/compose.yml` (KMS service + dependencies)
  - [ ] PostgreSQL service for KMS
  - [ ] Telemetry sidecar

- [ ] **Step 29**: Update SM tests
  - [ ] Update test imports: `internal/kms` → `internal/sm/kms`
  - [ ] Run tests: `go test ./internal/sm/kms/...`
  - [ ] Fix integration tests in `internal/test/e2e/`

- [ ] **Step 30**: Remove old KMS structure
  - [ ] Delete `internal/kms/` after validation
  - [ ] Delete `cmd/kms-server/` (replaced by `cmd/service/sm-kms/`)
  - [ ] Clean up old configs/deployments

---

## Phase 3: Code Migration - PKI Product (Steps 31-40)

### Step 31-35: Migrate CA to PKI Product

- [ ] **Step 31**: Create PKI product structure
  - [ ] Copy `internal/ca/` → `internal/pki/ca/`
  - [ ] Update package names: `package ca` → `package ca` (keep same, parent is `pki`)
  - [ ] Update import paths: `cryptoutil/internal/ca` → `cryptoutil/internal/pki/ca`

- [ ] **Step 32**: Update CA service imports project-wide
  - [ ] Find all `import "cryptoutil/internal/ca"` references
  - [ ] Replace with `import cryptoutilPkiCa "cryptoutil/internal/pki/ca"`
  - [ ] Update all call sites to use `cryptoutilPkiCa` alias

- [ ] **Step 33**: Create PKI product main.go
  - [ ] `internal/app/pki/main.go` with service registry for CA
  - [ ] Implement product-level CLI: `pki deployment <subcommand>`, `pki ca <subcommand>`

- [ ] **Step 34**: Create PKI CA service main.go
  - [ ] `internal/app/pki/ca/main.go` with service-specific logic
  - [ ] Wire up to runtime commands (up, down, live, ready, client)

- [ ] **Step 35**: Update PKI CA executables
  - [ ] `cmd/product/pki/main.go` (PKI product executable)
  - [ ] `cmd/service/pki-ca/main.go` (CA service executable)
  - [ ] Update `cmd/cryptoutil/main.go` to register PKI product

### Step 36-40: PKI Product API & Deployment

- [ ] **Step 36**: Refactor PKI CA API specs
  - [ ] Move `api/ca/` → `api/product/pki/ca/`
  - [ ] Update OpenAPI spec with product context
  - [ ] Regenerate API client/server code

- [ ] **Step 37**: Create PKI product configs
  - [ ] `configs/product/pki/product.yml` (product-level settings)
  - [ ] `configs/service/pki-ca/service.yml` (CA service settings)
  - [ ] Merge with existing `configs/ca/` settings

- [ ] **Step 38**: Create PKI product deployment
  - [ ] `deployments/product/pki/compose.yml` (CA service + dependencies)
  - [ ] Database service for CA (if needed)
  - [ ] Telemetry sidecar

- [ ] **Step 39**: Update PKI tests
  - [ ] Update test imports: `internal/ca` → `internal/pki/ca`
  - [ ] Run tests: `go test ./internal/pki/ca/...`
  - [ ] Fix integration tests in `internal/test/e2e/`

- [ ] **Step 40**: Remove old CA structure
  - [ ] Delete `internal/ca/` after validation
  - [ ] Delete `cmd/ca-server/` (replaced by `cmd/service/pki-ca/`)
  - [ ] Clean up old configs/deployments

---

## Phase 4: Code Migration - JOSE Product (Steps 41-50)

### Step 41-45: Migrate JOSE to JOSE Product with JA Service

- [ ] **Step 41**: Create JOSE product structure
  - [ ] Copy `internal/jose/` → `internal/jose/ja/` (JA = JOSE Appliance)
  - [ ] Update package names: `package jose` → `package ja`
  - [ ] Update import paths: `cryptoutil/internal/jose` → `cryptoutil/internal/jose/ja`

- [ ] **Step 42**: Update JOSE service imports project-wide
  - [ ] Find all `import "cryptoutil/internal/jose"` references
  - [ ] Replace with `import cryptoutilJoseJa "cryptoutil/internal/jose/ja"`
  - [ ] Update all call sites to use `cryptoutilJoseJa` alias

- [ ] **Step 43**: Create JOSE product main.go
  - [ ] `internal/app/jose/main.go` with service registry for JA
  - [ ] Implement product-level CLI: `jose deployment <subcommand>`, `jose ja <subcommand>`

- [ ] **Step 44**: Create JOSE JA service main.go
  - [ ] `internal/app/jose/ja/main.go` with service-specific logic
  - [ ] Wire up to runtime commands (up, down, live, ready, client)

- [ ] **Step 45**: Update JOSE JA executables
  - [ ] `cmd/product/jose/main.go` (JOSE product executable)
  - [ ] `cmd/service/jose-ja/main.go` (JA service executable)
  - [ ] Update `cmd/cryptoutil/main.go` to register JOSE product

### Step 46-50: JOSE Product API & Deployment

- [ ] **Step 46**: Refactor JOSE JA API specs
  - [ ] Move `api/jose/` → `api/product/jose/ja/`
  - [ ] Update OpenAPI spec with product context
  - [ ] Regenerate API client/server code

- [ ] **Step 47**: Create JOSE product configs
  - [ ] `configs/product/jose/product.yml` (product-level settings)
  - [ ] `configs/service/jose-ja/service.yml` (JA service settings)
  - [ ] Merge with existing `configs/jose/` settings

- [ ] **Step 48**: Create JOSE product deployment
  - [ ] `deployments/product/jose/compose.yml` (JA service + dependencies)
  - [ ] Database service for JA (if needed)
  - [ ] Telemetry sidecar

- [ ] **Step 49**: Update JOSE tests
  - [ ] Update test imports: `internal/jose` → `internal/jose/ja`
  - [ ] Run tests: `go test ./internal/jose/ja/...`
  - [ ] Fix integration tests in `internal/test/e2e/`

- [ ] **Step 50**: Remove old JOSE structure
  - [ ] Delete top-level `internal/jose/` crypto libs (keep as `internal/shared/crypto/jose/`)
  - [ ] Delete `cmd/jose-server/` (replaced by `cmd/service/jose-ja/`)
  - [ ] Clean up old configs/deployments

---

## Phase 5: Code Migration - Identity Product (Steps 51-65)

### Step 51-55: Migrate Identity Services to Product Structure

- [ ] **Step 51**: Create Identity product structure
  - [ ] Keep `internal/identity/` as product root
  - [ ] Create service subdirectories:
    - [ ] `internal/identity/authz/` (keep existing, add main.go)
    - [ ] `internal/identity/idp/` (keep existing, add main.go)
    - [ ] `internal/identity/rs/` (create new - Resource Server)
    - [ ] `internal/identity/rp/` (create new - Relying Party)
    - [ ] `internal/identity/spa/` (create new - Single Page App)

- [ ] **Step 52**: Create Identity service main.go files
  - [ ] `internal/app/identity/authz/main.go` (authz service logic)
  - [ ] `internal/app/identity/idp/main.go` (idp service logic)
  - [ ] `internal/app/identity/rs/main.go` (rs service logic)
  - [ ] `internal/app/identity/rp/main.go` (rp service logic)
  - [ ] `internal/app/identity/spa/main.go` (spa service logic)

- [ ] **Step 53**: Create Identity product main.go
  - [ ] `internal/app/identity/main.go` with service registry for 5 services
  - [ ] Implement product-level CLI: `identity deployment <subcommand>`, `identity <service> <subcommand>`

- [ ] **Step 54**: Update Identity service executables
  - [ ] `cmd/product/identity/main.go` (Identity product executable)
  - [ ] `cmd/service/identity-authz/main.go`
  - [ ] `cmd/service/identity-idp/main.go`
  - [ ] `cmd/service/identity-rs/main.go`
  - [ ] `cmd/service/identity-rp/main.go`
  - [ ] `cmd/service/identity-spa/main.go`
  - [ ] Update `cmd/cryptoutil/main.go` to register Identity product

- [ ] **Step 55**: Implement RS (Resource Server) service
  - [ ] `internal/identity/rs/server/` (REST API handlers)
  - [ ] `internal/identity/rs/middleware/` (access token validation)
  - [ ] `internal/identity/rs/domain/` (resource models)
  - [ ] Integration with authz service for token introspection

### Step 56-60: Implement RP (Relying Party) Service

- [ ] **Step 56**: Implement RP (Relying Party) service
  - [ ] `internal/identity/rp/server/` (REST API handlers)
  - [ ] `internal/identity/rp/oidc/` (OIDC client implementation)
  - [ ] `internal/identity/rp/session/` (session management)
  - [ ] Integration with authz and idp services

- [ ] **Step 57**: Implement SPA (Single Page App) service
  - [ ] `internal/identity/spa/server/` (static file serving + API)
  - [ ] `internal/identity/spa/ui/` (frontend assets)
  - [ ] `internal/identity/spa/oidc/` (OIDC flow implementation)
  - [ ] Integration with authz service for authorization code + PKCE flow

- [ ] **Step 58**: Refactor Identity API specs by service
  - [ ] Move `api/authz/` → `api/product/identity/authz/`
  - [ ] Move `api/idp/` → `api/product/identity/idp/`
  - [ ] Create `api/product/identity/rs/` (new)
  - [ ] Create `api/product/identity/rp/` (new)
  - [ ] Create `api/product/identity/spa/` (new)
  - [ ] Regenerate all API client/server code

- [ ] **Step 59**: Create Identity product configs
  - [ ] `configs/product/identity/product.yml` (shared settings across services)
  - [ ] `configs/service/identity-authz/service.yml`
  - [ ] `configs/service/identity-idp/service.yml`
  - [ ] `configs/service/identity-rs/service.yml`
  - [ ] `configs/service/identity-rp/service.yml`
  - [ ] `configs/service/identity-spa/service.yml`

- [ ] **Step 60**: Create Identity product deployment
  - [ ] `deployments/product/identity/compose.yml` (all 5 services)
  - [ ] PostgreSQL service for Identity (shared by authz/idp)
  - [ ] Telemetry sidecar for each service
  - [ ] Service discovery/networking between 5 services

### Step 61-65: Identity Product Tests & Cleanup

- [ ] **Step 61**: Update Identity authz tests
  - [ ] Ensure imports use product path: `internal/identity/authz`
  - [ ] Run tests: `go test ./internal/identity/authz/...`
  - [ ] Update integration tests

- [ ] **Step 62**: Update Identity idp tests
  - [ ] Ensure imports use product path: `internal/identity/idp`
  - [ ] Run tests: `go test ./internal/identity/idp/...`
  - [ ] Update integration tests

- [ ] **Step 63**: Create Identity rs/rp/spa tests
  - [ ] Create unit tests for rs service
  - [ ] Create unit tests for rp service
  - [ ] Create unit tests for spa service
  - [ ] Target 95%+ coverage for all 3 new services

- [ ] **Step 64**: Create Identity E2E tests
  - [ ] `internal/test/e2e/identity/` (E2E tests for all 5 services)
  - [ ] Test authorization code + PKCE flow (spa → authz → idp)
  - [ ] Test client credentials flow (rs → authz)
  - [ ] Test token introspection (rs → authz)
  - [ ] Test OIDC discovery (rp → authz)

- [ ] **Step 65**: Remove old Identity structure
  - [ ] Delete `cmd/identity-compose/` (replaced by deployments/product/identity/)
  - [ ] Delete `cmd/identity-demo/` (replaced by cmd/dev/demo-identity/)
  - [ ] Delete `cmd/identity-unified/` (replaced by cmd/product/identity/)
  - [ ] Clean up old configs

---

## Phase 6: Shared Libraries Consolidation (Steps 66-75)

### Step 66-70: Consolidate Crypto Libraries

- [ ] **Step 66**: Move JOSE crypto to shared
  - [ ] Verify `internal/shared/crypto/jose/` exists (should already be there)
  - [ ] Ensure no duplication with `internal/jose/ja/crypto/`
  - [ ] Update all imports to use `internal/shared/crypto/jose`

- [ ] **Step 67**: Consolidate certificate utilities
  - [ ] Ensure `internal/shared/crypto/certificate/` is product-agnostic
  - [ ] Move any product-specific logic to product packages
  - [ ] Update import aliases: `cryptoutilCert "cryptoutil/internal/shared/crypto/certificate"`

- [ ] **Step 68**: Consolidate TLS utilities
  - [ ] Ensure `internal/shared/crypto/tls/` is product-agnostic
  - [ ] Add suite→product→service CA chain validation
  - [ ] Update import aliases: `cryptoutilTls "cryptoutil/internal/shared/crypto/tls"`

- [ ] **Step 69**: Consolidate hashing utilities
  - [ ] Ensure `internal/shared/crypto/hash/` supports all FIPS-approved algorithms
  - [ ] Add algorithm registry for product-specific needs
  - [ ] Update import aliases: `cryptoutilHash "cryptoutil/internal/shared/crypto/hash"`

- [ ] **Step 70**: Consolidate key generation
  - [ ] Ensure `internal/shared/crypto/keygen/` is product-agnostic
  - [ ] Support RSA, ECDSA, EdDSA, AES, HMAC key generation
  - [ ] Update import aliases: `cryptoutilKeygen "cryptoutil/internal/shared/crypto/keygen"`

### Step 71-75: Consolidate Repository/Observability

- [ ] **Step 71**: Create shared repository patterns
  - [ ] `internal/shared/repository/` (base interfaces for GORM repositories)
  - [ ] Product-specific implementations remain in product packages
  - [ ] Shared transaction management patterns

- [ ] **Step 72**: Consolidate observability
  - [ ] Ensure `internal/shared/observability/` supports all products
  - [ ] Add product/service tags to all traces/metrics/logs
  - [ ] Update OTLP exporters with product context

- [ ] **Step 73**: Create shared middleware
  - [ ] `internal/runtime/transport/middleware/logging.go` (request logging)
  - [ ] `internal/runtime/transport/middleware/tracing.go` (OpenTelemetry tracing)
  - [ ] `internal/runtime/transport/middleware/metrics.go` (Prometheus metrics)
  - [ ] `internal/runtime/transport/middleware/auth.go` (mTLS validation)

- [ ] **Step 74**: Create shared config loading
  - [ ] `internal/shared/config/loader.go` (YAML config loading)
  - [ ] Support product/service hierarchy: suite → product → service
  - [ ] Environment variable overrides with namespacing

- [ ] **Step 75**: Consolidate validation utilities
  - [ ] Ensure `internal/shared/util/validation/` is product-agnostic
  - [ ] Add product-specific validators where needed
  - [ ] Update import aliases: `cryptoutilValidation "cryptoutil/internal/shared/util/validation"`

---

## Phase 7: Suite-Level Orchestration (Steps 76-85)

### Step 76-80: Suite CA Hierarchy Implementation

- [ ] **Step 76**: Implement `cryptoutil suite deployment initsuiteca`
  - [ ] Generate suite root CA (RSA 4096, 10-year validity)
  - [ ] Store in `configs/suite/ca/root-ca.crt` and secret storage
  - [ ] Create suite CA metadata (serial numbers, CRL distribution points)

- [ ] **Step 77**: Implement `cryptoutil suite deployment addproductca`
  - [ ] Issue product sub CA from suite root CA
  - [ ] Store in `configs/product/<product>/ca/product-ca.crt`
  - [ ] Add to suite CA registry

- [ ] **Step 78**: Implement `cryptoutil <product> deployment initproductca`
  - [ ] Verify suite root CA exists
  - [ ] Issue product sub CA from suite root CA
  - [ ] Initialize product-level CA operations

- [ ] **Step 79**: Implement `cryptoutil <product> deployment addserviceca`
  - [ ] Issue service sub CA from product sub CA
  - [ ] Store in `configs/service/<product>-<service>/ca/service-ca.crt`
  - [ ] Add to product CA registry

- [ ] **Step 80**: Implement service TLS cert issuance
  - [ ] `cryptoutil <product> <service> up` auto-generates TLS server cert from service CA
  - [ ] Server cert includes SPIFFE ID: `spiffe://cryptoutil/site/<site>/product/<product>/service/<service>`
  - [ ] Client cert validation in all service middleware

### Step 81-85: Suite Deployment & Orchestration

- [ ] **Step 81**: Create suite-level docker-compose
  - [ ] `deployments/suite/compose.yml` includes all 4 products
  - [ ] Init containers for CA hierarchy setup
  - [ ] Service dependencies (authz before idp, kms before others)

- [ ] **Step 82**: Implement service discovery
  - [ ] Docker Compose service names: `<product>-<service>` (e.g., `identity-authz`)
  - [ ] Environment variables for inter-service URLs
  - [ ] Health check dependencies in compose.yml

- [ ] **Step 83**: Implement mTLS between services
  - [ ] All inter-service calls use mTLS with service CA chain validation
  - [ ] SPIFFE ID validation in service middleware
  - [ ] Client cert generation for each service

- [ ] **Step 84**: Create suite-level E2E tests
  - [ ] `internal/test/e2e/suite/` (tests across all 4 products)
  - [ ] Test KMS protecting Identity secrets
  - [ ] Test PKI issuing certs for all services
  - [ ] Test JOSE operations for Identity tokens

- [ ] **Step 85**: Implement suite-level monitoring
  - [ ] Grafana dashboards by product and service
  - [ ] Aggregate metrics from all 8 services
  - [ ] Distributed tracing across product boundaries

---

## Phase 8: Testing & Documentation (Steps 86-95)

### Step 86-90: Coverage & Testing

- [ ] **Step 86**: Run full test suite on SM product
  - [ ] `go test ./internal/sm/... -cover`
  - [ ] Target 95%+ coverage for all packages
  - [ ] Fix any test failures from migration

- [ ] **Step 87**: Run full test suite on PKI product
  - [ ] `go test ./internal/pki/... -cover`
  - [ ] Target 95%+ coverage for all packages
  - [ ] Fix any test failures from migration

- [ ] **Step 88**: Run full test suite on JOSE product
  - [ ] `go test ./internal/jose/... -cover`
  - [ ] Target 95%+ coverage for all packages
  - [ ] Fix any test failures from migration

- [ ] **Step 89**: Run full test suite on Identity product
  - [ ] `go test ./internal/identity/... -cover`
  - [ ] Target 95%+ coverage for all 5 services
  - [ ] Fix any test failures from migration

- [ ] **Step 90**: Run gremlins mutation testing on all products
  - [ ] SM: `gremlins unleash --workers 2 ./internal/sm/kms/...` (target 98% efficacy)
  - [ ] PKI: `gremlins unleash --workers 2 ./internal/pki/ca/...` (target 98% efficacy)
  - [ ] JOSE: `gremlins unleash --workers 2 ./internal/jose/ja/...` (target 98% efficacy)
  - [ ] Identity: `gremlins unleash --workers 2 ./internal/identity/...` (target 98% efficacy)

### Step 91-95: Documentation & Runbooks

- [ ] **Step 91**: Update architecture documentation
  - [ ] `docs/architecture/PRODUCTS.md` (4 products overview)
  - [ ] `docs/architecture/SERVICES.md` (8 services overview)
  - [ ] `docs/architecture/PKI-HIERARCHY.md` (suite→product→service CA chain)

- [ ] **Step 92**: Create product runbooks
  - [ ] `docs/runbooks/sm/DEPLOYMENT.md` (SM product deployment)
  - [ ] `docs/runbooks/pki/DEPLOYMENT.md` (PKI product deployment)
  - [ ] `docs/runbooks/jose/DEPLOYMENT.md` (JOSE product deployment)
  - [ ] `docs/runbooks/identity/DEPLOYMENT.md` (Identity product deployment)

- [ ] **Step 93**: Create service runbooks
  - [ ] `docs/runbooks/sm/kms/OPERATIONS.md` (KMS service operations)
  - [ ] `docs/runbooks/pki/ca/OPERATIONS.md` (CA service operations)
  - [ ] `docs/runbooks/jose/ja/OPERATIONS.md` (JA service operations)
  - [ ] `docs/runbooks/identity/authz/OPERATIONS.md` (authz service operations)
  - [ ] `docs/runbooks/identity/idp/OPERATIONS.md` (idp service operations)
  - [ ] `docs/runbooks/identity/rs/OPERATIONS.md` (rs service operations)
  - [ ] `docs/runbooks/identity/rp/OPERATIONS.md` (rp service operations)
  - [ ] `docs/runbooks/identity/spa/OPERATIONS.md` (spa service operations)

- [ ] **Step 94**: Update API documentation
  - [ ] Generate OpenAPI docs for all 8 services
  - [ ] Publish to `docs/api/` with product/service hierarchy
  - [ ] Add mTLS and SPIFFE ID requirements

- [ ] **Step 95**: Create migration guide
  - [ ] `docs/MIGRATION-GUIDE.md` (how to migrate from old structure to new)
  - [ ] Breaking changes in import paths
  - [ ] Breaking changes in config file locations
  - [ ] Breaking changes in executable names

---

## Phase 9: CI/CD & Workflow Updates (Steps 96-100)

### Step 96-100: CI/CD Updates

- [ ] **Step 96**: Update GitHub Actions workflows
  - [ ] Update build matrix to include 4 products
  - [ ] Update test matrix to include 8 services
  - [ ] Add product-specific test jobs (run only on product changes)

- [ ] **Step 97**: Update Docker build scripts
  - [ ] Multi-stage builds for each product
  - [ ] Service-specific Docker images
  - [ ] Suite-level Docker image (all products)

- [ ] **Step 98**: Update pre-commit hooks
  - [ ] Add import path validation (no old paths)
  - [ ] Add config file validation (product/service structure)
  - [ ] Add API spec validation (product/service structure)

- [ ] **Step 99**: Update E2E test workflows
  - [ ] `cmd/dev/e2e/main.go` runs suite-level tests
  - [ ] `cmd/dev/e2e-<product>/main.go` runs product-level tests
  - [ ] GitHub Actions workflow for each product

- [ ] **Step 100**: Final validation & cleanup
  - [ ] Run full CI/CD pipeline
  - [ ] Validate all 8 services start/stop correctly
  - [ ] Validate mTLS between all services
  - [ ] Validate suite→product→service CA hierarchy
  - [ ] Delete all old structure directories
  - [ ] Update README.md with new architecture

---

## Success Criteria

### Phase Completion Criteria

1. **Phase 1-2 (SM Product)**: `go test ./internal/sm/... -cover` passes with 95%+ coverage
2. **Phase 3 (PKI Product)**: `go test ./internal/pki/... -cover` passes with 95%+ coverage
3. **Phase 4 (JOSE Product)**: `go test ./internal/jose/... -cover` passes with 95%+ coverage
4. **Phase 5 (Identity Product)**: `go test ./internal/identity/... -cover` passes with 95%+ coverage (all 5 services)
5. **Phase 6 (Shared)**: All shared libraries have no product-specific logic
6. **Phase 7 (Suite)**: Suite CA hierarchy fully implemented and tested
7. **Phase 8 (Testing)**: Gremlins mutation testing achieves 98%+ efficacy on all products
8. **Phase 9 (CI/CD)**: All GitHub Actions workflows pass

### Final Validation

- [ ] Suite-level executable works: `cryptoutil suite deployment initsuiteca`
- [ ] Product-level executables work: `identity deployment initproductca`
- [ ] Service-level executables work: `identity-authz up`
- [ ] Docker Compose works: `docker compose -f deployments/suite/compose.yml up`
- [ ] All 8 services communicate via mTLS
- [ ] SPIFFE IDs validated in all inter-service calls
- [ ] Full E2E test suite passes
- [ ] All GitHub Actions workflows pass

---

## Implementation Notes

### SPIFFE/SPIRE Decision

**Recommendation**: Implement private PKI hierarchy WITHOUT SPIFFE/SPIRE initially:

**Rationale**:

- SPIFFE/SPIRE adds significant complexity (separate control plane)
- Suite→product→service CA hierarchy achieves same goal
- Can migrate to SPIFFE/SPIRE later if needed (same SPIFFE ID format)
- Docker Compose E2E tests remain simple (no Spire server/agent containers)

**Implementation**:

- Use `internal/shared/crypto/pki/` for CA hierarchy management
- Embed SPIFFE IDs in TLS cert Subject Alternative Names (SAN)
- Validate SPIFFE IDs in service middleware (parse from client cert SAN)
- Service discovery via Docker Compose service names or Kubernetes services

### Windows Firewall Exception Prevention

**Issue**: Test executables (server.test.exe) trigger Windows Firewall prompts

**Root Cause**: Tests bind to 0.0.0.0 (all network interfaces) instead of 127.0.0.1 (loopback)

**Fix**: Update all test servers to bind to 127.0.0.1:

- Update `internal/runtime/transport/public_https.go` test config
- Update `internal/runtime/transport/admin_https.go` test config
- Add pre-commit hook to enforce 127.0.0.1 in test files

### Gremlins Mutation Testing Threshold

**Current**: 80% efficacy target
**New**: 98% efficacy target

**Update Locations**:

- `.github/instructions/01-04.testing.instructions.md`
- `docs/gremlins/MUTATIONS-HOWTO.md`
- `docs/gremlins/MUTATIONS-TASKS.md`
- All package-specific gremlins commands

---

## Timeline Estimate

- **Phase 1 (Foundation)**: 2-3 days
- **Phase 2 (SM Product)**: 1-2 days
- **Phase 3 (PKI Product)**: 1-2 days
- **Phase 4 (JOSE Product)**: 1-2 days
- **Phase 5 (Identity Product)**: 3-4 days (5 services)
- **Phase 6 (Shared)**: 1-2 days
- **Phase 7 (Suite)**: 2-3 days
- **Phase 8 (Testing)**: 2-3 days
- **Phase 9 (CI/CD)**: 1-2 days

**Total**: 15-25 days (depending on issues discovered)
