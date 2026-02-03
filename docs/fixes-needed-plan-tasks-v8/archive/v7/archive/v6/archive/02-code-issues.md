# Code Issues - Test Conformance Analysis

## Analysis Scope

Services analyzed:
- `internal/apps/template/service/` (service-template)
- `internal/apps/cipher/im/` (cipher-im)
- `internal/apps/jose/ja/` (jose-ja)
- `internal/kms/` (sm-kms)

## Summary

| Category | Status | Count |
|----------|--------|-------|
| Integration tests not using config-based server | ✅ Compliant | 0 |
| E2E tests not using docker compose | ✅ Compliant | 0 |
| Unit tests not table-driven | ⚠️ Non-conformant | 15+ files |
| Integration tests not table-driven | ⚠️ Non-conformant | 8+ files |
| E2E tests not table-driven | ✅ Compliant | 0 |

---

## 1. Integration Tests Not Using Server Instance via Configuration

**Status: ✅ COMPLIANT**

All services correctly use TestMain pattern with configuration-based server creation:

- **cipher-im**: `internal/apps/cipher/im/e2e/testmain_e2e_test.go` - Uses `ComposeManager` for Docker Compose lifecycle
- **identity**: `internal/apps/identity/e2e/testmain_e2e_test.go` - Uses `ComposeManager` for Docker Compose lifecycle
- **jose-ja**: `internal/apps/jose/ja/server/testmain_test.go` - Uses TestMain with configuration
- **jose-ja repository**: `internal/apps/jose/ja/repository/testmain_test.go` - Uses TestMain with shared GORM DB
- **sm-kms**: Uses TestMain pattern across `internal/kms/server/` tests

**Reference Implementation** (cipher-im E2E):
```go
func TestMain(m *testing.M) {
    ctx := context.Background()
    composeManager = cryptoutilAppsTemplateTestingE2e.NewComposeManager(cryptoutilSharedMagic.CipherE2EComposeFile)
    sharedHTTPClient = cryptoutilSharedCryptoTls.NewClientForTest()
    if err := composeManager.Start(ctx); err != nil { ... }
    if err := composeManager.WaitForMultipleServices(healthChecks, cryptoutilSharedMagic.CipherE2EHealthTimeout); err != nil { ... }
    exitCode := m.Run()
    _ = composeManager.Stop(ctx)
    os.Exit(exitCode)
}
```

---

## 2. E2E Tests Not Using Docker Compose

**Status: ✅ COMPLIANT**

All E2E tests properly use Docker Compose via `ComposeManager`:

| Service | E2E Test File | Compose File |
|---------|--------------|--------------|
| cipher-im | `internal/apps/cipher/im/e2e/testmain_e2e_test.go` | `deployments/cipher/compose.yml` |
| identity | `internal/apps/identity/e2e/testmain_e2e_test.go` | `deployments/identity/compose.yml` |

E2E tests use `cryptoutilAppsTemplateTestingE2e.ComposeManager` which properly:
1. Starts docker compose stack
2. Waits for health checks on all services
3. Runs tests
4. Stops docker compose stack on cleanup

---

## 3. Unit Tests Not Using Table-Driven Structure

**Status: ⚠️ NON-CONFORMANT**

### service-template

| File | Issue | Standalone Functions |
|------|-------|---------------------|
| `config/config_coverage_test.go` | 12 standalone functions instead of table-driven | `TestGetTLSPEMBytes_NilValue`, `TestGetTLSPEMBytes_NonBytesValue`, `TestNewForJOSEServer_DevMode`, `TestNewForJOSEServer_ProductionMode`, `TestNewForCAServer_DevMode`, `TestNewForCAServer_ProductionMode`, `TestRegisterAsBoolSetting`, `TestRegisterAsStringSetting`, `TestRegisterAsUint16Setting`, `TestRegisterAsStringSliceSetting`, `TestRegisterAsStringArraySetting`, `TestRegisterAsDurationSetting`, `TestRegisterAsIntSetting` |
| `server/domain/tenant_join_request_test.go` | 4 standalone functions | `TestTenantJoinRequest_TableName`, `TestTenantJoinRequest_StructCreation`, `TestTenantJoinRequest_StatusConstants`, `TestTenantJoinRequest_ClientIDMutuallyExclusive` |

### jose-ja

| File | Issue | Standalone Functions |
|------|-------|---------------------|
| `domain/models_test.go` | 4 standalone functions | `TestElasticJWK_TableName`, `TestMaterialJWK_TableName`, `TestAuditConfig_TableName`, `TestAuditLogEntry_TableName` |

### sm-kms

| File | Issue | Standalone Functions |
|------|-------|---------------------|
| `application/application_middleware_test.go` | 7 standalone functions testing basic auth middleware | `TestSwaggerUIBasicAuthMiddleware_NoAuthConfigured`, `TestSwaggerUIBasicAuthMiddleware_MissingAuthHeader`, `TestSwaggerUIBasicAuthMiddleware_InvalidAuthMethod`, `TestSwaggerUIBasicAuthMiddleware_InvalidBase64Encoding`, `TestSwaggerUIBasicAuthMiddleware_InvalidCredentialFormat`, `TestSwaggerUIBasicAuthMiddleware_InvalidCredentials`, `TestSwaggerUIBasicAuthMiddleware_ValidCredentials` |
| `demo/seed_test.go` | 6 standalone functions | `TestGenerateDemoTenantID`, `TestDefaultDemoTenants`, `TestDefaultDemoTenants_RegeneratesOnEachCall`, `TestDefaultDemoKeys`, `TestDemoKeyConfig_Fields`, `TestResetDemoData` |

### Recommended Refactoring Pattern

**Before (non-compliant):**
```go
func TestRegisterAsBoolSetting(t *testing.T) {
    setting := Setting{Value: true}
    result := RegisterAsBoolSetting(&setting)
    require.True(t, result)
}

func TestRegisterAsStringSetting(t *testing.T) {
    setting := Setting{Value: "test-value"}
    result := RegisterAsStringSetting(&setting)
    require.Equal(t, "test-value", result)
}
```

**After (compliant):**
```go
func TestRegisterAsSettings(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name     string
        setting  Setting
        register func(*Setting) interface{}
        want     interface{}
    }{
        {"bool setting", Setting{Value: true}, func(s *Setting) interface{} { return RegisterAsBoolSetting(s) }, true},
        {"string setting", Setting{Value: "test-value"}, func(s *Setting) interface{} { return RegisterAsStringSetting(s) }, "test-value"},
        // ... more cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            result := tt.register(&tt.setting)
            require.Equal(t, tt.want, result)
        })
    }
}
```

---

## 4. Integration Tests Not Using Table-Driven Structure

**Status: ⚠️ NON-CONFORMANT**

### jose-ja

| File | Issue | Standalone Functions |
|------|-------|---------------------|
| `repository/material_jwk_repository_test.go` | 9 standalone functions | `TestMaterialJWKRepository_Create`, `TestMaterialJWKRepository_GetByMaterialKID`, `TestMaterialJWKRepository_GetByID`, `TestMaterialJWKRepository_GetActiveMaterial`, `TestMaterialJWKRepository_ListByElasticJWK`, `TestMaterialJWKRepository_RotateMaterial`, `TestMaterialJWKRepository_RetireMaterial`, `TestMaterialJWKRepository_Delete`, `TestMaterialJWKRepository_CountMaterials` |
| `repository/additional_edge_cases_test.go` | 10+ standalone functions | `TestElasticJWKRepository_GetByIDWithInvalidID`, `TestElasticJWKRepository_GetWithSpecialCharactersInKID`, `TestElasticJWKRepository_ListWithEmptyDatabase`, `TestElasticJWKRepository_UpdateNonExistentJWK`, `TestElasticJWKRepository_DeleteAlreadyDeleted`, `TestMaterialJWKRepository_GetByIDEdgeCases`, `TestMaterialJWKRepository_GetByMaterialKIDWithSpecialChars`, `TestMaterialJWKRepository_GetActiveMaterialWhenNoneActive` |

### service-template

| File | Issue | Standalone Functions |
|------|-------|---------------------|
| `server/apis/registration_routes_test.go` | 3 standalone functions | `TestRegisterRegistrationRoutes_Integration`, `TestRegisterRegistrationRoutes_RateLimiting`, `TestRegisterJoinRequestManagementRoutes_Integration` |

### sm-kms

| File | Issue | Standalone Functions |
|------|-------|---------------------|
| `repository/sqlrepository/sql_provider_edge_cases_test.go` | 5 standalone functions | `TestNewSQLRepository_NilTelemetryService`, `TestNewSQLRepository_NilSettings`, `TestNewSQLRepository_ContainerModeInvalid`, `TestHealthCheck`, `TestHealthCheck_AfterShutdown` |

---

## 5. E2E Tests Not Using Table-Driven Structure

**Status: ✅ COMPLIANT**

E2E tests properly use table-driven patterns:

**Example from `cipher-im/e2e/e2e_test.go`:**
```go
func TestE2E_HealthChecks(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name      string
        publicURL string
    }{
        {sqliteContainer, sqlitePublicURL},
        {postgres1Container, postgres1PublicURL},
        {postgres2Container, postgres2PublicURL},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // ... test logic
        })
    }
}
```

---

## Positive Findings

### Correct Patterns in Use

1. **app.Test() for handler testing** - All handler tests use Fiber's in-memory `app.Test()` rather than real HTTP listeners
2. **TestMain for heavyweight resources** - Integration tests use TestMain to initialize shared database connections once
3. **t.Parallel() usage** - Most tests correctly call `t.Parallel()` for concurrent execution
4. **UUIDv7 for test data** - Tests use `cryptoutilSharedUtilRandom.GenerateUUIDv7()` for unique test data

---

## Action Items

### Priority 1: High Impact Refactoring

| File | Current | Target | Est. LOC Reduction |
|------|---------|--------|-------------------|
| `config_coverage_test.go` | 12 functions | 1 table-driven | ~80 lines |
| `application_middleware_test.go` | 7 functions | 1 table-driven | ~100 lines |
| `material_jwk_repository_test.go` | 9 functions | 2-3 table-driven | ~200 lines |

### Priority 2: Domain Model Tests

| File | Current | Target |
|------|---------|--------|
| `tenant_join_request_test.go` | 4 functions | 1 table-driven |
| `models_test.go` | 4 functions | 1 table-driven |

### Priority 3: Edge Case Tests

| File | Current | Target |
|------|---------|--------|
| `additional_edge_cases_test.go` | 10+ functions | 2-3 table-driven by category |
| `sql_provider_edge_cases_test.go` | 5 functions | 1 table-driven |

---

## References

- [03-02.testing.instructions.md](.github/instructions/03-02.testing.instructions.md) - Table-driven test requirements
- [07-01.testmain-integration-pattern.instructions.md](.github/instructions/07-01.testmain-integration-pattern.instructions.md) - TestMain pattern for GORM databases
