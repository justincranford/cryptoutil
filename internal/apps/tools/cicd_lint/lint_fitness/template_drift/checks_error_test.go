// Copyright (c) 2025 Justin Cranford

package template_drift

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestCheckTemplateCompliance_PublicWrapper tests the public wrapper using os.Chdir.
// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheckTemplateCompliance_PublicWrapper(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)

	root := projectRoot()

	require.NoError(t, os.Chdir(root))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test-template-compliance-public")
	require.NoError(t, CheckTemplateCompliance(logger))
}

func TestCheckTemplateComplianceInDir_Success(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-compliance-success")
	successFn := func(_ string) error { return nil }
	err := checkTemplateComplianceInDir(logger, t.TempDir(), successFn)
	require.NoError(t, err)
}

func TestCheckTemplateComplianceInDir_ComplianceError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-compliance-error")
	failFn := func(_ string) error {
		return fmt.Errorf("template-compliance violations:\ndeployments/test/Dockerfile: content drift")
	}
	err := checkTemplateComplianceInDir(logger, t.TempDir(), failFn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "template-compliance violations")
}

func TestCheckTemplateComplianceInDir_LoadError(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-compliance-load-error")
	loadErrFn := func(_ string) error {
		return fmt.Errorf("load templates: templates directory not found")
	}
	err := checkTemplateComplianceInDir(logger, t.TempDir(), loadErrFn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "load templates")
}

func TestDefaultComplianceFn_WithProjectRoot(t *testing.T) {
	t.Parallel()

	err := defaultComplianceFn(projectRoot())
	require.NoError(t, err)
}

func TestDefaultComplianceFn_MissingTemplatesDir(t *testing.T) {
	t.Parallel()

	err := defaultComplianceFn(t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "load templates")
}

func TestDefaultComplianceFn_MissingActualFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a minimal templates directory with one template.
	templDir := filepath.Join(tmpDir, cryptoutilSharedMagic.CICDTemplatesRelPath, "deployments", cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID)
	require.NoError(t, os.MkdirAll(templDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(
		filepath.Join(templDir, "Dockerfile"),
		[]byte("FROM __PS_ID__"),
		cryptoutilSharedMagic.CacheFilePermissions,
	))

	err := defaultComplianceFn(tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "template-compliance violations")
}
