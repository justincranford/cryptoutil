// Copyright (c) 2025 Justin Cranford

// Package no_unit_test_real_server enforces that unit tests do not start real HTTP/HTTPS servers.
// Unit tests MUST use fiber.App.Test() for handler testing instead of app.Listen().
// See ARCHITECTURE.md Section 10.2 Forbidden Pattern #2.
package no_unit_test_real_server

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

// bannedPatterns contains lowercase patterns indicating a real HTTP/HTTPS server is being started.
// These are HTTP server startup methods forbidden in unit tests.
// Note: ".listen(" is included but safePatterns takes priority for raw TCP socket calls.
var bannedPatterns = []string{
	".listen(",
"listenandserve(",
"listenandservetls(",
}

// safePatterns contains lowercase patterns for lines that should NOT trigger bans.
// These are raw TCP/TLS socket bindings used for port-conflict testing (not HTTP servers).
var safePatterns = []string{
"lc.listen(",
"net.listen(",
"tls.listen(",
"(&net.listenconfig{}).listen(",
}

// allowedSuffixes lists test file suffixes that are permitted to start real servers.
var allowedSuffixes = []string{
"_integration_test.go",
"_e2e_test.go",
}

// allowedPathFragments lists path segments permitted to start real servers.
var allowedPathFragments = []string{
"testing/testserver/",
"lint_fitness/",
"lint_gotest/",
}

// Check walks all test files in "." and reports real-server violations.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
return CheckInDir(logger, ".")
}

// CheckInDir walks test files from rootDir and reports violations.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
logger.Log("Checking for unit tests starting real HTTP servers...")

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

// CheckFiles checks the provided list of test files for real-server violations.
func CheckFiles(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
if len(testFiles) == 0 {
logger.Log("No unit test files to check for real server starts")

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
logger.Log(fmt.Sprintf("Found %d real-server start violation(s) in unit tests", totalViolations))
fmt.Fprintln(os.Stderr, "Use fiber.App.Test() for handler testing instead of app.Listen().")
fmt.Fprintln(os.Stderr, "See ARCHITECTURE.md Section 10.2 Forbidden Pattern #2.")

return fmt.Errorf("found %d real-server start violation(s) in unit tests", totalViolations)
}

logger.Log("\u2705 No real server starts found in unit tests")

return nil
}

// CheckFile scans a single file for banned HTTP-server start patterns.
func CheckFile(filePath string) []string {
content, err := os.ReadFile(filePath)
if err != nil {
return []string{fmt.Sprintf("error reading file: %v", err)}
}

var violations []string

scanner := bufio.NewScanner(strings.NewReader(string(content)))
lineNum := 0

for scanner.Scan() {
lineNum++

line := strings.ToLower(strings.TrimSpace(scanner.Text()))

// Skip comment lines.
if strings.HasPrefix(line, "//") {
continue
}

// Skip safe socket-binding patterns (net.ListenConfig, tls.Listen, etc.).
isSafe := false

for _, safe := range safePatterns {
if strings.Contains(line, safe) {
isSafe = true

break
}
}

if isSafe {
continue
}

for _, pattern := range bannedPatterns {
if strings.Contains(line, pattern) {
violations = append(violations, fmt.Sprintf(
"line %d: real server start %q - use fiber.App.Test() instead", lineNum, scanner.Text()))

break
}
}
}

return violations
}
