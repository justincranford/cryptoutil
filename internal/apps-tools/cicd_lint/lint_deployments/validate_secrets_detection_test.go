package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestValidateSecrets_NoSecretsDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
}

func TestValidateSecrets_NoConfigsDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
}

func TestValidateSecrets_NoComposeFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
}

func TestValidateSecrets_RealSmIM(t *testing.T) {
	t.Parallel()

	deploymentPath := findRealDeploymentPath(cryptoutilSharedMagic.OTLPServiceSMKMS)
	if deploymentPath == "" {
		t.Skip("sm-kms deployment not found")
	}

	result, err := ValidateSecrets(deploymentPath)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestValidateSecrets_DotNeverSecretFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, validatorNameSecrets)
	require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "unseal_1of5.secret.never"),
		[]byte(testShort), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.NotEmpty(t, result.Warnings)
	require.Contains(t, result.Warnings[0], "has 5 bytes")
	require.Contains(t, result.Warnings[0], "minimum recommended: 32")
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.DOCKER_COMPOSE_YML), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.Contains(t, result.Errors[0], testInlineSecret)
}

func TestValidateSecrets_ConfigNonStringValue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, cryptoutilSharedMagic.CICDConfigsDir)
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	config := "api-key: 12345\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, testAppYml), []byte(config), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
}

func TestValidateSecrets_ConfigSqliteURL(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, cryptoutilSharedMagic.CICDConfigsDir)
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	config := "database-password: sqlite:///tmp/test.db\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, testAppYml), []byte(config), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
}

func TestValidateSecrets_ConfigMemoryRef(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, cryptoutilSharedMagic.CICDConfigsDir)
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	config := "db-password: \":memory:\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, testAppYml), []byte(config), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
}

func TestFormatSecretValidationResult_Nil(t *testing.T) {
	t.Parallel()

	output := FormatSecretValidationResult(nil)
	require.Contains(t, output, cryptoutilSharedMagic.TestStatusSkip)
}

func TestFormatSecretValidationResult_Valid(t *testing.T) {
	t.Parallel()

	result := &SecretValidationResult{Valid: true}
	output := FormatSecretValidationResult(result)
	require.Contains(t, output, cryptoutilSharedMagic.TestStatusPass)
}

func TestFormatSecretValidationResult_Errors(t *testing.T) {
	t.Parallel()

	result := &SecretValidationResult{
		Valid:    false,
		Errors:   []string{"inline secret found"},
		Warnings: []string{"short secret"},
	}
	output := FormatSecretValidationResult(result)
	require.Contains(t, output, cryptoutilSharedMagic.TestStatusFail)
	require.Contains(t, output, "ERROR: inline secret found")
	require.Contains(t, output, "WARN: short secret")
}

func TestIsSecretFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{testSecretFile, "password.secret", true},
		{"secret never file", "unseal.secret.never", true},
		{"not secret", testReadmeMd, false},
		{"yml file", testConfigYml, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expect, isSecretFile(tc.input))
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
		{valuePassword, "db_password.secret", true},
		{valuePepper, valueHashPepperV3Secret, true},
		{"unseal", valueUnseal1of5Secret, true},
		{valueAPIKeyValue2E9BC6C9, "my_api_key.secret", true},
		{"secret_key", "secret_key.secret", true},
		{"database name", valuePostgresDatabaseSecret, false},
		{"username", valuePostgresUsernameSecret, false},
		{"url", valuePostgresURLSecret, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expect, isHighEntropySecretFile(tc.input))
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
		{valuePassword, testDatabasePassword, true},
		{"PASSWD", "POSTGRES_PASSWD", true},
		{"secret", "my-secret", true},
		{"TOKEN", "ACCESS_TOKEN", true},
		{valueAPIKey, valueAPIKey, true},
		{valueAPIKeyValue2E9BC6C9, "API_KEY", true},
		{valuePrivateKey, valuePrivateKey, true},
		{valuePepper, "hash-pepper", true},
		{testHostname, testHostname, false},
		{"port", "bind-port", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expect, isSecretFieldName(tc.input))
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
		{cryptoutilSharedMagic.TestDatabaseSQLite, "sqlite:///tmp/test.db", true},
		{"memory", cryptoutilSharedMagic.SQLiteMemoryPlaceholder, true},
		{"inline", "my-hardcoded-password", false},
		{"postgres URL", "postgres://user:pass@host/db", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expect, isSafeReference(tc.input))
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
		{cryptoutilSharedMagic.COMPOSE_YML, cryptoutilSharedMagic.COMPOSE_YML, true},
		{cryptoutilSharedMagic.COMPOSE_YAML, cryptoutilSharedMagic.COMPOSE_YAML, true},
		{cryptoutilSharedMagic.DOCKER_COMPOSE_YML, cryptoutilSharedMagic.DOCKER_COMPOSE_YML, true},
		{"no compose", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()

			if tc.filename != "" {
				require.NoError(t, os.WriteFile(filepath.Join(dir, tc.filename), []byte("services: {}"), cryptoutilSharedMagic.CacheFilePermissions))
			}

			result := findComposeFile(dir)
			if tc.found {
				require.NotEmpty(t, result)
				require.Contains(t, result, tc.filename)
			} else {
				require.Empty(t, result)
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
		{cryptoutilSharedMagic.EMPTY, "", true, false},
		{"1 byte", testX, false, true},
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
				require.False(t, result.Valid)
				require.NotEmpty(t, result.Errors)
			} else {
				require.True(t, result.Valid)
			}

			if tc.wantWarning {
				require.NotEmpty(t, result.Warnings)
			}

			if !tc.wantError && !tc.wantWarning {
				require.Empty(t, result.Errors)
				require.Empty(t, result.Warnings)
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
}

func TestValidateSecrets_ComposeInvalidYAMLSilentSkip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	invalidYAML := []byte("services:\n  - [invalid: yaml: {broken")
	require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), invalidYAML, cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Invalid YAML in compose is silently skipped (no error, result still valid).
	require.True(t, result.Valid)
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	require.True(t, result.Valid)
}

func findRealDeploymentPath(name string) string {
	candidates := []string{
		filepath.Join(defaultDeploymentsDir, name),
		filepath.Join("..", "..", "..", "..", defaultDeploymentsDir, name),
	}

	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(p)

			return abs
		}
	}

	return ""
}
