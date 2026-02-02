# Implementation Plan - Service Template & CICD Fixes

**Status**: 85% Complete (Phases 1-9 done, Phases 10-12 remaining)
**Created**: 2026-01-31
**Last Updated**: 2026-02-01

## Executive Summary

All phases 1-9 have been executed. Remaining work:
- **Task 5.1**: BLOCKED pending `StartApplicationListener` implementation completion
- **Phase 10**: Cleanup leftover coverage files (57+ files from LLM autonomous work)
- **Phase 11**: KMS ServerBuilder Extension (REQUIRED - must extend ServerBuilder for KMS-style services)
- **Phase 12**: KMS Before/After Comparison (REQUIRED - verify no loss of functionality or design intent)

Completed tasks archived to [completed.md](./completed.md).

All tests pass. All quality gates pass. Build is clean.

## Overview

This plan consolidates incomplete work from v6 analysis. It covers service-template fixes, CICD enforcement improvements, and test conformance issues.

## Technical Context

- **Language**: Go 1.25.5
- **Framework**: Fiber v2, GORM
- **Database**: PostgreSQL OR SQLite with GORM
- **Testing**: Table-driven tests, TestMain pattern, app.Test() for handlers

## Phases

### Phase 1: Copilot Instructions Updates - 2-3h ✅ COMPLETE

- Update 02-02.service-template.instructions.md
- Update 03-02.testing.instructions.md
- Update 03-08.server-builder.instructions.md

### Phase 2: CICD Enforcement Improvements - 6-8h ✅ COMPLETE

HIGH Priority linters added:
- Docker secrets pattern enforcement (#15)
- testify require over assert (#16)
- t.Parallel() enforcement (#17)
- Table-driven test pattern (#18)
- Hardcoded test passwords (#19)
- crypto/rand over math/rand (#20)
- No inline env vars in compose (#26)
- No InsecureSkipVerify (#28)
- golangci-lint v2 schema (#29 - CRITICAL)

### Phase 3: Deployment Fixes - 1-2h ✅ COMPLETE

- Fixed healthcheck path mismatch in KMS compose.yml (/admin/v1/ → /admin/api/v1/)
- Created template deployment (deployments/template/compose.yml)

### Phase 4: Critical Fixes (TODOs and Security) - 4-6h ✅ COMPLETE

- Complete registration handler TODOs (password hashing, validation)
- Add admin middleware to registration routes
- Implement realm lookup for multi-tenant deployments

### Phase 5: Test Architecture Refactoring - 6-8h ⚠️ PARTIAL

- Task 5.1 BLOCKED pending StartApplicationListener implementation
- All other tasks complete

### Phase 6: Coverage Improvements - 4-6h ✅ COMPLETE

- Repository package: 84.8% → 95%
- Application package: 89.8% → 95%
- Businesslogic package: 87.4% → 95%
- Config packages: 86.9%/87.1% → 95%

### Phase 7: Code Cleanup - 2-3h ✅ COMPLETE

- Investigated low-coverage functions
- Fixed config bug acknowledged in tests

### Phase 8: Race Condition Testing - 8-12h ✅ COMPLETE

- Enabled race detection across all packages
- Fixed race conditions in concurrent code
- Added proper synchronization primitives

### Phase 9: KMS Modernization Analysis - 15-20h ✅ COMPLETE

**STATUS**:
- Task 9.1: ✅ Complete - Analysis documented in `test-output/kms-migration-analysis/`
- Task 9.2: ✅ Complete (No Changes) - KMS already uses GORM via ORM wrapper
- Task 9.3: ⚠️ Architectural mismatch discovered - ServerBuilder needs extension
- Task 9.4: ✅ Complete (No Changes) - E2E tests already work

**Architectural Discovery**:
KMS has features that ServerBuilder currently lacks:
- Swagger UI with basic auth
- CSRF middleware
- OpenAPI handlers
- Security headers

### Phase 10: Cleanup Leftover Coverage Files (REQUIRED)

**Status**: Not Started
**Discovery**: LLM autonomous work left 57+ coverage files scattered in project root and internal directories

**User Decisions (from quizme-v1.md)**:
- **Q1**: Delete ALL files in test-output/ (clean slate)
- **Q2**: Detect leftover files in ALL directories including test-output/
- **Q3**: Pattern list: `*.out`, `*.cov`, `*.prof`, `*coverage*.html`, `*coverage*.txt`
- **Q4**: Auto-delete files if found, with warning message

**Tasks**:
- 10.1: Delete root-level coverage files (17 files)
- 10.2: Delete internal/ directory coverage files (4 files)
- 10.3: Delete ALL files in test-output/ (per user decision)
- 10.4: Add CICD linter to auto-delete leftover coverage files

**Estimated**: 2h

### Phase 11: KMS ServerBuilder Extension (REQUIRED)

**Status**: Not Started
**Rationale**: Service-template MUST support all KMS functionality to enable lateral migration

**CRITICAL**: KMS migrating to service-template MUST be a lateral move - no loss of functionality, architecture, design intent, or test intent. Service-template is designed to be the foundation of ALL 9 services including KMS.

**Tasks**:
- 11.1: Add `WithSwaggerUI()` to ServerBuilder (4h)
- 11.2: Add `WithOpenAPIHandlers()` to ServerBuilder (4h)
- 11.3: Add security headers to ServerBuilder (2h)
- 11.4: Migrate KMS to extended ServerBuilder (4h)

**Total estimated**: 14h

### Phase 12: KMS Before/After Comparison (REQUIRED)

**Status**: Not Started
**Rationale**: User requires comprehensive verification that service-template reproduces ALL KMS functionality

**CRITICAL**: Service-template has been built up through many iterations to serve as the foundation of all products and 9 services. KMS switching to service-template MUST be a lateral move in every way:
- No loss of functionality
- No divergence from design intent
- No loss of test coverage or test intent
- Maximum reusability
- Architecture preserved

**Tasks**:
- 12.1: Document KMS current architecture (endpoints, middleware, security, config)
- 12.2: Document KMS current test coverage and test intent
- 12.3: Compare KMS vs service-template feature parity
- 12.4: Verify all KMS tests pass with service-template backend
- 12.5: Document any intentional differences and their rationale
- 12.6: Create final comparison report

**Total estimated**: 6h

### Phase 13: KMS Full Migration to ServerBuilder (DEFERRED)

**Status**: Deferred (from Task 11.4)
**Rationale**: Full KMS migration requires significantly more effort than originally estimated (20-30 hours, not 4 hours)

**CRITICAL**: This phase handles the actual KMS migration that was originally Task 11.4. The scope is:
- Migrate KMS from custom application_listener.go (1223 lines) to ServerBuilder pattern
- Preserve ALL KMS functionality
- Maintain ALL KMS tests
- Create clean separation between infrastructure (ServerBuilder) and business logic (KMS handlers)

**Tasks**:
- 13.1: Create KMS server structure (similar to cipher-im pattern)
- 13.2: Migrate KMS handler registration to ServerBuilder
- 13.3: Migrate KMS middleware chain to ServerBuilder patterns
- 13.4: Migrate KMS health checks to ServerBuilder
- 13.5: Delete application_listener.go after all functionality migrated
- 13.6: Comprehensive KMS test verification

**Total estimated**: 20h (properly scoped)

**Architecture Reference**:
- Current KMS: ServerApplicationBasic → ServerApplicationCore → ServerApplicationListener (3-layer)
- Target: JoseServer pattern with ServerBuilder.WithPublicRouteRegistration()
- Reference: cipher-im server.go demonstrates proper ServerBuilder usage

## Technical Decisions

### Decision 1: Service Configuration Pattern
- **Chosen**: Create services from settings (e.g., `UnsealKeysServiceFromSettings`)
- **Rationale**: Consistent initialization, testable, configuration-driven
- **Impact**: All services should have `*FromSettings` factory functions

### Decision 2: Test Architecture Pattern
- **Chosen**: app.Test() for all handler tests, TestMain for heavyweight resources
- **Rationale**: Prevents Windows Firewall prompts, faster execution, no port binding

### Decision 3: Table-Driven Tests
- **Chosen**: Single table-driven test per error category
- **Rationale**: Reduced duplication, easier to add cases, faster execution

### Decision 4: Coverage File Cleanup (from quizme-v1.md)
- **Chosen**: Auto-delete leftover coverage files with warning
- **Patterns**: `*.out`, `*.cov`, `*.prof`, `*coverage*.html`, `*coverage*.txt`
- **Scope**: ALL directories including test-output/
- **Rationale**: Prevents sprawl, maintains clean workspace

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| ServerBuilder extension breaks existing services | Medium | High | Test all services after each change |
| KMS migration causes regression | Medium | High | Comprehensive before/after comparison (Phase 12) |
| Test refactoring breaks existing tests | Medium | Medium | Run tests after each file refactored |

## Quality Gates

- ✅ All tests pass (`go test ./...`)
- ✅ Coverage ≥95% production, ≥98% infrastructure
- ✅ Linting clean (`golangci-lint run`)
- ✅ No new TODOs without tracking
- ✅ No standalone tests (table-driven only)
- ✅ No real HTTPS listeners in unit tests
- ✅ All copilot instructions accurate

## Success Criteria

- [x] Phase 1-9 complete (analysis and preparation)
- [ ] Phase 10 complete (cleanup leftover files)
- [ ] Phase 11 complete (ServerBuilder extension for KMS)
- [ ] Phase 12 complete (KMS before/after comparison verified)
- [x] All quality gates pass
- [x] CI/CD green
- [x] Documentation updated
- [x] Race detection clean

## References

- Analysis docs archived in [archive/](./archive/)
- Copilot instructions in `.github/instructions/`
- Comparison table: [comparison-table.md](./comparison-table.md)
- v4 source: `docs/fixes-needed-plan-tasks-v4/` (archived after v6 complete)
