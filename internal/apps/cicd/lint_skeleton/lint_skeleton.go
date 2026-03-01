// Copyright (c) 2025 Justin Cranford

// Package lint_skeleton detects unreplaced skeleton template placeholder strings
// in Go source files that are outside the canonical skeleton-template directories.
// When a developer copies the skeleton-template service to create a new service,
// they MUST rename all occurrences of 'skeleton' to their new service name.
package lint_skeleton

import (
"errors"
"fmt"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
skeletonPlaceholders "cryptoutil/internal/apps/cicd/lint_skeleton/check_skeleton_placeholders"
)

// LinterFunc is the signature for lint_skeleton sub-linters.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger) error

// registeredLinters holds the ordered list of skeleton linters.
var registeredLinters = []struct {
name   string
linter LinterFunc
}{
{"check-skeleton-placeholders", skeletonPlaceholders.Check},
}

// Lint runs all registered skeleton linters sequentially.
// Continues on failure, collecting all errors before returning.
func Lint(logger *cryptoutilCmdCicdCommon.Logger) error {
var errs []error

for _, l := range registeredLinters {
logger.Log(fmt.Sprintf("Running %s", l.name))

if err := l.linter(logger); err != nil {
logger.LogError(err)
errs = append(errs, fmt.Errorf("%s: %w", l.name, err))
} else {
logger.Log(fmt.Sprintf("  ✅ %s passed", l.name))
}
}

if len(errs) > 0 {
return fmt.Errorf("lint-skeleton failed: %w", errors.Join(errs...))
}

return nil
}
