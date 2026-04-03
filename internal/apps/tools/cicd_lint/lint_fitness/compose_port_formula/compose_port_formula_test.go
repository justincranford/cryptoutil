// Copyright (c) 2025 Justin Cranford

package compose_port_formula

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// findProjectRoot walks up from the test file directory to find go.mod.
func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("Skipping — cannot find project root (no go.mod)")
		}

		dir = parent
	}
}

// buildComposeWithPorts creates a compose.yml in the given directory with the specified
// service name and port mapping.
func buildComposeWithPorts(t *testing.T, dir, serviceName, portMapping string) string {
	t.Helper()

	content := "services:\n" +
		"  " + serviceName + ":\n" +
		"    ports:\n" +
		"      - \"" + portMapping + "\"\n"

	path := filepath.Join(dir, "compose.yml")
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	return path
}

func TestCheck_PassesOnProjectRoot(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRoot(t)

	logger := cryptoutilCmdCicdCommon.NewLogger("test-compose-port-formula")

	err := CheckInDir(logger, projectRoot)
	require.NoError(t, err, "Project should pass compose port formula check")
}

func TestCheck_DelegatesToCheckInDir(t *testing.T) {
	t.Parallel()

	// Check() calls CheckInDir(logger, ".") from the workspace root.
	// Since tests run under the package directory, the project root (".")
	// is the workspace root which contains all compose files.
	logger := cryptoutilCmdCicdCommon.NewLogger("test-check-delegates")

	err := Check(logger)
	require.NoError(t, err, "Check() should pass on project root (delegates to CheckInDir)")
}

func TestCheckInDir_ValidServiceTierPorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		psID           string
		serviceVariant string
		variantOffset  int
	}{
		{
			name:           "sm-kms sqlite-1 service tier",
			psID:           cryptoutilSharedMagic.OTLPServiceSMKMS,
			serviceVariant: lintFitnessRegistry.ComposeVariantSQLite1,
			variantOffset:  0,
		},
		{
			name:           "sm-kms sqlite-2 service tier",
			psID:           cryptoutilSharedMagic.OTLPServiceSMKMS,
			serviceVariant: lintFitnessRegistry.ComposeVariantSQLite2,
			variantOffset:  1,
		},
		{
			name:           "sm-kms postgresql-1 service tier",
			psID:           cryptoutilSharedMagic.OTLPServiceSMKMS,
			serviceVariant: lintFitnessRegistry.ComposeVariantPostgres1,
			variantOffset:  2,
		},
		{
			name:           "sm-kms postgresql-2 service tier",
			psID:           cryptoutilSharedMagic.OTLPServiceSMKMS,
			serviceVariant: lintFitnessRegistry.ComposeVariantPostgres2,
			variantOffset:  3,
		},
		{
			name:           "sm-im sqlite-1 service tier",
			psID:           cryptoutilSharedMagic.OTLPServiceSMIM,
			serviceVariant: lintFitnessRegistry.ComposeVariantSQLite1,
			variantOffset:  0,
		},
		{
			name:           "jose-ja postgresql-2 service tier",
			psID:           cryptoutilSharedMagic.OTLPServiceJoseJA,
			serviceVariant: lintFitnessRegistry.ComposeVariantPostgres2,
			variantOffset:  3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			// SERVICE tier: deployments/{psid}/compose.yml
			svcDir := filepath.Join(tmpDir, "deployments", tc.psID)
			require.NoError(t, os.MkdirAll(svcDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

			basePort := lintFitnessRegistry.PublicPort(tc.psID)
			expectedHostPort := basePort + lintFitnessRegistry.PortTierOffsetService + tc.variantOffset
			svcName := lintFitnessRegistry.ComposeServiceName(tc.psID, tc.serviceVariant)

			portStr := strconv.Itoa(expectedHostPort) + cryptoutilSharedMagic.TestPort
			buildComposeWithPorts(t, svcDir, svcName, portStr)

			logger := cryptoutilCmdCicdCommon.NewLogger("test-valid-service-tier")
			err := CheckInDir(logger, tmpDir)
			require.NoError(t, err, "Valid service-tier port should pass")
		})
	}
}

func TestCheckInDir_ViolationServiceTierWrongPort(t *testing.T) {
	t.Parallel()

	// sm-kms sqlite-1 host port should be 8000, using 9999 → violation.
	psID := cryptoutilSharedMagic.OTLPServiceSMKMS

	tmpDir := t.TempDir()
	svcDir := filepath.Join(tmpDir, "deployments", psID)
	require.NoError(t, os.MkdirAll(svcDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	svcName := lintFitnessRegistry.ComposeServiceName(psID, lintFitnessRegistry.ComposeVariantSQLite1)
	buildComposeWithPorts(t, svcDir, svcName, "9999:8080")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-violation-service-tier")

	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "compose port formula violations")
	require.Contains(t, err.Error(), "host port 9999")
	require.Contains(t, err.Error(), "want 8000")
	require.Contains(t, err.Error(), "line 4", "violation must report correct 1-based line number")
}

func TestCheckInDir_ViolationProductTierWrongPort(t *testing.T) {
	t.Parallel()

	// sm-kms in PRODUCT tier: expected host port 18000, using 8000 → violation.
	psID := cryptoutilSharedMagic.OTLPServiceSMKMS
	product := "sm"

	tmpDir := t.TempDir()
	svcDir := filepath.Join(tmpDir, "deployments", product)
	require.NoError(t, os.MkdirAll(svcDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	svcName := lintFitnessRegistry.ComposeServiceName(psID, lintFitnessRegistry.ComposeVariantSQLite1)
	buildComposeWithPorts(t, svcDir, svcName, "8000:8080")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-violation-product-tier")

	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "compose port formula violations")
	require.Contains(t, err.Error(), "host port 8000")
	require.Contains(t, err.Error(), "want 18000")
}

func TestCheckInDir_ViolationSuiteTierWrongPort(t *testing.T) {
	t.Parallel()

	// sm-kms in SUITE tier: expected host port 28000, using 8000 → violation.
	psID := cryptoutilSharedMagic.OTLPServiceSMKMS

	suites := lintFitnessRegistry.AllSuites()
	require.NotEmpty(t, suites, "Registry must define at least one suite")

	suiteID := suites[0].ID

	tmpDir := t.TempDir()
	suiteDir := filepath.Join(tmpDir, "deployments", suiteID)
	require.NoError(t, os.MkdirAll(suiteDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	svcName := lintFitnessRegistry.ComposeServiceName(psID, lintFitnessRegistry.ComposeVariantSQLite1)
	buildComposeWithPorts(t, suiteDir, svcName, "8000:8080")

	logger := cryptoutilCmdCicdCommon.NewLogger("test-violation-suite-tier")

	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "compose port formula violations")
	require.Contains(t, err.Error(), "want 28000")
}

func TestCheckInDir_MissingComposeFileIsSkipped(t *testing.T) {
	t.Parallel()

	// Empty temp dir — no deployments, all compose files missing → no error (skipped).
	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-missing-compose-skipped")

	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err, "Missing compose files should be skipped, not counted as violations")
}

func TestCheckInDir_NonMatchingServicesIgnored(t *testing.T) {
	t.Parallel()

	// Compose file with only unrelated services — no violations.
	psID := cryptoutilSharedMagic.OTLPServiceSMKMS

	tmpDir := t.TempDir()
	svcDir := filepath.Join(tmpDir, "deployments", psID)
	require.NoError(t, os.MkdirAll(svcDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	content := "services:\n  unrelated-service:\n    ports:\n      - \"9999:8080\"\n"
	require.NoError(t, os.WriteFile(
		filepath.Join(svcDir, "compose.yml"),
		[]byte(content),
		cryptoutilSharedMagic.FilePermissionsDefault,
	))

	logger := cryptoutilCmdCicdCommon.NewLogger("test-non-matching-ignored")

	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err, "Non-matching service ports should be ignored")
}

func TestCheckInDir_ReadFileError(t *testing.T) {
	// Sequential seam test — must not use t.Parallel().
	original := portReadFileFn

	defer func() { portReadFileFn = original }()

	portReadFileFn = func(path string) ([]byte, error) {
		return nil, os.ErrPermission
	}

	// Create a deployments/{psid}/ dir so the file lookup attempts.
	psID := cryptoutilSharedMagic.OTLPServiceSMKMS

	tmpDir := t.TempDir()
	svcDir := filepath.Join(tmpDir, "deployments", psID)
	require.NoError(t, os.MkdirAll(svcDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	// Write a real file so os.IsNotExist returns false → triggers the read error path.
	require.NoError(t, os.WriteFile(
		filepath.Join(svcDir, "compose.yml"),
		[]byte("services: {}"),
		cryptoutilSharedMagic.FilePermissionsDefault,
	))

	logger := cryptoutilCmdCicdCommon.NewLogger("test-read-file-error")

	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot read")
}

func TestCheckTierPorts_ValidProductTierAllVariants(t *testing.T) {
	t.Parallel()

	// Verify all 3 variants for sm-im at PRODUCT tier.
	psID := cryptoutilSharedMagic.OTLPServiceSMIM
	basePort := lintFitnessRegistry.PublicPort(psID)

	tmpDir := t.TempDir()

	// Build compose.yml with all 3 services for sm-im at PRODUCT tier.
	var portEntries string

	for _, vm := range variantMappings {
		svcName := lintFitnessRegistry.ComposeServiceName(psID, vm.serviceVariant)
		hostPort := basePort + lintFitnessRegistry.PortTierOffsetProduct + vm.variantOffset
		portEntries += "  " + svcName + ":\n    ports:\n      - \"" + strconv.Itoa(hostPort) + ":8080\"\n"
	}

	content := "services:\n" + portEntries
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "compose.yml"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	violations := checkTierPorts(tmpDir, psID, "compose.yml", basePort, lintFitnessRegistry.PortTierOffsetProduct)
	require.Empty(t, violations, "All 3 product-tier variants correct should produce no violations")
}

// Sequential: modifies package-level allSuitesFn seam.
func TestCheckInDir_NoSuites(t *testing.T) {
	origSuites := allSuitesFn

	defer func() { allSuitesFn = origSuites }()

	allSuitesFn = func() []lintFitnessRegistry.Suite { return nil }

	// With no suites, only SERVICE and PRODUCT tiers should be checked.
	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test-no-suites")

	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err, "No suites defined should skip suite tier and produce no violations")
}

func TestCheckTierPorts_TwoCharLineNotPanic(t *testing.T) {
	t.Parallel()

	// A line exactly 2 chars long must not panic (boundary: len > 2 guard).
	psID := cryptoutilSharedMagic.OTLPServiceSMKMS
	basePort := lintFitnessRegistry.PublicPort(psID)

	tmpDir := t.TempDir()

	content := "services:\n  \n  " + lintFitnessRegistry.ComposeServiceName(psID, lintFitnessRegistry.ComposeVariantSQLite1) + ":\n    ports:\n      - \"" + strconv.Itoa(basePort) + ":8080\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "compose.yml"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))

	violations := checkTierPorts(tmpDir, psID, "compose.yml", basePort, lintFitnessRegistry.PortTierOffsetService)
	require.Empty(t, violations, "Two-char line must not cause panic or violation")
}
