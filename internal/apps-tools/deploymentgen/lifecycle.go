// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"slices"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	lifecycleExternalSecretIdentity = "external-secret-identity"
	lifecycleJobIdentity            = "lifecycle-job-identity"
	lifecycleBootstrapRBACIdentity  = "bootstrap-rbac-identity"
)

// LifecycleIdentity is a lifecycle-owned Kubernetes resource identity. It intentionally contains
// no Secret data, Job execution details, or other fields that would imply a deployable manifest.
type LifecycleIdentity struct {
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Owner     string `json:"owner"`
}

// LifecycleInventory lists graph identities that are provisioned outside the runtime and
// infrastructure Helm charts. It is an identity plan, not a Kubernetes manifest or runtime
// acceptance result.
type LifecycleInventory struct {
	Version    int                 `json:"version"`
	Run        string              `json:"run"`
	Identities []LifecycleIdentity `json:"identities"`
}

// BuildLifecycleInventory derives external Secret, lifecycle Job, and bootstrap RBAC identities
// from an already validated deployment inventory. Secret material and lifecycle execution
// semantics remain owned by their future provisioners.
func BuildLifecycleInventory(inventory Inventory) (LifecycleInventory, error) {
	if err := validateInventory(inventory); err != nil {
		return LifecycleInventory{}, fmt.Errorf("invalid inventory: %w", err)
	}

	result := LifecycleInventory{
		Version:    inventory.Version,
		Run:        inventory.Run,
		Identities: make([]LifecycleIdentity, 0),
	}

	appendObject := func(namespace string, object Object) {
		owner := lifecycleOwner(object)
		if owner == "" {
			return
		}

		result.Identities = append(result.Identities, LifecycleIdentity{
			Namespace: namespace,
			Kind:      object.Kind,
			Name:      object.Name,
			Owner:     owner,
		})
	}

	appendObjects := func(namespace string, objects []Object) {
		for _, object := range objects {
			appendObject(namespace, object)
		}
	}

	appendObjects(inventory.InfraNamespace, inventory.Infra)

	for _, release := range inventory.Releases {
		appendObjects(release.Namespace, release.Objects)

		for _, workload := range release.Workloads {
			appendObjects(release.Namespace, workload.Objects)

			for _, pod := range workload.Pods {
				appendObject(release.Namespace, Object{
					Kind: cryptoutilSharedMagic.KubernetesSecretKind,
					Name: pod.TLS,
				})
			}
		}
	}

	slices.SortFunc(result.Identities, func(left, right LifecycleIdentity) int {
		if namespaceOrder := compareString(left.Namespace, right.Namespace); namespaceOrder != 0 {
			return namespaceOrder
		}

		if kindOrder := compareString(left.Kind, right.Kind); kindOrder != 0 {
			return kindOrder
		}

		if nameOrder := compareString(left.Name, right.Name); nameOrder != 0 {
			return nameOrder
		}

		return compareString(left.Owner, right.Owner)
	})

	return result, nil
}

func lifecycleOwner(object Object) string {
	switch object.Kind {
	case cryptoutilSharedMagic.KubernetesSecretKind:
		return lifecycleExternalSecretIdentity
	case lifecycleJobKind:
		return lifecycleJobIdentity
	case "ServiceAccount", "Role", "RoleBinding":
		if object.Name == "bootstrap" {
			return lifecycleBootstrapRBACIdentity
		}
	}

	return ""
}

func compareString(left, right string) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}
