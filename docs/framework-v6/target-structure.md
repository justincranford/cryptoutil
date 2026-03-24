# Target Repository Structure - Framework v6

**Status**: CANONICAL TARGET — guides framework-v6 implementation
**Created**: 2026-03-24
**Purpose**: Define the complete, exact target state of every directory and file
in the repository. After framework-v6 implementation is complete, everything
listed here exists; everything **not** listed here is **deleted**.

> **Reading this document**: Directory trees use `*.go` / `*_test.go` wildcards
> for Go source files within established packages where individual file names
> are not structurally significant. All config files, secret files, documentation
> files, and deployment manifests are enumerated individually because their exact
> names are load-bearing for linters, tooling, and deployment.

---

## Framework v5 Mistakes Resolved in v6

| # | v5 Mistake | v6 Fix |
|---|-----------|--------|
| 1 | `doc-sync.agent.md` listed in B (agents section) despite being deleted | Removed from target |
| 2 | E.3 declared "FLAT PS-ID directories" but E.4 showed nested `{PRODUCT}/{SERVICE}/` — direct contradiction | Single canonical nested `configs/{PRODUCT}/{SERVICE}/` structure throughout |
| 3 | F.2 and F.3 each had a spurious duplicate trailing `unseal-5of5.secret` entry | Removed duplicate |
| 4 | `todos` tool name in UPDATE-TOOLS.md not yet updated to `todo` | `todo` used throughout |
| 5 | F.1 included `{PS-ID}-app-sqlite-2.yml` per-service but 3-tier strategy requires only 1 SQLite | Only `sqlite-1.yml` listed |
| 6 | `.vscode/mcp.json` absent (added in commit 672c4974e) | Included |
| 7 | `.github/actions/custom-cicd-lint/` listed but replaced by `download-cicd/` | `download-cicd/` used |
| 8 | `docs/UPDATE-TOOLS.md` not mentioned in docs section | Included |
| 9 | Stale docs (`ARCHITECTURE-TODO.md`, `ARCHITECTURE-INDEX.md`, etc.) planned for deletion in v5 but never deleted | Not listed → deleted by v6 |
| 10 | `deployments/template/` listed for deletion in v5 but never deleted | Not listed → deleted by v6 |
| 11 | Deployment-variant configs in `configs/sm/` listed for deletion in v5 but never deleted | Not listed → deleted by v6 |
| 12 | `configs/sm/kms/` has no canonical config (only deployment variants) | `configs/sm/kms/kms.yml` added as target |
| 13 | `configs/skeleton/skeleton-server.yml` at product level listed for deletion in v5 but never deleted | Not listed → deleted by v6 |
| 14 | Pending internal merges not reflected (docs_validation→lint_docs, github_cleanup→workflow, tls_generator→framework/tls, shared/apperr→framework/apperr) | v6 target reflects post-merge state |
| 15 | `deployments-all-files.json` at deployments root not addressed | Not listed → deleted by v6 |
| 16 | `docs/framework-v3/`, `docs/framework-v4/`, and other historical docs still present | Not listed → deleted by v6 |

---

## Entity Hierarchy (Canonical)

| Level | Variable | Instances |
|-------|----------|-----------|
| Suite | `{SUITE}` | `cryptoutil` |
| Product | `{PRODUCT}` | `identity`, `jose`, `pki`, `skeleton`, `sm` |
| Service | `{SERVICE}` | varies per product (see matrix below) |
| PS-ID | `{PS-ID}` = `{PRODUCT}-{SERVICE}` | 10 total (hyphen-separated, kebab-case) |
| PS\_ID | `{PS_ID}` = `{PRODUCT}_{SERVICE}` | 10 total (underscore variant for SQL/secrets) |
| Infra Tool | `cicd-lint`, `workflow` | 2 |

### Product-Service Matrix

| PS-ID | PS\_ID | Product | Service |
|-------|--------|---------|---------|
| `identity-authz` | `identity_authz` | identity | authz |
| `identity-idp` | `identity_idp` | identity | idp |
| `identity-rp` | `identity_rp` | identity | rp |
| `identity-rs` | `identity_rs` | identity | rs |
| `identity-spa` | `identity_spa` | identity | spa |
| `jose-ja` | `jose_ja` | jose | ja |
| `pki-ca` | `pki_ca` | pki | ca |
| `skeleton-template` | `skeleton_template` | skeleton | template |
| `sm-im` | `sm_im` | sm | im |
| `sm-kms` | `sm_kms` | sm | kms |

### Permission Convention

| Target | Octal |
|--------|-------|
| Directories | 750 |
| Source files (`.go`, `.yml`, `.yaml`, `.md`, `.sql`) | 640 |
| Secret files (`.secret`) | 440 |
| Secret marker files (`.secret.never`) | 440 |
| Executable scripts (`mvnw`) | 750 |
| Generated files (`*.gen.go`) | 640 |

---

## A. Root Level

### A.1 Root Files (KEEP)

```
{ROOT}/
├── .air.toml
├── .dockerignore
├── .editorconfig
├── .gitattributes
├── .gitignore
├── .gitleaks.toml
├── .gofumpt.toml
├── .golangci.yml
├── .gremlins.yaml
├── .markdownlint.jsonc
├── .nuclei-ignore
├── .pre-commit-config.yaml
├── .rgignore
├── .sqlfluff
├── go.mod
├── go.sum
├── LICENSE
├── pyproject.toml
└── README.md
```

### A.2 Root Junk Files — DELETE

All `*.exe`, `*.py`, `coverage*`, `*_coverage`, `*.test.exe`, `*.log`, and
similar build/test artifacts at root level. Git history preserves them.

### A.3 Root Hidden Directories

```
{ROOT}/
├── .cicd/                                 # CICD runtime caches (gitignored)
│   ├── circular-dep-cache.json
│   └── dep-cache.json
├── .ruff_cache/                           # Ruff Python linter cache (gitignored)
├── .semgrep/
│   └── rules/
│       └── go-testing.yml
├── .vscode/
│   ├── cspell.json
│   ├── extensions.json
│   ├── launch.json
│   ├── mcp.json                           # MCP server config (github + playwright)
│   └── settings.json
└── .zap/
    └── rules.tsv
```

---

## B. .github/ — GitHub & Copilot Configuration

```
.github/
├── copilot-instructions.md
├── agents/                                # 4 agents (no doc-sync)
│   ├── beast-mode.agent.md
│   ├── fix-workflows.agent.md
│   ├── implementation-execution.agent.md
│   └── implementation-planning.agent.md
├── actions/
│   ├── docker-compose-build/action.yml
│   ├── docker-compose-down/action.yml
│   ├── docker-compose-logs/action.yml
│   ├── docker-compose-up/action.yml
│   ├── docker-compose-verify/action.yml
│   ├── docker-images-pull/action.yml
│   ├── download-cicd/action.yml           # replaces custom-cicd-lint
│   ├── fuzz-test/action.yml
│   ├── go-setup/action.yml
│   ├── golangci-lint/action.yml
│   ├── security-scan-gitleaks/action.yml
│   ├── security-scan-trivy/action.yml
│   ├── security-scan-trivy2/action.yml
│   ├── workflow-job-begin/action.yml
│   └── workflow-job-end/action.yml
├── instructions/
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
├── skills/
│   ├── README.md
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
└── workflows/
    ├── ci-benchmark.yml
    ├── ci-coverage.yml
    ├── ci-dast.yml
    ├── ci-e2e.yml
    ├── ci-fitness.yml
    ├── ci-fuzz.yml
    ├── ci-gitleaks.yml
    ├── ci-identity-validation.yml
    ├── ci-load.yml
    ├── ci-mutation.yml
    ├── ci-quality.yml                     # includes cicd-lint job (no separate ci-cicd-lint.yml)
    ├── ci-race.yml
    ├── ci-sast.yml
    └── release.yml
```

---

## C. cmd/ — Binary Entry Points

**Rule**: Exactly 18 entries. Each `main.go` delegates to `internal/apps/`.

```
cmd/
├── cryptoutil/main.go                     # Suite CLI → internal/apps/cryptoutil/
├── identity/main.go                       # Product CLI → internal/apps/identity/
├── jose/main.go                           # Product CLI → internal/apps/jose/
├── pki/main.go                            # Product CLI → internal/apps/pki/
├── skeleton/main.go                       # Product CLI → internal/apps/skeleton/
├── sm/main.go                             # Product CLI → internal/apps/sm/
├── identity-authz/main.go                 # Service CLI → internal/apps/identity/authz/
├── identity-idp/main.go                   # Service CLI → internal/apps/identity/idp/
├── identity-rp/main.go                    # Service CLI → internal/apps/identity/rp/
├── identity-rs/main.go                    # Service CLI → internal/apps/identity/rs/
├── identity-spa/main.go                   # Service CLI → internal/apps/identity/spa/
├── jose-ja/main.go                        # Service CLI → internal/apps/jose/ja/
├── pki-ca/main.go                         # Service CLI → internal/apps/pki/ca/
├── skeleton-template/main.go             # Service CLI → internal/apps/skeleton/template/
├── sm-im/main.go                          # Service CLI → internal/apps/sm/im/
├── sm-kms/main.go                         # Service CLI → internal/apps/sm/kms/
├── cicd-lint/main.go                      # Tool CLI → internal/apps/tools/cicd_lint/
└── workflow/main.go                       # Tool CLI → internal/apps/tools/workflow/
```

---

## D. api/ — OpenAPI Specifications & Generated Code

**Rule**: One directory per PS-ID (10 total); no suite-level or product-level API dirs.

```
api/
└── {PS-ID}/                               # ×10
    ├── generate.go
    ├── openapi_spec.yaml
    ├── openapi_spec_components.yaml
    ├── openapi_spec_paths.yaml
    ├── openapi-gen_config_client.yaml
    ├── openapi-gen_config_models.yaml
    ├── openapi-gen_config_server.yaml
    ├── client/
    │   └── client.gen.go
    ├── models/
    │   └── models.gen.go
    └── server/
        └── server.gen.go
```

---

## E. configs/ — Canonical Application Configuration

**Principle**: `configs/` is the single source of truth for what the app needs,
independent of deployment environment. Deployment-specific overlays live in
`deployments/`.

**Structure**: `configs/{PRODUCT}/{SERVICE}/` nested hierarchy for all services.
Suite config at `configs/{SUITE}/`. No flat `configs/{PS-ID}/` at the root level.

```
configs/
├── cryptoutil/
│   └── cryptoutil.yml                     # Suite orchestration config
│
├── identity/
│   ├── policies/                          # Shared identity auth policies
│   │   ├── adaptive-auth.yml
│   │   ├── risk-scoring.yml
│   │   └── step-up.yml
│   ├── authz/
│   │   └── authz.yml
│   ├── idp/
│   │   └── idp.yml
│   ├── rp/
│   │   └── rp.yml
│   ├── rs/
│   │   └── rs.yml
│   └── spa/
│       └── spa.yml
│
├── jose/
│   └── ja/
│       └── jose-ja-server.yml
│
├── pki/
│   └── ca/
│       └── pki-ca-server.yml
│
├── skeleton/
│   └── template/
│       └── skeleton-template-server.yml
│
└── sm/
    ├── im/
    │   └── im.yml                         # canonical only — deployment variants deleted
    └── kms/
        └── kms.yml                        # CREATE: was missing; all deployment variants deleted
```

**Files to DELETE from configs/ (deployment variants and legacy):**

| File | Reason |
|------|--------|
| `configs/skeleton/skeleton-server.yml` | Product-level legacy file |
| `configs/sm/im/sm-im-pg-1.yml` | Deployment variant (belongs in deployments/) |
| `configs/sm/im/sm-im-pg-2.yml` | Deployment variant |
| `configs/sm/im/sm-im-sqlite.yml` | Deployment variant |
| `configs/sm/kms/sm-kms-pg-1.yml` | Deployment variant |
| `configs/sm/kms/sm-kms-pg-2.yml` | Deployment variant |
| `configs/sm/kms/sm-kms-sqlite.yml` | Deployment variant |

---

## F. deployments/ — Deployment Manifests

**Principle**: `deployments/` contains environment-specific manifests that
*consume* configuration from `configs/`. Each tier has its own secrets.

### F.1 Service-Level Deployments (×10)

Each service has exactly **4 config overlays** (1 common + 2 postgres + 1 sqlite)
matching the E2E test strategy: 2 PostgreSQL instances + 1 SQLite instance.

```
deployments/
└── {PS-ID}/                               # ×10 — identity-authz, identity-idp,
    │                                      #        identity-rp, identity-rs,
    │                                      #        identity-spa, jose-ja,
    │                                      #        pki-ca, skeleton-template,
    │                                      #        sm-im, sm-kms
    ├── compose.yml
    ├── Dockerfile
    ├── config/
    │   ├── {PS-ID}-app-common.yml         # shared: bind addresses, TLS, network
    │   ├── {PS-ID}-app-postgresql-1.yml   # postgres: database-driver + url
    │   ├── {PS-ID}-app-postgresql-2.yml   # postgres: database-driver + url
    │   └── {PS-ID}-app-sqlite-1.yml       # sqlite: database-driver + url
    └── secrets/                           # chmod 440
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

### F.2 Product-Level Deployments (×5)

Product secrets are **shared** across all services in the product. Browser,
service, and unseal credentials MUST NOT be set at product level (enforced by
`.secret.never` marker files). PostgreSQL and pepper MAY be shared at product
level.

```
deployments/
└── {PRODUCT}/                             # ×5 — identity, jose, pki, skeleton, sm
    ├── compose.yml
    ├── Dockerfile                         # CREATE: currently missing for all 5 products
    └── secrets/
        ├── browser-password.secret.never  # MUST NOT override at product level
        ├── browser-username.secret.never  # MUST NOT override at product level
        ├── service-password.secret.never  # MUST NOT override at product level
        ├── service-username.secret.never  # MUST NOT override at product level
        ├── hash-pepper-v3.secret
        ├── postgres-database.secret
        ├── postgres-password.secret
        ├── postgres-url.secret
        ├── postgres-username.secret
        ├── unseal-1of5.secret
        ├── unseal-2of5.secret
        ├── unseal-3of5.secret
        ├── unseal-4of5.secret
        └── unseal-5of5.secret
```

### F.3 Suite-Level Deployment (×1)

```
deployments/
└── cryptoutil-suite/
    ├── compose.yml
    ├── Dockerfile
    └── secrets/
        ├── browser-password.secret.never
        ├── browser-username.secret.never
        ├── service-password.secret.never
        ├── service-username.secret.never
        ├── hash-pepper-v3.secret
        ├── postgres-database.secret
        ├── postgres-password.secret
        ├── postgres-url.secret
        ├── postgres-username.secret
        ├── unseal-1of5.secret
        ├── unseal-2of5.secret
        ├── unseal-3of5.secret
        ├── unseal-4of5.secret
        └── unseal-5of5.secret
```

### F.4 Shared Infrastructure

```
deployments/
├── shared-telemetry/
│   ├── compose.yml
│   ├── cryptoutil.yml
│   ├── database.json
│   ├── health.json
│   ├── kms.json
│   ├── prometheus.yml
│   ├── dashboards.yaml
│   ├── prometheus.yaml
│   ├── cryptoutil-otel.yml
│   └── otel-collector-config.yaml
└── shared-postgres/
    ├── .sqlfluff
    ├── compose.yml
    ├── init-follower-databases.sql
    ├── init-leader-databases.sql
    ├── setup-logical-replication.sh
    ├── secrets/
    │   ├── postgres-database.secret
    │   ├── postgres-password.secret
    │   └── postgres-username.secret
```

### F.5 Files to DELETE from deployments/

| Path | Reason |
|------|--------|
| `deployments/template/` (entire dir) | Duplicate of `skeleton-template/`; reconciled in v5 |
| `deployments/deployments-all-files.json` | Metadata artifact, not a manifest |

---

## G. internal/ — Private Application Code

### G.1 internal/apps/ — Application Layer

```
internal/apps/
│
├── cryptoutil/                            # Suite orchestration
│   ├── cryptoutil.go                      #   Suite CLI dispatch (seam pattern)
│   └── *_test.go
│
├── {PRODUCT}/                             # ×5 — identity, jose, pki, skeleton, sm
│   ├── {PRODUCT}.go                       #   Product CLI dispatch
│   ├── *_test.go
│   ├── e2e/                               #   Product-level E2E tests
│   └── shared/                            #   Intra-product shared packages (optional)
│       └── (application-specific subdirs)
│
├── {PRODUCT}/{SERVICE}/                   # ×10 — e.g. sm/kms, sm/im, jose/ja, …
│   ├── {SERVICE}.go                       #   Service entry point (seam pattern)
│   ├── *_test.go
│   ├── server/                            #   HTTP handlers and route registration
│   │   └── *.go
│   ├── client/                            #   Domain-specific HTTP clients
│   │   └── *.go
│   ├── repository/                        #   GORM models + data-access methods
│   │   ├── *.go
│   │   ├── *_test.go
│   │   └── migrations/                    #   Domain migrations (2001+)
│   │       ├── 2001_init.up.sql
│   │       └── 2001_init.down.sql
│   ├── model/                             #   Internal domain value objects (optional)
│   │   └── *.go
│   ├── integration/                       #   Integration tests (optional)
│   │   └── *_integration_test.go
│   └── e2e/                               #   Service-level E2E tests
│       └── *.go
│
├── framework/                             # Shared service framework
│   ├── apperr/                            #   Application error types
│   │   │                                  #   MOVED from internal/shared/apperr/
│   │   └── *.go
│   ├── suite/
│   │   └── cli/
│   │       ├── suite_router.go            #   RouteSuite(), SuiteConfig, ProductEntry
│   │       └── suite_router_test.go
│   ├── product/
│   │   └── cli/
│   │       ├── product_router.go          #   RouteProduct(), ProductConfig, ServiceEntry
│   │       └── product_router_test.go
│   ├── tls/                               #   TLS certificate generation
│   │   │                                  #   MERGED: tls_generator from service/config/
│   │   ├── init.go
│   │   ├── init_test.go
│   │   └── export_test.go
│   └── service/
│       ├── cli/                           #   CLI infrastructure (cobra commands)
│       │   └── *.go
│       ├── client/                        #   HTTP client helpers
│       │   └── *.go
│       ├── config/                        #   Config loading and validation
│       │   └── *.go
│       │   # NOTE: config/tls_generator/ merged into framework/tls/ above
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
│       │   │   ├── migrations/            #   Framework migrations (1001-1999)
│       │   │   └── test_migrations/
│       │   ├── service/
│       │   ├── tenant/
│       │   ├── testutil/
│       │   ├── application.go
│       │   ├── contract.go
│       │   ├── contract_test.go
│       │   ├── public_server_base.go
│       │   ├── service_framework.go
│       │   ├── test_main.go
│       │   ├── ROUTE-REGISTRATION.md
│       │   └── *_test.go
│       ├── server_integration/
│       │   └── *.go
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
└── tools/
    ├── cicd_lint/
    │   ├── cicd.go
    │   ├── cicd_test.go
    │   ├── adaptive-sim/                  #   Adaptive simulation utilities
    │   │   └── *.go
    │   ├── common/                        #   Shared CICD utilities
    │   │   └── *.go
    │   ├── format_go/                     #   Go file formatter
    │   │   └── *.go
    │   ├── format_gotest/                 #   Go test formatter
    │   │   └── *.go
    │   ├── lint_compose/
    │   │   └── *.go
    │   ├── lint_deployments/
    │   │   └── *.go
    │   ├── lint_docs/                     #   Documentation linter
    │   │   │                              #   MERGED: docs_validation/ folded in here
    │   │   └── *.go
    │   ├── lint_fitness/
    │   │   ├── lint_fitness.go
    │   │   ├── lint_fitness_test.go
    │   │   ├── registry/
    │   │   │   ├── registry.go
    │   │   │   └── registry_test.go
    │   │   ├── admin_bind_address/
    │   │   ├── archive_detector/
    │   │   ├── banned_product_names/
    │   │   ├── bind_address_safety/
    │   │   ├── cgo_free_sqlite/
    │   │   ├── check_skeleton_placeholders/
    │   │   ├── cicd_coverage/
    │   │   ├── circular_deps/
    │   │   ├── cmd_anti_pattern/
    │   │   ├── cmd_main_pattern/
    │   │   ├── compose_db_naming/
    │   │   ├── compose_header_format/
    │   │   ├── compose_service_names/
    │   │   ├── configs_deployments_consistency/
    │   │   ├── configs_empty_dir/
    │   │   ├── configs_naming/
    │   │   ├── cross_service_import_isolation/
    │   │   ├── crypto_rand/
    │   │   ├── deployment_dir_completeness/
    │   │   ├── domain_layer_isolation/
    │   │   ├── entity_registry_completeness/
    │   │   ├── file_size_limits/
    │   │   ├── gen_config_initialisms/
    │   │   ├── health_endpoint_presence/
    │   │   ├── insecure_skip_verify/
    │   │   ├── legacy_dir_detection/
    │   │   ├── magic_e2e_compose_path/
    │   │   ├── magic_e2e_container_names/
    │   │   ├── migration_comment_headers/
    │   │   ├── migration_numbering/
    │   │   ├── migration_range_compliance/
    │   │   ├── no_hardcoded_passwords/
    │   │   ├── no_local_closed_db_helper/
    │   │   ├── no_postgres_in_non_e2e/
    │   │   ├── no_unit_test_real_db/
    │   │   ├── no_unit_test_real_server/
    │   │   ├── non_fips_algorithms/
    │   │   ├── otlp_service_name_pattern/
    │   │   ├── parallel_tests/
    │   │   ├── product_structure/
    │   │   ├── product_wiring/
    │   │   ├── require_api_dir/
    │   │   ├── require_framework_naming/
    │   │   ├── service_contract_compliance/
    │   │   ├── service_structure/
    │   │   ├── standalone_config_otlp_names/
    │   │   ├── standalone_config_presence/
    │   │   ├── test_patterns/
    │   │   └── tls_minimum_version/
    │   ├── lint_go/
    │   │   └── *.go
    │   ├── lint_go_mod/
    │   │   └── *.go
    │   ├── lint_golangci/
    │   │   └── *.go
    │   ├── lint_gotest/
    │   │   └── *.go
    │   ├── lint_ports/
    │   │   └── *.go
    │   ├── lint_text/
    │   │   └── *.go
    │   └── lint_workflow/
    │       └── *.go
    │   # NOTE: docs_validation/ merged into lint_docs/ above
    │   # NOTE: github_cleanup/ merged into workflow/ below
    │
    └── workflow/                          #   GitHub Actions workflow management
        │                                  #   MERGED: github_cleanup/ folded in here
        └── *.go
```

### G.2 internal/shared/ — Shared Libraries

```
internal/shared/
├── container/                             # Docker container utilities
│   └── *.go
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
│   └── *.go
├── magic/                                 # Named constants only; excluded from coverage
│   ├── magic_api.go
│   ├── magic_cicd.go
│   ├── magic_console.go
│   ├── magic_crypto.go
│   ├── magic_database.go
│   ├── magic_docker.go
│   ├── magic_framework.go
│   ├── magic_identity.go
│   ├── magic_identity_adaptive.go
│   ├── magic_identity_config.go
│   ├── magic_identity_http.go
│   ├── magic_identity_keys.go
│   ├── magic_identity_metrics.go
│   ├── magic_identity_mfa.go
│   ├── magic_identity_oauth.go
│   ├── magic_identity_oidc.go
│   ├── magic_identity_pbkdf2.go
│   ├── magic_identity_scopes.go
│   ├── magic_identity_testing.go
│   ├── magic_identity_timeouts.go
│   ├── magic_identity_uris.go
│   ├── magic_jose.go
│   ├── magic_memory.go
│   ├── magic_misc.go
│   ├── magic_network.go
│   ├── magic_orchestration.go
│   ├── magic_percent.go
│   ├── magic_pki.go
│   ├── magic_pki_ca.go
│   ├── magic_pkix.go
│   ├── magic_security.go
│   ├── magic_session.go
│   ├── magic_skeleton.go
│   ├── magic_sm.go
│   ├── magic_sm_im.go
│   ├── magic_telemetry.go
│   ├── magic_testing.go
│   ├── magic_unseal.go
│   └── magic_workflows.go
│   # NOTE: magic_demo.go deleted (demo code removed in v5)
│   # NOTE: magic_pkiinit.go deleted (pkiinit merged into framework/tls)
├── pool/
│   └── *.go
├── pwdgen/
│   └── *.go
├── telemetry/
│   └── *.go
├── testutil/
│   └── *.go
└── util/
    ├── cache/
    ├── combinations/
    ├── datetime/
    ├── files/
    ├── network/
    ├── poll/
    ├── random/
    ├── slice.go
    ├── slice_test.go
    ├── sysinfo/
    ├── thread/
    ├── yml_json.go
    └── yml_json_test.go
# NOTE: shared/apperr/ deleted — moved to internal/apps/framework/apperr/
```

---

## H. docs/ — Documentation

```
docs/
├── ARCHITECTURE.md                        # SSOT: Architecture reference
├── CONFIG-SCHEMA.md                       # Config file schema reference
├── DEV-SETUP.md                           # Developer setup guide
├── README.md                              # Documentation index
├── UPDATE-TOOLS.md                        # Agent tool matrix (todos column = todo)
└── framework-v6/                          # Active plan (THIS iteration)
    ├── lessons.md
    ├── plan.md
    ├── tasks.md
    └── target-structure.md               # THIS FILE
```

**Files/directories to DELETE from docs/:**

| Path | Reason |
|------|--------|
| `docs/ARCHITECTURE-TODO.md` | Superseded by plan tracking in framework-v*/tasks.md |
| `docs/ARCHITECTURE-INDEX.md` | Superseded by ARCHITECTURE.md built-in ToC |
| `docs/COPILOT-MULTI-PROJECT.md` | Stale reference doc |
| `docs/DEAD_CODE_REVIEW.md` | Completed review; no longer needed |
| `docs/VSCODE-CRASHES.md` | Stale troubleshooting doc |
| `docs/gremlins/` | Stale mutation testing notes |
| `docs/LESSONS/` | Cross-plan archive superseded by per-plan lessons.md |
| `docs/framework-brainstorm/` | Superseded by framework-v3+ |
| `docs/framework-v3/` | Historical (completed) |
| `docs/framework-v4/` | Historical (completed) |
| `docs/framework-v5/` | Historical (completed; this is framework-v6) |
| `docs/workflow-runtimes/` | Stale workflow analysis |

---

## I. test/ — External Test Suites

```
test/
└── load/                                  # Gatling load tests (Java 21 + Maven)
    ├── .gitignore
    ├── .mvn/
    ├── mvnw                               # chmod 750
    ├── mvnw.cmd
    ├── pom.xml
    ├── README.md
    └── src/
```

---

## J. pkg/ — Public Library Code (Reserved)

```
pkg/                                       # Currently empty; reserved for future public APIs
```

---

## K. Other Root Directories

```
scripts/                                   # Empty; keep (.gitkeep)
workflow-reports/                          # Ephemeral test output; gitignored, never committed
test-output/                               # Ephemeral test output; gitignored, never committed
testdata/                                  # DELETE: move contents to owning packages
```

---

## L. Secret File Naming Convention

All tiers use **identical `{purpose}.secret` names** with no tier prefix on
active secret files. Tier prefixes appear ONLY on `.secret.never` marker files.

### Active Secret Files

| File | Service tier value | Product/Suite tier value |
|------|--------------------|--------------------------|
| `browser-password.secret` | `{PS-ID}-browser-{base64-32}` | `.never` (MUST NOT share) |
| `browser-username.secret` | `{PS-ID}-browser-user` | `.never` (MUST NOT share) |
| `service-password.secret` | `{PS-ID}-service-{base64-32}` | `.never` (MUST NOT share) |
| `service-username.secret` | `{PS-ID}-service-user` | `.never` (MUST NOT share) |
| `hash-pepper-v3.secret` | `{PS-ID}-hash-pepper-v3-{base64-32}` | MUST be set per tier |
| `postgres-database.secret` | `{PS_ID}_database` | MUST be set per tier |
| `postgres-password.secret` | `{PS_ID}_database_pass-{base64-32}` | MUST be set per tier |
| `postgres-url.secret` | `postgres://{PS_ID}_database_user:…@{PS-ID}-postgres:5432/{PS_ID}_database` | MUST be set per tier |
| `postgres-username.secret` | `{PS_ID}_database_user` | MUST be set per tier |
| `unseal-Nof5.secret` (N=1..5) | `{SERVICE}-{hex-32}` | MUST be set per tier |

### Marker Files (`.secret.never`)

Present at product and suite level only. Purpose: explicit reminder that
browser/service credentials are service-specific and MUST NOT be shared.

| File | Content |
|------|---------|
| `browser-password.secret.never` | "MUST NOT be set at this level. Use service-specific secrets." |
| `browser-username.secret.never` | "MUST NOT be set at this level. Use service-specific secrets." |
| `service-password.secret.never` | "MUST NOT be set at this level. Use service-specific secrets." |
| `service-username.secret.never` | "MUST NOT be set at this level. Use service-specific secrets." |

---

## M. Agent Tool Matrix Reference

The `docs/UPDATE-TOOLS.md` table columns map to agent files in `.github/agents/`.
Correct tool name: **`todo`** (not `todos` — renamed in VS Code).

| Agent Column | Agent File |
|---|---|
| `beast-mode` | `beast-mode.agent.md` |
| `fix-wf` | `fix-workflows.agent.md` |
| `impl-exec` | `implementation-execution.agent.md` |
| `impl-plan` | `implementation-planning.agent.md` |

All four agents include `edit/insertEdit` in their `tools:` list.

---

## N. Framework v6 Change Summary

Changes required to reach this target from current repository state.

| Area | Current State | v6 Target | Action |
|------|--------------|-----------|--------|
| `docs/` stale files | 10+ stale docs/dirs present | Only essential docs + framework-v6/ | DELETE all listed in H |
| `deployments/template/` | Still present | Removed | DELETE |
| `deployments/deployments-all-files.json` | Present | Removed | DELETE |
| `configs/sm/kms/` canonical | No canonical config (only deployment variants) | `kms.yml` created | CREATE + DELETE variants |
| `configs/sm/im/` deployment variants | 3 variant files present | Deleted; only `im.yml` remains | DELETE variants |
| `configs/skeleton/skeleton-server.yml` | Present at product level | Removed | DELETE |
| `internal/apps/tools/cicd_lint/docs_validation/` | Separate package | Merged into `lint_docs/` | MERGE + DELETE dir |
| `internal/apps/tools/cicd_lint/github_cleanup/` | Separate package | Merged into `tools/workflow/` | MERGE + DELETE dir |
| `internal/apps/framework/service/config/tls_generator/` | Separate package | Merged into `framework/tls/` | MERGE + DELETE dir |
| `internal/shared/apperr/` | In shared/ | Moved to `framework/apperr/` | MOVE + DELETE old dir |
| `internal/shared/magic/magic_demo.go` | Present | Removed (demo deleted in v5) | DELETE |
| `internal/shared/magic/magic_pkiinit.go` | Present | Removed (pkiinit merged) | DELETE |
| `deployments/{PRODUCT}/Dockerfile` | Missing for all 5 products | Add to each product deployment | CREATE ×5 |
| `deployments/{PRODUCT}/secrets/*.secret.never` | Missing for all 5 products | Add 4 marker files per product | CREATE ×20 |
| `deployments/cryptoutil-suite/secrets/*.secret.never` | Missing | Add 4 marker files | CREATE ×4 |
| `testdata/` root dir | Present | Deleted; files moved to owning packages | DELETE |
| `docs/UPDATE-TOOLS.md` `todos` row | Named `todos` | Renamed to `todo` | DONE (this session) |
