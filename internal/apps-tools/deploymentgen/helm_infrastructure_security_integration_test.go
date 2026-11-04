// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	checkovPostgresHighUIDException = "CKV_K8S_40=The pinned PostgreSQL image's initdb requires its declared UID 999; UID 10001 fails user lookup."
	testInfraCollectorName          = "collector"
	testInfraLGTMName               = "lgtm"
)

func assertRenderedInfrastructureSecurity(t *testing.T, documents []helmRenderedDocument, inventory Inventory) {
	t.Helper()

	expectedUIDs := []struct {
		name     string
		runAs    int
		runGroup int
		fsGroup  int
		postgres bool
	}{
		{name: cryptoutilSharedMagic.KMSPostgresLeaderContainer, runAs: 999, runGroup: 999, fsGroup: 999, postgres: true},
		{name: cryptoutilSharedMagic.PKIInitPostgresFollowerService, runAs: 999, runGroup: 999, fsGroup: 999, postgres: true},
		{name: testInfraCollectorName, runAs: 10001, runGroup: 10001, fsGroup: 10001},
		{name: testInfraLGTMName, runAs: 10001, runGroup: 472, fsGroup: 472},
	}

	expectedNames := make([]string, 0, len(expectedUIDs))
	for _, expected := range expectedUIDs {
		if inventoryContainsObject(inventory.Infra, "StatefulSet", expected.name) {
			expectedNames = append(expectedNames, expected.name)
		}
	}

	slices.Sort(expectedNames)

	actualNames := make([]string, 0, len(expectedNames))
	actualByName := make(map[string]helmRenderedDocument, len(expectedNames))

	for _, document := range documents {
		if document.Kind != testStatefulSetKind {
			continue
		}

		actualNames = append(actualNames, document.Metadata.Name)
		actualByName[document.Metadata.Name] = document
	}

	slices.Sort(actualNames)
	require.Equal(t, expectedNames, actualNames)

	for _, expected := range expectedUIDs {
		if !slices.Contains(expectedNames, expected.name) {
			continue
		}

		t.Run(expected.name, func(t *testing.T) {
			t.Parallel()

			document := actualByName[expected.name]
			podSecurity := document.Spec.Template.Spec.SecurityContext
			require.True(t, podSecurity.RunAsNonRoot)
			require.Equal(t, expected.runAs, podSecurity.RunAsUser)
			require.Equal(t, expected.runGroup, podSecurity.RunAsGroup)
			require.Equal(t, expected.fsGroup, podSecurity.FSGroup)
			require.Len(t, document.Spec.Template.Spec.Containers, 1)

			container := document.Spec.Template.Spec.Containers[0]
			require.Equal(t, expected.name, container.Name)
			require.True(t, container.SecurityContext.RunAsNonRoot)
			require.Equal(t, expected.runAs, container.SecurityContext.RunAsUser)
			require.Equal(t, expected.runGroup, container.SecurityContext.RunAsGroup)
			require.NotNil(t, container.SecurityContext.AllowPrivilegeEscalation)
			require.False(t, *container.SecurityContext.AllowPrivilegeEscalation)
			require.NotNil(t, container.SecurityContext.ReadOnlyRootFilesystem)
			require.True(t, *container.SecurityContext.ReadOnlyRootFilesystem)
			require.Contains(t, container.SecurityContext.Capabilities.Drop, cryptoutilSharedMagic.TestDefaultLogLevelAll)

			if expected.postgres {
				require.Equal(
					t,
					checkovPostgresHighUIDException,
					document.Metadata.Annotations["checkov.io/skip1"],
				)

				return
			}

			require.NotContains(t, document.Metadata.Annotations, "checkov.io/skip1")
		})
	}
}

func TestHelmInfrastructureRequiresHighUIDForCollectorAndLGTM(t *testing.T) {
	t.Parallel()

	helmBinary := requirePinnedHelm(t)
	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	defaultsPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.yaml",
	)
	defaultsYAML, err := os.ReadFile(defaultsPath)
	require.NoError(t, err)
	defaults, err := DecodeInfrastructureChartValues(defaultsYAML)
	require.NoError(t, err)

	tests := []struct {
		name string
		edit func(*InfrastructureChartValues)
	}{
		{
			name: testInfraCollectorName,
			edit: func(values *InfrastructureChartValues) {
				values.Components.Collector.RunAsUser = 999
			},
		},
		{
			name: testInfraLGTMName,
			edit: func(values *InfrastructureChartValues) {
				values.Components.LGTM.RunAsUser = 472
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			values, projectionErr := BuildInfrastructureChartValues(model, inventory, defaults)
			require.NoError(t, projectionErr)
			test.edit(&values)

			valuesJSON, marshalErr := MarshalInfrastructureChartValues(values)
			require.NoError(t, marshalErr)

			valuesPath := filepath.Join(t.TempDir(), "infrastructure-values.json")
			require.NoError(t, os.WriteFile(valuesPath, valuesJSON, 0o600))

			chartPath := filepath.Join(
				"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
				"cryptoutil-infra",
			)

			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)

			defer cancel()

			command := exec.CommandContext(
				ctx, helmBinary, cryptoutilSharedMagic.SkeletonTemplateServiceName,
				"invalid-security-"+test.name, chartPath,
				"--namespace", inventory.InfraNamespace, "--kube-version", "1.37.1", "--values", valuesPath,
			)
			output, templateErr := command.CombinedOutput()
			require.Error(t, templateErr)
			require.Contains(t, string(output), "minimum: got")
			require.Contains(t, string(output), "want 10,000")
			require.Contains(t, string(output), "runAsUser")
		})
	}
}
