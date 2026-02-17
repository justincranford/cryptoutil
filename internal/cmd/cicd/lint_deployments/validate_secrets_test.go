package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSecrets_ValidDeployment(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "my_password.secret"),
		[]byte("this-is-a-very-long-secret-value-with-enough-entropy-for-security"), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateSecrets_PathNotFound(t *testing.T) {
	t.Parallel()

	result, err := ValidateSecrets("/nonexistent/path/xyz")
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0], "path not found")
}

func TestValidateSecrets_PathIsFile(t *testing.T) {
	t.Parallel()
	f := filepath.Join(t.TempDir(), "file.txt")
	require.NoError(t, os.WriteFile(f, []byte("data"), 0o600))

	result, err := ValidateSecrets(f)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "not a directory")
}

func TestValidateSecrets_EmptySecretFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "my_password.secret"), []byte(""), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0], "is empty")
}

func TestValidateSecrets_ShortSecretFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "hash_pepper.secret"), []byte("short"), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "has 5 bytes")
	assert.Contains(t, result.Warnings[0], "minimum recommended: 32")
}

func TestValidateSecrets_Base64LengthOK(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, 0o755))

	base64Value := "4b0beuNCVMFMjA4y2/WvULfaCZiE6TLnPctdiSdtVrI="
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "hash_pepper.secret"), []byte(base64Value), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)
}

func TestValidateSecrets_NonSecretFileIgnored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "readme.md"), []byte("short"), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateSecrets_NonHighEntropySecretIgnored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "postgres_database.secret"), []byte("mydb"), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)
}

func TestValidateSecrets_SecretDirSubdirectorySkipped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(secretsDir, "subdir"), 0o755))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_UnreadableSecretFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, 0o755))
	f := filepath.Join(secretsDir, "api_key.secret")
	require.NoError(t, os.WriteFile(f, []byte("data"), 0o600))
	require.NoError(t, os.Chmod(f, 0o000))

	t.Cleanup(func() { _ = os.Chmod(f, 0o600) })

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "cannot read secret file")
}

func TestValidateSecrets_UnreadableSecretsDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, 0o755))
	require.NoError(t, os.Chmod(secretsDir, 0o000))

	t.Cleanup(func() { _ = os.Chmod(secretsDir, 0o755) })

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "cannot read secrets directory")
}

func TestValidateSecrets_ConfigInlineSecret(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))

	config := "database-password: supersecretpassword123\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0], "inline secret")
}

func TestValidateSecrets_ConfigSafeReference(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))

	config := "database-password: file:///run/secrets/db_password\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateSecrets_ConfigNestedInlineSecret(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))

	config := "auth:\n  api-key: my-hardcoded-api-key-value\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "auth.api-key")
}

func TestValidateSecrets_ConfigEmptySecretValue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))

	config := "database-password: \"\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ConfigNonSecretField(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))

	config := "bind-public-port: 8080\nhost: localhost\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ConfigInvalidYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "bad.yml"), []byte(":\n  - :\n:"), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ConfigUnreadableFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))
	f := filepath.Join(configsDir, "unreadable.yml")
	require.NoError(t, os.WriteFile(f, []byte("data"), 0o600))
	require.NoError(t, os.Chmod(f, 0o000))

	t.Cleanup(func() { _ = os.Chmod(f, 0o600) })

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "cannot read config file")
}

func TestValidateSecrets_ConfigNonYAMLIgnored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "readme.md"), []byte("api-key: secret"), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ConfigSubdirIgnored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(configsDir, "subdir"), 0o755))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_UnreadableConfigsDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))
	require.NoError(t, os.Chmod(configsDir, 0o000))

	t.Cleanup(func() { _ = os.Chmod(configsDir, 0o755) })

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "cannot read configs directory")
}

func TestValidateSecrets_ComposeInlineSecret(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	compose := `services:
  myapp:
    image: myapp:latest
    environment:
      DB_PASSWORD: "hardcoded-password-value"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0], "inline secret")
}

func TestValidateSecrets_ComposeSecretFileRef(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	compose := `services:
  myapp:
    image: myapp:latest
    environment:
      DB_PASSWORD_FILE: "/run/secrets/db_password"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ComposeNonSecretEnv(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	compose := `services:
  myapp:
    image: myapp:latest
    environment:
      APP_NAME: "my-service"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ComposeEmptySecretValue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	compose := `services:
  myapp:
    image: myapp:latest
    environment:
      DB_PASSWORD: ""
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ComposeInvalidYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(":\n  - :\n:"), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ComposeUnreadable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "compose.yml")
	require.NoError(t, os.WriteFile(f, []byte("data"), 0o600))
	require.NoError(t, os.Chmod(f, 0o000))

	t.Cleanup(func() { _ = os.Chmod(f, 0o600) })

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "cannot read compose file")
}

func TestValidateSecrets_ComposeNonMapEnv(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	compose := `services:
  myapp:
    image: myapp:latest
    environment:
      - DB_PASSWORD=hardcoded
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_NoSecretsDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_NoConfigsDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_NoComposeFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_RealCipherIM(t *testing.T) {
	t.Parallel()

	deploymentPath := findRealDeploymentPath("cipher-im")
	if deploymentPath == "" {
		t.Skip("cipher-im deployment not found")
	}

	result, err := ValidateSecrets(deploymentPath)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestValidateSecrets_DotNeverSecretFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "unseal_1of5.secret.never"),
		[]byte("short"), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "has 5 bytes")
	assert.Contains(t, result.Warnings[0], "minimum recommended: 32")
}

func TestValidateSecrets_ComposeDockerComposeYml(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	compose := `services:
  myapp:
    image: myapp:latest
    environment:
      MY_SECRET: "inline-value"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(compose), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "inline secret")
}

func TestValidateSecrets_ConfigNonStringValue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))

	config := "api-key: 12345\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ConfigSqliteURL(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))

	config := "database-password: sqlite:///tmp/test.db\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ConfigMemoryRef(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, 0o755))

	config := "db-password: \":memory:\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestFormatSecretValidationResult_Nil(t *testing.T) {
	t.Parallel()

	output := FormatSecretValidationResult(nil)
	assert.Contains(t, output, "SKIP")
}

func TestFormatSecretValidationResult_Valid(t *testing.T) {
	t.Parallel()

	result := &SecretValidationResult{Valid: true}
	output := FormatSecretValidationResult(result)
	assert.Contains(t, output, "PASS")
}

func TestFormatSecretValidationResult_Errors(t *testing.T) {
	t.Parallel()

	result := &SecretValidationResult{
		Valid:    false,
		Errors:   []string{"inline secret found"},
		Warnings: []string{"short secret"},
	}
	output := FormatSecretValidationResult(result)
	assert.Contains(t, output, "FAIL")
	assert.Contains(t, output, "ERROR: inline secret found")
	assert.Contains(t, output, "WARN: short secret")
}

func TestIsSecretFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{"secret file", "password.secret", true},
		{"secret never file", "unseal.secret.never", true},
		{"not secret", "readme.md", false},
		{"yml file", "config.yml", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expect, isSecretFile(tc.input))
		})
	}
}

func TestIsHighEntropySecretFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{"password", "db_password.secret", true},
		{"pepper", "hash_pepper_v3.secret", true},
		{"unseal", "unseal_1of5.secret", true},
		{"api_key", "my_api_key.secret", true},
		{"secret_key", "secret_key.secret", true},
		{"database name", "postgres_database.secret", false},
		{"username", "postgres_username.secret", false},
		{"url", "postgres_url.secret", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expect, isHighEntropySecretFile(tc.input))
		})
	}
}

func TestIsSecretFieldName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{"password", "database-password", true},
		{"PASSWD", "POSTGRES_PASSWD", true},
		{"secret", "my-secret", true},
		{"TOKEN", "ACCESS_TOKEN", true},
		{"api-key", "api-key", true},
		{"api_key", "API_KEY", true},
		{"private-key", "private-key", true},
		{"pepper", "hash-pepper", true},
		{"hostname", "hostname", false},
		{"port", "bind-port", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expect, isSecretFieldName(tc.input))
		})
	}
}

func TestIsSafeReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{"docker secret", "file:///run/secrets/db_pass", true},
		{"file ref", "file:///path/to/file", true},
		{"sqlite", "sqlite:///tmp/test.db", true},
		{"memory", ":memory:", true},
		{"inline", "my-hardcoded-password", false},
		{"postgres URL", "postgres://user:pass@host/db", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expect, isSafeReference(tc.input))
		})
	}
}

func TestFindComposeFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		found    bool
	}{
		{"compose.yml", "compose.yml", true},
		{"compose.yaml", "compose.yaml", true},
		{"docker-compose.yml", "docker-compose.yml", true},
		{"docker-compose.yaml", "docker-compose.yaml", true},
		{"no compose", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()

			if tc.filename != "" {
				require.NoError(t, os.WriteFile(filepath.Join(dir, tc.filename), []byte("services: {}"), 0o600))
			}

			result := findComposeFile(dir)
			if tc.found {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, tc.filename)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestCheckSecretLength_BoundaryValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       string
		wantError   bool
		wantWarning bool
	}{
		{"empty", "", true, false},
		{"1 byte", "x", false, true},
		{"31 bytes", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false, true},
		{"32 bytes", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false, false},
		{"43 bytes", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false, false},
		{"100 bytes", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := &SecretValidationResult{Valid: true}
			checkSecretLength("test.secret", tc.value, result)

			if tc.wantError {
				assert.False(t, result.Valid)
				assert.NotEmpty(t, result.Errors)
			} else {
				assert.True(t, result.Valid)
			}

			if tc.wantWarning {
				assert.NotEmpty(t, result.Warnings)
			}

			if !tc.wantError && !tc.wantWarning {
				assert.Empty(t, result.Errors)
				assert.Empty(t, result.Warnings)
			}
		})
	}
}

func TestValidateSecrets_ComposeEnvNonStringValue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	compose := `services:
  myapp:
    image: myapp:latest
    environment:
      DB_PASSWORD: 12345
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ComposeEnvSafeFileRef(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	compose := `services:
  myapp:
    image: myapp:latest
    environment:
      DB_SECRET: "file:///run/secrets/db_secret"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ComposeEnvRunSecretsRef(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	compose := `services:
  myapp:
    image: myapp:latest
    environment:
      SOME_TOKEN_FILE: "/run/secrets/my_token"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ComposeEnvNonSecretKey(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	compose := `services:
  myapp:
    image: myapp:latest
    environment:
      APP_PORT: "8080"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), 0o600))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func findRealDeploymentPath(name string) string {
	candidates := []string{
		filepath.Join("deployments", name),
		filepath.Join("..", "..", "..", "..", "deployments", name),
	}

	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(p)

			return abs
		}
	}

	return ""
}
