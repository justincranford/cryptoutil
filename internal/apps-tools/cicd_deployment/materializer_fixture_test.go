// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	gen "cryptoutil/internal/apps-tools/deploymentgen"
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	json "encoding/json"
)

func executionMaterializerFixture(release gen.Release, workload gen.Workload) string {
	input := pki.CommandInput{
		Namespace: release.Namespace, APIEndpoint: "https://10.43.0.1:443",
		APIServerName: "kubernetes.default.svc", TrustFiles: map[string]string{},
	}
	for _, role := range []string{pkiPublicTrustRole, pkiAdminTrustRole, pki.SecretRoleHealthClient, pki.SecretRoleOTLPClient} {
		input.TrustFiles[role] = pki.TrustDirectory + "/tls.crt"
	}

	for _, pod := range workload.Pods {
		dns := pod.Name + "." + workload.Name + "-pods." + release.Namespace + ".svc.cluster.local"
		input.Pods = append(input.Pods, pki.CommandPod{
			PodName: pod.Name, ExpectedAdminDNS: dns,
			ExpectedPublicDNSNames: []string{dns, pod.Service + "." + release.Namespace + ".svc.cluster.local", workload.Name + "." + release.Namespace + ".svc.cluster.local"},
		})
	}

	raw, _ := json.Marshal(input)

	return string(raw)
}
