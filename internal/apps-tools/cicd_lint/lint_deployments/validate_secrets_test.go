package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestValidateSecrets_PathErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setup           func(t *testing.T) string
		wantErrContains []string
	}{
		{
			name: testPathNotFound,
			setup: func(t *testing.T) string {
				t.Helper()

				return "/nonexistent/path/xyz"
			},
			wantErrContains: []string{testValidateSecrets, testPathNotFound},
		},
		{
			name: "path is file not directory",
			setup: func(t *testing.T) string {
				t.Helper()
				f := filepath.Join(t.TempDir(), "file.txt")
				require.NoError(t, os.WriteFile(f, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))

				return f
			},
			wantErrContains: []string{testValidateSecrets, testNotADirectory},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := ValidateSecrets(tc.setup(t))
			require.NoError(t, err)
			require.False(t, result.Valid)
			require.NotEmpty(t, result.Errors)

			for _, s := range tc.wantErrContains {
				require.Contains(t, result.Errors[0], s)
			}
		})
	}
}

func TestValidateSecrets_SecretFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		filename         string
		content          string
		makeSubdir       bool
		wantValid        bool
		wantErrContains  []string
		wantWarnContains []string
	}{
		{name: "valid long secret", filename: testMyPasswordSecret, content: "this-is-a-very-long-secret-value-with-enough-entropy-for-security", wantValid: true},
		{name: "empty secret file", filename: testMyPasswordSecret, content: "", wantErrContains: []string{testValidateSecrets, "is empty", "ENG-HANDBOOK.md Section 13.3"}},
		{name: "short secret warns", filename: testHashPepperSecret, content: testShort, wantValid: true, wantWarnContains: []string{"has 5 bytes", "minimum recommended: 32"}},
		{name: "base64 value meets threshold", filename: testHashPepperSecret, content: "4b0beuNCVMFMjA4y2/WvULfaCZiE6TLnPctdiSdtVrI=", wantValid: true},
		{name: "non-secret file ignored", filename: testReadmeMd, content: testShort, wantValid: true},
		{name: "non-high-entropy secret ignored", filename: valuePostgresDatabaseSecret, content: "mydb", wantValid: true},
		{name: testSubdirectorySkipped, makeSubdir: true, wantValid: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			secretsDir := filepath.Join(dir, validatorNameSecrets)
			require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

			if tc.makeSubdir {
				require.NoError(t, os.Mkdir(filepath.Join(secretsDir, "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			} else {
				require.NoError(t, os.WriteFile(filepath.Join(secretsDir, tc.filename), []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))
			}

			result, err := ValidateSecrets(dir)
			require.NoError(t, err)
			require.Equal(t, tc.wantValid, result.Valid)

			if len(tc.wantErrContains) > 0 {
				require.NotEmpty(t, result.Errors)

				for _, s := range tc.wantErrContains {
					require.Contains(t, result.Errors[0], s)
				}
			} else {
				require.Empty(t, result.Errors)
			}

			if len(tc.wantWarnContains) > 0 {
				require.NotEmpty(t, result.Warnings)

				for _, s := range tc.wantWarnContains {
					require.Contains(t, result.Warnings[0], s)
				}
			}
		})
	}
}

func TestValidateSecrets_UnreadableSecretFile(t *testing.T) {
	t.Parallel()

	secretsDir := filepath.Join(t.TempDir(), validatorNameSecrets)
	require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(secretsDir, "api_key.secret"), []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))

	var readAttempted bool

	result := &SecretValidationResult{}
	validateSecretFileLengthsWithReadFile(secretsDir, result, func(string) ([]byte, error) {
		readAttempted = true

		return nil, os.ErrPermission
	})

	require.True(t, readAttempted)
	require.Len(t, result.Warnings, 1)
	require.Contains(t, result.Warnings[0], "cannot read secret file")
}

func TestValidateSecretHelpers_ReadFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		run         func(*testing.T, *SecretValidationResult)
		wantWarning string
	}{
		{
			name: "missing secrets directory",
			run: func(t *testing.T, result *SecretValidationResult) {
				t.Helper()
				validateSecretFileLengths(filepath.Join(t.TempDir(), "missing"), result)
			},
			wantWarning: testCannotReadSecretsDirectory,
		},
		{
			name: "missing configs directory",
			run: func(t *testing.T, result *SecretValidationResult) {
				t.Helper()
				validateConfigInlineSecrets(filepath.Join(t.TempDir(), "missing"), result)
			},
			wantWarning: testCannotReadConfigsDirectory,
		},
		{
			name: testMissingConfigFile,
			run: func(t *testing.T, result *SecretValidationResult) {
				t.Helper()
				checkConfigFileForInlineSecrets(filepath.Join(t.TempDir(), "missing.yml"), "missing.yml", result)
			},
			wantWarning: testCannotReadConfigFile,
		},
		{
			name: "missing compose file",
			run: func(t *testing.T, result *SecretValidationResult) {
				t.Helper()
				validateComposeInlineSecrets(filepath.Join(t.TempDir(), "missing.yml"), result)
			},
			wantWarning: "cannot read compose file",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := &SecretValidationResult{}
			tc.run(t, result)

			require.Len(t, result.Warnings, 1)
			require.Contains(t, result.Warnings[0], tc.wantWarning)
		})
	}
}

func TestValidateSecrets_ConfigFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		filename        string
		content         string
		makeSubdir      bool
		wantValid       bool
		wantErrContains []string
	}{
		{name: testInlineSecretDetected, filename: testAppYml, content: "database-password: supersecretpassword123\n", wantErrContains: []string{testInlineSecret}},
		{name: "safe file reference", filename: testAppYml, content: "database-password: file:///run/secrets/db_password\n", wantValid: true},
		{name: "nested inline secret", filename: testAppYml, content: "auth:\n  api-key: my-hardcoded-api-key-value\n", wantErrContains: []string{"auth.api-key"}},
		{name: testEmptySecretValueAccepted, filename: testAppYml, content: "database-password: \"\"\n", wantValid: true},
		{name: "non-secret field accepted", filename: testAppYml, content: "bind-public-port: 8080\nhost: localhost\n", wantValid: true},
		{name: testInvalidYAMLTolerated, filename: testBadYml, content: testStringValue, wantValid: true},
		{name: "non-YAML file ignored", filename: testReadmeMd, content: "api-key: secret", wantValid: true},
		{name: "subdirectory ignored", makeSubdir: true, wantValid: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			configsDir := filepath.Join(dir, cryptoutilSharedMagic.CICDConfigsDir)
			require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

			if tc.makeSubdir {
				require.NoError(t, os.Mkdir(filepath.Join(configsDir, "subdir"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			} else {
				require.NoError(t, os.WriteFile(filepath.Join(configsDir, tc.filename), []byte(tc.content), cryptoutilSharedMagic.CacheFilePermissions))
			}

			result, err := ValidateSecrets(dir)
			require.NoError(t, err)
			require.Equal(t, tc.wantValid, result.Valid)

			if len(tc.wantErrContains) > 0 {
				require.NotEmpty(t, result.Errors)

				for _, s := range tc.wantErrContains {
					require.Contains(t, result.Errors[0], s)
				}
			} else {
				require.Empty(t, result.Errors)
			}
		})
	}
}

func TestValidateSecrets_ComposeFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		compose         string
		wantValid       bool
		wantErrContains []string
	}{
		{
			name:            testInlineSecretDetected,
			compose:         "services:\n  myapp:\n    image: myapp:latest\n    environment:\n      DB_PASSWORD: \"hardcoded-password-value\"\n",
			wantErrContains: []string{testInlineSecret},
		},
		{name: "secret file reference accepted", compose: "services:\n  myapp:\n    image: myapp:latest\n    environment:\n      DB_PASSWORD_FILE: \"/run/secrets/db_password\"\n", wantValid: true},
		{name: "non-secret env accepted", compose: "services:\n  myapp:\n    image: myapp:latest\n    environment:\n      APP_NAME: \"my-service\"\n", wantValid: true},
		{name: testEmptySecretValueAccepted, compose: "services:\n  myapp:\n    image: myapp:latest\n    environment:\n      DB_PASSWORD: \"\"\n", wantValid: true},
		{name: testInvalidYAMLTolerated, compose: testStringValue, wantValid: true},
		{name: "non-map env format accepted", compose: "services:\n  myapp:\n    image: myapp:latest\n    environment:\n      - DB_PASSWORD=hardcoded\n", wantValid: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, cryptoutilSharedMagic.COMPOSE_YML), []byte(tc.compose), cryptoutilSharedMagic.CacheFilePermissions))

			result, err := ValidateSecrets(dir)
			require.NoError(t, err)
			require.Equal(t, tc.wantValid, result.Valid)

			if len(tc.wantErrContains) > 0 {
				require.NotEmpty(t, result.Errors)

				for _, s := range tc.wantErrContains {
					require.Contains(t, result.Errors[0], s)
				}
			} else {
				require.Empty(t, result.Errors)
			}
		})
	}
}
