package lint_deployments

import (
	"os"
	"path/filepath"
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
				require.NoError(t, os.WriteFile(filepath.Join(dir, testConfigYml),
					[]byte("service:\n  name: \"sm-kms\"\n  version: \"1.0.0\"\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "missing service name field",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, testConfigYml),
					[]byte("observability:\n  enabled: true\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: testNonYAMLFilesIgnored,
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(dir, valueReadmeMd), []byte("# readme"), cryptoutilSharedMagic.CacheFilePermissions))
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
				require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), content, cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.DOCKER_COMPOSE_YML), content, cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: testNestedDirectories,
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				nested := filepath.Join(dir, cryptoutilSharedMagic.ClaimSub, testDeep)
				require.NoError(t, os.MkdirAll(nested, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.WriteFile(filepath.Join(nested, testConfigYml),
					[]byte("service:\n  name: \"my-nested-service\"\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: testEmptyDirectory,
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
				require.NoError(t, os.WriteFile(filepath.Join(dir, testConfigYml),
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
		{name: "PascalCase", serviceName: testSmIm, wantErr: testSmIm},
		{name: "snake_case", serviceName: testSmKMSServiceValue38605A41, wantErr: testSmKMSServiceValue38605A41},
		{name: "UPPERCASE", serviceName: testSmKMSService, wantErr: testSmKMSService},
		{name: "spaces", serviceName: testSmImValueA70FF70D, wantErr: testSmImValueA70FF70D},
		{name: testLeadingHyphen, serviceName: testSmKMS, wantErr: testSmKMS},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			content := "service:\n  name: \"" + tc.serviceName + "\"\n"
			require.NoError(t, os.WriteFile(filepath.Join(dir, testConfigYml), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

			result, err := ValidateKebabCase(dir)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.False(t, result.Valid)
			require.True(t, len(result.Errors) > 0)
			require.Contains(t, result.Errors[0], "[ValidateKebabCase]")
			require.Contains(t, result.Errors[0], tc.wantErr)
			require.Contains(t, result.Errors[0], validatorNameKebabCase)
			require.Contains(t, result.Errors[0], testEngHandbookMdSection441)
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
			name: testPathDoesNotExist,
			setupFn: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantContains: testDoesNotExist,
		},
		{
			name: testPathIsFile,
			setupFn: func(t *testing.T) string {
				t.Helper()

				f := filepath.Join(t.TempDir(), "file.txt")
				require.NoError(t, os.WriteFile(f, []byte("content"), cryptoutilSharedMagic.CacheFilePermissions))

				return f
			},
			wantContains: testNotADirectory,
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

	dir := t.TempDir()
	filePath := filepath.Join(dir, testConfigYml)
	require.NoError(t, os.Mkdir(filePath, cryptoutilSharedMagic.CacheFilePermissions))

	result := &KebabCaseValidationResult{}
	validateConfigFieldKebabCase(filePath, testConfigYml, result)
	require.Len(t, result.Warnings, 1)
	require.Contains(t, result.Warnings[0], "Failed to read")
}

func TestValidateKebabCase_WalkError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	result := &KebabCaseValidationResult{Path: dir, Valid: true}
	validateKebabCaseWithWalk(dir, result, func(root string, walkFn filepath.WalkFunc) error {
		return walkFn(filepath.Join(root, "locked"), nil, os.ErrPermission)
	})

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
			config:    map[string]any{cryptoutilSharedMagic.ClaimName: testValue},
			fieldPath: cryptoutilSharedMagic.ClaimName,
			want:      testValue,
		},
		{
			name:      "nested field",
			config:    map[string]any{testService: map[string]any{cryptoutilSharedMagic.ClaimName: testMySvc}},
			fieldPath: valueServiceName,
			want:      testMySvc,
		},
		{
			name:      "deeply nested",
			config:    map[string]any{testA: map[string]any{testB: map[string]any{testC: testDeep}}},
			fieldPath: "a.b.c",
			want:      testDeep,
		},
		{
			name:      testMissingField,
			config:    map[string]any{"other": testValue},
			fieldPath: valueServiceName,
			want:      "",
		},
		{
			name:      "missing intermediate",
			config:    map[string]any{testService: "not-a-map"},
			fieldPath: valueServiceName,
			want:      "",
		},
		{
			name:      "non-string value",
			config:    map[string]any{testCount: cryptoutilSharedMagic.AnswerToLifeUniverseEverything},
			fieldPath: testCount,
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
			name:     testPassingResult,
			result:   &KebabCaseValidationResult{Path: testTest, Valid: true},
			contains: []string{testTest, cryptoutilSharedMagic.TestStatusPass},
		},
		{
			name: "failing result",
			result: &KebabCaseValidationResult{
				Path: testTest, Valid: false,
				Errors:   []string{testFieldError},
				Warnings: []string{testWarningMsg},
			},
			contains: []string{cryptoutilSharedMagic.TestStatusFail, testFieldError, testWarningMsg},
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
