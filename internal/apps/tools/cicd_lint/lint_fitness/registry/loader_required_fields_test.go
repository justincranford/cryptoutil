// Copyright (c) 2025 Justin Cranford
//

package registry_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
)

func TestLoadRegistry_EmptyRequiredFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		yaml         string
		wantContains string
	}{
		{
			name: "empty suite id",
			yaml: `
suites:
  - id: ""
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
`,
			wantContains: "suite",
		},
		{
			name: "empty product display_name",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s/
products:
  - id: ex
    display_name: ""
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
			wantContains: "display_name",
		},
		{
			name: "empty PS magic_file",
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
    magic_file: ""
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: "magic_file",
		},
		{
			name: "empty PS ps_id",
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
  - ps_id: ""
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
			wantContains: "ps_id",
		},
		{
			name: "empty PS product",
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
    product: ""
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
			wantContains: ".product:",
		},
		{
			name: "empty PS service",
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
    service: ""
    display_name: "Example Service"
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: ".service:",
		},
		{
			name: "empty PS display_name",
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
    display_name: ""
    internal_apps_dir: ex-svc/
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: ".display_name:",
		},
		{
			name: "empty PS internal_apps_dir",
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
    internal_apps_dir: ""
    magic_file: magic_ex.go
    base_port: 8000
    pg_host_port: 54320
    migration_range_start: 2001
    migration_range_end: 2999
    api_resources: []
`,
			wantContains: ".internal_apps_dir:",
		},
		{
			name: "PS internal_apps_dir missing trailing slash",
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
    internal_apps_dir: ex-svc
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

func TestLoadRegistry_SuiteFieldValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		yaml         string
		wantContains string
	}{
		{
			name: "empty display_name",
			yaml: `
suites:
  - id: s
    display_name: ""
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
`,
			wantContains: "display_name",
		},
		{
			name: "empty cmd_dir",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: ""
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
			wantContains: "cmd_dir",
		},
		{
			name: "cmd_dir missing trailing slash",
			yaml: `
suites:
  - id: s
    display_name: "S"
    cmd_dir: s
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
