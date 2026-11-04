// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestBuildRuntimeChartValuesIncludesRuntimeInputs(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	inventory = inventoryWithRelease(inventory, testProductIdentityMixedDiskSelector)

	values, err := BuildRuntimeChartValues(
		model, inventory, testProductIdentityMixedDiskSelector, chartImageFixture(scopes), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
		chartRuntimeInputFixture(inventory, testProductIdentityMixedDiskSelector),
	)
	require.NoError(t, err)

	encoded, err := MarshalRuntimeChartValues(values)
	require.NoError(t, err)

	for _, field := range []string{
		`"configurationMaps"`,
		`"frameworkConfigMapName"`,
		`"domainConfigMapName"`,
		`"resources"`,
		`"credentialSecretNames"`,
		`"healthClientSecretName"`,
		`"otlpClientSecretName"`,
	} {
		require.Contains(t, string(encoded), field)
	}
}

func TestBuildRuntimeChartValuesProjectsPerPoolResources(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)

	const selector = "p:identity:mixed-disk"

	inventory, err := build(model, scopes, selector)
	require.NoError(t, err)

	inputs := chartRuntimeInputFixture(inventory, selector)
	pgKey := runtimeWorkloadInputKey(selector, cryptoutilSharedMagic.OTLPServiceIdentityRP, cryptoutilSharedMagic.DeploymentPoolPG)
	diskKey := runtimeWorkloadInputKey(selector, cryptoutilSharedMagic.OTLPServiceIdentityRP, cryptoutilSharedMagic.DeploymentPoolDisk)
	pgInput := inputs.Workloads[pgKey]
	pgInput.Resources.Requests.Memory = "256Mi"
	inputs.Workloads[pgKey] = pgInput

	diskInput := inputs.Workloads[diskKey]
	diskInput.Resources.Requests.Memory = "192Mi"
	inputs.Workloads[diskKey] = diskInput

	values, err := BuildRuntimeChartValues(
		model, inventory, selector, chartImageFixture(scopes), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize}, inputs,
	)
	require.NoError(t, err)
	require.Len(t, values.Workloads, 10)

	projectedMemory := make(map[string]string, len(values.Workloads))
	for _, workload := range values.Workloads {
		if workload.PSID == cryptoutilSharedMagic.OTLPServiceIdentityRP {
			projectedMemory[workload.Pool] = workload.Resources.Requests.Memory
		}
	}

	require.Equal(t, "256Mi", projectedMemory[cryptoutilSharedMagic.DeploymentPoolPG])
	require.Equal(t, "192Mi", projectedMemory[cryptoutilSharedMagic.DeploymentPoolDisk])
}

func TestBuildRuntimeChartValuesProjectsOptionalAPIServerAudience(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	model.Bootstrap.APIServerAudience = "https://issuer.example.test/audience"

	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	inventory = inventoryWithRelease(inventory, testServiceSMKMSPGSelector)

	values, err := BuildRuntimeChartValues(
		model, inventory, testServiceSMKMSPGSelector, chartImageFixture(scopes), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
		chartRuntimeInputFixture(inventory, testServiceSMKMSPGSelector),
	)
	require.NoError(t, err)
	require.Equal(t, model.Bootstrap.APIServerAudience, values.APIServerAudience)
}

func TestBuildRuntimeChartValuesRejectsWhitespaceOnlyAPIServerAudience(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	model.Bootstrap.APIServerAudience = " \t\n "
	inventory = inventoryWithRelease(inventory, testServiceSMKMSPGSelector)

	_, err = BuildRuntimeChartValues(
		model, inventory, testServiceSMKMSPGSelector, chartImageFixture(scopes), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
		chartRuntimeInputFixture(inventory, testServiceSMKMSPGSelector),
	)
	require.ErrorContains(t, err, "API server audience must be omitted or nonempty")
}

func TestBuildRuntimeChartValuesRejectsConflictingConfigurationFiles(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	inventory = inventoryWithRelease(inventory, testProductIdentityMixedDiskSelector)

	inputs := chartRuntimeInputFixture(inventory, testProductIdentityMixedDiskSelector)
	pgKey := runtimeWorkloadInputKey(
		testProductIdentityMixedDiskSelector, cryptoutilSharedMagic.OTLPServiceIdentityRP, cryptoutilSharedMagic.DeploymentPoolPG,
	)
	diskKey := runtimeWorkloadInputKey(
		testProductIdentityMixedDiskSelector, cryptoutilSharedMagic.OTLPServiceIdentityRP, cryptoutilSharedMagic.DeploymentPoolDisk,
	)
	pgInput := inputs.Workloads[pgKey]
	diskInput := inputs.Workloads[diskKey]
	pgInput.FrameworkConfigFile = testSharedConfigFileName
	diskInput.FrameworkConfigFile = testSharedConfigFileName
	pgInput.FrameworkConfig = "framework: pg\n"
	diskInput.FrameworkConfig = "framework: disk\n"
	inputs.Workloads[pgKey] = pgInput
	inputs.Workloads[diskKey] = diskInput

	_, err = BuildRuntimeChartValues(
		model, inventory, testProductIdentityMixedDiskSelector, chartImageFixture(scopes), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize}, inputs,
	)
	require.ErrorContains(t, err, "runtime ConfigMap identity-rp-framework contains conflicting file shared.yml")
}

func TestBuildRuntimeChartValuesRejectsConflictingDomainConfigurationFiles(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	inventory = inventoryWithRelease(inventory, testProductIdentityMixedDiskSelector)

	inputs := chartRuntimeInputFixture(inventory, testProductIdentityMixedDiskSelector)
	pgKey := runtimeWorkloadInputKey(
		testProductIdentityMixedDiskSelector, cryptoutilSharedMagic.OTLPServiceIdentityRP, cryptoutilSharedMagic.DeploymentPoolPG,
	)
	diskKey := runtimeWorkloadInputKey(
		testProductIdentityMixedDiskSelector, cryptoutilSharedMagic.OTLPServiceIdentityRP, cryptoutilSharedMagic.DeploymentPoolDisk,
	)
	pgInput := inputs.Workloads[pgKey]
	diskInput := inputs.Workloads[diskKey]
	pgInput.DomainConfigFile = testSharedConfigFileName
	diskInput.DomainConfigFile = testSharedConfigFileName
	pgInput.DomainConfig = "domain: pg\n"
	diskInput.DomainConfig = "domain: disk\n"
	inputs.Workloads[pgKey] = pgInput
	inputs.Workloads[diskKey] = diskInput

	_, err = BuildRuntimeChartValues(
		model, inventory, testProductIdentityMixedDiskSelector, chartImageFixture(scopes), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize}, inputs,
	)
	require.ErrorContains(t, err, "runtime ConfigMap identity-rp-domain contains conflicting file shared.yml")
}

func TestBuildRuntimeChartValuesProjectsApprovedAPIServiceEdges(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	identityInventory := inventoryWithRelease(inventory, "p:identity:pg")
	require.Len(t, identityInventory.Releases, 1)
	identityRelease := identityInventory.Releases[0]
	require.Contains(t, identityRelease.Objects, Object{
		Kind: testNetworkPolicyKind, Name: testApprovedRPToAuthzEgressPolicyName,
	})
	require.Contains(t, identityRelease.Objects, Object{
		Kind: testNetworkPolicyKind, Name: testApprovedAuthzFromRPIngressPolicyName,
	})

	values, err := BuildRuntimeChartValues(
		model, identityInventory, identityRelease.Key, chartImageFixture(scopes), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
		chartRuntimeInputFixture(identityInventory, identityRelease.Key),
	)
	require.NoError(t, err)

	encoded, err := MarshalRuntimeChartValues(values)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"approvedAPIEdges"`)
	require.Contains(t, string(encoded), `"sourcePsid": "`+cryptoutilSharedMagic.OTLPServiceIdentityRP+`"`)
	require.Contains(t, string(encoded), `"targetPsid": "`+cryptoutilSharedMagic.OTLPServiceIdentityAuthz+`"`)

	serviceInventory := inventoryWithRelease(inventory, "s:identity-rp:pg")
	serviceValues, err := BuildRuntimeChartValues(
		model, serviceInventory, "s:identity-rp:pg", chartImageFixture(scopes), chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
		chartRuntimeInputFixture(serviceInventory, "s:identity-rp:pg"),
	)
	require.NoError(t, err)
	require.Empty(t, serviceValues.ApprovedAPIEdges)
}

func TestBuildRuntimeChartValuesRejectsInvalidRuntimeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		edit      func(RuntimeChartInputs) RuntimeChartInputs
		wantError string
	}{
		{
			name: "missing workload configuration",
			edit: func(inputs RuntimeChartInputs) RuntimeChartInputs {
				delete(
					inputs.Workloads,
					runtimeWorkloadInputKey(testServiceSMKMSPGSelector, cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.DeploymentPoolPG),
				)

				return inputs
			},
			wantError: "runtime workload configuration missing",
		},
		{
			name: "unsafe framework data key",
			edit: func(inputs RuntimeChartInputs) RuntimeChartInputs {
				key := runtimeWorkloadInputKey(
					testServiceSMKMSPGSelector, cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.DeploymentPoolPG,
				)
				input := inputs.Workloads[key]
				input.FrameworkConfigFile = testCredentialsTraversalPath
				inputs.Workloads[key] = input

				return inputs
			},
			wantError: testSafeKubernetesDataKeyError,
		},
		{
			name: "unsafe domain data key",
			edit: func(inputs RuntimeChartInputs) RuntimeChartInputs {
				key := runtimeWorkloadInputKey(
					testServiceSMKMSPGSelector, cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.DeploymentPoolPG,
				)
				input := inputs.Workloads[key]
				input.DomainConfigFile = testCredentialsTraversalPath
				inputs.Workloads[key] = input

				return inputs
			},
			wantError: testSafeKubernetesDataKeyError,
		},
		{
			name: "duplicate generated config argument",
			edit: func(inputs RuntimeChartInputs) RuntimeChartInputs {
				key := runtimeWorkloadInputKey(
					testServiceSMKMSPGSelector, cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.DeploymentPoolPG,
				)
				input := inputs.Workloads[key]
				input.Args = append(input.Args, "--config=/etc/cryptoutil/framework/"+input.FrameworkConfigFile)
				inputs.Workloads[key] = input

				return inputs
			},
			wantError: "must not duplicate generated config arguments",
		},
		{
			name: "resource request exceeds limit",
			edit: func(inputs RuntimeChartInputs) RuntimeChartInputs {
				key := runtimeWorkloadInputKey(
					testServiceSMKMSPGSelector, cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.DeploymentPoolPG,
				)
				input := inputs.Workloads[key]
				input.Resources.Requests.Memory = "1Gi"
				inputs.Workloads[key] = input

				return inputs
			},
			wantError: "resource requests must not exceed limits",
		},
		{
			name: "invalid CPU quantity",
			edit: func(inputs RuntimeChartInputs) RuntimeChartInputs {
				key := runtimeWorkloadInputKey(
					testServiceSMKMSPGSelector, cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.DeploymentPoolPG,
				)
				input := inputs.Workloads[key]
				input.Resources.Requests.CPU = "1000m!"
				inputs.Workloads[key] = input

				return inputs
			},
			wantError: "invalid positive CPU quantity",
		},
		{
			name: "memory quantity overflow",
			edit: func(inputs RuntimeChartInputs) RuntimeChartInputs {
				key := runtimeWorkloadInputKey(
					testServiceSMKMSPGSelector, cryptoutilSharedMagic.OTLPServiceSMKMS, cryptoutilSharedMagic.DeploymentPoolPG,
				)
				input := inputs.Workloads[key]
				input.Resources.Limits.Memory = "999999999999999999999999Ti"
				inputs.Workloads[key] = input

				return inputs
			},
			wantError: "memory quantity",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, scopes := fixture(t)
			inventory, err := build(model, scopes, "")
			require.NoError(t, err)

			inventory = inventoryWithRelease(inventory, testServiceSMKMSPGSelector)

			inputs := test.edit(chartRuntimeInputFixture(inventory, testServiceSMKMSPGSelector))
			_, err = BuildRuntimeChartValues(
				model, inventory, testServiceSMKMSPGSelector, chartImageFixture(scopes), chartMaterializerFixture(),
				ChartStorage{ClassName: testStorageClassName, Size: testStorageSize}, inputs,
			)
			require.ErrorContains(t, err, test.wantError)
		})
	}
}

func TestBuildRuntimeChartValuesRejectsInvalidPKIMaterializerResources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		edit      func(*PKIMaterializerImage)
		wantError string
	}{
		{
			name: "invalid memory quantity",
			edit: func(image *PKIMaterializerImage) {
				image.Resources.Requests.Memory = cryptoutilSharedMagic.DevSetupTestInvalidName
			},
			wantError: "PKI materializer has invalid resources: requests.memory",
		},
		{
			name: "request exceeds limit",
			edit: func(image *PKIMaterializerImage) {
				image.Resources.Requests.Memory = testMemoryQuantity128Mi
			},
			wantError: "PKI materializer has invalid resources: resource requests must not exceed limits",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, scopes := fixture(t)
			inventory, err := build(model, scopes, testServiceSMKMSPGSelector)
			require.NoError(t, err)

			pkiMaterializer := chartMaterializerFixture()
			test.edit(&pkiMaterializer)
			_, err = BuildRuntimeChartValues(
				model, inventory, testServiceSMKMSPGSelector, chartImageFixture(scopes), pkiMaterializer,
				ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
				chartRuntimeInputFixture(inventory, testServiceSMKMSPGSelector),
			)
			require.ErrorContains(t, err, test.wantError)
		})
	}
}
