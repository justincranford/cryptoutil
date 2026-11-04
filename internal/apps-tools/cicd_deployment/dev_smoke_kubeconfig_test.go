// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"encoding/base64"
	json "encoding/json"
	"strconv"
	"testing"

	gen "cryptoutil/internal/apps-tools/deploymentgen"

	"github.com/stretchr/testify/require"
)

func developmentKubeconfigIdentityFixture(t *testing.T) []byte {
	t.Helper()
	_, material, _ := publicationFixture(t)
	raw, err := json.Marshal(map[string]any{
		developmentKubeCurrent: developmentDefaultContext,
		developmentKubeClusters: []any{map[string]any{developmentName: developmentDefaultContext, testDevelopmentClusterField: map[string]any{
			developmentServerField: developmentDefaultServer, developmentKubeCA: base64.StdEncoding.EncodeToString(material[len(material)-1].Data[testPKICAKey]),
		}}},
		developmentKubeContexts: []any{map[string]any{developmentName: developmentDefaultContext, developmentKubeContext: map[string]any{
			testDevelopmentClusterField: developmentDefaultContext, developmentKubeUser: developmentDefaultContext,
		}}},
		developmentKubeUsers: []any{map[string]any{developmentName: developmentDefaultContext, developmentKubeUser: map[string]any{
			developmentKubeCertificate: base64.StdEncoding.EncodeToString(material[1].Data[testPKITLSCertKey]),
			developmentKubeKey:         base64.StdEncoding.EncodeToString(material[1].Data[testPKITLSPrivateKey]),
		}}},
	})
	require.NoError(t, err)

	return raw
}

func TestDevelopmentKubeconfigRejectsForgedTLSAndAmbientClient(t *testing.T) {
	t.Parallel()

	for mode := range 13 {
		t.Run(strconv.Itoa(mode), func(t *testing.T) {
			t.Parallel()

			var config map[string]any
			require.NoError(t, json.Unmarshal(developmentKubeconfigIdentityFixture(t), &config))
			clusters, ok := config[developmentKubeClusters].([]any)
			require.True(t, ok)
			clusterEntry, ok := clusters[0].(map[string]any)
			require.True(t, ok)
			cluster, ok := clusterEntry[testDevelopmentClusterField].(map[string]any)
			require.True(t, ok)
			users, ok := config[developmentKubeUsers].([]any)
			require.True(t, ok)
			entry, ok := users[0].(map[string]any)
			require.True(t, ok)
			user, ok := entry[developmentKubeUser].(map[string]any)
			require.True(t, ok)

			switch mode {
			case 0:
				cluster["insecure-skip-tls-verify"] = true
			case 1:
				cluster[developmentKubeCA] = ""
			case 2:
				cluster[developmentKubeCA] = "not-base64!"
			case testDevThirdCase:
				cluster[developmentKubeCA] = base64.StdEncoding.EncodeToString([]byte("not-PEM"))
			case 4:
				config[developmentKubeUsers] = []any{}
			case 5:
				entry[developmentName] = "another-user"
			case 6:
				user["exec"] = map[string]any{developmentCommandField: "not-authorized"}
			case 7:
				user[developmentKubeKey] = base64.StdEncoding.EncodeToString([]byte("forged-key"))
			case 8:
				user[developmentKubeCertificate] = 42
			case 9:
				contexts, ok := config[developmentKubeContexts].([]any)
				require.True(t, ok)
				contextEntry, ok := contexts[0].(map[string]any)
				require.True(t, ok)

				contextEntry[developmentKubeContext] = map[string]any{testDevelopmentClusterField: "another-cluster", developmentKubeUser: developmentDefaultContext}
			case 10:
				clusters[0] = "not-a-cluster-entry"
			case 11:
				contexts, ok := config[developmentKubeContexts].([]any)
				require.True(t, ok)

				contexts[0] = "not-a-context-entry"
			case 12:
				clusterEntry[testDevelopmentClusterField] = "not-a-cluster-object"
			}

			raw, err := json.Marshal(config)
			require.NoError(t, err)
			_, err = developmentKubeconfig(raw, gen.LocalK3sPlan{Run: gen.DevelopmentSmokeRun, APIHost: developmentLoopback, APIPort: 6445})
			require.Error(t, err)
		})
	}
}
