# Architecture Evolution Tasks - fixes-v8

**Status**: 0/45 tasks complete
**Created**: 2026-02-26
**Updated**: 2026-02-26

---

## Quality Mandate

ALL tasks MUST satisfy quality gates before marking complete:
- Build clean, lint clean, tests pass, coverage maintained
- Conventional commits with incremental history
- Evidence documented in test-output/ where applicable

---

## Phase 1: Architecture Documentation Hardening (8 tasks)

- [ ] 1.1 Run `cicd lint-docs validate-propagation` to check all @source/@propagate markers
- [ ] 1.2 Fix any stale propagation markers found in instruction files
- [ ] 1.3 Audit 68 lines >200 chars outside code blocks; fix non-table lines
- [ ] 1.4 Review 58 empty sections; categorize as intentional-placeholder vs incomplete
- [ ] 1.5 Document empty section findings (append to this file or plan.md)
- [ ] 1.6 Run full internal anchor validation on ARCHITECTURE.md
- [ ] 1.7 Run full file-link validation on ARCHITECTURE.md
- [ ] 1.8 Commit Phase 1 results: `docs: harden ARCHITECTURE.md post-structural-fix`

---

## Phase 2: Service-Template Readiness Evaluation (20 tasks)

### 2.1 Evaluation Framework (3 tasks)
- [ ] 2.1.1 Define scoring rubric (1-5 scale) for 10 dimensions
- [ ] 2.1.2 Create readiness scorecard template
- [ ] 2.1.3 Document evaluation methodology

### 2.2 SM Services (4 tasks)
- [ ] 2.2.1 Score sm-kms on 10 dimensions with evidence
- [ ] 2.2.2 Score sm-im on 10 dimensions with evidence
- [ ] 2.2.3 Compare sm-kms vs sm-im for alignment gaps
- [ ] 2.2.4 Document SM alignment findings

### 2.3 JOSE Service (2 tasks)
- [ ] 2.3.1 Score jose-ja on 10 dimensions with evidence
- [ ] 2.3.2 Compare jose-ja vs SM services for pattern consistency

### 2.4 PKI Service (2 tasks)
- [ ] 2.4.1 Score pki-ca on 10 dimensions with evidence
- [ ] 2.4.2 Compare pki-ca vs SM/JOSE services for pattern consistency

### 2.5 Identity Services (7 tasks)
- [ ] 2.5.1 Score identity-authz on 10 dimensions with evidence
- [ ] 2.5.2 Score identity-idp on 10 dimensions with evidence
- [ ] 2.5.3 Score identity-rs on 10 dimensions with evidence
- [ ] 2.5.4 Score identity-rp on 10 dimensions with evidence
- [ ] 2.5.5 Score identity-spa on 10 dimensions with evidence
- [ ] 2.5.6 Audit identity migration numbering (0002-0011 vs mandated 2001+)
- [ ] 2.5.7 Document identity readiness findings and architectural gaps

### 2.6 Summary (2 tasks)
- [ ] 2.6.1 Generate consolidated 9-service readiness scorecard
- [ ] 2.6.2 Commit Phase 2 results: `docs: service-template readiness evaluation`

---

## Phase 3: Identity Service Alignment Planning (10 tasks)

### 3.1 Migration Strategy (3 tasks)
- [ ] 3.1.1 Analyze identity migration 0002-0011 compatibility with template 1001-1999 range
- [ ] 3.1.2 Plan migration renumbering to 2001+ range (if needed)
- [ ] 3.1.3 Assess down-migration impact and rollback strategy

### 3.2 Architecture Analysis (3 tasks)
- [ ] 3.2.1 Evaluate shared domain vs per-service domain tradeoffs
- [ ] 3.2.2 Evaluate ServerManager vs per-service Application lifecycle
- [ ] 3.2.3 Document recommended architecture direction

### 3.3 Gap Analysis (3 tasks)
- [ ] 3.3.1 Scope identity-rp buildout (features, tests, migrations needed)
- [ ] 3.3.2 Scope identity-spa buildout (features, tests, migrations needed)
- [ ] 3.3.3 Plan E2E test decomposition (shared → per-service, if warranted)

### 3.4 Commit (1 task)
- [ ] 3.4.1 Commit Phase 3 results: `docs: identity service alignment plan`

---

## Phase 4: Next Architecture Step Execution (7 tasks)

### 4.1 Quick Wins (3 tasks)
- [ ] 4.1.1 Apply any config normalization fixes across identity services
- [ ] 4.1.2 Fix any missing health endpoint patterns
- [ ] 4.1.3 Fix any telemetry integration gaps

### 4.2 First Migration (2 tasks)
- [ ] 4.2.1 Execute highest-priority alignment task from Phase 3
- [ ] 4.2.2 Validate with full quality gate checks

### 4.3 Validation & Ship (2 tasks)
- [ ] 4.3.1 Run full test suite (unit + integration + E2E)
- [ ] 4.3.2 Commit Phase 4 results: `feat: <description of improvement>`

---

## Cross-Cutting Tasks

- [ ] CC-1 Keep docs/fixes-v8/plan.md Status field updated after each phase
- [ ] CC-2 Update ARCHITECTURE.md if any architectural decisions change
- [ ] CC-3 Push to remote after each phase completion for CI/CD validation

---

## Notes

### Pre-Existing Conditions
- Identity migrations use non-standard 0002-0011 range (predates current template migration spec)
- identity-rp and identity-spa are minimal implementations (~10 Go files, ~4 test files each)
- Identity uses a monolithic ServerManager pattern instead of per-service independent lifecycle
- All 9 services DO use NewServerBuilder (confirmed via grep)

### Deferred Items
None at this time.

---

## Evidence Archive

Evidence for completed tasks will be documented here as phases complete.

| Task | Evidence | Date |
|------|----------|------|
| - | - | - |
