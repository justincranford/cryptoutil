// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHelmInfrastructureChartMatchesIndependentInventory(t *testing.T) {
	t.Parallel()

	helmBinary := requirePinnedHelm(t)
	model, scopes := fixture(t)
	defaultsPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.yaml",
	)
	defaultsYAML, err := os.ReadFile(defaultsPath)
	require.NoError(t, err)
	defaults, err := DecodeInfrastructureChartValues(defaultsYAML)
	require.NoError(t, err)

	chartPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-infra",
	)
	outputDirectory := t.TempDir()
	selectors := []struct {
		name     string
		selector string
	}{
		{name: "full registry inventory"},
		{name: "sqlite-only selection", selector: testServiceSMKMSMemSelector},
	}

	for index, test := range selectors {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			inventory, buildErr := build(model, scopes, test.selector)
			require.NoError(t, buildErr)

			values, projectionErr := BuildInfrastructureChartValues(model, inventory, defaults)
			require.NoError(t, projectionErr)

			valuesJSON, marshalErr := MarshalInfrastructureChartValues(values)
			require.NoError(t, marshalErr)

			valuesPath := filepath.Join(outputDirectory, fmt.Sprintf("infra-%02d.json", index))
			require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))

			rendered, renderErr := renderHelmDocuments(
				t, helmBinary, chartPath, values.HelmRelease, values.Namespace, valuesPath, outputDirectory, index,
			)
			require.NoError(t, renderErr)

			expected := expectedInfrastructureChartObjects(inventory)
			renderedKeys, identityErr := objectKeysFromRendered(rendered, inventory.InfraNamespace)
			require.NoError(t, identityErr)
			require.Equal(t, objectKeys(expected, inventory.InfraNamespace), renderedKeys)
			assertRenderedDatabaseCatalog(t, rendered, inventory)
			assertRenderedInfrastructureSecurity(t, rendered, inventory)

			actualPVCs, pvcErr := controllerGeneratedPVCs(rendered, inventory.InfraNamespace)
			require.NoError(t, pvcErr)
			require.Equal(
				t, namespacedObjectKeys(expectedInfrastructureControllerPVCs(inventory)),
				namespacedObjectKeys(actualPVCs),
			)
		})
	}
}

func TestHelmInfrastructureRejectsInconsistentDatabaseCatalog(t *testing.T) {
	t.Parallel()

	helmBinary := requirePinnedHelm(t)
	model, scopes := fixture(t)
	defaultsPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.yaml",
	)
	defaultsYAML, err := os.ReadFile(defaultsPath)
	require.NoError(t, err)
	defaults, err := DecodeInfrastructureChartValues(defaultsYAML)
	require.NoError(t, err)

	chartPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-infra",
	)
	tests := []struct {
		name                string
		selector            string
		editDatabaseCatalog func(*InfrastructureChartValues)
		wantError           string
	}{
		{
			name:                "postgres requires database catalog",
			editDatabaseCatalog: func(values *InfrastructureChartValues) { values.Databases = []Database{} },
			wantError:           "PostgreSQL infrastructure requires a non-empty database catalog",
		},
		{
			name:     "catalog requires postgres",
			selector: testServiceSMKMSMemSelector,
			editDatabaseCatalog: func(values *InfrastructureChartValues) {
				values.Databases = []Database{{
					PSID: cryptoutilSharedMagic.OTLPServiceSMKMS, Name: "sm_kms", Role: "sm_kms_app",
					Publication: "sm_kms_pub", Subscription: "sm_kms_sub",
				}}
			},
			wantError: "database catalog entries require PostgreSQL infrastructure",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			inventory, buildErr := build(model, scopes, test.selector)
			require.NoError(t, buildErr)

			values, projectionErr := BuildInfrastructureChartValues(model, inventory, defaults)
			require.NoError(t, projectionErr)
			test.editDatabaseCatalog(&values)
			valuesJSON, marshalErr := MarshalInfrastructureChartValues(values)
			require.NoError(t, marshalErr)

			valuesPath := filepath.Join(t.TempDir(), "infrastructure-values.json")
			require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))

			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()

			command := exec.CommandContext(
				ctx, helmBinary, cryptoutilSharedMagic.SkeletonTemplateServiceName, values.HelmRelease, chartPath,
				"--namespace", values.Namespace, "--kube-version", "1.37.1", "--values", valuesPath,
			)
			output, templateErr := command.CombinedOutput()
			require.Error(t, templateErr)
			require.Contains(t, string(output), test.wantError)
		})
	}
}

func assertRenderedDatabaseCatalog(t *testing.T, documents []helmRenderedDocument, inventory Inventory) {
	t.Helper()

	expected := make([]Database, 0)
	for _, release := range inventory.Releases {
		expected = append(expected, release.Databases...)
	}

	slices.SortFunc(expected, func(left, right Database) int {
		if nameOrder := strings.Compare(left.Name, right.Name); nameOrder != 0 {
			return nameOrder
		}

		return strings.Compare(left.PSID, right.PSID)
	})

	catalogs := make([]helmRenderedDocument, 0, 1)

	for _, document := range documents {
		if document.Kind == testConfigMapKind && document.Metadata.Name == "database-catalog" {
			catalogs = append(catalogs, document)
		}
	}

	if len(expected) == 0 {
		require.Empty(t, catalogs, "SQLite-only inventory must not render a PostgreSQL database catalog")

		return
	}

	require.Len(t, catalogs, 1)

	var actual []Database
	require.NoError(t, yaml.Unmarshal([]byte(catalogs[0].Data["databases.yaml"]), &actual))
	require.Equal(t, expected, actual)
}
