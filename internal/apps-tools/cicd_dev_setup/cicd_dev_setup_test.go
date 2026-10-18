// Copyright (c) 2025-2026 Justin Cranford.
package cicd_dev_setup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilAppsToolsDevSetupChecker "cryptoutil/internal/apps-tools/cicd_dev_setup/pkg/checker"
	cryptoutilAppsToolsDevSetupConfig "cryptoutil/internal/apps-tools/cicd_dev_setup/pkg/config"
	cryptoutilAppsToolsDevSetupInstaller "cryptoutil/internal/apps-tools/cicd_dev_setup/pkg/installer"
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

const (
	testVersion200  = "2.0.0"
	testProjectRoot = "project-root"
)

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
	err      error
}

func (sr *stubReporter) ReportSummary(s *cryptoutilAppsToolsDevSetupReporter.Summary) error {
	sr.received = s

	return sr.err
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

func TestInternalMainWithInjectedDependencies(t *testing.T) {
	t.Parallel()

	runner := internalMain(
		func() (*cryptoutilAppsToolsDevSetupConfig.Config, error) {
			return &cryptoutilAppsToolsDevSetupConfig.Config{}, nil
		},
		func() (string, error) {
			return testProjectRoot, nil
		},
		newSetup,
	)

	var stdout, stderr strings.Builder
	require.NoError(t, runner(&stdout, &stderr))
	require.Contains(t, stdout.String(), "All dependencies installed successfully")
}

func TestInternalMainWithSetup_SetupRunnerFailure(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("injected setup runner failure")
	setup := &Setup{Reporter: &stubReporter{}}

	err := internalMainWithSetupWithRunner(context.Background(), setup, func(
		context.Context,
		*cryptoutilAppsToolsDevSetupConfig.Config,
		cryptoutilAppsToolsDevSetupChecker.Checker,
		cryptoutilAppsToolsDevSetupInstaller.Installer,
		cryptoutilAppsToolsDevSetupReporter.Reporter,
	) (*cryptoutilAppsToolsDevSetupReporter.Summary, error) {
		return nil, expectedErr
	})
	require.ErrorIs(t, err, expectedErr)
}

func TestInternalMainWithDependencies(t *testing.T) {
	t.Parallel()

	configError := fmt.Errorf("config unavailable")
	rootError := fmt.Errorf("root unavailable")
	reportError := fmt.Errorf("report unavailable")
	tests := []struct {
		name       string
		loadConfig func() (*cryptoutilAppsToolsDevSetupConfig.Config, error)
		findRoot   func() (string, error)
		reportErr  error
		wantErr    string
	}{
		{
			name: "configuration load fails",
			loadConfig: func() (*cryptoutilAppsToolsDevSetupConfig.Config, error) {
				return nil, configError
			},
			findRoot: func() (string, error) {
				return "", errors.New("root lookup should not be called")
			},
			wantErr: "failed to load configuration",
		},
		{
			name: "project root lookup fails",
			loadConfig: func() (*cryptoutilAppsToolsDevSetupConfig.Config, error) {
				return &cryptoutilAppsToolsDevSetupConfig.Config{}, nil
			},
			findRoot: func() (string, error) {
				return "", rootError
			},
			wantErr: "failed to determine project root",
		},
		{
			name: "reporter fails after setup",
			loadConfig: func() (*cryptoutilAppsToolsDevSetupConfig.Config, error) {
				return &cryptoutilAppsToolsDevSetupConfig.Config{}, nil
			},
			findRoot: func() (string, error) {
				return testProjectRoot, nil
			},
			reportErr: reportError,
			wantErr:   "failed to report results",
		},
		{
			name: "empty configuration completes",
			loadConfig: func() (*cryptoutilAppsToolsDevSetupConfig.Config, error) {
				return &cryptoutilAppsToolsDevSetupConfig.Config{}, nil
			},
			findRoot: func() (string, error) {
				return testProjectRoot, nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stdout := &strings.Builder{}
			stderr := &strings.Builder{}
			err := internalMainWithDependencies(stdout, stderr, tc.loadConfig, tc.findRoot, func(
				cfg *cryptoutilAppsToolsDevSetupConfig.Config,
				projectRoot string,
				out, errOut io.Writer,
			) *Setup {
				require.Equal(t, testProjectRoot, projectRoot)
				require.NotNil(t, cfg)

				return &Setup{
					Config:    cfg,
					Checker:   &stubChecker{},
					Installer: &stubInstaller{},
					Reporter:  &stubReporter{err: tc.reportErr},
				}
			})

			if tc.wantErr == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestMainWithRunner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		run      func(io.Writer, io.Writer) error
		wantCode int
		wantText string
	}{
		{
			name:     "success",
			run:      func(io.Writer, io.Writer) error { return nil },
			wantCode: exitCodeSuccess,
		},
		{
			name:     "failure",
			run:      func(io.Writer, io.Writer) error { return errors.New("setup failure") },
			wantCode: exitCodeFailure,
			wantText: "setup failure",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr strings.Builder
			require.Equal(t, tc.wantCode, mainWithRunner(&stdout, &stderr, tc.run))

			if tc.wantText != "" {
				require.Contains(t, stderr.String(), tc.wantText)
			}
		})
	}
}

func TestGetProjectRootWith(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		getwd   func() (string, error)
		stat    func(string) (os.FileInfo, error)
		wantErr string
		wantDir string
	}{
		{
			name: "working directory lookup fails",
			getwd: func() (string, error) {
				return "", errors.New("cwd unavailable")
			},
			stat:    os.Stat,
			wantErr: "failed to get current working directory",
		},
		{
			name: "finds root in parent directory",
			getwd: func() (string, error) {
				return filepath.Join("project", "nested"), nil
			},
			stat: func(path string) (os.FileInfo, error) {
				if path == filepath.Join("project", "go.mod") {
					return nil, nil
				}

				return nil, os.ErrNotExist
			},
			wantDir: "project",
		},
		{
			name: "root without go.mod fails",
			getwd: func() (string, error) {
				return string(filepath.Separator), nil
			},
			stat: func(string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			wantErr: "project root not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root, err := getProjectRootWith(tc.getwd, tc.stat)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantDir, root)
		})
	}
}

func TestGetProjectRoot_FindsRepositoryRoot(t *testing.T) {
	t.Parallel()

	root, err := getProjectRoot()
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(root, "go.mod"))
}

var _ cryptoutilAppsToolsDevSetupChecker.Checker = (*stubChecker)(nil)
