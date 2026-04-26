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

// --- ExtractSourceChunks (integration over temp dir) ---

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

	copilotContent := `<!-- @source from="docs/ENG-HANDBOOK.md" as="copilot-chunk" -->
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

	chunk := `<!-- @source from="docs/ENG-HANDBOOK.md" as="shared-chunk" -->`

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

	agentContent := "# Agent\n\n<!-- @source from=\"docs/ENG-HANDBOOK.md\" as=\"claude-chunk\" -->\nchunk content\n<!-- @/source -->\n"
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

// --- FormatCoverageValidationResults ---

func TestFormatCoverageValidationResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		result      *CoverageResult
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:        "clean result",
			result:      &CoverageResult{ManifestChunks: 40, ArchitectureChunks: 40},
			wantContain: []string{"Manifest chunks:      40", "Architecture chunks:  40", "All required @propagate chunks are covered"},
			wantAbsent:  []string{"ORPHANED CHUNKS", "MISSING @SOURCE BLOCKS"},
		},
		{
			name: "with violations",
			result: &CoverageResult{
				Violations:     []CoverageViolation{{ChunkID: "my-chunk", File: "instructions/a.md", Description: `@source block for chunk "my-chunk" not found in instructions/a.md`}},
				ManifestChunks: 1, ArchitectureChunks: 1,
			},
			wantContain: []string{"MISSING @SOURCE BLOCKS (1)", "my-chunk", "instructions/a.md", "Coverage validation FAILED"},
		},
		{
			name:        "with orphans",
			result:      &CoverageResult{OrphanedChunks: []string{"orphan-chunk"}, ManifestChunks: 0, ArchitectureChunks: 1},
			wantContain: []string{"ORPHANED CHUNKS (1)", "orphan-chunk", "Coverage validation FAILED"},
		},
		{
			name: "with both",
			result: &CoverageResult{
				Violations:     []CoverageViolation{{ChunkID: "missing-chunk", File: "a.md"}},
				OrphanedChunks: []string{"orphan-chunk"}, ManifestChunks: 1, ArchitectureChunks: 2,
			},
			wantContain: []string{"ORPHANED CHUNKS (1)", "MISSING @SOURCE BLOCKS (1)", "Coverage validation FAILED"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			report := FormatCoverageValidationResults(tc.result)

			for _, s := range tc.wantContain {
				require.Contains(t, report, s)
			}

			for _, s := range tc.wantAbsent {
				require.NotContains(t, report, s)
			}
		})
	}
}

// --- ValidateCoverageCommand ---

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

// --- validateCoverageWithRoot ---

func TestValidateCoverageWithRoot(t *testing.T) {
	t.Parallel()

	archWithExtra := minimalArchitectureContent() + `<!-- @propagate to=".github/instructions/x.md" as="extra-chunk" -->
`

	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name: "clean result",
			setup: func(t *testing.T) string {
				t.Helper()
				rootDir, _ := buildValidateRoot(t, minimalManifestYAML(), minimalArchitectureContent(),
					map[string]string{"01-01.terminology.instructions.md": sourceInstructionContent()})

				return rootDir
			},
			wantCode:   0,
			wantStdout: "All required @propagate chunks are covered",
		},
		{
			name: "manifest error",
			setup: func(t *testing.T) string {
				t.Helper()

				return t.TempDir()
			},
			wantCode:   1,
			wantStderr: "Error:",
		},
		{
			name: "violations",
			setup: func(t *testing.T) string {
				t.Helper()
				rootDir, _ := buildValidateRoot(t, minimalManifestYAML(), minimalArchitectureContent(),
					map[string]string{"01-01.terminology.instructions.md": rfcOnlySourceContent()})

				return rootDir
			},
			wantCode:   1,
			wantStdout: "MISSING @SOURCE BLOCKS",
		},
		{
			name: "orphans",
			setup: func(t *testing.T) string {
				t.Helper()
				rootDir, _ := buildValidateRoot(t, minimalManifestYAML(), archWithExtra,
					map[string]string{"01-01.terminology.instructions.md": sourceInstructionContent()})

				return rootDir
			},
			wantCode:   1,
			wantStdout: "ORPHANED CHUNKS",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := tc.setup(t)

			var stdout, stderr bytes.Buffer

			code := validateCoverageWithRoot(rootDir, &stdout, &stderr)

			require.Equal(t, tc.wantCode, code)

			if tc.wantStdout != "" {
				require.Contains(t, stdout.String(), tc.wantStdout)
			}

			if tc.wantStderr != "" {
				require.Contains(t, stderr.String(), tc.wantStderr)
			}
		})
	}
}
