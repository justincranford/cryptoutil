# cryptoutil Implementation Plan

## Overview

This plan outlines the implementation phases for cryptoutil, guided by the [constitution principles](../.specify/memory/constitution.md) and aligned with the [product specifications](./spec.md). All phases must adhere to FIPS 140-3 compliance, evidence-based completion, hierarchical key security, code quality excellence, and product architecture clarity.

**Key References:**

- **Constitution**: Immutable principles for FIPS compliance, evidence-based completion, security architecture, code quality, and product separation
- **Specification**: Detailed product capabilities (P1-P4) and current implementation status
- **Evidence Requirements**: All tasks require verifiable evidence (build clean, tests pass, coverage maintained, E2E demos work)

---

## Phase 1: Identity V2 Production Completion (Current Focus)

**Duration**: 1-2 weeks (accelerated from 2-4 weeks per grooming feedback)
**Goal**: Complete Identity product for production deployment

### 1.1 Login UI Implementation

| Task | Description | Priority | LOE |
|------|-------------|----------|-----|
| 1.1.1 | Create minimal HTML login template (server-rendered, no JavaScript) | HIGH | 2h |
| 1.1.2 | Add minimal CSS styling (accessible) | MEDIUM | 1h |
| 1.1.3 | Implement server-side form validation | HIGH | 2h |
| 1.1.4 | Add CSRF token handling | CRITICAL | 1h |
| 1.1.5 | Error message display | HIGH | 1h |

**Success Criteria**: `/oidc/v1/login` returns HTML form, processes credentials, redirects properly

### 1.2 Consent UI Implementation

| Task | Description | Priority | LOE |
|------|-------------|----------|-----|
| 1.2.1 | Create HTML consent template (minimal, server-rendered) | HIGH | 2h |
| 1.2.2 | Display client name, requested scopes, data access summary | HIGH | 1h |
| 1.2.3 | Show OAuth 2.1 compliant disclosure (configurable verbosity) | HIGH | 1h |
| 1.2.4 | Implement approve/deny actions | CRITICAL | 2h |

**Success Criteria**: `/oidc/v1/consent` displays scopes, processes user decision

### 1.3 Logout Flow Completion

| Task | Description | Priority | LOE |
|------|-------------|----------|-----|
| 1.3.1 | Clear server-side session | CRITICAL | 1h |
| 1.3.2 | Revoke associated tokens | HIGH | 2h |
| 1.3.3 | Redirect to post-logout URI | HIGH | 1h |
| 1.3.4 | Front-channel logout support | HIGH | 4h |
| 1.3.5 | Back-channel logout support | HIGH | 4h |

**Success Criteria**: `/oidc/v1/logout` terminates session, revokes tokens, supports all logout flows

### 1.4 Userinfo Endpoint Completion

| Task | Description | Priority | LOE |
|------|-------------|----------|-----|
| 1.4.1 | Extract user from token | CRITICAL | 1h |
| 1.4.2 | Return claims based on scopes | HIGH | 2h |
| 1.4.3 | Return JWT-signed response (OAuth 2.1 requirement) | CRITICAL | 2h |
| 1.4.4 | Add scope-based claim filtering | HIGH | 2h |

**Success Criteria**: `/oidc/v1/userinfo` returns JWT-signed user claims per token scopes

### 1.5 Security Hardening

| Task | Description | Priority | LOE |
|------|-------------|----------|-----|
| 1.5.1 | Implement PBKDF2-HMAC-SHA256 client secret hashing | CRITICAL | 2h |
| 1.5.2 | Token-user association: claims + DB (both for stateless validation and revocation) | CRITICAL | 2h |
| 1.5.3 | Token lifecycle cleanup: hybrid (on-access + periodic cleanup + DB TTL) | HIGH | 4h |
| 1.5.4 | Tiered rate limiting (IP + client + endpoint) | HIGH | 4h |
| 1.5.5 | Audit logging: all auth events + token introspection + revocation | HIGH | 4h |

**Success Criteria**: Secrets hashed with PBKDF2, tokens properly associated, cleanup running

---

## Phase 2: KMS Stabilization

**Duration**: 1-2 weeks
**Goal**: Ensure KMS demo reliability (100% reliable, must never fail) and OpenAPI documentation completeness

### 2.1 Demo Hardening

| Task | Description | Priority | LOE |
|------|-------------|----------|-----|
| 2.1.1 | Verify `go run ./cmd/demo kms` all steps (100% success rate) | CRITICAL | 2h |
| 2.1.2 | Add error recovery scenarios | MEDIUM | 4h |
| 2.1.3 | Document demo prerequisites | HIGH | 2h |

### 2.2 API Documentation

| Task | Description | Priority | LOE |
|------|-------------|----------|-----|
| 2.2.1 | Complete OpenAPI spec review (primary focus) | CRITICAL | 4h |
| 2.2.2 | Minimal executive summary | HIGH | 1h |

### 2.3 Integration Testing

| Task | Description | Priority | LOE |
|------|-------------|----------|-----|
| 2.3.1 | Add E2E test suite for key lifecycle (create, read, list, rotate) | HIGH | 8h |
| 2.3.2 | Test crypto operations (encrypt, decrypt, sign, verify) | HIGH | 4h |
| 2.3.3 | Test multi-tenant isolation (tenant A cannot access tenant B keys) | HIGH | 4h |
| 2.3.4 | Performance baseline (measure and document, no specific targets) | MEDIUM | 4h |

---

## Phase 3: Integration Demo

**Duration**: 1-2 weeks
**Goal**: Full stack demo working end-to-end: KMS + Identity + Telemetry + PostgreSQL

### 3.1 OAuth2 Client Configuration

| Task | Description | Priority | LOE |
|------|-------------|----------|-----|
| 3.1.1 | Pre-seed KMS as OAuth2 client (demo data includes KMS client) | HIGH | 1h |
| 3.1.2 | Bootstrap script for automated registration | HIGH | 2h |
| 3.1.3 | Implement token validation middleware (hybrid: local + introspection for revocation) | HIGH | 4h |
| 3.1.4 | Add resource-based scope authorization (elastickey:read, materialkey:create, etc.) | HIGH | 4h |

### 3.2 Token Validation in KMS

| Task | Description | Priority | LOE |
|------|-------------|----------|-----|
| 3.2.1 | Fetch JWKS from Identity | HIGH | 2h |
| 3.2.2 | Validate JWT signatures (local first) | CRITICAL | 2h |
| 3.2.3 | Check token expiration | CRITICAL | 1h |
| 3.2.4 | Verify required scopes | HIGH | 2h |
| 3.2.5 | Introspection for revocation check | HIGH | 2h |

### 3.3 Demo Script

| Task | Description | Priority | LOE |
|------|-------------|----------|-----|
| 3.3.1 | Update `go run ./cmd/demo integration` for full stack | HIGH | 4h |
| 3.3.2 | Docker Compose deployment healthy + demo passes (PRIMARY SUCCESS CRITERIA) | CRITICAL | 4h |
| 3.3.3 | Verify all UIs accessible (login, logout, consent) | CRITICAL | 2h |

**Success Criteria**: Docker Compose deployment healthy, all UIs accessible and working, `go run ./cmd/demo all` completes 7/7 steps

---

## Phase 4: Certificate Authority Foundation

**Duration**: 4-8 weeks
**Goal**: Implement core CA capabilities (Tasks 1-10 from docs/05-ca)

### 4.1 Domain Charter (Task 1)

| Deliverable | Description |
|-------------|-------------|
| `docs/ca/charter.md` | Scope, compliance obligations, non-goals |
| Scope matrix | Feature vs compliance mapping |
| Glossary | CA-specific terminology |

### 4.2 Configuration Schema (Task 2)

| Deliverable | Description |
|-------------|-------------|
| `docs/ca/config-schema.yaml` | JSON Schema for CA config |
| Validation utilities | Config validation tool |
| Sample configs | Example configurations |

### 4.3 Crypto Provider Abstractions (Task 3)

| Deliverable | Description |
|-------------|-------------|
| `internal/ca/crypto/provider.go` | Provider interface |
| Memory implementation | In-memory key storage |
| Filesystem implementation | File-based key storage |
| HSM stubs | Future HSM integration points |

### 4.4 Profile Engines (Tasks 4-5)

| Deliverable | Description |
|-------------|-------------|
| `internal/ca/profile/subject` | Subject template resolution |
| `internal/ca/profile/certificate` | Certificate policy rendering |
| Profile library | 20+ predefined profiles |

### 4.5 CA Hierarchy (Tasks 6-8)

| Deliverable | Description |
|-------------|-------------|
| `cmd/ca/root-bootstrap` | Root CA creation CLI |
| Intermediate provisioning | Subordinate CA workflow |
| Issuing CA lifecycle | Rotation and monitoring |

### 4.6 Enrollment API (Task 9)

| Deliverable | Description |
|-------------|-------------|
| `api/ca/openapi_spec.yaml` | OpenAPI specification |
| Generated handlers | oapi-codegen output |
| CSR processing | Certificate request handling |

### 4.7 Revocation Services (Task 10)

| Deliverable | Description |
|-------------|-------------|
| CRL generation | Certificate Revocation Lists |
| OCSP responder | Online status protocol |
| Delta CRLs | Incremental updates |

---

## Phase 5: Production Hardening

**Duration**: 2-4 weeks
**Goal**: Security, observability, deployment readiness

### 5.1 Security Hardening

| Task | Description |
|------|-------------|
| STRIDE threat model | Document attack surfaces |
| gosec configuration | Security linting rules |
| HSM adapter design | Future HSM integration |
| Penetration testing | External security review |

### 5.2 Observability Completion

| Task | Description |
|------|-------------|
| Complete metrics | All operations instrumented |
| Grafana dashboards | Pre-built visualizations |
| Alert rules | Proactive monitoring |
| Runbooks | Incident response procedures |

### 5.3 Deployment Automation

| Task | Description |
|------|-------------|
| Docker Compose optimization | Production-ready configs |
| Kubernetes manifests | K8s deployment options |
| Terraform modules | Infrastructure as code |
| CI/CD pipeline completion | Full automation |

---

## Phase 6: Advanced Features (Future)

**Duration**: 8+ weeks
**Goal**: Enterprise capabilities

### 6.1 Advanced MFA

- Email OTP delivery
- SMS OTP delivery
- FIDO2 WebAuthn completion
- Risk-based authentication

### 6.2 Enterprise Identity

- SAML 2.0 support
- LDAP/AD integration
- SCIM provisioning
- Group-based authorization

### 6.3 CA Advanced Features (Tasks 11-20)

- Time-stamping service
- Registration Authority
- CT log submission
- ACME protocol support
- EST/CMP/SCEP protocols

---

## Risk Management

### Technical Risks

| Risk | Mitigation | Owner |
|------|-----------|-------|
| UI implementation delays | Use simple HTML/CSS, avoid frameworks | Phase 1 |
| HSM integration complexity | Design for pluggable providers | Phase 4 |
| Performance under load | Early benchmarking, connection pooling | Ongoing |
| Security vulnerabilities | gosec, DAST scanning, external review | Ongoing |

### Schedule Risks

| Risk | Mitigation |
|------|-----------|
| Scope creep | Strict adherence to phase boundaries |
| Resource constraints | Prioritize critical path items |
| Dependency delays | Early identification, parallel work |

---

## Success Metrics

### Phase 1 Success

- [ ] `/oidc/v1/login` returns HTML, processes credentials
- [ ] `/oidc/v1/consent` displays scopes, records decision
- [ ] `/oidc/v1/logout` terminates session, revokes tokens
- [ ] `/oidc/v1/userinfo` returns claims per scopes
- [ ] Client secrets hashed (PBKDF2-HMAC-SHA256)
- [ ] Token cleanup job running

### Phase 2-3 Success

- [ ] `go run ./cmd/demo all` completes 7/7 steps
- [ ] Docker Compose deployment healthy
- [ ] Zero critical TODOs in identity code

### Phase 4 Success

- [ ] Root CA bootstrap functional
- [ ] Certificate enrollment API operational
- [ ] CRL/OCSP services running
- [ ] 85%+ test coverage in `internal/ca/`

### Overall Success

- [ ] All linting passes (`golangci-lint run`)
- [ ] Coverage targets met
- [ ] Zero CRITICAL/HIGH vulnerabilities
- [ ] Documentation complete

---

*Plan Version: 1.0.0*
*Last Updated: December 2025*
*Next Review: End of Phase 1*
