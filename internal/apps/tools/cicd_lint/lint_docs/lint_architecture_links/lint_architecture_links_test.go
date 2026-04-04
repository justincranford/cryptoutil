// Copyright (c) 2025 Justin Cranford

package lint_architecture_links

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

const (
	testArchPath            = "docs/ARCHITECTURE.md"
	testInstructionFileName = "01-01.test.instructions.md"
	testAgentFileName       = "my-agent.agent.md"

	testArchContentSimple  = "# Executive Summary\n"
	testArchContentQuality = "# Executive Summary\n## 2.5 Quality Strategy\n"
)

// findTestProjectRoot walks up from the current directory to find go.mod.
func findTestProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}

		parent := filepath.Dir(dir)

		require.NotEqual(t, parent, dir, "go.mod not found in any parent directory")

		dir = parent
	}
}

func TestCheck_Integration(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-lint-architecture-links")
	err := Check(logger)

	require.NoError(t, err, "all current instruction/agent/skill file anchors should resolve to real ARCHITECTURE.md headings")
}

func TestHeadingToAnchor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple heading",
			input:    "Quality Strategy",
			expected: "quality-strategy",
		},
		{
			name:     "heading with dots removed",
			input:    "2.5 Quality Strategy",
			expected: "25-quality-strategy",
		},
		{
			name:     "heading with ampersand creates double hyphen",
			input:    "PKI Architecture & Strategy",
			expected: "pki-architecture--strategy",
		},
		{
			name:     "heading with underscore preserved",
			input:    "format_go Self-Modification Protection - CRITICAL",
			expected: "format_go-self-modification-protection---critical",
		},
		{
			name:     "heading with section number and underscores",
			input:    "11.2.8 format_go Self-Modification Protection - CRITICAL",
			expected: "1128-format_go-self-modification-protection---critical",
		},
		{
			name:     "parens removed",
			input:    "Headless Authentication Methods (13 total)",
			expected: "headless-authentication-methods-13-total",
		},
		{
			name:     "leading spaces trimmed by caller",
			input:    "Simple Heading",
			expected: "simple-heading",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "all special chars removed",
			input:    "!@#$%",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := headingToAnchor(tt.input)

			require.Equal(t, tt.expected, got)
		})
	}
}

func TestExtractAnchors(t *testing.T) {
	t.Parallel()

	content := `# Executive Summary
## 2.5 Quality Strategy
### 11.2.8 format_go Self-Modification Protection - CRITICAL
#### 6.9.2 Headless Authentication Methods (13 total)

Some regular content line without heading.

## 6.5 PKI Architecture & Strategy
`

	anchors := extractAnchors(content)

	expected := []string{
		"executive-summary",
		"25-quality-strategy",
		"1128-format_go-self-modification-protection---critical",
		"692-headless-authentication-methods-13-total",
		"65-pki-architecture--strategy",
	}

	for _, anchor := range expected {
		require.Contains(t, anchors, anchor, "expected anchor %q to be extracted", anchor)
	}
}

func TestCheckWithFS_AllAnchorsValid(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	archContent := "# Executive Summary\n## 2.5 Quality Strategy\n### 10.2 Unit Testing Strategy\n"

	walkFn := func(absDir string, fn fs.WalkDirFunc) error {
		name := testInstructionFileName

		return fn(filepath.Join(absDir, name), &fakeDirEntry{name: name, isDir: false}, nil)
	}

	archAbsPath := filepath.Join(root, filepath.FromSlash(testArchPath))

	readFileFn := func(name string) ([]byte, error) {
		if name == archAbsPath {
			return []byte(archContent), nil
		}

		return []byte("See [ARCHITECTURE.md Section 2.5](../../docs/ARCHITECTURE.md#25-quality-strategy) for details.\n"), nil
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.NoError(t, err)
}

func TestCheckWithFS_BrokenAnchor(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	archContent := testArchContentQuality

	walkCallCount := 0
	walkFn := func(absDir string, fn fs.WalkDirFunc) error {
		walkCallCount++
		if walkCallCount != 1 {
			return nil
		}

		name := testInstructionFileName

		return fn(filepath.Join(absDir, name), &fakeDirEntry{name: name, isDir: false}, nil)
	}

	readFileFn := func(name string) ([]byte, error) {
		if name == filepath.Join(root, filepath.FromSlash(testArchPath)) {
			return []byte(archContent), nil
		}

		return []byte("See [ARCHITECTURE.md Section 99](../../docs/ARCHITECTURE.md#99-nonexistent-section) for details.\n"), nil
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "1 broken ARCHITECTURE.md anchor(s) found")
}

func TestCheckWithFS_GetwdError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	getwdFn := func() (string, error) { return "", fmt.Errorf("getwd failed") }

	walkFn := func(_ string, _ fs.WalkDirFunc) error {
		return fmt.Errorf("should not be called")
	}

	readFileFn := func(_ string) ([]byte, error) {
		return nil, fmt.Errorf("should not be called")
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get working directory")
}

func TestCheckWithFS_ReadArchError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	walkFn := func(_ string, _ fs.WalkDirFunc) error {
		return fmt.Errorf("should not be called")
	}

	readFileFn := func(_ string) ([]byte, error) {
		return nil, fmt.Errorf("disk read error")
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read ARCHITECTURE.md")
}

func TestCheckWithFS_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	archContent := testArchContentSimple
	callCount := 0

	readFileFn := func(name string) ([]byte, error) {
		callCount++
		if callCount == 1 {
			return []byte(archContent), nil
		}

		return nil, fmt.Errorf("should not be called for walk targets")
	}

	walkFn := func(_ string, _ fs.WalkDirFunc) error {
		return fmt.Errorf("permission denied")
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk")
}

func TestCheckWithFS_ReadTargetFileError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	archContent := testArchContentSimple
	callCount := 0

	readFileFn := func(_ string) ([]byte, error) {
		callCount++
		if callCount == 1 {
			return []byte(archContent), nil
		}

		return nil, fmt.Errorf("disk read error")
	}

	walkFn := func(absDir string, fn fs.WalkDirFunc) error {
		name := testInstructionFileName

		return fn(filepath.Join(absDir, name), &fakeDirEntry{name: name, isDir: false}, nil)
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read")
}

func TestCheckWithFS_WalkEntryError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	archContent := testArchContentSimple

	readFileFn := func(_ string) ([]byte, error) {
		return []byte(archContent), nil
	}

	walkFn := func(absDir string, fn fs.WalkDirFunc) error {
		return fn(filepath.Join(absDir, "bad.md"), &fakeDirEntry{name: "bad.md", isDir: false}, fmt.Errorf("stat error"))
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to access")
}

func TestCheckWithFS_SkipsNonMdFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	archContent := testArchContentSimple

	var readCalled []string

	readFileFn := func(name string) ([]byte, error) {
		readCalled = append(readCalled, filepath.Base(name))

		return []byte(archContent), nil
	}

	walkCallCount := 0
	walkFn := func(absDir string, fn fs.WalkDirFunc) error {
		walkCallCount++
		if walkCallCount != 1 {
			return nil
		}

		for _, name := range []string{"README.txt", "data.yaml", "valid.md"} {
			if err := fn(filepath.Join(absDir, name), &fakeDirEntry{name: name, isDir: false}, nil); err != nil {
				return err
			}
		}

		return nil
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.NoError(t, err)
	// readFileFn is called once for ARCHITECTURE.md and once for valid.md only.
	require.Len(t, readCalled, 2)
}

func TestCheckWithFS_SkipsExcludedScaffoldDirs(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	archContent := testArchContentSimple

	var filesRead []string

	readFileFn := func(name string) ([]byte, error) {
		filesRead = append(filesRead, filepath.Base(name))

		return []byte(archContent), nil
	}

	walkCallCount := 0
	walkFn := func(absDir string, fn fs.WalkDirFunc) error {
		walkCallCount++
		// Simulate a scaffold dir entry â€” fn returns fs.SkipDir for excluded dirs.
		scaffoldDir := filepath.Join(absDir, "instruction-scaffold")

		err := fn(scaffoldDir, &fakeDirEntry{name: "instruction-scaffold", isDir: true}, nil)
		if err != nil && !errors.Is(err, fs.SkipDir) {
			return err
		}
		// In real WalkDir, fs.SkipDir just skips that directory's children â€”
		// the walk continues with the next sibling entry.
		// Only emit the instruction file on the first walk call.
		if walkCallCount == 1 {
			name := testInstructionFileName

			return fn(filepath.Join(absDir, name), &fakeDirEntry{name: name, isDir: false}, nil)
		}

		return nil
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.NoError(t, err)
	// Only ARCHITECTURE.md and the instruction file should be read (not scaffold files).
	require.Len(t, filesRead, 2)
}

func TestCheckWithFS_MultipleBrokenAnchors(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	archContent := testArchContentSimple

	fileContent := `See ARCHITECTURE.md#broken-one.
See ARCHITECTURE.md#broken-two.
See ARCHITECTURE.md#executive-summary valid.
`
	callCount := 0

	readFileFn := func(_ string) ([]byte, error) {
		callCount++
		if callCount == 1 {
			return []byte(archContent), nil
		}

		return []byte(fileContent), nil
	}

	walkCallCount := 0
	walkFn := func(absDir string, fn fs.WalkDirFunc) error {
		walkCallCount++
		if walkCallCount != 1 {
			return nil
		}

		name := testInstructionFileName

		return fn(filepath.Join(absDir, name), &fakeDirEntry{name: name, isDir: false}, nil)
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "2 broken ARCHITECTURE.md anchor(s) found")
}

func TestCheckWithFS_AgentFilesChecked(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	archContent := testArchContentQuality
	callCount := 0

	readFileFn := func(_ string) ([]byte, error) {
		callCount++
		if callCount == 1 {
			return []byte(archContent), nil
		}

		return []byte("See ARCHITECTURE.md#99-missing for details.\n"), nil
	}

	// Only provide an agent file (not instruction) to verify agents dir is scanned.
	walkCallCount := 0

	walkFn := func(absDir string, fn fs.WalkDirFunc) error {
		walkCallCount++
		// Only call fn on second walk (agents dir).
		if walkCallCount != 2 { //nolint:mnd // second walk = agents dir
			return nil
		}

		name := testAgentFileName

		return fn(filepath.Join(absDir, name), &fakeDirEntry{name: name, isDir: false}, nil)
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "1 broken ARCHITECTURE.md anchor(s) found")
}

// fakeDirEntry implements fs.DirEntry for testing.
type fakeDirEntry struct {
	name  string
	isDir bool
}

func (f *fakeDirEntry) Name() string               { return f.name }
func (f *fakeDirEntry) IsDir() bool                { return f.isDir }
func (f *fakeDirEntry) Type() fs.FileMode          { return 0 }
func (f *fakeDirEntry) Info() (fs.FileInfo, error) { return nil, nil }
