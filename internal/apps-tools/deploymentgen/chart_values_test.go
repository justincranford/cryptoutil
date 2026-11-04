// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestBuildRuntimeChartValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		releaseKey    string
		wantReleases  int
		wantWorkloads int
		wantPools     map[string]int
	}{
		{name: testFullRegistryMatrixCaseName, wantReleases: 55, wantWorkloads: 147},
		{
			name: "identity mixed disk product", releaseKey: testProductIdentityMixedDiskSelector,
			wantReleases: 1, wantWorkloads: 10, wantPools: map[string]int{cryptoutilSharedMagic.DeploymentPoolPG: 5, cryptoutilSharedMagic.DeploymentPoolDisk: 5},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, scopes := fixture(t)
			inventory, err := build(model, scopes, "")
			require.NoError(t, err)

			images := chartImageFixture(scopes)
			values := make([]RuntimeChartValues, 0)

			for _, release := range inventory.Releases {
				if test.releaseKey != "" && release.Key != test.releaseKey {
					continue
				}

				releaseInventory := inventoryWithRelease(inventory, release.Key)
				value, buildErr := BuildRuntimeChartValues(
					model, releaseInventory, release.Key, images, chartMaterializerFixture(),
					ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
					chartRuntimeInputFixture(releaseInventory, release.Key),
				)
				require.NoError(t, buildErr)

				values = append(values, value)
			}

			require.Len(t, values, test.wantReleases)

			workloadCount := 0
			poolCounts := map[string]int{}

			for _, value := range values {
				require.Equal(t, model.Run, value.Run)
				require.NotEmpty(t, value.ReleaseKey)
				require.NotEmpty(t, value.Namespace)
				require.NotEmpty(t, value.Workloads)
				encoded, encodeErr := MarshalRuntimeChartValues(value)
				require.NoError(t, encodeErr)
				require.Equal(t, byte('\n'), encoded[len(encoded)-1])

				for _, workload := range value.Workloads {
					workloadCount++
					poolCounts[workload.Pool]++
					require.Equal(t, workload.MaxReplicas, len(workload.InstanceServices))
					require.Equal(t, workload.MaxReplicas, len(workload.TLSLeafSecretNames))
					require.Equal(t, workload.PSID+"-framework", workload.FrameworkConfigMapName)
					require.Equal(t, workload.PSID+"-domain", workload.DomainConfigMapName)
					require.Equal(t, []string{"unseal", "hash-pepper", workload.PSID + "-credentials"}, workload.CredentialSecretNames)
					require.Contains(t, workload.Image.Args, "--backend="+workload.Pool)
					require.Regexp(t, `@sha256:[0-9a-f]{64}$`, workload.Image.Reference)
					require.NotContains(t, workload.Image.Reference, "-----BEGIN")

					for _, instance := range workload.InstanceServices {
						require.Equal(t, instance.PodName+"-health-client", instance.HealthClientSecretName)
						require.Equal(t, instance.PodName+"-otlp-client", instance.OTLPClientSecretName)
						require.Contains(t, workload.TLSLeafSecretNames, instance.TLSLeafSecretName)
					}

					if workload.Pool == cryptoutilSharedMagic.DeploymentPoolPG {
						require.Equal(t, workload.PSID+"-db", workload.DatabaseSecretName)
					} else {
						require.Empty(t, workload.DatabaseSecretName)
					}
				}
			}

			require.Equal(t, test.wantWorkloads, workloadCount)
			require.NotEmpty(t, values[0].ConfigurationMaps)

			for pool, wantCount := range test.wantPools {
				require.Equal(t, wantCount, poolCounts[pool])
			}
		})
	}
}

func TestBuildRuntimeChartValuesErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*Model, *Inventory, map[string]ContainerImage, *PKIMaterializerImage, *ChartStorage) string
	}{
		{
			name: "materializer requires image digest",
			edit: func(_ *Model, _ *Inventory, _ map[string]ContainerImage, image *PKIMaterializerImage, _ *ChartStorage) string {
				image.Reference = "example.invalid/pki:latest"

				return testMaterializerRequirements
			},
		},
		{
			name: "materializer requires command",
			edit: func(_ *Model, _ *Inventory, _ map[string]ContainerImage, image *PKIMaterializerImage, _ *ChartStorage) string {
				image.Command = nil

				return testMaterializerRequirements
			},
		},
		{
			name: "storage class is validated",
			edit: func(_ *Model, _ *Inventory, _ map[string]ContainerImage, _ *PKIMaterializerImage, storage *ChartStorage) string {
				storage.ClassName = "Not.Valid"

				return testInvalidKubernetesValues
			},
		},
		{
			name: "storage size is validated",
			edit: func(_ *Model, _ *Inventory, _ map[string]ContainerImage, _ *PKIMaterializerImage, storage *ChartStorage) string {
				storage.Size = "0Gi"

				return testInvalidKubernetesValues
			},
		},
		{
			name: "model and inventory run must match",
			edit: func(model *Model, _ *Inventory, _ map[string]ContainerImage, _ *PKIMaterializerImage, _ *ChartStorage) string {
				model.Run = testDifferentIdentity

				return testIdentityMismatchError
			},
		},
		{
			name: "model and inventory version must match",
			edit: func(model *Model, _ *Inventory, _ map[string]ContainerImage, _ *PKIMaterializerImage, _ *ChartStorage) string {
				model.Version++

				return testIdentityMismatchError
			},
		},
		{
			name: "invalid Kubernetes API service address",
			edit: func(model *Model, _ *Inventory, _ map[string]ContainerImage, _ *PKIMaterializerImage, _ *ChartStorage) string {
				model.Bootstrap.APIServiceIP = testNotAnIPAddress

				return testNonOverlappingIPv4Error
			},
		},
		{
			name: "inventory must be valid",
			edit: func(_ *Model, inventory *Inventory, _ map[string]ContainerImage, _ *PKIMaterializerImage, _ *ChartStorage) string {
				inventory.Releases = append(inventory.Releases, inventory.Releases[0])

				return testInvalidInventoryCaseName
			},
		},
		{
			name: testMissingRelease,
			edit: func(_ *Model, _ *Inventory, _ map[string]ContainerImage, _ *PKIMaterializerImage, _ *ChartStorage) string {
				return "release not found"
			},
		},
		{
			name: "PS-ID image must exist",
			edit: func(_ *Model, _ *Inventory, images map[string]ContainerImage, _ *PKIMaterializerImage, _ *ChartStorage) string {
				delete(images, cryptoutilSharedMagic.OTLPServiceSMKMS)

				return "image missing for PS-ID"
			},
		},
		{
			name: "app image must include digest",
			edit: func(_ *Model, _ *Inventory, images map[string]ContainerImage, _ *PKIMaterializerImage, _ *ChartStorage) string {
				image := images[cryptoutilSharedMagic.OTLPServiceSMKMS]
				image.Reference = "example.invalid/sm-kms:latest"
				images[cryptoutilSharedMagic.OTLPServiceSMKMS] = image

				return "lowercase sha256 digest"
			},
		},
		{
			name: "app probe commands are required",
			edit: func(_ *Model, _ *Inventory, images map[string]ContainerImage, _ *PKIMaterializerImage, _ *ChartStorage) string {
				image := images[cryptoutilSharedMagic.OTLPServiceSMKMS]
				image.Probes.Startup = nil
				images[cryptoutilSharedMagic.OTLPServiceSMKMS] = image

				return testProbeCommandError
			},
		},
		{
			name: "release image command is required",
			edit: func(_ *Model, _ *Inventory, images map[string]ContainerImage, _ *PKIMaterializerImage, _ *ChartStorage) string {
				image := images[cryptoutilSharedMagic.OTLPServiceSMKMS]
				image.Command = nil
				images[cryptoutilSharedMagic.OTLPServiceSMKMS] = image

				return testProbeCommandError
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, scopes := fixture(t)
			inventory, err := build(model, scopes, "")
			require.NoError(t, err)

			images := chartImageFixture(scopes)
			materializer := chartMaterializerFixture()
			storage := ChartStorage{ClassName: testStorageClassName, Size: testStorageSize}

			releaseKey := testServiceSMKMSPGSelector
			if test.name == testMissingRelease {
				releaseKey = "s:sm-kms:missing"
			}

			want := ""
			if test.edit != nil {
				want = test.edit(&model, &inventory, images, &materializer, &storage)
			}

			_, err = BuildRuntimeChartValues(
				model, inventory, releaseKey, images, materializer, storage,
				chartRuntimeInputFixture(inventory, releaseKey),
			)
			require.ErrorContains(t, err, want)
		})
	}
}

func TestMarshalRuntimeChartValuesDeterministic(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	images := chartImageFixture(scopes)

	values, err := BuildRuntimeChartValues(
		model, inventory, testProductIdentityMixedDiskSelector, images, chartMaterializerFixture(),
		ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
		chartRuntimeInputFixture(inventory, testProductIdentityMixedDiskSelector),
	)
	require.NoError(t, err)

	first, err := MarshalRuntimeChartValues(values)
	require.NoError(t, err)
	second, err := MarshalRuntimeChartValues(values)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Contains(t, string(first), `"name": "identity-authz-disk"`)
	require.NotContains(t, string(first), cryptoutilSharedMagic.CICDTemplateAlpineVersion)
}

func chartImageFixture(scopes []Scope) map[string]ContainerImage {
	images := make(map[string]ContainerImage)

	for _, scope := range scopes {
		if scope.Tier != TierService {
			continue
		}

		for _, psid := range scope.Members {
			images[psid] = ContainerImage{
				Reference: "registry.example.invalid/" + psid + "@sha256:" + strings.Repeat("a", 64),
				Command:   []string{"/app/" + psid, "serve"},
				Args:      []string{"--log-level=info"},
				Drain:     []string{"/app/" + psid, "drain"},
				Probes: ProbeCommands{
					Startup:   []string{"/app/" + psid, "startupz"},
					Liveness:  []string{"/app/" + psid, "livez"},
					Readiness: []string{"/app/" + psid, "readyz"},
				},
			}
		}
	}

	return images
}

func chartRuntimeInputFixture(inventory Inventory, releaseKey string) RuntimeChartInputs {
	inputs := RuntimeChartInputs{Workloads: make(map[string]RuntimeWorkloadInput)}

	for _, release := range inventory.Releases {
		if release.Key != releaseKey {
			continue
		}

		for _, workload := range release.Workloads {
			inputs.Workloads[runtimeWorkloadInputKey(release.Key, workload.PSID, workload.Pool)] = RuntimeWorkloadInput{
				MaterializerConfig:  materializerInputFixture(release, workload),
				FrameworkConfigFile: fmt.Sprintf("%s-%s-framework.yml", workload.PSID, workload.Pool),
				FrameworkConfig:     fmt.Sprintf("# framework config for %s/%s/%s\n", release.Key, workload.PSID, workload.Pool),
				DomainConfigFile:    fmt.Sprintf("%s-%s-domain.yml", workload.PSID, workload.Pool),
				DomainConfig:        fmt.Sprintf("# domain config for %s/%s/%s\n", release.Key, workload.PSID, workload.Pool),
				Args:                []string{"--backend=" + workload.Pool},
				Resources: ResourceRequirements{
					Requests: ResourceSet{CPU: testCPUQuantity100m, Memory: testMemoryQuantity128Mi},
					Limits:   ResourceSet{CPU: "500m", Memory: "512Mi"},
				},
			}
		}
	}

	return inputs
}

func inventoryWithRelease(inventory Inventory, releaseKey string) Inventory {
	selected := inventory
	selected.Releases = make([]Release, 0, 1)

	for _, release := range inventory.Releases {
		if release.Key == releaseKey {
			selected.Releases = append(selected.Releases, release)

			break
		}
	}

	return selected
}

func chartMaterializerFixture() PKIMaterializerImage {
	return PKIMaterializerImage{
		Reference: "registry.example.invalid/pki-materializer@sha256:" + strings.Repeat("b", 64),
		Command:   []string{materializerSuiteExecutable},
		Args:      []string{"pki-materialize", "--instance=/etc/cryptoutil/instance/instance.yaml", "--input=/etc/cryptoutil/instance/materializer.json"},
		Resources: ResourceRequirements{
			Requests: ResourceSet{CPU: "25m", Memory: "32Mi"},
			Limits:   ResourceSet{CPU: testCPUQuantity100m, Memory: "64Mi"},
		},
	}
}
