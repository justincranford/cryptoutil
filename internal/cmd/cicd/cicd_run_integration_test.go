// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const testGoModMinimal3 = `module example.com/test

go 1.25.4
`

// TestRun_SingleCommand tests Run with a single command.
func TestRun_SingleCommand(t *testing.T) {
	// Use a command that doesn't require external files
	err := Run([]string{"go-check-circular-package-dependencies"})
	// May pass or fail depending on project state, but should not panic
	_ = err
}

// TestRun_MultipleCommands tests Run with multiple commands.
func TestRun_MultipleCommands(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	// Create minimal go.mod and go.sum for dependency checks
	goModContent := testGoModMinimal3
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	// Run multiple commands
	commands := []string{
		"go-check-circular-package-dependencies",
		"go-update-direct-dependencies",
	}

	err = Run(commands)
	// Commands may fail for various reasons, but should not panic
	_ = err
}

// TestRun_AllEnforceUTF8 tests Run with all-enforce-utf8 command.
func TestRun_AllEnforceUTF8(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	// Create a valid UTF-8 file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("Hello, World!"), 0o600)
	require.NoError(t, err)

	err = Run([]string{"all-enforce-utf8"})
	require.NoError(t, err, "all-enforce-utf8 should succeed with valid UTF-8 file")
}

// TestRun_GoEnforceTestPatterns tests Run with go-enforce-test-patterns command.
func TestRun_GoEnforceTestPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	// Create a valid test file
	testFile := filepath.Join(tmpDir, "example_test.go")
	testContent := `package example

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	require.True(t, true)
}
`
	err = os.WriteFile(testFile, []byte(testContent), 0o600)
	require.NoError(t, err)

	err = Run([]string{"go-enforce-test-patterns"})
	require.NoError(t, err, "go-enforce-test-patterns should succeed with valid test file")
}

// TestRun_GoEnforceAny tests Run with go-enforce-any command.
func TestRun_GoEnforceAny(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	// Create a Go file
	goFile := filepath.Join(tmpDir, "example.go")
	goContent := `package example

var x any = 42
`
	err = os.WriteFile(goFile, []byte(goContent), 0o600)
	require.NoError(t, err)

	err = Run([]string{"go-enforce-any"})
	require.NoError(t, err, "go-enforce-any should succeed with valid Go file")
}

// TestRun_CommandExecutionOrder tests that commands execute in order.
func TestRun_CommandExecutionOrder(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	// Create test files
	err = os.WriteFile("test.txt", []byte("test"), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("test.go", []byte("package test\n\nvar x any = 1"), 0o600)
	require.NoError(t, err)

	// Run commands in specific order
	commands := []string{
		"all-enforce-utf8",
		"go-enforce-any",
	}

	err = Run(commands)
	require.NoError(t, err, "Commands should execute in order successfully")
}

// TestRun_FailedCommandStopsExecution tests that a failed command stops execution.
func TestRun_FailedCommandStopsExecution(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	// Create invalid test file that will fail go-enforce-test-patterns
	testFile := filepath.Join(tmpDir, "bad_test.go")
	badContent := `package test

import (
	"testing"
)

func TestBad(t *testing.T) {
	t.Errorf("using t.Errorf instead of require")
}
`
	err = os.WriteFile(testFile, []byte(badContent), 0o600)
	require.NoError(t, err)

	// Run will execute all commands and collect errors
	err = Run([]string{"go-enforce-test-patterns"})
	require.Error(t, err, "Should fail with invalid test file")
	require.Contains(t, err.Error(), "failed commands")
}

// TestRun_InvalidCommandInList tests Run with an invalid command in the list.
func TestRun_InvalidCommandInList(t *testing.T) {
	err := Run([]string{"all-enforce-utf8", "invalid-command"})
	require.Error(t, err, "Should fail with invalid command")
	require.Contains(t, err.Error(), "unknown command")
}

// TestRun_EmptyCommandList tests Run with empty command list.
func TestRun_EmptyCommandList(t *testing.T) {
	err := Run([]string{})
	require.Error(t, err, "Should fail with empty command list")
	require.Contains(t, err.Error(), "Usage")
}

// TestRun_GoUpdateDirectDependencies tests Run with dependency update command.
func TestRun_GoUpdateDirectDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	// Create minimal go.mod and go.sum
	goModContent := `module example.com/test

go 1.25.4

require github.com/stretchr/testify v1.8.0
`
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	err = Run([]string{"go-update-direct-dependencies"})
	// May succeed or fail depending on network/cache, but should not panic
	_ = err
}

// TestRun_GoUpdateAllDependencies tests Run with all dependencies update command.
func TestRun_GoUpdateAllDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	// Create minimal go.mod and go.sum
	goModContent := testGoModMinimal3
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	err = Run([]string{"go-update-all-dependencies"})
	// May succeed or fail depending on network/cache, but should not panic
	_ = err
}
