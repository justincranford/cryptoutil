package lint_deployments

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestValidateTelemetry_ValidCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFn func(t *testing.T) string
	}{
		{
			name: "valid configs",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				writeConfig(t, dir, "config-1.yml", map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSvc1Value43DBC711,
					valueOtlpEndpoint: testHTTPCollector4317, valueOtlpEnvironment: valueDevelopment,
				})
				writeConfig(t, dir, "config-2.yml", map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSvc2,
					valueOtlpEndpoint: testHTTPCollector4317, valueOtlpEnvironment: valueDevelopment,
				})

				return dir
			},
		},
		{
			name: "disabled OTLP",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				writeConfig(t, dir, testConfigYml, map[string]string{valueOtlp: testFalse})

				return dir
			},
		},
		{
			name: "non-bool OTLP skipped",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(
					filepath.Join(dir, testConfigYml),
					[]byte("otlp: \"yes\"\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "no OTLP field",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(
					filepath.Join(dir, testConfigYml),
					[]byte("bind-public-port: 8080\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: "invalid YAML type skipped",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(
					filepath.Join(dir, testConfigYml),
					[]byte("- item1\n- item2\n"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
		{
			name: testSubdirectorySkipped,
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				writeConfig(t, dir, testConfigYml, map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSvc1Value43DBC711, valueOtlpEndpoint: testHTTPCollector4317,
				})

				return dir
			},
		},
		{
			name: "non-YAML skipped",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				require.NoError(t, os.WriteFile(
					filepath.Join(dir, "readme.txt"),
					[]byte("not yaml"), cryptoutilSharedMagic.CacheFilePermissions))

				return dir
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := ValidateTelemetry(tc.setupFn(t))
			require.NoError(t, err)
			require.NotNil(t, result)
			require.True(t, result.Valid)
			require.Empty(t, result.Errors)
		})
	}
}

func TestValidateTelemetry_ValidWithWarnings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFn        func(t *testing.T) string
		wantWarnSubstr string
	}{
		{
			name: "endpoint missing port",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				writeConfig(t, dir, testConfigYml, map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSvc1Value43DBC711, valueOtlpEndpoint: "http://collector",
				})

				return dir
			},
			wantWarnSubstr: "missing port",
		},
		{
			name: testUnusualScheme,
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				writeConfig(t, dir, testConfigYml, map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSvc1Value43DBC711, valueOtlpEndpoint: "ftp://collector:4317",
				})

				return dir
			},
			wantWarnSubstr: testUnusualScheme,
		},
		{
			name: "duplicate service names",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				writeConfig(t, dir, "config-1.yml", map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSameName, valueOtlpEndpoint: testHTTPCollector4317,
				})
				writeConfig(t, dir, "config-2.yml", map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSameName, valueOtlpEndpoint: testHTTPCollector4317,
				})

				return dir
			},
			wantWarnSubstr: "Duplicate otlp-service",
		},
		{
			name: "inconsistent endpoints",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				writeConfig(t, dir, "config-1.yml", map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSvc1Value43DBC711, valueOtlpEndpoint: "http://collector-a:4317",
				})
				writeConfig(t, dir, "config-2.yml", map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSvc2, valueOtlpEndpoint: "http://collector-b:4317",
				})

				return dir
			},
			wantWarnSubstr: "Inconsistent",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := ValidateTelemetry(tc.setupFn(t))
			require.NoError(t, err)
			require.NotNil(t, result)
			require.True(t, result.Valid)
			require.NotEmpty(t, result.Warnings)
			require.Contains(t, result.Warnings[0], tc.wantWarnSubstr)
		})
	}
}

func TestValidateTelemetry_Violations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupFn      func(t *testing.T) string
		wantContains string
	}{
		{
			name: testPathNotFound,
			setupFn: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantContains: testNotFound,
		},
		{
			name: testPathIsFile,
			setupFn: func(t *testing.T) string {
				t.Helper()

				f := filepath.Join(t.TempDir(), "file.yml")
				require.NoError(t, os.WriteFile(f, []byte("test"), cryptoutilSharedMagic.CacheFilePermissions))

				return f
			},
			wantContains: testNotADirectory,
		},
		{
			name: "empty endpoint",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				writeConfig(t, dir, testConfigYml, map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSvc1Value43DBC711,
				})

				return dir
			},
			wantContains: cryptoutilSharedMagic.EMPTY,
		},
		{
			name: "invalid endpoint URL",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				writeConfig(t, dir, testConfigYml, map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSvc1Value43DBC711, valueOtlpEndpoint: "://bad-url",
				})

				return dir
			},
			wantContains: cryptoutilSharedMagic.DevSetupTestInvalidName,
		},
		{
			name: "endpoint missing host",
			setupFn: func(t *testing.T) string {
				t.Helper()

				dir := t.TempDir()
				writeConfig(t, dir, testConfigYml, map[string]string{
					valueOtlp: testTrue, valueOtlpService: testSvc1Value43DBC711, valueOtlpEndpoint: "http://:4317",
				})

				return dir
			},
			wantContains: "missing host",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := ValidateTelemetry(tc.setupFn(t))
			require.NoError(t, err)
			require.NotNil(t, result)
			require.False(t, result.Valid)
			require.Contains(t, result.Errors[0], tc.wantContains)
		})
	}
}

func TestValidateTelemetry_AcceptedSchemes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
	}{
		{name: cryptoutilSharedMagic.ProtocolHTTP, endpoint: testHTTPCollector4317},
		{name: cryptoutilSharedMagic.ProtocolHTTPS, endpoint: "https://collector:4318"},
		{name: "grpc", endpoint: "grpc://collector:4317"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			writeConfig(t, dir, testConfigYml, map[string]string{
				valueOtlp: testTrue, valueOtlpService: testSvc1Value43DBC711, valueOtlpEndpoint: tc.endpoint,
			})

			result, err := ValidateTelemetry(dir)
			require.NoError(t, err)
			require.True(t, result.Valid)

			for _, w := range result.Warnings {
				require.NotContains(t, w, testUnusualScheme)
			}
		})
	}
}

func TestValidateTelemetry_UnreadableFile(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	dir := t.TempDir()
	require.NoError(t, os.Symlink("/nonexistent/broken.yml",
		filepath.Join(dir, "broken.yml")))

	result, err := ValidateTelemetry(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
}

func TestValidateTelemetry_UnreadableDir(t *testing.T) {
	t.Parallel()

	result := &TelemetryValidationResult{Valid: true}
	entries := collectOTLPEntries("/nonexistent/dir", result)
	require.Nil(t, entries)
	require.True(t, result.Valid)
}

func TestValidateTelemetry_EndpointNormalization(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeConfig(t, dir, "config-1.yml", map[string]string{
		valueOtlp: testTrue, valueOtlpService: testSvc1Value43DBC711, valueOtlpEndpoint: "http://collector:4317/",
	})
	writeConfig(t, dir, "config-2.yml", map[string]string{
		valueOtlp: testTrue, valueOtlpService: testSvc2, valueOtlpEndpoint: testHTTPCollector4317,
	})

	result, err := ValidateTelemetry(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)

	for _, w := range result.Warnings {
		require.NotContains(t, w, "Inconsistent")
	}
}

func TestNormalizeEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "no trailing slash", input: testHTTPHost4317, expected: testHTTPHost4317},
		{name: "single trailing slash", input: "http://host:4317/", expected: testHTTPHost4317},
		{name: "multiple trailing slashes", input: "http://host:4317///", expected: testHTTPHost4317},
		{name: cryptoutilSharedMagic.EMPTY, input: "", expected: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.expected, normalizeEndpoint(tc.input))
		})
	}
}

func TestFormatTelemetryValidationResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *TelemetryValidationResult
		contains []string
	}{
		{
			name:     testNilResult,
			result:   nil,
			contains: []string{"No telemetry validation result"},
		},
		{
			name:     testValid,
			result:   &TelemetryValidationResult{Valid: true},
			contains: []string{testPassed},
		},
		{
			name: testErrorsAndWarnings,
			result: &TelemetryValidationResult{
				Valid:    false,
				Errors:   []string{testErr1},
				Warnings: []string{testWarn1},
			},
			contains: []string{cryptoutilSharedMagic.TaskFailed, testErrorErr1, testWarningWarn1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatTelemetryValidationResult(tc.result)
			for _, want := range tc.contains {
				require.Contains(t, output, want)
			}
		})
	}
}

func TestValidateTelemetry_RealSmIM(t *testing.T) {
	t.Parallel()

	configDir := filepath.Join("testdata", cryptoutilSharedMagic.CICDConfigsDir, valueSm, "im")
	if _, err := os.Stat(configDir); err != nil {
		configDir = filepath.Join("..", "..", "..", "..", cryptoutilSharedMagic.CICDConfigsDir, valueSm, "im")
	}

	if _, err := os.Stat(configDir); err != nil {
		t.Skip("Real sm-kms configs not found")
	}

	result, err := ValidateTelemetry(configDir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid, "Real sm-kms OTLP config validation failed: %v", result.Errors)
}

// writeConfig creates a YAML config file from key-value pairs.
func writeConfig(t *testing.T, dir, name string, fields map[string]string) {
	t.Helper()

	var content string

	for k, v := range fields {
		if v == testTrue || v == testFalse {
			content += k + ": " + v + "\n"
		} else {
			content += k + ": \"" + v + "\"\n"
		}
	}

	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), cryptoutilSharedMagic.CacheFilePermissions))
}
