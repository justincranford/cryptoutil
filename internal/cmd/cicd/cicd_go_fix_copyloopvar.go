package cicd

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	cryptoutilCmd "cryptoutil/internal/cmd/cicd/common"
	"cryptoutil/internal/cmd/cicd/fix/copyloopvar"
)

// goFixCopyLoopVar removes unnecessary loop variable copies (x := x) in Go 1.22+.
// In Go 1.22+, loop variables are automatically per-iteration, making explicit copies redundant.
func goFixCopyLoopVar(logger *cryptoutilCmd.Logger, files []string) error {
	logger.Log("Starting copyloopvar auto-fix")

	// Get root directory from files.
	if len(files) == 0 {
		logger.Log("No files provided")
		return nil
	}

	// Use first file's directory as root.
	rootDir := filepath.Dir(files[0])

	// Get Go version from runtime.
	goVersion := strings.TrimPrefix(runtime.Version(), "go")

	// Call copyloopvar package.
	_, _, issuesFixed, err := copyloopvar.Fix(logger, rootDir, goVersion)
	if err != nil {
		return fmt.Errorf("copyloopvar fix failed: %w", err)
	}

	if issuesFixed > 0 {
		return fmt.Errorf("fixed %d loop variable copies - please review changes", issuesFixed)
	}

	logger.Log("No loop variable copies needed fixing")
	return nil
}
