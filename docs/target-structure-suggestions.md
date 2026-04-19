# target-structure.md → ENG-HANDBOOK.md Suggestions

## Executive Summary

Analysis of [target-structure.md](target-structure.md) against [ENG-HANDBOOK.md §4.4](ENG-HANDBOOK.md#44-code-organization) reveals that the handbook covers broad structural patterns and naming conventions, but is missing concrete enumerations, file permission standards, and several inventory-level details that the target-structure document provides. The following additions are suggested.

1. [File Permission Convention Table](#1-file-permission-convention-table) — octal permissions for each file type (source, secret, executable) are not documented in the handbook.
2. [Root-Level File Inventory](#2-root-level-file-inventory) — the complete enumeration of 23+ root config files (`.air.toml`, `.gitleaks.toml`, etc.) is absent.
3. [Root Hidden Directory Inventory](#3-root-hidden-directory-inventory) — `.cicd-lint/`, `.semgrep/`, `.vscode/` (including `mcp.json`), `.well-known/`, and `.zap/` directories are not described.
4. [.github/ Top-Level File Catalog](#4-github-top-level-file-catalog) — `copilot-instructions.md`, `dependabot.yml`, `SECURITY.md`, and their peer files are not listed.
5. [GitHub Actions Catalog](#5-github-actions-catalog) — the 15 reusable actions under `.github/actions/` are not enumerated with purpose descriptions.
6. [Concrete Service Subdirectory Inventory](#6-concrete-service-subdirectory-inventory) — each PS-ID's actual subdirectories (e.g., pki-ca's 15+ dirs, identity-authz's dpop/pkce/unified dirs) are absent from the handbook.
7. [Identity Shared Package Catalog](#7-identity-shared-package-catalog) — the packages in `internal/apps/identity/` shared across all identity services are not listed.
8. [Complete magic/ File Listing](#8-complete-magic-file-listing) — the 42 `magic_*.go` files in `internal/shared/magic/` are enumerated in target-structure.md but not in the handbook.
9. [Other Top-Level Directories](#9-other-top-level-directories) — `scripts/`, `workflow-reports/`, `test-output/` (gitignored ephemeral dirs) and `pkg/` (reserved) are not described.
10. [Dockerfile Parameterization Table](#10-dockerfile-parameterization-table) — the table of image.title, binary path, EXPOSE, HEALTHCHECK, and ENTRYPOINT values per deployment tier is absent.
11. [Pending Work Inventory](#11-pending-work-inventory) — the documented gap (product-level Dockerfiles missing for all 5 products) is not captured in the handbook.

---

## Details

### 1. File Permission Convention Table

**Current state in ENG-HANDBOOK.md**: No file permission standards are stated.

**Suggested addition to §4.4.1 or §13**:

| Target | Permission | Octal | Description |
|--------|-----------|-------|-------------|
| Directories | `drwxr-x---` | 750 | Owner rwx, group rx, others no access |
| Source files (`.go`, `.yml`, `.yaml`, `.md`, `.sql`) | `-rw-r-----` | 640 | Owner rw, group r, others no access |
| Secret files (`.secret`) | `-r--r-----` | 440 | Owner/group read only, no other access |
| Secret marker files (`.secret.never`) | `-r--r-----` | 440 | Same as secrets — read-only reminder |
| Executable scripts (`mvnw`) | `-rwxr-x---` | 750 | Owner rwx, group rx, others no access |
| Generated files (`*.gen.go`) | `-rw-r-----` | 640 | Same as source files |

---

### 2. Root-Level File Inventory

**Current state in ENG-HANDBOOK.md**: No enumeration of root config files. The handbook describes patterns but not what files must exist.

**Suggested addition to §4.4.1** — reference or inline the canonical list:

Root files that MUST exist (any omission = configuration gap):

```
.air.toml            .dockerignore        .editorconfig
.gitattributes       .gitignore           .gitleaks.toml
.gofumpt.toml        .golangci.yml        .gremlins.yaml
.markdownlint.jsonc  .nuclei-ignore       .pre-commit-config.yaml
.rgignore            .sqlfluff            .yamlfmt
CLAUDE.md            go.mod               go.sum
LICENSE              NOTICE               pyproject.toml
README.md            robots.txt           TERMS.md
```

Root files that must NEVER be committed (build/test artifacts):
- `*.exe`, `*.py`, `coverage*`, `*_coverage`, `*.test.exe`

---

### 3. Root Hidden Directory Inventory

**Current state in ENG-HANDBOOK.md**: Not described.

**Suggested addition to §4.4**:

| Directory | Purpose |
|-----------|---------|
| `.cicd-lint/` | cicd-lint runtime caches (gitignored): `circular-dep-cache.json`, `dep-cache.json` |
| `.ruff_cache/` | Ruff Python linter cache (gitignored) |
| `.semgrep/rules/` | Semgrep SAST rules: `go-testing.yml` |
| `.vscode/` | VS Code workspace: `cspell.json`, `extensions.json`, `launch.json`, `mcp.json`, `settings.json` |
| `.well-known/` | Well-known URIs (RFC 8615): `tdm-reservation.txt` (Text & Data Mining reservation) |
| `.zap/` | OWASP ZAP DAST config: `rules.tsv` |

Note: `.vscode/mcp.json` configures MCP server integrations for Claude Code and Copilot Chat.

---

### 4. .github/ Top-Level File Catalog

**Current state in ENG-HANDBOOK.md**: Section 2.1 describes agents, skills, and instruction files but does not enumerate the top-level `.github/` files.

**Suggested addition to §2.1.4 or a new §2.1.0**:

| File | Purpose |
|------|---------|
| `copilot-instructions.md` | Copilot config hub: loads all instruction files via `@` imports |
| `dependabot.yml` | Dependabot automated dependency update configuration |
| `SECURITY.md` | Security policy and vulnerability reporting process |
| `versions-rules.xml` | Version constraint rules for dependency validation |
| `workflows-outdated-action-exemptions.json` | Exemptions for known-outdated workflow actions |

---

### 5. GitHub Actions Catalog

**Current state in ENG-HANDBOOK.md**: §B.7 lists reusable actions but without per-action purpose descriptions. The count (15) and download-cicd rename are not captured.

**Suggested addition to §B.7**:

| Action | Purpose |
|--------|---------|
| `docker-compose-build` | Build Docker images via Compose |
| `docker-compose-down` | Tear down Compose services |
| `docker-compose-logs` | Collect container logs |
| `docker-compose-up` | Start Compose services |
| `docker-compose-verify` | Verify Compose health |
| `docker-images-pull` | Parallel Docker image pre-pull |
| `download-cicd` | Download cicd-lint binary (renamed from `custom-cicd-lint`) |
| `fuzz-test` | Execute fuzz tests |
| `go-setup` | Go toolchain setup with caching |
| `golangci-lint` | golangci-lint v2 execution |
| `security-scan-gitleaks` | Gitleaks secret detection |
| `security-scan-trivy` | Trivy: manual install + CLI (supports scan-files) |
| `security-scan-trivy2` | Trivy: official aquasecurity/trivy-action (simpler) |
| `workflow-job-begin` | Job telemetry start |
| `workflow-job-end` | Job telemetry end |

---

### 6. Concrete Service Subdirectory Inventory

**Current state in ENG-HANDBOOK.md**: §4.4.4 describes the generic service pattern (`server/`, `service/`, `repository/`, etc.) but does not enumerate the actual subdirectories discovered in each PS-ID.

**Suggested addition to §4.4.4**:

| PS-ID | Actual Subdirectories |
|-------|----------------------|
| `identity-authz` | `clientauth/`, `dpop/`, `e2e/`, `pkce/`, `server/`, `unified/` |
| `identity-idp` | `auth/`, `server/`, `templates/`, `unified/`, `userauth/` |
| `identity-rp` | `server/`, `unified/` |
| `identity-rs` | `server/`, `unified/` |
| `identity-spa` | `server/`, `unified/` |
| `jose-ja` | `e2e/`, `model/`, `repository/`, `server/`, `service/` |
| `pki-ca` | `api/`, `bootstrap/`, `cli/`, `compliance/`, `config/`, `crypto/`, `domain/`, `domain-v2/`, `intermediate/`, `observability/`, `profile/`, `repository-v2/`, `security/`, `server/`, `service/`, `storage/` |
| `skeleton-template` | `domain/`, `e2e/`, `repository/`, `server/` |
| `sm-im` | `client/`, `e2e/`, `integration/`, `model/`, `repository/`, `server/`, `testing/` |
| `sm-kms` | `client/`, `e2e/`, `server/` |

---

### 7. Identity Shared Package Catalog

**Current state in ENG-HANDBOOK.md**: The handbook mentions `internal/apps/identity/` exists for shared identity code but does not list the packages.

**Suggested addition to §4.4.4**:

Identity shared packages at `internal/apps/identity/` (shared across all 5 identity services):

| Package | Purpose |
|---------|---------|
| `apperr/` | Identity-specific error types |
| `config/` | Shared identity configuration |
| `domain/` | Shared identity domain types |
| `email/` | Email sending |
| `issuer/` | Token issuer |
| `jobs/` | Background jobs |
| `mfa/` | Multi-factor authentication |
| `repository/` (with `orm/`, `migrations/`) | Shared identity data access |
| `rotation/` | Key/token rotation |

---

### 8. Complete magic/ File Listing

**Current state in ENG-HANDBOOK.md**: §11.1.4 states "ALL magic constants MUST be in `internal/shared/magic/`" but does not enumerate the files.

**Suggested addition to §11.1.4** (or §4.4.5):

`internal/shared/magic/` contains 42 `magic_*.go` files organized by domain:

| File | Domain |
|------|--------|
| `magic_api.go`, `magic_cicd.go`, `magic_console.go` | Infrastructure |
| `magic_crypto.go`, `magic_database.go`, `magic_docker.go` | Core systems |
| `magic_framework.go`, `magic_memory.go`, `magic_misc.go` | Framework |
| `magic_network.go`, `magic_orchestration.go`, `magic_percent.go` | Network/ops |
| `magic_security.go`, `magic_session.go`, `magic_telemetry.go` | Security/obs |
| `magic_testing.go`, `magic_test_fixtures.go` | Testing |
| `magic_unseal.go`, `magic_workflows.go` | Deployment |
| `magic_identity*.go` (11 files) | Identity product |
| `magic_jose.go` | JOSE product |
| `magic_pki.go`, `magic_pkiinit.go`, `magic_pkix.go`, `magic_pki_ca.go` | PKI product |
| `magic_skeleton.go` | Skeleton product |
| `magic_sm.go`, `magic_sm_im.go` | SM product |
| `magic_cicd_test.go` | Test for cicd constants |

---

### 9. Other Top-Level Directories

**Current state in ENG-HANDBOOK.md**: Not described.

**Suggested addition to §4.4.1**:

| Directory | Status | Purpose |
|-----------|--------|---------|
| `scripts/` | Gittracked, empty | Reserved for Go-based scripts; contains `.gitkeep` only. Part of standard Go project layout. |
| `workflow-reports/` | Gitignored | Ephemeral GitHub Actions workflow output; never committed. |
| `test-output/` | Gitignored | Ephemeral test output (autoapprove logs, coverage reports); never committed. |
| `pkg/` | Gittracked, empty | Reserved for future public library APIs; currently empty. |

---

### 10. Dockerfile Parameterization Table

**Current state in ENG-HANDBOOK.md**: §12.2.1 describes the 4-stage Dockerfile pattern but does not tabulate the per-tier differences in image labels, binary path, EXPOSE, HEALTHCHECK, and ENTRYPOINT.

**Suggested addition to §12.2.1**:

| Field | Service (`{PS-ID}`) | Product (`{PRODUCT}`) | Suite (`{SUITE}`) |
|-------|---------------------|----------------------|-------------------|
| `image.title` LABEL | `{SUITE}-{PS-ID}` | `{SUITE}-{PRODUCT}` | `{SUITE}` |
| Binary built | `./cmd/{PS-ID}` | `./cmd/{SUITE}` | `./cmd/{SUITE}` |
| `EXPOSE` | `8080` | Product-range (e.g., `18000`) | Suite-range (e.g., `28000`) |
| `HEALTHCHECK` | `/app/{PS-ID} livez` | Same pattern, product port | Same pattern, suite port |
| `ENTRYPOINT` | `["/sbin/tini", "--", "/app/{PS-ID}"]` | `["/sbin/tini", "--", "/app/{SUITE}", "{PRODUCT}", "start"]` | `["/sbin/tini", "--", "/app/{SUITE}"]` |

Current state: 10 service Dockerfiles + 1 suite Dockerfile exist. Product-level Dockerfiles (5) are pending creation.

---

### 11. Pending Work Inventory

**Current state in ENG-HANDBOOK.md**: No explicit tracking of known structural gaps.

**Suggested addition to §4.4.6 or §12.2**:

| Area | Current State | Target State | Action |
|------|--------------|-------------|--------|
| `deployments/{PRODUCT}/Dockerfile` | Missing in all 5 products | Present in all 5 products | CREATE — blocked pending suite binary architecture decision |

Note: Product Dockerfiles may become necessary if the architecture migrates to a single suite binary (`./cmd/cryptoutil`). Currently each PS-ID builds its own binary, making product-level Dockerfiles redundant.
