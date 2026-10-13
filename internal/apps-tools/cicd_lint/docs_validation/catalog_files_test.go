// Copyright (c) 2025-2026 Justin Cranford.
package docs_validation

import (
	"io/fs"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	catalogTestTargetPath     = ".catalog-target.md"
	catalogTestCopilotPath    = ".github/agents/catalog-test.agent.md"
	catalogTestClaudePath     = ".claude/agents/catalog-test.md"
	catalogTestChunkID        = "catalog-test-chunk"
	catalogTestBody           = "catalog test body"
	catalogTestEmptyAppendix  = "Appendix without catalog entries.\n"
	catalogTestNoEntriesError = "no file catalog entries found"
)

func TestCheckCatalogFiles_AppendixEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		appendix         string
		files            map[string]string
		wantChecked      int
		wantViolations   int
		wantErrorMessage string
	}{
		{
			name: "single-file entry matches disk",
			appendix: "<!-- @file-catalog path=\"" + catalogTestTargetPath + "\" -->\n" +
				catalogTestBody + "\n<!-- @/file-catalog -->\n",
			files:       map[string]string{catalogTestTargetPath: catalogTestBody + "\n"},
			wantChecked: 1,
		},
		{
			name: "single-file entry mismatch",
			appendix: "<!-- @file-catalog path=\"" + catalogTestTargetPath + "\" -->\n" +
				catalogTestBody + "\n<!-- @/file-catalog -->\n",
			files:       map[string]string{catalogTestTargetPath: "stale catalog body\n"},
			wantChecked: 1, wantViolations: 1,
		},
		{
			name: "pair entries match disk",
			appendix: "<!-- @file-catalog-pair copilot=\"" + catalogTestCopilotPath + "\" claude=\"" + catalogTestClaudePath + "\" -->\n" +
				"<!-- @copilot-frontmatter:start -->\n---\nname: copilot-catalog-test\n---\n<!-- @copilot-frontmatter:end -->\n" +
				"<!-- @claude-frontmatter:start -->\n---\nname: claude-catalog-test\n---\n<!-- @claude-frontmatter:end -->\n" +
				"<!-- @file-body:start -->\n" + catalogTestBody + "\n<!-- @file-body:end -->\n" +
				"<!-- @/file-catalog-pair -->\n",
			files: map[string]string{
				catalogTestCopilotPath: "---\nname: copilot-catalog-test\n---\n" + catalogTestBody + "\n",
				catalogTestClaudePath:  "---\nname: claude-catalog-test\n---\n" + catalogTestBody + "\n",
			},
			wantChecked: 2,
		},
		{
			name:             "catalog with no entries is rejected",
			appendix:         catalogTestEmptyAppendix,
			wantErrorMessage: catalogTestNoEntriesError,
		},
		{
			name:             "missing Appendix D is reported by file checker",
			wantErrorMessage: "failed to read " + cryptoutilSharedMagic.CICDDocsAppendixDPath,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			files := make(map[string][]byte, len(tc.files)+1)
			if tc.appendix != "" {
				files[cryptoutilSharedMagic.CICDDocsAppendixDPath] = []byte(tc.appendix)
			}

			for path, content := range tc.files {
				files[path] = []byte(content)
			}

			result, err := CheckCatalogFiles(".", catalogTestReadFile(files))
			if tc.wantErrorMessage != "" {
				require.ErrorContains(t, err, tc.wantErrorMessage)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantChecked, result.Checked)
			require.Len(t, result.Violations, tc.wantViolations)
		})
	}
}

func TestCheckCatalogPropagation_UsesHandbookAndAppendix(t *testing.T) {
	t.Parallel()

	handbook := "<!-- @to-appendix as=\"" + catalogTestChunkID + "\" appendixes=\"" + catalogTestTargetPath + "\" -->\n" +
		catalogTestBody + "\n<!-- @/to-appendix -->\n"
	appendix := "<!-- @file-catalog path=\"" + catalogTestTargetPath + "\" -->\n" +
		"<!-- @from-eng-handbook as=\"" + catalogTestChunkID + "\" -->\n" + catalogTestBody + "\n" +
		"<!-- @/from-eng-handbook -->\n<!-- @/file-catalog -->\n"

	tests := []struct {
		name             string
		files            map[string]string
		wantChecked      int
		wantViolations   int
		wantErrorMessage string
	}{
		{
			name: "matching propagated chunk",
			files: map[string]string{
				cryptoutilSharedMagic.CICDDocsHandbookPath:  handbook,
				cryptoutilSharedMagic.CICDDocsAppendixDPath: appendix,
			},
			wantChecked: 1,
		},
		{
			name: "mismatching catalog chunk",
			files: map[string]string{
				cryptoutilSharedMagic.CICDDocsHandbookPath: handbook,
				cryptoutilSharedMagic.CICDDocsAppendixDPath: "<!-- @file-catalog path=\"" + catalogTestTargetPath + "\" -->\n" +
					"<!-- @from-eng-handbook as=\"" + catalogTestChunkID + "\" -->\nstale catalog body\n" +
					"<!-- @/from-eng-handbook -->\n<!-- @/file-catalog -->\n",
			},
			wantChecked: 1, wantViolations: 1,
		},
		{
			name: "noncatalogued target is ignored",
			files: map[string]string{
				cryptoutilSharedMagic.CICDDocsHandbookPath:  handbook,
				cryptoutilSharedMagic.CICDDocsAppendixDPath: "<!-- @file-catalog path=\"unrelated.md\" -->\ncontent\n<!-- @/file-catalog -->\n",
			},
		},
		{
			name: "empty Appendix D catalog is rejected",
			files: map[string]string{
				cryptoutilSharedMagic.CICDDocsHandbookPath:  handbook,
				cryptoutilSharedMagic.CICDDocsAppendixDPath: catalogTestEmptyAppendix,
			},
			wantErrorMessage: catalogTestNoEntriesError,
		},
		{
			name: "missing handbook is reported",
			files: map[string]string{
				cryptoutilSharedMagic.CICDDocsAppendixDPath: appendix,
			},
			wantErrorMessage: "failed to read " + cryptoutilSharedMagic.CICDDocsHandbookPath,
		},
		{
			name: "missing Appendix D is reported by propagation checker",
			files: map[string]string{
				cryptoutilSharedMagic.CICDDocsHandbookPath: handbook,
			},
			wantErrorMessage: "failed to read " + cryptoutilSharedMagic.CICDDocsAppendixDPath,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			files := make(map[string][]byte, len(tc.files))
			for path, content := range tc.files {
				files[path] = []byte(content)
			}

			result, err := CheckCatalogPropagation(".", catalogTestReadFile(files))
			if tc.wantErrorMessage != "" {
				require.ErrorContains(t, err, tc.wantErrorMessage)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantChecked, result.Checked)
			require.Len(t, result.Violations, tc.wantViolations)
		})
	}
}

func catalogTestReadFile(files map[string][]byte) func(string) ([]byte, error) {
	return func(path string) ([]byte, error) {
		content, ok := files[path]
		if !ok {
			return nil, fs.ErrNotExist
		}

		return content, nil
	}
}
