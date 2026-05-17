# Target Repository Structure

**Status**: CANONICAL TARGET — Living reference document
**Created**: 2026-03-26
**Last Updated**: 2026-04-05
**Purpose**: Define the complete, parameterized target state of every directory and file in the
repository. Originally created during framework-v6, now maintained as a living spec in framework-v8.
This document supersedes framework-v5/target-structure.md (deleted — git history preserves).

**RULE**: Everything listed here MUST exist. Everything NOT listed is deleted.

**Directory/File Count Derivation Principle**: All file and directory counts in this document MUST be shown as a formula derived from the entity multipliers above (e.g., `4 global + 12×10 PS-IDs = 124`). Raw counts without formulas are unverifiable during review.

---

## Entity Hierarchy (Canonical)

| Level | Variable | Instances | Count |
|-------|----------|-----------|-------|
| Suite | `{SUITE}` | `cryptoutil` | 1 |
| Product | `{PRODUCT}` | `sm`, `jose`, `pki`, `identity`, `skeleton` | 5 |
| Service | `{SERVICE}` | varies per product (see below) | 10 total |
| PS-ID | `{PS-ID}` = `{PRODUCT}-{SERVICE}` | see table below | 10 |
| PS_ID | `{PS_ID}` = `{PRODUCT}_{SERVICE}` | underscore variant for SQL/secrets | 10 |
| Infra Tool | N/A | `cicd-lint`, `cicd-workflow` | 2 |
| Framework | N/A | `framework` | 1 |

### Product-Service Matrix

| PS-ID | PS_ID | Product | Service | Display Name |
|-------|-------|---------|---------|-------------|
| `sm-kms` | `sm_kms` | sm | kms | Secrets Manager Key Management |
| `sm-im` | `sm_im` | sm | im | Secrets Manager Instant Messenger |
| `jose-ja` | `jose_ja` | jose | ja | JOSE JWK Authority |
| `pki-ca` | `pki_ca` | pki | ca | PKI Certificate Authority |
| `identity-authz` | `identity_authz` | identity | authz | Identity Authorization Server |
| `identity-idp` | `identity_idp` | identity | idp | Identity Provider |
| `identity-rs` | `identity_rs` | identity | rs | Identity Resource Server |
| `identity-rp` | `identity_rp` | identity | rp | Identity Relying Party |
| `identity-spa` | `identity_spa` | identity | spa | Identity Single Page App |
| `skeleton-template` | `skeleton_template` | skeleton | template | Skeleton Template |

### Permission Convention

| Target | Permission | Octal | Description |
|--------|-----------|-------|-------------|
| Directories | `drwxr-x---` | 750 | Owner rwx, group rx, others no access |
| Source files (`.go`, `.yml`, `.yaml`, `.md`, `.sql`) | `-rw-r-----` | 640 | Owner rw, group r, others no access |
| Secret files (`.secret`) | `-r--r-----` | 440 | Owner/group r only, no other |
| Secret marker files (`.secret.never`) | `-r--r-----` | 440 | Same as secrets |
| Executable scripts (`mvnw`) | `-rwxr-x---` | 750 | Owner rwx, group rx, others no access |
| Generated files (`*.gen.go`) | `-rw-r-----` | 640 | Same as source |

---

## A. Root Level

### A.1 Root Files (KEEP — legitimate project config) `drwxr-x---`

```
{ROOT}/                                    # drwxr-x---
├── .air.toml                              # Air live-reload config
├── .dockerignore                          # Docker build context exclusions
├── .editorconfig                          # Editor formatting standards (indent, line endings)
├── .gitattributes                         # Git line ending and diff config
├── .gitignore                             # Git ignore rules
├── .gitleaks.toml                         # Gitleaks secret detection config
├── .gofumpt.toml                          # gofumpt Go formatting config
├── .golangci.yml                          # golangci-lint v2 linter config
├── .gremlins.yaml                         # Gremlins mutation testing config
├── .markdownlint.jsonc                    # Markdown linting rules
├── .nuclei-ignore                         # Nuclei DAST scan exclusions
├── .pre-commit-config.yaml                # Pre-commit hook definitions
├── .rgignore                              # ripgrep ignore patterns
├── .sqlfluff                              # SQL linting config
├── .yamlfmt                               # yamlfmt YAML formatter config
├── CLAUDE.md                              # Claude Code project instructions
├── go.mod                                 # Go module definition
├── go.sum                                 # Go module dependency checksums
├── LICENSE                                # Project license
├── NOTICE                                 # Third-party attribution notices
├── pyproject.toml                         # Python project config (pre-commit tooling)
├── README.md                              # Project README
├── robots.txt                             # Web crawler control
└── TERMS.md                               # Terms of service
```

### A.2 Root Files (DELETE — junk artifacts)

All `*.exe`, `*.py`, `coverage*`, `*_coverage`, `*.test.exe` files at root are build/test
artifacts that must never be committed.

### A.3 Root Hidden Directories `drwxr-x---`

```
{ROOT}/
├── .cicd-lint/                             # CICD-lint runtime caches (gitignored)
│   ├── circular-dep-cache.json            #   Circular dependency analysis cache
│   └── dep-cache.json                     #   Dependency analysis cache
├── .ruff_cache/                           # Ruff Python linter cache (gitignored)
├── .semgrep/                              # Semgrep SAST rules
│   └── rules/
│       └── go-testing.yml                 #   Go testing SAST rules
├── .vscode/                               # VS Code workspace settings
│   ├── cspell.json                        #   Spell checking dictionary
│   ├── extensions.json                    #   Recommended extensions
│   ├── launch.json                        #   Debug launch configs
│   ├── mcp.json                           #   MCP server configuration (v6 NEW)
│   └── settings.json                      #   Workspace settings
├── .well-known/                           # Well-known URIs (RFC 8615)
│   └── tdm-reservation.txt               #   Text & Data Mining reservation
└── .zap/                                  # OWASP ZAP DAST config
    └── rules.tsv                          #   ZAP scan rules
```

---

## B. .github/ & .claude/ — GitHub, Copilot & Claude Configuration `drwxr-x---`

### B.0 Top-Level .github/ Files

```
.github/
├── copilot-instructions.md                # Copilot config hub (loads instructions/)
├── dependabot.yml                         # Dependabot automated dependency updates
├── SECURITY.md                            # Security policy and vulnerability reporting
├── versions-rules.xml                     # Version constraint rules
└── workflows-outdated-action-exemptions.json  # Exemptions for outdated workflow actions
```

### B.1 Agents (4 agents — `doc-sync` deleted)

```
.github/agents/
├── beast-mode.agent.md                    # Continuous autonomous execution
├── fix-workflows.agent.md                 # CI/CD workflow fixer
├── implementation-execution.agent.md      # Plan execution agent
└── implementation-planning.agent.md       # Plan creation agent
```

### B.2 Actions (15 actions — `download-cicd` replaces `custom-cicd-lint`)

```
.github/actions/
├── docker-compose-build/action.yml
├── docker-compose-down/action.yml
├── docker-compose-logs/action.yml
├── docker-compose-up/action.yml
├── docker-compose-verify/action.yml
├── docker-images-pull/action.yml          # Parallel Docker image pre-pull
├── download-cicd/action.yml               # Download cicd-lint binary (was custom-cicd-lint)
├── fuzz-test/action.yml
├── go-setup/action.yml                    # Go toolchain setup with caching
├── golangci-lint/action.yml               # golangci-lint v2 execution
├── security-scan-gitleaks/action.yml
├── security-scan-trivy/action.yml         # Manual Trivy install + CLI (supports scan-files)
├── security-scan-trivy2/action.yml        # Official aquasecurity/trivy-action (simpler)
├── workflow-job-begin/action.yml          # Job telemetry start
└── workflow-job-end/action.yml            # Job telemetry end
```

### B.3 Instructions (18 files)

```
.github/instructions/
├── 01-01.terminology.instructions.md
├── 01-02.beast-mode.instructions.md
├── 02-01.architecture.instructions.md
├── 02-02.versions.instructions.md
├── 02-03.observability.instructions.md
├── 02-04.openapi.instructions.md
├── 02-05.security.instructions.md
├── 02-06.authn.instructions.md
├── 03-01.coding.instructions.md
├── 03-02.testing.instructions.md
├── 03-03.golang.instructions.md
├── 03-04.data-infrastructure.instructions.md
├── 03-05.linting.instructions.md
├── 04-01.deployment.instructions.md
├── 05-01.cross-platform.instructions.md
├── 05-02.git.instructions.md
├── 06-01.evidence-based.instructions.md
└── 06-02.agent-format.instructions.md
```

### B.4 Skills (13 skills + README)

```
.github/skills/
├── README.md
├── psid-template-sync/SKILL.md
├── coverage-analysis/SKILL.md
├── customization-scaffold/SKILL.md
├── fips-audit/SKILL.md
├── fitness-function-gen/SKILL.md
├── migration-create/SKILL.md
├── new-service/SKILL.md
├── openapi-codegen/SKILL.md
├── propagation-check/SKILL.md
├── sync-copilot-claude/SKILL.md
├── test-benchmark-gen/SKILL.md
├── test-fuzz-gen/SKILL.md
└── test-table-driven/SKILL.md
```

### B.5 Workflows (15 workflows)

```
.github/workflows/
├── ci-benchmark.yml                       # Benchmark testing
├── ci-coverage.yml                        # Code coverage analysis
├── ci-dast.yml                            # Dynamic application security testing
├── ci-e2e.yml                             # End-to-end testing
├── ci-fitness.yml                         # Architecture fitness functions
├── ci-fuzz.yml                            # Fuzz testing
├── ci-github-cleanup.yml                  # GitHub Actions storage cleanup
├── ci-gitleaks.yml                        # Secret detection
├── ci-identity-validation.yml             # Identity service validation
├── ci-load.yml                            # Load testing (Gatling)
├── ci-mutation.yml                        # Mutation testing (gremlins)
├── ci-quality.yml                         # Build + lint + unit tests (includes cicd-lint)
├── ci-race.yml                            # Race condition detection
├── ci-sast.yml                            # Static application security testing
└── release.yml                            # Release workflow
```

**NOTE**: The `ci-cicd-lint.yml` separate workflow is consolidated INTO `ci-quality.yml` as a
job step. No standalone cicd-lint workflow in target state.

### B.6 .claude/ — Claude Code Configuration (Dual Canonical Pairs)

Every Copilot agent and skill has a Claude Code counterpart. See `06-02.agent-format.instructions.md`
for the dual canonical file strategy and drift linting (`lint-agent-drift`, `lint-skill-command-drift`).

```
.claude/
├── settings.local.json                    # Claude Code workspace settings
├── agents/                                # Claude agents (4 — mirrors .github/agents/)
│   ├── beast-mode.md
│   ├── fix-workflows.md
│   ├── implementation-execution.md
│   └── implementation-planning.md
└── skills/                                # Claude skills (13 — mirrors .github/skills/)
    ├── psid-template-sync/SKILL.md
    ├── coverage-analysis/SKILL.md
    ├── customization-scaffold/SKILL.md
    ├── fips-audit/SKILL.md
    ├── fitness-function-gen/SKILL.md
    ├── migration-create/SKILL.md
    ├── new-service/SKILL.md
    ├── openapi-codegen/SKILL.md
    ├── propagation-check/SKILL.md
    ├── sync-copilot-claude/SKILL.md
    ├── test-benchmark-gen/SKILL.md
    ├── test-fuzz-gen/SKILL.md
    └── test-table-driven/SKILL.md
```

---

## C. cmd/ — Binary Entry Points `drwxr-x---`

**Pattern**: Flat directories — every entry is a direct child of `cmd/`. No nesting.
Each entry has exactly one `main.go` that delegates to `internal/apps/`.

**Canonical templates**: `api/cryptosuite-registry/templates/cmd/{__PS_ID__,__PRODUCT__,__SUITE__}/main.go`
enforced by lint-fitness `cmd-ps-id-template`, `cmd-product-template`, `cmd-suite-template`.

**Rigid structure (all three types)**:

| Type | Required file | Invariants |
|------|--------------|------------|
| `cmd/{PS-ID}/` | `main.go` | `package main`; imports `os` + `cryptoutil/internal/apps/{PS-ID}`; calls `os.Exit(<alias>.<PascalService>(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))` |
| `cmd/{PRODUCT}/` | `main.go` | `package main`; imports `os` + `cryptoutil/internal/apps/{PRODUCT}`; calls `os.Exit(<alias>.<PascalProduct>(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))` |
| `cmd/{SUITE}/` | `main.go` | `package main`; imports `os` + `cryptoutil/internal/apps/{SUITE}`; calls `os.Exit(<alias>.Suite(os.Args, os.Stdin, os.Stdout, os.Stderr))` — uses full `os.Args`, NOT `os.Args[1:]` |

```
cmd/                                                  # drwxr-x---  (18 flat entries)
│
│   # {SUITE}/main.go — Suite CLI → internal/apps/{SUITE}/ (×1)
├── cryptoutil/main.go                                # {SUITE}=cryptoutil
│
│   # {PRODUCT}/main.go — Product CLI → internal/apps/{PRODUCT}/ (×5)
├── identity/main.go                                  # {PRODUCT}=identity
├── jose/main.go                                      # {PRODUCT}=jose
├── pki/main.go                                       # {PRODUCT}=pki
├── skeleton/main.go                                  # {PRODUCT}=skeleton
├── sm/main.go                                        # {PRODUCT}=sm
│
│   # {PS-ID}/main.go — Service CLI → internal/apps/{PS-ID}/ (×10)
├── identity-authz/main.go                            # {PS-ID}=identity-authz
├── identity-idp/main.go                              # {PS-ID}=identity-idp
├── identity-rp/main.go                               # {PS-ID}=identity-rp
├── identity-rs/main.go                               # {PS-ID}=identity-rs
├── identity-spa/main.go                              # {PS-ID}=identity-spa
├── jose-ja/main.go                                   # {PS-ID}=jose-ja
├── pki-ca/main.go                                    # {PS-ID}=pki-ca
├── skeleton-template/main.go                         # {PS-ID}=skeleton-template
├── sm-im/main.go                                     # {PS-ID}=sm-im
├── sm-kms/main.go                                    # {PS-ID}=sm-kms
│
│   # {INFRA-TOOL}/main.go — Infrastructure tools (×2)
├── cicd-lint/main.go                                 # {INFRA-TOOL}=cicd-lint
└── cicd-workflow/main.go                             # {INFRA-TOOL}=cicd-workflow
```

**Total**: 18 flat entries (1 suite + 5 products + 10 services + 2 infra tools).

---

## D. api/ — OpenAPI Specs and Generated Code `drwxr-x---`

**Pattern**: One directory per PS-ID. Each contains the OpenAPI spec files and oapi-codegen
generated code. Plus a `cryptosuite-registry/` directory for the machine-readable entity registry.

```
api/                                                  # drwxr-x---
├── cryptosuite-registry/                             # Machine-readable entity registry (SSOT)
│   ├── registry.yaml                                 #   Canonical YAML entity registry
│   ├── registry-schema.json                          #   JSON Schema validating registry.yaml
│   └── templates/                                    #   Parameterized canonical deployment templates
│       ├── configs/
│       │   └── __PS_ID__/                            #     Standalone config templates (×1, expands to ×10)
│       │       └── __PS_ID__-framework.yml           #       Framework config template
│       └── deployments/
│           ├── __PS_ID__/                            #     PS-ID templates (×1, expands to ×10)
│           │   ├── Dockerfile                        #       Dockerfile template
│           │   ├── compose.yml                       #       Compose template
│           │   ├── config/                           #       Config overlay templates
│           │   │   ├── __PS_ID__-app-framework-common.yml
│           │   │   ├── __PS_ID__-app-framework-sqlite-1.yml
│           │   │   ├── __PS_ID__-app-framework-sqlite-2.yml
│           │   │   ├── __PS_ID__-app-framework-postgresql-1.yml
│           │   │   └── __PS_ID__-app-framework-postgresql-2.yml
│           │   └── secrets/                          #       Secrets templates (15 files)
│           │       ├── unseal-{1..5}of5.secret       #         Unseal key shards
│           │       ├── hash-pepper-v3.secret         #         Hash pepper
│           │       ├── postgres-{url,username,password,database}.secret
│           │       ├── browser-{username,password}.secret
│           │       ├── service-{username,password}.secret
│           │       └── issuing-ca-key.secret         #         PKI CA key (PS-ID level only)
│           ├── __PRODUCT__/                          #     Product templates (×1, expands to ×5)
│           │   ├── compose.yml                       #       Product compose template
│           │   └── secrets/                          #       Secrets templates (15 files)
│           │       ├── unseal-{1..5}of5.secret
│           │       ├── hash-pepper-v3.secret
│           │       ├── postgres-{url,username,password,database}.secret
│           │       ├── browser-{username,password}.secret.never  # Marker only
│           │       ├── service-{username,password}.secret.never  # Marker only
│           │       └── issuing-ca-key.secret.never              # Marker only
│           ├── __SUITE__/                            #     Suite templates (×1)
│           │   ├── compose.yml                       #       Suite compose template
│           │   └── secrets/                          #       Secrets templates (15 files)
│           │       ├── unseal-{1..5}of5.secret
│           │       ├── hash-pepper-v3.secret
│           │       ├── postgres-{url,username,password,database}.secret
│           │       ├── browser-{username,password}.secret.never
│           │       ├── service-{username,password}.secret.never
│           │       └── issuing-ca-key.secret.never
│           ├── shared-postgres/                      #     Shared PostgreSQL static templates
│           │   ├── compose.yml
│           │   ├── init-databases.sql
│           │   ├── init-users.sql
│           │   ├── postgresql-leader.conf
│           │   ├── postgresql-follower.conf
│           │   ├── setup-logical-replication.sh
│           │   └── secrets/                          #       postgres-{database,username,password}.secret
│           └── shared-telemetry/                     #     Shared telemetry static templates
│               ├── compose.yml
│               └── otel/
│                   └── otel-collector-config.yaml
├── {PS-ID}/                                          # One dir per service (×10)
│   ├── openapi_spec_components.yaml                  #   Reusable components
│   ├── openapi_spec_paths.yaml                       #   API endpoints
│   ├── openapi-gen_config_client.yaml                #   oapi-codegen client config
│   ├── openapi-gen_config_model.yaml                 #   oapi-codegen model config
│   ├── openapi-gen_config_server.yaml                #   oapi-codegen server config
│   ├── client/                                       #   Generated client code
│   │   └── client.gen.go
│   ├── model/                                        #   Generated model code
│   │   └── models.gen.go
│   └── server/                                       #   Generated server code
│       └── server.gen.go
```

**All 10 PS-IDs**: `identity-authz`, `identity-idp`, `identity-rp`, `identity-rs`,
`identity-spa`, `jose-ja`, `pki-ca`, `skeleton-template`, `sm-im`, `sm-kms`.

---

## E. configs/ — Service Configuration Files `drwxr-x---`

### E.1 Suite Config

**Pattern**: `configs/{SUITE}/{SUITE}.yml`

```
configs/
└── {SUITE}/
    └── {SUITE}.yml                        # Suite-level config (logging, telemetry)
```

**Concrete** (`{SUITE}=cryptoutil`):

```
configs/
└── cryptoutil/
    └── cryptoutil.yml
```

### E.2 Product Configs — NOT APPLICABLE

Product-level config directories (`configs/{PRODUCT}/{PRODUCT}.yml`) are NOT used.
Products (cmd/identity, cmd/jose, etc.) are CLI dispatchers that recurse to their
constituent service binaries — they do not have their own config files. All config
is at the service level (E.3) or suite level (E.1).

### E.3 Service Configs (10 services — FLAT `configs/{PS-ID}/`)

Each service has its own flat directory at `configs/{PS-ID}/` containing exactly
one config file named `{PS-ID}.yml`. NO nested product subdirectories.

Config file name pattern: `{PS-ID}.yml` (e.g., `sm-im.yml`, NOT `im.yml`).

```
configs/
├── identity-authz/
│   ├── identity-authz.yml                 # Service config for identity-authz
│   └── domain/                            # Exception: authorization domain configs (Decision 4=A)
│       └── policies/
│           ├── adaptive-authorization.yml # RENAMED from adaptive-auth.yml (`auth` is a banned term because it is ambiguous)
│           ├── risk-scoring.yml
│           └── step-up.yml
├── identity-idp/
│   └── identity-idp.yml
├── identity-rp/
│   └── identity-rp.yml
├── identity-rs/
│   └── identity-rs.yml
├── identity-spa/
│   └── identity-spa.yml
├── jose-ja/
│   └── jose-ja.yml
├── pki-ca/
│   ├── pki-ca.yml
│   └── profiles/                          # Exception: certificate profiles (Decision 3=B)
│       │                                  # 25 YAML certificate profile definitions;
│       │                                  # valid subdir because they are real config data,
│       │                                  # NOT deployment variants or schema
│       └── (25 *.yaml profile files)      # e.g. root-ca.yaml, tls-server.yaml, etc.
├── skeleton-template/
│   └── skeleton-template.yml
├── sm-im/
│   └── sm-im.yml
└── sm-kms/
    └── sm-kms.yml
```

---

## F. deployments/ — Service Deployments `drwxr-x---`

### F.1 Per-Service Deployment (10 services × identical pattern)

Each service has exactly the same structure. 5 config overlay files (NOT 4).

```
deployments/{PS-ID}/                                  # drwxr-x---
├── compose.yml                                       # Docker Compose service definition
├── Dockerfile                                        # Service Docker image build
├── config/
│   ├── {PS-ID}-app-common.yml                        #   Common: bind addresses, TLS, network
│   ├── {PS-ID}-app-sqlite-1.yml                      #   SQLite in-memory instance 1
│   ├── {PS-ID}-app-sqlite-2.yml                      #   SQLite in-memory instance 2 (REQUIRED)
│   ├── {PS-ID}-app-postgresql-1.yml                  #   PostgreSQL logical instance 1
│   └── {PS-ID}-app-postgresql-2.yml                  #   PostgreSQL logical instance 2
└── secrets/                                          # 14 secret files
    ├── hash-pepper-v3.secret                         #   {PS-ID}-hash-pepper-v3-{base64-random-32-bytes}
    ├── browser-username.secret                       #   {PS-ID}-browser-user
    ├── browser-password.secret                       #   {PS-ID}-browser-pass-{base64-random-32-bytes}
    ├── service-username.secret                       #   {PS-ID}-service-user
    ├── service-password.secret                       #   {PS-ID}-service-pass-{base64-random-32-bytes}
    ├── postgres-username.secret                      #   {PS_ID}_database_user
    ├── postgres-password.secret                      #   {PS_ID}_database_pass-{base64-random-32-bytes}
    ├── postgres-database.secret                      #   {PS_ID}_database
    ├── postgres-url.secret                           #   postgres://{PS_ID}_database_user:{PS_ID}_database_pass@{PS-ID}-postgres:5432/{PS_ID}_database?sslmode=disable
    ├── unseal-1of5.secret                            #   {PS-ID}-unseal-key-1-of-5-{base64-random-32-bytes}
    ├── unseal-2of5.secret                            #   {PS-ID}-unseal-key-2-of-5-{base64-random-32-bytes}
    ├── unseal-3of5.secret                            #   {PS-ID}-unseal-key-3-of-5-{base64-random-32-bytes}
    ├── unseal-4of5.secret                            #   {PS-ID}-unseal-key-4-of-5-{base64-random-32-bytes}
    └── unseal-5of5.secret                            #   {PS-ID}-unseal-key-5-of-5-{base64-random-32-bytes}
```

**All 10 services** (`identity-authz`, `identity-idp`, `identity-rp`, `identity-rs`,
`identity-spa`, `jose-ja`, `pki-ca`, `skeleton-template`, `sm-im`, `sm-kms`) follow
this identical structure.

### F.2 Per-Product Deployment (5 products)

Each product has a deployment directory with a compose.yml and secrets. Product-level
Dockerfiles are intentionally absent; PRODUCT domains reuse included PS-ID builders and PS-ID images.

```
deployments/{PRODUCT}/                                # drwxr-x---
├── compose.yml                                       # Product-level Docker Compose
└── secrets/
    ├── hash-pepper-v3.secret                         # {PRODUCT}-hash-pepper-v3-{base64-random-32-bytes}
    ├── browser-username.secret.never                 # MUST use `.never` filename extension at product level; these are service-level creds only
    ├── browser-password.secret.never                 # MUST use `.never` filename extension at product level; these are service-level creds only
    ├── service-username.secret.never                 # MUST use `.never` filename extension at product level; these are service-level creds only
    ├── service-password.secret.never                 # MUST use `.never` filename extension at product level; these are service-level creds only
    ├── postgres-username.secret                      # {PRODUCT}_database_user
    ├── postgres-password.secret                      # {PRODUCT}_database_pass-{base64-random-32-bytes}
    ├── postgres-database.secret                      # {PRODUCT}_database
    ├── postgres-url.secret                           # postgres://{PRODUCT}_database_user:{PRODUCT}_database_pass@{PRODUCT}-postgres:5432/{PRODUCT}_database?sslmode=disable
    ├── unseal-1of5.secret                            # {PRODUCT}-unseal-key-1-of-5-{base64-random-32-bytes}
    ├── unseal-2of5.secret                            # {PRODUCT}-unseal-key-2-of-5-{base64-random-32-bytes}
    ├── unseal-3of5.secret                            # {PRODUCT}-unseal-key-3-of-5-{base64-random-32-bytes}
    ├── unseal-4of5.secret                            # {PRODUCT}-unseal-key-4-of-5-{base64-random-32-bytes}
    └── unseal-5of5.secret                            # {PRODUCT}-unseal-key-5-of-5-{base64-random-32-bytes}
```

**Total per product**: 4 `.secret.never` + 10 `.secret` = 14 files + compose.yml.

**All 5 products** (`identity`, `jose`, `pki`, `skeleton`, `sm`) follow this identical structure.

### F.3 Suite Deployment

**Pattern**: `deployments/{SUITE}/`

The suite deployment directory uses the bare `{SUITE}` name (e.g., `cryptoutil`),
consistent with all other naming conventions. Contains compose.yml and secrets.

```
deployments/{SUITE}/                                  # drwxr-x---
├── compose.yml                                       # Suite-level Docker Compose
└── secrets/
    ├── hash-pepper-v3.secret                         # {SUITE}-hash-pepper-v3-{base64-random-32-bytes}
    ├── browser-username.secret.never                 # MUST use `.never` filename extension at suite level; these are service-level creds only
    ├── browser-password.secret.never                 # MUST use `.never` filename extension at suite level; these are service-level creds only
    ├── service-username.secret.never                 # MUST use `.never` filename extension at suite level; these are service-level creds only
    ├── service-password.secret.never                 # MUST use `.never` filename extension at suite level; these are service-level creds only
    ├── postgres-username.secret                      # {SUITE}_database_user
    ├── postgres-password.secret                      # {SUITE}_database_pass-{base64-random-32-bytes}
    ├── postgres-database.secret                      # {SUITE}_database
    ├── postgres-url.secret                           # postgres://{SUITE}_database_user:{SUITE}_database_pass@{SUITE}-postgres:5432/{SUITE}_database?sslmode=disable
    ├── unseal-1of5.secret                            # {SUITE}-unseal-key-1-of-5-{base64-random-32-bytes}
    ├── unseal-2of5.secret                            # {SUITE}-unseal-key-2-of-5-{base64-random-32-bytes}
    ├── unseal-3of5.secret                            # {SUITE}-unseal-key-3-of-5-{base64-random-32-bytes}
    ├── unseal-4of5.secret                            # {SUITE}-unseal-key-4-of-5-{base64-random-32-bytes}
    └── unseal-5of5.secret                            # {SUITE}-unseal-key-5-of-5-{base64-random-32-bytes}
```

**Total**: 4 `.secret.never` + 10 `.secret` = 14 files + compose.yml.

### F.4 Shared Infrastructure Deployments

```
deployments/
├── shared-telemetry/
│   └── compose.yml                                   # otel-collector-contrib + grafana-otel-lgtm
└── shared-postgres/
    └── compose.yml                                   # Shared PostgreSQL container
                                                      # Every service gets a logical database in this
                                                      # instance; credentials shared at suite/product/
                                                      # service level as appropriate
```

**`/certs` Docker Volume**: Each PS-ID's `pki-init` init-container generates a TLS certificate tree
into a named Docker volume mounted at `/certs`. The directory layout follows the 14-category
keystore/truststore pattern defined in [`docs/tls-structure.md`](tls-structure.md). Per PS-ID scope:
90 directories (assuming 2 realms; count is `|realms|`-dependent — see the Directory Count Summary in `tls-structure.md`), each
containing `SAME-AS-DIR-NAME.{p12,crt,key}` (keystores) or `SAME-AS-DIR-NAME.{p12,crt}`
(truststores). See [`docs/tls-structure.md`](tls-structure.md) for the full specification.

### F.5 Dockerfile Parameterization

All Dockerfiles are PS-ID Dockerfiles. PRODUCT and SUITE deployment domains are compose-only layers
that reuse PS-ID builder services and PS-ID images.

**Pattern**: `deployments/{PS-ID}/Dockerfile`

| Field | PS-ID Dockerfile |
|-------|------------------|
| `image.title` LABEL | `{SUITE}-{PS-ID}` |
| `image.description` LABEL | Service-specific description |
| Binary built | `./cmd/{PS-ID}` |
| `EXPOSE` | 8080 (container public) |
| `HEALTHCHECK` | `CMD /app/{PS-ID} livez || exit 1` |
| `ENTRYPOINT` | `["/sbin/tini", "--", "/app/{PS-ID}"]` |

**Current state**: 10 PS-ID Dockerfiles exist. 0 product-level Dockerfiles exist. 0 suite-level Dockerfiles exist. This is the intended 10-image deployment model.

---

## G. internal/ — Private Application Code `drwxr-x---`

### G.1 internal/apps/ — Application Layer

**Structure**: `internal/apps/{SUITE | PRODUCT | PS-ID | framework | tools}`

Services live at flat `internal/apps/{PS-ID}/` (NOT nested under their product).
`cmd/{PS-ID}/main.go` delegates to `internal/apps/{PS-ID}/{PS-ID}.go`.
Product directories (`internal/apps/{PRODUCT}/`) contain ONLY product-level code
(`{PRODUCT}.go`, shared packages) — NO service subdirectories.

#### G.1.1 Suite & Product Pattern

**Canonical templates**: `api/cryptosuite-registry/templates/internal/apps/{__SUITE__,__PRODUCT__}/MANIFEST.yaml`
enforced by lint-fitness `apps-suite-template`, `apps-product-template`.

**Suite rigid structure** (`internal/apps/cryptoutil/` — exactly 1 suite):

| File/Dir | Status | Purpose |
|----------|--------|---------|
| `cryptoutil.go` | **REQUIRED** | Suite CLI dispatch via `RouteSuite()` |
| `cryptoutil_test.go` | **REQUIRED** | Suite router tests |
| `e2e/` | OPTIONAL | Full-suite E2E tests |

**Product rigid structure** (`internal/apps/{PRODUCT}/` — 5 products):

| File/Dir | Status | Purpose |
|----------|--------|---------|
| `{PRODUCT}.go` | **REQUIRED** | Product CLI dispatch via `RouteProduct()` |
| `{PRODUCT}_test.go` | **REQUIRED** | Product router tests |
| `{SERVICE}/` (any) | **FORBIDDEN** | Service code belongs at `internal/apps/{PS-ID}/`, NOT nested under product |
| shared packages | OPTIONAL | Varies by product; `identity/` has `apperr/`, `config/`, `domain/`, etc. |

**Known product violations** (service-named subdirs — GAP tasks in V17):

| Product | Forbidden dirs | Correct location |
|---------|---------------|-----------------|
| `sm/` | `im/`, `kms/` | `internal/apps/sm-im/`, `internal/apps/sm-kms/` |
| `jose/` | `ja/` | `internal/apps/jose-ja/` |
| `pki/` | `ca/` | `internal/apps/pki-ca/` |
| `skeleton/` | `template/` | `internal/apps/skeleton-template/` |

```
internal/apps/                                        # drwxr-x---
│
│   # Suite (×1)
├── cryptoutil/
│   ├── cryptoutil.go                                 #   REQUIRED: Suite CLI dispatch
│   └── cryptoutil_test.go                            #   REQUIRED: Suite tests
│
│   # Products (×5)
├── {PRODUCT}/                                        # identity, jose, pki, skeleton, sm
│   ├── {PRODUCT}.go                                  #   REQUIRED: Product CLI dispatch
│   ├── {PRODUCT}_test.go                             #   REQUIRED: Product tests
│   └── (shared packages only)/                       #   OPTIONAL: NO service subdirectories
```

#### G.1.2 Service Pattern (`{PS-ID}/`)

Each service lives at `internal/apps/{PS-ID}/` (flat, NOT nested under product).

**Canonical template**: `api/cryptosuite-registry/templates/internal/apps/__PS_ID__/MANIFEST.yaml`
enforced by lint-fitness `apps-ps-id-template`.

**ROOT FILE RULE**: ALL files at the PS-ID root MUST start with the `{SERVICE}_` prefix.
The root contains ONLY CLI integration code — no server logic, no HTTP handlers, no OpenAPI.

**PS-ID root rigid structure** (CLI files only — all 10 PS-IDs):

| File/Dir | Status | Purpose |
|----------|--------|---------|
| `{SERVICE}.go` | **REQUIRED** | Service entry point (`Kms()`, `Ja()`, `Ca()`, etc.) |
| `{SERVICE}_usage.go` | **REQUIRED** | CLI usage string via `BuildUsageMain()` |
| `{SERVICE}_test.go` | **REQUIRED** | CLI integration tests (help, version, unknown-subcommand) |
| `server/` | **REQUIRED** | All server implementation, swagger, and integration tests |
| `e2e/` | OPTIONAL | Docker Compose E2E tests |
| `client/` | OPTIONAL | Typed HTTP client (sm-kms, sm-im only) |
| `testing/` | OPTIONAL | Test helpers shared across packages |

**`server/` rigid structure** (all server code lives here, NOT at PS-ID root):

| File | Status | Purpose |
|------|--------|---------|
| `server.go` | **REQUIRED** | Admin server implementation |
| `swagger.go` | **REQUIRED** | OpenAPI/swagger serving via `builder.WithSwagger()` |
| `swagger_test.go` | **REQUIRED** | Swagger serving tests |
| `testmain_test.go` | **REQUIRED** | `TestMain` for integration test heavyweight setup |
| `{SERVICE}_lifecycle_test.go` | **REQUIRED** | Start/stop/graceful-shutdown across dual ports |
| `{SERVICE}_port_conflict_test.go` | **REQUIRED** | Deterministic failure when ports already in use |
| `public_server.go` | OPTIONAL | Public API server (absent in sm-kms legacy structure) |

**Current gap matrix** (✓ = correct location · MOVE = exists at PS-ID root, must migrate to `server/` · MISS = does not exist anywhere):

| Invariant | sm-kms | sm-im | jose-ja | pki-ca | id-authz | id-idp | id-rs | id-rp | id-spa | skel-tmpl |
|-----------|:------:|:-----:|:-------:|:------:|:--------:|:------:|:-----:|:-----:|:------:|:---------:|
| root `{SVC}.go` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| root `{SVC}_usage.go` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| root `{SVC}_test.go` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `server/server.go` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `server/swagger.go` | MOVE | MOVE | MOVE | MOVE | MOVE | MOVE | MOVE | **MISS** | **MISS** | MOVE |
| `server/swagger_test.go` | MOVE | MOVE | MOVE | MOVE | MOVE | MOVE | MOVE | **MISS** | **MISS** | MOVE |
| `server/testmain_test.go` | **MISS** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `server/{SVC}_lifecycle_test.go` | MOVE | MOVE | MOVE | MOVE¹ | MOVE² | MOVE² | **MISS** | **MISS** | **MISS** | MOVE |
| `server/{SVC}_port_conflict_test.go` | MOVE | MOVE | MOVE | MOVE | **MISS** | **MISS** | **MISS** | **MISS** | **MISS** | MOVE |

¹ pki-ca `server/` has `server_lifecycle_test.go` — rename to `ca_lifecycle_test.go` on move.
² identity-authz/idp have `service_lifecycle_test.go` at root — rename to `authz_`/`idp_lifecycle_test.go` on move.

**{SVC} = service component** (`kms`, `im`, `ja`, `ca`, `authz`, `idp`, `rs`, `rp`, `spa`, `template`)

```
├── {PS-ID}/                                          # Flat PS-ID directory (×10 total)
│   ├── {SERVICE}.go                                  #   REQUIRED: Service entry point (CLI only)
│   ├── {SERVICE}_usage.go                            #   REQUIRED: CLI usage string
│   ├── {SERVICE}_test.go                             #   REQUIRED: CLI integration tests
│   ├── server/                                       #   REQUIRED: All server code + tests
│   │   ├── server.go                                 #     Admin server
│   │   ├── public_server.go                          #     Public server (OPTIONAL: absent in sm-kms)
│   │   ├── swagger.go                                #     OpenAPI serving (NOT at PS-ID root)
│   │   ├── swagger_test.go                           #     Swagger tests
│   │   ├── testmain_test.go                          #     Integration TestMain
│   │   ├── {SERVICE}_lifecycle_test.go               #     Lifecycle tests
│   │   ├── {SERVICE}_port_conflict_test.go           #     Port conflict tests
│   │   └── (handler/route/service/contract files)   #     Per-service implementation
│   ├── e2e/                                          #   OPTIONAL: Docker Compose E2E tests
│   ├── client/                                       #   OPTIONAL: Typed HTTP client
│   └── (domain packages)/                            #   OPTIONAL: Varies by service complexity
```

**Concrete service subdirectories** (discovered from actual codebase):

| PS-ID | Subdirectories |
|-------|---------------|
| `identity-authz` | `client/`, `clientauth/`, `dpop/`, `e2e/`, `pkce/`, `server/`, `unified/` |
| `identity-idp` | `auth/`, `client/`, `e2e/`, `server/`, `templates/`, `unified/`, `userauth/` |
| `identity-rp` | `client/`, `e2e/`, `server/`, `unified/` |
| `identity-rs` | `client/`, `e2e/`, `server/`, `unified/` |
| `identity-spa` | `client/`, `e2e/`, `server/`, `unified/` |
| `jose-ja` | `client/`, `e2e/`, `model/`, `repository/`, `server/`, `service/` |
| `pki-ca` | `api/`, `bootstrap/`, `cli/`, `compliance/`, `config/`, `crypto/`, `domain/`, `domain-v2/`, `intermediate/`, `observability/`, `profile/`, `repository-v2/`, `security/`, `server/`, `service/`, `storage/` |
| `skeleton-template` | `client/`, `domain/`, `e2e/`, `repository/`, `server/` |
| `sm-im` | `client/`, `e2e/`, `integration/`, `model/`, `repository/`, `server/`, `testing/` |
| `sm-kms` | `client/`, `e2e/`, `server/` |

**Identity shared packages** (at `internal/apps/identity/`, shared across identity services):

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

#### G.1.3 Framework & Tools

```
internal/
├── apps-framework/                                   # Service framework (shared by ALL services)
│   ├── product/                                      #   Product-level framework
│   │   └── cli/
│   │       ├── product_router.go                     #     RouteProduct(), ProductConfig, ServiceEntry
│   │       └── product_router_test.go
│   ├── suite/                                        #   Suite-level framework
│   │   └── cli/
│   │       ├── suite_router.go                       #     RouteSuite(), SuiteConfig, ProductEntry
│   │       └── suite_router_test.go
│   ├── tls/                                          #   TLS certificate generation (merged: tls_generator + pkiinit)
│   │                                                 #   Generates /certs volume with 14-category keystore/truststore
│   │                                                 #   layout — see docs/tls-structure.md for full specification
│   └── service/                                      #   Service-level framework
│       ├── cli/
│       ├── client/
│       ├── config/                                            #   Shared config types (ServerConfig, DatabaseConfig, etc.)
│       ├── ratelimit/                                         #   Rate limiter (moved from identity/ratelimit)
│       ├── server/
│       │   ├── apis/
│       │   ├── application/
│       │   ├── barrier/
│       │   │   └── unsealkeysservice/
│       │   ├── builder/
│       │   ├── businesslogic/
│       │   ├── domain/
│       │   ├── listener/
│       │   ├── middleware/
│       │   ├── realm/
│       │   ├── realms/
│       │   ├── repository/
│       │   │   ├── migrations/                       #     Framework migrations (1001-1999)
│       │   │   └── test_migrations/
│       │   ├── service/
│       │   └── tenant/
│       ├── server_integration/
│       ├── testing/
│       │   ├── assertions/
│       │   ├── contract/
│       │   ├── e2e_helpers/
│       │   ├── e2e_infra/
│       │   ├── fixtures/
│       │   ├── healthclient/
│       │   ├── httpservertests/
│       │   ├── testdb/
│       │   └── testserver/
│       └── testutil/
│
├── apps-tools/                                       # Infrastructure tooling
│   ├── cicd_lint/                                    #   Custom linting and formatting tools
│   │   ├── cicd.go                                   #     CLI entry point + command registration
│   │   ├── cicd_test.go
│   │   ├── adaptive-sim/                             #     Adaptive simulation tools
│   │   ├── common/                                   #     Shared linter utilities
│   │   ├── docs_validation/                          #     Documentation validation (propagation checks)
│   │   ├── format_go/                                #     Go file formatting (any, copyloopvar)
│   │   ├── format_gotest/                            #     Go test file formatting (t.Helper)
│   │   ├── github_cleanup/                           #     GitHub Actions storage cleanup
│   │   ├── lint_compose/                             #     Docker Compose file linting
│   │   ├── lint_deployments/                         #     Deployment structure validator (8 validators)
│   │   ├── lint_docs/                                #     Documentation linter (includes docs_validation)
│   │   ├── lint_fitness/                             #     Architecture fitness functions (68 linters)
│   │   │   ├── lint_fitness.go                       #       Fitness runner
│   │   │   ├── lint_fitness_test.go
│   │   │   ├── lint-fitness-registry.yaml             #       Machine-readable linter category registry
│   │   │   ├── registry/                             #       Entity registry (SSOT)
│   │   │   │   ├── registry.go
│   │   │   │   └── registry_test.go
│   │   │   └── (68 linter directories)               #       See Section M for full list
│   │   ├── lint_go/                                  #     Go package linting
│   │   ├── lint_golangci/                            #     golangci-lint config validation
│   │   ├── lint_gotest/                              #     Go test file linting
│   │   ├── lint_go_mod/                              #     Go module linting
│   │   ├── lint_openapi/                             #     OpenAPI spec validation
│   │   ├── lint_ports/                               #     Port assignment validation
│   │   ├── lint_security/                            #     Security-focused linting
│   │   ├── lint_text/                                #     UTF-8 text file linting
│   │   └── lint_workflow/                            #     GitHub Actions workflow linting
│   │
│   └── cicd_workflow/                                #   GitHub Actions workflow management
│       └── *.go                                      #     run + cleanup subcommands
```

### G.2 internal/shared/ — Shared Libraries `drwxr-x---`

```
internal/shared/                                      # drwxr-x---
├── apperr/                                           # Application error types
├── container/
├── crypto/
│   ├── asn1/
│   ├── certificate/
│   ├── digests/
│   ├── hash/
│   ├── jose/
│   ├── keygen/
│   ├── keygenpooltest/
│   ├── password/
│   ├── pbkdf2/
│   └── tls/
├── database/
├── magic/                                            # Named constants (SSOT, excluded from coverage)
│   │                                                 # 42 files (all magic_*.go pattern)
│   ├── magic_api.go
│   ├── magic_cicd.go
│   ├── magic_cicd_test.go
│   ├── magic_console.go
│   ├── magic_crypto.go
│   ├── magic_database.go
│   ├── magic_docker.go
│   ├── magic_framework.go
│   ├── magic_identity.go                             # Identity product constants
│   ├── magic_identity_adaptive.go                    # Identity adaptive auth
│   ├── magic_identity_config.go                      # Identity config
│   ├── magic_identity_http.go                        # Identity HTTP
│   ├── magic_identity_keys.go                        # Identity keys
│   ├── magic_identity_metrics.go                     # Identity metrics
│   ├── magic_identity_mfa.go                         # Identity MFA
│   ├── magic_identity_oauth.go                       # Identity OAuth
│   ├── magic_identity_oidc.go                        # Identity OIDC
│   ├── magic_identity_pbkdf2.go                      # Identity PBKDF2
│   ├── magic_identity_scopes.go                      # Identity scopes
│   ├── magic_identity_testing.go                     # Identity testing
│   ├── magic_identity_timeouts.go                    # Identity timeouts
│   ├── magic_identity_uris.go                        # Identity URIs
│   ├── magic_jose.go                                 # JOSE product constants
│   ├── magic_memory.go                               # Memory constants
│   ├── magic_misc.go
│   ├── magic_network.go
│   ├── magic_orchestration.go
│   ├── magic_percent.go
│   ├── magic_pki.go                                  # PKI product constants
│   ├── magic_pkiinit.go                              # PKI init constants
│   ├── magic_pkix.go                                 # PKIX constants
│   ├── magic_pki_ca.go                               # PKI-CA service constants
│   ├── magic_security.go
│   ├── magic_session.go
│   ├── magic_skeleton.go                             # Skeleton product constants
│   ├── magic_sm.go                                   # SM product constants
│   ├── magic_sm_im.go                                # SM-IM service constants
│   ├── magic_telemetry.go
│   ├── magic_testing.go
│   ├── magic_test_fixtures.go                        # Test fixture constants
│   ├── magic_unseal.go
│   └── magic_workflows.go
├── pool/
├── pwdgen/
├── telemetry/
├── testutil/
└── util/
    ├── cache/
    ├── combinations/
    ├── datetime/
    ├── files/
    ├── network/
    ├── poll/
    ├── random/
    ├── sysinfo/
    └── thread/
```

### G.3 INFRA_TOOL CLI Wiring Rule `drwxr-x---`

**RULE**: All INFRA_TOOL CLI wiring files are at the tool root (`internal/apps-tools/{TOOL}/`),
NOT in a nested `cmd/` subdirectory within the tool.

```
internal/apps-tools/cicd_lint/                       # drwxr-x---
├── cicd.go                                           #   CLI entry point + command dispatch (at tool root)
└── cicd_test.go
```

**Known violation**: `internal/apps-tools/cicd_lint/cmd/cicd.go` (and `cicd_test.go`) currently
exists as a nested thin-wrapper layer. These files must be merged into
`internal/apps-tools/cicd_lint/cicd.go` and the `cmd/` subdirectory deleted. See Section N.

---

## H. docs/ — Documentation `drwxr-x---`

```
docs/                                                 # drwxr-x---
├── ENG-HANDBOOK.md                                   # SSOT: Architecture reference (5080+ lines)
├── DEV-SETUP.md                                      # Developer setup guide
├── README.md                                         # Documentation index
├── required-propagations.yaml                        # @propagate coverage completeness manifest
├── target-structure.md                               # THIS FILE — canonical target structure
└── framework-v17/                                    # Framework-v17 implementation artifacts (in progress)
    ├── lessons.md                                    #   Lessons learned (filled during execution)
    ├── plan.md                                       #   Implementation plan (6 phases, 40+ tasks)
    ├── tasks.md                                      #   Task checklist
    └── quizme-v1.md                                  #   Open architectural questions (answer before Phase 5)
```

---

## I. test/ — External Test Suites `drwxr-x---`

```
test/                                                 # drwxr-x---
└── load/                                             # Gatling load tests (Java 21 + Maven)
    │                                                 # Needs refactoring: cover all 10 service-level,
    │                                                 # all 5 product-level, and 1 suite-level load tests
    ├── .gitignore
    ├── .mvn/                                         #   Maven wrapper
    ├── mvnw                                          #   Maven wrapper (Unix, chmod 750)
    ├── mvnw.cmd                                      #   Maven wrapper (Windows)
    ├── pom.xml
    ├── README.md
    ├── src/
    └── target/                                       #   Maven build output (gitignored)
```

---

## J. pkg/ — Public Library Code (Reserved) `drwxr-x---`

```
pkg/                                                  # Currently empty, reserved for future public APIs
```

---

## K. Other Directories

```
scripts/                                              # Currently empty (.gitkeep only)
                                                      # Part of Go project structure, keep empty
workflow-reports/                                     # Ephemeral test output, never Git tracked (gitignored)
test-output/                                          # Ephemeral test output, never Git tracked (gitignored)
```

---

## L. Secret File Naming Convention

All tiers (service, product, suite) use **identical `{purpose}.secret` filenames** —
no tier prefix on active secret filenames. The **value inside** each secret contains
the tier-specific prefix (e.g., `{PS-ID}-`, `{PRODUCT}-`, `{SUITE}-`).

`.secret.never` marker files exist ONLY at product and suite tiers as explicit
reminders that browser/service credentials are service-level concerns.

| Secret Purpose | Filename | Service Value Pattern | Product Value Pattern | Suite Value Pattern |
|---------------|----------|-----------------------|-----------------------|---------------------|
| Hash pepper v3 | `hash-pepper-v3.secret` | `{PS-ID}-hash-pepper-v3-{base64-random-32-bytes}` | `{PRODUCT}-hash-pepper-v3-{base64-random-32-bytes}` | `{SUITE}-hash-pepper-v3-{base64-random-32-bytes}` |
| Browser username | `browser-username.secret` | `{PS-ID}-browser-user` | `.never` only | `.never` only |
| Browser password | `browser-password.secret` | `{PS-ID}-browser-pass-{base64-random-32-bytes}` | `.never` only | `.never` only |
| Service username | `service-username.secret` | `{PS-ID}-service-user` | `.never` only | `.never` only |
| Service password | `service-password.secret` | `{PS-ID}-service-pass-{base64-random-32-bytes}` | `.never` only | `.never` only |
| PostgreSQL username | `postgres-username.secret` | `{PS_ID}_database_user` | `{PRODUCT}_database_user` | `{SUITE}_database_user` |
| PostgreSQL password | `postgres-password.secret` | `{PS_ID}_database_pass-{base64-random-32-bytes}` | `{PRODUCT}_database_pass-{base64-random-32-bytes}` | `{SUITE}_database_pass-{base64-random-32-bytes}` |
| PostgreSQL database | `postgres-database.secret` | `{PS_ID}_database` | `{PRODUCT}_database` | `{SUITE}_database` |
| PostgreSQL URL | `postgres-url.secret` | `postgres://{PS_ID}_database_user:{PS_ID}_database_pass@{PS-ID}-postgres:5432/{PS_ID}_database?sslmode=disable` | `postgres://{PRODUCT}_database_user:{PRODUCT}_database_pass@{PRODUCT}-postgres:5432/{PRODUCT}_database?sslmode=disable` | `postgres://{SUITE}_database_user:{SUITE}_database_pass@{SUITE}-postgres:5432/{SUITE}_database?sslmode=disable` |
| Unseal shard N | `unseal-{N}of5.secret` | `{PS-ID}-unseal-key-N-of-5-{base64-random-32-bytes}` | `{PRODUCT}-unseal-key-N-of-5-{base64-random-32-bytes}` | `{SUITE}-unseal-key-N-of-5-{base64-random-32-bytes}` |

**`.secret.never` marker files** — present at product and suite tiers as explicit reminders:

| Tier | Files Present | Content |
|------|-------------|---------|
| Product (×5) | `browser-password.secret.never`, `browser-username.secret.never`, `service-password.secret.never`, `service-username.secret.never` | "MUST NEVER be used at product level. Use service-specific secrets." |
| Suite (×1) | Same 4 filenames | "MUST NEVER be used at suite level. Use service-specific secrets." |

**Total `.secret.never` files**: 4 per product × 5 products + 4 suite = **24 files**.

---

## M. Fitness Linter Coverage (68 existing + 12 planned in V17 = 80 target)

**Current**: 68 linters. **V17 target**: 80 linters (68 + 12 new in Phases 2–4; see `docs/framework-v17/`).

**All fitness linter directories** (alphabetical; `†` = new in framework-v17, not yet implemented):

```
lint_fitness/
├── lint-fitness-registry.yaml             # Machine-readable linter category registry
├── admin_bind_address/                    # Admin 127.0.0.1:9090 bind enforcement
├── api_path_registry/                     # API path registry validation (v7 NEW)
├── apps_product_no_service_dirs/          # † No service-named subdirs in product dirs (v17 Phase 2)
├── apps_product_template/                 # † MANIFEST.yaml-driven product structure (v17 Phase 4)
├── apps_ps_id_required_files/             # † Registry-driven PS-ID entry+usage file checks (v17 Phase 2)
├── apps_ps_id_server_package/             # † server/server.go + server/public_server.go (v17 Phase 2)
├── apps_ps_id_swagger_presence/           # † server/swagger.go + server/swagger_test.go (v17 Phase 2)
├── apps_ps_id_template/                   # † MANIFEST.yaml-driven PS-ID structure check (v17 Phase 4)
├── apps_ps_id_test_patterns/              # † server/testmain + lifecycle + port_conflict (v17 Phase 2)
├── apps_suite_required_files/             # † cryptoutil.go + cryptoutil_test.go (v17 Phase 2)
├── apps_suite_template/                   # † MANIFEST.yaml-driven suite structure (v17 Phase 4)
├── cmd_product_template/                  # † cmd/{PRODUCT}/main.go structural invariants (v17 Phase 4)
├── cmd_ps_id_template/                    # † cmd/{PS-ID}/main.go structural invariants (v17 Phase 4)
├── cmd_suite_template/                    # † cmd/{SUITE}/main.go structural invariants (v17 Phase 4)
├── archive_detector/                      # No archived/orphaned directories
├── banned_product_names/                  # Legacy product name detection
├── bind_address_safety/                   # Bind address safety (no 0.0.0.0 in tests)
├── cgo_free_sqlite/                       # CGO-free SQLite driver enforcement
├── check_skeleton_placeholders/           # Skeleton template placeholder validation
├── cicd_coverage/                         # CICD coverage enforcement
├── circular_deps/                         # Circular dependency detection
├── cmd_anti_pattern/                      # cmd/ anti-pattern detection
├── cmd_entry_whitelist/                   # Only 18 allowed cmd/ entries
├── cmd_main_pattern/                      # cmd/*/main.go pattern validation
├── compose_db_naming/                     # Docker Compose DB naming conventions
├── compose_header_format/                 # Docker Compose header format
├── compose_port_formula/                  # Compose port formula validation (v7 NEW)
├── compose_service_names/                 # Docker Compose service name conventions
├── configs_deployments_consistency/       # configs/ ↔ deployments/ structural mirror
├── configs_empty_dir/                     # No empty config directories
├── configs_naming/                        # Flat configs/{PS-ID}/ naming pattern
├── config_overlay_freshness/              # Config overlay template freshness (v7 NEW)
├── cross_service_import_isolation/        # Service import isolation enforcement
├── (removed: crypto_rand/ — now enforced by golangci-lint fips-rand deny rule)
├── deployment_dir_completeness/           # Deployment directory completeness
├── dockerfile_labels/                     # Dockerfile OCI label validation
├── domain_layer_isolation/                # Domain layer isolation enforcement
├── entity_registry_completeness/          # Entity registry vs filesystem sync
├── entity_registry_schema/                # Entity registry YAML schema validation (v7 NEW)
├── file_size_limits/                      # File size limit enforcement (500 lines)
├── fitness_registry_completeness/         # Fitness linter registry completeness (v7 NEW)
├── gen_config_initialisms/                # oapi-codegen initialism consistency
├── health_endpoint_presence/              # Health endpoint presence in services
├── health_path_completeness/              # Health path completeness matrix (v7 NEW)
├── import_alias_formula/                  # Import alias formula enforcement (v7 NEW)
├── infra_tool_naming/                     # Infrastructure tool naming conventions
├── insecure_skip_verify/                  # InsecureSkipVerify detection
├── legacy_dir_detection/                  # Legacy directory detection
├── magic_constant_location/               # Magic constants in internal/shared/magic/
├── magic_e2e_compose_path/                # E2E compose path magic constants
├── magic_e2e_container_names/             # E2E container name magic constants
├── migration_comment_headers/             # Migration file comment headers
├── migration_numbering/                   # Migration file numbering
├── migration_range_compliance/            # Framework (1001-1999) vs domain (2001+)
├── non_fips_algorithms/                   # FIPS 140-3 algorithm enforcement
├── no_hardcoded_passwords/                # No hardcoded password detection
├── no_local_closed_db_helper/             # No local closed DB helpers
├── no_postgres_in_non_e2e/                # PostgreSQL only in E2E tests
├── no_unit_test_real_db/                  # No real DB in unit tests
├── no_unit_test_real_server/              # No real server in unit tests
├── otlp_service_name_pattern/             # OTLP service name pattern enforcement
├── parallel_tests/                        # t.Parallel() enforcement
├── pki_ca_profile_schema/                 # PKI-CA certificate profile schema validation (v7 NEW)
├── product_structure/                     # Product directory structure validation
├── product_wiring/                        # Product wiring validation
├── registry/                              # Entity registry (SSOT)
├── require_api_dir/                       # api/ directory requirement per service
├── require_framework_naming/              # Framework naming convention enforcement
├── root_junk_detection/                   # Root directory junk file detection
├── secret_content/                        # Secret file content validation
├── secret_naming/                         # Secret file naming conventions
├── service_contract_compliance/           # ServiceServer compile-time assertion enforcement
├── service_structure/                     # Service directory structure validation
├── standalone_config_otlp_names/          # Standalone config OTLP name consistency
├── standalone_config_presence/            # Standalone config file presence
├── subcommand_completeness/               # CLI subcommand completeness matrix (v7 NEW)
├── template_consistency/                  # Skeleton template consistency
├── test_file_suffix_structure/            # Test file suffix structural rules (v7 NEW)
├── test_patterns/                         # Test pattern enforcement
├── tls_minimum_version/                   # TLS 1.3+ minimum version enforcement
└── unseal_secret_content/                 # Unseal key value pattern validation
```

**Selection of linters by scope**:

| Linter | Scope | Rule |
|--------|-------|------|
| `root_junk_detection` | `{ROOT}/` | No `*.exe`, `*.py`, `coverage*`, `*.test.exe` at root |
| `cmd_entry_whitelist` | `cmd/` | Only 18 allowed entries (1 suite + 5 products + 10 services + 2 infra tools) |
| `configs_naming` | `configs/` | Validates flat `{PS-ID}/{PS-ID}.yml` pattern; rejects nested `{PRODUCT}/{SERVICE}/`; allows `pki-ca/profiles/` and `identity-authz/domain/policies/` exceptions |
| `secret_naming` | `deployments/*/secrets/` | All tiers use `{purpose}.secret` names; `.never` markers enforced at product/suite |
| `unseal_secret_content` | `deployments/*/secrets/unseal-*.secret` | Validates unseal secret value patterns; rejects generic `dev-unseal-key-N-of-5` placeholders |
| `template_consistency` | `deployments/skeleton-template/` | Hyphens in secret names (not underscores) |
| `entity_registry_completeness` | (cross-cutting) | Verify `configs/{PS-ID}/` existence for all registered PS-IDs |
| `dockerfile_labels` | `deployments/*/Dockerfile` | Validates LABEL `org.opencontainers.image.title` matches deployment tier |

---

## N. Remaining Work (Pending Items)

### N.1 PS-ID Root-to-server/ File Migration (V17 Phase 5)

All files currently at PS-ID root that do not start with `{SERVICE}_` must move to `server/`.
See framework-v17/ plan.md Phase 5 and the gap matrix in G.1.2 for per-PS-ID details.

| Area | Current State | Target State | Action |
|------|--------------|-------------|--------|
| `swagger.go`, `swagger_test.go` at root | 8 PS-IDs have these at root | All in `server/` | MOVE |
| `testmain_test.go` at root | 9 PS-IDs have at root (sm-kms missing) | All in `server/` | MOVE (+ CREATE for sm-kms) |
| `{SVC}_lifecycle_test.go` at root | 7 PS-IDs have at root | All in `server/` | MOVE (+ CREATE for id-rs, id-rp, id-spa) |
| `{SVC}_port_conflict_test.go` at root | 5 PS-IDs have at root | All in `server/` | MOVE (+ CREATE for id-authz, id-idp, id-rs, id-rp, id-spa) |
| Non-`{SERVICE}_`-prefixed files at root | sm-im, id-authz, id-idp, id-rs have extra root files | None | MOVE to `server/` or appropriate subpackage |
| `swagger.go`/`swagger_test.go` creation | identity-rp, identity-spa missing entirely | Present in `server/` | CREATE in `server/` |

### N.2 Product Service-Dir Cleanup (V17 Phase 5)

Service-named subdirectories inside product directories violate the flat PS-ID layout rule.

| Product | Forbidden dirs | Correct location | Action |
|---------|---------------|-----------------|--------|
| `sm/` | `im/`, `kms/` | `internal/apps/sm-im/`, `internal/apps/sm-kms/` | Audit + DELETE if redundant |
| `jose/` | `ja/` | `internal/apps/jose-ja/` | Audit + DELETE if redundant |
| `pki/` | `ca/` | `internal/apps/pki-ca/` | Audit + DELETE if redundant |
| `skeleton/` | `template/` | `internal/apps/skeleton-template/` | Audit + DELETE if redundant |

### N.3 INFRA_TOOL Nested cmd/ Removal

| Area | Current State | Target State | Action |
|------|--------------|-------------|--------|
| `internal/apps-tools/cicd_lint/cmd/` | Thin-wrapper `cmd/cicd.go` nested inside tool | Logic merged into `cicd_lint/cicd.go`, `cmd/` deleted | MERGE + DELETE |

### N.4 Deployments

| Area | Current State | Target State | Action |
|------|--------------|-------------|--------|
| `deployments/` product Dockerfile | Absent in all 5 products | Intentionally absent in all 5 products; PRODUCT domains reuse PS-ID images | KEEP ABSENT |

### N.5 Framework V17 Linter Implementation

| Phase | Linters | Status |
|-------|---------|--------|
| V17 Phase 2 | 6 new linters (`apps-ps-id-*`, `apps-product-no-service-dirs`, `apps-suite-required-files`) | ❌ TODO |
| V17 Phase 3 | Register + integrate Phase 2 linters | ❌ TODO |
| V17 Phase 4 | 6 template-compliance linters (`apps-*-template`, `cmd-*-template`) | ❌ TODO |
