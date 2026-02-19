// Copyright (c) 2025 Justin Cranford

package lint_golangci

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"

	"github.com/stretchr/testify/require"
)

func TestLintGolangCIConfig_NoConfigFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{}

	err := LintGolangCIConfig(logger, filesByExtension)
	require.NoError(t, err, "lint should pass with no config files")
}

func TestLintGolangCIConfig_ValidV2Config(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".golangci.yml")

	validContent := `linters-settings:
  wsl_v5:
    allow-assign-and-anything: false
    allow-cuddle-declarations: false
  revive:
    rules:
      - name: exported
        disabled: true

linters:
  enable:
    - gofmt
    - govet
    - revive
    - staticcheck
`
	err := os.WriteFile(configFile, []byte(validContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {configFile},
	}

	err = LintGolangCIConfig(logger, filesByExtension)
	require.NoError(t, err, "lint should pass with valid v2 config")
}

func TestLintGolangCIConfig_DeprecatedOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name: "deprecated wsl key (should use wsl_v5)",
			content: `linters-settings:
  wsl:
    allow-assign-and-anything: false
`,
			wantErr:     true,
			errContains: "v2 compatibility violations",
		},
		{
			name: "deprecated force-err-cuddling option",
			content: `linters-settings:
  wsl_v5:
    force-err-cuddling: true
`,
			wantErr:     true,
			errContains: "v2 compatibility violations",
		},
		{
			name: "deprecated ignore-words option",
			content: `linters-settings:
  misspell:
    ignore-words:
      - color
`,
			wantErr:     true,
			errContains: "v2 compatibility violations",
		},
		{
			name: "deprecated deadcode linter",
			content: `linters:
  enable:
    - deadcode
    - gofmt
`,
			wantErr:     true,
			errContains: "v2 compatibility violations",
		},
		{
			name: "deprecated structcheck linter",
			content: `linters:
  enable:
    - structcheck
`,
			wantErr:     true,
			errContains: "v2 compatibility violations",
		},
		{
			name: "deprecated varcheck linter",
			content: `linters:
  enable:
    - varcheck
`,
			wantErr:     true,
			errContains: "v2 compatibility violations",
		},
		{
			name: "deprecated golint linter",
			content: `linters:
  enable:
    - golint
`,
			wantErr:     true,
			errContains: "v2 compatibility violations",
		},
		{
			name: "deprecated interfacer linter",
			content: `linters:
  enable:
    - interfacer
`,
			wantErr:     true,
			errContains: "v2 compatibility violations",
		},
		{
			name: "deprecated maligned linter",
			content: `linters:
  enable:
    - maligned
`,
			wantErr:     true,
			errContains: "v2 compatibility violations",
		},
		{
			name: "valid v2 config with wsl_v5",
			content: `linters-settings:
  wsl_v5:
    allow-cuddle-declarations: false
`,
			wantErr: false,
		},
		{
			name: "valid v2 config with revive",
			content: `linters:
  enable:
    - revive
    - gofmt
    - govet
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			configFile := filepath.Join(tempDir, ".golangci.yml")

			err := os.WriteFile(configFile, []byte(tt.content), 0o600)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			filesByExtension := map[string][]string{
				"yml": {configFile},
			}

			err = LintGolangCIConfig(logger, filesByExtension)

			if tt.wantErr {
				require.Error(t, err, "lint should fail for deprecated config")
				require.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err, "lint should pass for valid v2 config")
			}
		})
	}
}

func TestCheckGolangCIConfig_NonExistentFile(t *testing.T) {
	t.Parallel()

	violations, err := checkGolangCIConfig("/nonexistent/file.yml")
	require.Error(t, err, "should fail for non-existent file")
	require.Nil(t, violations)
}

func TestCheckGolangCIConfig_MultipleViolations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".golangci.yml")

	content := `linters-settings:
  wsl:
    force-err-cuddling: true
  misspell:
    ignore-words:
      - color

linters:
  enable:
    - deadcode
    - structcheck
`
	err := os.WriteFile(configFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := checkGolangCIConfig(configFile)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(violations), 4, "should detect multiple violations")
}

func TestFindGolangCIConfigFiles_AllFormats(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create test files.
	ymlFile := filepath.Join(tempDir, ".golangci.yml")
	yamlFile := filepath.Join(tempDir, "golangci.yaml")
	tomlFile := filepath.Join(tempDir, "golangci.toml")
	otherFile := filepath.Join(tempDir, "other.yml")

	for _, file := range []string{ymlFile, yamlFile, tomlFile, otherFile} {
		err := os.WriteFile(file, []byte("test"), 0o600)
		require.NoError(t, err)
	}

	filesByExtension := map[string][]string{
		"yml":  {ymlFile, otherFile},
		"yaml": {yamlFile},
		"toml": {tomlFile},
	}

	files := FindGolangCIConfigFiles(filesByExtension)
	require.Len(t, files, 3, "should find yml, yaml, and toml config files")
	require.Contains(t, files, ymlFile)
	require.Contains(t, files, yamlFile)
	require.Contains(t, files, tomlFile)
	require.NotContains(t, files, otherFile, "should not include non-golangci yml files")
}

// TestLintGolangCIConfig_FileReadError tests that LintGolangCIConfig logs a warning
// and continues when a config file cannot be opened (covers the warning+continue path).
func TestLintGolangCIConfig_FileReadError(t *testing.T) {
t.Parallel()

logger := cryptoutilCmdCicdCommon.NewLogger("test")
// Pass a nonexistent .golangci.yml path - FindGolangCIConfigFiles will include it
// but checkGolangCIConfig will fail to open it.
filesByExtension := map[string][]string{
"yml": {"/nonexistent/path/.golangci.yml"},
}

// Should not return error - file errors are logged as warnings and skipped.
err := LintGolangCIConfig(logger, filesByExtension)
require.NoError(t, err, "LintGolangCIConfig should continue after file open error")
}

// TestCheckGolangCIConfig_SectionExitOnTopLevelKey tests that section tracking is reset
// when a non-section top-level key is encountered (covers lines 136-140).
func TestCheckGolangCIConfig_SectionExitOnTopLevelKey(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
configFile := filepath.Join(tmpDir, ".golangci.yml")

// Config with linters-settings followed by a different top-level section.
// The 'severity:' key triggers the section-exit reset logic.
content := `linters-settings:
  wsl_v5:
    allow-assign-and-anything: false
severity:
  default-severity: minor
linters:
  enable:
    - gofmt
`
err := os.WriteFile(configFile, []byte(content), 0o600)
require.NoError(t, err)

violations, err := checkGolangCIConfig(configFile)
require.NoError(t, err, "checkGolangCIConfig should succeed with valid config")
require.Empty(t, violations, "Should have no violations with valid v2 config")
}

// TestCheckGolangCIConfig_CommentTopLevelKey tests that comment lines at top level
// do NOT reset section tracking (covers the false-branch of inner if).
func TestCheckGolangCIConfig_CommentTopLevelKey(t *testing.T) {
t.Parallel()

tmpDir := t.TempDir()
configFile := filepath.Join(tmpDir, ".golangci.yml")

// Config with a comment at top-level after linters-settings.
// The '# comment:' line should NOT exit sections.
content := `linters-settings:
  wsl_v5:
    allow-assign-and-anything: false
# comment: this should not exit sections
  wsl_v5:
    force-err-cuddling: false
`
err := os.WriteFile(configFile, []byte(content), 0o600)
require.NoError(t, err)

_, err = checkGolangCIConfig(configFile)
require.NoError(t, err, "checkGolangCIConfig should succeed")
}
