// Copyright (c) 2025 Justin Cranford

package health_endpoint_presence

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

func mkServiceWithServer(t *testing.T, root, product, service string) {
	t.Helper()

	serverDir := filepath.Join(root, "internal", "apps", product, service, "server")
	require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"),
		[]byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
}

func TestCheckInDir_ServiceWithHealthEndpoints_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	mkServiceWithServer(t, tmp, "sm", "im")
	// Write a file containing both required health patterns.
	serverFile := filepath.Join(tmp, "internal", "apps", "sm", "im", "server", "server.go")
	require.NoError(t, os.WriteFile(serverFile, []byte(`package server
// registers livez and readyz endpoints
func (s *Server) Start() { }
`), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_ServiceMissingHealthEndpoints_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	mkServiceWithServer(t, tmp, "sm", "im")
	// server.go has no health endpoint references.
	serverFile := filepath.Join(tmp, "internal", "apps", "sm", "im", "server", "server.go")
	require.NoError(t, os.WriteFile(serverFile, []byte(`package server
func (s *Server) Start() { }
`), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "health endpoint presence violations")
}

func TestCheckInDir_SkipCicdProduct_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	cicdDir := filepath.Join(tmp, "internal", "apps", "cicd", "linter", "server")
	require.NoError(t, os.MkdirAll(cicdDir, cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(cicdDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_SkipSkeletonProduct_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	skeletonDir := filepath.Join(tmp, "internal", "apps", cryptoutilSharedMagic.SkeletonProductName, cryptoutilSharedMagic.SkeletonTemplateServiceName, "server")
	require.NoError(t, os.MkdirAll(skeletonDir, cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(skeletonDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_SkipArchivedService_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// Archived service has no healthcheck, but should be skipped.
	archivedDir := filepath.Join(tmp, "internal", "apps", "sm", "_old", "server")
	require.NoError(t, os.MkdirAll(archivedDir, cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(archivedDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_NoServices_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "internal", "apps"), cryptoutilSharedMagic.DirPermissions))
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckServiceHealth_PatternsFoundInSubdir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	svcDir := filepath.Join(tmp, "internal", "apps", "sm", "im")
	require.NoError(t, os.MkdirAll(svcDir, cryptoutilSharedMagic.DirPermissions))
	// Pattern in a non-server.go file (still in service dir).
	require.NoError(t, os.WriteFile(filepath.Join(svcDir, "routes.go"),
		[]byte("package sm\n// sets up livez and readyz\n"), cryptoutilSharedMagic.CacheFilePermissions))

	svc := serviceID{product: "sm", service: "im"}
	violations := checkServiceHealth(svc, filepath.Join(tmp, "internal", "apps"))
	require.Empty(t, violations)
}

func TestDiscoverServices_EmptyAppsDir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	services, err := discoverServices(tmp)
	require.NoError(t, err)
	require.Empty(t, services)
}
