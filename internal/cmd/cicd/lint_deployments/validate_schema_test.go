package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// validConfigContent returns a minimal valid flat kebab-case config.
func validConfigContent() string {
	return `bind-public-protocol: "https"
bind-public-address: "127.0.0.1"
bind-public-port: 8080
bind-private-protocol: "https"
bind-private-address: "127.0.0.1"
bind-private-port: 9090
tls-public-mode: "auto"
tls-private-mode: "auto"
otlp: true
otlp-service: "my-service"
otlp-environment: "development"
otlp-endpoint: "http://otel:4317"
`
}

func TestValidateSchema_ValidConfig(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(validConfigContent()), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSchema(path)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
}

func TestValidateSchema_MissingRequiredFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "missing bind-public-protocol",
			content: "bind-public-address: \"0.0.0.0\"\nbind-public-port: 8080\nbind-private-protocol: \"https\"\nbind-private-address: \"127.0.0.1\"\nbind-private-port: 9090\ntls-public-mode: \"auto\"\ntls-private-mode: \"auto\"\notlp: true\n",
			wantErr: "bind-public-protocol",
		},
		{
			name:    "missing bind-private-address",
			content: "bind-public-protocol: \"https\"\nbind-public-address: \"0.0.0.0\"\nbind-public-port: 8080\nbind-private-protocol: \"https\"\nbind-private-port: 9090\ntls-public-mode: \"auto\"\ntls-private-mode: \"auto\"\notlp: true\n",
			wantErr: "bind-private-address",
		},
		{
			name:    "missing otlp",
			content: "bind-public-protocol: \"https\"\nbind-public-address: \"0.0.0.0\"\nbind-public-port: 8080\nbind-private-protocol: \"https\"\nbind-private-address: \"127.0.0.1\"\nbind-private-port: 9090\ntls-public-mode: \"auto\"\ntls-private-mode: \"auto\"\n",
			wantErr: "otlp",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "config.yml")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))

			result, err := ValidateSchema(path)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.False(t, result.Valid)
			require.True(t, containsSubstring(result.Errors, tc.wantErr))
		})
	}
}

func TestValidateSchema_InvalidTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "port as string",
			content: validConfigContent() + "# override\n",
			wantErr: "", // valid - no type error expected
		},
		{
			name:    "protocol as int",
			content: "bind-public-protocol: 123\nbind-public-address: \"0.0.0.0\"\nbind-public-port: 8080\nbind-private-protocol: \"https\"\nbind-private-address: \"127.0.0.1\"\nbind-private-port: 9090\ntls-public-mode: \"auto\"\ntls-private-mode: \"auto\"\notlp: true\n",
			wantErr: "must be a string",
		},
		{
			name:    "otlp as string",
			content: "bind-public-protocol: \"https\"\nbind-public-address: \"0.0.0.0\"\nbind-public-port: 8080\nbind-private-protocol: \"https\"\nbind-private-address: \"127.0.0.1\"\nbind-private-port: 9090\ntls-public-mode: \"auto\"\ntls-private-mode: \"auto\"\notlp: \"yes\"\n",
			wantErr: "must be a boolean",
		},
		{
			name:    "port as bool",
			content: "bind-public-protocol: \"https\"\nbind-public-address: \"0.0.0.0\"\nbind-public-port: true\nbind-private-protocol: \"https\"\nbind-private-address: \"127.0.0.1\"\nbind-private-port: 9090\ntls-public-mode: \"auto\"\ntls-private-mode: \"auto\"\notlp: true\n",
			wantErr: "must be an integer",
		},
		{
			name:    "cors-allowed-origins as string",
			content: validConfigContent() + "cors-allowed-origins: \"single-value\"\n",
			wantErr: "must be a string array",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "config.yml")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))

			result, err := ValidateSchema(path)
			require.NoError(t, err)
			require.NotNil(t, result)

			if tc.wantErr == "" {
				require.True(t, result.Valid)
			} else {
				require.False(t, result.Valid)
				require.True(t, containsSubstring(result.Errors, tc.wantErr))
			}
		})
	}
}

func TestValidateSchema_InvalidEnumValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "invalid protocol",
			content: "bind-public-protocol: \"http\"\nbind-public-address: \"0.0.0.0\"\nbind-public-port: 8080\nbind-private-protocol: \"https\"\nbind-private-address: \"127.0.0.1\"\nbind-private-port: 9090\ntls-public-mode: \"auto\"\ntls-private-mode: \"auto\"\notlp: true\n",
			wantErr: "not in allowed values",
		},
		{
			name:    "invalid private address",
			content: "bind-public-protocol: \"https\"\nbind-public-address: \"0.0.0.0\"\nbind-public-port: 8080\nbind-private-protocol: \"https\"\nbind-private-address: \"0.0.0.0\"\nbind-private-port: 9090\ntls-public-mode: \"auto\"\ntls-private-mode: \"auto\"\notlp: true\n",
			wantErr: "not in allowed values",
		},
		{
			name:    "invalid tls mode",
			content: "bind-public-protocol: \"https\"\nbind-public-address: \"0.0.0.0\"\nbind-public-port: 8080\nbind-private-protocol: \"https\"\nbind-private-address: \"127.0.0.1\"\nbind-private-port: 9090\ntls-public-mode: \"invalid\"\ntls-private-mode: \"auto\"\notlp: true\n",
			wantErr: "not in allowed values",
		},
		{
			name:    "invalid otlp-environment",
			content: "bind-public-protocol: \"https\"\nbind-public-address: \"0.0.0.0\"\nbind-public-port: 8080\nbind-private-protocol: \"https\"\nbind-private-address: \"127.0.0.1\"\nbind-private-port: 9090\ntls-public-mode: \"auto\"\ntls-private-mode: \"auto\"\notlp: true\notlp-environment: \"staging\"\n",
			wantErr: "not in allowed values",
		},
		{
			name:    "invalid session algorithm",
			content: "bind-public-protocol: \"https\"\nbind-public-address: \"0.0.0.0\"\nbind-public-port: 8080\nbind-private-protocol: \"https\"\nbind-private-address: \"127.0.0.1\"\nbind-private-port: 9090\ntls-public-mode: \"auto\"\ntls-private-mode: \"auto\"\notlp: true\nbrowser-session-algorithm: \"HMAC\"\n",
			wantErr: "not in allowed values",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "config.yml")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))

			result, err := ValidateSchema(path)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.False(t, result.Valid)
			require.True(t, containsSubstring(result.Errors, tc.wantErr))
		})
	}
}

func TestValidateSchema_UnknownFields(t *testing.T) {
	t.Parallel()

	content := validConfigContent() + "unknown-field: \"value\"\nanother-unknown: 42\n"
	path := filepath.Join(t.TempDir(), "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSchema(path)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Unknown fields produce warnings, not errors.
	require.True(t, result.Valid)
	require.True(t, len(result.Warnings) >= 2)
	require.True(t, containsSubstring(result.Warnings, "unknown-field"))
	require.True(t, containsSubstring(result.Warnings, "another-unknown"))
}

func TestValidateSchema_FileNotFound(t *testing.T) {
	t.Parallel()

	result, err := ValidateSchema(filepath.Join(t.TempDir(), "nonexistent.yml"))
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.True(t, containsSubstring(result.Errors, "cannot read file"))
}

func TestValidateSchema_InvalidYAML(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "broken.yml")
	require.NoError(t, os.WriteFile(path, []byte("invalid: [yaml: {broken"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSchema(path)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.True(t, containsSubstring(result.Errors, "YAML parse error"))
}

func TestValidateSchema_EmptyFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "empty.yml")
	require.NoError(t, os.WriteFile(path, []byte(""), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSchema(path)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid, "empty file should produce warning, not error")
	require.True(t, containsSubstring(result.Warnings, "empty"))
}

func TestValidateSchema_CORSArrayWithNonString(t *testing.T) {
	t.Parallel()

	content := validConfigContent() + "cors-allowed-origins:\n  - \"http://ok\"\n  - 123\n"
	path := filepath.Join(t.TempDir(), "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSchema(path)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Valid)
	require.True(t, containsSubstring(result.Errors, "must be a string"))
}

func TestValidateSchema_OptionalFieldsValid(t *testing.T) {
	t.Parallel()

	content := validConfigContent() +
		"cors-max-age: 3600\n" +
		"cors-allowed-origins:\n  - \"http://localhost:8080\"\n" +
		"browser-session-algorithm: \"JWS\"\n" +
		"browser-session-jws-algorithm: \"HS256\"\n" +
		"service-session-algorithm: \"Opaque\"\n" +
		"database-url: \"file:///run/secrets/db.secret\"\n"

	path := filepath.Join(t.TempDir(), "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(content), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSchema(path)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
}

func TestIsIntLike(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  any
		want bool
	}{
		{name: "int", val: cryptoutilSharedMagic.AnswerToLifeUniverseEverything, want: true},
		{name: "int64", val: int64(cryptoutilSharedMagic.AnswerToLifeUniverseEverything), want: true},
		{name: "float64", val: float64(42.0), want: true},
		{name: "string", val: "42", want: false},
		{name: "bool", val: true, want: false},
		{name: "nil", val: nil, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, isIntLike(tc.val))
		})
	}
}

func TestFormatSchemaValidationResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   *SchemaValidationResult
		contains []string
	}{
		{
			name:     "passing",
			result:   &SchemaValidationResult{Path: "/test", Valid: true},
			contains: []string{cryptoutilSharedMagic.TestStatusPass, "/test"},
		},
		{
			name: "failing with errors and warnings",
			result: &SchemaValidationResult{
				Path: "/test", Valid: false,
				Errors:   []string{"missing field"},
				Warnings: []string{"unknown field"},
			},
			contains: []string{cryptoutilSharedMagic.TestStatusFail, "missing field", "unknown field"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output := FormatSchemaValidationResult(tc.result)
			for _, s := range tc.contains {
				require.Contains(t, output, s)
			}
		})
	}
}
