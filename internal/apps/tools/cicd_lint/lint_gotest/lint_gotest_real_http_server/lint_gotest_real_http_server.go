// Copyright (c) 2025 Justin Cranford

// Package lint_gotest_real_http_server enforces that test files do not use
// httptest.NewServer to start real HTTP listeners when testing service handlers.
// Per ARCHITECTURE.md §10.2, handler tests must use Fiber's app.Test() (in-memory,
// no network binding), which avoids port conflicts, Windows firewall prompts,
// and TIME_WAIT delays. The exemption for files in client/ directories applies
// because client code tests legitimately mock remote servers via httptest.NewServer.
package lint_gotest_real_http_server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintGoTestCommon "cryptoutil/internal/apps/tools/cicd_lint/lint_gotest/common"
)

// realHTTPServerPattern is the string to detect in test files.
const realHTTPServerPattern = "httptest.NewServer("

// isExemptedFile returns true for test files that legitimately need httptest.NewServer.
// Client code tests mock remote HTTP servers (OAuth2, federation endpoints) via httptest.NewServer.
// These are NOT testing local Fiber handlers — they test HTTP client code.
func isExemptedFile(filePath string) bool {
	// Normalize to forward slashes for cross-platform path matching.
	filePath = filepath.ToSlash(filePath)

	// HTTP client tests: mocking remote OAuth2/OIDC/federation endpoints.
	if strings.Contains(filePath, "/client/") {
		return true
	}

	// Realm and federation tests: mocking remote IdP/federation endpoints.
	if strings.Contains(filePath, "/realm/") {
		return true
	}

	// Client auth tests (e.g., clientauth/): revocation, token introspection, etc.
	if strings.Contains(filePath, "clientauth") {
		return true
	}

	// Test framework helpers (e.g., testing/healthclient/): mock health endpoints.
	if strings.Contains(filePath, "/testing/") {
		return true
	}

	// Network utility tests: test HTTP client utility functions like DoRequest.
	if strings.Contains(filePath, "/util/network/") {
		return true
	}

	// Backchannel logout tests: mock backchannel logout endpoints.
	if strings.Contains(filePath, "backchannel") {
		return true
	}

	return false
}

// Check scans test files for httptest.NewServer usage that starts real HTTP listeners.
// Returns an error if any violations are found.
func Check(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Checking for real HTTP server usage in test files...")

	filteredFiles := lintGoTestCommon.FilterExcludedTestFiles(testFiles)

	if len(filteredFiles) == 0 {
		logger.Log("Real HTTP server check completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check for httptest.NewServer usage", len(filteredFiles)))

	totalIssues := 0

	for _, filePath := range filteredFiles {
		if isExemptedFile(filePath) {
			continue
		}

		issues := checkRealHTTPServer(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		logger.Log(fmt.Sprintf("Found %d real HTTP server violations", totalIssues))
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Fix: Replace httptest.NewServer() with Fiber app.Test() pattern.")
		fmt.Fprintln(os.Stderr, "  app := fiber.New(fiber.Config{DisableStartupMessage: true})")
		fmt.Fprintln(os.Stderr, "  app.Get(\"/path\", handler)")
		fmt.Fprintln(os.Stderr, "  req := httptest.NewRequest(\"GET\", \"/path\", nil)")
		fmt.Fprintln(os.Stderr, "  resp, err := app.Test(req, -1)  // in-memory, no network")

		return fmt.Errorf("found %d real HTTP server violations", totalIssues)
	}

	fmt.Fprintln(os.Stderr, "\n✅ No real HTTP server usage found in handler tests")

	logger.Log("Real HTTP server check completed")

	return nil
}

// CheckRealHTTPServerUsage checks a test file for httptest.NewServer usage.
func CheckRealHTTPServerUsage(filePath string) []string {
	return checkRealHTTPServer(filePath)
}

// checkRealHTTPServer returns violation messages for each httptest.NewServer call found.
func checkRealHTTPServer(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		if strings.Contains(line, realHTTPServerPattern) {
			issues = append(issues, fmt.Sprintf("line %d: real HTTP server started via httptest.NewServer: %s", i+1, strings.TrimSpace(line)))
		}
	}

	return issues
}
