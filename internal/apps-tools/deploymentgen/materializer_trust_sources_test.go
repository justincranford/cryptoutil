// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaterializerTrustFilesExistInOwnedImportedSource(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		edit func(*RuntimeChartValues, *LifecycleChartValues)
		want bool
	}{
		{"present", nil, false},
		{"unparseable trust input", func(r *RuntimeChartValues, _ *LifecycleChartValues) {
			r.Workloads[0].MaterializerConfig = "not-json-trust"
		}, true},
		{"missing owned trust", func(_ *RuntimeChartValues, l *LifecycleChartValues) { l.Secrets = nil }, true},
		{"missing filename", func(r *RuntimeChartValues, l *LifecycleChartValues) {
			for index := range l.Secrets {
				if l.Secrets[index].Name == r.Workloads[0].TrustSecretName {
					delete(l.Secrets[index].Files, cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate)
				}
			}
		}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			model, inventory, input := lifecycleDeployFixture(t, "s:sm-kms:mem")
			lifecycle, err := BuildLifecycleChartValues(model, inventory, input)
			require.NoError(t, err)
			runtime, err := BuildRuntimeChartValues(model, inventory, inventory.Releases[0].Key, chartValuesInputFixture(t).Images, chartMaterializerFixture(),
				ChartStorage{ClassName: testStorageClassName, Size: testStorageSize}, chartRuntimeInputFixture(inventory, inventory.Releases[0].Key))
			require.NoError(t, err)

			if test.edit != nil {
				test.edit(&runtime, &lifecycle)
			}

			err = validateMaterializerTrustSources(runtime, lifecycle)
			if test.want {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
