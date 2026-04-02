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

func TestSplitMarkdownFrontmatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		content          string
		wantFrontmatter  string
		wantBody         string
		wantErrSubstring string
	}{
		{
			name:            "typical agent file",
			content:         "---\nname: copilot-beast\ndescription: Test agent.\n---\n# Body\n\nContent here.\n",
			wantFrontmatter: "name: copilot-beast\ndescription: Test agent.",
			wantBody:        "# Body\n\nContent here.\n",
		},
		{
			name:            "crlf normalized",
			content:         "---\r\nname: test\r\n---\r\n# Body\r\n",
			wantFrontmatter: "name: test",
			wantBody:        "# Body\n",
		},
		{
			name:            "empty body",
			content:         "---\nname: test\n---\n",
			wantFrontmatter: "name: test",
			wantBody:        "",
		},
		{
			name:             "no frontmatter delimiter",
			content:          "just content\nno delimiter",
			wantErrSubstring: "does not begin with YAML frontmatter delimiter",
		},
		{
			name:             "unclosed frontmatter",
			content:          "---\nname: test\nno closing delimiter",
			wantErrSubstring: "cannot find closing YAML frontmatter delimiter",
		},
		{
			name:            "multiline frontmatter",
			content:         "---\nname: copilot-test\ndescription: Multi-line.\nargument-hint: '<dir>'\n---\n# Body\n",
			wantFrontmatter: "name: copilot-test\ndescription: Multi-line.\nargument-hint: '<dir>'",
			wantBody:        "# Body\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fm, body, err := splitMarkdownFrontmatter(tc.content)

			if tc.wantErrSubstring != "" {
				require.ErrorContains(t, err, tc.wantErrSubstring)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantFrontmatter, fm)
			require.Equal(t, tc.wantBody, body)
		})
	}
}

func TestParseAgentFrontmatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		content          string
		wantName         string
		wantDescription  string
		wantArgHint      string
		wantBody         string
		wantErrSubstring string
	}{
		{
			name:            "all fields present",
			content:         "---\nname: copilot-beast-mode\ndescription: Autonomous execution agent.\nargument-hint: '<dir>'\n---\n# Body content\n",
			wantName:        "copilot-beast-mode",
			wantDescription: "Autonomous execution agent.",
			wantArgHint:     "<dir>",
			wantBody:        "# Body content\n",
		},
		{
			name:            "no argument-hint",
			content:         "---\nname: copilot-fix\ndescription: Fix workflows.\n---\n# Fix stuff.\n",
			wantName:        "copilot-fix",
			wantDescription: "Fix workflows.",
			wantArgHint:     "",
			wantBody:        "# Fix stuff.\n",
		},
		{
			name:             "invalid yaml",
			content:          "---\n: invalid: yaml: here\n---\n# Body\n",
			wantErrSubstring: "YAML parse error",
		},
		{
			name:             "no frontmatter",
			content:          "no frontmatter here",
			wantErrSubstring: "does not begin with YAML frontmatter delimiter",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fm, body, err := parseAgentFrontmatter(tc.content)

			if tc.wantErrSubstring != "" {
				require.ErrorContains(t, err, tc.wantErrSubstring)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantName, fm.Name)
			require.Equal(t, tc.wantDescription, fm.Description)
			require.Equal(t, tc.wantArgHint, fm.ArgumentHint)
			require.Equal(t, tc.wantBody, body)
		})
	}
}

const (
	testCopilotDescription = "Use this agent for autonomous execution."
	testBody               = "# BODY\n\nThis is the agent body content.\n\nIt spans multiple lines.\n"
)

// makeAgentFiles writes a matching Copilot/Claude agent pair under rootDir.
func makeAgentFiles(t *testing.T, rootDir, baseName, copilotDesc, claudeDesc, copilotArgHint, claudeArgHint, copilotBody, claudeBody string) {
	t.Helper()

	copilotFrontmatter := fmt.Sprintf("name: copilot-%s\ndescription: %s", baseName, copilotDesc)
	if copilotArgHint != "" {
		copilotFrontmatter += fmt.Sprintf("\nargument-hint: '%s'", copilotArgHint)
	}

	claudeFrontmatter := fmt.Sprintf("name: claude-%s\ndescription: %s", baseName, claudeDesc)
	if claudeArgHint != "" {
		claudeFrontmatter += fmt.Sprintf("\nargument-hint: '%s'", claudeArgHint)
	}

	copilotContent := fmt.Sprintf("---\n%s\n---\n%s", copilotFrontmatter, copilotBody)
	claudeContent := fmt.Sprintf("---\n%s\n---\n%s", claudeFrontmatter, claudeBody)

	copilotDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "agents")
	claudeDir := filepath.Join(rootDir, ".claude", "agents")

	require.NoError(t, os.MkdirAll(copilotDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(copilotDir, baseName+".agent.md"), []byte(copilotContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, baseName+".md"), []byte(claudeContent), cryptoutilSharedMagic.FilePermissionsDefault))
}

func TestCheckAgentDrift_AllPairsMatch(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	makeAgentFiles(t, rootDir, "beast-mode", testCopilotDescription, testCopilotDescription, "", "", testBody, testBody)
	makeAgentFiles(t, rootDir, "fix-workflows", "Fix workflows agent.", "Fix workflows agent.", "", "", testBody, testBody)

	result, err := CheckAgentDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Empty(t, result.Violations)
	require.Equal(t, 2, result.Checked)
}

func TestCheckAgentDrift_DescriptionMismatch(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	makeAgentFiles(t, rootDir, "beast-mode", "Short description.", "Long informative description.", "", "", testBody, testBody)

	result, err := CheckAgentDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "description", result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, "Short description")
	require.Contains(t, result.Violations[0].Detail, "Long informative description")
}

func TestCheckAgentDrift_MissingClaudeFile(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Write only the Copilot file.
	copilotDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "agents")
	claudeDir := filepath.Join(rootDir, ".claude", "agents")

	require.NoError(t, os.MkdirAll(copilotDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeDir, 0o700))

	content := fmt.Sprintf("---\nname: copilot-orphan\ndescription: %s\n---\n%s", testCopilotDescription, testBody)
	require.NoError(t, os.WriteFile(filepath.Join(copilotDir, "orphan.agent.md"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckAgentDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "missing", result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, ".claude/agents/orphan.md")
}

func TestCheckAgentDrift_ArgumentHintMismatch(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	makeAgentFiles(t, rootDir, "planner", testCopilotDescription, testCopilotDescription, "<dir>", "<path>", testBody, testBody)

	result, err := CheckAgentDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "argument-hint", result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, "<dir>")
	require.Contains(t, result.Violations[0].Detail, "<path>")
}

func TestCheckAgentDrift_BodyMismatch(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	claudeBody := testBody + "\n## Extra Section\n\nAdded to Claude only.\n"
	makeAgentFiles(t, rootDir, "executor", testCopilotDescription, testCopilotDescription, "", "", testBody, claudeBody)

	result, err := CheckAgentDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "body", result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, "body content differs")
}

func TestCheckAgentDrift_CRLFNormalization(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Copilot file uses CRLF, Claude file uses LF — should pass after normalization.
	copilotContent := "---\r\nname: copilot-crlf\r\ndescription: CRLF agent.\r\n---\r\n# Body\r\n\r\nContent.\r\n"
	claudeContent := "---\nname: claude-crlf\ndescription: CRLF agent.\n---\n# Body\n\nContent.\n"

	copilotDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "agents")
	claudeDir := filepath.Join(rootDir, ".claude", "agents")

	require.NoError(t, os.MkdirAll(copilotDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(copilotDir, "crlf.agent.md"), []byte(copilotContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "crlf.md"), []byte(claudeContent), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckAgentDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Empty(t, result.Violations, "CRLF vs LF differences must not be flagged as violations")
}

func TestCheckAgentDrift_NamePrefixViolation(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Copilot file uses wrong prefix (no "copilot-") in name field.
	copilotContent := fmt.Sprintf("---\nname: wrong-beast\ndescription: %s\n---\n%s", testCopilotDescription, testBody)
	claudeContent := fmt.Sprintf("---\nname: claude-beast\ndescription: %s\n---\n%s", testCopilotDescription, testBody)

	copilotDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "agents")
	claudeDir := filepath.Join(rootDir, ".claude", "agents")

	require.NoError(t, os.MkdirAll(copilotDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(copilotDir, "beast.agent.md"), []byte(copilotContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "beast.md"), []byte(claudeContent), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckAgentDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, cryptoutilSharedMagic.CICDAgentFrontMatterNameField, result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, `must have prefix "copilot-"`)
}

func TestCheckAgentDrift_ClaudeNameMismatch(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Claude file uses wrong name (not "claude-<base>").
	copilotContent := fmt.Sprintf("---\nname: copilot-beast\ndescription: %s\n---\n%s", testCopilotDescription, testBody)
	claudeContent := fmt.Sprintf("---\nname: wrong-beast\ndescription: %s\n---\n%s", testCopilotDescription, testBody)

	copilotDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "agents")
	claudeDir := filepath.Join(rootDir, ".claude", "agents")

	require.NoError(t, os.MkdirAll(copilotDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(copilotDir, "beast.agent.md"), []byte(copilotContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "beast.md"), []byte(claudeContent), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := CheckAgentDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, cryptoutilSharedMagic.CICDAgentFrontMatterNameField, result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, `"claude-beast"`)
}

func TestCheckAgentDrift_EmptyAgentsDir(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirGithubInstructions, "agents"), 0o700))
	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, ".claude", "agents"), 0o700))

	result, err := CheckAgentDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Empty(t, result.Violations)
	require.Equal(t, 0, result.Checked)
}

func TestCheckAgentDrift_AgentsDirMissing(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	// No .github/agents directory created.

	_, err := CheckAgentDrift(rootDir, rootedReadFile(rootDir))

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot read .github/agents")
}

func TestCheckAgentDrift_MultipleViolations(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Two pairs: one clean, one with description + body drift.
	makeAgentFiles(t, rootDir, "clean", testCopilotDescription, testCopilotDescription, "", "", testBody, testBody)

	driftedBody := testBody + "extra line only in Claude\n"
	makeAgentFiles(t, rootDir, "drifted", "Copilot desc.", "Claude desc.", "", "", testBody, driftedBody)

	result, err := CheckAgentDrift(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Equal(t, 2, result.Checked)
	require.Len(t, result.Violations, 2)

	fields := []string{result.Violations[0].Field, result.Violations[1].Field}
	require.Contains(t, fields, "description")
	require.Contains(t, fields, "body")
}

// Sequential: modifies findProjectRootFn package-level seam.
func TestAgentDriftCommand_NoPairsDir(t *testing.T) {
	// AgentDriftCommand reads from project root via findProjectRootFn.
	var stdout, stderr bytes.Buffer

	orig := findProjectRootFn

	t.Cleanup(func() { findProjectRootFn = orig })

	// Point root at a temp dir that has no .github/agents directory.
	tmpDir := t.TempDir()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	exitCode := AgentDriftCommand(&stdout, &stderr)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "cannot read .github/agents")
}

// Sequential: modifies findProjectRootFn package-level seam.
func TestAgentDriftCommand_AllClean(t *testing.T) {
	tmpDir := t.TempDir()

	makeAgentFiles(t, tmpDir, "beast-mode", testCopilotDescription, testCopilotDescription, "", "", testBody, testBody)

	orig := findProjectRootFn

	t.Cleanup(func() { findProjectRootFn = orig })

	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	var stdout, stderr bytes.Buffer

	exitCode := AgentDriftCommand(&stdout, &stderr)

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "All agent pairs are in sync")
	require.Empty(t, stderr.String())
}

// Sequential: modifies findProjectRootFn package-level seam.
func TestAgentDriftCommand_WithViolation(t *testing.T) {
	tmpDir := t.TempDir()

	makeAgentFiles(t, tmpDir, "drifted", "Copilot desc.", "Claude desc.", "", "", testBody, testBody)

	orig := findProjectRootFn

	t.Cleanup(func() { findProjectRootFn = orig })

	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	var stdout, stderr bytes.Buffer

	exitCode := AgentDriftCommand(&stdout, &stderr)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stdout.String(), "1 violation(s) found")
	require.Contains(t, stderr.String(), "agent-drift: 1 violation(s)")
}

func TestFormatAgentDriftResults_Clean(t *testing.T) {
	t.Parallel()

	result := &AgentDriftResult{Checked: 4}
	report := formatAgentDriftResults(result)

	require.Contains(t, report, "Checked 4 Copilot/Claude Code agent pairs")
	require.Contains(t, report, "All agent pairs are in sync")
	require.NotContains(t, report, "violation")
}

func TestFormatAgentDriftResults_WithViolations(t *testing.T) {
	t.Parallel()

	result := &AgentDriftResult{
		Checked: 3,
		Violations: []AgentDriftViolation{
			{
				CopilotFile: ".github/agents/foo.agent.md",
				ClaudeFile:  ".claude/agents/foo.md",
				Field:       "description",
				Detail:      "description mismatch: copilot vs claude",
			},
			{
				CopilotFile: ".github/agents/bar.agent.md",
				ClaudeFile:  ".claude/agents/bar.md",
				Field:       "body",
				Detail:      "body content differs",
			},
		},
	}

	report := formatAgentDriftResults(result)

	require.Contains(t, report, "Checked 3 Copilot/Claude Code agent pairs")
	require.Contains(t, report, "2 violation(s) found")
	require.Contains(t, report, "field=description")
	require.Contains(t, report, "field=body")
	require.Contains(t, report, ".github/agents/foo.agent.md")
	require.Contains(t, report, ".claude/agents/bar.md")

	// Verify ordering.
	require.True(t, strings.Index(report, "[1]") < strings.Index(report, "[2]"))
}

func TestFormatAgentDriftResults_MissingClaudeFile(t *testing.T) {
	t.Parallel()

	result := &AgentDriftResult{
		Checked: 1,
		Violations: []AgentDriftViolation{
			{
				CopilotFile: ".github/agents/orphan.agent.md",
				Field:       "missing",
				Detail:      "Claude Code agent file not found",
			},
		},
	}

	report := formatAgentDriftResults(result)

	require.Contains(t, report, "1 violation(s) found")
	require.Contains(t, report, ".github/agents/orphan.agent.md")
	// No claude line since ClaudeFile is empty.
	require.NotContains(t, report, "claude:  ")
}
