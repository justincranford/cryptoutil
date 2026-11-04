// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testLifecycleFieldX             = "x"
	testLifecycleFieldY             = "y"
	testLifecycleUnknownObject      = "unknown or duplicate"
	testLifecycleInitCommand        = "init"
	testLifecycleSigningError       = "signing or application"
	testLifecycleOverwriteError     = "cannot overwrite"
	testLifecycleSecretKind         = cryptoutilSharedMagic.KubernetesSecretKind
	testLifecycleRoleKind           = "Role"
	testLifecycleRoleBindingKind    = "RoleBinding"
	testLifecycleServiceAccountKind = "ServiceAccount"
	testLifecycleResourcesField     = "resources"
	testLifecycleReservedOrdinals   = 3
)

func TestLifecycleStrictDecodeAndFieldStructure(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		testOpenJSONObject, testNullDocument, `{"version":"one"}`, `{"version":1,"version":1}`,
		`{"version":1,"extra":true}`, `{"version":1,"secrets":[],"jobs":[{"Name":"wrong"}]}`,
	} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			_, err := DecodeLifecycleInput([]byte(raw))
			require.Error(t, err)
		})
	}

	tests := []struct {
		name        string
		left, right any
		want        bool
	}{
		{"matching scalar", testLifecycleFieldX, testLifecycleFieldX, true},
		{"different scalar", testLifecycleFieldX, testLifecycleFieldY, false},
		{"scalar versus object", nil, map[string]any{}, false},
		{"object versus scalar", map[string]any{}, nil, false},
		{"object length", map[string]any{testLifecycleFieldX: 1}, map[string]any{}, false},
		{"object field", map[string]any{testLifecycleFieldX: 1}, map[string]any{testLifecycleFieldY: 1}, false},
		{"array versus scalar", []any{}, nil, false},
		{"array length", []any{1}, []any{}, false},
		{"array mismatch", []any{map[string]any{testLifecycleFieldX: 1}}, []any{map[string]any{testLifecycleFieldY: 1}}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.want, lifecycleFieldsMatch(test.left, test.right))
		})
	}
}

func TestLifecycleJobPermissionBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*LifecycleInput)
		want string
	}{
		{"unknown Job", func(v *LifecycleInput) { v.Jobs[0].Name = testLifecycleUnknownName }, testLifecycleUnknownObject},
		{"duplicate Job", func(v *LifecycleInput) { v.Jobs[1] = v.Jobs[0] }, testLifecycleUnknownObject},
		{"blank argument", func(v *LifecycleInput) { v.Jobs[0].Args = []string{" "} }, "empty argument"},
		{"database PKI invocation", func(v *LifecycleInput) { v.Jobs[0].Args = []string{lifecyclePKICommand} }, "must not invoke PKI"},
		{"database writes", func(v *LifecycleInput) { v.Jobs[0].WriteSecretNames = []string{lifecyclePostgresBootstrap} }, "must not have Secret write"},
		{"database signer access", func(v *LifecycleInput) { v.Jobs[0].SecretNames = []string{lifecycleAuthoritySecret} }, testLifecycleSigningError},
		{"PKI command mismatch", func(v *LifecycleInput) { v.Jobs[1].Args = []string{testLifecycleInitCommand} }, "suite pki-init"},
		{"PKI credential overwrite", func(v *LifecycleInput) { v.Jobs[1].WriteSecretNames = []string{lifecyclePostgresBootstrap} }, testLifecycleOverwriteError},
		{"cross-member migration access", func(v *LifecycleInput) {
			for index := range v.Jobs {
				if strings.HasSuffix(v.Jobs[index].Name, "-migrate") {
					v.Jobs[index].SecretNames = []string{"sm-kms-db"}
				}
			}
		}, "only its member"},
		{"missing consumer key", func(v *LifecycleInput) { delete(v.Secrets[0].Files, "ca.crt") }, "consumer key"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, _, input := lifecycleDeployFixture(t, testServiceSMKMSPGSelector)
			test.edit(&input)

			var err error

			model.Revision, err = LifecycleRevision(input)
			require.NoError(t, err)
			inventory, err := build(model, RegisteredScopes(), testServiceSMKMSPGSelector)
			require.NoError(t, err)
			_, err = BuildLifecycleChartValues(model, inventory, input)
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestLifecycleModelAndSchemaBoundaries(t *testing.T) {
	t.Parallel()

	model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
	model.Run = testDifferentIdentity
	_, err := BuildLifecycleChartValues(model, inventory, input)
	require.ErrorContains(t, err, "identities differ")

	model.Run = inventory.Run
	model.Bootstrap.APIHost = testNotAnIPAddress
	_, err = BuildLifecycleChartValues(model, inventory, input)
	require.Error(t, err)
	model, inventory, input = lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
	inventory.InfraNamespace = ""
	_, err = BuildLifecycleChartValues(model, inventory, input)
	require.Error(t, err)

	model, inventory, input = lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
	model.Bootstrap.APIServerAudience = "explicit-api-audience"
	values, err := BuildLifecycleChartValues(model, inventory, input)
	require.NoError(t, err)
	require.Equal(t, model.Bootstrap.APIServerAudience, values.APIAudience)
	raw, err := json.Marshal(values)
	require.NoError(t, err)
	schema, err := os.ReadFile(filepath.Join("..", "..", "..", lifecycleValuesSchemaPath))
	require.NoError(t, err)
	require.NoError(t, ValidateLifecycleChartValuesSchema(schema, raw))
	require.Error(t, ValidateLifecycleChartValuesSchema(schema, []byte(`{"jobs":[]}`)))
}
