// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
)

const (
	pkiPublishCommand    = "pki-publish"
	pkiPublicTrustRole   = "public-server"
	pkiAdminTrustRole    = "admin-server"
	pkiPostgresTrustRole = "postgres-client"
)

// PKIPublicationInput contains only explicit identity expectations and references
// to already protected imports, never certificate/key/token/credential contents.
type PKIPublicationInput struct {
	Pods []PKIPublicationPod `json:"pods"`
}

type PKIPublicationPod struct {
	Namespace                string                                 `json:"namespace"`
	PodName                  string                                 `json:"podName"`
	ExpectedAdminDNS         string                                 `json:"expectedAdminDNS"`
	ExpectedPublicDNSNames   []string                               `json:"expectedPublicDNSNames"`
	Trust                    map[string]PKITrustReference           `json:"trust"`
	PostgreSQLClientBindings map[string]pki.PostgreSQLClientBinding `json:"postgresqlClientBindings"`
}

type PKITrustReference struct {
	Secret string `json:"secret"`
	Key    string `json:"key"`
}

func pkiPublishMain(args []string, stdout, stderr io.Writer) int {
	process := lifecycleExecutionProcess(executeLifecycleRequest)
	process.publish = publishPKIRequest

	return lifecycleCommandWithIO(args, stdout, stderr, process, pkiPublishCommand)
}

func decodePKIPublication(raw []byte) (PKIPublicationInput, error) {
	var input PKIPublicationInput
	if len(raw) == 0 || len(raw) > 1024*1024 {
		return input, errors.New("publication input outside size bound")
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	var unique map[string]any
	if err := yaml.Unmarshal(raw, &unique); err != nil {
		return PKIPublicationInput{}, errors.New("duplicate or invalid publication properties")
	}

	if err := decoder.Decode(&input); err != nil {
		return PKIPublicationInput{}, errors.New("invalid publication input")
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return PKIPublicationInput{}, errors.New("trailing publication input")
	}

	return input, nil
}

func publishPKIRequest(ctx context.Context, input gen.BootstrapPreflightInput, secrets []gen.LifecycleSecret,
	path string, config execution.KubectlConfig, policy execution.Policy,
) (execution.Result, error) {
	manifest, err := readPKIPublicationInput(path)
	if err != nil {
		return execution.Result{}, err
	}

	return publishPKIWithRunner(ctx, input, secrets, manifest, config, policy, execution.ExecRunner{})
}

func readPKIPublicationInput(path string) (PKIPublicationInput, error) {
	file, err := os.Open(path)
	if err != nil {
		return PKIPublicationInput{}, errors.New("publication input unavailable")
	}

	raw, readErr := io.ReadAll(io.LimitReader(file, 1024*1024+1))

	closeErr := file.Close()
	if readErr != nil || closeErr != nil {
		return PKIPublicationInput{}, errors.New("publication input read failed")
	}

	manifest, err := decodePKIPublication(raw)
	if err != nil {
		return PKIPublicationInput{}, err
	}

	return manifest, nil
}

func publishPKIWithRunner(ctx context.Context, input gen.BootstrapPreflightInput, secrets []gen.LifecycleSecret,
	manifest PKIPublicationInput, config execution.KubectlConfig, policy execution.Policy, runner execution.Runner,
) (execution.Result, error) {
	if ctx == nil || policy.MaxDuration <= 0 || policy.MaxDuration > 24*time.Hour ||
		policy.OperationTimeout <= 0 || policy.OperationTimeout > time.Hour || policy.OperationTimeout > policy.MaxDuration ||
		input.Lifecycle == nil {
		return execution.Result{}, errors.New("publication requires explicit bounded policy and lifecycle")
	}

	ctx, cancel := context.WithTimeout(ctx, policy.MaxDuration)
	defer cancel()

	if err := ctx.Err(); err != nil {
		return execution.Result{}, fmt.Errorf("publication canceled: %w", err)
	}

	// Validate the entire publication before rendering or contacting Kubernetes.
	prepared, err := buildPKIPublication(input.RuntimeValues, secrets, manifest, time.Now().UTC())
	if err != nil {
		return execution.Result{}, err
	}

	values, err := gen.BuildLifecycleChartValues(input.Model, input.Inventory, *input.Lifecycle)
	if err != nil {
		return execution.Result{}, fmt.Errorf("publication graph: %w", err)
	}

	client, err := execution.NewKubectlClient(config, runner)
	if err != nil {
		return execution.Result{}, fmt.Errorf("publication client: %w", err)
	}

	op, operationCancel := context.WithTimeout(ctx, policy.OperationTimeout)
	err = client.Prepare(op, values)

	operationCancel()

	if err != nil {
		return execution.Result{}, fmt.Errorf("publication prepare: %w", err)
	}

	for _, boundary := range values.Namespaces {
		op, cancel := context.WithTimeout(ctx, policy.OperationTimeout)
		err := client.Require(op, execution.Identity{Kind: "Namespace", Name: boundary.Name, Run: values.Run, Release: boundary.Release})

		cancel()

		if err != nil {
			return execution.Result{}, fmt.Errorf("publication namespace: %w", err)
		}
	}

	for _, secret := range prepared {
		op, cancel := context.WithTimeout(ctx, policy.OperationTimeout)
		err := client.EnsureSecret(op, secret)

		cancel()

		if err != nil {
			return execution.Result{}, fmt.Errorf("publication Secret: %w", err)
		}
	}

	return execution.Result{Secrets: len(prepared)}, nil
}

func buildPKIPublication(values []gen.RuntimeChartValues, imports []gen.LifecycleSecret,
	manifest PKIPublicationInput, now time.Time,
) ([]gen.LifecycleSecret, error) {
	entries := make(map[string]PKIPublicationPod, len(manifest.Pods))
	for _, pod := range manifest.Pods {
		key := pod.Namespace + "/" + pod.PodName
		if _, exists := entries[key]; exists {
			return nil, errors.New("duplicate publication pod")
		}

		entries[key] = pod
	}

	var result []gen.LifecycleSecret

	for _, release := range values {
		for _, workload := range release.Workloads {
			instance := pki.InstanceConfig{
				PSID: workload.PSID, Pool: workload.Pool, Run: release.Run,
				Release: release.HelmRelease, ClusterDomain: release.ClusterDomain, TrustSecret: workload.TrustSecretName,
				FrameworkConfigMap: workload.FrameworkConfigMapName, FrameworkConfigFile: workload.FrameworkConfigFile,
				DomainConfigMap: workload.DomainConfigMapName, DomainConfigFile: workload.DomainConfigFile,
				HealthClientDirectory: pki.HealthClientDirectory, OTLPClientDirectory: pki.OTLPClientDirectory,
				DatabaseSecret: workload.DatabaseSecretName,
			}
			for _, service := range workload.InstanceServices {
				instance.InstanceServices = append(instance.InstanceServices, pki.InstanceService{
					Name: service.Name, PodName: service.PodName,
					TLSLeafSecretName: service.TLSLeafSecretName, HealthClientSecretName: service.HealthClientSecretName,
					OTLPClientSecretName: service.OTLPClientSecretName,
				})
			}

			for _, service := range instance.InstanceServices {
				key := release.Namespace + "/" + service.PodName

				pod, exists := entries[key]
				if !exists {
					return nil, errors.New("publication does not cover every reserved pod")
				}

				delete(entries, key)

				trust, err := publicationTrust(release.Namespace, pod, imports)
				if err != nil {
					return nil, err
				}

				built, err := pki.BuildPublication(instance, pki.RuntimeInputs{
					Namespace: release.Namespace, PodName: service.PodName,
					ExpectedAdminDNS: pod.ExpectedAdminDNS, ExpectedPublicDNSNames: pod.ExpectedPublicDNSNames,
				}, trust, imports, now)
				if err != nil {
					return nil, fmt.Errorf("publication validation: %w", err)
				}

				result = append(result, built...)
			}
		}
	}

	if len(entries) != 0 || len(result) == 0 {
		return nil, errors.New("publication has unowned pods or no runtime graph")
	}

	return result, nil
}

func publicationTrust(namespace string, pod PKIPublicationPod, imports []gen.LifecycleSecret) (pki.TrustInputs, error) {
	trust := pki.TrustInputs{PostgreSQLClientBindings: pod.PostgreSQLClientBindings}
	for role, destination := range map[string]*[]byte{
		pkiPublicTrustRole: &trust.PublicServerPEM, pkiAdminTrustRole: &trust.AdminServerPEM,
		pki.SecretRoleHealthClient: &trust.HealthClientPEM, pki.SecretRoleOTLPClient: &trust.OTLPClientPEM, pkiPostgresTrustRole: &trust.PostgreSQLClientPEM,
	} {
		reference, exists := pod.Trust[role]
		if !exists {
			continue
		}

		for _, secret := range imports {
			if secret.Metadata.Namespace == namespace && secret.Metadata.Name == reference.Secret &&
				secret.Metadata.Labels["cu.io/owner"] == "external-secret-import" {
				*destination = secret.Data[reference.Key]
			}
		}

		if len(*destination) == 0 {
			return pki.TrustInputs{}, errors.New("publication trust reference unavailable")
		}
	}

	for role := range pod.Trust {
		switch role {
		case pkiPublicTrustRole, pkiAdminTrustRole, pki.SecretRoleHealthClient, pki.SecretRoleOTLPClient, pkiPostgresTrustRole:
		default:
			return pki.TrustInputs{}, errors.New("unknown publication trust role")
		}
	}

	return trust, nil
}
