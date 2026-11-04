// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	"path"
	"testing"

	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func requireRenderedMaterializerContract(t *testing.T, rendered []helmRenderedDocument) {
	t.Helper()

	for _, object := range rendered {
		switch object.Kind {
		case "ConfigMap":
			raw, exists := object.Data["materializer.json"]
			if !exists {
				continue
			}

			config, err := pki.DecodeInstanceConfig([]byte(object.Data["instance.yaml"]))
			require.NoError(t, err)
			input, err := pki.DecodeCommandInput([]byte(raw))
			require.NoError(t, err)
			require.NoError(t, pki.ValidateCommandInput(input, config))
			require.NotContains(t, raw, cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey)
		case "StatefulSet":
			init := object.Spec.Template.Spec.InitContainers[0]
			require.Equal(t, []string{pki.SuiteExecutable}, init.Command)
			require.Equal(t, []string{pki.Command, "--instance=" + pki.InstancePath, "--input=" + pki.InputPath}, init.Args)

			env := map[string]string{}
			for _, value := range init.Env {
				env[value.Name] = value.ValueFrom.FieldRef.FieldPath
			}

			require.Equal(t, "metadata.name", env[pki.PodNameEnvironment])
			require.Equal(t, "metadata.namespace", env[pki.NamespaceEnvironment])

			for _, volume := range object.Spec.Template.Spec.Volumes {
				if volume.Name != testRuntimeBootstrapTokenVolume {
					continue
				}

				require.Len(t, volume.Projected.Sources, 2)
				require.Equal(t, path.Base(cryptoutilSharedMagic.DeploymentMaterializerProjectedTokenPath), volume.Projected.Sources[0].ServiceAccountToken.Path)
				ca := volume.Projected.Sources[1].ConfigMap
				require.Equal(t, "kube-root-ca.crt", ca.Name)
				require.Len(t, ca.Items, 1)
				require.Equal(t, path.Base(cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath), ca.Items[0].Key)
				require.Equal(t, path.Base(cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath), ca.Items[0].Path)
			}

			mounts := map[string]helmRenderedVolumeMount{}
			for _, m := range init.VolumeMounts {
				mounts[m.Name] = m
			}

			require.Equal(t, path.Dir(cryptoutilSharedMagic.DeploymentMaterializerProjectedTokenPath), mounts["bootstrap-api-token"].MountPath)
			require.Equal(t, path.Dir(cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath), mounts["bootstrap-api-token"].MountPath)
			require.True(t, mounts["bootstrap-api-token"].ReadOnly)
			require.Equal(t, pki.TrustDirectory, mounts["trust"].MountPath)
			require.True(t, mounts["trust"].ReadOnly)
		}
	}
}
