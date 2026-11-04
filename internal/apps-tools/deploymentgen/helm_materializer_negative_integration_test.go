// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelmMaterializerRejectsTamperedExplicitInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*testing.T, map[string]any)
	}{
		{"namespace mismatch", func(_ *testing.T, i map[string]any) { i["namespace"] = testOtherNamespace }},
		{"API mismatch", func(_ *testing.T, i map[string]any) { i["apiEndpoint"] = "https://192.0.2.2:443" }},
		{"unknown root", func(_ *testing.T, i map[string]any) { i[cryptoutilSharedMagic.STRATEGY_UNKNOWN] = true }},
		{"trust path escape", func(t *testing.T, i map[string]any) {
			materializerJSONMap(t, i["trustFiles"])["public-server"] = "/run/cryptoutil/trust/.."
		}},
		{"missing trust", func(t *testing.T, i map[string]any) { delete(materializerJSONMap(t, i["trustFiles"]), "admin-server") }},
		{"missing pods", func(_ *testing.T, i map[string]any) { i["pods"] = []any{} }},
		{"unknown pod property", func(t *testing.T, i map[string]any) {
			materializerFirstPod(t, i)[cryptoutilSharedMagic.STRATEGY_UNKNOWN] = true
		}},
		{"unowned pod", func(t *testing.T, i map[string]any) { materializerFirstPod(t, i)["podName"] = "unowned-pod" }},
		{"admin mismatch", func(t *testing.T, i map[string]any) {
			materializerFirstPod(t, i)["expectedAdminDNS"] = "unowned-admin.example"
		}},
		{"public mismatch", func(t *testing.T, i map[string]any) { materializerFirstPod(t, i)["expectedPublicDNSNames"] = []any{} }},
		{"duplicate public", func(t *testing.T, i map[string]any) {
			p := materializerFirstPod(t, i)
			dns, ok := p["expectedPublicDNSNames"].([]any)
			require.True(t, ok)
			require.NotEmpty(t, dns)
			p["expectedPublicDNSNames"] = append(dns, dns[0])
		}},
		{"missing PG bindings", func(t *testing.T, i map[string]any) { delete(materializerFirstPod(t, i), "postgresqlClientBindings") }},
	}

	helm := requirePinnedHelm(t)
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			model, scopes := fixture(t)
			inventory, err := build(model, scopes, "s:sm-kms:pg")
			require.NoError(t, err)

			release := inventory.Releases[0]
			values, err := BuildRuntimeChartValues(model, inventory, release.Key, chartImageFixture(scopes), chartMaterializerFixture(),
				ChartStorage{ClassName: testStorageClassName, Size: testStorageSize}, chartRuntimeInputFixture(inventory, release.Key))
			require.NoError(t, err)

			var input map[string]any
			require.NoError(t, json.Unmarshal([]byte(values.Workloads[0].MaterializerConfig), &input))
			test.mutate(t, input)
			raw, err := json.Marshal(input)
			require.NoError(t, err)

			values.Workloads[0].MaterializerConfig = string(raw)
			raw, err = MarshalRuntimeChartValues(values)
			require.NoError(t, err)
			directory := t.TempDir()
			file := filepath.Join(directory, "explicit-values.json")
			require.NoError(t, os.WriteFile(file, raw, 0o600))
			_, err = renderHelmDocuments(t, helm, filepath.Join("..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes", "cryptoutil-runtime"),
				values.HelmRelease, values.Namespace, file, directory, index)
			require.Error(t, err)
		})
	}
}

func materializerJSONMap(t *testing.T, value any) map[string]any {
	t.Helper()

	result, ok := value.(map[string]any)
	require.True(t, ok)

	return result
}

func materializerFirstPod(t *testing.T, input map[string]any) map[string]any {
	t.Helper()

	pods, ok := input["pods"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, pods)

	return materializerJSONMap(t, pods[0])
}
