# Prioritized Recommendations

This document synthesizes the brainstorm into a prioritized action list.
Each recommendation includes rationale, effort estimate, and expected return.

---

## Priority 0 — Do These First (High ROI, Modest Effort)

These items unblock everything else and have the best effort-to-value ratio.

### P0-1: ServiceContract Interface

File: internal/apps/framework/service/framework/contract.go (new)

What: Define a Go interface every cryptoutil service must implement.
Why:  Compile-time enforcement of framework contracts. Free once written.
When: Before starting the next service migration (pki-ca or identity-authz).
Effort: 1-2 days
Return: Prevents architectural drift permanently. Catches missing methods immediately.

Minimum viable interface:
  ServiceID() string
  ProductID() string
  DomainMigrations() fs.FS
  RegisterPublicRoutes(app *fiber.App, res*ServiceResources) error

Extended interface (optional, for hook-capable services):
  PostStart(res *ServiceResources) error
  RegisterAdminRoutes(app*fiber.App, res *ServiceResources) error

---

### P0-2: air Live Reload Configuration

File: .air.toml (new, per service or one root file with service targets)

What: Install and configure air for hot reload during development.
Why:  2-3x faster inner development loop. Huge quality-of-life improvement.
When: Immediately. Takes 2 hours.
Effort: 2 hours
Return: 2-3x faster feedback loop during development.

Install: go install github.com/air-verse/air@latest
Config: .air.toml with cmd = 'go build -o tmp/{service} ./cmd/{service}'

---

### P0-3: Promote Skeleton to Full CRUD Reference

Files: internal/apps/skeleton/template/ (expand from 21 to ~60 files)

What: Add complete domain layer (item entity, repository, service, handlers)
      with full test coverage to skeleton-template.
Why:  Provides a clear reference for every new service. Eliminates the
      'read 316 template files to understand the framework' problem.
When: Before starting identity service migrations.
Effort: 1-2 weeks
Return: All future service migrations start from a working reference.

What to add to skeleton:
  domain/item.go + domain/item_test.go
  repository/item_repository.go + repository/item_repository_test.go
  service/item_service.go + service/item_service_test.go
  server/apis/items_handler.go + server/apis/items_handler_test.go
  client/item_client.go + client/item_client_test.go
  api/openapi_spec_paths.yaml (CRUD endpoints for Item)

---

## Priority 1 — High Value (Medium Effort, Enables Scaling)

### P1-1: cicd new-service Scaffolding Tool

Files: cmd/cicd/new_service/ (new), skeleton/**/*.tmpl (new templates)

What: CLI tool that generates a complete new service from skeleton templates.
Why:  Reduces service migration from 3 weeks to 3 days.
When: After skeleton is promoted to full CRUD reference (P0-3).
Effort: 2-3 weeks
Return: 3-5x faster per service migration. 10 services * 3 days = 30 days total.

Command: go run ./cmd/cicd new-service --product sm --service newname --entity Item
Output: internal/apps/sm/newname/ with full service skeleton that compiles and tests pass.

Implementation approach:
1. Convert skeleton files to text/template templates (*.tmpl)
2. Define ScaffoldContext struct with product, service, entity variables
3. Render templates with context into output directory
4. Run go build to verify output compiles
5. Run go test ./... to verify tests pass

---

### P1-2: Cross-Service Contract Test Suite

Files: internal/framework/contract_test.go (new)

What: Test suite run against all services to verify consistent behavior.
Why:  Divergence caught immediately in CI rather than discovered at runtime.
When: After ServiceContract interface exists (P0-1).
Effort: 1 week
Return: Catches behavioral regressions across all services. Documents expected behavior.

Tests to include:
- /admin/api/v1/livez returns 200
- /admin/api/v1/readyz returns 200 when DB ready, 503 when not
- /admin/api/v1/shutdown drains and stops
- /service/api/v1/* rejects unauthenticated requests with 401
- /browser/api/v1/* rejects cross-origin requests appropriately
- Error responses always contain code + message fields
- All services include trace_id in error responses

---

### P1-3: cicd diff-skeleton Conformance Tool

Files: cmd/cicd/diff_skeleton/ (new)

What: Tool that shows structural differences between a service and skeleton.
Why:  Prevents drift from compounding. Surfaces issues before they are expensive.
When: After skeleton is promoted (P0-3).
Effort: 1 week
Return: Services stay conformant. Maintenance overhead reduced by 30%+.

Command: go run ./cmd/cicd diff-skeleton --service sm-im
Command: go run ./cmd/cicd diff-skeleton --all-services

Output shows:
- Files present in skeleton but missing from service (possibly required)
- Pattern drift (e.g., session middleware not applied to route group)
- Outdated API usage (e.g., old WithBarrier() vs new BarrierModule usage)

---

### P1-4: Shared Test Infrastructure Package

Files: internal/framework/testing/ (new) or expand internal/apps/skeleton/testing/

What: Common test utilities shared by all 10 services.
Why:  TestMain duplication is ~50 lines per service = 500 lines of identical code.
      Each TestMain pattern change requires updating 10 copies.
When: After skeleton reference implementation is stable.
Effort: 1 week
Return: 2x test writing speed. Pattern changes propagate automatically.

Package includes:
- NewInMemorySQLiteDB() for unit tests
- NewPostgresTestContainer() for integration tests
- StandardHealthClient for contract tests
- StandardFixtures{} for test data creation (tenant, realm, user)
- StandardAssertions for HTTP response validation

---

## Priority 2 — Strategic Improvements (Medium-High Effort)

### P2-1: Service Manifest Declaration

Files: internal/apps/framework/service/framework/manifest.go (new)

What: ServiceManifest struct that declares service capabilities.
      Framework reads it and auto-configures accordingly.
Why:  Eliminates manual WithBarrier(), WithSessions(), WithJWTAuth() calls.
      Framework makes the right decisions based on declared capabilities.
When: After ServiceContract interface is stable.
Effort: 2-3 weeks (framework changes + update all services)
Return: Reduces server.go boilerplate by 50%. Clearer intent per service.

### P2-2: Architecture Fitness Functions

Files: internal/arch/ (new), .github/workflows/ci-fitness.yml (new)

What: Automated tests that verify architectural properties never regress.
Why:  ARCHITECTURE.md documents constraints. Fitness functions enforce them.
When: After basic tooling is in place.
Effort: 2-3 weeks
Return: Architectural constraints enforced automatically in CI. Documentation stays accurate.

Initial fitness functions (see 07-fitness-functions.md for full list):
- No cross-service imports
- All services have required directory structure
- No file exceeds 500 lines
- All services have health endpoints
- All TLS configs use minimum TLS 1.3
- Migration version ranges are correct

### P2-3: OpenAPI-to-Repository Code Generation

Files: cmd/cicd/generate/ (new), templates for repository + service layers

What: Generate repository CRUD and service stubs from OpenAPI spec.
Why:  Reducing handwritten code reduces bugs and speeds iteration.
When: After scaffolding tool (P1-1) is stable.
Effort: 3-4 weeks
Return: Each new entity takes hours instead of days to implement fully.

---

## Priority 3 — Big Bets (High Effort, High Payoff)

### P3-1: Module System (fx or Wire)

What: Replace monolithic ServerBuilder options with composable modules.
      Each module (Barrier, Sessions, OpenAPI, etc.) is independently testable.
Why:  Enables mix-and-match capabilities. Makes framework more discoverable.
      Identity services (OAuth 2.1, OIDC) need different modules than SM/JOSE.
When: After all 10 services are migrated. Framework stability required first.
Effort: 4-6 weeks (framework rewrite + update all services)
Return: Huge long-term maintainability. Enables true plug-and-play service assembly.

Consider: Uber fx for runtime DI, Google Wire for compile-time DI.
fx.Module maps almost perfectly to the cryptoutil module concept.

### P3-2: Extract Framework as Separate Go Module

What: Extract internal/apps/framework/service/ to its own versioned Go module.
Why:  Clean contract boundary. Versioned evolution. Framework can be reused.
When: After Module System is implemented and stable.
Effort: 4-6 weeks
Return: Framework becomes independently testable and publishable.
        Other developers learning cryptography can use the framework.

---

## Summary Table

| ID   | Name                           | Effort   | Return | Priority |
|------|--------------------------------|----------|--------|----------|
| P0-1 | ServiceContract interface      | 2 days   | 10x    | Now      |
| P0-2 | air live reload                | 2 hours  | 3x     | Now      |
| P0-3 | Skeleton full CRUD reference   | 2 weeks  | 10x    | Now      |
| P1-1 | cicd new-service scaffolding   | 3 weeks  | 10x    | Soon     |
| P1-2 | Cross-service contract tests   | 1 week   | 5x     | Soon     |
| P1-3 | cicd diff-skeleton tool        | 1 week   | 5x     | Soon     |
| P1-4 | Shared test infrastructure     | 1 week   | 5x     | Soon     |
| P2-1 | Service Manifest declaration   | 3 weeks  | 5x     | Later    |
| P2-2 | Architecture fitness functions | 3 weeks  | 3x     | Later    |
| P2-3 | OpenAPI-to-repository codegen  | 4 weeks  | 10x    | Later    |
| P3-1 | Module system (fx/Wire)        | 6 weeks  | 5x     | Future   |
| P3-2 | Extract framework module       | 6 weeks  | 3x     | Future   |

---

## First 30 Days Action Plan

Week 1:
- Install and configure air live reload for sm-im and jose-ja
- Write ServiceContract interface (draft)
- Make skeleton compile with ServiceContract
- Review if sm-im, jose-ja, sm-kms satisfy ServiceContract

Week 2:
- Add full CRUD domain layer to skeleton-template (Item entity)
- Write skeleton OpenAPI spec for Item CRUD
- Start shared test infrastructure package

Week 3:
- Complete shared test infrastructure
- Write first 5 cross-service contract tests (health endpoints)
- Begin cicd diff-skeleton tool design

Week 4:
- Complete cicd diff-skeleton tool
- Run diff-skeleton on sm-im, jose-ja, sm-kms
- Fix identified divergence
- Begin cicd new-service scaffolding tool

End of Day 30: Foundation is set for 10x migration velocity for remaining 7 services.
