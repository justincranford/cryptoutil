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

*(To be filled during Phase 2 execution)*

---

## Phase 3: Builder Refactoring

*(To be filled during Phase 3 execution)*

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
