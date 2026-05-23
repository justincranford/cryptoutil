// Copyright (c) 2025-2026 Justin Cranford.
package e2e_infra

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestResetCertOutputDir_MakesArtifactsWritable(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	composePath := filepath.Join(tempDir, "compose.yml")

	require.NoError(t, os.WriteFile(composePath, []byte("services: {}"), cryptoutilSharedMagic.CacheFilePermissions))

	certsDir := filepath.Join(tempDir, "certs")
	require.NoError(t, os.MkdirAll(filepath.Join(certsDir, "nested"), cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(certsDir, cryptoutilSharedMagic.PKIInitAdminCABundleFilename), []byte("stale"), 0o400))
	require.NoError(t, os.WriteFile(filepath.Join(certsDir, "tls-config.yml"), []byte("stale"), 0o400))
	require.NoError(t, os.WriteFile(filepath.Join(certsDir, "nested", "leaf.pem"), []byte("stale"), 0o400))
	require.NoError(t, os.WriteFile(filepath.Join(certsDir, ".gitkeep"), []byte{}, cryptoutilSharedMagic.CacheFilePermissions))

	cm := &ComposeManager{ComposeFile: composePath}
	require.NoError(t, cm.resetCertOutputDir())

	_, err := os.Stat(filepath.Join(certsDir, ".gitkeep"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(certsDir, cryptoutilSharedMagic.PKIInitAdminCABundleFilename))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(certsDir, "nested"))
	require.NoError(t, err)
}

func TestResetCertOutputDir_MissingDirectory(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	composePath := filepath.Join(tempDir, "compose.yml")
	require.NoError(t, os.WriteFile(composePath, []byte("services: {}"), cryptoutilSharedMagic.CacheFilePermissions))

	cm := &ComposeManager{ComposeFile: composePath}
	require.NoError(t, cm.resetCertOutputDir())
}

func TestBuildDockerExecArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		compose  string
		profiles []string
		service  string
		command  []string
		want     []string
	}{
		{
			name:    "no-profiles",
			compose: "deployments/sm-kms/compose.yml",
			service: cryptoutilSharedMagic.PKIInitPostgresLeaderService,
			command: []string{"psql", "--username=user", "--command", "SELECT 1"},
			want:    []string{"compose", "-f", "deployments/sm-kms/compose.yml", "exec", cryptoutilSharedMagic.PKIInitPostgresLeaderService, "psql", "--username=user", "--command", "SELECT 1"},
		},
		{
			name:     "with-profile",
			compose:  "deployments/sm-kms/compose.yml",
			profiles: []string{cryptoutilSharedMagic.DockerServicePostgres},
			service:  cryptoutilSharedMagic.KMSE2ESQLiteContainer,
			command:  []string{"/bin/sh"},
			want:     []string{"compose", "-f", "deployments/sm-kms/compose.yml", "--profile", cryptoutilSharedMagic.DockerServicePostgres, "exec", cryptoutilSharedMagic.KMSE2ESQLiteContainer, "/bin/sh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cm := &ComposeManager{ComposeFile: tt.compose, Profiles: tt.profiles}
			got := cm.BuildDockerExecArgs(tt.service, tt.command...)
			require.Equal(t, tt.want, got)
		})
	}
}
