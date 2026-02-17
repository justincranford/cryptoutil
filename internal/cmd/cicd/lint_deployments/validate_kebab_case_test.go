package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateKebabCase_ValidServiceName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := `service:
  name: "cipher-im"
  version: "1.0.0"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"), []byte(content), 0o600))

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateKebabCase_InvalidServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		wantErr     string
	}{
		{name: "PascalCase", serviceName: "CipherIM", wantErr: "CipherIM"},
		{name: "snake_case", serviceName: "cipher_im", wantErr: "cipher_im"},
		{name: "UPPERCASE", serviceName: "CIPHER-IM", wantErr: "CIPHER-IM"},
		{name: "spaces", serviceName: "cipher im", wantErr: "cipher im"},
		{name: "leading hyphen", serviceName: "-cipher", wantErr: "-cipher"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			content := "service:\n  name: \"" + tc.serviceName + "\"\n"
			require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"), []byte(content), 0o600))

			result, err := ValidateKebabCase(dir)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.Valid)
			assert.True(t, len(result.Errors) > 0)
			assert.Contains(t, result.Errors[0], tc.wantErr)
			assert.Contains(t, result.Errors[0], "kebab-case")
		})
	}
}

func TestValidateKebabCase_PathDoesNotExist(t *testing.T) {
	t.Parallel()

	result, err := ValidateKebabCase(filepath.Join(t.TempDir(), "nonexistent"))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "does not exist")
}

func TestValidateKebabCase_PathIsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o600))

	result, err := ValidateKebabCase(filePath)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "not a directory")
}

func TestValidateKebabCase_MissingServiceNameField(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := "observability:\n  enabled: true\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"), []byte(content), 0o600))

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid, "Missing field should not fail validation")
	assert.Empty(t, result.Errors)
}

func TestValidateKebabCase_NonYAMLFilesIgnored(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "data.json"), []byte("{}"), 0o600))

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid)
}

func TestValidateKebabCase_ComposeFilesSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Compose files have service names validated by ValidateNaming, not ValidateKebabCase.
	content := "services:\n  Invalid_Name:\n    image: nginx\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(content), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(content), 0o600))

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid, "Compose files should be skipped by ValidateKebabCase")
}

func TestValidateKebabCase_InvalidYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.yml"), []byte("invalid: [yaml: {"), 0o600))

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, len(result.Warnings) > 0)
	assert.Contains(t, result.Warnings[0], "Failed to parse")
}

func TestValidateKebabCase_UnreadableFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "config.yml")
	require.NoError(t, os.WriteFile(filePath, []byte("service:\n  name: ok\n"), 0o600))
	require.NoError(t, os.Chmod(filePath, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(filePath, 0o600)
	})

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, len(result.Warnings) > 0)
	assert.Contains(t, result.Warnings[0], "Failed to read")
}

func TestValidateKebabCase_NestedDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	nested := filepath.Join(dir, "sub", "deep")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	content := "service:\n  name: \"my-nested-service\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(nested, "config.yml"), []byte(content), 0o600))

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid)
}

func TestValidateKebabCase_EmptyDirectory(t *testing.T) {
	t.Parallel()

	result, err := ValidateKebabCase(t.TempDir())
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Valid)
}

func TestValidateKebabCase_WalkError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subDir := filepath.Join(dir, "locked")
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "config.yml"), []byte("key: val"), 0o600))
	require.NoError(t, os.Chmod(subDir, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(subDir, 0o755)
	})

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, len(result.Warnings) > 0)
	assert.Contains(t, result.Warnings[0], "error accessing")
}

func TestValidateKebabCase_ServiceNameNotString(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := "service:\n  name: 123\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"), []byte(content), 0o600))

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Integer value should be silently skipped (not a string).
	assert.True(t, result.Valid)
}

func TestGetNestedField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    map[string]any
		fieldPath string
		want      string
	}{
		{
			name:      "simple field",
			config:    map[string]any{"name": "value"},
			fieldPath: "name",
			want:      "value",
		},
		{
			name:      "nested field",
			config:    map[string]any{"service": map[string]any{"name": "my-svc"}},
			fieldPath: "service.name",
			want:      "my-svc",
		},
		{
			name:      "deeply nested",
			config:    map[string]any{"a": map[string]any{"b": map[string]any{"c": "deep"}}},
			fieldPath: "a.b.c",
			want:      "deep",
		},
		{
			name:      "missing field",
			config:    map[string]any{"other": "value"},
			fieldPath: "service.name",
			want:      "",
		},
		{
			name:      "missing intermediate",
			config:    map[string]any{"service": "not-a-map"},
			fieldPath: "service.name",
			want:      "",
		},
		{
			name:      "non-string value",
			config:    map[string]any{"count": 42},
			fieldPath: "count",
			want:      "",
		},
		{
			name:      "nil map",
			config:    nil,
			fieldPath: "anything",
			want:      "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := getNestedField(tc.config, tc.fieldPath)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFormatKebabCaseValidationResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *KebabCaseValidationResult
		contains []string
	}{
		{
			name:     "passing result",
			result:   &KebabCaseValidationResult{Path: "/test", Valid: true},
			contains: []string{"/test", "PASS"},
		},
		{
			name: "failing result",
			result: &KebabCaseValidationResult{
				Path: "/test", Valid: false,
				Errors:   []string{"field error"},
				Warnings: []string{"warning msg"},
			},
			contains: []string{"FAIL", "field error", "warning msg"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatKebabCaseValidationResult(tc.result)
			for _, s := range tc.contains {
				assert.Contains(t, output, s)
			}
		})
	}
}
