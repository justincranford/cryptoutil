// Copyright (c) 2025 Justin Cranford

// Package no_unit_test_real_db enforces that unit tests do not create real PostgreSQL database containers.
// Unit tests MUST use testdb.NewInMemorySQLiteDB() instead.
// See ARCHITECTURE.md Section 10.3 TestMain Pattern (Forbidden Pattern #3).
package no_unit_test_real_db

import (
"bufio"
"fmt"
"io/fs"
"os"
"path/filepath"
"strings"

cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// bannedPatterns contains lowercase patterns indicating a real PostgreSQL container is being started.
// These are banned in unit tests outside TestMain - use testdb.NewInMemorySQLiteDB() for unit tests
// or place in TestMain for per-package setup.
var bannedPatterns = []string{
"postgres.runcontainer(",
"postgresmodule.run(",
".newpostgrestestcontainer(",
}

// allowedSuffixes lists test file suffixes that are permitted to create real DB containers.
var allowedSuffixes = []string{
"_integration_test.go",
"_e2e_test.go",
}

// allowedPathFragments lists path segments permitted to create real DB containers.
var allowedPathFragments = []string{
"testing/testdb/",
"lint_fitness/",
"lint_gotest/",
}

// Check walks all test files in "." and reports real-database violations.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
return CheckInDir(logger, ".")
}

// CheckInDir walks test files from rootDir and reports violations.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
logger.Log("Checking for unit tests creating real database containers...")

var testFiles []string

err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
if walkErr != nil {
return walkErr
}

if d.IsDir() {
switch d.Name() {
case cryptoutilSharedMagic.CICDExcludeDirGit, cryptoutilSharedMagic.CICDExcludeDirVendor:
return filepath.SkipDir
}

return nil
}

if !strings.HasSuffix(path, "_test.go") {
return nil
}

normalized := filepath.ToSlash(path)

for _, suffix := range allowedSuffixes {
if strings.HasSuffix(normalized, suffix) {
return nil
}
}

for _, fragment := range allowedPathFragments {
if strings.Contains(normalized, fragment) {
return nil
}
}

testFiles = append(testFiles, path)

return nil
})
if err != nil {
return fmt.Errorf("walking test files: %w", err)
}

return CheckFiles(logger, testFiles)
}

// CheckFiles checks the provided list of test files for real-database violations.
func CheckFiles(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
if len(testFiles) == 0 {
logger.Log("No unit test files to check for real database containers")

return nil
}

totalViolations := 0

for _, filePath := range testFiles {
issues := CheckFile(filePath)

if len(issues) > 0 {
for _, issue := range issues {
fmt.Fprintf(os.Stderr, "%s: %s\n", filePath, issue)
}

totalViolations += len(issues)
}
}

if totalViolations > 0 {
logger.Log(fmt.Sprintf("Found %d real-database container violation(s) in unit tests", totalViolations))
fmt.Fprintln(os.Stderr, "Use testdb.NewInMemorySQLiteDB() for unit tests or place containers in TestMain.")
fmt.Fprintln(os.Stderr, "See ARCHITECTURE.md Section 10.3 TestMain Pattern.")

return fmt.Errorf("found %d real-database container violation(s) in unit tests", totalViolations)
}

logger.Log("\u2705 No real database containers found in unit tests")

return nil
}

// CheckFile scans a single file for banned real-database container patterns.
// Lines inside a TestMain function block are exempt - TestMain is the approved location
// for per-package heavyweight resource setup per ARCHITECTURE.md Section 10.3.
func CheckFile(filePath string) []string {
content, err := os.ReadFile(filePath)
if err != nil {
return []string{fmt.Sprintf("error reading file: %v", err)}
}

var violations []string

scanner := bufio.NewScanner(strings.NewReader(string(content)))
lineNum := 0
braceDepth := 0
insideTestMain := false
testMainEntryDepth := 0

for scanner.Scan() {
lineNum++

rawLine := scanner.Text()
line := strings.ToLower(strings.TrimSpace(rawLine))

// Detect TestMain function entry.
if strings.HasPrefix(line, "func testmain(") {
insideTestMain = true
testMainEntryDepth = braceDepth
}

// Count braces to track scope depth.
for _, ch := range rawLine {
switch ch {
case '{':
braceDepth++
case '}':
braceDepth--
}
}

// Detect when we exit TestMain scope.
if insideTestMain && braceDepth <= testMainEntryDepth {
insideTestMain = false
}

// Skip comment lines.
if strings.HasPrefix(line, "//") {
continue
}

// Lines inside TestMain are allowed - it's the approved setup location.
if insideTestMain {
continue
}

for _, pattern := range bannedPatterns {
if strings.Contains(line, pattern) {
violations = append(violations, fmt.Sprintf(
"line %d: real database container %q - use testdb.NewInMemorySQLiteDB() instead", lineNum, rawLine))

break
}
}
}

return violations
}
