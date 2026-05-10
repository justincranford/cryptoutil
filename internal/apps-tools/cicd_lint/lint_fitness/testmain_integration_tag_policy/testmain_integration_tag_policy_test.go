// Copyright (c) 2025-2026 Justin Cranford.
package testmain_integration_tag_policy_test

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	lintFitnessTestmainIntegrationTagPolicy "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/testmain_integration_tag_policy"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// untaggedTestMain is a valid testmain_test.go with no build tags.
const untaggedTestMain = `// Copyright (c) 2025-2026 Justin Cranford.
//
// TestMain for package integration tests.

package server

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
`

// taggedTestMain contains a testmain_test.go with a forbidden //go:build tag.
// Note: the legacy //+ build line is embedded as a non-directive comment to avoid
// triggering false positives in our own linter during compilation.
const taggedTestMain = "/" + `/go:build integration` + "\n" + `// ` + `+build integration` + "\n" +
	"\n" +
	`// Copyright (c) 2025-2026 Justin Cranford.` + "\n" +
	"\n" +
	`package server` + "\n" +
	"\n" +
	`import (` + "\n" +
	`	"os"` + "\n" +
	`	"testing"` + "\n" +
	`)` + "\n" +
	"\n" +
	`func TestMain(m *testing.M) {` + "\n" +
	`	os.Exit(m.Run())` + "\n" +
	`}`

// e2eTaggedTestMain contains a testmain_test.go with a forbidden //go:build e2e tag.
const e2eTaggedTestMain = "/" + `/go:build e2e` + "\n" +
	"\n" +
	`package server` + "\n" +
	"\n" +
	`import (` + "\n" +
	`	"os"` + "\n" +
	`	"testing"` + "\n" +
	`)` + "\n" +
	"\n" +
	`func TestMain(m *testing.M) {` + "\n" +
	`	os.Exit(m.Run())` + "\n" +
	`}`

// makeInternalDir creates internal/<relPath> under dir and returns its full path.
func makeInternalDir(t *testing.T, dir, relPath string) string {
	t.Helper()

	fullPath := filepath.Join(dir, "internal", relPath)
	require.NoError(t, os.MkdirAll(fullPath, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	return fullPath
}

func TestFindViolations_MissingInternalDir_ReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// No internal/ directory created.
	violations, err := lintFitnessTestmainIntegrationTagPolicy.FindViolations(dir)

	require.Error(t, err)
	require.Nil(t, violations)
}

func TestFindViolations_NoTestmainFiles_NoViolations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	internalDir := makeInternalDir(t, dir, "apps/sm-kms/server")

	// Write a non-testmain file.
	require.NoError(t, os.WriteFile(filepath.Join(internalDir, "some_test.go"), []byte("package server\n"), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessTestmainIntegrationTagPolicy.FindViolations(dir)

	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolations_UntaggedTestmain_NoViolations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	serverDir := makeInternalDir(t, dir, "apps/sm-kms/server")
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte(untaggedTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessTestmainIntegrationTagPolicy.FindViolations(dir)

	require.NoError(t, err)
	require.Empty(t, violations)
}

func TestFindViolations_IntegrationTaggedTestmain_ReturnsViolation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	serverDir := makeInternalDir(t, dir, "apps/sm-kms/server")
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte(taggedTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessTestmainIntegrationTagPolicy.FindViolations(dir)

	require.NoError(t, err)
	require.NotEmpty(t, violations)
	require.Contains(t, violations[0].Tag, "//go:build integration")
}

func TestFindViolations_E2ETaggedTestmain_ReturnsViolation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	serverDir := makeInternalDir(t, dir, "apps/sm-kms/server")
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte(e2eTaggedTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessTestmainIntegrationTagPolicy.FindViolations(dir)

	require.NoError(t, err)
	require.NotEmpty(t, violations)
	require.Contains(t, violations[0].Tag, "//go:build e2e")
}

func TestFindViolations_LegacyBuildTagLine_ReturnsViolation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	serverDir := makeInternalDir(t, dir, "apps/sm-kms/server")

	// The tagged file has BOTH //go:build and // +build lines.
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte(taggedTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessTestmainIntegrationTagPolicy.FindViolations(dir)

	require.NoError(t, err)
	// taggedTestMain has BOTH //go:build and // +build — expect two violations.
	require.GreaterOrEqual(t, len(violations), 2, "expected violations for both //go:build and // +build lines")
}

func TestFindViolations_MultipleFiles_DetectsBothTagged(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	serverDir := makeInternalDir(t, dir, "apps/sm-kms/server")
	clientDir := makeInternalDir(t, dir, "apps/sm-kms/client")

	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte(untaggedTestMain), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(clientDir, "testmain_test.go"), []byte(taggedTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessTestmainIntegrationTagPolicy.FindViolations(dir)

	require.NoError(t, err)

	// Only client has a tagged testmain — server's is clean.
	for _, v := range violations {
		require.Contains(t, v.File, "client", "violation should be in client directory, not server")
	}

	require.NotEmpty(t, violations)
}

func TestCheckInDir_NoViolations_ReturnsNil(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	serverDir := makeInternalDir(t, dir, "apps/sm-kms/server")
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte(untaggedTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	logger := cryptoutilCmdCicdCommon.NewLogger("testmain-integration-tag-policy-test")
	err := lintFitnessTestmainIntegrationTagPolicy.CheckInDir(logger, dir)

	require.NoError(t, err)
}

func TestCheckInDir_WithViolations_ReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	serverDir := makeInternalDir(t, dir, "apps/sm-kms/server")
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte(taggedTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	logger := cryptoutilCmdCicdCommon.NewLogger("testmain-integration-tag-policy-test")
	err := lintFitnessTestmainIntegrationTagPolicy.CheckInDir(logger, dir)

	require.Error(t, err)
}

func TestCheck_CalledFromTestWorkdir(t *testing.T) {
	t.Parallel()

	// Check() calls CheckInDir(".") from the process working directory.
	// The test binary runs in the package directory, not the repo root,
	// so the call must not panic.
	logger := cryptoutilCmdCicdCommon.NewLogger("testmain-integration-tag-policy-test")
	_ = lintFitnessTestmainIntegrationTagPolicy.Check(logger)
}

func TestFindViolations_SkipsGitAndVendorDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// A clean testmain in sm-kms/server.
	serverDir := makeInternalDir(t, dir, "apps/sm-kms/server")
	require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte(untaggedTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	// Tagged testmains inside .git and vendor — must be skipped.
	gitDir := makeInternalDir(t, dir, cryptoutilSharedMagic.CICDExcludeDirGit+"/apps/sm-kms/server")
	vendorDir := makeInternalDir(t, dir, cryptoutilSharedMagic.CICDExcludeDirVendor+"/apps/sm-kms/server")
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "testmain_test.go"), []byte(taggedTestMain), cryptoutilSharedMagic.FilePermissionsDefault))
	require.NoError(t, os.WriteFile(filepath.Join(vendorDir, "testmain_test.go"), []byte(taggedTestMain), cryptoutilSharedMagic.FilePermissionsDefault))

	violations, err := lintFitnessTestmainIntegrationTagPolicy.FindViolations(dir)

	require.NoError(t, err)
	require.Empty(t, violations, "testmain files in .git and vendor should be skipped")
}
