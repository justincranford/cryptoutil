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

// makeSkillCommandPair writes a matched Copilot skill + Claude command pair under rootDir.
// The Claude command content includes the expected back-reference to the skill file.
func makeSkillCommandPair(t *testing.T, rootDir, skillName string) {
	t.Helper()

	skillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", skillName)
	commandDir := filepath.Join(rootDir, ".claude", "commands")

	require.NoError(t, os.MkdirAll(skillDir, 0o700))
	require.NoError(t, os.MkdirAll(commandDir, 0o700))

	skillContent := fmt.Sprintf("# %s\n\nThis is the skill content for %s.\n", skillName, skillName)
	skillRef := fmt.Sprintf(".github/skills/%s/SKILL.md", skillName)
	commandContent := fmt.Sprintf("# %s command\n\nFull Copilot original: [%s](%s)\n\nBrief description.\n", skillName, skillName, skillRef)

	require.NoError(t, os.WriteFile(filepath.Join(skillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte(skillContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(commandDir, skillName+".md"), []byte(commandContent), cryptoutilSharedMagic.FilePermissionsDefault))
}

func TestCheckSkillCommandDrift_AllPairs(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	makeSkillCommandPair(t, rootDir, "test-table-driven")
	makeSkillCommandPair(t, rootDir, "coverage-analysis")
	makeSkillCommandPair(t, rootDir, "fips-audit")

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Empty(t, result.Violations)
	require.Equal(t, 3, result.Checked)
}

func TestCheckSkillCommandDrift_MissingCommandFile(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Skill exists but no Claude command.
	skillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "orphan-skill")
	commandDir := filepath.Join(rootDir, ".claude", "commands")

	require.NoError(t, os.MkdirAll(skillDir, 0o700))
	require.NoError(t, os.MkdirAll(commandDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte("# Orphan skill\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "missing", result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, "orphan-skill")
	require.Contains(t, result.Violations[0].Detail, ".claude/commands/orphan-skill.md")
}

func TestCheckSkillCommandDrift_MissingSkillRef(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	skillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "test-fuzz")
	commandDir := filepath.Join(rootDir, ".claude", "commands")

	require.NoError(t, os.MkdirAll(skillDir, 0o700))
	require.NoError(t, os.MkdirAll(commandDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte("# Fuzz skill\n"), cryptoutilSharedMagic.FilePermissionsDefault))
	// Command exists but does NOT reference the skill file.
	require.NoError(t, os.WriteFile(filepath.Join(commandDir, "test-fuzz.md"), []byte("# Fuzz command\n\nNo back-reference here.\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "missing-reference", result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, ".github/skills/test-fuzz/SKILL.md")
}

func TestCheckSkillCommandDrift_OrphanCommand(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Claude command exists but no matching skill dir.
	commandDir := filepath.Join(rootDir, ".claude", "commands")
	skillsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills")

	require.NoError(t, os.MkdirAll(commandDir, 0o700))
	require.NoError(t, os.MkdirAll(skillsDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(commandDir, "orphan-command.md"), []byte("# Orphan command\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "orphan", result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, ".claude/commands/orphan-command.md")
	require.Contains(t, result.Violations[0].Detail, ".github/skills/orphan-command/SKILL.md")
}

func TestCheckSkillCommandDrift_MissingSKILLmd(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Skill directory exists but SKILL.md is missing.
	skillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "no-skill-file")
	commandDir := filepath.Join(rootDir, ".claude", "commands")

	require.NoError(t, os.MkdirAll(skillDir, 0o700))
	require.NoError(t, os.MkdirAll(commandDir, 0o700))
	// Only write the command; skill dir is empty.
	require.NoError(t, os.WriteFile(filepath.Join(commandDir, "no-skill-file.md"), []byte("# Command\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	// Expect missing-skill-file violation AND orphan violation for the command.
	fieldViolated := make(map[string]bool)
	for _, v := range result.Violations {
		fieldViolated[v.Field] = true
	}

	require.True(t, fieldViolated["missing-skill-file"], "expected missing-skill-file violation")
	require.True(t, fieldViolated["orphan"], "expected orphan violation for command with no skill")
}

func TestCheckSkillCommandDrift_EmptyDirs(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills"), 0o700))
	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, ".claude", "commands"), 0o700))

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

	// Only create skills dir, no commands dir.
	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills"), 0o700))

	_, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot read .claude/commands")
}

func TestCheckSkillCommandDrift_SkillsIgnoreFiles(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// README.md should be ignored (not a skill directory).
	skillsBaseDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills")
	commandDir := filepath.Join(rootDir, ".claude", "commands")

	require.NoError(t, os.MkdirAll(skillsBaseDir, 0o700))
	require.NoError(t, os.MkdirAll(commandDir, 0o700))

	// File in .github/skills/ root (not a subdirectory) — should be ignored.
	require.NoError(t, os.WriteFile(filepath.Join(skillsBaseDir, "README.md"), []byte("# README\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Empty(t, result.Violations, "files in .github/skills/ root must be ignored (only subdirs are skills)")
	require.Equal(t, 0, result.Checked)
}

func TestCheckSkillCommandDrift_CommandsIgnoreNonMd(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	skillDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "my-skill")
	commandDir := filepath.Join(rootDir, ".claude", "commands")

	require.NoError(t, os.MkdirAll(skillDir, 0o700))
	require.NoError(t, os.MkdirAll(commandDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte("# Skill\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	skillRef := ".github/skills/my-skill/SKILL.md"
	require.NoError(t, os.WriteFile(filepath.Join(commandDir, "my-skill.md"), []byte(fmt.Sprintf("# Cmd\n\nRef: %s\n", skillRef)), cryptoutilSharedMagic.FilePermissionsDefault))
	// Non-.md file in commands dir — should be ignored.
	require.NoError(t, os.WriteFile(filepath.Join(commandDir, "notes.txt"), []byte("ignored\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Empty(t, result.Violations)
}

func TestFormatSkillCommandDriftResults_Clean(t *testing.T) {
	t.Parallel()

	result := &SkillCommandDriftResult{Checked: 14}
	report := formatSkillCommandDriftResults(result)

	require.Contains(t, report, "Checked 14 Copilot skill / Claude Code command pairs")
	require.Contains(t, report, "All skill/command pairs are in sync")
	require.NotContains(t, report, "violation")
}

func TestFormatSkillCommandDriftResults_WithViolations(t *testing.T) {
	t.Parallel()

	result := &SkillCommandDriftResult{
		Checked: cryptoutilSharedMagic.DefaultEmailOTPLength,
		Violations: []SkillCommandDriftViolation{
			{
				SkillFile:   ".github/skills/foo/SKILL.md",
				CommandFile: ".claude/commands/foo.md",
				Field:       "missing-reference",
				Detail:      "Claude Code command does not reference the skill",
			},
			{
				SkillFile:   ".github/skills/bar/SKILL.md",
				CommandFile: ".claude/commands/bar.md",
				Field:       "missing",
				Detail:      "Claude Code command file not found",
			},
		},
	}

	report := formatSkillCommandDriftResults(result)

	require.Contains(t, report, fmt.Sprintf("Checked %d Copilot skill / Claude Code command pairs", cryptoutilSharedMagic.DefaultEmailOTPLength))
	require.Contains(t, report, "2 violation(s) found")
	require.Contains(t, report, "field=missing-reference")
	require.Contains(t, report, "field=missing")
	require.Contains(t, report, ".github/skills/foo/SKILL.md")
	require.Contains(t, report, ".claude/commands/bar.md")

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

// Sequential: modifies findProjectRootFn package-level seam.
func TestSkillCommandDriftCommand_AllClean(t *testing.T) {
	tmpDir := t.TempDir()

	makeSkillCommandPair(t, tmpDir, "my-skill")

	orig := findProjectRootFn

	t.Cleanup(func() { findProjectRootFn = orig })

	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	var stdout, stderr bytes.Buffer

	exitCode := SkillCommandDriftCommand(&stdout, &stderr)

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "All skill/command pairs are in sync")
	require.Empty(t, stderr.String())
}

// Sequential: modifies findProjectRootFn package-level seam.
func TestSkillCommandDriftCommand_WithViolation(t *testing.T) {
	tmpDir := t.TempDir()

	// Missing Claude command.
	skillDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "broken")
	commandDir := filepath.Join(tmpDir, ".claude", "commands")

	require.NoError(t, os.MkdirAll(skillDir, 0o700))
	require.NoError(t, os.MkdirAll(commandDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, cryptoutilSharedMagic.CICDSkillFileName), []byte("# Broken skill\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	orig := findProjectRootFn

	t.Cleanup(func() { findProjectRootFn = orig })

	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	var stdout, stderr bytes.Buffer

	exitCode := SkillCommandDriftCommand(&stdout, &stderr)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stdout.String(), "violation(s) found")
	require.Contains(t, stderr.String(), "skill-command-drift:")
}
