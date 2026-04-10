package lint_deployments

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestValidateKebabCase_ValidCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFn func(t *testing.T) string
	}{
		{
			name: "valid service name",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"),
					[]byte("service:\n  name: \"sm-im\"\n  version: \"1.0.0\"\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "missing service name field",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"),
					[]byte("observability:\n  enabled: true\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "non-YAML files ignored",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "data.json"), []byte("{}"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "compose files skipped",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				content := []byte("services:\n  Invalid_Name:\n    image: nginx\n")
				require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), content, cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-compose.yml"), content, cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "nested directories",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				nested := filepath.Join(dir, cryptoutilSharedMagic.ClaimSub, "deep")
				require.NoError(t, os.MkdirAll(nested, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(filepath.Join(nested, "config.yml"),
					[]byte("service:\n  name: \"my-nested-service\"\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "empty directory",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return t.TempDir()
			},
		},
		{
			name: "service name not string",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"),
					[]byte("service:\n  name: 123\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := ValidateKebabCase(tc.setupFn(t))
			require.NoError(t, err)
			require.NotNil(t, result)
			require.True(t, result.Valid)
			require.Empty(t, result.Errors)
		})
	}
}

func TestValidateKebabCase_InvalidServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		wantErr     string
	}{
		{name: "PascalCase", serviceName: "SmIM", wantErr: "SmIM"},
		{name: "snake_case", serviceName: "sm_im", wantErr: "sm_im"},
		{name: "UPPERCASE", serviceName: "SM-IM", wantErr: "SM-IM"},
		{name: "spaces", serviceName: "sm im", wantErr: "sm im"},
		{name: "leading hyphen", serviceName: "-sm-im", wantErr: "-sm-im"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			content := "service:\n  name: \"" + tc.serviceName + "\"\n"
			require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

			result, err := ValidateKebabCase(dir)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.False(t, result.Valid)
			require.True(t, len(result.Errors) > 0)
			require.Contains(t, result.Errors[0], "[ValidateKebabCase]")
			require.Contains(t, result.Errors[0], tc.wantErr)
			require.Contains(t, result.Errors[0], "kebab-case")
			require.Contains(t, result.Errors[0], "ENG-HANDBOOK.md Section 4.4.1")
		})
	}
}

func TestValidateKebabCase_Violations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupFn      func(t *testing.T) string
		wantContains string
	}{
		{
			name: "path does not exist",
			setupFn: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantContains: "does not exist",
		},
		{
			name: "path is file",
			setupFn: func(t *testing.T) string {
				t.Helper()

				f := filepath.Join(t.TempDir(), "file.txt")
				require.NoError(t, os.WriteFile(f, []byte("content"), cryptoutilSharedMagic.CacheFilePermissions))

				return f
			},
			wantContains: "not a directory",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := ValidateKebabCase(tc.setupFn(t))
			require.NoError(t, err)
			require.NotNil(t, result)
			require.False(t, result.Valid)
			require.Contains(t, result.Errors[0], "[ValidateKebabCase]")
			require.Contains(t, result.Errors[0], tc.wantContains)
		})
	}
}

func TestValidateKebabCase_InvalidYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.yml"), []byte("invalid: [yaml: {"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, len(result.Warnings) > 0)
	require.Contains(t, result.Warnings[0], "Failed to parse")
}

func TestValidateKebabCase_UnreadableFile(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	dir := t.TempDir()
	filePath := filepath.Join(dir, "config.yml")
	require.NoError(t, os.WriteFile(filePath, []byte("service:\n  name: ok\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.Chmod(filePath, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(filePath, cryptoutilSharedMagic.CacheFilePermissions)
	})

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, len(result.Warnings) > 0)
	require.Contains(t, result.Warnings[0], "Failed to read")
}

func TestValidateKebabCase_WalkError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	dir := t.TempDir()
	subDir := filepath.Join(dir, "locked")
	require.NoError(t, os.MkdirAll(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "config.yml"), []byte("key: val"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.Chmod(subDir, 0o000))

	t.Cleanup(func() {
		_ = os.Chmod(subDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
	})

	result, err := ValidateKebabCase(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, len(result.Warnings) > 0)
	require.Contains(t, result.Warnings[0], "error accessing")
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
			config:    map[string]any{cryptoutilSharedMagic.ClaimName: "value"},
			fieldPath: cryptoutilSharedMagic.ClaimName,
			want:      "value",
		},
		{
			name:      "nested field",
			config:    map[string]any{"service": map[string]any{cryptoutilSharedMagic.ClaimName: "my-svc"}},
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
			config:    map[string]any{"count": cryptoutilSharedMagic.AnswerToLifeUniverseEverything},
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
			require.Equal(t, tc.want, got)
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
			contains: []string{"/test", cryptoutilSharedMagic.TestStatusPass},
		},
		{
			name: "failing result",
			result: &KebabCaseValidationResult{
				Path: "/test", Valid: false,
				Errors:   []string{"field error"},
				Warnings: []string{"warning msg"},
			},
			contains: []string{cryptoutilSharedMagic.TestStatusFail, "field error", "warning msg"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatKebabCaseValidationResult(tc.result)
			for _, s := range tc.contains {
				require.Contains(t, output, s)
			}
		})
	}
}
