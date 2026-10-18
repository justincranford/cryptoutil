// Copyright (c) 2025-2026 Justin Cranford.
package docs_validation

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestSplitMarkdownFrontmatter_ClosingDelimiterAtEOF(t *testing.T) {
	t.Parallel()

	frontmatter, body, err := splitMarkdownFrontmatter("---\nname: copilot-test\n---")

	require.NoError(t, err)
	require.Equal(t, "name: copilot-test", frontmatter)
	require.Empty(t, body)
}

func TestCheckAgentDrift_ReadAndParseErrors(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	agentsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDGithubAgentsDir)
	require.NoError(t, os.MkdirAll(filepath.Join(agentsDir, "nested"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute))

	for _, name := range []string{
		"read-error.agent.md",
		"copilot-parse-error.agent.md",
		"claude-parse-error.agent.md",
		"missing-claude.agent.md",
	} {
		require.NoError(t, os.WriteFile(
			filepath.Join(agentsDir, name),
			[]byte("placeholder"),
			cryptoutilSharedMagic.CacheFilePermissions,
		))
	}

	copilotParseErrorPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubAgentsDir, "copilot-parse-error.agent.md"))
	copilotParseErrorClaudePath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDClaudeAgentsDir, "copilot-parse-error.md"))
	claudeParseErrorCopilotPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubAgentsDir, "claude-parse-error.agent.md"))
	claudeParseErrorPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDClaudeAgentsDir, "claude-parse-error.md"))
	missingClaudeCopilotPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubAgentsDir, "missing-claude.agent.md"))
	missingClaudePath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDClaudeAgentsDir, "missing-claude.md"))
	validCopilot := []byte("---\nname: copilot-claude-parse-error\n---\nshared body\n")

	readFile := func(path string) ([]byte, error) {
		switch path {
		case filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubAgentsDir, "read-error.agent.md")):
			return nil, errors.New("injected Copilot read failure")
		case copilotParseErrorPath, copilotParseErrorClaudePath:
			return []byte("---\n: invalid: yaml: here\n---\n"), nil
		case claudeParseErrorCopilotPath:
			return validCopilot, nil
		case claudeParseErrorPath:
			return []byte("not frontmatter"), nil
		case missingClaudeCopilotPath:
			return []byte("---\nname: copilot-missing-claude\n---\nshared body\n"), nil
		case missingClaudePath:
			return nil, fs.ErrNotExist
		default:
			return nil, fs.ErrNotExist
		}
	}

	result, err := CheckAgentDrift(rootDir, readFile)

	require.NoError(t, err)
	require.Equal(t, 2, result.Checked)
	require.Len(t, result.Violations, 4)
	fields := []string{
		result.Violations[0].Field,
		result.Violations[1].Field,
		result.Violations[2].Field,
		result.Violations[3].Field,
	}
	require.ElementsMatch(t, []string{"file", valueFrontmatter, valueFrontmatter, valueMissing}, fields)
}

func TestAgentDriftCommand_RootLookupFailure(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := agentDriftCommand(&stdout, &stderr, func() (string, error) {
		return "", errors.New("injected root lookup failure")
	})

	require.Equal(t, 1, exitCode)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "injected root lookup failure")
}

func TestCheckCatalogFiles_ReadErrorsAndMismatches(t *testing.T) {
	t.Parallel()

	appendix := "<!-- @file-catalog path=\"" + catalogTestTargetPath + "\" -->\n" +
		catalogTestBody + "\n<!-- @/file-catalog -->\n" +
		"<!-- @file-catalog-pair copilot=\"" + catalogTestCopilotPath + "\" claude=\"" + catalogTestClaudePath + "\" -->\n" +
		"<!-- @copilot-frontmatter:start -->\n---\nname: copilot-catalog-test\n---\n<!-- @copilot-frontmatter:end -->\n" +
		"<!-- @claude-frontmatter:start -->\n---\nname: claude-catalog-test\n---\n<!-- @claude-frontmatter:end -->\n" +
		"<!-- @file-body:start -->\n" + catalogTestBody + "\n<!-- @file-body:end -->\n" +
		"<!-- @/file-catalog-pair -->\n"

	tests := []struct {
		name       string
		files      map[string][]byte
		wantFields []string
	}{
		{
			name: "disk read failures",
			files: map[string][]byte{
				cryptoutilSharedMagic.CICDDocsAppendixDPath: []byte(appendix),
			},
			wantFields: []string{valueDiskRead, valueDiskRead, valueDiskRead},
		},
		{
			name: "single and pair content mismatches",
			files: map[string][]byte{
				cryptoutilSharedMagic.CICDDocsAppendixDPath: []byte(appendix),
				catalogTestTargetPath:                       []byte("stale single\n"),
				catalogTestCopilotPath:                      []byte("stale Copilot\n"),
				catalogTestClaudePath:                       []byte("stale Claude\n"),
			},
			wantFields: []string{valueContentMismatch, valueContentMismatch, valueContentMismatch},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			readFile := catalogTestReadFile(tc.files)
			result, err := CheckCatalogFiles(".", readFile)

			require.NoError(t, err)
			require.Equal(t, len(tc.wantFields), result.Checked)
			require.Len(t, result.Violations, len(tc.wantFields))

			for i, violation := range result.Violations {
				require.Equal(t, tc.wantFields[i], violation.Field)
			}
		})
	}
}

func TestCheckCatalogPropagation_EmptyAndMissingSourceBlocks(t *testing.T) {
	t.Parallel()

	handbook := "<!-- @to-appendix as=\"" + catalogTestChunkID + "\" appendixes=\"   \" -->\n" +
		"<!-- @/to-appendix -->\n" +
		"<!-- @to-appendix as=\"" + catalogTestChunkID + "\" appendixes=\"" + catalogTestTargetPath + "\" -->\n" +
		catalogTestBody + "\n<!-- @/to-appendix -->\n"
	appendix := "<!-- @file-catalog path=\"" + catalogTestTargetPath + "\" -->\n" +
		"content without a source marker\n<!-- @/file-catalog -->\n"
	files := map[string][]byte{
		cryptoutilSharedMagic.CICDDocsHandbookPath:  []byte(handbook),
		cryptoutilSharedMagic.CICDDocsAppendixDPath: []byte(appendix),
	}

	result, err := CheckCatalogPropagation(".", catalogTestReadFile(files))

	require.NoError(t, err)
	require.Equal(t, 1, result.Checked)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "source-missing", result.Violations[0].Field)
}
