// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHelmDeployableLifecycleMatchesCompleteIndependentGraph(t *testing.T) {
	t.Parallel()

	helm := requirePinnedHelm(t)
	tests := []struct {
		name, selector  string
		reserveOrdinals bool
	}{
		{"full registry matrix", "", false},
		{"PostgreSQL selection", testServiceSMKMSPGSelector, false},
		{"SQLite-only selection", testServiceSMKMSMemSelector, false},
		{"minimum one maximum three", testServiceSMKMSPGSelector, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, inventory, input := lifecycleDeployFixture(t, test.selector, func(model *Model) {
				if test.reserveOrdinals {
					model.Replicas = bounds{Min: 1, Max: testLifecycleReservedOrdinals}
					model.Overrides = map[string]bounds{}
				}
			})
			values, err := BuildLifecycleChartValues(model, inventory, input)
			require.NoError(t, err)
			valuesJSON, err := marshalCanonicalJSON(values)
			require.NoError(t, err)

			root := filepath.Join("..", "..", "..")
			lifecyclePath := filepath.Join(root, canonicalChartSourceRoot, "cryptoutil-lifecycle")
			schema, err := os.ReadFile(filepath.Join(lifecyclePath, "values.schema.json"))
			require.NoError(t, err)
			require.NoError(t, validateJSONSchemaDocument(
				"https://cryptoutil.dev/schemas/k8s-lifecycle-values.schema.json",
				"lifecycle values schema", "lifecycle values", schema, valuesJSON,
			))
			directory := t.TempDir()
			valuesPath := filepath.Join(directory, "lifecycle.json")
			require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))
			documents := renderLifecycleDocuments(t, helm, lifecyclePath, valuesPath)

			imported, err := ImportLifecycleSecrets(model, inventory, input, lifecycleSecretFiles(input))
			require.NoError(t, err)

			assertLifecycleImportedReferences(t, documents, imported)

			keys := make([]string, 0)
			for _, secret := range imported {
				keys = append(keys, "Secret/"+secret.Metadata.Namespace+"/"+secret.Metadata.Name)
			}

			for _, document := range documents {
				keys = append(keys, document.Kind+"/"+document.Metadata.Namespace+"/"+document.Metadata.Name)
				require.NotEqual(t, testLifecycleSecretKind, document.Kind, "Helm must never store private material")

				if document.Kind == lifecycleJobKind {
					assertLifecycleJobSecurity(t, document)
				}
			}

			for index, release := range inventory.Releases {
				runtimeValues, err := BuildRuntimeChartValues(
					model, inventoryWithRelease(inventory, release.Key), release.Key, chartImageFixture(RegisteredScopes()),
					chartMaterializerFixture(), ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
					chartRuntimeInputFixture(inventory, release.Key),
				)
				require.NoError(t, err)
				runtimeValues, err = AddLifecycleRuntimeQuota(runtimeValues, values)
				require.NoError(t, err)
				runtimeJSON, err := MarshalRuntimeChartValues(runtimeValues)
				require.NoError(t, err)

				file := filepath.Join(directory, fmt.Sprintf("runtime-%d.json", index))
				require.NoError(t, os.WriteFile(file, runtimeJSON, 0o600))
				runtimeDocuments, err := renderHelmDocuments(
					t, helm, filepath.Join(root, canonicalChartSourceRoot, "cryptoutil-runtime"),
					release.Helm, release.Namespace, file, directory, index+1,
				)
				require.NoError(t, err)
				assertLifecycleImportedReferences(t, runtimeDocuments, imported)

				resourceKeys, err := objectKeysFromRendered(runtimeDocuments, release.Namespace)
				require.NoError(t, err)

				keys = append(keys, resourceKeys...)
				pvcs, err := controllerGeneratedPVCs(runtimeDocuments, release.Namespace)
				require.NoError(t, err)

				keys = append(keys, namespacedObjectKeys(pvcs)...)
			}

			defaultsYAML, err := os.ReadFile(filepath.Join(root, infrastructureChartDefaultsPath))
			require.NoError(t, err)
			defaults, err := DecodeInfrastructureChartValues(defaultsYAML)
			require.NoError(t, err)
			infraValues, err := BuildInfrastructureChartValues(model, inventory, defaults)
			require.NoError(t, err)
			infraValues, err = AddLifecycleInfrastructureQuota(infraValues, values)
			require.NoError(t, err)
			infraJSON, err := MarshalInfrastructureChartValues(infraValues)
			require.NoError(t, err)

			infraFile := filepath.Join(directory, "infra.json")
			require.NoError(t, os.WriteFile(infraFile, infraJSON, 0o600))
			infraDocuments, err := renderHelmDocuments(
				t, helm, filepath.Join(root, canonicalChartSourceRoot, "cryptoutil-infra"),
				infraValues.HelmRelease, inventory.InfraNamespace, infraFile, directory, len(inventory.Releases)+1,
			)
			require.NoError(t, err)
			assertLifecycleImportedReferences(t, infraDocuments, imported)

			infraKeys, err := objectKeysFromRendered(infraDocuments, inventory.InfraNamespace)
			require.NoError(t, err)

			keys = append(keys, infraKeys...)
			pvcs, err := controllerGeneratedPVCs(infraDocuments, inventory.InfraNamespace)
			require.NoError(t, err)

			keys = append(keys, namespacedObjectKeys(pvcs)...)
			slices.Sort(keys)
			require.Equal(t, expectedCombinedInventoryResourceKeys(inventory), keys)
		})
	}
}

func renderLifecycleDocuments(t *testing.T, helm, chart, valuesPath string) []helmRenderedDocument {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, helm, cryptoutilSharedMagic.SkeletonTemplateServiceName, "cu-lifecycle", chart, "--kube-version", "1.37.1", "--values", valuesPath)
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
	decoder := yaml.NewDecoder(bytes.NewReader(output))
	documents := make([]helmRenderedDocument, 0)

	for {
		var document helmRenderedDocument

		err := decoder.Decode(&document)
		if errors.Is(err, io.EOF) {
			break
		}

		require.NoError(t, err)

		if document.Kind == "" {
			continue
		}

		require.NotEmpty(t, document.Metadata.Namespace)
		require.NotEmpty(t, document.Metadata.Name)
		documents = append(documents, document)
	}

	return documents
}

func assertLifecycleJobSecurity(t *testing.T, job helmRenderedDocument) {
	t.Helper()

	pod := job.Spec.Template.Spec
	require.NotNil(t, pod.AutomountServiceAccountToken)
	require.False(t, *pod.AutomountServiceAccountToken)
	require.True(t, pod.SecurityContext.RunAsNonRoot)
	require.Equal(t, 10001, pod.SecurityContext.RunAsUser)
	require.Equal(t, "RuntimeDefault", pod.SecurityContext.SeccompProfile.Type)
	require.Len(t, pod.Containers, 1)
	container := pod.Containers[0]
	require.False(t, *container.SecurityContext.AllowPrivilegeEscalation)
	require.True(t, *container.SecurityContext.ReadOnlyRootFilesystem)
	require.Equal(t, []string{cryptoutilSharedMagic.TestDefaultLogLevelAll}, container.SecurityContext.Capabilities.Drop)

	token := false

	for _, mount := range container.VolumeMounts {
		if mount.Name == testRuntimeBootstrapTokenVolume {
			token = true

			require.True(t, mount.ReadOnly)
		}
	}

	require.Equal(t, strings.HasPrefix(job.Metadata.Name, "pki-init-") ||
		strings.HasPrefix(job.Metadata.Name, "infra-pki-"), token)
	require.Empty(t, job.Metadata.Annotations["helm.sh/hook"])
}

func assertLifecycleImportedReferences(t *testing.T, documents []helmRenderedDocument, imported []LifecycleSecret) {
	t.Helper()

	secrets := make(map[string]LifecycleSecret, len(imported))
	for _, secret := range imported {
		secrets[secret.Metadata.Namespace+"/"+secret.Metadata.Name] = secret
	}

	for _, document := range documents {
		pod := document.Spec.Template.Spec

		check := func(name string) {
			if name == "" {
				return
			}

			_, exists := secrets[document.Metadata.Namespace+"/"+name]
			require.True(t, exists, "rendered %s/%s references missing imported Secret %s", document.Kind, document.Metadata.Name, name)
		}
		for _, volume := range pod.Volumes {
			check(volume.Secret.SecretName)

			for _, source := range volume.Projected.Sources {
				check(source.Secret.Name)
			}
		}

		for _, container := range append(append([]helmRenderedContainer(nil), pod.InitContainers...), pod.Containers...) {
			for _, source := range container.EnvFrom {
				check(source.SecretRef.Name)
			}

			for _, variable := range container.Env {
				reference := variable.ValueFrom.SecretKeyRef
				check(reference.Name)

				if reference.Name != "" {
					secret := secrets[document.Metadata.Namespace+"/"+reference.Name]
					_, exists := secret.Data[reference.Key]
					require.True(t, exists, "rendered Secret key %s/%s is missing", reference.Name, reference.Key)
				}
			}
		}
	}
}
