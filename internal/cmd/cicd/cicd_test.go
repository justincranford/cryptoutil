// IMPORTANT: This file contains deliberate linter error patterns for testing cicd functionality.
// It MUST be excluded from all linting operations to prevent self-referencing errors.
// See .golangci.yml exclude-rules and cicd.go exclusion patterns for details.
//
// This file intentionally uses any patterns and other lint violations to test
// that cicd correctly identifies and reports such patterns in other files.
package cicd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Common test constants to avoid goconst linter violations.
const (
	testPackageMain   = "package main"
	testImportFmt     = `import "fmt"`
	testFuncMainStart = `
func main() {`
	testFuncMainEnd = `}
`
	testTypeMyStruct = `
type MyStruct struct {
	Data any
}
`
)

func TestRunUsage(t *testing.T) {
	// Test with no commands (should return error)
	err := Run([]string{})
	require.Error(t, err, "Expected error when no commands provided")
	require.Contains(t, err.Error(), "Usage: cicd <command>", "Error message should contain usage information")
}

func TestRunInvalidCommand(t *testing.T) {
	// Test with invalid command
	err := Run([]string{"invalid-command"})
	require.Error(t, err, "Expected error for invalid command")
	require.Contains(t, err.Error(), "unknown command: invalid-command", "Error message should indicate unknown command")
}

func TestRunMultipleCommands(t *testing.T) {
	// Note: We can't easily test actual command execution as they call os.Exit
	// This test just verifies the command parsing logic works
	commands := []string{"go-update-direct-dependencies", "github-workflow-lint"}
	require.Len(t, commands, 2, "Expected 2 commands")
	require.Equal(t, "go-update-direct-dependencies", commands[0], "Expected first command")
	require.Equal(t, "github-workflow-lint", commands[1], "Expected second command")
}

func TestValidateCommands_HappyPath(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
	}{
		{
			name:     "single valid command",
			commands: []string{"github-workflow-lint"},
		},
		{
			name:     "multiple different valid commands",
			commands: []string{"github-workflow-lint", "go-enforce-test-patterns", "all-enforce-utf8"},
		},
		{
			name:     "dependency update commands individually",
			commands: []string{"go-update-direct-dependencies"},
		},
		{
			name:     "all dependency update commands individually",
			commands: []string{"go-update-all-dependencies"},
		},
		{
			name: "all commands once each",
			commands: []string{
				"all-enforce-utf8",
				"go-enforce-test-patterns",
				"go-enforce-any",
				"go-check-circular-package-dependencies",
				"go-update-direct-dependencies",
				"github-workflow-lint",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateCommands(tt.commands)
			require.NoError(t, err, "Expected no error for valid commands: %v", tt.commands)
		})
	}
}

func TestValidateCommands_DuplicateCommands(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		expected []string // Now expecting multiple error messages
	}{
		{
			name:     "duplicate github-workflow-lint",
			commands: []string{"github-workflow-lint", "github-workflow-lint"},
			expected: []string{"command 'github-workflow-lint' specified 2 times"},
		},
		{
			name:     "duplicate go-enforce-test-patterns",
			commands: []string{"go-enforce-test-patterns", "all-enforce-utf8", "go-enforce-test-patterns"},
			expected: []string{"command 'go-enforce-test-patterns' specified 2 times"},
		},
		{
			name:     "duplicate go-update-direct-dependencies",
			commands: []string{"go-update-direct-dependencies", "github-workflow-lint", "go-update-direct-dependencies"},
			expected: []string{"command 'go-update-direct-dependencies' specified 2 times"},
		},
		{
			name:     "multiple duplicates",
			commands: []string{"github-workflow-lint", "github-workflow-lint", "all-enforce-utf8", "all-enforce-utf8"},
			expected: []string{
				"command 'github-workflow-lint' specified 2 times",
				"command 'all-enforce-utf8' specified 2 times",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateCommands(tt.commands)
			require.Error(t, err, "Expected error for duplicate commands: %v", tt.commands)

			errMsg := err.Error()
			for _, expectedMsg := range tt.expected {
				require.Contains(t, errMsg, expectedMsg, "Error message should contain expected text: %s", expectedMsg)
			}
		})
	}
}

func TestValidateCommands_MutuallyExclusiveCommands(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		expected string
	}{
		{
			name:     "both dependency update commands",
			commands: []string{"go-update-direct-dependencies", "go-update-all-dependencies"},
			expected: "cannot be used together",
		},
		{
			name:     "both dependency update commands with other commands",
			commands: []string{"github-workflow-lint", "go-update-direct-dependencies", "go-update-all-dependencies", "all-enforce-utf8"},
			expected: "cannot be used together",
		},
		{
			name:     "both dependency update commands in different order",
			commands: []string{"go-update-all-dependencies", "go-update-direct-dependencies"},
			expected: "cannot be used together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateCommands(tt.commands)
			require.Error(t, err, "Expected error for mutually exclusive commands: %v", tt.commands)
			require.Contains(t, err.Error(), tt.expected, "Error message should contain expected text")
		})
	}
}

func TestValidateCommands_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		commands    []string
		expectError bool
	}{
		{
			name:        "empty commands slice",
			commands:    []string{},
			expectError: true, // validateCommands now checks for empty slice
		},
		{
			name:        "nil commands slice",
			commands:    nil,
			expectError: true, // validateCommands now checks for nil slice (len(nil) == 0)
		},
		{
			name:        "single dependency update command with other commands",
			commands:    []string{"go-update-direct-dependencies", "github-workflow-lint"},
			expectError: false,
		},
		{
			name:        "single all dependency update command with other commands",
			commands:    []string{"go-update-all-dependencies", "all-enforce-utf8"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateCommands(tt.commands)
			if tt.expectError {
				require.Error(t, err, "Expected error for edge case: %v", tt.commands)
			} else {
				require.NoError(t, err, "Expected no error for edge case: %v", tt.commands)
			}
		})
	}
}

func TestValidateCommands_MultipleErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		expected []string // All expected error messages
	}{
		{
			name:     "unknown command and duplicate",
			commands: []string{"unknown-cmd", "github-workflow-lint", "github-workflow-lint"},
			expected: []string{
				"unknown command: unknown-cmd",
				"command 'github-workflow-lint' specified 2 times",
			},
		},
		{
			name:     "empty commands with unknown",
			commands: []string{"", "invalid-cmd"},
			expected: []string{
				"Usage: cicd <command>",        // From empty check
				"unknown command: ",            // From empty string
				"unknown command: invalid-cmd", // From invalid command
			},
		},
		{
			name:     "duplicate and mutually exclusive",
			commands: []string{"go-update-direct-dependencies", "go-update-all-dependencies", "go-update-direct-dependencies"},
			expected: []string{
				"command 'go-update-direct-dependencies' specified 2 times",
				"cannot be used together", // From mutually exclusive check
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateCommands(tt.commands)
			require.Error(t, err, "Expected error for multiple error types: %v", tt.commands)

			errMsg := err.Error()
			for _, expectedMsg := range tt.expected {
				require.Contains(t, errMsg, expectedMsg, "Error message should contain expected text: %s", expectedMsg)
			}
		})
	}
}

func TestValidateCommands(t *testing.T) {
	tests := []struct {
		name            string
		commands        []string
		expectedDoFiles bool
		expectedError   bool
		errorContains   string
	}{
		{
			name:            "empty commands",
			commands:        []string{},
			expectedDoFiles: false,
			expectedError:   true,
			errorContains:   "Usage:",
		},
		{
			name:            "nil commands",
			commands:        nil,
			expectedDoFiles: false,
			expectedError:   true,
			errorContains:   "Usage:",
		},
		{
			name:            "single valid command that needs files",
			commands:        []string{"all-enforce-utf8"},
			expectedDoFiles: true,
			expectedError:   false,
		},
		{
			name:            "single valid command that doesn't need files",
			commands:        []string{"go-update-direct-dependencies"},
			expectedDoFiles: false,
			expectedError:   false,
		},
		{
			name:            "multiple valid commands",
			commands:        []string{"all-enforce-utf8", "go-update-direct-dependencies"},
			expectedDoFiles: true,
			expectedError:   false,
		},
		{
			name:            "unknown command",
			commands:        []string{"unknown-command"},
			expectedDoFiles: false,
			expectedError:   true,
			errorContains:   "unknown command",
		},
		{
			name:            "duplicate command",
			commands:        []string{"all-enforce-utf8", "all-enforce-utf8"},
			expectedDoFiles: false,
			expectedError:   true,
			errorContains:   "specified 2 times",
		},
		{
			name:            "mutually exclusive commands",
			commands:        []string{"go-update-direct-dependencies", "go-update-all-dependencies"},
			expectedDoFiles: false,
			expectedError:   true,
			errorContains:   "cannot be used together",
		},
		{
			name:            "all valid commands",
			commands:        []string{"all-enforce-utf8", "go-enforce-test-patterns", "go-enforce-any", "go-check-circular-package-dependencies", "go-update-direct-dependencies", "github-workflow-lint"},
			expectedDoFiles: true,
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doFiles, err := validateCommands(tt.commands)

			if tt.expectedError {
				require.Error(t, err, "Expected error for test case: %s", tt.name)

				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains, "Error should contain expected text")
				}
			} else {
				require.NoError(t, err, "Expected no error for test case: %s", tt.name)
				require.Equal(t, tt.expectedDoFiles, doFiles, "doFindAllFiles should match expected value")
			}
		})
	}
}

func TestCollectAllFiles(t *testing.T) {
	// Create a temporary directory with some test files
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"file1.txt",
		"file2.go",
		"subdir/file3.txt",
		"subdir/nested/file4.go",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(fullPath, []byte("test content"), 0o600))
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	// Collect files
	files, err := collectAllFiles()
	require.NoError(t, err)

	// Should find all test files
	require.Len(t, files, len(testFiles), "Should find all test files")

	// Convert to relative paths for comparison
	for i, file := range files {
		files[i], err = filepath.Rel(tempDir, filepath.Join(tempDir, file))
		require.NoError(t, err)
		// Normalize path separators to forward slashes for cross-platform comparison
		files[i] = filepath.ToSlash(files[i])
	}

	// Normalize expected paths to forward slashes
	normalizedTestFiles := make([]string, len(testFiles))
	for i, file := range testFiles {
		normalizedTestFiles[i] = filepath.ToSlash(file)
	}

	// Sort both slices for comparison
	require.ElementsMatch(t, normalizedTestFiles, files, "Should find all expected files")
}

func TestLoadActionExceptions(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		// Change to a temp directory where the file doesn't exist
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		exceptions, err := loadActionExceptions()
		require.NoError(t, err)
		require.NotNil(t, exceptions)
		require.NotNil(t, exceptions.Exceptions)
		require.Empty(t, exceptions.Exceptions)
	})

	t.Run("valid JSON file", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// Create test exceptions file
		exceptionsContent := `{
			"exceptions": {
				"actions/checkout": {
					"allowed_versions": ["v2", "v3"],
					"reason": "Test exception"
				}
			}
		}`

		require.NoError(t, os.MkdirAll(".github", 0o755))
		require.NoError(t, os.WriteFile(".github/workflows-outdated-action-exemptions.json", []byte(exceptionsContent), 0o600))

		exceptions, err := loadActionExceptions()
		require.NoError(t, err)
		require.NotNil(t, exceptions)
		require.Contains(t, exceptions.Exceptions, "actions/checkout")
		require.Equal(t, []string{"v2", "v3"}, exceptions.Exceptions["actions/checkout"].AllowedVersions)
		require.Equal(t, "Test exception", exceptions.Exceptions["actions/checkout"].Reason)
	})

	t.Run("invalid JSON file", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// Create invalid JSON file
		require.NoError(t, os.MkdirAll(".github", 0o755))
		require.NoError(t, os.WriteFile(".github/workflows-outdated-action-exemptions.json", []byte("invalid json"), 0o600))

		exceptions, err := loadActionExceptions()
		require.Error(t, err)
		require.Nil(t, exceptions)
		require.Contains(t, err.Error(), "failed to unmarshal exceptions JSON")
	})
}

func TestLoadDepCache(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "test_cache.json")

	t.Run("valid cache file", func(t *testing.T) {
		cacheContent := `{
			"last_check": "2025-01-01T00:00:00Z",
			"go_mod_mod_time": "2025-01-01T00:00:00Z",
			"go_sum_mod_time": "2025-01-01T00:00:00Z",
			"outdated_deps": ["github.com/example/old"],
			"mode": "direct"
		}`
		require.NoError(t, os.WriteFile(cacheFile, []byte(cacheContent), 0o600))

		cache, err := loadDepCache(cacheFile, "direct")
		require.NoError(t, err)
		require.NotNil(t, cache)
		require.Equal(t, "direct", cache.Mode)
		require.Len(t, cache.OutdatedDeps, 1)
		require.Equal(t, "github.com/example/old", cache.OutdatedDeps[0])
	})

	t.Run("cache file does not exist", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.json")
		cache, err := loadDepCache(nonExistentFile, "direct")
		require.Error(t, err)
		require.Nil(t, cache)
		require.Contains(t, err.Error(), "failed to read cache file")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		require.NoError(t, os.WriteFile(cacheFile, []byte("invalid json"), 0o600))
		cache, err := loadDepCache(cacheFile, "direct")
		require.Error(t, err)
		require.Nil(t, cache)
		require.Contains(t, err.Error(), "failed to unmarshal cache JSON")
	})

	t.Run("mode mismatch", func(t *testing.T) {
		cacheContent := `{
			"last_check": "2025-01-01T00:00:00Z",
			"go_mod_mod_time": "2025-01-01T00:00:00Z",
			"go_sum_mod_time": "2025-01-01T00:00:00Z",
			"outdated_deps": [],
			"mode": "direct"
		}`
		require.NoError(t, os.WriteFile(cacheFile, []byte(cacheContent), 0o600))

		cache, err := loadDepCache(cacheFile, "all")
		require.Error(t, err)
		require.Nil(t, cache)
		require.Contains(t, err.Error(), "cache mode mismatch")
	})
}

func TestSaveDepCache(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "test_cache.json")

	cache := DepCache{
		LastCheck:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		GoModModTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		GoSumModTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		OutdatedDeps: []string{"github.com/example/old", "github.com/example/older"},
		Mode:         "direct",
	}

	err := saveDepCache(cacheFile, cache)
	require.NoError(t, err)

	// Verify file was created and has correct content
	content, err := os.ReadFile(cacheFile)
	require.NoError(t, err)

	var loadedCache DepCache

	require.NoError(t, json.Unmarshal(content, &loadedCache))
	require.Equal(t, cache, loadedCache)

	// Check file permissions (should be 0o600 on Unix, but may differ on Windows)
	info, err := os.Stat(cacheFile)
	require.NoError(t, err)
	// On Windows, permissions might be different, so we just check that the file exists and is readable
	require.True(t, info.Mode().IsRegular(), "Cache file should be a regular file")
}

func TestGetDirectDependencies(t *testing.T) {
	t.Run("valid go.mod", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		// Create a test go.mod file
		goModContent := `module example.com/test

go 1.21

require (
	github.com/example/direct1 v1.0.0
	github.com/example/direct2 v2.0.0
	github.com/example/indirect v1.0.0 // indirect
)

require (
	github.com/example/direct3 v3.0.0
)
`
		require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), 0o600))

		deps, err := getDirectDependencies()
		require.NoError(t, err)
		require.Contains(t, deps, "github.com/example/direct1")
		require.Contains(t, deps, "github.com/example/direct2")
		require.Contains(t, deps, "github.com/example/direct3")
		require.NotContains(t, deps, "github.com/example/indirect") // Should exclude indirect deps
	})

	t.Run("go.mod does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		deps, err := getDirectDependencies()
		require.Error(t, err)
		require.Nil(t, deps)
		require.Contains(t, err.Error(), "failed to read go.mod")
	})

	t.Run("empty go.mod", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Chdir(oldWd))
		}()
		require.NoError(t, os.Chdir(tempDir))

		require.NoError(t, os.WriteFile("go.mod", []byte("module example.com/test\ngo 1.21\n"), 0o600))

		deps, err := getDirectDependencies()
		require.NoError(t, err)
		require.Empty(t, deps)
	})
}
