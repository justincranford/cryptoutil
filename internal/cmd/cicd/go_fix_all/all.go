// Copyright (c) 2025 Justin Cranford

package go_fix_all

import (
	"fmt"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
	"cryptoutil/internal/cmd/cicd/go_fix_copyloopvar"
	"cryptoutil/internal/cmd/cicd/go_fix_thelper"
)

// Fix runs all auto-fix commands in sequence.
// Returns aggregated statistics across all fix commands.
func Fix(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, goVersion string) (int, int, int, error) {
	logger.Log("Starting fix-all: running all auto-fix commands")

	var (
		totalProcessed, totalModified, totalIssuesFixed int
		errors                                          []error
	)

	// Run copyloopvar fixes.
	logger.Log("Running copyloopvar fixes")

	processed, modified, issuesFixed, err := go_fix_copyloopvar.Fix(logger, rootDir, goVersion)
	if err != nil {
		errors = append(errors, fmt.Errorf("copyloopvar failed: %w", err))
	} else {
		totalProcessed += processed
		totalModified += modified
		totalIssuesFixed += issuesFixed
		logger.Log(fmt.Sprintf("copyloopvar: processed=%d, modified=%d, fixed=%d", processed, modified, issuesFixed))
	}

	// Run thelper fixes.
	logger.Log("Running thelper fixes")

	processed, modified, issuesFixed, err = go_fix_thelper.Fix(logger, rootDir)
	if err != nil {
		errors = append(errors, fmt.Errorf("thelper failed: %w", err))
	} else {
		totalProcessed += processed
		totalModified += modified
		totalIssuesFixed += issuesFixed
		logger.Log(fmt.Sprintf("thelper: processed=%d, modified=%d, fixed=%d", processed, modified, issuesFixed))
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("fix-all completed with %d errors", len(errors)))

		var errMsgs []string
		for _, err := range errors {
			errMsgs = append(errMsgs, err.Error())
		}

		return totalProcessed, totalModified, totalIssuesFixed, fmt.Errorf("fix-all failures:\n%s", strings.Join(errMsgs, "\n"))
	}

	logger.Log(fmt.Sprintf("fix-all completed: processed=%d, modified=%d, fixed=%d", totalProcessed, totalModified, totalIssuesFixed))

	return totalProcessed, totalModified, totalIssuesFixed, nil
}
