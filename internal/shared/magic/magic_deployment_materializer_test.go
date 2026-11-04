// Copyright (c) 2025-2026 Justin Cranford.
package magic

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeploymentMaterializerGenericAliases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		legacy  string
		generic string
	}{
		{
			name:    "Kubernetes Secret kind",
			legacy:  DeploymentMaterializerKubernetesSecretKind,
			generic: KubernetesSecretKind,
		},
		{
			name:    "Kubernetes opaque Secret type",
			legacy:  DeploymentMaterializerSecretTypeOpaque,
			generic: KubernetesSecretTypeOpaque,
		},
		{
			name:    "PS-ID label",
			legacy:  DeploymentMaterializerSecretLabelPSID,
			generic: KubernetesLabelPSID,
		},
		{
			name:    "run label",
			legacy:  DeploymentMaterializerSecretLabelRun,
			generic: KubernetesLabelRun,
		},
		{
			name:    "release label",
			legacy:  DeploymentMaterializerSecretLabelRelease,
			generic: KubernetesLabelRelease,
		},
		{
			name:    "admin private key filename",
			legacy:  DeploymentMaterializerSecretDataAdminPrivateKey,
			generic: AdminPrivateKeyFileName,
		},
		{
			name:    "secrets path segment",
			legacy:  DeploymentMaterializerSecretAPIMiddle,
			generic: DeploymentSecretsDirectoryPathSegment,
		},
		{
			name:    "PostgreSQL pool",
			legacy:  DeploymentMaterializerPoolPG,
			generic: DeploymentPoolPG,
		},
		{
			name:    "memory pool",
			legacy:  DeploymentMaterializerPoolMem,
			generic: DeploymentPoolMem,
		},
		{
			name:    "disk pool",
			legacy:  DeploymentMaterializerPoolDisk,
			generic: DeploymentPoolDisk,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.generic, tc.legacy)
		})
	}
}
