// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestLifecycleMaterializerPublicationKeys(t *testing.T) {
	t.Parallel()

	for _, omitted := range []string{
		"", cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminCertificate,
		cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminPrivateKey, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey,
	} {
		t.Run("omitted-"+omitted, func(t *testing.T) {
			t.Parallel()

			files := make(map[string]string)

			for _, key := range []string{
				cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicPrivateKey,
				cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminCertificate, cryptoutilSharedMagic.DeploymentMaterializerSecretDataAdminPrivateKey,
			} {
				if key != omitted {
					files[key] = key
				}
			}

			err := validateLifecycleSecretSource(LifecycleSecretSource{Name: "sm-kms-mem-0-tls", Type: lifecycleOpaqueSecretType, Files: files})
			if omitted != "" {
				require.ErrorContains(t, err, omitted)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
