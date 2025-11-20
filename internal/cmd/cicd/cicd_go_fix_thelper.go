package cicd

import (
	"fmt"
	"path/filepath"

	cryptoutilCmd "cryptoutil/internal/cmd/cicd/common"
	"cryptoutil/internal/cmd/cicd/fix/thelper"
)

// goFixTHelper adds t.Helper() to test helper functions that are missing it.
// Test helper functions are identified by naming patterns (setup*, check*, assert*, verify*, helper*).
func goFixTHelper(logger *cryptoutilCmd.Logger, files []string) error {
	logger.Log("Starting thelper auto-fix")

	// Get root directory from files.
	if len(files) == 0 {
		logger.Log("No files provided")

		return nil
	}

	// Use first file's directory as root.
	rootDir := filepath.Dir(files[0])

	// Call thelper package.
	_, _, issuesFixed, err := thelper.Fix(logger, rootDir)
	if err != nil {
		return fmt.Errorf("thelper fix failed: %w", err)
	}

	if issuesFixed > 0 {
		return fmt.Errorf("added t.Helper() to %d test helper functions - please review changes", issuesFixed)
	}

	logger.Log("No test helper functions needed t.Helper()")

	return nil
}
