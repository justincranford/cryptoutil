package installer

import (
	"context"
	"cryptoutil/internal/apps-tools/dev_setup/pkg/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"testing"
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
	tests := []struct {
		name          string
		dep           *config.Dependency
		executeErr    error
		expectErr     bool
		expectCmdName string
	}{
		{
			name: "Install via uv pip",
			dep: &config.Dependency{
				Name:       "pre-commit",
				InstallCmd: "uv pip install pre-commit",
			},
			executeErr: nil,
			expectErr:  false,
		},
		{
			name: "Install via go install",
			dep: &config.Dependency{
				Name:       "golangci-lint",
				InstallCmd: "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2",
			},
			executeErr: nil,
			expectErr:  false,
		},
		{
			name: "Run pre-commit install",
			dep: &config.Dependency{
				Name:       "pre-commit-hooks",
				InstallCmd: "pre-commit install",
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

			if tc.expectErr && err == nil {
				t.Errorf("expected error, got nil")
			}

			if !tc.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if executor.CallCount != 1 {
				t.Errorf("expected 1 call, got %d", executor.CallCount)
			}
		})
	}
}

func TestInstall_Sad(t *testing.T) {
	tests := []struct {
		name       string
		dep        *config.Dependency
		executeErr error
		expectErr  bool
	}{
		{
			name: "Installation fails",
			dep: &config.Dependency{
				Name:       "missing-tool",
				InstallCmd: "apt-get install missing-tool",
			},
			executeErr: fmt.Errorf("package not found"),
			expectErr:  true,
		},
		{
			name: "No install command defined",
			dep: &config.Dependency{
				Name:       cryptoutilSharedMagic.DevSetupTestInvalidToolName,
				InstallCmd: "",
			},
			expectErr: true,
		},
		{
			name: "Invalid install command format",
			dep: &config.Dependency{
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

			if tc.expectErr && err == nil {
				t.Errorf("expected error, got nil")
			}

			if !tc.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestInstall_WithContext(t *testing.T) {
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
		dep := &config.Dependency{
			Name:       "test",
			InstallCmd: "test-cmd",
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Simulate timeout

		err := installer.Install(ctx, dep)
		if err == nil {
			t.Errorf("expected error from context, got nil")
		}
	})
}

func TestSystemExecutor(t *testing.T) {
	t.Run("Execute valid command", func(t *testing.T) {
		executor := NewSystemExecutor()
		// Use go version which is guaranteed to exist on systems where tests run
		stdout, stderr, err := executor.Execute(context.Background(), "go", "version")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Output might have newlines, just check it's not empty
		output := stdout + stderr
		if output == "" {
			t.Errorf("expected output, got empty string")
		}
	})

	t.Run("Execute invalid command", func(t *testing.T) {
		executor := NewSystemExecutor()

		_, _, err := executor.Execute(context.Background(), "nonexistent-command-xyz-123", "arg")
		if err == nil {
			t.Errorf("expected error for invalid command, got nil")
		}
	})
}
