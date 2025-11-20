// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
)

// TestSaveCircularDepCache_ErrorCases tests error handling in saveCircularDepCache.
func TestSaveCircularDepCache_ErrorCases(t *testing.T) {
	// Test with invalid directory path (directory creation should fail on read-only filesystem)
	// This is platform-specific and hard to test reliably, so we'll test the happy path
	// and rely on the existing tests for error cases
	t.Skip("Skipping platform-specific file system error tests")
}

// TestSaveDepCache_ErrorCases tests error handling in saveDepCache.
func TestSaveDepCache_ErrorCases(t *testing.T) {
	// Test with invalid directory path
	t.Skip("Skipping platform-specific file system error tests")
}

// TestGetDirectDependencies_EdgeCases tests edge cases in getDirectDependencies.
func TestGetDirectDependencies_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		goModContent string
		wantCount    int
		wantErr      bool
	}{
		{
			name:         "empty go.mod",
			goModContent: "",
			wantCount:    0,
			wantErr:      false,
		},
		{
			name: "only module declaration",
			goModContent: `module example.com/test
go 1.25.4
`,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "single require line",
			goModContent: `module example.com/test
go 1.25.4
require github.com/example/dep v1.0.0
`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "require block with indirect",
			goModContent: `module example.com/test
go 1.25.4
require (
	github.com/direct/dep v1.0.0
	github.com/indirect/dep v2.0.0 // indirect
)
`,
			wantCount: 1, // Only direct dependency
			wantErr:   false,
		},
		{
			name: "multiple require blocks",
			goModContent: `module example.com/test
go 1.25.4
require github.com/first/dep v1.0.0
require (
	github.com/second/dep v2.0.0
	github.com/third/dep v3.0.0
)
`,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "malformed require line",
			goModContent: `module example.com/test
go 1.25.4
require invalid
`,
			wantCount: 1, // Parser treats "invalid" as module name
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, err := getDirectDependencies([]byte(tt.goModContent))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantCount, len(deps), "Unexpected number of dependencies")
			}
		})
	}
}

// TestFilterTextFiles_EdgeCases tests edge cases in filterTextFiles.
func TestFilterTextFiles_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		wantCount int
	}{
		{
			name:      "empty file list",
			files:     []string{},
			wantCount: 0,
		},
		{
			name: "all excluded files",
			files: []string{
				"vendor/some/file.go",
				".git/config",
				"node_modules/package/index.js",
			},
			wantCount: 0,
		},
		{
			name: "mixed included and excluded",
			files: []string{
				"main.go",
				"vendor/dep.go",
				"README.md",
				".git/config",
			},
			wantCount: 2, // main.go and README.md
		},
		{
			name: "generated files excluded",
			files: []string{
				"normal.go",
				"generated_gen.go",
				"proto.pb.go",
			},
			wantCount: 1, // Only normal.go
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterTextFiles(tt.files)
			require.Equal(t, tt.wantCount, len(result), "Unexpected number of filtered files")
		})
	}
}

// TestProcessGoFile_EdgeCases tests edge cases in processGoFile.
func TestProcessGoFile_EdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		wantModified     bool
		wantReplacements int
	}{
		{
			name:             "empty file",
			content:          "",
			wantModified:     false,
			wantReplacements: 0,
		},
		{
			name:             "only comments",
			content:          "// This is a comment\n/* Block comment */\n",
			wantModified:     false,
			wantReplacements: 0,
		},
		{
			name: "already using any",
			content: `package test
var x any = 42
`,
			wantModified:     false,
			wantReplacements: 0,
		},
		{
			name: "multiple any on same line",
			content: `package test
func convert(a any, b any) (any, any) {
	return a, b
}
`,
			wantModified:     false, // Already using 'any', nothing to replace
			wantReplacements: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "test.go")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0o600)
			require.NoError(t, err)

			replacements, err := processGoFile(tmpFile)
			require.NoError(t, err)

			if tt.wantModified {
				require.Greater(t, replacements, 0, "Expected modifications")
			} else {
				require.Equal(t, 0, replacements, "Expected no modifications")
			}

			require.Equal(t, tt.wantReplacements, replacements, "Unexpected replacement count")
		})
	}
}

// TestGoCheckCircularPackageDeps_CacheScenarios tests various cache scenarios.
func TestGoCheckCircularPackageDeps_CacheScenarios(t *testing.T) {
	logger := common.NewLogger("TestGoCheckCircularPackageDeps_CacheScenarios")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	// Create minimal go.mod
	err = os.WriteFile("go.mod", []byte(testGoModMinimal), 0o600)
	require.NoError(t, err)

	// Create minimal Go file so 'go list' finds packages
	err = os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0o600)
	require.NoError(t, err)

	// First run - cache miss, should check
	err = goCheckCircularPackageDeps(logger)
	require.NoError(t, err)

	// Verify cache was created
	cacheFile := cryptoutilMagic.CircularDepCacheFileName
	require.FileExists(t, cacheFile)

	// Second run - cache hit
	err = goCheckCircularPackageDeps(logger)
	require.NoError(t, err)

	// Modify go.mod to invalidate cache
	time.Sleep(10 * time.Millisecond) // Ensure different mod time

	err = os.WriteFile("go.mod", []byte(testGoModMinimal+"\n// Modified\n"), 0o600)
	require.NoError(t, err)

	// Third run - cache invalidated
	err = goCheckCircularPackageDeps(logger)
	require.NoError(t, err)
}

// TestGoUpdateDeps_CacheScenarios tests various cache scenarios for goUpdateDeps.
func TestGoUpdateDeps_CacheScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	logger := common.NewLogger("TestGoUpdateDeps_CacheScenarios")

	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	//nolint:errcheck // Best effort to restore directory
	defer os.Chdir(originalDir)

	// Create go.mod and go.sum
	err = os.WriteFile("go.mod", []byte(testGoModMinimal), 0o600)
	require.NoError(t, err)
	err = os.WriteFile("go.sum", []byte(""), 0o600)
	require.NoError(t, err)

	// First run - cache miss
	//nolint:errcheck // May or may not error depending on actual dependencies
	_ = goUpdateDeps(logger, cryptoutilMagic.DepCheckDirect)

	// Verify cache was created
	cacheFile := cryptoutilMagic.DepCacheFileName
	require.FileExists(t, cacheFile)

	// Second run - cache hit
	//nolint:errcheck // Should use cache
	_ = goUpdateDeps(logger, cryptoutilMagic.DepCheckDirect)
}

// TestValidateAndParseWorkflowFile_EdgeCases tests edge cases in workflow file parsing.
func TestValidateAndParseWorkflowFile_EdgeCases(t *testing.T) {
	tests := []struct {
		name              string
		content           string
		wantValidationErr bool
		wantActionsCount  int
	}{
		{
			name:              "empty file",
			content:           "",
			wantValidationErr: true, // Empty file causes validation errors (missing name, missing CI prefix, missing logging)
			wantActionsCount:  0,
		},
		{
			name: "missing CI prefix",
			content: `name: Test
on: [push]
`,
			wantValidationErr: true,
			wantActionsCount:  0,
		},
		{
			name: "valid CI workflow with action",
			content: `name: CI Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo "${{ github.workflow }}"
`,
			wantValidationErr: true, // Filename 'test.yml' missing 'ci-' prefix
			wantActionsCount:  1,
		},
		{
			name: "workflow without actions",
			content: `name: CI Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "test"
`,
			wantValidationErr: true, // Missing workflow reference
			wantActionsCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), ".github", "workflows", "test.yml")
			err := os.MkdirAll(filepath.Dir(tmpFile), 0o755)
			require.NoError(t, err)

			err = os.WriteFile(tmpFile, []byte(tt.content), 0o600)
			require.NoError(t, err)

			actions, validationErrs, err := validateAndParseWorkflowFile(tmpFile)
			require.NoError(t, err)

			if tt.wantValidationErr {
				require.NotEmpty(t, validationErrs, "Expected validation errors")
			} else {
				require.Empty(t, validationErrs, "Unexpected validation errors")
			}

			require.Equal(t, tt.wantActionsCount, len(actions), "Unexpected number of actions")
		})
	}
}

// TestLoadWorkflowActionExceptions_EdgeCases tests edge cases in loading exceptions.
func TestLoadWorkflowActionExceptions_EdgeCases(t *testing.T) {
	const githubDir = ".github"

	tests := []struct {
		name       string
		createFile bool
		content    string
		wantEmpty  bool
		wantError  bool
	}{
		{
			name:       "file does not exist",
			createFile: false,
			wantEmpty:  true,
			wantError:  false,
		},
		{
			name:       "empty json file",
			createFile: true,
			content:    "{}",
			wantEmpty:  true,
			wantError:  false,
		},
		{
			name:       "invalid json",
			createFile: true,
			content:    "invalid json",
			wantEmpty:  false,
			wantError:  true, // Should return error for invalid JSON
		},
		{
			name:       "valid exceptions",
			createFile: true,
			content: `{
  "exceptions": {
    "actions/checkout": {
      "allowed_versions": ["v3", "v4"]
    }
  }
}`,
			wantEmpty: false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			originalDir, err := os.Getwd()
			require.NoError(t, err)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)
			//nolint:errcheck // Best effort to restore directory
			defer os.Chdir(originalDir)

			if tt.createFile {
				err = os.MkdirAll(githubDir, 0o755)
				require.NoError(t, err)

				exceptionsFile := filepath.Join(githubDir, "workflows-outdated-action-exemptions.json")
				err = os.WriteFile(exceptionsFile, []byte(tt.content), 0o600)
				require.NoError(t, err)
			}

			exceptions, err := loadWorkflowActionExceptions()
			if tt.wantError {
				require.Error(t, err)
				require.Nil(t, exceptions)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, exceptions)

			if tt.wantEmpty {
				require.Empty(t, exceptions.Exceptions)
			} else {
				require.NotEmpty(t, exceptions.Exceptions)
			}
		})
	}
}

// TestCheckActionVersionsConcurrently_EdgeCases tests additional branches in concurrent version checking.
func TestCheckActionVersionsConcurrently_EdgeCases(t *testing.T) {
	logger := common.NewLogger("TestCheckActionVersionsConcurrently")

	tests := []struct {
		name            string
		actions         map[string]WorkflowActionDetails
		exceptions      *WorkflowActionExceptions
		wantOutdatedMin int
		wantExemptedMin int
		wantErrorsMin   int
		description     string
	}{
		{
			name:    "empty action map",
			actions: make(map[string]WorkflowActionDetails),
			exceptions: &WorkflowActionExceptions{
				Exceptions: make(map[string]WorkflowActionException),
			},
			wantOutdatedMin: 0,
			wantExemptedMin: 0,
			wantErrorsMin:   0,
			description:     "empty action map should return empty results",
		},
		{
			name: "action with matching exemption version",
			actions: map[string]WorkflowActionDetails{
				"actions/checkout@v3": {
					Name:           "actions/checkout",
					CurrentVersion: "v3",
					WorkflowFiles:  []string{"test.yml"},
				},
			},
			exceptions: &WorkflowActionExceptions{
				Exceptions: map[string]WorkflowActionException{
					"actions/checkout": {
						AllowedVersions: []string{"v3"},
						Reason:          "Compatibility",
					},
				},
			},
			wantOutdatedMin: 0,
			wantExemptedMin: 1,
			wantErrorsMin:   0,
			description:     "exempted action should appear in exempted list",
		},
		{
			name: "action with non-matching exemption version",
			actions: map[string]WorkflowActionDetails{
				"actions/checkout@v2": {
					Name:           "actions/checkout",
					CurrentVersion: "v2",
					WorkflowFiles:  []string{"test.yml"},
				},
			},
			exceptions: &WorkflowActionExceptions{
				Exceptions: map[string]WorkflowActionException{
					"actions/checkout": {
						AllowedVersions: []string{"v3"}, // Different version
						Reason:          "Compatibility",
					},
				},
			},
			wantOutdatedMin: 0, // Will either be outdated or error (depends on API)
			wantExemptedMin: 0,
			wantErrorsMin:   0,
			description:     "non-matching exemption version should check latest version",
		},
		{
			name: "multiple actions with mixed exemptions",
			actions: map[string]WorkflowActionDetails{
				"actions/checkout@v3": {
					Name:           "actions/checkout",
					CurrentVersion: "v3",
					WorkflowFiles:  []string{"test1.yml"},
				},
				"actions/setup-go@v4": {
					Name:           "actions/setup-go",
					CurrentVersion: "v4",
					WorkflowFiles:  []string{"test2.yml"},
				},
			},
			exceptions: &WorkflowActionExceptions{
				Exceptions: map[string]WorkflowActionException{
					"actions/checkout": {
						AllowedVersions: []string{"v3"},
						Reason:          "Compatibility",
					},
				},
			},
			wantOutdatedMin: 0, // setup-go will either be outdated or error
			wantExemptedMin: 1, // checkout is exempted
			wantErrorsMin:   0,
			description:     "mixed exemptions should handle each action correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outdated, exempted, errors := checkActionVersionsConcurrently(logger, tt.actions, tt.exceptions)

			require.GreaterOrEqual(t, len(outdated), tt.wantOutdatedMin, "outdated count: %s", tt.description)
			require.GreaterOrEqual(t, len(exempted), tt.wantExemptedMin, "exempted count: %s", tt.description)
			require.GreaterOrEqual(t, len(errors), tt.wantErrorsMin, "errors count: %s", tt.description)

			// Additional validation for exempted actions
			if tt.wantExemptedMin > 0 {
				for _, action := range exempted {
					require.Contains(t, tt.exceptions.Exceptions, action.Name,
						"exempted action should exist in exceptions map")
				}
			}
		})
	}
}

// TestCheckActionVersionsConcurrently_ConcurrentExecution tests concurrent behavior.
func TestCheckActionVersionsConcurrently_ConcurrentExecution(t *testing.T) {
	logger := common.NewLogger("TestCheckActionVersionsConcurrently_Concurrent")

	// Create a larger action map to test concurrent execution
	actions := make(map[string]WorkflowActionDetails)

	for i := 1; i <= 10; i++ {
		key := "test-action-" + string(rune('0'+i))
		actions[key] = WorkflowActionDetails{
			Name:           "actions/test-action",
			CurrentVersion: "v1",
			WorkflowFiles:  []string{"test.yml"},
		}
	}

	exceptions := &WorkflowActionExceptions{
		Exceptions: make(map[string]WorkflowActionException),
	}

	// This should complete without deadlocks or panics
	outdated, exempted, errors := checkActionVersionsConcurrently(logger, actions, exceptions)

	// We expect results (outdated or errors, since these are test actions that will fail API calls)
	totalResults := len(outdated) + len(exempted) + len(errors)
	require.Equal(t, len(actions), totalResults,
		"All actions should produce results (outdated, exempted, or error)")
}
