// Copyright (c) 2025 Justin Cranford

package service_contract_compliance

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

// makeServiceDir creates the internal/apps/<product>/<service>/server/ tree.
// Returns the path to server.go.
func makeServiceDir(t *testing.T, root, product, service, serverContent string) string {
	t.Helper()

	serverDir := filepath.Join(root, "internal", "apps", product, service, "server")
	require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.DirPermissions))
	serverPath := filepath.Join(serverDir, "server.go")
	require.NoError(t, os.WriteFile(serverPath, []byte(serverContent), cryptoutilSharedMagic.CacheFilePermissions))

	return serverPath
}

const goodServerGo = "package server\n\nimport someapi \"example.com/api\"\n\nvar _ someapi.ServiceServer = (*MyServer)(nil)\n\ntype MyServer struct{}\n"

const missingAssertionServerGo = "package server\n\ntype MyServer struct{}\n\nfunc (s *MyServer) Hello() {}\n"

func TestCheckInDir_WithValidAssertion_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeServiceDir(t, tmp, cryptoutilSharedMagic.JoseProductName, "ja", goodServerGo)
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_MissingAssertion_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeServiceDir(t, tmp, cryptoutilSharedMagic.JoseProductName, "ja", missingAssertionServerGo)
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "service contract compliance")
}

func TestCheckInDir_EmptyDir_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// Create apps dir so ReadDir doesn't error.
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "internal", "apps"), cryptoutilSharedMagic.DirPermissions))
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_SkeletonService_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// skeleton product is always excluded from compliance checks.
	makeServiceDir(t, tmp, cryptoutilSharedMagic.SkeletonProductName, cryptoutilSharedMagic.SkeletonTemplateServiceName, missingAssertionServerGo)
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_CicdService_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeServiceDir(t, tmp, "cicd", "cmd", missingAssertionServerGo)
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_MultipleServices_AllValid_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeServiceDir(t, tmp, cryptoutilSharedMagic.JoseProductName, "ja", goodServerGo)
	makeServiceDir(t, tmp, cryptoutilSharedMagic.PKIProductName, "ca", goodServerGo)
	makeServiceDir(t, tmp, "sm", cryptoutilSharedMagic.KMSServiceName, goodServerGo)
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_MultipleServices_OneFails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	makeServiceDir(t, tmp, cryptoutilSharedMagic.JoseProductName, "ja", goodServerGo)
	makeServiceDir(t, tmp, cryptoutilSharedMagic.PKIProductName, "ca", missingAssertionServerGo)
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
}

func TestDiscoverServices_NoAppsDir_Error(t *testing.T) {
	t.Parallel()
	// discoverServices reads the appsDir; missing dir causes ReadDir error.
	_, err := discoverServices("/nonexistent/apps", os.ReadDir)
	require.Error(t, err)
}

func TestDiscoverServices_EmptyAppsDir_ReturnsNil(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	appsDir := filepath.Join(tmp, "apps")
	require.NoError(t, os.MkdirAll(appsDir, cryptoutilSharedMagic.DirPermissions))
	services, err := discoverServices(appsDir, os.ReadDir)
	require.NoError(t, err)
	require.Empty(t, services)
}

func TestDiscoverServices_WithService_ReturnsIt(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// Create <tmp>/apps/<product>/<service>/server/server.go
	serverDir := filepath.Join(tmp, "apps", cryptoutilSharedMagic.JoseProductName, "ja", "server")
	require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
	services, err := discoverServices(filepath.Join(tmp, "apps"), os.ReadDir)
	require.NoError(t, err)
	require.Len(t, services, 1)
	require.Equal(t, cryptoutilSharedMagic.JoseProductName, services[0].product)
	require.Equal(t, "ja", services[0].service)
}

func TestCheckServerFile_WithAssertion_NoViolation(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "server.go")
	require.NoError(t, os.WriteFile(p, []byte(goodServerGo), cryptoutilSharedMagic.CacheFilePermissions))

	var violations []string

	svc := serviceID{product: cryptoutilSharedMagic.JoseProductName, service: "ja"}
	err := checkServerFile(p, svc, &violations, os.ReadFile)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckServerFile_MissingAssertion_AddsViolation(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "server.go")
	require.NoError(t, os.WriteFile(p, []byte(missingAssertionServerGo), cryptoutilSharedMagic.CacheFilePermissions))

	var violations []string

	svc := serviceID{product: cryptoutilSharedMagic.JoseProductName, service: "ja"}
	err := checkServerFile(p, svc, &violations, os.ReadFile)
	require.NoError(t, err) // returns nil, appends to violations
	require.NotEmpty(t, violations)
}

func TestCheckServerFile_NonexistentFile_AddsViolation(t *testing.T) {
	t.Parallel()

	var violations []string

	svc := serviceID{product: cryptoutilSharedMagic.JoseProductName, service: "ja"}
	// Missing file is reported as a violation (not an error) per implementation.
	err := checkServerFile("/nonexistent/server.go", svc, &violations, os.ReadFile)
	require.NoError(t, err)
	require.NotEmpty(t, violations)
}

func TestDiscoverServices_NonDirFileInAppsDir_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	appsDir := filepath.Join(tmp, "apps")
	require.NoError(t, os.MkdirAll(appsDir, cryptoutilSharedMagic.DirPermissions))
	// Create a FILE (not directory) directly in appsDir - triggers !p.IsDir() continue.
	require.NoError(t, os.WriteFile(filepath.Join(appsDir, "README.md"), []byte(cryptoutilSharedMagic.CICDExcludeDirDocs), cryptoutilSharedMagic.CacheFilePermissions))
	services, err := discoverServices(appsDir, os.ReadDir)
	require.NoError(t, err)
	require.Empty(t, services)
}

func TestDiscoverServices_NonDirFileInProductDir_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	appsDir := filepath.Join(tmp, "apps")
	productDir := filepath.Join(appsDir, cryptoutilSharedMagic.JoseProductName)
	require.NoError(t, os.MkdirAll(productDir, cryptoutilSharedMagic.DirPermissions))
	// Create a FILE (not directory) in the product dir - triggers !s.IsDir() continue.
	require.NoError(t, os.WriteFile(filepath.Join(productDir, "README.md"), []byte(cryptoutilSharedMagic.CICDExcludeDirDocs), cryptoutilSharedMagic.CacheFilePermissions))

	services, err := discoverServices(appsDir, os.ReadDir)
	require.NoError(t, err)
	require.Empty(t, services)
}

func TestDiscoverServices_ReadDirSeamError(t *testing.T) {
	t.Parallel()

	stubReadDirFn := func(_ string) ([]os.DirEntry, error) {
		return nil, fmt.Errorf("injected readdir error")
	}

	err := checkInDir(newTestLogger(), t.TempDir(), stubReadDirFn, os.ReadFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to discover services")
}

func TestCheckServerFile_ReadFileError(t *testing.T) {
	t.Parallel()

	stubReadFileFn := func(_ string) ([]byte, error) {
		return nil, fmt.Errorf("injected read error")
	}

	tmp := t.TempDir()
	makeServiceDir(t, tmp, "sm", "im", "package server\ntype MyServer struct{}\n")

	err := checkInDir(newTestLogger(), tmp, os.ReadDir, stubReadFileFn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check")
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
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

	logger := cryptoutilCmdCicdCommon.NewLogger("test-service-contract-compliance")

	err = Check(logger)
	require.NoError(t, err)
}

func TestDiscoverServices_ArchivedDirSkipped(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	appsDir := filepath.Join(tmp, "apps")

	productDir := filepath.Join(appsDir, cryptoutilSharedMagic.JoseProductName)
	require.NoError(t, os.MkdirAll(productDir, cryptoutilSharedMagic.DirPermissions))

	// Create a dir prefixed with _ (archived).
	archivedDir := filepath.Join(productDir, "_archived_service")
	require.NoError(t, os.MkdirAll(archivedDir, cryptoutilSharedMagic.DirPermissions))

	services, err := discoverServices(appsDir, os.ReadDir)
	require.NoError(t, err)
	require.Empty(t, services)
}

func TestCheckInDir_NoAppsDir_Error(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	// No internal/apps dir created — discoverServices should fail.
	err := CheckInDir(newTestLogger(), tmp)
	// Should error because internal/apps doesn't exist.
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to discover services")
}
