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

// SkillCommandDriftViolation describes a single skill drift error between Copilot and Claude skill files.
type SkillCommandDriftViolation struct {
	SkillFile       string
	ClaudeSkillFile string
	Field           string
	Detail          string
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

// extractBody returns the content after the YAML frontmatter block (everything
// after the closing `---` delimiter). If no frontmatter is present, the entire
// content is returned. Leading/trailing whitespace is trimmed from the result.
func extractBody(content string) string {
	lines := strings.SplitAfter(content, "\n")
	inFrontmatter := false
	fmEndIdx := -1

	for i, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r\n")
		if line == cryptoutilSharedMagic.CICDYAMLFrontmatterDelimiter {
			if !inFrontmatter {
				inFrontmatter = true

				continue
			}

			fmEndIdx = i + 1

			break
		}
	}

	if fmEndIdx < 0 {
		return strings.TrimSpace(content)
	}

	return strings.TrimSpace(strings.Join(lines[fmEndIdx:], ""))
}

// CheckSkillCommandDrift validates that every Copilot skill in .github/skills/NAME/
// has a matching Claude Code skill at .claude/skills/NAME/SKILL.md, and that
// frontmatter fields and body content are identical between the two files.
// It also validates that both files contain a ## Key Rules section.
func CheckSkillCommandDrift(rootDir string, readFileFn func(string) ([]byte, error)) (*SkillCommandDriftResult, error) {
	result := &SkillCommandDriftResult{}

	skillsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDGithubSkillsDir)

	// Step 1: Scan and collect all skill names from .github/skills/ subdirectories.
	skillEntries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", cryptoutilSharedMagic.CICDGithubSkillsDir, err)
	}

	skillNames := make([]string, 0, len(skillEntries))

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

	// Step 2: For each skill, validate the corresponding Claude skill file.
	for _, skillName := range skillNames {
		skillFilePath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubSkillsDir, skillName, cryptoutilSharedMagic.CICDSkillFileName))
		claudeSkillRelPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDClaudeSkillsDir, skillName, cryptoutilSharedMagic.CICDSkillFileName))

		skillContent, skillReadErr := readFileFn(skillFilePath)
		if skillReadErr != nil {
			continue // Missing skill file already flagged in Step 1.
		}

		skillStr := string(skillContent)

		claudeContent, claudeReadErr := readFileFn(claudeSkillRelPath)
		if claudeReadErr != nil {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile:       skillFilePath,
				ClaudeSkillFile: claudeSkillRelPath,
				Field:           "missing",
				Detail:          fmt.Sprintf("Claude Code skill file not found for skill %q: expected %s", skillName, claudeSkillRelPath),
			})

			result.Checked++

			continue
		}

		result.Checked++

		claudeStr := string(claudeContent)

		// Validate that the Claude skill has YAML frontmatter.
		if !hasFrontmatter(claudeStr) {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile:       skillFilePath,
				ClaudeSkillFile: claudeSkillRelPath,
				Field:           "missing-frontmatter",
				Detail:          fmt.Sprintf("Claude Code skill %s is missing YAML frontmatter (must begin with ---)", claudeSkillRelPath),
			})
		} else {
			// Validate description field matches.
			skillDesc := extractFrontmatterField(skillStr, "description")
			claudeDesc := extractFrontmatterField(claudeStr, "description")

			if skillDesc != "" && claudeDesc != skillDesc {
				result.Violations = append(result.Violations, SkillCommandDriftViolation{
					SkillFile:       skillFilePath,
					ClaudeSkillFile: claudeSkillRelPath,
					Field:           "description-mismatch",
					Detail:          fmt.Sprintf("Claude Code skill %s description does not match Copilot skill: claude=%q copilot=%q", claudeSkillRelPath, claudeDesc, skillDesc),
				})
			}

			// Validate argument-hint field matches (only when skill has one).
			skillHint := extractFrontmatterField(skillStr, "argument-hint")
			if skillHint != "" {
				claudeHint := extractFrontmatterField(claudeStr, "argument-hint")
				if claudeHint != skillHint {
					result.Violations = append(result.Violations, SkillCommandDriftViolation{
						SkillFile:       skillFilePath,
						ClaudeSkillFile: claudeSkillRelPath,
						Field:           "argument-hint-mismatch",
						Detail:          fmt.Sprintf("Claude Code skill %s argument-hint does not match Copilot skill: claude=%q copilot=%q", claudeSkillRelPath, claudeHint, skillHint),
					})
				}
			}
		}

		// Validate body content is identical between Copilot and Claude skills.
		skillBody := extractBody(skillStr)
		claudeBody := extractBody(claudeStr)

		if skillBody != claudeBody {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile:       skillFilePath,
				ClaudeSkillFile: claudeSkillRelPath,
				Field:           "body-mismatch",
				Detail:          fmt.Sprintf("Claude Code skill %s body content does not match Copilot skill %s", claudeSkillRelPath, skillFilePath),
			})
		}

		// Validate ## Key Rules section in Copilot skill.
		if !hasMarkdownSection(skillStr, "## Key Rules") {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile: skillFilePath,
				Field:     "missing-key-rules",
				Detail:    fmt.Sprintf("Copilot skill %s is missing the '## Key Rules' section", skillFilePath),
			})
		}

		// Validate ## Key Rules section in Claude skill.
		if !hasMarkdownSection(claudeStr, "## Key Rules") {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile:       skillFilePath,
				ClaudeSkillFile: claudeSkillRelPath,
				Field:           "missing-key-rules",
				Detail:          fmt.Sprintf("Claude Code skill %s is missing the '## Key Rules' section", claudeSkillRelPath),
			})
		}
	}

	// Step 3: Reverse check — every Claude skill must have a matching Copilot skill.
	claudeSkillsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDClaudeSkillsDir)

	claudeEntries, err := os.ReadDir(claudeSkillsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", cryptoutilSharedMagic.CICDClaudeSkillsDir, err)
	}

	for _, entry := range claudeEntries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		claudeSkillRelPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDClaudeSkillsDir, skillName, cryptoutilSharedMagic.CICDSkillFileName))
		expectedCopilotPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubSkillsDir, skillName, cryptoutilSharedMagic.CICDSkillFileName))

		_, skillReadErr := readFileFn(expectedCopilotPath)
		if skillReadErr != nil {
			result.Violations = append(result.Violations, SkillCommandDriftViolation{
				SkillFile:       expectedCopilotPath,
				ClaudeSkillFile: claudeSkillRelPath,
				Field:           "orphan",
				Detail:          fmt.Sprintf("Claude Code skill %s has no matching Copilot skill at %s", claudeSkillRelPath, expectedCopilotPath),
			})
		}
	}

	return result, nil
}

// formatSkillCommandDriftResults formats the skill drift report for output.
func formatSkillCommandDriftResults(result *SkillCommandDriftResult) string {
	var sb strings.Builder

	sb.WriteString("=== Skill Drift Check ===\n\n")
	sb.WriteString(fmt.Sprintf("Checked %d Copilot skill / Claude Code skill pairs\n\n", result.Checked))

	if len(result.Violations) == 0 {
		sb.WriteString("✅ All skill pairs are in sync\n")

		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("❌ %d violation(s) found:\n\n", len(result.Violations)))

	for i, v := range result.Violations {
		sb.WriteString(fmt.Sprintf("[%d] field=%s\n", i+1, v.Field))

		if v.SkillFile != "" {
			sb.WriteString(fmt.Sprintf("    copilot-skill: %s\n", v.SkillFile))
		}

		if v.ClaudeSkillFile != "" {
			sb.WriteString(fmt.Sprintf("    claude-skill:  %s\n", v.ClaudeSkillFile))
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
