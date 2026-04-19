# Tasks — Framework v13: v10-v12 Cleanup

**Status**: 0 of 30 tasks complete (0%)
**Last Updated**: 2026-06-30
**Created**: 2026-06-30

---

## Quality Mandate - MANDATORY

**Quality Attributes (NO EXCEPTIONS)**:
- ✅ **Correctness**: ALL code must be functionally correct with comprehensive tests
- ✅ **Completeness**: NO phases or tasks or steps skipped, NO features de-prioritized, NO shortcuts
- ✅ **Thoroughness**: Evidence-based validation at every step
- ✅ **Reliability**: Quality gates enforced (≥95%/98% coverage/mutation)
- ✅ **Efficiency**: Optimized for maintainability and performance, NOT implementation speed
- ✅ **Accuracy**: Changes must address root cause, not just symptoms
- ❌ **Time Pressure**: NEVER rush, NEVER skip validation, NEVER defer quality checks
- ❌ **Premature Completion**: NEVER mark phases or tasks or steps complete without objective evidence

**ALL issues are blockers - NO exceptions:**

- ✅ **Fix issues immediately** — When unknowns discovered, blockers identified, tests fail, or quality gates are not met, STOP and address
- ✅ **Treat as BLOCKING**: ALL issues block progress to next task
- ✅ **Document root causes** — Root cause analysis is part of implementation, not optional
- ✅ **NEVER defer**: No "we'll fix later", no "non-critical", no "nice-to-have"
- ✅ **NEVER skip**: Cannot mark phase or task or step complete with known issues
- ✅ **NEVER de-prioritize quality** — Evidence-based verification is ALWAYS highest priority

---

## Task Status Legend

| Symbol | Meaning | When to Use |
|--------|---------|-------------|
| ❌ | Not started | Task not yet begun |
| 🔄 | In progress | Currently being worked on |
| ✅ | Complete | Task finished with evidence |
| ⏳ | Blocked | Requires external dependency (MUST have resolution plan) |

---

## Task Checklist

### Phase 1: v12 Docker-Deferred TLS Smoke Test

**Phase Objective**: Verify ALL v12 TLS/mTLS configuration actually works in running Docker Compose containers.

#### Task 1.1: Docker Desktop Verification
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: None
- **Description**: Verify Docker Desktop is running and `docker compose` is available.
- **Acceptance Criteria**:
  - [ ] `docker ps` succeeds
  - [ ] `docker compose version` shows v5+
  - [ ] Docker engine is responsive

#### Task 1.2: Start sm-kms Docker Compose Stack
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: Task 1.1
- **Description**: Build and start the sm-kms PS-ID deployment with all 4 instances (sqlite-1, sqlite-2, postgres-1, postgres-2) using `docker compose up --wait`.
- **Acceptance Criteria**:
  - [ ] `docker compose -f deployments/sm-kms/compose.yml up --wait` succeeds
  - [ ] All 4 app instances reach healthy status
  - [ ] PostgreSQL containers are running and accepting connections
  - [ ] pki-init container completed cert generation

#### Task 1.3: Verify PostgreSQL Server TLS (Cat 10)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: Task 1.2
- **Description**: Verify PostgreSQL leader and follower accept TLS connections using Cat 10 server certificates.
- **Acceptance Criteria**:
  - [ ] `psql` connects to leader with `sslmode=verify-full`
  - [ ] `psql` connects to follower with `sslmode=verify-full`
  - [ ] `pg_stat_ssl` confirms SSL=true for connections
  - [ ] Non-TLS connections are rejected (if HBA rules enforce)

#### Task 1.4: Verify PostgreSQL Client mTLS (Cat 12)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: Task 1.2
- **Description**: Verify app instances connect to PostgreSQL with client certificates (Cat 12).
- **Acceptance Criteria**:
  - [ ] App instances' PostgreSQL connections show client cert in `pg_stat_ssl`
  - [ ] GORM DSN includes correct cert paths
  - [ ] Connection works for both postgres-1 and postgres-2 app instances

#### Task 1.5: Verify PostgreSQL Replication TLS (Cat 13)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: Task 1.2
- **Description**: Verify leader↔follower replication uses mTLS (Cat 13 replication certs).
- **Acceptance Criteria**:
  - [ ] `pg_stat_ssl` shows SSL=true for replication connection
  - [ ] Replication status is streaming
  - [ ] Cat 13 replication cert is mounted and used

#### Task 1.6: Verify Public TLS Endpoint (Cat 2)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: Task 1.2
- **Description**: Verify public endpoint `/service/api/v1/health` responds over TLS using Cat 2 public server cert.
- **Acceptance Criteria**:
  - [ ] `curl --cacert <ca-cert> https://127.0.0.1:<port>/service/api/v1/health` succeeds
  - [ ] Response is HTTP 200 with valid health payload
  - [ ] All 4 instances respond correctly at their assigned ports

#### Task 1.7: Verify Admin mTLS Endpoint (Cat 3)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: Task 1.2
- **Description**: Verify admin endpoint `/admin/api/v1/livez` responds over mTLS using Cat 3 admin cert (from inside container, since admin is 127.0.0.1:9090).
- **Acceptance Criteria**:
  - [ ] `docker compose exec` reaches admin endpoint from inside container
  - [ ] `/admin/api/v1/livez` responds HTTP 200
  - [ ] `/admin/api/v1/readyz` responds HTTP 200
  - [ ] All 4 instances have functioning admin endpoints

#### Task 1.8: Fix Configuration Issues
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: —
- **Dependencies**: Tasks 1.3-1.7
- **Description**: Fix any configuration issues discovered during TLS verification (cert paths, HBA rules, volume mounts, GORM DSN, etc.).
- **Acceptance Criteria**:
  - [ ] All issues from Tasks 1.3-1.7 resolved
  - [ ] Docker compose stack restarts cleanly with fixes
  - [ ] Evidence logged in `test-output/v13-phase1/`

#### Task 1.9: Phase 1 Post-Mortem
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: Tasks 1.1-1.8
- **Description**: Update lessons.md with Phase 1 findings.
- **Acceptance Criteria**:
  - [ ] lessons.md Phase 1 section populated
  - [ ] Evidence archived in `test-output/v13-phase1/`
  - [ ] Commit: `docs(framework-v13): phase 1 post-mortem`

---

### Phase 2: TLS/mTLS E2E Go Tests

**Phase Objective**: Write Go E2E tests that programmatically validate TLS certificate chains and mTLS authentication in running Docker Compose deployments.

#### Task 2.1: CA-Validated TLS Client Setup
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: —
- **Dependencies**: Phase 1 complete
- **Description**: Add CA-validated TLS HTTP client to `e2e_infra` package, using pki-init-generated CA cert rather than `InsecureSkipVerify: true`.
- **Acceptance Criteria**:
  - [ ] Function creates `http.Client` with CA cert pool from pki-init output
  - [ ] Client validates server certificate chain
  - [ ] Client rejects connections to servers with untrusted certs
  - [ ] Unit tests cover happy path and error paths

#### Task 2.2: Migrate Existing PS-ID E2E Tests
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: —
- **Dependencies**: Task 2.1
- **Description**: Update all 4 existing PS-ID E2E tests (skeleton-template, jose-ja, sm-im, sm-kms) to use CA-validated TLS client instead of `InsecureSkipVerify: true`.
- **Acceptance Criteria**:
  - [ ] skeleton-template E2E uses CA-validated client
  - [ ] jose-ja E2E uses CA-validated client
  - [ ] sm-im E2E uses CA-validated client
  - [ ] sm-kms E2E uses CA-validated client
  - [ ] All 4 E2E tests pass with real TLS validation
  - [ ] Zero occurrences of `InsecureSkipVerify: true` in E2E test files

#### Task 2.3: TLS Chain Validation Tests (Table-Driven)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1.5h
- **Actual**: —
- **Dependencies**: Task 2.1
- **Description**: Write table-driven E2E tests for public TLS chain validation: happy path (correct CA), sad path (wrong CA, expired cert, hostname mismatch).
- **Acceptance Criteria**:
  - [ ] Happy path: connection succeeds with correct CA cert
  - [ ] Sad path: connection fails with wrong CA cert
  - [ ] Sad path: connection fails with no CA cert (system roots)
  - [ ] Tests are table-driven with `t.Parallel()` where applicable
  - [ ] Tests use `ComposeManager` for Docker Compose lifecycle

#### Task 2.4: Admin mTLS Validation Tests (Table-Driven)
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: —
- **Dependencies**: Task 2.1
- **Description**: Write table-driven E2E tests for admin mTLS: happy path (correct client cert), sad path (no client cert, wrong client cert).
- **Acceptance Criteria**:
  - [ ] Happy path: admin endpoint accessible with correct client cert
  - [ ] Sad path: admin endpoint rejects connection without client cert
  - [ ] Sad path: admin endpoint rejects connection with wrong client cert
  - [ ] Tests verify both `/admin/api/v1/livez` and `/admin/api/v1/readyz`

#### Task 2.5: PostgreSQL mTLS Connection Test
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: Task 2.1
- **Description**: Write E2E test that verifies app connects to PostgreSQL with client cert (Cat 12/14).
- **Acceptance Criteria**:
  - [ ] Test programmatically verifies PostgreSQL mTLS via `pg_stat_ssl`
  - [ ] Test confirms `ssl=true` and client cert present

#### Task 2.6: Phase 2 Post-Mortem
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: Tasks 2.1-2.5
- **Description**: Update lessons.md with Phase 2 findings.
- **Acceptance Criteria**:
  - [ ] lessons.md Phase 2 section populated
  - [ ] Evidence archived in `test-output/v13-phase2/`
  - [ ] Commit: `docs(framework-v13): phase 2 post-mortem`

---

### Phase 3: v11 Mutation & Race Testing

**Phase Objective**: Execute mutation testing and race detection that v11 deferred to Linux.

#### Task 3.1: Install and Verify gremlins
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: None (can run in parallel with Phases 1-2 if Docker-independent)
- **Description**: Install `gremlins` mutation testing tool on Linux and verify it works.
- **Acceptance Criteria**:
  - [ ] `gremlins` binary available in PATH
  - [ ] `gremlins unleash --help` outputs usage info

#### Task 3.2: Run Mutation Testing on pki-init
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Actual**: —
- **Dependencies**: Task 3.1
- **Description**: Run `gremlins unleash` on `internal/apps/tools/pki_init/` packages.
- **Acceptance Criteria**:
  - [ ] gremlins mutation score ≥95% for pki-init packages
  - [ ] Survivors analyzed (true survivors vs equivalent mutations)
  - [ ] Results logged in `test-output/v13-phase3/mutation-report.txt`

#### Task 3.3: Run Race Detection on pki-init
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: None
- **Description**: Run `go test -race -count=2` on pki-init packages.
- **Acceptance Criteria**:
  - [ ] Zero data races detected
  - [ ] All tests pass under race detector
  - [ ] Results logged in `test-output/v13-phase3/race-report.txt`

#### Task 3.4: Fix Survivors and Races
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: Tasks 3.2, 3.3
- **Description**: Fix any mutation survivors (add tests) and race conditions (fix synchronization).
- **Acceptance Criteria**:
  - [ ] All fixable survivors addressed with new tests
  - [ ] All races fixed
  - [ ] Re-run confirms improvements

#### Task 3.5: Phase 3 Post-Mortem
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: Tasks 3.1-3.4
- **Description**: Update lessons.md with Phase 3 findings.
- **Acceptance Criteria**:
  - [ ] lessons.md Phase 3 section populated
  - [ ] Evidence archived in `test-output/v13-phase3/`
  - [ ] Commit: `docs(framework-v13): phase 3 post-mortem`

---

### Phase 4: Documentation Synchronization

**Phase Objective**: Resolve documentation drift between tls-structure.md, tls-structure-suggestions.md, and ENG-HANDBOOK.md.

#### Task 4.1: Review tls-structure-suggestions.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: None (can run in parallel with Docker phases)
- **Description**: Review `docs/tls-structure-suggestions.md` and categorize each suggestion as: merge, reject, or defer.
- **Acceptance Criteria**:
  - [ ] All suggestions categorized
  - [ ] Rationale for each decision documented

#### Task 4.2: Merge Applicable Suggestions to ENG-HANDBOOK.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 45m
- **Actual**: —
- **Dependencies**: Task 4.1
- **Description**: Merge accepted suggestions into ENG-HANDBOOK.md §6.11.3 (TLS cert categories).
- **Acceptance Criteria**:
  - [ ] ENG-HANDBOOK.md §6.11.3 reflects current 14-category cert structure
  - [ ] Propagation markers updated where needed
  - [ ] Commit: `docs(eng-handbook): update TLS cert category structure`

#### Task 4.3: Verify tls-structure.md Consistency
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: Task 4.2
- **Description**: Verify `docs/tls-structure.md` is consistent with updated ENG-HANDBOOK.md §6.11.3.
- **Acceptance Criteria**:
  - [ ] No contradictions between tls-structure.md and ENG-HANDBOOK.md
  - [ ] Post-v12 cert structure (Cat 9, Cat 14 changes) accurately reflected

#### Task 4.4: Run lint-docs Validation
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: Task 4.3
- **Description**: Run `go run ./cmd/cicd-lint lint-docs` to verify propagation integrity.
- **Acceptance Criteria**:
  - [ ] lint-docs passes with zero errors
  - [ ] All @propagate/@source blocks in sync

#### Task 4.5: Delete tls-structure-suggestions.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 5m
- **Actual**: —
- **Dependencies**: Task 4.2
- **Description**: Delete `docs/tls-structure-suggestions.md` after merging (git history preserves content).
- **Acceptance Criteria**:
  - [ ] File deleted
  - [ ] Commit: `chore(docs): remove merged tls-structure-suggestions.md`

#### Task 4.6: Phase 4 Post-Mortem
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: Tasks 4.1-4.5
- **Description**: Update lessons.md with Phase 4 findings.
- **Acceptance Criteria**:
  - [ ] lessons.md Phase 4 section populated
  - [ ] Evidence archived in `test-output/v13-phase4/`
  - [ ] Commit: `docs(framework-v13): phase 4 post-mortem`

---

### Phase 5: Template Docker Validation

**Phase Objective**: Verify v10's template directory produces functionally correct Docker Compose deployments.

#### Task 5.1: Start skeleton-template Docker Compose Stack
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: Docker Desktop running (Task 1.1 or separate verification)
- **Description**: Run `docker compose up --wait` for skeleton-template deployment.
- **Acceptance Criteria**:
  - [ ] `docker compose -f deployments/skeleton-template/compose.yml up --wait` succeeds
  - [ ] All 4 instances (sqlite-1, sqlite-2, postgres-1, postgres-2) reach healthy status
  - [ ] pki-init container completed cert generation

#### Task 5.2: Verify Health Endpoints
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: Task 5.1
- **Description**: Verify all 4 instances respond on correct health endpoints.
- **Acceptance Criteria**:
  - [ ] `/service/api/v1/health` responds HTTP 200 on all 4 instances
  - [ ] `/browser/api/v1/health` responds HTTP 200 on all 4 instances
  - [ ] Admin endpoints accessible from inside containers

#### Task 5.3: Run Template-Compliance Linter
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: None (can run independently)
- **Description**: Run `go run ./cmd/cicd-lint lint-fitness` to confirm template structural compliance.
- **Acceptance Criteria**:
  - [ ] template-compliance fitness linter passes
  - [ ] No structural discrepancies

#### Task 5.4: Compare Template vs Actual Deployment
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 30m
- **Actual**: —
- **Dependencies**: Tasks 5.1, 5.3
- **Description**: Compare template output against actual skeleton-template deployment files for any functional discrepancies the structural linter might miss.
- **Acceptance Criteria**:
  - [ ] Comparison documented
  - [ ] Any discrepancies logged and assessed
  - [ ] Functional issues (if any) tracked as fix tasks

#### Task 5.5: Phase 5 Post-Mortem
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: Tasks 5.1-5.4
- **Description**: Update lessons.md with Phase 5 findings.
- **Acceptance Criteria**:
  - [ ] lessons.md Phase 5 section populated
  - [ ] Evidence archived in `test-output/v13-phase5/`
  - [ ] Commit: `docs(framework-v13): phase 5 post-mortem`

---

### Phase 6: Knowledge Propagation

**Phase Objective**: Apply lessons learned from Phases 1-5 to permanent artifacts.

#### Task 6.1: Review All Lessons
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: Phases 1-5 complete
- **Description**: Review lessons.md from all phases, categorize by artifact impact.
- **Acceptance Criteria**:
  - [ ] All lessons categorized (ENG-HANDBOOK, agents, skills, instructions, code, tests, workflows, docs)
  - [ ] Priority assigned per lesson

#### Task 6.2: Update ENG-HANDBOOK.md
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 20m
- **Actual**: —
- **Dependencies**: Task 6.1
- **Description**: Update ENG-HANDBOOK.md with patterns/decisions discovered during Phases 1-5.
- **Acceptance Criteria**:
  - [ ] New patterns documented
  - [ ] @propagate markers added where content is shared with instruction files
  - [ ] Commit: `docs(eng-handbook): v13 knowledge propagation`

#### Task 6.3: Update Agents, Skills, Instructions
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 15m
- **Actual**: —
- **Dependencies**: Task 6.1
- **Description**: Update agent, skill, and instruction files where v13 work exposed gaps.
- **Acceptance Criteria**:
  - [ ] Relevant agents updated (both Copilot and Claude variants)
  - [ ] Skills updated if applicable
  - [ ] Instructions updated if applicable
  - [ ] Separate semantic commits per artifact type

#### Task 6.4: Verify Propagation Integrity
- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 10m
- **Actual**: —
- **Dependencies**: Tasks 6.2, 6.3
- **Description**: Run `go run ./cmd/cicd-lint lint-docs validate-propagation` to verify all propagation is consistent.
- **Acceptance Criteria**:
  - [ ] lint-docs validate-propagation passes with zero errors
  - [ ] No drift between ENG-HANDBOOK.md and instruction files

---

## Cross-Cutting Tasks

### Testing
- [ ] Unit tests ≥95% coverage (production), ≥98% (infrastructure/utility)
- [ ] E2E tests validate real TLS chains (no InsecureSkipVerify)
- [ ] Mutation testing ≥95% for pki-init
- [ ] Race detector clean on pki-init
- [ ] All tests pass with `go test ./... -shuffle=on`

### Code Quality
- [ ] `golangci-lint run` clean
- [ ] `golangci-lint run --build-tags e2e,integration` clean
- [ ] `go build ./...` clean
- [ ] `go build -tags e2e,integration ./...` clean
- [ ] No new TODOs without tracking

### Documentation
- [ ] ENG-HANDBOOK.md §6.11.3 current (14-category cert structure)
- [ ] tls-structure.md consistent with ENG-HANDBOOK.md
- [ ] tls-structure-suggestions.md deleted after merge
- [ ] lint-docs passes with zero errors

### Deployment
- [ ] sm-kms Docker Compose starts and all TLS works
- [ ] skeleton-template Docker Compose starts and is functional
- [ ] template-compliance linter passes

---

## Notes / Deferred Work

- **Item 7 (deferred to v14)**: Full E2E framework redesign — registry-driven config, shared TestMain factory, 16-deployment orchestrator. Out of scope for v13.
- **Item 8 (deleted)**: "Time/tokens not relevant" — not a valid retrospective item.
- **Item 13 (deleted)**: "Time/tokens not relevant" — not a valid retrospective item.
- **Item 15 (deleted)**: "Wrong" per user assessment — not applicable.

---

## Evidence Archive

- `test-output/v13-phase1/` — Docker TLS smoke test logs
- `test-output/v13-phase2/` — E2E TLS test results
- `test-output/v13-phase3/` — Mutation and race testing results
- `test-output/v13-phase4/` — Documentation sync verification
- `test-output/v13-phase5/` — Template Docker validation logs
