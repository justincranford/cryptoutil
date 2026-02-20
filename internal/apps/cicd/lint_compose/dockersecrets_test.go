// Copyright (c) 2025 Justin Cranford

package lint_compose

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintComposeDockerSecrets "cryptoutil/internal/apps/cicd/lint_compose/docker_secrets"

	"github.com/stretchr/testify/require"
)

func TestLintDockerSecrets_NoComposeFiles(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{}

	err := lintComposeDockerSecrets.Check(logger, filesByExtension)
	require.NoError(t, err, "lint should pass with no compose files")
}

func TestLintDockerSecrets_ValidComposeFile(t *testing.T) {
	t.Parallel()

	// Create temp dir with valid compose file using Docker secrets.
	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, "docker-compose.yml")

	validContent := `version: '3.8'
services:
  app:
    image: myapp:latest
    secrets:
      - postgres_password
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password

secrets:
  postgres_password:
    file: ./secrets/postgres_password.secret
`
	err := os.WriteFile(composeFile, []byte(validContent), 0o600)
	require.NoError(t, err)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	filesByExtension := map[string][]string{
		"yml": {composeFile},
	}

	err = lintComposeDockerSecrets.Check(logger, filesByExtension)
	require.NoError(t, err, "lint should pass with valid Docker secrets pattern")
}

func TestLintDockerSecrets_InlinePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name: "inline POSTGRES_PASSWORD",
			content: `version: '3.8'
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: mysecretpassword
`,
			wantErr:     true,
			errContains: "inline credential violations",
		},
		{
			name: "inline POSTGRES_USER",
			content: `version: '3.8'
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_USER: admin
`,
			wantErr:     true,
			errContains: "inline credential violations",
		},
		{
			name: "inline DATABASE_URL",
			content: `version: '3.8'
services:
  app:
    image: myapp:latest
    environment:
      DATABASE_URL: postgres://user:pass@localhost:5432/db
`,
			wantErr:     true,
			errContains: "inline credential violations",
		},
		{
			name: "inline API_KEY",
			content: `version: '3.8'
services:
  app:
    image: myapp:latest
    environment:
      API_KEY: sk-1234567890
`,
			wantErr:     true,
			errContains: "inline credential violations",
		},
		{
			name: "inline SECRET_KEY",
			content: `version: '3.8'
services:
  app:
    image: myapp:latest
    environment:
      SECRET_KEY: my-secret-key-value
`,
			wantErr:     true,
			errContains: "inline credential violations",
		},
		{
			name: "valid _FILE pattern",
			content: `version: '3.8'
services:
  db:
    image: postgres:16
    secrets:
      - postgres_password
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password

secrets:
  postgres_password:
    file: ./secrets/postgres_password.secret
`,
			wantErr: false,
		},
		{
			name: "valid file:// URI pattern",
			content: `version: '3.8'
services:
  app:
    image: myapp:latest
    secrets:
      - db_url
    command:
      - --database-url=file:///run/secrets/db_url

secrets:
  db_url:
    file: ./secrets/db_url.secret
`,
			wantErr: false,
		},
		{
			name: "commented out credential",
			content: `version: '3.8'
services:
  db:
    image: postgres:16
    environment:
      # POSTGRES_PASSWORD: old_password
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
`,
			wantErr: false,
		},
		{
			name: "non-environment credential reference",
			content: `version: '3.8'
services:
  app:
    image: myapp:latest
    ports:
      - "8080:8080"
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			composeFile := filepath.Join(tempDir, "docker-compose.yml")

			err := os.WriteFile(composeFile, []byte(tt.content), 0o600)
			require.NoError(t, err)

			logger := cryptoutilCmdCicdCommon.NewLogger("test")
			filesByExtension := map[string][]string{
				"yml": {composeFile},
			}

			err = lintComposeDockerSecrets.Check(logger, filesByExtension)

			if tt.wantErr {
				require.Error(t, err, "lint should fail for inline credentials")
				require.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err, "lint should pass for valid Docker secrets pattern")
			}
		})
	}
}

func TestCheckComposeFileSecrets_NonExistentFile(t *testing.T) {
	t.Parallel()

	violations, err := lintComposeDockerSecrets.CheckComposeFileSecrets("/nonexistent/file.yml")
	require.Error(t, err, "should fail for non-existent file")
	require.Nil(t, violations)
}

func TestCheckComposeFileSecrets_MultipleViolations(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, "docker-compose.yml")

	content := `version: '3.8'
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_USER: admin
      POSTGRES_PASSWORD: secretpass
      POSTGRES_DB: mydb
`
	err := os.WriteFile(composeFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := lintComposeDockerSecrets.CheckComposeFileSecrets(composeFile)
	require.NoError(t, err)
	require.Len(t, violations, 3, "should detect 3 inline credentials")
}

func TestCheckComposeFileSecrets_MixedValidInvalid(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, "docker-compose.yml")

	content := `version: '3.8'
services:
  db:
    image: postgres:16
    secrets:
      - postgres_password
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
      POSTGRES_USER: admin

secrets:
  postgres_password:
    file: ./secrets/postgres_password.secret
`
	err := os.WriteFile(composeFile, []byte(content), 0o600)
	require.NoError(t, err)

	violations, err := lintComposeDockerSecrets.CheckComposeFileSecrets(composeFile)
	require.NoError(t, err)
	require.Len(t, violations, 1, "should detect 1 inline credential (POSTGRES_USER)")
}
