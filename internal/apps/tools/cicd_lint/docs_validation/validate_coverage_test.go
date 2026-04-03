// Copyright (c) 2025 Justin Cranford

package docs_validation

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// rfcOnlySourceContent returns instruction file content with ONLY the rfc-2119-keywords @source block.
// Used in tests where emphasis-keywords is intentionally missing.
func rfcOnlySourceContent() string {
	return `<!-- @source from="docs/ARCHITECTURE.md" as="rfc-2119-keywords" -->
content
<!-- @/source -->
`
}

// minimalManifestYAML returns a valid YAML manifest with two entries.
func minimalManifestYAML() string {
	return `required_propagations:
  - chunk_id: rfc-2119-keywords
    source_file: docs/ARCHITECTURE.md
    required_targets:
      - .github/instructions/01-01.terminology.instructions.md
  - chunk_id: emphasis-keywords
    source_file: docs/ARCHITECTURE.md
    required_targets:
      - .github/instructions/01-01.terminology.instructions.md
`
}

// minimalArchitectureContent returns ARCHITECTURE.md content with two @propagate markers.
func minimalArchitectureContent() string {
	return `# Architecture
<!-- @propagate to=".github/instructions/01-01.terminology.instructions.md" as="rfc-2119-keywords" -->
some content
<!-- @/propagate -->
<!-- @propagate to=".github/instructions/01-01.terminology.instructions.md" as="emphasis-keywords" -->
more content
<!-- @/propagate -->
`
}

// sourceInstructionContent returns instruction file content with @source blocks.
func sourceInstructionContent() string {
	return `# Instruction File
<!-- @source from="docs/ARCHITECTURE.md" as="rfc-2119-keywords" -->
RFC keyword content
<!-- @/source -->
<!-- @source from="docs/ARCHITECTURE.md" as="emphasis-keywords" -->
Emphasis content
<!-- @/source -->
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

// -----------------------------------------------------------------------
// LoadPropagationsManifest
// -----------------------------------------------------------------------

func TestLoadPropagationsManifest_HappyPath(t *testing.T) {
	t.Parallel()

	readFile := makeFakeReadFile(map[string]string{
		cryptoutilSharedMagic.CICDRequiredPropagationsManifest: minimalManifestYAML(),
	})

	manifest, err := LoadPropagationsManifest(readFile)

	require.NoError(t, err)
	require.Len(t, manifest.RequiredPropagations, 2)
	require.Equal(t, "rfc-2119-keywords", manifest.RequiredPropagations[0].ChunkID)
	require.Equal(t, "docs/ARCHITECTURE.md", manifest.RequiredPropagations[0].SourceFile)
	require.Len(t, manifest.RequiredPropagations[0].RequiredTargets, 1)
}

func TestLoadPropagationsManifest_InvalidYAML(t *testing.T) {
	t.Parallel()

	readFile := makeFakeReadFile(map[string]string{
		cryptoutilSharedMagic.CICDRequiredPropagationsManifest: "!!! not valid: yaml: [",
	})

	manifest, err := LoadPropagationsManifest(readFile)

	require.Error(t, err)
	require.Nil(t, manifest)
	require.Contains(t, err.Error(), "failed to parse")
}

func TestLoadPropagationsManifest_FileNotFound(t *testing.T) {
	t.Parallel()

	readFile := makeFakeReadFile(map[string]string{})

	manifest, err := LoadPropagationsManifest(readFile)

	require.Error(t, err)
	require.Nil(t, manifest)
	require.Contains(t, err.Error(), "failed to read")
}

func TestLoadPropagationsManifest_EmptyManifest(t *testing.T) {
	t.Parallel()

	readFile := makeFakeReadFile(map[string]string{
		cryptoutilSharedMagic.CICDRequiredPropagationsManifest: "required_propagations: []\n",
	})

	manifest, err := LoadPropagationsManifest(readFile)

	require.NoError(t, err)
	require.Empty(t, manifest.RequiredPropagations)
}

// -----------------------------------------------------------------------
// ExtractPropagateChunks
// -----------------------------------------------------------------------

func TestExtractPropagateChunks_HappyPath(t *testing.T) {
	t.Parallel()

	readFile := makeFakeReadFile(map[string]string{
		"docs/ARCHITECTURE.md": minimalArchitectureContent(),
	})

	chunks, err := ExtractPropagateChunks(readFile)

	require.NoError(t, err)
	require.Equal(t, []string{"emphasis-keywords", "rfc-2119-keywords"}, chunks) // sorted
}

func TestExtractPropagateChunks_Deduplicated(t *testing.T) {
	t.Parallel()

	// Same chunk referenced twice (e.g., duplicate marker).
	content := `<!-- @propagate to=".github/instructions/a.md" as="my-chunk" -->
<!-- @propagate to=".github/instructions/b.md" as="my-chunk" -->
`
	readFile := makeFakeReadFile(map[string]string{
		"docs/ARCHITECTURE.md": content,
	})

	chunks, err := ExtractPropagateChunks(readFile)

	require.NoError(t, err)
	require.Equal(t, []string{"my-chunk"}, chunks)
}

func TestExtractPropagateChunks_FiltersInvalidChunkIDs(t *testing.T) {
	t.Parallel()

	// Grammar example line should be ignored.
	content := `@propagate-open  ::= '<!-- @propagate to="' PATH_LIST '" as="' CHUNK_ID '" -->'
<!-- @propagate to=".github/instructions/a.md" as="valid-chunk" -->
`
	readFile := makeFakeReadFile(map[string]string{
		"docs/ARCHITECTURE.md": content,
	})

	chunks, err := ExtractPropagateChunks(readFile)

	require.NoError(t, err)
	require.Equal(t, []string{"valid-chunk"}, chunks)
}

func TestExtractPropagateChunks_EmptyFile(t *testing.T) {
	t.Parallel()

	readFile := makeFakeReadFile(map[string]string{
		"docs/ARCHITECTURE.md": "",
	})

	chunks, err := ExtractPropagateChunks(readFile)

	require.NoError(t, err)
	require.Empty(t, chunks)
}

func TestExtractPropagateChunks_FileNotFound(t *testing.T) {
	t.Parallel()

	readFile := makeFakeReadFile(map[string]string{})

	chunks, err := ExtractPropagateChunks(readFile)

	require.Error(t, err)
	require.Nil(t, chunks)
	require.Contains(t, err.Error(), "failed to read docs/ARCHITECTURE.md")
}

// -----------------------------------------------------------------------
// extractSourceChunksFromContent
// -----------------------------------------------------------------------

func TestExtractSourceChunksFromContent_SingleMatch(t *testing.T) {
	t.Parallel()

	result := make(map[string][]string)
	content := `<!-- @source from="docs/ARCHITECTURE.md" as="my-chunk" -->
content
<!-- @/source -->
`
	extractSourceChunksFromContent("instructions/test.md", content, result)

	require.Len(t, result, 1)
	require.Equal(t, []string{"instructions/test.md"}, result["my-chunk"])
}

func TestExtractSourceChunksFromContent_MultipleMatches(t *testing.T) {
	t.Parallel()

	result := make(map[string][]string)
	content := `<!-- @source from="docs/ARCHITECTURE.md" as="chunk-a" -->
a
<!-- @/source -->
<!-- @source from="docs/ARCHITECTURE.md" as="chunk-b" -->
b
<!-- @/source -->
`
	extractSourceChunksFromContent("file.md", content, result)

	require.Len(t, result, 2)
	require.Equal(t, []string{"file.md"}, result["chunk-a"])
	require.Equal(t, []string{"file.md"}, result["chunk-b"])
}

func TestExtractSourceChunksFromContent_NoMatches(t *testing.T) {
	t.Parallel()

	result := make(map[string][]string)
	extractSourceChunksFromContent("file.md", "no @source blocks here", result)

	require.Empty(t, result)
}

func TestExtractSourceChunksFromContent_FiltersInvalidChunkIDs(t *testing.T) {
	t.Parallel()

	// Source block grammar example line should be filtered.
	result := make(map[string][]string)
	content := `@source-open     ::= '<!-- @source from="' PATH '" as="' CHUNK_ID '" -->'
<!-- @source from="docs/ARCHITECTURE.md" as="valid-chunk" -->
`
	extractSourceChunksFromContent("file.md", content, result)

	require.Len(t, result, 1)
	require.Contains(t, result, "valid-chunk")
	require.NotContains(t, result, "' CHUNK_ID '")
}

func TestExtractSourceChunksFromContent_AppendsToPriorEntries(t *testing.T) {
	t.Parallel()

	result := map[string][]string{
		"my-chunk": {"already-there.md"},
	}

	content := `<!-- @source from="docs/ARCHITECTURE.md" as="my-chunk" -->`
	extractSourceChunksFromContent("second.md", content, result)

	require.Len(t, result["my-chunk"], 2)
}

// -----------------------------------------------------------------------
// ExtractSourceChunks (integration over temp dir)
// -----------------------------------------------------------------------

func TestExtractSourceChunks_HappyPath(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	instrDir := rootDir + "/.github/instructions"
	require.NoError(t, os.MkdirAll(instrDir, 0o700))
	require.NoError(t, os.WriteFile(instrDir+"/01.instructions.md", []byte(sourceInstructionContent()), cryptoutilSharedMagic.FilePermissionsDefault))

	readFile := rootedReadFile(rootDir)

	result, err := ExtractSourceChunks(rootDir, readFile)

	require.NoError(t, err)
	require.Contains(t, result, "rfc-2119-keywords")
	require.Contains(t, result, "emphasis-keywords")
}

func TestExtractSourceChunks_CopilotInstructionsFile(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	copilotDir := rootDir + "/.github"
	require.NoError(t, os.MkdirAll(copilotDir, 0o700))

	copilotContent := `<!-- @source from="docs/ARCHITECTURE.md" as="copilot-chunk" -->
content
<!-- @/source -->
`
	require.NoError(t, os.WriteFile(copilotDir+"/copilot-instructions.md", []byte(copilotContent), cryptoutilSharedMagic.FilePermissionsDefault))

	readFile := rootedReadFile(rootDir)

	result, err := ExtractSourceChunks(rootDir, readFile)

	require.NoError(t, err)
	require.Contains(t, result, "copilot-chunk")
}

func TestExtractSourceChunks_SkipsNonMatchingFiles(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	instrDir := rootDir + "/.github/instructions"
	require.NoError(t, os.MkdirAll(instrDir, 0o700))

	// This file does not match *.instructions.md.
	require.NoError(t, os.WriteFile(instrDir+"/README.md", []byte(sourceInstructionContent()), cryptoutilSharedMagic.FilePermissionsDefault))

	readFile := rootedReadFile(rootDir)

	result, err := ExtractSourceChunks(rootDir, readFile)

	require.NoError(t, err)
	require.Empty(t, result) // copilot-instructions.md also absent
}

func TestExtractSourceChunks_NonExistentDirsSkipped(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	readFile := rootedReadFile(rootDir)

	// No .github/instructions or .github/agents dirs created.
	result, err := ExtractSourceChunks(rootDir, readFile)

	require.NoError(t, err)
	require.Empty(t, result)
}

func TestExtractSourceChunks_SortsFileLists(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	instrDir := rootDir + "/.github/instructions"
	require.NoError(t, os.MkdirAll(instrDir, 0o700))

	chunk := `<!-- @source from="docs/ARCHITECTURE.md" as="shared-chunk" -->`

	require.NoError(t, os.WriteFile(instrDir+"/z.instructions.md", []byte(chunk), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(instrDir+"/a.instructions.md", []byte(chunk), cryptoutilSharedMagic.FilePermissionsDefault))

	readFile := rootedReadFile(rootDir)

	result, err := ExtractSourceChunks(rootDir, readFile)

	require.NoError(t, err)
	// Verify sorted order.
	files := result["shared-chunk"]
	require.Len(t, files, 2)
	require.True(t, strings.Compare(files[0], files[1]) < 0)
}

func TestExtractSourceChunks_ClaudeAgentsDir(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	claudeDir := rootDir + "/.claude/agents"
	require.NoError(t, os.MkdirAll(claudeDir, 0o700))

	agentContent := "# Agent\n\n<!-- @source from=\"docs/ARCHITECTURE.md\" as=\"claude-chunk\" -->\nchunk content\n<!-- @/source -->\n"
	require.NoError(t, os.WriteFile(claudeDir+"/myagent.md", []byte(agentContent), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := ExtractSourceChunks(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Contains(t, result, "claude-chunk")
	require.Contains(t, result["claude-chunk"], ".claude/agents/myagent.md")
}

func TestExtractSourceChunks_ReadFileFails(t *testing.T) {
	t.Parallel()

	// Real dir with a matching instruction file, but readFile returns an error for that file.
	rootDir := t.TempDir()
	instrDir := rootDir + "/.github/instructions"
	require.NoError(t, os.MkdirAll(instrDir, 0o700))
	require.NoError(t, os.WriteFile(instrDir+"/01.instructions.md", []byte("content"), cryptoutilSharedMagic.FilePermissionsDefault))

	// readFile fails for any non-manifest/non-arch path.
	failingReadFile := func(path string) ([]byte, error) {
		// Allow copilot-instructions.md to fail silently (err == nil not required).
		return nil, fmt.Errorf("read error: %s", path)
	}

	result, err := ExtractSourceChunks(rootDir, failingReadFile)

	// ExtractSourceChunks silently skips read errors.
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestExtractSourceChunks_DirectoryEntrySkipped(t *testing.T) {
	t.Parallel()

	// Create a subdirectory inside .github/instructions — it should be skipped.
	rootDir := t.TempDir()
	instrDir := rootDir + "/.github/instructions"
	subDir := instrDir + "/subdir"
	require.NoError(t, os.MkdirAll(subDir, 0o700))

	// Also add a valid file so we can confirm the dir entry was skipped without error.
	require.NoError(t, os.WriteFile(instrDir+"/01.instructions.md", []byte(sourceInstructionContent()), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := ExtractSourceChunks(rootDir, rootedReadFile(rootDir))

	require.NoError(t, err)
	require.Contains(t, result, "rfc-2119-keywords")
}

// -----------------------------------------------------------------------
// ValidateCoverage
// -----------------------------------------------------------------------

func buildValidateRoot(t *testing.T, manifestContent, archContent string, instructionFiles map[string]string) (string, func(string) ([]byte, error)) {
	t.Helper()

	rootDir := t.TempDir()

	// Write manifest.
	docsDir := rootDir + "/docs"
	require.NoError(t, os.MkdirAll(docsDir, 0o700))
	require.NoError(t, os.WriteFile(docsDir+"/required-propagations.yaml", []byte(manifestContent), cryptoutilSharedMagic.FilePermissionsDefault))

	// Write ARCHITECTURE.md.
	require.NoError(t, os.WriteFile(docsDir+"/ARCHITECTURE.md", []byte(archContent), cryptoutilSharedMagic.FilePermissionsDefault))

	// Write instruction files.
	instrDir := rootDir + "/.github/instructions"
	require.NoError(t, os.MkdirAll(instrDir, 0o700))

	for name, content := range instructionFiles {
		require.NoError(t, os.WriteFile(instrDir+"/"+name, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))
	}

	return rootDir, rootedReadFile(rootDir)
}

func TestValidateCoverage_AllChunksCovered(t *testing.T) {
	t.Parallel()

	rootDir, readFile := buildValidateRoot(t,
		minimalManifestYAML(),
		minimalArchitectureContent(),
		map[string]string{
			"01-01.terminology.instructions.md": sourceInstructionContent(),
		},
	)

	result, err := ValidateCoverage(rootDir, readFile)

	require.NoError(t, err)
	require.Empty(t, result.Violations)
	require.Empty(t, result.OrphanedChunks)
	require.Equal(t, 2, result.ManifestChunks)
	require.Equal(t, 2, result.ArchitectureChunks)
}

func TestValidateCoverage_MissingSourceBlock(t *testing.T) {
	t.Parallel()

	// Instruction file only has rfc-2119-keywords, not emphasis-keywords.
	rootDir, readFile := buildValidateRoot(t,
		minimalManifestYAML(),
		minimalArchitectureContent(),
		map[string]string{
			"01-01.terminology.instructions.md": rfcOnlySourceContent(),
		},
	)

	result, err := ValidateCoverage(rootDir, readFile)

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Equal(t, "emphasis-keywords", result.Violations[0].ChunkID)
	require.Contains(t, result.Violations[0].Description, "emphasis-keywords")
}

func TestValidateCoverage_OrphanedChunk(t *testing.T) {
	t.Parallel()

	// ARCHITECTURE.md has "extra-chunk" but manifest only lists rfc-2119-keywords + emphasis-keywords.
	archContent := minimalArchitectureContent() + `<!-- @propagate to=".github/instructions/x.md" as="extra-chunk" -->
`
	rootDir, readFile := buildValidateRoot(t,
		minimalManifestYAML(),
		archContent,
		map[string]string{
			"01-01.terminology.instructions.md": sourceInstructionContent(),
		},
	)

	result, err := ValidateCoverage(rootDir, readFile)

	require.NoError(t, err)
	require.Len(t, result.OrphanedChunks, 1)
	require.Equal(t, "extra-chunk", result.OrphanedChunks[0])
}

func TestValidateCoverage_ViolationsAndOrphansTogether(t *testing.T) {
	t.Parallel()

	// Instruction file only has rfc-2119-keywords (emphasis-keywords missing → violation),
	// and ARCHITECTURE.md has extra-chunk (→ orphan).
	archContent := minimalArchitectureContent() + `<!-- @propagate to=".github/instructions/x.md" as="extra-chunk" -->
`
	rootDir, readFile := buildValidateRoot(t,
		minimalManifestYAML(),
		archContent,
		map[string]string{
			"01-01.terminology.instructions.md": rfcOnlySourceContent(),
		},
	)

	result, err := ValidateCoverage(rootDir, readFile)

	require.NoError(t, err)
	require.Len(t, result.Violations, 1)
	require.Len(t, result.OrphanedChunks, 1)
}

func TestValidateCoverage_ManifestLoadError(t *testing.T) {
	t.Parallel()

	// rootDir has no ARCHITECTURE.md or manifest → manifest error first.
	rootDir := t.TempDir()
	readFile := rootedReadFile(rootDir)

	result, err := ValidateCoverage(rootDir, readFile)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to read")
}

func TestValidateCoverage_ExtractSourceChunksError(t *testing.T) {
	t.Parallel()

	rootDir, readFile := buildValidateRoot(t,
		minimalManifestYAML(),
		minimalArchitectureContent(),
		map[string]string{},
	)

	result, err := validateCoverage(rootDir, readFile, func(_ string, _ func(string) ([]byte, error)) (map[string][]string, error) {
		return nil, fmt.Errorf("simulated ExtractSourceChunks error")
	})

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "simulated ExtractSourceChunks error")
}

func TestValidateCoverage_ArchitectureMDError(t *testing.T) {
	t.Parallel()

	// Manifest exists but ARCHITECTURE.md absent.
	rootDir := t.TempDir()
	docsDir := rootDir + "/docs"
	require.NoError(t, os.MkdirAll(docsDir, 0o700))
	require.NoError(t, os.WriteFile(docsDir+"/required-propagations.yaml", []byte(minimalManifestYAML()), cryptoutilSharedMagic.FilePermissionsDefault))

	result, err := ValidateCoverage(rootDir, rootedReadFile(rootDir))

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "ARCHITECTURE.md")
}

func TestValidateCoverage_ViolationsSortedByChunkAndFile(t *testing.T) {
	t.Parallel()

	// Manifest has 3 entries across 2 files, none present in instruction files.
	manifestYAML := `required_propagations:
  - chunk_id: zzz-last
    source_file: docs/ARCHITECTURE.md
    required_targets:
      - .github/instructions/b.instructions.md
  - chunk_id: aaa-first
    source_file: docs/ARCHITECTURE.md
    required_targets:
      - .github/instructions/a.instructions.md
      - .github/instructions/b.instructions.md
`
	archContent := `<!-- @propagate to=".github/instructions/a.md" as="aaa-first" -->
<!-- @propagate to=".github/instructions/b.md" as="zzz-last" -->
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

// -----------------------------------------------------------------------
// FormatCoverageValidationResults
// -----------------------------------------------------------------------

func TestFormatCoverageValidationResults_CleanResult(t *testing.T) {
	t.Parallel()

	result := &CoverageResult{
		ManifestChunks:     40,
		ArchitectureChunks: 40,
	}

	report := FormatCoverageValidationResults(result)

	require.Contains(t, report, "Manifest chunks:      40")
	require.Contains(t, report, "Architecture chunks:  40")
	require.Contains(t, report, "All required @propagate chunks are covered")
	// Verify empty sections are NOT printed (kills len()>0 boundary mutations).
	require.NotContains(t, report, "ORPHANED CHUNKS")
	require.NotContains(t, report, "MISSING @SOURCE BLOCKS")
}

func TestFormatCoverageValidationResults_WithViolations(t *testing.T) {
	t.Parallel()

	result := &CoverageResult{
		Violations: []CoverageViolation{
			{ChunkID: "my-chunk", File: "instructions/a.md", Description: `@source block for chunk "my-chunk" not found in instructions/a.md`},
		},
		ManifestChunks:     1,
		ArchitectureChunks: 1,
	}

	report := FormatCoverageValidationResults(result)

	require.Contains(t, report, "MISSING @SOURCE BLOCKS (1)")
	require.Contains(t, report, "my-chunk")
	require.Contains(t, report, "instructions/a.md")
	require.Contains(t, report, "Coverage validation FAILED")
}

func TestFormatCoverageValidationResults_WithOrphans(t *testing.T) {
	t.Parallel()

	result := &CoverageResult{
		OrphanedChunks:     []string{"orphan-chunk"},
		ManifestChunks:     0,
		ArchitectureChunks: 1,
	}

	report := FormatCoverageValidationResults(result)

	require.Contains(t, report, "ORPHANED CHUNKS (1)")
	require.Contains(t, report, "orphan-chunk")
	require.Contains(t, report, "Coverage validation FAILED")
}

func TestFormatCoverageValidationResults_WithBoth(t *testing.T) {
	t.Parallel()

	result := &CoverageResult{
		Violations:         []CoverageViolation{{ChunkID: "missing-chunk", File: "a.md"}},
		OrphanedChunks:     []string{"orphan-chunk"},
		ManifestChunks:     1,
		ArchitectureChunks: 2,
	}

	report := FormatCoverageValidationResults(result)

	require.Contains(t, report, "ORPHANED CHUNKS (1)")
	require.Contains(t, report, "MISSING @SOURCE BLOCKS (1)")
	require.Contains(t, report, "Coverage validation FAILED")
}

// -----------------------------------------------------------------------
// ValidateCoverageCommand
// -----------------------------------------------------------------------

func TestValidateCoverageCommand_RootError(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := validateCoverageCommand(&stdout, &stderr, func() (string, error) {
		return "", fmt.Errorf("simulated root error")
	})

	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "simulated root error")
}

func TestValidateCoverageCommand_CleanProject(t *testing.T) {
	t.Parallel()

	rootDir, _ := buildValidateRoot(t,
		minimalManifestYAML(),
		minimalArchitectureContent(),
		map[string]string{
			"01-01.terminology.instructions.md": sourceInstructionContent(),
		},
	)

	var stdout, stderr bytes.Buffer

	code := validateCoverageCommand(&stdout, &stderr, func() (string, error) {
		return rootDir, nil
	})

	require.Equal(t, 0, code)
	require.Contains(t, stdout.String(), "All required @propagate chunks are covered")
}

// -----------------------------------------------------------------------
// validateCoverageWithRoot
// -----------------------------------------------------------------------

func TestValidateCoverageWithRoot_CleanResult(t *testing.T) {
	t.Parallel()

	rootDir, _ := buildValidateRoot(t,
		minimalManifestYAML(),
		minimalArchitectureContent(),
		map[string]string{
			"01-01.terminology.instructions.md": sourceInstructionContent(),
		},
	)

	var stdout, stderr bytes.Buffer

	code := validateCoverageWithRoot(rootDir, &stdout, &stderr)

	require.Equal(t, 0, code)
	require.Contains(t, stdout.String(), "All required @propagate chunks are covered")
	require.Empty(t, stderr.String())
}

func TestValidateCoverageWithRoot_ManifestError(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir() // no manifest file

	var stdout, stderr bytes.Buffer

	code := validateCoverageWithRoot(rootDir, &stdout, &stderr)

	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "Error:")
}

func TestValidateCoverageWithRoot_Violations(t *testing.T) {
	t.Parallel()

	rootDir, _ := buildValidateRoot(t,
		minimalManifestYAML(),
		minimalArchitectureContent(),
		map[string]string{
			"01-01.terminology.instructions.md": rfcOnlySourceContent(),
		},
	)

	var stdout, stderr bytes.Buffer

	code := validateCoverageWithRoot(rootDir, &stdout, &stderr)

	require.Equal(t, 1, code)
	require.Contains(t, stdout.String(), "MISSING @SOURCE BLOCKS")
}

func TestValidateCoverageWithRoot_Orphans(t *testing.T) {
	t.Parallel()

	// ARCHITECTURE.md has extra-chunk not in manifest.
	archContent := minimalArchitectureContent() + `<!-- @propagate to=".github/instructions/x.md" as="extra-chunk" -->
`
	rootDir, _ := buildValidateRoot(t,
		minimalManifestYAML(),
		archContent,
		map[string]string{
			"01-01.terminology.instructions.md": sourceInstructionContent(),
		},
	)

	var stdout, stderr bytes.Buffer

	code := validateCoverageWithRoot(rootDir, &stdout, &stderr)

	require.Equal(t, 1, code)
	require.Contains(t, stdout.String(), "ORPHANED CHUNKS")
}
