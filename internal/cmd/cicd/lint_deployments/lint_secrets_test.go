package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateProductSecrets_AllPresent verifies no errors when all product secrets exist.
func TestValidateProductSecrets_AllPresent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := filepath.Join(tmpDir, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, dirPermissions))

	productName := "identity"

	// Create hash_pepper secret.
	require.NoError(t, os.WriteFile(
		filepath.Join(secretsDir, productName+"-hash_pepper.secret"),
		[]byte("pepper"), filePermissions))

	// Create all .never files.
	neverFiles := []string{
		"-unseal_1of5.secret.never",
		"-unseal_2of5.secret.never",
		"-unseal_3of5.secret.never",
		"-unseal_4of5.secret.never",
		"-unseal_5of5.secret.never",
		"-postgres_username.secret.never",
		"-postgres_password.secret.never",
		"-postgres_database.secret.never",
		"-postgres_url.secret.never",
	}

	for _, suffix := range neverFiles {
		require.NoError(t, os.WriteFile(
			filepath.Join(secretsDir, productName+suffix),
			[]byte("never"), filePermissions))
	}

	result := &ValidationResult{Valid: true}
	validateProductSecrets(tmpDir, productName, result)

	assert.True(t, result.Valid, "expected valid when all product secrets present")
	assert.Empty(t, result.MissingSecrets, "expected no missing secrets")
	assert.Empty(t, result.Errors, "expected no errors")
}

// TestValidateProductSecrets_Missing verifies errors when product secrets are missing.
func TestValidateProductSecrets_Missing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := filepath.Join(tmpDir, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, dirPermissions))

	result := &ValidationResult{Valid: true}
	validateProductSecrets(tmpDir, "sm", result)

	assert.False(t, result.Valid, "expected invalid when secrets missing")
	assert.NotEmpty(t, result.MissingSecrets, "expected missing secrets reported")

	// Should report missing hash_pepper + 9 .never files = 10 total.
	expectedMissing := 10
	assert.Len(t, result.MissingSecrets, expectedMissing)
}

// TestValidateSuiteSecrets_AllPresent verifies no errors when all suite secrets exist.
func TestValidateSuiteSecrets_AllPresent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := filepath.Join(tmpDir, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, dirPermissions))

	// Create hash_pepper.
	require.NoError(t, os.WriteFile(
		filepath.Join(secretsDir, "cryptoutil-hash_pepper.secret"),
		[]byte("pepper"), filePermissions))

	// Create all .never files.
	neverFiles := []string{
		"cryptoutil-unseal_1of5.secret.never",
		"cryptoutil-unseal_2of5.secret.never",
		"cryptoutil-unseal_3of5.secret.never",
		"cryptoutil-unseal_4of5.secret.never",
		"cryptoutil-unseal_5of5.secret.never",
		"cryptoutil-postgres_username.secret.never",
		"cryptoutil-postgres_password.secret.never",
		"cryptoutil-postgres_database.secret.never",
		"cryptoutil-postgres_url.secret.never",
	}

	for _, name := range neverFiles {
		require.NoError(t, os.WriteFile(
			filepath.Join(secretsDir, name),
			[]byte("never"), filePermissions))
	}

	result := &ValidationResult{Valid: true}
	validateSuiteSecrets(tmpDir, result)

	assert.True(t, result.Valid, "expected valid when all suite secrets present")
	assert.Empty(t, result.MissingSecrets, "expected no missing secrets")
}

// TestValidateSuiteSecrets_Missing verifies errors when suite secrets are missing.
func TestValidateSuiteSecrets_Missing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	secretsDir := filepath.Join(tmpDir, "secrets")
	require.NoError(t, os.MkdirAll(secretsDir, dirPermissions))

	result := &ValidationResult{Valid: true}
	validateSuiteSecrets(tmpDir, result)

	assert.False(t, result.Valid, "expected invalid when suite secrets missing")

	// Should report missing hash_pepper + 9 .never files = 10 total.
	expectedMissing := 10
	assert.Len(t, result.MissingSecrets, expectedMissing)
}

// TestCheckHardcodedCredentials tests credential detection in compose files.
func TestCheckHardcodedCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		wantValid bool
		wantErrs  int
	}{
		{
			name: "clean compose no credentials",
			content: `name: test
services:
  web:
    image: nginx
`,
			wantValid: true,
			wantErrs:  0,
		},
		{
			name: "hardcoded POSTGRES_USER detected",
			content: `services:
  db:
    environment:
      POSTGRES_USER: admin
`,
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "POSTGRES_USER_FILE is safe",
			content: `services:
  db:
    environment:
      POSTGRES_USER_FILE: /run/secrets/postgres_user
`,
			wantValid: true,
			wantErrs:  0,
		},
		{
			name: "hardcoded POSTGRES_PASSWORD detected",
			content: `services:
  db:
    environment:
      POSTGRES_PASSWORD: secret123
`,
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "hardcoded POSTGRES_DB detected",
			content: `services:
  db:
    environment:
      POSTGRES_DB: mydb
`,
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "hardcoded postgresql URL detected",
			content: `services:
  app:
    environment:
      DATABASE_URL: postgresql://user:pass@db:5432/mydb
`,
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "postgres URL with secrets ref is safe",
			content: `services:
  app:
    environment:
      DATABASE_URL: file:///run/secrets/postgres_url.secret
`,
			wantValid: true,
			wantErrs:  0,
		},
		{
			name: "commented postgres URL is safe",
			content: `services:
  app:
    environment:
      # DATABASE_URL: postgres://user:pass@db:5432/mydb
`,
			wantValid: true,
			wantErrs:  0,
		},
		{
			name: "multiple credential violations",
			content: `services:
  db:
    environment:
      POSTGRES_USER: admin
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: mydb
`,
			wantValid: false,
			wantErrs:  3,
		},
		{
			name:      "line numbers are correct",
			content:   "line1\nline2\nPOSTGRES_USER: admin\nline4\nPOSTGRES_DB: mydb\n",
			wantValid: false,
			wantErrs:  2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			composePath := filepath.Join(tmpDir, "compose.yml")
			require.NoError(t, os.WriteFile(composePath, []byte(tc.content), filePermissions))

			result := &ValidationResult{Valid: true}
			checkHardcodedCredentials(tmpDir, result)

			assert.Equal(t, tc.wantValid, result.Valid)
			assert.Len(t, result.Errors, tc.wantErrs)

			if tc.name == "line numbers are correct" {
				assert.Contains(t, result.Errors[0], ":3:", "POSTGRES_USER should be on line 3")
				assert.Contains(t, result.Errors[1], ":5:", "POSTGRES_DB should be on line 5")
			}
		})
	}
}

// TestCheckHardcodedCredentials_NoComposeFile verifies no error when compose file is missing.
func TestCheckHardcodedCredentials_NoComposeFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	result := &ValidationResult{Valid: true}

	checkHardcodedCredentials(tmpDir, result)

	assert.True(t, result.Valid, "should remain valid when no compose file exists")
	assert.Empty(t, result.Errors)
}
