// Copyright (c) 2025 Justin Cranford

package cross_service_import_isolation

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func newTestLogger() *cryptoutilCmdCicdCommon.Logger {
	return cryptoutilCmdCicdCommon.NewLogger("test")
}

// mkServiceDir creates internal/apps/<product>/<service>/server/ in root.
func mkServiceDir(t *testing.T, root, product, service string) {
	t.Helper()

	dir := filepath.Join(root, "internal", "apps", product, service, "server")
	require.NoError(t, os.MkdirAll(dir, cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "server.go"),
		[]byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
}

// goFileWithImports writes a valid Go file with properly tab-indented block imports.
// Each element of imports should be the full import text without leading tab
// (e.g. `"fmt"` or `myAlias "some/package"`).
func goFileWithImports(t *testing.T, path, pkg string, imports []string) {
	t.Helper()

	var body string

	body += "package " + pkg + "\n\n"

	body += "import (\n"
	for _, imp := range imports {
		body += "\t" + imp + "\n"
	}

	body += ")\n"

	require.NoError(t, os.MkdirAll(filepath.Dir(path), cryptoutilSharedMagic.DirPermissions))
	require.NoError(t, os.WriteFile(path, []byte(body), cryptoutilSharedMagic.CacheFilePermissions))
}

func TestCheckInDir_NoServices_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "internal", "apps"), cryptoutilSharedMagic.DirPermissions))
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_CleanImports_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	mkServiceDir(t, tmp, "sm", "im")
	// Imports of shared/ and template/ are always allowed.
	goFileWithImports(t,
		filepath.Join(tmp, "internal", "apps", "sm", "im", "handler.go"),
		"im",
		[]string{
			`"fmt"`,
			`"cryptoutil/internal/shared/magic"`,
			`"cryptoutil/internal/apps/framework/service/server"`,
		},
	)
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_SameProductImport_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	mkServiceDir(t, tmp, cryptoutilSharedMagic.IdentityProductName, cryptoutilSharedMagic.IDPServiceName)
	mkServiceDir(t, tmp, cryptoutilSharedMagic.IdentityProductName, cryptoutilSharedMagic.AuthzServiceName)
	goFileWithImports(t,
		filepath.Join(tmp, "internal", "apps", cryptoutilSharedMagic.IdentityProductName, cryptoutilSharedMagic.IDPServiceName, "handler.go"),
		cryptoutilSharedMagic.IDPServiceName,
		[]string{`"cryptoutil/internal/apps/identity/authz/clientauth"`},
	)
	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_CrossProductImport_Fails(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	mkServiceDir(t, tmp, "sm", "im")
	mkServiceDir(t, tmp, cryptoutilSharedMagic.PKIProductName, "ca")
	goFileWithImports(t,
		filepath.Join(tmp, "internal", "apps", "sm", "im", "handler.go"),
		"im",
		[]string{`"cryptoutil/internal/apps/pki/ca/somepkg"`},
	)
	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cross-service import isolation violation")
}

func TestCheckInDir_SkipArchivedProduct_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	archivedDir := filepath.Join(tmp, "internal", "apps", "_archived", "svc", "server")
	require.NoError(t, os.MkdirAll(archivedDir, cryptoutilSharedMagic.DirPermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_SkipCicdProduct_Passes(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	cicdDir := filepath.Join(tmp, "internal", "apps", "cicd", "linter", "server")
	require.NoError(t, os.MkdirAll(cicdDir, cryptoutilSharedMagic.DirPermissions))

	err := CheckInDir(newTestLogger(), tmp)
	require.NoError(t, err)
}

func TestCheckInDir_BadRootDir_Error(t *testing.T) {
	t.Parallel()

	err := CheckInDir(newTestLogger(), "/nonexistent/path/xyz")
	require.Error(t, err)
}

func TestCollectServices_EmptyAppsDir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	appsDir := filepath.Join(tmp, "internal", "apps")
	require.NoError(t, os.MkdirAll(appsDir, cryptoutilSharedMagic.DirPermissions))
	services, err := collectServices(appsDir)
	require.NoError(t, err)
	require.Empty(t, services)
}

func TestIsViolation_AllowedImports(t *testing.T) {
	t.Parallel()

	self := serviceRef{product: "sm", service: "im"}
	allServices := []serviceRef{{product: cryptoutilSharedMagic.PKIProductName, service: "ca"}}

	tests := []struct {
		name       string
		importPath string
		wantViol   bool
	}{
		{"non-apps import", "github.com/some/lib", false},
		{"template import", "cryptoutil/internal/apps/framework/service/server", false},
		{"cicd import", "cryptoutil/internal/apps/cicd/common", false},
		{"same-product import", "cryptoutil/internal/apps/sm/kms/something", false},
		{"cross-product violation", "cryptoutil/internal/apps/pki/ca/something", true},
		{"non-service cross-product", "cryptoutil/internal/apps/pki/shared/lib", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isViolation(tc.importPath, self, allServices)
			require.Equal(t, tc.wantViol, got)
		})
	}
}

func TestExtractImports_ValidFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	goFile := filepath.Join(tmp, "valid.go")
	// Block import with proper tab indentation (required by importLinePattern ^\s+).
	content := "package foo\n\nimport (\n" +
		"\t\"fmt\"\n" +
		"\tmyAlias \"some/package\"\n" +
		")\n\nfunc main() {}\n"
	require.NoError(t, os.WriteFile(goFile, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
	imports, err := extractImports(goFile)
	require.NoError(t, err)
	require.Contains(t, imports, "fmt")
	require.Contains(t, imports, "some/package")
}

func TestExtractImports_NonexistentFile_Error(t *testing.T) {
	t.Parallel()

	_, err := extractImports("/nonexistent/file.go")
	require.Error(t, err)
}

func TestExtractImports_EmptyFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	goFile := filepath.Join(tmp, "empty.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package empty\n"), cryptoutilSharedMagic.CacheFilePermissions))
	imports, err := extractImports(goFile)
	require.NoError(t, err)
	require.Empty(t, imports)
}

func TestIsViolation_NonAppsImport_ReturnsFalse(t *testing.T) {
	t.Parallel()

	self := serviceRef{product: "sm", service: "im"}
	result := isViolation("github.com/some/external/lib", self, nil)
	require.False(t, result)
}

func TestIsViolation_InsufficientPathParts_ReturnsFalse(t *testing.T) {
	t.Parallel()

	self := serviceRef{product: "sm", service: "im"}
	// Only one part after apps/ prefix - not enough to identify service.
	result := isViolation("cryptoutil/internal/apps/pki", self, nil)
	require.False(t, result)
}

func TestIsViolation_CrossProductNotInServices_ReturnsFalse(t *testing.T) {
	t.Parallel()

	self := serviceRef{product: "sm", service: "im"}
	// pki/ca is not in allServices list, so not a service-to-service violation.
	result := isViolation("cryptoutil/internal/apps/pki/ca/something", self, []serviceRef{})
	require.False(t, result)
}

func TestCheckInDir_SkeletonProduct_Skipped(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	mkServiceDir(t, tmp, cryptoutilSharedMagic.SkeletonProductName, cryptoutilSharedMagic.SkeletonTemplateServiceName)
	// skeleton is excluded from collectServices.
	mkServiceDir(t, tmp, cryptoutilSharedMagic.PKIProductName, "ca")
	goFileWithImports(t,
		filepath.Join(tmp, "internal", "apps", cryptoutilSharedMagic.SkeletonProductName, cryptoutilSharedMagic.SkeletonTemplateServiceName, "handler.go"),
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
		[]string{`"cryptoutil/internal/apps/pki/ca/somepkg"`},
	)
	err := CheckInDir(newTestLogger(), tmp)
	// skeleton is excluded so no violation detected.
	require.NoError(t, err)
}

// Sequential: modifies package-level crossServiceWalkFn seam.
func TestCheckInDir_WalkError(t *testing.T) {
	orig := crossServiceWalkFn

	t.Cleanup(func() { crossServiceWalkFn = orig })

	crossServiceWalkFn = func(_ string, _ filepath.WalkFunc) error {
		return fmt.Errorf("injected walk error")
	}

	tmp := t.TempDir()
	mkServiceDir(t, tmp, "sm", "im")

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to scan service")
}

// Sequential: modifies package-level crossServiceWalkFn seam.
func TestCheckInDir_WalkCallbackError(t *testing.T) {
	orig := crossServiceWalkFn

	t.Cleanup(func() { crossServiceWalkFn = orig })

	crossServiceWalkFn = func(_ string, fn filepath.WalkFunc) error {
		return fn("bad/path", nil, fmt.Errorf("injected callback error"))
	}

	tmp := t.TempDir()
	mkServiceDir(t, tmp, "sm", "im")

	err := CheckInDir(newTestLogger(), tmp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to scan service")
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

	logger := cryptoutilCmdCicdCommon.NewLogger("test-cross-service-import-isolation")

	err = Check(logger)
	require.NoError(t, err)
}
