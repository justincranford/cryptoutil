# Implementation Plan - Lint Enhancement & Technical Debt (V9)

**Status**: In Progress - Phase 2 & 3 Complete
**Created**: 2026-02-04
**Updated**: 2026-02-05
**Purpose**: Enhance lint tools and address technical debt from V8

## Background

V8 successfully completed all core objectives:
- Port standards alignment (8050-8130 series)
- Health path standardization (`/admin/api/v1/livez` on port 9090)
- Admin port security (9090 never exposed to host)
- lint-ports and lint-compose tools pass

V9 carries forward deferred Phase 19 enhancements and addresses new improvements.

## Executive Decisions

### Decision 1: lint-ports Enhancement Scope
**Options**:
- A: Only container port validation
- B: Container port + host port range validation
- C: Container port + host port range + health path validation
- D: Full scope (container, host, health, compose, docs)
- E: [blank - user fills in]

**Rationale**: V8 success criteria are met with current lint-ports (legacy port detection). Enhanced validation would catch configuration drift proactively.

### Decision 2: lint_go Pre-existing Issues
**Options**:
- A: Fix lint_go issues in V9 (errcheck, goconst, gosec)
- B: Defer lint_go issues to separate cleanup plan
- C: [blank - user fills in]

**Rationale**: `golangci-lint run ./...` shows 7+ issues in internal/cmd/cicd/lint_go package. These are pre-existing and not related to port/health work.

---

## V9 Scope

### Phase 1: lint-ports Enhancement (Deferred from V8 Phase 19)

**Objective**: Enhance lint-ports to validate:
1. Container ports match expected values per service
2. Host port ranges are within allocated range
3. Health paths use `/admin/api/v1/livez` on port 9090
4. Compose files have correct port mappings
5. Documentation references correct ports

**Tasks**:
- 1.1: Add container port validation to lint_ports.go
- 1.2: Add host port range validation
- 1.3: Add health path validation
- 1.4: Add compose file port validation
- 1.5: Add documentation port validation
- 1.6: Update lint_ports tests for new features
- 1.7: Verify all existing tests still pass

**Estimated**: 4-6 hours

### Phase 2: lint_go Technical Debt ✅ COMPLETE

**Objective**: Fix pre-existing linting issues in lint_go package

**Issues Identified**:
- errcheck: `file.Close()` return value not checked (2 occurrences)
- goconst: string `)` has 2 occurrences (make constant)
- gosec G306: WriteFile permissions should be 0600 or less (4 occurrences in tests)

**Tasks**:
- 2.1: Fix errcheck issues in cryptopatterns.go ✅
- 2.2: Fix goconst issue in lint_go.go ✅
- 2.3: Fix gosec issues in cryptopatterns_test.go ✅
- 2.4: Run `golangci-lint run ./...` and verify 0 issues in lint_go ✅

**Estimated**: 1-2 hours
**Actual**: 0.35h
**Evidence**: Commits c5fcf648, fb27e4a2

### Phase 3: Identity E2E Docker Investigation ✅ COMPLETE

**Objective**: Investigate and fix identity-authz E2E test Docker issues

**Background**: E2E tests for identity services fail with Docker infrastructure issues. This is NOT related to port/health work but needs investigation.

**Root Causes Identified**:
1. **Wrong health check paths**: Docker Compose and Dockerfiles used `/health` on wrong ports (18000, 18100, etc.) instead of `/admin/api/v1/livez` on port 9090
2. **Port conflict**: Both identity-authz and identity-idp were configured for port 8100
3. **Config override bug**: All 5 identity services unconditionally overwrite config-specified port with default port value, ignoring config file settings

**Fixes Applied**:
1. Fixed health paths in compose.e2e.yml, compose.simple.yml, and all 5 Dockerfiles
2. Changed identity-idp from port 8100 to 8101 in E2E config
3. Fixed config override logic in all 5 identity services to only apply default when port=0

**Tasks**:
- 3.1: Analyze identity E2E test failures ✅
- 3.2: Fix Docker Compose configuration issues ✅
- 3.3: Verify E2E tests pass ✅

**Estimated**: 2-4 hours
**Actual**: 2.25h
**Evidence**: All 5 services healthy, E2E tests pass in 5.170s

---

## Success Criteria

- [ ] lint-ports validates container ports, host ranges, health paths
- [x] `golangci-lint run ./internal/cmd/cicd/lint_go/...` shows 0 issues
- [x] identity E2E tests pass
- [x] All existing tests continue to pass

## Port Standards Reference (from V8)

| Service | Container Port | Host Port Range | Admin Port |
|---------|----------------|-----------------|------------|
| sm-kms | 8080 | 8080-8089 | 9090 |
| cipher-im | 8070 | 8070-8079 | 9090 |
| jose-ja | 8060 | 8060-8069 | 9090 |
| pki-ca | 8050 | 8050-8059 | 9090 |
| identity-authz | 8100 | 8100-8109 | 9090 |
| identity-idp | 8101 | 8100-8109 | 9090 |
| identity-rs | 8110 | 8110-8119 | 9090 |
| identity-rp | 8120 | 8120-8129 | 9090 |
| identity-spa | 8130 | 8130-8139 | 9090 |

**Note**: identity-idp uses port 8101 in E2E to avoid conflict with identity-authz on 8100.

**Health Path Standard**: `/admin/api/v1/livez` on port 9090
