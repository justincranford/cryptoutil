package lint_deployments

import (
	"os"
	"path/filepath"
	"runtime"
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
			name: "path not found",
			setup: func(t *testing.T) string {
				t.Helper()

				return "/nonexistent/path/xyz"
			},
			wantErrContains: []string{"[ValidateSecrets]", "path not found"},
		},
		{
			name: "path is file not directory",
			setup: func(t *testing.T) string {
				t.Helper()
				f := filepath.Join(t.TempDir(), "file.txt")
				require.NoError(t, os.WriteFile(f, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))

				return f
			},
			wantErrContains: []string{"[ValidateSecrets]", "not a directory"},
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
		{name: "valid long secret", filename: "my_password.secret", content: "this-is-a-very-long-secret-value-with-enough-entropy-for-security", wantValid: true},
		{name: "empty secret file", filename: "my_password.secret", content: "", wantErrContains: []string{"[ValidateSecrets]", "is empty", "ENG-HANDBOOK.md Section 12.6"}},
		{name: "short secret warns", filename: "hash_pepper.secret", content: "short", wantValid: true, wantWarnContains: []string{"has 5 bytes", "minimum recommended: 32"}},
		{name: "base64 value meets threshold", filename: "hash_pepper.secret", content: "4b0beuNCVMFMjA4y2/WvULfaCZiE6TLnPctdiSdtVrI=", wantValid: true},
		{name: "non-secret file ignored", filename: "readme.md", content: "short", wantValid: true},
		{name: "non-high-entropy secret ignored", filename: "postgres_database.secret", content: "mydb", wantValid: true},
		{name: "subdirectory skipped", makeSubdir: true, wantValid: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			secretsDir := filepath.Join(dir, "secrets")
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

func TestValidateSecrets_UnreadablePaths(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		t.Skip("os.Chmod 0o000 does not restrict access on Windows NTFS")
	}

	tests := []struct {
		name             string
		setup            func(t *testing.T) string
		wantWarnContains string
	}{
		{
			name: "unreadable secret file",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				secretsDir := filepath.Join(dir, "secrets")
				require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				f := filepath.Join(secretsDir, "api_key.secret")
				require.NoError(t, os.WriteFile(f, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, os.Chmod(f, 0o000))
				t.Cleanup(func() { _ = os.Chmod(f, cryptoutilSharedMagic.CacheFilePermissions) })

				return dir
			},
			wantWarnContains: "cannot read secret file",
		},
		{
			name: "unreadable secrets directory",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				secretsDir := filepath.Join(dir, "secrets")
				require.NoError(t, os.Mkdir(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.Chmod(secretsDir, 0o000))
				t.Cleanup(func() {
					_ = os.Chmod(secretsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
				})

				return dir
			},
			wantWarnContains: "cannot read secrets directory",
		},
		{
			name: "unreadable config file",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				configsDir := filepath.Join(dir, cryptoutilSharedMagic.CICDConfigsDir)
				require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				f := filepath.Join(configsDir, "unreadable.yml")
				require.NoError(t, os.WriteFile(f, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, os.Chmod(f, 0o000))
				t.Cleanup(func() { _ = os.Chmod(f, cryptoutilSharedMagic.CacheFilePermissions) })

				return dir
			},
			wantWarnContains: "cannot read config file",
		},
		{
			name: "unreadable configs directory",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				configsDir := filepath.Join(dir, cryptoutilSharedMagic.CICDConfigsDir)
				require.NoError(t, os.Mkdir(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, os.Chmod(configsDir, 0o000))
				t.Cleanup(func() {
					_ = os.Chmod(configsDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
				})

				return dir
			},
			wantWarnContains: "cannot read configs directory",
		},
		{
			name: "unreadable compose file",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				f := filepath.Join(dir, "compose.yml")
				require.NoError(t, os.WriteFile(f, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, os.Chmod(f, 0o000))
				t.Cleanup(func() { _ = os.Chmod(f, cryptoutilSharedMagic.CacheFilePermissions) })

				return dir
			},
			wantWarnContains: "cannot read compose file",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := ValidateSecrets(tc.setup(t))
			require.NoError(t, err)
			require.NotEmpty(t, result.Warnings)
			require.Contains(t, result.Warnings[0], tc.wantWarnContains)
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
		{name: "inline secret detected", filename: "app.yml", content: "database-password: supersecretpassword123\n", wantErrContains: []string{"inline secret"}},
		{name: "safe file reference", filename: "app.yml", content: "database-password: file:///run/secrets/db_password\n", wantValid: true},
		{name: "nested inline secret", filename: "app.yml", content: "auth:\n  api-key: my-hardcoded-api-key-value\n", wantErrContains: []string{"auth.api-key"}},
		{name: "empty secret value accepted", filename: "app.yml", content: "database-password: \"\"\n", wantValid: true},
		{name: "non-secret field accepted", filename: "app.yml", content: "bind-public-port: 8080\nhost: localhost\n", wantValid: true},
		{name: "invalid YAML tolerated", filename: "bad.yml", content: ":\n  - :\n:", wantValid: true},
		{name: "non-YAML file ignored", filename: "readme.md", content: "api-key: secret", wantValid: true},
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
			name:            "inline secret detected",
			compose:         "services:\n  myapp:\n    image: myapp:latest\n    environment:\n      DB_PASSWORD: \"hardcoded-password-value\"\n",
			wantErrContains: []string{"inline secret"},
		},
		{name: "secret file reference accepted", compose: "services:\n  myapp:\n    image: myapp:latest\n    environment:\n      DB_PASSWORD_FILE: \"/run/secrets/db_password\"\n", wantValid: true},
		{name: "non-secret env accepted", compose: "services:\n  myapp:\n    image: myapp:latest\n    environment:\n      APP_NAME: \"my-service\"\n", wantValid: true},
		{name: "empty secret value accepted", compose: "services:\n  myapp:\n    image: myapp:latest\n    environment:\n      DB_PASSWORD: \"\"\n", wantValid: true},
		{name: "invalid YAML tolerated", compose: ":\n  - :\n:", wantValid: true},
		{name: "non-map env format accepted", compose: "services:\n  myapp:\n    image: myapp:latest\n    environment:\n      - DB_PASSWORD=hardcoded\n", wantValid: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(tc.compose), cryptoutilSharedMagic.CacheFilePermissions))

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
