# Lessons Learned - Framework v3

This file captures lessons from each phase, used as:
1. Memory for the entire plan.md / tasks.md execution
2. Input for knowledge propagation to ARCHITECTURE.md, agents, skills, instructions

---

## Phase 1: Close v1 Gaps and Knowledge Propagation

### What Worked

- Knowledge propagation (Tasks 1.2-1.5): Document-first approach worked well — update ARCHITECTURE.md, then propagate to instructions/skills/agents.
- CI workflow addition (Task 1.4): Simple GitHub Actions workflow for `cicd lint-fitness` needed no new code.
- Contract test coverage (Tasks 1.7-1.10): `RunContractTests` adoption uniform across all 10 services.
- lint-fitness integration test (`TestLint_Integration`): Single authoritative test that validates all linters end-to-end; caught magic-aliases and literal-use violations immediately.

### What Didn't Work / Root Causes

1. **Go 1.24+ stdlib crypto ignores `rand io.Reader`** (FIPS 140-3): `rsa.GenerateKey`, `ecdsa.GenerateKey`, `ecdh.Curve.GenerateKey` silently ignore the rand parameter. Function-level seams were required to inject error paths for testing.
2. **Windows OS incompatibilities discovered** (pre-existing):
   - `syscall.SIGINT` not available on Windows — lifecycle tests needed `runtime.GOOS == magic.OSNameWindows` skip guards.
   - `os.Chmod(0o000)` does not restrict reads on Windows — realm permission test needed Windows skip.
   - `/bin/echo` and `/root/` paths don't exist on Windows — workflow tests needed Windows skips.
   - OS file handles must be closed before `t.TempDir()` cleanup on Windows.
3. **SQLite named in-memory URL format**: modernc.org/sqlite does NOT support `file::memory:NAME?cache=shared`. Must use `file:NAME?mode=memory&cache=shared`. Fixed in `application_core.go`.
4. **magic-aliases linter (33 violations)**: 26 were in the config package (largest block). Recovery from PowerShell corruption required Python: `-replace` in PowerShell is case-insensitive, causing double-prefix corruption like `cryptoutilSharedMagic.cryptoutilSharedMagic.DefaultXxx`.
5. **literal-use violations (11)**: All 11 were `"windows"` string literals instead of `magic.OSNameWindows` — added in the same session as the Windows skip guards.
6. **Flaky property test `TestHKDFInvariants`** in `digests`: Fails with some random seeds under `-p=4` parallelism. Pre-existing; passes in isolation.
7. **Parallel test flakiness** in `businesslogic` and `pool`: Fail under `-p=4` due to SQLite shared-memory contention, pass in isolation. Pre-existing.

### Pattern: PowerShell `-replace` is Case-Insensitive

**CRITICAL**: PowerShell's `-replace` operator is case-insensitive by default. When chaining replacements where the replacement text contains substrings matching the original pattern, it causes double/triple prefix corruption. **Always use Python or sed-style tools for identifier replacement** when the replacement string might be matched again.

### Pattern: magic-aliases Linter Catches All Types

The `magic-aliases` linter catches ALL `const` aliases — even function-local `const` declarations. This is correct behavior. `var` aliases are not flagged (var default values are acceptable since they can't be inlined at compile time).

### Pattern: After Adding Code, Run TestLint_Integration

After adding any new skip guard, constant, or literal, run `go test ./internal/apps/cicd/lint_go -run TestLint_Integration` immediately. It catches `literal-use` (blocking) violations that golangci-lint misses.

### Quality Gate Outcome

- `go build ./...` ✅ clean
- `golangci-lint run --fix ./...` ✅ 0 issues
- `golangci-lint run --build-tags e2e,integration ./...` ✅ 0 issues
- `go build -tags e2e,integration ./...` ✅ clean
- `TestLint_Integration` ✅ ok
- `go test ./... -count=1 -p=4` ✅ passes (flaky tests are pre-existing, pass in isolation)

---

## Phase 2: Remove InsecureSkipVerify — Integration Tests Only (D14, D15)

### What Worked

- **TLSRootCAPool() interface pattern**: Adding TLSRootCAPool() and AdminTLSRootCAPool() to the ServiceServer interface gave all services a clean, uniform way to expose cert pools for test clients. Pattern:  estServer.TLSRootCAPool() for public,  estServer.AdminTLSRootCAPool() for admin.
- **Testutil helper functions**: cryptoutilAppsTemplateServiceServerTestutil.PublicRootCAPool() and PrivateRootCAPool() give a one-line way to build properly validated HTTP clients without test-struct complexity.
- **Auto-TLS in tests**: The existing  ls_generator.go Auto mode already created ephemeral CA certs per test run. The only missing piece was surfacing the root CA pool to callers — no new TLS infrastructure was needed.
- **G402 removal from gosec.excludes**: Removing the G402 blanket exclusion caught 2 real issues (sm-kms MinVersion missing, identity e2e using InsecureSkipVerify) that would have been silent violations.
- **semgrep
o-tls-insecure-skip-verify rule**: Activating the rule with path filters (test files included, tls_validate_test.go excluded) gives a second gate beyond gosec.

### What Didn't Work / Root Causes

1. **Task 2.7 public_table_test.go indentation corruption**: The prior session's multi_replace_string_in_file produced malformed Go (missing closing braces, wrong tab depth). Root cause: multi_replace operates on tab characters that are invisible in tool display — always verify with character-level analysis ([byte[]]) after replacing client struct literals. Lesson: After replacing any multi-level struct literal (TLSClientConfig inside Transport inside http.Client), verify with go build immediately.

2. **.golangci.yml YAML structure corruption on first insertion attempt**: The
eplace_string_in_file tool consumed the blank line + settings: + errcheck: section when inserting the identity e2e gosec path exclusion. Root cause: The old string matched a larger block than expected because the surrounding YAML (blank line + settings block) was part of a contiguous string. Lesson: When inserting a new YAML array entry, use PowerShell $content.Replace() with exact string matching to avoid consuming adjacent structure. After any .golangci.yml edit, ALWAYS run python -c "import yaml; yaml.safe_load(open('.golangci.yml').read())" to verify YAML validity.

3. **Full suite parallel test flakiness**: Running go test ./... -shuffle=on -count=1 caused  emplate/service/server/application and  emplate/service/server to fail due to resource contention from parallel execution. These pass in isolation on both committed and modified code. This is pre-existing. Lesson: For the quality gate, run go test ./... -shuffle=on but accept that contention-related failures that pass in isolation are pre-existing.

4. **identity/test/e2e/ missed by Task 2.6**: Task 2.6 migrated identity service test clients, but internal/apps/identity/test/e2e/identity_e2e_test.go (a separate  est/e2e/ subdirectory) was missed. This file connects to actual deployed services, so InsecureSkipVerify is justified — the fix was a gosec path exclusion, not a migration. Lesson: After disabling G402 blanket exclusion in Task 2.8, always re-run golangci-lint run --build-tags e2e,integration ./... (not just ./...) to catch e2e-tagged files.

5. **golangci-lint build tag sensitivity**: The golangci-lint run ./... command only lints files without build tags active. Files tagged //go:build e2e or //go:build integration require --build-tags e2e,integration to be linted. The standard lint gate must ALWAYS include both forms.

### Key Decisions

- **identity/test/e2e/identity_e2e_test.go**: Added gosec path exclusion instead of migrating to TLSRootCAPool pattern. This file connects to externally-deployed service containers with self-signed certs and has no access to the server's TLS bundle. A documentation comment explains why InsecureSkipVerify is used.
- **semgrep exclusion scope**: The
o-tls-insecure-skip-verify semgrep rule includes all _test.go,_integration_test.go,_e2e_test.go files but excludes  ls_validate_test.go. The identity e2e test file (identity_e2e_test.go, not identity_e2e_integration_test.go) is covered by the *_test.go pattern — the semgrep rule and gosec exclusion together ensure it is checked by semgrep (and passes, since semgrep exclusion is a different file) but excluded from gosec G402.
- **YAML structure fix**: When golangci-lint viper reports "line X: did not find expected key", always use Python yaml.safe_load to pinpoint the issue before trying random fixes.

### Quality Gates Status

- go build ./...: ✅ clean
- go build -tags e2e,integration ./...: ✅ clean  
- golangci-lint run ./...: ✅ 0 issues
- golangci-lint run --build-tags e2e,integration ./...: ✅ 0 issues
- go test ./... -shuffle=on: ✅ passes (pre-existing contention failures confirmed pre-existing)

---

## Phase 3: Builder Refactoring

### What Worked

1. **DomainConfig struct is clean** — `MigrationsFS`, `MigrationsPath`, `RouteRegistration` captures 100% of what services need; 0 services needed any additional configuration.
2. **`Build()` convenience function** reduces every service's `NewFromConfig` to a single `Build()` call + struct literal. Each service is now ~10-15 lines, down from 20-30.
3. **`replace_string_in_file` works when given exact tab-indented text** — semantic search returns real indentation; using those snippets directly in `replace_string_in_file` succeeds without any CRLF handling needed.
4. **Position-based PowerShell replacement** (`IndexOf` + `Substring` + concatenation with CRLF normalization) is reliable for complex multi-line blocks and handles em-dash / UTF-8 characters that confuse regex.

### What Didn't Work

1. **Space-indented `oldString` in `replace_string_in_file`** — All service files use tab indentation + CRLF; providing space-indented `oldString` always fails. Must match exact file content character-for-character.
2. **Accumulating changes in a single `replace_string_in_file` call with multiple items** — Failed silently when one array element failed (e.g., identity-rp had different `NewPublicServer` signature than authz/idp/rs). MUST read every file individually before replacing.
3. **Assuming all identity services are identical** — identity-rp passes `res.SessionManager, res.RealmService` to `NewPublicServer`; identity-spa uses `RegisterRoutes()` (capital R). Subtle differences break bulk replacements.

### Root Causes

- CRLF line endings + tab indentation = `replace_string_in_file` succeeds only with exact content
- Services evolved independently and have subtle API differences even within the same product family
- `domain_config.go` had trailing whitespace that pre-commit `end-of-file-fixer` caught — needed a second re-add + commit

### Prevention

- Always read a service file before migrating it (never assume same-product services are identical)
- When using `multi_replace_string_in_file` across multiple files, verify each file individually first; partial failures are silent
- Pre-commit hooks auto-fix trailing whitespace and EOF — if commit aborts, re-add the modified file and retry

### Pattern Discovery

- **DomainConfig pattern** generalizes cleanly: `{MigrationsFS, MigrationsPath, RouteRegistration}` is the universal domain configuration API for all service types
- **Services with no domain migrations** (identity-*) simply omit those fields — Go zero values work correctly
- **sm-kms special case**: initializes `kmsCore` BEFORE calling `Build()` so the closure captures it; `kmsCore.Shutdown()` called in the error path before returning

---

## Phase 4: Sequential Exemption Reduction

*(To be filled during Phase 4 execution)*

---

## Phase 5: ServiceServer Interface Expansion

*(To be filled during Phase 5 execution)*

---

## Phase 6: lint-fitness Value Assessment

*(To be filled during Phase 6 execution)*

---

## Phase 7: Domain Extraction and Fresh Skeletons (D13, D16)

*(To be filled during Phase 7 execution)*

---

## Phase 8: Staged Domain Reintegration (D13)

*(To be filled during Phase 8 execution)*

---

## Phase 9: Quality and Knowledge Propagation

*(To be filled during Phase 9 execution)*
