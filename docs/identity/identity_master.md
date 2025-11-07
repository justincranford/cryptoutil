# IDENTITY MODULE - MASTER COORDINATION FILE
# OAuth 2.1 + OpenID Connect Identity Provider
# Multi-File Session Management

## 🎯 SESSION OVERVIEW

This identity module implementation is structured across multiple focused files for optimal Copilot Chat sessions.

**Total Tasks**: 8
**Estimated Timeline**: ~3-4 hours total
**Architecture**: OAuth 2.1 AuthZ Server + OIDC IdP Server + Resource Server

## 📋 SESSION WORKFLOW

1. **Start with Task 1** (`01_foundation_setup.md`)
2. **Complete each task** using its dedicated file
3. **Commit after each task** with semantic message
4. **Reference this master file** for overall coordination
5. **Use regrouping points** to plan next steps

## 📁 FILE STRUCTURE

```text
workflow-reports/identity/
├── identity_master.md              # This coordination file
├── 01_foundation_setup.md          # Task 1: Domain models & config
├── 02_storage_interfaces.md        # Task 2: Database abstractions
├── 03_token_operations.md          # Task 3: JWT issuance/validation
├── 04_authz_server_core.md        # Task 4: OAuth 2.1 endpoints
├── 05_oidc_identity_provider.md    # Task 5: OIDC user auth
├── 06_http_servers_apis.md         # Task 6: Fiber servers & APIs
├── 07_spa_relying_party.md         # Task 7: React/Vue SPA client
└── 08_integration_testing.md       # Task 8: E2E & integration tests
```

## 🚀 TASK STATUS OVERVIEW

| Task | File | Status | Time | Description |
|------|------|--------|------|-------------|
| 1 | `01_foundation_setup.md` | status:ready | 15 min | Domain models, errors, basic config |
| 2 | `02_storage_interfaces.md` | status:pending | 20 min | Database interfaces & implementations |
| 3 | `03_token_operations.md` | status:pending | 25 min | JWT token operations with cryptoutil |
| 4 | `04_authz_server_core.md` | status:pending | 30 min | OAuth 2.1 server with PKCE |
| 5 | `05_oidc_identity_provider.md` | status:pending | 30 min | OIDC user authentication |
| 6 | `06_http_servers_apis.md` | status:pending | 35 min | HTTP servers, CLI clients, admin APIs |
| 7 | `07_spa_relying_party.md` | status:pending | 20 min | SPA relying party application |
| 8 | `08_integration_testing.md` | status:pending | 40 min | Complete spec coverage testing |

```text
workflow-reports/identity/
├── identity_master.md              # This coordination file
├── 01_foundation_setup.md          # Task 1: Domain models & config
├── 02_storage_interfaces.md        # Task 2: Database abstractions
├── 03_token_operations.md          # Task 3: JWT issuance/validation
├── 04_authz_server_core.md        # Task 4: OAuth 2.1 endpoints
├── 05_oidc_identity_provider.md    # Task 5: OIDC user auth
├── 06_http_servers_apis.md         # Task 6: Fiber servers & APIs
├── 07_spa_relying_party.md         # Task 7: React/Vue SPA client
└── 08_integration_testing.md       # Task 8: E2E & integration tests
```

## 🚀 TASK STATUS OVERVIEW

| Task | File | Status | Time | Description |
|------|------|--------|------|-------------|
| 1 | `01_foundation_setup.md` | status:ready | 15 min | Domain models, errors, basic config |
| 2 | `02_storage_interfaces.md` | status:pending | 20 min | Database interfaces & implementations |
| 3 | `03_token_operations.md` | status:pending | 25 min | JWT token operations with cryptoutil |
| 4 | `04_authz_server_core.md` | status:pending | 30 min | OAuth 2.1 server with PKCE |
| 5 | `05_oidc_identity_provider.md` | status:pending | 30 min | OIDC user authentication |
| 6 | `06_http_servers_apis.md` | status:pending | 35 min | HTTP servers, CLI clients, admin APIs |
| 7 | `07_spa_relying_party.md` | status:pending | 20 min | SPA relying party application |
| 8 | `08_integration_testing.md` | status:pending | 40 min | Complete spec coverage testing |

```

## 📋 GLOBAL CONSTRAINTS (Reference for all tasks)

### Dependencies

**ONLY use existing go.mod dependencies:**

- `lestrrat-go/jwx/v3` - JWT operations
- `google/uuid` - UUID generation
- `gofiber/fiber/v2` - HTTP server
- `gorm.io/gorm` - Database ORM
- `github.com/stretchr/testify` - Testing

### Code Reuse

**ONLY allowed from:**

- `internal/common/container`
- `internal/common/crypto`
- `internal/common/pool`
- `internal/common/telemetry`
- `internal/common/util`

### Magic Values

**ALL constants** → `/internal/identity/<package>/magic*.go` files (linting configuration updated to allow this)

### Import Aliases

**CRITICAL**: When creating identity packages, add corresponding import aliases to `.golangci.yml`:

```yaml
# Add to .golangci.yml importas section:
- pkg: cryptoutil/internal/identity/authz
  alias: cryptoutilIdentityAuthz
- pkg: cryptoutil/internal/identity/idp
  alias: cryptoutilIdentityIdp
- pkg: cryptoutil/internal/identity/rs
  alias: cryptoutilIdentityRs
- pkg: cryptoutil/internal/identity/common
  alias: cryptoutilIdentityCommon
```

### Architecture

- **AuthZ Server**: OAuth 2.1 authorization server
- **IdP Server**: OpenID Connect identity provider
- **RS Server**: Resource server with token validation
- **All services**: Independently deployable with HTTP/HTTPS

## 📋 IMPLEMENTATION NOTES

### Critical Constraints

- **DEPENDENCY CONSTRAINT**: Use ONLY existing go.mod dependencies - NO new 3rd-party libraries
- **CONFIGURATION**: All authz, idp, rs implementations must be fully parameterized with NO hardcoded defaults
- **TOKEN FORMATS**: Support JWS, JWE, and UUID access tokens using existing jwx and uuid libraries
- **OPENAPI**: Generate separate component-based specs for authz, idp, and rs services
- **SERVERS**: Each service (authz, idp, rs) must support independent startup as GoFiber servers with optional HTTP/HTTPS
- **MAGIC VALUES**: ALL constant magic values MUST go in `/internal/identity/<PACKAGE>/magic*.go` files
- **CLI/AGENT CLIENTS**: Working CLI clients and agent clients required for each of authz, idp, rs
- **ADMIN APIs**: Authz and idp must offer admin APIs for model CRUD
- **SPA RELYING PARTY**: Working SPA relying party is required

### Implementation Guidance

- **OAUTH2/OIDC COVERAGE**: Complete OAuth 2.1 and OIDC specifications must be covered in design, implementation, and ESPECIALLY integration/E2E tests
- **SERVICE INITIALIZATION**: Study `ServerApplicationBasic` pattern for `TelemetryService` and `JWKGenService` initialization
- **DECOUPLING**: `/internal/identity/` must be maximally decoupled for future extraction into separate repository

## 🎯 ACCEPTANCE CRITERIA SUMMARY

### Functional Requirements

- ✅ Complete OAuth 2.1 authorization code flow with PKCE
- ✅ Client credentials grant type, refresh token flow
- ✅ OIDC ID tokens and userinfo endpoint
- ✅ Token introspection and revocation
- ✅ Secure client registration and management
- ✅ User authentication and profile management
- ✅ Resource server with token validation
- ✅ Independent server startup for authz, idp, rs services
- ✅ Working CLI clients and agent clients for each service
- ✅ Admin APIs for authz and idp model CRUD operations
- ✅ Working SPA relying party application
- ✅ Complete OAuth 2.1 and OIDC specification coverage

### Security Requirements

- ✅ JWT token signing with RSA keys
- ✅ Secure password hashing
- ✅ Rate limiting and abuse prevention
- ✅ Audit logging for security events
- ✅ Input validation and sanitization
- ✅ Secure session management
- ✅ Magic values properly isolated in dedicated files
- ✅ Maximum decoupling for microservice extraction

### Quality Requirements

- ✅ 95%+ code coverage with comprehensive tests for each task
- ✅ Parameterized unit tests covering happy and sad paths
- ✅ Edge case coverage and table-driven testing
- ✅ OpenAPI 3.0 component-based specifications
- ✅ Integration with existing cryptoutil patterns
- ✅ Structured logging and metrics
- ✅ Performance benchmarks meeting requirements
- ✅ Independent server startup for authz, idp, rs with HTTP/HTTPS support
- ✅ Complete OAuth 2.1 and OIDC specification coverage in integration and E2E tests
- ✅ Fully parameterized configurations (no hardcoded defaults)

### Operational Requirements

- ✅ Health check endpoints
- ✅ Graceful shutdown handling
- ✅ Configuration validation
- ✅ Database migrations
- ✅ Docker containerization ready
- ✅ Fully parameterized configurations with no hardcoded defaults

## 🔄 SESSION MANAGEMENT

### Current Session Focus

**Active Task**: None (Ready to start Task 1)
**Next Action**: Open `01_foundation_setup.md` and begin implementation

### Cross-Task Validation Checkpoints

**Before starting each task (except Task 1), perform these quick validations:**

1. **Dependency Check**: Verify previous task's outputs exist and compile
2. **Integration Test**: Run basic compilation test to ensure interfaces match
3. **Import Validation**: Confirm only allowed `internal/common/*` imports are used

**Example validation commands:**

```bash
# Check previous task outputs exist
go build ./internal/identity/...

# Quick compilation test
go test -c ./internal/identity/... 2>&1 | head -20

# Import validation (should only show allowed internal/common/*)
go list -f '{{.Imports}}' ./internal/identity/... | grep -v "internal/common"
```

### Regrouping Protocol

After each task completion:

1. **Validate** completion criteria are met
2. **Cross-Task Check**: Run validation checkpoint for next task readiness
3. **Commit** with semantic message: `feat: complete Task X - [brief description]`
4. **Update** this master file with new status
5. **Plan** next task and open its file
6. **Reference** constraints section as needed

### Rollback Procedures

**If a task needs significant rework:**

1. **Assess Impact**: Check which downstream tasks are affected
2. **Stash Changes**: `git stash` current work if needed
3. **Reset to Last Good**: `git reset --hard HEAD~X` (where X = number of commits to undo)
4. **Reapply Selectively**: `git cherry-pick` only the commits you want to keep
5. **Update Status**: Mark affected tasks as `status:pending` in this file
6. **Restart**: Begin the reworked task with fresh context

**Emergency rollback (if multiple tasks affected):**

```bash
# Create backup branch
git branch backup-before-rollback

# Reset to before problematic task
git reset --hard <commit-hash-before-issue>

# Update all affected task statuses to pending
# Restart from the earliest affected task
```

### Emergency Reset

If Copilot loses context:

1. Close all files except the active task file
2. Re-read the task file's goal and constraints
3. Restart with focused prompt: "You are working on Task X from `filename`. Focus only on this task."

---

**Ready to begin? Start with Task 1: Open `01_foundation_setup.md`**
 
 
 
 
 
 
 
 
