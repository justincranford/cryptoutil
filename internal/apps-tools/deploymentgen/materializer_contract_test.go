// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	json "encoding/json"
	"testing"

	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"

	"github.com/stretchr/testify/require"
)

func TestMaterializerProjectionRejectsUnownedInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*pki.CommandInput)
		raw    string
	}{
		{name: "malformed", raw: "not-json"},
		{name: "unowned namespace input", mutate: func(i *pki.CommandInput) { i.Namespace = "unowned" }},
		{name: "API address", mutate: func(i *pki.CommandInput) { i.APIEndpoint = "https://192.0.2.1:443" }},
		{name: "API identity", mutate: func(i *pki.CommandInput) { i.APIServerName = "wrong.example" }},
		{name: "missing pod", mutate: func(i *pki.CommandInput) { i.Pods = nil }},
		{name: "trust escape", mutate: func(i *pki.CommandInput) { i.TrustFiles["public-server"] = "/etc/forged" }},
		{name: testMaterializerProjectionHappy},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			model, scopes := fixture(t)
			inventory, err := build(model, scopes, "s:sm-kms:mem")
			require.NoError(t, err)

			release := inventory.Releases[0]
			input := chartRuntimeInputFixture(inventory, release.Key)
			key := runtimeWorkloadInputKey(release.Key, release.Workloads[0].PSID, release.Workloads[0].Pool)
			value := input.Workloads[key]
			config, err := pki.DecodeCommandInput([]byte(value.MaterializerConfig))
			require.NoError(t, err)

			if test.mutate != nil {
				test.mutate(&config)
			}

			raw, err := json.Marshal(config)
			require.NoError(t, err)

			if test.raw != "" {
				raw = []byte(test.raw)
			}

			value.MaterializerConfig = string(raw)
			input.Workloads[key] = value

			_, err = BuildRuntimeChartValues(model, inventory, release.Key, chartImageFixture(scopes), chartMaterializerFixture(),
				ChartStorage{ClassName: testStorageClassName, Size: testStorageSize}, input)
			if test.name == testMaterializerProjectionHappy {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
