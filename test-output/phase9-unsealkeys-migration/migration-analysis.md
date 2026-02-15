# UnsealKeysService Migration Analysis

**Date**: 2026-02-06
**Task**: Phase 9 Task 9.1 - Analyze UnsealKeysService Files

## Files to Migrate (16 total)

### Production Code (5 files)

1. unseal_keys_service.go - Interface + implementations
2. unseal_keys_service_from_settings.go - Factory pattern
3. unseal_keys_service_sharedsecrets.go - Shared secrets implementation
4. unseal_keys_service_simple.go - Simple implementation
5. unseal_keys_service_sysinfo.go - System info implementation

### Test Code (11 files)

1. unseal_keys_service_from_settings_test.go
2. unseal_keys_service_from_settings_additional_test.go
3. unseal_keys_service_sharedsecrets_test.go
4. unseal_keys_service_simple_test.go
5. unseal_keys_service_simple_test_util.go
6. unseal_keys_service_sysinfo_test.go
7. unseal_keys_service_sysinfo_test_util.go
8. unseal_keys_service_additional_coverage_test.go
9. unseal_keys_service_comprehensive_test.go
10. unseal_keys_service_edge_cases_test.go
11. unseal_keys_service_error_paths_test.go

## External Dependencies (Packages Importing UnsealKeysService)

### Template Service (11 imports - PRIMARY CONSUMER)

1. internal/apps/template/service/server/apis/registration_integration_test.go
2. internal/apps/template/service/server/barrier/barrier_service.go
3. internal/apps/template/service/server/barrier/root_keys_service.go
4. internal/apps/template/service/server/barrier/rotation_service.go
5. internal/apps/template/service/server/barrier/barrier_service_test.go
6. internal/apps/template/service/server/barrier/rotation_service_test.go
7. internal/apps/template/service/server/barrier/rotation_handlers_test.go
8. internal/apps/template/service/server/barrier/key_services_test.go
9. internal/apps/template/service/server/businesslogic/session_manager_service_test.go
10. internal/apps/template/service/server/application/application_basic.go
11. internal/apps/template/service/server/builder/server_builder.go

### Other Services (3 imports - WILL NEED UPDATES)

1. internal/apps/jose/ja/service/testmain_test.go
2. internal/apps/cipher/im/server/apis/messages_test.go
3. internal/apps/cipher/im/repository/testmain_test.go
4. internal/apps/sm/kms/server/application/application_basic.go

## Internal Dependencies (What UnsealKeysService Imports)

### Standard Library

- fmt

### Internal Shared Packages (STILL ACCESSIBLE after migration)

- cryptoutil/internal/shared/crypto/digests
- cryptoutil/internal/shared/crypto/jose
- cryptoutil/internal/shared/crypto/keygen
- cryptoutil/internal/shared/magic
- cryptoutil/internal/shared/util/combinations

### External Packages

- github.com/google/uuid
- github.com/lestrrat-go/jwx/v3/jwk

## Migration Strategy

### Step 1: Create Target Directory

- Path: `internal/apps/template/service/server/barrier/unsealkeysservice/`
- This makes unsealkeysservice a LOCAL package within template's barrier module

### Step 2: Move All 16 Files

- Copy files from `internal/shared/barrier/unsealkeysservice/` to new location
- Preserve all file names and structure

### Step 3: Update Imports in Template Service (11 files)

OLD: `cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"`
NEW: `cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"`

OR even better (local package reference):
NEW: `unsealkeysservice "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"`

### Step 4: Update Imports in Other Services (4 files)

These services will need to import from the NEW location:
- jose-ja: testmain_test.go
- cipher-im: messages_test.go, testmain_test.go  
- sm-kms: application_basic.go

They will continue using:
`cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"`

### Step 5: Delete Old Location

- Remove `internal/shared/barrier/unsealkeysservice/` entirely
- Verify no references remain

### Step 6: Test Everything

- Build: `go build ./...`
- Test: `go test ./internal/apps/template/...`
- Test: `go test ./internal/apps/jose/...`
- Test: `go test ./internal/apps/cipher/...`
- Test: `go test ./internal/apps/sm/...`
- Lint: `golangci-lint run ./...`
- Coverage: Verify ≥95% maintained in template barrier

## Risk Assessment

### LOW RISK

- All imports are explicit (easy to find and replace)
- No circular dependencies detected
- Internal dependencies remain accessible (all in internal/shared/)
- Test files will move with production code (no orphaned tests)

### MITIGATION

- Make changes in single commit for atomic rollback if needed
- Run tests after each major step (create, move, refactor imports, delete)
- Keep old location until ALL tests pass with new location

## Expected Impact

### Build Time

- No change (same number of packages)

### Test Coverage

- Should INCREASE (template's coverage includes unsealkeysservice now)
- Current: Template + Shared counted separately
- After: All in template barrier module

### Import Complexity

- Template: SIMPLER (local package reference instead of cross-module)
- Other services: SAME (still need full path, but to template instead of shared)

## Acceptance Criteria Verification

✅ **List: All 16 Go files** - COMPLETE (5 production + 11 test files listed above)
✅ **Analyze: External dependencies** - COMPLETE (15 imports found, 11 in template, 4 in other services)
✅ **Analyze: Internal dependencies** - COMPLETE (all in internal/shared/, remain accessible)
✅ **Document: Migration strategy** - COMPLETE (6-step strategy defined above)

## Next Steps

Ready to proceed to Task 9.2: Create Package in Template
