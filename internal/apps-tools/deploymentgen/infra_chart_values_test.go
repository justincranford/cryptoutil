// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestBuildInfrastructureChartValuesMatchesSchema(t *testing.T) {
	t.Parallel()

	defaultsPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.yaml",
	)
	defaultsYAML, err := os.ReadFile(defaultsPath)
	require.NoError(t, err)

	defaults, err := DecodeInfrastructureChartValues(defaultsYAML)
	require.NoError(t, err)

	schemaPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.schema.json",
	)
	schema, err := os.ReadFile(schemaPath)
	require.NoError(t, err)

	model, scopes := fixture(t)
	tests := []struct {
		name                string
		selector            string
		wantClients         int
		wantDatabases       int
		wantPostgres        bool
		wantPostgresClients int
	}{
		{
			name: testFullRegistryMatrixCaseName, wantClients: 55, wantDatabases: 63,
			wantPostgres: true, wantPostgresClients: 33,
		},
		{name: "SQLite-only selected release", selector: testServiceSMKMSMemSelector, wantClients: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			inventory, buildErr := build(model, scopes, test.selector)
			require.NoError(t, buildErr)

			values, projectionErr := BuildInfrastructureChartValues(model, inventory, defaults)
			require.NoError(t, projectionErr)
			require.Equal(t, "cu-"+model.Run+"-infra", values.HelmRelease)
			require.Equal(t, inventory.InfraNamespace, values.Namespace)
			require.Equal(t, model.ClusterDomain, values.ClusterDomain)
			require.Len(t, values.Clients, test.wantClients)
			require.Len(t, values.Databases, test.wantDatabases)
			require.Equal(t, test.wantPostgres, values.Components.PostgresLeader.Enabled)
			require.Equal(t, test.wantPostgres, values.Components.PostgresFollower.Enabled)
			require.True(t, values.Components.Collector.Enabled)
			require.True(t, values.Components.LGTM.Enabled)
			require.Equal(t, model.Infrastructure.Collector, values.Components.Collector.Replicas)

			postgresClients := 0

			for _, client := range values.Clients {
				if client.Postgres {
					postgresClients++
				}
			}

			require.Equal(t, test.wantPostgresClients, postgresClients)

			valuesJSON, marshalErr := MarshalInfrastructureChartValues(values)
			require.NoError(t, marshalErr)
			require.Equal(t, byte('\n'), valuesJSON[len(valuesJSON)-1])
			require.NoError(t, ValidateInfrastructureChartValuesSchema(schema, valuesJSON))

			second, marshalErr := MarshalInfrastructureChartValues(values)
			require.NoError(t, marshalErr)
			require.Equal(t, valuesJSON, second)
		})
	}
}

func TestDecodeInfrastructureChartValuesRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	defaultsPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.yaml",
	)
	defaults, err := os.ReadFile(defaultsPath)
	require.NoError(t, err)

	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{name: "valid canonical values", input: defaults},
		{name: "non-JSON YAML value", input: []byte("run: .nan\n"), want: "encode infrastructure values JSON"},
		{name: "invalid typed value", input: []byte("components: true\n"), want: "decode typed infrastructure values"},
		{
			name:  "unknown root property",
			input: []byte(strings.Replace(string(defaults), "run: review", "run: review\nunknownProperty: true", 1)),
			want:  testUnknownJSONField,
		},
		{name: "malformed YAML", input: []byte("components: ["), want: infrastructureValuesYAMLDecodeError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := DecodeInfrastructureChartValues(test.input)
			if test.want == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestBuildInfrastructureChartValuesRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	defaultsPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.yaml",
	)
	defaultsYAML, err := os.ReadFile(defaultsPath)
	require.NoError(t, err)
	defaults, err := DecodeInfrastructureChartValues(defaultsYAML)
	require.NoError(t, err)

	tests := []struct {
		name     string
		selector string
		edit     func(*Model, *Inventory, *InfrastructureChartValues)
		wantErr  string
	}{
		{
			name: testModelInventoryIdentityMismatch,
			edit: func(_ *Model, inventory *Inventory, _ *InfrastructureChartValues) {
				inventory.Run = "other"
			},
			wantErr: testModelInventoryIdentityMismatch,
		},
		{
			name: testInvalidBootstrapCaseName,
			edit: func(model *Model, _ *Inventory, _ *InfrastructureChartValues) {
				model.Bootstrap.APIHost = testNonLoopbackIPv4
			},
			wantErr: testLoopbackIPv4Error,
		},
		{
			name: testInvalidInventoryCaseName,
			edit: func(_ *Model, inventory *Inventory, _ *InfrastructureChartValues) {
				inventory.Infra = append(inventory.Infra, inventory.Infra[0])
			},
			wantErr: testInvalidInventoryCaseName,
		},
		{
			name: "missing storage class",
			edit: func(_ *Model, _ *Inventory, values *InfrastructureChartValues) {
				values.Storage.ClassName = ""
			},
			wantErr: "infrastructure storage class is required",
		},
		{
			name: "collector replicas differ from canonical chart",
			edit: func(model *Model, _ *Inventory, _ *InfrastructureChartValues) {
				model.Infrastructure.Collector = 1
			},
			wantErr: "do not match canonical infrastructure chart replicas",
		},
		{
			name: "PostgreSQL graph missing leader",
			edit: func(_ *Model, inventory *Inventory, _ *InfrastructureChartValues) {
				inventory.Infra = withoutInventoryObject(
					inventory.Infra, testStatefulSetKind, cryptoutilSharedMagic.PKIInitPostgresLeaderService,
				)
			},
			wantErr: testInfrastructureDependencyMismatch,
		},
		{
			name: "infrastructure graph missing collector",
			edit: func(_ *Model, inventory *Inventory, _ *InfrastructureChartValues) {
				inventory.Infra = withoutInventoryObject(inventory.Infra, testStatefulSetKind, "collector")
			},
			wantErr: testInfrastructureDependencyMismatch,
		},
		{
			name: "infrastructure graph missing LGTM",
			edit: func(_ *Model, inventory *Inventory, _ *InfrastructureChartValues) {
				inventory.Infra = withoutInventoryObject(inventory.Infra, testStatefulSetKind, "lgtm")
			},
			wantErr: testInfrastructureDependencyMismatch,
		},
		{
			name: "component pod count differs",
			edit: func(_ *Model, inventory *Inventory, _ *InfrastructureChartValues) {
				inventory.InfraPods = withoutInventoryObject(inventory.InfraPods, testPodKind, "collector-0")
			},
			wantErr: "infrastructure pod identities disagree with model component replicas",
		},
		{
			name: "PostgreSQL pod count differs",
			edit: func(_ *Model, inventory *Inventory, _ *InfrastructureChartValues) {
				inventory.InfraPods = withoutInventoryObject(inventory.InfraPods, testPodKind, "postgres-follower-0")
			},
			wantErr: testPostgresPodIdentityMismatch,
		},
		{
			name:     "PostgreSQL pod exists for SQLite-only inventory",
			selector: testServiceSMKMSMemSelector,
			edit: func(_ *Model, inventory *Inventory, _ *InfrastructureChartValues) {
				inventory.InfraPods = append(inventory.InfraPods, Object{Kind: testPodKind, Name: testPostgresLeaderPodName})
			},
			wantErr: testPostgresPodIdentityMismatch,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, scopes := fixture(t)
			inventory, buildErr := build(model, scopes, test.selector)
			require.NoError(t, buildErr)

			testDefaults := defaults
			test.edit(&model, &inventory, &testDefaults)

			_, err := BuildInfrastructureChartValues(model, inventory, testDefaults)
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestInfrastructureChartValuesSchemaRejectsInvalidAPIServerAddress(t *testing.T) {
	t.Parallel()

	defaultsPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.yaml",
	)
	defaultsYAML, err := os.ReadFile(defaultsPath)
	require.NoError(t, err)
	values, err := DecodeInfrastructureChartValues(defaultsYAML)
	require.NoError(t, err)

	values.APIServerIP = testNotAnIPAddress

	valuesJSON, err := MarshalInfrastructureChartValues(values)
	require.NoError(t, err)

	schemaPath := filepath.Join(
		"..", "..", "..", "api", "cryptosuite-registry", "templates", "kubernetes",
		"cryptoutil-infra", "values.schema.json",
	)
	schema, err := os.ReadFile(schemaPath)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateInfrastructureChartValuesSchema(schema, valuesJSON), testSchemaMismatchError)
}

func withoutInventoryObject(objects []Object, kind, name string) []Object {
	result := make([]Object, 0, len(objects))
	for _, object := range objects {
		if object.Kind == kind && object.Name == name {
			continue
		}

		result = append(result, object)
	}

	return result
}
