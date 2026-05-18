// Copyright (c) 2025-2026 Justin Cranford.
// Package lint_deprecated_crlf_phrasing blocks stale historical CRLF wording in
// active instructions and agents. These phrases are semantically obsolete after
// LF-everywhere policy adoption and can mislead maintainers.
package lint_deprecated_crlf_phrasing

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var bannedPhrases = []string{
	"renormalize crlf files to lf",
	"mixing crlf fixes",
}

var targetDirs = []string{
	cryptoutilSharedMagic.CICDGithubInstructionsDir,
	cryptoutilSharedMagic.CICDGithubAgentsDir,
	cryptoutilSharedMagic.CICDClaudeAgentsDir,
}

// Check scans docs/agent instruction content for obsolete CRLF historical wording.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkWithFS(logger, os.Getwd, filepath.WalkDir, os.Open)
}

func checkWithFS(
	logger *cryptoutilCmdCicdCommon.Logger,
	getwdFn func() (string, error),
	walkDirFn func(root string, fn fs.WalkDirFunc) error,
	openFn func(name string) (*os.File, error),
) error {
	root, err := findProjectRoot(getwdFn)
	if err != nil {
		return fmt.Errorf("lint-deprecated-crlf-phrasing: %w", err)
	}

	var violations []string

	for _, relDir := range targetDirs {
		absDir := filepath.Join(root, filepath.FromSlash(relDir))

		walkErr := walkDirFn(absDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if !strings.HasSuffix(strings.ToLower(path), ".md") {
				return nil
			}

			file, openErr := openFn(path)
			if openErr != nil {
				return openErr
			}

			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					logger.Log(fmt.Sprintf("failed to close %s: %v", path, closeErr))
				}
			}()

			scanner := bufio.NewScanner(file)

			lineNumber := 0
			for scanner.Scan() {
				lineNumber++

				line := strings.ToLower(scanner.Text())
				for _, phrase := range bannedPhrases {
					if strings.Contains(line, phrase) {
						relPath, relErr := filepath.Rel(root, path)
						if relErr != nil {
							relPath = path
						}

						violations = append(violations, fmt.Sprintf("%s:%d contains banned phrase %q", filepath.ToSlash(relPath), lineNumber, phrase))
					}
				}
			}

			if scanErr := scanner.Err(); scanErr != nil {
				return fmt.Errorf("failed to scan %s: %w", path, scanErr)
			}

			return nil
		})
		if walkErr != nil {
			return fmt.Errorf("lint-deprecated-crlf-phrasing: failed to scan %s: %w", relDir, walkErr)
		}
	}

	if len(violations) > 0 {
		for _, violation := range violations {
			logger.Log("  " + violation)
		}

		return fmt.Errorf("lint-deprecated-crlf-phrasing: %d violation(s) found", len(violations))
	}

	logger.Log("lint-deprecated-crlf-phrasing: no deprecated CRLF wording found")

	return nil
}

func findProjectRoot(getwdFn func() (string, error)) (string, error) {
	dir, err := getwdFn()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}

		dir = parent
	}
}
