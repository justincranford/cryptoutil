// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"fmt"
	"slices"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// maxExpandedPods keeps planner output bounded at over 33 times the 299 pods
// in the default full inventory while including inactive replica capacity.
const maxExpandedPods = 10_000

// Object identifies a namespaced or cluster-scoped Kubernetes resource.
type Object struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type object = Object

// Pod is a stable maximum-capacity application pod identity.
type Pod struct {
	Name       string `json:"name"`
	Service    string `json:"service"`
	TLS        string `json:"tlsSecret"`
	PVC        string `json:"pvc,omitempty"`
	FQDN       string `json:"fqdn"`
	Host       string `json:"host"`
	LegacySlot string `json:"legacySlot"`
	Active     bool   `json:"active"`
}

type pod = Pod

// Workload describes one PS-ID/backend pool and its configured pod identities.
type Workload struct {
	PSID        string   `json:"psid"`
	Pool        string   `json:"pool"`
	Name        string   `json:"name"`
	Bounds      bounds   `json:"replicas"`
	ServiceFQDN string   `json:"serviceFqdn,omitempty"`
	Host        string   `json:"host,omitempty"`
	Objects     []object `json:"objects"`
	Pods        []pod    `json:"pods"`
}

type workload = Workload

// Database identifies the isolated logical database and replication objects.
type Database struct {
	PSID         string `json:"psid"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	Publication  string `json:"publication"`
	Subscription string `json:"subscription"`
}

type database = Database

// Release describes one independently namespaced registry scope and mode.
type Release struct {
	Key       string     `json:"key"`
	Namespace string     `json:"namespace"`
	Helm      string     `json:"helmRelease"`
	Objects   []object   `json:"objects"`
	Databases []database `json:"databases"`
	Workloads []workload `json:"workloads"`
}

type release = Release

// Inventory is the deterministic complete graph for selected release scopes.
type Inventory struct {
	Version        int       `json:"version"`
	Run            string    `json:"run"`
	InfraNamespace string    `json:"infraNamespace"`
	Infra          []object  `json:"infra"`
	InfraPods      []object  `json:"infraPods"`
	Releases       []release `json:"releases"`
}

type inventory = Inventory

func named(kind string, names ...string) []object {
	result := make([]object, 0, len(names))
	for _, name := range names {
		result = append(result, object{Kind: kind, Name: name})
	}

	return result
}

func build(m model, scopes []Scope, selectKey string) (inventory, error) {
	if err := validate(m, scopes); err != nil {
		return inventory{}, err
	}

	if err := preflightPodCapacity(m, scopes, selectKey); err != nil {
		return inventory{}, err
	}

	inv := inventory{
		Version: m.Version, Run: m.Run, InfraNamespace: "cu-" + m.Run + "-infra",
		Releases: []release{},
	}

	hasPG := false

	for _, scope := range scopes {
		for _, v := range m.Modes {
			key := string(scope.Tier) + ":" + scope.ID + ":" + v.Name
			if selectKey != "" && selectKey != key {
				continue
			}

			rel := release{
				Key: key, Namespace: "cu-" + m.Run + "-" + string(scope.Tier) + "-" + scope.ID + "-" + v.Name,
				Helm:      "cu-" + string(scope.Tier) + "-" + scope.ID + "-" + v.Name,
				Databases: []database{},
			}

			rel.Objects = append(rel.Objects, named("NetworkPolicy", "default-deny", "allow-dns-and-bootstrap-api", "allow-approved-release-edges")...)
			for _, dependency := range registeredAPIDependenciesForMembers(scope.Members) {
				rel.Objects = append(rel.Objects, named(
					"NetworkPolicy", apiServiceDependencyPolicyNames(dependency)...,
				)...)
			}

			rel.Objects = append(rel.Objects, named("ResourceQuota", "budget")...)
			rel.Objects = append(rel.Objects, named(cryptoutilSharedMagic.KubernetesSecretKind, "unseal", lifecyclePepperSecret)...)
			rel.Objects = append(rel.Objects, named("ServiceAccount", "bootstrap")...)
			rel.Objects = append(rel.Objects, named("Role", "bootstrap")...)
			rel.Objects = append(rel.Objects, named("RoleBinding", "bootstrap")...)
			rel.Objects = append(rel.Objects, named("Job", "pki-init-"+m.Revision)...)

			releaseHasPG := false

			for _, id := range scope.Members {
				rel.Objects = append(rel.Objects, named("ConfigMap", id+"-framework", id+"-domain")...)
				rel.Objects = append(rel.Objects, named(cryptoutilSharedMagic.KubernetesSecretKind, id+"-credentials", id+"-e2e-tls", id+"-workload-trust")...)

				for _, pool := range v.Pools {
					w := buildWorkload(m, rel.Namespace, id, pool)
					rel.Workloads = append(rel.Workloads, w)

					if pool != cryptoutilSharedMagic.DeploymentPoolPG {
						continue
					}

					hasPG = true
					releaseHasPG = true
					sqlID := strings.ReplaceAll(strings.Join([]string{m.Run, string(scope.Tier), scope.ID, v.Name, id}, "_"), "-", "_")
					rel.Databases = append(rel.Databases, database{
						PSID: id, Name: "d_" + sqlID, Role: "u_" + sqlID,
						Publication: "pub_" + sqlID, Subscription: "sub_" + sqlID,
					})
					rel.Objects = append(rel.Objects, named(cryptoutilSharedMagic.KubernetesSecretKind, id+"-db", id+"-migration-db")...)
					rel.Objects = append(rel.Objects, named("Job", id+"-migrate-"+m.Revision)...)
				}
			}

			if releaseHasPG {
				rel.Objects = append(rel.Objects, named("NetworkPolicy", "allow-postgres-leader")...)
			}

			inv.Releases = append(inv.Releases, rel)
		}
	}

	if len(inv.Releases) == 0 {
		return inv, fmt.Errorf("selector matched no release: %s", selectKey)
	}

	inv.Infra = append(inv.Infra, named("StatefulSet", "collector", "lgtm")...)
	inv.Infra = append(inv.Infra, named("Service", "collector", "collector-pods", "lgtm", "lgtm-pods")...)
	inv.Infra = append(inv.Infra, named("PersistentVolumeClaim", "data-lgtm-0")...)
	inv.Infra = append(inv.Infra, named("ConfigMap", "collector", "lgtm")...)
	inv.Infra = append(inv.Infra, named(
		cryptoutilSharedMagic.KubernetesSecretKind, "pki-authorities",
		"collector-server-tls", "collector-client-ca", "lgtm-server-ca", "collector-lgtm-client-tls",
		"lgtm-server-tls", "lgtm-client-ca", "lgtm-admin",
	)...)
	inv.Infra = append(inv.Infra, named(
		"NetworkPolicy", "default-deny", "allow-dns", "allow-bootstrap-api",
		"allow-client-telemetry", "allow-collector-egress-to-lgtm", "allow-collector-ingress-to-lgtm",
	)...)
	inv.Infra = append(inv.Infra, named("ServiceAccount", "bootstrap")...)
	inv.Infra = append(inv.Infra, named("Role", "bootstrap")...)
	inv.Infra = append(inv.Infra, named("RoleBinding", "bootstrap")...)
	inv.Infra = append(inv.Infra, named("Job", "infra-pki-"+m.Revision)...)
	inv.Infra = append(inv.Infra, named("ResourceQuota", "budget")...)
	inv.Infra = append(inv.Infra, named("PodDisruptionBudget", "collector")...)

	inv.InfraPods = append(inv.InfraPods, named("Pod", "lgtm-0")...)
	for i := range m.Infrastructure.Collector {
		inv.InfraPods = append(inv.InfraPods, named("Pod", fmt.Sprintf("collector-%d", i))...)
		inv.Infra = append(inv.Infra, named("PersistentVolumeClaim", fmt.Sprintf("data-collector-%d", i))...)
	}

	if hasPG {
		inv.Infra = append(inv.Infra, named("NetworkPolicy", "allow-postgres-edges", "allow-postgres-follower-to-leader")...)
		inv.Infra = append(inv.Infra, named("NetworkPolicy", "allow-bootstrap-postgres-egress", "allow-postgres-from-bootstrap")...)
		inv.InfraPods = append(inv.InfraPods, named("Pod", "postgres-leader-0", "postgres-follower-0")...)
		inv.Infra = append(inv.Infra, named("StatefulSet", cryptoutilSharedMagic.PKIInitPostgresLeaderService, cryptoutilSharedMagic.PKIInitPostgresFollowerService)...)
		inv.Infra = append(inv.Infra, named("Service", cryptoutilSharedMagic.PKIInitPostgresLeaderService, "postgres-leader-pods", cryptoutilSharedMagic.PKIInitPostgresFollowerService, "postgres-follower-pods")...)
		inv.Infra = append(inv.Infra, named("PersistentVolumeClaim", "data-postgres-leader-0", "data-postgres-follower-0")...)
		inv.Infra = append(inv.Infra, named(
			cryptoutilSharedMagic.KubernetesSecretKind, "postgres-leader-tls", "postgres-follower-tls",
			"postgres-client-ca", "postgres-bootstrap", "postgres-follower-bootstrap",
		)...)
		inv.Infra = append(inv.Infra, named(
			"ConfigMap", cryptoutilSharedMagic.KMSPostgresLeaderContainer,
			cryptoutilSharedMagic.PKIInitPostgresFollowerService, "database-catalog",
		)...)
		inv.Infra = append(inv.Infra, named("Job", "db-init-"+m.Revision, "replication-"+m.Revision)...)
	}

	return inv, validateInventory(inv)
}

func preflightPodCapacity(m model, scopes []Scope, selectKey string) error {
	total := 0
	addCapacity := func(count int) error {
		if count < 0 || count > maxExpandedPods-total {
			return fmt.Errorf("expanded pod capacity exceeds limit of %d", maxExpandedPods)
		}

		total += count

		return nil
	}
	addMemberCapacity := func(id string, v mode) error {
		for _, pool := range v.Pools {
			b, ok := m.Overrides[id+"/"+pool]
			if !ok {
				b = m.Replicas
			}

			if err := addCapacity(b.Max); err != nil {
				return err
			}
		}

		return nil
	}

	matchedRelease := false
	hasPG := false

	for _, scope := range scopes {
		for _, v := range m.Modes {
			key := string(scope.Tier) + ":" + scope.ID + ":" + v.Name
			if selectKey != "" && selectKey != key {
				continue
			}

			matchedRelease = true

			for _, id := range scope.Members {
				if slices.Contains(v.Pools, cryptoutilSharedMagic.DeploymentPoolPG) {
					hasPG = true
				}

				if err := addMemberCapacity(id, v); err != nil {
					return err
				}
			}
		}
	}

	if !matchedRelease {
		return nil
	}

	if err := addCapacity(1); err != nil { // LGTM pod.
		return err
	}

	if err := addCapacity(m.Infrastructure.Collector); err != nil {
		return err
	}

	if hasPG {
		if err := addCapacity(2); err != nil { // PostgreSQL leader and follower.
			return err
		}
	}

	return nil
}

func buildWorkload(m model, ns, id, pool string) workload {
	b, ok := m.Overrides[id+"/"+pool]
	if !ok {
		b = m.Replicas
	}

	w := workload{PSID: id, Pool: pool, Name: id + "-" + pool, Bounds: b}
	w.Objects = append(w.Objects, named("StatefulSet", w.Name)...)
	w.Objects = append(w.Objects, named("Service", w.Name+"-pods")...)
	w.Objects = append(w.Objects, named("ServiceAccount", w.Name)...)
	w.Objects = append(w.Objects, named("Role", w.Name+"-cert-reader")...)
	w.Objects = append(w.Objects, named("RoleBinding", w.Name+"-cert-reader")...)
	w.Objects = append(w.Objects, named("PodDisruptionBudget", w.Name)...)

	w.Objects = append(w.Objects, named("ConfigMap", w.Name+"-config")...)
	if pool == cryptoutilSharedMagic.DeploymentPoolPG {
		w.ServiceFQDN = w.Name + "." + ns + ".svc." + m.ClusterDomain
		w.Host = w.Name + "." + ns + "." + m.LocalDomain
		w.Objects = append(w.Objects, named("Service", w.Name)...)
		w.Objects = append(w.Objects, named("HorizontalPodAutoscaler", w.Name)...)
	}

	for i := range b.Max {
		name := fmt.Sprintf("%s-%d", w.Name, i)
		instance := fmt.Sprintf("%s-i%02d", w.Name, i+1)
		backend := map[string]string{cryptoutilSharedMagic.DeploymentPoolPG: "postgresql", cryptoutilSharedMagic.DeploymentPoolMem: poolBackendSQLite, cryptoutilSharedMagic.DeploymentPoolDisk: poolBackendSQLite}[pool]

		p := pod{
			Name: name, Service: instance, TLS: name + "-tls",
			FQDN:       name + "." + w.Name + "-pods." + ns + ".svc." + m.ClusterDomain,
			Host:       instance + "." + ns + "." + m.LocalDomain,
			LegacySlot: fmt.Sprintf("%s-%s-%d", id, backend, i+1), Active: i < b.Min,
		}
		if pool == cryptoutilSharedMagic.DeploymentPoolDisk {
			p.PVC = "data-" + name
		}

		w.Pods = append(w.Pods, p)
		w.Objects = append(w.Objects, named(
			cryptoutilSharedMagic.KubernetesSecretKind, name+"-health-client", name+"-otlp-client",
		)...)
	}

	return w
}

func sqlIdentifier(value string) bool {
	if len(value) == 0 || len(value) > 63 {
		return false
	}

	for i := 0; i < len(value); i++ {
		c := value[i]
		if (c < 'a' || c > 'z') && c != '_' && (i == 0 || c < '0' || c > '9') {
			return false
		}
	}

	return true
}

func validateInventory(inv inventory) error {
	identities := map[string]bool{}

	check := func(namespace, kind, name string) error {
		key := namespace + "/" + kind + "/" + name
		if !label(name) || identities[key] {
			return fmt.Errorf("invalid or duplicate identity: %s", key)
		}

		identities[key] = true

		return nil
	}
	if err := check("", "Namespace", inv.InfraNamespace); err != nil {
		return err
	}

	for _, o := range inv.Infra {
		if err := check(inv.InfraNamespace, o.Kind, o.Name); err != nil {
			return err
		}
	}

	for _, o := range inv.InfraPods {
		if err := check(inv.InfraNamespace, o.Kind, o.Name); err != nil {
			return err
		}
	}

	sqlNames := map[string]bool{}

	for _, rel := range inv.Releases {
		if err := check("", "Namespace", rel.Namespace); err != nil {
			return err
		}

		if !label(rel.Helm) || len(rel.Helm) > 53 {
			return fmt.Errorf("invalid Helm release: %s", rel.Helm)
		}

		objects := append([]object{}, rel.Objects...)
		for _, db := range rel.Databases {
			for _, name := range []string{db.Name, db.Role, db.Publication, db.Subscription} {
				if !sqlIdentifier(name) || sqlNames[name] {
					return fmt.Errorf("invalid or duplicate SQL identifier: %s", name)
				}

				sqlNames[name] = true
			}
		}

		for _, w := range rel.Workloads {
			if (w.ServiceFQDN != "" && !dnsName(w.ServiceFQDN)) || (w.Host != "" && !dnsName(w.Host)) {
				return fmt.Errorf("invalid DNS name (RFC 1123 labels, at most 253 bytes): %s", w.Name)
			}

			objects = append(objects, w.Objects...)
			for _, p := range w.Pods {
				if !dnsName(p.FQDN) || !dnsName(p.Host) {
					return fmt.Errorf("invalid DNS name (RFC 1123 labels, at most 253 bytes): %s", p.Name)
				}

				objects = append(objects, named("Pod", p.Name)...)
				objects = append(objects, named("Service", p.Service)...)

				objects = append(objects, named(cryptoutilSharedMagic.KubernetesSecretKind, p.TLS)...)
				if p.PVC != "" {
					objects = append(objects, named("PersistentVolumeClaim", p.PVC)...)
				}
			}
		}

		for _, o := range objects {
			if err := check(rel.Namespace, o.Kind, o.Name); err != nil {
				return err
			}
		}
	}

	return nil
}
