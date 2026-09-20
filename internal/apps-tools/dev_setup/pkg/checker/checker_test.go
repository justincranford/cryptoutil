package checker

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
}

func (me *MockExecutor) Execute(ctx context.Context, cmd string, args ...string) (string, string, error) {
	if me.ExecuteFunc != nil {
		return me.ExecuteFunc(ctx, cmd, args...)
	}

	return "", "", fmt.Errorf("not implemented")
}

func TestCheck_Happy(t *testing.T) {
	tests := []struct {
		name          string
		dep           *config.Dependency
		executeOutput string
		executeErr    error
		expectFound   bool
		expectVersion string
	}{
		{
			name: "Tool installed with version",
			dep: &config.Dependency{
				Name:             "python",
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
			dep: &config.Dependency{
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
			dep: &config.Dependency{
				Name:     "pre-commit",
				CheckCmd: "pre-commit --version",
			},
			executeOutput: "pre-commit 3.0.0",
			executeErr:    nil,
			expectFound:   true,
			expectVersion: "",
		},
		{
			name: "Tool version with v prefix",
			dep: &config.Dependency{
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
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if found != tc.expectFound {
				t.Errorf("expected found=%v, got %v", tc.expectFound, found)
			}

			if tc.dep.DetectVersionCmd != "" && version != tc.expectVersion {
				t.Errorf("expected version=%q, got %q", tc.expectVersion, version)
			}
		})
	}
}

func TestCheck_Sad(t *testing.T) {
	tests := []struct {
		name        string
		dep         *config.Dependency
		executeErr  error
		expectFound bool
		expectErr   bool
	}{
		{
			name: "Tool not installed",
			dep: &config.Dependency{
				Name:     "missing-tool",
				CheckCmd: "missing-tool --version",
			},
			executeErr:  fmt.Errorf("command not found"),
			expectFound: false,
			expectErr:   false,
		},
		{
			name: "No check command defined",
			dep: &config.Dependency{
				Name:     cryptoutilSharedMagic.DevSetupTestInvalidName,
				CheckCmd: "",
			},
			expectErr: true,
		},
		{
			name: "Invalid check command",
			dep: &config.Dependency{
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

			if tc.expectErr && err == nil {
				t.Errorf("expected error, got nil")
			}

			if !tc.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tc.expectErr && found != tc.expectFound {
				t.Errorf("expected found=%v, got %v", tc.expectFound, found)
			}
		})
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{cryptoutilSharedMagic.DevSetupTestVersion123, cryptoutilSharedMagic.DevSetupTestVersion123},
		{"Python 3.14.0", cryptoutilSharedMagic.DevSetupTestPythonVersionParsed},
		{"go version go1.26.1 linux/amd64", cryptoutilSharedMagic.DevSetupTestGoVersionParsed},
		{"version 2.12.2", cryptoutilSharedMagic.DevSetupTestGolangciVersion},
		{"v1.0.0-alpha", "1.0.0-alpha"},
		{"golangci-lint has version v2.12.2", cryptoutilSharedMagic.DevSetupTestGolangciVersion},
		{cryptoutilSharedMagic.DevSetupTestNoVersionHere, cryptoutilSharedMagic.DevSetupTestNoVersionHere},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := extractVersion(tc.input)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}
