// Copyright (c) 2025-2026 Justin Cranford.
package magic

const (
	// KubernetesSecretKind is the Kubernetes API kind for a Secret resource.
	KubernetesSecretKind = "Secret"
	// KubernetesSecretTypeOpaque is the Kubernetes API type for an opaque Secret.
	KubernetesSecretTypeOpaque = "Opaque"

	// KubernetesLabelPSID is the shared Kubernetes label key for a PS-ID.
	KubernetesLabelPSID = "cu.io/psid"
	// KubernetesLabelRun is the shared Kubernetes label key for a deployment run.
	KubernetesLabelRun = "cu.io/run"
	// KubernetesLabelRelease is the shared Kubernetes label key for a deployment release.
	KubernetesLabelRelease = "cu.io/release"
)
