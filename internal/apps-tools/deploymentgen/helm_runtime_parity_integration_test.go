// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestHelmRuntimeChartMatchesIndependentInventoryForFullMatrix(t *testing.T) {
	t.Parallel()

	helmBinary := requirePinnedHelm(t)
	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	images := chartImageFixture(scopes)
	chartPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-runtime",
	)
	outputDirectory := t.TempDir()

	for index, release := range inventory.Releases {
		t.Run(release.Key, func(t *testing.T) {
			t.Parallel()

			releaseInventory := inventoryWithRelease(inventory, release.Key)
			values, projectionErr := BuildRuntimeChartValues(
				model, releaseInventory, release.Key, images, chartMaterializerFixture(),
				ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
				chartRuntimeInputFixture(releaseInventory, release.Key),
			)
			require.NoError(t, projectionErr)

			valuesJSON, marshalErr := MarshalRuntimeChartValues(values)
			require.NoError(t, marshalErr)

			valuesPath := filepath.Join(outputDirectory, fmt.Sprintf("runtime-%02d.json", index))
			require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))

			rendered, renderErr := renderHelmDocuments(
				t, helmBinary, chartPath, values.HelmRelease, values.Namespace, valuesPath, outputDirectory, index,
			)
			require.NoError(t, renderErr)
			require.Empty(t, runtimeRenderedSecurityViolations(rendered))
			requireRenderedMaterializerContract(t, rendered)

			expected := expectedRuntimeChartObjects(release)
			renderedKeys, identityErr := objectKeysFromRendered(rendered, release.Namespace)
			require.NoError(t, identityErr)
			require.Equal(t, objectKeys(expected, release.Namespace), renderedKeys)

			actualPVCs, pvcErr := controllerGeneratedPVCs(rendered, release.Namespace)
			require.NoError(t, pvcErr)
			require.Equal(
				t, namespacedObjectKeys(expectedRuntimeControllerPVCs(release)),
				namespacedObjectKeys(actualPVCs),
			)
		})
	}
}

func TestHelmRuntimeChartRendersRegistryAPIEdgeSelectors(t *testing.T) {
	t.Parallel()

	helmBinary := requirePinnedHelm(t)
	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	const releaseKey = "p:identity:pg"

	releaseInventory := inventoryWithRelease(inventory, releaseKey)
	values, err := BuildRuntimeChartValues(
		model, releaseInventory, releaseKey, chartImageFixture(scopes), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
		chartRuntimeInputFixture(releaseInventory, releaseKey),
	)
	require.NoError(t, err)
	valuesJSON, err := MarshalRuntimeChartValues(values)
	require.NoError(t, err)

	outputDirectory := t.TempDir()
	valuesPath := filepath.Join(outputDirectory, "identity-api-edge.json")
	require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))

	chartPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-runtime",
	)
	rendered, err := renderHelmDocuments(
		t, helmBinary, chartPath, values.HelmRelease, values.Namespace, valuesPath, outputDirectory, 0,
	)
	require.NoError(t, err)

	var (
		egressPolicy  *helmRenderedDocument
		ingressPolicy *helmRenderedDocument
	)

	for index := range rendered {
		switch rendered[index].Metadata.Name {
		case testApprovedRPToAuthzEgressPolicyName:
			egressPolicy = &rendered[index]
		case testApprovedAuthzFromRPIngressPolicyName:
			ingressPolicy = &rendered[index]
		}
	}

	require.NotNil(t, egressPolicy)
	require.Equal(t, map[string]string{
		cryptoutilSharedMagic.KubernetesLabelRun: model.Run, cryptoutilSharedMagic.KubernetesLabelRelease: values.HelmRelease,
		cryptoutilSharedMagic.KubernetesLabelPSID: cryptoutilSharedMagic.OTLPServiceIdentityRP,
	}, egressPolicy.Spec.PodSelector.MatchLabels)
	require.Len(t, egressPolicy.Spec.Egress, 1)
	require.Len(t, egressPolicy.Spec.Egress[0].To, 1)
	require.Equal(t, map[string]string{
		cryptoutilSharedMagic.KubernetesLabelRun: model.Run, cryptoutilSharedMagic.KubernetesLabelRelease: values.HelmRelease,
		cryptoutilSharedMagic.KubernetesLabelPSID: cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
	}, egressPolicy.Spec.Egress[0].To[0].PodSelector.MatchLabels)
	require.Equal(t, []helmNetworkPolicyPort{{Protocol: testTCPProtocol, Port: 8080}}, egressPolicy.Spec.Egress[0].Ports)

	require.NotNil(t, ingressPolicy)
	require.Equal(t, map[string]string{
		cryptoutilSharedMagic.KubernetesLabelRun: model.Run, cryptoutilSharedMagic.KubernetesLabelRelease: values.HelmRelease,
		cryptoutilSharedMagic.KubernetesLabelPSID: cryptoutilSharedMagic.OTLPServiceIdentityAuthz,
	}, ingressPolicy.Spec.PodSelector.MatchLabels)
	require.Len(t, ingressPolicy.Spec.Ingress, 1)
	require.Len(t, ingressPolicy.Spec.Ingress[0].From, 1)
	require.Equal(t, map[string]string{
		cryptoutilSharedMagic.KubernetesLabelRun: model.Run, cryptoutilSharedMagic.KubernetesLabelRelease: values.HelmRelease,
		cryptoutilSharedMagic.KubernetesLabelPSID: cryptoutilSharedMagic.OTLPServiceIdentityRP,
	}, ingressPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels)
	require.Equal(t, []helmNetworkPolicyPort{{Protocol: testTCPProtocol, Port: 8080}}, ingressPolicy.Spec.Ingress[0].Ports)
}

func TestHelmRuntimeChartReservesMaximumOrdinalSecretsButOnlyRunsMinimumPods(t *testing.T) {
	t.Parallel()

	helmBinary := requirePinnedHelm(t)
	model, scopes := fixture(t)
	model.Overrides["sm-kms/disk"] = Bounds{Min: 1, Max: testLifecycleReservedOrdinals}

	const selector = "s:sm-kms:disk"

	inventory, err := build(model, scopes, selector)
	require.NoError(t, err)
	require.Len(t, inventory.Releases, 1)

	release := inventory.Releases[0]
	workload := release.Workloads[0]
	require.Equal(t, 3, len(workload.Pods))
	require.True(t, workload.Pods[0].Active)
	require.False(t, workload.Pods[1].Active)
	require.False(t, workload.Pods[2].Active)

	activePVCs := 0

	for _, pod := range workload.Pods {
		if pod.Active && pod.PVC != "" {
			activePVCs++
		}
	}

	require.Equal(t, workload.Bounds.Min, activePVCs)

	values, err := BuildRuntimeChartValues(
		model, inventory, selector, chartImageFixture(scopes), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
		chartRuntimeInputFixture(inventory, selector),
	)
	require.NoError(t, err)
	valuesJSON, err := MarshalRuntimeChartValues(values)
	require.NoError(t, err)

	outputDirectory := t.TempDir()
	valuesPath := filepath.Join(outputDirectory, "minimum-replicas.json")
	require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))

	chartPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-runtime",
	)
	rendered, err := renderHelmDocuments(
		t, helmBinary, chartPath, values.HelmRelease, values.Namespace, valuesPath, outputDirectory, 0,
	)
	require.NoError(t, err)

	expected := expectedRuntimeChartObjects(release)
	renderedKeys, identityErr := objectKeysFromRendered(rendered, release.Namespace)
	require.NoError(t, identityErr)
	require.Equal(t, objectKeys(expected, release.Namespace), renderedKeys)

	actualPVCs, pvcErr := controllerGeneratedPVCs(rendered, release.Namespace)
	require.NoError(t, pvcErr)
	require.Equal(
		t, namespacedObjectKeys(expectedRuntimeControllerPVCs(release)), namespacedObjectKeys(actualPVCs),
	)

	var (
		statefulSet     *helmRenderedDocument
		certificateRole *helmRenderedDocument
	)

	for index := range rendered {
		switch {
		case rendered[index].Kind == testStatefulSetKind && rendered[index].Metadata.Name == workload.Name:
			statefulSet = &rendered[index]
		case rendered[index].Kind == "Role" && rendered[index].Metadata.Name == workload.Name+"-cert-reader":
			certificateRole = &rendered[index]
		}
	}

	require.NotNil(t, statefulSet)
	require.Equal(t, workload.Bounds.Min, statefulSet.Spec.Replicas)
	require.Len(t, statefulSet.Spec.VolumeClaimTemplates, 1)
	require.Equal(t, "data", statefulSet.Spec.VolumeClaimTemplates[0].Metadata.Name)

	for _, document := range rendered {
		require.NotEqual(t, testPersistentVolumeClaimKind, document.Kind)
	}

	require.NotNil(t, certificateRole)
	require.Len(t, certificateRole.Rules, 1)
	require.Equal(t, []string{"secrets"}, certificateRole.Rules[0].Resources)
	require.Equal(t, []string{"get"}, certificateRole.Rules[0].Verbs)
	require.Len(t, certificateRole.Rules[0].ResourceNames, 9)

	for _, pod := range workload.Pods {
		require.Contains(t, certificateRole.Rules[0].ResourceNames, pod.TLS)
		require.Contains(t, certificateRole.Rules[0].ResourceNames, pod.Name+"-health-client")
		require.Contains(t, certificateRole.Rules[0].ResourceNames, pod.Name+"-otlp-client")
	}
}
