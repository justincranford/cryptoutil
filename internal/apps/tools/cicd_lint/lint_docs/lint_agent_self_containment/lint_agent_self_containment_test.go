// Copyright (c) 2025 Justin Cranford

package lint_agent_self_containment

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testBeastModeAgentFile = "beast-mode.agent.md"

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

	logger := cryptoutilCmdCicdCommon.NewLogger("test-lint-agent-self-containment")
	err := Check(logger)

	require.NoError(t, err, "all current agents should reference ARCHITECTURE.md")
}

func TestCheckWithFS_AllAgentsCompliant(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	walkFn := func(_ string, fn fs.WalkDirFunc) error {
		files := []struct {
			name    string
			content string
		}{
			{testBeastModeAgentFile, "# Beast Mode\n\nSee [ARCHITECTURE.md Section 14.7](docs/ARCHITECTURE.md#147).\n"},
			{"fix-workflows.agent.md", "# Fix Workflows\n\nSee ARCHITECTURE.md Section 9.\n"},
		}

		for _, f := range files {
			if err := fn(filepath.Join(root, cryptoutilSharedMagic.CICDGithubAgentsDir, f.name), &fakeAgentDirEntry{name: f.name, isDir: false}, nil); err != nil {
				return err
			}
		}

		return nil
	}

	readFileFn := func(name string) ([]byte, error) {
		switch filepath.Base(name) {
		case testBeastModeAgentFile:
			return []byte("# Beast Mode\n\nSee [ARCHITECTURE.md Section 14.7](docs/ARCHITECTURE.md#147).\n"), nil
		case "fix-workflows.agent.md":
			return []byte("# Fix Workflows\n\nSee ARCHITECTURE.md Section 9.\n"), nil
		}

		return nil, fmt.Errorf("unexpected file: %s", name)
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.NoError(t, err)
}

func TestCheckWithFS_AgentMissingReference(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	walkFn := func(_ string, fn fs.WalkDirFunc) error {
		files := []string{"beast-mode.agent.md", "noncompliant.agent.md"}

		for _, name := range files {
			if err := fn(filepath.Join(root, cryptoutilSharedMagic.CICDGithubAgentsDir, name), &fakeAgentDirEntry{name: name, isDir: false}, nil); err != nil {
				return err
			}
		}

		return nil
	}

	readFileFn := func(name string) ([]byte, error) {
		switch filepath.Base(name) {
		case "beast-mode.agent.md":
			return []byte("# Beast Mode\nSee ARCHITECTURE.md Section 14.\n"), nil
		case "noncompliant.agent.md":
			return []byte("# Noncompliant Agent\n\nNo architecture reference here.\n"), nil
		}

		return nil, fmt.Errorf("unexpected file: %s", name)
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "1 agent(s) missing ARCHITECTURE.md references")
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

func TestCheckWithFS_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	walkFn := func(_ string, _ fs.WalkDirFunc) error {
		return fmt.Errorf("permission denied")
	}

	readFileFn := func(_ string) ([]byte, error) {
		return nil, fmt.Errorf("should not be called")
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk")
}

func TestCheckWithFS_ReadFileError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	walkFn := func(_ string, fn fs.WalkDirFunc) error {
		name := "agent.agent.md"

		return fn(filepath.Join(root, cryptoutilSharedMagic.CICDGithubAgentsDir, name), &fakeAgentDirEntry{name: name, isDir: false}, nil)
	}

	readFileFn := func(_ string) ([]byte, error) {
		return nil, fmt.Errorf("disk read error")
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

	walkFn := func(_ string, fn fs.WalkDirFunc) error {
		return fn("bad-path", &fakeAgentDirEntry{name: "bad-path", isDir: false}, fmt.Errorf("stat error"))
	}

	readFileFn := func(_ string) ([]byte, error) {
		return nil, nil
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to access")
}

func TestCheckWithFS_SkipsDirectoriesAndNonAgentFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	var readCalled []string

	walkFn := func(_ string, fn fs.WalkDirFunc) error {
		entries := []struct {
			name  string
			isDir bool
		}{
			{"subdir", true},
			{"README.md", false},
			{"beast-mode.agent.md", false},
		}

		for _, e := range entries {
			path := filepath.Join(root, cryptoutilSharedMagic.CICDGithubAgentsDir, e.name)
			if err := fn(path, &fakeAgentDirEntry{name: e.name, isDir: e.isDir}, nil); err != nil {
				return err
			}
		}

		return nil
	}

	readFileFn := func(name string) ([]byte, error) {
		readCalled = append(readCalled, filepath.Base(name))

		return []byte("# Agent\nSee ARCHITECTURE.md Section 2.\n"), nil
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.NoError(t, err)
	require.Equal(t, []string{"beast-mode.agent.md"}, readCalled, "should only read .agent.md files")
}

func TestCheckWithFS_EmptyDirectory(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	walkFn := func(_ string, _ fs.WalkDirFunc) error {
		return nil
	}

	readFileFn := func(_ string) ([]byte, error) {
		return nil, nil
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "no .agent.md files found")
}

func TestCheckWithFS_MultipleViolations(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	root := findTestProjectRoot(t)

	getwdFn := func() (string, error) { return root, nil }

	walkFn := func(_ string, fn fs.WalkDirFunc) error {
		for _, name := range []string{"a.agent.md", "b.agent.md", "c.agent.md"} {
			path := filepath.Join(root, cryptoutilSharedMagic.CICDGithubAgentsDir, name)
			if err := fn(path, &fakeAgentDirEntry{name: name, isDir: false}, nil); err != nil {
				return err
			}
		}

		return nil
	}

	readFileFn := func(_ string) ([]byte, error) {
		return []byte("# No reference here\n"), nil
	}

	err := checkWithFS(logger, getwdFn, walkFn, readFileFn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "3 agent(s) missing ARCHITECTURE.md references")
}

// fakeAgentDirEntry implements fs.DirEntry for testing.
type fakeAgentDirEntry struct {
	name  string
	isDir bool
}

func (f *fakeAgentDirEntry) Name() string               { return f.name }
func (f *fakeAgentDirEntry) IsDir() bool                { return f.isDir }
func (f *fakeAgentDirEntry) Type() fs.FileMode          { return 0 }
func (f *fakeAgentDirEntry) Info() (fs.FileInfo, error) { return nil, nil }
