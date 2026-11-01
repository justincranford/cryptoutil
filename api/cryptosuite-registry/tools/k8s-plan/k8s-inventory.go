//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.

package main

import (
	"fmt"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// maxExpandedPods keeps planner output bounded at over 33 times the 299 pods
// in the default full inventory while including inactive replica capacity.
const maxExpandedPods = 10_000

type object struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type pod struct {
	Name       string `json:"name"`
	Service    string `json:"service"`
	TLS        string `json:"tlsSecret"`
	PVC        string `json:"pvc,omitempty"`
	FQDN       string `json:"fqdn"`
	Host       string `json:"host"`
	LegacySlot string `json:"legacySlot"`
	Active     bool   `json:"active"`
}

type workload struct {
	PSID        string   `json:"psid"`
	Pool        string   `json:"pool"`
	Name        string   `json:"name"`
	Bounds      bounds   `json:"replicas"`
	ServiceFQDN string   `json:"serviceFqdn,omitempty"`
	Host        string   `json:"host,omitempty"`
	Objects     []object `json:"objects"`
	Pods        []pod    `json:"pods"`
}

type database struct {
	PSID         string `json:"psid"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	Publication  string `json:"publication"`
	Subscription string `json:"subscription"`
}

type release struct {
	Key       string     `json:"key"`
	Namespace string     `json:"namespace"`
	Helm      string     `json:"helmRelease"`
	Objects   []object   `json:"objects"`
	Databases []database `json:"databases"`
	Workloads []workload `json:"workloads"`
}

type inventory struct {
	Version        int       `json:"version"`
	Run            string    `json:"run"`
	InfraNamespace string    `json:"infraNamespace"`
	Infra          []object  `json:"infra"`
	InfraPods      []object  `json:"infraPods"`
	Releases       []release `json:"releases"`
}

func named(kind string, names ...string) []object {
	result := make([]object, 0, len(names))
	for _, name := range names {
		result = append(result, object{Kind: kind, Name: name})
	}

	return result
}

func build(m model, r registry, selectKey string) (inventory, error) {
	if err := validate(m, r); err != nil {
		return inventory{}, err
	}

	if err := preflightPodCapacity(m, r, selectKey); err != nil {
		return inventory{}, err
	}

	inv := inventory{
		Version: m.Version, Run: m.Run, InfraNamespace: "cu-" + m.Run + "-infra",
		Releases: []release{},
	}

	type scope struct{ tier, name string }

	var scopes []scope
	for _, s := range r.Members {
		scopes = append(scopes, scope{tierService, s.PSID})
	}

	for _, p := range r.Products {
		scopes = append(scopes, scope{tierProduct, p.ID})
	}

	for _, s := range r.Suites {
		scopes = append(scopes, scope{tierSuite, s.ID})
	}

	hasPG := false

	for _, s := range scopes {
		for _, v := range m.Modes {
			key := s.tier + ":" + s.name + ":" + v.Name
			if selectKey != "" && selectKey != key {
				continue
			}

			rel := release{
				Key: key, Namespace: "cu-" + m.Run + "-" + s.tier + "-" + s.name + "-" + v.Name,
				Helm:      "cu-" + s.tier + "-" + s.name + "-" + v.Name,
				Databases: []database{},
			}
			rel.Objects = append(rel.Objects, named("NetworkPolicy", "default-deny", "allow-dns", "allow-public", "allow-dependencies", "allow-bootstrap-api")...)
			rel.Objects = append(rel.Objects, named("ResourceQuota", "budget")...)
			rel.Objects = append(rel.Objects, named("Secret", "unseal", "hash-pepper")...)
			rel.Objects = append(rel.Objects, named("ServiceAccount", "bootstrap")...)
			rel.Objects = append(rel.Objects, named("Role", "bootstrap")...)
			rel.Objects = append(rel.Objects, named("RoleBinding", "bootstrap")...)
			rel.Objects = append(rel.Objects, named("Job", "pki-init-"+m.Revision)...)

			for _, member := range r.Members {
				if (s.tier == tierService && member.PSID != s.name) || (s.tier == tierProduct && member.Product != s.name) {
					continue
				}

				id := member.PSID
				rel.Objects = append(rel.Objects, named("ConfigMap", id+"-framework", id+"-domain")...)
				rel.Objects = append(rel.Objects, named("Secret", id+"-credentials", id+"-e2e-tls")...)

				for _, pool := range v.Pools {
					w := buildWorkload(m, rel.Namespace, id, pool)
					rel.Workloads = append(rel.Workloads, w)

					if pool != poolPG {
						continue
					}

					hasPG = true
					sqlID := strings.ReplaceAll(strings.Join([]string{m.Run, s.tier, s.name, v.Name, id}, "_"), "-", "_")
					rel.Databases = append(rel.Databases, database{
						PSID: id, Name: "d_" + sqlID, Role: "u_" + sqlID,
						Publication: "pub_" + sqlID, Subscription: "sub_" + sqlID,
					})
					rel.Objects = append(rel.Objects, named("Secret", id+"-db", id+"-migration-db")...)
					rel.Objects = append(rel.Objects, named("Job", id+"-migrate-"+m.Revision)...)
				}
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
	inv.Infra = append(inv.Infra, named("Secret", "telemetry-tls", "pki-authorities")...)
	inv.Infra = append(inv.Infra, named("NetworkPolicy", "default-deny", "allow-dns", "allow-clients", "allow-infra", "allow-bootstrap-api")...)
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
		inv.InfraPods = append(inv.InfraPods, named("Pod", "postgres-leader-0", "postgres-follower-0")...)
		inv.Infra = append(inv.Infra, named("StatefulSet", cryptoutilSharedMagic.PKIInitPostgresLeaderService, cryptoutilSharedMagic.PKIInitPostgresFollowerService)...)
		inv.Infra = append(inv.Infra, named("Service", cryptoutilSharedMagic.PKIInitPostgresLeaderService, "postgres-leader-pods", cryptoutilSharedMagic.PKIInitPostgresFollowerService, "postgres-follower-pods")...)
		inv.Infra = append(inv.Infra, named("PersistentVolumeClaim", "data-postgres-leader-0", "data-postgres-follower-0")...)
		inv.Infra = append(inv.Infra, named("Secret", "postgres-tls", "postgres-bootstrap")...)
		inv.Infra = append(inv.Infra, named("ConfigMap", cryptoutilSharedMagic.DockerServicePostgres, "database-catalog")...)
		inv.Infra = append(inv.Infra, named("Job", "db-init-"+m.Revision, "replication-"+m.Revision)...)
	}

	return inv, validateInventory(inv)
}

func preflightPodCapacity(m model, r registry, selectKey string) error {
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

	for _, service := range r.Members {
		for _, v := range m.Modes {
			key := "s:" + service.PSID + ":" + v.Name
			if selectKey != "" && selectKey != key {
				continue
			}

			matchedRelease = true

			for _, pool := range v.Pools {
				if pool == poolPG {
					hasPG = true
				}
			}

			if err := addMemberCapacity(service.PSID, v); err != nil {
				return err
			}
		}
	}

	for _, product := range r.Products {
		for _, v := range m.Modes {
			key := "p:" + product.ID + ":" + v.Name
			if selectKey != "" && selectKey != key {
				continue
			}

			matchedRelease = true

			for _, service := range r.Members {
				if service.Product != product.ID {
					continue
				}

				for _, pool := range v.Pools {
					if pool == poolPG {
						hasPG = true
					}
				}

				if err := addMemberCapacity(service.PSID, v); err != nil {
					return err
				}
			}
		}
	}

	for _, suite := range r.Suites {
		for _, v := range m.Modes {
			key := "u:" + suite.ID + ":" + v.Name
			if selectKey != "" && selectKey != key {
				continue
			}

			matchedRelease = true

			for _, service := range r.Members {
				for _, pool := range v.Pools {
					if pool == poolPG {
						hasPG = true
					}
				}

				if err := addMemberCapacity(service.PSID, v); err != nil {
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
	if pool == poolPG {
		w.ServiceFQDN = w.Name + "." + ns + ".svc." + m.ClusterDomain
		w.Host = w.Name + "." + ns + "." + m.LocalDomain
		w.Objects = append(w.Objects, named("Service", w.Name)...)
		w.Objects = append(w.Objects, named("HorizontalPodAutoscaler", w.Name)...)
	}

	for i := range b.Max {
		name := fmt.Sprintf("%s-%d", w.Name, i)
		instance := fmt.Sprintf("%s-i%02d", w.Name, i+1)
		backend := map[string]string{poolPG: "postgresql", poolMem: poolBackendSQLite, poolDisk: poolBackendSQLite}[pool]

		p := pod{
			Name: name, Service: instance, TLS: name + "-tls",
			FQDN:       name + "." + w.Name + "-pods." + ns + ".svc." + m.ClusterDomain,
			Host:       instance + "." + ns + "." + m.LocalDomain,
			LegacySlot: fmt.Sprintf("%s-%s-%d", id, backend, i+1), Active: i < b.Min,
		}
		if pool == poolDisk {
			p.PVC = "data-" + name
		}

		w.Pods = append(w.Pods, p)
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

				objects = append(objects, named("Secret", p.TLS)...)
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
