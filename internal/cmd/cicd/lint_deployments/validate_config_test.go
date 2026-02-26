package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    string
		wantValid  bool
		wantErrors []string
		wantWarns  []string
	}{
		{
			name: "valid config with all fields",
			content: `bind-public-protocol: https
bind-public-address: 0.0.0.0
bind-public-port: 8080
bind-private-protocol: https
bind-private-address: 127.0.0.1
bind-private-port: 9090
database-url: "file:///run/secrets/db_url"
otlp: true
otlp-service: my-service
otlp-endpoint: otel-collector:4317
`,
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "valid config with sqlite database",
			content: `bind-public-protocol: https
bind-public-address: 127.0.0.1
bind-public-port: 8080
bind-private-protocol: https
bind-private-address: 127.0.0.1
bind-private-port: 9090
database-url: "sqlite:///data/app.db"
`,
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "valid config with memory database",
			content: `bind-public-protocol: https
bind-public-address: 127.0.0.1
bind-public-port: 8080
bind-private-protocol: https
bind-private-address: 127.0.0.1
bind-private-port: 9090
database-url: ":memory:"
`,
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name:       "invalid YAML",
			content:    "key: [unclosed bracket",
			wantValid:  false,
			wantErrors: []string{"YAML parse error"},
		},
		{
			name:      "empty config",
			content:   "---\n",
			wantValid: true,
			wantWarns: []string{"config file is empty"},
		},
		{
			name: "invalid bind address",
			content: `bind-public-address: not-an-ip
bind-private-address: 127.0.0.l
`,
			wantValid: false,
			wantErrors: []string{
				"'bind-public-address' is not a valid IP address",
				"'bind-private-address' is not a valid IP address",
			},
		},
		{
			name: "bind address wrong type",
			content: `bind-public-address: 12345
`,
			wantValid:  false,
			wantErrors: []string{"'bind-public-address' must be a string"},
		},
		{
			name: "port out of range zero",
			content: `bind-public-port: 0
`,
			wantValid:  false,
			wantErrors: []string{"'bind-public-port' must be between"},
		},
		{
			name: "port out of range high",
			content: `bind-public-port: 70000
`,
			wantValid:  false,
			wantErrors: []string{"'bind-public-port' must be between"},
		},
		{
			name: "port wrong type",
			content: `bind-public-port: "not-a-number"
`,
			wantValid:  false,
			wantErrors: []string{"'bind-public-port' must be an integer"},
		},
		{
			name: "protocol not https",
			content: `bind-public-protocol: http
bind-private-protocol: grpc
`,
			wantValid: false,
			wantErrors: []string{
				"'bind-public-protocol' must be \"https\"",
				"'bind-private-protocol' must be \"https\"",
			},
		},
		{
			name: "protocol wrong type",
			content: `bind-public-protocol: 443
`,
			wantValid:  false,
			wantErrors: []string{"'bind-public-protocol' must be a string"},
		},
		{
			name: "admin bind policy violation",
			content: `bind-private-address: 0.0.0.0
`,
			wantValid:  false,
			wantErrors: []string{"POLICY VIOLATION"},
		},
		{
			name: "inline postgres credentials",
			content: `database-url: "postgres://user:pass@localhost:5432/mydb"
`,
			wantValid:  false,
			wantErrors: []string{"inline database credentials"},
		},
		{
			name: "inline postgresql credentials",
			content: `database-url: "postgresql://admin:secret@db:5432/prod"
`,
			wantValid:  false,
			wantErrors: []string{"inline database credentials"},
		},
		{
			name: "unexpected database URL format",
			content: `database-url: "mysql://user:pass@localhost/mydb"
`,
			wantValid: true,
			wantWarns: []string{"unexpected format"},
		},
		{
			name: "otlp enabled missing required fields",
			content: `otlp: true
`,
			wantValid: false,
			wantErrors: []string{
				"'otlp-service' is required when 'otlp' is true",
				"'otlp-endpoint' is required when 'otlp' is true",
			},
		},
		{
			name: "otlp disabled no errors",
			content: `otlp: false
`,
			wantValid: true,
		},
		{
			name: "otlp wrong type",
			content: `otlp: "yes"
`,
			wantValid: true,
			wantWarns: []string{"'otlp' should be a boolean"},
		},
		{
			name: "otlp enabled with all required",
			content: `otlp: true
otlp-service: test-svc
otlp-endpoint: collector:4317
`,
			wantValid: true,
		},
		{
			name: "minimal valid config",
			content: `log-level: INFO
`,
			wantValid: true,
		},
		{
			name: "valid port at boundaries",
			content: `bind-public-port: 1
bind-private-port: 65535
`,
			wantValid: true,
		},
		{
			name: "empty database URL skipped",
			content: `database-url: ""
`,
			wantValid: true,
		},
		{
			name: "multiple errors combined",
			content: `bind-public-address: invalid
bind-public-port: 0
bind-public-protocol: http
bind-private-address: 0.0.0.0
database-url: "postgres://user:pass@host/db"
otlp: true
`,
			wantValid: false,
			wantErrors: []string{
				"not a valid IP address",
				"must be between",
				"must be \"https\"",
				"POLICY VIOLATION",
				"inline database credentials",
				"'otlp-service' is required",
				"'otlp-endpoint' is required",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yml")
			require.NoError(t, os.WriteFile(configFile, []byte(tc.content), filePermissions))

			result, err := ValidateConfigFile(configFile)
			require.NoError(t, err)
			assert.Equal(t, tc.wantValid, result.Valid, "Valid mismatch")

			for _, wantErr := range tc.wantErrors {
				assert.True(t, containsSubstring(result.Errors, wantErr),
					"expected error containing %q in %v", wantErr, result.Errors)
			}

			for _, wantWarn := range tc.wantWarns {
				assert.True(t, containsSubstring(result.Warnings, wantWarn),
					"expected warning containing %q in %v", wantWarn, result.Warnings)
			}
		})
	}
}

func TestValidateConfigFile_FileNotFound(t *testing.T) {
	t.Parallel()

	result, err := ValidateConfigFile("/nonexistent/path/config.yml")
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.True(t, containsSubstring(result.Errors, "cannot read file"))
}

func TestToInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		wantVal int
		wantOK  bool
	}{
		{name: "int", input: cryptoutilSharedMagic.AnswerToLifeUniverseEverything, wantVal: cryptoutilSharedMagic.AnswerToLifeUniverseEverything, wantOK: true},
		{name: "int64", input: int64(99), wantVal: 99, wantOK: true},
		{name: "float64", input: float64(cryptoutilSharedMagic.DemoServerPort), wantVal: cryptoutilSharedMagic.DemoServerPort, wantOK: true},
		{name: "string", input: "not-a-number", wantVal: 0, wantOK: false},
		{name: "bool", input: true, wantVal: 0, wantOK: false},
		{name: "nil", input: nil, wantVal: 0, wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			val, ok := toInt(tc.input)
			assert.Equal(t, tc.wantVal, val)
			assert.Equal(t, tc.wantOK, ok)
		})
	}
}

func TestFormatConfigValidationResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		result     *ConfigValidationResult
		wantPASS   bool
		wantSubstr []string
	}{
		{
			name: "passing result",
			result: &ConfigValidationResult{
				Path:  "config.yml",
				Valid: true,
			},
			wantPASS:   true,
			wantSubstr: []string{"[PASS]", "config.yml"},
		},
		{
			name: "failing result with errors and warnings",
			result: &ConfigValidationResult{
				Path:     "bad.yml",
				Valid:    false,
				Errors:   []string{"port out of range", "invalid address"},
				Warnings: []string{"deprecated field"},
			},
			wantPASS:   false,
			wantSubstr: []string{"[FAIL]", "bad.yml", "ERROR: port out of range", "ERROR: invalid address", "WARNING: deprecated field"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatConfigValidationResult(tc.result)

			if tc.wantPASS {
				assert.Contains(t, output, "[PASS]")
			} else {
				assert.Contains(t, output, "[FAIL]")
			}

			for _, s := range tc.wantSubstr {
				assert.Contains(t, output, s)
			}
		})
	}
}

func TestMainValidateConfig_NoArgs(t *testing.T) {
	t.Parallel()

	exitCode := mainValidateConfig(nil)
	assert.Equal(t, 1, exitCode)
}

func TestMainValidateConfig_NonexistentFile(t *testing.T) {
	t.Parallel()

	exitCode := mainValidateConfig([]string{"/nonexistent/config.yml"})
	assert.Equal(t, 1, exitCode)
}

func TestMainValidateConfig_ValidFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "valid.yml")
	content := `bind-public-protocol: https
bind-private-protocol: https
bind-private-address: 127.0.0.1
`
	require.NoError(t, os.WriteFile(configFile, []byte(content), filePermissions))

	exitCode := mainValidateConfig([]string{configFile})
	assert.Equal(t, 0, exitCode)
}

func TestMainValidateConfig_InvalidFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.yml")
	content := `bind-public-protocol: http
`
	require.NoError(t, os.WriteFile(configFile, []byte(content), filePermissions))

	exitCode := mainValidateConfig([]string{configFile})
	assert.Equal(t, 1, exitCode)
}

func TestValidateAdminBindPolicy_NonStringType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yml")
	content := `bind-private-address: 12345
`
	require.NoError(t, os.WriteFile(configFile, []byte(content), filePermissions))

	result, err := ValidateConfigFile(configFile)
	require.NoError(t, err)
	// Non-string type error is caught by validateBindAddresses, not duplicated by admin policy.
	assert.False(t, result.Valid)
	assert.True(t, containsSubstring(result.Errors, "must be a string"))
}

func TestValidateConfigSecretRefs_NonStringType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yml")
	content := `database-url: 12345
`
	require.NoError(t, os.WriteFile(configFile, []byte(content), filePermissions))

	result, err := ValidateConfigFile(configFile)
	require.NoError(t, err)
	// Non-string database-url is silently skipped.
	assert.True(t, result.Valid)
}

func TestValidateConfigSecretRefs_NoDatabaseURL(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yml")
	content := `log-level: INFO
`
	require.NoError(t, os.WriteFile(configFile, []byte(content), filePermissions))

	result, err := ValidateConfigFile(configFile)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}
