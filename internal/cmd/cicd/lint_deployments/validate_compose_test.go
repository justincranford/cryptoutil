package lint_deployments

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateComposeFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    string
		wantValid  bool
		wantErrors []string
		wantWarns  []string
	}{
		{
			name: "valid compose file",
			content: `services:
  myapp:
    image: myapp:latest
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "wget", "-q", "-O", "/dev/null", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
secrets:
  my_secret.secret:
    file: ./secrets/my_secret.secret
`,
			wantValid: true,
		},
		{
			name:       "invalid YAML",
			content:    "services:\n  myapp:\n    image: [invalid",
			wantValid:  false,
			wantErrors: []string{"YAML parse error"},
		},
		{
			name:       "no services defined",
			content:    "version: '3'\n",
			wantValid:  false,
			wantErrors: []string{"no services defined"},
		},
		{
			name: "port conflict",
			content: `services:
  app1:
    image: app1:latest
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "true"]
  app2:
    image: app2:latest
    ports:
      - "8080:9090"
    healthcheck:
      test: ["CMD", "true"]
`,
			wantValid:  false,
			wantErrors: []string{"port conflict"},
		},
		{
			name: "missing healthcheck",
			content: `services:
  myapp:
    image: myapp:latest
`,
			wantValid:  false,
			wantErrors: []string{"missing healthcheck"},
		},
		{
			name: "builder exempt from healthcheck",
			content: `services:
  builder-myapp:
    image: myapp:latest
    entrypoint: ["sh", "-c"]
    command: ["echo 'done'"]
`,
			wantValid: true,
		},
		{
			name: "healthcheck prefix exempt",
			content: `services:
  healthcheck-secrets:
    image: alpine:latest
    command: ["ls", "/run/secrets"]
`,
			wantValid: true,
		},
		{
			name: "undefined secret no secrets section",
			content: `services:
  myapp:
    image: myapp:latest
    secrets:
      - my_secret.secret
    healthcheck:
      test: ["CMD", "true"]
`,
			wantValid:  false,
			wantErrors: []string{"no secrets section defined"},
		},
		{
			name: "undefined secret with secrets section",
			content: `services:
  myapp:
    image: myapp:latest
    secrets:
      - missing_secret.secret
    healthcheck:
      test: ["CMD", "true"]
secrets:
  other_secret.secret:
    file: ./secrets/other.secret
`,
			wantValid:  false,
			wantErrors: []string{"references undefined secret"},
		},
		{
			name: "hardcoded credentials",
			content: `services:
  mydb:
    image: postgres:18
    environment:
      POSTGRES_PASSWORD: mysecretpassword
    healthcheck:
      test: ["CMD", "pg_isready"]
`,
			wantValid:  false,
			wantErrors: []string{"hardcoded credentials"},
		},
		{
			name: "credentials via file reference safe",
			content: `services:
  mydb:
    image: postgres:18
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password.secret
    healthcheck:
      test: ["CMD", "pg_isready"]
`,
			wantValid: true,
		},
		{
			name: "credentials via variable reference safe",
			content: `services:
  mydb:
    image: postgres:18
    environment:
      POSTGRES_PASSWORD: ${DB_PASS}
    healthcheck:
      test: ["CMD", "pg_isready"]
`,
			wantValid: true,
		},
		{
			name: "dangerous bind mount docker.sock",
			content: `services:
  myapp:
    image: myapp:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    healthcheck:
      test: ["CMD", "true"]
`,
			wantValid:  false,
			wantErrors: []string{"dangerous bind mount detected"},
		},
		{
			name: "depends on unknown service warns",
			content: `services:
  myapp:
    image: myapp:latest
    depends_on:
      external-svc:
        condition: service_started
    healthcheck:
      test: ["CMD", "true"]
`,
			wantValid: true,
			wantWarns: []string{"not defined locally"},
		},
		{
			name: "depends on list format valid",
			content: `services:
  myapp:
    image: myapp:latest
    depends_on:
      - valid-svc
    healthcheck:
      test: ["CMD", "true"]
  valid-svc:
    image: valid:latest
    healthcheck:
      test: ["CMD", "true"]
`,
			wantValid: true,
		},
		{
			name: "echo entrypoint exempt from healthcheck",
			content: `services:
  init-svc:
    image: alpine:latest
    entrypoint: ["sh", "-c", "echo init done"]
`,
			wantValid: true,
		},
		{
			name: "environment list format credentials",
			content: `services:
  mydb:
    image: postgres:18
    environment:
      - POSTGRES_PASSWORD=hardcoded123
    healthcheck:
      test: ["CMD", "pg_isready"]
`,
			wantValid:  false,
			wantErrors: []string{"hardcoded credentials"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			composePath := filepath.Join(dir, "compose.yml")
			require.NoError(t, os.WriteFile(composePath, []byte(tc.content), filePermissions))

			result, err := ValidateComposeFile(composePath)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.wantValid, result.Valid, "validity mismatch; errors=%v warnings=%v", result.Errors, result.Warnings)

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

func TestValidateComposeFile_FileNotFound(t *testing.T) {
	t.Parallel()

	result, err := ValidateComposeFile("/nonexistent/compose.yml")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.True(t, containsSubstring(result.Errors, "YAML parse error"))
}

func TestValidateComposeFile_WithIncludes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	includeDir := filepath.Join(dir, "shared")
	require.NoError(t, os.MkdirAll(includeDir, dirPermissions))

	includeContent := `secrets:
  shared_secret.secret:
    file: ./secrets/shared.secret
services:
  shared-svc:
    image: shared:latest
    healthcheck:
      test: ["CMD", "true"]
`
	require.NoError(t, os.WriteFile(filepath.Join(includeDir, "compose.yml"),
		[]byte(includeContent), filePermissions))

	mainContent := `include:
  - path: shared/compose.yml
services:
  myapp:
    image: myapp:latest
    secrets:
      - shared_secret.secret
    healthcheck:
      test: ["CMD", "true"]
`
	mainPath := filepath.Join(dir, "compose.yml")
	require.NoError(t, os.WriteFile(mainPath, []byte(mainContent), filePermissions))

	result, err := ValidateComposeFile(mainPath)
	require.NoError(t, err)
	assert.True(t, result.Valid, "should be valid with included secrets: %v", result.Errors)
}

// containsSubstring checks if any string in slice contains the given substring.
func containsSubstring(slice []string, substr string) bool {
for _, s := range slice {
if strings.Contains(s, substr) {
return true
}
}

return false
}
