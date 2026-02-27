// Copyright (c) 2025 Justin Cranford

package cmd_main_pattern

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
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

	// No cmd/ directory present - should succeed silently.
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
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

func TestCheckInDir_WalkError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	tmpDir := t.TempDir()

	// Create cmd/ but with inaccessible subdirectory.
	cmdDir := filepath.Join(tmpDir, "cmd")
	require.NoError(t, os.MkdirAll(cmdDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	badDir := filepath.Join(cmdDir, "badapp")
	require.NoError(t, os.MkdirAll(badDir, 0o000))
	t.Cleanup(func() { _ = os.Chmod(badDir, 0o700) })

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

// Sequential: uses os.Chdir (global process state).
func TestCheck_DelegatesCheckInDir(t *testing.T) {
	// NOTE: Cannot use t.Parallel() - test changes working directory.
	origDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")

	err = Check(logger)
	require.NoError(t, err) // no cmd/ dir in temp dir → no violations
}
