# Framework v4 - Lessons Learned

---

## Phase 1: Fix Legacy sm-kms-pg- Naming and Add OTLP Service Name Check

**Completed**: commit `dc5970d47`

### Lessons

1. **PowerShell heredoc strips leading whitespace**: `@'...'@` removes leading spaces/tabs from each line. Go files written via PowerShell heredoc have zero indentation and fail `gofumpt`. Fix: always run `gofmt -w` + `gofumpt -w` immediately after writing any Go file via PowerShell.

2. **Magic literal checker covers more than expected**: The `lint-go` `literal-use` linter enforces magic constants for service name strings (`"sm-im"`, `"sm-kms"`, `"jose-ja"`), file permission octals (`0o600`, `0o755`), and path segment strings (`"im"`, `"kms"`). Tests must use `cryptoutilSharedMagic.*` constants for all known service names and permissions — even in table-driven test data fields.

3. **backtick strings in `wantErrContain` are exempt**: Composite strings like `` `got "sm-kms-pg-1", want "sm-kms-postgres-1"` `` inside backtick raw string literals are NOT flagged by the literal checker. Only standalone string assignments trigger the check.

4. **Multi_replace_string_in_file fails on mixed whitespace**: The tool uses exact string matching. If the file uses tabs and the replacement string uses spaces, it will fail silently. When in doubt, read the file first with `Get-Content -Raw` to see exact bytes.

5. **`configs/orphaned/` exclusion is mandatory**: The `orphaned/` directory contains legacy configs with intentionally incorrect naming. Any new OTLP name check MUST explicitly skip this directory or it will produce false positives on every run.

---

## Phase 2: Registry-Driven Foundation and Entity Registry Check

**Completed**: commit `7bae5aee3`

### Lessons

1. **`lint-go` literal-use flags test case `name:` fields too**: The `magic-usage` linter flags string literals in table-driven test struct fields — including `name: "sm-im"` used only as a `t.Run` label. These must use the corresponding magic constant (e.g., `name: cryptoutilSharedMagic.OTLPServiceSMIM`) even though the semantic meaning is just "test display name". Rationale: the linter enforces consistency mechanically, not semantically.

2. **Count literals need semantic constants**: Literal integers like `5` and `10` in `require.Len()` assertions are flagged. The linter suggested `JWECompactParts = 5` (wrong semantics) and `PercentageBasis10 = 10` (also wrong). The correct fix was to add `SuiteProductCount = 5` to `magic_cicd.go` alongside the existing `SuiteServiceCount = 10`. This is the right abstraction: the count is meaningful domain knowledge, not just any arbitrary integer.

3. **Display name strings need title-case constants**: `"Skeleton"` (title-case display name for the Skeleton product) is a magic literal requiring `SkeletonProductNameTitleCase`. Scan existing magic constants before creating new ones — `SkeletonProductNameTitleCase` already existed.

4. **Suite ID `"cryptoutil"` uses `DefaultOTLPServiceDefault`**: The suite name matches the default OTLP service name. Reuse `DefaultOTLPServiceDefault = "cryptoutil"` even in non-OTLP contexts where it refers to the suite identity.

5. **`configs/{PRODUCT}/{SERVICE}/` structure already existed**: All 10 `configs/pki/ca/`, `configs/jose/ja/`, etc. directories were created in the prior session and were already git-tracked. The entity-registry-completeness check confirmed all 10 pass on the real workspace immediately. No remediation was needed.

6. **Magic file names (e.g., `"magic_sm.go"`) are NOT flagged by literal-use**: The checker does not have entries for the magic file names themselves. Only service ID strings (`"sm-im"`) and product names (`"pki"`) are registered in the checker's allowlist. When writing `TestCheckInDir_MissingMagicFile`, the `magicFile:` strings like `"magic_sm.go"` can remain as-is.

7. **Registry defensive copies prevent mutation in tests**: `AllProducts()`, `AllProductServices()`, `AllSuites()` each use `copy()` to return independent slices. The `TestAllReturnsIndependentCopies` test validates this invariant. This is important because callers in fitness checks iterate the registry without expecting surprises from prior test runs.

---

## Phase 3: Banned Name Detection

**Commit**: `f7a9c84a6`

### Self-Referential Exclusion

The banned-product-names checker defines the banned phrases as Go string literals. The checker's own directory (`banned_product_names/`) must be excluded from scanning to prevent self-referential false positives. General rule: any checker that defines the forbidden patterns as literals must exclude its own package directory.

### Documentation Exclusion

Planning docs under `docs/` legitimately reference old product names when explaining the migration history (e.g., "we renamed Cipher IM to SM IM"). These are valid documentation, not accidental regressions. Excluding `docs/` from the production drift check is correct.

### Test File Exclusion

Test files (e.g., `legacy_dir_detection_test.go`) use banned phrases as negative test data — asserting that the checker *detects* them. Scanning `_test.go` files would cause the integration tests themselves to fail. Skip `_test.go` files in all banned-phrase walk functions.

### `test-output/` Exclusion

The `test-output/` directory contains historical session artifacts (e.g., coverage files, workflow reports). Its contents are not maintained and may reference old names incidentally. Exclude it from banned-phrase scans.

### The Check Caught a Real Violation

The `legacy-dir-detection` integration test revealed that `internal/apps/cipher/` still existed on disk. This was the check working correctly — the directory was a leftover of the cipher→SM migration (contained only the test binary `cipher.test.exe` and temp test directories, no Go source). Removal was required before the check could pass in CI. The lesson: write the check, run it, fix what it finds.

### New Magic Constants Unlock Pre-Existing Violations

Adding `CICDExcludeDirDocs`, `CICDExcludeDirTestOutput`, and `CICDExcludeDirBannedProductNamesCheck` immediately created 18 new `[literal-use]` blocking violations across 12 files — these were pre-existing literal usages that became blocking once the constant was defined. All must be fixed in the same commit. New constants should always be followed by a full `lint-go` run to catch newly-unblocked violations.

---

## Phase 4: Deployment Directory Completeness

**Commit**: `239a78f07`

### Pre-Audit Before Writing the Check

Auditing all 10 PS deployment directories before writing the check confirmed all were complete. This enabled writing a correct integration test (`TestCheck_RealWorkspace`) with confidence it would pass.

### Pattern Reuse

The implementation closely mirrors `entity_registry_completeness` — both iterate `lintFitnessRegistry.AllProductServices()`, accumulate violations, and return a single aggregated error. This consistent pattern makes the code predictable and the tests straightforward.

### Early Return on Missing Config Dir

If `config/` directory is missing, returning early avoids cascading "missing config file" violations for all 4 files. This keeps error output actionable: one clear error instead of 5.

### Test Granularity

Each failure mode (missing Dockerfile, compose.yml, secrets/, config/, each config file) gets its own test. Table-driven tests for Dockerfile and config-file subtypes maintain coverage without duplication.

---

## Phase 5: Compose File Header and Service Name Validation

### What Went Well

- All 3 checks (`compose-header-format`, `compose-service-names`, `compose-db-naming`) implemented and passing in one session.
- Registry-driven pattern: all 10 PS validated automatically — no hardcoding of PS lists.
- 20 tests written covering integration (real workspace), all-correct, missing-file, and wrong-value cases.

### Surprises / Root Causes

- **5 compose headers had drift**: 4 PS (identity-idp/rp/rs/spa) had lowercase PS-ID on line 3; sm-kms had wrong display name on line 5 (`SM Key Management Service` vs. registry `Secrets Manager Key Management`). Discovered only at lint-fitness run time.
- **magic_cicd.go uses TAB indentation**: `replace_string_in_file` with space-indented oldString fails silently with "Could not find matching text". Always match exact whitespace. Workaround: use string concatenation `$tab = "`t"` in PowerShell heredocs.
- **PowerShell heredoc strips tab indentation**: Writing Go files with `@'...'@` in PowerShell 5.1 causes lines to lose leading tabs. Use string concatenation with `${tab}` to build files `requiring` tab indentation.
- **`format-go` hook auto-fixed `interface{}` → `any`**: Pre-commit auto-modified compose_service_names.go. Always re-stage and retry commit after hook modifications.

### Patterns Established

- New magic constants added when literal values have domain semantics: `CICDComposeHeaderLinesToCheck`, `CICDComposeLine3Index`, `CICDComposeLine5Index`, `CICDTempDirPermissions`.
- Test files must use `cryptoutilSharedMagic.FilePermissions` for `0o644` and `cryptoutilSharedMagic.DirPermissions` / `CICDTempDirPermissions` for `0o755`.
- `IME2ESQLiteContainer`, `IME2EPostgreSQL1Container`, `IME2EPostgreSQL2Container` exist in shared/magic for SM-IM E2E container names.
- For the `{PS-ID}-db-postgres-1` db container service suffix, no dedicated constant exists — compute as `psID + "-db-postgres-1"` dynamically.

### Quality Gates Summary

- `go build ./...`: ✅ clean
- `golangci-lint run`: ✅ 0 issues
- `go test ./internal/apps/cicd/lint_fitness/...`: ✅ all pass
- `go run ./cmd/cicd lint-fitness`: ✅ SUCCESS
- `go run ./cmd/cicd lint-go`: ✅ 0 literal-use [blocking] violations

---

## Phase 6: Magic Constants Cross-Reference Validation

### Drift Found

- `magic_identity.go`: `IdentityE2EComposeFile` used 4 `../` levels but the active e2e tests are in `identity/authz/e2e/` (5 levels from root). The constant path was stale from before the authz sub-service restructuring. Fixed to use 5 levels and updated comment to reference `identity/authz/e2e`.

### Implementation Notes

- **Regex approach**: Used `regexp.MustCompile` with `(\w+)\s*=\s*"([^"]+)"` to parse Go constant string assignments. No AST needed — simpler and faster for this validation.
- **Deduplication by MagicFile**: Multiple PS (e.g. all 5 identity PS) share `magic_identity.go`. The checker deduplicates by `ps.MagicFile` so each magic file is scanned exactly once, using the first PS entry in the registry to get `InternalAppsDir`.
- **Skip sentinel**: For `magic-e2e-container-names`, a PS is skipped if no constant with suffix `E2ESQLiteContainer` is found in the magic file. This handles identity (uses per-service naming) and pki-ca (no E2E tests).
- **Path resolution**: For `magic-e2e-compose-path`, the checker resolves `*E2EComposeFile` relative to `rootDir/internal/apps/{InternalAppsDir}e2e/`. The `filepath.Clean(filepath.Join(...))` chain handles `../` traversal correctly.
- **PowerShell backtick issue**: Cannot write Go files containing backtick-delimited regex literals using Python one-liners in PowerShell — PowerShell intercepts backticks before Python runs. Solution: use VS Code file editor tools (`replace_string_in_file`) directly instead of terminal commands.

### Patterns Established

- Magic file checkers follow the same dedup-by-MagicFile pattern to handle shared magic files.
- `magic-e2e-container-names` checks the 3-tuple `{SQLite, PostgreSQL1, PostgreSQL2}Container` per PS; missing constituent = violation when SQLite sentinel is present.
- `magic-e2e-compose-path` verifies E2EComposeFile constants resolve to actual files; constants must be correct relative to `InternalAppsDir/e2e/`.
- Test files write fake magic sources using Go string concatenation (no backtick literals in the test code itself) to avoid BOM and quoting issues.
- `goconst` linter catches duplicate string literals across test functions — extract to package-level constants when a string is used 2+ times.

---

## Phase 7: Standalone Config File Presence and Naming

### Implementation Notes

- **Allowlist pattern**: Only `sm-im` and `sm-kms` use the standardized `configs/{product}/{service}/` layout. Implemented via `configAllowlist` map keyed on `OTLPServiceSMIM` / `OTLPServiceSMKMS` magic constants (no literal strings, lint-go compliant).
- **Two-checker split**: `standalone-config-presence` checks file existence; `standalone-config-otlp-names` checks OTLP values. This mirrors the separation of concerns in `compose_service_names` + `compose_db_naming`.
- **Missing file is not an OTLP error**: `standalone_config_otlp_names` uses `os.IsNotExist` to skip absent files silently — file absence is `standalone-config-presence`'s domain.
- **Registry-driven vs filesystem-scan**: `standalone_config_otlp_names` iterates `AllProductServices()` + allowlist filter, unlike `otlp_service_name_pattern` which scans the filesystem. Both validate the same invariant from different directions.
- **Windows path separator in tests**: `filepath.Rel` returns OS-native separators, so test assertions must use component names (e.g. `cryptoutilSharedMagic.KMSServiceName`) rather than slash-joined paths. Discovered during first test run.
- **No drift found**: Both `sm-im` and `sm-kms` configs already conform to `{ps-id}-sqlite-1`, `{ps-id}-postgres-1`, `{ps-id}-postgres-2` naming.

### Commit

- `697b41951`: `feat(lint-fitness): add standalone config presence and OTLP name checks (Phase 7)` — 5 files, 500 insertions

---

## Phase 8: Migration Comment Header Validation

**Root Causes Found**: 20 domain migration files across 5 product-services had non-conforming comment headers — abbreviated names (`JOSE-JA`, `SM IM`, `KMS Business Tables Migration`) instead of full registry `DisplayName` values.

**Key Lessons**:
1. **`Contains` vs exact-match**: Using `strings.Contains` for the check (not a prefix match) lets files keep additional descriptive text after "database schema" while still enforcing the registry name. This is the right trade-off.
2. **Walk for migrations dirs**: Not all PS use `repository/migrations/` — some use `repository-v2/migrations/`. Walking `internal/apps/{InternalAppsDir}` for any `migrations` directory is robust to this variation.
3. **Archived dirs need skip**: Dirs prefixed with `_` (archived/disabled) must be excluded or false violations accumulate.
4. **Domain min = 2001**: Framework migrations (1001-1999) are intentionally generic — only domain-specific (2001+) migrations are owned per-PS.
5. **banned-product-names catches its own fixer**: A code comment in the checker using an example old name triggered the checker itself. Lesson: use generic descriptions in examples, not actual old product names.
6. **20 violations in one shot**: All found by `TestCheck_RealWorkspace` in a single test run. Write the real-workspace integration test first — it reveals all violations immediately.
7. **SQLFluff ran on commit**: The pre-commit `sqlfluff fix` hook ran over all 20 changed SQL files and passed cleanly — no SQL formatting regressions.

---

## Phase 9: ARCHITECTURE.md Updates and CICD Tool Catalog

**Root Causes Found**: ARCHITECTURE.md Section 9.11 was stale — showed "23 total" when 43 checks existed, "lint_gotest (5)" when it was 4 (the 5th was from lint_skeleton), and "New checks (8)" when there were actually 28 new checks across framework-v3 Phase 4 and framework-v4 Phases 1-8.

**Key Lessons**:
1. **Count discrepancy compounds**: Each new check added without updating ARCHITECTURE.md widens the gap. Updating at Phase 9 with the full catalog is the right approach — maintaining accuracy vs. keeping docs synchronized incrementally are both valid strategies.
2. **Catalog grouping matters**: Separating `lint_go (10)`, `lint_gotest (4)`, `lint_skeleton (1)`, `framework-v3 Phase 4 (15)`, and `framework-v4 (13)` makes the history clear and auditable.
3. **lint-docs validates propagation not counts**: `go run ./cmd/cicd lint-docs` only checks `@propagate`/`@source` marker integrity, not free-form counts. Manual review of counts is still required.
4. **Entity Registry sub-section**: Documenting registry location, types, and update procedure in ARCHITECTURE.md ensures new contributors know the single source of truth and how to extend it.
5. **ci-fitness.yml**: The workflow doesn't need updating — it just runs `lint-fitness` without mentioning a count.

---

## Phase 10: Knowledge Propagation

### Summary

Phase 10 propagated entity registry patterns and banned product name documentation to permanent instruction and skill artifacts.

### What Worked

- Adding `CICDExcludeDirGithubInstructions = ".github"` magic constant and using it in `banned_product_names.go` excludedDirs cleanly excluded `.github/instructions/` from the banned-phrase scan without any test failures in that checker.
- The fitness-function-gen SKILL.md update (23→43 checks + Registry-Driven Check Pattern section) documented the established coding pattern for new contributors.
- lint-docs validate-propagation passed with 270 valid refs, 0 broken after all instruction file updates.

### Lessons Learned

1. **New magic constants propagate immediately to literal-use checker**: When adding a magic constant like `CICDExcludeDirGithubInstructions`, the `magic-usage` linter (`TestLint_Integration`) immediately flags all existing literal usages of that string value as blocking violations. This is expected behavior — the fix is to update all usages to reference the new constant (not add exceptions). In this case, 17 usages of `".github"` across `lint_workflow` and `workflow` packages required updating.

2. **replace_string_in_file tool can lose content when oldString includes trailing code**: If the `oldString` matches content that is followed immediately by other code on the same line (due to a previous bad edit), the tool will find no match. Use PowerShell `[System.IO.File]::ReadAllText()` + `.Replace()` + `[System.IO.File]::WriteAllText()` for robust string replacement when tool-based replacement fails.

3. **Documentation files need scanner exclusions for banned phrases**: Instruction files and skill files that document why certain names are banned will necessarily contain those banned phrases. Add the parent directory to the excluded dirs list using a dedicated magic constant (same pattern as `CICDExcludeDirBannedProductNamesCheck`).

4. **Verify build immediately after each replace_string_in_file call**: A single mismatched replacement can silently lose a line or merge adjacent lines. `go build ./...` is the fastest verification — any undefined symbol immediately reveals corruption before it compounds.

### Key Metrics

- Files changed in Phase 10: 13 (1 magic constant, 1 banned-names checker, 7 test files, 1 implementation file, 1 instruction doc, 1 skill doc, 1 plan.md)
- Blocking violations fixed: 17 (all `".github"` literal-use in lint_workflow + workflow packages)
- Quality gate: all 43 lint_fitness checks pass, `go test ./internal/apps/cicd/...` all OK, `TestLint_Integration` OK
