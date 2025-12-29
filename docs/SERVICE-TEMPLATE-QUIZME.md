# Service Template Refactoring - Clarification Questions

**Date**: 2025-12-29
**Context**: Technical debt cleanup discovered during Phase 10-12 implementation

---

## Critical Issues Identified

The following technical debt and pattern violations were discovered:

1. **CGO Dependency**: go.mod contains `github.com/mattn/go-sqlite3` (CGO-dependent) instead of `modernc.org/sqlite`
2. **Missing Import Aliases**: 20+ imports of `cryptoutil/internal/learn/*` without importas aliases
3. **Hardcoded Secrets**: Multiple test files contain hardcoded JWT secrets
4. **Hardcoded Constants**: Test wait times, intervals outside magic package
5. **Magic Location**: `internal/learn/magic/` should be `internal/shared/magic/magic_learn.go`
6. **Duplicated Migrations**: ApplyMigrations duplicated across 3 services
7. **Hardcoded Validation**: Username/password rules hardcoded instead of realm-based
8. **Missing TestMain**: E2E and server tests create servers per-test instead of TestMain pattern
9. **Test User Pattern**: Missing fixed prefix + randomized suffix for concurrent test safety
10. **ServiceTemplate Incomplete**: Server initialization not using full template pattern

---

## Q1: ServiceTemplate Shared Infrastructure (CRITICAL - Blocks Other Services)

**Current State**:

- Each service duplicates: sqlDB setup, migrations, telemetry init, JWK services, TLS config
- learn-im duplicates 200+ lines of infrastructure code from kms reference

**Proposed Pattern**:

```go
type ServiceTemplate struct {
    sqlDB            *sql.DB
    gormDB           *gorm.DB
    telemetryService *TelemetryService
    jwkGenService    *JWKGenService
    publicTLSCfg     *tls.Config
    adminTLSCfg      *tls.Config
    migrationsApplied bool
    // Reusable HTTP server infrastructure
}

type LearnIMServer struct {
    template *ServiceTemplate  // Embedded template
    // Service-specific repos
    userRepo         UserRepository
    messageRepo      MessageRepository
}
```

**Options**:

- **A**: ServiceTemplate contains ALL infrastructure (DB, telemetry, crypto, TLS, migrations)
- **B**: ServiceTemplate contains ONLY network (TLS, ports) - services handle DB/crypto
- **C**: Two-tier: BaseTemplate (network) + DataTemplate (DB/migrations)
- **D**: Keep current duplication pattern
- **E**: Other: ___________

**Recommended**: **A** (Maximum reuse, consistent patterns, faster development)

A

---

## Q2: Magic Values Consolidation

**Current State**:

- `internal/learn/magic/magic.go` contains username/password lengths, JWT config
- Pattern inconsistent with identity/kms using `internal/shared/magic/magic_*.go`

**Options**:

- **A**: Move ALL to `internal/shared/magic/magic_learn.go` (centralized, auditable)
- **B**: Keep service-specific in `internal/<service>/magic/` (service isolation)
- **C**: Hybrid: Common in shared, unique in service
- **D**: Eliminate service magic packages
- **E**: Other: ___________

**Recommended**: **A** (Consistency with existing services)

A

---

## Q3: Realm Configuration for Validation Rules

**Current Violation** in `auth_handlers.go`:

```go
if len(username) < 3 || len(username) > 50 {
    return nil, ErrInvalidUsername
}
```

**Proposed Realm YAML**:

```yaml
realms:
  - name: username-password-file
    validation:
      username:
        min: 8
        max: 64
        pattern: "^[a-zA-Z0-9_-]+$"
      password:
        min: 12
        max: 64
        complexity: ["uppercase", "lowercase", "digit", "special"]
```

**Options**:

- **A**: Full realm config with validation section (flexible, enterprise-ready)
- **B**: Hardcoded defaults + optional realm override (backward compatible)
- **C**: Keep hardcoded (validation rules rarely change)
- **D**: Use ServerSettings fields (global, not realm-specific)
- **E**: Other: ___________

**Recommended**: **A** (Enterprise deployments need per-realm policies)

A

---

## Q4: Migrations Code Extraction (DRY Principle)

**Current Duplication**:

- `internal/learn/repository/migrations.go` - 80 lines
- `internal/identity/repository/migrations.go` - ~80 lines
- `internal/kms/server/repository/sqlrepository/sql_migrations.go` - ~80 lines

**Pattern**:

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS

func ApplyMigrations(db *sql.DB, dbType DatabaseType) error { ... }
```

**Options**:

- **A**: Extract to `internal/template/server/migrations.go`, services pass embed.FS
- **B**: Extract to `internal/shared/database/migrations.go` as utility
- **C**: Keep duplicated (migration logic may diverge)
- **D**: Builder pattern: `NewMigrationRunner(embedFS).Apply(db)`
- **E**: Other: ___________

**Recommended**: **A** (Template pattern, service autonomy for migration files)

A

---

## Q5: Import Alias Enforcement via cicd Subcommand

**Current Violations** (20+ instances):

```go
import "cryptoutil/internal/learn/repository"  // ❌ Should be aliased
import cryptoutilRepository "cryptoutil/internal/learn/repository"  // ✅ Correct
```

**Proposed cicd Subcommand**:

```bash
go run ./cmd/cicd go-check-importas
# Checks ALL cryptoutil/internal/* imports have aliases per .golangci.yml
```

**Options**:

- **A**: ALL `cryptoutil/internal/*` imports require aliases
- **B**: Only `cryptoutil/internal/<service>/*` imports need aliases
- **C**: Only cross-service imports need aliases
- **D**: No enforcement
- **E**: Other: ___________

**Recommended**: **A** (Consistency, refactoring safety, follows shared pattern)

A

---

## Q6: TestMain Pattern Implementation Priority

**Current Violations**:

- `internal/learn/e2e/*_test.go` - creates server per test (slow, fragile)
- `internal/learn/server/*_test.go` - creates server per test
- `internal/template/server/test_main_test.go` - missing TestMain infrastructure

**TestMain Benefits**:

- Faster: Server created once, reused across all tests
- Robust: Validates real concurrency patterns
- Efficient: Fewer ports/connections

**Options**:

- **A**: template/server → learn/e2e → learn/server (foundation first)
- **B**: learn/e2e → learn/server → template/server (immediate value)
- **C**: Fix all simultaneously
- **D**: Template only, services opt-in
- **E**: Other: ___________

**Recommended**: **A** (Template ensures pattern correctness before service adoption)

A

---

## Q7: Test Secrets Generation Pattern

**Current Violations**:

```go
const testJWTSecret = "learn-im-test-secret-e2e"  // ❌ SAST warning
const jwtSecret = "learn-im-dev-secret-change-in-production"  // ❌ SAST warning
```

**Options**:

- **A**: UUIDv7: `googleUuid.NewV7().String()`
- **B**: Prefix + UUID: `"test-jwt-" + googleUuid.NewV7().String()`
- **C**: Magic constant: `cryptoutilMagic.TestJWTSecret`
- **D**: Derive from test name: `deriveSecret(t.Name())`
- **E**: Other: ___________

**Recommended**: **B** (Prefix aids debugging, UUID prevents SAST warnings)

E; use "test-jwt-" prefix with RandomString(43) suffix from internal/shared/util/random; 43 chars gives ~258 bits of entropy, exceeding NIST 256-bit recommendation

---

## Q8: Test User/Password Pattern

**Requested Pattern**:

```go
username := "test-user-" + randomSuffix
password := "Test-Pass-" + randomSuffix
```

**Benefits**: Concurrent test safety, SAST compliance, debuggable logs

**Options**:

- **A**: Prefix + UUIDv7: `"test-user-" + googleUuid.NewV7().String()`
- **B**: Prefix + randomString: `"test-user-" + randomString(8)`
- **C**: Fully random: `randomString(16)`
- **D**: Magic + test name: `cryptoutilMagic.TestUsername + t.Name()`
- **E**: Other: USE GenerateUsername, GeneratePassword, GenerateDomain, and GenerateEmailAddress from internal\shared\util\random\usernames_passwords_test.go

**Recommended**: **A** (UUID provides timestamp ordering, no collisions)

E

Update .github\instructions\03-02.testing.instructions.md to use GenerateUsername, GeneratePassword, GenerateDomain, and GenerateEmailAddress from internal\shared\util\random\usernames_passwords_test.go methods

---

## Q9: Localhost vs 127.0.0.1 Pattern

**Current**: "localhost" hardcoded ~50 times in `internal/learn/*`

**Options**:

- **A**: Create `cryptoutilMagic.Localhost = "localhost"`, enforce via linter
- **B**: Keep "localhost" hardcoded (universally understood)
- **C**: Use `cryptoutilMagic.IPv4Loopback` ("127.0.0.1") everywhere
- **D**: Service-specific: `learnMagic.Localhost`
- **E**: Other: ___________

**Recommended**: **C** (Avoids IPv4/IPv6 ambiguity, Windows Firewall compliance)

E; use HostnameLocalhost from internal\shared\magic\magic_network.go

---

## Q10: Phase Optimization Strategy

**Current SERVICE-TEMPLATE.md**: Phases 1-10 complete, 11-12 partial, 13 deferred

**Discovered Tech Debt** requires new phases for:

- CGO removal
- Import alias enforcement
- TestMain pattern migration
- Hardcoded secrets removal
- Magic values consolidation
- ServiceTemplate extraction

**Options**:

- **A**: Infrastructure first (CGO, importas, magic, migrations, template) → Service refactors
- **B**: Service-first (Complete learn-im fully) → Extract template
- **C**: Mixed (Critical violations) → Template → Services
- **D**: Defer template, complete learn-im only
- **E**: Other: ___________

**Recommended**: **A** (Prevents propagating debt to future services)

A

---

## Summary

| # | Topic | Recommended | Rationale |
|---|-------|-------------|-----------|
| Q1 | ServiceTemplate | A - Full infrastructure | Max reuse, consistency |
| Q2 | Magic consolidation | A - Centralize in shared | Follow existing pattern |
| Q3 | Realm validation | A - YAML config | Enterprise flexibility |
| Q4 | Migrations DRY | A - Template extraction | Service autonomy preserved |
| Q5 | Import aliases | A - ALL internal/* | Consistency, safety |
| Q6 | TestMain priority | A - Template first | Foundation before adoption |
| Q7 | Test secrets | B - Prefix + UUID | Debugging + SAST compliance |
| Q8 | User/pass pattern | A - Prefix + UUIDv7 | Timestamp ordering |
| Q9 | Localhost handling | C - Always 127.0.0.1 | No ambiguity |
| Q10 | Phase strategy | A - Infrastructure first | Prevent debt propagation |
