# Implementation Plan - Framework V18+V19: ENG-HANDBOOK.md Propagation + Prescriptive MANIFEST + Identity Conformance Migration

**Status**: Planning
**Created**: 2026-04-26
**Last Updated**: 2026-04-27
**Predecessors**: V17 (commit `a747ac2ea` — 87 linters, partial migration)
**Purpose**: Two merged work streams:

(1) **V18 — ENG-HANDBOOK.md Knowledge Propagation**: Propagate implementation-specific details
from four reference docs (`target-structure.md`, `tls-structure.md`, `deployment-templates.md`,
`claude-structure.md`) into the canonical `ENG-HANDBOOK.md`, plus one targeted fix to
`tls-structure.md` for the Admin CA Bundle section. After V18 phases complete, the four
suggestion docs are deleted.

(2) **V19 — Prescriptive MANIFEST.yaml + Identity Conformance Migration**: Expand
`api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml` to be fully
prescriptive, extend the `apps_ps_id_template` linter to validate all new MANIFEST fields,
and complete V17 deferred conformance migration for identity services (GAPs A through G)
plus sm-im root cleanup.

---

## Quality Mandate — MANDATORY

| Attribute | Requirement |
|-----------|-------------|
| Correctness | ALL additions accurate; no copy-paste errors |
| Completeness | ALL suggestion items addressed; NO items skipped |
| Thoroughness | Evidence-based validation at every step |
| Reliability | `lint-docs` and `lint-fitness` clean after every phase |
| Efficiency | Optimized for maintainability; NOT implementation speed |
| Accuracy | Root cause addressed; not just symptoms |
| NO Time Pressure | NEVER rush; NEVER skip validation; NEVER defer quality checks |
| NO Premature Completion | Objective evidence required before marking complete |

**ALL issues are blockers.** Fix immediately. NEVER defer.

---

## Overview

This merged plan executes V18 documentation work (Phases 1–5) before V19 code work
(Phases 6–13). V18 must run first to ensure ENG-HANDBOOK.md quality before structural
migration expands the surface area of affected docs.

**V18 Phases** (documentation-only; no Go code):
- **Phase 1**: ENG-HANDBOOK.md additions from `target-structure.md` (11 items → §2.1, §4.4, §B.7, §11.1.4, §12.2.1)
- **Phase 2**: `tls-structure.md` fix + ENG-HANDBOOK.md additions from `tls-structure.md` (5 items → §6.5, §6.11)
- **Phase 3**: ENG-HANDBOOK.md additions from `deployment-templates.md` (11 items → §12.2.1, §12.3.1, §12.3.3, §12.3.5, §13.2, §13.6)
- **Phase 4**: ENG-HANDBOOK.md additions from `claude-structure.md` (11 items → §2.1.1, §2.1.5, §14.11)
- **Phase 5**: lint-docs full verification (propagation drift check)

**V19 Phases** (Go code changes):
- **Phase 6**: Prescriptive MANIFEST.yaml + `apps_ps_id_template` linter extension (4h)
- **Phase 7**: Identity services server code migration — authz, idp, rs, rp, spa (20h)
- **Phase 8**: sm-im root cleanup (2h)
- **Phase 9**: Create missing `server/` subdirectory packages for 5 identity services (12h)
- **Phase 10**: Create missing `client/` packages for 8 PS-IDs (8h)
- **Phase 11**: Create missing `e2e/` packages for 5 PS-IDs (8h)
- **Phase 12**: Remove `knownExclusions` + final validation (3h)
- **Phase 13**: Knowledge propagation (2h)

---

## Technical Context

- **Language**: Go 1.26.1; CGO_ENABLED=0 (MANDATORY)
- **Linting**: golangci-lint v2.7.2+
- **DB**: PostgreSQL (E2E) / SQLite in-memory (unit/integration)
- **V18 key tool**: `go run ./cmd/cicd-lint lint-docs`
- **V19 key tool**: `go run ./cmd/cicd-lint lint-fitness`
- **Predecessor**: V17 commit `a747ac2ea` (87 linters, partial migration; GAPs A–G deferred)
- **V18 Affected Files** (documentation-only):
  - `docs/ENG-HANDBOOK.md` — 38 additions across §2, §4, §6, §11, §12, §13, §14
  - `docs/tls-structure.md` — Admin CA Bundle fix
  - 4 suggestion docs deleted
- **V19 Affected Files** (Go code):
  - `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml` — 6 new field categories
  - `internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/` — 2 files, 5 new check functions
  - `internal/apps/identity-{authz,idp,rs,rp,spa}/` — ~115 files moved/created across 5 services
  - `internal/apps/sm-im/` — 8 files moved/deleted (root cleanup)
  - `internal/apps/{jose-ja,pki-ca,identity-*,skeleton-template}/client/` — 8 new dirs
  - `internal/apps/{pki-ca,identity-idp,identity-rs,identity-rp,identity-spa}/e2e/` — 5 new dirs

---

## Background

### V18 Source Documents (Reference Docs)

| Doc | Purpose | ENG-HANDBOOK.md Target |
|-----|---------|----------------------|
| `docs/target-structure.md` | Concrete file/dir inventory, permissions, catalogs | §2.1, §4.4, §4.4.1, §4.4.4, §B.7, §11.1.4, §12.2.1 |
| `docs/tls-structure.md` | PKI cert categories, naming, count formulas, TLS modes | §6.5, §6.11 |
| `docs/deployment-templates.md` | Template rules (DF/CO/CF/SC/PC/SU), parameterization | §12.2.1, §12.3.1, §12.3.3, §12.3.5, §13.2, §13.6 |
| `docs/claude-structure.md` | Claude Code `.claude/` structure, skill/agent fields | §2.1.1, §2.1.5, §14.11 |

### V18 Suggestion Docs (Deleted After V18 Phases Complete)

Four suggestion docs were created to track documentation gaps. After V18 completes all items,
the suggestion docs are deleted since all remaining work is captured in ENG-HANDBOOK.md:

- `docs/tls-structure-suggestions.md` — Items 1–5 actionable; Items 6–7 obsolete (V12/V13 complete)
- `docs/target-structure-suggestions.md` — All 11 items actionable
- `docs/deployment-templates-suggestions.md` — All 11 items actionable
- `docs/claude-structure-suggestions.md` — All 11 items actionable

### V19 Deferred GAPs from V17 (carried into this plan)

| GAP | PS-ID | Description | Current State |
|-----|-------|-------------|---------------|
| A / Task 5.2 | identity-authz | Move 60+ handler/route/service files from root → server/ | ALL at PS-ID root |
| A / Task 5.3 | identity-idp | Move 40+ handler/route/service files from root → server/ | ALL at PS-ID root |
| A / Task 5.4 | identity-rs | Move service.go, validator.go, tests from root → server/ | At PS-ID root |
| A / Task 5.5 | identity-rp | Create lifecycle_test.go and port_conflict_test.go in server/ | Missing entirely |
| A / Task 5.6 | identity-spa | Create lifecycle_test.go and port_conflict_test.go in server/ | Missing entirely |
| A / Task 5.7 | sm-im | Move http_test.go, im_database_test.go, etc. from root → server/ | At PS-ID root |

### V19 Server/ Subdirectory State — All 10 PS-IDs

| PS-ID | server/apis/ | server/config/ | server/model/ | server/repository/ | Notes |
|-------|-------------|---------------|--------------|---------------------|-------|
| sm-kms | ❌ | ❌ | ❌ | ✅ | Legacy: businesslogic/ handler/ instead — DEFERRED V20 |
| sm-im | ✅ | ✅ | ✅ | ✅ | Canonical modern pattern |
| jose-ja | ✅ | ✅ | ✅ | ✅ | Canonical + service/ layer |
| pki-ca | ❌ | ✅ | ❌ | ❌ | Complex CA — DEFERRED V20 |
| identity-authz | ❌ | ✅ | ❌ | ❌ | All code at root (Phase 7) |
| identity-idp | ❌ | ✅ | ❌ | ❌ | All code at root (Phase 7) |
| identity-rs | ❌ | ✅ | ❌ | ❌ | Partial code at root (Phase 7) |
| identity-rp | ❌ | ✅ | ❌ | ❌ | Minimal server/ (Phase 7) |
| identity-spa | ❌ | ✅ | ❌ | ❌ | Minimal server/ (Phase 7) |
| skeleton-template | ✅ | ✅ | ✅ | ✅ | Canonical template |

### V19 client/ and e2e/ State — All 10 PS-IDs

| PS-ID | client/ | e2e/ |
|-------|---------|------|
| sm-kms | ✅ | ✅ |
| sm-im | ✅ | ✅ |
| jose-ja | ❌ | ✅ |
| pki-ca | ❌ | ❌ |
| identity-authz | ❌ | ✅ |
| identity-idp | ❌ | ❌ |
| identity-rs | ❌ | ❌ |
| identity-rp | ❌ | ❌ |
| identity-spa | ❌ | ❌ |
| skeleton-template | ❌ | ✅ |

**Gap summary**: 8 PS-IDs missing client/; 5 PS-IDs missing e2e/.

---

## Target Structure — Canonical PS-ID Layout (V19)

The FULL target structure every PS-ID must eventually reach:

```
internal/apps/{PS-ID}/
├── {SERVICE}.go                              REQUIRED — CLI entry: server/health/livez/readyz/shutdown
├── {SERVICE}_usage.go                        REQUIRED — CLI usage string via BuildUsageMain()
├── {SERVICE}_cli_test.go                     REQUIRED — CLI integration tests
│                                              EXCEPTION: sm-im uses im_cli_commands_test.go + im_cli_url_test.go
├── client/                                   REQUIRED — typed HTTP client package
│   ├── *.go (non-test, domain-named)          REQUIRED — at least one non-test .go file
│   └── *_test.go                              OPTIONAL
├── e2e/                                      REQUIRED — E2E Docker Compose integration tests
│   ├── testmain_e2e_test.go                  REQUIRED — TestMain: starts Compose, waits for health endpoint
│   └── {SERVICE}_e2e_test.go                 REQUIRED — primary E2E scenarios
│                                              ALL .go files MUST have //go:build e2e tag
├── testing/ (optional)                       OPTIONAL — shared test helpers
└── server/                                   REQUIRED — ALL server implementation; NOTHING domain at root
    ├── server.go                             REQUIRED — admin HTTPS server: livez/readyz/shutdown + mTLS
    ├── public_server.go                      REQUIRED — public HTTPS server: browser/ + service/ paths
    │                                          EXCEPTION: sm-kms (legacy — no public_server.go)
    ├── swagger.go                            REQUIRED — OpenAPI spec serving
    ├── swagger_test.go                       REQUIRED — swagger serving tests
    ├── testmain_test.go                      REQUIRED — integration TestMain: shared server+DB
    ├── {SERVICE}_lifecycle_test.go           REQUIRED — dual-port startup, graceful shutdown
    ├── {SERVICE}_port_conflict_test.go       REQUIRED — deterministic failure when ports bound
    ├── apis/                                 REQUIRED — HTTP handler implementations
    │   └── *.go / *_test.go                  EXCEPTION: sm-kms (legacy); pki-ca (deferred V20)
    ├── config/                               REQUIRED — server configuration package
    │   ├── config.go                         REQUIRED
    │   ├── config_test.go                    REQUIRED
    │   └── config_test_helper.go             REQUIRED
    │                                          EXCEPTION for config_test_helper.go: pki-ca, identity services
    ├── model/                                REQUIRED — GORM persistence models
    │   └── *.go / *_test.go                  EXCEPTION: sm-kms, pki-ca (deferred V20)
    └── repository/                           REQUIRED — database access layer
        ├── migrations/                       REQUIRED subdir
        │   ├── NNNN_name.up.sql
        │   └── NNNN_name.down.sql
        ├── migrations.go                     REQUIRED — //go:embed migrations/*.sql + Migrate()
        └── {domain}_repository.go            EXCEPTION: sm-kms, pki-ca (deferred V20)
```

**Root file rule (MANDATORY)**: ALL files at the PS-ID root MUST start with `{SERVICE}_` prefix
OR be named `testmain_test.go`. FORBIDDEN at root: `swagger.go`, `handlers_*.go`, `routes.go`,
`service.go`, `middleware.go`, `http_test.go`, `validator.go`, any non-CLI implementation file.

---

## Phase 0: Pre-flight Build Health [Status: ☐ TODO]

**Objective**: Verify clean baseline before any V18 or V19 changes.

- `go build ./...` exits 0
- `go build -tags e2e,integration ./...` exits 0
- `go run ./cmd/cicd-lint lint-fitness` exits 0 (all 87 linters pass)
- `go run ./cmd/cicd-lint lint-docs` exits 0
- `go test ./internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/...` exits 0

**Success**: Output archived in `test-output/v18v19-phase0/`.

---

## Phase 1: ENG-HANDBOOK.md Additions from target-structure.md [Status: ☐ TODO]

**Phase Objective**: Add 11 missing catalog entries, tables, and inventory sections from
`target-structure.md` into the appropriate ENG-HANDBOOK.md sections.

| Task | Item | ENG-HANDBOOK.md Section | Content |
|------|------|------------------------|---------|
| 1.1 | File Permission Convention Table | §4.4.1 | octal permission table (640/440/750/etc.) |
| 1.2 | Root-Level File Inventory | §4.4.1 | 23+ root config files that MUST exist |
| 1.3 | Root Hidden Directory Inventory | §4.4 | `.cicd-lint/`, `.semgrep/`, `.vscode/`, `.well-known/`, `.zap/` |
| 1.4 | .github/ Top-Level File Catalog | §2.1 | copilot-instructions.md, dependabot.yml, SECURITY.md, etc. |
| 1.5 | GitHub Actions Catalog | §B.7 | 15 reusable actions with purpose descriptions |
| 1.6 | Concrete Service Subdirectory Inventory | §4.4.4 | Per-PS-ID actual subdirectories (e.g., pki-ca 15+ dirs) |
| 1.7 | Identity Shared Package Catalog | §4.4.4 | `internal/apps/identity/` shared packages (9 packages) |
| 1.8 | Complete magic/ File Listing | §11.1.4 | 42 `magic_*.go` files organized by domain |
| 1.9 | Other Top-Level Directories | §4.4.1 | `scripts/`, `workflow-reports/`, `test-output/`, `pkg/` |
| 1.10 | Dockerfile Parameterization Table | §12.2.1 | Per-tier image.title, binary, EXPOSE, HEALTHCHECK, ENTRYPOINT |
| 1.11 | Pending Work Inventory | §4.4.6 | Product Dockerfiles missing (all 5 products) |
| 1.12 | Post-phase lint check | — | `lint-docs`, `lint-fitness` |

**Success**: All 11 items propagated to ENG-HANDBOOK.md; `lint-docs` exits 0.
**Post-Mortem**: After quality gates pass, update `lessons.md` with lessons learned.

---

## Phase 2: tls-structure.md Fix + ENG-HANDBOOK.md Additions from tls-structure.md [Status: ☐ TODO]

**Phase Objective**: Fix the partial Admin CA Bundle documentation in `tls-structure.md`, then add
5 items from `tls-structure.md` into ENG-HANDBOOK.md §6.5 and §6.11.

Items 6–7 from `tls-structure-suggestions.md` are **OBSOLETE** (V12/V13 phases complete;
`pki-init-order.md` is now a standalone doc at `docs/pki-init-order.md`).

| Task | Item | Target Doc | Content |
|------|------|-----------|---------|
| 2.0 | Admin CA Bundle fix | `tls-structure.md` | Add `--cert`/`--key`/`--ca-cert` flags to Policy Alignment section |
| 2.1 | Admin CA Bundle → ENG-HANDBOOK.md | §6.5 (PKI Architecture) | Admin mTLS client trust requirement |
| 2.2 | tls-config.yml Dynamic Cert Pattern | §6.11 (TLS Config) | TLSModeAutoGenerate/PreGenerated/Mixed table; SAME-AS-DIR-NAME convention |
| 2.3 | Realm Dynamic Binding (Decision 8) | §6.5 | pki-init realm list from registry.yaml; per-PS-ID realm defaults |
| 2.4 | postgres vs postgres-1/2 Naming | §6.11 | Shared domain vs per-instance naming rationale |
| 2.5 | Directory Count Formula Derivation | §6.11 or tls-structure.md | 26 global + 64 per-PS-ID × 10 = 630 formula with Cat 9 correction |
| 2.6 | Post-phase lint check | — | `lint-docs`, `lint-fitness` |

**Success**: Admin CA Bundle fixed; 5 items propagated to ENG-HANDBOOK.md; `lint-docs` exits 0.
**Post-Mortem**: After quality gates pass, update `lessons.md` with lessons learned.

---

## Phase 3: ENG-HANDBOOK.md Additions from deployment-templates.md [Status: ☐ TODO]

**Phase Objective**: Add 11 items covering parameterization tables, Dockerfile/Compose/config
enforceable rule catalogs, PostgreSQL mTLS cert reference, inconsistency inventory, and template
syntax spec into ENG-HANDBOOK.md §12–§13.

| Task | Item | ENG-HANDBOOK.md Section | Content |
|------|------|------------------------|---------|
| 3.0 | Pre-phase lint baseline | — | Verify lint-docs clean before changes |
| 3.1 | Complete Parameterization Table | §13.6 or §13.7 | Entity, port, build/container params; PS-ID matrix |
| 3.2 | Container UID/GID Security Rationale | §12.2.1 | UID 65532 rationale, ARG parameterization, debug override |
| 3.3 | Dockerfile Rules DF-01–DF-24 | §12.2.1 | 24 machine-checkable Dockerfile rules |
| 3.4 | Compose Rules CO-01–CO-22 | §12.3.1 | 22 compose rules including named-volume mandate |
| 3.5 | Deployment Config Rules CF-01–CF-17 | §13.2 | 17 config overlay rules (incl. PostgreSQL mTLS cert paths) |
| 3.6 | Standalone Config Rules SC-01–SC-06 | §13.2 | 6 standalone config rules (127.0.0.1 bind, port consistency) |
| 3.7 | Product/Suite Compose Rules PC/SU | §12.3.5 | PC-01–PC-06 (product) and SU-01–SU-04 (suite) rules |
| 3.8 | Secret File Value Patterns | §12.3.3 | Complete 14-secret table with exact value format patterns |
| 3.9 | PostgreSQL mTLS Cert Reference Table | §6.11.4 | Per-node cert ownership; PKI Cat 10–14 reference |
| 3.10 | Current Inconsistency Inventory | §13.6 or Appendix M | Dockerfile Pattern A/B/C bugs; config snake_case PS-IDs |
| 3.11 | Template Syntax Specification | §13.6 | `__KEY__` format; path-level vs content-level; template catalog |
| 3.12 | Post-phase lint check | — | `lint-docs`, `lint-fitness` |

**Success**: All 11 items propagated to ENG-HANDBOOK.md; `lint-docs` exits 0.
**Post-Mortem**: After quality gates pass, update `lessons.md` with lessons learned.

---

## Phase 4: ENG-HANDBOOK.md Additions from claude-structure.md [Status: ☐ TODO]

**Phase Objective**: Add 11 items covering `.claude/` directory structure, CLAUDE.md format,
skill/agent frontmatter fields, dynamic context injection, path-scoped rules, and agentskills.io
open standard into ENG-HANDBOOK.md §2.1 and §14.11.

| Task | Item | ENG-HANDBOOK.md Section | Content |
|------|------|------------------------|---------|
| 4.0 | Pre-phase lint baseline | — | Verify lint-docs clean before changes |
| 4.1 | .claude/ Directory Structure Reference | §14.11 or §2.1 | Full `.claude/` tree (agents/, skills/, rules/, settings.json, etc.) |
| 4.2 | User-Level ~/.claude/ Structure | §14.11 | `~/.claude/` layout; loading order vs project CLAUDE.md |
| 4.3 | CLAUDE.md Format and Loading Behavior | §14.11 | User message delivery, /compact survival, @import syntax |
| 4.4 | Required CLAUDE.md Sections for cryptoutil | §14.11 | Canonical section structure for this project's CLAUDE.md |
| 4.5 | Complete Skill Frontmatter Fields | §2.1.5 | `allowed-tools`, `model`, `effort`, `context`, `agent`, `paths`, `shell` |
| 4.6 | Dynamic Context Injection Syntax | §2.1.5 | Backtick-bang blocks; `$ARGUMENTS`, `$0/$N`, `${CLAUDE_SESSION_ID}` |
| 4.7 | Skill Body Structure Template | §2.1.5 | Recommended SKILL.md body structure for cryptoutil |
| 4.8 | Sub-Agent Frontmatter Fields | §2.1.1 | `disallowedTools`, `permissionMode`, `maxTurns`, `skills`, `memory`, `color` |
| 4.9 | Path-Scoped Rules (.claude/rules/) | §14.11 | Auto-load behavior, `paths` frontmatter, recommended cryptoutil rule files |
| 4.10 | agentskills.io Open Standard Context | §2.1.5 | Cross-agent shared frontmatter; multi-tool adoption |
| 4.11 | CLAUDE.md Length and Scoping Strategy | §14.11 | Per-directory CLAUDE.md, path-scoped rules for monorepos |
| 4.12 | Post-phase lint check | — | `lint-docs`, `lint-fitness` |

**Success**: All 11 items propagated to ENG-HANDBOOK.md; `lint-docs` exits 0.
**Post-Mortem**: After quality gates pass, update `lessons.md` with lessons learned.

---

## Phase 5: lint-docs Full Verification [Status: ☐ TODO]

**Phase Objective**: Run `lint-docs` and verify zero propagation drift across all modified
ENG-HANDBOOK.md sections and their corresponding instruction files. This is the quality gate
between V18 documentation work (Phases 1–4) and V19 code work (Phases 6–13).

| Task | Description |
|------|-------------|
| 5.1 | Run `go run ./cmd/cicd-lint lint-docs` — verify ALL checks pass |
| 5.2 | Fix any `validate-chunks`, `validate-coverage`, or `lint-agent-drift` violations found |
| 5.3 | Delete four suggestion docs after all items addressed: `tls-structure-suggestions.md`, `target-structure-suggestions.md`, `deployment-templates-suggestions.md`, `claude-structure-suggestions.md` |

**Success**: Zero `lint-docs` violations; suggestion docs deleted.
**Post-Mortem**: After quality gates pass, update `lessons.md` with lessons learned.

---

## ENG-HANDBOOK.md Section Map (V18)

| Phase | Suggestion Doc | ENG-HANDBOOK.md Sections Modified |
|-------|---------------|----------------------------------|
| Phase 1 | target-structure-suggestions.md | §2.1, §4.4, §4.4.1, §4.4.4, §4.4.6, §B.7, §11.1.4, §12.2.1 |
| Phase 2 | tls-structure-suggestions.md | §6.5, §6.11 + tls-structure.md fix |
| Phase 3 | deployment-templates-suggestions.md | §6.11.4, §12.2.1, §12.3.1, §12.3.3, §12.3.5, §13.2, §13.6 |
| Phase 4 | claude-structure-suggestions.md | §2.1.1, §2.1.5, §14.11 |
| Phase 5 | — | Verification + suggestion doc deletion |

---

## Phase 6: Prescriptive MANIFEST.yaml + Linter Extension (4h) [Status: ☐ TODO]

**Phase Objective**: Expand `__PS_ID__ MANIFEST.yaml` to fully specify server/ subdirectory
structure, config/ files, repository/ files/dirs, and required e2e/ files. Extend the linter
to validate all new MANIFEST fields.

#### MANIFEST Fields to ADD

```yaml
required_dirs:
  - server
  - client   # knownExclusions: jose-ja, pki-ca, identity-*, skeleton-template
  - e2e      # knownExclusions: pki-ca, identity-idp, identity-rs, identity-rp, identity-spa

required_server_dirs:
  - apis        # knownExclusions: sm-kms, pki-ca, all 5 identity services
  - config      # NO exclusions
  - model       # knownExclusions: sm-kms, pki-ca, all 5 identity services
  - repository  # knownExclusions: sm-kms, pki-ca, all 5 identity services

required_server_config_files:
  - config.go
  - config_test.go
  - config_test_helper.go   # knownExclusions: pki-ca, identity-{authz,idp,rs,rp,spa}

required_server_repository_files:
  - migrations.go           # knownExclusions: pki-ca, identity-{authz,idp,rs,rp,spa}

required_server_repository_dirs:
  - migrations              # knownExclusions: pki-ca, identity-{authz,idp,rs,rp,spa}

required_e2e_files:
  - testmain_e2e_test.go    # knownExclusions: pki-ca, identity-idp, identity-rs, identity-rp, identity-spa
  - __SERVICE___e2e_test.go # same knownExclusions
```

#### New Check Functions (apps_ps_id_template.go)

1. `checkServerDirs` — verifies `server/{dir}` for each RequiredServerDirs entry
2. `checkServerConfigFiles` — verifies `server/config/{file}` for each entry
3. `checkServerRepositoryFiles` — verifies `server/repository/{file}` for each entry
4. `checkServerRepositoryDirs` — verifies `server/repository/{dir}` for each entry
5. `checkE2EFiles` — verifies `e2e/{file}` with `__SERVICE__` substitution

**Affected Files**:
```
api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml    MODIFY
internal/apps-tools/cicd_lint/lint_fitness/apps_ps_id_template/
    apps_ps_id_template.go                                                    MODIFY
    apps_ps_id_template_test.go                                               MODIFY
```

**Success**: `lint-fitness` exits 0; MANIFEST is fully prescriptive; `apps_ps_id_template`
linter has ≥98% coverage; all 5 new check functions have corresponding test cases.

---

## Phase 7: Identity Services Server Code Migration (20h) [Status: ☐ TODO]

**Phase Objective**: Complete V17 GAPs — move domain code from PS-ID root → server/ for
identity-authz, identity-idp, identity-rs, identity-rp, identity-spa.

**Package declaration rule**: Files at root (`package identity_authz`) move to server/
(`package server`). Test files (`package identity_authz_test`) become `package server_test`.

#### 7a: identity-authz — ~60 files from root → server/

- **Keep at root**: `authz.go`, `authz_usage.go`, `authz_cli_test.go`, `authz_contract_test.go`
- **Move to server/**: swagger.go, service.go, all handlers_*.go → server/apis/, domain files
- **Create in server/**: `authz_lifecycle_test.go`, `authz_port_conflict_test.go`

#### 7b: identity-idp — ~40 files from root → server/

- **Keep at root**: `idp.go`, `idp_usage.go`, `idp_cli_test.go`, `idp_contract_test.go`
- **Move to server/**: swagger.go, service.go, all handlers_*.go → server/apis/, auth/, templates/
- **Create in server/**: `idp_lifecycle_test.go`, `idp_port_conflict_test.go`

#### 7c: identity-rs — ~7 files from root → server/

- **Keep at root**: `rs.go`, `rs_usage.go`, `rs_cli_test.go`, `rs_contract_test.go`
- **Move to server/**: swagger.go, service.go, validator.go + tests
- **Create in server/**: `rs_lifecycle_test.go`, `rs_port_conflict_test.go`

#### 7d: identity-rp — Move rp_test.go + create tests

- **Move from root → server/**: `rp_test.go` (HTTP handler test, not CLI)
- **Create in server/**: `rp_lifecycle_test.go`, `rp_port_conflict_test.go`

#### 7e: identity-spa — Move spa_test.go + create tests

- **Move from root → server/**: `spa_test.go`
- **Create in server/**: `spa_lifecycle_test.go`, `spa_port_conflict_test.go`

**Affected Files**:
```
internal/apps/identity-authz/server/   + swagger.go, lifecycle_test.go, port_conflict_test.go
internal/apps/identity-authz/          [~60 files moved from root to server/]
internal/apps/identity-idp/server/     + swagger.go, lifecycle_test.go, port_conflict_test.go
internal/apps/identity-idp/            [~40 files moved from root to server/]
internal/apps/identity-rs/server/      + swagger.go, service.go, validator.go, lifecycle_test.go, port_conflict_test.go
internal/apps/identity-rp/server/      + lifecycle_test.go, port_conflict_test.go; rp_test.go moved
internal/apps/identity-spa/server/     + lifecycle_test.go, port_conflict_test.go; spa_test.go moved
```

**Success**: All identity service roots have ONLY CLI files; all server/ directories have
swagger.go + testmain_test.go + lifecycle + port_conflict tests; `go test ./internal/apps/identity-.../...` passes.

---

## Phase 8: sm-im Root Cleanup (2h) [Status: ☐ TODO]

**Phase Objective**: Complete V17 GAP Task 5.7 — move all non-CLI test files from sm-im root to server/.

**Files to MOVE** from root → `server/`:
```
http_test.go, http_errors_test.go, response_body_test.go
im_database_test.go, im_server_lifecycle_test.go
im_lifecycle_test.go, im_port_conflict_test.go
```

**Files to DELETE** from root (redundant — server/ copy retained):
```
testmain_test.go
```

**Files to KEEP at root**: `im.go`, `im_usage.go`, `im_cli_commands_test.go`, `im_cli_url_test.go`

**Success**: sm-im root has ONLY the 4 CLI files; `go test ./internal/apps/sm-im/...` passes.

---

## Phase 9: Create Missing server/ Subdirectory Packages (12h) [Status: ☐ TODO]

**Phase Objective**: Create `server/apis/`, `server/model/`, `server/repository/` (with
`migrations/`) for all 5 identity services. sm-kms and pki-ca are **EXPLICITLY DEFERRED to V20**.

| PS-ID | server/apis/ | server/model/ | server/repository/ + migrations/ |
|-------|-------------|--------------|-----------------------------------|
| identity-authz | ✅ Phase 7 populates | CREATE | CREATE |
| identity-idp | ✅ Phase 7 populates | CREATE | CREATE |
| identity-rs | ✅ Phase 7 populates | CREATE | CREATE |
| identity-rp | CREATE (minimal handler) | CREATE | CREATE |
| identity-spa | CREATE (minimal handler) | CREATE | CREATE |

Each new `repository/` must have: `migrations/` subdir, migration SQL files using the range
from `api/cryptosuite-registry/registry.yaml`, and `migrations.go` with `//go:embed`.

**Affected Files**:
```
internal/apps/identity-{authz,idp,rs,rp,spa}/server/apis/        NEW
internal/apps/identity-{authz,idp,rs,rp,spa}/server/model/       NEW
internal/apps/identity-{authz,idp,rs,rp,spa}/server/repository/  NEW
internal/apps/identity-{authz,idp,rs,rp,spa}/server/repository/migrations/  NEW
internal/apps/identity-{authz,idp,rs,rp,spa}/server/repository/migrations.go  NEW
```

**Success**: All 5 identity services have `server/apis/`, `server/model/`, `server/repository/migrations/`;
`go build ./...` exits 0; knownExclusions reduced by 21 entries.

---

## Phase 10: Create Missing client/ Packages (8h) [Status: ☐ TODO]

**Phase Objective**: Create typed HTTP client packages for the 8 PS-IDs currently missing them.

| PS-ID | Minimum client/ content |
|-------|------------------------|
| jose-ja | `client.go` — GetJWKS, CreateJWK, RotateJWK |
| pki-ca | `client.go` — IssueCert, RevokeCert, GetCRL |
| identity-authz | `client.go` — Authorize, Introspect, Token |
| identity-idp | `client.go` — Login, Logout, JWKS |
| identity-rs | `client.go` — ValidateToken, GetResources |
| identity-rp | `client.go` — Callback, Logout |
| identity-spa | `client.go` — minimal API surface |
| skeleton-template | `client.go` — placeholder client |

**Affected Files**:
```
internal/apps/{jose-ja,pki-ca,identity-authz,identity-idp,identity-rs,
               identity-rp,identity-spa,skeleton-template}/client/  NEW (8 dirs)
```

**Success**: All 10 PS-IDs have `client/`; `go build ./...` exits 0; `required_dirs: client`
knownExclusions emptied.

---

## Phase 11: Create Missing e2e/ Packages (8h) [Status: ☐ TODO]

**Phase Objective**: Create E2E test packages for 5 PS-IDs that currently lack them.

| PS-ID | Required files |
|-------|---------------|
| pki-ca | `testmain_e2e_test.go`, `ca_e2e_test.go` |
| identity-idp | `testmain_e2e_test.go`, `idp_e2e_test.go` |
| identity-rs | `testmain_e2e_test.go`, `rs_e2e_test.go` |
| identity-rp | `testmain_e2e_test.go`, `rp_e2e_test.go` |
| identity-spa | `testmain_e2e_test.go`, `spa_e2e_test.go` |

ALL `.go` files in `e2e/` MUST have `//go:build e2e` as the first line.

**Affected Files**:
```
internal/apps/{pki-ca,identity-idp,identity-rs,identity-rp,identity-spa}/e2e/  NEW (5 dirs)
```

**Success**: All 10 PS-IDs have `e2e/`; `go build -tags e2e ./...` exits 0.

---

## Phase 12: Remove knownExclusions + Final Validation (3h) [Status: ☐ TODO]

**Phase Objective**: After Phases 6–11, all PS-IDs (except sm-kms and pki-ca) conform to the
canonical template. Remove temporary knownExclusions from `apps_ps_id_template`.

**Permanent exceptions (never removed)**:
- `sm-kms`: `public_server.go` optional (legacy)
- `sm-im`: `__SERVICE___cli_test.go` exclusion (uses two files)
- `sm-kms`, `pki-ca`: `required_server_dirs: [apis, model, repository]` — deferred V20

**Temporary exclusions to REMOVE** after Phases 6–11 complete:
- All identity service exclusions for `required_server_dirs`
- All identity service exclusions for `required_server_config_files: config_test_helper.go`
- All identity service exclusions for `required_server_repository_files`
- All identity service exclusions for `required_dirs: client`
- All identity service exclusions for `required_dirs: e2e`
- All identity service exclusions for `required_e2e_files`

**Success**: `go run ./cmd/cicd-lint lint-fitness` exits 0; only 3 permanent exceptions remain.

---

## Phase 13: Knowledge Propagation (2h) [Status: ☐ TODO]

**Phase Objective**: Apply lessons learned to permanent project artifacts. NEVER skip.

- Review `lessons.md` from all prior phases
- Update ENG-HANDBOOK.md: canonical PS-ID structure spec, MANIFEST field catalog, migration patterns
- Update `docs/target-structure.md`: reflect new canonical structure after V19 migration
- Update instruction files where V19 work surfaces new coding/testing patterns
- Update `.github/skills/fitness-function-gen/SKILL.md` with recursive MANIFEST pattern
- Verify propagation: `go run ./cmd/cicd-lint lint-docs` exits 0
- Commit all artifact updates with separate semantic commits per artifact type

---

## Affected Files Summary

| Stream | Formula | Count |
|--------|---------|-------|
| V18: ENG-HANDBOOK.md | 1 file, ~38 additions | 1 |
| V18: tls-structure.md fix | 1 file | 1 |
| V18: delete suggestion docs | 4 docs | 4 |
| V19: MANIFEST.yaml | 1 file modified | 1 |
| V19: apps_ps_id_template | 2 files modified | 2 |
| V19: identity service migrations | ~110 files moved across 5 services | ~110 |
| V19: sm-im root cleanup | 8 files moved/deleted | 8 |
| V19: server/ subdir packages | 5 × 5 dirs (apis,model,repo,repo/migrations) = 25 dirs | ~25 files |
| V19: client/ packages | 8 new dirs × 1 file | 8 |
| V19: e2e/ packages | 5 new dirs × 2 files | 10 |

---

## Executive Decisions

### Decision 1: MANIFEST.yaml Granularity Scope

**Decision**: Option C selected — `required_server_dirs` + `required_server_config_files` +
`required_server_repository_files` + `required_e2e_files`.

**Rationale**: config/, repository/, and e2e/ have fixed, non-domain file sets. apis/ and model/
vary by service — prescribing filenames would require per-PS-ID MANIFEST variants.

### Decision 2: client/ and e2e/ Promotion

**Decision**: Option A — promote to `required_dirs` with large `knownExclusions`.

**Rationale**: Makes requirement explicit; knownExclusions list makes migration debt visible in
every CI run until resolved.

### Decision 3: sm-kms and pki-ca server/ Subdirectory Migration

**Decision**: Option B — defer to V20.

**Rationale**: pki-ca's CA architecture is fundamentally different. sm-kms uses legacy
businesslogic/ ORM structure. Both require dedicated analysis sub-plans.

### Decision 4: identity-authz/idp handler file destination

**Decision**: Option B — move handlers_*.go to `server/apis/` as `package apis`.

**Rationale**: Matches sm-im canonical pattern. Separates HTTP concerns from domain concerns.

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| identity-authz circular import after package split | Medium | High | `go build ./...` after EVERY file move; fix cycles immediately |
| Package declaration changes break test package references | High | Medium | All `_test` packages updated before attempting `go test` |
| Large identity-authz migration (~60 files) introduces build errors | High | Medium | Move in batches: swagger first, service next, handlers last |
| sm-im test moves break existing CI | Low | Medium | Run `go test ./internal/apps/sm-im/...` after each file move |
| migrations/ numbering conflicts with framework range | Low | High | Verify against registry.yaml migration ranges before creating files |
| pki-ca server/ conflict with existing cmd/ structure | Low | High | pki-ca DEFERRED — knownExclusions maintained |
| V18 ENG-HANDBOOK.md additions create propagation drift | Low | Medium | Run lint-docs after every phase; fix before proceeding |

---

## Quality Gates — MANDATORY

**Per-Action**:
- ✅ `go test ./...` — 100% passing, zero skips
- ✅ `go build ./...` AND `go build -tags e2e,integration ./...`
- ✅ `golangci-lint run` AND `golangci-lint run --build-tags e2e,integration`
- ✅ No new TODOs without tracking in tasks.md

**Coverage Targets**:
- ✅ `apps_ps_id_template` package: ≥98% after Phase 6
- ✅ Identity service packages: ≥95% after Phase 7
- ✅ New `client/` packages: ≥95%
- ✅ New `e2e/` packages: excluded (E2E-tagged)

**Per-Phase**:
- ✅ `go run ./cmd/cicd-lint lint-fitness` exits 0 after each phase (Phases 6–12)
- ✅ `go run ./cmd/cicd-lint lint-docs` exits 0 after Phases 1–5 and Phase 13
- ✅ Race detector clean: `go test -race -count=2 ./...`

**ENG-HANDBOOK.md Cross-References**:
- [§5.1 Service Framework Pattern](../../docs/ENG-HANDBOOK.md#51-service-framework-pattern)
- [§10.2 Unit Testing](../../docs/ENG-HANDBOOK.md#102-unit-testing-strategy)
- [§10.4 E2E Testing](../../docs/ENG-HANDBOOK.md#104-e2e-testing-strategy)
- [§11.2 Quality Gates](../../docs/ENG-HANDBOOK.md#112-quality-gates)
- [§14.1 Coding Standards](../../docs/ENG-HANDBOOK.md#141-coding-standards)
- [§14.2 Version Control](../../docs/ENG-HANDBOOK.md#142-version-control)
- [§14.8 Phase Post-Mortem](../../docs/ENG-HANDBOOK.md#148-phase-post-mortem--knowledge-propagation)

---

## Success Criteria

**V18 (Documentation)**:
- [ ] All 38 ENG-HANDBOOK.md additions complete (11+5+11+11 items)
- [ ] `tls-structure.md` Admin CA Bundle section fixed
- [ ] All 4 suggestion docs deleted
- [ ] `go run ./cmd/cicd-lint lint-docs` exits 0

**V19 (Code)**:
- [ ] `__PS_ID__ MANIFEST.yaml` is fully prescriptive (all 6 new field categories)
- [ ] `apps_ps_id_template` linter validates all MANIFEST fields with ≥98% coverage
- [ ] All 10 PS-ID roots contain ONLY `{SERVICE}_`-prefixed CLI files
- [ ] All 10 PS-ID `server/` dirs have `swagger.go`, `testmain_test.go`, lifecycle, port_conflict tests
- [ ] All 5 identity services have `server/apis/`, `server/model/`, `server/repository/migrations/`
- [ ] All 10 PS-IDs have `client/`
- [ ] All 10 PS-IDs have `e2e/`
- [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0 with only 3 permanent exceptions

**Overall**:
- [ ] All quality gates passing; CI/CD green
- [ ] Evidence archived in `test-output/v18v19-*/`

---

## Out of Scope

- No changes to `docs/pki-init-order.md` (Items 6–7 from tls-structure-suggestions.md obsolete)
- No changes to deployment artifacts (`deployments/`, `configs/`)
- sm-kms server/ subdirectory migration (businesslogic/ → apis/model/) — DEFERRED to V20
- pki-ca server/ subdirectory migration (complex CA structure) — DEFERRED to V20
