# cryptoutil Iteration Template (000)

## Purpose

This directory contains templates for planning and executing cryptoutil iterations using Spec Kit methodology.

**Version**: 1.0.0
**Created**: December 6, 2025
**Maintained By**: Constitution + Copilot Instructions

---

## Template Files

| File | Purpose | Primary User |
|------|---------|--------------|
| `spec.md` | Product specification, requirements, architecture | Human + LLM |
| `plan.md` | Implementation plan, technical approach, timeline | Human + LLM |
| `tasks.md` | Granular task breakdown with dependencies | LLM Agent |
| `README.md` | This file - usage guide | Human |

---

## When to Use These Templates

### Start of New Iteration

1. Copy `specs/000-cryptoutil-template/` to `specs/NNN-cryptoutil/`
2. Rename NNN to actual iteration number (e.g., `003-cryptoutil`)
3. Fill in spec.md with requirements and goals
4. Fill in plan.md with implementation approach
5. Fill in tasks.md with detailed task breakdown
6. Create additional files:
   - `PROGRESS.md` - Session log and status tracking
   - `EXECUTIVE-SUMMARY.md` - Stakeholder overview
   - `CLARIFICATIONS.md` - Ambiguity resolution
   - `ANALYSIS.md` - Coverage analysis (after /speckit.analyze)
   - `CHECKLIST-ITERATION-NNN.md` - Completion validation (after /speckit.checklist)

---

## Template Features

### Mandatory Requirements Included

#### Functional Requirements
- ✅ API endpoints with priorities
- ✅ Supported features and algorithms
- ✅ FIPS 140-3 compliance tracking
- ✅ Dependencies and prerequisites

#### Non-Functional Requirements (NFR)
- ✅ NFR1: Security (FIPS, secrets, TLS, audit logging)
- ✅ NFR2: Performance (response time, throughput, database queries)
- ✅ NFR3: Reliability (uptime, error rates, graceful shutdown)
- ✅ NFR4: Quality (coverage 95%/100%/100%, linting, file sizes, mutation score ≥80%)
- ✅ NFR5: Testability (table-driven, parallel, benchmarks, fuzz, property, integration)
- ✅ NFR6: Observability (structured logging, OTLP, Prometheus, health endpoints)
- ✅ NFR7: Deployment (Docker, container size, startup time, YAML config)

#### Testing Requirements
- ✅ Unit tests: Table-driven with `t.Parallel()`, ≥95% production, ≥100% infra/util
- ✅ Integration tests: Docker Compose, real database, `//go:build integration` tag
- ✅ Benchmark tests: All hot paths, `*_bench_test.go` files
- ✅ Fuzz tests: All parsers/validators, ≥15s fuzz time, `*_fuzz_test.go` files
- ✅ Property-based tests: gopter, invariants, round-trip validation
- ✅ Mutation tests: gremlins, ≥80% mutation score, baseline per package
- ✅ E2E tests: Full stack, demo scripts, real telemetry

#### Quality Gates
- ✅ Pre-commit: build, lint, file size, encoding
- ✅ Pre-push: tests, coverage, benchmarks, dependencies
- ✅ Pre-merge: CI passing, code review, integration tests, Docker deploy

---

## Coverage Targets (Constitution v2.0)

| Code Type | Target | Tool |
|-----------|--------|------|
| Production | ≥95% | `go test -cover ./internal/product/...` |
| Infrastructure (cicd) | ≥100% | `go test -cover ./internal/cmd/cicd/...` |
| Utility | 100% | `go test -cover ./internal/common/util/...` |
| Mutation Score | ≥80% | `gremlins unleash` |

**Note**: Coverage targets incremented from 90/95/100 to 95/100/100 as of Constitution v2.0.

---

## FIPS 140-3 Requirements

All cryptographic operations MUST use NIST FIPS 140-3 approved algorithms:

| Algorithm Type | Approved | BANNED |
|----------------|----------|--------|
| Symmetric | AES ≥128 bits | 3DES, DES |
| Asymmetric | RSA ≥2048 bits, EC NIST curves, EdDSA | RSA <2048 |
| Hash | SHA-256, SHA-384, SHA-512 | MD5, SHA-1 |
| Password Hash | PBKDF2-HMAC-SHA256/384/512 | bcrypt, scrypt, Argon2 |
| MAC | HMAC-SHA256/384/512 | - |

FIPS mode is ALWAYS enabled and MUST NEVER be disabled.

---

## File Size Limits

| Limit | Lines | Action Required |
|-------|-------|-----------------|
| Soft | 300 | Consider refactoring |
| Medium | 400 | Should refactor |
| Hard | 500 | MUST refactor |

**Rationale**: Optimal for human and LLM agent development and reviews.

---

## Secret Management

**CRITICAL**: NEVER use environment variables for secrets in ANY deployment.

**Required Approaches**:
- Docker: Mount secrets to `/run/secrets/`, reference with `file://` URLs
- Kubernetes: Mount secrets as files, reference directly
- Local Dev: Use YAML config files or Docker secrets (same as prod)

---

## Iteration Workflow

### Pre-Implementation Phase

1. **Clarify** (`/speckit.clarify` or manual)
   - Resolve all `[NEEDS CLARIFICATION]` markers in spec.md
   - Document resolutions in `CLARIFICATIONS.md`
   - Get user input on ambiguous requirements

2. **Analyze** (`/speckit.analyze` or manual)
   - Create `ANALYSIS.md` with requirement-to-task coverage matrix
   - Identify gaps and missing tasks
   - Validate all requirements have corresponding tasks

3. **Plan Review**
   - Review spec.md, plan.md, tasks.md for completeness
   - Validate dependencies and timelines
   - Confirm LOE estimates are realistic

### Implementation Phase

4. **Execute Tasks**
   - Follow tasks.md order respecting dependencies
   - Update PROGRESS.md after each session
   - Commit incrementally (not just at end)

5. **Continuous Validation**
   - Run tests after each task
   - Check coverage after each task
   - Run `golangci-lint` frequently
   - Update PROGRESS.md with evidence

### Post-Implementation Phase

6. **Checklist** (`/speckit.checklist` or manual)
   - Create `CHECKLIST-ITERATION-NNN.md`
   - Verify all completion criteria
   - Document evidence for each gate

7. **Executive Summary**
   - Create/update `EXECUTIVE-SUMMARY.md`
   - Stakeholder overview of deliverables
   - Manual testing guide
   - Known issues and limitations

8. **Retrospective**
   - What went well
   - What needs improvement
   - Lessons learned for next iteration
   - Update constitution/instructions if needed

---

## Integration with Constitution

These templates implement requirements from `.specify/memory/constitution.md`:

### Section I: Product Delivery
- Four products (JOSE, Identity, KMS, CA)
- Standalone and united deployment modes
- SQLite (dev) + PostgreSQL (prod)
- YAML configuration (no env vars for secrets)

### Section II: Cryptographic Compliance
- FIPS 140-3 approved algorithms only
- Secret management via Docker/K8s secrets
- Data at rest encryption requirements

### Section III: KMS Hierarchical Security
- Multi-layer key barrier architecture
- Shared unseal secrets for interoperability

### Section IV: Go Testing Requirements
- Table-driven tests with `t.Parallel()`
- Test helpers with `t.Helper()`
- No magic values (UUIDv7 or magic constants)
- Dynamic port allocation (port 0 pattern)
- File naming: `*_test.go`, `*_bench_test.go`, `*_fuzz_test.go`, `*_integration_test.go`

### Section V: Code Quality Excellence
- Fix ALL linting errors (no exceptions)
- NEVER use `//nolint:` except for documented bugs
- UTF-8 without BOM encoding
- File size limits: 300/400/500 lines
- Coverage: 95%/100%/100% (production/infra/utility)
- Pre-commit/pre-push hooks enforcement

---

## Integration with Copilot Instructions

These templates implement patterns from `.github/instructions/*.md`:

### 01-01.coding.instructions.md
- Named default variables (not inline literals)
- Consistent parameter/return order
- Prefer switch over if/else chains

### 01-02.testing.instructions.md
- Table-driven tests mandatory
- `t.Parallel()` for concurrency testing
- Dynamic port allocation pattern
- Test file organization (unit, bench, fuzz, integration)
- Coverage targets: 95/100/100

### 01-03.golang.instructions.md
- Go version consistency (1.25.5+)
- Static linking with debug symbols
- Standard Go Project Layout
- Import alias conventions (cryptoutilPackage)
- Crypto acronyms ALL CAPS
- NEVER os.Exit() in library/test code

### 01-04.database.instructions.md
- GORM ORM (not sql.DB directly)
- Cross-database compatibility (PostgreSQL + SQLite)
- UUID type: TEXT (not UUID - breaks SQLite)
- Nullable UUIDs: NullableUUID type (not pointers)
- JSON fields: `serializer:json` (not `type:json`)
- SQLite: WAL mode, busy_timeout, MaxOpenConns configuration

### 01-05.security.instructions.md
- FIPS 140-3 compliance mandatory
- Docker/K8s secrets (never env vars)
- TLS 1.3+, never InsecureSkipVerify
- 127.0.0.1 for localhost (not "localhost")

### 01-06.linting.instructions.md
- ALL linting errors MANDATORY to fix
- NEVER `//nolint:` except documented bugs
- golangci-lint v2.6.2+
- wsl, godot, mnd, errcheck rules
- UTF-8 without BOM enforcement

---

## References

| Document | Path | Purpose |
|----------|------|---------|
| Constitution | `.specify/memory/constitution.md` | Immutable principles |
| Copilot Instructions | `.github/instructions/*.md` | Coding patterns |
| Feature Template | `docs/feature-template/` | Multi-day feature planning |
| Spec Kit (External) | github/spec-kit | Spec-driven methodology |
| Iteration 1 | `specs/001-cryptoutil/` | Identity V2 + KMS reference |
| Iteration 2 | `specs/002-cryptoutil/` | JOSE + CA Server reference |

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2025-01-06 | Initial template creation |
|  |  | Coverage targets: 95/100/100 |
|  |  | Mutation testing: ≥80% |
|  |  | Comprehensive NFR sections |
|  |  | All test types included |

---

*Template Version: 1.0.0*
*Maintained By: Spec Kit + Constitution + Copilot Instructions*
