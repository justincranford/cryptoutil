package lint_deployments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMainGenerateListings tests the generate-listings CLI subcommand.
func TestMainGenerateListings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) (string, string)
		wantCode int
	}{
		{
			name: "valid directories generate listings",
			setup: func(t *testing.T) (string, string) {
				t.Helper()

				tmpDir := t.TempDir()
				deploymentsDir := filepath.Join(tmpDir, "deployments")
				configsDir := filepath.Join(tmpDir, "configs")
				require.NoError(t, os.MkdirAll(filepath.Join(deploymentsDir, "jose-ja"), dirPermissions))
				require.NoError(t, os.MkdirAll(configsDir, dirPermissions))
				require.NoError(t, os.WriteFile(
					filepath.Join(deploymentsDir, "jose-ja", "compose.yml"),
					[]byte("name: jose-ja\n"), filePermissions))

				return deploymentsDir, configsDir
			},
			wantCode: 0,
		},
		{
			name: "nonexistent deployments dir fails",
			setup: func(t *testing.T) (string, string) {
				t.Helper()

				return "/nonexistent-deploy", "/nonexistent-config"
			},
			wantCode: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deployDir, configDir := tc.setup(t)
			got := mainGenerateListings([]string{deployDir, configDir})
			assert.Equal(t, tc.wantCode, got)
		})
	}
}

// TestMainValidateMirror tests the validate-mirror CLI subcommand.
func TestMainValidateMirror(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) (string, string)
		wantCode int
	}{
		{
			name: "valid mirror passes",
			setup: func(t *testing.T) (string, string) {
				t.Helper()

				tmpDir := t.TempDir()
				deploymentsDir := filepath.Join(tmpDir, "deployments")
				configsDir := filepath.Join(tmpDir, "configs")
				require.NoError(t, os.MkdirAll(filepath.Join(deploymentsDir, "jose-ja"), dirPermissions))
				require.NoError(t, os.MkdirAll(filepath.Join(configsDir, "jose"), dirPermissions))

				return deploymentsDir, configsDir
			},
			wantCode: 0,
		},
		{
			name: "missing mirror fails",
			setup: func(t *testing.T) (string, string) {
				t.Helper()

				tmpDir := t.TempDir()
				deploymentsDir := filepath.Join(tmpDir, "deployments")
				configsDir := filepath.Join(tmpDir, "configs")
				require.NoError(t, os.MkdirAll(filepath.Join(deploymentsDir, "jose-ja"), dirPermissions))
				require.NoError(t, os.MkdirAll(configsDir, dirPermissions))

				return deploymentsDir, configsDir
			},
			wantCode: 1,
		},
		{
			name: "nonexistent directory errors",
			setup: func(t *testing.T) (string, string) {
				t.Helper()

				return "/nonexistent-mirror-deploy", "/nonexistent-mirror-config"
			},
			wantCode: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deployDir, configDir := tc.setup(t)
			got := mainValidateMirror([]string{deployDir, configDir})
			assert.Equal(t, tc.wantCode, got)
		})
	}
}

// TestMainValidateCompose tests the validate-compose CLI subcommand.
func TestMainValidateCompose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		setup    func(t *testing.T) string
		wantCode int
	}{
		{
			name:     "no args fails",
			args:     []string{},
			wantCode: 1,
		},
		{
			name: "valid compose file passes",
			setup: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()
				composePath := filepath.Join(tmpDir, "compose.yml")
				content := "name: test\nservices:\n  web:\n    image: nginx\n    healthcheck:\n      test: [\"CMD\", \"curl\", \"-f\", \"http://localhost\"]\n      interval: 30s\n      timeout: 10s\n      retries: 3\n"
				require.NoError(t, os.WriteFile(composePath, []byte(content), filePermissions))

				return composePath
			},
			wantCode: 0,
		},
		{
			name: "nonexistent file fails",
			setup: func(t *testing.T) string {
				t.Helper()

				return "/nonexistent/compose.yml"
			},
			wantCode: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var args []string
			if tc.args != nil {
				args = tc.args
			} else if tc.setup != nil {
				path := tc.setup(t)
				args = []string{path}
			}

			got := mainValidateCompose(args)
			assert.Equal(t, tc.wantCode, got)
		})
	}
}

// TestMainSubcommandRouting tests that Main routes subcommands correctly.
func TestMainSubcommandRouting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		setup    func(t *testing.T) []string
		wantCode int
	}{
		{
			name:     "validate-compose without args",
			args:     []string{"validate-compose"},
			wantCode: 1,
		},
		{
			name:     "validate-config without args",
			args:     []string{"validate-config"},
			wantCode: 1,
		},
		{
			name:     "nonexistent directory for default lint",
			args:     []string{"/nonexistent-dir-for-lint"},
			wantCode: 1,
		},
		{
			name: "generate-listings via Main routing",
			setup: func(t *testing.T) []string {
				t.Helper()

				tmpDir := t.TempDir()
				deploymentsDir := filepath.Join(tmpDir, "deployments")
				configsDir := filepath.Join(tmpDir, "configs")
				require.NoError(t, os.MkdirAll(filepath.Join(deploymentsDir, "jose-ja"), dirPermissions))
				require.NoError(t, os.MkdirAll(configsDir, dirPermissions))
				require.NoError(t, os.WriteFile(
					filepath.Join(deploymentsDir, "jose-ja", "compose.yml"),
					[]byte("name: jose-ja\n"), filePermissions))

				return []string{"generate-listings", deploymentsDir, configsDir}
			},
			wantCode: 0,
		},
		{
			name: "validate-mirror via Main routing",
			setup: func(t *testing.T) []string {
				t.Helper()

				tmpDir := t.TempDir()
				deploymentsDir := filepath.Join(tmpDir, "deployments")
				configsDir := filepath.Join(tmpDir, "configs")
				require.NoError(t, os.MkdirAll(filepath.Join(deploymentsDir, "jose-ja"), dirPermissions))
				require.NoError(t, os.MkdirAll(filepath.Join(configsDir, "jose"), dirPermissions))

				return []string{"validate-mirror", deploymentsDir, configsDir}
			},
			wantCode: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			args := tc.args
			if tc.setup != nil {
				args = tc.setup(t)
			}

			got := Main(args)
			assert.Equal(t, tc.wantCode, got)
		})
	}
}

// TestMainGenerateListings_ConfigsFailure tests failure when configs dir is invalid
// but deployments dir is valid.
func TestMainGenerateListings_ConfigsFailure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	deploymentsDir := filepath.Join(tmpDir, "deployments")
	require.NoError(t, os.MkdirAll(deploymentsDir, dirPermissions))

	got := mainGenerateListings([]string{deploymentsDir, "/nonexistent-configs"})
	assert.Equal(t, 1, got)
}

// TestMainDefaultLintDeployments tests the default lint-deployments path through Main.
func TestMainDefaultLintDeployments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		wantCode int
	}{
		{
			name: "valid deployment structure returns zero",
			setup: func(t *testing.T) string {
				t.Helper()

				tmpDir := t.TempDir()

				return tmpDir
			},
			wantCode: 0,
		},
		{
			name: "nonexistent base dir fails",
			setup: func(t *testing.T) string {
				t.Helper()

				return "/nonexistent-deploy-base"
			},
			wantCode: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			baseDir := tc.setup(t)
			got := Main([]string{baseDir})
			assert.Equal(t, tc.wantCode, got)
		})
	}
}
