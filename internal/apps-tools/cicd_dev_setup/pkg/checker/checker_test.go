package checker

import (
	"context"
	"fmt"
	"testing"

	cryptoutilAppsToolsDevSetupConfig "cryptoutil/internal/apps-tools/cicd_dev_setup/pkg/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// MockExecutor implements Executor for testing.
type MockExecutor struct {
	ExecuteFunc func(ctx context.Context, cmd string, args ...string) (string, string, error)
}

func (me *MockExecutor) Execute(ctx context.Context, cmd string, args ...string) (string, string, error) {
	if me.ExecuteFunc != nil {
		return me.ExecuteFunc(ctx, cmd, args...)
	}

	return "", "", fmt.Errorf("not implemented")
}

func TestCheck_Happy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		dep           *cryptoutilAppsToolsDevSetupConfig.Dependency
		executeOutput string
		executeErr    error
		expectFound   bool
		expectVersion string
	}{
		{
			name: "Tool installed with version",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:             cryptoutilSharedMagic.DevSetupToolTypePython,
				CheckCmd:         cryptoutilSharedMagic.DevSetupTestPythonCmd,
				DetectVersionCmd: cryptoutilSharedMagic.DevSetupTestPythonCmd,
			},
			executeOutput: cryptoutilSharedMagic.DevSetupTestPythonVersion,
			executeErr:    nil,
			expectFound:   true,
			expectVersion: cryptoutilSharedMagic.DevSetupTestPythonVersionParsed,
		},
		{
			name: "Tool installed (go version format)",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:             "go",
				CheckCmd:         cryptoutilSharedMagic.DevSetupTestGoCmd,
				DetectVersionCmd: cryptoutilSharedMagic.DevSetupTestGoCmd,
			},
			executeOutput: cryptoutilSharedMagic.DevSetupTestGoVersionOutput,
			executeErr:    nil,
			expectFound:   true,
			expectVersion: cryptoutilSharedMagic.DevSetupTestGoVersionParsed,
		},
		{
			name: "Tool installed (no version detection)",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:     "pre-commit",
				CheckCmd: cryptoutilSharedMagic.DevSetupPreCommitCheckCmd,
			},
			executeOutput: "pre-commit 3.0.0",
			executeErr:    nil,
			expectFound:   true,
			expectVersion: "",
		},
		{
			name: "Tool version with v prefix",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:             "golangci-lint",
				CheckCmd:         cryptoutilSharedMagic.DevSetupTestGolangciCmd,
				DetectVersionCmd: cryptoutilSharedMagic.DevSetupTestGolangciCmd,
			},
			executeOutput: cryptoutilSharedMagic.DevSetupTestGolangciOutput,
			executeErr:    nil,
			expectFound:   true,
			expectVersion: cryptoutilSharedMagic.DevSetupTestGolangciVersion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			executor := &MockExecutor{
				ExecuteFunc: func(ctx context.Context, cmd string, args ...string) (string, string, error) {
					return tc.executeOutput, "", tc.executeErr
				},
			}

			checker := NewDependencyChecker(executor)

			found, version, err := checker.Check(context.Background(), tc.dep)
			require.NoError(t, err)
			require.Equal(t, tc.expectFound, found)

			if tc.dep.DetectVersionCmd != "" {
				require.Equal(t, tc.expectVersion, version)
			}
		})
	}
}

func TestCheck_Sad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		dep         *cryptoutilAppsToolsDevSetupConfig.Dependency
		executeErr  error
		expectFound bool
		expectErr   bool
	}{
		{
			name: "Tool not installed",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:     "missing-tool",
				CheckCmd: "missing-tool --version",
			},
			executeErr:  fmt.Errorf("command not found"),
			expectFound: false,
			expectErr:   false,
		},
		{
			name: "No check command defined",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:     cryptoutilSharedMagic.DevSetupTestInvalidName,
				CheckCmd: "",
			},
			expectErr: true,
		},
		{
			name: "Invalid check command",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:     cryptoutilSharedMagic.DevSetupTestInvalidName,
				CheckCmd: "",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			executor := &MockExecutor{
				ExecuteFunc: func(ctx context.Context, cmd string, args ...string) (string, string, error) {
					return "", "", tc.executeErr
				},
			}

			checker := NewDependencyChecker(executor)
			found, _, err := checker.Check(context.Background(), tc.dep)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectFound, found)
			}
		})
	}
}

func TestExtractVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{cryptoutilSharedMagic.DevSetupTestVersion123, cryptoutilSharedMagic.DevSetupTestVersion123},
		{cryptoutilSharedMagic.DevSetupTestPythonVersion, cryptoutilSharedMagic.DevSetupTestPythonVersionParsed},
		{cryptoutilSharedMagic.DevSetupTestGoVersionOutput, cryptoutilSharedMagic.DevSetupTestGoVersionParsed},
		{"version 2.12.2", cryptoutilSharedMagic.DevSetupTestGolangciVersion},
		{"v1.0.0-alpha", "1.0.0-alpha"},
		{cryptoutilSharedMagic.DevSetupTestGolangciOutput, cryptoutilSharedMagic.DevSetupTestGolangciVersion},
		{cryptoutilSharedMagic.DevSetupTestNoVersionHere, cryptoutilSharedMagic.DevSetupTestNoVersionHere},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := extractVersion(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
