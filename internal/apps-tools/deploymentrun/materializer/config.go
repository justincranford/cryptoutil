// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gopkg.in/yaml.v3"
)

const (
	expectedHealthClientDirectory = HealthClientDirectory
	expectedOTLPClientDirectory   = OTLPClientDirectory
)

var (
	dnsLabelPattern     = regexp.MustCompile(`^[a-z0-9](?:[-a-z0-9]*[a-z0-9])?$`)
	dnsSubdomainPattern = regexp.MustCompile(`^[a-z0-9](?:[-a-z0-9]*[a-z0-9])?(?:\.[a-z0-9](?:[-a-z0-9]*[a-z0-9])?)*$`)
	labelValuePattern   = regexp.MustCompile(`^[A-Za-z0-9](?:[-A-Za-z0-9_.]*[A-Za-z0-9])?$`)
	configFilePattern   = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,253}$`)
)

// InstanceConfig is the strict instance.yaml contract emitted by the runtime chart.
type InstanceConfig struct {
	PSID                   string                  `yaml:"psid"`
	Pool                   string                  `yaml:"pool"`
	Run                    string                  `yaml:"run"`
	Release                string                  `yaml:"release"`
	ClusterDomain          string                  `yaml:"cluster-domain"`
	TrustSecret            string                  `yaml:"trust-secret"`
	FrameworkConfigMap     string                  `yaml:"framework-config-map"`
	FrameworkConfigFile    string                  `yaml:"framework-config-file"`
	DomainConfigMap        string                  `yaml:"domain-config-map"`
	DomainConfigFile       string                  `yaml:"domain-config-file"`
	CredentialSecretMounts []CredentialSecretMount `yaml:"credential-secret-mounts"`
	HealthClientDirectory  string                  `yaml:"health-client-directory"`
	OTLPClientDirectory    string                  `yaml:"otlp-client-directory"`
	InstanceServices       []InstanceService       `yaml:"instance-services"`
	DatabaseSecret         string                  `yaml:"database-secret,omitempty"`
}

// CredentialSecretMount describes an application credential mount; the materializer
// validates its identity but never reads its Secret.
type CredentialSecretMount struct {
	SecretName string `yaml:"secret-name"`
	MountPath  string `yaml:"mount-path"`
}

// InstanceService identifies a single workload and its three materializer inputs.
type InstanceService struct {
	Name                   string `yaml:"name"`
	PodName                string `yaml:"pod-name"`
	TLSLeafSecretName      string `yaml:"tls-leaf-secret-name"`
	HealthClientSecretName string `yaml:"health-client-secret-name"`
	OTLPClientSecretName   string `yaml:"otlp-client-secret-name"`
}

// DecodeInstanceConfig rejects unknown fields, duplicate YAML keys, and trailing documents.
func DecodeInstanceConfig(data []byte) (InstanceConfig, error) {
	var config InstanceConfig
	if len(bytes.TrimSpace(data)) == 0 {
		return config, fmt.Errorf("instance configuration is empty")
	}

	if len(data) > cryptoutilSharedMagic.DeploymentMaterializerMaximumInstanceYAMLBytes {
		return InstanceConfig{}, fmt.Errorf("instance configuration exceeds the size limit")
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&config); err != nil {
		return InstanceConfig{}, fmt.Errorf("decode instance configuration: %w", err)
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err != nil {
			return InstanceConfig{}, fmt.Errorf("decode trailing instance configuration: %w", err)
		}

		return InstanceConfig{}, fmt.Errorf("instance configuration has a trailing YAML document")
	}

	if err := config.validate(); err != nil {
		return InstanceConfig{}, err
	}

	return config, nil
}

// SelectPodService returns only the exact pod entry; it never falls back to another ordinal.
func SelectPodService(config InstanceConfig, podName string) (InstanceService, error) {
	if !validDNSSubdomain(podName) {
		return InstanceService{}, fmt.Errorf("pod name is invalid")
	}

	var match InstanceService

	matches := 0

	for _, service := range config.InstanceServices {
		if service.PodName == podName {
			match = service
			matches++
		}
	}

	if matches != 1 {
		return InstanceService{}, fmt.Errorf("expected exactly one service entry for pod")
	}

	return match, nil
}

func (config InstanceConfig) validate() error {
	if !validDNSLabel(config.PSID) || !validPool(config.Pool) {
		return fmt.Errorf("instance psid or pool identity is invalid")
	}

	if !validDNSLabel(config.PSID + "-" + config.Pool) {
		return fmt.Errorf("derived workload identity is invalid")
	}

	if !validRequiredLabelValue(config.Run) || !validRequiredLabelValue(config.Release) {
		return fmt.Errorf("instance run or release identity is invalid")
	}

	if !validDNSSubdomain(config.ClusterDomain) {
		return fmt.Errorf("instance cluster-domain is invalid")
	}

	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "trust-secret", value: config.TrustSecret},
		{name: "framework-config-map", value: config.FrameworkConfigMap},
		{name: "domain-config-map", value: config.DomainConfigMap},
	} {
		if !validDNSSubdomain(field.value) {
			return fmt.Errorf("instance %s is invalid", field.name)
		}
	}

	if !safeConfigPath(config.FrameworkConfigFile) || !safeConfigPath(config.DomainConfigFile) {
		return fmt.Errorf("instance configuration file path is invalid")
	}

	if config.HealthClientDirectory != expectedHealthClientDirectory || config.OTLPClientDirectory != expectedOTLPClientDirectory {
		return fmt.Errorf("instance client directories do not match the fixed PKI layout")
	}

	if config.DatabaseSecret != "" && !validDNSSubdomain(config.DatabaseSecret) {
		return fmt.Errorf("instance database-secret is invalid")
	}

	if len(config.InstanceServices) == 0 {
		return fmt.Errorf("instance-services must not be empty")
	}

	if err := validateCredentialMounts(config.CredentialSecretMounts); err != nil {
		return err
	}

	pods := make(map[string]struct{}, len(config.InstanceServices))
	secrets := make(map[string]struct{}, len(config.InstanceServices)*3)

	workloadName := config.PSID + "-" + config.Pool
	for _, service := range config.InstanceServices {
		if !validDNSLabel(service.Name) || !validDNSSubdomain(service.PodName) {
			return fmt.Errorf("instance service workload or pod identity is invalid")
		}

		if !validInstanceServiceName(service.Name, workloadName) {
			return fmt.Errorf("instance service name does not match psid and pool identity")
		}

		if _, exists := pods[service.PodName]; exists {
			return fmt.Errorf("instance service pod name is duplicated")
		}

		pods[service.PodName] = struct{}{}

		for _, name := range []string{
			service.TLSLeafSecretName,
			service.HealthClientSecretName,
			service.OTLPClientSecretName,
		} {
			if !validDNSSubdomain(name) {
				return fmt.Errorf("instance service Secret name is invalid")
			}

			if _, exists := secrets[name]; exists {
				return fmt.Errorf("instance service Secret name is duplicated")
			}

			secrets[name] = struct{}{}
		}
	}

	nonPKISecrets := []string{config.TrustSecret, config.DatabaseSecret}
	for _, mount := range config.CredentialSecretMounts {
		nonPKISecrets = append(nonPKISecrets, mount.SecretName)
	}

	for _, name := range nonPKISecrets {
		if name == "" {
			continue
		}

		if _, exists := secrets[name]; exists {
			return fmt.Errorf("materializer Secret reference overlaps a trust, database, or credential Secret")
		}
	}

	return nil
}

func validInstanceServiceName(name string, workloadName string) bool {
	prefix := workloadName + "-i"
	if !strings.HasPrefix(name, prefix) {
		return false
	}

	ordinal := name[len(prefix):]
	if len(ordinal) < 2 {
		return false
	}

	hasNonzeroDigit := false

	for _, character := range ordinal {
		if character < '0' || character > '9' {
			return false
		}

		if character != '0' {
			hasNonzeroDigit = true
		}
	}

	return hasNonzeroDigit
}

func validateCredentialMounts(mounts []CredentialSecretMount) error {
	secretNames := make(map[string]struct{}, len(mounts))

	mountPaths := make(map[string]struct{}, len(mounts))
	for _, mount := range mounts {
		if !validDNSSubdomain(mount.SecretName) || !safeAbsoluteDirectory(mount.MountPath) {
			return fmt.Errorf("credential Secret mount is invalid")
		}

		if strings.HasPrefix(mount.MountPath, "/run/cryptoutil/pki/") || mount.MountPath == "/run/cryptoutil/pki" {
			return fmt.Errorf("credential Secret mount overlaps the PKI directory")
		}

		if _, exists := secretNames[mount.SecretName]; exists {
			return fmt.Errorf("credential Secret mount name is duplicated")
		}

		if _, exists := mountPaths[mount.MountPath]; exists {
			return fmt.Errorf("credential Secret mount path is duplicated")
		}

		secretNames[mount.SecretName] = struct{}{}
		mountPaths[mount.MountPath] = struct{}{}
	}

	return nil
}

func validDNSLabel(value string) bool {
	return len(value) <= 63 && dnsLabelPattern.MatchString(value)
}

func validPool(value string) bool {
	switch value {
	case cryptoutilSharedMagic.DeploymentMaterializerPoolPG, cryptoutilSharedMagic.DeploymentMaterializerPoolMem, cryptoutilSharedMagic.DeploymentMaterializerPoolDisk:
		return true
	default:
		return false
	}
}

func validDNSSubdomain(value string) bool {
	return len(value) <= 253 && dnsSubdomainPattern.MatchString(value)
}

func validRequiredLabelValue(value string) bool {
	return len(value) <= 63 && labelValuePattern.MatchString(value)
}

func safeConfigPath(value string) bool {
	return value != cryptoutilSharedMagic.CurrentDirectory && value != cryptoutilSharedMagic.DeploymentMaterializerPathParent && configFilePattern.MatchString(value)
}

func safeAbsoluteDirectory(value string) bool {
	if !strings.HasPrefix(value, cryptoutilSharedMagic.DeploymentMaterializerPathRoot) || value == cryptoutilSharedMagic.DeploymentMaterializerPathRoot || strings.Contains(value, `\`) ||
		strings.ContainsRune(value, 0) || path.Clean(value) != value {
		return false
	}

	return true
}
