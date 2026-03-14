# Quizme v4 — Definitive Decision Questions

**Purpose**: Answer the two remaining "E" (need more info) answers from quizme-v3.
**Created**: 2026-03-15
**Context**: Q1 (skeleton-template role) has been answered "E" FOUR TIMES across v1/v2/v3. This is the definitive treatment with maximum depth.

---

## Q1: What should skeleton-template be? (D12 — 4th iteration)

### Executive Summary

You have 10 product-services. ALL use the same framework (`internal/apps/template/`). The framework bootstraps everything: HTTPS servers, TLS, database, auth, CORS, CSRF, sessions, barrier, telemetry, health checks. Your product-services are thin wrappers that inject domain-specific code.

The question: **What should the skeleton-template service contain?**

### How Creating a New Service Works Today

When you want to create the 11th service (e.g., `audit-log`), here's what happens:

1. **You invoke `/new-service audit-log`** (Copilot skill)
2. Copilot copies `internal/apps/skeleton/template/` → `internal/apps/audit/log/`
3. Copilot renames all `skeleton` → `audit`, `template` → `log`
4. You get a running service with ZERO domain logic — just health endpoints

**What you'd get (today, Option A) after `/new-service audit-log`:**

```go
// internal/apps/audit/log/server/server.go (ENTIRE file, ~90 lines)

type AuditLogServer struct {
    app *Application
    db  *gorm.DB
    // Framework services (all auto-wired by builder):
    telemetryService      *TelemetryService
    jwkGenService         *JWKGenService
    barrierService        *BarrierService
    sessionManagerService *SessionManagerService
    realmService          RealmService
    realmRepo             TenantRealmRepository
}

func NewFromConfig(ctx context.Context, cfg *AuditLogServerSettings) (*AuditLogServer, error) {
    builder := NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

    builder.WithDomainMigrations(repository.MigrationsFS, "migrations")

    builder.WithPublicRouteRegistration(func(base *PublicServerBase, res *ServiceResources) error {
        // YOUR DOMAIN CODE GOES HERE
        // Currently empty — just health endpoints from the framework
        return nil
    })

    resources, err := builder.Build()
    // ... create server struct from resources ...
    return server, nil
}
```

**What you do NEXT** (the part that varies by option):

### Option A: Keep Skeleton Minimal (Status Quo)

After `/new-service`, you open sm-im side-by-side and copy the pattern:

```go
// You look at sm-im and add YOUR domain code inside WithPublicRouteRegistration:

builder.WithPublicRouteRegistration(func(base *PublicServerBase, res *ServiceResources) error {
    // Step 1: Create your repositories (1 line each)
    auditEntryRepo := repository.NewAuditEntryRepository(res.DB)

    // Step 2: Create your public server (wires handlers to repos)
    publicServer, err := NewPublicServer(base, res.SessionManager, res.RealmService, auditEntryRepo)
    if err != nil { return err }

    // Step 3: Register routes
    return publicServer.registerRoutes()
})
```

Then you create your domain files:
- `repository/audit_entry.go` — GORM model + repository
- `repository/migrations/2001_init.up.sql` — domain DB schema
- `server/public_server.go` — HTTP handlers
- `server/config/config.go` — domain config
- `domain/model.go` — business logic models (if needed)

**Effort to add domain**: ~30-60 min with Copilot assistance. You always have sm-im and jose-ja as real working references.

### Option B: Add Minimal CRUD Example to Skeleton

After `/new-service`, you get the SAME skeleton BUT with one working example endpoint already included:

```go
// You'd get this pre-wired inside WithPublicRouteRegistration:

builder.WithPublicRouteRegistration(func(base *PublicServerBase, res *ServiceResources) error {
    // Example: one CRUD repository (DELETE or rename for your domain)
    itemRepo := repository.NewItemRepository(res.DB)

    publicServer, err := NewPublicServer(base, res.SessionManager, res.RealmService, itemRepo)
    if err != nil { return err }

    return publicServer.registerRoutes()
})
```

Plus you'd get example files you can rename/delete:
- `repository/item.go` — Example GORM model + CRUD repository (~40 lines)
- `repository/migrations/2001_init.up.sql` — Example table
- `server/public_server.go` — Example GET/POST/PUT/DELETE handlers (~60 lines)

**What you'd do**: Rename `Item` → `AuditEntry`, adjust fields, done.

**Effort to set up**: ~4h to create the example. ~15 min/quarter to keep it current.

### Side-by-Side Comparison

| Aspect | A: Minimal (Status Quo) | B: CRUD Example |
|--------|------------------------|-----------------|
| What `/new-service` generates | Empty shell + health | Shell + 1 working CRUD endpoint |
| First thing developer does | Open sm-im, copy pattern | Rename example, adjust fields |
| Time to first domain endpoint | ~30-60 min | ~10-15 min |
| Maintenance cost | Zero | ~15 min/quarter |
| Creation cost | Zero (already done) | ~4 hours one-time |
| Risk of example going stale | N/A | Low (lint-fitness catches drift) |
| Copilot context needed | More (must read sm-im) | Less (example is self-contained) |
| Useful for non-AI developers | Less (must find reference) | More (example in place) |

### "Does It Actually Matter?" Analysis

**The honest answer**: For YOU (solo developer + Copilot), the difference is negligible. Here's why:

1. **Copilot already has full codebase context**. When you say "add a CRUD endpoint for audit entries," Copilot looks at sm-im and jose-ja regardless of what's in the skeleton.
2. **You'll create maybe 1-2 more services** before the architecture stabilizes. The `/new-service` skill runs at most a handful more times.
3. **The skeleton's primary value is as a lint-fitness validation target and documentation**, not as a daily-use template.

**When B matters**: If you ever onboard another developer or use Copilot in a workspace where it can't see the full codebase (e.g., reviewing a single service in isolation).

### My Recommendation

**Option A (keep minimal)**. Confidence: **HIGH**.

Rationale:
- You already have 3+ working CRUD services as references (sm-im, jose-ja, sm-kms)
- Copilot generates domain code from those references more effectively than from a toy example
- Zero maintenance cost
- The only scenario where B wins (onboarding a new developer) is not your current situation
- If you later want B, it's a ~4h task with no dependencies — can add anytime

### Important Clarification: How This Is Like Spring Boot

Your current architecture already works like Spring Boot:

| Spring Boot | Your Framework |
|-------------|---------------|
| `@SpringBootApplication` | `NewServerBuilder(ctx, cfg)` + `builder.Build()` |
| Auto-configuration (web, security, JPA) | Builder auto-wires TLS, auth, CORS, CSRF, sessions, barrier, telemetry |
| `@RestController` + `@RequestMapping` | `WithPublicRouteRegistration(func(base, res) error { ... })` |
| `application.yml` | `config-skeleton-template.yml` |
| JPA entities + `@Repository` | GORM models + `NewXxxRepository(db)` |
| `src/main/resources/db/migration/` | `repository/migrations/2001_init.up.sql` |
| Spring Boot Starter (parent POM) | `internal/apps/template/` (shared framework) |

**Key insight**: Your services are ALREADY as thin as Spring Boot apps. The `NewFromConfig` function is your `@SpringBootApplication` class. The `WithPublicRouteRegistration` callback is where you wire your `@RestController` equivalents.

### Answer Options

| Option | What it means |
|--------|--------------|
| **A) Keep minimal** | Skeleton stays as-is. Use sm-im/jose-ja as CRUD references. Zero maintenance. |
| **B) Add CRUD example** | Add ~100 lines of example CRUD to skeleton. ~4h work, ~15 min/quarter maintenance. |

**Answer**: ___

---

## Q2: Docker Testing Strategy for Services with PostgreSQL (D14 supplement)

### Executive Summary

Your services currently use SQLite in-memory for unit and integration tests. This is fast (<2ms startup) and CGO-free. But real production uses PostgreSQL. The question is: **when and how should you add PostgreSQL-backed tests?**

### Current State (What You Have Today)

```
Unit Tests:       SQLite in-memory (testdb.NewInMemorySQLiteDB)
                  → <2ms startup, in-process, no Docker needed
                  → Tests ALL GORM logic, migrations, constraints

Integration Tests: SQLite in-memory (same as unit)
                   → Server starts with real DB, real TLS, real middleware
                   → Tests HTTP endpoints end-to-end (Fiber app.Test or real HTTPS)

E2E Tests:        Docker Compose with PostgreSQL
                   → Full production stack (HTTPS + PostgreSQL + OTel)
                   → Not yet implemented for most services
```

### Why PostgreSQL Tests Might Matter

SQLite and PostgreSQL have subtle differences:

| Behavior | SQLite | PostgreSQL | Risk |
|----------|--------|------------|------|
| UUID type | TEXT only | Native UUID or TEXT | Low (we use TEXT everywhere) |
| JSON queries | Limited | Full JSONB support | Medium (if you use JSON queries) |
| Transaction isolation | Serializable | Read Committed default | Low (WAL mode similar) |
| Concurrent writes | Single writer + WAL | MVCC multi-writer | Medium (race conditions) |
| Case sensitivity | Platform-dependent | Always case-sensitive | Low (we use LOWER()) |
| Schema migrations | Loose DDL | Strict DDL | Low (we test migrations) |

### Option A: SQLite-Only Until E2E (Current Approach)

- Unit + integration tests use SQLite in-memory
- PostgreSQL tested only in E2E (Docker Compose)
- **Risk**: PostgreSQL-specific bugs caught late (E2E tests are slow to iterate on)
- **Benefit**: Fast tests, no Docker dependency for unit/integration

### Option B: Add PostgreSQL Testcontainer Tests for Critical Services

- Unit + integration tests STILL use SQLite (fast, no Docker)
- Add SEPARATE integration tests with `//go:build integration` that use PostgreSQL testcontainers
- Run these in CI and before major releases
- **Pattern you already have** (from testdb package):

```go
// This helper already exists in your codebase:
container, db := testdb.NewPostgresTestContainer(ctx, t)
// Takes ~3-5 seconds to start PostgreSQL in a container
// Tests run against real PostgreSQL with real driver
```

- **Risk**: Slower tests (~5s startup per test suite), Docker dependency
- **Benefit**: Catches PostgreSQL-specific issues before E2E

### Option C: PostgreSQL Testcontainers for ALL Services

- Every service gets PostgreSQL testcontainer integration tests
- Run alongside SQLite tests (both execute in CI)
- **Risk**: 10 services × ~5s = ~50s extra test time in CI
- **Benefit**: Maximum confidence in PostgreSQL compatibility

### My Recommendation

**Option B** (PostgreSQL testcontainers for critical services only). Confidence: **HIGH**.

Critical services = sm-kms, jose-ja, pki-ca (they have complex DB interactions). Simple services (identity-rp/rs/spa) don't need PostgreSQL-specific tests because their DB usage is trivial (sessions only, managed by framework).

### Answer Options

| Option | What it means |
|--------|--------------|
| **A) SQLite-only until E2E** | No PostgreSQL testcontainers. Trust E2E to catch PostgreSQL issues. |
| **B) PostgreSQL for critical services** | Add PostgreSQL testcontainer tests for sm-kms, jose-ja, pki-ca. |
| **C) PostgreSQL for ALL services** | Every service gets PostgreSQL testcontainer tests. |

**Answer**: ___

---

## Q3: Terminology — "service-template" vs "service-framework" (NEW)

### Executive Summary

You raised a concern that the name "service-template" causes LLM reasoning friction. When Copilot sees "template" it may confuse the FRAMEWORK code (`internal/apps/template/`) with the SKELETON service (`internal/apps/skeleton/template/`).

### Current Naming

| Thing | Current Name | What It Is |
|-------|-------------|-----------|
| The 64K-line framework | `internal/apps/template/service/` | ALL reusable infrastructure |
| The bare-bones starter service | `internal/apps/skeleton/template/` | Minimal service for `/new-service` |
| Import aliases | `cryptoutilAppsTemplateService*` | Used in every service |
| Config references | `ServiceTemplateServerSettings` | Embedded in every service config |

### The Friction

"Template" appears in both the FRAMEWORK (the 64K-line engine) and the SKELETON (the 19-file starter). When you say "service-template" to Copilot, it might think you mean the skeleton-template service, not the shared framework.

### Option A: Keep "service-template" (Status Quo)

- No renaming work
- Copilot adapts via instruction files and context
- Mild ongoing ambiguity

### Option B: Rename to "service-framework"

- `internal/apps/template/` → `internal/apps/framework/`
- `cryptoutilAppsTemplateService*` → `cryptoutilAppsFrameworkService*`
- `ServiceTemplateServerSettings` → `ServiceFrameworkServerSettings`
- **Scope**: ~340 files reference the framework. Renaming is a large but mechanical refactor.
- **Benefit**: Zero ambiguity. "Framework" = the engine. "Skeleton" = the starter.
- **Risk**: Large diff, merge conflicts if branches exist, import path changes across all services

### Option C: Rename to "service-engine" or "service-core"

- Same as B but with a shorter name
- `cryptoutilAppsEngineService*` or `cryptoutilAppsCoreService*`
- **Benefit**: Even more distinct from "template" and "skeleton"

### My Recommendation

**Option A (keep as-is)** for now. Confidence: **MEDIUM**.

Rationale: The renaming is a pure mechanical refactor that can happen anytime. Right now it would create a massive diff that slows down the more important framework-v3 work (TLS, builder refactoring, sequential exemption reduction). If the ambiguity causes persistent Copilot confusion, revisit after Phase 3 (builder refactoring) since that's already touching all service files.

### Answer Options

| Option | What it means |
|--------|--------------|
| **A) Keep "service-template"** | No rename. Address ambiguity via instruction files. |
| **B) Rename to "service-framework"** | Rename after Phase 3 (builder refactoring). ~340 files change. |
| **C) Rename to "service-engine"** | Same as B but with "engine" instead of "framework". |

**Answer**: ___

---

## Summary of Pending Decisions

| Decision | Status | Quizme |
|----------|--------|--------|
| D12: Skeleton-template role | **Pending Q1 above** | v4 Q1 |
| D14 supplement: Docker testing | **Pending Q2 above** | v4 Q2 |
| NEW: Terminology rename | **Pending Q3 above** | v4 Q3 |
