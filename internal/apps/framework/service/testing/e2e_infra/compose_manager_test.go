// Copyright (c) 2025 Justin Cranford

package e2e_infra

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

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
