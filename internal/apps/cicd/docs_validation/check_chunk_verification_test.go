// Copyright (c) 2025 Justin Cranford

package docs_validation

import (
	"bytes"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testFileA = "a.md"
	testFileB = "b.md"
	testFileC = "c.md"
)

func TestChunkMappings(t *testing.T) {
	t.Parallel()

	mappings := chunkMappings()
	require.NotEmpty(t, mappings, "chunkMappings must not be empty")

	for _, m := range mappings {
		require.NotEmpty(t, m.ArchSection, "ArchSection must not be empty")
		require.NotEmpty(t, m.Description, "Description must not be empty")
		require.NotEmpty(t, m.DestFile, "DestFile must not be empty")
		require.NotEmpty(t, m.MarkerText, "MarkerText must not be empty")
	}
}

func TestVerifyChunks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mappings   []ChunkMapping
		readFile   func(string) ([]byte, error)
		wantPass   bool
		wantFound  []bool
		wantErrors []bool
	}{
		{
			name: "all chunks found",
			mappings: []ChunkMapping{
				{ArchSection: "1.1", Description: "Test A", DestFile: testFileA, MarkerText: "marker-a"},
				{ArchSection: "2.2", Description: "Test B", DestFile: testFileB, MarkerText: "marker-b"},
			},
			readFile: func(path string) ([]byte, error) {
				switch path {
				case testFileA:
					return []byte("content with marker-a here"), nil
				case testFileB:
					return []byte("content with marker-b here"), nil
				default:
					return nil, fmt.Errorf("file not found: %s", path)
				}
			},
			wantPass:   true,
			wantFound:  []bool{true, true},
			wantErrors: []bool{false, false},
		},
		{
			name: "missing marker",
			mappings: []ChunkMapping{
				{ArchSection: "1.1", Description: "Test A", DestFile: testFileA, MarkerText: "marker-a"},
				{ArchSection: "2.2", Description: "Test B", DestFile: testFileB, MarkerText: "marker-b"},
			},
			readFile: func(path string) ([]byte, error) {
				switch path {
				case testFileA:
					return []byte("content with marker-a here"), nil
				case testFileB:
					return []byte("content without the expected text"), nil
				default:
					return nil, fmt.Errorf("file not found: %s", path)
				}
			},
			wantPass:   false,
			wantFound:  []bool{true, false},
			wantErrors: []bool{false, false},
		},
		{
			name: "file read error",
			mappings: []ChunkMapping{
				{ArchSection: "1.1", Description: "Test A", DestFile: "missing.md", MarkerText: "marker-a"},
			},
			readFile: func(_ string) ([]byte, error) {
				return nil, fmt.Errorf("permission denied")
			},
			wantPass:   false,
			wantFound:  []bool{false},
			wantErrors: []bool{true},
		},
		{
			name:     "empty mappings",
			mappings: []ChunkMapping{},
			readFile: func(_ string) ([]byte, error) {
				return nil, fmt.Errorf("should not be called")
			},
			wantPass:   true,
			wantFound:  []bool{},
			wantErrors: []bool{},
		},
		{
			name: "all markers missing",
			mappings: []ChunkMapping{
				{ArchSection: "1.1", Description: "Test A", DestFile: testFileA, MarkerText: "missing-a"},
				{ArchSection: "2.2", Description: "Test B", DestFile: testFileB, MarkerText: "missing-b"},
			},
			readFile: func(_ string) ([]byte, error) {
				return []byte("empty content"), nil
			},
			wantPass:   false,
			wantFound:  []bool{false, false},
			wantErrors: []bool{false, false},
		},
		{
			name: "mixed errors and missing",
			mappings: []ChunkMapping{
				{ArchSection: "1.1", Description: "Found", DestFile: testFileA, MarkerText: "exists"},
				{ArchSection: "2.2", Description: "Error", DestFile: "bad.md", MarkerText: "x"},
				{ArchSection: "3.3", Description: "Missing", DestFile: testFileC, MarkerText: "absent"},
			},
			readFile: func(path string) ([]byte, error) {
				switch path {
				case testFileA:
					return []byte("content with exists here"), nil
				case "bad.md":
					return nil, fmt.Errorf("disk error")
				case testFileC:
					return []byte("no match"), nil
				default:
					return nil, fmt.Errorf("unknown: %s", path)
				}
			},
			wantPass:   false,
			wantFound:  []bool{true, false, false},
			wantErrors: []bool{false, true, false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			results, allPassed := VerifyChunks(tc.mappings, tc.readFile)

			require.Equal(t, tc.wantPass, allPassed, "allPassed mismatch")
			require.Len(t, results, len(tc.wantFound), "result count mismatch")

			for i, r := range results {
				require.Equal(t, tc.wantFound[i], r.Found, "Found mismatch at index %d", i)

				if tc.wantErrors[i] {
					require.Error(t, r.Error, "expected error at index %d", i)
				} else {
					require.NoError(t, r.Error, "unexpected error at index %d", i)
				}
			}
		})
	}
}

func TestFormatVerificationResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		results   []ChunkVerificationResult
		allPassed bool
		wantParts []string
	}{
		{
			name: "all pass",
			results: []ChunkVerificationResult{
				{Mapping: ChunkMapping{ArchSection: "1.1", Description: "A"}, Found: true},
				{Mapping: ChunkMapping{ArchSection: "2.2", Description: "B"}, Found: true},
			},
			allPassed: true,
			wantParts: []string{"PASS [1.1] A", "PASS [2.2] B", "2 PASS, 0 FAIL", "verified successfully"},
		},
		{
			name: "some fail - missing",
			results: []ChunkVerificationResult{
				{Mapping: ChunkMapping{ArchSection: "1.1", Description: "A"}, Found: true},
				{Mapping: ChunkMapping{ArchSection: "2.2", Description: "B", DestFile: testFileB, MarkerText: "marker"}, Found: false},
			},
			allPassed: false,
			wantParts: []string{"PASS [1.1] A", "FAIL [2.2] B", "Missing marker", "1 PASS, 1 FAIL", cryptoutilSharedMagic.TaskFailed},
		},
		{
			name: "error result",
			results: []ChunkVerificationResult{
				{Mapping: ChunkMapping{ArchSection: "1.1", Description: "A"}, Found: false, Error: fmt.Errorf("read error")},
			},
			allPassed: false,
			wantParts: []string{"FAIL [1.1] A", "read error", "0 PASS, 1 FAIL", cryptoutilSharedMagic.TaskFailed},
		},
		{
			name:      "empty results",
			results:   []ChunkVerificationResult{},
			allPassed: true,
			wantParts: []string{"0 PASS, 0 FAIL", "verified successfully"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatVerificationResults(tc.results, tc.allPassed)

			for _, part := range tc.wantParts {
				require.Contains(t, output, part, "missing expected part: %s", part)
			}
		})
	}
}

func TestCheckChunkVerification_Integration(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := CheckChunkVerification(&stdout, &stderr)

	output := stdout.String()
	require.Contains(t, output, "Chunk Verification Report")
	require.Contains(t, output, "Summary")
	require.Equal(t, 0, exitCode, "all chunks should pass in real project: %s", output)
}

func TestCheckChunkVerification_AllMappingsValid(t *testing.T) {
	t.Parallel()

	rootDir, err := findProjectRoot()
	require.NoError(t, err, "must find project root")

	mappings := chunkMappings()
	results, allPassed := VerifyChunks(mappings, rootedReadFile(rootDir))

	require.True(t, allPassed, "all mappings must pass against real project files")

	for _, r := range results {
		t.Run(r.Mapping.ArchSection+"_"+r.Mapping.Description, func(t *testing.T) {
			t.Parallel()

			require.True(t, r.Found, "marker %q not found in %s", r.Mapping.MarkerText, r.Mapping.DestFile)
			require.NoError(t, r.Error)
		})
	}
}

func TestCheckChunkVerification_MissingChunkDetection(t *testing.T) {
	t.Parallel()

	mappings := []ChunkMapping{
		{ArchSection: "99.99", Description: "Nonexistent", DestFile: "fake.md", MarkerText: "impossible"},
	}

	results, allPassed := VerifyChunks(mappings, func(_ string) ([]byte, error) {
		return []byte("no match here"), nil
	})

	require.False(t, allPassed)
	require.Len(t, results, 1)
	require.False(t, results[0].Found)
	require.NoError(t, results[0].Error)
}

func TestFindProjectRoot(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	require.NoError(t, err)
	require.NotEmpty(t, root)
	require.FileExists(t, root+"/go.mod")
}

func TestCheckChunkVerificationWithRoot_Failure(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	exitCode := checkChunkVerificationWithRoot(t.TempDir(), &stdout)

	require.Equal(t, 1, exitCode, "should fail when no instruction files present")
	require.Contains(t, stdout.String(), cryptoutilSharedMagic.TestStatusFail)
	require.Contains(t, stdout.String(), cryptoutilSharedMagic.TaskFailed)
}

func TestCheckChunkVerificationWithRoot_Success(t *testing.T) {
	t.Parallel()

	rootDir, err := findProjectRoot()
	require.NoError(t, err)

	var stdout bytes.Buffer

	exitCode := checkChunkVerificationWithRoot(rootDir, &stdout)

	require.Equal(t, 0, exitCode, "should pass when using real project root: %s", stdout.String())
	require.Contains(t, stdout.String(), "verified successfully")
}

func TestRootedReadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rootDir string
		path    string
		wantErr bool
	}{
		{
			name:    "nonexistent file",
			rootDir: t.TempDir(),
			path:    "test.txt",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			readFn := rootedReadFile(tc.rootDir)
			_, err := readFn(tc.path)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
