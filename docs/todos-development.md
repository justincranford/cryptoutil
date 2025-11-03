# Cryptoutil Development Workflow & Configuration TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 16, 2025
**Status**: Development workflow enhancements planned for ongoing maintenance - Pre-commit automation analysis added

---

## 🟢 LOW - Development Workflow & Configuration

### Task DW1: Implement 12-Factor App Standards Compliance
- **Description**: Ensure application follows 12-factor app methodology for cloud-native deployment
- **12-Factor Requirements**:
  - **I. Codebase**: One codebase tracked in revision control, many deploys - **Status: ✅ IMPLEMENTED** (Single Git repository with clear versioning)
  - **II. Dependencies**: Explicitly declare and isolate dependencies - **Status: ✅ IMPLEMENTED** (Go modules with explicit dependency management)
  - **III. Config**: Store config in the environment - **Status: ✅ IMPLEMENTED** (YAML configs + environment variables for secrets)
  - **IV. Backing services**: Treat backing services as attached resources - **Status: ✅ IMPLEMENTED** (Database via connection strings)
  - **V. Build, release, run**: Strictly separate build and run stages - **Status: ✅ IMPLEMENTED** (Dockerfile with distinct build/validation/runtime stages)
  - **VI. Processes**: Execute the app as one or more stateless processes - **Status: ❓ PARTIALLY IMPLEMENTED** (Appears stateless but needs verification)
  - **VII. Port binding**: Export services via port binding - **Status: ✅ IMPLEMENTED** (Binds to configurable ports 8080/9090)
  - **VIII. Concurrency**: Scale out via the process model - **Status: ❓ NEEDS AUDIT** (Horizontal scaling capability needs verification)
  - **IX. Disposability**: Maximize robustness with fast startup and graceful shutdown - **Status: ✅ IMPLEMENTED** (Signal handling + health checks)
  - **X. Dev/prod parity**: Keep development, staging, and production as similar as possible - **Status: ✅ IMPLEMENTED** (Docker compose environments)
  - **XI. Logs**: Treat logs as event streams - **Status: ✅ IMPLEMENTED** (Structured slog logging as event streams)
  - **XII. Admin processes**: Run admin/management tasks as one-off processes - **Status: ❓ NEEDS AUDIT** (Admin task separation needs verification)
- **Current State**: 8/12 factors fully implemented, 2 partially implemented, 2 need audit
- **Action Items**:
  - Audit Factor VI (stateless processes) - verify no local file storage or in-memory state
  - Audit Factor VIII (concurrency) - verify horizontal scaling capability with multiple instances
  - Audit Factor XII (admin processes) - verify admin tasks run as separate processes
  - Document final 12-factor compliance status
  - Update deployment configurations for any missing factors
- **Files**: Docker configs, deployment files, application architecture
- **Expected Outcome**: Cloud-native, scalable application following industry best practices
- **Priority**: LOW - Best practices alignment
- **Timeline**: Ongoing maintenance

### Task DW2: Implement Hot Config File Reload
- **Description**: Add ability to reload configuration files without restarting the server
- **Current State**: Configuration loaded only at startup
- **Action Items**:
  - Add file watcher for config files (development mode only)
  - Implement graceful config reload with validation
  - Add reload endpoint for runtime config updates
  - Handle config reload failures gracefully
  - Add configuration versioning/checksum validation
- **Files**: `internal/common/config/config.go`, server startup code
- **Expected Outcome**: Development workflow improvement with live config reloading
- **Priority**: LOW - Developer experience enhancement
- **Timeline**: Q1 2026

---

## 🟡 MEDIUM - Database & Dependencies

### Task DB1: Migrate from lib/pq to pgx PostgreSQL driver
- **Description**: The lib/pq PostgreSQL driver is in maintenance mode and recommends migrating to pgx
- **Current State**: ✅ **COMPLETED** - Successfully migrated from `github.com/lib/pq` to `github.com/jackc/pgx/v5/stdlib`
- **Migration Details**:
  - ✅ Replaced `github.com/lib/pq` with `github.com/jackc/pgx/v5/stdlib` for GORM compatibility
  - ✅ Updated import in `internal/server/repository/sqlrepository/gormdb.go`
  - ✅ Updated all Go files that imported `github.com/lib/pq` (6 files total)
  - ✅ Updated PostgreSQL error handling from `*pq.Error` to `*pgconn.PgError`
  - ✅ Updated golang-migrate to use pgx/v5 driver instead of postgres (lib/pq) driver
  - ✅ Tested database connectivity with both SQLite and PostgreSQL backends
  - ✅ Verified all existing functionality works with new driver
  - ✅ Note: `lib/pq` remains as indirect dependency due to golang-migrate internal dependencies (acceptable)
- **Files Updated**:
  - `go.mod` - Removed lib/pq, kept pgx/v5
  - `internal/server/repository/sqlrepository/gormdb.go` - Updated import
  - `internal/server/repository/sqlrepository/sql_schema_util.go` - Updated import
  - `internal/server/repository/sqlrepository/sql_provider_test.go` - Updated import
  - `internal/server/repository/sqlrepository/sql_migrations.go` - Updated to use pgx migration driver
  - `internal/server/repository/orm/business_entities_operations.go` - Updated import and error type
  - `internal/server/repository/orm/orm_repository.go` - Updated import
  - `internal/server/repository/orm/orm_transaction_test.go` - Updated import
  - `internal/server/barrier/barrier_service_test.go` - Updated import
- **URLs**:
  - lib/pq status: https://github.com/lib/pq/blob/master/README.md (maintenance mode)
  - pgx replacement: https://github.com/jackc/pgx (actively maintained)
  - pgx migration driver: https://github.com/golang-migrate/migrate/tree/master/database/pgx/v5
- **Action Items**:
  - ✅ Update go.mod to replace lib/pq with pgx/v5
  - ✅ Update GORM dialector configuration for pgx
  - ✅ Update all Go file imports from lib/pq to pgx/v5/stdlib
  - ✅ Update PostgreSQL error handling to use pgconn.PgError
  - ✅ Update golang-migrate to use pgx/v5 driver instead of postgres driver
  - ✅ Run full test suite with both database backends
  - ✅ Update any pq-specific connection parameters if needed
- **Files**: `go.mod`, `internal/server/repository/sqlrepository/gormdb.go`, and 6 other Go files
- **Expected Outcome**: ✅ Modern, actively maintained PostgreSQL driver with better performance and features
- **Priority**: MEDIUM - Dependency modernization
- **Timeline**: ✅ **COMPLETED** Q1 2026

---

## 🟢 LOW - Documentation & API Management

### Task DOC1: API Versioning Strategy Documentation
- **Description**: Document comprehensive API versioning strategy and deprecation policy
- **Current State**: Basic API versioning exists but not formally documented
- **Action Items**:
  - Document API versioning conventions (URL-based, header-based, etc.)
  - Create API deprecation policy and timeline
  - Document backward compatibility guarantees
  - Create migration guides for API changes
- **Files**: `docs/api-versioning.md`, OpenAPI specifications
- **Expected Outcome**: Clear API evolution and compatibility guidelines
- **Priority**: Low - API management
