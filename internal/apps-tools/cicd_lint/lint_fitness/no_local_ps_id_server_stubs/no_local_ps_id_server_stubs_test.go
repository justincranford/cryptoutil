// Copyright (c) 2025-2026 Justin Cranford.
package no_local_ps_id_server_stubs_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	lintFitnessNoLocalPSIDServerStubs "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/no_local_ps_id_server_stubs"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testServerTypeName = "Server"

const stubPublicServerGoSrc = `package server
import "context"
type stubPublicServer struct{}
func (s *stubPublicServer) Start(_ context.Context) error { return nil }
func (s *stubPublicServer) Shutdown(_ context.Context) error { return nil }
func (s *stubPublicServer) ActualPort() int { return 8443 }
func (s *stubPublicServer) PublicBaseURL() string { return "" }
`

const stubAdminServerGoSrc = `package server
import ("context"; "crypto/x509")
type stubAdminServer struct{}
func (s *stubAdminServer) Start(_ context.Context) error { return nil }
func (s *stubAdminServer) Shutdown(_ context.Context) error { return nil }
func (s *stubAdminServer) ActualPort() int { return 9090 }
func (s *stubAdminServer) SetReady(_ bool) {}
func (s *stubAdminServer) AdminBaseURL() string { return "https://localhost:9090" }
func (s *stubAdminServer) AdminTLSRootCAPool() *x509.CertPool { return nil }
`

// Sequential: changes the process working directory to exercise Check.
func TestCheck_WorkspaceRoot(t *testing.T) {
	original, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(t.TempDir()))
	defer func() {
		require.NoError(t, os.Chdir(original))
	}()

	err = lintFitnessNoLocalPSIDServerStubs.Check(cryptoutilCmdCicdCommon.NewLogger("test-check-workspace-root"))
	require.NoError(t, err)
}

func TestFindViolations_NoViolations_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	violations, err := lintFitnessNoLocalPSIDServerStubs.FindViolations(dir)

	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolations_ExcludedDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for _, excludedDir := range []string{
		cryptoutilSharedMagic.CICDExcludeDirVendor,
		cryptoutilSharedMagic.CICDExcludeDirGit,
	} {
		path := filepath.Join(dir, excludedDir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS)
		require.NoError(t, os.MkdirAll(path, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(path, "server_test.go"), []byte(stubPublicServerGoSrc), cryptoutilSharedMagic.FilePermissionsDefault))
	}

	violations, err := lintFitnessNoLocalPSIDServerStubs.FindViolations(dir)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolations_NoViolations_NonTestFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	psDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.CMD_SERVER)
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
	psDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.CMD_SERVER)
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
	psDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceIdentityAuthz, cryptoutilSharedMagic.CMD_SERVER)
	require.NoError(t, os.MkdirAll(psDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	require.NoError(t, os.WriteFile(filepath.Join(psDir, "server_test.go"), []byte(stubAdminServerGoSrc), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessNoLocalPSIDServerStubs.FindViolations(dir)

	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Equal(t, "stubAdminServer", violations[0].StructName)
	require.Equal(t, "IAdminServer", violations[0].Interface)
}

func TestFindViolations_FrameworkExcluded(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	fwDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.FrameworkProductName, "service", cryptoutilSharedMagic.CMD_SERVER)
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
	psDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceIdentityAuthz, cryptoutilSharedMagic.CMD_SERVER)
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
	psDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.CMD_SERVER)
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

func TestCheckInDir_ReturnsError_WhenWalkFails(t *testing.T) {
	t.Parallel()

	err := lintFitnessNoLocalPSIDServerStubs.CheckInDir(
		cryptoutilCmdCicdCommon.NewLogger("test-no-local-ps-id-server-stubs-walk-error"),
		filepath.Join(t.TempDir(), "missing-root"),
	)
	require.ErrorContains(t, err, "failed to walk directory")
}

func TestFindViolations_WalkErrors(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	tests := []struct {
		name   string
		walkFn func(string, fs.WalkDirFunc) error
	}{
		{
			name: "walk function error",
			walkFn: func(string, fs.WalkDirFunc) error {
				return os.ErrPermission
			},
		},
		{
			name: "callback error",
			walkFn: func(path string, callback fs.WalkDirFunc) error {
				return callback(path, nil, os.ErrPermission)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := lintFitnessNoLocalPSIDServerStubs.ExportedFindViolationsWithWalkDir(rootDir, tc.walkFn)
			require.ErrorIs(t, err, os.ErrPermission)
			require.ErrorContains(t, err, "walking")
		})
	}
}

func TestFindViolations_CheckFileError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	psDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.CMD_SERVER)
	require.NoError(t, os.MkdirAll(psDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(psDir, "server_test.go"), []byte("package server"), cryptoutilSharedMagic.FilePermissionsDefault))

	_, err := lintFitnessNoLocalPSIDServerStubs.ExportedFindViolationsWithFunctions(
		dir,
		filepath.WalkDir,
		func(string) ([]lintFitnessNoLocalPSIDServerStubs.Violation, error) {
			return nil, fs.ErrPermission
		},
	)

	require.ErrorIs(t, err, fs.ErrPermission)
}

func TestPSIDTestFileClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "missing internal apps prefix", path: "other/internal/services/sm-kms/server_test.go"},
		{name: "missing product service directory", path: "internal/apps/server_test.go"},
		{name: "framework excluded", path: "internal/apps/framework/service/server_test.go"},
		{name: "tools excluded", path: "internal/apps/tools/cicd_lint/server_test.go"},
		{name: "skeleton excluded", path: filepath.Join("internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName, "server_test.go")},
		{name: "product service included", path: "internal/apps/sm-kms/server/server_test.go", want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, lintFitnessNoLocalPSIDServerStubs.ExportedIsPSIDTestFile(tc.path))
		})
	}
}

func TestCheckFile_OpenAndParseErrors(t *testing.T) {
	t.Parallel()

	missingPath := filepath.Join(t.TempDir(), "missing_test.go")
	_, err := lintFitnessNoLocalPSIDServerStubs.ExportedCheckFile(missingPath)
	require.ErrorContains(t, err, "reading")

	invalidPath := filepath.Join(t.TempDir(), "invalid_test.go")
	require.NoError(t, os.WriteFile(invalidPath, []byte("package broken\nfunc ("), cryptoutilSharedMagic.FilePermissionsDefault))
	violations, err := lintFitnessNoLocalPSIDServerStubs.ExportedCheckFile(invalidPath)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestReceiverTypeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{name: "identifier", expr: &ast.Ident{Name: testServerTypeName}, want: testServerTypeName},
		{name: "pointer", expr: &ast.StarExpr{X: &ast.Ident{Name: testServerTypeName}}, want: testServerTypeName},
		{name: "unsupported expression", expr: &ast.ArrayType{}, want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, lintFitnessNoLocalPSIDServerStubs.ExportedReceiverTypeName(tc.expr))
		})
	}
}

func TestCollectReceiverMethods_HandlesFunctionsAndInvalidReceivers(t *testing.T) {
	t.Parallel()

	source := `package example
func helper() {}
type ` + testServerTypeName + ` struct{}
func (s *` + testServerTypeName + `) Serve() {}
func (s *example.` + testServerTypeName + `) Invalid() {}
`
	file, err := parser.ParseFile(token.NewFileSet(), "server.go", source, 0)
	require.NoError(t, err)

	file.Decls = append(file.Decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{{
				Type: &ast.SelectorExpr{X: &ast.Ident{Name: "example"}, Sel: &ast.Ident{Name: testServerTypeName}},
			}},
		},
		Name: &ast.Ident{Name: "Invalid"},
	})

	methods := lintFitnessNoLocalPSIDServerStubs.ExportedCollectReceiverMethods(file)
	require.Equal(t, map[string]bool{"Serve": true}, methods[testServerTypeName])
	require.NotContains(t, methods, "")
}

func TestFindTypeLine_MissingType(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "types.go", "package example\nimport \"fmt\"\nfunc helper() { fmt.Println(\"ok\") }\ntype Existing struct{}\n", 0)
	require.NoError(t, err)

	require.Zero(t, lintFitnessNoLocalPSIDServerStubs.ExportedFindTypeLine(fset, file, "Missing"))
}
