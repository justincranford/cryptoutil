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
| `README.md` | This file - usage guide | Human |
| `spec.md` | Product specification, requirements, architecture | Human + LLM |
| `clarify.md` | Clarify underspecified areas in spec | Human + Agent |
| `plan.md` | Implementation plan, technical approach, timeline | Human + LLM |
| `tasks.md` | Granular task breakdown with dependencies | Human + Agent |
| `analyze.md` | Cross-artifact consistency check | Human + Agent |
| `implement/DETAILED.md` | Track progress of tasks  | Human + Agent |
| `implement/EXECUTIVE.md` | Track high-level notes for draft after tasks, polished summary after last task is done | Human + Agent |

---

## Spec Kit Workflow

See [GitHub Spec Kit](https://github.com/github/spec-kit).

1. [ ] **Constitution** (`/speckit.constitution`)
   - Create/update project governing principles
   - Define development guidelines
   - Document: `.specify/memory/constitution.md`

2. [ ] **Specify** (`/speckit.specify`)
   - Define WHAT to build (requirements, user stories); after Constitution
   - Focus on intent, not implementation
   - Document: `specs/NNN-cryptoutil/spec.md`

3. [ ] **Clarify** (`/speckit.clarify` - formerly `/quizme`)
   - Clarify underspecified areas in spec; after Specify, before Plan
   - Ask questions about ambiguous requirements
   - Document: `specs/NNN-cryptoutil/clarify.md`

4. [ ] **Plan** (`/speckit.plan`)
   - Define HOW to build (tech stack, architecture); after Clarify, before Tasks
   - Technical implementation approach
   - Document: `specs/NNN-cryptoutil/plan.md`

5. [ ] **Tasks** (`/speckit.tasks`)
   - Generate actionable task list from plan; after Plan, before Analyze
   - Include dependencies and LOE (Level of Effort) estimates
   - Document: `specs/NNN-cryptoutil/tasks.md`

6. [ ] **Analyze** (`/speckit.analyze`)
   - Cross-artifact consistency check; after Tasks, Before Implement
   - Requirement-to-task coverage analysis
   - Document: `specs/NNN-cryptoutil/analyze.md`

7. [ ] **Implement** (`/speckit.implement`)
   - Execute all tasks according to plan; after Analyze, iteratively updated and committed
   - Run tests+coverage before finishing tasks; test failures and coverage regression block task completion
   - Track progress in `specs/NNN-cryptoutil/implement/DETAILED.md` with TWO sections:
     1. Checklist of all tasks from tasks.md (maintains same order for cross-reference)
     2. Append-only timeline of task implementation (time-ordered, may be out of order from section 1)

8. [ ] **Executive Summary**
   - Stakeholder overview; timeline of high-level notes can be iteratively appended and committed during Implement
   - Customer Demonstrability; docker compose up+down standalone per product, docker compose up+down suite of all four products, e2e demo commands, demo videos
   - Risk Tracking: Known issues, limitations, missing features+tasks, incomplete features+tasks, areas of improvement
   - Post Mortem: What went well, What needs improvement, lessons learned, checklist of suggestions to add/update/delete in copilot instructions/constitution/next speckit iteration/speckit template
   - Track in `specs/NNN-cryptoutil/implement/EXECUTIVE.md`

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
- ✅ NFR4: Quality (linting, formatting, file sizes, ≥coverage 95%, mutation score ≥98%)
- ✅ NFR5: Testability (table-driven, happy+sad use case coverage, parallel, benchmarks, fuzz, property, integration, e2e, docker compose up+down)
- ✅ NFR6: Observability (structured logging, OTLP, Prometheus, health endpoints)
- ✅ NFR7: Deployment (Docker, container size, startup time, YAML config)

#### Testing Requirements

- ✅ Unit tests: Table-driven with `t.Parallel()`, ≥95% production/infra/util
- ✅ Integration tests: Docker Compose, real database, `//go:build integration` tag
- ✅ Benchmark tests: All hot paths, `*_bench_test.go` files
- ✅ Fuzz tests: All parsers/validators, ≥15s fuzz time, `*_fuzz_test.go` files
- ✅ Property-based tests: gopter, invariants, round-trip validation
- ✅ Mutation tests: gremlins, ≥98% mutation score, baseline per package
- ✅ Docker Compose: Full stack, real database, test database, real telemetry, standlone+suite, docker secrets, no environment variables
- ✅ E2E tests: Full stack, demo scripts, real telemetry

#### Quality Gates

- ✅ Pre-commit: build, lint, format, file size, encoding
- ✅ Pre-push: tests, coverage, benchmarks, dependencies
- ✅ Pre-merge: CI workflows passing, code review, integration tests, Docker deploy, dast, sast, quality

---

## Coverage Targets (Constitution v2.0)

| Code Type | Target | Tool |
|-----------|--------|------|
| Production | ≥95% | `go test -cover ./internal/product/...` |
| Infrastructure (cicd) | ≥95% | `go test -cover ./internal/cmd/cicd/...` |
| Utility | ≥95% | `go test -cover ./internal/common/util/...` |
| Mutation Score | ≥98% | `gremlins unleash` |

**Note**: Coverage targets ≥95% as of Constitution v2.0.

---

## FIPS 140-3 Requirements

All cryptographic operations MUST use NIST FIPS 140-3 approved algorithms:

| Algorithm Type | Approved | BANNED |
|----------------|----------|--------|
| Symmetric | AES ≥128 bits, AES-HS ≥256 bits | 3DES, DES |
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
- Data at rest encryption requirements; confidentiality and integrity
- Data in transit encryption requirements; confidentiality and integrity

### Section III: KMS Hierarchical Security

- Multi-layer key barrier architecture
- Shared unseal secrets in application microservices that share a database, for interoperability

### Section IV: Go Production Requirements

- Go version consistency (1.25.5+)
- Static linking with debug symbols; no CGO dependency
- GORM ORM (not sql.DB directly)
- Cross-database compatibility (PostgreSQL + SQLite)
- UUID type: TEXT type in databases (not UUID - breaks SQLite)
- Nullable UUIDs: NullableUUID type (not pointers)
- SQLite: WAL mode, busy_timeout, MaxOpenConns configuration
- JSON fields: `serializer:json` (not `type:json`)
- 127.0.0.1 for localhost (not "localhost") inside Docker Compose to avoid IPv6 vs IPv4 stack split issues

### Section V: Go Testing Requirements

- Table-driven tests with `t.Parallel()`
- Test helpers with `t.Helper()`
- Coverage: 95%/ (production/infra/utility)
- Cover happy and sad path use cases
- No magic values (must use random UUIDv7 or magic constants in `magic` package)
- Dynamic port allocation for concurrent testing (port 0 pattern)
- File naming: `*_test.go`, `*_bench_test.go`, `*_fuzz_test.go`, `*_property_test.go`, `*_integration_test.go`
- NEVER os.Exit() in library/test code

### Section VI: Code Quality Excellence

- Test artifacts MUST be held to the same HIGHEST QUALITY as production artifacts; code, config, workflows, automation, security, concurrency, performance, scalability, robustness, etc
- Fix ALL linting and formatting errors (no exceptions); including tests
- NEVER use `//nolint:` except for documented bugs
- UTF-8 without BOM encoding
- File size limits: 300/400/500 lines
- Pre-commit/pre-push hooks enforcement
- Consistent parameter/return order
- Prefer switch over if/else chains
- Standard Go Project Layout
- Import alias conventions (cryptoutilPackage)

---

*Template Version: 1.0.0*
*Maintained By: Spec Kit + Constitution + Copilot Instructions*
