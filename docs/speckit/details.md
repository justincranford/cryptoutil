# Speckit Archive: Detailed Delta Inventory

## Purpose

Section-by-section comparison of `docs/speckit/constitution.md` (archived) against `docs/ARCHITECTURE.md` (authoritative source of truth) and `.github/instructions/*.instructions.md`.

---

## Section I: Product Delivery Requirements

### Delta 1: Product Count and Classification

- **Constitution**: "four Products (9 total services: 8 product services + 1 demo service)" — Cipher classified as "Demo"
- **ARCHITECTURE.md §3.1**: "five cryptographic-based products" — Cipher is a full product (Product #3)
- **Resolution**: ARCHITECTURE.md is authoritative. Cipher is a full product, not a demo.

### Delta 2: Path Structure

- **Constitution**: `internal/infra/*` (infrastructure) and `internal/product/*` (products)
- **ARCHITECTURE.md §4.4.1**: `internal/shared/*` (shared utilities) and `internal/apps/*` (applications)
- **Resolution**: ARCHITECTURE.md matches the actual codebase. Constitution used planned names that were never implemented.

### Delta 3: Configuration via Environment Variables

- **Constitution**: "Support configuration via environment variables, CLI parameters, and YAML files"
- **ARCHITECTURE.md §9.2.1**: "Docker Secrets > YAML > CLI parameters (NO environment variables)"
- **Instruction 02-02**: "CRITICAL: Environment variables are NOT supported for configuration"
- **Resolution**: ARCHITECTURE.md and instructions are authoritative. Env vars are explicitly banned.

### No Delta: Standalone/United Mode, Docker Compose, SQLite/PostgreSQL

- Both documents agree on standalone and united mode requirements.
- Both require Docker Compose support and dual database support (SQLite dev, PostgreSQL prod).

---

## Section II: Cryptographic Compliance

### No Significant Delta

- Both documents agree on CGO ban, FIPS 140-3, approved/banned algorithms, hash registry selection, Docker/K8s secrets, TLS 1.3+, CRLDP+OCSP.
- Minor wording differences only. All content covered in ARCHITECTURE.md §6.4 and instructions 02-07, 02-08, 02-09.

---

## Section III: Service Architecture Requirements

### Delta 4: Private Endpoint Configurability

- **Constitution**: "ALWAYS 127.0.0.1:9090 (NEVER configurable, NEVER exposed)"
- **ARCHITECTURE.md §5.3.3**: Port 0 for tests (dynamic), 9090 for production, address always 127.0.0.1
- **Resolution**: ARCHITECTURE.md is authoritative. Port is configurable for testing (port 0 mandatory for tests to avoid TIME_WAIT).

### Delta 5: Service Federation Patterns

- **Constitution**: "Circuit breakers, fallback modes, retry strategies"
- **ARCHITECTURE.md §3.3 / Instruction 02-01**: "No circuit breakers, no retry logic" — multi-level failover instead
- **Resolution**: ARCHITECTURE.md is authoritative. The architecture explicitly rejects circuit breakers and retry logic in favor of multi-level failover.

### No Delta: Dual HTTPS, Container Support, Federation Requirement

- Both agree on dual HTTPS endpoints, container support as mandatory, and configurable federation.

---

## Section IV: Testing Requirements

### Delta 6: Mutation Testing Targets

- **Constitution**: "≥85% Phase 4, ≥98% Phase 5+ gremlins score per package"
- **ARCHITECTURE.md §2.5**: "≥95% mandatory minimum, ≥98% ideal efficacy"
- **Instruction 03-02**: "≥98% ideal efficacy (all packages), ≥95% mandatory minimum"
- **Resolution**: ARCHITECTURE.md and instructions are authoritative. The phase-based targets from constitution are superseded by flat category-based targets.

### No Delta: Concurrency, Coverage, Test Execution Time

- Both agree on t.Parallel(), ≥95%/≥98% coverage, <15s per package, <180s total.

---

## Section V: Code Quality Requirements

### Delta 7: Token Stop Condition

- **Constitution**: "Token usage ≥ 990,000 (NOT 90k - ACTUAL 990,000!)"
- **ARCHITECTURE.md / Instructions**: No explicit token limit. Beast-mode instructions say "Time/token pressure does NOT exist"
- **Resolution**: Instructions are authoritative. Token pressure is explicitly rejected as a constraint.

### No Delta: Linting, Evidence-Based Completion, File Size Limits

- Both agree on zero linting exceptions, evidence-based completion, 300/400/500 line limits.

---

## Section VI: Development Workflow (Speckit Lifecycle)

### Delta 8: Speckit Lifecycle (REMOVED)

- **Constitution**: 8-step mandatory speckit lifecycle (`/speckit.constitution` through `/speckit.checklist`)
- **ARCHITECTURE.md / Instructions**: No speckit workflow. Development workflow uses `docs/todos-*.md` for task tracking.
- **Resolution**: Speckit infrastructure has been fully removed. This entire section is obsolete.

### Delta 9: Pre/Post-Implementation Gates

- **Constitution**: References `/speckit.clarify`, `/speckit.analyze`, `/speckit.checklist` as mandatory gates
- **ARCHITECTURE.md §11.2**: Quality gates defined as pre-commit, pre-push, and CI/CD gates (no speckit references)
- **Resolution**: ARCHITECTURE.md quality gates supersede speckit-based gates.

---

## Section VII: Service Template Requirements

### No Significant Delta

- Both agree on template extraction, dual HTTPS, health checks, telemetry, middleware, YAML+secrets config.
- Both agree on migration priority: cipher-im first, sm-kms never (reference implementation).
- Minor wording differences only. All content covered in ARCHITECTURE.md §5.1-5.2 and instruction 02-02.

---

## Section VIII: Governance and Standards

### Delta 10: PROGRESS.md Reference

- **Constitution**: "PROGRESS.md (in specs/NNN-cryptoutil/) is authoritative status source"
- **Current State**: No PROGRESS.md exists. Task tracking uses `docs/todos-*.md`.
- **Resolution**: Constitution reference is obsolete. `docs/todos-*.md` is the current task tracking mechanism.

### No Delta: RFC 2119 Terminology, Decision Authority, Documentation Standards

- Both agree on RFC 2119 usage, technical decision authority, and lean documentation.

---

## Section IX: Amendment Process

- **Constitution**: Formal amendment process (unanimous/majority consent, 48-hour review)
- **ARCHITECTURE.md**: No formal amendment process section (changes tracked via git history)
- **Resolution**: Constitution's amendment process is historical governance. ARCHITECTURE.md uses standard version control governance.

---

## Conclusion

All 10 deltas are resolved in favor of ARCHITECTURE.md and `.github/instructions/`. No content from constitution.md is missing from the authoritative documents. The constitution is safely archived for historical reference only.
