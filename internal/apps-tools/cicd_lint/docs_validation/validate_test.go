// Copyright (c) 2025-2026 Justin Cranford.
package docs_validation

import (
	"fmt"
	"os"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// rfcOnlySourceContent returns instruction file content with ONLY the rfc-2119-keywords @from-eng-handbook block.
// Used in tests where emphasis-keywords is intentionally missing.
func rfcOnlySourceContent() string {
	return `<!-- @from-eng-handbook as="rfc-2119-keywords" -->
content
<!-- @/from-eng-handbook -->
`
}

// minimalManifestYAML returns a valid YAML manifest with two entries.
func minimalManifestYAML() string {
	return `required_propagations:
  - chunk_id: rfc-2119-keywords
    source_file: docs/ENG-HANDBOOK.md
    required_targets:
      - .github/instructions/01-01.terminology.instructions.md
  - chunk_id: emphasis-keywords
    source_file: docs/ENG-HANDBOOK.md
    required_targets:
      - .github/instructions/01-01.terminology.instructions.md
`
}

// minimalArchitectureContent returns ENG-HANDBOOK.md content with two @to-appendix markers.
func minimalArchitectureContent() string {
	return `# Architecture
<!-- @to-appendix as="rfc-2119-keywords" appendixes=".github/instructions/01-01.terminology.instructions.md" -->
some content
<!-- @/to-appendix -->
<!-- @to-appendix as="emphasis-keywords" appendixes=".github/instructions/01-01.terminology.instructions.md" -->
more content
<!-- @/to-appendix -->
`
}

// minimalArchitectureWithTargetMismatchContent returns ENG-HANDBOOK.md content with a target list mismatch.
func minimalArchitectureWithTargetMismatchContent() string {
	return `# Architecture
<!-- @to-appendix as="rfc-2119-keywords" appendixes=".github/instructions/other.instructions.md" -->
some content
<!-- @/to-appendix -->
`
}

// sourceInstructionContent returns instruction file content with @from-eng-handbook blocks.
func sourceInstructionContent() string {
	return `# Instruction File
<!-- @from-eng-handbook as="rfc-2119-keywords" -->
RFC keyword content
<!-- @/from-eng-handbook -->
<!-- @from-eng-handbook as="emphasis-keywords" -->
Emphasis content
<!-- @/from-eng-handbook -->
`
}

// makeFakeReadFile creates a readFile func backed by a map[relPath]content.
func makeFakeReadFile(files map[string]string) func(string) ([]byte, error) {
	return func(path string) ([]byte, error) {
		if content, ok := files[path]; ok {
			return []byte(content), nil
		}

		return nil, fmt.Errorf("file not found: %s", path)
	}
}

// buildValidateRoot sets up a temp dir with manifest, architecture, and instruction files.
func buildValidateRoot(t *testing.T, manifestContent, archContent string, instructionFiles map[string]string) (string, func(string) ([]byte, error)) {
	t.Helper()

	rootDir := t.TempDir()
	docsDir := rootDir + "/docs"
	require.NoError(t, os.MkdirAll(docsDir, 0o700))
	require.NoError(t, os.WriteFile(docsDir+"/required-propagations.yaml", []byte(manifestContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(docsDir+"/ENG-HANDBOOK.md", []byte(archContent), cryptoutilSharedMagic.FilePermissionsDefault))

	instrDir := rootDir + "/.github/instructions"
	require.NoError(t, os.MkdirAll(instrDir, 0o700))

	for name, content := range instructionFiles {
		require.NoError(t, os.WriteFile(instrDir+"/"+name, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))
	}

	return rootDir, rootedReadFile(rootDir)
}

// --- LoadPropagationsManifest ---

func TestLoadPropagationsManifest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   map[string]string
		wantErr string
		wantLen int
	}{
		{
			name:    "happy path",
			files:   map[string]string{cryptoutilSharedMagic.CICDRequiredPropagationsManifest: minimalManifestYAML()},
			wantLen: 2,
		},
		{
			name:    "invalid YAML",
			files:   map[string]string{cryptoutilSharedMagic.CICDRequiredPropagationsManifest: "!!! not valid: yaml: ["},
			wantErr: "failed to parse",
		},
		{
			name:    "file not found",
			files:   map[string]string{},
			wantErr: "failed to read",
		},
		{
			name:  "empty manifest",
			files: map[string]string{cryptoutilSharedMagic.CICDRequiredPropagationsManifest: "required_propagations: []\n"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			manifest, err := LoadPropagationsManifest(makeFakeReadFile(tc.files))
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Nil(t, manifest)
				require.Contains(t, err.Error(), tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Len(t, manifest.RequiredPropagations, tc.wantLen)

			if tc.wantLen > 0 {
				require.Equal(t, "rfc-2119-keywords", manifest.RequiredPropagations[0].ChunkID)
				require.Equal(t, "docs/ENG-HANDBOOK.md", manifest.RequiredPropagations[0].SourceFile)
				require.Len(t, manifest.RequiredPropagations[0].RequiredTargets, 1)
			}
		})
	}
}

// --- ExtractPropagateChunks ---

func TestExtractPropagateChunks(t *testing.T) {
	t.Parallel()

	deduplicatedContent := `<!-- @to-appendix as="my-chunk" appendixes=".github/instructions/a.md" -->
some content
<!-- @/to-appendix -->
<!-- @to-appendix as="my-chunk" appendixes=".github/instructions/b.md" -->
some content
<!-- @/to-appendix -->
`
	grammarFilterContent := `@to-appendix-open  ::= '<!-- @to-appendix as="' CHUNK_ID '" appendixes="' PATH_LIST '" -->'
<!-- @to-appendix as="valid-chunk" appendixes=".github/instructions/a.md" -->
some content
<!-- @/to-appendix -->
`

	tests := []struct {
		name       string
		files      map[string]string
		wantErr    string
		wantChunks []string
	}{
		{
			name:       "happy path",
			files:      map[string]string{"docs/ENG-HANDBOOK.md": minimalArchitectureContent()},
			wantChunks: []string{"emphasis-keywords", "rfc-2119-keywords"},
		},
		{
			name:       "deduplicated",
			files:      map[string]string{"docs/ENG-HANDBOOK.md": deduplicatedContent},
			wantChunks: []string{"my-chunk"},
		},
		{
			name:       "filters invalid chunk IDs",
			files:      map[string]string{"docs/ENG-HANDBOOK.md": grammarFilterContent},
			wantChunks: []string{"valid-chunk"},
		},
		{
			name:  "empty file",
			files: map[string]string{"docs/ENG-HANDBOOK.md": ""},
		},
		{
			name:    "file not found",
			files:   map[string]string{},
			wantErr: "failed to read docs/ENG-HANDBOOK.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			chunks, err := ExtractPropagateChunks(makeFakeReadFile(tc.files))
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Nil(t, chunks)
				require.Contains(t, err.Error(), tc.wantErr)

				return
			}

			require.NoError(t, err)

			if tc.wantChunks == nil {
				require.Empty(t, chunks)
			} else {
				require.Equal(t, tc.wantChunks, chunks)
			}
		})
	}
}

// --- extractSourceChunksFromContent ---

func TestExtractSourceChunksFromContent(t *testing.T) {
	t.Parallel()

	singleContent := `<!-- @from-eng-handbook as="my-chunk" -->
content
<!-- @/from-eng-handbook -->
`
	multiContent := `<!-- @from-eng-handbook as="chunk-a" -->
a
<!-- @/from-eng-handbook -->
<!-- @from-eng-handbook as="chunk-b" -->
b
<!-- @/from-eng-handbook -->
`
	grammarContent := `@from-eng-handbook-open     ::= '<!-- @from-eng-handbook as="' CHUNK_ID '" -->'
<!-- @from-eng-handbook as="valid-chunk" -->
`

	tests := []struct {
		name          string
		filePath      string
		content       string
		initialResult map[string][]string
		validate      func(t *testing.T, result map[string][]string)
	}{
		{
			name: "single match", filePath: "instructions/test.md", content: singleContent,
			validate: func(t *testing.T, result map[string][]string) {
				t.Helper()
				require.Len(t, result, 1)
				require.Equal(t, []string{"instructions/test.md"}, result["my-chunk"])
			},
		},
		{
			name: "multiple matches", filePath: "file.md", content: multiContent,
			validate: func(t *testing.T, result map[string][]string) {
				t.Helper()
				require.Len(t, result, 2)
				require.Equal(t, []string{"file.md"}, result["chunk-a"])
				require.Equal(t, []string{"file.md"}, result["chunk-b"])
			},
		},
		{
			name: "no matches", filePath: "file.md", content: "no @from-eng-handbook blocks here",
			validate: func(t *testing.T, result map[string][]string) {
				t.Helper()
				require.Empty(t, result)
			},
		},
		{
			name: "filters invalid chunk IDs", filePath: "file.md", content: grammarContent,
			validate: func(t *testing.T, result map[string][]string) {
				t.Helper()
				require.Len(t, result, 1)
				require.Contains(t, result, "valid-chunk")
				require.NotContains(t, result, "' CHUNK_ID '")
			},
		},
		{
			name: "appends to prior entries", filePath: "second.md",
			content:       `<!-- @from-eng-handbook as="my-chunk" -->`,
			initialResult: map[string][]string{"my-chunk": {"already-there.md"}},
			validate: func(t *testing.T, result map[string][]string) {
				t.Helper()
				require.Len(t, result["my-chunk"], 2)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := make(map[string][]string)

			if tc.initialResult != nil {
				for k, v := range tc.initialResult {
					result[k] = v
				}
			}

			extractSourceChunksFromContent(tc.filePath, tc.content, result)
			tc.validate(t, result)
		})
	}
}

func TestExtractSourceChunks_IncludesSkills(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	docsDir := rootDir + "/docs"
	require.NoError(t, os.MkdirAll(docsDir, 0o700))
	require.NoError(t, os.WriteFile(docsDir+"/ENG-HANDBOOK.md", []byte("# Architecture\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	ghSkillDir := rootDir + "/.github/skills/sample-skill"
	claudeSkillDir := rootDir + "/.claude/skills/sample-skill"

	require.NoError(t, os.MkdirAll(ghSkillDir, 0o700))
	require.NoError(t, os.MkdirAll(claudeSkillDir, 0o700))

	skillContent := `<!-- @from-eng-handbook as="sample-skill-chunk" -->
sample
<!-- @/from-eng-handbook -->
`
	require.NoError(t, os.WriteFile(ghSkillDir+"/SKILL.md", []byte(skillContent), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(claudeSkillDir+"/SKILL.md", []byte(skillContent), cryptoutilSharedMagic.FilePermissionsDefault))

	chunks, err := ExtractSourceChunks(rootDir, rootedReadFile(rootDir))
	require.NoError(t, err)
	require.Contains(t, chunks, "sample-skill-chunk")
	require.Contains(t, chunks["sample-skill-chunk"], ".github/skills/sample-skill/SKILL.md")
	require.Contains(t, chunks["sample-skill-chunk"], ".claude/skills/sample-skill/SKILL.md")
}

// --- ValidateCoverage ---

func TestValidateCoverage(t *testing.T) {
	t.Parallel()

	archWithExtra := minimalArchitectureContent() + `<!-- @to-appendix as="extra-chunk" appendixes=".github/instructions/x.md" -->
extra content
<!-- @/to-appendix -->
`

	tests := []struct {
		name             string
		manifestContent  string
		archContent      string
		instructionFiles map[string]string
		validate         func(t *testing.T, result *CoverageResult)
	}{
		{
			name:             "all chunks covered",
			manifestContent:  minimalManifestYAML(),
			archContent:      minimalArchitectureContent(),
			instructionFiles: map[string]string{"01-01.terminology.instructions.md": sourceInstructionContent()},
			validate: func(t *testing.T, result *CoverageResult) {
				t.Helper()
				require.Empty(t, result.Violations)
				require.Empty(t, result.OrphanedChunks)
				require.Equal(t, 2, result.ManifestChunks)
				require.Equal(t, 2, result.ArchitectureChunks)
			},
		},
		{
			name:             "missing from-eng-handbook block",
			manifestContent:  minimalManifestYAML(),
			archContent:      minimalArchitectureContent(),
			instructionFiles: map[string]string{"01-01.terminology.instructions.md": rfcOnlySourceContent()},
			validate: func(t *testing.T, result *CoverageResult) {
				t.Helper()
				require.Len(t, result.Violations, 1)
				require.Equal(t, "emphasis-keywords", result.Violations[0].ChunkID)
				require.Contains(t, result.Violations[0].Description, "emphasis-keywords")
			},
		},
		{
			name:             "orphaned chunk",
			manifestContent:  minimalManifestYAML(),
			archContent:      archWithExtra,
			instructionFiles: map[string]string{"01-01.terminology.instructions.md": sourceInstructionContent()},
			validate: func(t *testing.T, result *CoverageResult) {
				t.Helper()
				require.Len(t, result.OrphanedChunks, 1)
				require.Equal(t, "extra-chunk", result.OrphanedChunks[0])
			},
		},
		{
			name:             "violations and orphans together",
			manifestContent:  minimalManifestYAML(),
			archContent:      archWithExtra,
			instructionFiles: map[string]string{"01-01.terminology.instructions.md": rfcOnlySourceContent()},
			validate: func(t *testing.T, result *CoverageResult) {
				t.Helper()
				require.Len(t, result.Violations, 1)
				require.Len(t, result.OrphanedChunks, 1)
			},
		},
		{
			name:             "to-appendix target mismatch against manifest",
			manifestContent:  minimalManifestYAML(),
			archContent:      minimalArchitectureWithTargetMismatchContent(),
			instructionFiles: map[string]string{"01-01.terminology.instructions.md": rfcOnlySourceContent()},
			validate: func(t *testing.T, result *CoverageResult) {
				t.Helper()
				require.NotEmpty(t, result.CompositionIssues)
				require.Contains(t, strings.Join(result.CompositionIssues, "\n"), "manifest target")
			},
		},
		{
			name:            "unstable numeric semantic chunk ids are rejected",
			manifestContent: minimalManifestYAML(),
			archContent: `# Architecture
<!-- @to-appendix as="section-13-4-rules" appendixes=".github/instructions/01-01.terminology.instructions.md" -->
some content
<!-- @/to-appendix -->
`,
			instructionFiles: map[string]string{"01-01.terminology.instructions.md": rfcOnlySourceContent()},
			validate: func(t *testing.T, result *CoverageResult) {
				t.Helper()
				require.NotEmpty(t, result.CompositionIssues)
				require.Contains(t, strings.Join(result.CompositionIssues, "\n"), "unstable semantic chunk id")
			},
		},
		{
			name:            "empty appendix target rejected",
			manifestContent: minimalManifestYAML(),
			archContent: `# Architecture
<!-- @to-appendix as="rfc-2119-keywords" appendixes=".github/instructions/01-01.terminology.instructions.md, " -->
some content
<!-- @/to-appendix -->
`,
			instructionFiles: map[string]string{"01-01.terminology.instructions.md": rfcOnlySourceContent()},
			validate: func(t *testing.T, result *CoverageResult) {
				t.Helper()
				require.NotEmpty(t, result.CompositionIssues)
				require.Contains(t, strings.Join(result.CompositionIssues, "\n"), "empty appendix target")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir, readFile := buildValidateRoot(t, tc.manifestContent, tc.archContent, tc.instructionFiles)
			result, err := ValidateCoverage(rootDir, readFile)

			require.NoError(t, err)
			tc.validate(t, result)
		})
	}
}

func TestValidateCoverage_ManifestLoadError(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	result, err := ValidateCoverage(rootDir, rootedReadFile(rootDir))

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to read")
}

func TestValidateCoverage_ExtractSourceChunksError(t *testing.T) {
	t.Parallel()

	rootDir, readFile := buildValidateRoot(t, minimalManifestYAML(), minimalArchitectureContent(), map[string]string{})

	result, err := validateCoverage(rootDir, readFile, func(_ string, _ func(string) ([]byte, error)) (map[string][]string, error) {
		return nil, fmt.Errorf("simulated ExtractSourceChunks error")
	})

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "simulated ExtractSourceChunks error")
}

func TestValidateCoverage_ArchitectureMDError(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	docsDir := rootDir + "/docs"
	require.NoError(t, os.MkdirAll(docsDir, 0o700))
	require.NoError(t, os.WriteFile(docsDir+"/required-propagations.yaml", []byte(minimalManifestYAML()), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := ValidateCoverage(rootDir, rootedReadFile(rootDir))

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "ENG-HANDBOOK.md")
}

func TestValidateCoverage_ViolationsSortedByChunkAndFile(t *testing.T) {
	t.Parallel()

	manifestYAML := `required_propagations:
  - chunk_id: zzz-last
    source_file: docs/ENG-HANDBOOK.md
    required_targets:
      - .github/instructions/b.instructions.md
  - chunk_id: aaa-first
    source_file: docs/ENG-HANDBOOK.md
    required_targets:
      - .github/instructions/a.instructions.md
      - .github/instructions/b.instructions.md
`
	archContent := `<!-- @to-appendix as="aaa-first" appendixes=".github/instructions/a.md" -->
aaa
<!-- @/to-appendix -->
<!-- @to-appendix as="zzz-last" appendixes=".github/instructions/b.md" -->
zzz
<!-- @/to-appendix -->
`
	rootDir, readFile := buildValidateRoot(t, manifestYAML, archContent, map[string]string{})

	result, err := ValidateCoverage(rootDir, readFile)

	require.NoError(t, err)
	require.Len(t, result.Violations, 3)
	// First two should be aaa-first.
	require.Equal(t, "aaa-first", result.Violations[0].ChunkID)
	require.Equal(t, "aaa-first", result.Violations[1].ChunkID)
	// Last should be zzz-last.
	require.Equal(t, "zzz-last", result.Violations[2].ChunkID)
	// Verify secondary sort by File within same ChunkID (kills sort negation mutation).
	require.Contains(t, result.Violations[0].File, "a.instructions.md")
	require.Contains(t, result.Violations[1].File, "b.instructions.md")
}

// TestValidateCoverage_ViolationsSortedByHandbookOrder verifies that violations sort by the
// @to-appendix marker line order in ENG-HANDBOOK.md before chunk alphabetical order.
func TestValidateCoverage_ViolationsSortedByHandbookOrder(t *testing.T) {
	t.Parallel()

	manifestYAML := `required_propagations:
  - chunk_id: zzz-early
    source_file: docs/ENG-HANDBOOK.md
    required_targets:
      - .github/instructions/a.instructions.md
  - chunk_id: aaa-late
    source_file: docs/ENG-HANDBOOK.md
    required_targets:
      - .github/instructions/a.instructions.md
`

	archContent := `# Handbook
<!-- @to-appendix as="zzz-early" appendixes=".github/instructions/a.instructions.md" -->
early content
<!-- @/to-appendix -->

<!-- @to-appendix as="aaa-late" appendixes=".github/instructions/a.instructions.md" -->
late content
<!-- @/to-appendix -->
`

	rootDir, readFile := buildValidateRoot(t, manifestYAML, archContent, map[string]string{})

	result, err := ValidateCoverage(rootDir, readFile)

	require.NoError(t, err)
	require.Len(t, result.Violations, 2)
	// zzz-early must come BEFORE aaa-late because its marker appears first in the handbook.
	require.Equal(t, "zzz-early", result.Violations[0].ChunkID)
	require.Equal(t, "aaa-late", result.Violations[1].ChunkID)
}
