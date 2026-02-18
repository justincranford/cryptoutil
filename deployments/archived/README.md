# Archived Deployment Directories

This directory contains deprecated deployment configurations that have been archived during the deployment architecture refactoring.

## Archived Directories

### compose-legacy/ (Archived: 2026-02-17)

**Original Location**: `deployments/compose/`

**Reason for Archival**: 
- Breaks SUITE→PRODUCT→SERVICE hierarchical pattern
- Used by legacy E2E tests in `internal/test/e2e/` 
- Does NOT use service-template ComposeManager pattern
- Custom infrastructure duplicates template functionality

**Replaced By**:
- SUITE-level: `deployments/cryptoutil-suite/`
- PRODUCT-level: `deployments/{PRODUCT}/`
- SERVICE-level: `deployments/{PRODUCT}-{SERVICE}/`
- Modern E2E pattern: `internal/apps/{PRODUCT}/{SERVICE}/e2e/` using ComposeManager

**Migration Path**:
- Migrate E2E tests to use ComposeManager from `internal/apps/template/testing/e2e/`
- Follow cipher-im or identity E2E patterns (RECOMMENDED)
- See `test-output/phase1/e2e-patterns.txt` for analysis

**DO NOT DELETE**: May be needed for reference during E2E migration (Phase 6)

### kms-legacy/ (Removed: 2026-02-17)

**Original Location**: `deployments/kms/`

**Reason for Removal**:
- Empty directory (contained only an empty `config/` subdirectory)
- Legacy name predating the `PRODUCT-SERVICE` naming convention
- Replaced by `deployments/sm-kms/` (SERVICE-level) and `deployments/sm/` (PRODUCT-level)
- CI workflow references updated to use `deployments/cryptoutil-suite/Dockerfile`

**Note**: Directory was empty so `git mv` was not possible; directory was simply removed.

