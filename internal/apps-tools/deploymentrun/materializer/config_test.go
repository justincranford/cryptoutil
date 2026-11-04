// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const validInstanceYAML = `psid: identity-authz
pool: pg
run: run-17
release: release-17
cluster-domain: cluster.local
trust-secret: trust-17
framework-config-map: framework-17
framework-config-file: framework.yaml
domain-config-map: domain-17
domain-config-file: domain.yaml
credential-secret-mounts:
  - secret-name: credential-17
    mount-path: /run/cryptoutil/credentials
health-client-directory: /run/cryptoutil/pki/health-client
otlp-client-directory: /run/cryptoutil/pki/otlp-client
instance-services:
  - name: identity-authz-pg-i01
    pod-name: identity-authz-pg-0
    tls-leaf-secret-name: identity-authz-pg-tls-0
    health-client-secret-name: identity-authz-pg-health-0
    otlp-client-secret-name: identity-authz-pg-otlp-0
  - name: identity-authz-pg-i02
    pod-name: identity-authz-pg-1
    tls-leaf-secret-name: identity-authz-pg-tls-1
    health-client-secret-name: identity-authz-pg-health-1
    otlp-client-secret-name: identity-authz-pg-otlp-1
database-secret: identity-authz-database
`

func TestDecodeInstanceConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{name: "valid strict chart contract", input: validInstanceYAML},
		{name: "optional database Secret omitted", input: strings.Replace(validInstanceYAML, "database-secret: identity-authz-database\n", "", 1)},
		{name: "empty document", input: " \n", wantErr: cryptoutilSharedMagic.EMPTY},
		{name: "oversized instance YAML", input: strings.Repeat("x", cryptoutilSharedMagic.DeploymentMaterializerMaximumInstanceYAMLBytes+1), wantErr: testSizeLimitError},
		{name: "unknown top-level field", input: validInstanceYAML + "future-field: forbidden\n", wantErr: "field future-field not found"},
		{name: "duplicate YAML key", input: strings.Replace(validInstanceYAML, "pool: pg\n", "pool: pg\npool: pg\n", 1), wantErr: "already defined"},
		{name: testTrailingYAMLDocumentError, input: validInstanceYAML + "---\npsid: other\n", wantErr: testTrailingYAMLDocumentError},
		{name: "malformed trailing YAML document", input: validInstanceYAML + "---\ninvalid: [\n", wantErr: "decode trailing instance configuration"},
		{name: "duplicate pod identity", input: strings.Replace(validInstanceYAML, testPostgreSQLPodNameOne, "identity-authz-pg-0", 1), wantErr: "pod name is duplicated"},
		{name: "unsafe config file path", input: strings.Replace(validInstanceYAML, "framework.yaml", "../framework.yaml", 1), wantErr: testFilePathInvalidError},
		{name: "client directory override", input: strings.Replace(validInstanceYAML, expectedHealthClientDirectory, "/tmp/health", 1), wantErr: testFixedPKILayoutError},
		{name: "overlapping credential mount", input: strings.Replace(validInstanceYAML, "/run/cryptoutil/credentials", cryptoutilSharedMagic.DeploymentMaterializerHealthClientDirectory, 1), wantErr: testOverlapsPKIDirectoryError},
		{name: "duplicate secret reference", input: strings.Replace(validInstanceYAML, "identity-authz-pg-tls-1", "identity-authz-pg-tls-0", 1), wantErr: testDuplicateSecretNameError},
		{name: "TLS Secret cannot alias trust Secret", input: strings.Replace(validInstanceYAML, "identity-authz-pg-tls-0", "trust-17", 1), wantErr: testOverlapsSensitiveSecretError},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			config, err := DecodeInstanceConfig([]byte(testCase.input))
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, cryptoutilSharedMagic.OTLPServiceIdentityAuthz, config.PSID)
			require.Len(t, config.InstanceServices, 2)
		})
	}
}

func TestGeneratedInstanceYAMLPools(t *testing.T) {
	t.Parallel()

	const generatedInstanceYAML = `psid: sm-kms
pool: {{POOL}}
run: run-17
release: release-17
cluster-domain: cluster.local
trust-secret: trust-17
framework-config-map: framework-sm-kms
framework-config-file: framework.yaml
domain-config-map: domain-sm-kms
domain-config-file: domain.yaml
credential-secret-mounts: []
health-client-directory: /run/cryptoutil/pki/health-client
otlp-client-directory: /run/cryptoutil/pki/otlp-client
instance-services:
  - name: sm-kms-{{POOL}}-i01
    pod-name: sm-kms-{{POOL}}-0
    tls-leaf-secret-name: sm-kms-{{POOL}}-tls-0
    health-client-secret-name: sm-kms-{{POOL}}-health-0
    otlp-client-secret-name: sm-kms-{{POOL}}-otlp-0
database-secret: sm-kms-database
`

	for _, pool := range []string{cryptoutilSharedMagic.DeploymentMaterializerPoolPG, cryptoutilSharedMagic.DeploymentMaterializerPoolMem, cryptoutilSharedMagic.DeploymentMaterializerPoolDisk} {
		t.Run(pool, func(t *testing.T) {
			t.Parallel()

			input := strings.ReplaceAll(generatedInstanceYAML, "{{POOL}}", pool)
			config, err := DecodeInstanceConfig([]byte(input))
			require.NoError(t, err)
			require.Equal(t, pool, config.Pool)

			service, err := SelectPodService(config, "sm-kms-"+pool+"-0")
			require.NoError(t, err)
			require.Equal(t, "sm-kms-"+pool+"-i01", service.Name)
			require.Equal(t, "sm-kms-"+pool+"-0", service.PodName)

			namespace := "runtime-17"
			serviceDNS := service.Name + "." + namespace + ".svc." + config.ClusterDomain
			workloadDNS := "sm-kms-" + pool + "." + namespace + ".svc." + config.ClusterDomain
			realPodDNS := service.PodName + ".sm-kms-" + pool + "-pods." + namespace + ".svc." + config.ClusterDomain
			oldOrdinalDerivedPodDNS := service.PodName + "." + service.Name + "." + namespace + ".svc." + config.ClusterDomain
			externalSNI := testExternalLoginSNI
			expected := []string{serviceDNS, realPodDNS, workloadDNS, externalSNI}

			names, err := validateExpectedPublicDNSNames(config, service, namespace, expected)
			require.NoError(t, err)
			require.Equal(t, expected, names)
			require.NotContains(t, names, oldOrdinalDerivedPodDNS)

			expected[1] = oldOrdinalDerivedPodDNS
			_, err = validateExpectedPublicDNSNames(config, service, namespace, expected)
			require.ErrorContains(t, err, testMissingDerivedServiceOrPodFQDNError)
		})
	}
}

func TestValidInstanceServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		want        bool
	}{
		{name: "canonical first ordinal", serviceName: "identity-authz-pg-i01", want: true},
		{name: "multi-digit ordinal", serviceName: "identity-authz-pg-i100", want: true},
		{name: "wrong workload prefix", serviceName: testOtherWorkloadServiceName},
		{name: "missing ordinal", serviceName: "identity-authz-pg"},
		{name: "short ordinal", serviceName: "identity-authz-pg-i1"},
		{name: "nondigit ordinal", serviceName: "identity-authz-pg-i0a"},
		{name: "zero ordinal", serviceName: "identity-authz-pg-i00"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.want, validInstanceServiceName(testCase.serviceName, "identity-authz-pg"))
		})
	}
}

func TestSelectPodService(t *testing.T) {
	t.Parallel()

	config, err := DecodeInstanceConfig([]byte(validInstanceYAML))
	require.NoError(t, err)

	tests := []struct {
		name    string
		podName string
		want    string
		wantErr string
	}{
		{name: "exact selected pod", podName: testPostgreSQLPodNameOne, want: "identity-authz-pg-health-1"},
		{name: "no ordinal fallback", podName: testUnconfiguredPostgreSQLPodName, wantErr: "exactly one"},
		{name: "invalid pod name", podName: "../identity-authz-pg-0", wantErr: "pod name is invalid"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			service, selectErr := SelectPodService(config, testCase.podName)
			if testCase.wantErr != "" {
				require.ErrorContains(t, selectErr, testCase.wantErr)

				return
			}

			require.NoError(t, selectErr)
			require.Equal(t, testCase.want, service.HealthClientSecretName)
		})
	}

	t.Run("duplicate direct input is rejected", func(t *testing.T) {
		t.Parallel()

		duplicate := config
		duplicate.InstanceServices = append(append([]InstanceService(nil), config.InstanceServices...), config.InstanceServices[1])
		_, selectErr := SelectPodService(duplicate, testPostgreSQLPodNameOne)
		require.ErrorContains(t, selectErr, "exactly one")
	})
}

func TestInstanceConfigValidation(t *testing.T) {
	t.Parallel()

	base, err := DecodeInstanceConfig([]byte(validInstanceYAML))
	require.NoError(t, err)

	tests := []struct {
		name    string
		mutate  func(*InstanceConfig)
		wantErr string
	}{
		{name: "invalid PSID", mutate: func(config *InstanceConfig) { config.PSID = "Identity Authz" }, wantErr: testPSIDOrPoolIdentityError},
		{name: "invalid pool", mutate: func(config *InstanceConfig) { config.Pool = "pool/name" }, wantErr: testPSIDOrPoolIdentityError},
		{name: "instance service name must match workload identity", mutate: func(config *InstanceConfig) { config.InstanceServices[0].Name = testOtherWorkloadServiceName }, wantErr: "instance service name does not match"},
		{name: "derived workload identity exceeds DNS label limit", mutate: func(config *InstanceConfig) { config.PSID = strings.Repeat("a", 61) }, wantErr: "derived workload identity"},
		{name: "legacy PostgreSQL pool alias", mutate: func(config *InstanceConfig) { config.Pool = "pgpool" }, wantErr: testPSIDOrPoolIdentityError},
		{name: "legacy memory pool alias", mutate: func(config *InstanceConfig) { config.Pool = "mempool" }, wantErr: testPSIDOrPoolIdentityError},
		{name: "legacy disk pool alias", mutate: func(config *InstanceConfig) { config.Pool = "diskpool" }, wantErr: testPSIDOrPoolIdentityError},
		{name: "invalid run label", mutate: func(config *InstanceConfig) { config.Run = "" }, wantErr: testRunOrReleaseIdentityError},
		{name: "invalid release label", mutate: func(config *InstanceConfig) { config.Release = "bad/value" }, wantErr: testRunOrReleaseIdentityError},
		{name: "invalid cluster domain", mutate: func(config *InstanceConfig) { config.ClusterDomain = "cluster..local" }, wantErr: "cluster-domain"},
		{name: "invalid trust Secret", mutate: func(config *InstanceConfig) { config.TrustSecret = "../trust" }, wantErr: "trust-secret"},
		{name: "invalid framework map", mutate: func(config *InstanceConfig) { config.FrameworkConfigMap = "" }, wantErr: "framework-config-map"},
		{name: "invalid domain map", mutate: func(config *InstanceConfig) { config.DomainConfigMap = "../domain" }, wantErr: "domain-config-map"},
		{name: "invalid framework config path", mutate: func(config *InstanceConfig) { config.FrameworkConfigFile = "" }, wantErr: testFilePathInvalidError},
		{name: "config file path cannot be filesystem root", mutate: func(config *InstanceConfig) {
			config.FrameworkConfigFile = cryptoutilSharedMagic.DeploymentMaterializerPathRoot
		}, wantErr: testFilePathInvalidError},
		{name: "absolute config filename is rejected", mutate: func(config *InstanceConfig) { config.FrameworkConfigFile = testAbsoluteFrameworkConfigFile }, wantErr: testFilePathInvalidError},
		{name: "nested config filename is rejected", mutate: func(config *InstanceConfig) { config.DomainConfigFile = "nested/domain.yaml" }, wantErr: testFilePathInvalidError},
		{name: "traversal config filename is rejected", mutate: func(config *InstanceConfig) { config.FrameworkConfigFile = "../framework.yaml" }, wantErr: testFilePathInvalidError},
		{name: "NUL config filename is rejected", mutate: func(config *InstanceConfig) { config.FrameworkConfigFile = testFrameworkConfigFilenameWithNUL }, wantErr: testFilePathInvalidError},
		{name: "dot config filename is rejected", mutate: func(config *InstanceConfig) { config.DomainConfigFile = testDotPath }, wantErr: testFilePathInvalidError},
		{name: "invalid domain config path", mutate: func(config *InstanceConfig) { config.DomainConfigFile = "domain\\config.yaml" }, wantErr: testFilePathInvalidError},
		{name: "invalid health directory", mutate: func(config *InstanceConfig) { config.HealthClientDirectory = "/tmp/health" }, wantErr: testFixedPKILayoutError},
		{name: "invalid OTLP directory", mutate: func(config *InstanceConfig) { config.OTLPClientDirectory = "" }, wantErr: testFixedPKILayoutError},
		{name: "invalid optional database Secret", mutate: func(config *InstanceConfig) { config.DatabaseSecret = testInvalidSecretName }, wantErr: "database-secret"},
		{name: "no workloads", mutate: func(config *InstanceConfig) { config.InstanceServices = nil }, wantErr: "must not be empty"},
		{
			name: "invalid credential reference",
			mutate: func(config *InstanceConfig) {
				config.CredentialSecretMounts[0].SecretName = testParentSecretName
			},
			wantErr: testInvalidCredentialMountError,
		},
		{
			name: "invalid credential mount path",
			mutate: func(config *InstanceConfig) {
				config.CredentialSecretMounts[0].MountPath = "/run/../secret"
			},
			wantErr: testInvalidCredentialMountError,
		},
		{
			name: "credential mount NUL is rejected",
			mutate: func(config *InstanceConfig) {
				config.CredentialSecretMounts[0].MountPath = "/run/credentials\x00"
			},
			wantErr: testInvalidCredentialMountError,
		},
		{
			name: "credential mount cannot target filesystem root",
			mutate: func(config *InstanceConfig) {
				config.CredentialSecretMounts[0].MountPath = "/"
			},
			wantErr: testInvalidCredentialMountError,
		},
		{
			name: "duplicate credential Secret",
			mutate: func(config *InstanceConfig) {
				config.CredentialSecretMounts = append(config.CredentialSecretMounts, config.CredentialSecretMounts[0])
			},
			wantErr: "Secret mount name is duplicated",
		},
		{
			name: "duplicate credential path",
			mutate: func(config *InstanceConfig) {
				config.CredentialSecretMounts = append(config.CredentialSecretMounts, CredentialSecretMount{
					SecretName: "credential-18", MountPath: config.CredentialSecretMounts[0].MountPath,
				})
			},
			wantErr: "Secret mount path is duplicated",
		},
		{
			name: "credential mount overlaps PKI",
			mutate: func(config *InstanceConfig) {
				config.CredentialSecretMounts[0].MountPath = "/run/cryptoutil/pki/other"
			},
			wantErr: testOverlapsPKIDirectoryError,
		},
		{
			name: "invalid workload name",
			mutate: func(config *InstanceConfig) {
				config.InstanceServices[0].Name = testInvalidSecretName
			},
			wantErr: testWorkloadOrPodIdentityError,
		},
		{
			name: testInvalidPodIdentityCaseName,
			mutate: func(config *InstanceConfig) {
				config.InstanceServices[0].PodName = testInvalidSecretName
			},
			wantErr: testWorkloadOrPodIdentityError,
		},
		{
			name: testInvalidSecretNameCaseName,
			mutate: func(config *InstanceConfig) {
				config.InstanceServices[0].HealthClientSecretName = testParentSecretName
			},
			wantErr: "Secret name is invalid",
		},
		{
			name: "duplicated Secret name across roles",
			mutate: func(config *InstanceConfig) {
				config.InstanceServices[0].HealthClientSecretName = config.InstanceServices[0].TLSLeafSecretName
			},
			wantErr: testDuplicateSecretNameError,
		},
		{
			name: "TLS Secret cannot alias database Secret",
			mutate: func(config *InstanceConfig) {
				config.InstanceServices[0].TLSLeafSecretName = config.DatabaseSecret
			},
			wantErr: testOverlapsSensitiveSecretError,
		},
		{
			name: "TLS Secret cannot alias credential Secret",
			mutate: func(config *InstanceConfig) {
				config.InstanceServices[0].TLSLeafSecretName = config.CredentialSecretMounts[0].SecretName
			},
			wantErr: testOverlapsSensitiveSecretError,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			config := base
			config.InstanceServices = append([]InstanceService(nil), base.InstanceServices...)
			config.CredentialSecretMounts = append([]CredentialSecretMount(nil), base.CredentialSecretMounts...)
			testCase.mutate(&config)
			require.ErrorContains(t, config.validate(), testCase.wantErr)
		})
	}
}

func TestSafeConfigPathBasenamePattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{name: "one character", value: "a", valid: true},
		{name: "allowed basename punctuation", value: "framework_config-v2.yaml", valid: true},
		{name: "253 character maximum", value: strings.Repeat("a", 253), valid: true},
		{name: "254 character basename", value: strings.Repeat("a", 254)},
		{name: cryptoutilSharedMagic.EMPTY, value: ""},
		{name: "dot", value: testDotPath},
		{name: "dot dot", value: ".."},
		{name: "nested path", value: "nested/file.yaml"},
		{name: "absolute path", value: testAbsoluteFrameworkConfigFile},
		{name: "backslash path", value: `nested\file.yaml`},
		{name: "NUL", value: testFrameworkConfigFilenameWithNUL},
		{name: "space", value: "framework file.yaml"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, testCase.valid, safeConfigPath(testCase.value))
		})
	}
}
