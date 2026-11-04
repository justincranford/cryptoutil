// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Materialize strictly parses instance.yaml, selects exactly the current pod, fetches only
// its TLS/health/OTLP Secret references, validates every leaf, then atomically publishes all files.
func (materializer *Materializer) Materialize(
	ctx context.Context,
	instanceYAML []byte,
	inputs RuntimeInputs,
	trust TrustInputs,
) error {
	if ctx == nil {
		return fmt.Errorf("materializer context is nil")
	}

	boundedContext, cancel := context.WithTimeout(ctx, cryptoutilSharedMagic.DeploymentMaterializerOverallTimeout)
	defer cancel()

	ctx = boundedContext
	if err := contextError(ctx); err != nil {
		return fmt.Errorf("materializer canceled before initialization: %w", err)
	}

	if !validDNSLabel(inputs.Namespace) || !validDNSSubdomain(inputs.PodName) {
		return fmt.Errorf("the Downward API pod identity is invalid")
	}

	if !validDNSSubdomain(inputs.ExpectedAdminDNS) {
		return fmt.Errorf("expected administrative DNS name is invalid")
	}

	if _, err := validateAPIEndpoint(inputs.APIEndpoint, inputs.APIServerName); err != nil {
		return err
	}

	if !filepath.IsAbs(materializer.outputDirectory) || filepath.Clean(materializer.outputDirectory) != materializer.outputDirectory {
		return fmt.Errorf("materializer output directory is not a clean absolute path")
	}

	config, err := DecodeInstanceConfig(instanceYAML)
	if err != nil {
		return err
	}

	service, err := SelectPodService(config, inputs.PodName)
	if err != nil {
		return err
	}

	expectedPublicDNSNames, err := validateExpectedPublicDNSNames(
		config, service, inputs.Namespace, inputs.ExpectedPublicDNSNames,
	)
	if err != nil {
		return err
	}

	expectedAdminDNS, err := expectedAdminServerDNS(config, service, inputs.Namespace)
	if err != nil {
		return err
	}

	if inputs.ExpectedAdminDNS != expectedAdminDNS {
		return fmt.Errorf("expected administrative DNS name differs from the derived pod identity")
	}

	if err := preflightOutputDirectory(materializer.fileSystem, materializer.outputDirectory); err != nil {
		return err
	}

	if err := validateTrustInputs(config, trust, inputs.PodName); err != nil {
		return err
	}

	var postgresBinding *PostgreSQLClientBinding

	if config.Pool == cryptoutilSharedMagic.DeploymentMaterializerPoolPG {
		binding, exists := trust.PostgreSQLClientBindings[inputs.PodName]
		if !exists {
			return fmt.Errorf("trusted PostgreSQL client binding for the current pod is missing")
		}

		postgresBinding = &binding
	}

	apiCAPEM, err := materializer.readFile(cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath)
	if err != nil {
		return fmt.Errorf("read projected Kubernetes API CA bundle: %w", err)
	}

	if len(apiCAPEM) == 0 || len(apiCAPEM) > cryptoutilSharedMagic.DeploymentMaterializerMaximumAPICABundleBytes {
		return fmt.Errorf("projected Kubernetes API CA bundle is empty or oversized")
	}

	apiClient, err := newAPIHTTPClient(apiCAPEM, inputs.APIServerName)
	if err != nil {
		return err
	}
	defer apiClient.CloseIdleConnections()

	tokenBytes, err := materializer.readFile(cryptoutilSharedMagic.DeploymentMaterializerProjectedTokenPath)
	if err != nil {
		return fmt.Errorf("read projected service-account token: %w", err)
	}

	token, err := validateToken(tokenBytes)
	if err != nil {
		return err
	}

	doer := materializer.httpDoer
	if doer == nil {
		doer = apiClient
	}

	tlsBundle, err := fetchSecret(
		ctx, doer, token, inputs.APIEndpoint, inputs.Namespace, service.TLSLeafSecretName, config, service, roleTLSBundle,
	)
	if err != nil {
		return err
	}

	healthSecret, err := fetchSecret(
		ctx, doer, token, inputs.APIEndpoint, inputs.Namespace, service.HealthClientSecretName, config, service, roleHealth,
	)
	if err != nil {
		return err
	}

	otlpSecret, err := fetchSecret(
		ctx, doer, token, inputs.APIEndpoint, inputs.Namespace, service.OTLPClientSecretName, config, service, roleOTLP,
	)
	if err != nil {
		return err
	}

	files, err := buildMaterializedFiles(
		config, service, inputs.Namespace, expectedAdminDNS, expectedPublicDNSNames,
		postgresBinding, trust, tlsBundle, healthSecret, otlpSecret, materializer.now(),
	)
	if err != nil {
		return err
	}

	if err := contextError(ctx); err != nil {
		return fmt.Errorf("materializer canceled before publication: %w", err)
	}

	return publishFiles(ctx, materializer.fileSystem, materializer.outputDirectory, files)
}

func validateTrustInputs(
	config InstanceConfig,
	trust TrustInputs,
	podName string,
) error {
	anchors := []struct {
		name string
		pem  []byte
	}{
		{name: rolePublic, pem: trust.PublicServerPEM},
		{name: roleAdmin, pem: trust.AdminServerPEM},
		{name: cryptoutilSharedMagic.DeploymentMaterializerSecretRoleHealth, pem: trust.HealthClientPEM},
		{name: cryptoutilSharedMagic.DeploymentMaterializerSecretRoleOTLP, pem: trust.OTLPClientPEM},
	}
	if config.Pool == cryptoutilSharedMagic.DeploymentMaterializerPoolPG {
		anchors = append(anchors, struct {
			name string
			pem  []byte
		}{name: rolePostgres, pem: trust.PostgreSQLClientPEM})
	}

	for _, anchor := range anchors {
		if _, err := parseTrustRoots(anchor.pem); err != nil {
			return fmt.Errorf("trusted CA anchors for %s are invalid: %w", anchor.name, err)
		}
	}

	if config.Pool != cryptoutilSharedMagic.DeploymentMaterializerPoolPG {
		return nil
	}

	if err := validatePostgreSQLClientBindings(trust.PostgreSQLClientBindings, config.InstanceServices); err != nil {
		return err
	}

	if _, exists := trust.PostgreSQLClientBindings[podName]; !exists {
		return fmt.Errorf("trusted PostgreSQL client binding for the current pod is missing")
	}

	return nil
}

func validatePostgreSQLClientBindings(
	bindings map[string]PostgreSQLClientBinding,
	services []InstanceService,
) error {
	if len(bindings) == 0 {
		return fmt.Errorf("trusted PostgreSQL client bindings are empty")
	}

	configuredPods := make(map[string]struct{}, len(services))
	for _, service := range services {
		configuredPods[service.PodName] = struct{}{}
	}

	if len(bindings) != len(configuredPods) {
		return fmt.Errorf("trusted PostgreSQL client bindings do not exactly cover configured pods")
	}

	commonNames := make(map[string]string, len(bindings))
	for podName, binding := range bindings {
		if !validDNSSubdomain(podName) || !validBindingText(binding.CommonName) ||
			!validBindingText(binding.DatabaseRole) {
			return fmt.Errorf("trusted PostgreSQL client binding is invalid")
		}

		if _, exists := configuredPods[podName]; !exists {
			return fmt.Errorf("trusted PostgreSQL client bindings do not exactly cover configured pods")
		}

		if _, exists := commonNames[binding.CommonName]; exists {
			return fmt.Errorf("trusted PostgreSQL client common names are not unique per pod")
		}

		commonNames[binding.CommonName] = podName
	}

	return nil
}

func validBindingText(value string) bool {
	if value == "" || len(value) > 63 || strings.TrimSpace(value) != value {
		return false
	}

	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}

	return true
}

func validateExpectedPublicDNSNames(
	config InstanceConfig,
	service InstanceService,
	namespace string,
	expected []string,
) ([]string, error) {
	if len(expected) > cryptoutilSharedMagic.DeploymentMaterializerMaximumExpectedPublicDNSNames {
		return nil, fmt.Errorf("expected public DNS identities exceed the size limit")
	}

	workloadDNS := service.Name + "." + namespace + ".svc." + config.ClusterDomain
	governingHeadlessService := config.PSID + "-" + config.Pool + "-pods"
	podDNS := service.PodName + "." + governingHeadlessService + "." + namespace + ".svc." + config.ClusterDomain
	required := []string{workloadDNS, podDNS}

	seen := make(map[string]struct{}, len(expected))
	for _, name := range expected {
		if !validDNSSubdomain(name) {
			return nil, fmt.Errorf("expected public DNS identity is invalid")
		}

		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("expected public DNS identity is duplicated")
		}

		seen[name] = struct{}{}
	}

	for _, name := range required {
		if !validDNSSubdomain(name) {
			return nil, fmt.Errorf("derived public service or pod FQDN is invalid")
		}

		if _, exists := seen[name]; !exists {
			return nil, fmt.Errorf("expected public DNS identities omit a derived service or pod FQDN")
		}
	}

	return append([]string(nil), expected...), nil
}

func expectedAdminServerDNS(config InstanceConfig, service InstanceService, namespace string) (string, error) {
	expected := service.PodName + "." + config.PSID + "-" + config.Pool + "-pods." +
		namespace + ".svc." + config.ClusterDomain
	if !validDNSSubdomain(expected) {
		return "", fmt.Errorf("derived administrative pod DNS identity is invalid")
	}

	return expected, nil
}
