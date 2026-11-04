// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const (
	lifecycleImportCheckFormat  = "import-check"
	lifecycleJobKind            = "Job"
	lifecyclePKICommand         = cryptoutilSharedMagic.PSIDPKIInit
	lifecycleInfraPKIName       = "infra-pki"
	lifecycleAuthoritySecret    = "pki-authorities"
	lifecyclePepperSecret       = "hash-pepper"
	lifecyclePostgresBootstrap  = "postgres-bootstrap"
	lifecycleTLSSecretType      = "kubernetes.io/tls"
	lifecycleOpaqueSecretType   = cryptoutilSharedMagic.KubernetesSecretTypeOpaque
	lifecycleRevisionFormat     = "revision"
	lifecycleValuesFormat       = "values"
	lifecycleSecretPlanFormat   = "secret-plan"
	lifecycleRootPath           = "."
	lifecycleMaxSecretBytes     = 1 << 20
	lifecycleMaxDeadlineSeconds = 3600
	lifecycleMaxBackoffLimit    = 3
	lifecycleValuesSchemaPath   = "api/cryptosuite-registry/templates/kubernetes/cryptoutil-lifecycle/values.schema.json"
	lifecyclePasswordKey        = "password"
	lifecycleUnsealSecret       = "unseal"
)
