// Copyright (c) 2025 Justin Cranford

package github_cleanup

import (
	"errors"
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// CleanerFunc is the signature for github_cleanup sub-cleaners.
type CleanerFunc func(*CleanupConfig) error

// registeredCleaners holds the ordered list of cleanup operations.
var registeredCleaners = []struct {
	name    string
	cleaner CleanerFunc
}{
	{"cleanup-runs", CleanupRuns},
	{"cleanup-artifacts", CleanupArtifacts},
	{"cleanup-caches", CleanupCaches},
}

// Cleanup parses args and runs all registered cleanup operations sequentially.
// Continues on failure, collecting all errors before returning.
// Flags: --confirm, --max-age-days=N, --keep-min-runs=N, --repo=owner/repo.
func Cleanup(logger *cryptoutilCmdCicdCommon.Logger, args []string) error {
	cfg := NewDefaultConfig(logger)

	if err := ParseArgs(args, cfg); err != nil {
		return fmt.Errorf("github-cleanup: invalid arguments: %w", err)
	}

	var errs []error

	for _, c := range registeredCleaners {
		logger.Log(fmt.Sprintf("Running %s", c.name))

		if err := c.cleaner(cfg); err != nil {
			logger.LogError(err)
			errs = append(errs, fmt.Errorf("%s: %w", c.name, err))
		} else {
			logger.Log(fmt.Sprintf("  \u2705 %s completed", c.name))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("github-cleanup failed: %w", errors.Join(errs...))
	}

	return nil
}
