// Copyright (c) 2025 Justin Cranford
//

package registry_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

func TestLoadRegistry_ProductFieldValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		yaml         string
		wantContains string
	}{
		{
			name: "empty product ID",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ""
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
			wantContains: ".id:",
		},
		{
			name: "empty product internal_apps_dir",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ""
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
			wantContains: "internal_apps_dir",
		},
		{
			name: "empty product cmd_dir",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ""
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
			wantContains: "cmd_dir",
		},
		{
			name: "product cmd_dir missing trailing slash",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex/
    cmd_dir: ex
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
			wantContains: "must end with '/'",
		},
		{
			name: "product internal_apps_dir missing trailing slash",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: "Example"
    internal_apps_dir: ex
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
			wantContains: "must end with '/'",
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

func TestLoadRegistry_PSFormatValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		yaml         string
		wantContains string
	}{
		{
			name: "PS ID not matching product-service",
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
  - ps_id: wrong-name
    product: ex
    service: svc
    display_name: "Example Service"
    internal_apps_dir: wrong-name/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: `expected "ex-svc"`,
		},
		{
			name: "internal_apps_dir not matching PS ID",
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
    internal_apps_dir: wrong-dir/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: `expected "ex-svc/"`,
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

func TestLoadRegistry_PSPortValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		basePort     string
		pgHostPort   string
		wantContains string
	}{
		{name: "negative base_port", basePort: "-1", pgHostPort: "54320", wantContains: "base_port"},
		{name: "zero base_port", basePort: "0", pgHostPort: "54320", wantContains: "base_port"},
		{name: "negative pg_host_port", basePort: "8000", pgHostPort: "-1", wantContains: "pg_host_port"},
		{name: "zero pg_host_port", basePort: "8000", pgHostPort: "0", wantContains: "pg_host_port"},
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
    base_port: ` + tc.basePort + `
    pg_host_port: ` + tc.pgHostPort + `
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`

			path := writeRegistryYAML(t, yaml)
			_, err := lintFitnessRegistry.LoadRegistry(path)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantContains)
		})
	}
}
