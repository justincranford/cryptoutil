// Copyright (c) 2025 Justin Cranford

package docs_validation

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// SkillCommandDriftViolation describes a single skill/command drift error.
type SkillCommandDriftViolation struct {
	SkillFile   string
	CommandFile string
	Field       string
	Detail      string
}

// SkillCommandDriftResult holds the result of the skill/command drift check.
type SkillCommandDriftResult struct {
	Violations []SkillCommandDriftViolation
	Checked    int
}

// CheckSkillCommandDrift validates that every Copilot skill in .github/skills/NAME/
// has a matching Claude Code command at .claude/commands/NAME.md, and that each
// Claude command file contains a reference back to its Copilot skill file.
func CheckSkillCommandDrift(rootDir string, readFileFn func(string) ([]byte, error)) (*SkillCommandDriftResult, error) {
	result := &SkillCommandDriftResult{}

	skillsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDGithubSkillsDir)

	// Step 1: Scan and collect all skill names from .github/skills/ subdirectories.
	skillEntries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", cryptoutilSharedMagic.CICDGithubSkillsDir, err)
	}

	var skillNames []string

	for _, entry := range skillEntries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		skillFilePath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubSkillsDir, skillName, cryptoutilSharedMagic.CICDSkillFileName))

		_, readErr := readFileFn(skillFilePath)
		if readErr != nil {
			// Directory exists but SKILL.md is missing — flag it.
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile: skillFilePath,
				Field:     "missing-skill-file",
				Detail:    fmt.Sprintf("skill directory %q exists but %s is missing", skillName, cryptoutilSharedMagic.CICDSkillFileName),
			})

			continue
		}

		skillNames = append(skillNames, skillName)
	}

	sort.Strings(skillNames)

	// Step 2: For each skill, validate the corresponding Claude command file.
	for _, skillName := range skillNames {
		skillFilePath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubSkillsDir, skillName, cryptoutilSharedMagic.CICDSkillFileName))
		commandRelPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDClaudeCommandsDir, skillName+".md"))

		commandContent, commandReadErr := readFileFn(commandRelPath)
		if commandReadErr != nil {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile:   skillFilePath,
				CommandFile: commandRelPath,
				Field:       "missing",
				Detail:      fmt.Sprintf("Claude Code command file not found for skill %q: expected %s", skillName, commandRelPath),
			})

			result.Checked++

			continue
		}

		result.Checked++

		// Validate that the Claude command references the Copilot skill file.
		expectedSkillRef := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubSkillsDir, skillName, cryptoutilSharedMagic.CICDSkillFileName))
		if !strings.Contains(string(commandContent), expectedSkillRef) {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile:   skillFilePath,
				CommandFile: commandRelPath,
				Field:       "missing-reference",
				Detail:      fmt.Sprintf("Claude Code command %s does not reference the Copilot skill file %q", commandRelPath, expectedSkillRef),
			})
		}
	}

	// Step 3: Reverse check — every Claude command must have a matching skill.
	commandsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDClaudeCommandsDir)

	commandEntries, err := os.ReadDir(commandsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", cryptoutilSharedMagic.CICDClaudeCommandsDir, err)
	}

	for _, entry := range commandEntries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		skillName := strings.TrimSuffix(name, ".md")
		commandRelPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDClaudeCommandsDir, name))
		expectedSkillFilePath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubSkillsDir, skillName, cryptoutilSharedMagic.CICDSkillFileName))

		_, skillReadErr := readFileFn(expectedSkillFilePath)
		if skillReadErr != nil {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile:   expectedSkillFilePath,
				CommandFile: commandRelPath,
				Field:       "orphan",
				Detail:      fmt.Sprintf("Claude Code command %s has no matching Copilot skill at %s", commandRelPath, expectedSkillFilePath),
			})
		}
	}

	return result, nil
}

// formatSkillCommandDriftResults formats the skill/command drift report for output.
func formatSkillCommandDriftResults(result *SkillCommandDriftResult) string {
	var sb strings.Builder

	sb.WriteString("=== Skill/Command Drift Check ===\n\n")
	sb.WriteString(fmt.Sprintf("Checked %d Copilot skill / Claude Code command pairs\n\n", result.Checked))

	if len(result.Violations) == 0 {
		sb.WriteString("✅ All skill/command pairs are in sync\n")

		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("❌ %d violation(s) found:\n\n", len(result.Violations)))

	for i, v := range result.Violations {
		sb.WriteString(fmt.Sprintf("[%d] field=%s\n", i+1, v.Field))

		if v.SkillFile != "" {
			sb.WriteString(fmt.Sprintf("    skill:   %s\n", v.SkillFile))
		}

		if v.CommandFile != "" {
			sb.WriteString(fmt.Sprintf("    command: %s\n", v.CommandFile))
		}

		sb.WriteString(fmt.Sprintf("    %s\n\n", v.Detail))
	}

	return sb.String()
}

// SkillCommandDriftCommand runs the skill/command drift check and writes results to stdout/stderr.
// Returns 0 on success, 1 on violations.
func SkillCommandDriftCommand(stdout, stderr io.Writer) int {
	rootDir, err := findProjectRootFn()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "skill-command-drift: cannot determine project root: %v\n", err)

		return 1
	}

	readFileFn := func(relPath string) ([]byte, error) {
		return os.ReadFile(filepath.Join(rootDir, filepath.FromSlash(relPath)))
	}

	result, err := CheckSkillCommandDrift(rootDir, readFileFn)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "skill-command-drift: %v\n", err)

		return 1
	}

	_, _ = fmt.Fprint(stdout, formatSkillCommandDriftResults(result))

	if len(result.Violations) > 0 {
		_, _ = fmt.Fprintf(stderr, "skill-command-drift: %d violation(s) found\n", len(result.Violations))

		return 1
	}

	return 0
}
