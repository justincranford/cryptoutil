# Cryptoutil Documentation Plans Index

**Last Updated**: 2025-11-21
**Purpose**: Organized index of all documentation plan directories, ordered newest to oldest for systematic review

---

## Plan Directories (Newest → Oldest)

### 01-mixed (2025-11-21)
**Status**: Active/Current
**Last Updated**: 2025-11-21 10:35 AM
**Purpose**: Mixed cross-cutting concerns and ongoing maintenance tasks
**Files**:
- `todos-development.md` - Development workflow & configuration (12-factor compliance ✅ COMPLETE)
- `todos-infrastructure.md` - Infrastructure & deployment tasks
- `todos-observability.md` - Observability & monitoring enhancements
- `todos-quality.md` - Code quality & testing improvements
- `todos-security.md` - Security hardening & compliance
- `todos-testing.md` - Testing infrastructure (⚠️ GORM AutoMigrate blocker active)

**Key Items**:
- ✅ 12-Factor App Compliance COMPLETE
- ✅ PostgreSQL driver migration (lib/pq → pgx/v5) COMPLETE
- ⚠️ GORM AutoMigrate failure blocking identity integration tests
- Hot config reload (LOW priority)
- API versioning documentation (LOW priority)

---

### 02-refactor (2025-11-11)
**Status**: Archive/Reference
**Last Updated**: 2025-11-11 12:04 AM
**Purpose**: General refactoring initiatives and code improvements
**Files**:
- `README.md` - Refactoring plans and guidelines

**Key Items**:
- TBD - needs review and update based on current project state

---

### 03-identityV2 (2025-11-11)
**Status**: Archive/Reference - Superseded by newer identity implementation
**Last Updated**: 2025-11-11 12:04 AM
**Purpose**: Second iteration of identity/OAuth2/OIDC implementation
**Files**: 28 files including:
- `README.md` - Identity V2 master plan
- `identityV2_master.md` - Master tracking document
- `requirements.yml` - Requirements specification
- `task-01-*.md` through `task-20-*.md` - 20 task breakdowns
- Supporting docs: topology, config normalization, dependency graph, etc.

**Key Items**:
- ⚠️ Task 10.5 blocked by GORM AutoMigrate failure
- Tasks 10.6-20 depend on 10.5 completion
- Comprehensive OAuth2/OIDC implementation plan
- Docker Compose orchestration suite
- E2E testing framework

---

### 04-identity (2025-11-11)
**Status**: Archive/Reference - Original identity implementation (older than 03-identityV2)
**Last Updated**: 2025-11-11 12:04 AM
**Purpose**: Original identity/OAuth2/OIDC implementation plans
**Files**: 22 files including:
- `identity_master.md` - Original master plan
- `01_foundation_setup.md` through `15_user_auth_hardware.md` - 15 task breakdowns
- `16_gap_analysis.md` - Gap analysis document
- `17_18_19_summary.md` - Docker Compose & E2E summary
- Task-specific files: `task-07-client-auth-cli.md`, `task-08-key-rotation-ops.md`
- `e2e_coverage_report.md` and `e2e_coverage.html` - E2E coverage analysis

**Key Items**:
- Original implementation plan (superseded by 03-identityV2)
- Completed tasks documented in gap analysis
- Historical reference for implementation decisions

---

## Review Order

Work through plans in **reverse chronological order** (newest first):

1. **01-mixed** (Active) - Review and update based on current project state
2. **02-refactor** (Archive) - Review and update or archive obsolete items
3. **03-identityV2** (Newer identity plan) - Review progress, update blocker status
4. **04-identity** (Older identity plan) - Archive or consolidate with 03-identityV2

---

## Status Legend

- ✅ **COMPLETE** - Task fully implemented and verified
- ⚠️ **BLOCKED** - Active blocker preventing progress
- 🔄 **IN PROGRESS** - Currently being worked on
- 📋 **PLANNED** - Not yet started
- 🗄️ **ARCHIVED** - Obsolete or superseded

---

## Next Steps

1. Complete review of 01-mixed todos (ongoing)
2. Update 02-refactor based on current state
3. Review 03-identityV2 and update blocker status
4. Archive or consolidate 04-identity with 03-identityV2
