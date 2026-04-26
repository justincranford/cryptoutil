// Copyright (c) 2025 Justin Cranford

package no_local_ps_id_server_stubs_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	lintFitnessNoLocalPSIDServerStubs "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/no_local_ps_id_server_stubs"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const stubPublicServerGoSrc = `package server
import "context"
type stubPublicServer struct{}
func (s *stubPublicServer) Start(_ context.Context) error { return nil }
func (s *stubPublicServer) Shutdown(_ context.Context) error { return nil }
func (s *stubPublicServer) ActualPort() int { return 8443 }
func (s *stubPublicServer) PublicBaseURL() string { return "" }
`

func TestFindViolations_NoViolations_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	violations, err := lintFitnessNoLocalPSIDServerStubs.FindViolations(dir)

	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolations_NoViolations_NonTestFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	psDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, "server")
	require.NoError(t, os.MkdirAll(psDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Production (non-test) file defining stub-like methods should NOT be flagged.
	require.NoError(t, os.WriteFile(filepath.Join(psDir, "server.go"), []byte(stubPublicServerGoSrc), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessNoLocalPSIDServerStubs.FindViolations(dir)

	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolations_Violation_LocalPublicStub(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	psDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, "server")
	require.NoError(t, os.MkdirAll(psDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	content := `package server
import "context"
type stubPublicServer struct{ startErr error }
func (s *stubPublicServer) Start(_ context.Context) error { return s.startErr }
func (s *stubPublicServer) Shutdown(_ context.Context) error { return nil }
func (s *stubPublicServer) ActualPort() int { return 8443 }
func (s *stubPublicServer) PublicBaseURL() string { return "https://localhost:8443" }
`
	require.NoError(t, os.WriteFile(filepath.Join(psDir, "server_test.go"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessNoLocalPSIDServerStubs.FindViolations(dir)

	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Equal(t, "stubPublicServer", violations[0].StructName)
	require.Equal(t, "IPublicServer", violations[0].Interface)
}

func TestFindViolations_Violation_LocalAdminStub(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	psDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceJoseJA, "server")
	require.NoError(t, os.MkdirAll(psDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	content := `package server
import ("context"; "crypto/x509")
type stubAdminServer struct{}
func (s *stubAdminServer) Start(_ context.Context) error { return nil }
func (s *stubAdminServer) Shutdown(_ context.Context) error { return nil }
func (s *stubAdminServer) ActualPort() int { return 9090 }
func (s *stubAdminServer) SetReady(_ bool) {}
func (s *stubAdminServer) AdminBaseURL() string { return "https://localhost:9090" }
func (s *stubAdminServer) AdminTLSRootCAPool() *x509.CertPool { return nil }
`
	require.NoError(t, os.WriteFile(filepath.Join(psDir, "server_test.go"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessNoLocalPSIDServerStubs.FindViolations(dir)

	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Equal(t, "stubAdminServer", violations[0].StructName)
	require.Equal(t, "IAdminServer", violations[0].Interface)
}

func TestFindViolations_FrameworkExcluded(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	fwDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.FrameworkProductName, "service", "server")
	require.NoError(t, os.MkdirAll(fwDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Framework files are excluded — no violation should fire even with stub methods.
	content := `package server_test
import "context"
type mockPublicServer struct{}
func (m *mockPublicServer) Start(_ context.Context) error { return nil }
func (m *mockPublicServer) Shutdown(_ context.Context) error { return nil }
func (m *mockPublicServer) ActualPort() int { return 8443 }
func (m *mockPublicServer) PublicBaseURL() string { return "" }
`
	require.NoError(t, os.WriteFile(filepath.Join(fwDir, "application_test.go"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessNoLocalPSIDServerStubs.FindViolations(dir)

	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolations_ToolsExcluded(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	toolsDir := filepath.Join(dir, "internal", "apps", "tools", "cicd_lint")
	require.NoError(t, os.MkdirAll(toolsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	content := `package cicd_test
import "context"
type stubPublicServer struct{}
func (s *stubPublicServer) Start(_ context.Context) error { return nil }
func (s *stubPublicServer) Shutdown(_ context.Context) error { return nil }
func (s *stubPublicServer) ActualPort() int { return 8443 }
func (s *stubPublicServer) PublicBaseURL() string { return "" }
`
	require.NoError(t, os.WriteFile(filepath.Join(toolsDir, "foo_test.go"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessNoLocalPSIDServerStubs.FindViolations(dir)

	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolations_PartialMethodSet_NotFlagged(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	psDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServicePKICA, "server")
	require.NoError(t, os.MkdirAll(psDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Only 3 of the 4 IPublicServer methods — should not be flagged.
	content := `package server
import "context"
type partialStub struct{}
func (s *partialStub) Start(_ context.Context) error { return nil }
func (s *partialStub) Shutdown(_ context.Context) error { return nil }
func (s *partialStub) ActualPort() int { return 8443 }
`
	require.NoError(t, os.WriteFile(filepath.Join(psDir, "server_test.go"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessNoLocalPSIDServerStubs.FindViolations(dir)

	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckInDir_ReturnsError_WhenViolationsFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	psDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, "server")
	require.NoError(t, os.MkdirAll(psDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	require.NoError(t, os.WriteFile(filepath.Join(psDir, "server_test.go"), []byte(stubPublicServerGoSrc), cryptoutilSharedMagic.FilePermissionsDefault))

	logger := cryptoutilCmdCicdCommon.NewLogger("test-no-local-ps-id-server-stubs")
	err := lintFitnessNoLocalPSIDServerStubs.CheckInDir(logger, dir)

	require.Error(t, err)
	require.Contains(t, err.Error(), "local PS-ID server stub")
}

func TestCheckInDir_ReturnsNil_WhenClean(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logger := cryptoutilCmdCicdCommon.NewLogger("test-no-local-ps-id-server-stubs")
	err := lintFitnessNoLocalPSIDServerStubs.CheckInDir(logger, dir)

	require.NoError(t, err)
}
