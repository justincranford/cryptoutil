//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.

package main

import (
	"bytes"
	json "encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func fixture(t *testing.T) (model, registry) {
	t.Helper()

	mBytes, err := os.ReadFile(filepath.Join("..", "..", "k8s", k8sModelFileName))
	require.NoError(t, err)
	m, err := decodeModel(mBytes)
	require.NoError(t, err)
	rBytes, err := os.ReadFile(filepath.Join("..", "..", k8sRegistryFileName))
	require.NoError(t, err)
	r, err := decodeRegistry(rBytes)
	require.NoError(t, err)

	return m, r
}

func TestInventoryMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		selector string
		releases int
		pools    int
		pods     int
		dbs      int
	}{
		{"", 55, 147, 294, 63},
		{"s:sm-kms:mixed-mem", 1, 2, 4, 1},
		{"p:identity:mixed-disk", 1, 10, 20, 5},
		{testSuitePGSelector, 1, 7, 14, 7},
		{testProductSMMemSelector, 1, 1, 2, 0},
		{"p:skeleton:disk", 1, 1, 2, 0},
	}
	for _, tc := range tests {
		t.Run("selector="+tc.selector, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			inv, err := build(m, r, tc.selector)
			require.NoError(t, err)
			require.Len(t, inv.Releases, tc.releases)

			pools, pods, dbs, activeSQLitePVCs := 0, 0, 0, 0
			for _, rel := range inv.Releases {
				dbs += len(rel.Databases)

				pools += len(rel.Workloads)
				for _, w := range rel.Workloads {
					pods += w.Bounds.Min
					require.Len(t, w.Pods, w.Bounds.Max)

					for _, p := range w.Pods {
						if p.PVC != "" && p.Active {
							activeSQLitePVCs++
						}
					}
				}
			}

			require.Equal(t, tc.pools, pools)
			require.Equal(t, tc.pods, pods)
			require.Equal(t, tc.dbs, dbs)

			if tc.selector == "" {
				require.Equal(t, 84, activeSQLitePVCs)
			}

			repeated, err := build(m, r, tc.selector)
			require.NoError(t, err)
			require.Equal(t, inv, repeated)
			require.Equal(t, render(m, r, inv), render(m, r, repeated))
		})
	}
}

func TestInventoryValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*model, *registry)
		want string
	}{
		{"bad run", func(m *model, _ *registry) { m.Run = "Upper" }, testInvalidVersion},
		{"bad version", func(m *model, _ *registry) { m.Version = 2 }, testInvalidVersion},
		{"bad domain", func(m *model, _ *registry) { m.ClusterDomain = "cluster..local" }, testInvalidDomain},
		{"domain label too long", func(m *model, _ *registry) { m.ClusterDomain = strings.Repeat("a", 64) }, testInvalidDomain},
		{"long domain suffix", func(m *model, _ *registry) { m.ClusterDomain = strings.Repeat("a", 254) }, "domain too long"},
		{"bad bounds", func(m *model, _ *registry) { m.Replicas = bounds{testThree, 2} }, testReplicasKey},
		{"zero bounds", func(m *model, _ *registry) { m.Replicas = bounds{0, 2} }, testReplicasKey},
		{"unknown override", func(m *model, _ *registry) { m.Overrides["missing/pg"] = bounds{2, 2} }, "override"},
		{"bad override", func(m *model, _ *registry) { m.Overrides["sm-kms/pg"] = bounds{2, 1} }, "replicas"},
		{"unknown pool", func(m *model, _ *registry) { m.Modes[0].Pools = []string{"not-a-pool"} }, testPoolError},
		{"duplicate pool", func(m *model, _ *registry) { m.Modes[0].Pools = []string{poolPG, poolPG} }, testPoolError},
		{"duplicate mode", func(m *model, _ *registry) { m.Modes = append(m.Modes, m.Modes[0]) }, "mode"},
		{"empty modes", func(m *model, _ *registry) { m.Modes = nil }, testNonemptyError},
		{"empty pool list", func(m *model, _ *registry) { m.Modes[0].Pools = nil }, "empty mode"},
		{"orphan service", func(_ *model, r *registry) { r.Members[0].Product = "missing" }, testPSIDError},
		{"duplicate service", func(_ *model, r *registry) { r.Members = append(r.Members, r.Members[0]) }, testPSIDError},
		{"duplicate product", func(_ *model, r *registry) { r.Products = append(r.Products, r.Products[0]) }, testProductIdentity},
		{"duplicate suite", func(_ *model, r *registry) { r.Suites = append(r.Suites, r.Suites[0]) }, testSuiteIdentity},
		{"empty registry", func(_ *model, r *registry) { r.Members = nil }, testNonemptyError},
		{"empty product", func(_ *model, r *registry) { r.Products = append(r.Products, product{ID: testExtraIdentity}) }, testEmptyProductError},
		{"two primaries", func(m *model, _ *registry) { m.Infrastructure.Leader = 2 }, "singletons"},
		{"long names", func(m *model, _ *registry) { m.Revision = "r" + string(bytes.Repeat([]byte("x"), 62)) }, testDuplicateIdentityError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			tc.edit(&m, &r)
			_, err := build(m, r, "")
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestInventoryIsolationAndScaling(t *testing.T) {
	t.Parallel()
	m, r := fixture(t)
	m.Overrides["sm-kms/pg"] = bounds{1, testThree}
	first, err := build(m, r, testServiceSMKMSPGSelector)
	require.NoError(t, err)

	w := first.Releases[0].Workloads[0]
	require.Equal(t, bounds{1, testThree}, w.Bounds)
	require.Len(t, w.Pods, testThree)
	require.True(t, w.Pods[0].Active)
	require.False(t, w.Pods[1].Active)
	require.Equal(t, "sm-kms-postgresql-1", w.Pods[0].LegacySlot)
	require.Equal(t, "sm-kms-pg-0", w.Pods[0].Name)
	require.Contains(t, string(render(m, r, first)), "reserved")
	m.Run = "other"
	second, err := build(m, r, testServiceSMKMSPGSelector)
	require.NoError(t, err)
	require.NotEqual(t, first.InfraNamespace, second.InfraNamespace)
	require.NotEqual(t, first.Releases[0].Namespace, second.Releases[0].Namespace)
	require.NotEqual(t, first.Releases[0].Databases[0].Name, second.Releases[0].Databases[0].Name)
	require.NotEqual(t, first.Releases[0].Workloads[0].Pods[0].Host, second.Releases[0].Workloads[0].Pods[0].Host)
	require.Equal(t, first.Releases[0].Helm, second.Releases[0].Helm)
	second.Releases = append(second.Releases, second.Releases[0])
	require.ErrorContains(t, validateInventory(second), testDuplicateIdentityError)
}

func TestValidationMapErrorOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*model, *registry)
		want string
	}{
		{
			name: "overrides",
			edit: func(m *model, _ *registry) {
				m.Overrides["z-invalid/pg"] = bounds{Min: 1, Max: 1}
				m.Overrides["a-invalid/pg"] = bounds{Min: 1, Max: 1}
			},
			want: "a-invalid/pg",
		},
		{
			name: "empty products",
			edit: func(_ *model, r *registry) {
				r.Products = append(r.Products, product{ID: "z-empty"}, product{ID: "a-empty"})
			},
			want: "empty product: a-empty",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			tc.edit(&m, &r)
			require.ErrorContains(t, validate(m, r), tc.want)
		})
	}
}

func TestExpandedPodCapacity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		selector string
		edit     func(*model)
		want     string
	}{
		{
			name:     "selected maximum exceeds limit",
			selector: testServiceSMKMSPGSelector,
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: maxExpandedPods + 1}
			},
			want: testExpandedPodCapacityError,
		},
		{
			name:     "unselected maximum does not count",
			selector: testServiceSMKMSMemSelector,
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: int(^uint(0) >> 1)}
			},
		},
		{
			name:     "overflow-sized maximum rejected before allocation",
			selector: testServiceSMKMSPGSelector,
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: int(^uint(0) >> 1)}
			},
			want: testExpandedPodCapacityError,
		},
		{
			name:     "product expansion is preflighted",
			selector: "p:sm:pg",
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: maxExpandedPods + 1}
			},
			want: "expanded pod capacity",
		},
		{
			name:     "suite expansion is preflighted",
			selector: testSuitePGSelector,
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: maxExpandedPods + 1}
			},
			want: testExpandedPodCapacityError,
		},
		{
			name:     "postgres infrastructure capacity is included",
			selector: testServiceSMKMSPGSelector,
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: maxExpandedPods - 3}
			},
			want: testExpandedPodCapacityError,
		},
		{
			name:     "base infrastructure pod capacity is included",
			selector: testServiceSMKMSPGSelector,
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: maxExpandedPods}
			},
			want: testExpandedPodCapacityError,
		},
		{
			name:     "infrastructure maximum is included",
			selector: testServiceSMKMSMemSelector,
			edit: func(m *model) {
				m.Infrastructure.Collector = maxExpandedPods + 1
			},
			want: testExpandedPodCapacityError,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			tc.edit(&m)
			originalOverride := m.Overrides["sm-kms/pg"]
			originalCollector := m.Infrastructure.Collector

			_, err := build(m, r, tc.selector)
			if tc.want != "" {
				require.ErrorContains(t, err, tc.want)
				require.Equal(t, originalOverride, m.Overrides["sm-kms/pg"])
				require.Equal(t, originalCollector, m.Infrastructure.Collector)

				return
			}

			require.NoError(t, err)
			require.Equal(t, originalOverride, m.Overrides["sm-kms/pg"])
			require.Equal(t, originalCollector, m.Infrastructure.Collector)
		})
	}
}

func TestSelectorProjection(t *testing.T) {
	t.Parallel()
	m, r := fixture(t)
	all, err := build(m, r, "")
	require.NoError(t, err)

	for _, rel := range all.Releases {
		t.Run(rel.Key, func(t *testing.T) {
			t.Parallel()

			one, err := build(m, r, rel.Key)
			require.NoError(t, err)
			require.Equal(t, rel, one.Releases[0])
			data, err := json.Marshal(one)
			require.NoError(t, err)
			require.True(t, json.Valid(data))
		})
	}

	_, err = build(m, r, "p:pki:pg")
	require.ErrorContains(t, err, "matched no release")
}

func TestRegistryExpansionAndSQLiteIsolation(t *testing.T) {
	t.Parallel()
	m, r := fixture(t)
	r.Products = append(r.Products, product{ID: testExampleValue})
	r.Members = append(r.Members, member{PSID: "example-service", Product: testExampleValue})
	all, err := build(m, r, "")
	require.NoError(t, err)
	require.Len(t, all.Releases, (8+4+1)*5)

	for _, pool := range []string{poolMem, poolDisk} {
		t.Run(pool, func(t *testing.T) {
			t.Parallel()

			one, err := build(m, r, "s:sm-kms:"+pool)
			require.NoError(t, err)

			w := one.Releases[0].Workloads[0]
			require.NotContains(t, w.Objects, object{"Service", w.Name})
			require.Empty(t, one.Releases[0].Databases)
			require.NotContains(t, one.InfraPods, object{"Pod", "postgres-leader-0"})
			require.NotEqual(t, w.Pods[0].Service, w.Pods[1].Service)
			require.Equal(t, pool == poolDisk, w.Pods[0].PVC != "")
		})
	}
}

func TestExpandedIdentityValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*inventory)
		want string
	}{
		{"infra namespace", func(i *inventory) { i.InfraNamespace = testInvalidKubernetesIdentity }, testDuplicateIdentityError},
		{"infra object", func(i *inventory) { i.Infra[0].Name = testInvalidKubernetesIdentity }, testDuplicateIdentityError},
		{"infra pod", func(i *inventory) { i.InfraPods[0].Name = testInvalidKubernetesIdentity }, testDuplicateIdentityError},
		{"helm release", func(i *inventory) { i.Releases[0].Helm = testInvalidKubernetesIdentity }, "Helm"},
		{"SQL", func(i *inventory) { i.Releases[0].Databases[0].Role = i.Releases[0].Databases[0].Name }, testSQLIdentifierError},
		{"SQL invalid punctuation", func(i *inventory) { i.Releases[0].Databases[0].Name = "db-name" }, testSQLIdentifierError},
		{"SQL leading digit", func(i *inventory) { i.Releases[0].Databases[0].Name = "1database" }, testSQLIdentifierError},
		{"SQL overlong", func(i *inventory) { i.Releases[0].Databases[0].Name = strings.Repeat("a", 64) }, testSQLIdentifierError},
		{"pod DNS", func(i *inventory) {
			i.Releases[0].Workloads[0].Pods[0].FQDN = string(bytes.Repeat([]byte("a"), 254))
		}, testDNSNameError},
		{"pooled DNS", func(i *inventory) {
			i.Releases[0].Workloads[0].ServiceFQDN = string(bytes.Repeat([]byte("a"), 254))
		}, testDNSNameError},
		{"release object", func(i *inventory) { i.Releases[0].Objects[0].Name = testInvalidKubernetesIdentity }, testDuplicateIdentityError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			inv, err := build(m, r, testServiceSMKMSPGSelector)
			require.NoError(t, err)
			tc.edit(&inv)
			require.ErrorContains(t, validateInventory(inv), tc.want)
		})
	}
}

func TestGeneratedSQLIdentifierValidation(t *testing.T) {
	t.Parallel()
	m, r := fixture(t)
	m.Run = "abcdefgh"
	m.Overrides = nil
	m.Modes = []mode{{Name: "bbbbbbbbbb", Pools: []string{poolPG}}}
	id := strings.Repeat("a", 20)
	r.Members = []member{{PSID: id, Product: testProductIdentity}}
	r.Products = []product{{ID: testProductIdentity}}
	r.Suites = []suite{{ID: testSuiteIdentity}}

	_, err := build(m, r, "s:"+id+":bbbbbbbbbb")
	require.ErrorContains(t, err, "invalid or duplicate SQL identifier")
}

func TestRegistryOrder(t *testing.T) {
	t.Parallel()

	r, err := decodeRegistry([]byte("suites: [{id: z}, {id: a}]\nproducts: [{id: z}, {id: a}]\nproduct_services: [{ps_id: z}, {ps_id: a}]\n"))
	require.NoError(t, err)
	require.Equal(t, "a", r.Suites[0].ID)
	require.Equal(t, "a", r.Products[0].ID)
	require.Equal(t, "a", r.Members[0].PSID)
}
