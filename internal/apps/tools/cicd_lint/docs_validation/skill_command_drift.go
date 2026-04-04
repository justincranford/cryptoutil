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

// extractFrontmatterField extracts the value of a named field from a YAML
// frontmatter block delimited by `---` markers. Returns an empty string if the
// field is absent or the file has no frontmatter.
func extractFrontmatterField(content, field string) string {
	lines := strings.SplitAfter(content, "\n")
	inFrontmatter := false

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r\n")

		if line == cryptoutilSharedMagic.CICDYAMLFrontmatterDelimiter {
			if !inFrontmatter {
				inFrontmatter = true

				continue
			}

			break
		}

		if !inFrontmatter {
			return ""
		}

		prefix := field + ":"
		if !strings.HasPrefix(line, prefix) {
			continue
		}

		val := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		// Strip surrounding double-quotes if present.
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}

		return val
	}

	return ""
}

// hasFrontmatter reports whether a file content begins with a YAML frontmatter block.
func hasFrontmatter(content string) bool {
	lines := strings.SplitN(content, "\n", cryptoutilSharedMagic.CICDFrontmatterFirstNLines)
	if len(lines) == 0 {
		return false
	}

	return strings.TrimRight(lines[0], "\r") == cryptoutilSharedMagic.CICDYAMLFrontmatterDelimiter
}

// hasMarkdownSection reports whether the content contains the given Markdown section heading.
func hasMarkdownSection(content, heading string) bool {
	return strings.Contains(content, heading)
}

// CheckSkillCommandDrift validates that every Copilot skill in .github/skills/NAME/
// has a matching Claude Code command at .claude/commands/NAME.md, and that each
// Claude command file contains a reference back to its Copilot skill file.
// It also validates that command frontmatter matches skill frontmatter and that
// both files contain a ## Key Rules section.
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

		skillContent, skillReadErr := readFileFn(skillFilePath)
		if skillReadErr != nil {
			continue // Missing skill file already flagged in Step 1.
		}

		skillStr := string(skillContent)

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

		commandStr := string(commandContent)

		// Validate that the Claude command references the Copilot skill file.
		expectedSkillRef := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubSkillsDir, skillName, cryptoutilSharedMagic.CICDSkillFileName))
		if !strings.Contains(commandStr, expectedSkillRef) {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile:   skillFilePath,
				CommandFile: commandRelPath,
				Field:       "missing-reference",
				Detail:      fmt.Sprintf("Claude Code command %s does not reference the Copilot skill file %q", commandRelPath, expectedSkillRef),
			})
		}

		// Validate that the Claude command has YAML frontmatter.
		if !hasFrontmatter(commandStr) {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile:   skillFilePath,
				CommandFile: commandRelPath,
				Field:       "missing-frontmatter",
				Detail:      fmt.Sprintf("Claude Code command %s is missing YAML frontmatter (must begin with ---)", commandRelPath),
			})
		} else {
			// Validate description field matches.
			skillDesc := extractFrontmatterField(skillStr, "description")
			cmdDesc := extractFrontmatterField(commandStr, "description")

			if skillDesc != "" && cmdDesc != skillDesc {
				result.Violations = append(result.Violations, SkillCommandDriftViolation{
					SkillFile:   skillFilePath,
					CommandFile: commandRelPath,
					Field:       "description-mismatch",
					Detail:      fmt.Sprintf("Claude Code command %s description does not match skill: command=%q skill=%q", commandRelPath, cmdDesc, skillDesc),
				})
			}

			// Validate argument-hint field matches (only when skill has one).
			skillHint := extractFrontmatterField(skillStr, "argument-hint")
			if skillHint != "" {
				cmdHint := extractFrontmatterField(commandStr, "argument-hint")
				if cmdHint != skillHint {
					result.Violations = append(result.Violations, SkillCommandDriftViolation{
						SkillFile:   skillFilePath,
						CommandFile: commandRelPath,
						Field:       "argument-hint-mismatch",
						Detail:      fmt.Sprintf("Claude Code command %s argument-hint does not match skill: command=%q skill=%q", commandRelPath, cmdHint, skillHint),
					})
				}
			}
		}

		// Validate ## Key Rules section in skill.
		if !hasMarkdownSection(skillStr, "## Key Rules") {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile: skillFilePath,
				Field:     "missing-key-rules",
				Detail:    fmt.Sprintf("Copilot skill %s is missing the '## Key Rules' section", skillFilePath),
			})
		}

		// Validate ## Key Rules section in command.
		if !hasMarkdownSection(commandStr, "## Key Rules") {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile:   skillFilePath,
				CommandFile: commandRelPath,
				Field:       "missing-key-rules",
				Detail:      fmt.Sprintf("Claude Code command %s is missing the '## Key Rules' section", commandRelPath),
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
	return skillCommandDriftCommand(stdout, stderr, findProjectRoot)
}

func skillCommandDriftCommand(stdout, stderr io.Writer, rootFn func() (string, error)) int {
	rootDir, err := rootFn()
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
