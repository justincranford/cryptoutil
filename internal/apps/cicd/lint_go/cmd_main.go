package lint_go

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

const mainGoFilename = "main.go"

// checkCmdMainPattern checks that all main.go files under cmd/ follow the ARCHITECTURE.md 4.4.3 pattern.
// Required pattern: func main() { os.Exit(cryptoutilApps<SOMETHING>.<SOMETHING>(os.Args, os.Stdin, os.Stdout, os.Stderr)) }.
func checkCmdMainPattern(logger *cryptoutilCmdCicdCommon.Logger) error {
	errors := []string{}

	rootDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cmdDir := filepath.Join(rootDir, "cmd")
	if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
		return nil // No cmd directory, skip check
	}

	err = filepath.Walk(cmdDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Base(path) == mainGoFilename {
			if err := checkMainGoFile(path); err != nil {
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

// checkMainGoFile verifies a single main.go file follows the required pattern.
func checkMainGoFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Required pattern: func main() { os.Exit(cryptoutilApps<Something>.<Something>(os.Args, os.Stdin, os.Stdout, os.Stderr)) }
	// Allow whitespace variations but enforce one-liner with full os.Args
	pattern := `func\s+main\(\)\s*\{\s*os\.Exit\(cryptoutil[A-Z][a-zA-Z0-9]*\.[A-Z][a-zA-Z0-9]*\(os\.Args,\s*os\.Stdin,\s*os\.Stdout,\s*os\.Stderr\)\)\s*\}`

	matched, err := regexp.Match(pattern, content)
	if err != nil {
		return fmt.Errorf("regex error: %w", err)
	}

	if !matched {
		return fmt.Errorf("does not match required pattern: func main() { os.Exit(cryptoutilApps<Something>.<Something>(os.Args, os.Stdin, os.Stdout, os.Stderr)) }")
	}

	return nil
}
