# Phase 3: JOSE-JA ServerBuilder Integration

## Status: STARTING

## Objective

Wire JOSE-JA domain, repositories, and handlers into the service template architecture using ServerBuilder pattern.

## Prerequisites (Completed ✅)

- ✅ Phase 2.3: Handler tests passing (70/70)
- ✅ Phase 2.3: Repository tests passing (31/31)
- ✅ Phase 2.3: All realm_id removed from domain, migrations, repositories, handlers
- ✅ Phase 2.3: Git commit with evidence

## Phase 3 Tasks

### Task 3.1: Create Server Package Structure ⏳ NEXT

**Files to create**:
```
internal/apps/jose/ja/server/
  ├── server.go                 # Main server entrypoint
  ├── application/
  │   └── application.go        # Wires DI container
  └── config/
      └── settings.go           # Already exists
```

**server.go responsibilities**:
- NewFromConfig() factory method
- ServerBuilder DI setup
- Route registration callbacks
- Start() / Shutdown() delegation

**application.go responsibilities**:
- DI container setup (repositories, services, handlers)
- Database initialization
- Migration application
- Health check configuration

### Task 3.2: ServerBuilder DI Container

**Inject into Builder**:
- ElasticJWKRepository
- MaterialJWKRepository
- AuditConfigRepository
- AuditLogRepository
- JWKHandler (with all 9 HTTP handlers)

**Pattern**:
```go
builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

builder.WithDomainMigrations(repository.MigrationsFS, "migrations")

builder.WithPublicRouteRegistration(func(
    base *cryptoutilTemplateServer.PublicServerBase,
    res *cryptoutilTemplateBuilder.ServiceResources,
) error {
    // Create repositories
    elasticRepo := repository.NewElasticJWKRepository(res.DB)
    materialRepo := repository.NewMaterialJWKRepository(res.DB)
    auditConfigRepo := repository.NewAuditConfigRepository(res.DB)
    auditLogRepo := repository.NewAuditLogRepository(res.DB)

    // Create handler
    handler := apis.NewJWKHandler(elasticRepo, materialRepo, auditConfigRepo, auditLogRepo)

    // Register routes
    base.Router.Post("/elastic-jwks", handler.HandleCreateElasticJWK())
    base.Router.Get("/elastic-jwks/:kid", handler.HandleGetElasticJWK())
    // ... 7 more routes

    return nil
})

resources, err := builder.Build()
```

### Task 3.3: Route Registration

**Elastic JWK Routes** (4 routes):
```
POST   /service/api/v1/elastic-jwks
GET    /service/api/v1/elastic-jwks/:kid
GET    /service/api/v1/elastic-jwks
DELETE /service/api/v1/elastic-jwks/:kid
```

**Material JWK Routes** (4 routes):
```
POST   /service/api/v1/elastic-jwks/:kid/materials
GET    /service/api/v1/elastic-jwks/:kid/materials
GET    /service/api/v1/elastic-jwks/:kid/materials/active
POST   /service/api/v1/elastic-jwks/:kid/materials/rotate
```

**Utility Routes** (4 routes - not implemented yet):
```
GET    /service/api/v1/jwks
POST   /service/api/v1/sign
POST   /service/api/v1/verify
POST   /service/api/v1/encrypt
POST   /service/api/v1/decrypt
```

### Task 3.4: Integration Tests

**Create**: `internal/apps/jose/ja/server/server_integration_test.go`

**Test scenarios**:
- Server starts successfully
- All 9 routes accessible
- Database migrations applied
- Health endpoints responding (livez, readyz)
- Graceful shutdown works

**Pattern**: Use TestMain to start server once, run all integration tests, shutdown

### Task 3.5: E2E Tests

**Create**: `test/e2e/jose/ja_test.go`

**Test scenarios**:
- Full request/response cycle for each handler
- PostgreSQL backend (real database, not mocks)
- TLS enabled (HTTPS)
- Tenant context isolation
- Error responses (401, 400, 404, 500)

**Pattern**: Docker Compose with PostgreSQL + JOSE-JA service

## Evidence Requirements

- [ ] `go build ./internal/apps/jose/ja/...` succeeds
- [ ] Integration tests pass (server starts, routes work)
- [ ] E2E tests pass (full request/response cycles)
- [ ] Health checks work (livez/readyz)
- [ ] Graceful shutdown works
- [ ] Git commit with evidence

## Time Estimate

- Server structure: 30 minutes
- ServerBuilder DI: 45 minutes
- Route registration: 30 minutes
- Integration tests: 45 minutes
- E2E tests: 1 hour
- **Total**: 3-4 hours

## Next Steps After Phase 3

- Phase 4: Business logic layer (audit sampling, JWKS generation)
- Phase 5: HTTP API completeness (browser paths, OpenAPI)
- Phase 6: End-to-end testing (Docker Compose, full stack)
- Phase 7: Configuration (YAML files, secrets)
- Phase 8: Deployment (Docker, Kubernetes)
- Phase 9: Documentation (README, API docs)
