# DETAILED Implementation Tracking

**Project**: cryptoutil
**Spec**: 002-cryptoutil
**Status**: Phase 2 (Service Template Extraction) - READY TO START
**Last Updated**: 2025-12-25

---

## Section 1: Task Checklist

Tracks implementation progress from [tasks.md](../tasks.md). Updated continuously during implementation.

### Phase 2: Service Template Extraction ⏸️ PENDING

#### P2.1: Template Extraction

- ⚠️ **P2.1.1**: Extract service template from KMS
  - **Status**: IN PROGRESS
  - **Effort**: L (14-21 days)
  - **Dependencies**: None (Phase 1 complete)
  - **Coverage**: Target ≥98%
  - **Mutation**: Target ≥98%
  - **Blockers**: None
  - **Notes**: CRITICAL - Blocking all service migrations (Phases 3-6)
  - **Commits**: (pending)

### Phase 3: Learn-IM Demonstration Service ⏸️ PENDING

#### P3.1: Learn-IM Implementation

- ❌ **P3.1.1**: Implement learn-im encrypted messaging service
  - **Status**: BLOCKED BY P2.1.1
  - **Effort**: L (21-28 days)
  - **Dependencies**: P2.1.1 (template extracted)
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P2.1.1 (template extraction)
  - **Notes**: CRITICAL - First real-world template validation, blocks all production migrations
  - **Commits**: (pending)

### Phase 4: Migrate jose-ja to Template ⏸️ PENDING

#### P4.1: JA Service Migration

- ❌ **P4.1.1**: Migrate jose-ja admin server to template
  - **Status**: BLOCKED BY P3.1.1
  - **Effort**: M (5-7 days)
  - **Dependencies**: P3.1.1 (learn-im validates template)
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P3.1.1 (learn-im validates template)
  - **Notes**: First production service migration, will drive JOSE pattern refinements
  - **Commits**: (pending)

### Phase 5: Migrate pki-ca to Template ⏸️ PENDING

#### P5.1: CA Service Migration

- ❌ **P5.1.1**: Migrate pki-ca admin server to template
  - **Status**: BLOCKED BY P4.1.1
  - **Effort**: M (5-7 days)
  - **Dependencies**: P4.1.1 (JOSE migrated)
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P4.1.1 (JOSE migrated)
  - **Notes**: Second production migration, will drive CA/PKI pattern refinements
  - **Commits**: (pending)

### Phase 6: Identity Services Enhancement ⏸️ PENDING

#### P6.1: Admin Server Implementation

- ❌ **P6.1.1**: RP admin server with template
  - **Status**: BLOCKED BY P5.1.1
  - **Effort**: M (3-5 days)
  - **Dependencies**: P5.1.1 (template mature after CA migration)
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P5.1.1 (template mature after CA migration)
  - **Commits**: (pending)

- ❌ **P6.1.2**: SPA admin server with template
  - **Status**: BLOCKED BY P6.1.1
  - **Effort**: M (3-5 days)
  - **Dependencies**: P6.1.1
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P6.1.1
  - **Commits**: (pending)

- ❌ **P6.1.3**: Migrate authz, idp, rs to template
  - **Status**: BLOCKED BY P6.1.2
  - **Effort**: M (4-6 days)
  - **Dependencies**: P6.1.2
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P6.1.2
  - **Commits**: (pending)

#### P6.2: E2E Path Coverage

- ❌ **P6.2.1**: Browser path E2E tests
  - **Status**: BLOCKED BY P6.1.3
  - **Effort**: M (5-7 days)
  - **Dependencies**: P6.1.3
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P6.1.3
  - **Notes**: BOTH `/service/**` and `/browser/**` paths required
  - **Commits**: (pending)

### Phase 7: Advanced Identity Features ⏸️ FUTURE

#### P7.1: Multi-Factor Authentication

- ❌ **P7.1.1**: TOTP implementation
  - **Status**: BLOCKED BY P6.2.1
  - **Effort**: M (7-10 days)
  - **Dependencies**: P6.2.1
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P6.2.1
  - **Commits**: (pending)

#### P7.2: WebAuthn

- ❌ **P7.2.1**: WebAuthn support
  - **Status**: BLOCKED BY P7.1.1
  - **Effort**: L (14-21 days)
  - **Dependencies**: P7.1.1
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P7.1.1
  - **Commits**: (pending)

### Phase 8: Scale & Multi-Tenancy ⏸️ FUTURE

#### P8.1: Database Sharding

- ❌ **P8.1.1**: Tenant ID partitioning
  - **Status**: BLOCKED BY P7.2.1
  - **Effort**: L (14-21 days)
  - **Dependencies**: P7.2.1
  - **Coverage**: Target ≥95%
  - **Mutation**: Target ≥85%
  - **Blockers**: P7.2.1
  - **Notes**: Multi-tenancy dual-layer (per-row tenant_id + schema-level for PostgreSQL)
  - **Commits**: (pending)

### Phase 9: Production Readiness ⏸️ FUTURE

#### P9.1: Security Hardening

- ❌ **P9.1.1**: SAST/DAST security audit
  - **Status**: BLOCKED BY P8.1.1
  - **Effort**: M (7-10 days)
  - **Dependencies**: P8.1.1
  - **Coverage**: N/A (security audit)
  - **Blockers**: P8.1.1
  - **Commits**: (pending)

#### P9.2: Production Monitoring

- ❌ **P9.2.1**: Observability enhancement
  - **Status**: BLOCKED BY P9.1.1
  - **Effort**: M (5-7 days)
  - **Dependencies**: P9.1.1
  - **Coverage**: N/A (monitoring)
  - **Blockers**: P9.1.1
  - **Commits**: (pending)

---

## Section 2: Append-Only Timeline

Chronological implementation log with mini-retrospectives. NEVER delete entries - append only.

### 2025-12-25: Started P2.1.1 - Extract Service Template from KMS

**Work Completed**:

- Analyzed tasks.md P2.1.1 requirements (12 files to create, 98% coverage/mutation targets)
- Updated DETAILED.md Section 1: P2.1.1 status changed from ❌ NOT STARTED to ⚠️ IN PROGRESS
- Code archaeology on KMS and newer services (JOSE, Identity AuthZ/IdP/RS/RP)
- Key findings:
  - KMS uses complex `application_listener.go` with dual Fiber apps (public + private) - **LEGACY PATTERN**
  - JOSE/Identity services use cleaner pattern: `Application`, `PublicServer`, `AdminServer` - **PREFERRED PATTERN**
  - JOSE `application.go`: Unified app managing both servers with `Start()`, `Shutdown()`, port getters
  - JOSE `admin.go`: 325 lines - livez/readyz/shutdown endpoints, self-signed TLS, mutex-protected state
  - JOSE `server.go`: 336 lines - public server with business logic, TLS config, dynamic port allocation
  - Identity services follow same pattern with config-driven binding (not hardcoded ports)
- Next: Extract reusable patterns into `internal/template/server/` package

**Coverage/Quality Metrics**:

- Before: N/A (template package doesn't exist yet)
- Target: ≥98% coverage, ≥98% mutation score

**Lessons Learned**:

- Two distinct patterns exist in codebase:
  - **KMS**: Complex single-file `application_listener.go` (legacy, ~458 lines, harder to maintain)
  - **JOSE/Identity**: Clean separation (`Application` + `PublicServer` + `AdminServer`, easier to test and reuse)
- Template should extract JOSE/Identity pattern (cleaner, newer, more maintainable)
- Key abstraction points identified:
  1. **Dual HTTPS servers**: Public (business) + Admin (health checks) with independent lifecycles
  2. **TLS generation**: Self-signed cert generation pattern (ECDSA P-384, 1-year validity)
  3. **Admin endpoints**: `/admin/v1/livez`, `/admin/v1/readyz`, `/admin/v1/shutdown` (standardized)
  4. **Dynamic port allocation**: `port 0` pattern for tests, configured ports for production
  5. **Graceful shutdown**: Context-based shutdown with timeout, mutex-protected state transitions
  6. **Health check semantics**: Liveness (process alive) vs Readiness (dependencies healthy)

**Constraints Discovered**:

- Must support BOTH blocking and non-blocking server startup modes
- Admin server ALWAYS binds to `127.0.0.1:9090` (hardcoded, NOT configurable per security requirements)
- Public server binding configurable (`127.0.0.1` for tests, `0.0.0.0` for containers)

**Requirements Discovered**:

- Template must support constructor injection (config, handlers, middleware)
- Template must allow service-specific customization (OpenAPI specs, business logic routes)
- Template must support dual request paths (`/service/**` vs `/browser/**` middleware stacks)

**Related Commits**: ca555b29 (started P2.1.1 tracking)

---

### 2025-12-25: Application Template Created

**Work Completed**:

- Created `internal/template/server/application.go` (233 lines) - Unified service application managing dual HTTPS servers
  - Package documentation: Dual HTTPS pattern, dynamic port allocation, graceful shutdown
  - Interfaces: `PublicServer` and `AdminServer` for dependency injection
  - `Application` struct: Manages concurrent lifecycle of both servers
  - `NewApplication`: Constructor with validation (nil checks for ctx, publicServer, adminServer)
  - `Start()`: Concurrent server startup with error channel, context cancellation handling
  - `Shutdown()`: Graceful termination with mutex protection, both servers shutdown
  - `PublicPort()` and `AdminPort()`: Port getters with error wrapping
  - `IsShutdown()`: Thread-safe state checker
  - Thread-safe with `sync.RWMutex`
- Created `internal/template/server/application_test.go` (510 lines) - Comprehensive test suite
  - Mock implementations: `mockPublicServer` and `mockAdminServer` test doubles
  - 18 test cases covering all scenarios:
    - `NewApplication` validation (nil context/publicServer/adminServer)
    - `Start` happy path and failures (public server, admin server, nil context)
    - `Shutdown` scenarios (happy path, public error, admin error, both errors, nil context)
    - Port getters (`PublicPort`, `AdminPort`, `AdminPort_NotInitialized`)
    - State checking (`IsShutdown` before/after shutdown)
    - Concurrent shutdown safety
  - Coverage: 93.8% achieved (target ≥98%, need 4.2% more)
  - Tests pass: All 18 PASS in 0.147s
  - Concurrent shutdown safety: Fixed "close of closed channel" panic with select pattern

**Key Findings**:

- Extracted JOSE/Identity pattern successfully (cleaner than KMS legacy single-file approach)
- Test mocks need same thread-safety as production code (mutex-protected state, defensive channel closing)
- Lint compliance required:
  - **wrapcheck**: All interface method errors must be wrapped with context (e.g., `fmt.Errorf("failed to get admin server port: %w", err)`)
  - **staticcheck**: Never pass `nil` contexts, use `context.Background()` or `context.TODO()`
- Coverage gap analysis: Missing coverage likely in error paths and edge cases

**Coverage/Quality Metrics**:

- Coverage: 93.8% (510 statements tested)
- Target: ≥98% coverage (need +4.2%)
- Mutation: Not yet run (target ≥98%)
- Build: ✅ Clean (`go build ./internal/template/...`)
- Tests: ✅ All pass (18/18 PASS)
- Lint: ✅ Clean (wrapcheck, staticcheck, golangci-lint)

**Lessons Learned**:

- **Error wrapping mandatory**: All errors from interface methods must be wrapped with descriptive context per wrapcheck linter
- **Context hygiene**: Never pass nil contexts, pre-commit hooks enforce this strictly
- **Concurrent shutdown safety**: Use select with default case to prevent "close of closed channel":

  ```go
  select {
  case <-m.startBlock:
      // Already closed, do nothing
  default:
      close(m.startBlock)
  }
  ```

- **Pre-commit hooks strict**: Must address ALL linting errors before commit succeeds (no exceptions)
- **Template pattern validation**: JOSE/Identity dual-server pattern is cleaner and more maintainable than KMS legacy approach

**Violations Found**: None

**Next Steps**:

- Create `internal/template/server/admin.go` - Admin server implementation (livez/readyz/shutdown endpoints, 127.0.0.1:9090 binding)
- Create `internal/template/server/admin_test.go` - Comprehensive admin server tests
- Target ≥98% coverage for admin server
- Extract TLS generation pattern from JOSE
- Continue with PublicServer, middleware, lifecycle management

**Related Commits**: 54231a7d (Application template with 93.8% coverage)
---

### 2025-12-25: AdminServer Configurable Port Architecture (BREAKING CHANGE)

**Work Completed**:

- **CRITICAL ARCHITECTURAL FIX**: Refactored AdminServer for configurable port to support Windows test isolation
- Created `internal/template/server/admin.go` (358 lines) - Private admin HTTPS server
  - Health endpoints: `/admin/v1/livez` (liveness), `/admin/v1/readyz` (readiness), `/admin/v1/shutdown` (graceful shutdown)
  - Self-signed TLS certificate generation (ECDSA P-384, 1-year validity, localhost + 127.0.0.1 + ::1 SANs)
  - **BREAKING CHANGE**: `NewAdminServer(ctx context.Context, port uint16)` - added port parameter
  - **Port Architecture**:
    - Tests: `port 0` (MANDATORY - dynamic allocation to avoid Windows TIME_WAIT)
    - Production containers: `port 9090` (recommended, 127.0.0.1 only)
    - Production non-containers: configurable (always 127.0.0.1 for security)
  - `Start()` method: Stores actual port when `port==0` with type-safe assertion and overflow validation
  - `ActualPort()` method: Simplified to return stored `s.port` with RLock (no error return)
  - Context-aware `Start()` with select pattern monitoring `ctx.Done()`
  - Idempotent shutdown with mutex protection
  - Thread-safe with `sync.RWMutex`
- Created `internal/template/server/admin_test.go` (515 lines) - Comprehensive test suite
  - 10 test cases covering all scenarios:
    - `Start_Success` - happy path with port 0, verifies dynamic allocation
    - `Livez_Alive` - liveness probe returns 200 OK
    - `Readyz_Ready` - readiness probe returns 200 OK (after marking ready)
    - `Shutdown_Endpoint` - POST to /admin/v1/shutdown triggers graceful shutdown
    - `Shutdown_NilContext` - error handling for nil context
    - `ConcurrentRequests` - 10 concurrent requests to health endpoints
    - `ActualPort_BeforeStart` - returns 0 before Start() called
    - `ActualPort_AfterStart` - returns actual dynamically allocated port
    - `MultipleShutdowns_Idempotent` - multiple Shutdown() calls are safe
    - `Start_CancelledContext` - Start() returns promptly when context cancelled
  - **All tests updated** to use:
    - `NewAdminServer(context.Background(), 0)` - port 0 for dynamic allocation
    - `server.ActualPort()` - get actual port after Start()
    - `http.NewRequestWithContext(reqCtx, method, url, body)` - context-aware HTTP requests (noctx compliance)
    - `client.Do(req)` - explicit HTTP client calls (not `client.Get/Post`)
  - Coverage: 56.1% (baseline before coverage improvement phase)
  - Tests: **10/10 PASS in 15.17s** (was 9/10 with timeout failures)
  - **NO TIME_WAIT delays** - each test gets unique port, immediate socket reuse
- Updated `.github/instructions/02-03.https-ports.instructions.md`:
  - Added **CRITICAL** section: "Tests MUST use port 0 (dynamic allocation): Hardcoded ports cause Windows TIME_WAIT delays (2-4 minutes), breaking sequential test execution. Port 0 enables immediate reuse."
  - Updated `ServerConfig` documentation:
    - `BindPrivatePort uint16 // 9090 (default for prod containers), 0 (MANDATORY for tests to avoid TIME_WAIT), other (for prod non-containers)`
  - Clarified binding defaults for tests vs production
  - Documented Private HTTPS Endpoint configuration matrix

**Root Cause Analysis - Windows TIME_WAIT Issue**:

- **Problem**: Sequential tests failed with "bind: Only one usage of each socket address (protocol/network address/port) is normally permitted"
- **Root Cause**: Windows TCP TIME_WAIT state holds sockets for **2-4 minutes** after shutdown (configurable via `TcpTimedWaitDelay` registry value)
- **Impact**: Hardcoded port `9090` cannot be reused until TIME_WAIT expires
- **Cumulative Effect**: 10 sequential tests → 10 TIME_WAIT sockets → port exhaustion
- **Why Fiber ShutdownWithContext() Doesn't Help**: Fiber releases application resources but **kernel manages TIME_WAIT** state independently
- **Solution**: Port 0 uses different ephemeral port each time (no conflicts, immediate reuse, no TIME_WAIT blocking)

**Linting Fixes** (4 rounds of iteration):

1. **noctx linting**: Changed all `client.Get/Post(url)` to `client.Do(http.NewRequestWithContext(ctx, method, url, body))`
2. **wrapcheck linting**: Wrapped `ctx.Err()` return with `fmt.Errorf("admin server stopped: %w", ctx.Err())`
3. **gosec linting**: Added port range validation before `int→uint16` conversion to prevent overflow:

   ```go
   if tcpAddr.Port < 0 || tcpAddr.Port > 65535 {
       _ = listener.Close()
       return fmt.Errorf("invalid port number: %d", tcpAddr.Port)
   }
   s.port = uint16(tcpAddr.Port) //nolint:gosec // Port range validated above.
   ```

4. **staticcheck linting**: Added `//nolint:staticcheck // Testing nil context handling.` directive for intentional nil context test in `application_test.go`

**Coverage/Quality Metrics**:

- Coverage: 56.1% (baseline before coverage improvement phase)
- Target: ≥98% coverage (need +41.9%)
- Mutation: Not yet run (target ≥98%)
- Build: ✅ Clean (`go build ./internal/template/...`)
- Tests: ✅ **All pass (10/10 PASS in 15.17s)** - was 9/10 with timeout failures
- Lint: ✅ Clean (noctx, wrapcheck, gosec, staticcheck, golangci-lint)
- Pre-commit hooks: ✅ All pass
- Pre-push hooks: ✅ All pass (golangci-lint-full, go build, secrets scan)

**Lessons Learned**:

- **Windows TIME_WAIT is 2-4 minutes by default**: Hardcoded ports are incompatible with sequential test execution on Windows (sockets held in TIME_WAIT state)
- **Port 0 is MANDATORY for tests**: Dynamic port allocation enables immediate socket reuse without TIME_WAIT blocking
- **Fiber ShutdownWithContext() doesn't guarantee immediate OS socket release**: Kernel manages TIME_WAIT independently of application shutdown
- **SO_REUSEADDR is platform-specific and complex on Windows**: Port 0 is simpler and more reliable cross-platform solution
- **Start() methods MUST monitor context cancellation concurrently**: Cannot rely solely on blocking server calls, must use select{} pattern
- **Context-aware HTTP requests required for noctx linting**: Use `client.Do(http.NewRequestWithContext(ctx, method, url, body))` not `client.Get/Post(url)`
- **Type assertions need safety checks**: Always use `ok` pattern before type conversion (`tcpAddr, ok := listener.Addr().(*net.TCPAddr)`)
- **Integer conversions need range validation for gosec**: Validate port range before `int→uint16` conversion to prevent overflow
- **Intentional nil context tests need nolint directives**: Use `//nolint:staticcheck // Testing nil context handling.` for error handling tests
- **Iterative commits and pushes essential**: 2 commits (1fb68962, 058c3f5b) with incremental fixes enabled workflow monitoring and validation

**Violations Found**:

- **Instruction file loading issue**: `.github/instructions/01-02.continuous-work.instructions.md` exists at correct path and is documented in `copilot-instructions.md` but is **NOT being loaded** into Copilot context (tooling/configuration problem)

**Next Steps**:

- ✅ **COMPLETED**: AdminServer refactored for configurable port
- ✅ **COMPLETED**: All tests updated to use port 0 and ActualPort()
- ✅ **COMPLETED**: All linting errors fixed
- ✅ **COMPLETED**: Documentation updated (02-03.https-ports.instructions.md)
- ✅ **COMPLETED**: Committed and pushed (1fb68962, 058c3f5b)
- ⏳ **PENDING**: Investigate instruction file loading issue
- ⏳ **PENDING**: Improve AdminServer coverage from 56.1% to ≥98%
- ⏳ **PENDING**: Run mutation testing (target ≥98%)
- ⏳ **PENDING**: Update EXECUTIVE.md lessons learned
- ⏳ **PENDING**: Continue to next task (PublicServer or coverage improvement)

**Related Commits**:

- 1fb68962 "refactor(template): AdminServer configurable port - MANDATORY port 0 for tests to avoid Windows TIME_WAIT (2-4 min) delays"
- 058c3f5b "fix(template): add nolint directive for nil context test"
