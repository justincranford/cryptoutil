# Tasks - Service Template & CICD Fixes (V6)

**Status**: V6 COMPLETE (Phase 13 done) | One blocked task remains | See V7 for unified service migration
**Last Updated**: 2026-02-02

## Summary

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1-4 | ✅ Complete | Instructions, CICD, Deployment, Critical Fixes |
| Phase 5 | ⚠️ Partial | Test Architecture (5.1 BLOCKED - StartApplicationListener) |
| Phase 6-8 | ✅ Complete | Coverage, Cleanup, Race Detection |
| Phase 9 | ✅ Complete | KMS Analysis |
| Phase 10 | ✅ Complete | Cleanup (10.1-10.4) |
| Phase 11 | ✅ Complete | ServerBuilder Extensions (11.1-11.3) |
| Phase 12 | ✅ Complete | KMS Before/After Comparison |
| Phase 13 | ✅ Complete | KMS ServerBuilder Migration |

**All completed tasks**: See [completed.md](./completed.md)
**V7 unified plan**: See [../fixes-needed-plan-tasks-v7/](../fixes-needed-plan-tasks-v7/)

---

## Remaining Incomplete Tasks

### Task 5.1: Refactor Listener Tests to app.Test() (BLOCKED)

- **Status**: ❌ BLOCKED
- **Blocker**: `StartApplicationListener` not yet implemented (returns "implementation in progress" error)
- **Estimated**: 3h
- **Files**:
  - `internal/apps/template/service/server/listener/servers_test.go`
  - `internal/apps/template/service/server/listener/application_listener_test.go`
- **Description**: Replace real HTTPS listeners with Fiber app.Test() for in-memory testing
- **Next Steps**:
  1. Complete `StartApplicationListener` implementation first
  2. THEN refactor tests to use app.Test() pattern
- **Note**: This blocker is resolved by V7 plan - unified service-template will complete this implementation

---

## V6 Outcome Summary

**What V6 accomplished**:
- Extended ServerBuilder with SwaggerUI, OpenAPI, Security Headers
- Analyzed KMS architecture for migration
- Created database/barrier/migration abstraction modes

**What went wrong**:
- Phase 13 created OPTIONAL modes (disabled DB, disabled barrier, etc.) instead of MANDATORY modes
- KMS should use GORM like cipher-im and jose-ja, not a custom SQLRepository
- KMS should use service-template barrier, not a separate shared barrier
- KMS should use service-template JWT authentication with realms
- The abstractions fragmented rather than unified the services

**V7 corrects this**: See [../fixes-needed-plan-tasks-v7/](../fixes-needed-plan-tasks-v7/) for the unified approach
