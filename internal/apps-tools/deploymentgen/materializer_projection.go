// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	"fmt"
)

const materializerSuiteExecutable = pki.SuiteExecutable

func validateMaterializerProjection(release RuntimeChartValues, workload RuntimeChartWorkload) error {
	input, err := pki.DecodeCommandInput([]byte(workload.MaterializerConfig))
	if err != nil {
		return fmt.Errorf("materializer workload input: %w", err)
	}

	if input.Namespace != release.Namespace || input.APIEndpoint != "https://"+release.APIServerIP+":443" ||
		input.APIServerName != "kubernetes.default.svc" {
		return fmt.Errorf("materializer endpoint/namespace must match explicit chart target")
	}

	config := pki.InstanceConfig{
		PSID: workload.PSID, Pool: workload.Pool, Run: release.Run,
		Release: release.HelmRelease, ClusterDomain: release.ClusterDomain, TrustSecret: workload.TrustSecretName,
		FrameworkConfigMap: workload.FrameworkConfigMapName, FrameworkConfigFile: workload.FrameworkConfigFile,
		DomainConfigMap: workload.DomainConfigMapName, DomainConfigFile: workload.DomainConfigFile,
		HealthClientDirectory: pki.HealthClientDirectory, OTLPClientDirectory: pki.OTLPClientDirectory,
		DatabaseSecret: workload.DatabaseSecretName,
	}
	for _, service := range workload.InstanceServices {
		config.InstanceServices = append(config.InstanceServices, pki.InstanceService{
			Name: service.Name, PodName: service.PodName,
			TLSLeafSecretName: service.TLSLeafSecretName, HealthClientSecretName: service.HealthClientSecretName, OTLPClientSecretName: service.OTLPClientSecretName,
		})
	}

	if err := pki.ValidateCommandInput(input, config); err != nil {
		return fmt.Errorf("materializer identities/trust contract: %w", err)
	}

	return nil
}
