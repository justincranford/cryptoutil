// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestHelmCombinedLifecycleIdentityParityForPGAndSQLite(t *testing.T) {
	t.Parallel()

	helmBinary := requirePinnedHelm(t)
	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	selectedReleases := make([]Release, 0, 2)
	chartKeys := make([]string, 0)
	outputDirectory := t.TempDir()
	runtimeChartPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-runtime",
	)
	selectors := []string{testServiceSMKMSPGSelector, testServiceSMKMSMemSelector}

	for index, selector := range selectors {
		releaseInventory, buildErr := build(model, scopes, selector)
		require.NoError(t, buildErr)
		require.Len(t, releaseInventory.Releases, 1)

		release := releaseInventory.Releases[0]
		selectedReleases = append(selectedReleases, release)

		values, projectionErr := BuildRuntimeChartValues(
			model,
			releaseInventory,
			selector,
			chartImageFixture(scopes),
			chartMaterializerFixture(),
			ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
			chartRuntimeInputFixture(releaseInventory, selector),
		)
		require.NoError(t, projectionErr)

		valuesJSON, marshalErr := MarshalRuntimeChartValues(values)
		require.NoError(t, marshalErr)

		valuesPath := filepath.Join(outputDirectory, fmt.Sprintf("runtime-%02d.json", index))
		require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))

		rendered, renderErr := renderHelmDocuments(
			t,
			helmBinary,
			runtimeChartPath,
			values.HelmRelease,
			values.Namespace,
			valuesPath,
			outputDirectory,
			index,
		)
		require.NoError(t, renderErr)

		renderedKeys, identityErr := objectKeysFromRendered(rendered, release.Namespace)
		require.NoError(t, identityErr)

		chartKeys = append(chartKeys, renderedKeys...)

		controllerPVCs, pvcErr := controllerGeneratedPVCs(rendered, release.Namespace)
		require.NoError(t, pvcErr)

		chartKeys = append(chartKeys, namespacedObjectKeys(controllerPVCs)...)
	}

	infrastructureDefaults, err := os.ReadFile(filepath.Join(
		"..", "..", "..", filepath.FromSlash(infrastructureChartDefaultsPath),
	))
	require.NoError(t, err)

	defaults, err := DecodeInfrastructureChartValues(infrastructureDefaults)
	require.NoError(t, err)

	infrastructureValues, err := BuildInfrastructureChartValues(model, inventory, defaults)
	require.NoError(t, err)

	infrastructureJSON, err := MarshalInfrastructureChartValues(infrastructureValues)
	require.NoError(t, err)

	infrastructureValuesPath := filepath.Join(outputDirectory, "infrastructure.json")
	require.NoError(t, os.WriteFile(infrastructureValuesPath, infrastructureJSON, 0o600))

	infrastructureChartPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-infra",
	)
	infrastructureDocuments, err := renderHelmDocuments(
		t,
		helmBinary,
		infrastructureChartPath,
		infrastructureValues.HelmRelease,
		infrastructureValues.Namespace,
		infrastructureValuesPath,
		outputDirectory,
		len(selectors),
	)
	require.NoError(t, err)

	infrastructureKeys, err := objectKeysFromRendered(infrastructureDocuments, inventory.InfraNamespace)
	require.NoError(t, err)

	chartKeys = append(chartKeys, infrastructureKeys...)

	infrastructurePVCs, err := controllerGeneratedPVCs(infrastructureDocuments, inventory.InfraNamespace)
	require.NoError(t, err)

	chartKeys = append(chartKeys, namespacedObjectKeys(infrastructurePVCs)...)

	combinedInventory := inventory
	combinedInventory.Releases = selectedReleases

	lifecycle, err := BuildLifecycleInventory(combinedInventory)
	require.NoError(t, err)

	expectedGraphKeys := expectedCombinedInventoryResourceKeys(combinedInventory)
	expectedLifecycle := expectedLifecycleInventoryIdentities(combinedInventory)
	require.Equal(
		t,
		expectedLifecycle,
		lifecycle.Identities,
		"generated lifecycle identities and owner partitions must exactly match the independent graph",
	)

	expectedLifecycleKeys := lifecycleIdentityKeys(expectedLifecycle)
	expectedChartKeys := subtractLifecycleParityKeys(expectedGraphKeys, expectedLifecycleKeys)
	chartKeys = sortedLifecycleParityKeys(chartKeys)
	require.Equal(
		t,
		expectedChartKeys,
		chartKeys,
		"canonical runtime and infrastructure charts plus controller PVCs must render exactly the graph-owned resources",
	)

	combinedKeys := append(append([]string(nil), chartKeys...), lifecycleIdentityKeys(lifecycle.Identities)...)
	require.Equal(
		t,
		expectedGraphKeys,
		sortedLifecycleParityKeys(combinedKeys),
		"chart-rendered resources, controller PVC identities, and lifecycle identities must cover the selected graph exactly",
	)
}

func expectedLifecycleInventoryIdentities(inventory Inventory) []LifecycleIdentity {
	identities := make([]LifecycleIdentity, 0)
	appendObject := func(namespace string, object Object) {
		owner := ""

		switch object.Kind {
		case testLifecycleSecretKind:
			owner = "external-secret-identity"
		case lifecycleJobKind:
			owner = "lifecycle-job-identity"
		case testLifecycleServiceAccountKind, testLifecycleRoleKind, testLifecycleRoleBindingKind:
			if object.Name == testLifecycleBootstrapResourceName {
				owner = "bootstrap-rbac-identity"
			}
		}

		if owner == "" {
			return
		}

		identities = append(identities, LifecycleIdentity{
			Namespace: namespace,
			Kind:      object.Kind,
			Name:      object.Name,
			Owner:     owner,
		})
	}

	for _, object := range inventory.Infra {
		appendObject(inventory.InfraNamespace, object)
	}

	for _, release := range inventory.Releases {
		for _, object := range release.Objects {
			appendObject(release.Namespace, object)
		}

		for _, workload := range release.Workloads {
			for _, object := range workload.Objects {
				appendObject(release.Namespace, object)
			}

			for _, pod := range workload.Pods {
				appendObject(release.Namespace, Object{Kind: cryptoutilSharedMagic.KubernetesSecretKind, Name: pod.TLS})
			}
		}
	}

	slices.SortFunc(identities, func(left, right LifecycleIdentity) int {
		if order := strings.Compare(left.Namespace, right.Namespace); order != 0 {
			return order
		}

		if order := strings.Compare(left.Kind, right.Kind); order != 0 {
			return order
		}

		if order := strings.Compare(left.Name, right.Name); order != 0 {
			return order
		}

		return strings.Compare(left.Owner, right.Owner)
	})

	return identities
}

func lifecycleIdentityKeys(identities []LifecycleIdentity) []string {
	objects := make([]namespacedObject, len(identities))
	for index, identity := range identities {
		objects[index] = namespacedObject{
			Kind: identity.Kind, Namespace: identity.Namespace, Name: identity.Name,
		}
	}

	return namespacedObjectKeys(objects)
}

func expectedCombinedInventoryResourceKeys(inventory Inventory) []string {
	objects := make([]namespacedObject, 0)
	appendObject := func(namespace string, object Object) {
		if object.Kind == "Pod" {
			return
		}

		objects = append(objects, namespacedObject{
			Kind: object.Kind, Namespace: namespace, Name: object.Name,
		})
	}

	for _, object := range inventory.Infra {
		appendObject(inventory.InfraNamespace, object)
	}

	for _, release := range inventory.Releases {
		for _, object := range release.Objects {
			appendObject(release.Namespace, object)
		}

		for _, workload := range release.Workloads {
			for _, object := range workload.Objects {
				appendObject(release.Namespace, object)
			}

			for _, pod := range workload.Pods {
				appendObject(release.Namespace, Object{Kind: testServiceKind, Name: pod.Service})
				appendObject(release.Namespace, Object{
					Kind: cryptoutilSharedMagic.KubernetesSecretKind, Name: pod.TLS,
				})

				if pod.Active && pod.PVC != "" {
					appendObject(release.Namespace, Object{
						Kind: testPersistentVolumeClaimKind, Name: pod.PVC,
					})
				}
			}
		}
	}

	return namespacedObjectKeys(objects)
}

func subtractLifecycleParityKeys(keys, excluded []string) []string {
	excludedCounts := make(map[string]int, len(excluded))
	for _, key := range excluded {
		excludedCounts[key]++
	}

	result := make([]string, 0, len(keys)-len(excluded))
	for _, key := range keys {
		if excludedCounts[key] > 0 {
			excludedCounts[key]--

			continue
		}

		result = append(result, key)
	}

	return result
}

func sortedLifecycleParityKeys(keys []string) []string {
	sorted := append([]string(nil), keys...)
	slices.Sort(sorted)

	return sorted
}
