// Copyright (c) 2025-2026 Justin Cranford.
package template_drift

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// testBase64String is a 47-character base64 string used across multiple comparison tests.
const testBase64String = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuv"

// projectRoot returns the path to the project root for integration tests.
func projectRoot() string {
	return filepath.Join("..", "..", "..", "..", "..")
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
	templDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDTemplatesRelPath, "deployments", cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID)
	require.NoError(t, os.MkdirAll(templDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(templDir, "Dockerfile"), []byte("FROM "+cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(templDir, ".gitkeep"), []byte(""), cryptoutilSharedMagic.CacheFilePermissions))

	templates, err := LoadTemplatesDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, templates, 1)
	require.Equal(t, "FROM "+cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID, templates["deployments/"+cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID+"/Dockerfile"])
}

func TestLoadTemplatesDir_SkipsStructuralMetaFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	baseTemplDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDTemplatesRelPath)

	// Deployment template (should be loaded).
	deploymentsDir := filepath.Join(baseTemplDir, "deployments", cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID)
	require.NoError(t, os.MkdirAll(deploymentsDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(deploymentsDir, "Dockerfile"), []byte("FROM alpine"), cryptoutilSharedMagic.CacheFilePermissions))

	// Go source template (should be LOADED â€” cmd/ is no longer skipped).
	cmdDir := filepath.Join(baseTemplDir, "cmd", cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID)
	require.NoError(t, os.MkdirAll(cmdDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("//go:build ignore\n\npackage main"), cryptoutilSharedMagic.CacheFilePermissions))

	// Go source template under internal/ (should be LOADED â€” only MANIFEST.yaml is skipped).
	internalDir := filepath.Join(baseTemplDir, "internal", "apps", cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID)
	require.NoError(t, os.MkdirAll(internalDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(internalDir, "MANIFEST.yaml"), []byte("required_root_files: []"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(internalDir, "__SERVICE___usage.go"), []byte("//go:build ignore\n\npackage main\nvar _ = \"__SERVICE__\""), cryptoutilSharedMagic.CacheFilePermissions))

	templates, err := LoadTemplatesDir(tmpDir)
	require.NoError(t, err)

	// MANIFEST.yaml is a structural meta-file and must be excluded.
	_, hasManifest := templates["internal/apps/"+cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID+"/MANIFEST.yaml"]
	require.False(t, hasManifest, "MANIFEST.yaml is a structural meta-file and must be skipped")

	// Deployment Dockerfile must be loaded.
	_, hasDockerfile := templates["deployments/"+cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID+"/Dockerfile"]
	require.True(t, hasDockerfile, "deployment template must be loaded")

	// cmd/ Go source template (non-usage) must be SKIPPED â€” cmd/ templates other than *_usage.go are excluded.
	_, hasCmdMain := templates["cmd/"+cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID+"/main.go"]
	require.False(t, hasCmdMain, "non-usage Go source template in cmd/ must be skipped")

	// internal/ Go source template (*_usage.go) must be loaded (//go:build ignore header stripped).
	usageContent, hasUsage := templates["internal/apps/"+cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID+"/__SERVICE___usage.go"]
	require.True(t, hasUsage, "usage Go source template in internal/ must be loaded")
	require.NotContains(t, usageContent, "//go:build ignore", "//go:build ignore tag must be stripped")
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
		"deployments/" + cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID + "/Dockerfile": "FROM " + cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID + ":latest",
	}

	expected := BuildExpectedFS(templates)
	// One expanded entry per PS-ID.
	require.Len(t, expected, len(cryptoutilRegistry.AllProductServices()))

	// Spot-check one expansion.
	content, ok := expected["deployments/"+cryptoutilSharedMagic.OTLPServiceSMKMS+"/Dockerfile"]
	require.True(t, ok)
	require.Contains(t, content, cryptoutilSharedMagic.OTLPServiceSMKMS)
	require.NotContains(t, content, cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID)
}

func TestBuildExpectedFS_ProductExpansion(t *testing.T) {
	t.Parallel()

	templates := map[string]string{
		"deployments/" + cryptoutilSharedMagic.CICDTemplateExpansionKeyProduct + "/compose.yml": "product: " + cryptoutilSharedMagic.CICDTemplateExpansionKeyProduct,
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
		"deployments/" + cryptoutilSharedMagic.CICDTemplateExpansionKeySuite + "/Dockerfile": "FROM " + cryptoutilSharedMagic.CICDTemplateExpansionKeySuite + ":latest",
	}

	expected := BuildExpectedFS(templates)
	// 1 suite â†’ 1 entry.
	require.Len(t, expected, 1)

	content, ok := expected["deployments/"+cryptoutilSharedMagic.DefaultOTLPServiceDefault+"/Dockerfile"]
	require.True(t, ok)
	require.Contains(t, content, cryptoutilSharedMagic.DefaultOTLPServiceDefault)
}

func TestBuildExpectedFS_StaticPath(t *testing.T) {
	t.Parallel()

	templates := map[string]string{
		"deployments/shared-telemetry/compose.yml": "suite: " + cryptoutilSharedMagic.CICDTemplateExpansionKeySuite,
	}

	expected := BuildExpectedFS(templates)
	require.Len(t, expected, 1)

	content, ok := expected["deployments/shared-telemetry/compose.yml"]
	require.True(t, ok)
	require.Contains(t, content, cryptoutilSharedMagic.DefaultOTLPServiceDefault)
	require.NotContains(t, content, cryptoutilSharedMagic.CICDTemplateExpansionKeySuite)
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
		"deployments/__PS_ID__/secrets/password.secret": "__PS_ID__-password-" + cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder,
	}

	expected := BuildExpectedFS(templates)

	content, ok := expected["deployments/sm-kms/secrets/password.secret"]
	require.True(t, ok)
	// __BASE64_CHAR43__ should NOT be substituted â€” it's a comparison placeholder.
	require.Contains(t, content, cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder)
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

	// The actual file has a real base64 value instead of __BASE64_CHAR43__.
	filePath := filepath.Join(tmpDir, "deployments", "test", "secrets", "password.secret")
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), cryptoutilSharedMagic.CICDTempDirPermissions))

	realB64 := testBase64String // 47 chars
	require.NoError(t, os.WriteFile(filePath, []byte("test-password-"+realB64), cryptoutilSharedMagic.CacheFilePermissions))

	expected := map[string]string{
		"deployments/test/secrets/password.secret": "test-password-" + cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder,
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
		"prefix-"+cryptoutilSharedMagic.CICDTemplateBase64Char43Placeholder, "prefix-"+realB64)
	require.Empty(t, diff)
}
