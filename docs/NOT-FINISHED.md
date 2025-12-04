# Unfinished Work Tracker

**Last Updated**: January 2026
**Purpose**: Single source of truth for incomplete/not-started work across all documentation

---

## Quick Reference

| Area | Status | Priority | Location |
|------|--------|----------|----------|
| CA Subsystem | ✅ Complete | - | `internal/ca/` |
| Identity V2 | ✅ Production Ready | - | `internal/identity/` |
| KMS | ✅ Production Ready | - | `internal/kms/` |
| Phase 5 Hardening | ⚠️ Partial | HIGH | Multiple |
| Future MFA | ❌ Not Started | LOW | `internal/identity/` |
| Infrastructure Refactor | ⏸️ Deferred | MEDIUM | `docs/01-refactor/` |
| Products Architecture | ⏸️ Deferred | LOW | `docs/03-products/` |

---

## Active Work Streams

### 1. Phase 5: Production Hardening (PARTIAL)

**Location**: `specs/001-cryptoutil/plan.md` - Phase 5

| Task | Status | Notes |
|------|--------|-------|
| STRIDE threat model | ✅ Complete (CA) | CA-specific threat model in `docs/05-ca/threat-model.md` |
| gosec configuration | ✅ Complete | Security linting configured |
| HSM adapter design | ✅ Stubs complete | Future HSM integration points ready |
| Penetration testing | ❌ Not started | External security review needed |
| Complete metrics | ⚠️ Partial | HTTP metrics complete, custom metrics pending |
| Grafana dashboards | ✅ Complete | Basic dashboards deployed |
| Alert rules | ✅ Complete (K8s) | Kubernetes PrometheusRules defined |
| Runbooks | ⚠️ Partial | Some runbooks in `docs/runbooks/` |
| Docker Compose | ✅ Complete | Production-ready configs |
| Kubernetes manifests | ✅ Complete (CA) | CA K8s manifests complete |
| Terraform modules | ❌ Not started | Infrastructure as code deferred |
| CI/CD completion | ⚠️ Partial | Core workflows complete |

---

## Deferred Work (Future Roadmap)

### 2. Infrastructure Refactor (DEFERRED)

**Location**: `docs/01-refactor/`
**Status**: Deferred to post-production hardening
**Reason**: Focus shifted to CA completion and production readiness

| Task | File | Status |
|------|------|--------|
| Service Group Taxonomy | `service-groups.md` | ⏸️ Planned |
| Dependency Analysis | `dependency-analysis.md` | ⏸️ Planned |
| Blueprint | `blueprint.md` | ⏸️ Planned |
| Import Alias Policy | `import-aliases.md`, `importas-migration.md` | ⏸️ Planned |
| CLI Strategy | `cli-strategy.md`, `cli-*.md` | ⏸️ Planned |
| Shared Utility Extraction | `shared-utilities.md` | ⏸️ Planned |
| Build Pipeline Impact | `pipeline-impact.md` | ⏸️ Planned |
| Workflow Updates | `workflow-updates.md` | ⏸️ Planned |
| Testing Validation | `testing-validation.md` | ⏸️ Planned |
| Documentation Sync | `documentation.md` | ⏸️ Planned |
| Tooling Updates | `tooling.md` | ⏸️ Planned |
| Observability Updates | `observability-updates.md` | ⏸️ Planned |
| KMS Extraction | `kms-extraction.md` | ⏸️ Planned |
| Identity Extraction | `identity-extraction.md` | ⏸️ Planned |
| CA Preparation | `ca-preparation.md` | ✅ Complete |

**Recommendation**: Revisit after Phase 5 hardening complete (Q2 2026)

---

### 3. Products Architecture Vision (DEFERRED)

**Location**: `docs/03-products/`
**Status**: Vision documented, implementation deferred
**Reason**: Current structure adequate for immediate needs

Key documents:

- `README.md` - Products/Infrastructure vision
- `infrastructure/I01-configuration.md` - Configuration component spec
- `passthru1/`, `passthru2/`, `passthru3/` - Implementation sprints

**Recommendation**: Revisit when adding P5+ products (Q3-Q4 2026)

---

### 4. Future MFA Factors (NOT STARTED)

**Location**: `specs/001-cryptoutil/spec.md` - P2 Identity MFA Factors

| Factor | Status | Priority |
|--------|--------|----------|
| TOTP | ✅ Working | - |
| Passkey/WebAuthn | ✅ Working | - |
| Hardware Security Keys | ❌ Not Implemented | HIGH |
| Email OTP | ⚠️ Partial | MEDIUM |
| SMS OTP | ⚠️ Partial | MEDIUM |
| HOTP | ❌ Not Implemented | LOW |
| Recovery Codes | ❌ Not Implemented | MEDIUM |
| Push Notifications | ❌ Not Required | - |
| Phone Call OTP | ❌ Not Required | - |

**Reference**: `docs/webauthn/browser-compatibility.md` for WebAuthn status

---

## TODO Files Status

### Active TODO Files

| File | Purpose | Items | Priority |
|------|---------|-------|----------|
| `04-general-stuff/todos-development.md` | Dev workflow | Hot reload, API versioning | LOW |
| `04-general-stuff/todos-infrastructure.md` | Infrastructure | K8s manifests, artifact consolidation | MEDIUM |
| `04-general-stuff/todos-observability.md` | Monitoring | Grafana expansion, metrics exposition | MEDIUM |
| `04-general-stuff/todos-quality.md` | Code quality | Identity TODOs (tracked separately) | LOW |
| `04-general-stuff/todos-security.md` | Security | Cookie HttpOnly, JSON parsing | MEDIUM |
| `04-general-stuff/todos-testing.md` | Testing | Pattern recommendations | LOW |

### Archived TODO Files

| File | Reason |
|------|--------|
| `04-general-stuff/archive/todos-database-schema-RESOLVED.md` | GORM issues fixed |
| `04-general-stuff/archive/task-17-gap-analysis-progress-HISTORICAL.md` | Historical reference |

---

## Identity V2 Tracked Items

**Primary Tracker**: `docs/02-identityV2/PROJECT-STATUS.md`

| Category | Status | Notes |
|----------|--------|-------|
| OAuth 2.1 Core | ✅ Complete | All R01-R11 tasks done |
| Secret Rotation | ✅ Complete | P5.01-P5.08 all phases done |
| Login/Consent UI | ✅ Working | HTML forms rendered |
| OIDC Discovery | ✅ Working | All well-known endpoints |
| Token Lifecycle | ✅ Working | Cleanup jobs running |

**Remaining Items** (tracked as future features):

- client_secret_jwt (HIGH priority)
- private_key_jwt (HIGH priority)
- session_cookie for SPA (Required)
- MFA enrollment/factors endpoints

---

## Completed Work (Archive Candidates)

The following directories contain completed work that could be archived:

1. `docs/02-identityV2/passthru0/` through `passthru6/` - Historical sprints
2. `docs/02-identityV2/identityV1-probably-mostly-completed/` - Legacy identity docs
3. `docs/archive/cicd-refactoring-nov2025/` - Completed refactoring
4. `docs/archive/codecov-nov2025/` - Completed codecov setup
5. `docs/archive/golangci-v2-migration-nov2025/` - Completed migration

---

## Speckit Best Practices for Work Verification

### Verifying Completed Work

1. **Check constitution.md**: Ensure work adheres to core principles
2. **Update spec.md**: Mark features complete with ✅ status
3. **Update plan.md**: Update phase status and success criteria
4. **Run evidence checks**:
   - `go build ./...` clean
   - `golangci-lint run` clean
   - `go test ./...` pass
   - Coverage maintained
5. **Update PROJECT-STATUS.md**: Single source of truth

### Identifying Subsequent Work

1. **Review spec.md**: Look for ❌, ⚠️, or missing status indicators
2. **Review plan.md**: Check unchecked items in Success Criteria
3. **Grep for TODOs**: `grep -r "TODO\|FIXME" internal/`
4. **Check NOT-FINISHED.md**: This document
5. **Run grooming sessions**: Use speckit grooming for validation

---

*Document Version: 1.0.0*
*Maintained alongside specs/001-cryptoutil/spec.md and plan.md*
