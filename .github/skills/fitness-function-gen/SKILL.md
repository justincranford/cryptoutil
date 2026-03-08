---
name: fitness-function-gen
description: "Generate a new architecture fitness function (linter) for the cryptoutil lint-fitness framework. Use when adding a new architectural invariant that should be enforced via go run ./cmd/cicd lint-fitness across every service."
argument-hint: "[linter-name] [architectural rule description]"
---

Generate a new architecture fitness function for the cryptoutil lint-fitness framework.

## Purpose

The lint-fitness framework runs 23 architectural invariant checks on every CI push. Use this skill when:

- Adding a new architectural rule from ARCHITECTURE.md that must be enforced programmatically
- Migrating a soft architectural guideline to a hard enforced check
- Extending compliance checking for a new pattern (e.g., new file naming conventions)

## Fitness Function Registration

Every fitness function MUST:
1. Live in internal/apps/cicd/lint_fitness/<linter-name>/
2. Export a Check(logger *cryptoutilCmdCicdCommon.Logger) error function
3. Be registered in internal/apps/cicd/lint_fitness/lint_fitness.go
4. Achieve =98% test coverage (infrastructure/utility target)

## Directory Structure

`
internal/apps/cicd/lint_fitness/
+-- lint_fitness.go                 # Registration + Lint() orchestrator
+-- your-linter-name/               # kebab-case directory
    +-- your_linter_name.go         # Implementation: package your_linter_name
    +-- your_linter_name_test.go    # Tests (95%+)
`

## Implementation Template

`go
// Copyright (c) 2025 Justin Cranford

// Package your_linter_name enforces ARCHITECTURE.md Section X.Y rule name.
// Brief description of what this linter enforces.
package your_linter_name

import (
"fmt"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// Check enforces [rule] from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
return CheckInDir(logger, ".")
}

// CheckInDir enforces [rule] under rootDir. Separate for testability.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
logger.Log("Checking [rule]...")

var violations []string

// Walk files and collect violations
// ...

if len(violations) > 0 {
for _, v := range violations {
logger.Log(fmt.Sprintf("VIOLATION: %s", v))
}
return fmt.Errorf("[rule] check found %d violation(s)", len(violations))
}

logger.Log("[Rule] check passed")
return nil
}
`

## Registration in lint_fitness.go

Add to the `registeredLinters` slice in `internal/apps/cicd/lint_fitness/lint_fitness.go`:

`go
import (
    // ... existing imports
    lintFitnessYourLinter "cryptoutil/internal/apps/cicd/lint_fitness/your-linter-name"
)

var registeredLinters = []struct { name string; linter LinterFunc }{
    // ... existing linters
    {"your-linter-name", lintFitnessYourLinter.Check}, // Add here
}
`

## Test Template

`go
// Copyright (c) 2025 Justin Cranford

package your_linter_name

import (
"os"
"path/filepath"
"testing"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
return cryptoutilCmdCicdCommon.NewLogger("test")
}

func TestCheckInDir_CompliantFile_Passes(t *testing.T) {
t.Parallel()
tmp := t.TempDir()
// Write a file that PASSES the rule
require.NoError(t, os.WriteFile(
filepath.Join(tmp, "compliant.go"),
[]byte("package foo\n// compliant content\n"),
cryptoutilSharedMagic.CacheFilePermissions,
))
require.NoError(t, CheckInDir(newTestLogger(), tmp))
}

func TestCheckInDir_ViolatingFile_Fails(t *testing.T) {
t.Parallel()
tmp := t.TempDir()
// Write a file that VIOLATES the rule
require.NoError(t, os.WriteFile(
filepath.Join(tmp, "violating.go"),
[]byte("package foo\n// violating content\n"),
cryptoutilSharedMagic.CacheFilePermissions,
))
err := CheckInDir(newTestLogger(), tmp)
require.Error(t, err)
require.Contains(t, err.Error(), "violation")
}

func TestCheck_WorkspaceRoot_Passes(t *testing.T) {
t.Parallel()
// Runs against the actual workspace - should always pass
require.NoError(t, Check(newTestLogger()))
}
`

## Critical Notes

- **CheckInDir pattern**: Always separate Check (calls .) from CheckInDir (parameterized root). Tests use CheckInDir(logger, tmp) for isolation.
- **Error aggregation**: NEVER short-circuit. Collect ALL violations before returning. Report them all, then return one consolidated error.
- **File permissions**: Use cryptoutilSharedMagic.CacheFilePermissions for test files.
- **t.Parallel()**: MANDATORY on all tests EXCEPT those using os.Chdir. Add // Sequential: comment for those.
- **The fitness check runs on CI**: Adding a linter that fails on existing code is a CI blocker. Always test against the actual codebase root first.

## After Creation

1. Run go run ./cmd/cicd lint-fitness � must pass with your new linter included.
2. Run tests: go test ./internal/apps/cicd/lint_fitness/... � must achieve =98% coverage.
3. Update lint_fitness_test.go TestLint_Success count if it has a hardcoded linter count.
4. Commit with ci(cicd): add [linter-name] fitness function.

## References

Read [ARCHITECTURE.md Section 9.10](../../../docs/ARCHITECTURE.md#910-cicd-command-architecture) for CICD command architecture.
Read [ARCHITECTURE.md Section 10.2.5](../../../docs/ARCHITECTURE.md#1025-sequential-test-exemption) for // Sequential: comment exemption.
Read [ARCHITECTURE.md Section 11.3](../../../docs/ARCHITECTURE.md#113-code-quality-standards) for test coverage targets (=98% for infrastructure/utility code).
