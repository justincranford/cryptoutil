// Copyright (c) 2025-2026 Justin Cranford.
package template_drift

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestBuildParams(t *testing.T) {
	t.Parallel()

	params := buildParams(cryptoutilSharedMagic.OTLPServiceJoseJA)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceJoseJA, params[cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID])
	require.Equal(t, "JOSE-JA", params["__PS_ID_UPPER__"])
	require.Equal(t, cryptoutilSharedMagic.DefaultOTLPServiceDefault, params[cryptoutilSharedMagic.CICDTemplateExpansionKeySuite])
	require.Equal(t, cryptoutilSharedMagic.CICDTemplateGoVersion, params["__GO_VERSION__"])
	require.Equal(t, cryptoutilSharedMagic.CICDTemplateContainerUID, params["__CONTAINER_UID__"])
	require.Equal(t, cryptoutilSharedMagic.CICDTemplateContainerGID, params["__CONTAINER_GID__"])
	require.NotEmpty(t, params["__PRODUCT_DISPLAY_NAME__"])
	require.NotEmpty(t, params["__PS_DISPLAY_NAME__"])
	require.NotEmpty(t, params["__PS_PUBLIC_PORT_BASE__"])
}

func TestBuildInstanceParams(t *testing.T) {
	t.Parallel()

	params := buildInstanceParams(cryptoutilSharedMagic.OTLPServiceSMKMS, 1, int(cryptoutilSharedMagic.DefaultPublicPortCryptoutil))
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceSMKMS, params[cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID])
	require.Equal(t, "1", params["__INSTANCE_NUM__"])
	require.Equal(t, "8000", params["__PS_PUBLIC_PORT__"])
	require.Equal(t, "SM-KMS", params["__PS_ID_UPPER__"])
}

func TestBuildProductParams(t *testing.T) {
	t.Parallel()

	params := buildProductParams("sm")
	require.Equal(t, "sm", params[cryptoutilSharedMagic.CICDTemplateExpansionKeyProduct])
	require.Equal(t, "SM", params["__PRODUCT_UPPER__"])
	require.NotEmpty(t, params["__PRODUCT_INCLUDE_LIST__"])
	require.NotEmpty(t, params["__PRODUCT_SERVICE_OVERRIDES__"])
	require.NotEmpty(t, params["__PRODUCT_INIT_PS_ID__"])
}

func TestBuildSuiteParams(t *testing.T) {
	t.Parallel()

	params := buildSuiteParams(cryptoutilSharedMagic.DefaultOTLPServiceDefault)
	require.Equal(t, cryptoutilSharedMagic.DefaultOTLPServiceDefault, params[cryptoutilSharedMagic.CICDTemplateExpansionKeySuite])
	require.NotEmpty(t, params["__SUITE_INCLUDE_LIST__"])
	require.NotEmpty(t, params["__SUITE_SERVICE_OVERRIDES__"])
	require.NotEmpty(t, params["__SUITE_INIT_PS_ID__"])
}

func TestBuildStaticParams(t *testing.T) {
	t.Parallel()

	params := buildStaticParams()
	require.Equal(t, cryptoutilSharedMagic.DefaultOTLPServiceDefault, params[cryptoutilSharedMagic.CICDTemplateExpansionKeySuite])
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
	// Should expand to many files (>260: deployment + config + secrets + *_usage.go per PS-ID).
	require.Greater(t, len(expected), 260)
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
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDTemplatesRelPath), cryptoutilSharedMagic.CICDTempDirPermissions))

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
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, cryptoutilSharedMagic.CICDTemplatesRelPath), cryptoutilSharedMagic.CICDTempDirPermissions))

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
	templatesDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDTemplatesRelPath)
	require.NoError(t, os.MkdirAll(templatesDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	// Create a sub-directory whose name ends in .yml â€” os.ReadFile on a directory is an error.
	fakeFilePath := filepath.Join(templatesDir, "config.yml")
	require.NoError(t, os.MkdirAll(fakeFilePath, cryptoutilSharedMagic.CICDTempDirPermissions))

	_, err := loadTemplatesDirFn(tmpDir, func(_ string, fn fs.WalkDirFunc) error {
		// Fake: claim the directory-with-.yml-name is a regular file.
		return fn(fakeFilePath, &fakeDirEntry{name: "config.yml", isDir: false}, nil)
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "walk templates directory")
}

// TestCompareBase64Placeholder_TrailingTooShort verifies that a trailing __BASE64_CHAR43__
// segment whose actual value is shorter than 43 characters reports a "too short" error.
func TestCompareBase64Placeholder_TrailingTooShort(t *testing.T) {
	t.Parallel()

	// "SHORT" is only 5 characters â€” well below the 43-char minimum.
	diff := compareBase64Placeholder("prefix-"+cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder, "prefix-SHORT")
	require.NotEmpty(t, diff)
	require.Contains(t, diff, "too short")
}
