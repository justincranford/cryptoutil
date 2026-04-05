// Package cmd_main_pattern checks that all main.go files follow the required pattern.
package cmd_main_pattern

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

const mainGoFilename = "main.go"

// mainPattern is the compiled regex for the required main.go pattern.
// Uses MustCompile since the pattern is a constant that is always valid.
// Accepts both os.Args (suite/infrastructure) and os.Args[1:] (product/service) patterns.
var mainPattern = regexp.MustCompile(`func\s+main\(\)\s*\{\s*os\.Exit\(cryptoutil[A-Z][a-zA-Z0-9]*\.[A-Z][a-zA-Z0-9]*\(os\.Args(\[1:\])?,\s*os\.Stdin,\s*os\.Stdout,\s*os\.Stderr\)\)\s*\}`)

// Check checks that all main.go files under cmd/ follow the ENG-HANDBOOK.md 4.4.3 pattern.
// Required pattern: func main() { os.Exit(cryptoutilApps<SOMETHING>.<SOMETHING>(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }.
// Also accepts os.Args (without [1:]) for suite/infrastructure binaries.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir checks all main.go files under rootDir/cmd/ follow the required pattern.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDir(logger, rootDir, filepath.Walk)
}

// checkInDir is the internal implementation that accepts a walkFn for testing.
func checkInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, walkFn func(string, filepath.WalkFunc) error) error {
	errors := []string{}

	cmdDir := filepath.Join(rootDir, "cmd")
	if _, statErr := os.Stat(cmdDir); os.IsNotExist(statErr) {
		return fmt.Errorf("cmd/ directory not found at %s", cmdDir)
	}

	err := walkFn(cmdDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Base(path) == mainGoFilename {
			if err := CheckMainGoFile(path); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", path, err))
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk cmd directory: %w", err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("cmd/ main() pattern violations:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// CheckMainGoFile verifies a single main.go file follows the required pattern.
func CheckMainGoFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Required pattern: func main() { os.Exit(cryptoutilApps<Something>.<Something>(os.Args, os.Stdin, os.Stdout, os.Stderr)) }
	// Allow whitespace variations but enforce one-liner with full os.Args
	if !mainPattern.Match(content) {
		return fmt.Errorf("does not match required pattern: func main() { os.Exit(cryptoutilApps<Something>.<Something>(os.Args or os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }")
	}

	return nil
}
