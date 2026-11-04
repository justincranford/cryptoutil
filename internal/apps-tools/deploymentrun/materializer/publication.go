// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"fmt"
	"maps"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// PublicationSecret is the in-memory Kubernetes Secret transport contract.
type PublicationSecret struct {
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Metadata   PublicationMetadata `json:"metadata"`
	Type       string              `json:"type"`
	Data       map[string][]byte   `json:"data"`
}

type PublicationMetadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}

// BuildPublication validates protected imported PKI bytes with exactly the same
// certificate, key, trust, DNS, EKU and PostgreSQL identity policy as the consumer.
// It only returns independent in-memory Secret objects. It generates no keys,
// changes no files and performs no Kubernetes operations.
func BuildPublication(
	config InstanceConfig, inputs RuntimeInputs, trust TrustInputs,
	imported []PublicationSecret, now time.Time,
) ([]PublicationSecret, error) {
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("publication instance contract: %w", err)
	}

	if !validDNSLabel(inputs.Namespace) || now.IsZero() {
		return nil, fmt.Errorf("publication requires explicit namespace and validation time")
	}

	service, err := SelectPodService(config, inputs.PodName)
	if err != nil {
		return nil, err
	}

	adminDNS, err := expectedAdminServerDNS(config, service, inputs.Namespace)
	if err != nil || adminDNS != inputs.ExpectedAdminDNS {
		return nil, fmt.Errorf("publication administrative identity mismatch")
	}

	if err := validateTrustInputs(config, trust, inputs.PodName); err != nil {
		return nil, err
	}

	bindings := []struct{ name, role string }{
		{service.TLSLeafSecretName, roleTLSBundle},
		{service.HealthClientSecretName, roleHealth},
		{service.OTLPClientSecretName, roleOTLP},
	}

	result := make([]PublicationSecret, 0, len(bindings))
	for _, binding := range bindings {
		var selected *PublicationSecret

		for index := range imported {
			secret := &imported[index]
			if secret.Metadata.Namespace == inputs.Namespace && secret.Metadata.Name == binding.name {
				if selected != nil {
					return nil, fmt.Errorf("publication has duplicate Secret imports")
				}

				selected = secret
			}
		}

		if selected == nil || selected.Metadata.Labels["cu.io/owner"] != "external-secret-import" ||
			selected.Metadata.Labels[labelRun] != config.Run {
			return nil, fmt.Errorf("publication missing owned imported Secret")
		}

		total := 0
		for _, value := range selected.Data {
			if len(value) > cryptoutilSharedMagic.DeploymentMaterializerMaximumCertificatePEMBytes-total {
				return nil, fmt.Errorf("publication Secret exceeds the payload bound")
			}

			total += len(value)
		}

		labels := maps.Clone(selected.Metadata.Labels)
		for key, value := range map[string]string{
			labelWorkload: config.PSID + "-" + config.Pool, labelPSID: config.PSID,
			labelPool: config.Pool, labelRun: config.Run, labelRelease: config.Release, labelRole: binding.role,
		} {
			if existing, exists := labels[key]; exists && existing != value {
				return nil, fmt.Errorf("publication forged identity label")
			}

			labels[key] = value
		}

		api := apiSecret{APIVersion: selected.APIVersion, Kind: selected.Kind, Type: selected.Type, Data: selected.Data}

		api.Metadata.Name, api.Metadata.Namespace, api.Metadata.Labels = binding.name, inputs.Namespace, labels
		if err := validateSecretIdentity(api, inputs.Namespace, binding.name, config, service, binding.role); err != nil {
			return nil, err
		}

		secret := *selected
		secret.Metadata.Labels = labels

		secret.Data = make(map[string][]byte, len(selected.Data))
		for key, value := range selected.Data {
			secret.Data[key] = append([]byte(nil), value...)
		}

		result = append(result, secret)
	}

	var postgres *PostgreSQLClientBinding

	if config.Pool == cryptoutilSharedMagic.DeploymentMaterializerPoolPG {
		binding := trust.PostgreSQLClientBindings[inputs.PodName]
		postgres = &binding
	}

	if _, err := buildMaterializedFiles(config, service, inputs.Namespace, inputs.ExpectedAdminDNS,
		inputs.ExpectedPublicDNSNames, postgres, trust, result[0].Data, result[1].Data, result[2].Data, now); err != nil {
		return nil, fmt.Errorf("publication certificate validation: %w", err)
	}

	return result, nil
}
