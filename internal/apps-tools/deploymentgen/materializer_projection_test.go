// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
)

func materializerInputFixture(release Release, workload Workload) string {
	input := pki.CommandInput{
		Namespace: release.Namespace, APIEndpoint: "https://10.43.0.1:443",
		APIServerName: "kubernetes.default.svc", TrustFiles: map[string]string{},
	}
	for _, role := range []string{"public-server", "admin-server", cryptoutilSharedMagic.DeploymentMaterializerSecretRoleHealth, cryptoutilSharedMagic.DeploymentMaterializerSecretRoleOTLP} {
		input.TrustFiles[role] = pki.TrustDirectory + "/tls.crt"
	}

	bindings := make(map[string]pki.PostgreSQLClientBinding)
	for _, pod := range workload.Pods {
		bindings[pod.Name] = pki.PostgreSQLClientBinding{CommonName: "client-" + pod.Name, DatabaseRole: workload.PSID}
	}

	for _, pod := range workload.Pods {
		podDNS := pod.Name + "." + workload.Name + "-pods." + release.Namespace + ".svc.cluster.local"

		entry := pki.CommandPod{PodName: pod.Name, ExpectedAdminDNS: podDNS, ExpectedPublicDNSNames: []string{
			pod.Service + "." + release.Namespace + ".svc.cluster.local", workload.Name + "." + release.Namespace + ".svc.cluster.local", podDNS,
		}}
		if workload.Pool == cryptoutilSharedMagic.DeploymentPoolPG {
			entry.PostgreSQLClientBindings = bindings
			input.TrustFiles["postgres-client"] = pki.TrustDirectory + "/tls.crt"
		}

		input.Pods = append(input.Pods, entry)
	}

	raw, _ := json.Marshal(input)

	return string(raw)
}
