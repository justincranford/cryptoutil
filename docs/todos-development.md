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
