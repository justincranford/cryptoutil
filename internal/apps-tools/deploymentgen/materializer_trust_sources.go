// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"path"

	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
)

func validateMaterializerTrustSources(runtime RuntimeChartValues, lifecycle LifecycleChartValues) error {
	for _, workload := range runtime.Workloads {
		input, err := pki.DecodeCommandInput([]byte(workload.MaterializerConfig))
		if err != nil {
			return fmt.Errorf("materializer trust input rejected")
		}

		var keys map[string]string

		for _, source := range lifecycle.Secrets {
			if source.Namespace == runtime.Namespace && source.Name == workload.TrustSecretName {
				keys = source.Files

				break
			}
		}

		for _, file := range input.TrustFiles {
			if keys[path.Base(file)] == "" {
				return fmt.Errorf("materializer explicit issuer file is absent from owned trust Secret")
			}
		}
	}

	return nil
}
