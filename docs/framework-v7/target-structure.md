# Target Repository Structure

**Status**: CANONICAL TARGET — Living reference document
**Created**: 2026-03-26
**Last Updated**: 2026-03-29
**Purpose**: Define the complete, parameterized target state of every directory and file in the
repository. Originally created during framework-v6, now maintained as a living spec in framework-v7.
This document supersedes framework-v5/target-structure.md (deleted — git history preserves).

**RULE**: Everything listed here MUST exist. Everything NOT listed is deleted.

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
├── .cicd/                                 # CICD runtime caches (gitignored)
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

## B. .github/ — GitHub & Copilot Configuration `drwxr-x---`

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

### B.4 Skills (14 skills + README)

```
.github/skills/
├── README.md
├── agent-scaffold/SKILL.md
├── contract-test-gen/SKILL.md
├── coverage-analysis/SKILL.md
├── fips-audit/SKILL.md
├── fitness-function-gen/SKILL.md
├── instruction-scaffold/SKILL.md
├── migration-create/SKILL.md
├── new-service/SKILL.md
├── openapi-codegen/SKILL.md
├── propagation-check/SKILL.md
├── skill-scaffold/SKILL.md
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

---

## C. cmd/ — Binary Entry Points `drwxr-x---`

**Pattern**: Flat directories — every entry is a direct child of `cmd/`. No nesting.
Each entry has exactly one `main.go` that delegates to `internal/apps/`.

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
generated code.

```
api/                                                  # drwxr-x---
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
Dockerfiles do NOT yet exist (CREATE pending — see Section N).

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

**Total per product**: 4 `.secret.never` + 10 `.secret` = 14 files (no Dockerfile yet).

**All 5 products** (`identity`, `jose`, `pki`, `skeleton`, `sm`) follow this identical structure.

### F.3 Suite Deployment

**Pattern**: `deployments/{SUITE}/`

The suite deployment directory uses the bare `{SUITE}` name (e.g., `cryptoutil`),
consistent with all other naming conventions. Contains a Dockerfile, compose.yml,
and secrets.

```
deployments/{SUITE}/                                  # drwxr-x---
├── compose.yml                                       # Suite-level Docker Compose
├── Dockerfile                                        # Suite Docker image build
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

**Total**: 4 `.secret.never` + 10 `.secret` = 14 files + Dockerfile + compose.yml.

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

### F.5 Dockerfile Parameterization

All Dockerfiles follow identical multi-stage structure (validation → builder → runtime).
Parameterized fields differ by deployment tier.

**Pattern**: `deployments/{DEPLOYMENT-DIR}/Dockerfile`

| Field | Service (`{PS-ID}`) | Product (`{PRODUCT}`) | Suite (`{SUITE}`) |
|-------|---------------------|----------------------|-------------------|
| `image.title` LABEL | `{SUITE}-{PS-ID}` | `{SUITE}-{PRODUCT}` | `{SUITE}` |
| `image.description` LABEL | Service-specific description | Product-specific description | Suite-level description |
| Binary built | `./cmd/{SUITE}` (always suite binary) | `./cmd/{SUITE}` | `./cmd/{SUITE}` |
| `EXPOSE` | 8080 (container public) | Service-range (e.g., 18000) | Suite-range (e.g., 28000) |
| `HEALTHCHECK` | `wget --no-check-certificate -qO- https://127.0.0.1:8080/browser/api/v1/health` | Same pattern, product port | Same pattern, suite port |
| `ENTRYPOINT` | `["/app/{SUITE}", "{PS-ID}", "start"]` | `["/app/{SUITE}", "{PRODUCT}", "start"]` | `["/app/{SUITE}"]` |

**Current state**: 10 service-level + 1 suite-level Dockerfiles exist. 0 product-level Dockerfiles exist (CREATE pending).

---

## G. internal/ — Private Application Code `drwxr-x---`

### G.1 internal/apps/ — Application Layer

**Structure**: `internal/apps/{SUITE | PRODUCT | PS-ID | framework | tools}`

Services live at flat `internal/apps/{PS-ID}/` (NOT nested under their product).
`cmd/{PS-ID}/main.go` delegates to `internal/apps/{PS-ID}/{PS-ID}.go`.
Product directories (`internal/apps/{PRODUCT}/`) contain ONLY product-level code
(`{PRODUCT}.go`, shared packages) — NO service subdirectories.

#### G.1.1 Suite & Product Pattern

```
internal/apps/                                        # drwxr-x---
│
│   # Suite orchestration (×1, {SUITE}=cryptoutil)
├── cryptoutil/
│   ├── cryptoutil.go                                 #   Suite CLI dispatch (seam pattern)
│   ├── *_test.go
│   └── e2e/                                          #   E2E tests (full suite docker compose)
│
│   # Product level (×5)
├── {PRODUCT}/                                        # identity, jose, pki, skeleton, sm
│   ├── {PRODUCT}.go                                  #   Product CLI dispatch
│   ├── *_test.go
│   ├── e2e/                                          #   E2E tests (full product docker compose)
│   └── (shared packages)/                            #   Shared within product (optional, varies)
```

#### G.1.2 Service Pattern (`{PS-ID}/`)

Each service lives at `internal/apps/{PS-ID}/` (flat, NOT nested under product). The generic pattern:

```
├── {PS-ID}/                                          # Flat PS-ID directory (×10 total)
│   ├── {PS-ID}.go                                    #   Service entry point (seam pattern)
│   ├── *_test.go
│   ├── client/                                       #   HTTP client (optional)
│   ├── e2e/                                          #   E2E tests (service docker compose)
│   ├── integration/                                  #   Integration tests (optional)
│   ├── model/                                        #   Domain models (optional)
│   ├── repository/                                   #   Data access layer (optional)
│   │   ├── *.go                                      #     GORM entity models + repository methods
│   │   ├── *_test.go
│   │   └── migrations/                               #     Domain migrations (2001+)
│   ├── server/                                       #   HTTP server setup
│   ├── service/                                      #   Business logic (optional)
│   └── testing/                                      #   Test helpers (optional)
```

**Concrete service subdirectories** (discovered from actual codebase):

| PS-ID | Subdirectories |
|-------|---------------|
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

**Identity shared packages** (at `internal/apps/identity/`, shared across identity services):

| Package | Purpose |
|---------|---------|
| `apperr/` | Identity-specific error types |
| `authz/` | Authorization logic shared across identity services |
| `config/` | Shared identity configuration |
| `domain/` | Shared identity domain types |
| `email/` | Email sending |
| `idp/` | Identity provider shared logic |
| `issuer/` | Token issuer |
| `jobs/` | Background jobs |
| `mfa/` | Multi-factor authentication |
| `ratelimit/` | Rate limiting |
| `repository/` (with `orm/`, `migrations/`) | Shared identity data access |
| `rotation/` | Key/token rotation |
| `rp/` | Relying party shared logic |
| `rs/` | Resource server shared logic |
| `spa/` | Single page app shared logic |

#### G.1.3 Framework & Tools

```
internal/apps/
├── framework/                                        # Service framework (shared by ALL services)
│   ├── product/                                      #   Product-level framework
│   │   └── cli/
│   │       ├── product_router.go                     #     RouteProduct(), ProductConfig, ServiceEntry
│   │       └── product_router_test.go
│   ├── suite/                                        #   Suite-level framework
│   │   └── cli/
│   │       ├── suite_router.go                       #     RouteSuite(), SuiteConfig, ProductEntry
│   │       └── suite_router_test.go
│   ├── tls/                                          #   TLS certificate generation (merged: tls_generator + pkiinit)
│   └── service/                                      #   Service-level framework
│       ├── cli/
│       ├── client/
│       ├── config/
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
├── tools/                                            # Infrastructure tooling
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
│   │   ├── lint_fitness/                             #     Architecture fitness functions (59 linters)
│   │   │   ├── lint_fitness.go                       #       Fitness runner
│   │   │   ├── lint_fitness_test.go
│   │   │   ├── registry/                             #       Entity registry (SSOT)
│   │   │   │   ├── registry.go
│   │   │   │   └── registry_test.go
│   │   │   └── (59 linter directories)               #       See Section M for full list
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
├── apperr/                                           # Application error types (MOVE to framework/apperr/ pending)
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

### G.3 internal/cmd/ — CLI Wiring `drwxr-x---`

```
internal/cmd/                                         # drwxr-x---
└── cicd_lint/                                        # cicd-lint CLI wiring
    ├── cicd.go                                       #   Bridges cmd/cicd-lint/main.go → tools/cicd_lint/
    └── cicd_test.go
```

**Note**: `internal/cmd/cicd_lint/` is the thin wiring layer between `cmd/cicd-lint/main.go`
and `internal/apps/tools/cicd_lint/`. It contains the CLI entry point and argument parsing.

---

## H. docs/ — Documentation `drwxr-x---`

```
docs/                                                 # drwxr-x---
├── ARCHITECTURE.md                                   # SSOT: Architecture reference (5080+ lines)
├── DEV-SETUP.md                                      # Developer setup guide
├── README.md                                         # Documentation index
├── gremlins/                                         # ORPHANED empty dir (DELETE — content is at framework-v7/gremlins/)
├── workflow-runtimes/                                # ORPHANED empty dir (DELETE — content is at framework-v7/workflow-runtimes/)
└── framework-v7/                                     # Ongoing reference documentation
    ├── README.md                                     # Index of living docs
    ├── PORT-REORDERING.md                            # Port reassignment plan (completed)
    ├── STALE.md                                      # Stale content tracking (all items resolved)
    ├── target-structure.md                           # THIS FILE (canonical target structure)
    ├── gremlins/                                     # Mutation testing reference
    │   ├── MUTATIONS-HOWTO.md
    │   ├── MUTATIONS-TASKS.md
    │   ├── mutation-analysis.md
    │   └── mutation-baseline-results.md
    └── workflow-runtimes/                            # CI/CD operational reference
        ├── README.md
        └── GITHUB-STORAGE-CLEANUP.md
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

## M. Fitness Linter Coverage (59 linters)

**All 59 fitness linter directories** (alphabetical):

```
lint_fitness/
├── admin_bind_address/                    # Admin 127.0.0.1:9090 bind enforcement
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
├── compose_service_names/                 # Docker Compose service name conventions
├── configs_deployments_consistency/       # configs/ ↔ deployments/ structural mirror
├── configs_empty_dir/                     # No empty config directories
├── configs_naming/                        # Flat configs/{PS-ID}/ naming pattern
├── cross_service_import_isolation/        # Service import isolation enforcement
├── crypto_rand/                           # crypto/rand enforcement (never math/rand)
├── deployment_dir_completeness/           # Deployment directory completeness
├── dockerfile_labels/                     # Dockerfile OCI label validation
├── domain_layer_isolation/                # Domain layer isolation enforcement
├── entity_registry_completeness/          # Entity registry vs filesystem sync
├── file_size_limits/                      # File size limit enforcement (500 lines)
├── gen_config_initialisms/                # oapi-codegen initialism consistency
├── health_endpoint_presence/              # Health endpoint presence in services
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
├── product_structure/                     # Product directory structure validation
├── product_wiring/                        # Product wiring validation
├── registry/                              # Entity registry (SSOT)
├── require_api_dir/                       # api/ directory requirement per service
├── require_framework_naming/              # Framework naming convention enforcement
├── root_junk_detection/                   # Root directory junk file detection
├── secret_content/                        # Secret file content validation
├── secret_naming/                         # Secret file naming conventions
├── service_contract_compliance/           # Service contract test presence
├── service_structure/                     # Service directory structure validation
├── standalone_config_otlp_names/          # Standalone config OTLP name consistency
├── standalone_config_presence/            # Standalone config file presence
├── template_consistency/                  # Skeleton template consistency
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

| Area | Current State | Target State | Action |
|------|--------------|-------------|--------|
| Root files | 9 junk `*_coverage`/`cover` artifacts | Clean project config only | DELETE artifacts |
| `docs/gremlins/` | Empty orphaned directory at `docs/` root | Deleted (content is at `docs/framework-v7/gremlins/`) | DELETE |
| `docs/workflow-runtimes/` | Empty orphaned directory at `docs/` root | Deleted (content is at `docs/framework-v7/workflow-runtimes/`) | DELETE |
| `deployments/` product Dockerfile | Missing in all 5 products | Present in all 5 products | CREATE |
| `internal/shared/apperr/` | Present | Moved to `internal/apps/framework/apperr/` | MOVE |
| `testdata/` | Present (`adaptive-sim/` subdirectory) | Deleted (move to owning package) | DELETE |
| `sm-im` orphaned test DBs | ~130 `test_*.db` files committed | Gitignored and removed from tracking | ADD to `.gitignore`, `git rm --cached` |
