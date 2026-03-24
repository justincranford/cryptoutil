# Target Repository Structure - Framework v6

**Status**: CANONICAL TARGET — Post-v6 implementation state
**Created**: 2026-03-26
**Last Updated**: 2026-03-26
**Purpose**: Define the complete, parameterized target state of every directory and file in the
repository after all framework-v6 phases complete. This document supersedes framework-v5/target-structure.md.

**RULE**: Everything listed here MUST exist after v6 completes. Everything NOT listed is deleted.

---

## Corrections from framework-v5/target-structure.md

The following errors and contradictions in v5/target-structure.md are resolved in this document:

| # | v5 Error | v6 Correction | Decision |
|---|----------|--------------|---------|
| 1 | E.4 used nested `configs/{PRODUCT}/{SERVICE}/` dirs | E.4 now flat `configs/{PS-ID}/` per E.3 | 2=B |
| 2 | E.4 used `{SERVICE}.yml` config filename | Config files named `{PS-ID}.yml` | 2=B |
| 3 | sqlite-2 overlay missing from F.1 (only sqlite-1) | F.1 has BOTH sqlite-1 AND sqlite-2 | RC-3 |
| 4 | F.1 unseal example showed `im-{hex}` (SERVICE prefix) | `{PS-ID}-unseal-key-N-of-5-{hex}` | 1=A |
| 5 | F.2 had duplicate `unseal-5of5.secret` entry | Single entry, no duplicate | RC-1 |
| 6 | postgres-database value was `{PS_ID}` | `{PS_ID}_database` | 6=A |
| 7 | postgres-username value was `{PS_ID}_user` | `{PS_ID}_database_user` | 6=A |
| 8 | `configs/pki-ca/profiles/` not documented | Documented as valid exception | 3=B |
| 9 | `configs/identity-authz/domain/policies/` absent | Documented with rename | 4=A |
| 10 | `deployments/template/` still shown as existing | Deleted (merged → skeleton-template) | 5=C |
| 11 | `doc-sync.agent.md` listed in B | Not listed (agent deleted) | — |
| 12 | `custom-cicd-lint/action.yml` in B | Renamed to `download-cicd/action.yml` | — |
| 13 | `.vscode/mcp.json` missing from A.3 | Added to A.3 | — |
| 14 | `docs/UPDATE-TOOLS.md` missing from H | Added to H | — |
| 15 | Product unseal used `dev-unseal-key-N-of-5` | `{PRODUCT}-unseal-key-N-of-5-{hex}` | 1=A |
| 16 | Suite unseal used `suite-` prefix | `cryptoutil-unseal-key-N-of-5-{hex}` | 1=A |

---

## Entity Hierarchy (Canonical)

| Level | Variable | Instances | Count |
|-------|----------|-----------|-------|
| Suite | `{SUITE}` | `cryptoutil` | 1 |
| Product | `{PRODUCT}` | `identity`, `jose`, `pki`, `skeleton`, `sm` | 5 |
| Service | `{SERVICE}` | varies per product (see below) | 10 total |
| PS-ID | `{PS-ID}` = `{PRODUCT}-{SERVICE}` | see table below | 10 |
| PS_ID | `{PS_ID}` = `{PRODUCT}_{SERVICE}` | underscore variant for SQL/secrets | 10 |
| Infra Tool | N/A | `cicd-lint`, `workflow` | 2 |
| Framework | N/A | `framework` | 1 |

### Product-Service Matrix

| PS-ID | PS_ID | Product | Service | Display Name |
|-------|-------|---------|---------|-------------|
| `identity-authz` | `identity_authz` | identity | authz | Identity Authorization Server |
| `identity-idp` | `identity_idp` | identity | idp | Identity Provider |
| `identity-rp` | `identity_rp` | identity | rp | Identity Relying Party |
| `identity-rs` | `identity_rs` | identity | rs | Identity Resource Server |
| `identity-spa` | `identity_spa` | identity | spa | Identity Single Page App |
| `jose-ja` | `jose_ja` | jose | ja | JOSE JWK Authority |
| `pki-ca` | `pki_ca` | pki | ca | PKI Certificate Authority |
| `skeleton-template` | `skeleton_template` | skeleton | template | Skeleton Template |
| `sm-im` | `sm_im` | sm | im | Secrets Manager Instant Messenger |
| `sm-kms` | `sm_kms` | sm | kms | Secrets Manager Key Management |

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
├── go.mod                                 # Go module definition
├── go.sum                                 # Go module dependency checksums
├── LICENSE                                # Project license
├── pyproject.toml                         # Python project config (pre-commit tooling)
└── README.md                              # Project README
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

### B.5 Workflows (14 workflows)

```
.github/workflows/
├── ci-benchmark.yml                       # Benchmark testing
├── ci-coverage.yml                        # Code coverage analysis
├── ci-dast.yml                            # Dynamic application security testing
├── ci-e2e.yml                             # End-to-end testing
├── ci-fitness.yml                         # Architecture fitness functions
├── ci-fuzz.yml                            # Fuzz testing
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

### B.6 What Gets DELETED from .github/

| File | Reason |
|------|--------|
| `agents/doc-sync.agent.md` | Agent deleted — functionality not required |
| `actions/custom-cicd-lint/` | Renamed to `download-cicd/` |

---

## C. cmd/ — Binary Entry Points `drwxr-x---`

**Pattern**: Flat directories; each entry has exactly one `main.go` that delegates to `internal/apps/`.

```
cmd/                                                  # drwxr-x---
├── cryptoutil/main.go                                # Suite CLI → internal/apps/cryptoutil/
├── {PRODUCT}/main.go                                 # Product CLI → internal/apps/{PRODUCT}/ (×5)
│   ├── identity/main.go
│   ├── jose/main.go
│   ├── pki/main.go
│   ├── skeleton/main.go
│   └── sm/main.go
├── {PS-ID}/main.go                                   # Service CLI → internal/apps/{PRODUCT}/{SERVICE}/ (×10)
│   ├── identity-authz/main.go
│   ├── identity-idp/main.go
│   ├── identity-rp/main.go
│   ├── identity-rs/main.go
│   ├── identity-spa/main.go
│   ├── jose-ja/main.go
│   ├── pki-ca/main.go
│   ├── skeleton-template/main.go
│   ├── sm-im/main.go
│   └── sm-kms/main.go
└── {INFRA-TOOL}/main.go                             # Infra tools (×2)
    ├── cicd-lint/main.go
    └── workflow/main.go
```

**Total**: 18 entries (1 suite + 5 products + 10 services + 2 infra tools).

**DELETE from cmd/**:

| Entry | Reason |
|-------|--------|
| `cmd/demo/` | Dead code |
| `cmd/identity-compose/` | Non-standard entry point |
| `cmd/identity-demo/` | Dead code |

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

```
configs/
└── cryptoutil/
    └── cryptoutil.yml                     # Suite-level config (logging, telemetry)
```

### E.2 Product Configs (5 products — FLAT, one dir per product)

Each product has its own flat directory at `configs/{PRODUCT}/` containing exactly
one config file named `{PRODUCT}.yml`. NO nested service subdirectories.

```
configs/
├── identity/
│   └── identity.yml                       # Product-level config
├── jose/
│   └── jose.yml
├── pki/
│   └── pki.yml
├── skeleton/
│   └── skeleton.yml
└── sm/
    └── sm.yml
```

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
│           ├── adaptive-authorization.yml # RENAMED from adaptive-auth.yml (banned term)
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

### E.4 What Gets DELETED from configs/

Deletion order: service subdirs first, then empty product dirs.

| Current Location | Reason |
|-----------------|--------|
| `configs/sm/im/` | Service configs moved to flat `configs/sm-im/` |
| `configs/sm/kms/` | Service configs moved to flat `configs/sm-kms/` |
| `configs/jose/ja/` | Service configs moved to flat `configs/jose-ja/` |
| `configs/pki/ca/` | Service configs moved to flat `configs/pki-ca/` |
| `configs/skeleton/template/` | Service configs moved to flat `configs/skeleton-template/` |
| `configs/identity/authz/` | Service configs moved to flat `configs/identity-authz/` |
| `configs/identity/idp/` | Service configs moved to flat `configs/identity-idp/` |
| `configs/identity/rp/` | Service configs moved to flat `configs/identity-rp/` |
| `configs/identity/rs/` | Service configs moved to flat `configs/identity-rs/` |
| `configs/identity/spa/` | Service configs moved to flat `configs/identity-spa/` |
| `configs/identity/policies/` | Moved to `configs/identity-authz/domain/policies/` |
| `configs/skeleton/skeleton-server.yml` | Orphaned product-level file (non-canonical name) |
| `configs/sm/im/sm-im-pg-1.yml` | Deployment variant — belongs in deployments/ |
| `configs/sm/im/sm-im-pg-2.yml` | Deployment variant — belongs in deployments/ |
| `configs/sm/im/sm-im-sqlite.yml` | Deployment variant — belongs in deployments/ |
| `configs/sm/kms/sm-kms-pg-1.yml` | Deployment variant — belongs in deployments/ |
| `configs/sm/kms/sm-kms-pg-2.yml` | Deployment variant — belongs in deployments/ |
| `configs/sm/kms/sm-kms-sqlite.yml` | Deployment variant — belongs in deployments/ |
| `configs/pki-ca/pki-ca-config-schema.yaml` | Schema hardcoded in Go, not a config file |
| `configs/identity/development.yml` | Environment config — not in canonical config spec |
| `configs/identity/production.yml` | Environment config — not in canonical config spec |
| `configs/identity/test.yml` | Environment config — not in canonical config spec |
| `configs/identity/profiles/` | Compose profiles — not in spec |
| `configs/orphaned/` | Archived orphaned configs — delete after v6 review |

After all service subdirs are moved to flat structure, the parent product directories
`configs/sm/`, `configs/jose/`, `configs/pki/`, `configs/skeleton/` have no more nested
subdirs and contain only the product-level `{PRODUCT}.yml` file. The `configs/identity/`
directory also contains only `identity.yml` after policies and service subdirs are removed.

**No orphaned deployment-variant files** (`*-pg-1.yml`, `*-sqlite.yml`, etc.) remain
in configs/ — those belong in `deployments/{PS-ID}/config/`.

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
    ├── unseal-1of5.secret                            #   {PS-ID}-unseal-key-1-of-5-{hex-random-32-bytes}
    ├── unseal-2of5.secret                            #   {PS-ID}-unseal-key-2-of-5-{hex-random-32-bytes}
    ├── unseal-3of5.secret                            #   {PS-ID}-unseal-key-3-of-5-{hex-random-32-bytes}
    ├── unseal-4of5.secret                            #   {PS-ID}-unseal-key-4-of-5-{hex-random-32-bytes}
    └── unseal-5of5.secret                            #   {PS-ID}-unseal-key-5-of-5-{hex-random-32-bytes}
```

**Concrete examples**:

```
# sm-im secrets (PS-ID=sm-im, PS_ID=sm_im)
hash-pepper-v3.secret  →  sm-im-hash-pepper-v3-Qrst6789Uvwx0123Yzab4567Cdef8901
unseal-1of5.secret  →  sm-im-unseal-key-1-of-5-a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
postgres-username.secret  →  sm_im_database_user
postgres-password.secret  →  sm_im_database_pass-Abcd1234Efgh5678Ijkl9012Mnop3456
postgres-database.secret  →  sm_im_database
postgres-url.secret  →  postgres://sm_im_database_user:sm_im_database_pass-Abcd1234...@sm-im-postgres:5432/sm_im_database?sslmode=disable
browser-username.secret  →  sm-im-browser-user
browser-password.secret  →  sm-im-browser-pass-Ghij2345Klmn6789Opqr0123Stuv4567
service-username.secret  →  sm-im-service-user
service-password.secret  →  sm-im-service-pass-Wxyz8901Abcd2345Efgh6789Ijkl0123

# pki-ca secrets (PS-ID=pki-ca, PS_ID=pki_ca)
unseal-1of5.secret  →  pki-ca-unseal-key-1-of-5-{unique-hex-NOT-copied-from-sm-kms}
postgres-database.secret  →  pki_ca_database
postgres-username.secret  →  pki_ca_database_user
```

**All 10 services** (`identity-authz`, `identity-idp`, `identity-rp`, `identity-rs`,
`identity-spa`, `jose-ja`, `pki-ca`, `skeleton-template`, `sm-im`, `sm-kms`) follow
this identical structure.

### F.2 Per-Product Deployment (5 products)

Each product has a deployment directory with a Dockerfile, compose.yml, and secrets.
Product-level secrets are shared across all services in the product.

```
deployments/{PRODUCT}/                                # drwxr-x---
├── compose.yml                                       # Product-level Docker Compose
├── Dockerfile                                        # Product Docker image (v6 CREATE)
└── secrets/
    ├── hash-pepper-v3.secret                         # {PRODUCT}-hash-pepper-v3-{base64-random-32-bytes}
    ├── browser-username.secret.never                 # MUST NEVER be used at product level
    ├── browser-password.secret.never                 # MUST NEVER be used at product level
    ├── service-username.secret.never                 # MUST NEVER be used at product level
    ├── service-password.secret.never                 # MUST NEVER be used at product level
    ├── postgres-username.secret                      # {PRODUCT}_database_user
    ├── postgres-password.secret                      # {PRODUCT}_database_pass-{base64-random-32-bytes}
    ├── postgres-database.secret                      # {PRODUCT}_database
    ├── postgres-url.secret                           # postgres://{PRODUCT}_database_user:{PRODUCT}_database_pass@{PRODUCT}-postgres:5432/{PRODUCT}_database?sslmode=disable
    ├── unseal-1of5.secret                            # {PRODUCT}-unseal-key-1-of-5-{hex-random-32-bytes}
    ├── unseal-2of5.secret                            # {PRODUCT}-unseal-key-2-of-5-{hex-random-32-bytes}
    ├── unseal-3of5.secret                            # {PRODUCT}-unseal-key-3-of-5-{hex-random-32-bytes}
    ├── unseal-4of5.secret                            # {PRODUCT}-unseal-key-4-of-5-{hex-random-32-bytes}
    └── unseal-5of5.secret                            # {PRODUCT}-unseal-key-5-of-5-{hex-random-32-bytes}
```

**Total per product**: 4 `.secret.never` + 10 `.secret` = 14 files.

**Concrete example** (`sm` product, `PRODUCT=sm`):

```
hash-pepper-v3.secret           →  sm-hash-pepper-v3-Abcd1234Efgh5678Ijkl9012Mnop3456
browser-username.secret.never   →  (content: "MUST NEVER be used at product level")
browser-password.secret.never   →  (content: "MUST NEVER be used at product level")
service-username.secret.never   →  (content: "MUST NEVER be used at product level")
service-password.secret.never   →  (content: "MUST NEVER be used at product level")
postgres-username.secret        →  sm_database_user
postgres-password.secret        →  sm_database_pass-Qrst6789Uvwx0123Yzab4567Cdef8901
postgres-database.secret        →  sm_database
postgres-url.secret             →  postgres://sm_database_user:sm_database_pass-...@sm-postgres:5432/sm_database?sslmode=disable
unseal-1of5.secret              →  sm-unseal-key-1-of-5-{hex-random-32-bytes}
unseal-2of5.secret              →  sm-unseal-key-2-of-5-{hex-random-32-bytes}
unseal-3of5.secret              →  sm-unseal-key-3-of-5-{hex-random-32-bytes}
unseal-4of5.secret              →  sm-unseal-key-4-of-5-{hex-random-32-bytes}
unseal-5of5.secret              →  sm-unseal-key-5-of-5-{hex-random-32-bytes}
```

**All 5 products** (`identity`, `jose`, `pki`, `skeleton`, `sm`) follow this identical structure.

### F.3 Suite Deployment

```
deployments/cryptoutil-suite/                         # drwxr-x---
├── compose.yml                                       # Suite-level Docker Compose
└── secrets/
    ├── hash-pepper-v3.secret                         # cryptoutil-hash-pepper-v3-{base64-random-32-bytes}
    ├── browser-username.secret.never                 # MUST NEVER be used at suite level
    ├── browser-password.secret.never                 # MUST NEVER be used at suite level
    ├── service-username.secret.never                 # MUST NEVER be used at suite level
    ├── service-password.secret.never                 # MUST NEVER be used at suite level
    ├── postgres-username.secret                      # cryptoutil_database_user
    ├── postgres-password.secret                      # cryptoutil_database_pass-{base64-random-32-bytes}
    ├── postgres-database.secret                      # cryptoutil_database
    ├── postgres-url.secret                           # postgres://cryptoutil_database_user:cryptoutil_database_pass@cryptoutil-postgres:5432/cryptoutil_database?sslmode=disable
    ├── unseal-1of5.secret                            # cryptoutil-unseal-key-1-of-5-{hex-random-32-bytes}
    ├── unseal-2of5.secret                            # cryptoutil-unseal-key-2-of-5-{hex-random-32-bytes}
    ├── unseal-3of5.secret                            # cryptoutil-unseal-key-3-of-5-{hex-random-32-bytes}
    ├── unseal-4of5.secret                            # cryptoutil-unseal-key-4-of-5-{hex-random-32-bytes}
    └── unseal-5of5.secret                            # cryptoutil-unseal-key-5-of-5-{hex-random-32-bytes}
```

**Total**: 4 `.secret.never` + 10 `.secret` = 14 files. No Dockerfile (suite orchestrates via compose only).

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

### F.5 What Gets DELETED from deployments/

| Current Location | Reason |
|-----------------|--------|
| `deployments/template/` | Duplicate of `deployments/skeleton-template/` — merge then delete (Decision 5=C) |
| `deployments/archived/` | Dead code |
| `deployments/shared-citus/` | Citus removed — only PostgreSQL and SQLite supported |
| `deployments/deployments-all-files.json` | Build artifact, not in spec |
| `deployments/pki-ca/README.md` | Not in spec |
| `deployments/{PRODUCT}/secrets/{PRODUCT}-postgres-username.secret.never` | Legacy prefixed marker (all products) |
| `deployments/{PRODUCT}/secrets/{PRODUCT}-postgres-password.secret.never` | Legacy prefixed marker (all products) |
| `deployments/{PRODUCT}/secrets/{PRODUCT}-postgres-database.secret.never` | Legacy prefixed marker (all products) |
| `deployments/{PRODUCT}/secrets/{PRODUCT}-postgres-url.secret.never` | Legacy prefixed marker (all products) |
| `deployments/{PRODUCT}/secrets/{PRODUCT}-unseal-{1..5}of5.secret.never` | Legacy prefixed marker (all products) |
| `deployments/{PRODUCT}/secrets/sm-hash-pepper.secret` | Legacy file (only in sm) |
| `deployments/cryptoutil-suite/secrets/{SUITE}-hash-pepper.secret.never` | Legacy prefixed marker |
| `deployments/cryptoutil-suite/secrets/{SUITE}-postgres-*.secret.never` | Legacy prefixed markers |
| `deployments/cryptoutil-suite/secrets/{SUITE}-unseal-{1..5}of5.secret.never` | Legacy prefixed markers |

---

## G. internal/ — Private Application Code `drwxr-x---`

### G.1 internal/apps/ — Application Layer

```
internal/apps/                                        # drwxr-x---
├── {SUITE}/                                          # Suite orchestration
│   ├── {SUITE}.go                                    #   Suite CLI dispatch (seam pattern)
│   ├── *_test.go
│   └── e2e/                                          #   E2E tests (full suite docker compose)
│
├── {PRODUCT}/                                        # Product level (×5)
│   ├── {PRODUCT}.go                                  #   Product CLI dispatch
│   ├── *_test.go
│   ├── e2e/                                          #   E2E tests (full product docker compose)
│   └── shared/                                       #   Shared within product (optional)
│       └── (shared packages)/
│           ├── *.go
│           └── *_test.go
│
├── {PRODUCT}/{SERVICE}/                              # Service implementation (×N per product, 10 total)
│   ├── {SERVICE}.go                                  #   Service entry point (seam pattern)
│   ├── *_test.go
│   ├── integration/                                  #   Integration tests
│   ├── e2e/                                          #   E2E tests (service docker compose)
│   ├── repository/                                   #   Data access layer
│   │   ├── *.go                                      #     GORM entity models + repository methods
│   │   ├── *_test.go
│   │   └── migrations/                               #     Domain migrations (2001+)
│   │       ├── 2001_init.up.sql
│   │       └── 2001_init.down.sql
│   ├── model/                                        #   Domain models (optional)
│   │   └── *.go
│   └── handler/                                      #   HTTP handlers (optional)
│       └── *.go
│
├── framework/                                        # Service framework (shared by ALL services)
│   ├── apperr/                                       #   Application error types (moved from shared/apperr/)
│   ├── suite/                                        #   Suite-level framework
│   │   └── cli/
│   │       ├── suite_router.go                       #     RouteSuite(), SuiteConfig, ProductEntry
│   │       └── suite_router_test.go
│   ├── product/                                      #   Product-level framework
│   │   └── cli/
│   │       ├── product_router.go                     #     RouteProduct(), ProductConfig, ServiceEntry
│   │       └── product_router_test.go
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
│   │   ├── common/
│   │   ├── format_go/
│   │   ├── format_gotest/
│   │   ├── lint_compose/
│   │   ├── lint_deployments/                         #   Deployment structure validator (8 validators)
│   │   ├── lint_docs/                                #   Documentation linter (includes docs_validation)
│   │   ├── lint_fitness/                             #   Architecture fitness functions
│   │   │   ├── registry/                             #     Entity registry (SSOT)
│   │   │   │   ├── registry.go
│   │   │   │   └── registry_test.go
│   │   │   ├── banned_product_names/
│   │   │   ├── circular_deps/
│   │   │   ├── configs_naming/                       #     Validates FLAT configs/{PS-ID}/ pattern
│   │   │   ├── entity_registry_completeness/
│   │   │   ├── file_size/
│   │   │   ├── parallel_tests/
│   │   │   ├── test_patterns/
│   │   │   └── ... (44+ linters)
│   │   ├── lint_go/
│   │   ├── lint_golangci/
│   │   ├── lint_gotest/
│   │   ├── lint_go_mod/
│   │   ├── lint_ports/
│   │   ├── lint_text/
│   │   └── lint_workflow/
│   │
│   └── workflow/                                     #   GitHub Actions workflow management
│       └── *.go                                      #     run + cleanup subcommands
│
└── (DELETE)
    ├── demo/                                         #   Dead code
    └── pkiinit/                                      #   Merged → framework/tls/
```

**Consolidations required**:

- `docs_validation/` → merged into `lint_docs/` (single documentation linter)
- `github_cleanup/` → merged into `tools/workflow/` (subcommands: run, cleanup)
- `configs_naming` fitness linter rewritten to validate **flat** `configs/{PS-ID}/` pattern

### G.2 internal/shared/ — Shared Libraries `drwxr-x---`

```
internal/shared/                                      # drwxr-x---
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
│   ├── magic_api.go
│   ├── magic_cicd.go
│   ├── magic_console.go
│   ├── magic_crypto.go
│   ├── magic_database.go
│   ├── magic_docker.go
│   ├── magic_framework.go
│   ├── magic_{PRODUCT}.go                            # Per-product constants (×5)
│   ├── magic_{PRODUCT}_{topic}.go                    # Per-product topic files (identity has ~12)
│   ├── magic_misc.go
│   ├── magic_network.go
│   ├── magic_orchestration.go
│   ├── magic_percent.go
│   ├── magic_security.go
│   ├── magic_session.go
│   ├── magic_telemetry.go
│   ├── magic_testing.go
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

---

## H. docs/ — Documentation `drwxr-x---`

```
docs/                                                 # drwxr-x---
├── ARCHITECTURE.md                                   # SSOT: Architecture reference (5080+ lines)
├── CONFIG-SCHEMA.md                                  # Config file schema reference
├── DEV-SETUP.md                                      # Developer setup guide
├── README.md                                         # Documentation index
├── UPDATE-TOOLS.md                                   # VS Code / MCP tool catalog and update guide
└── framework-v6/                                     # Current active plan
    ├── plan.md
    ├── tasks.md
    ├── lessons.md
    └── target-structure.md                           # THIS FILE
```

**DELETE** (historical and stale docs):

| Entry | Reason |
|-------|--------|
| `docs/framework-v3/` | Historical plan (completed) |
| `docs/framework-v4/` | Historical plan (completed) |
| `docs/framework-v5/` | Superseded by framework-v6/ |
| `docs/LESSONS/` | Cross-plan lessons archive (superseded by per-plan lessons.md) |
| `docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md` | After merge into ARCHITECTURE.md |
| `docs/ARCHITECTURE-INDEX.md` | Superseded by ARCHITECTURE.md ToC |
| `docs/ARCHITECTURE-TODO.md` | Superseded by plan tracking |
| `docs/COPILOT-MULTI-PROJECT.md` | Stale reference doc |
| `docs/DEAD_CODE_REVIEW.md` | Completed, no longer needed |
| `docs/VSCODE-CRASHES.md` | Stale troubleshooting doc |
| `docs/demo-brainstorm/` | Demos archived |
| `docs/framework-brainstorm/` | Superseded by framework-v3+ |
| `docs/gremlins/` | Stale mutation testing notes |
| `docs/workflow-runtimes/` | Stale workflow analysis |

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

All tiers (service, product, suite) use **identical `{purpose}.secret` names** —
no tier prefix on active secret files. Tier prefixes appear ONLY on `.secret.never`
marker files.

| Secret Purpose | Filename | Service Value Pattern | Product Value Pattern | Suite Value Pattern |
|---------------|----------|-----------------------|-----------------------|---------------------|
| Hash pepper v3 | `hash-pepper-v3.secret` | `{PS-ID}-hash-pepper-v3-{base64-random-32-bytes}` | `{PRODUCT}-hash-pepper-v3-{base64}` | `cryptoutil-hash-pepper-v3-{base64}` |
| Browser username | `browser-username.secret` | `{PS-ID}-browser-user` | `.never` only | `.never` only |
| Browser password | `browser-password.secret` | `{PS-ID}-browser-pass-{base64-random-32-bytes}` | `.never` only | `.never` only |
| Service username | `service-username.secret` | `{PS-ID}-service-user` | `.never` only | `.never` only |
| Service password | `service-password.secret` | `{PS-ID}-service-pass-{base64-random-32-bytes}` | `.never` only | `.never` only |
| PostgreSQL username | `postgres-username.secret` | `{PS_ID}_database_user` | `{PRODUCT}_database_user` | `cryptoutil_database_user` |
| PostgreSQL password | `postgres-password.secret` | `{PS_ID}_database_pass-{base64-random-32-bytes}` | `{PRODUCT}_database_pass-{base64}` | `cryptoutil_database_pass-{base64}` |
| PostgreSQL database | `postgres-database.secret` | `{PS_ID}_database` | `{PRODUCT}_database` | `cryptoutil_database` |
| PostgreSQL URL | `postgres-url.secret` | `postgres://{PS_ID}_database_user:{PS_ID}_database_pass@{PS-ID}-postgres:5432/{PS_ID}_database?sslmode=disable` | `postgres://{PRODUCT}_database_user:{PRODUCT}_database_pass@{PRODUCT}-postgres:5432/{PRODUCT}_database?sslmode=disable` | `postgres://cryptoutil_database_user:cryptoutil_database_pass@cryptoutil-postgres:5432/cryptoutil_database?sslmode=disable` |
| Unseal shard N | `unseal-{N}of5.secret` | `{PS-ID}-unseal-key-N-of-5-{hex-random-32-bytes}` | `{PRODUCT}-unseal-key-N-of-5-{hex-random-32-bytes}` | `cryptoutil-unseal-key-N-of-5-{hex-random-32-bytes}` |

**`.secret.never` marker files** — present at product and suite tiers as explicit reminders:

| Tier | Files Present | Content |
|------|-------------|---------|
| Product (×5) | `browser-password.secret.never`, `browser-username.secret.never`, `service-password.secret.never`, `service-username.secret.never` | "MUST NEVER be used at product level. Use service-specific secrets." |
| Suite (×1) | Same 4 filenames | "MUST NEVER be used at suite level. Use service-specific secrets." |

**Total `.secret.never` files**: 4 per product × 5 products + 4 suite = **24 files**.

---

## M. Fitness Linter Coverage (New/Enhanced in v6)

| Linter | Scope | Rule |
|--------|-------|------|
| `root-junk-detection` | `{ROOT}/` | No `*.exe`, `*.py`, `coverage*`, `*.test.exe` at root |
| `cmd-entry-whitelist` | `cmd/` | Only 18 allowed entries (1 suite + 5 products + 10 services + 2 infra tools) |
| `configs-structure` | `configs/` | Must follow flat `{SUITE}/`, `{PRODUCT}/`, `{PS-ID}/` hierarchy (Decision 2=B) |
| `configs-naming` (rewritten) | `configs/` | Validates flat `{PS-ID}/{PS-ID}.yml` pattern; rejects nested `{PRODUCT}/{SERVICE}/`; allows `pki-ca/profiles/` and `identity-authz/domain/policies/` exceptions |
| `configs-no-deployment` | `configs/` | No deployment variants (`*-pg-1.yml`, `*-sqlite.yml`) or environment files |
| `secret-naming` | `deployments/*/secrets/` | All tiers use `{purpose}.secret` names; `.never` markers enforced at product/suite |
| `template-consistency` | `deployments/skeleton-template/` | Hyphens in secret names (not underscores) |
| `archive-detection` | `**/*archived*/`, `**/*orphaned*/` | No archived/orphaned directories |
| `entity-registry-completeness` | (existing, enhanced) | Verify `configs/{PS-ID}/` existence for all registered PS-IDs |

---

## N. Change Summary (Current → Post-v6 Target)

| Area | Current State | Target State | Action |
|------|--------------|-------------|--------|
| Root files | ~80+ junk artifacts | Clean project config only | DELETE artifacts |
| `.vscode/mcp.json` | Missing | Present | CREATE |
| `cmd/` | 18 entries + extras | 18 entries exactly | DELETE demo, identity-compose, identity-demo |
| `api/` | Missing components for some services | All 10 PS-IDs with full generated spec | CREATE missing |
| `configs/` | Nested `{PRODUCT}/{SERVICE}/` dirs | Flat `{PS-ID}/` dirs + `{PRODUCT}/` dirs | RESTRUCTURE (Decision 2=B) |
| `configs/` service filenames | `{SERVICE}.yml` (e.g., `im.yml`) | `{PS-ID}.yml` (e.g., `sm-im.yml`) | RENAME |
| `configs/pki-ca/profiles/` | At `configs/pki/ca/profiles/` | At `configs/pki-ca/profiles/` | MOVE (Decision 3=B) |
| `configs/identity/policies/` | At `configs/identity/policies/` | At `configs/identity-authz/domain/policies/` | MOVE + RENAME (Decision 4=A) |
| `configs/identity/policies/adaptive-auth.yml` | `adaptive-auth.yml` (banned term) | `adaptive-authorization.yml` | RENAME (Decision 4=A) |
| `deployments/` service sqlite-2 | Missing in all 10 services | Present in all 10 services | CREATE (RC-3) |
| `deployments/` product Dockerfile | Missing in all 5 products | Present in all 5 products | CREATE |
| `deployments/template/` | Still exists | Deleted (merged → skeleton-template) | MERGE + DELETE (Decision 5=C) |
| `deployments/` archived | Present | Deleted | DELETE |
| `deployments/` shared-citus | Present | Deleted | DELETE |
| `deployments/deployments-all-files.json` | Present | Deleted | DELETE |
| Service unseal prefix | `{SERVICE}-{hex}` (e.g., `im-{hex}`) | `{PS-ID}-unseal-key-N-of-5-{hex}` | FIX (Decision 1=A) |
| Product unseal value | `dev-unseal-key-N-of-5` | `{PRODUCT}-unseal-key-N-of-5-{hex}` | FIX (Decision 1=A) |
| Suite unseal prefix | `suite-` | `cryptoutil-` | FIX (Decision 1=A) |
| `pki-ca` unseal | Copy of sm-kms values | Unique `pki-ca-` prefixed values | REGENERATE |
| Service postgres DB | `{PS_ID}` (e.g., `sm_im`) | `{PS_ID}_database` | FIX (Decision 6=A) |
| Service postgres user | `{PS_ID}_user` | `{PS_ID}_database_user` | FIX (Decision 6=A) |
| Product postgres DB | Not standardized | `{PRODUCT}_database` | FIX (Decision 6=A) |
| `.secret.never` at product | 0 files | 4 files per product (20 total) | CREATE (RC-3) |
| `.secret.never` at suite | 0 files | 4 files (24 total with products) | CREATE (RC-3) |
| Legacy prefixed `.never` files | Present at products/suite | Deleted | DELETE |
| `internal/apps/tools/cicd_lint/configs_naming/` | Validates nested pattern | Validates flat pattern | REWRITE |
| `internal/` demo, pkiinit | Present | Deleted / merged into `framework/tls/` | DELETE / MERGE |
| `internal/shared/apperr/` | Present | Moved to `internal/apps/framework/apperr/` | MOVE |
| `internal/apps/tools/docs_validation/` | Separate dir | Merged into `lint_docs/` | MERGE |
| `internal/apps/tools/github_cleanup/` | Separate dir | Merged into `tools/workflow/` as subcommand | MERGE |
| `.github/agents/doc-sync.agent.md` | Present | Deleted | DELETE |
| `.github/actions/custom-cicd-lint/` | Present | Renamed to `download-cicd/` | RENAME |
| `docs/framework-v5/` | Active plan | Historical — delete | DELETE |
| `docs/UPDATE-TOOLS.md` | Missing | Present | CREATE |
| `docs/` stale (framework-v3/v4, LESSONS/, etc.) | Present | Deleted | DELETE |
| `testdata/` | Present (1 sample file) | Deleted (move to owning package) | DELETE |

---

## O. Open Questions

All questions resolved via plan.md quizme answers:

- Decision 1=A: Unseal naming `{PS-ID}-unseal-key-N-of-5-{hex-random-32-bytes}`
- Decision 2=B: Flat `configs/{PS-ID}/` (NOT nested `configs/{PRODUCT}/{SERVICE}/`)
- Decision 3=B: Keep `configs/pki-ca/profiles/` as valid subdir exception
- Decision 4=A: Identity policies → `configs/identity-authz/domain/policies/`; rename `adaptive-auth.yml` → `adaptive-authorization.yml`
- Decision 5=C: Merge `deployments/template/` into `deployments/skeleton-template/` then delete
- Decision 6=A: Postgres DB = `{PS_ID}_database`, Postgres user = `{PS_ID}_database_user`
