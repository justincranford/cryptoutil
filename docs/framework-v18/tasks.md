# Tasks - Framework V18: ENG-HANDBOOK.md Knowledge Propagation

**Status**: 0 of 47 tasks complete (0%)
**Last Updated**: 2026-04-26
**Created**: 2026-04-26

## Quality Mandate — MANDATORY

| Attribute | Requirement |
|-----------|-------------|
| Correctness | ALL additions accurate; no copy-paste errors |
| Completeness | NO phases/tasks/steps skipped; NO shortcuts |
| Thoroughness | Evidence-based validation at every step |
| Reliability | `lint-docs` and `lint-fitness` clean after every phase |
| Efficiency | Optimized for maintainability; NOT implementation speed |
| Accuracy | Root cause addressed; not just symptoms |
| NO Time Pressure | NEVER rush; NEVER skip validation; NEVER defer quality checks |
| NO Premature Completion | Objective evidence required before marking complete |

**ALL issues are blockers.** Fix immediately. NEVER defer.

---

## Task Status Legend — MANDATORY

| Symbol | Meaning | When to Use |
|--------|---------|-------------|
| ❌ | Not started | Task not yet begun |
| 🔄 | In progress | Currently being worked on |
| ✅ | Complete | Task finished with evidence |
| ⏳ | Blocked | Requires external dependency (MUST have resolution plan) |

---

## Phase 1: ENG-HANDBOOK.md Additions from target-structure.md

**Phase Objective**: Add 11 missing catalog entries, tables, and inventory sections from
`target-structure.md` into the appropriate ENG-HANDBOOK.md sections.

### Task 1.0: Build Health Pre-Flight

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: None
- **Description**: Verify the codebase compiles and all existing linters pass before starting.
  Establish a clean baseline.
- **Acceptance Criteria**:
  - [ ] `go build ./...` exits 0
  - [ ] `go build -tags e2e,integration ./...` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
- **Files**: None (verification only)

### Task 1.1: Add File Permission Convention Table → §4.4.1

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.0
- **Description**: Add the octal permission table for directories (750), source files (640),
  secret files (440), executables (750), and generated files (640) to ENG-HANDBOOK.md §4.4.1.
- **Source**: `docs/target-structure-suggestions.md` Item 1
- **Acceptance Criteria**:
  - [ ] Table added to §4.4.1 with all 6 permission rows
  - [ ] Matches content from `docs/target-structure.md` permission section
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 1.2: Add Root-Level File Inventory → §4.4.1

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.1
- **Description**: Add the enumeration of 23+ root config files that MUST exist (`.air.toml`,
  `.gitleaks.toml`, `CLAUDE.md`, etc.) plus files that must NEVER be committed.
- **Source**: `docs/target-structure-suggestions.md` Item 2
- **Acceptance Criteria**:
  - [ ] "Must exist" file list added to §4.4.1
  - [ ] "Must never be committed" list added
  - [ ] Matches `docs/target-structure.md` root file catalog
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 1.3: Add Root Hidden Directory Inventory → §4.4

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.2
- **Description**: Add table of hidden root directories: `.cicd-lint/`, `.ruff_cache/`,
  `.semgrep/rules/`, `.vscode/`, `.well-known/`, `.zap/` with their purposes.
- **Source**: `docs/target-structure-suggestions.md` Item 3
- **Acceptance Criteria**:
  - [ ] Table added to §4.4 with all 6 hidden dirs and their purposes
  - [ ] `.vscode/mcp.json` noted as MCP server config
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 1.4: Add .github/ Top-Level File Catalog → §2.1

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.3
- **Description**: Add table of `.github/` top-level files: `copilot-instructions.md`,
  `dependabot.yml`, `SECURITY.md`, `versions-rules.xml`, `workflows-outdated-action-exemptions.json`.
- **Source**: `docs/target-structure-suggestions.md` Item 4
- **Acceptance Criteria**:
  - [ ] Table added to §2.1 (new §2.1.0 or §2.1.4) with all 5 files and purposes
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 1.5: Add GitHub Actions Catalog → §B.7

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.4
- **Description**: Add the 15 reusable actions in `.github/actions/` with per-action purpose
  descriptions to §B.7, including the `download-cicd` rename from `custom-cicd-lint`.
- **Source**: `docs/target-structure-suggestions.md` Item 5
- **Acceptance Criteria**:
  - [ ] Table of all 15 actions added to §B.7
  - [ ] `download-cicd` rename documented
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 1.6: Add Concrete Service Subdirectory Inventory → §4.4.4

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.5
- **Description**: Add per-PS-ID actual subdirectory table to §4.4.4. Each of the 10 PS-IDs
  lists its actual subdirs (e.g., `pki-ca` has 15+ dirs including `domain-v2/`, `intermediate/`).
- **Source**: `docs/target-structure-suggestions.md` Item 6
- **Acceptance Criteria**:
  - [ ] Table with all 10 PS-IDs and their actual subdirs added to §4.4.4
  - [ ] Matches `docs/target-structure.md` service subdirectory section
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 1.7: Add Identity Shared Package Catalog → §4.4.4

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.6
- **Description**: Add table of `internal/apps/identity/` shared packages (9 packages:
  `apperr/`, `config/`, `domain/`, `email/`, `issuer/`, `jobs/`, `mfa/`, `repository/`, `rotation/`).
- **Source**: `docs/target-structure-suggestions.md` Item 7
- **Acceptance Criteria**:
  - [ ] Table of all 9 shared identity packages added to §4.4.4
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 1.8: Add Complete magic/ File Listing → §11.1.4

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.7
- **Description**: Add the domain-organized listing of all 42 `magic_*.go` files in
  `internal/shared/magic/` to §11.1.4 (or §4.4.5).
- **Source**: `docs/target-structure-suggestions.md` Item 8
- **Acceptance Criteria**:
  - [ ] File listing table with domain groupings added to §11.1.4
  - [ ] Count verified against actual filesystem
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 1.9: Add Other Top-Level Directories → §4.4.1

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.8
- **Description**: Add table of other top-level directories: `scripts/` (gittracked, empty),
  `workflow-reports/` (gitignored), `test-output/` (gitignored), `pkg/` (reserved).
- **Source**: `docs/target-structure-suggestions.md` Item 9
- **Acceptance Criteria**:
  - [ ] Table with status and purpose for all 4 dirs added to §4.4.1
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 1.10: Add Dockerfile Parameterization Table → §12.2.1

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.9
- **Description**: Add per-deployment-tier table showing image.title, binary path, EXPOSE,
  HEALTHCHECK, and ENTRYPOINT differences between service, product, and suite tiers.
- **Source**: `docs/target-structure-suggestions.md` Item 10
- **Acceptance Criteria**:
  - [ ] Three-column (service/product/suite) table added to §12.2.1
  - [ ] Note about pending product Dockerfiles included
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 1.11: Add Pending Work Inventory → §4.4.6

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.10
- **Description**: Add table documenting known structural gap: product-level Dockerfiles missing
  for all 5 products. Include the "blocked pending suite binary architecture decision" note.
- **Source**: `docs/target-structure-suggestions.md` Item 11
- **Acceptance Criteria**:
  - [ ] Pending work table added to §4.4.6 (or §12.2 if §4.4.6 does not exist)
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 1.12: Post-Phase 1 Lint Check

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 1.1–1.11
- **Description**: Run all relevant linters after Phase 1 additions to ENG-HANDBOOK.md.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
  - [ ] `go build ./...` exits 0
- **Files**: None (verification only)

---

## Phase 2: tls-structure.md Fix + ENG-HANDBOOK.md Additions from tls-structure.md

**Phase Objective**: Fix the partial Admin CA Bundle documentation, then add 5 items from
`tls-structure.md` into ENG-HANDBOOK.md §6.5 and §6.11.

Note: Items 6–7 from `tls-structure-suggestions.md` are OBSOLETE — the V12/V13 phases are
complete and `docs/pki-init-order.md` is now a standalone reference doc.

### Task 2.0: Fix Admin CA Bundle Documentation in tls-structure.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 1.12
- **Description**: The Admin CA Bundle section at `tls-structure.md` line ~246 exists but does
  not mention the `--cert`, `--key`, and `--ca-cert` CLI flags used by the `livez` subcommand.
  Add this to the "Policy Alignment" section of `tls-structure.md`.
- **Source**: `docs/tls-structure-suggestions.md` Item 1 (PARTIAL fix)
- **Acceptance Criteria**:
  - [ ] `--cert`, `--key`, `--ca-cert` flags documented in tls-structure.md Admin mTLS section
  - [ ] Truststore file path and its role explained
  - [ ] Docker HEALTHCHECK mount requirement noted
- **Files**: `docs/tls-structure.md`

### Task 2.1: Add Admin CA Bundle → ENG-HANDBOOK.md §6.5

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.0
- **Description**: Add the Admin mTLS client trust requirement (admin port 9090 mTLS, `livez`
  CLI flags, Docker HEALTHCHECK mount) to ENG-HANDBOOK.md §6.5 PKI Architecture.
- **Source**: `docs/tls-structure-suggestions.md` Item 1
- **Acceptance Criteria**:
  - [ ] Admin mTLS client trust block added to §6.5
  - [ ] Cross-reference to §5.5 (Docker HEALTHCHECK pattern) included
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 2.2: Add tls-config.yml Dynamic Cert Pattern → §6.11

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.1
- **Description**: Add a "Runtime TLS Configuration" subsection to §6.11 documenting the three
  TLS modes (AutoGenerate, PreGenerated, Mixed), when to use each, and the SAME-AS-DIR-NAME
  file naming convention with a code example.
- **Source**: `docs/tls-structure-suggestions.md` Item 2
- **Acceptance Criteria**:
  - [ ] Three-mode table added (AutoGenerate/PreGenerated/Mixed)
  - [ ] SAME-AS-DIR-NAME convention explained with code example
  - [ ] `tls-config.yml` purpose described
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 2.3: Add Realm Dynamic Binding (Decision 8) → §6.5

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.2
- **Description**: Add Decision 8 (Realm Dynamic Binding) explanation to §6.5: how pki-init
  reads realm lists from `registry.yaml`, per-PS-ID defaults, what happens when a realm is
  added, and why the realm appears in directory name (not SAN/CN).
- **Source**: `docs/tls-structure-suggestions.md` Item 3
- **Acceptance Criteria**:
  - [ ] Decision 8 explanation added with per-PS-ID realm defaults table
  - [ ] "When a realm is added" behavior documented
  - [ ] Rationale for directory-name embedding (not SAN/CN) explained
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 2.4: Add postgres vs postgres-1/2 Naming → §6.11

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.3
- **Description**: Add "PostgreSQL Naming Conventions: Shared Domain vs. Individual Instances"
  subsection to §6.11 explaining when `postgres` (shared, Cat 4–5) vs `postgres-1`/`postgres-2`
  (individual, Cat 6–7, 14) naming is used and why.
- **Source**: `docs/tls-structure-suggestions.md` Item 4
- **Acceptance Criteria**:
  - [ ] Two-pattern table (shared domain vs individual) added with rationale
  - [ ] Cross-reference to cert categories included
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 2.5: Add Directory Count Formula Derivation → §6.11 or tls-structure.md

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.4
- **Description**: Add the directory count formula derivation (26 global + 64 per-PS-ID × 10 =
  630) with Cat 9 correction to either the tls-structure.md count summary or §6.11 in
  ENG-HANDBOOK.md. Per §14.1.2: raw counts without formulas are unverifiable.
- **Source**: `docs/tls-structure-suggestions.md` Item 5
- **Acceptance Criteria**:
  - [ ] Per-PS-ID formula: 26 global + 64 per-PS-ID = 90
  - [ ] Per-SUITE formula: correct Cat 9 breakdown showing 630 total
  - [ ] Formula added to both tls-structure.md count table and §6.11
- **Files**: `docs/tls-structure.md`, `docs/ENG-HANDBOOK.md`

### Task 2.6: Post-Phase 2 Lint Check

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 2.0–2.5
- **Description**: Run all relevant linters after Phase 2 changes.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
- **Files**: None (verification only)

---

## Phase 3: ENG-HANDBOOK.md Additions from deployment-templates.md

**Phase Objective**: Add 11 items covering parameterization tables, enforceable rule catalogs
(DF/CO/CF/SC/PC/SU), PostgreSQL mTLS cert reference, inconsistency inventory, and template
syntax specification into ENG-HANDBOOK.md §12–§13.

### Task 3.0: Pre-Phase 3 Lint Baseline

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 2.6
- **Description**: Verify lint-docs passes cleanly before Phase 3 changes.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
- **Files**: None

### Task 3.1: Add Complete Parameterization Table → §13.6/§13.7

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.0
- **Description**: Add four parameterization tables to §13.6 or new §13.7: Entity Parameters,
  Port Parameters, Build/Container Parameters, and Complete PS-ID Parameter Matrix (all 10 PS-IDs).
- **Source**: `docs/deployment-templates-suggestions.md` Item 1
- **Acceptance Criteria**:
  - [ ] Entity parameters table with `{SUITE}`, `{PS-ID}`, `{PS_ID}`, `{PRODUCT}`, display names
  - [ ] Port parameters table with 7 port parameters and formulas
  - [ ] Build/container parameters table with 10 entries
  - [ ] PS-ID matrix table with all 10 rows
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 3.2: Add Container UID/GID Security Rationale → §12.2.1

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.1
- **Description**: Add container UID 65532 security rationale (blast radius reduction) and the
  ARG parameterization strategy with debug override procedure to §12.2.1.
- **Source**: `docs/deployment-templates-suggestions.md` Item 2
- **Acceptance Criteria**:
  - [ ] UID 65532 rationale documented
  - [ ] ARG parameterization reason documented (de-duplication + debug override)
  - [ ] `--build-arg CONTAINER_UID=0` debug override noted with "NEVER in CI/CD" warning
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 3.3: Add Dockerfile Rules DF-01–DF-24 → §12.2.1

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.2
- **Description**: Add the complete 24-rule Dockerfile enforceable rule catalog to §12.2.1
  with requirement and rationale for each rule.
- **Source**: `docs/deployment-templates-suggestions.md` Item 3
- **Acceptance Criteria**:
  - [ ] All 24 rules (DF-01 through DF-24) added as a table
  - [ ] Each rule has requirement and rationale columns
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 3.4: Add Compose Rules CO-01–CO-22 → §12.3.1

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.3
- **Description**: Add the complete 22-rule Docker Compose enforceable rule catalog to §12.3.1,
  including the named-volume mandate (CO-21/CO-22) for portability across Docker Desktop/Swarm/K8s.
- **Source**: `docs/deployment-templates-suggestions.md` Item 4
- **Acceptance Criteria**:
  - [ ] All 22 rules (CO-01 through CO-22) added as a table
  - [ ] Named-volume rationale (CO-21/CO-22) explicitly stated
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 3.5: Add Deployment Config Rules CF-01–CF-17 → §13.2

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.4
- **Description**: Add 17 deployment config overlay rules to §13.2, including PostgreSQL mTLS
  cert path rules (CF-13–CF-17) and SQLite-specific exclusions.
- **Source**: `docs/deployment-templates-suggestions.md` Item 5
- **Acceptance Criteria**:
  - [ ] All 17 rules (CF-01 through CF-17) added as a table
  - [ ] PostgreSQL cert path rules CF-13–CF-17 present
  - [ ] CF-17 SQLite exclusion noted
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 3.6: Add Standalone Config Rules SC-01–SC-06 → §13.2

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.5
- **Description**: Add 6 standalone config rules to §13.2, including the critical SC-02
  requiring `127.0.0.1` (not `0.0.0.0`) to prevent Windows firewall popups.
- **Source**: `docs/deployment-templates-suggestions.md` Item 6
- **Acceptance Criteria**:
  - [ ] All 6 rules (SC-01 through SC-06) added as a table
  - [ ] SC-02 cross-references §5.6 Windows Firewall Prevention
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 3.7: Add Product/Suite Compose Rules → §12.3.5

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.6
- **Description**: Add product compose rules (PC-01–PC-06) and suite compose rules (SU-01–SU-04)
  to §12.3.5. Includes the `!override` tag requirement (PC-03) and port offset formulas.
- **Source**: `docs/deployment-templates-suggestions.md` Item 7
- **Acceptance Criteria**:
  - [ ] PC-01–PC-06 table added
  - [ ] SU-01–SU-04 table added
  - [ ] Port offset formulas (SERVICE+10000 / SERVICE+20000) confirmed
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 3.8: Add Secret File Value Patterns → §12.3.3

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.7
- **Description**: Add the complete 14-secret table with filename and exact value format pattern
  to §12.3.3. Includes the `postgres-url.secret` base DSN note (no `sslmode=` param).
- **Source**: `docs/deployment-templates-suggestions.md` Item 8
- **Acceptance Criteria**:
  - [ ] All 14 rows present in table
  - [ ] `postgres-url.secret` note about `sslmode=` separation from YAML `database-ssl*` fields
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 3.9: Add PostgreSQL mTLS Cert Reference Table → §6.11.4

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.8
- **Description**: Add the PKI category reference table for PostgreSQL nodes (Cat 10–14) and
  the logical cert ownership per node table to §6.11.4.
- **Source**: `docs/deployment-templates-suggestions.md` Item 9
- **Acceptance Criteria**:
  - [ ] PKI category reference table (7 rows) added to §6.11.4
  - [ ] Logical cert ownership table (5 nodes) added
  - [ ] SQLite instances explicitly noted as having NO PostgreSQL certs
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 3.10: Add Current Inconsistency Inventory → §13.6 or Appendix M

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.9
- **Description**: Add the three Dockerfile divergence patterns (A/B/C), specific per-PS-ID
  bugs, and config key naming inconsistencies to §13.6 or a new Appendix M.
- **Source**: `docs/deployment-templates-suggestions.md` Item 10
- **Acceptance Criteria**:
  - [ ] Pattern A/B/C table with affected PS-IDs and key deviations
  - [ ] Specific bug table (identity-spa wrong binary copy, skeleton-template jose refs, sm-im UID)
  - [ ] Config snake_case vs kebab-case inconsistency table
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 3.11: Add Template Syntax Specification → §13.6

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.10
- **Description**: Add the `__KEY__` placeholder format specification (double underscores,
  ALL_CAPS), path-level vs content-level placeholder behavior, and the template file catalog
  with instantiation counts to §13.6.
- **Source**: `docs/deployment-templates-suggestions.md` Item 11
- **Acceptance Criteria**:
  - [ ] `__KEY__` format rationale (vs `${VAR}` conflicts) documented
  - [ ] Path-level vs content-level placeholder distinction documented
  - [ ] Template file catalog with instantiation counts (`×10`, `×5`, `×1`)
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 3.12: Post-Phase 3 Lint Check

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 3.1–3.11
- **Description**: Run all relevant linters after Phase 3 additions.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
- **Files**: None (verification only)

---

## Phase 4: ENG-HANDBOOK.md Additions from claude-structure.md

**Phase Objective**: Add 11 items covering `.claude/` directory structure, CLAUDE.md format,
skill/agent frontmatter fields, dynamic context injection, path-scoped rules, and agentskills.io
open standard into ENG-HANDBOOK.md §2.1.1, §2.1.5, and §14.11.

### Task 4.0: Pre-Phase 4 Lint Baseline

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 3.12
- **Description**: Verify lint-docs passes cleanly before Phase 4 changes.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
- **Files**: None

### Task 4.1: Add .claude/ Directory Structure Reference → §14.11 or §2.1

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.0
- **Description**: Add the canonical `.claude/` directory tree (agents/, skills/, rules/,
  settings.json, agent-memory/, worktrees/) to §14.11 or §2.1. Note legacy `.claude/commands/`
  removal.
- **Source**: `docs/claude-structure-suggestions.md` Item 1
- **Acceptance Criteria**:
  - [ ] Directory tree added with all subdirs and their purposes
  - [ ] `.claude/commands/` removal noted
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.2: Add User-Level ~/.claude/ Structure → §14.11

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.1
- **Description**: Add the `~/.claude/` user-level structure (CLAUDE.md, agents/, skills/,
  rules/, projects/<proj>/memory/) and note that user-level loads before project-level.
- **Source**: `docs/claude-structure-suggestions.md` Item 2
- **Acceptance Criteria**:
  - [ ] `~/.claude/` directory tree added to §14.11
  - [ ] Loading order (user-level before project-level) documented
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.3: Add CLAUDE.md Format and Loading Behavior → §14.11

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.2
- **Description**: Document CLAUDE.md delivery as user message (not system prompt), /compact
  survival, @path import syntax (max 5 hops), HTML comment stripping, and the 200-line target.
- **Source**: `docs/claude-structure-suggestions.md` Item 3
- **Acceptance Criteria**:
  - [ ] "Delivery: user message, not system prompt" documented
  - [ ] /compact survival behavior documented
  - [ ] @import syntax and HTML comment stripping noted
  - [ ] 200-line target reiterated
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.4: Add Required CLAUDE.md Sections → §14.11

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.3
- **Description**: Document the required section structure for cryptoutil's CLAUDE.md:
  Architecture Source of Truth, Instruction Files, Agents table, Skills table.
- **Source**: `docs/claude-structure-suggestions.md` Item 4
- **Acceptance Criteria**:
  - [ ] Canonical CLAUDE.md section structure added to §14.11
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.5: Add Complete Skill Frontmatter Fields → §2.1.5

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.4
- **Description**: Add the complete SKILL.md frontmatter field table to §2.1.5 including all
  Claude Code-specific fields: `allowed-tools`, `model`, `effort`, `context`, `agent`, `paths`,
  `shell`. Note `disable-model-invocation` is Copilot-only.
- **Source**: `docs/claude-structure-suggestions.md` Item 5
- **Acceptance Criteria**:
  - [ ] All 12 frontmatter fields documented with Required column and description
  - [ ] `disable-model-invocation` explicitly marked Copilot-only
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.6: Add Dynamic Context Injection Syntax → §2.1.5

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.5
- **Description**: Document backtick-bang inline command execution in SKILL.md bodies and all
  string substitution variables (`$ARGUMENTS`, `$0/$N`, `${CLAUDE_SESSION_ID}`, `${CLAUDE_SKILL_DIR}`).
- **Source**: `docs/claude-structure-suggestions.md` Item 6
- **Acceptance Criteria**:
  - [ ] Backtick-bang syntax example added
  - [ ] Substitution variables table with all 4 entries
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.7: Add Skill Body Structure Template → §2.1.5

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.6
- **Description**: Document the recommended SKILL.md body structure: one-paragraph summary,
  `## Key Rules` (mandatory, 6-12 rules), `## Template / Workflow`, full reference link.
- **Source**: `docs/claude-structure-suggestions.md` Item 7
- **Acceptance Criteria**:
  - [ ] Recommended body structure added as a code block example
  - [ ] `## Key Rules` mandatory status reiterated
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.8: Add Sub-Agent Frontmatter Fields → §2.1.1

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.7
- **Description**: Add the complete Claude Code agent frontmatter table to §2.1.1 including
  `disallowedTools`, `permissionMode`, `maxTurns`, `skills`, `memory`, `color`. Note that
  `tools:` MUST be OMIT in Claude agents (inherits all).
- **Source**: `docs/claude-structure-suggestions.md` Item 8
- **Acceptance Criteria**:
  - [ ] All 11 frontmatter fields documented in table
  - [ ] `tools: OMIT` in Claude agents explicitly called out
  - [ ] Subagent isolation behavior (no parent conversation history) documented
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.9: Add Path-Scoped Rules (.claude/rules/) → §14.11

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.5h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.8
- **Description**: Document `.claude/rules/` auto-load behavior: `paths` frontmatter key
  for lazy loading vs. no `paths` for session-launch loading. Add recommended cryptoutil
  rule files (`framework.md` and `tests.md`).
- **Source**: `docs/claude-structure-suggestions.md` Item 9
- **Acceptance Criteria**:
  - [ ] Load behavior (with/without `paths`) documented
  - [ ] Example rule file format shown
  - [ ] Recommended cryptoutil rule files table added
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.10: Add agentskills.io Open Standard Context → §2.1.5

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.9
- **Description**: Document the agentskills.io open standard provenance, multi-tool adoption
  (Gemini CLI, Copilot, OpenAI Codex, Amp, Kiro, Qodo, VS Code), and shared frontmatter
  constraints (`name` ≤64 chars, `description` ≤1024 chars).
- **Source**: `docs/claude-structure-suggestions.md` Item 10
- **Acceptance Criteria**:
  - [ ] agentskills.io standard mentioned with tool adoption list
  - [ ] Shared frontmatter constraints documented
  - [ ] Cross-tool body identity requirement documented
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.11: Add CLAUDE.md Length and Scoping Strategy → §14.11

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.10
- **Description**: Document the scaling strategy for large monorepos: per-directory CLAUDE.md
  files, path-scoped rules for cryptoutil (`framework.md`, `tests.md`), and the adherence
  monitoring technique.
- **Source**: `docs/claude-structure-suggestions.md` Item 11
- **Acceptance Criteria**:
  - [ ] Subdirectory CLAUDE.md lazy loading documented
  - [ ] Cryptoutil-specific rule file recommendations added
  - [ ] Adherence monitoring tip (extract violated sections to rule files)
- **Files**: `docs/ENG-HANDBOOK.md`

### Task 4.12: Post-Phase 4 Lint Check

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Tasks 4.1–4.11
- **Description**: Run all relevant linters after Phase 4 additions.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
  - [ ] `go run ./cmd/cicd-lint lint-fitness` exits 0
- **Files**: None (verification only)

---

## Phase 5: Knowledge Propagation Verification

**Phase Objective**: Run `lint-docs` and verify zero propagation drift across all modified
ENG-HANDBOOK.md sections and their corresponding instruction files.

### Task 5.1: Run Full lint-docs Verification

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 0.25h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 4.12
- **Description**: Run the complete `lint-docs` pipeline to verify no propagation drift was
  introduced by the ENG-HANDBOOK.md additions in Phases 1–4.
- **Acceptance Criteria**:
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0
  - [ ] `validate-chunks` passes (no @propagate/source drift)
  - [ ] `validate-coverage` passes
  - [ ] `lint-agent-drift` passes (agent pairs in sync)
  - [ ] `lint-skill-command-drift` passes
- **Files**: None (verification only)

### Task 5.2: Fix Any Propagation Violations Found

- **Status**: ❌
- **Owner**: LLM Agent
- **Estimated**: 1h
- **Actual**: [Fill when complete]
- **Dependencies**: Task 5.1
- **Description**: If Task 5.1 finds any violations, fix them before declaring V18 complete.
  Common issues: new ENG-HANDBOOK.md content that duplicates an existing @propagate block
  (causing validate-chunks to detect drift), or new sections that should be @propagate targets.
- **Acceptance Criteria**:
  - [ ] All `lint-docs` violations resolved
  - [ ] `go run ./cmd/cicd-lint lint-docs` exits 0 (clean)
  - [ ] `go build ./...` exits 0
- **Files**: `docs/ENG-HANDBOOK.md`, instruction files (if propagation drift discovered)
