// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratedRuntimeChartValuesMatchSchemaForFullMatrix(t *testing.T) {
	t.Parallel()

	schemaPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-runtime", "values.schema.json",
	)
	schema, err := os.ReadFile(schemaPath)
	require.NoError(t, err)

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	for _, release := range inventory.Releases {
		t.Run(release.Key, func(t *testing.T) {
			t.Parallel()

			releaseInventory := inventoryWithRelease(inventory, release.Key)
			values, buildErr := BuildRuntimeChartValues(
				model, releaseInventory, release.Key, chartImageFixture(scopes), chartMaterializerFixture(),
				ChartStorage{ClassName: testStorageClassName, Size: testStorageSize},
				chartRuntimeInputFixture(releaseInventory, release.Key),
			)
			require.NoError(t, buildErr)

			valuesJSON, marshalErr := MarshalRuntimeChartValues(values)
			require.NoError(t, marshalErr)
			require.NoError(t, ValidateRuntimeChartValuesSchema(schema, valuesJSON), release.Key)
		})
	}
}
