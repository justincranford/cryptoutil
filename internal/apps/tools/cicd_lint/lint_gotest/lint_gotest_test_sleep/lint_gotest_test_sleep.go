// Copyright (c) 2025 Justin Cranford

// Package lint_gotest_test_sleep enforces that test files do not use time.Sleep.
// Per ENG-HANDBOOK.md §10.2, time.Sleep in tests indicates a poorly designed test
// that relies on timing rather than proper synchronization primitives (channels,
// sync.WaitGroup, context cancellation, or signal injection). Use channels and
// context-based coordination to make tests deterministic and fast.
//
// Exception categories (see isExemptedFile for the full list):
//   - Rate limiter tests: must wait for actual rate limit windows to test rate limiting
//   - Session cleanup tests: must wait for session expiry timers to fire
//   - Key rotation tests: must wait for rotation intervals to trigger
//   - Connection pool tests: must wait for pool idle period timing
//   - Telemetry/sidecar tests: must wait for async push/flush intervals
//   - Admin shutdown and concurrent access tests: must wait for async shutdown sequences
package lint_gotest_test_sleep

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintGoTestCommon "cryptoutil/internal/apps/tools/cicd_lint/lint_gotest/common"
)

// testSleepPattern is the string to detect in test files.
const testSleepPattern = "time.Sleep("

// isExemptedFile returns true for test files that legitimately require time.Sleep
// because they test time-dependent system behaviors where no synchronization
// primitive can replace actual wall clock time.
func isExemptedFile(filePath string) bool {
	// Normalize to forward slashes for cross-platform path matching.
	filePath = filepath.ToSlash(filePath)

	// Rate limiter tests: must wait for actual rate limit time windows.
	if strings.Contains(filePath, "rate_limiter") {
		return true
	}

	// Session manager lifecycle tests: must wait for session expiry, cleanup, alg timing.
	if strings.Contains(filePath, "session_manager") {
		return true
	}

	// Cleanup tests: must wait for cleanup job timer intervals.
	if strings.HasSuffix(filePath, "_cleanup_test.go") ||
		strings.HasSuffix(filePath, "cleanup_test.go") ||
		strings.HasSuffix(filePath, "_cleanup_integration_test.go") ||
		strings.HasSuffix(filePath, "cleanup_integration_test.go") {
		return true
	}

	// Rotation tests: must wait for key rotation intervals.
	if strings.HasSuffix(filePath, "_rotation_test.go") {
		return true
	}

	// Pool tests: must wait for pool idle timeout expiry.
	if strings.Contains(filePath, "/pool/") {
		return true
	}

	// Telemetry tests (sidecar, service, comprehensive): must wait for async OTLP flush.
	if strings.Contains(filePath, "/telemetry/") {
		return true
	}

	// Application listener tests: must wait for async listener start/stop.
	if strings.HasSuffix(filePath, "_listener_test.go") ||
		strings.HasSuffix(filePath, "_listener_db_test.go") ||
		strings.HasSuffix(filePath, "_listener_send_test.go") {
		return true
	}

	// Application-level tests: use mock server with configurable startDelay.
	if strings.HasSuffix(filePath, "application_test.go") {
		return true
	}

	// Concurrent admin access tests: must wait for concurrent request timing.
	if strings.HasSuffix(filePath, "_concurrent_test.go") {
		return true
	}

	// Admin shutdown tests: must wait for graceful shutdown sequences.
	if strings.HasSuffix(filePath, "_shutdown_test.go") ||
		strings.HasSuffix(filePath, "_health_shutdown_test.go") {
		return true
	}

	// Coverage2 tests: businesslogic timing coverage tests.
	if strings.HasSuffix(filePath, "_coverage2_test.go") {
		return true
	}

	// Public and table-scan tests in repository layer (timing-dependent DB ops).
	if strings.HasSuffix(filePath, "public_test.go") ||
		strings.HasSuffix(filePath, "_table_test.go") {
		return true
	}

	// Integration tests (flagged with build tags; test timing-dependent cross-service flows).
	if strings.HasSuffix(filePath, "integration_test.go") ||
		strings.HasSuffix(filePath, "_integration_test.go") {
		return true
	}

	// Repository tests: tenant and other repository timing tests.
	if strings.HasSuffix(filePath, "_repository_test.go") {
		return true
	}

	// Test server helper: testserver package uses sleep in timing probe.
	if strings.Contains(filePath, "/testserver/") {
		return true
	}

	// Device authorization tests: must wait for device code expiry.
	if strings.HasSuffix(filePath, "_authorization_test.go") {
		return true
	}

	// WebAuthn authenticator tests: must wait for authenticator timeouts.
	if strings.HasSuffix(filePath, "_authenticator_test.go") {
		return true
	}

	// Policy loader cache tests: must wait for cache TTL expiry.
	if strings.HasSuffix(filePath, "_cache_test.go") {
		return true
	}

	// Revocation tests: must wait for token revocation propagation.
	if strings.HasSuffix(filePath, "_revocation_test.go") {
		return true
	}

	// High-coverage tests (highcov): comprehensive timing tests.
	if strings.HasSuffix(filePath, "_highcov_test.go") {
		return true
	}

	// Handler cert ops tests: PKI certificate operation timing.
	if strings.HasSuffix(filePath, "_cert_ops_test.go") {
		return true
	}

	// Observability tests: must wait for telemetry interval.
	if strings.HasSuffix(filePath, "observability_test.go") {
		return true
	}

	// RA cancel tests: must wait for CA operation timeouts.
	if strings.HasSuffix(filePath, "_cancel_test.go") {
		return true
	}

	// SM-IM HTTP error tests and E2E tests.
	if strings.HasSuffix(filePath, "http_errors_test.go") {
		return true
	}

	// TCP/network test utilities: must wait for connection timeouts.
	if strings.Contains(filePath, "/util/network/") ||
		strings.Contains(filePath, "/util/thread/") {
		return true
	}

	// logger_test.go in cicd_lint: tests async log flush behavior.
	if strings.HasSuffix(filePath, "logger_test.go") {
		return true
	}

	// Outdated dependency cache test: must wait for cache TTL.
	if strings.Contains(filePath, "outdated_deps") {
		return true
	}

	// Hardware error validation and webauthn tests.
	if strings.HasSuffix(filePath, "_error_validation_test.go") {
		return true
	}

	// Certificate tests (timing-based validation windows).
	if strings.Contains(filePath, "/certificate/") {
		return true
	}

	// Client auth revocation tests.
	if strings.Contains(filePath, "/clientauth/") {
		return true
	}

	// Identity authz handler tests.
	if strings.Contains(filePath, "identity-authz/handlers") {
		return true
	}

	return false
}

// Check scans test files for time.Sleep usage.
// Returns an error if any non-exempted violations are found.
func Check(logger *cryptoutilCmdCicdCommon.Logger, testFiles []string) error {
	logger.Log("Checking for time.Sleep usage in test files...")

	filteredFiles := lintGoTestCommon.FilterExcludedTestFiles(testFiles)

	if len(filteredFiles) == 0 {
		logger.Log("Test sleep check completed (no test files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d test files to check for time.Sleep usage", len(filteredFiles)))

	totalIssues := 0

	for _, filePath := range filteredFiles {
		if isExemptedFile(filePath) {
			continue
		}

		issues := checkTestSleep(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		logger.Log(fmt.Sprintf("Found %d time.Sleep violations", totalIssues))
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Fix: Replace time.Sleep with proper synchronization primitives.")
		fmt.Fprintln(os.Stderr, "  - Use channels: <-doneCh or select { case <-doneCh: case <-ctx.Done(): }")
		fmt.Fprintln(os.Stderr, "  - Use sync.WaitGroup: wg.Wait()")
		fmt.Fprintln(os.Stderr, "  - Inject timing functions and advance a fake clock in tests")

		return fmt.Errorf("found %d time.Sleep violations", totalIssues)
	}

	fmt.Fprintln(os.Stderr, "\n✅ No non-exempted time.Sleep usage found in test files")

	logger.Log("Test sleep check completed")

	return nil
}

// CheckTestSleepUsage checks a test file for time.Sleep usage.
func CheckTestSleepUsage(filePath string) []string {
	return checkTestSleep(filePath)
}

// checkTestSleep returns violation messages for each time.Sleep call found.
func checkTestSleep(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		if strings.Contains(line, testSleepPattern) {
			issues = append(issues, fmt.Sprintf("line %d: time.Sleep in test: %s", i+1, strings.TrimSpace(line)))
		}
	}

	return issues
}
