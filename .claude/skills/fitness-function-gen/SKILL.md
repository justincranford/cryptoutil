---
name: fitness-function-gen
description: "Generate a new architecture fitness function (linter) for the cryptoutil lint-fitness framework. Use when adding a new architectural invariant that should be enforced via go run ./cmd/cicd-lint lint-fitness across every service."
argument-hint: "[linter-name] [architectural rule description]"
---

Generate a new architecture fitness function for the cryptoutil lint-fitness framework.

## Purpose

Use this skill when an architectural rule from `docs/ENG-HANDBOOK.md` must be enforced by `go run ./cmd/cicd-lint lint-fitness` rather than by review alone.

- Adding a new architectural rule from ENG-HANDBOOK.md that must be enforced programmatically
- Migrating a soft architectural guideline to a hard enforced check
- Extending compliance checking for a new pattern (e.g., new file naming conventions)

Use `psid-template-sync` instead when the change is only a template-instantiation update and does not require a new linter.

## Key Rules

- Register the new checker in `internal/apps-tools/cicd_lint/lint_fitness/lint_fitness.go`
- Export `Check(logger *cryptoutilCmdCicdCommon.Logger) error` and a testable `CheckInDir(...)` or equivalent helper
- MUST return hard error (`fmt.Errorf`) on absent required directories (never `return nil`)
- Prefer `fs.FS`, `io.Reader`, or explicit function parameters for filesystem and input seams so error paths are unit-testable
- Tests ≥98% line coverage (infrastructure/utility target)
- Validator error aggregation: collect ALL violations before returning (never short-circuit)
- Run the checker against the real workspace before committing it so pre-existing violations are fixed in the same change

## Fitness Function Registration

Every fitness function MUST:
1. Live in internal/apps-tools/cicd_lint/lint_fitness/<linter-name>/
2. Export a Check(logger *cryptoutilCmdCicdCommon.Logger) error function
3. Be registered in internal/apps-tools/cicd_lint/lint_fitness/lint_fitness.go
4. Achieve =98% test coverage (infrastructure/utility target)

## Directory Structure

```text
internal/apps-tools/cicd_lint/lint_fitness/
+-- lint_fitness.go
+-- your-linter-name/
    +-- your_linter_name.go
    +-- your_linter_name_test.go
```

## Implementation Template

```go
// Package your_linter_name enforces ENG-HANDBOOK.md Section X.Y.
package your_linter_name

import (
    "fmt"

    cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
)

func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
    return CheckInDir(logger, ".")
}

func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
    logger.Log("Checking [rule]...")

    var violations []string

    // Walk files and collect violations.

    if len(violations) > 0 {
        for _, violation := range violations {
            logger.Log(fmt.Sprintf("VIOLATION: %s", violation))
        }

        return fmt.Errorf("[rule] check found %d violation(s)", len(violations))
    }

    logger.Log("[rule] check passed")

    return nil
}
```

## Registration in lint_fitness.go

Add to the `registeredLinters` slice in `internal/apps-tools/cicd_lint/lint_fitness/lint_fitness.go`:

```go
import (
    // ... existing imports
    lintFitnessYourLinter "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/your-linter-name"
)

var registeredLinters = []struct { name string; linter LinterFunc }{
    // ... existing linters
    {"your-linter-name", lintFitnessYourLinter.Check}, // Add here
}
```

## Test Template

```go
package your_linter_name

import (
    "os"
    "path/filepath"
    "testing"

    cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
    cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
    "github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
    return cryptoutilCmdCicdCommon.NewLogger("test")
}

func TestCheckInDir_CompliantFile_Passes(t *testing.T) {
    t.Parallel()

    tmp := t.TempDir()
    require.NoError(t, os.WriteFile(
        filepath.Join(tmp, "compliant.go"),
        []byte("package foo\n// compliant content\n"),
        cryptoutilSharedMagic.FilePermissionsDefault,
    ))

    require.NoError(t, CheckInDir(newTestLogger(), tmp))
}

func TestCheckInDir_ViolatingFile_Fails(t *testing.T) {
    t.Parallel()

    tmp := t.TempDir()
    require.NoError(t, os.WriteFile(
        filepath.Join(tmp, "violating.go"),
        []byte("package foo\n// violating content\n"),
        cryptoutilSharedMagic.FilePermissionsDefault,
    ))

    err := CheckInDir(newTestLogger(), tmp)
    require.Error(t, err)
    require.Contains(t, err.Error(), "violation")
}

func TestCheck_RealWorkspace_Passes(t *testing.T) {
    t.Parallel()

    require.NoError(t, Check(newTestLogger()))
}
```

## Registry-Driven Check Pattern

For checks that must validate EVERY product-service uniformly, use the registry-driven pattern instead of hardcoding names:

```go
import (
    lintFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
)

func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
    logger.Log("Checking [rule]...")
    var violations []string
    for _, ps := range lintFitnessRegistry.AllProductServices() {
        // Check each PS using ps.PSID, ps.DisplayName, ps.InternalAppsDir, etc.
        psDir := filepath.Join(rootDir, "internal", "apps", ps.InternalAppsDir)
        if err := checkPS(ps, psDir); err != nil {
            violations = append(violations, err.Error())
        }
    }
    if len(violations) > 0 {
        for _, v := range violations { logger.Log(fmt.Sprintf("VIOLATION: %s", v)) }
        return fmt.Errorf("[rule] found %d violation(s)", len(violations))
    }
    logger.Log("[Rule] check passed")
    return nil
}
```

**Registry fields**: `ps.PSID` (e.g. `sm-kms`), `ps.Product`, `ps.Service`, `ps.DisplayName` (e.g. `Secrets Manager Key Management`), `ps.InternalAppsDir` (e.g. `sm-kms/`), `ps.MagicFile`.

**When to use registry-driven**: When the rule applies to all product-services (naming patterns, config presence, migration headers, compose structure). When the rule is service-specific or cross-cutting, use the simpler `rootDir` walk pattern.

**Real-workspace test is mandatory**: Add `TestCheck_RealWorkspace` that calls `Check(logger)` against the actual workspace. This test reveals existing violations before the check is first committed, so fix those violations in the same change.

## Critical Notes

- **CheckInDir pattern**: Always separate Check (calls .) from CheckInDir (parameterized root). Tests use CheckInDir(logger, tmp) for isolation.
- **Error aggregation**: NEVER short-circuit. Collect ALL violations before returning. Report them all, then return one consolidated error.
- **File permissions**: Use `cryptoutilSharedMagic.FilePermissionsDefault` for test files (0o600). Use `cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute` for directories (0o755). Never use raw octal literals — the `magic-usage` linter enforces this.
- **t.Parallel()**: MANDATORY on all tests EXCEPT those using os.Chdir. Add // Sequential: comment for those.
- **The fitness check runs on CI**: Adding a linter that fails on existing code is a CI blocker. Always test against the actual codebase root first.

## After Creation

1. Run `go run ./cmd/cicd-lint lint-fitness` and require it to pass with the new linter registered.
2. Run `go test ./internal/apps-tools/cicd_lint/lint_fitness/...` and keep coverage at or above 98% for the touched package set.
3. Update lint_fitness_test.go TestLint_Success count if it has a hardcoded linter count.
4. Commit with `ci(cicd): add [linter-name] fitness function`.

## References

Read [ENG-HANDBOOK.md Section 9.10 CICD Command Architecture](../../../docs/ENG-HANDBOOK.md#910-cicd-command-architecture) for checker registration and command boundaries.

Read [ENG-HANDBOOK.md Section 9.11 Architecture Fitness Functions](../../../docs/ENG-HANDBOOK.md#911-architecture-fitness-functions) for the existing fitness-linter model and registry-driven enforcement approach.

Read [ENG-HANDBOOK.md Section 10.2.5 Sequential Test Exemption](../../../docs/ENG-HANDBOOK.md#1025-sequential-test-exemption) for the `// Sequential:` exception.

Read [ENG-HANDBOOK.md Section 11.3 Code Quality Standards](../../../docs/ENG-HANDBOOK.md#113-code-quality-standards) for the 98% infrastructure coverage target.
