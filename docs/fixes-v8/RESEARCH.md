# Service Template Reusability Research

**Phases 5-8 Analysis** — Documenting findings from skeleton-template (Phase 5), pki-ca clean-slate (Phase 6), template reusability (Phase 7), and CICD linter enhancements (Phase 8).

---

## Task 7.1: Minimal File Set for a Conforming Product-Service

### Required Source Files (7 files + 2 SQL)

| # | File | Purpose | Auto-Gen? |
|---|------|---------|-----------|
| 1 | `SERVICE.go` | Service entry point: CLI routing via `RouteService()`, `serverStart()`, `client()`, `init()` | Template (90%+ boilerplate) |
| 2 | `SERVICE_usage.go` | Usage string constants (8 strings: main, server, client, init, health, livez, readyz, shutdown) | Template (100% boilerplate, only names change) |
| 3 | `server/server.go` | Server struct wrapping `Application`, `NewFromConfig()` using builder, `Start()`, `Shutdown()`, port/URL accessors | Template (95%+ boilerplate) |
| 4 | `server/config/config.go` | `Settings` struct embedding `*ServiceTemplateServerSettings`, `ParseWithFlagSet()`, `Validate()` | Template (90%+ boilerplate) |
| 5 | `domain/model.go` | Domain model struct(s) with GORM tags, `TableName()` method | Hand-written (domain-specific) |
| 6 | `repository/migrations.go` | `MigrationsFS` embed, `mergedFS` type, `GetMergedMigrationsFS()`, `ApplyMigrations()` | Template (95%+ boilerplate) |
| 7 | `repository/migrations/2001_*.up.sql` | Migration SQL for domain tables | Hand-written (domain-specific) |
| 8 | `repository/migrations/2001_*.down.sql` | Migration rollback SQL | Hand-written (domain-specific) |

### Required Test Files (10-12 files)

| # | File | Purpose | Auto-Gen? |
|---|------|---------|-----------|
| 1 | `testmain_test.go` | Package-level TestMain: server lifecycle, shared HTTP client, base URLs | Template (95%+ boilerplate) |
| 2 | `SERVICE_cli_test.go` | CLI subcommand help, version, not-implemented, parse error, create error | Template (90%+ boilerplate) |
| 3 | `SERVICE_lifecycle_test.go` | Full server start → SIGINT → graceful shutdown | Template (95%+ boilerplate) |
| 4 | `SERVICE_port_conflict_test.go` | Port conflict error path | Template (100% boilerplate) |
| 5 | `server/testmain_test.go` | Server-level TestMain with `MustStartAndWaitForDualPorts` | Template (95%+ boilerplate) |
| 6 | `server/server_test.go` | Unit tests: nil ctx, nil cfg, invalid DB URL | Template (90%+ boilerplate) |
| 7 | `server/server_integration_test.go` | Integration: lifecycle, ports, accessors, health endpoints, shutdown | Template (90%+ boilerplate) |
| 8 | `server/config/config_test_helper.go` | `NewTestConfig()`, `DefaultTestConfig()` | Template (95%+ boilerplate) |
| 9 | `server/config/config_test.go` | Parse defaults, custom port, invalid flag, nil base | Template (90%+ boilerplate) |
| 10 | `domain/model_test.go` | TableName, fields, zero value | Template (80%+ boilerplate) |
| 11 | `repository/migrations_test.go` | Embedded FS, merged FS, read/stat, apply, error paths | Template (90%+ boilerplate) |
| 12 | `e2e/e2e_test.go` (optional) | E2E with Docker containers | Template (90%+ boilerplate) |

### Required Wiring Files (outside service dir)

| # | File | Change Required |
|---|------|----------------|
| 1 | `cmd/PRODUCT-SERVICE/main.go` | Trivial: 5-line delegation to service entry point |
| 2 | `internal/apps/PRODUCT/PRODUCT.go` | Add `ServiceEntry` to `RouteProduct()` call |
| 3 | `internal/apps/PRODUCT/PRODUCT_test.go` | Add service routing test |
| 4 | `internal/apps/cryptoutil/cryptoutil.go` | Add case to suite switch (if new product) |
| 5 | `internal/shared/magic/magic_PRODUCT.go` | Add magic constants (port, service ID, OTLP name) |

### Comparison: Skeleton vs PKI-CA vs SM-KMS

| Metric | skeleton-template | pki-ca (skeleton) | sm-kms (full service) |
|--------|-------------------|--------------------|-----------------------|
| Source files (non-test) | 7 | 7 | 41 |
| Test files | 12 | 10 | 78 |
| SQL migrations | 2 | 2 | 4 |
| **Total files** | **21** | **19** | **123** |
| Domain complexity | Minimal (1 model) | Minimal (1 model) | Full (handlers, business logic, ORM mappers, middleware) |

**Finding**: A conforming skeleton requires ~19-21 files (7 source + 10-12 test + 2 SQL). A full-featured service like sm-kms adds ~100+ files of domain-specific business logic.

---

## Task 7.2: Template Friction Points

### Copy-Paste Boilerplate

1. **`repository/migrations.go`** — The `mergedFS` type (Open, ReadFile, Stat, ReadDir) is identical across all services. Only the embed path and migration function name change. ~80 lines of pure boilerplate per service.

2. **`server/server.go`** — The `PKICAServer`/`SkeletonTemplateServer` wrapper struct is identical pattern: store `app` + `db`, delegate all methods (`Start`, `Shutdown`, `PublicPort`, `AdminPort`, `SetReady`, `PublicBaseURL`, `AdminBaseURL`, `DB`). ~90 lines of identical structure, only the type name changes.

3. **`SERVICE_usage.go`** — 8 string constants that differ only in product/service name. Pure template with search-and-replace of names.

4. **`SERVICE.go`** — The `serverStart()` function follows an identical pattern: parse config → create server → set ready → start in goroutine → wait for signal → shutdown. ~80 lines of identical flow logic.

5. **Test files** — Of the 10-12 test files, ~90% is boilerplate that only changes type names and magic constant references.

### Missing Helper Functions

1. **No `mergedFS` helper** — Each service re-implements the `mergedFS` type. A shared `migration.NewMergedFS(domainFS, templateFS)` would eliminate ~80 lines per service.

2. **No server wrapper generator** — The `ServerWrapper` struct (delegate all Application methods) is ~90 lines of pure delegation. A generic wrapper or code generator would help.

3. **No test template generator** — The 10+ test files are algorithmically derivable from the service name, magic constants, and domain model.

### API Surface Confusion

1. **`ServiceResources` fields vs methods** — `resources.DB` is a field, `resources.Application` is a field. This was initially coded as method calls (`resources.DB()`, `resources.Application()`), causing build errors. The API surface should be more discoverable (or documented in godoc).

2. **`Application.Shutdown(ctx)` signature** — The single-context signature was initially assumed to be variadic, causing build errors.

3. **Port return types** — `PublicPort()` and `AdminPort()` return `int`, not `uint16`. This caused build errors when implementing the server wrapper.

---

## Task 7.3: Product Wiring Analysis

### RouteProduct/RouteService Patterns

The `RouteProduct` and `RouteService` functions in `internal/apps/template/service/cli/` are well-designed:

- **RouteProduct** (69 lines): Takes `ProductConfig`, args, and a `[]ServiceEntry`. Routes by service name (first arg). Clean pattern.
- **RouteService** (101 lines): Takes `ServiceConfig`, args, and handler functions. Routes by subcommand. Includes built-in health/livez/readyz/shutdown via HTTP client. Clean pattern.
- **Health commands** (288 lines): Shared HTTP client for all health subcommands. Reusable across all services.

**Assessment**: The routing pattern is clean and scalable. No significant improvements needed.

### Suite Router Scalability

The suite router (`cryptoutil.go`) uses a `switch` statement for product routing. Currently 5 products (sm, identity, jose, pki, skeleton). This scales well up to ~20 products. For beyond that, a map-based approach would be cleaner, but the current switch is fine.

**No changes recommended** for product/suite routing.

---

## Task 7.4: Enhancement Proposals

### P0: Must-Fix (High Impact, Low Effort)

| # | Problem | Solution | LOE | Impact |
|---|---------|----------|-----|--------|
| P0.1 | `mergedFS` type duplicated in every service (~80 lines each) | Extract shared `migration.NewMergedFS(domainFS, templateFS)` in `internal/apps/template/service/` | 2h | Eliminates ~80 lines per service × 10 services = 800 lines |
| P0.2 | `ServiceResources` field vs method confusion | Add comprehensive godoc to `ServiceResources` struct fields documenting they are fields, not methods | 0.5h | Prevents build errors for new service authors |

### P1: Should-Fix (Medium Impact, Medium Effort)

| # | Problem | Solution | LOE | Impact |
|---|---------|----------|-----|--------|
| P1.1 | Server wrapper struct is ~90 lines of pure delegation per service | Provide `ServerBase` base struct with generic delegation, services embed it | 4h | Eliminates ~90 lines per service × 10 = 900 lines |
| P1.2 | Usage strings differ only in product/service names | Create `GenerateUsageTexts(productName, serviceName, port)` function | 1h | Eliminates `_usage.go` file per service entirely |
| P1.3 | `serverStart()` function is identical pattern across services | Extract shared `RunServer(ctx, parseConfig, createServer)` | 2h | Reduces each `SERVICE.go` by ~50 lines |

### P2: Nice-to-Have (Low Impact, Higher Effort)

| # | Problem | Solution | LOE | Impact |
|---|---------|----------|-----|--------|
| P2.1 | 10+ test files are algorithmically derivable | Code generator: `go generate` command producing skeleton test files from service name and model | 8h | Faster service creation, but tests will need manual tuning |
| P2.2 | Config parsing is repetitive (embed + override port + OTLP name) | Generic config factory with port/OTLP params | 2h | Minor reduction per service |

### Priority Summary

**Immediate wins (P0)**: Extract shared `mergedFS`, improve godoc → saves ~800 lines, prevents confusion.
**Next iteration (P1)**: Server base struct, usage generator, shared server runner → saves ~2,000+ lines across all services.
**Future (P2)**: Code generation for test scaffolding → productivity improvement for new services.

## Task 9.1: Skeleton Creation Patterns

### Step-by-Step Guide for Creating a New Product-Service

**Prerequisites**: Product name (kebab-case), service name (kebab-case), assigned port range from ARCHITECTURE.md Section 3.4.

#### Step 1: Magic Constants
Add constants to `internal/shared/magic/magic_PRODUCT.go`:
- Service ID, product name, service name
- Public port, admin port, PostgreSQL port
- OTLP service name
- Default bind addresses

#### Step 2: Product Entry Point
Create `cmd/PRODUCT/main.go` — 5-line delegation to `internal/apps/PRODUCT/PRODUCT.go`.

#### Step 3: Service Entry Point
Create `cmd/PRODUCT-SERVICE/main.go` — 5-line delegation to `internal/apps/PRODUCT/SERVICE/SERVICE.go`.

#### Step 4: Product Router
Create `internal/apps/PRODUCT/PRODUCT.go` with `RouteProduct()` registering the service via `cli.RouteProduct()`.

Create `internal/apps/PRODUCT/PRODUCT_test.go` with product routing tests.

#### Step 5: Service Implementation (7 files)
1. `internal/apps/PRODUCT/SERVICE/SERVICE.go` — CLI routing, `serverStart()`, `client()`
2. `internal/apps/PRODUCT/SERVICE/SERVICE_usage.go` — 8 usage string constants
3. `internal/apps/PRODUCT/SERVICE/server/server.go` — Server struct, builder integration
4. `internal/apps/PRODUCT/SERVICE/server/config/config.go` — Settings struct, flag parsing
5. `internal/apps/PRODUCT/SERVICE/domain/model.go` — Domain model with GORM tags
6. `internal/apps/PRODUCT/SERVICE/repository/migrations.go` — Embedded migrations FS
7. `internal/apps/PRODUCT/SERVICE/repository/migrations/2001_*.{up,down}.sql` — SQL migrations

#### Step 6: Test Files (10-12 files)
Follow the test file inventory from Task 7.1 (see above). All tests follow established patterns — copy from skeleton-template or pki-ca and search/replace names.

#### Step 7: Suite Wiring
Add product case to `internal/apps/cryptoutil/cryptoutil.go` switch statement.

#### Step 8: Deployments & Configs
Create `deployments/PRODUCT-SERVICE/compose.yml` and `configs/PRODUCT/config-SERVICE.yml`.

#### Step 9: Validation
- `go build ./...` — clean
- `go test ./... -shuffle=on` — all pass
- `golangci-lint run` — zero issues
- `go run ./cmd/cicd lint-go` — all validators pass
- `go run ./cmd/cicd lint-deployments validate-all` — all validators pass

### Common Pitfalls

| Pitfall | Symptom | Fix |
|---------|---------|-----|
| `ServiceResources` fields used as methods | `resources.DB()` build error | Use `resources.DB` (field, not method) |
| `Application.Shutdown` assumed variadic | `Shutdown(ctx, timeout)` build error | Use `Shutdown(ctx)` (single context) |
| Port return type mismatch | `uint16` vs `int` | `PublicPort()` and `AdminPort()` return `int` |
| Missing `mergedFS` for migrations | golang-migrate version mismatch | Use `GetMergedMigrationsFS()` pattern |
| Hardcoded ports in tests | Windows TIME_WAIT failures | ALWAYS use port 0 in tests |
| `localhost` in Alpine containers | IPv6 resolution failures | Use `127.0.0.1` explicitly |

---

## Task 9.2: Template Learnings (Consolidated)

### Strengths
1. **Builder pattern** (`NewServerBuilder`) eliminates ~48,000 lines of boilerplate per service
2. **Dual HTTPS** (public + admin) is cleanly separated
3. **Health checks** (livez/readyz/shutdown) are automatic
4. **CLI routing** (`RouteProduct`/`RouteService`) scales well to 20+ services
5. **Migration merging** handles template + domain migrations seamlessly
6. **GORM cross-DB** compatibility (PostgreSQL + SQLite) works with `type:text` pattern

### Weaknesses (from Task 7.2)
1. **Copy-paste boilerplate**: `mergedFS` (~80 lines), server wrapper (~90 lines), usage strings, `serverStart()` — all duplicated per service
2. **API surface confusion**: Field vs method on `ServiceResources`, `Shutdown` signature, port return types
3. **No code generation**: 90%+ of test files are algorithmically derivable but hand-written

### Enhancement Roadmap (from Task 7.4)
- **P0 (immediate)**: Extract shared `mergedFS` helper, improve `ServiceResources` godoc
- **P1 (next iteration)**: `ServerBase` embed, usage string generator, shared `RunServer()`
- **P2 (future)**: Code generator for test scaffolding

---

## Task 9.3: Identity Services Roadmap

### Current State
- 5 identity services: authz, idp, rs, rp, spa
- All use `NewServerBuilder` (confirmed in Phase 2 scoring)
- Shared domain layer: 44 domain files + 47 repository files (Phase 3 finding)
- Shared migration numbering: 0002-0011 range (non-standard, excluded from CICD validator)
- identity-authz and identity-idp: 133/129 files respectively (complex, advanced services)
- identity-rs/rp/spa: 18/10/10 files (minimal, early-stage)

### Planned Approach: Archive + Skeleton

Following the proven pki-ca pattern (Phase 6):

1. **Archive** existing identity service code to `_SERVICE-archived/` directories
2. **Create skeleton** for each service using Step-by-Step Guide (Task 9.1)
3. **Incrementally restore** domain logic from archive, adapting to template patterns
4. **Independent deployability** (ED-7): Each service gets own DB, migration range (2001+), and deployment config
5. **Shared E2E** (ED-10): Single E2E suite tests all 5 services together (decided)

### Migration Priority
Per ARCHITECTURE.md: `sm-im -> jose-ja -> sm-kms -> pki-ca -> identity services`

The first 4 are migrated. Identity services are the final frontier.

### Independent Deployability Requirements (ED-7)
- Separate PostgreSQL databases per service
- Migration numbering in 2001+ range (replace legacy 0002-0011)
- Individual `deployments/identity-SERVICE/compose.yml` configs
- Per-service health endpoints

### Shared E2E Strategy (ED-10)
- Single E2E test suite exercising all 5 identity services
- Tests verify cross-service flows (authn → authz → resource access)
- Docker Compose brings up all 5 services + shared PostgreSQL

### Timeline
- **Post-fixes-v8**: Identity migration is deferred to dedicated work phase
- **Estimated effort**: 2-4 weeks (archiving + skeleton + restore per service)

---

## Task 9.4: Three-Tier Architecture Vision

### Architecture Layers

```
┌─────────────────────────────────────────────────┐
│                 Service Tier                      │
│   sm-kms, sm-im, jose-ja, pki-ca,               │
│   identity-{authz,idp,rs,rp,spa}                │
│   (Business logic, domain models, APIs)          │
├─────────────────────────────────────────────────┤
│              Stereotype Tier                      │
│   skeleton-template (current)                     │
│   (Validates template patterns, reference impl)  │
├─────────────────────────────────────────────────┤
│                 Base Tier                         │
│   service-template (current)                      │
│   (Builder, health, TLS, migrations, telemetry)  │
└─────────────────────────────────────────────────┘
```

### Base Tier: service-template
- **Location**: `internal/apps/template/service/`
- **Purpose**: Core infrastructure shared by ALL services
- **Components**: ServerBuilder, Application, dual HTTPS listeners, health checks, barrier/unseal, session management, realm management, telemetry, GORM database
- **Change frequency**: Low (infrastructure changes)
- **Impact of change**: ALL services affected

### Stereotype Tier: skeleton-template
- **Location**: `internal/apps/skeleton/template/`
- **Purpose**: Reference implementation demonstrating correct template usage
- **Components**: Minimal domain model, standard file layout, full test coverage
- **Change frequency**: When template patterns evolve
- **Validation role**: If skeleton-template breaks, the template change is wrong

### Service Tier: Business Services
- **Location**: `internal/apps/{sm,jose,pki,identity}/SERVICE/`
- **Purpose**: Domain-specific business logic
- **Components**: Domain models, handlers, middleware, business logic, domain migrations
- **Change frequency**: High (feature development)
- **Template compliance**: Validated by Phase 8 CICD linters

### Long-Term Workflow

```
1. Change base tier (service-template)
   ↓
2. Validate stereotype tier (skeleton-template)
   - If skeleton breaks → fix base or skeleton
   - If skeleton passes → proceed
   ↓
3. Roll out to service tier (all 9 other business services)
   - CICD linters catch structural non-compliance
   - Per-service tests catch functional regressions
```

### Benefits
- **Canary validation**: skeleton-template is the first consumer of any template change
- **Structural enforcement**: CICD linters (Phase 8) prevent drift from template patterns
- **Independent evolution**: Services add domain logic without breaking template compliance
- **New service creation**: Follow skeleton-template as reference (~19-21 files to start)

---

## Cross-Reference

- **Architecture**: See [ARCHITECTURE.md Section 5.1](../../docs/ARCHITECTURE.md#51-service-template-pattern) for service template pattern.
- **Phase 5**: skeleton-template creation (validates pattern reproducibility).
- **Phase 6**: pki-ca clean-slate (validates pattern for real service).
- **Phase 8**: CICD linter enhancements (structural validators based on these findings).
- **Phase 9**: Documentation consolidation (this section).
