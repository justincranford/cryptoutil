// Copyright (c) 2025 Justin Cranford

package e2e_infra

import (
	"testing"

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
			service: "postgres-leader",
			command: []string{"psql", "--username=user", "--command", "SELECT 1"},
			want:    []string{"compose", "-f", "deployments/sm-kms/compose.yml", "exec", "postgres-leader", "psql", "--username=user", "--command", "SELECT 1"},
		},
		{
			name:     "with-profile",
			compose:  "deployments/sm-kms/compose.yml",
			profiles: []string{"postgres"},
			service:  "sm-kms-app-sqlite-1",
			command:  []string{"/bin/sh"},
			want:     []string{"compose", "-f", "deployments/sm-kms/compose.yml", "--profile", "postgres", "exec", "sm-kms-app-sqlite-1", "/bin/sh"},
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
