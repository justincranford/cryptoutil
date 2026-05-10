// Copyright (c) 2025-2026 Justin Cranford.
package testmain_orchestration_policy_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	lintFitnessTestmainOrchestrationPolicy "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/testmain_orchestration_policy"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testSubPkgServer = "server"
	testSubPkgClient = "client"
)

// compliantTestMain contains a valid testmain_test.go that imports test_orch_integration.
const compliantTestMain = `package server

import (
	cryptoutilTestOrcIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	"testing"
	"os"
)

func TestMain(m *testing.M) {
	srv := cryptoutilTestOrcIntegration.StartIntegrationServerForTestMain(nil, nil, nil)
	os.Exit(m.Run())
	_ = srv
}
`

// violatingTestMain contains a testmain_test.go that does NOT import test_orch_integration.
const violatingTestMain = `package server

import (
	"testing"
	"os"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
`

// makeServerDir creates the canonical server/ directory for a given PS-ID under dir.
func makeServerDir(t *testing.T, dir, psid string) string {
	t.Helper()

	serverDir := filepath.Join(dir, "internal", "apps", psid, testSubPkgServer)
	require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	return serverDir
}

// writePSIDServerTestMain writes a testmain_test.go into server/ for a given PS-ID.
func writePSIDServerTestMain(t *testing.T, dir, psid, content string) {
	t.Helper()

	serverDir := makeServerDir(t, dir, psid)
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte(content), cryptoutilSharedMagic.FilePermissionsDefault))
}

func TestFindViolations_EmptyAppsDir_ReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// No internal/apps directory created.
	violations, err := lintFitnessTestmainOrchestrationPolicy.FindViolations(dir)

	require.Error(t, err)
	require.Nil(t, violations)
}

func TestFindViolations_MissingServerTestMain_ReturnsViolation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create server/ directory but no testmain_test.go inside.
	makeServerDir(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS)

	violations, err := lintFitnessTestmainOrchestrationPolicy.FindViolations(dir)

	require.NoError(t, err)
	require.NotEmpty(t, violations)

	// At least one violation should mention the sm-kms PS-ID.
	var found bool

	for _, v := range violations {
		if v.Reason != "" && v.File != "" {
			found = true

			break
		}
	}

	require.True(t, found, "expected at least one violation for missing testmain_test.go")
}

func TestFindViolations_CompliantServerTestMain_NoViolation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePSIDServerTestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, compliantTestMain)

	violations, err := lintFitnessTestmainOrchestrationPolicy.FindViolations(dir)

	require.NoError(t, err)
	// Violations may still occur for other PS-IDs since this is a minimal dir,
	// but none should involve sm-kms/server.
	for _, v := range violations {
		require.NotContains(t, v.File, cryptoutilSharedMagic.OTLPServiceSMKMS+"/server",
			"sm-kms/server should have no violations")
	}
}

func TestFindViolations_ViolatingServerTestMain_ReturnsViolation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePSIDServerTestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, violatingTestMain)

	violations, err := lintFitnessTestmainOrchestrationPolicy.FindViolations(dir)

	require.NoError(t, err)

	var found bool

	for _, v := range violations {
		if filepath.Base(filepath.Dir(v.File)) == testSubPkgServer {
			found = true

			break
		}
	}

	require.True(t, found, "expected a violation for sm-kms/server testmain missing test_orch_integration import")
}

func TestFindViolations_CompliantClientTestMain_NoViolation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create compliant server testmain for sm-kms.
	writePSIDServerTestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, compliantTestMain)

	// Create compliant client testmain.
	clientDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, testSubPkgClient)
	require.NoError(t, os.MkdirAll(clientDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(clientDir, "testmain_test.go"), []byte(compliantTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessTestmainOrchestrationPolicy.FindViolations(dir)

	require.NoError(t, err)

	for _, v := range violations {
		require.NotContains(t, v.File, cryptoutilSharedMagic.OTLPServiceSMKMS,
			"sm-kms should have no violations when both server and client testmains are compliant")
	}
}

func TestFindViolations_ViolatingClientTestMain_ReturnsViolation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Compliant server testmain.
	writePSIDServerTestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, compliantTestMain)

	// Violating client testmain.
	clientDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, testSubPkgClient)
	require.NoError(t, os.MkdirAll(clientDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(clientDir, "testmain_test.go"), []byte(violatingTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessTestmainOrchestrationPolicy.FindViolations(dir)

	require.NoError(t, err)

	var found bool

	for _, v := range violations {
		if filepath.Base(filepath.Dir(v.File)) == testSubPkgClient {
			found = true

			break
		}
	}

	require.True(t, found, "expected a violation for sm-kms/client testmain missing test_orch_integration import")
}

func TestCheckInDir_ReturnsError_WhenViolationsExist(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create server/ dir without testmain.
	makeServerDir(t, dir, cryptoutilSharedMagic.OTLPServiceJoseJA)

	logger := cryptoutilCmdCicdCommon.NewLogger("testmain-orchestration-policy-test")
	err := lintFitnessTestmainOrchestrationPolicy.CheckInDir(logger, dir)

	require.Error(t, err)
}

func TestCheckInDir_NoViolations_ReturnsNil(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	for _, psid := range []string{
		cryptoutilSharedMagic.OTLPServiceSMKMS,
		cryptoutilSharedMagic.OTLPServiceSMIM,
		cryptoutilSharedMagic.OTLPServiceJoseJA,
		cryptoutilSharedMagic.OTLPServicePKICA,
		cryptoutilSharedMagic.OTLPServiceSkeletonTemplate,
		cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
		cryptoutilSharedMagic.OTLPServiceIdentityIDP,
		cryptoutilSharedMagic.OTLPServiceIdentityRP,
		cryptoutilSharedMagic.OTLPServiceIdentityRS,
		cryptoutilSharedMagic.OTLPServiceIdentitySPA,
	} {
		writePSIDServerTestMain(t, dir, psid, compliantTestMain)
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("testmain-orchestration-policy-test")
	err := lintFitnessTestmainOrchestrationPolicy.CheckInDir(logger, dir)

	require.NoError(t, err)
}

func TestCheck_CalledFromTestWorkdir(t *testing.T) {
	t.Parallel()

	// Check() runs CheckInDir(".") from the process working directory.
	// The test binary runs in the package directory, not the repo root,
	// so internal/apps won't exist — the call must return without panicking.
	logger := cryptoutilCmdCicdCommon.NewLogger("testmain-orchestration-policy-test")
	_ = lintFitnessTestmainOrchestrationPolicy.Check(logger)
}

func TestWalkTestMainFiles_ReturnsServerAndClientPaths(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	writePSIDServerTestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, compliantTestMain)

	clientDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, testSubPkgClient)
	require.NoError(t, os.MkdirAll(clientDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(clientDir, "testmain_test.go"), []byte(compliantTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	paths, err := lintFitnessTestmainOrchestrationPolicy.WalkTestMainFiles(dir)

	require.NoError(t, err)
	require.Len(t, paths, 2)

	var hasServer, hasClient bool

	for _, p := range paths {
		base := filepath.Base(filepath.Dir(p))
		if base == testSubPkgServer {
			hasServer = true
		}

		if base == testSubPkgClient {
			hasClient = true
		}
	}

	require.True(t, hasServer, "expected server/testmain_test.go in walk results")
	require.True(t, hasClient, "expected client/testmain_test.go in walk results")
}

func TestWalkTestMainFiles_EmptyAppsDir_ReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := lintFitnessTestmainOrchestrationPolicy.WalkTestMainFiles(dir)

	require.Error(t, err)
}

func TestFindViolations_ServerIsDirectory_ReturnsViolation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create testmain_test.go as a DIRECTORY to hit the info.IsDir() branch.
	fakeFile := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.OTLPServiceSMKMS, testSubPkgServer, "testmain_test.go")
	require.NoError(t, os.MkdirAll(fakeFile, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	violations, err := lintFitnessTestmainOrchestrationPolicy.FindViolations(dir)

	require.NoError(t, err)
	require.NotEmpty(t, violations)
}

func TestWalkTestMainFiles_SkipsGitAndVendorDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Compliant server testmain for sm-kms (will be found).
	writePSIDServerTestMain(t, dir, cryptoutilSharedMagic.OTLPServiceSMKMS, compliantTestMain)

	// Testmain files inside .git and vendor directories (must be skipped).
	gitDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.CICDExcludeDirGit, testSubPkgServer)
	vendorDir := filepath.Join(dir, "internal", "apps", cryptoutilSharedMagic.CICDExcludeDirVendor, testSubPkgServer)

	require.NoError(t, os.MkdirAll(gitDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.MkdirAll(vendorDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "testmain_test.go"), []byte(compliantTestMain), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(vendorDir, "testmain_test.go"), []byte(compliantTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	paths, err := lintFitnessTestmainOrchestrationPolicy.WalkTestMainFiles(dir)

	require.NoError(t, err)

	for _, p := range paths {
		require.NotContains(t, p, cryptoutilSharedMagic.CICDExcludeDirGit, "should not walk into .git")
		require.NotContains(t, p, cryptoutilSharedMagic.CICDExcludeDirVendor, "should not walk into vendor")
	}

	require.Len(t, paths, 1, "only sm-kms/server testmain should be found")
}
