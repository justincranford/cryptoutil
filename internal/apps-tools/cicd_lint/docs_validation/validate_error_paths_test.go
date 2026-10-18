// Copyright (c) 2025-2026 Justin Cranford.
package docs_validation

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testFileZ            = "z.md"
	testCatalogPathField = "path"
)

const (
	testValidChunkID        = "valid-chunk"
	testCatalogContentField = "content"
)

func TestCoverageExtractorsCatalogAndReadErrors(t *testing.T) {
	t.Parallel()

	handbook := strings.Join([]string{
		cryptoutilSharedMagic.CICDFileCatalogStartMarker + ` path="catalog.md" -->`,
		`<!-- @to-appendix as="catalog-hidden" appendixes="target.md" -->`,
		cryptoutilSharedMagic.CICDFileCatalogEndMarker,
		cryptoutilSharedMagic.CICDFileCatalogPairEndMarker,
		`<!-- @to-appendix as="not_a_chunk" appendixes="target.md" -->`,
		`<!-- @to-appendix as="valid-chunk" appendixes="target.md" -->`,
		`<!-- @to-appendix as="unstable-13-4-id" appendixes="target.md" -->`,
		`<!-- @to-appendix as="empty-target" appendixes=", " -->`,
	}, "\n")

	readFile := func(path string) ([]byte, error) {
		if path != cryptoutilSharedMagic.CICDDocsHandbookPath {
			return nil, errors.New("unexpected path")
		}

		return []byte(handbook), nil
	}

	chunks, err := ExtractPropagateChunks(readFile)
	require.NoError(t, err)
	require.Equal(t, []string{"empty-target", "unstable-13-4-id", testValidChunkID}, chunks)

	targets, lineNumbers, issues, err := extractToAppendixMappings(readFile)
	require.NoError(t, err)
	require.Equal(t, []string{testTargetMd}, targets[testValidChunkID])
	require.NotZero(t, lineNumbers[testValidChunkID])
	require.NotContains(t, targets, "catalog-hidden")
	require.NotContains(t, targets, "not_a_chunk")
	require.Equal(t, []string{testTargetMd}, targets["unstable-13-4-id"])
	require.Contains(t, strings.Join(issues, "\n"), "unstable semantic chunk id")
	require.Contains(t, strings.Join(issues, "\n"), "empty appendix target")

	readError := func(string) ([]byte, error) {
		return nil, errors.New("injected read failure")
	}
	chunks, err = ExtractPropagateChunks(readError)
	require.ErrorContains(t, err, "failed to read docs/ENG-HANDBOOK.md")
	require.Nil(t, chunks)

	_, _, _, err = extractToAppendixMappings(readError)
	require.ErrorContains(t, err, "failed to read docs/ENG-HANDBOOK.md")
}

func TestValidateCoverage_AppendixMappingReadError(t *testing.T) {
	t.Parallel()

	handbookReads := 0
	readFile := func(path string) ([]byte, error) {
		switch path {
		case cryptoutilSharedMagic.CICDRequiredPropagationsManifest:
			return []byte(minimalManifestYAML()), nil
		case cryptoutilSharedMagic.CICDDocsHandbookPath:
			handbookReads++
			if handbookReads == 1 {
				return []byte(minimalArchitectureContent()), nil
			}

			return nil, errors.New("injected appendix mapping read failure")
		default:
			return nil, nil
		}
	}

	result, err := validateCoverage(t.TempDir(), readFile, ExtractSourceChunks)

	require.ErrorContains(t, err, "injected appendix mapping read failure")
	require.Nil(t, result)
}

func TestExtractSourceChunks_SkillReadError(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	skillPath := filepath.Join(rootDir, cryptoutilSharedMagic.CICDGithubSkillsDir, "broken", cryptoutilSharedMagic.CICDSkillFileName)
	require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(skillPath, []byte(sourceInstructionContent()), cryptoutilSharedMagic.CacheFilePermissions))

	readFile := func(string) ([]byte, error) {
		return nil, errors.New("injected read failure")
	}
	result, err := ExtractSourceChunks(rootDir, readFile)

	require.NoError(t, err)
	require.Empty(t, result)
}

func TestComputeCoverage_SkipsDirectoriesNonmatchesAndReadErrors(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	instructionsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDGithubInstructionsDir)
	require.NoError(t, os.MkdirAll(filepath.Join(instructionsDir, "nested"), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(instructionsDir, "readme.md"), []byte(cryptoutilSharedMagic.PKCEMethodPlain), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(instructionsDir, "01-fails.instructions.md"), []byte("content"), cryptoutilSharedMagic.CacheFilePermissions))

	readFile := func(string) ([]byte, error) {
		return nil, errors.New("injected read failure")
	}
	result, err := ComputeCoverage(rootDir, readFile)

	require.NoError(t, err)
	require.Zero(t, result.TotalFiles)
	require.Zero(t, result.CoveredFiles)
	require.Empty(t, result.ZeroCoverageFiles)
}

func TestValidatePropagation_ScannerAndDisplaySkips(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	instructionsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDGithubInstructionsDir)
	agentsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDGithubAgentsDir)
	docsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirDocs)
	require.NoError(t, os.MkdirAll(filepath.Join(instructionsDir, "nested"), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.MkdirAll(agentsDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.MkdirAll(docsDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	handbook := "## 1.1 Numeric Heading\n## Plain Heading\n"
	instructionPath := filepath.Join(instructionsDir, "display.instructions.md")
	instructionContent := strings.Join([]string{
		`See [ ](../../docs/ENG-HANDBOOK.md#11-numeric-heading).`,
		`See [Plain](../../docs/ENG-HANDBOOK.md#plain-heading).`,
		`See [Section 2.1](../../docs/ENG-HANDBOOK.md#plain-heading).`,
	}, "\n")
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, filepath.Base(cryptoutilSharedMagic.CICDDocsHandbookPath)), []byte(handbook), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(instructionPath, []byte(instructionContent), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(instructionsDir, "README.md"), []byte("not matched"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(instructionsDir, "failed.instructions.md"), []byte("not readable"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(agentsDir, "failed.agent.md"), []byte("not readable"), cryptoutilSharedMagic.CacheFilePermissions))

	readFile := func(relPath string) ([]byte, error) {
		relPath = filepath.ToSlash(relPath)

		switch relPath {
		case cryptoutilSharedMagic.CICDDocsHandbookPath:
			return []byte(handbook), nil
		case cryptoutilSharedMagic.CICDCopilotInstructionsFile,
			filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubInstructionsDir, "failed.instructions.md")),
			filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubAgentsDir, "failed.agent.md")):
			return nil, errors.New("injected read failure")
		default:
			return os.ReadFile(filepath.Join(rootDir, filepath.FromSlash(relPath)))
		}
	}

	result, err := ValidatePropagation(rootDir, readFile)
	require.NoError(t, err)
	require.Len(t, result.ValidRefs, strings.Count(instructionContent, "]("))
	require.Empty(t, result.BrokenRefs)
	require.Empty(t, result.DisplayTextWarnings)

	var stdout, stderr bytes.Buffer
	require.Zero(t, validatePropagationWithRoot(rootDir, &stdout, &stderr))
	require.Empty(t, stderr.String())
}

func TestFormatCatalogCommandResultsWithViolations(t *testing.T) {
	t.Parallel()

	fileReport := formatCatalogFilesResults(&CatalogFilesResult{
		Checked: 2,
		Violations: []CatalogViolation{
			{File: testFileZ, Field: testCatalogContentField, Detail: testMismatch},
			{File: testFileA, Field: testCatalogPathField, Detail: valueMissing},
			{File: testFileA, Field: testCatalogContentField, Detail: valueMissing},
		},
	})
	require.Contains(t, fileReport, "VIOLATIONS (3)")
	require.Less(t, strings.Index(fileReport, testFileA), strings.Index(fileReport, testFileZ))
	require.Less(t, strings.Index(fileReport, "["+testCatalogContentField+"] "+testFileA), strings.Index(fileReport, "["+testCatalogPathField+"] "+testFileA))

	propagationReport := formatCatalogPropagationResults(&CatalogPropagationResult{
		Checked: 2,
		Violations: []CatalogViolation{
			{File: testFileZ, Field: testCatalogContentField, Detail: testMismatch},
			{File: testFileA, Field: testCatalogPathField, Detail: valueMissing},
			{File: testFileA, Field: testCatalogContentField, Detail: valueMissing},
		},
	})
	require.Contains(t, propagationReport, "VIOLATIONS (3)")
	require.Less(t, strings.Index(propagationReport, testFileA), strings.Index(propagationReport, testFileZ))
	require.Less(t, strings.Index(propagationReport, "["+testCatalogContentField+"] "+testFileA), strings.Index(propagationReport, "["+testCatalogPathField+"] "+testFileA))
}

func TestFormatCatalogCommandResultsWithoutViolations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format func() string
		want   string
	}{
		{
			name: testCatalogFilesCaseName,
			format: func() string {
				return formatCatalogFilesResults(&CatalogFilesResult{Checked: 2})
			},
			want: "All 2 catalog entries match their files on disk.",
		},
		{
			name: testCatalogPropagationCaseName,
			format: func() string {
				return formatCatalogPropagationResults(&CatalogPropagationResult{Checked: testCatalogPropagationCheckedCount})
			},
			want: fmt.Sprintf("All %d catalogued chunk(s) found with matching content.", testCatalogPropagationCheckedCount),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			report := tc.format()

			require.Contains(t, report, tc.want)
			require.NotContains(t, report, "VIOLATIONS")
		})
	}
}

func TestFormatCoverageValidationResults_CompositionIssues(t *testing.T) {
	t.Parallel()

	report := FormatCoverageValidationResults(&CoverageResult{
		CompositionIssues: []string{"first issue", "second issue"},
	})

	require.Contains(t, report, "SECTION/APPENDIX COMPOSITION ISSUES (2, review-order)")
	require.Contains(t, report, "first issue")
	require.Contains(t, report, "second issue")
	require.Contains(t, report, "Fix @to-appendix marker composition issues")
}

func TestFormatChunkResults_SortsAllIssueGroups(t *testing.T) {
	t.Parallel()

	result := &ChunkValidationResult{
		Results: []ChunkResult{
			{ChunkID: "z-file", PropagateBlock: PropagateBlock{TargetFile: testFileZ, LineNumber: 2}, Status: ChunkStatusFileNotFound},
			{ChunkID: "a-file", PropagateBlock: PropagateBlock{TargetFile: testFileA, LineNumber: 1}, Status: ChunkStatusFileNotFound},
			{ChunkID: "z-missing", PropagateBlock: PropagateBlock{TargetFile: testFileZ, LineNumber: 4}, Status: ChunkStatusMissing},
			{ChunkID: "a-missing", PropagateBlock: PropagateBlock{TargetFile: testFileA, LineNumber: 3}, Status: ChunkStatusMissing},
			{ChunkID: "z-mismatch", PropagateBlock: PropagateBlock{TargetFile: testFileZ, LineNumber: 6}, SourceBlock: &SourceBlock{LineNumber: 8}, Status: ChunkStatusMismatch},
			{ChunkID: "a-mismatch", PropagateBlock: PropagateBlock{TargetFile: testFileA, LineNumber: 5}, SourceBlock: &SourceBlock{LineNumber: 7}, Status: ChunkStatusMismatch},
			{ChunkID: "matched", Status: ChunkStatusMatch},
		},
		Matched:    1,
		Mismatched: 2,
		Missing:    2,
		FileErrors: 2,
	}

	report := FormatChunkResults(result)

	require.Contains(t, report, "FILE NOT FOUND (2)")
	require.Contains(t, report, "MISSING @from-eng-handbook BLOCKS (2)")
	require.Contains(t, report, "CONTENT MISMATCHES (2)")
	require.Less(t, strings.Index(report, "a-file"), strings.Index(report, "z-file"))
	require.Less(t, strings.Index(report, "a-missing"), strings.Index(report, "z-missing"))
	require.Less(t, strings.Index(report, "a-mismatch"), strings.Index(report, "z-mismatch"))
}

func TestValidateChunksWithRoot_CleanResult(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	docsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirDocs)
	require.NoError(t, os.MkdirAll(docsDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, filepath.Base(cryptoutilSharedMagic.CICDDocsHandbookPath)), []byte("# Empty handbook\n"), cryptoutilSharedMagic.CacheFilePermissions))

	var stdout, stderr bytes.Buffer

	exitCode := validateChunksWithRoot(rootDir, &stdout, &stderr)

	require.Zero(t, exitCode)
	require.Empty(t, stderr.String())
	require.Contains(t, stdout.String(), "All propagated chunks are in sync")
}

func TestValidateChunksWithRoot_MissingTargetFile(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	docsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirDocs)
	require.NoError(t, os.MkdirAll(docsDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	handbook := join(
		`<!-- @to-appendix as="missing-target" appendixes="`+testFileA+`" -->`,
		"canonical content",
		"<!-- @/to-appendix -->",
	)
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, filepath.Base(cryptoutilSharedMagic.CICDDocsHandbookPath)), []byte(handbook), cryptoutilSharedMagic.CacheFilePermissions))

	var stdout, stderr bytes.Buffer

	exitCode := validateChunksWithRoot(rootDir, &stdout, &stderr)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stdout.String(), "FILE NOT FOUND")
	require.Empty(t, stderr.String())
}

func TestValidatePropagationWithRoot_BrokenReference(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	docsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirDocs)
	instructionsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDGithubInstructionsDir)

	require.NoError(t, os.MkdirAll(docsDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.MkdirAll(instructionsDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, filepath.Base(cryptoutilSharedMagic.CICDDocsHandbookPath)), []byte("## Existing Heading\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(instructionsDir, "broken.instructions.md"), []byte("See [Missing](../../docs/ENG-HANDBOOK.md#missing-heading).\n"), cryptoutilSharedMagic.CacheFilePermissions))

	var stdout, stderr bytes.Buffer

	exitCode := validatePropagationWithRoot(rootDir, &stdout, &stderr)

	require.Equal(t, 1, exitCode)
	require.Contains(t, stdout.String(), "BROKEN REFERENCES (1)")
	require.Empty(t, stderr.String())
}

func TestValidatePropagation_EmptyDisplayText(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	docsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDExcludeDirDocs)
	instructionsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDGithubInstructionsDir)

	require.NoError(t, os.MkdirAll(docsDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.MkdirAll(instructionsDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, filepath.Base(cryptoutilSharedMagic.CICDDocsHandbookPath)), []byte("## Valid Heading\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(instructionsDir, "empty-link.instructions.md"), []byte("[](../../docs/ENG-HANDBOOK.md#valid-heading)\n"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidatePropagation(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Len(t, result.ValidRefs, 1)
	require.Empty(t, result.ValidRefs[0].DisplayText)
	require.Empty(t, result.DisplayTextWarnings)
}

func TestSkillCommandDriftErrorPaths(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	rootErr := skillCommandDriftCommand(&stdout, &stderr, func() (string, error) {
		return "", errors.New("injected root error")
	})
	require.Equal(t, 1, rootErr)
	require.Contains(t, stderr.String(), "cannot determine project root")

	stdout.Reset()
	stderr.Reset()
	checkErr := skillCommandDriftCommand(&stdout, &stderr, func() (string, error) {
		return t.TempDir(), nil
	})
	require.Equal(t, 1, checkErr)
	require.Contains(t, stderr.String(), "cannot read")
}

func TestSkillCommandDrift_ReportsSkillReadFailureAfterDiscovery(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, cryptoutilSharedMagic.CICDGithubSkillsDir, "disappearing"), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.MkdirAll(filepath.Join(rootDir, cryptoutilSharedMagic.CICDClaudeSkillsDir), cryptoutilSharedMagic.CICDTempDirPermissions))

	skillPath := filepath.ToSlash(filepath.Join(cryptoutilSharedMagic.CICDGithubSkillsDir, "disappearing", cryptoutilSharedMagic.CICDSkillFileName))
	readCount := 0
	readFile := func(path string) ([]byte, error) {
		if path != skillPath {
			return nil, errors.New("unexpected path")
		}

		readCount++
		if readCount == 1 {
			return []byte("skill content"), nil
		}

		return nil, errors.New("injected transient read failure")
	}

	result, err := CheckSkillCommandDrift(rootDir, readFile)

	require.NoError(t, err)
	require.Equal(t, 2, readCount)
	require.Equal(t, 1, result.Checked)
	require.Len(t, result.Violations, 1)
	require.Equal(t, valueMissingSkillFile, result.Violations[0].Field)
	require.Contains(t, result.Violations[0].Detail, "injected transient read failure")
}

func TestCheckChunkVerificationCommandRootError(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := checkChunkVerificationCommand(&stdout, &stderr, func() (string, error) {
		return "", errors.New("injected root error")
	})

	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "injected root error")
	require.Empty(t, stdout.String())
}

func TestValidateCoverage_SortsMissingChunksByIDWhenHandbookLineIsUnknown(t *testing.T) {
	t.Parallel()

	manifest := `required_propagations:
  - chunk_id: z-chunk
    source_file: docs/ENG-HANDBOOK.md
    required_targets:
      - .github/instructions/z.instructions.md
  - chunk_id: a-chunk
    source_file: docs/ENG-HANDBOOK.md
    required_targets:
      - .github/instructions/a.instructions.md
`
	readFile := makeFakeReadFile(map[string]string{
		cryptoutilSharedMagic.CICDRequiredPropagationsManifest: manifest,
		cryptoutilSharedMagic.CICDDocsHandbookPath:             "# Handbook without appendix markers",
	})

	result, err := validateCoverage(t.TempDir(), readFile, ExtractSourceChunks)

	require.NoError(t, err)
	require.Len(t, result.Violations, 2)
	require.Equal(t, "a-chunk", result.Violations[0].ChunkID)
	require.Equal(t, "z-chunk", result.Violations[1].ChunkID)
}

// Sequential: changes the process working directory.
func TestFindProjectRootWithoutGoMod(t *testing.T) {
	testDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})
	require.NoError(t, os.Chdir(testDir))

	root, err := findProjectRoot()
	require.Empty(t, root)
	require.ErrorContains(t, err, "go.mod not found")
}
