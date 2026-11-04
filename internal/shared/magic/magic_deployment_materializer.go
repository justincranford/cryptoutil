// Copyright (c) 2025-2026 Justin Cranford.

package magic

import "time"

const (
	DeploymentMaterializerDefaultOutputDirectory = "/run/cryptoutil/pki/output"
	DeploymentMaterializerHealthClientDirectory  = "/run/cryptoutil/pki/health-client"
	DeploymentMaterializerOTLPClientDirectory    = "/run/cryptoutil/pki/otlp-client"
	DeploymentMaterializerProjectedTokenPath     = "/run/cryptoutil/pki/input/token"
	DeploymentMaterializerProjectedAPICAPath     = "/run/cryptoutil/pki/input/ca.crt"

	DeploymentMaterializerSecretLabelWorkloadName = "cu.io/workload"
	DeploymentMaterializerSecretLabelPSID         = KubernetesLabelPSID
	DeploymentMaterializerSecretLabelPool         = "cu.io/pool"
	DeploymentMaterializerSecretLabelRun          = KubernetesLabelRun
	DeploymentMaterializerSecretLabelRelease      = KubernetesLabelRelease
	DeploymentMaterializerSecretLabelRole         = "cu.io/materializer-role"

	DeploymentMaterializerSecretRoleTLSBundle  = "tls-bundle"
	DeploymentMaterializerSecretRoleHealth     = "health-client"
	DeploymentMaterializerSecretRoleOTLP       = "otlp-client"
	DeploymentMaterializerSecretTypeOpaque     = KubernetesSecretTypeOpaque
	DeploymentMaterializerKubernetesAPIVersion = "v1"
	DeploymentMaterializerKubernetesSecretKind = KubernetesSecretKind

	DeploymentMaterializerSecretDataPublicCertificate     = "public.crt"
	DeploymentMaterializerSecretDataPublicPrivateKey      = "public.key"
	DeploymentMaterializerSecretDataAdminCertificate      = "admin.crt"
	DeploymentMaterializerSecretDataAdminPrivateKey       = AdminPrivateKeyFileName
	DeploymentMaterializerSecretDataPostgreSQLCertificate = "postgres.crt"
	DeploymentMaterializerSecretDataPostgreSQLPrivateKey  = "postgres.key"
	DeploymentMaterializerSecretDataCertificate           = "tls.crt"
	DeploymentMaterializerSecretDataPrivateKey            = "tls.key"

	DeploymentMaterializerPublicServerCertificateFile     = "public-server/tls.crt"
	DeploymentMaterializerPublicServerPrivateKeyFile      = "public-server/tls.key"
	DeploymentMaterializerAdminServerCertificateFile      = "admin-server/tls.crt"
	DeploymentMaterializerAdminServerPrivateKeyFile       = "admin-server/tls.key"
	DeploymentMaterializerHealthClientCertificateFile     = "health-client/tls.crt"
	DeploymentMaterializerHealthClientPrivateKeyFile      = "health-client/tls.key"
	DeploymentMaterializerOTLPClientCertificateFile       = "otlp-client/tls.crt"
	DeploymentMaterializerOTLPClientPrivateKeyFile        = "otlp-client/tls.key"
	DeploymentMaterializerPostgreSQLClientCertificateFile = "postgres-client/tls.crt"
	DeploymentMaterializerPostgreSQLClientPrivateKeyFile  = "postgres-client/tls.key"

	DeploymentMaterializerSecretAPIPrefix = "/api/v1/namespaces/"
	DeploymentMaterializerSecretAPIMiddle = DeploymentSecretsDirectoryPathSegment
	DeploymentMaterializerPathRoot        = "/"
	DeploymentMaterializerPathParent      = ".."

	DeploymentMaterializerPoolPG   = DeploymentPoolPG
	DeploymentMaterializerPoolMem  = DeploymentPoolMem
	DeploymentMaterializerPoolDisk = DeploymentPoolDisk

	DeploymentMaterializerMaximumTokenBytes             = 64 * 1024
	DeploymentMaterializerMaximumAPICABundleBytes       = 1024 * 1024
	DeploymentMaterializerMaximumSecretBodyBytes        = 2 * 1024 * 1024
	DeploymentMaterializerMaximumInstanceYAMLBytes      = 1 * 1024 * 1024
	DeploymentMaterializerMaximumCertificatePEMBytes    = 1 * 1024 * 1024
	DeploymentMaterializerMaximumExpectedPublicDNSNames = 16

	DeploymentMaterializerAPIRequestTimeout = 10 * time.Second
	DeploymentMaterializerOverallTimeout    = 30 * time.Second
)
