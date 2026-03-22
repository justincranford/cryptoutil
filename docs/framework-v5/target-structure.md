# Target Repository Structure - Framework v5

**Status**: DRAFT — User review required before execution
**Created**: 2025-06-22
**Last Updated**: 2025-06-22
**Purpose**: Define the complete, parameterized target state of every directory
and file in the repository. Once approved, fitness linters enforce this structure
and automation moves current state toward it.

---

## Entity Hierarchy (Canonical)

| Level | Variable | Instances | Count |
|-------|----------|-----------|-------|
| Suite | `{SUITE}` | `cryptoutil` | 1 |
| Product | `{PRODUCT}` | `identity`, `jose`, `pki`, `skeleton`, `sm` | 5 |
| Service | `{SERVICE}` | varies per product (see below) | 10 total |
| PS-ID | `{PS-ID}` = `{PRODUCT}-{SERVICE}` | see table below | 10 |
| PS_ID | `{PS_ID}` = `{PRODUCT}_{SERVICE}` | underscore variant for SQL/secrets | 10 |
| Infra Tool | N/A | `cicd`, `workflow` | 2 |
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

All directory and file permissions shown in this document follow this convention:

| Target | Permission | Octal | Description |
|--------|-----------|-------|-------------|
| Directories | `drwxr-xr-x` | 755 | Owner rwx, group/other rx |
| Source files (`.go`, `.yml`, `.yaml`, `.md`, `.sql`) | `-rw-r--r--` | 644 | Owner rw, group/other r |
| Secret files (`.secret`) | `-r--r-----` | 440 | Owner/group r only, no other |
| Secret marker files (`.secret.never`) | `-r--r-----` | 440 | Same as secrets |
| Executable scripts (`mvnw`) | `-rwxr-xr-x` | 755 | Owner rwx, group/other rx |
| Generated files (`*.gen.go`) | `-rw-r--r--` | 644 | Same as source |

---

## A. Root Level

### A.1 Root Files (KEEP — legitimate project config) `drwxr-xr-x`

```
{ROOT}/                                    # drwxr-xr-x
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

### A.2 Root Files (DELETE — junk artifacts, ~80+ files)

All `*.exe`, `*.py`, `coverage*`, `*_coverage`, `*.test.exe` files at root are
build/test artifacts that should never be committed. Git history preserves them.

### A.3 Root Hidden Directories `drwxr-xr-x`

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
│   └── settings.json                      #   Workspace settings
└── .zap/                                  # OWASP ZAP DAST config
    └── rules.tsv                          #   ZAP scan rules
```

---

## B. .github/ — GitHub & Copilot Configuration `drwxr-xr-x`

```
.github/
├── copilot-instructions.md                # Copilot config hub (loads instructions/)
├── agents/                                # Copilot chat agents
│   ├── beast-mode.agent.md                #   Continuous autonomous execution
│   ├── doc-sync.agent.md                  #   Documentation sync agent
│   ├── fix-workflows.agent.md             #   CI/CD workflow fixer
│   ├── implementation-execution.agent.md  #   Plan execution agent
│   └── implementation-planning.agent.md   #   Plan creation agent
├── actions/                               # Reusable GitHub Actions
│   ├── custom-cicd-lint/action.yml        #   Custom CICD lint composite action
│   ├── docker-compose-build/action.yml
│   ├── docker-compose-down/action.yml
│   ├── docker-compose-logs/action.yml
│   ├── docker-compose-up/action.yml
│   ├── docker-compose-verify/action.yml
│   ├── docker-images-pull/action.yml      #   Parallel Docker image pre-pull
│   ├── fuzz-test/action.yml
│   ├── go-setup/action.yml                #   Go toolchain setup with caching
│   ├── golangci-lint/action.yml           #   golangci-lint v2 execution
│   ├── security-scan-gitleaks/action.yml
│   ├── security-scan-trivy/action.yml
│   ├── security-scan-trivy2/action.yml
│   ├── workflow-job-begin/action.yml      #   Job telemetry start
│   └── workflow-job-end/action.yml        #   Job telemetry end
├── instructions/                          # Copilot instruction files (auto-loaded alpha order)
│   ├── 01-01.terminology.instructions.md
│   ├── 01-02.beast-mode.instructions.md
│   ├── 02-01.architecture.instructions.md
│   ├── 02-02.versions.instructions.md
│   ├── 02-03.observability.instructions.md
│   ├── 02-04.openapi.instructions.md
│   ├── 02-05.security.instructions.md
│   ├── 02-06.authn.instructions.md
│   ├── 03-01.coding.instructions.md
│   ├── 03-02.testing.instructions.md
│   ├── 03-03.golang.instructions.md
│   ├── 03-04.data-infrastructure.instructions.md
│   ├── 03-05.linting.instructions.md
│   ├── 04-01.deployment.instructions.md
│   ├── 05-01.cross-platform.instructions.md
│   ├── 05-02.git.instructions.md
│   ├── 06-01.evidence-based.instructions.md
│   └── 06-02.agent-format.instructions.md
├── skills/                                # Copilot skills (slash commands)
│   ├── README.md                          #   Skill catalog
│   ├── agent-scaffold/SKILL.md
│   ├── contract-test-gen/SKILL.md
│   ├── coverage-analysis/SKILL.md
│   ├── fips-audit/SKILL.md
│   ├── fitness-function-gen/SKILL.md
│   ├── instruction-scaffold/SKILL.md
│   ├── migration-create/SKILL.md
│   ├── new-service/SKILL.md
│   ├── openapi-codegen/SKILL.md
│   ├── propagation-check/SKILL.md
│   ├── skill-scaffold/SKILL.md
│   ├── test-benchmark-gen/SKILL.md
│   ├── test-fuzz-gen/SKILL.md
│   └── test-table-driven/SKILL.md
└── workflows/                             # GitHub Actions CI/CD workflows
    ├── ci-benchmark.yml                   #   Benchmark testing
    ├── ci-coverage.yml                    #   Code coverage analysis
    ├── ci-dast.yml                        #   Dynamic application security testing
    ├── ci-e2e.yml                         #   End-to-end testing
    ├── ci-fitness.yml                     #   Architecture fitness functions
    ├── ci-fuzz.yml                        #   Fuzz testing
    ├── ci-gitleaks.yml                    #   Secret detection
    ├── ci-identity-validation.yml         #   Identity service validation
    ├── ci-load.yml                        #   Load testing (Gatling)
    ├── ci-mutation.yml                    #   Mutation testing (gremlins)
    ├── ci-quality.yml                     #   Build + lint + unit tests
    ├── ci-race.yml                        #   Race condition detection
    ├── ci-sast.yml                        #   Static application security testing
    ├── cicd-lint-deployments.yml          #   Deployment structure validation
    └── release.yml                        #   Release workflow
```

---

## C. cmd/ — Binary Entry Points `drwxr-xr-x`

**Pattern**: Each entry has exactly one `main.go` that delegates to `internal/apps/`.

```
cmd/                                                  # drwxr-xr-x
├── {SUITE}/main.go                                   # Suite CLI → internal/apps/{SUITE}/
│                                                     #   e.g. cmd/cryptoutil/main.go
│
├── {PRODUCT}/main.go                                 # Product CLI → internal/apps/{PRODUCT}/  (×5)
│                                                     #   cmd/identity/main.go
│                                                     #   cmd/jose/main.go
│                                                     #   cmd/pki/main.go
│                                                     #   cmd/skeleton/main.go
│                                                     #   cmd/sm/main.go
│
├── {PS-ID}/main.go                                   # Service CLI → internal/apps/{PRODUCT}/{SERVICE}/  (×10)
│                                                     #   cmd/identity-authz/main.go
│                                                     #   cmd/identity-idp/main.go
│                                                     #   cmd/identity-rp/main.go
│                                                     #   cmd/identity-rs/main.go
│                                                     #   cmd/identity-spa/main.go
│                                                     #   cmd/jose-ja/main.go
│                                                     #   cmd/pki-ca/main.go
│                                                     #   cmd/skeleton-template/main.go
│                                                     #   cmd/sm-im/main.go
│                                                     #   cmd/sm-kms/main.go
│
├── cicd/main.go                                      # CICD tool → internal/apps/tools/cicd/
└── workflow/main.go                                  # Workflow runner → internal/apps/tools/workflow/
```

**Allowed entries**: 1 suite + 5 products + 10 services + 2 infra = **18 total**

**DELETE** (per Decision 2): `cmd/demo/`, `cmd/identity-compose/`, `cmd/identity-demo/`

---

## D. api/ — OpenAPI Specifications & Generated Code `drwxr-xr-x`

**Pattern**: One directory per product-service, containing the OpenAPI spec and
oapi-codegen output.

```
api/                                                  # drwxr-xr-x
└── {PS-ID}/                                          # (×10)
    ├── generate.go                                   # //go:generate oapi-codegen directives
    ├── openapi_spec.yaml                             # OpenAPI 3.0.3 specification (SSOT for API)
    ├── openapi_spec_components.yaml                  # Reusable schema components (required)
    ├── openapi_spec_paths.yaml                       # API path definitions (required)
    ├── openapi-gen_config_client.yaml                # oapi-codegen client generation config
    ├── openapi-gen_config_models.yaml                # oapi-codegen models generation config
    ├── openapi-gen_config_server.yaml                # oapi-codegen server generation config
    ├── client/
    │   └── client.gen.go                             # Generated HTTP client
    ├── models/
    │   └── models.gen.go                             # Generated request/response models
    └── server/
        └── server.gen.go                             # Generated strict server interface
```

---

## E. configs/ — Canonical Application Configuration (SSOT) `drwxr-xr-x`

**Principle**: `configs/` is the **single source of truth** for what the app needs —
environment-agnostic, reusable configuration. Deployment-specific overlays live
in `deployments/`.

### E.1 Suite Level

```
configs/                                              # drwxr-xr-x
└── {SUITE}/                                          # Suite-level config
    └── {SUITE}.yml                                   # Suite orchestration config
                                                      #   e.g. configs/cryptoutil/cryptoutil.yml
```

### E.2 Product Level (Parameterized ×5)

```
configs/                                              # drwxr-xr-x
└── {PRODUCT}/                                        # Product directory
    └── {PRODUCT}.yml                                 # Canonical product domain config
                                                      #   e.g. configs/sm/sm.yml
                                                      #   e.g. configs/jose/jose.yml
                                                      #   e.g. configs/pki/pki.yml
                                                      #   e.g. configs/identity/identity.yml
                                                      #   e.g. configs/skeleton/skeleton.yml
```

### E.3 Product-Service Level (Parameterized ×10)

**FLAT PS-ID directories** — NOT nested `configs/{PRODUCT}/{SERVICE}/`.

```
configs/                                              # drwxr-xr-x
└── {PS-ID}/                                          # Service directory (flat, one per PS-ID)
    └── {PS-ID}.yml                                   # Canonical service domain config
                                                      #   e.g. configs/sm-kms/sm-kms.yml
                                                      #   e.g. configs/sm-im/sm-im.yml
                                                      #   e.g. configs/jose-ja/jose-ja.yml
                                                      #   e.g. configs/pki-ca/pki-ca.yml
                                                      #   e.g. configs/identity-authz/identity-authz.yml
                                                      #   e.g. configs/identity-idp/identity-idp.yml
                                                      #   e.g. configs/identity-rp/identity-rp.yml
                                                      #   e.g. configs/identity-rs/identity-rs.yml
                                                      #   e.g. configs/identity-spa/identity-spa.yml
                                                      #   e.g. configs/skeleton-template/skeleton-template.yml
```

### E.4 Product-Level Shared Config (where applicable)

Some products have shared configuration that applies across all services within
that product. These live at the product level under `configs/{PRODUCT}/{SERVICE}/domain/`
for domain-specific configuration files.

```
configs/                                              # drwxr-xr-x
├── identity/                                         # Identity product
│   ├── identity.yml                                  # Product-level config (from E.2)
│   ├── authz/
│   │   ├── authz.yml                                 # AuthZ service domain config
│   │   └── domain/                                   # Domain-specific configuration files
│   ├── idp/
│   │   ├── idp.yml                                   # IdP service domain config
│   │   └── domain/                                   # Domain-specific configuration files
│   ├── rp/
│   │   ├── rp.yml                                    # RP service domain config
│   │   └── domain/                                   # Domain-specific configuration files
│   ├── rs/
│   │   ├── rs.yml                                    # RS service domain config
│   │   └── domain/                                   # Domain-specific configuration files
│   └── spa/
│       ├── spa.yml                                   # SPA service domain config
│       └── domain/                                   # Domain-specific configuration files
│
├── pki/
│   └── ca/
│       ├── ca.yml                                    # CA service domain config (rename from ca-server.yml)
│       └── domain/                                   # Domain-specific configuration files
│
├── jose/
│   └── ja/
│       ├── ja.yml                                    # JA service domain config
│       └── domain/                                   # Domain-specific configuration files
│
├── skeleton/
│   └── template/
│       ├── template.yml                              # Template service domain config
│       └── domain/                                   # Domain-specific configuration files
│
└── sm/
    ├── im/
    │   ├── im.yml                                    # IM service domain config
    │   └── domain/                                   # Domain-specific configuration files
    └── kms/
        ├── kms.yml                                   # KMS service domain config
        └── domain/                                   # Domain-specific configuration files
```

### E.5 What MOVES OUT of configs/ (to deployments/)

These files are **deployment-specific** and belong in `deployments/`, not `configs/`:

| Current Location | Reason to Move | Target |
|-----------------|----------------|--------|
| `configs/identity/development.yml` | Environment-specific | `deployments/identity/config/` or delete |
| `configs/identity/production.yml` | Environment-specific | `deployments/identity/config/` or delete |
| `configs/identity/test.yml` | Environment-specific | `deployments/identity/config/` or delete |
| `configs/identity/profiles/` (authz-idp.yml, ci.yml, demo.yml, full-stack.yml) | Compose deployment profiles | `deployments/identity/config/` |
| `configs/identity/*/authz-docker.yml` | Docker-specific overlay | `deployments/identity-authz/config/` |
| `configs/identity/*/idp-docker.yml` | Docker-specific overlay | `deployments/identity-idp/config/` |
| `configs/identity/*/rp-docker.yml` | Docker-specific overlay | `deployments/identity-rp/config/` |
| `configs/identity/*/rs-docker.yml` | Docker-specific overlay | `deployments/identity-rs/config/` |
| `configs/identity/*/spa-docker.yml` | Docker-specific overlay | `deployments/identity-spa/config/` |
| `configs/sm/im/config-pg-1.yml` | Deployment variant | Already in `deployments/sm-im/config/` |
| `configs/sm/im/config-pg-2.yml` | Deployment variant | Already in `deployments/sm-im/config/` |
| `configs/sm/im/config-sqlite.yml` | Deployment variant | Already in `deployments/sm-im/config/` |
| `configs/sm/kms/config-pg-1.yml` | Deployment variant | Already in `deployments/sm-kms/config/` |
| `configs/sm/kms/config-pg-2.yml` | Deployment variant | Already in `deployments/sm-kms/config/` |
| `configs/sm/kms/config-sqlite.yml` | Deployment variant | Already in `deployments/sm-kms/config/` |

### E.6 What Gets DELETED from configs/

| Current Location | Reason |
|-----------------|--------|
| `configs/orphaned/` | Dead code (Decision 3) |
| `configs/ca/` (contents) | Moved to `configs/pki/ca/` |
| `configs/configs-all-files.json` | Metadata artifact, not config |
| `configs/jose/jose-server.yml` | Replaced by `configs/jose/ja/ja.yml` |
| `configs/skeleton/skeleton-server.yml` | Replaced by `configs/skeleton/template/template.yml` |

---

## F. deployments/ — Deployment Manifests & Wiring `drwxr-xr-x`

**Principle**: `deployments/` contains environment-specific deployment manifests that
CONSUME configuration from `configs/`. Config files here are deployment overlays, NOT
the canonical source.

### F.1 Service-Level Deployments (×10)

```
deployments/                                          # drwxr-xr-x
└── {PS-ID}/                                          # One per product-service
    ├── compose.yml                                   # Docker Compose manifest
    ├── Dockerfile                                    # Multi-stage build (builder → validator → runtime)
    ├── config/                                       # Deployment-specific config overlays
    │   ├── {PS-ID}-app-common.yml                    #   Shared across all instances (bind addresses,
    │   │                                             #     TLS paths, Docker network hostnames)
    │   ├── {PS-ID}-app-sqlite-1.yml                  #   SQLite in-memory instance 1
    │   │                                             #     database-driver: sqlite
    │   │                                             #     database-url: "file::memory:?cache=shared"
    │   ├── {PS-ID}-app-sqlite-2.yml                  #   SQLite in-memory instance 2
    │   │                                             #     database-driver: sqlite
    │   │                                             #     database-url: "file::memory:?cache=shared"
    │   ├── {PS-ID}-app-postgresql-1.yml              #   PostgreSQL logical database instance 1
    │   │                                             #     database-driver: postgres
    │   │                                             #     database-url: file:///run/secrets/postgres-url.secret
    │   └── {PS-ID}-app-postgresql-2.yml              #   PostgreSQL logical database instance 2
    │                                                 #     database-driver: postgres
    │                                                 #     database-url: file:///run/secrets/postgres-url.secret
    └── secrets/                                      # Docker secrets (chmod 440)
        ├── browser-password.secret                   #   value: {PS-ID}-browser-{base64-random}
        │                                             #     e.g. sm-im-browser-ZRWjFFiRHMGps8E+xiwt1A==
        ├── browser-username.secret                   #   value: {PS-ID}-browser-user
        │                                             #     e.g. sm-im-browser-user
        ├── hash-pepper-v3.secret                     #   value: {base64-random-32-bytes}
        │                                             #     e.g. txvakOQ3is9DDHfdAtjoa8sl2AjgqOHZSk0ggjOlk0M=
        ├── postgres-database.secret                  #   value: {PS_ID}
        │                                             #     e.g. sm_im
        ├── postgres-password.secret                  #   value: {PS_ID}_pass
        │                                             #     e.g. sm_im_pass
        ├── postgres-url.secret                       #   value: postgres://{PS_ID}_user:{PS_ID}_pass@{PS-ID}-postgres:5432/{PS_ID}?sslmode=disable
        │                                             #     e.g. postgres://sm_im_user:sm_im_pass@sm-im-postgres:5432/sm_im?sslmode=disable
        ├── postgres-username.secret                  #   value: {PS_ID}_user
        │                                             #     e.g. sm_im_user
        ├── service-password.secret                   #   value: {PS-ID}-service-{base64-random}
        │                                             #     e.g. sm-im-service-cIu5DadDObrS+rP49XwrYw==
        ├── service-username.secret                   #   value: {PS-ID}-service-user
        │                                             #     e.g. sm-im-service-user
        ├── unseal-1of5.secret                        #   value: {SERVICE}-{hex-random-32-bytes}
        │                                             #     e.g. im-0d6dfc52f2517a2820e11859fe9e4f3c
        ├── unseal-2of5.secret                        #   value: {SERVICE}-{hex-random-32-bytes}
        ├── unseal-3of5.secret                        #   value: {SERVICE}-{hex-random-32-bytes}
        ├── unseal-4of5.secret                        #   value: {SERVICE}-{hex-random-32-bytes}
        └── unseal-5of5.secret                        #   value: {SERVICE}-{hex-random-32-bytes}
```

### F.2 Product-Level Deployments (×5)

```
deployments/                                          # drwxr-xr-x
└── {PRODUCT}/                                        # One per product
    ├── compose.yml                                   # Product-level compose (delegates to services)
    ├── Dockerfile                                    # Product-level multi-stage build
    └── secrets/                                      # Product-level secrets (shared across services)
        ├── hash-pepper-v3.secret                     #   value: dev-hash-pepper-v3
        ├── postgres-database.secret                  #   value: cryptoutil
        ├── postgres-password.secret                  #   value: cryptoutil-dev-password
        ├── postgres-url.secret                       #   value: postgres://cryptoutil:cryptoutil-dev-password@{PRODUCT}-postgres:5432/cryptoutil?sslmode=disable
        ├── postgres-username.secret                  #   value: cryptoutil
        ├── sm-hash-pepper.secret                     #   value: CHANGE_ME_IN_PRODUCTION (legacy, product-specific)
        ├── unseal-1of5.secret                        #   value: dev-unseal-key-1-of-5
        ├── unseal-2of5.secret                        #   value: dev-unseal-key-2-of-5
        ├── unseal-3of5.secret                        #   value: dev-unseal-key-3-of-5
        ├── unseal-4of5.secret                        #   value: dev-unseal-key-4-of-5
        ├── unseal-5of5.secret                        #   value: dev-unseal-key-5-of-5
        │
        │   # .secret.never marker files (KEEP as explicit reminders)
        ├── {PRODUCT}-postgres-database.secret.never  #   "PostgreSQL secrets MUST NOT be shared at {PRODUCT} level"
        ├── {PRODUCT}-postgres-password.secret.never  #   Use service-specific secrets instead
        ├── {PRODUCT}-postgres-url.secret.never       #   See: ARCHITECTURE.md Section 12.3.4
        ├── {PRODUCT}-postgres-username.secret.never
        ├── {PRODUCT}-unseal-1of5.secret.never        #   "Unseal secrets MUST NOT be shared at {PRODUCT} level"
        ├── {PRODUCT}-unseal-2of5.secret.never
        ├── {PRODUCT}-unseal-3of5.secret.never
        ├── {PRODUCT}-unseal-4of5.secret.never
        └── {PRODUCT}-unseal-5of5.secret.never
```

### F.3 Suite-Level Deployment (×1)

```
deployments/                                          # drwxr-xr-x
└── {SUITE}-suite/                                    # e.g. cryptoutil-suite/
    ├── compose.yml                                   # Suite-level compose (all 5 products → transitively all 10 services)
    ├── Dockerfile                                    # Suite-level multi-stage build
    └── secrets/                                      # Suite-level secrets
        ├── postgres-database.secret                  #   value: (shared suite-level dev value)
        ├── postgres-password.secret                  #   value: (shared suite-level dev value)
        ├── postgres-url.secret                       #   value: (shared suite-level dev value)
        ├── postgres-username.secret                  #   value: (shared suite-level dev value)
        ├── unseal-1of5.secret                        #   value: (shared suite-level dev value)
        ├── unseal-2of5.secret                        #   value: (shared suite-level dev value)
        ├── unseal-3of5.secret                        #   value: (shared suite-level dev value)
        ├── unseal-4of5.secret                        #   value: (shared suite-level dev value)
        ├── unseal-5of5.secret                        #   value: (shared suite-level dev value)
        │
        │   # .secret.never marker files (KEEP as explicit reminders)
        ├── {SUITE}-hash-pepper.secret.never          #   "Secrets MUST NOT be shared at {SUITE} level"
        ├── {SUITE}-postgres-database.secret.never    #   Use service-specific secrets instead
        ├── {SUITE}-postgres-password.secret.never    #   See: ARCHITECTURE.md Section 12.3.4
        ├── {SUITE}-postgres-url.secret.never
        ├── {SUITE}-postgres-username.secret.never
        ├── {SUITE}-unseal-1of5.secret.never
        ├── {SUITE}-unseal-2of5.secret.never
        ├── {SUITE}-unseal-3of5.secret.never
        ├── {SUITE}-unseal-4of5.secret.never
        └── {SUITE}-unseal-5of5.secret.never
```

### F.4 Shared Infrastructure

```
deployments/                                          # drwxr-xr-x
├── shared-telemetry/                                 # OpenTelemetry + Grafana LGTM
│   └── compose.yml                                   #   otel-collector-contrib + grafana-otel-lgtm
├── shared-postgres/                                  # Shared PostgreSQL container
│   └── compose.yml                                   #   PostgreSQL for multi-service sharing;
│                                                     #   every service gets a logical database in this instance
└── shared-citus/                                     # Shared Citus (distributed PostgreSQL)
    └── compose.yml                                   #   Citus for distributed scenarios;
                                                      #   every service gets a logical schema in this instance
```

### F.5 Template

**`deployments/skeleton-template/`** IS the deployment template. The current
`deployments/template/` directory duplicates it and must be reconciled.

**FIX**: Template secrets currently use underscores (`hash_pepper_v3.secret`). Rename
to hyphens (`hash-pepper-v3.secret`) to match all other tiers.

### F.6 What Gets DELETED from deployments/

| Current Location | Reason |
|-----------------|--------|
| `deployments/archived/` | Dead code (Decision 3) |
| `deployments/template/` | Duplicate of `deployments/skeleton-template/` — reconcile and remove |

---

## G. internal/ — Private Application Code `drwxr-xr-x`

### G.1 internal/apps/ — Application Layer

```
internal/apps/                                        # drwxr-xr-x
├── {SUITE}/                                          # Suite orchestration
│   ├── {SUITE}.go                                    #   Suite CLI dispatch (seam pattern)
│   │                                                 #     Called by cmd/{SUITE}/main.go
│   │                                                 #     Delegates to products
│   ├── *_test.go                                     #   Unit tests
│   └── e2e/                                          #   E2E tests (orchestrates docker compose of full suite)
│
├── {PRODUCT}/                                        # Product level (×5)
│   ├── {PRODUCT}.go                                  #   Product CLI dispatch (seam pattern)
│   │                                                 #     Called by cmd/{PRODUCT}/main.go
│   │                                                 #     Delegates to services
│   ├── *_test.go                                     #   Unit tests
│   ├── e2e/                                          #   E2E tests (orchestrates docker compose of full product)
│   └── shared/                                       #   Shared within product (optional)
│       └── (shared packages)/                        #   e.g. identity/shared/domain/, identity/shared/config/
│           ├── *.go
│           └── *_test.go
│
├── {PRODUCT}/{SERVICE}/                              # Service implementation (×N per product, 10 total)
│   ├── {SERVICE}.go                                  #   Service entry point (seam pattern)
│   │                                                 #     Called by cmd/{PS-ID}/main.go
│   │                                                 #     Delegates to framework
│   ├── *_test.go                                     #   Unit tests
│   ├── integration/                                  #   Integration tests
│   ├── e2e/                                          #   E2E tests (orchestrates docker compose of service)
│   ├── repository/                                   #   Data access layer
│   │   ├── *.go                                      #     GORM entity models + repository methods
│   │   │                                             #     (models live alongside their data access code)
│   │   ├── *_test.go                                 #     Unit tests
│   │   └── migrations/                               #     Domain migrations (2001+)
│   │       ├── 2001_init.up.sql
│   │       └── 2001_init.down.sql
│   ├── model/                                        #   Domain models (optional)
│   │   └── *.go                                      #     Internal domain value objects and aggregates
│   │                                                 #     NOT API models (those are in api/{PS-ID}/models/)
│   │                                                 #     NOT GORM models (those are in repository/)
│   └── handler/                                      #   HTTP handlers (optional)
│       └── *.go                                      #     Domain-specific handlers beyond generated
│                                                     #     strict server (api/{PS-ID}/server/)
│
├── framework/                                        # Service framework (shared by ALL services)
│   ├── suite/                                        #   Suite-level framework (orchestration, routing)
│   ├── product/                                      #   Product-level framework (product CLI, aggregation)
│   └── service/                                      #   Service-level framework
│       ├── cli/                                      #     CLI infrastructure (cobra commands)
│       ├── client/                                   #     HTTP client helpers
│       ├── config/                                   #     Config loading and validation
│       │   └── tls_generator/                        #     Auto TLS certificate generation
│       ├── server/                                   #     Server infrastructure
│       │   ├── apis/                                 #       API route registration
│       │   ├── application/                          #       Application lifecycle
│       │   ├── barrier/                              #       Encryption at rest (Unseal → Root → Intermediate → Content)
│       │   │   └── unsealkeysservice/                #       Unseal key management
│       │   ├── builder/                              #       Server builder pattern (constructor injection)
│       │   ├── businesslogic/                        #       Business logic layer
│       │   ├── domain/                               #       Domain types (realm, tenant, session)
│       │   ├── listener/                             #       Dual HTTPS listeners (public + admin)
│       │   ├── middleware/                            #       HTTP middleware (CORS, CSRF, rate limiting, auth)
│       │   ├── realm/                                #       Authentication, authorization, and identity realm
│       │   ├── realms/                               #       AuthN, AuthZ, and identity realm registry
│       │   ├── repository/                           #       Database repository
│       │   │   ├── migrations/                       #       Framework migrations (1001-1999)
│       │   │   └── test_migrations/                  #       Test fixture migrations
│       │   ├── service/                              #       Service layer
│       │   └── tenant/                               #       Multi-tenancy
│       ├── server_integration/                       #     Integration test suite
│       ├── testing/                                  #     Shared test infrastructure
│       │   ├── assertions/                           #       Response validation helpers
│       │   ├── contract/                             #       Cross-service contract tests
│       │   ├── e2e_helpers/                          #       E2E test helpers
│       │   ├── e2e_infra/                            #       E2E infrastructure setup
│       │   ├── fixtures/                             #       Test data fixtures (tenants, realms, users)
│       │   ├── healthclient/                         #       Health endpoint test client
│       │   ├── httpservertests/                      #       HTTP server test helpers
│       │   ├── testdb/                               #       Test database helpers (SQLite in-memory, Postgres container)
│       │   └── testserver/                           #       Test server start/wait helpers
│       └── testutil/                                 #     Framework test utilities
│
├── tools/                                            # Infrastructure tooling
│   ├── cicd/                                         #   Custom linting and formatting tools
│   │   ├── common/                                   #     Shared CICD utilities
│   │   ├── format_go/                                #     Go code formatter
│   │   ├── format_gotest/                            #     Go test formatter
│   │   ├── lint_compose/                             #     Docker Compose linter
│   │   ├── lint_deployments/                         #     Deployment structure validator
│   │   │   └── (8 validators)                        #       Naming, schema, ports, secrets, etc.
│   │   ├── lint_docs/                                #     Documentation linter (includes docs_validation/)
│   │   ├── lint_fitness/                             #     Architecture fitness functions
│   │   │   ├── registry/                             #       Entity registry (SSOT for products/services)
│   │   │   │   ├── registry.go                       #         AllProducts(), AllProductServices(), AllSuites()
│   │   │   │   └── registry_test.go
│   │   │   ├── banned_product_names/                 #       Legacy name detection
│   │   │   ├── circular_deps/                        #       Circular dependency detection
│   │   │   ├── entity_registry_completeness/         #       Registry vs filesystem drift
│   │   │   ├── file_size/                            #       File size limit enforcement
│   │   │   ├── parallel_tests/                       #       t.Parallel() enforcement
│   │   │   ├── test_patterns/                        #       Test pattern enforcement
│   │   │   └── ... (44+ linters)                     #       All fitness function linters
│   │   ├── lint_go/                                  #     Go code linter
│   │   ├── lint_golangci/                            #     golangci-lint config validator
│   │   ├── lint_gotest/                              #     Go test linter
│   │   ├── lint_go_mod/                              #     Go module linter
│   │   ├── lint_ports/                               #     Port assignment validator
│   │   ├── lint_text/                                #     UTF-8/text linter
│   │   └── lint_workflow/                            #     GitHub Actions workflow linter
│   │
│   └── workflow/                                     #   GitHub Actions workflow management
│       └── *.go                                      #     Workflow runner + cleanup (consolidate github_cleanup/)
│
└── (DELETE)
    ├── demo/                                         #   Dead code (Decision 2)
    └── pkiinit/                                      #   Dead code (Decision 2)
```

**Consolidation required**:

- `docs_validation/` → merge into `lint_docs/` (single documentation linter)
- `github_cleanup/` → merge into `tools/workflow/` (single workflow tool)

### G.2 internal/shared/ — Shared Libraries `drwxr-xr-x`

```
internal/shared/                                      # drwxr-xr-x
├── apperr/                                           # Application error types
├── container/                                        # Docker container utilities
├── crypto/                                           # Cryptographic libraries
│   ├── asn1/                                         #   ASN.1 encoding/decoding
│   ├── certificate/                                  #   X.509 certificate operations
│   ├── digests/                                      #   Cryptographic digest functions
│   ├── hash/                                         #   Versioned hash service (PBKDF2, HKDF)
│   ├── jose/                                         #   JOSE/JWK/JWS/JWE operations
│   ├── keygen/                                       #   Key generation (RSA, ECDSA, EdDSA)
│   ├── keygenpooltest/                               #   Key generation pool test helpers
│   ├── password/                                     #   Password generation
│   ├── pbkdf2/                                       #   PBKDF2 key derivation
│   └── tls/                                          #   TLS certificate generation
├── database/                                         # Database utilities (GORM helpers)
├── magic/                                            # Named constants (SSOT, excluded from coverage)
│   ├── magic_api.go                                  #   API path constants
│   ├── magic_cicd.go                                 #   CICD command constants
│   ├── magic_console.go                              #   Console output constants
│   ├── magic_crypto.go                               #   Cryptographic constants
│   ├── magic_database.go                             #   Database constants
│   ├── magic_docker.go                               #   Docker constants
│   ├── magic_framework.go                            #   Framework constants
│   ├── magic_{PRODUCT}.go                            #   Per-product constants (×5)
│   ├── magic_{PRODUCT}_{topic}.go                    #   Per-product topic files (identity has ~12)
│   ├── magic_misc.go                                 #   Miscellaneous constants
│   ├── magic_network.go                              #   Network constants
│   ├── magic_orchestration.go                        #   Orchestration constants
│   ├── magic_percent.go                              #   Percentage constants
│   ├── magic_security.go                             #   Security constants
│   ├── magic_session.go                              #   Session constants
│   ├── magic_telemetry.go                            #   Telemetry constants
│   ├── magic_testing.go                              #   Testing constants
│   ├── magic_unseal.go                               #   Unseal key constants
│   └── magic_workflows.go                            #   Workflow constants
├── pool/                                             # High-performance key generation pool
├── pwdgen/                                           # Password generation utilities
├── telemetry/                                        # OpenTelemetry setup and management
├── testutil/                                         # Shared test utility helpers
└── util/                                             # General utilities
    ├── cache/                                        #   In-memory cache
    ├── combinations/                                 #   Combinatorial helpers
    ├── datetime/                                     #   Date/time utilities
    ├── files/                                        #   File system utilities
    ├── network/                                      #   Network utilities
    ├── poll/                                         #   Polling/retry helpers
    ├── random/                                       #   Secure random generation
    ├── sysinfo/                                      #   System information
    └── thread/                                       #   Thread/goroutine utilities
```

---

## H. docs/ — Documentation `drwxr-xr-x`

```
docs/                                                 # drwxr-xr-x
├── ARCHITECTURE.md                                   # SSOT: Architecture reference (5080+ lines)
├── CONFIG-SCHEMA.md                                  # Config file schema reference
├── DEV-SETUP.md                                      # Developer setup guide
├── README.md                                         # Documentation index
└── framework-v5/                                     # Current plan (active)
    ├── plan.md
    ├── tasks.md
    ├── lessons.md
    └── target-structure.md                           # THIS FILE
```

**DELETE** (per Decisions 2-3):

- `docs/framework-v3/` — Historical plan (completed, dead code)
- `docs/framework-v4/` — Historical plan (completed, dead code)
- `docs/LESSONS/` — Cross-plan lessons archive (superseded by per-plan lessons.md)
- `docs/ARCHITECTURE-COMPOSE-MULTIDEPLOY.md` (after merge into ARCHITECTURE.md)
- `docs/ARCHITECTURE-INDEX.md` (superseded by ARCHITECTURE.md ToC)
- `docs/ARCHITECTURE-TODO.md` (superseded by plan tracking)
- `docs/COPILOT-MULTI-PROJECT.md` (stale reference doc)
- `docs/DEAD_CODE_REVIEW.md` (completed, no longer needed)
- `docs/VSCODE-CRASHES.md` (stale troubleshooting doc)
- `docs/demo-brainstorm/` (demos archived per Decision 2)
- `docs/framework-brainstorm/` (superseded by framework-v3+)
- `docs/gremlins/` (stale mutation testing notes)
- `docs/workflow-runtimes/` (stale workflow analysis)

---

## I. test/ — External Test Suites `drwxr-xr-x`

```
test/                                                 # drwxr-xr-x
└── load/                                             # Gatling load tests (Java 21 + Maven)
    │                                                 # Needs refactoring: cover all 10 service-level,
    │                                                 # all 5 product-level, and 1 suite-level load tests
    ├── .gitignore
    ├── .mvn/                                         #   Maven wrapper
    ├── mvnw                                          #   Maven wrapper (Unix, chmod 755)
    ├── mvnw.cmd                                      #   Maven wrapper (Windows)
    ├── pom.xml                                       #   Maven POM
    ├── README.md                                     #   Load test documentation
    ├── src/                                           #   Gatling test sources
    └── target/                                       #   Maven build output (gitignored)
```

---

## J. pkg/ — Public Library Code (Reserved) `drwxr-xr-x`

```
pkg/                                                  # Currently empty, reserved for future public APIs
```

---

## K. Other Directories

```
scripts/                                              # Currently empty (.gitkeep only)
                                                      # Part of Go project structure, keep empty
testdata/                                             # Test data files
└── adaptive-sim/
    └── sample-auth-logs.json                         #   Sample auth logs for adaptive simulation
workflow-reports/                                      # Ephemeral test output, never Git tracked (gitignored)
test-output/                                          # Ephemeral test output, never Git tracked (gitignored)
```

---

## L. Secret File Naming Convention

All tiers (service, product, suite) use **identical `{purpose}.secret` names** —
no tier prefix on active secret files. Tier prefixes appear ONLY on `.secret.never`
marker files.

| Secret Purpose | Filename | Value Pattern (Service) | Value Pattern (Product/Suite) |
|---------------|----------|------------------------|-------------------------------|
| Browser password | `browser-password.secret` | `{PS-ID}-browser-{base64-random}` | (not at product/suite) |
| Browser username | `browser-username.secret` | `{PS-ID}-browser-user` | (not at product/suite) |
| Hash pepper v3 | `hash-pepper-v3.secret` | `{base64-random-32-bytes}` | `dev-hash-pepper-v3` |
| PostgreSQL database | `postgres-database.secret` | `{PS_ID}` | `cryptoutil` |
| PostgreSQL password | `postgres-password.secret` | `{PS_ID}_pass` | `cryptoutil-dev-password` |
| PostgreSQL URL | `postgres-url.secret` | `postgres://{PS_ID}_user:{PS_ID}_pass@{PS-ID}-postgres:5432/{PS_ID}?sslmode=disable` | `postgres://cryptoutil:cryptoutil-dev-password@{PRODUCT}-postgres:5432/cryptoutil?sslmode=disable` |
| PostgreSQL username | `postgres-username.secret` | `{PS_ID}_user` | `cryptoutil` |
| Service password | `service-password.secret` | `{PS-ID}-service-{base64-random}` | (not at product/suite) |
| Service username | `service-username.secret` | `{PS-ID}-service-user` | (not at product/suite) |
| Unseal shard N | `unseal-{N}of5.secret` | `{SERVICE}-{hex-random-32-bytes}` | `dev-unseal-key-{N}-of-5` |

**.secret.never marker files** (KEEP as explicit reminders):

| Tier | Pattern | Content |
|------|---------|---------|
| Product | `{PRODUCT}-{purpose}.secret.never` | "MUST NOT be shared at {PRODUCT} level. Use service-specific secrets." |
| Suite | `{SUITE}-{purpose}.secret.never` | "MUST NOT be shared at {SUITE} level. Use service-specific secrets." |

---

## M. Fitness Linter Coverage (New/Enhanced)

These fitness sub-linters of `lint-fitness` enforce the target structure:

| Linter | Scope | Rule |
|--------|-------|------|
| `root-junk-detection` | `{ROOT}/` | No *.exe, *.py, coverage*,*.test.exe at root |
| `cmd-entry-whitelist` | `cmd/` | Only 18 allowed entries (1 suite + 5 products + 10 services + 2 infra tools) |
| `configs-structure` | `configs/` | Must follow `{SUITE}/`, `{PRODUCT}/`, `{PS-ID}/` hierarchy |
| `configs-no-deployment` | `configs/` | No deployment variants, docker overlays, or environment files |
| `secret-naming` | `deployments/*/secrets/` | All tiers use identical `{purpose}.secret` names, with enforced `.never` marker exceptions |
| `template-consistency` | `deployments/skeleton-template/` | Template uses hyphens in secret names, matching services |
| `archive-detection` | `**/*archived*/`, `**/*orphaned*/` | No archived/orphaned directories |
| `entity-registry-completeness` | (existing, enhanced) | Verify `configs/{PS-ID}/` existence |

---

## N. Change Summary (Current → Target)

| Area | Current State | Target State | Action |
|------|--------------|-------------|--------|
| Root files | ~80+ junk artifacts | Clean project config only | DELETE artifacts |
| `cmd/` | 18 entries + extras | 18 entries exactly | DELETE demo, identity-compose, identity-demo |
| `api/` | Missing components/paths for some | All 10 PS-IDs with full spec | CREATE missing |
| `configs/` | Mixed nesting + orphaned | Flat `{SUITE}/` + `{PRODUCT}/` + `{PS-ID}/` + shared E.4 | RESTRUCTURE |
| `deployments/` service | Missing sqlite-2 config | All 4 config overlays per service | CREATE missing |
| `deployments/` product | No Dockerfile | Dockerfile added | CREATE |
| `deployments/` .never | Some marked for delete | KEEP all as explicit reminders | KEEP |
| `deployments/` template | Separate `template/` dir | Use `skeleton-template/` only | RECONCILE |
| `deployments/` archived | Still present | Deleted | DELETE |
| `internal/apps/` suite | No explicit entry point | `{SUITE}.go` seam + e2e/ | CREATE |
| `internal/apps/` product | No explicit entry point | `{PRODUCT}.go` seam + shared/ + e2e/ | CREATE |
| `internal/apps/` service | Nested `{PRODUCT}/{SERVICE}/` | Keep nested, add integration/ + e2e/ | ADD dirs |
| `internal/apps/` cicd | Under `internal/apps/cicd/` | Under `internal/apps/tools/cicd/` | MOVE |
| `internal/apps/` workflow | Under `internal/apps/workflow/` | Under `internal/apps/tools/workflow/` | MOVE |
| `internal/apps/` docs\_validation | Separate from lint\_docs | Merged into lint\_docs/ | MERGE |
| `internal/apps/` github\_cleanup | Under cicd/ | Merged into tools/workflow/ | MERGE |
| `internal/apps/` framework | Only service/ | Add suite/ + product/ | CREATE |
| `internal/apps/` demo, pkiinit | Present | Deleted | DELETE |
| `docs/` | Historical plans + stale docs | Only active plan + core docs | DELETE stale |
| `test/load/` | Single basic test | Cover all 10+5+1 tiers | REFACTOR |
| Secret naming | Inconsistent across tiers | Uniform `{purpose}.secret` + `.never` markers | STANDARDIZE |

---

## O. Open Questions (Quizme v2 Candidates)

These questions require user decisions before implementation:

1. **tools/cicd/ naming**: The name `cicd/` is misleading since it only contains
   custom linting and formatting tools, not CI/CD pipeline management. Possible renames:
   `tools/lint/`, `tools/quality/`, `tools/checks/`. Keep as-is?

2. **tls\_generator/ restructuring**: Should `framework/service/config/tls_generator/`
   move to `framework/tls/` to support suite/product/service TLS needs? Should
   `internal/apps/pkiinit/` merge into it?

3. **apperr/ location**: Should `internal/shared/apperr/` move to
   `internal/apps/framework/` since it is part of the application error layer
   rather than a general shared utility?

4. **framework/suite/ and framework/product/ scope**: What concrete packages
   belong under these new directories? Suite orchestration CLI? Product aggregation
   CLI? Or are these empty scaffolds awaiting future framework extraction?

5. **testdata/ directory**: Keep or delete? Currently only contains one sample
   file (`adaptive-sim/sample-auth-logs.json`).
