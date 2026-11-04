// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"errors"
	"os"
	"slices"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestBuildLifecycleInventory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		selector    string
		wantOwners  map[string]int
		wantPresent []LifecycleIdentity
	}{
		{
			name:     "PostgreSQL release and shared infrastructure",
			selector: testServiceSMKMSPGSelector,
			wantOwners: map[string]int{
				lifecycleExternalSecretIdentity: 26,
				lifecycleJobIdentity:            5,
				lifecycleBootstrapRBACIdentity:  6,
			},
			wantPresent: []LifecycleIdentity{
				{
					Namespace: testLifecyclePostgresNamespace, Kind: cryptoutilSharedMagic.KubernetesSecretKind,
					Name: "sm-kms-migration-db", Owner: lifecycleExternalSecretIdentity,
				},
				{
					Namespace: testLifecyclePostgresNamespace, Kind: testLifecycleJobKind,
					Name: "sm-kms-migrate-r1", Owner: lifecycleJobIdentity,
				},
				{
					Namespace: testLifecyclePostgresNamespace, Kind: testLifecycleRoleBindingKind,
					Name: testLifecycleBootstrapResourceName, Owner: lifecycleBootstrapRBACIdentity,
				},
				{
					Namespace: testLifecycleInfrastructureNamespace, Kind: testLifecycleJobKind,
					Name: "replication-r1", Owner: lifecycleJobIdentity,
				},
				{
					Namespace: testLifecycleInfrastructureNamespace, Kind: cryptoutilSharedMagic.KubernetesSecretKind,
					Name: lifecyclePostgresBootstrap, Owner: lifecycleExternalSecretIdentity,
				},
			},
		},
		{
			name:     "SQLite-only release and infrastructure",
			selector: testServiceSMKMSMemSelector,
			wantOwners: map[string]int{
				lifecycleExternalSecretIdentity: 19,
				lifecycleJobIdentity:            2,
				lifecycleBootstrapRBACIdentity:  6,
			},
			wantPresent: []LifecycleIdentity{
				{
					Namespace: "cu-review-s-sm-kms-mem", Kind: testLifecycleJobKind,
					Name: "pki-init-r1", Owner: lifecycleJobIdentity,
				},
				{
					Namespace: testLifecycleInfrastructureNamespace, Kind: cryptoutilSharedMagic.KubernetesSecretKind,
					Name: lifecycleAuthoritySecret, Owner: lifecycleExternalSecretIdentity,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, scopes := fixture(t)
			inventory, err := build(model, scopes, test.selector)
			require.NoError(t, err)

			lifecycle, err := BuildLifecycleInventory(inventory)
			require.NoError(t, err)
			require.Equal(t, model.Version, lifecycle.Version)
			require.Equal(t, model.Run, lifecycle.Run)

			owners := make(map[string]int)
			for _, identity := range lifecycle.Identities {
				owners[identity.Owner]++
			}

			require.Equal(t, test.wantOwners, owners)
			require.True(t, slices.IsSortedFunc(lifecycle.Identities, func(left, right LifecycleIdentity) int {
				if order := compareString(left.Namespace, right.Namespace); order != 0 {
					return order
				}

				if order := compareString(left.Kind, right.Kind); order != 0 {
					return order
				}

				if order := compareString(left.Name, right.Name); order != 0 {
					return order
				}

				return compareString(left.Owner, right.Owner)
			}))

			for _, expected := range test.wantPresent {
				require.Contains(t, lifecycle.Identities, expected)
			}
		})
	}
}

func TestBuildLifecycleInventoryRejectsInvalidGraph(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, testServiceSMKMSPGSelector)
	require.NoError(t, err)

	inventory.InfraNamespace = ""

	_, err = BuildLifecycleInventory(inventory)
	require.ErrorContains(t, err, "invalid inventory")
}

func TestGeneratorCLILifecycleIdentities(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	modelJSON, err := json.Marshal(model)
	require.NoError(t, err)

	inventory, err := build(model, scopes, testServiceSMKMSPGSelector)
	require.NoError(t, err)
	expected, err := BuildLifecycleInventory(inventory)
	require.NoError(t, err)
	expectedJSON, err := marshalCanonicalJSON(expected)
	require.NoError(t, err)

	tests := []struct {
		name        string
		args        []string
		wantCode    int
		wantOutput  []byte
		wantMessage string
	}{
		{
			name:       "preview identities for selected release",
			args:       []string{cliFormatFlag, lifecycleIdentitiesFormat, cliSelectFlag, testServiceSMKMSPGSelector},
			wantOutput: expectedJSON,
		},
		{
			name:        "write is not an artifact mode",
			args:        []string{cliWriteFlag, cliFormatFlag, lifecycleIdentitiesFormat},
			wantCode:    2,
			wantMessage: cliInvalidFlagsMessage,
		},
		{
			name:        "check is not an artifact mode",
			args:        []string{cliCheckFlag, cliFormatFlag, lifecycleIdentitiesFormat},
			wantCode:    2,
			wantMessage: cliInvalidFlagsMessage,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			readFile := func(path string) ([]byte, error) {
				if path == defaultPlannerPaths().model {
					return modelJSON, nil
				}

				return nil, errors.New("unexpected read")
			}

			writeFile := func(string, []byte, os.FileMode) error {
				t.Fatal("lifecycle identities must not write artifacts")

				return nil
			}

			var stdout, stderr bytes.Buffer

			code := internalMain(test.args, &stdout, &stderr, readFile, writeFile)
			require.Equal(t, test.wantCode, code, stderr.String())

			if test.wantOutput != nil {
				require.Equal(t, test.wantOutput, stdout.Bytes())
			}

			if test.wantMessage != "" {
				require.Contains(t, stderr.String(), test.wantMessage)
			}
		})
	}
}

func TestLifecycleInventoryIsDeterministic(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, "")
	require.NoError(t, err)

	first, err := BuildLifecycleInventory(inventory)
	require.NoError(t, err)
	second, err := BuildLifecycleInventory(inventory)
	require.NoError(t, err)

	firstJSON, err := marshalCanonicalJSON(first)
	require.NoError(t, err)
	secondJSON, err := marshalCanonicalJSON(second)
	require.NoError(t, err)
	require.Equal(t, firstJSON, secondJSON)
}

func TestLifecycleInventoryDoesNotReadExternalMaterial(t *testing.T) {
	t.Parallel()

	model, scopes := fixture(t)
	inventory, err := build(model, scopes, testServiceSMKMSPGSelector)
	require.NoError(t, err)

	lifecycle, err := BuildLifecycleInventory(inventory)
	require.NoError(t, err)
	encoded, err := marshalCanonicalJSON(lifecycle)
	require.NoError(t, err)

	require.NotContains(t, string(encoded), "privateKey")
	require.NotContains(t, string(encoded), "password")
	require.NotContains(t, string(encoded), "data:")
}
