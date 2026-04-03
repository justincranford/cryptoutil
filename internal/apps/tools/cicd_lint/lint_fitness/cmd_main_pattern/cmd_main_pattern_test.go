// Copyright (c) 2025 Justin Cranford

package cmd_main_pattern

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

func TestCheckMainGoFile_Valid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainFile := filepath.Join(tmpDir, "main.go")

	content := `package main

import (
"os"
cryptoutilAppsFoo "cryptoutil/internal/apps/foo"
)

func main() { os.Exit(cryptoutilAppsFoo.InternalMain(os.Args, os.Stdin, os.Stdout, os.Stderr)) }
`
	err := os.WriteFile(mainFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	err = CheckMainGoFile(mainFile)
	require.NoError(t, err)
}

func TestCheckMainGoFile_ValidArgsSlice(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainFile := filepath.Join(tmpDir, "main.go")

	content := `package main

import (
"os"
cryptoutilAppsFoo "cryptoutil/internal/apps/foo"
)

func main() { os.Exit(cryptoutilAppsFoo.InternalMain(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
`
	err := os.WriteFile(mainFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	err = CheckMainGoFile(mainFile)
	require.NoError(t, err)
}

func TestCheckMainGoFile_MissingPattern(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	mainFile := filepath.Join(tmpDir, "main.go")

	// Missing the required pattern (no os.Exit wrapper).
	content := `package main

func main() {
println("hello world")
}
`
	err := os.WriteFile(mainFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	err = CheckMainGoFile(mainFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not match required pattern")
}

func TestCheckMainGoFile_FileNotFound(t *testing.T) {
	t.Parallel()

	err := CheckMainGoFile("/nonexistent/path/main.go")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read file")
}

func TestCheckMainGoFile_PatternVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid single-line pattern",
			content: `package main
import "os"
import cryptoutilAppsBar "cryptoutil/internal/apps/bar"
func main() { os.Exit(cryptoutilAppsBar.Main(os.Args, os.Stdin, os.Stdout, os.Stderr)) }
`,
			wantErr: false,
		},
		{
			name: "valid single-line pattern with args slice",
			content: `package main
import "os"
import cryptoutilAppsBar "cryptoutil/internal/apps/bar"
func main() { os.Exit(cryptoutilAppsBar.Main(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
`,
			wantErr: false,
		},
		{
			name: "missing os.Exit",
			content: `package main
import "os"
import cryptoutilAppsBar "cryptoutil/internal/apps/bar"
func main() { cryptoutilAppsBar.Main(os.Args, os.Stdin, os.Stdout, os.Stderr) }
`,
			wantErr: true,
		},
		{
			name: "missing os.Args",
			content: `package main
import "os"
import cryptoutilAppsBar "cryptoutil/internal/apps/bar"
func main() { os.Exit(cryptoutilAppsBar.Main(nil, os.Stdin, os.Stdout, os.Stderr)) }
`,
			wantErr: true,
		},
		{
			name: "lowercase cryptoutil prefix - invalid",
			content: `package main
import "os"
func main() { os.Exit(myApp.Run(os.Args, os.Stdin, os.Stdout, os.Stderr)) }
`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			mainFile := filepath.Join(tmpDir, "main.go")
			err := os.WriteFile(mainFile, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			err = CheckMainGoFile(mainFile)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckInDir_NoCmdDir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// No cmd/ directory present - hard error (cmd/ is required in all repositories).
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cmd/ directory not found")
}

func TestCheckInDir_WithValidMainFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create cmd/myapp/main.go with valid pattern.
	cmdDir := filepath.Join(tmpDir, "cmd", "myapp")
	require.NoError(t, os.MkdirAll(cmdDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	content := `package main
import "os"
import cryptoutilAppsMyApp "cryptoutil/internal/apps/myapp"
func main() { os.Exit(cryptoutilAppsMyApp.Run(os.Args, os.Stdin, os.Stdout, os.Stderr)) }
`
	err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	err = CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_WithInvalidMainFile(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create cmd/myapp/main.go with INVALID pattern.
	cmdDir := filepath.Join(tmpDir, "cmd", "myapp")
	require.NoError(t, os.MkdirAll(cmdDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	content := `package main

func main() {
println("bad pattern")
}
`
	err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	err = CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cmd/ main() pattern violations")
}

// Sequential: modifies package-level cmdMainWalkFn seam.
func TestCheckInDir_WalkError(t *testing.T) {
	orig := cmdMainWalkFn

	t.Cleanup(func() { cmdMainWalkFn = orig })

	cmdMainWalkFn = func(_ string, _ filepath.WalkFunc) error {
		return fmt.Errorf("injected walk error")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create cmd/ so the stat check passes.
	cmdDir := filepath.Join(tmpDir, "cmd")
	require.NoError(t, os.MkdirAll(cmdDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to walk cmd directory")
}

func TestCheckInDir_IgnoresNonMainFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create cmd/ with non-main.go files (should be ignored).
	cmdDir := filepath.Join(tmpDir, "cmd", "myapp")
	require.NoError(t, os.MkdirAll(cmdDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	content := `package main

func helper() {}
`
	err := os.WriteFile(filepath.Join(cmdDir, "helper.go"), []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	err = CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheck_DelegatesCheckInDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create an empty cmd/ dir so the hard-error-on-absent-cmd-dir check passes.
	if err := os.MkdirAll(filepath.Join(tmpDir, "cmd"), 0o755); err != nil {
		t.Fatalf("setup: mkdir cmd/: %v", err)
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err) // empty cmd/ dir → no violations (no main.go files to check)
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

	logger := cryptoutilCmdCicdCommon.NewLogger("test-cmd-main-pattern")

	err = Check(logger)
	require.NoError(t, err)
}
