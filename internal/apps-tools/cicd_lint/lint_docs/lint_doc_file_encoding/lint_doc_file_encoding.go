// Copyright (c) 2025-2026 Justin Cranford.
// Package lint_doc_file_encoding enforces UTF-8 (without BOM) and LF-only
// line endings for documentation policy artifacts.
package lint_doc_file_encoding

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	claudeLocal  = ".claude/settings.local.json"
	skillsReadme = ".github/skills/README.md"
)

var nullByte byte

var policyDirectories = []string{
	cryptoutilSharedMagic.CICDGithubInstructionsDir,
	cryptoutilSharedMagic.CICDGithubAgentsDir,
	cryptoutilSharedMagic.CICDGithubSkillsDir,
	cryptoutilSharedMagic.CICDClaudeAgentsDir,
	cryptoutilSharedMagic.CICDClaudeSkillsDir,
}

// Check enforces UTF-8 without BOM and LF-only line endings for all agent
// tooling policy artifacts (handbook, instructions, agents, skills, and root
// control files).
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkWithRootFinder(logger, findProjectRoot)
}

func checkWithRootFinder(
	logger *cryptoutilCmdCicdCommon.Logger,
	findRoot func() (string, error),
) error {
	rootDir, err := findRoot()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}

	return CheckInDir(logger, rootDir)
}

// CheckInDir enforces UTF-8 without BOM and LF-only line endings for all agent
// tooling policy artifacts in rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithOperations(logger, rootDir, productionFileOperations())
}

func checkInDirWithOperations(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, operations fileOperations) error {
	logger.Log("Checking agent-tooling policy file encoding and line endings...")

	files, err := collectPolicyFilesWithOperations(rootDir, operations)
	if err != nil {
		return fmt.Errorf("collect policy files: %w", err)
	}

	var violations []string

	for _, filePath := range files {
		issues, err := checkFileWithReadFile(filePath, operations.readFile)
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: %v", filePath, err))

			continue
		}

		for _, issue := range issues {
			violations = append(violations, fmt.Sprintf("%s: %s", filePath, issue))
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)

		return fmt.Errorf("agent tooling encoding policy violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("lint-doc-file-encoding: all agent-tooling artifacts are UTF-8 without BOM and LF-only")

	return nil
}

func collectPolicyFiles(rootDir string) ([]string, error) {
	return collectPolicyFilesWithOperations(rootDir, productionFileOperations())
}

func collectPolicyFilesWithOperations(rootDir string, operations fileOperations) ([]string, error) {
	files := []string{
		filepath.Join(rootDir, cryptoutilSharedMagic.CICDDocsHandbookPath),
		filepath.Join(rootDir, cryptoutilSharedMagic.CICDCopilotInstructionsFile),
		filepath.Join(rootDir, cryptoutilSharedMagic.CICDClaudeMdFile),
		filepath.Join(rootDir, claudeLocal),
		filepath.Join(rootDir, skillsReadme),
	}

	for _, relativeDir := range policyDirectories {
		dirPath := filepath.Join(rootDir, relativeDir)

		if _, err := operations.stat(dirPath); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return nil, fmt.Errorf("stat %q: %w", relativeDir, err)
		}

		walkErr := operations.walkDir(dirPath, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if entry.IsDir() {
				return nil
			}

			files = append(files, path)

			return nil
		})
		if walkErr != nil {
			return nil, fmt.Errorf("walk %q: %w", relativeDir, walkErr)
		}
	}

	seen := make(map[string]struct{}, len(files))

	unique := make([]string, 0, len(files))
	for _, f := range files {
		clean := filepath.Clean(f)
		if _, ok := seen[clean]; ok {
			continue
		}

		seen[clean] = struct{}{}
		unique = append(unique, clean)
	}

	sort.Strings(unique)

	return unique, nil
}

func checkFile(filePath string) ([]string, error) {
	return checkFileWithReadFile(filePath, os.ReadFile)
}

func checkFileWithReadFile(filePath string, readFile func(string) ([]byte, error)) ([]string, error) {
	data, err := readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var issues []string

	if len(data) >= 2 {
		if (data[0] == 0xFF && data[1] == 0xFE) || (data[0] == 0xFE && data[1] == 0xFF) {
			issues = append(issues, "contains UTF-16 BOM (UTF-16 permanently banned)")
		}
	}

	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		issues = append(issues, "contains UTF-8 BOM")
	}

	if !utf8.Valid(data) {
		issues = append(issues, "contains invalid UTF-8 byte sequences")
	}

	for i, b := range data {
		if b == nullByte {
			issues = append(issues, fmt.Sprintf("contains NUL byte at offset %d (likely UTF-16)", i))

			break
		}
	}

	for i := 0; i < len(data)-1; i++ {
		if data[i] == '\r' && data[i+1] == '\n' {
			issues = append(issues, "contains CRLF line endings (CRLF permanently banned)")

			break
		}
	}

	return issues, nil
}

func findProjectRoot() (string, error) {
	return findProjectRootWithGetwd(os.Getwd)
}

func findProjectRootWithGetwd(getwd func() (string, error)) (string, error) {
	return findProjectRootWithOperations(getwd, os.Stat)
}

func findProjectRootWithOperations(
	getwd func() (string, error),
	stat func(string) (os.FileInfo, error),
) (string, error) {
	dir, err := getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, statErr := stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}

		dir = parent
	}
}

type fileOperations struct {
	stat     func(string) (os.FileInfo, error)
	walkDir  func(string, fs.WalkDirFunc) error
	readFile func(string) ([]byte, error)
}

func productionFileOperations() fileOperations {
	return fileOperations{
		stat:     os.Stat,
		walkDir:  filepath.WalkDir,
		readFile: os.ReadFile,
	}
}
