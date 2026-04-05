// Copyright (c) 2025 Justin Cranford

package docs_validation

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// makeSkillPair writes a matched Copilot skill + Claude skill pair under rootDir.
// Both files contain compliant YAML frontmatter, identical body content, and ## Key Rules.
func makeSkillPair(t *testing.T, rootDir, skillName string) {
	t.Helper()

	copilotSkillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", skillName)
	claudeSkillDir := filepath.Join(rootDir, ".claude", "skills", skillName)

	require.NoError(t, os.MkdirAll(copilotSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillDir, 0o700))

	const (
		testDesc = "Test skill description."
		testHint = "[arg]"
	)

	body := fmt.Sprintf("\n## Purpose\n\nThis is the skill for %s.\n\n## Key Rules\n\n- Rule one.\n- Rule two.\n", skillName)
	copilotContent := fmt.Sprintf("---\nname: %s\ndescription: %q\nargument-hint: %q\ndisable-model-invocation: true\n---\n%s", skillName, testDesc, testHint, body)
	claudeContent := fmt.Sprintf("---\nname: %s\ndescription: %q\nargument-hint: %q\n---\n%s", skillName, testDesc, testHint, body)

	require.NoError(t, os.WriteFile(filepath.Join(copilotSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(copilotContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(claudeSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(claudeContent), cryptoutilSharedMagic.FilePermissionsDefault))
}

func TestCheckSkillCommandDrift_AllPairs(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	makeSkillPair(t, rootDir, "test-table-driven")
	makeSkillPair(t, rootDir, "coverage-analysis")
	makeSkillPair(t, rootDir, "fips-audit")

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Empty(t, result.Violations)
	require.Equal(t, 3, result.Checked)
}

func TestCheckSkillCommandDrift_MissingClaudeSkillFile(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Copilot skill exists but no Claude skill.
	copilotSkillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "orphan-skill")
	claudeSkillsDir := filepath.Join(rootDir, ".claude", "skills")

	require.NoError(t, os.MkdirAll(copilotSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillsDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(copilotSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte("---\nname: orphan-skill\ndescription: \"Test.\"\n---\n\n## Key Rules\n\n- Rule.\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "missing", result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, "orphan-skill")
	require.Contains(t, result.Violations[0].Detail, ".claude/skills/orphan-skill/SKILL.md")
}

func TestCheckSkillCommandDrift_BodyMismatch(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	copilotSkillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "test-fuzz")
	claudeSkillDir := filepath.Join(rootDir, ".claude", "skills", "test-fuzz")

	require.NoError(t, os.MkdirAll(copilotSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillDir, 0o700))

	// Copilot skill with one body.
	copilotContent := "---\nname: test-fuzz\ndescription: \"Fuzz skill.\"\nargument-hint: \"[arg]\"\n---\n\n## Purpose\n\nCopilot body.\n\n## Key Rules\n\n- Rule one.\n"
	require.NoError(t, os.WriteFile(filepath.Join(copilotSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(copilotContent), cryptoutilSharedMagic.FilePermissionsDefault))

	// Claude skill with different body.
	claudeContent := "---\nname: test-fuzz\ndescription: \"Fuzz skill.\"\nargument-hint: \"[arg]\"\n---\n\n## Purpose\n\nDifferent Claude body.\n\n## Key Rules\n\n- Rule one.\n"
	require.NoError(t, os.WriteFile(filepath.Join(claudeSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(claudeContent), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "body-mismatch", result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, ".claude/skills/test-fuzz/SKILL.md")
}

func TestCheckSkillCommandDrift_OrphanClaudeSkill(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Claude skill exists but no matching Copilot skill dir.
	claudeSkillDir := filepath.Join(rootDir, ".claude", "skills", "orphan-skill")
	copilotSkillsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills")

	require.NoError(t, os.MkdirAll(claudeSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(copilotSkillsDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(claudeSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte("# Orphan claude skill\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "orphan", result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, ".claude/skills/orphan-skill/SKILL.md")
	require.Contains(t, result.Violations[0].Detail, ".github/skills/orphan-skill/SKILL.md")
}

func TestCheckSkillCommandDrift_MissingSKILLmd(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Copilot skill directory exists but SKILL.md is missing.
	copilotSkillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "no-skill-file")
	claudeSkillsDir := filepath.Join(rootDir, ".claude", "skills")

	require.NoError(t, os.MkdirAll(copilotSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillsDir, 0o700))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "missing-skill-file", result.Violations[0].Field)
}

func TestCheckSkillCommandDrift_EmptyDirs(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills"), 0o700))
	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, ".claude", "skills"), 0o700))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Empty(t, result.Violations)
	require.Equal(t, 0, result.Checked)
}

func TestCheckSkillCommandDrift_SkillsDirMissing(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	_, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot read .github/skills")
}

func TestCheckSkillCommandDrift_CommandsDirMissing(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Only create Copilot skills dir and a skill; no Claude skills dir.
	copilotSkillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "some-skill")
	require.NoError(t, os.MkdirAll(copilotSkillDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(copilotSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte("---\nname: some-skill\ndescription: \"Test.\"\n---\n\n## Key Rules\n\n- Rule.\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	_, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot read .claude/skills")
}

func TestCheckSkillCommandDrift_SkillsIgnoreFiles(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// README.md should be ignored (not a skill directory).
	skillsBaseDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills")
	claudeSkillsDir := filepath.Join(rootDir, ".claude", "skills")

	require.NoError(t, os.MkdirAll(skillsBaseDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillsDir, 0o700))

	// File in .github/skills/ root (not a subdirectory) — should be ignored.
	require.NoError(t, os.WriteFile(filepath.Join(skillsBaseDir, "README.md"), []byte("# README\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Empty(t, result.Violations, "files in .github/skills/ root must be ignored (only subdirs are skills)")
	require.Equal(t, 0, result.Checked)
}

func TestCheckSkillCommandDrift_ClaudeSkillsIgnoreFiles(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	makeSkillPair(t, rootDir, "my-skill")

	// Non-directory entry in .claude/skills/ root — should be ignored in reverse scan.
	claudeSkillsDir := filepath.Join(rootDir, ".claude", "skills")
	require.NoError(t, os.WriteFile(filepath.Join(claudeSkillsDir, "notes.txt"), []byte("ignored\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Empty(t, result.Violations)
}

func TestFormatSkillCommandDriftResults_Clean(t *testing.T) {
	t.Parallel()

	result := &SkillCommandDriftResult{Checked: 14}
	report := formatSkillCommandDriftResults(result)

	require.Contains(t, report, "Checked 14 Copilot skill / Claude Code skill pairs")
	require.Contains(t, report, "All skill pairs are in sync")
	require.NotContains(t, report, "violation")
}

func TestFormatSkillCommandDriftResults_WithViolations(t *testing.T) {
	t.Parallel()

	result := &SkillCommandDriftResult{
		Checked: cryptoutilSharedMagic.DefaultEmailOTPLength,
		Violations: []SkillCommandDriftViolation{
			{
				SkillFile:       ".github/skills/foo/SKILL.md",
				ClaudeSkillFile: ".claude/skills/foo/SKILL.md",
				Field:           "body-mismatch",
				Detail:          "Claude Code skill body does not match Copilot skill",
			},
			{
				SkillFile:       ".github/skills/bar/SKILL.md",
				ClaudeSkillFile: ".claude/skills/bar/SKILL.md",
				Field:           "missing",
				Detail:          "Claude Code skill file not found",
			},
		},
	}

	report := formatSkillCommandDriftResults(result)

	require.Contains(t, report, fmt.Sprintf("Checked %d Copilot skill / Claude Code skill pairs", cryptoutilSharedMagic.DefaultEmailOTPLength))
	require.Contains(t, report, "2 violation(s) found")
	require.Contains(t, report, "field=body-mismatch")
	require.Contains(t, report, "field=missing")
	require.Contains(t, report, ".github/skills/foo/SKILL.md")
	require.Contains(t, report, ".claude/skills/bar/SKILL.md")

	require.True(t, strings.Index(report, "[1]") < strings.Index(report, "[2]"))
}

func TestSkillCommandDriftCommand_NoDirs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	readFile := rootedReadFile(tmpDir)
	readFileErr := func(path string) ([]byte, error) {
		return readFile(path)
	}

	// Directly call CheckSkillCommandDrift with a temp root that has no skills dir.
	_, err := CheckSkillCommandDrift(tmpDir, readFileErr)

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot read .github/skills")
}

func TestSkillCommandDriftCommand_AllClean(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	makeSkillPair(t, tmpDir, "my-skill")

	var stdout, stderr bytes.Buffer

	exitCode := skillCommandDriftCommand(&stdout, &stderr, func() (string, error) { return tmpDir, nil })

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "All skill pairs are in sync")
	require.Empty(t, stderr.String())
}

func TestSkillCommandDriftCommand_WithViolation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Missing Claude skill.
	copilotSkillDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "broken")
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")

	require.NoError(t, os.MkdirAll(copilotSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillsDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(copilotSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte("---\nname: broken\ndescription: \"Test.\"\n---\n\n## Key Rules\n\n- Rule.\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	var stdout, stderr bytes.Buffer

	exitCode := skillCommandDriftCommand(&stdout, &stderr, func() (string, error) { return tmpDir, nil })

	require.Equal(t, 1, exitCode)
	require.Contains(t, stdout.String(), "violation(s) found")
	require.Contains(t, stderr.String(), "skill-command-drift:")
}

func TestCheckSkillCommandDrift_MissingFrontmatter(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	copilotSkillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "no-fm")
	claudeSkillDir := filepath.Join(rootDir, ".claude", "skills", "no-fm")

	require.NoError(t, os.MkdirAll(copilotSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillDir, 0o700))

	copilotContent := "---\nname: no-fm\ndescription: \"A skill.\"\nargument-hint: \"[arg]\"\n---\n\n## Key Rules\n\n- Rule one.\n"
	// Claude skill has no frontmatter block.
	claudeContent := "# No frontmatter skill\n\n## Key Rules\n\n- Rule one.\n"

	require.NoError(t, os.WriteFile(filepath.Join(copilotSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(copilotContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(claudeSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(claudeContent), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)

	fmViolation := false

	for _, v := range result.Violations {
		if v.Field == "missing-frontmatter" {
			fmViolation = true

			require.Contains(t, v.Detail, ".claude/skills/no-fm/SKILL.md")
		}
	}

	require.True(t, fmViolation, "expected missing-frontmatter violation")
}

func TestCheckSkillCommandDrift_DescriptionMismatch(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	copilotSkillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "desc-mismatch")
	claudeSkillDir := filepath.Join(rootDir, ".claude", "skills", "desc-mismatch")

	require.NoError(t, os.MkdirAll(copilotSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillDir, 0o700))

	copilotContent := "---\nname: desc-mismatch\ndescription: \"Original description.\"\nargument-hint: \"[arg]\"\n---\n\n## Key Rules\n\n- Rule one.\n"
	claudeContent := "---\nname: desc-mismatch\ndescription: \"Different description.\"\nargument-hint: \"[arg]\"\n---\n\n## Key Rules\n\n- Rule one.\n"

	require.NoError(t, os.WriteFile(filepath.Join(copilotSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(copilotContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(claudeSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(claudeContent), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)

	fields := make(map[string]bool)
	for _, v := range result.Violations {
		fields[v.Field] = true
	}

	require.True(t, fields["description-mismatch"], "expected description-mismatch violation")
	require.NotEmpty(t, result.Violations[0].Detail)
}

func TestCheckSkillCommandDrift_ArgumentHintMismatch(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	copilotSkillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "hint-mismatch")
	claudeSkillDir := filepath.Join(rootDir, ".claude", "skills", "hint-mismatch")

	require.NoError(t, os.MkdirAll(copilotSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillDir, 0o700))

	copilotContent := "---\nname: hint-mismatch\ndescription: \"A skill.\"\nargument-hint: \"[correct-arg]\"\n---\n\n## Key Rules\n\n- Rule one.\n"
	claudeContent := "---\nname: hint-mismatch\ndescription: \"A skill.\"\nargument-hint: \"[wrong-arg]\"\n---\n\n## Key Rules\n\n- Rule one.\n"

	require.NoError(t, os.WriteFile(filepath.Join(copilotSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(copilotContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(claudeSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(claudeContent), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)

	fields := make(map[string]bool)
	for _, v := range result.Violations {
		fields[v.Field] = true
	}

	require.True(t, fields["argument-hint-mismatch"], "expected argument-hint-mismatch violation")
}

func TestCheckSkillCommandDrift_SkillMissingKeyRules(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	copilotSkillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "no-kr-skill")
	claudeSkillDir := filepath.Join(rootDir, ".claude", "skills", "no-kr-skill")

	require.NoError(t, os.MkdirAll(copilotSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillDir, 0o700))

	// Copilot skill is missing ## Key Rules.
	noKeyRulesContent := "---\nname: no-kr-skill\ndescription: \"A skill.\"\nargument-hint: \"[arg]\"\n---\n\n## Purpose\n\nNo Key Rules here.\n"

	require.NoError(t, os.WriteFile(filepath.Join(copilotSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(noKeyRulesContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(claudeSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(noKeyRulesContent), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)

	fields := make(map[string]bool)
	for _, v := range result.Violations {
		fields[v.Field] = true
	}

	require.True(t, fields["missing-key-rules"], "expected missing-key-rules violation for skill")
}

func TestCheckSkillCommandDrift_ClaudeSkillMissingKeyRules(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	copilotSkillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "no-kr-claude")
	claudeSkillDir := filepath.Join(rootDir, ".claude", "skills", "no-kr-claude")

	require.NoError(t, os.MkdirAll(copilotSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillDir, 0o700))

	copilotContent := "---\nname: no-kr-claude\ndescription: \"A skill.\"\nargument-hint: \"[arg]\"\n---\n\n## Key Rules\n\n- Rule one.\n"
	// Claude skill is missing ## Key Rules.
	claudeContent := "---\nname: no-kr-claude\ndescription: \"A skill.\"\nargument-hint: \"[arg]\"\n---\n\n## Purpose\n\nNo Key Rules here.\n"

	require.NoError(t, os.WriteFile(filepath.Join(copilotSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(copilotContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(claudeSkillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(claudeContent), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)

	fields := make(map[string]bool)
	for _, v := range result.Violations {
		fields[v.Field] = true
	}

	require.True(t, fields["missing-key-rules"], "expected missing-key-rules violation for Claude skill")
}
