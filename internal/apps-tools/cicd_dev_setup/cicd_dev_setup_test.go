// Copyright (c) 2025-2026 Justin Cranford.
package cicd_dev_setup

import (
	"context"
	"fmt"
	"testing"

	cryptoutilAppsToolsDevSetupChecker "cryptoutil/internal/apps-tools/cicd_dev_setup/pkg/checker"
	cryptoutilAppsToolsDevSetupConfig "cryptoutil/internal/apps-tools/cicd_dev_setup/pkg/config"
	cryptoutilAppsToolsDevSetupReporter "cryptoutil/internal/apps-tools/cicd_dev_setup/pkg/reporter"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// stubChecker returns a fixed (installed, version, err) sequence per call, so tests
// can simulate "not installed initially, then found after install" or "still not
// found after install" without a real system executor.
type stubChecker struct {
	responses []checkResponse
	call      int
}

type checkResponse struct {
	installed bool
	version   string
	err       error
}

func (sc *stubChecker) Check(_ context.Context, _ *cryptoutilAppsToolsDevSetupConfig.Dependency) (bool, string, error) {
	if sc.call >= len(sc.responses) {
		return false, "", fmt.Errorf("no more stub responses")
	}

	r := sc.responses[sc.call]
	sc.call++

	return r.installed, r.version, r.err
}

type stubInstaller struct {
	err error
}

func (si *stubInstaller) Install(_ context.Context, _ *cryptoutilAppsToolsDevSetupConfig.Dependency) error {
	return si.err
}

const testVersion200 = "2.0.0"

func oneDepConfig() *cryptoutilAppsToolsDevSetupConfig.Config {
	return &cryptoutilAppsToolsDevSetupConfig.Config{
		Groups: []*cryptoutilAppsToolsDevSetupConfig.Group{
			{
				Name: "Test Group",
				Dependencies: []*cryptoutilAppsToolsDevSetupConfig.Dependency{
					{Name: "tool", CheckCmd: "tool --version", InstallCmd: "install tool"},
				},
				BlockOnFailure: false,
			},
		},
	}
}

func TestRunSetup_AlreadyInstalled(t *testing.T) {
	t.Parallel()

	chk := &stubChecker{responses: []checkResponse{{installed: true, version: cryptoutilSharedMagic.ServiceVersion}}}
	inst := &stubInstaller{}

	summary, err := runSetup(context.Background(), oneDepConfig(), chk, inst, nil)
	require.NoError(t, err)
	require.False(t, summary.HasErrors, "expected no errors when dependency already installed")

	dep := summary.Groups["Test Group"].Dependencies[0]
	require.Equal(t, cryptoutilSharedMagic.DevSetupStatusInstalled, dep.Status)
	require.Equal(t, 1, chk.call, "expected exactly 1 check call (idempotent skip)")
}

func TestRunSetup_InstallThenVerified(t *testing.T) {
	t.Parallel()

	chk := &stubChecker{responses: []checkResponse{
		{installed: false},
		{installed: true, version: testVersion200},
	}}
	inst := &stubInstaller{}

	summary, err := runSetup(context.Background(), oneDepConfig(), chk, inst, nil)
	require.NoError(t, err)
	require.False(t, summary.HasErrors, "expected no errors when install succeeds and verification confirms it")

	dep := summary.Groups["Test Group"].Dependencies[0]
	require.Equal(t, cryptoutilSharedMagic.DevSetupStatusInstalled, dep.Status)
	require.Equal(t, testVersion200, dep.ActualVersion)
}

func TestRunSetup_InstallSucceedsButVerifyFails(t *testing.T) {
	t.Parallel()

	chk := &stubChecker{responses: []checkResponse{
		{installed: false},
		{installed: false},
	}}
	inst := &stubInstaller{}

	summary, err := runSetup(context.Background(), oneDepConfig(), chk, inst, nil)
	require.NoError(t, err)

	// A genuine post-install verification failure must be reported as an error,
	// not silently masked as "Installed" (this replaces the previous overly-broad
	// masking behavior that always claimed success even when tools were unusable).
	require.True(t, summary.HasErrors, "expected HasErrors=true when tool remains unverifiable after install")

	dep := summary.Groups["Test Group"].Dependencies[0]
	require.Equal(t, cryptoutilSharedMagic.DevSetupStatusVerifyFailed, dep.Status)
	require.NotNil(t, dep.Error, "expected non-nil error explaining the verification failure")
}

func TestRunSetup_InstallCommandFails(t *testing.T) {
	t.Parallel()

	chk := &stubChecker{responses: []checkResponse{{installed: false}}}
	inst := &stubInstaller{err: fmt.Errorf("install exited non-zero")}

	summary, err := runSetup(context.Background(), oneDepConfig(), chk, inst, nil)
	require.NoError(t, err)
	require.True(t, summary.HasErrors, "expected HasErrors=true when install command itself fails")

	dep := summary.Groups["Test Group"].Dependencies[0]
	require.Equal(t, cryptoutilSharedMagic.DevSetupStatusInstallFailed, dep.Status)
}

func TestRunSetup_CheckCommandErrors(t *testing.T) {
	t.Parallel()

	chk := &stubChecker{responses: []checkResponse{{err: fmt.Errorf("check exec failed")}}}
	inst := &stubInstaller{}

	summary, err := runSetup(context.Background(), oneDepConfig(), chk, inst, nil)
	require.NoError(t, err)
	require.True(t, summary.HasErrors, "expected HasErrors=true when the check command itself errors")

	dep := summary.Groups["Test Group"].Dependencies[0]
	require.Equal(t, cryptoutilSharedMagic.DevSetupStatusCheckFailed, dep.Status)
}

// stubReporter records the summary it receives so tests can assert internalMainWithSetup
// forwards it and always returns nil (post-checkout must never block on tool issues).
type stubReporter struct {
	received *cryptoutilAppsToolsDevSetupReporter.Summary
}

func (sr *stubReporter) ReportSummary(s *cryptoutilAppsToolsDevSetupReporter.Summary) error {
	sr.received = s

	return nil
}

func TestInternalMainWithSetup_NeverBlocksOnDependencyErrors(t *testing.T) {
	t.Parallel()

	chk := &stubChecker{responses: []checkResponse{
		{installed: false},
		{installed: false},
	}}
	inst := &stubInstaller{}
	rep := &stubReporter{}

	setup := &Setup{Config: oneDepConfig(), Checker: chk, Installer: inst, Reporter: rep}

	err := internalMainWithSetup(context.Background(), setup)
	require.NoError(t, err, "expected nil error (post-checkout must not block)")
	require.NotNil(t, rep.received, "expected cryptoutilAppsToolsDevSetupReporter to receive a summary")
	require.True(t, rep.received.HasErrors, "expected summary to accurately report HasErrors=true even though exit is graceful")
}

func TestInternalMainWithSetup_NilSetup(t *testing.T) {
	t.Parallel()

	err := internalMainWithSetup(context.Background(), nil)
	require.Error(t, err, "expected error for nil setup")
}

var _ cryptoutilAppsToolsDevSetupChecker.Checker = (*stubChecker)(nil)
