package installer

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
	CallCount   int
}

func (me *MockExecutor) Execute(ctx context.Context, cmd string, args ...string) (string, string, error) {
	me.CallCount++
	if me.ExecuteFunc != nil {
		return me.ExecuteFunc(ctx, cmd, args...)
	}

	return "", "", fmt.Errorf("not implemented")
}

func TestInstall_Happy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		dep           *cryptoutilAppsToolsDevSetupConfig.Dependency
		executeErr    error
		expectErr     bool
		expectCmdName string
	}{
		{
			name: "Install via uv pip",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:       "pre-commit",
				InstallCmd: "uv pip install pre-commit",
			},
			executeErr: nil,
			expectErr:  false,
		},
		{
			name: "Install via go install",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:       "golangci-lint",
				InstallCmd: cryptoutilSharedMagic.DevSetupGolangciInstallCmd,
			},
			executeErr: nil,
			expectErr:  false,
		},
		{
			name: "Run pre-commit install",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:       "pre-commit-hooks",
				InstallCmd: cryptoutilSharedMagic.DevSetupGitHooksInstallCmd,
			},
			executeErr: nil,
			expectErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			executor := &MockExecutor{
				ExecuteFunc: func(ctx context.Context, cmd string, args ...string) (string, string, error) {
					return "", "", tc.executeErr
				},
			}

			installer := NewDependencyInstaller(executor, "/tmp/project")
			err := installer.Install(context.Background(), tc.dep)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, 1, executor.CallCount)
		})
	}
}

func TestInstall_Sad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dep        *cryptoutilAppsToolsDevSetupConfig.Dependency
		executeErr error
		expectErr  bool
	}{
		{
			name: "Installation fails",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:       "missing-tool",
				InstallCmd: "apt-get install missing-tool",
			},
			executeErr: fmt.Errorf("package not found"),
			expectErr:  true,
		},
		{
			name: "No install command defined",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:       cryptoutilSharedMagic.DevSetupTestInvalidToolName,
				InstallCmd: "",
			},
			expectErr: true,
		},
		{
			name: "Invalid install command format",
			dep: &cryptoutilAppsToolsDevSetupConfig.Dependency{
				Name:       cryptoutilSharedMagic.DevSetupTestInvalidToolName,
				InstallCmd: "",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			executor := &MockExecutor{
				ExecuteFunc: func(ctx context.Context, cmd string, args ...string) (string, string, error) {
					return cryptoutilSharedMagic.DevSetupTestSomeStderr, cryptoutilSharedMagic.DevSetupTestSomeStderr, tc.executeErr
				},
			}

			installer := NewDependencyInstaller(executor, "/tmp/project")
			err := installer.Install(context.Background(), tc.dep)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInstall_WithContext(t *testing.T) {
	t.Parallel()
	t.Run("Installation respects context deadline", func(t *testing.T) {
		executor := &MockExecutor{
			ExecuteFunc: func(ctx context.Context, cmd string, args ...string) (string, string, error) {
				// Check if context was passed correctly
				select {
				case <-ctx.Done():
					return "", "", ctx.Err()
				default:
					return "", "", nil
				}
			},
		}

		installer := NewDependencyInstaller(executor, "/tmp/project")
		dep := &cryptoutilAppsToolsDevSetupConfig.Dependency{
			Name:       "test",
			InstallCmd: "test-cmd",
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Simulate timeout

		err := installer.Install(ctx, dep)
		require.Error(t, err, "expected error from context")
	})
}

func TestSystemExecutor(t *testing.T) {
	t.Parallel()
	t.Run("Execute valid command", func(t *testing.T) {
		executor := NewSystemExecutor()
		// Use go version which is guaranteed to exist on systems where tests run
		stdout, stderr, err := executor.Execute(context.Background(), "go", cryptoutilSharedMagic.CLIVersionCommand)
		require.NoError(t, err)

		// Output might have newlines, just check it's not empty
		output := stdout + stderr
		require.NotEmpty(t, output, "expected output")
	})

	t.Run("Execute invalid command", func(t *testing.T) {
		executor := NewSystemExecutor()

		_, _, err := executor.Execute(context.Background(), "nonexistent-command-xyz-123", "arg")
		require.Error(t, err, "expected error for invalid command")
	})
}
