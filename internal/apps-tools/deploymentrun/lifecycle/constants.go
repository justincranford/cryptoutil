// Copyright (c) 2025-2026 Justin Cranford.

package lifecycle

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const (
	namespaceKind    = "Namespace"
	secretKind       = cryptoutilSharedMagic.KubernetesSecretKind
	jobKind          = "Job"
	configMapKind    = "ConfigMap"
	statefulSetKind  = "StatefulSet"
	importOwner      = "external-secret-import"
	bootstrapName    = "bootstrap"
	dash             = "-"
	outputFlag       = "-o"
	runLabel         = cryptoutilSharedMagic.KubernetesLabelRun
	infraPKIName     = "infra-pki"
	databaseInitName = "db-init"
	apiVersion       = "v1"
	jsonOutput       = "json"
)
