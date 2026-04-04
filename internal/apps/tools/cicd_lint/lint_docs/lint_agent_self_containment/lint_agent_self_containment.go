// Copyright (c) 2025 Justin Cranford

// Package lint_agent_self_containment validates that every Copilot agent file
// in .github/agents/ contains at least one reference to ARCHITECTURE.md,
// enforcing the agent self-containment requirement from ARCHITECTURE.md §2.1.1.
// Agents without ARCHITECTURE.md references are non-compliant and must be updated.
package lint_agent_self_containment

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

const (
	agentsRelDir        = ".github/agents"
	agentFileSuffix     = ".agent.md"
	architectureRefText = "ARCHITECTURE.md"
)

// Check runs the lint-agent-self-containment check, automatically finding the project root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return checkWithFS(logger, os.Getwd, filepath.WalkDir, os.ReadFile)
}

// checkWithFS runs the check using injected filesystem operations for testability.
func checkWithFS(
	logger *cryptoutilCmdCicdCommon.Logger,
	getwdFn func() (string, error),
	walkFn func(root string, fn fs.WalkDirFunc) error,
	readFileFn func(name string) ([]byte, error),
) error {
	root, err := findProjectRoot(getwdFn)
	if err != nil {
		return fmt.Errorf("lint-agent-self-containment: %w", err)
	}

	agentsDir := filepath.Join(root, agentsRelDir)

	logger.Log(fmt.Sprintf("Checking agents in %s for ARCHITECTURE.md references...", agentsRelDir))

	var (
		violations []string
		checked    int
	)

	err = walkFn(agentsDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("failed to access %s: %w", path, walkErr)
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), agentFileSuffix) {
			return nil
		}

		content, readErr := readFileFn(path)
		if readErr != nil {
			return fmt.Errorf("failed to read %s: %w", path, readErr)
		}

		checked++

		if !strings.Contains(string(content), architectureRefText) {
			violations = append(violations, path)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("lint-agent-self-containment: failed to walk %s: %w", agentsRelDir, err)
	}

	if checked == 0 {
		return fmt.Errorf("lint-agent-self-containment: no .agent.md files found in %s", agentsRelDir)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			logger.Log(fmt.Sprintf("  agent missing ARCHITECTURE.md reference: %s", v))
		}

		return fmt.Errorf("lint-agent-self-containment: %d agent(s) missing ARCHITECTURE.md references", len(violations))
	}

	logger.Log(fmt.Sprintf("lint-agent-self-containment: all %d agents reference ARCHITECTURE.md", checked))

	return nil
}

// findProjectRoot walks up from cwd to find the directory containing go.mod.
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
