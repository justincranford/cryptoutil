// Copyright (c) 2025 Justin Cranford
//

package registry_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

func TestLoadRegistry_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := lintFitnessRegistry.LoadRegistry("/nonexistent/path/that/does/not/exist/registry.yaml")
	require.Error(t, err, "LoadRegistry must fail when file does not exist")
	require.Contains(t, err.Error(), "read registry")
}

func TestLoadRegistry_InvalidYAMLSyntax(t *testing.T) {
	t.Parallel()

	path := writeRegistryYAML(t, "suites: [\n  bad: yaml\n  unclosed:")
	_, err := lintFitnessRegistry.LoadRegistry(path)
	require.Error(t, err, "LoadRegistry must fail on invalid YAML syntax")
}

func TestLoadRegistry_EmptyCollections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		yaml         string
		wantContains string
	}{
		{
			name: "empty suites",
			yaml: `
suites: []
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: "suites: must contain at least one entry",
		},
		{
			name: "empty products",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products: []
product_services:
  - ps_id: s-svc
    product: s
    service: svc
    display_name: "S Service"
    internal_apps_dir: s-svc/
    magic_file: magic_s.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: "products: must contain at least one entry",
		},
		{
			name: "empty product services",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services: []
`,
			wantContains: "product_services: must contain at least one entry",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := writeRegistryYAML(t, tc.yaml)
			_, err := lintFitnessRegistry.LoadRegistry(path)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantContains)
		})
	}
}

func TestLoadRegistry_DuplicateIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		yaml         string
		wantContains string
	}{
		{
			name: "duplicate suite ID",
			yaml: `
suites:
  - id: dup
    display_name: "Dup Suite"
    cmd_dir: dup/
  - id: dup
    display_name: "Dup Suite 2"
    cmd_dir: dup/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: "duplicate id dup",
		},
		{
			name: "duplicate product ID",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
  - id: ex
    display_name: "Example Dup"
    internal_apps_dir: ex2/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: "duplicate id ex",
		},
		{
			name: "duplicate PS ID",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service Dup"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8001
    pg_host_port: 54321
    migration_range_start: 3001
    migration_range_end: 3999
    api_resources: []
`,
			wantContains: "duplicate ps_id ex-svc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := writeRegistryYAML(t, tc.yaml)
			_, err := lintFitnessRegistry.LoadRegistry(path)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantContains)
		})
	}
}

func TestLoadRegistry_OverlappingPorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		yaml         string
		wantContains string
	}{
		{
			name: "overlapping base ports",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
  - ps_id: ex-alt
    product: ex
    service: alt
    display_name: "Example Alt"
    internal_apps_dir: ex-alt/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54321
    migration_range_start: 3001
    migration_range_end: 3999
    api_resources: []
`,
			wantContains: "8000 already assigned to ex-svc",
		},
		{
			name: "overlapping PG ports",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
  - ps_id: ex-alt
    product: ex
    service: alt
    display_name: "Example Alt"
    internal_apps_dir: ex-alt/
    magic_file: magic_ex.go
    base_port: 8001
    pg_host_port: 54320
    migration_range_start: 3001
    migration_range_end: 3999
    api_resources: []
`,
			wantContains: "54320 already assigned to ex-svc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := writeRegistryYAML(t, tc.yaml)
			_, err := lintFitnessRegistry.LoadRegistry(path)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantContains)
		})
	}
}

func TestLoadRegistry_OverlappingMigrationRanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		start1 int
		end1   int
		start2 int
		end2   int
	}{
		{
			name:   "large overlap",
			start1: 2001,
			end1:   3999,
			start2: 2999,
			end2:   4999,
		},
		{
			name:   "exact boundary overlap start equals end",
			start1: 2001,
			end1:   2999,
			start2: 2999,
			end2:   4001,
		},
		{
			name:   "exact boundary overlap end equals start",
			start1: 2999,
			end1:   4001,
			start2: 2001,
			end2:   2999,
		},
		{
			name:   "single point overlap",
			start1: 2001,
			end1:   2500,
			start2: 2500,
			end2:   migrationRangeTestEnd,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: ` + strconv.Itoa(tc.start1) + `
    migration_range_end: ` + strconv.Itoa(tc.end1) + `
    api_resources: []
  - ps_id: ex-alt
    product: ex
    service: alt
    display_name: "Example Alt"
    internal_apps_dir: ex-alt/
    magic_file: magic_ex.go
    base_port: 8001
    pg_host_port: 54321
    migration_range_start: ` + strconv.Itoa(tc.start2) + `
    migration_range_end: ` + strconv.Itoa(tc.end2) + `
    api_resources: []
`

			path := writeRegistryYAML(t, yaml)
			_, err := lintFitnessRegistry.LoadRegistry(path)
			require.Error(t, err)
			require.Contains(t, err.Error(), "overlaps with")
		})
	}
}

func TestLoadRegistry_MigrationRangeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		start        int
		end          int
		wantContains string
	}{
		{name: "end equals start", start: 2001, end: 2001, wantContains: "migration_range_end"},
		{name: "end less than start", start: migrationRangeTestEnd, end: 2001, wantContains: "migration_range_end"},
		{name: "start below minimum", start: migrationRangeTestBelowMinStart, end: 999, wantContains: "migration_range_start"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			yaml := `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex/
product_services:
  - ps_id: ex-svc
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: ` + strconv.Itoa(tc.start) + `
    migration_range_end: ` + strconv.Itoa(tc.end) + `
    api_resources: []
`

			path := writeRegistryYAML(t, yaml)
			_, err := lintFitnessRegistry.LoadRegistry(path)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantContains)
		})
	}
}
