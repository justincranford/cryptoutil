// Copyright (c) 2025 Justin Cranford

// Package common provides shared utilities for lint_gotest subpackages.
package common

import "strings"

// FilterExcludedTestFiles filters out test files that should be excluded from linting.
func FilterExcludedTestFiles(testFiles []string) []string {
filteredTestFiles := make([]string, 0, len(testFiles))

for _, path := range testFiles {
// Exclude cicd test files as they contain deliberate patterns for testing cicd functionality.
// Also exclude edge_cases_test.go, testmain_test.go, e2e_test.go, sessions_test.go, admin_test.go.
if strings.HasSuffix(path, "cicd_test.go") ||
strings.HasSuffix(path, "cicd.go") ||
strings.HasSuffix(path, "cicd_enforce_test_patterns_test.go") ||
strings.HasSuffix(path, "cicd_enforce_test_patterns_integration_test.go") ||
strings.HasSuffix(path, "cicd_run_integration_test.go") ||
strings.Contains(path, "lint_gotest") ||
strings.HasSuffix(path, "_edge_cases_test.go") ||
strings.HasSuffix(path, "testmain_test.go") ||
strings.HasSuffix(path, "e2e_test.go") ||
strings.HasSuffix(path, "sessions_test.go") ||
strings.HasSuffix(path, "admin_test.go") {
continue
}

filteredTestFiles = append(filteredTestFiles, path)
}

return filteredTestFiles
}
