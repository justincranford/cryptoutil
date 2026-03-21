# Target Repository Structure - Framework v5

**Status**: DRAFT — User review required before execution
**Created**: 2026-03-22
**Purpose**: Define the complete, parameterized target state of every directory and file in the repository. Once approved, fitness linters enforce this structure and automation moves current state toward it.

---

## Entity Hierarchy (Canonical)

| Level | Variable | Instances | Count |
|-------|----------|-----------|-------|
| Suite | `{SUITE}` | `cryptoutil` | 1 |
| Product | `{PRODUCT}` | `identity`, `jose`, `pki`, `skeleton`, `sm` | 5 |
| Service | `{SERVICE}` | varies per product (see below) | 10 total |
| PS-ID | `{PS-ID}` = `{PRODUCT}-{SERVICE}` | see table below | 10 |
| Infra Tool | N/A | `cicd`, `workflow` | 2 |
| Framework | N/A | `framework` | 1 |

### Product-Service Matrix

| PS-ID | Product | Service | Display Name |
|-------|---------|---------|-------------|
| `identity-authz` | identity | authz | Identity Authorization Server |
| `identity-idp` | identity | idp | Identity Provider |
| `identity-rp` | identity | rp | Identity Relying Party |
| `identity-rs` | identity | rs | Identity Resource Server |
| `identity-spa` | identity | spa | Identity Single Page App |
| `jose-ja` | jose | ja | JOSE JWK Authority |
| `pki-ca` | pki | ca | PKI Certificate Authority |
| `skeleton-template` | skeleton | template | Skeleton Template |
| `sm-im` | sm | im | Secrets Manager Instant Messenger |
| `sm-kms` | sm | kms | Secrets Manager Key Management |

---

## A. Root Level

### A.1 Root Files (KEEP — legitimate project config)

```
{ROOT}/
├── .air.toml                          # Air live-reload config
├── .dockerignore                      # Docker build context exclusions
├── .editorconfig                      # Editor formatting standards (indent, line endings)
├── .gitattributes                     # Git line ending and diff config
├── .gitignore                         # Git ignore rules
├── .gitleaks.toml                     # Gitleaks secret detection config
├── .gofumpt.toml                      # gofumpt Go formatting config
├── .golangci.yml                      # golangci-lint v2 linter config
├── .gremlins.yaml                     # Gremlins mutation testing config
├── .markdownlint.jsonc                # Markdown linting rules
├── .nuclei-ignore                     # Nuclei DAST scan exclusions
├── .pre-commit-config.yaml            # Pre-commit hook definitions
├── .rgignore                          # ripgrep ignore patterns
├── .sqlfluff                          # SQL linting config
├── go.mod                             # Go module definition
├── go.sum                             # Go module dependency checksums
├── LICENSE                            # Project license
├── pyproject.toml                     # Python project config (pre-commit tooling)
└── README.md                          # Project README
```

### A.2 Root Files (DELETE — junk artifacts, ~80+ files)

All `*.exe`, `*.py`, `coverage*`, `*_coverage`, `*.test.exe` files at root are build/test artifacts that should never be committed. Git history preserves them.

### A.3 Root Hidden Directories

```
{ROOT}/
├── .cicd/                             # CICD runtime caches (gitignored)
│   ├── circular-dep-cache.json        #   Circular dependency analysis cache
│   └── dep-cache.json                 #   Dependency analysis cache
├── .ruff_cache/                       # Ruff Python linter cache (gitignored)
├── .semgrep/                          # Semgrep SAST rules
│   └── rules/
│       └── go-testing.yml             #   Go testing SAST rules
├── .vscode/                           # VS Code workspace settings
│   ├── cspell.json                    #   Spell checking dictionary
│   ├── extensions.json                #   Recommended extensions
│   ├── launch.json                    #   Debug launch configs
│   └── settings.json                  #   Workspace settings
└── .zap/                              # OWASP ZAP DAST config
    └── rules.tsv                      #   ZAP scan rules
```

---

## B. .github/ — GitHub & Copilot Configuration

```
.github/
├── copilot-instructions.md            # Copilot config hub (loads instructions/)
├── agents/                            # Copilot chat agents
│   ├── beast-mode.agent.md            #   Continuous autonomous execution
│   ├── doc-sync.agent.md              #   Documentation sync agent
│   ├── fix-workflows.agent.md         #   CI/CD workflow fixer
│   ├── implementation-execution.agent.md  # Plan execution agent
│   └── implementation-planning.agent.md   # Plan creation agent
├── actions/                           # Reusable GitHub Actions
│   ├── custom-cicd-lint/action.yml    #   Custom CICD lint composite action
│   ├── docker-compose-build/action.yml
│   ├── docker-compose-down/action.yml
│   ├── docker-compose-logs/action.yml
│   ├── docker-compose-up/action.yml
│   ├── docker-compose-verify/action.yml
│   ├── docker-images-pull/action.yml  #   Parallel Docker image pre-pull
│   ├── fuzz-test/action.yml
│   ├── go-setup/action.yml            #   Go toolchain setup with caching
│   ├── golangci-lint/action.yml       #   golangci-lint v2 execution
│   ├── security-scan-gitleaks/action.yml
│   ├── security-scan-trivy/action.yml
│   ├── security-scan-trivy2/action.yml
│   ├── workflow-job-begin/action.yml  #   Job telemetry start
│   └── workflow-job-end/action.yml    #   Job telemetry end
├── instructions/                      # Copilot instruction files (auto-loaded alpha order)
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
├── skills/                            # Copilot skills (slash commands)
│   ├── README.md                      #   Skill catalog
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
└── workflows/                         # GitHub Actions CI/CD workflows
    ├── ci-benchmark.yml               #   Benchmark testing
    ├── ci-coverage.yml                #   Code coverage analysis
    ├── ci-dast.yml                    #   Dynamic application security testing
    ├── ci-e2e.yml                     #   End-to-end testing
    ├── ci-fitness.yml                 #   Architecture fitness functions
    ├── ci-fuzz.yml                    #   Fuzz testing
    ├── ci-gitleaks.yml                #   Secret detection
    ├── ci-identity-validation.yml     #   Identity service validation
    ├── ci-load.yml                    #   Load testing (Gatling)
    ├── ci-mutation.yml                #   Mutation testing (gremlins)
    ├── ci-quality.yml                 #   Build + lint + unit tests
    ├── ci-race.yml                    #   Race condition detection
    ├── ci-sast.yml                    #   Static application security testing
    ├── cicd-lint-deployments.yml      #   Deployment structure validation
    └── release.yml                    #   Release workflow
```

---

## C. cmd/ — Binary Entry Points

**Pattern**: Each entry has exactly one `main.go` that delegates to `internal/apps/`.

```
cmd/
├── {SUITE}/main.go                    # Suite CLI → internal/apps/{SUITE}/
│                                      #   e.g. cmd/cryptoutil/main.go
│
├── {PRODUCT}/main.go                  # Product CLI → internal/apps/{PRODUCT}/  (×5)
│                                      #   cmd/identity/main.go
│                                      #   cmd/jose/main.go
│                                      #   cmd/pki/main.go
│                                      #   cmd/skeleton/main.go
│                                      #   cmd/sm/main.go
│
├── {PS-ID}/main.go                    # Service CLI → internal/apps/{PRODUCT}/{SERVICE}/  (×10)
│                                      #   cmd/identity-authz/main.go
│                                      #   cmd/identity-idp/main.go
│                                      #   cmd/identity-rp/main.go
│                                      #   cmd/identity-rs/main.go
│                                      #   cmd/identity-spa/main.go
│                                      #   cmd/jose-ja/main.go
│                                      #   cmd/pki-ca/main.go
│                                      #   cmd/skeleton-template/main.go
│                                      #   cmd/sm-im/main.go
│                                      #   cmd/sm-kms/main.go
│
├── cicd/main.go                       # CICD tool → internal/apps/cicd/
└── workflow/main.go                   # Workflow runner → internal/apps/workflow/
```

**Allowed entries**: 1 suite + 5 products + 10 services + 2 infra = **18 total**

**DELETE** (per Decision 2): `cmd/demo/`, `cmd/identity-compose/`, `cmd/identity-demo/`

---

## D. api/ — OpenAPI Specifications & Generated Code

**Pattern**: One directory per product-service, containing the OpenAPI spec and oapi-codegen output.

```
api/
└── {PS-ID}/                           # (×10)
    ├── generate.go                    # //go:generate oapi-codegen directives
    ├── openapi_spec.yaml              # OpenAPI 3.0.3 specification (SSOT for API)
    ├── openapi_spec_components.yaml   # Reusable schema components (optional)
    ├── openapi_spec_paths.yaml        # API path definitions (optional)
    ├── openapi-gen_config_client.yaml # oapi-codegen client generation config
    ├── openapi-gen_config_models.yaml # oapi-codegen models generation config
    ├── openapi-gen_config_server.yaml # oapi-codegen server generation config
    ├── client/
    │   └── client.gen.go              # Generated HTTP client
    ├── models/
    │   └── models.gen.go              # Generated request/response models
    └── server/
        └── server.gen.go              # Generated strict server interface
```

---

## E. configs/ — Canonical Application Configuration (SSOT)

**Principle**: `configs/` is the **single source of truth** for what the app needs —
environment-agnostic, reusable configuration. Deployment-specific overlays live in `deployments/`.

### E.1 Suite Level

```
configs/
└── {SUITE}/                           # Suite-level config
    └── {SUITE}.yml                    # Suite orchestration config
                                       #   e.g. configs/cryptoutil/cryptoutil.yml
```

### E.2 Product-Service Level (Parameterized ×10)

```
configs/
└── {PRODUCT}/                         # Product directory
    └── {SERVICE}/                     # Service directory
        └── {SERVICE}.yml              # Canonical service domain config
                                       #   e.g. configs/sm/kms/kms.yml
                                       #   e.g. configs/jose/ja/ja.yml
                                       #   e.g. configs/pki/ca/ca.yml
```

### E.3 Product-Level Shared Config (where applicable)

Some products have shared configuration that applies across all services within
that product. These live at the product level, NOT duplicated per service.

```
configs/
├── identity/                          # Identity product
│   ├── authz/
│   │   └── authz.yml                 # AuthZ service domain config
│   ├── idp/
│   │   └── idp.yml                   # IdP service domain config
│   ├── rp/
│   │   └── rp.yml                    # RP service domain config
│   ├── rs/
│   │   └── rs.yml                    # RS service domain config
│   ├── spa/
│   │   └── spa.yml                   # SPA service domain config
│   └── policies/                      # Auth policies (shared across identity services)
│       ├── adaptive-auth.yml          #   Adaptive authentication rules
│       ├── risk-scoring.yml           #   Risk scoring parameters
│       └── step-up.yml               #   Step-up authentication thresholds
│
├── pki/
│   └── ca/
│       ├── ca.yml                     # CA service domain config (rename from ca-server.yml)
│       ├── ca-config-schema.yaml      # CA config schema definition
│       └── profiles/                  # Certificate profiles (domain config)
│           ├── code-signing.yaml      #   Code signing cert profile
│           ├── database-server.yaml   #   Database server cert profile
│           ├── device.yaml            #   Device cert profile
│           └── ... (23 profile files) #   All other cert profiles
│
├── jose/
│   └── ja/
│       └── ja.yml                     # JA service domain config
│
├── skeleton/
│   └── template/
│       └── template.yml               # Template service domain config
│
└── sm/
    ├── im/
    │   └── im.yml                     # IM service domain config
    └── kms/
        └── kms.yml                    # KMS service domain config
```

### E.4 What MOVES OUT of configs/ (to deployments/)

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

### E.5 What Gets DELETED from configs/

| Current Location | Reason |
|-----------------|--------|
| `configs/orphaned/` | Dead code (Decision 3) |
| `configs/ca/` (contents) | Moved to `configs/pki/ca/` |
| `configs/configs-all-files.json` | Metadata artifact, not config |
| `configs/jose/jose-server.yml` | Replaced by `configs/jose/ja/ja.yml` |
| `configs/skeleton/skeleton-server.yml` | Replaced by `configs/skeleton/template/template.yml` |

---

## F. deployments/ — Deployment Manifests & Wiring

**Principle**: `deployments/` contains environment-specific deployment manifests that
CONSUME configuration from `configs/`. Config files here are deployment overlays, NOT
the canonical source.

### F.1 Service-Level Deployments (×10)

```
deployments/
└── {PS-ID}/                           # One per product-service
    ├── compose.yml                    # Docker Compose manifest
    ├── Dockerfile                     # Multi-stage build (builder → validator → runtime)
    ├── config/                        # Deployment-specific config overlays
    │   ├── {PS-ID}-app-common.yml     #   Shared across all instances (bind addresses,
    │   │                              #     TLS paths, Docker network hostnames)
    │   ├── {PS-ID}-app-sqlite-1.yml   #   SQLite instance config (in-memory DB)
    │   ├── {PS-ID}-app-postgresql-1.yml  # PostgreSQL instance 1 config
    │   └── {PS-ID}-app-postgresql-2.yml  # PostgreSQL instance 2 config
    └── secrets/                       # Docker secrets (chmod 440)
        ├── browser-password.secret    #   Browser client password
        ├── browser-username.secret    #   Browser client username
        ├── hash-pepper-v3.secret      #   PBKDF2 hash pepper (version 3)
        ├── postgres-database.secret   #   PostgreSQL database name
        ├── postgres-password.secret   #   PostgreSQL password
        ├── postgres-url.secret        #   PostgreSQL connection URL
        ├── postgres-username.secret   #   PostgreSQL username
        ├── service-password.secret    #   Service client password
        ├── service-username.secret    #   Service client username
        ├── unseal-1of5.secret         #   Unseal key shard 1 of 5
        ├── unseal-2of5.secret         #   Unseal key shard 2 of 5
        ├── unseal-3of5.secret         #   Unseal key shard 3 of 5
        ├── unseal-4of5.secret         #   Unseal key shard 4 of 5
        └── unseal-5of5.secret         #   Unseal key shard 5 of 5
```

### F.2 Product-Level Deployments (×5)

```
deployments/
└── {PRODUCT}/                         # One per product
    ├── compose.yml                    # Product-level compose (delegates to services)
    └── secrets/                       # Product-level secrets (shared across services)
        ├── browser-password.secret
        ├── browser-username.secret
        ├── hash-pepper-v3.secret
        ├── postgres-database.secret
        ├── postgres-password.secret
        ├── postgres-url.secret
        ├── postgres-username.secret
        ├── service-password.secret
        ├── service-username.secret
        ├── unseal-1of5.secret
        ├── unseal-2of5.secret
        ├── unseal-3of5.secret
        ├── unseal-4of5.secret
        └── unseal-5of5.secret
```

**DELETE from product secrets**: All `{PRODUCT}-{purpose}.secret.never` legacy files.

### F.3 Suite-Level Deployment (×1)

```
deployments/
└── {SUITE}-suite/                     # e.g. cryptoutil-suite/
    ├── compose.yml                    # Suite-level compose (all 10 services)
    ├── Dockerfile                     # Suite-level build
    └── secrets/                       # Suite-level secrets
        ├── browser-password.secret
        ├── browser-username.secret
        ├── hash-pepper-v3.secret
        ├── postgres-database.secret
        ├── postgres-password.secret
        ├── postgres-url.secret
        ├── postgres-username.secret
        ├── service-password.secret
        ├── service-username.secret
        ├── unseal-1of5.secret
        ├── unseal-2of5.secret
        ├── unseal-3of5.secret
        ├── unseal-4of5.secret
        └── unseal-5of5.secret
```

**DELETE from suite secrets**: All `{SUITE}-{purpose}.secret.never` legacy files.

### F.4 Shared Infrastructure

```
deployments/
├── shared-telemetry/                  # OpenTelemetry + Grafana LGTM
│   └── compose.yml                    #   otel-collector-contrib + grafana-otel-lgtm
├── shared-postgres/                   # Shared PostgreSQL container
│   └── compose.yml                    #   PostgreSQL for multi-service sharing
└── shared-citus/                      # Shared Citus (distributed PostgreSQL)
    └── compose.yml                    #   Citus for distributed scenarios
```

### F.5 Template

```
deployments/
└── template/                          # Deployment template (for new services)
    ├── compose.yml                    # Template compose manifest
    ├── config/
    │   ├── template-app-common.yml
    │   ├── template-app-sqlite-1.yml
    │   ├── template-app-postgresql-1.yml
    │   └── template-app-postgresql-2.yml
    └── secrets/                       # Template secrets (hyphenated, matching services)
        ├── browser-password.secret
        ├── browser-username.secret
        ├── hash-pepper-v3.secret
        ├── postgres-database.secret
        ├── postgres-password.secret
        ├── postgres-url.secret
        ├── postgres-username.secret
        ├── service-password.secret
        ├── service-username.secret
        ├── unseal-1of5.secret
        ├── unseal-2of5.secret
        ├── unseal-3of5.secret
        ├── unseal-4of5.secret
        └── unseal-5of5.secret
```

**FIX**: Template secrets currently use underscores (`hash_pepper_v3.secret`). Rename to hyphens (`hash-pepper-v3.secret`) to match all other tiers.

### F.6 What Gets DELETED from deployments/

| Current Location | Reason |
|-----------------|--------|
| `deployments/archived/` | Dead code (Decision 3) |
| `deployments/{PRODUCT}/secrets/{PRODUCT}-*.secret.never` | Legacy placeholder files |
| `deployments/{SUITE}-suite/secrets/{SUITE}-*.secret.never` | Legacy placeholder files |

---

## G. internal/ — Private Application Code

### G.1 internal/apps/ — Application Layer

```
internal/apps/
├── {SUITE}/                           # Suite orchestration
│   └── *.go                           #   Suite CLI dispatch, delegates to products
│
├── {PRODUCT}/                         # Product level (×5)
│   ├── {SERVICE}/                     # Service implementation (×N per product)
│   │   ├── *.go                       #   Service business logic
│   │   ├── *_test.go                  #   Unit tests
│   │   ├── repository/               #   Data access layer (optional)
│   │   │   ├── *.go
│   │   │   ├── *_test.go
│   │   │   └── migrations/           #   Domain migrations (2001+)
│   │   │       ├── 2001_init.up.sql
│   │   │       └── 2001_init.down.sql
│   │   ├── model/                     #   Domain models (optional)
│   │   │   └── *.go
│   │   └── handler/                   #   HTTP handlers (optional)
│   │       └── *.go
│   └── (shared packages)/            #   Shared within product (optional)
│       └── *.go                       #   e.g. identity/domain/, identity/config/
│
├── cicd/                              # CICD tooling (infrastructure)
│   ├── common/                        #   Shared CICD utilities
│   ├── docs_validation/               #   Documentation validation
│   ├── format_go/                     #   Go code formatter
│   ├── format_gotest/                 #   Go test formatter
│   ├── github_cleanup/                #   GitHub Actions cleanup
│   ├── lint_compose/                  #   Docker Compose linter
│   ├── lint_deployments/              #   Deployment structure validator
│   │   └── (8 validators)            #   Naming, schema, ports, secrets, etc.
│   ├── lint_docs/                     #   Documentation linter
│   ├── lint_fitness/                  #   Architecture fitness functions
│   │   ├── registry/                  #   Entity registry (SSOT for products/services)
│   │   │   ├── registry.go            #   AllProducts(), AllProductServices(), AllSuites()
│   │   │   └── registry_test.go
│   │   ├── banned_product_names/      #   Legacy name detection
│   │   ├── circular_deps/             #   Circular dependency detection
│   │   ├── entity_registry_completeness/ # Registry vs filesystem drift
│   │   ├── file_size/                 #   File size limit enforcement
│   │   ├── parallel_tests/            #   t.Parallel() enforcement
│   │   ├── test_patterns/             #   Test pattern enforcement
│   │   └── ... (44+ linters)         #   All fitness function linters
│   ├── lint_go/                       #   Go code linter
│   ├── lint_golangci/                 #   golangci-lint config validator
│   ├── lint_gotest/                   #   Go test linter
│   ├── lint_go_mod/                   #   Go module linter
│   ├── lint_ports/                    #   Port assignment validator
│   ├── lint_text/                     #   UTF-8/text linter
│   └── lint_workflow/                 #   GitHub Actions workflow linter
│
├── framework/                         # Service framework (shared by ALL services)
│   └── service/
│       ├── cli/                       #   CLI infrastructure (cobra commands)
│       ├── client/                    #   HTTP client helpers
│       ├── config/                    #   Config loading and validation
│       │   └── tls_generator/         #   Auto TLS certificate generation
│       ├── server/                    #   Server infrastructure
│       │   ├── apis/                  #   API route registration
│       │   ├── application/           #   Application lifecycle
│       │   ├── barrier/               #   Encryption at rest (Unseal → Root → Intermediate → Content)
│       │   │   └── unsealkeysservice/ #   Unseal key management
│       │   ├── builder/               #   Server builder pattern (constructor injection)
│       │   ├── businesslogic/         #   Business logic layer
│       │   ├── domain/                #   Domain types (realm, tenant, session)
│       │   ├── listener/              #   Dual HTTPS listeners (public + admin)
│       │   ├── middleware/            #   HTTP middleware (CORS, CSRF, rate limiting, auth)
│       │   ├── realm/                 #   Authentication realm
│       │   ├── realms/                #   AuthN realm registry
│       │   ├── repository/            #   Database repository
│       │   │   ├── migrations/        #   Framework migrations (1001-1999)
│       │   │   └── test_migrations/   #   Test fixture migrations
│       │   ├── service/               #   Service layer
│       │   └── tenant/                #   Multi-tenancy
│       ├── server_integration/        #   Integration test suite
│       ├── testing/                   #   Shared test infrastructure
│       │   ├── assertions/            #   Response validation helpers
│       │   ├── contract/              #   Cross-service contract tests
│       │   ├── e2e_helpers/           #   E2E test helpers
│       │   ├── e2e_infra/             #   E2E infrastructure setup
│       │   ├── fixtures/              #   Test data fixtures (tenants, realms, users)
│       │   ├── healthclient/          #   Health endpoint test client
│       │   ├── httpservertests/        #   HTTP server test helpers
│       │   ├── testdb/                #   Test database helpers (SQLite in-memory, Postgres container)
│       │   └── testserver/            #   Test server start/wait helpers
│       └── testutil/                  #   Framework test utilities
│
└── workflow/                          # Workflow runner (infrastructure)
    └── *.go
```

**DELETE** (per Decision 2): `internal/apps/demo/`, `internal/apps/pkiinit/`

### G.2 internal/shared/ — Shared Libraries

```
internal/shared/
├── apperr/                            # Application error types
├── container/                         # Docker container utilities
├── crypto/                            # Cryptographic libraries
│   ├── asn1/                          #   ASN.1 encoding/decoding
│   ├── certificate/                   #   X.509 certificate operations
│   ├── digests/                       #   Cryptographic digest functions
│   ├── hash/                          #   Versioned hash service (PBKDF2, HKDF)
│   ├── jose/                          #   JOSE/JWK/JWS/JWE operations
│   ├── keygen/                        #   Key generation (RSA, ECDSA, EdDSA)
│   ├── keygenpooltest/                #   Key generation pool test helpers
│   ├── password/                      #   Password generation
│   ├── pbkdf2/                        #   PBKDF2 key derivation
│   └── tls/                           #   TLS certificate generation
├── database/                          # Database utilities (GORM helpers)
├── magic/                             # Named constants (SSOT, excluded from coverage)
│   ├── magic_api.go                   #   API path constants
│   ├── magic_cicd.go                  #   CICD command constants
│   ├── magic_console.go               #   Console output constants
│   ├── magic_crypto.go                #   Cryptographic constants
│   ├── magic_database.go              #   Database constants
│   ├── magic_docker.go                #   Docker constants
│   ├── magic_framework.go             #   Framework constants
│   ├── magic_{PRODUCT}.go             #   Per-product constants (×5)
│   ├── magic_{PRODUCT}_{topic}.go     #   Per-product topic files (identity has ~12)
│   ├── magic_misc.go                  #   Miscellaneous constants
│   ├── magic_network.go               #   Network constants
│   ├── magic_orchestration.go         #   Orchestration constants
│   ├── magic_percent.go               #   Percentage constants
│   ├── magic_security.go              #   Security constants
│   ├── magic_session.go               #   Session constants
│   ├── magic_telemetry.go             #   Telemetry constants
│   ├── magic_testing.go               #   Testing constants
│   ├── magic_unseal.go                #   Unseal key constants
│   └── magic_workflows.go             #   Workflow constants
├── pool/                              # High-performance key generation pool
├── pwdgen/                            # Password generation utilities
├── telemetry/                         # OpenTelemetry setup and management
├── testutil/                          # Shared test utility helpers
└── util/                              # General utilities
    ├── cache/                         #   In-memory cache
    ├── combinations/                  #   Combinatorial helpers
    ├── datetime/                      #   Date/time utilities
    ├── files/                         #   File system utilities
    ├── network/                       #   Network utilities
    ├── poll/                          #   Polling/retry helpers
    ├── random/                        #   Secure random generation
    ├── sysinfo/                       #   System information
    └── thread/                        #   Thread/goroutine utilities
```

---

## H. docs/ — Documentation

```
docs/
├── ARCHITECTURE.md                    # SSOT: Architecture reference (5080+ lines)
├── CONFIG-SCHEMA.md                   # Config file schema reference
├── DEV-SETUP.md                       # Developer setup guide
├── README.md                          # Documentation index
├── framework-v5/                      # Current plan (active)
│   ├── plan.md
│   ├── tasks.md
│   ├── lessons.md
│   └── target-structure.md            # THIS FILE
├── framework-v3/                      # Historical plan (completed)
├── framework-v4/                      # Historical plan (completed)
└── LESSONS/                           # Cross-plan lessons archive
```

**DELETE** (per Decisions 2-3):
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

## I. test/ — External Test Suites

```
test/
└── load/                              # Gatling load tests (Java 21 + Maven)
    ├── .gitignore
    ├── .mvn/                          #   Maven wrapper
    ├── mvnw                           #   Maven wrapper (Unix)
    ├── mvnw.cmd                       #   Maven wrapper (Windows)
    ├── pom.xml                        #   Maven POM
    ├── README.md                      #   Load test documentation
    ├── src/                           #   Gatling test sources
    └── target/                        #   Maven build output (gitignored)
```

---

## J. pkg/ — Public Library Code (Reserved)

```
pkg/                                   # Currently empty, reserved for future public APIs
```

---

## K. Other Directories

```
scripts/                               # Currently empty (.gitkeep only)
testdata/                              # Test data files
└── adaptive-sim/
    └── sample-auth-logs.json          #   Sample auth logs for adaptive simulation
workflow-reports/                       # Workflow analysis reports (gitignored or delete)
test-output/                           # Test evidence (gitignored)
```

**DELETE**: `workflow-reports/` (stale analysis docs, superseded by plan tracking)
**DELETE**: `scripts/` (empty, no shell scripts per coding standards — Go/Python only)

---

## L. Secret File Naming Convention

### Canonical Secret Names (all tiers)

| Secret File | Purpose | Content |
|------------|---------|---------|
| `browser-password.secret` | Browser client password | Dev default value |
| `browser-username.secret` | Browser client username | Dev default value |
| `hash-pepper-v3.secret` | PBKDF2 hash pepper v3 | 32+ byte random |
| `postgres-database.secret` | PostgreSQL database name | e.g. `cryptoutil` |
| `postgres-password.secret` | PostgreSQL password | Dev default value |
| `postgres-url.secret` | PostgreSQL connection URL | Full DSN |
| `postgres-username.secret` | PostgreSQL username | Dev default value |
| `service-password.secret` | Service client password | Dev default value |
| `service-username.secret` | Service client username | Dev default value |
| `unseal-1of5.secret` | Unseal key shard 1 of 5 | HKDF-derived (NEVER modify) |
| `unseal-2of5.secret` | Unseal key shard 2 of 5 | HKDF-derived |
| `unseal-3of5.secret` | Unseal key shard 3 of 5 | HKDF-derived |
| `unseal-4of5.secret` | Unseal key shard 4 of 5 | HKDF-derived |
| `unseal-5of5.secret` | Unseal key shard 5 of 5 | HKDF-derived |

**Rules**:
- **Naming**: `{purpose}.secret` — kebab-case, NO product/suite prefix, NO `.never` suffix
- **Identical names at all tiers**: SERVICE, PRODUCT, SUITE, template all use the same file names
- **Permissions**: `chmod 440` (r--r-----)
- **Never modify unseal keys**: Breaks HKDF deterministic derivation

---

## M. Fitness Linter Coverage (New/Enhanced)

These fitness linters enforce the target structure:

| Linter | Scope | Rule |
|--------|-------|------|
| `root-junk-detection` | `{ROOT}/` | No *.exe, *.py, coverage*,*.test.exe at root |
| `cmd-entry-whitelist` | `cmd/` | Only 18 allowed entries (1 suite + 5 products + 10 services + 2 infra) |
| `configs-structure` | `configs/` | Must follow `{PRODUCT}/{SERVICE}/` hierarchy |
| `configs-no-deployment` | `configs/` | No deployment variants, docker overlays, or environment files |
| `secret-naming` | `deployments/*/secrets/` | All tiers use identical `{purpose}.secret` names |
| `secret-no-legacy` | `deployments/*/secrets/` | No `.never` suffix files |
| `template-consistency` | `deployments/template/` | Template uses hyphens in secret names, matching services |
| `archive-detection` | `**/*archived*/`, `**/*orphaned*/` | No archived/orphaned directories |
| `entity-registry-completeness` | (existing, enhanced) | Verify configs/{PRODUCT}/{SERVICE}/ existence |

---

## N. Summary of Changes from Current State

### Files/Directories to DELETE (~250+ items)

| Category | Count | Examples |
|----------|-------|---------|
| Root junk files | ~80 | *.exe, *.py, coverage_* |
| Archived directories | 9 dirs, ~161 files | internal/apps/identity/_archived/, deployments/archived/ |
| Demo entries | 3 cmd + 1 internal | cmd/demo/, cmd/identity-compose/, cmd/identity-demo/, internal/apps/demo/ |
| Legacy secrets | ~25 | {PRODUCT}-*.secret.never, cryptoutil-*.secret.never |
| Stale docs | ~10 files + 4 dirs | ARCHITECTURE-COMPOSE-MULTIDEPLOY.md, demo-brainstorm/, etc. |
| Misc | ~5 | configs/configs-all-files.json, configs/orphaned/, scripts/, internal/apps/pkiinit/ |

### Files/Directories to MOVE

| From | To | Reason |
|------|-----|--------|
| `configs/ca/*` | `configs/pki/ca/` | Product name is `pki`, not `ca` |
| `configs/identity/*-docker.yml` | `deployments/identity-*/config/` | Docker-specific overlay |
| `configs/identity/development.yml` etc. | `deployments/identity/config/` or delete | Environment-specific |
| `configs/identity/profiles/` | `deployments/identity/config/` | Compose deployment profiles |
| `configs/sm/im/config-*.yml` | Delete (already in deployments/) | Deployment variants |
| `configs/sm/kms/config-*.yml` | Delete (already in deployments/) | Deployment variants |

### Files to RENAME

| From | To | Reason |
|------|-----|--------|
| `configs/pki/ca/ca-server.yml` | `configs/pki/ca/ca.yml` | Standardize naming |
| `deployments/template/secrets/hash_pepper_v3.secret` | `hash-pepper-v3.secret` | Underscore → hyphen |
| `deployments/template/secrets/postgres_*.secret` | `postgres-*.secret` | Underscore → hyphen |
| `deployments/template/secrets/unseal_*of5.secret` | `unseal-*of5.secret` | Underscore → hyphen |

### Content to MERGE

| From | Into | Section |
|------|------|---------|
| `ARCHITECTURE-COMPOSE-MULTIDEPLOY.md` | `ARCHITECTURE.md` | Section 12.3 |
