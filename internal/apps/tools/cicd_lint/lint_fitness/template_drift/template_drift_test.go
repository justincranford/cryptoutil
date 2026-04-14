// Copyright (c) 2025 Justin Cranford

package template_drift

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// testBase64String is a 47-character base64 string used across multiple comparison tests.
const testBase64String = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuv"

// projectRoot returns the path to the project root for integration tests.
func projectRoot() string {
	return filepath.Join("..", "..", "..", "..", "..", "..")
}

func TestCheckTemplateCompliance(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-compliance")
	err := checkTemplateComplianceInDir(logger, projectRoot(), defaultComplianceFn)
	require.NoError(t, err)
}

func TestLoadTemplatesDir_Happy(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create templates dir structure.
	templDir := filepath.Join(tmpDir, templatesRelPath, "deployments", "__PS_ID__")
	require.NoError(t, os.MkdirAll(templDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(templDir, "Dockerfile"), []byte("FROM __PS_ID__"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(templDir, ".gitkeep"), []byte(""), cryptoutilSharedMagic.CacheFilePermissions))

	templates, err := LoadTemplatesDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, templates, 1)
	require.Equal(t, "FROM __PS_ID__", templates["deployments/__PS_ID__/Dockerfile"])
}

func TestLoadTemplatesDir_NonExistentRoot(t *testing.T) {
	t.Parallel()

	_, err := LoadTemplatesDir(filepath.Join(t.TempDir(), "nonexistent"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "templates directory not found")
}

func TestBuildExpectedFS_PSIDExpansion(t *testing.T) {
	t.Parallel()

	templates := map[string]string{
		"deployments/__PS_ID__/Dockerfile": "FROM __PS_ID__:latest",
	}

	expected := BuildExpectedFS(templates)
	// One expanded entry per PS-ID.
	require.Len(t, expected, len(cryptoutilRegistry.AllProductServices()))

	// Spot-check one expansion.
	content, ok := expected["deployments/"+cryptoutilSharedMagic.OTLPServiceSMKMS+"/Dockerfile"]
	require.True(t, ok)
	require.Contains(t, content, cryptoutilSharedMagic.OTLPServiceSMKMS)
	require.NotContains(t, content, "__PS_ID__")
}

func TestBuildExpectedFS_ProductExpansion(t *testing.T) {
	t.Parallel()

	templates := map[string]string{
		"deployments/__PRODUCT__/compose.yml": "product: __PRODUCT__",
	}

	expected := BuildExpectedFS(templates)
	// One expanded entry per product.
	require.Len(t, expected, len(cryptoutilRegistry.AllProducts()))

	content, ok := expected["deployments/"+cryptoutilSharedMagic.SMProductName+"/compose.yml"]
	require.True(t, ok)
	require.Contains(t, content, cryptoutilSharedMagic.SMProductName)
}

func TestBuildExpectedFS_SuiteExpansion(t *testing.T) {
	t.Parallel()

	templates := map[string]string{
		"deployments/__SUITE__/Dockerfile": "FROM __SUITE__:latest",
	}

	expected := BuildExpectedFS(templates)
	// 1 suite → 1 entry.
	require.Len(t, expected, 1)

	content, ok := expected["deployments/"+cryptoutilSharedMagic.DefaultOTLPServiceDefault+"/Dockerfile"]
	require.True(t, ok)
	require.Contains(t, content, cryptoutilSharedMagic.DefaultOTLPServiceDefault)
}

func TestBuildExpectedFS_StaticPath(t *testing.T) {
	t.Parallel()

	templates := map[string]string{
		"deployments/shared-telemetry/compose.yml": "suite: __SUITE__",
	}

	expected := BuildExpectedFS(templates)
	require.Len(t, expected, 1)

	content, ok := expected["deployments/shared-telemetry/compose.yml"]
	require.True(t, ok)
	require.Contains(t, content, cryptoutilSharedMagic.DefaultOTLPServiceDefault)
	require.NotContains(t, content, "__SUITE__")
}

func TestBuildExpectedFS_ContentSubstitution(t *testing.T) {
	t.Parallel()

	templates := map[string]string{
		"deployments/__PS_ID__/config/__PS_ID__-app.yml": "psid: __PS_ID__\nupper: __PS_ID_UPPER__",
	}

	expected := BuildExpectedFS(templates)

	content, ok := expected["deployments/jose-ja/config/jose-ja-app.yml"]
	require.True(t, ok)
	require.Contains(t, content, "psid: jose-ja")
	require.Contains(t, content, "upper: JOSE-JA")
}

func TestBuildExpectedFS_SecretsExpansion(t *testing.T) {
	t.Parallel()

	templates := map[string]string{
		"deployments/__PS_ID__/secrets/password.secret": "__PS_ID__-password-BASE64_CHAR43",
	}

	expected := BuildExpectedFS(templates)

	content, ok := expected["deployments/sm-kms/secrets/password.secret"]
	require.True(t, ok)
	// BASE64_CHAR43 should NOT be substituted — it's a comparison placeholder.
	require.Contains(t, content, "BASE64_CHAR43")
	require.Contains(t, content, "sm-kms-password")
}

func TestCompareExpectedFS_AllMatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a file on disk.
	filePath := filepath.Join(tmpDir, "deployments", "test", "Dockerfile")
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filePath, []byte("FROM scratch"), cryptoutilSharedMagic.CacheFilePermissions))

	expected := map[string]string{
		"deployments/test/Dockerfile": "FROM scratch",
	}

	err := CompareExpectedFS(expected, tmpDir)
	require.NoError(t, err)
}

func TestCompareExpectedFS_ContentMismatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "deployments", "test", "Dockerfile")
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filePath, []byte("FROM wrong"), cryptoutilSharedMagic.CacheFilePermissions))

	expected := map[string]string{
		"deployments/test/Dockerfile": "FROM scratch",
	}

	err := CompareExpectedFS(expected, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "content drift")
}

func TestCompareExpectedFS_MissingFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	expected := map[string]string{
		"deployments/test/Dockerfile": "FROM scratch",
	}

	err := CompareExpectedFS(expected, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "deployments/test/Dockerfile")
}

func TestCompareExpectedFS_SecretsPlaceholder(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// The actual file has a real base64 value instead of BASE64_CHAR43.
	filePath := filepath.Join(tmpDir, "deployments", "test", "secrets", "password.secret")
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), cryptoutilSharedMagic.CICDTempDirPermissions))

	realB64 := testBase64String // 47 chars
	require.NoError(t, os.WriteFile(filePath, []byte("test-password-"+realB64), cryptoutilSharedMagic.CacheFilePermissions))

	expected := map[string]string{
		"deployments/test/secrets/password.secret": "test-password-BASE64_CHAR43",
	}

	err := CompareExpectedFS(expected, tmpDir)
	require.NoError(t, err)
}

func TestChooseComparison_ExactMatch(t *testing.T) {
	t.Parallel()

	diff := chooseComparison("deployments/sm-kms/Dockerfile", "FROM test", "FROM test")
	require.Empty(t, diff)
}

func TestChooseComparison_PKICASuperset(t *testing.T) {
	t.Parallel()

	diff := chooseComparison("deployments/pki-ca/compose.yml", "a\nb", "a\nextra\nb")
	require.Empty(t, diff)
}

func TestChooseComparison_PKICAConfigPrefix(t *testing.T) {
	t.Parallel()

	diff := chooseComparison(
		"deployments/pki-ca/config/pki-ca-app-framework-common.yml",
		"a\nb",
		"a\nb\nextra",
	)
	require.Empty(t, diff)
}

func TestChooseComparison_StandaloneConfigPrefix(t *testing.T) {
	t.Parallel()

	diff := chooseComparison("configs/sm-kms/sm-kms-framework.yml", "a\nb", "a\nb\nextra")
	require.Empty(t, diff)
}

func TestChooseComparison_Base64Placeholder(t *testing.T) {
	t.Parallel()

	realB64 := testBase64String // 47 chars
	diff := chooseComparison("deployments/sm-kms/secrets/password.secret",
		"prefix-BASE64_CHAR43", "prefix-"+realB64)
	require.Empty(t, diff)
}

func TestBuildParams(t *testing.T) {
	t.Parallel()

	params := buildParams(cryptoutilSharedMagic.OTLPServiceJoseJA)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceJoseJA, params["__PS_ID__"])
	require.Equal(t, "JOSE-JA", params["__PS_ID_UPPER__"])
	require.Equal(t, cryptoutilSharedMagic.DefaultOTLPServiceDefault, params["__SUITE__"])
	require.Equal(t, "1.26.1", params["__GO_VERSION__"])
	require.Equal(t, "65532", params["__CONTAINER_UID__"])
	require.Equal(t, "65532", params["__CONTAINER_GID__"])
	require.NotEmpty(t, params["__PRODUCT_DISPLAY_NAME__"])
	require.NotEmpty(t, params["__SERVICE_DISPLAY_NAME__"])
	require.NotEmpty(t, params["__SERVICE_APP_PORT_BASE__"])
}

func TestBuildInstanceParams(t *testing.T) {
	t.Parallel()

	params := buildInstanceParams(cryptoutilSharedMagic.OTLPServiceSMKMS, 1, int(cryptoutilSharedMagic.DefaultPublicPortCryptoutil))
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceSMKMS, params["__PS_ID__"])
	require.Equal(t, "1", params["__INSTANCE_NUM__"])
	require.Equal(t, "8000", params["__SERVICE_APP_PORT__"])
	require.Equal(t, "SM-KMS", params["__PS_ID_UPPER__"])
}

func TestBuildProductParams(t *testing.T) {
	t.Parallel()

	params := buildProductParams("sm")
	require.Equal(t, "sm", params["__PRODUCT__"])
	require.Equal(t, "SM", params["__PRODUCT_UPPER__"])
	require.NotEmpty(t, params["__PRODUCT_INCLUDE_LIST__"])
	require.NotEmpty(t, params["__PRODUCT_SERVICE_OVERRIDES__"])
	require.NotEmpty(t, params["__PRODUCT_INIT_PS_ID__"])
}

func TestBuildSuiteParams(t *testing.T) {
	t.Parallel()

	params := buildSuiteParams(cryptoutilSharedMagic.DefaultOTLPServiceDefault)
	require.Equal(t, cryptoutilSharedMagic.DefaultOTLPServiceDefault, params["__SUITE__"])
	require.NotEmpty(t, params["__SUITE_INCLUDE_LIST__"])
	require.NotEmpty(t, params["__SUITE_SERVICE_OVERRIDES__"])
	require.NotEmpty(t, params["__SUITE_INIT_PS_ID__"])
}

func TestBuildStaticParams(t *testing.T) {
	t.Parallel()

	params := buildStaticParams()
	require.Equal(t, cryptoutilSharedMagic.DefaultOTLPServiceDefault, params["__SUITE__"])
	require.Equal(t, cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault, params["__IMAGE_TAG__"])
}

func TestSubstituteParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		params map[string]string
		want   string
	}{
		{
			name:   "single substitution",
			input:  "hello __NAME__",
			params: map[string]string{"__NAME__": "world"},
			want:   "hello world",
		},
		{
			name:   "multiple substitutions",
			input:  "__A__ and __B__",
			params: map[string]string{"__A__": "x", "__B__": "y"},
			want:   "x and y",
		},
		{
			name:   "no placeholders",
			input:  "no change",
			params: map[string]string{"__X__": "y"},
			want:   "no change",
		},
		{
			name:   "empty params",
			input:  "__A__",
			params: map[string]string{},
			want:   "__A__",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := substituteParams(tt.input, tt.params)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBuildProductIncludeList(t *testing.T) {
	t.Parallel()

	list := buildProductIncludeList([]string{cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.OTLPServiceSMIM})
	require.Contains(t, list, "include:")
	require.Contains(t, list, "- path: ../"+cryptoutilSharedMagic.OTLPServiceSMKMS+"/compose.yml")
	require.Contains(t, list, "- path: ../"+cryptoutilSharedMagic.OTLPServiceSMIM+"/compose.yml")
}

func TestBuildProductServiceOverrides(t *testing.T) {
	t.Parallel()

	overrides := buildProductServiceOverrides(cryptoutilSharedMagic.SMProductName, []string{cryptoutilSharedMagic.OTLPServiceSMKMS})
	require.Contains(t, overrides, "sm-kms-app-sqlite-1:")
	require.Contains(t, overrides, "ports: !override")
}

func TestBuildSuiteIncludeList(t *testing.T) {
	t.Parallel()

	products := []cryptoutilRegistry.Product{
		{ID: cryptoutilSharedMagic.SMProductName},
		{ID: cryptoutilSharedMagic.JoseProductName},
	}

	list := buildSuiteIncludeList(products)
	require.Contains(t, list, "include:")
	require.Contains(t, list, "- path: ../"+cryptoutilSharedMagic.SMProductName+"/compose.yml")
	require.Contains(t, list, "- path: ../"+cryptoutilSharedMagic.JoseProductName+"/compose.yml")
}

func TestBuildSuiteServiceOverrides(t *testing.T) {
	t.Parallel()

	overrides := buildSuiteServiceOverrides()
	require.Contains(t, overrides, "sm-kms-app-sqlite-1:")
	require.Contains(t, overrides, "!override")
}

func TestBuildProductPSIDListDisplay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		productID string
		psIDs     []string
		want      string
	}{
		{
			name:      "two services",
			productID: cryptoutilSharedMagic.SMProductName,
			psIDs:     []string{cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.OTLPServiceSMIM},
			want:      "SM product (2 services: kms, im)",
		},
		{
			name:      "single service",
			productID: cryptoutilSharedMagic.JoseProductName,
			psIDs:     []string{cryptoutilSharedMagic.OTLPServiceJoseJA},
			want:      "JOSE product (1 service: ja)",
		},
		{
			name:      "empty",
			productID: "test",
			psIDs:     []string{},
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildProductPSIDListDisplay(tt.productID, tt.psIDs)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPluralS(t *testing.T) {
	t.Parallel()

	require.Equal(t, "", pluralS(1))
	require.Equal(t, "s", pluralS(0))
	require.Equal(t, "s", pluralS(2))
}

// TestCheckTemplateCompliance_IntegrationWithProjectRoot runs against the actual project.
func TestCheckTemplateCompliance_IntegrationWithProjectRoot(t *testing.T) {
	t.Parallel()

	templates, err := LoadTemplatesDir(projectRoot())
	require.NoError(t, err)
	require.NotEmpty(t, templates)

	expected := BuildExpectedFS(templates)
	require.NotEmpty(t, expected)
	// Should expand to many files (>300).
	require.Greater(t, len(expected), 300)
}

// fakeDirEntry is a minimal fs.DirEntry implementation for seam-injection tests.
type fakeDirEntry struct {
	name  string
	isDir bool
}

func (f *fakeDirEntry) Name() string               { return f.name }
func (f *fakeDirEntry) IsDir() bool                { return f.isDir }
func (f *fakeDirEntry) Type() fs.FileMode          { return 0 }
func (f *fakeDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// TestLoadTemplatesDirFn_WalkCallbackError exercises the path where the WalkDirFunc
// receives a non-nil error argument from the OS (e.g., permission denied on a dir entry).
func TestLoadTemplatesDirFn_WalkCallbackError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, templatesRelPath), cryptoutilSharedMagic.CICDTempDirPermissions))

	injectedErr := errors.New("permission denied")

	_, err := loadTemplatesDirFn(tmpDir, func(_ string, fn fs.WalkDirFunc) error {
		// Simulate the OS passing a non-nil err to the WalkDirFunc callback.
		return fn(tmpDir, &fakeDirEntry{name: "locked-dir", isDir: true}, injectedErr)
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "walk templates directory")
}

// TestLoadTemplatesDirFn_WalkError exercises the outer walk error path where
// filepath.WalkDir itself returns an error (not via the callback).
func TestLoadTemplatesDirFn_WalkError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, templatesRelPath), cryptoutilSharedMagic.CICDTempDirPermissions))

	_, err := loadTemplatesDirFn(tmpDir, func(_ string, _ fs.WalkDirFunc) error {
		return errors.New("walk failed")
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "walk templates directory")
}

// TestLoadTemplatesDirFn_ReadFileError exercises the os.ReadFile error path.
// We inject a fake DirEntry reporting a regular file but point it at a directory
// on disk; os.ReadFile on a directory fails on all platforms.
func TestLoadTemplatesDirFn_ReadFileError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, templatesRelPath)
	require.NoError(t, os.MkdirAll(templatesDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	// Create a sub-directory whose name ends in .yml — os.ReadFile on a directory is an error.
	fakeFilePath := filepath.Join(templatesDir, "config.yml")
	require.NoError(t, os.MkdirAll(fakeFilePath, cryptoutilSharedMagic.CICDTempDirPermissions))

	_, err := loadTemplatesDirFn(tmpDir, func(_ string, fn fs.WalkDirFunc) error {
		// Fake: claim the directory-with-.yml-name is a regular file.
		return fn(fakeFilePath, &fakeDirEntry{name: "config.yml", isDir: false}, nil)
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "walk templates directory")
}

// TestCompareBase64Placeholder_TrailingTooShort verifies that a trailing BASE64_CHAR43
// segment whose actual value is shorter than 43 characters reports a "too short" error.
func TestCompareBase64Placeholder_TrailingTooShort(t *testing.T) {
	t.Parallel()

	// "SHORT" is only 5 characters — well below the 43-char minimum.
	diff := compareBase64Placeholder("prefix-BASE64_CHAR43", "prefix-SHORT")
	require.NotEmpty(t, diff)
	require.Contains(t, diff, "too short")
}
