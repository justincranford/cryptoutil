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

// writeSkillFile creates a single skill file in the given directory.
func writeSkillFile(t *testing.T, dir, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.CICDSkillFileName), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))
}

func TestCheckSkillCommandDrift(t *testing.T) {
	t.Parallel()

	ghSkills := func(root, name string) string {
		return filepath.Join(root, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", name)
	}

	clSkills := func(root, name string) string {
		return filepath.Join(root, ".claude", "skills", name)
	}

	const (
		validSkill         = "---\nname: %s\ndescription: \"Test.\"\n---\n\n## Key Rules\n\n- Rule.\n"
		validSkillWithHint = "---\nname: %s\ndescription: \"%s\"\nargument-hint: \"%s\"\n---\n\n## Key Rules\n\n- Rule one.\n"
	)

	tests := []struct {
		name             string
		setup            func(t *testing.T, root string)
		wantErr          string
		wantChecked      int
		checkChecked     bool
		wantFields       []string
		wantDetailSubstr []string
	}{
		{
			name: "all pairs matched",
			setup: func(t *testing.T, root string) {
				t.Helper()
				makeSkillPair(t, root, "test-table-driven")
				makeSkillPair(t, root, "coverage-analysis")
				makeSkillPair(t, root, "fips-audit")
			},
			checkChecked: true,
			wantChecked:  3,
		},
		{
			name: "missing claude skill file",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeSkillFile(t, ghSkills(root, "orphan-skill"), fmt.Sprintf(validSkill, "orphan-skill"))
				require.NoError(t, os.MkdirAll(clSkills(root, ""), 0o700))
			},
			wantFields:       []string{"missing"},
			wantDetailSubstr: []string{"orphan-skill", ".claude/skills/orphan-skill/SKILL.md"},
		},
		{
			name: "body mismatch",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeSkillFile(t, ghSkills(root, "test-fuzz"),
					"---\nname: test-fuzz\ndescription: \"Fuzz skill.\"\nargument-hint: \"[arg]\"\n---\n\n## Purpose\n\nCopilot body.\n\n## Key Rules\n\n- Rule one.\n")
				writeSkillFile(t, clSkills(root, "test-fuzz"),
					"---\nname: test-fuzz\ndescription: \"Fuzz skill.\"\nargument-hint: \"[arg]\"\n---\n\n## Purpose\n\nDifferent Claude body.\n\n## Key Rules\n\n- Rule one.\n")
			},
			wantFields:       []string{"body-mismatch"},
			wantDetailSubstr: []string{".claude/skills/test-fuzz/SKILL.md"},
		},
		{
			name: "orphan claude skill",
			setup: func(t *testing.T, root string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(ghSkills(root, ""), 0o700))
				writeSkillFile(t, clSkills(root, "orphan-skill"), "# Orphan claude skill\n")
			},
			wantFields:       []string{"orphan"},
			wantDetailSubstr: []string{".claude/skills/orphan-skill/SKILL.md", ".github/skills/orphan-skill/SKILL.md"},
		},
		{
			name: "missing SKILL.md in copilot dir",
			setup: func(t *testing.T, root string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(ghSkills(root, "no-skill-file"), 0o700))
				require.NoError(t, os.MkdirAll(clSkills(root, ""), 0o700))
			},
			wantFields: []string{"missing-skill-file"},
		},
		{
			name: "empty dirs no violations",
			setup: func(t *testing.T, root string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(ghSkills(root, ""), 0o700))
				require.NoError(t, os.MkdirAll(clSkills(root, ""), 0o700))
			},
			checkChecked: true,
			wantChecked:  0,
		},
		{
			name: "skills dir missing error",
			setup: func(t *testing.T, _ string) {
				t.Helper()
			},
			wantErr: "cannot read .github/skills",
		},
		{
			name: "claude skills dir missing error",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeSkillFile(t, ghSkills(root, "some-skill"), fmt.Sprintf(validSkill, "some-skill"))
			},
			wantErr: "cannot read .claude/skills",
		},
		{
			name: "ignores files in copilot skills root",
			setup: func(t *testing.T, root string) {
				t.Helper()

				skillsBase := ghSkills(root, "")
				require.NoError(t, os.MkdirAll(skillsBase, 0o700))
				require.NoError(t, os.MkdirAll(clSkills(root, ""), 0o700))
				require.NoError(t, os.WriteFile(filepath.Join(skillsBase, "README.md"), []byte("# README\n"), cryptoutilSharedMagic.FilePermissionsDefault))
			},
			checkChecked: true,
			wantChecked:  0,
		},
		{
			name: "ignores files in claude skills root",
			setup: func(t *testing.T, root string) {
				t.Helper()
				makeSkillPair(t, root, "my-skill")
				require.NoError(t, os.WriteFile(filepath.Join(clSkills(root, ""), "notes.txt"), []byte("ignored\n"), cryptoutilSharedMagic.FilePermissionsDefault))
			},
		},
		{
			name: "missing frontmatter in claude skill",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeSkillFile(t, ghSkills(root, "no-fm"),
					fmt.Sprintf(validSkillWithHint, "no-fm", "A skill.", "[arg]"))
				writeSkillFile(t, clSkills(root, "no-fm"),
					"# No frontmatter skill\n\n## Key Rules\n\n- Rule one.\n")
			},
			wantFields:       []string{"missing-frontmatter"},
			wantDetailSubstr: []string{".claude/skills/no-fm/SKILL.md"},
		},
		{
			name: "description mismatch",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeSkillFile(t, ghSkills(root, "desc-mismatch"), fmt.Sprintf(validSkillWithHint, "desc-mismatch", "Original description.", "[arg]"))
				writeSkillFile(t, clSkills(root, "desc-mismatch"), fmt.Sprintf(validSkillWithHint, "desc-mismatch", "Different description.", "[arg]"))
			},
			wantFields: []string{"description-mismatch"},
		},
		{
			name: "argument hint mismatch",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeSkillFile(t, ghSkills(root, "hint-mismatch"), fmt.Sprintf(validSkillWithHint, "hint-mismatch", "A skill.", "[correct-arg]"))
				writeSkillFile(t, clSkills(root, "hint-mismatch"), fmt.Sprintf(validSkillWithHint, "hint-mismatch", "A skill.", "[wrong-arg]"))
			},
			wantFields: []string{"argument-hint-mismatch"},
		},
		{
			name: "copilot skill missing key rules",
			setup: func(t *testing.T, root string) {
				t.Helper()

				noKR := "---\nname: no-kr-skill\ndescription: \"A skill.\"\nargument-hint: \"[arg]\"\n---\n\n## Purpose\n\nNo Key Rules here.\n"
				writeSkillFile(t, ghSkills(root, "no-kr-skill"), noKR)
				writeSkillFile(t, clSkills(root, "no-kr-skill"), noKR)
			},
			wantFields: []string{"missing-key-rules"},
		},
		{
			name: "claude skill missing key rules",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeSkillFile(t, ghSkills(root, "no-kr-claude"),
					fmt.Sprintf(validSkillWithHint, "no-kr-claude", "A skill.", "[arg]"))
				writeSkillFile(t, clSkills(root, "no-kr-claude"),
					"---\nname: no-kr-claude\ndescription: \"A skill.\"\nargument-hint: \"[arg]\"\n---\n\n## Purpose\n\nNo Key Rules here.\n")
			},
			wantFields: []string{"missing-key-rules"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := t.TempDir()
			tc.setup(t, rootDir)

			result, err := CheckSkillCommandDrift(rootDir, rootedReadFile(rootDir))
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)

			if tc.checkChecked {
				require.Equal(t, tc.wantChecked, result.Checked)
			}

			fields := make(map[string]bool)
			for _, v := range result.Violations {
				fields[v.Field] = true
			}

			for _, f := range tc.wantFields {
				require.True(t, fields[f], "expected violation field %q", f)
			}

			if len(tc.wantFields) == 0 && tc.wantErr == "" {
				require.Empty(t, result.Violations)
			}

			for _, substr := range tc.wantDetailSubstr {
				found := false

				for _, v := range result.Violations {
					if strings.Contains(v.Detail, substr) {
						found = true

						break
					}
				}

				require.True(t, found, "expected detail containing %q", substr)
			}
		})
	}
}

func TestFormatSkillCommandDriftResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		result          *SkillCommandDriftResult
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "clean report",
			result:          &SkillCommandDriftResult{Checked: 14},
			wantContains:    []string{"Checked 14 Copilot skill / Claude Code skill pairs", "All skill pairs are in sync"},
			wantNotContains: []string{"violation"},
		},
		{
			name: "report with violations",
			result: &SkillCommandDriftResult{
				Checked: cryptoutilSharedMagic.DefaultEmailOTPLength,
				Violations: []SkillCommandDriftViolation{
					{SkillFile: ".github/skills/foo/SKILL.md", ClaudeSkillFile: ".claude/skills/foo/SKILL.md", Field: "body-mismatch", Detail: "Claude Code skill body does not match Copilot skill"},
					{SkillFile: ".github/skills/bar/SKILL.md", ClaudeSkillFile: ".claude/skills/bar/SKILL.md", Field: "missing", Detail: "Claude Code skill file not found"},
				},
			},
			wantContains: []string{
				fmt.Sprintf("Checked %d Copilot skill / Claude Code skill pairs", cryptoutilSharedMagic.DefaultEmailOTPLength),
				"2 violation(s) found", "field=body-mismatch", "field=missing",
				".github/skills/foo/SKILL.md", ".claude/skills/bar/SKILL.md",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			report := formatSkillCommandDriftResults(tc.result)

			for _, s := range tc.wantContains {
				require.Contains(t, report, s)
			}

			for _, s := range tc.wantNotContains {
				require.NotContains(t, report, s)
			}
		})
	}
}

func TestFormatSkillCommandDriftResults_ViolationOrdering(t *testing.T) {
	t.Parallel()

	result := &SkillCommandDriftResult{
		Checked: cryptoutilSharedMagic.DefaultEmailOTPLength,
		Violations: []SkillCommandDriftViolation{
			{SkillFile: ".github/skills/foo/SKILL.md", ClaudeSkillFile: ".claude/skills/foo/SKILL.md", Field: "body-mismatch", Detail: "mismatch"},
			{SkillFile: ".github/skills/bar/SKILL.md", ClaudeSkillFile: ".claude/skills/bar/SKILL.md", Field: "missing", Detail: "not found"},
		},
	}

	report := formatSkillCommandDriftResults(result)
	require.True(t, strings.Index(report, "[1]") < strings.Index(report, "[2]"))
}

func TestSkillCommandDriftCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setup           func(t *testing.T, root string)
		wantExitCode    int
		wantStdout      []string
		wantStderrEmpty bool
	}{
		{
			name: "all clean",
			setup: func(t *testing.T, root string) {
				t.Helper()
				makeSkillPair(t, root, "my-skill")
			},
			wantExitCode:    0,
			wantStdout:      []string{"All skill pairs are in sync"},
			wantStderrEmpty: true,
		},
		{
			name: "with violation",
			setup: func(t *testing.T, root string) {
				t.Helper()

				copilotDir := filepath.Join(root, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "skills", "broken")
				writeSkillFile(t, copilotDir, "---\nname: broken\ndescription: \"Test.\"\n---\n\n## Key Rules\n\n- Rule.\n")
				require.NoError(t, os.MkdirAll(filepath.Join(root, ".claude", "skills"), 0o700))
			},
			wantExitCode: 1,
			wantStdout:   []string{"violation(s) found"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			tc.setup(t, tmpDir)

			var stdout, stderr bytes.Buffer

			exitCode := skillCommandDriftCommand(&stdout, &stderr, func() (string, error) { return tmpDir, nil })
			require.Equal(t, tc.wantExitCode, exitCode)

			for _, s := range tc.wantStdout {
				require.Contains(t, stdout.String(), s)
			}

			if tc.wantStderrEmpty {
				require.Empty(t, stderr.String())
			}
		})
	}
}
