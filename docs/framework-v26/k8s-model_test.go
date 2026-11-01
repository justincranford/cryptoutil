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

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

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
		{"bad run", func(m *model, _ *registry) { m.Run = "Upper" }, "invalid version"},
		{"bad version", func(m *model, _ *registry) { m.Version = 2 }, "invalid version"},
		{"bad domain", func(m *model, _ *registry) { m.ClusterDomain = "cluster..local" }, "invalid domain"},
		{"domain label too long", func(m *model, _ *registry) { m.ClusterDomain = strings.Repeat("a", 64) }, "invalid domain"},
		{"long domain suffix", func(m *model, _ *registry) { m.ClusterDomain = strings.Repeat("a", 254) }, "domain too long"},
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
		{"empty product", func(_ *model, r *registry) { r.Products = append(r.Products, product{ID: "extra"}) }, "empty product"},
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
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "zero-length input", input: "", want: "model:"},
		{name: "truncated object", input: `{`, want: "model:"},
		{name: "unknown field", input: `{"unexpected": true}`, want: "unknown field"},
		{name: "wrong field type", input: `{"version": "one"}`, want: "model:"},
		{name: "null", input: `null`, want: "top-level JSON object"},
		{name: "array", input: `[]`, want: "top-level JSON object"},
		{name: "string", input: `"value"`, want: "top-level JSON object"},
		{name: "duplicate top-level key", input: `{"version": 1, "version": 1}`, want: "duplicate JSON object key"},
		{name: "duplicate nested key", input: `{"overrides": {"sm-kms/pg": {"min": 1, "min": 2}}}`, want: "duplicate JSON object key"},
		{name: "invalid array value", input: `{"value": [1,]}`, want: "model:"},
		{name: "truncated array value", input: `{"value": [1`, want: "model:"},
		{name: "truncated object value", input: `{"value": 1`, want: "model:"},
		{name: "missing object value", input: `{"value":`, want: "model:"},
		{name: "mismatched object delimiter", input: `{"value": 1]`, want: "model:"},
		{name: "truncated nested array", input: `{"value": [1`, want: "model:"},
		{name: "mismatched array delimiter", input: `{"value": [1}`, want: "model:"},
		{name: "nested array value error", input: `{"value": [[1,]]}`, want: "model:"},
		{name: "unexpected array value delimiter", input: `{"value": ]}`, want: "model:"},
		{name: "multiple documents", input: `{} {}`, want: "exactly one"},
		{name: "trailing content", input: `{} trailing`, want: "exactly one"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := decodeModel([]byte(tc.input))
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestRegistryDecodingStrictness(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "malformed YAML", input: "suites: [\n", want: "registry:"},
		{name: "null document", input: "null\n", want: "non-null YAML mapping"},
		{name: "sequence document", input: "- suites\n", want: "non-null YAML mapping"},
		{name: "scalar document", input: "registry\n", want: "non-null YAML mapping"},
		{name: "duplicate top-level key", input: "suites: []\nsuites: []\n", want: "duplicate YAML mapping key"},
		{name: "duplicate nested key", input: "suites: [{id: first, id: second}]\n", want: "duplicate YAML mapping key"},
		{name: "unknown top-level field", input: "unexpected: true\n", want: "field unexpected not found"},
		{name: "unknown nested field", input: "suites: [{id: first, unexpected: true}]\n", want: "field unexpected not found"},
		{name: "field from another mapping type", input: "suites: [{id: first, internal_apps_dir: first/}]\n", want: "field internal_apps_dir not found"},
		{name: "multiple documents", input: "suites: []\n---\nproducts: []\n", want: "exactly one YAML document"},
		{name: "malformed second document", input: "suites: []\n---\n[\n", want: "exactly one YAML document"},
		{name: "single mapping", input: "suites: []\nproducts: []\nproduct_services: []\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := decodeRegistry([]byte(tc.input))
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.want)
		})
	}
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

func TestExpandedDNSNameBoundary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		extraByte   int
		domainBytes int
		wantError   string
		wantDomain  bool
	}{
		{name: "full name at DNS limit"},
		{name: "expanded name exceeds limit", extraByte: 1, wantError: "invalid DNS name"},
		{name: "suffix above legacy cutoff remains valid", domainBytes: 201, wantError: "invalid DNS name", wantDomain: true},
		{name: "maximum suffix length is permitted", domainBytes: 253, wantError: "invalid DNS name", wantDomain: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m, r := fixture(t)
			prefix := "sm-kms-pg-0.sm-kms-pg-pods.cu-" + m.Run + "-s-sm-kms-pg.svc."
			domainLength := 253 - len(prefix) + tc.extraByte
			if tc.domainBytes > 0 {
				domainLength = tc.domainBytes
			}
			var labels []string
			for remaining := domainLength; remaining > 0; {
				if len(labels) > 0 {
					remaining--
				}
				labelLength := min(63, remaining)
				labels = append(labels, "1"+strings.Repeat("a", labelLength-1))
				remaining -= labelLength
			}
			m.ClusterDomain = strings.Join(labels, ".")
			if tc.wantDomain {
				require.Greater(t, len(m.ClusterDomain), 200)
				require.NoError(t, validate(m, r))
			}

			inv, err := build(m, r, "s:sm-kms:pg")
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, 253, len(inv.Releases[0].Workloads[0].Pods[0].FQDN))
			require.True(t, dnsLabel(strings.Split(m.ClusterDomain, ".")[0]))
		})
	}
}

func TestDNSValidationHelpers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		label       string
		dnsName     string
		wantLabel   bool
		wantDNSName bool
	}{
		{name: "numeric label", label: "123", dnsName: "123.example", wantLabel: true, wantDNSName: true},
		{name: "zero-length name", dnsName: "", wantDNSName: false},
		{name: "missing domain label", label: "", dnsName: "example..local", wantDNSName: false},
		{name: "label too long", label: strings.Repeat("a", cryptoutilSharedMagic.FQDNLabelMaxLength+1), dnsName: "example.local", wantDNSName: true},
		{name: "name too long", label: "example", dnsName: strings.Repeat("a", cryptoutilSharedMagic.FQDNMaxLength+1), wantLabel: true, wantDNSName: false},
		{name: "leading hyphen", label: "-example", dnsName: "-example.local", wantDNSName: false},
		{name: "trailing hyphen", label: "example-", dnsName: "example-.local", wantDNSName: false},
		{name: "uppercase", label: "Example", dnsName: "Example.local", wantDNSName: false},
		{name: "invalid character", label: "bad_name", dnsName: "bad_name.local", wantDNSName: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantLabel, dnsLabel(tc.label))
			require.Equal(t, tc.wantDNSName, dnsName(tc.dnsName))
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
			selector: "s:sm-kms:pg",
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: maxExpandedPods + 1}
			},
			want: "expanded pod capacity",
		},
		{
			name:     "unselected maximum does not count",
			selector: "s:sm-kms:mem",
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: int(^uint(0) >> 1)}
			},
		},
		{
			name:     "overflow-sized maximum rejected before allocation",
			selector: "s:sm-kms:pg",
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: int(^uint(0) >> 1)}
			},
			want: "expanded pod capacity",
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
			selector: "u:cryptoutil:pg",
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: maxExpandedPods + 1}
			},
			want: "expanded pod capacity",
		},
		{
			name:     "postgres infrastructure capacity is included",
			selector: "s:sm-kms:pg",
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: maxExpandedPods - 3}
			},
			want: "expanded pod capacity",
		},
		{
			name:     "base infrastructure pod capacity is included",
			selector: "s:sm-kms:pg",
			edit: func(m *model) {
				m.Overrides["sm-kms/pg"] = bounds{Min: 1, Max: maxExpandedPods}
			},
			want: "expanded pod capacity",
		},
		{
			name:     "infrastructure maximum is included",
			selector: "s:sm-kms:mem",
			edit: func(m *model) {
				m.Infrastructure.Collector = maxExpandedPods + 1
			},
			want: "expanded pod capacity",
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
	r.Products = append(r.Products, product{ID: "example"})
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
		{"SQL invalid punctuation", func(i *inventory) { i.Releases[0].Databases[0].Name = "db-name" }, "SQL identifier"},
		{"SQL leading digit", func(i *inventory) { i.Releases[0].Databases[0].Name = "1database" }, "SQL identifier"},
		{"SQL overlong", func(i *inventory) { i.Releases[0].Databases[0].Name = strings.Repeat("a", 64) }, "SQL identifier"},
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

func TestGeneratedSQLIdentifierValidation(t *testing.T) {
	t.Parallel()
	m, r := fixture(t)
	m.Run = "abcdefgh"
	m.Overrides = nil
	m.Modes = []mode{{Name: "bbbbbbbbbb", Pools: []string{"pg"}}}
	id := strings.Repeat("a", 20)
	r.Members = []member{{PSID: id, Product: "product"}}
	r.Products = []product{{ID: "product"}}
	r.Suites = []suite{{ID: "suite"}}

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
