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

func TestValidateChunks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		archContent    string
		files          map[string]string // additional files.
		wantErr        bool
		wantMatched    int
		wantMismatched int
		wantMissing    int
		wantFileErrors int
	}{
		{
			name: "all match",
			archContent: join(
				"# Arch",
				`<!-- @propagate to="target.md" as="chunk-a" -->`,
				"Line alpha",
				"<!-- @/propagate -->",
			),
			files: map[string]string{
				"target.md": join(
					"# Target",
					`<!-- @source from="docs/ENG-HANDBOOK.md" as="chunk-a" -->`,
					"Line alpha",
					"<!-- @/source -->",
				),
			},
			wantMatched: 1,
		},
		{
			name: "mismatch",
			archContent: join(
				`<!-- @propagate to="target.md" as="chunk-b" -->`,
				"New content",
				"<!-- @/propagate -->",
			),
			files: map[string]string{
				"target.md": join(
					`<!-- @source from="docs/ENG-HANDBOOK.md" as="chunk-b" -->`,
					"Old content",
					"<!-- @/source -->",
				),
			},
			wantMismatched: 1,
		},
		{
			name: "missing source block",
			archContent: join(
				`<!-- @propagate to="target.md" as="chunk-c" -->`,
				"Content here",
				"<!-- @/propagate -->",
			),
			files:       map[string]string{"target.md": "# Just a heading\n"},
			wantMissing: 1,
		},
		{
			name: "file not found",
			archContent: join(
				`<!-- @propagate to="nonexistent.md" as="chunk-d" -->`,
				"Some content",
				"<!-- @/propagate -->",
			),
			wantFileErrors: 1,
		},
		{
			name:    "architecture missing",
			wantErr: true,
		},
		{
			name: "multiple blocks same file",
			archContent: join(
				`<!-- @propagate to="multi.md" as="first" -->`,
				"First block",
				"<!-- @/propagate -->",
				"",
				`<!-- @propagate to="multi.md" as="second" -->`,
				"Second block",
				"<!-- @/propagate -->",
			),
			files: map[string]string{
				"multi.md": join(
					`<!-- @source from="docs/ENG-HANDBOOK.md" as="first" -->`,
					"First block",
					"<!-- @/source -->",
					`<!-- @source from="docs/ENG-HANDBOOK.md" as="second" -->`,
					"Second block",
					"<!-- @/source -->",
				),
			},
			wantMatched: 2,
		},
		{
			name: "duplicate file not found counted per block",
			archContent: join(
				`<!-- @propagate to="gone.md" as="aaa" -->`,
				"A content",
				"<!-- @/propagate -->",
				`<!-- @propagate to="gone.md" as="bbb" -->`,
				"B content",
				"<!-- @/propagate -->",
			),
			wantFileErrors: 2,
		},
		{
			name: "multi-target matches both files",
			archContent: join(
				`<!-- @propagate to="first.md, second.md" as="multi" -->`,
				"Multi content",
				"<!-- @/propagate -->",
			),
			files: map[string]string{
				"first.md": join(
					`<!-- @source from="docs/ENG-HANDBOOK.md" as="multi" -->`,
					"Multi content",
					"<!-- @/source -->",
				),
				"second.md": join(
					`<!-- @source from="docs/ENG-HANDBOOK.md" as="multi" -->`,
					"Multi content",
					"<!-- @/source -->",
				),
			},
			wantMatched: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			files := make(map[string]string)
			for k, v := range tc.files {
				files[k] = v
			}

			if tc.archContent != "" {
				files["docs/ENG-HANDBOOK.md"] = tc.archContent
			}

			readFile := func(path string) ([]byte, error) {
				c, ok := files[path]
				if !ok {
					return nil, fmt.Errorf("file not found: %s", path)
				}

				return []byte(c), nil
			}

			result, err := ValidateChunks(".", readFile)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantMatched, result.Matched)
			require.Equal(t, tc.wantMismatched, result.Mismatched)
			require.Equal(t, tc.wantMissing, result.Missing)
			require.Equal(t, tc.wantFileErrors, result.FileErrors)
		})
	}
}

func TestFormatChunkResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		result       *ChunkValidationResult
		wantContains []string
		wantExcludes []string
	}{
		{
			name: "all match",
			result: &ChunkValidationResult{
				Results: []ChunkResult{{ChunkID: "a", Status: ChunkStatusMatch}, {ChunkID: "b", Status: ChunkStatusMatch}},
				Matched: 2,
			},
			wantContains: []string{"2 chunks, 2 matched, 0 mismatched", "All propagated chunks are in sync."},
			wantExcludes: []string{"STALE", cryptoutilSharedMagic.TestStatusFail, "FILE NOT FOUND", "MISSING @source BLOCKS", "CONTENT MISMATCHES"},
		},
		{
			name: "with mismatch",
			result: &ChunkValidationResult{
				Results: []ChunkResult{{
					ChunkID: "stale-chunk", PropagateBlock: PropagateBlock{TargetFile: "target.md", LineNumber: 3},
					SourceBlock: &SourceBlock{LineNumber: 4}, Status: ChunkStatusMismatch,
				}},
				Mismatched: 1,
			},
			wantContains: []string{"CONTENT MISMATCHES (1)", "STALE [stale-chunk]", cryptoutilSharedMagic.TaskFailed},
		},
		{
			name: "with missing",
			result: &ChunkValidationResult{
				Results: []ChunkResult{{
					ChunkID: "absent", PropagateBlock: PropagateBlock{TargetFile: "target.md", LineNumber: 2},
					Status: ChunkStatusMissing,
				}},
				Missing: 1,
			},
			wantContains: []string{"MISSING @source BLOCKS (1)", cryptoutilSharedMagic.TestStatusFail + " [absent]"},
		},
		{
			name: "with file not found",
			result: &ChunkValidationResult{
				Results: []ChunkResult{{
					ChunkID: "gone", PropagateBlock: PropagateBlock{TargetFile: "missing.md", LineNumber: 3},
					Status: ChunkStatusFileNotFound,
				}},
				FileErrors: 1,
			},
			wantContains: []string{"FILE NOT FOUND (1)", cryptoutilSharedMagic.TestStatusFail + " [gone]"},
		},
		{
			name: "mixed issues",
			result: &ChunkValidationResult{
				Results: []ChunkResult{
					{ChunkID: "ok", Status: ChunkStatusMatch},
					{ChunkID: "stale", PropagateBlock: PropagateBlock{TargetFile: "t.md", LineNumber: 1}, SourceBlock: &SourceBlock{LineNumber: 2}, Status: ChunkStatusMismatch},
					{ChunkID: "absent", PropagateBlock: PropagateBlock{TargetFile: "t.md", LineNumber: 3}, Status: ChunkStatusMissing},
					{ChunkID: "gone", PropagateBlock: PropagateBlock{TargetFile: "x.md", LineNumber: 4}, Status: ChunkStatusFileNotFound},
				},
				Matched: 1, Mismatched: 1, Missing: 1, FileErrors: 1,
			},
			wantContains: []string{"4 chunks", "1 matched", "1 mismatched", "1 missing", "1 file errors", cryptoutilSharedMagic.TaskFailed},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatChunkResults(tc.result)

			for _, s := range tc.wantContains {
				require.Contains(t, output, s)
			}

			for _, s := range tc.wantExcludes {
				require.NotContains(t, output, s)
			}
		})
	}
}

func TestValidateChunksCommand_Integration(t *testing.T) {
	t.Parallel()

	rootDir := findChunksProjectRoot(t)

	var stdout, stderr bytes.Buffer

	exitCode := validateChunksWithRoot(rootDir, &stdout, &stderr)

	require.Equal(t, 0, exitCode, "validate-chunks should pass on real project: stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "chunks")
	require.Contains(t, stdout.String(), "All propagated chunks are in sync.")
	require.Empty(t, stderr.String())
}

func TestValidateChunksWithRoot_BadRoot(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := validateChunksWithRoot("/nonexistent/path", &stdout, &stderr)

	require.Equal(t, 1, exitCode)
	require.NotEmpty(t, stderr.String())
}

func TestValidateChunksCommand_FindRootError(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := validateChunksCommand(&stdout, &stderr, func() (string, error) {
		return "", fmt.Errorf("injected root error")
	})
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "injected root error")
}

func TestValidateChunksCommand_Success(t *testing.T) {
	t.Parallel()

	root := findChunksProjectRoot(t)

	var stdout, stderr bytes.Buffer

	exitCode := validateChunksCommand(&stdout, &stderr, func() (string, error) {
		return root, nil
	})
	require.Equal(t, 0, exitCode, "validate-chunks should pass on real project: stderr=%s", stderr.String())
}

// join is a test helper to join lines with newlines.
func join(lines ...string) string {
	return strings.Join(lines, "\n")
}

// findChunksProjectRoot navigates upward from CWD to find go.mod.
func findChunksProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod)")
		}

		dir = parent
	}
}
