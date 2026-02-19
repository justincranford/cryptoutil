// Copyright (c) 2025 Justin Cranford

package lint_magic

import (
"fmt"
"strings"
)

// AllResult aggregates the results of both validators.
type AllResult struct {
// Valid is false when either validator found issues.
Valid bool

// Duplicates is the result from ValidateDuplicates.
Duplicates *DuplicatesResult

// Usage is the result from ValidateUsage.
Usage *UsageResult

// Errors lists top-level errors that prevented a validator from running.
Errors []string
}

// ValidateAll runs ValidateDuplicates and ValidateUsage sequentially and
// returns an aggregated AllResult.  Both validators are always run to
// completion so that all problems are visible in a single pass.
func ValidateAll(magicDir, rootDir string) (*AllResult, error) {
result := &AllResult{Valid: true}

dupResult, err := ValidateDuplicates(magicDir)
if err != nil {
result.Errors = append(result.Errors, fmt.Sprintf("validate-duplicates error: %v", err))
result.Valid = false
} else {
result.Duplicates = dupResult
if !dupResult.Valid {
result.Valid = false
}
}

usageResult, err := ValidateUsage(magicDir, rootDir)
if err != nil {
result.Errors = append(result.Errors, fmt.Sprintf("validate-usage error: %v", err))
result.Valid = false
} else {
result.Usage = usageResult
if !usageResult.Valid {
result.Valid = false
}
}

return result, nil
}

// FormatAllResult formats both validator results into a single human-readable
// CI/CD report, consistent with the lint-deployments output style.
func FormatAllResult(result *AllResult) string {
var sb strings.Builder

if len(result.Errors) > 0 {
for _, e := range result.Errors {
fmt.Fprintf(&sb, "ERROR: %s\n", e)
}

fmt.Fprint(&sb, "\n")
}

if result.Duplicates != nil {
fmt.Fprint(&sb, FormatDuplicatesResult(result.Duplicates))
}

if result.Usage != nil {
fmt.Fprint(&sb, FormatUsageResult(result.Usage))
}

fmt.Fprint(&sb, "\n")

if result.Valid {
fmt.Fprint(&sb, "validate-all: OK\n")
} else {
fmt.Fprint(&sb, "validate-all: FAIL\n")
}

return sb.String()
}
