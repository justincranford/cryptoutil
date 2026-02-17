# Deployments Contents Check

**Date**: 2025-02-17
**Status**: ✅ ALL VALIDATION PASSED

## Validation Results (66/66 Pass)

- ✅ naming (deployments)
- ✅ naming (configs)
- ✅ kebab-case (configs)
- ✅ schema (9 config files validated)
- ✅ template-pattern (deployments/template)
- ✅ ports (14 deployment directories)
- ✅ telemetry (configs)
- ✅ admin (19 deployment directories)
- ✅ secrets (15 deployment directories)

**Total Duration**: 25ms

## Issues Resolved

### 1. orphaned deployments/ca/ Directory
**Status**: ✅ RESOLVED
- Deleted per user request
- Reason: No corresponding product/service exists

### 2. Underscore Secret Filenames
**Status**: ✅ RESOLVED
- Template directory secrets RETAIN underscores (validator requirement)
- All other deployments use kebab-case (hyphens)
- 202 files successfully renamed from underscores to hyphens across service deployments
- All secret references in YAML files updated

### 3. Empty .yml Directories
**Status**: ✅ RESOLVED
- Removed 5 empty directories with .yml extensions:
  - deployments/jose/config/jose-sqlite.yml
  - deployments/jose/config/jose-common.yml
  - deployments/identity/config/authz-e2e.yml
  - deployments/kms/config/common.yml
  - deployments/kms/config/sqlite.yml

### 4. OPTIONAL -> REQUIRED Migration
**Status**: ✅ RESOLVED
- All 289 JSON entries in deployments-all-files.json now marked REQUIRED
- All 116 JSON entries in configs-all-files.json now marked REQUIRED

## File Inventory

**Deployments tracked**: 289 entries (287 actual files + 2 historical references)
**Configs tracked**: 116 entries
**Total validators**: 66
**Validation success rate**: 100%

### Template Directory Special Handling

The `deployments/template/` directory uses UNDERSCORE naming for secrets as required by the template validator:
- `unseal_1of5.secret` through `unseal_5of5.secret`
- `postgres_database.secret`, `postgres_password.secret`, `postgres_username.secret`, `postgres_url.secret`
- `hash_pepper_v3.secret`

This is intentional and validates correctly. All other deployments use kebab-case (hyphens).

### Historical References (Not on Disk) 

- `pki-ca/README.md` - Tracked in JSON but file deleted
- `shared-postgres/setup-logical-replication.sh` - Tracked in JSON but file deleted

These remain in JSON for historical tracking (status=REQUIRED).

## Fixes Applied

1. Deleted `deployments/ca/` directory
2. Renamed `deployments_all_files.json` → `deployments-all-files.json`
3. Renamed `configs_all_files.json` → `configs-all-files.json`
4. Changed 57 OPTIONAL → REQUIRED in deployments JSON
5. Changed 56 OPTIONAL → REQUIRED in configs JSON
6. Renamed 202 files from underscore to hyphen naming (service deployments only)
7. Updated all deployment/config YAML files with new secret filenames
8. Reverted template secrets back to underscores (validator requirement)
9. Updated template compose/config files to reference underscore secrets
10. Removed 5 empty directories with .yml extensions
11. Regenerated JSON listings with kebab-case filenames
12. Updated internal/cmd/cicd/lint_deployments Go code to use kebab-case filenames

## Commits

1. First commit (7 files changed):
   - Deleted deployments/ca/
   - Renamed JSON files to kebab-case
   - Changed OPTIONAL → REQUIRED
   - Updated Go code references

2. Second commit (493 files changed):
   - Renamed 202 underscore files to hyphens (service deployments)
   - Updated 287 YAML files with new secret names
   - Regenerated JSON listings
   - Bulk sed operations automated

3. Third commit (16 files changed):
   - Reverted template secrets to underscores
   - Removed 5 empty .yml directories
   - Final JSON regeneration
   - All validators passing

## Conclusion

All deployment structure validation now passes. Kebab-case naming enforced consistently except for template directory (intentional exception). All entries marked REQUIRED for rigid validation.
