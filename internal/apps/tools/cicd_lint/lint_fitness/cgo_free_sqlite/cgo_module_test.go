// Copyright (c) 2025 Justin Cranford

package cgo_free_sqlite

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

var nonexistentGoMod = "/nonexistent/path/go.mod"

func TestCheckGoModForCGO(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		content        string
		wantErr        string
		wantLen        int
		wantContains   []string
		useNonexistent bool
	}{
		{
			name: "valid file",
			content: "module example.com/myproject\n\ngo 1.21\n\nrequire (\n" +
				"\tmodernc.org/sqlite v1.29.0\n" +
				"\tgithub.com/golang-migrate/migrate/v4 v4.17.0\n)\n",
			wantLen: 0,
		},
		{
			name: "banned modules",
			content: "module example.com/myproject\n\ngo 1.21\n\nrequire (\n" +
				"\tgithub.com/mattn/go-sqlite3 v1.14.19\n" +
				"\tgithub.com/golang-migrate/migrate/v4/database/sqlite3 v4.17.0\n)\n",
			wantLen:      2,
			wantContains: []string{"go-sqlite3", "database/sqlite3"},
		},
		{
			name: "indirect module skipped",
			content: "module example.com/myproject\n\ngo 1.21\n\nrequire (\n" +
				"\tgithub.com/mattn/go-sqlite3 v1.14.19 // indirect\n)\n",
			wantLen: 0,
		},
		{
			name:           "file not found",
			useNonexistent: true,
			wantErr:        "failed to open go.mod",
		},
		{
			name:    "scanner error",
			content: "module test\n// " + strings.Repeat("x", 70000) + "\n",
			wantErr: "error reading go.mod",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var goModFile string
			if tc.useNonexistent {
				goModFile = "/nonexistent/path/go.mod"
			} else {
				tmpDir := t.TempDir()
				goModFile = filepath.Join(tmpDir, "go.mod")
				require.NoError(t, os.WriteFile(goModFile, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))
			}

			violations, err := CheckGoModForCGO(goModFile)

			switch {
			case tc.wantErr != "":
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
			default:
				require.NoError(t, err)
				require.Len(t, violations, tc.wantLen)

				joined := strings.Join(violations, "\n")
				for _, want := range tc.wantContains {
					require.Contains(t, joined, want)
				}
			}
		})
	}
}

func TestCheckRequiredCGOModule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		content        string
		wantFound      bool
		wantErr        string
		useNonexistent bool
	}{
		{
			name: "found",
			content: "module example.com/myproject\n\ngo 1.21\n\nrequire (\n" +
				"\tmodernc.org/sqlite v1.29.0\n)\n",
			wantFound: true,
		},
		{
			name: "not found",
			content: "module example.com/myproject\n\ngo 1.21\n\nrequire (\n" +
				"\tgithub.com/some/other/module v1.0.0\n)\n",
			wantFound: false,
		},
		{
			name:           "file not found",
			useNonexistent: true,
			wantErr:        "failed to open go.mod",
		},
		{
			name:    "scanner error",
			content: "module test\n// " + strings.Repeat("x", 70000) + "\n",
			wantErr: "error reading go.mod",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var goModFile string
			if tc.useNonexistent {
				goModFile = nonexistentGoMod
			} else {
				tmpDir := t.TempDir()
				goModFile = filepath.Join(tmpDir, "go.mod")
				require.NoError(t, os.WriteFile(goModFile, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))
			}

			found, err := CheckRequiredCGOModule(goModFile)

			switch {
			case tc.wantErr != "":
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
			default:
				require.NoError(t, err)
				require.Equal(t, tc.wantFound, found)
			}
		})
	}
}

func TestCheckGoFileForCGO(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		filename       string
		subdir         string
		content        string
		wantEmpty      bool
		wantContains   string
		wantErr        string
		useNonexistent bool
	}{
		{
			name:     "clean file",
			filename: "clean.go",
			content: "package main\n\nimport (\n" +
				"\t\"modernc.org/sqlite\"\n" +
				"\t\"github.com/golang-migrate/migrate/v4/database/sqlite\"\n)\n\n" +
				"func main() {\n\t// Using CGO-free sqlite\n}\n",
			wantEmpty: true,
		},
		{
			name:     "banned import",
			filename: "banned.go",
			content: "package main\n\nimport (\n" +
				"\t_ \"github.com/mattn/go-sqlite3\"\n)\n\n" +
				"func main() {\n}\n",
			wantContains: "banned CGO import detected",
		},
		{
			name:     "banned migrate import",
			filename: "banned_migrate.go",
			content: "package main\n\nimport (\n" +
				"\t_ \"github.com/golang-migrate/migrate/v4/database/sqlite3\"\n)\n\n" +
				"func main() {\n}\n",
			wantContains: "banned CGO migrate import detected",
		},
		{
			name:      "lint_go directory skipped",
			filename:  "lint_go.go",
			subdir:    "lint_go",
			content:   "package main\n\nimport (\n\t_ \"github.com/mattn/go-sqlite3\"\n)\n",
			wantEmpty: true,
		},
		{
			name:           "file not found",
			useNonexistent: true,
			wantErr:        "failed to open",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var filePath string
			if tc.useNonexistent {
				filePath = "/nonexistent/path/file.go"
			} else {
				tmpDir := t.TempDir()

				dir := tmpDir
				if tc.subdir != "" {
					dir = filepath.Join(tmpDir, tc.subdir)
					require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				}

				filePath = filepath.Join(dir, tc.filename)
				require.NoError(t, os.WriteFile(filePath, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))
			}

			violations, err := CheckGoFileForCGO(filePath)

			switch {
			case tc.wantErr != "":
				require.Error(t, err)
				require.Nil(t, violations)
				require.Contains(t, err.Error(), tc.wantErr)
			case tc.wantEmpty:
				require.NoError(t, err)
				require.Empty(t, violations)
			default:
				require.NoError(t, err)
				require.NotEmpty(t, violations)
				require.Contains(t, strings.Join(violations, "\n"), tc.wantContains)
			}
		})
	}
}

// Sequential: redirects os.Stderr (global process state, cannot run in parallel).
func TestPrintCGOViolations(t *testing.T) {
	tests := []struct {
		name             string
		goModViolations  []string
		importViolations []string
		hasRequired      bool
		wantContains     []string
		wantNotContains  []string
	}{
		{
			name:             "all types",
			goModViolations:  []string{"go.mod:5: banned CGO module"},
			importViolations: []string{"file.go:10: banned CGO import"},
			hasRequired:      false,
			wantContains:     []string{"CGO validation failed", "go.mod violations", "Import violations", "Required module missing"},
		},
		{
			name:            "go.mod only",
			goModViolations: []string{"go.mod:5: banned module"},
			hasRequired:     true,
			wantContains:    []string{"go.mod violations"},
			wantNotContains: []string{"Import violations", "Required module missing"},
		},
		{
			name:             "import only",
			importViolations: []string{"file.go:10: banned import"},
			hasRequired:      true,
			wantContains:     []string{"Import violations"},
			wantNotContains:  []string{"go.mod violations", "Required module missing"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			PrintCGOViolations(tc.goModViolations, tc.importViolations, tc.hasRequired)

			_ = w.Close()
			os.Stderr = oldStderr

			output, _ := io.ReadAll(r)
			outputStr := string(output)

			for _, want := range tc.wantContains {
				require.Contains(t, outputStr, want)
			}

			for _, notWant := range tc.wantNotContains {
				require.NotContains(t, outputStr, notWant)
			}
		})
	}
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck(t *testing.T) {
	tests := []struct {
		name       string
		setupFiles map[string]string
		wantErr    string
	}{
		{
			name: "with required module",
			setupFiles: map[string]string{
				"go.mod":  "module testmod\n\ngo 1.21\n\nrequire (\n\tmodernc.org/sqlite v1.30.0\n)\n",
				"main.go": testMainContent,
			},
		},
		{
			name: "missing required module",
			setupFiles: map[string]string{
				"go.mod":  "module testmod\n\ngo 1.21\n",
				"main.go": testMainContent,
			},
			wantErr: "CGO validation failed",
		},
		{
			name:    "no go.mod",
			wantErr: "failed to check go.mod",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origDir, err := os.Getwd()
			require.NoError(t, err)

			defer func() { require.NoError(t, os.Chdir(origDir)) }()

			tmpDir := t.TempDir()
			require.NoError(t, os.Chdir(tmpDir))

			for name, content := range tc.setupFiles {
				require.NoError(t, os.WriteFile(name, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
			}

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			err = Check(logger)

			switch {
			case tc.wantErr != "":
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
			default:
				require.NoError(t, err)
			}
		})
	}
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_WalkError(t *testing.T) {
	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod does not enforce POSIX permissions on Windows")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	// Write go.mod with required module so CheckGoModForCGO passes.
	goModContent := "module testmod\n\ngo 1.21\n\nrequire (\n\tmodernc.org/sqlite v1.30.0\n)\n"
	require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), cryptoutilSharedMagic.CacheFilePermissions))

	// Create chmod 0000 subdir to trigger walk error.
	require.NoError(t, os.MkdirAll("locked", 0o700))
	require.NoError(t, os.Chmod("locked", 0o000))

	t.Cleanup(func() { _ = os.Chmod(filepath.Join(tmpDir, "locked"), 0o700) })

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check Go files")
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-cgo-free-sqlite")

	err = Check(logger)
	require.NoError(t, err)
}
