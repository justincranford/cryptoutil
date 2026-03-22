// Copyright (c) 2025 Justin Cranford

package domain_layer_isolation

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
}

func TestCheckInDir_NoViolations(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "internal", "apps", "sm", "im", "domain", "message.go"), `package domain

import (
"time"
"github.com/google/uuid"
)

type Message struct {
ID        uuid.UUID
CreatedAt time.Time
}
`)
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_DomainImportsServer_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "internal", "apps", "sm", "im", "domain", "bad.go"), `package domain

import (
"cryptoutil/internal/apps/sm/im/server/repository"
)

var _ = repository.User{}
`)
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "domain layer isolation")
}

func TestCheckInDir_DomainImportsClient_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "internal", "apps", "sm", "im", "domain", "bad.go"), `package domain

import (
"cryptoutil/internal/apps/sm/im/client"
)

var _ = client.Client{}
`)
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
}

func TestCheckInDir_DomainImportsAPI_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "internal", "apps", "sm", "im", "domain", "bad.go"), `package domain

import (
"cryptoutil/internal/apps/sm/im/api/model"
)

var _ = model.Message{}
`)
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
}

func TestCheckInDir_ImportContainsServerPath_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "internal", "apps", "sm", "im", "domain", "bad.go"), `package domain

import (
"cryptoutil/internal/apps/sm/im/server/internal/sub"
)

var _ = sub.X
`)
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
}

func TestCheckInDir_TestFileInDomain_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "internal", "apps", "sm", "im", "domain", "message_test.go"), `package domain

import (
"cryptoutil/internal/apps/sm/im/server/repository"
"testing"
)

func TestFoo(t *testing.T) { _ = repository.User{} }
`)
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_EmptyDir_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "internal", "apps"), cryptoutilSharedMagic.DirPermissions))
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestIsDomainFile_Various(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		wantBool bool
	}{
		{"direct domain file", filepath.Join("internal", "apps", "sm", "im", "domain", "message.go"), true},
		{"server file", filepath.Join("internal", "apps", "sm", "im", "server", "server.go"), false},
		{"nested domain", filepath.Join("internal", "apps", "sm", "im", "domain", "nested", "file.go"), true},
		{"root file", "main.go", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isDomainFile(tc.path)
			require.Equal(t, tc.wantBool, got)
		})
	}
}

func TestExtractImportPath_Various(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		line  string
		wantP string
	}{
		{"quoted import", `"fmt"`, "fmt"},
		{"aliased import", `myAlias "some/package"`, "some/package"},
		{"blank import", `_ "some/init"`, "some/init"},
		{"empty line", "", ""},
		{"no quotes", "just text", ""},
		{"single quote", `"only_one`, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := extractImportPath(tc.line)
			require.Equal(t, tc.wantP, got)
		})
	}
}

func TestScanDomainFile_CleanFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	goFile := filepath.Join(tmp, "message.go")
	require.NoError(t, os.WriteFile(goFile, []byte(`package domain

import (
"fmt"
"time"
)

func init() { fmt.Println(time.Now()) }
`), cryptoutilSharedMagic.CacheFilePermissions))
	violations, err := scanDomainFile(goFile, tmp)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestScanDomainFile_WithViolation(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	goFile := filepath.Join(tmp, "bad.go")
	require.NoError(t, os.WriteFile(goFile, []byte(`package domain

import (
"cryptoutil/internal/apps/sm/im/server/repository"
)

var _ = repository.User{}
`), cryptoutilSharedMagic.CacheFilePermissions))
	violations, err := scanDomainFile(goFile, tmp)
	require.NoError(t, err)
	require.Len(t, violations, 1)
}

func TestScanDomainFile_NonexistentFile_Error(t *testing.T) {
	t.Parallel()

	_, err := scanDomainFile("/nonexistent/file.go", "/tmp")
	require.Error(t, err)
}

// Sequential: modifies package-level domainIsolationWalkFn seam.
func TestCheckInDir_WalkError(t *testing.T) {
	orig := domainIsolationWalkFn

	t.Cleanup(func() { domainIsolationWalkFn = orig })

	domainIsolationWalkFn = func(_ string, _ filepath.WalkFunc) error {
		return fmt.Errorf("injected walk error")
	}

	err := CheckInDir(newTestLogger(), t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "filesystem walk failed")
}

// Sequential: modifies package-level domainIsolationWalkFn seam.
func TestCheckInDir_WalkCallbackError(t *testing.T) {
	orig := domainIsolationWalkFn

	t.Cleanup(func() { domainIsolationWalkFn = orig })

	domainIsolationWalkFn = func(_ string, fn filepath.WalkFunc) error {
		return fn("bad/path", nil, fmt.Errorf("injected callback error"))
	}

	err := CheckInDir(newTestLogger(), t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "filesystem walk failed")
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

	logger := cryptoutilCmdCicdCommon.NewLogger("test-domain-layer-isolation")

	err = Check(logger)
	require.NoError(t, err)
}
