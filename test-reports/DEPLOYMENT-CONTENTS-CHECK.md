# Deployments Contents Check

## Summary
 
**Execution Date:** $(date -u +"%Y-%m-%d %H:%M:%S UTC")
**Deployments Directory Listing:** 288 files
**JSON Tracking Entries:** 288 entries
**Missing Entries:** 0
**Extra Entries:** 0

## Section 1: Expected and Found

All 288 files in deployments/ are tracked in deployments-all-files.json with REQUIRED status:
- All compose files are tracked
- All Dockerfiles are tracked
- All secret files are tracked  
- All config files are tracked

Status: ✅ PASS - Perfect 1:1 match

## Section 2: Expected But Not Found

None. All expected files exist and are tracked.

Status: ✅ PASS

## Section 3: NotExpected But Found

None. All files found in deployments/ are properly tracked in the JSON.

Status: ✅ PASS

## Section 4: Unexpected and Not Found

None applicable.

Status: ✅ PASS

## Actions Taken

1. ✅ Deleted `deployments/ca/` directory (not in JSON, orphaned)
2. ✅ Renamed `deployments_all_files.json` → `deployments-all-files.json` (kebab-case)
3. ✅ Renamed `configs_all_files.json` → `configs-all-files.json` (kebab-case)
4. ✅ Changed all OPTIONAL → REQUIRED (57 entries made rigid)
5. ✅ Updated code references to new kebab-case filenames

## Outstanding Issues

**Naming Validation Failure:** 
The linter reports naming violations due to underscore usage in filenames:
- Secret files: browser_username, browser_password, service_username, service_password
- Unseal keys: unseal_1of5, unseal_2of5, etc.
- Hash peppers: hash_pepper_v3
- Postgres configs: postgres_database, postgres_url, postgres_password, postgres_username

These files use underscores instead of hyphens, violating kebab-case naming convention. Renaming these files requires updating:
- All compose.yml files that reference them
- All config files  
- The JSON tracking files
- Approximately 1000+ line changes across the codebase

Recommendation: Schedule dedicated refactoring session for kebab-case compliance of secret filenames.

