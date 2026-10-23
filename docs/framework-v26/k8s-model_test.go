//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.

package main

import (
	"bytes"
	json "encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func fixture(t *testing.T) (model, registry) {
	t.Helper()
	mBytes, err := os.ReadFile("k8s-model.json")
	require.NoError(t, err)
	m, err := decodeModel(mBytes)
	require.NoError(t, err)
	rBytes, err := os.ReadFile(filepath.Join("..", "..", "api", "cryptosuite-registry", "registry.yaml"))
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
		{"u:cryptoutil:pg", 1, 7, 14, 7},
		{"p:sm:mem", 1, 1, 2, 0},
		{"p:skeleton:disk", 1, 1, 2, 0},
	}
	for _, tc := range tests {
		t.Run("selector="+tc.selector, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			inv, err := build(m, r, tc.selector)
			require.NoError(t, err)
			require.Len(t, inv.Releases, tc.releases)
			pools, pods, dbs := 0, 0, 0
			for _, rel := range inv.Releases {
				dbs += len(rel.Databases)
				pools += len(rel.Workloads)
				for _, w := range rel.Workloads {
					pods += w.Bounds.Min
					require.Len(t, w.Pods, w.Bounds.Max)
				}
			}
			require.Equal(t, tc.pools, pools)
			require.Equal(t, tc.pods, pods)
			require.Equal(t, tc.dbs, dbs)
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
		{"bad run", func(m *model, _ *registry) { m.Run = "Upper" }, "invalid version"},
		{"bad version", func(m *model, _ *registry) { m.Version = 2 }, "invalid version"},
		{"bad domain", func(m *model, _ *registry) { m.ClusterDomain = "cluster..local" }, "invalid domain"},
		{"long domain", func(m *model, _ *registry) { m.ClusterDomain = string(bytes.Repeat([]byte("a"), 201)) }, "domain too long"},
		{"bad bounds", func(m *model, _ *registry) { m.Replicas = bounds{3, 2} }, "replicas"},
		{"zero bounds", func(m *model, _ *registry) { m.Replicas = bounds{0, 2} }, "replicas"},
		{"unknown override", func(m *model, _ *registry) { m.Overrides["missing/pg"] = bounds{2, 2} }, "override"},
		{"bad override", func(m *model, _ *registry) { m.Overrides["sm-kms/pg"] = bounds{2, 1} }, "replicas"},
		{"unknown pool", func(m *model, _ *registry) { m.Modes[0].Pools = []string{"not-a-pool"} }, "pool"},
		{"duplicate pool", func(m *model, _ *registry) { m.Modes[0].Pools = []string{"pg", "pg"} }, "pool"},
		{"duplicate mode", func(m *model, _ *registry) { m.Modes = append(m.Modes, m.Modes[0]) }, "mode"},
		{"empty modes", func(m *model, _ *registry) { m.Modes = nil }, "nonempty"},
		{"empty pool list", func(m *model, _ *registry) { m.Modes[0].Pools = nil }, "empty mode"},
		{"orphan service", func(_ *model, r *registry) { r.Members[0].Product = "missing" }, "PS-ID"},
		{"duplicate service", func(_ *model, r *registry) { r.Members = append(r.Members, r.Members[0]) }, "PS-ID"},
		{"duplicate product", func(_ *model, r *registry) { r.Products = append(r.Products, r.Products[0]) }, "product"},
		{"duplicate suite", func(_ *model, r *registry) { r.Suites = append(r.Suites, r.Suites[0]) }, "suite"},
		{"empty registry", func(_ *model, r *registry) { r.Members = nil }, "nonempty"},
		{"empty product", func(_ *model, r *registry) { r.Products = append(r.Products, entity{"extra"}) }, "empty product"},
		{"two primaries", func(m *model, _ *registry) { m.Infrastructure.Leader = 2 }, "singletons"},
		{"long names", func(m *model, _ *registry) { m.Revision = "r" + string(bytes.Repeat([]byte("x"), 62)) }, "duplicate identity"},
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
	m.Overrides["sm-kms/pg"] = bounds{1, 3}
	first, err := build(m, r, "s:sm-kms:pg")
	require.NoError(t, err)
	w := first.Releases[0].Workloads[0]
	require.Equal(t, bounds{1, 3}, w.Bounds)
	require.Len(t, w.Pods, 3)
	require.True(t, w.Pods[0].Active)
	require.False(t, w.Pods[1].Active)
	require.Equal(t, "sm-kms-postgresql-1", w.Pods[0].LegacySlot)
	require.Equal(t, "sm-kms-pg-0", w.Pods[0].Name)
	require.Contains(t, string(render(m, r, first)), "reserved")
	m.Run = "other"
	second, err := build(m, r, "s:sm-kms:pg")
	require.NoError(t, err)
	require.NotEqual(t, first.InfraNamespace, second.InfraNamespace)
	require.NotEqual(t, first.Releases[0].Namespace, second.Releases[0].Namespace)
	require.NotEqual(t, first.Releases[0].Databases[0].Name, second.Releases[0].Databases[0].Name)
	require.NotEqual(t, first.Releases[0].Workloads[0].Pods[0].Host, second.Releases[0].Workloads[0].Pods[0].Host)
	require.Equal(t, first.Releases[0].Helm, second.Releases[0].Helm)
	second.Releases = append(second.Releases, second.Releases[0])
	require.ErrorContains(t, validateInventory(second), "duplicate identity")
}

func TestModelDecoding(t *testing.T) {
	t.Parallel()
	tests := []struct{ input, want string }{
		{`{`, "model:"},
		{`{"unexpected": true}`, "unknown field"},
		{`{} {}`, "exactly one"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			_, err := decodeModel([]byte(tc.input))
			require.ErrorContains(t, err, tc.want)
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
	r.Products = append(r.Products, entity{"example"})
	r.Members = append(r.Members, member{PSID: "example-service", Product: "example"})
	all, err := build(m, r, "")
	require.NoError(t, err)
	require.Len(t, all.Releases, (8+4+1)*5)
	for _, pool := range []string{"mem", "disk"} {
		t.Run(pool, func(t *testing.T) {
			t.Parallel()
			one, err := build(m, r, "s:sm-kms:"+pool)
			require.NoError(t, err)
			w := one.Releases[0].Workloads[0]
			require.NotContains(t, w.Objects, object{"Service", w.Name})
			require.Empty(t, one.Releases[0].Databases)
			require.NotContains(t, one.InfraPods, object{"Pod", "postgres-leader-0"})
			require.NotEqual(t, w.Pods[0].Service, w.Pods[1].Service)
			require.Equal(t, pool == "disk", w.Pods[0].PVC != "")
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
		{"infra namespace", func(i *inventory) { i.InfraNamespace = "Invalid" }, "duplicate identity"},
		{"infra object", func(i *inventory) { i.Infra[0].Name = "Invalid" }, "duplicate identity"},
		{"infra pod", func(i *inventory) { i.InfraPods[0].Name = "Invalid" }, "duplicate identity"},
		{"helm release", func(i *inventory) { i.Releases[0].Helm = "Invalid" }, "Helm"},
		{"SQL", func(i *inventory) { i.Releases[0].Databases[0].Role = i.Releases[0].Databases[0].Name }, "SQL identifier"},
		{"pod DNS", func(i *inventory) {
			i.Releases[0].Workloads[0].Pods[0].FQDN = string(bytes.Repeat([]byte("a"), 254))
		}, "DNS name"},
		{"pooled DNS", func(i *inventory) {
			i.Releases[0].Workloads[0].ServiceFQDN = string(bytes.Repeat([]byte("a"), 254))
		}, "DNS name"},
		{"release object", func(i *inventory) { i.Releases[0].Objects[0].Name = "Invalid" }, "duplicate identity"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			inv, err := build(m, r, "s:sm-kms:pg")
			require.NoError(t, err)
			tc.edit(&inv)
			require.ErrorContains(t, validateInventory(inv), tc.want)
		})
	}
}

func TestRegistryOrder(t *testing.T) {
	t.Parallel()
	r, err := decodeRegistry([]byte("suites: [{id: z}, {id: a}]\nproducts: [{id: z}, {id: a}]\nproduct_services: [{ps_id: z}, {ps_id: a}]\n"))
	require.NoError(t, err)
	require.Equal(t, "a", r.Suites[0].ID)
	require.Equal(t, "a", r.Products[0].ID)
	require.Equal(t, "a", r.Members[0].PSID)
}
