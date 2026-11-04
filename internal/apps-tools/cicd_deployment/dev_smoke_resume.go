// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"errors"
	"strings"
)

const (
	developmentKubeSystem     = "kube-system"
	developmentKubePublic     = "kube-public"
	developmentKubeNodeLease  = "kube-node-lease"
	developmentListKind       = "List"
	developmentJobKind        = "Job"
	developmentPodKind        = "Pod"
	developmentGetCommand     = "get"
	developmentSystemNSCount  = 4
	developmentComposeProject = "com.docker.compose.project"
	developmentNodeIDLabel    = "cu.io/node-id"
	developmentLocalVolume    = cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault
	developmentItems          = "items"
)

// Recovery does not adopt an application deployment. Before any image import,
// Secret publication or Helm mutation, require only K3s system namespaces and
// system workloads, with no cryptoutil run-labelled objects. This is a recovery
// guard, not cluster certification or application acceptance.
func validateDevelopmentBootstrapInventory(raw []byte) error {
	var list struct {
		Kind  string
		Items []struct {
			Kind     string
			Metadata struct {
				Name, Namespace string
				Labels          map[string]string
			}
		}
	}
	if err := json.Unmarshal(raw, &list); err != nil || list.Kind != developmentListKind || len(list.Items) == 0 {
		return errors.New("bootstrap recovery live inventory rejected")
	}

	namespaces := map[string]bool{}

	for _, object := range list.Items {
		if object.Metadata.Name == "" || object.Metadata.Labels[cryptoutilSharedMagic.KubernetesLabelRun] != "" {
			return errors.New("bootstrap recovery existing application identity rejected")
		}

		switch object.Kind {
		case developmentNamespaceKind:
			switch object.Metadata.Name {
			case developmentDefaultContext, developmentKubeSystem, developmentKubePublic, developmentKubeNodeLease:
				if namespaces[object.Metadata.Name] {
					return errors.New("bootstrap recovery duplicate namespace rejected")
				}

				namespaces[object.Metadata.Name] = true
			default:
				return errors.New("bootstrap recovery existing application namespace rejected")
			}
		case developmentPodKind, developmentJobKind:
			if object.Metadata.Namespace != developmentKubeSystem {
				return errors.New("bootstrap recovery existing application workload rejected")
			}

			if !developmentBootstrapSystemWorkload(object.Kind, object.Metadata.Name) {
				return errors.New("bootstrap recovery non-baseline system workload rejected")
			}
		default:
			return errors.New("bootstrap recovery unexpected inventory kind rejected")
		}
	}

	if len(namespaces) != developmentSystemNSCount {
		return errors.New("bootstrap recovery incomplete system namespace inventory rejected")
	}

	return nil
}

// These are the enabled baseline components of the canonical, pinned K3s
// bootstrap. A namespace alone is not permission to adopt arbitrary workloads.
func developmentBootstrapSystemWorkload(kind, name string) bool {
	if kind == developmentJobKind {
		return name == "helm-install-traefik" || name == "helm-install-traefik-crd"
	}

	for _, prefix := range []string{
		"coredns-", "local-path-provisioner-", "metrics-server-", "traefik-", "svclb-traefik-",
		"helm-install-traefik-", "helm-install-traefik-crd-",
	} {
		if strings.HasPrefix(name, prefix) && len(name) > len(prefix) {
			return true
		}
	}

	return false
}
