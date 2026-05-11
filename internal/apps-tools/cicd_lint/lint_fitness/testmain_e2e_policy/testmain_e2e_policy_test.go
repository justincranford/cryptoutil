// Copyright (c) 2025-2026 Justin Cranford.
package testmain_e2e_policy_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	lintFitnessTestmainE2EPolicy "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/testmain_e2e_policy"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	compliantE2ETestMain = `//go:build e2e

package e2e_test

import (
	"os"
	"testing"

	cryptoutilTestOrchE2e "cryptoutil/internal/apps-framework/service/test_orch_e2e"
)

func TestMain(m *testing.M) {
	os.Exit(cryptoutilTestOrchE2e.SetupE2ETestMain(m, cryptoutilTestOrchE2e.E2ETestConfig{}, nil))
}
`

	legacyE2ETestMain = `//go:build e2e

package e2e_test

import (
	"os"
	"testing"

	cryptoutilAppsFrameworkTestingE2eInfra "cryptoutil/internal/apps-framework/service/testing/e2e_infra"
)

func TestMain(m *testing.M) {
	os.Exit(cryptoutilAppsFrameworkTestingE2eInfra.SetupE2ETestMain(m, cryptoutilAppsFrameworkTestingE2eInfra.E2ETestConfig{}, nil))
}
`

	missingRequiredOnlyE2ETestMain = `//go:build e2e

package e2e_test

import (
	"testing"
)
`

	forbiddenWithRequiredE2ETestMain = `//go:build e2e

package e2e_test

import (
	"os"
	"testing"

	cryptoutilTestOrchE2e "cryptoutil/internal/apps-framework/service/test_orch_e2e"
	cryptoutilAppsFrameworkTestingE2eInfra "cryptoutil/internal/apps-framework/service/testing/e2e_infra"
)

func TestMain(m *testing.M) {
	os.Exit(cryptoutilTestOrchE2e.SetupE2ETestMain(m, cryptoutilTestOrchE2e.E2ETestConfig{}, nil))
}
`
)

func writeE2ETestMain(t *testing.T, root, psid, content string) {
	t.Helper()

	e2eDir := filepath.Join(root, "internal", "apps", psid, "e2e")
	require.NoError(t, os.MkdirAll(e2eDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(e2eDir, "testmain_e2e_test.go"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))
}

func TestFindViolations_MissingAppsDir_ReturnsError(t *testing.T) {
	t.Parallel()

	violations, err := lintFitnessTestmainE2EPolicy.FindViolations(t.TempDir())
	require.Nil(t, violations)
	require.Error(t, err)
}

func TestFindViolations_Compliant_NoViolations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeE2ETestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, compliantE2ETestMain)

	violations, err := lintFitnessTestmainE2EPolicy.FindViolations(dir)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolations_LegacyImport_Violations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeE2ETestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, legacyE2ETestMain)

	violations, err := lintFitnessTestmainE2EPolicy.FindViolations(dir)
	require.NoError(t, err)
	require.Len(t, violations, 2)
	require.Contains(t, violations[0].Reason, "must import")
	require.Contains(t, violations[1].Reason, "must not import")
}

func TestFindViolations_MissingRequiredOnly_Violation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeE2ETestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, missingRequiredOnlyE2ETestMain)

	violations, err := lintFitnessTestmainE2EPolicy.FindViolations(dir)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0].Reason, "must import")
}

func TestFindViolations_ForbiddenWithRequired_Violation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeE2ETestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, forbiddenWithRequiredE2ETestMain)

	violations, err := lintFitnessTestmainE2EPolicy.FindViolations(dir)
	require.NoError(t, err)
	require.Len(t, violations, 1)
	require.Contains(t, violations[0].Reason, "must not import")
}

func TestFindViolations_SkipsGitAndVendorDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeE2ETestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, compliantE2ETestMain)

	gitDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.CICDExcludeDirGit, "e2e")
	vendorDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.CICDExcludeDirVendor, "e2e")

	require.NoError(t, os.MkdirAll(gitDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(vendorDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "testmain_e2e_test.go"), []byte(legacyE2ETestMain), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(vendorDir, "testmain_e2e_test.go"), []byte(legacyE2ETestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessTestmainE2EPolicy.FindViolations(dir)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolations_IgnoresNonTargetFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	nonTargetDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, "e2e")
	require.NoError(t, os.MkdirAll(nonTargetDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(nonTargetDir, "other_test.go"), []byte("package e2e_test"), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessTestmainE2EPolicy.FindViolations(dir)
	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestCheckInDir_WithViolation_ReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeE2ETestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, legacyE2ETestMain)

	logger := cryptoutilCmdCicdCommon.NewLogger("testmain-e2e-policy-test")
	err := lintFitnessTestmainE2EPolicy.CheckInDir(logger, dir)
	require.Error(t, err)
}

func TestCheck_CalledFromTestWorkdir(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("testmain-e2e-policy-test")
	_ = lintFitnessTestmainE2EPolicy.Check(logger)
}

func TestLintAndCheck_FromTemporaryWorkdir(t *testing.T) {
	dir := t.TempDir()
	writeE2ETestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, compliantE2ETestMain)

	original, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(original))
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("testmain-e2e-policy-test")
	require.NoError(t, lintFitnessTestmainE2EPolicy.Lint(logger))
	require.NoError(t, lintFitnessTestmainE2EPolicy.Check(logger))
}
