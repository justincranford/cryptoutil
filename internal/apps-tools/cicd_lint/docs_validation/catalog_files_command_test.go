// Copyright (c) 2025-2026 Justin Cranford.
package docs_validation

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testCatalogRootUnavailable   = "root unavailable"
	testCatalogReadFailure       = "failed to read"
	testCatalogRootLookupFailure = "root lookup failure"
	testViolationSummary         = "violation(s) found"
	testViolationsHeading        = "VIOLATIONS"
)

func TestCatalogFilesCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rootError  error
		appendix   string
		fileBody   string
		wantCode   int
		wantOutput string
		wantError  string
	}{
		{
			name:      testCatalogRootLookupFailure,
			rootError: errors.New(testCatalogRootUnavailable),
			wantCode:  1,
			wantError: testCatalogRootUnavailable,
		},
		{
			name:      "missing appendix",
			wantCode:  1,
			wantError: testCatalogReadFailure,
		},
		{
			name:       "catalog entry mismatch",
			appendix:   catalogCommandEntry("catalog target"),
			fileBody:   "stale content\n",
			wantCode:   1,
			wantOutput: testViolationsHeading,
			wantError:  testViolationSummary,
		},
		{
			name:       "catalog entry matches",
			appendix:   catalogCommandEntry(catalogTestBody),
			fileBody:   catalogTestBody + "\n",
			wantOutput: "All 1 catalog entries match",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if tc.appendix != "" {
				writeCatalogCommandFile(t, root, cryptoutilSharedMagic.CICDDocsAppendixDPath, tc.appendix)
				writeCatalogCommandFile(t, root, catalogTestTargetPath, tc.fileBody)
			}

			var stdout, stderr strings.Builder

			code := catalogFilesCommand(&stdout, &stderr, func() (string, error) {
				if tc.rootError != nil {
					return "", tc.rootError
				}

				return root, nil
			})

			require.Equal(t, tc.wantCode, code)

			if tc.wantOutput != "" {
				require.Contains(t, stdout.String(), tc.wantOutput)
			}

			if tc.wantError != "" {
				require.Contains(t, stderr.String(), tc.wantError)
			}
		})
	}
}

func TestCatalogPropagationCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		rootError  error
		handbook   string
		appendix   string
		wantCode   int
		wantOutput string
		wantError  string
	}{
		{
			name:      testCatalogRootLookupFailure,
			rootError: errors.New("root unavailable"),
			wantCode:  1,
			wantError: "root unavailable",
		},
		{
			name:      "missing files",
			wantCode:  1,
			wantError: "failed to read",
		},
		{
			name:       "catalog propagation mismatch",
			handbook:   catalogCommandHandbook(catalogTestBody),
			appendix:   catalogCommandPropagationEntry("stale content"),
			wantCode:   1,
			wantOutput: "VIOLATIONS",
			wantError:  "violation(s) found",
		},
		{
			name:       "catalog propagation matches",
			handbook:   catalogCommandHandbook(catalogTestBody),
			appendix:   catalogCommandPropagationEntry(catalogTestBody),
			wantOutput: "All 1 catalogued chunk(s) found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if tc.handbook != "" {
				writeCatalogCommandFile(t, root, cryptoutilSharedMagic.CICDDocsHandbookPath, tc.handbook)
				writeCatalogCommandFile(t, root, cryptoutilSharedMagic.CICDDocsAppendixDPath, tc.appendix)
			}

			var stdout, stderr strings.Builder

			code := catalogPropagationCommand(&stdout, &stderr, func() (string, error) {
				if tc.rootError != nil {
					return "", tc.rootError
				}

				return root, nil
			})

			require.Equal(t, tc.wantCode, code)

			if tc.wantOutput != "" {
				require.Contains(t, stdout.String(), tc.wantOutput)
			}

			if tc.wantError != "" {
				require.Contains(t, stderr.String(), tc.wantError)
			}
		})
	}
}

func TestCatalogCommandWrappers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command func(io.Writer, io.Writer) int
	}{
		{name: "catalog files", command: CatalogFilesCommand},
		{name: "catalog propagation", command: CatalogPropagationCommand},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr strings.Builder

			code := tc.command(&stdout, &stderr)
			require.Equal(t, 0, code)
			require.Empty(t, stderr.String())
		})
	}
}

func TestValidationCommandWrappers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command func(io.Writer, io.Writer) int
	}{
		{name: "agent drift", command: AgentDriftCommand},
		{name: "skill command drift", command: SkillCommandDriftCommand},
		{name: "propagation coverage", command: PropagationCoverageCommand},
		{name: "validate chunks", command: ValidateChunksCommand},
		{name: "validate coverage", command: ValidateCoverageCommand},
		{name: "validate propagation", command: ValidatePropagationCommand},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr strings.Builder
			require.Equal(t, 0, tc.command(&stdout, &stderr))
			require.Empty(t, stderr.String())
		})
	}
}

func catalogCommandEntry(body string) string {
	return "<!-- @file-catalog path=\"" + catalogTestTargetPath + "\" -->\n" +
		body + "\n<!-- @/file-catalog -->\n"
}

func catalogCommandHandbook(body string) string {
	return "<!-- @to-appendix as=\"" + catalogTestChunkID + "\" appendixes=\"" + catalogTestTargetPath + "\" -->\n" +
		body + "\n<!-- @/to-appendix -->\n"
}

func catalogCommandPropagationEntry(body string) string {
	return "<!-- @file-catalog path=\"" + catalogTestTargetPath + "\" -->\n" +
		"<!-- @from-eng-handbook as=\"" + catalogTestChunkID + "\" -->\n" +
		body + "\n<!-- @/from-eng-handbook -->\n<!-- @/file-catalog -->\n"
}

func writeCatalogCommandFile(t *testing.T, rootDir, relativePath, content string) {
	t.Helper()

	filePath := filepath.Join(rootDir, filepath.FromSlash(relativePath))
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
}
