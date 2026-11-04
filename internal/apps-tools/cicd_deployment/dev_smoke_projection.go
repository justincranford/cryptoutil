// Copyright (c) 2025-2026 Justin Cranford.
package cicd_deployment

import (
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

func developmentPublication(values []gen.RuntimeChartValues) (PKIPublicationInput, error) {
	result := PKIPublicationInput{}

	for _, release := range values {
		for _, workload := range release.Workloads {
			var input pki.CommandInput
			if err := json.Unmarshal([]byte(workload.MaterializerConfig), &input); err != nil {
				return result, errors.New("development publication identities rejected")
			}

			for _, pod := range input.Pods {
				entry := PKIPublicationPod{
					Namespace: release.Namespace, PodName: pod.PodName, ExpectedAdminDNS: pod.ExpectedAdminDNS,
					ExpectedPublicDNSNames: pod.ExpectedPublicDNSNames, Trust: map[string]PKITrustReference{},
				}
				for role := range input.TrustFiles {
					entry.Trust[role] = PKITrustReference{Secret: workload.TrustSecretName, Key: "ca.crt"}
				}

				result.Pods = append(result.Pods, entry)
			}
		}
	}

	return result, nil
}

func waitDevelopment(ctx context.Context, check func() error) error {
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("development wait canceled: %w", err)
		}

		if err := check(); err == nil {
			return nil
		}

		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()

			return fmt.Errorf("development wait canceled: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

func developmentKubeconfig(raw []byte, cluster gen.LocalK3sPlan) ([]byte, error) {
	var config map[string]any
	if err := yaml.Unmarshal(raw, &config); err != nil || config["current-context"] != developmentDefaultContext {
		return nil, errors.New("owned K3s kubeconfig rejected")
	}

	clusters, ok := config["clusters"].([]any)
	if !ok || len(clusters) != 1 {
		return nil, errors.New("owned K3s kubeconfig cluster binding rejected")
	}

	entry, ok := clusters[0].(map[string]any)
	if !ok {
		return nil, errors.New("owned K3s kubeconfig cluster rejected")
	}

	details, ok := entry["cluster"].(map[string]any)
	if !ok || details[cryptoutilSharedMagic.CMD_SERVER] != "https://127.0.0.1:6443" || details["certificate-authority-data"] == nil {
		return nil, errors.New("owned K3s kubeconfig TLS rejected")
	}

	details[cryptoutilSharedMagic.CMD_SERVER] = "https://" + net.JoinHostPort(cluster.APIHost, strconv.Itoa(cluster.APIPort))

	contexts, ok := config["contexts"].([]any)
	if !ok || len(contexts) != 1 {
		return nil, errors.New("owned K3s kubeconfig context binding rejected")
	}

	current, ok := contexts[0].(map[string]any)
	if !ok || current[cryptoutilSharedMagic.ClaimName] != developmentDefaultContext {
		return nil, errors.New("owned K3s kubeconfig context rejected")
	}

	if err := validateDevelopmentKubeconfigIdentity(config, details, current); err != nil {
		return nil, err
	}

	current[cryptoutilSharedMagic.ClaimName], config["current-context"] = cluster.Run, cluster.Run

	encoded, err := yaml.Marshal(config)
	if err != nil {
		return nil, errors.New("protected kubeconfig encoding failed")
	}

	return encoded, nil
}
