# Tasks - Framework v2

**Status**: 0 of 0 tasks complete (Phase 2 detailed; other phases TBD during execution)
**Last Updated**: 2026-03-08
**Created**: 2026-03-08

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- OK **Correctness**: ALL code must be functionally correct with comprehensive tests
- OK **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- OK **Thoroughness**: Evidence-based validation at every step
- OK **Reliability**: Quality gates enforced (>=95%/98% coverage/mutation)
- OK **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- OK **Accuracy**: Changes must address root cause, not just symptoms
- NO **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- NO **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

---

## Task Checklist

### Phase 1: Close v1 Gaps

**Phase Objective**: Close framework gaps identified in v1 (CI workflow, contract tests, quality gates)

> Tasks to be detailed when this phase begins execution. See plan.md Phase 1.

---

### Phase 2: Remove InsecureSkipVerify (G402)

**Phase Objective**: Generate real TLS cert chains for all test servers, replace InsecureSkipVerify: true
with proper CA-trusting TLS configs, remove G402 from gosec.excludes.

**Background**: ???ec G402 (InsecureSkipVerify) is currently excluded in .golangci.yml because tests
use InsecureSkipVerify: true to bypass TLS cert validation. The correct fix is to expose the test
server's CA cert at startup so test clients can validate the cert chain properly.

**Scope**: ~10 services x N test files. All integration/E2E tests that call HTTPS test servers.

#### Task 2.1: Add TLS Test Bundle to service-template testserver

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Add TLS cert bundle generation to the shared testserver infrastructure
- **Acceptance Criteria**:
  - [ ] `NewTestTLSBundle(t)` in `internal/apps/template/service/testing/testserver/` generates self-signed CA + server cert
  - [ ] `TLSClientConfig(t *testing.T, bundle *TestTLSBundle) *tls.Config` returns config trusting the test CA cert
  - [ ] `testserver.StartAndWait()` accepts optional TLS bundle or auto-generates one
  - [ ] Server exposes `TLSBundle()` accessor so test setup can retrieve the CA cert
  - [ ] Unit tests for TLS bundle generation (>=95% coverage)
  - [ ] Build clean: `go build ./internal/apps/template/service/testing/...`
  - [ ] No linting errors: `golangci-lint run ./internal/apps/template/service/testing/...`
- **Files**:
  - `internal/apps/template/service/testing/testserver/tls_bundle.go` (new)
  - `internal/apps/template/service/testing/testserver/tls_bundle_test.go` (new)
  - `internal/apps/template/service/testing/testserver/testserver.go` (update StartAndWait)

#### Task 2.2: Migrate sm-im test HTTP clients

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in sm-im tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in sm-im test files
  - [ ] All sm-im tests pass: `go test ./internal/apps/sm/im/...`
  - [ ] No linting errors: `golangci-lint run ./internal/apps/sm/im/...`
- **Files**: `internal/apps/sm/im/**/*_test.go`, `internal/apps/sm/im/**/*_integration_test.go`

#### Task 2.3: Migrate jose-ja test HTTP clients

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in jose-ja tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in jose-ja test files
  - [ ] All jose-ja tests pass: `go test ./internal/apps/jose/...`
  - [ ] No linting errors: `golangci-lint run ./internal/apps/jose/...`
- **Files**: `internal/apps/jose/**/*_test.go`

#### Task 2.4: Migrate sm-kms test HTTP clients

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in sm-kms tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in sm-kms test files
  - [ ] All sm-kms tests pass: `go test ./internal/apps/sm/kms/...`
  - [ ] No linting errors: `golangci-lint run ./internal/apps/sm/kms/...`
- **Files**: `internal/apps/sm/kms/**/*_test.go`

#### Task 2.5: Migrate pki-ca test HTTP clients

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in pki-ca tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in pki-ca test files
  - [ ] All pki-ca tests pass: `go test ./internal/apps/pki/...`
  - [ ] No linting errors: `golangci-lint run ./internal/apps/pki/...`
- **Files**: `internal/apps/pki/**/*_test.go`

#### Task 2.6: Migrate identity service test HTTP clients (all 5)

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 2h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in identity-authz/idp/rp/rs/spa tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in identity service test files
  - [ ] All identity tests pass: `go test ./internal/apps/identity/...`
  - [ ] No linting errors: `golangci-lint run ./internal/apps/identity/...`
- **Files**: `internal/apps/identity/**/*_test.go`

#### Task 2.7: Migrate skeleton-template test HTTP clients

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.1
- **Description**: Replace InsecureSkipVerify: true with TLSClientConfig(t) in skeleton-template tests
- **Acceptance Criteria**:
  - [ ] Zero `InsecureSkipVerify: true` in skeleton-template test files
  - [ ] All template tests pass: `go test ./internal/apps/template/...`
  - [ ] `go test ./internal/apps/skeleton/...`
  - [ ] No linting errors
- **Files**: `internal/apps/template/**/*_test.go`, `internal/apps/skeleton/**/*_test.go`

#### Task 2.8: Remove G402 from gosec.excludes and activate semgrep rule

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 2.2 through 2.7 (all InsecureSkipVerify removed)
- **Description**: Remove G402 exclusion from .golangci.yml, activate the semgrep rule
- **Acceptance Criteria**:
  - [ ] `G402` removed from `gosec.excludes` list in `.golangci.yml`
  - [ ] `no-tls-insecure-skip-verify` rule uncommented in `.semgrep/rules/go-testing.yml`
  - [ ] `golangci-lint run ./...` passes with G402 enabled (zero violations)
  - [ ] `go test ./... -shuffle=on` passes
  - [ ] `go build ./...` clean
- **Files**:
  - `.golangci.yml` (remove G402 from excludes)
  - `.semgrep/rules/go-testing.yml` (uncomment rule)
- **Evidence**: `test-output/phase2/golangci-g402-clean.log`

#### Task 2.9: Full suite validation and phase completion

- **Status**: TODO
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.8
- **Description**: Full quality gate run, coverage verification, phase post-mortem
- **Acceptance Criteria**:
  - [ ] `go build ./...` and `go build -tags e2e,integration ./...` clean
  - [ ] `golangci-lint run` AND `golangci-lint run --build-tags e2e,integration` clean
  - [ ] `go test ./... -shuffle=on` passes (zero failures, zero skips)
  - [ ] `go test -race -count=2 ./...` clean
  - [ ] Coverage maintained: `go test ./... -coverprofile=coverage.out`
  - [ ] Git commit with conventional commit message
  - [ ] `lessons.md` updated with phase post-mortem
- **Evidence**: `test-output/phase2/full-validation.log`

---

### Phase 3: PKI-CA Domain Completion

**Phase Objective**: Certificate issuance, revocation, CRL, OCSP (see plan.md Phase 3)

> Tasks to be detailed when this phase begins execution.

---

### Phase 4: Identity Foundation (authz)

**Phase Objective**: OAuth 2.1 authorization server core (see plan.md Phase 4)

> Tasks to be detailed when this phase begins execution.

---

### Phase 5: Identity Provider (idp)

**Phase Objective**: OIDC provider, user authentication flows (see plan.md Phase 5)

> Tasks to be detailed when this phase begins execution.

---

### Phase 6: Identity Services (rp, rs, spa)

**Phase Objective**: Relying party, resource server, single page app (see plan.md Phase 6)

> Tasks to be detailed when this phase begins execution.

---

### Phase 7: Quality & Polish

**Phase Objective**: Coverage, mutation, benchmarking, documentation (see plan.md Phase 7)

> Tasks to be detailed when this phase begins execution.

---

## Cross-Cutting Tasks

### Semgrep Rules Maintenance

- [ ] After each phase: review `.semgrep/rules/go-testing.yml` for new relevant patterns
- [ ] After Phase 2 complete: uncomment `no-tls-insecure-skip-verify` in go-testing.yml

---

## ARCHITECTURE.md Cross-References

| Topic | Section |
|-------|---------|
| TLS Configuration | [Section 6.4](../../docs/ARCHITECTURE.md#64-cryptographic-architecture) |
| Test HTTP Client Patterns | [Section 10.3.4](../../docs/ARCHITECTURE.md#1034-test-http-client-patterns) |
| Integration Testing | [Section 10.3](../../docs/ARCHITECTURE.md#103-integration-testing-strategy) |
| Shared Test Infrastructure | [Section 10.3.6](../../docs/ARCHITECTURE.md#1036-shared-test-infrastructure) |
| Quality Gates | [Section 11.2](../../docs/ARCHITECTURE.md#112-quality-gates) |
| Security Architecture | [Section 6](../../docs/ARCHITECTURE.md#6-security-architecture) |
| Service Template | [Section 5.1](../../docs/ARCHITECTURE.md#51-service-template-pattern) |
| Post-Mortem & Knowledge Propagation | [Section 13.8](../../docs/ARCHITECTURE.md#138-phase-post-mortem--knowledge-propagation) |
