---
name: fitness-function-gen
description: "Generate a new architecture fitness function (linter) for the cryptoutil lint-fitness framework. Use when adding a new architectural invariant that should be enforced via go run ./cmd/cicd-lint lint-fitness across every service."
argument-hint: "[linter-name] [architectural rule description]"
---

Create a new architecture fitness function (linter) for the cicd_lint/lint_fitness framework.

**Full Copilot original**: [.github/skills/fitness-function-gen/SKILL.md](.github/skills/fitness-function-gen/SKILL.md)

Provide: linter name (e.g., `check-migration-ranges`), what invariant it checks.

## Key Rules

- Register in `registeredLinters` slice in `lint_fitness.go` and `lint-fitness-registry.yaml`
- Function signature: `Check(logger *cryptoutilCmdCicdCommon.Logger) error`
- MUST return hard error (`fmt.Errorf`) on absent required directories (never `return nil`)
- Use function-param injection for OS I/O (never package-level `var walkFn = filepath.Walk`)
- Tests ≥98% line coverage (infrastructure/utility target)
- Validator error aggregation: collect ALL violations before returning (never short-circuit)

## Directory Structure

```
internal/apps/tools/cicd_lint/lint_fitness/{linter-name}/
├── {linter_name}.go       # Check function
└── {linter_name}_test.go  # ≥98% coverage required
```

## Required Exports

```go
// {linter_name}.go
package {linter_name}

import "log/slog"

// Check validates the architectural invariant from the repo root.
func Check(logger *slog.Logger) error {
    return CheckInDir(logger, ".")
}

// CheckInDir validates the invariant from a specific root directory.
func CheckInDir(logger *slog.Logger, rootDir string) error {
    // Implementation
    return nil
}
```

## Registration

In `internal/apps/tools/cicd_lint/lint_fitness/lint_fitness.go`:
```go
import cryptoutilToolsCicdLintLintFitness{Name} "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/{linter-name}"

var registeredLinters = []linter{
    // ... existing linters ...
    {name: "{linter-name}", fn: cryptoutilToolsCicdLintLintFitness{Name}.Check},
}
```

## Registry-Driven Pattern (for PS-ID uniform checks)

```go
func CheckInDir(logger *slog.Logger, rootDir string) error {
    registryPath := filepath.Join(rootDir, "api", "cryptosuite-registry", "registry.yaml")
    registry, err := loadRegistry(registryPath)
    if err != nil {
        return fmt.Errorf("load registry: %w", err)
    }

    var errs []error
    for _, psID := range registry.ProductServices {
        dir := filepath.Join(rootDir, "internal", "apps", psID.InternalAppsDir)
        if err := checkPSID(logger, psID, dir); err != nil {
            errs = append(errs, err)
        }
    }

    return errors.Join(errs...)
}
```

## 55 Existing Checks

The fitness framework already validates 55 invariants. Before creating a new one, run:
```bash
go run cmd/cicd-lint/main.go lint-fitness
```
to see what is already checked, and avoid duplicates.

## Test Requirements

- ≥98% coverage (infrastructure/utility target)
- Test with both valid and invalid directory structures
- Table-driven tests with t.Parallel()
