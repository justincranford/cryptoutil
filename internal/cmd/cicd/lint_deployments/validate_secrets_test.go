package lint_deployments

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
	require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "my_password.secret"),
		[]byte("this-is-a-very-long-secret-value-with-enough-entropy-for-security"), cryptoutilSharedMagic.CacheFilePermissions))

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
	require.NoError(t, os.WriteFile(f, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(f)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "not a directory")
}

func TestValidateSecrets_EmptySecretFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "my_password.secret"), []byte(""), cryptoutilSharedMagic.CacheFilePermissions))

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
	require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "hash_pepper.secret"), []byte("short"), cryptoutilSharedMagic.CacheFilePermissions))

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
	require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	base64Value := "4b0beuNCVMFMjA4y2/WvULfaCZiE6TLnPctdiSdtVrI="
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "hash_pepper.secret"), []byte(base64Value), cryptoutilSharedMagic.CacheFilePermissions))

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
	require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "readme.md"), []byte("short"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateSecrets_NonHighEntropySecretIgnored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "postgres_database.secret"), []byte("mydb"), cryptoutilSharedMagic.CacheFilePermissions))

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
	require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.Mkdir(filepath.Join(secretsDir, "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_UnreadableSecretFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	f := filepath.Join(secretsDir, "api_key.secret")
	require.NoError(t, os.WriteFile(f, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.Chmod(f, 0o000))

	t.Cleanup(func() { _ = os.Chmod(f, cryptoutilSharedMagic.CacheFilePermissions) })

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "cannot read secret file")
}

func TestValidateSecrets_UnreadableSecretsDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.Chmod(secretsDir, 0o000))

	t.Cleanup(func() { _ = os.Chmod(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute) })

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "cannot read secrets directory")
}

func TestValidateSecrets_ConfigInlineSecret(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	config := "database-password: supersecretpassword123\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), cryptoutilSharedMagic.CacheFilePermissions))

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
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	config := "database-password: file:///run/secrets/db_password\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestValidateSecrets_ConfigNestedInlineSecret(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	config := "auth:\n  api-key: my-hardcoded-api-key-value\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "auth.api-key")
}

func TestValidateSecrets_ConfigEmptySecretValue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	config := "database-password: \"\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ConfigNonSecretField(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	config := "bind-public-port: 8080\nhost: localhost\n"
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "app.yml"), []byte(config), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ConfigInvalidYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "bad.yml"), []byte(":\n  - :\n:"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ConfigUnreadableFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	f := filepath.Join(configsDir, "unreadable.yml")
	require.NoError(t, os.WriteFile(f, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.Chmod(f, 0o000))

	t.Cleanup(func() { _ = os.Chmod(f, cryptoutilSharedMagic.CacheFilePermissions) })

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "cannot read config file")
}

func TestValidateSecrets_ConfigNonYAMLIgnored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(configsDir, "readme.md"), []byte("api-key: secret"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ConfigSubdirIgnored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.Mkdir(filepath.Join(configsDir, "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_UnreadableConfigsDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.Chmod(configsDir, 0o000))

	t.Cleanup(func() { _ = os.Chmod(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute) })

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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ComposeInvalidYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(":\n  - :\n:"), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestValidateSecrets_ComposeUnreadable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "compose.yml")
	require.NoError(t, os.WriteFile(f, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.Chmod(f, 0o000))

	t.Cleanup(func() { _ = os.Chmod(f, cryptoutilSharedMagic.CacheFilePermissions) })

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
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(compose), cryptoutilSharedMagic.CacheFilePermissions))

	result, err := ValidateSecrets(dir)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

