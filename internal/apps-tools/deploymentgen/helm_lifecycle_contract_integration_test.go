// Copyright (c) 2025-2026 Justin Cranford.

//go:build integration

package deploymentgen

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const (
	testLifecycleSchemaFailure          = "schema"
	testLifecycleGetVerb                = "get"
	testLifecycleCPUField               = "cpu"
	testLifecycleMemoryField            = "memory"
	testLifecycleKindField              = "kind"
	testLifecycleSecretsResource        = "secrets"
	testLifecycleNameField              = cryptoutilSharedMagic.ClaimName
	testLifecycleDatabaseOperationError = "exact suite operation"
)

func TestHelmLifecycleRejectsForgedContracts(t *testing.T) {
	t.Parallel()

	helm := requirePinnedHelm(t)
	tests := []struct {
		name string
		edit func(*LifecycleChartValues)
		want string
	}{
		{"missing image pin", func(v *LifecycleChartValues) { v.Jobs[0].Image = "unpinned" }, testLifecycleSchemaFailure},
		{"missing resource requests", func(v *LifecycleChartValues) { v.Jobs[0].Resources = ResourceRequirements{} }, testLifecycleSchemaFailure},
		{"render missing deadline", func(v *LifecycleChartValues) { v.Jobs[0].ActiveDeadlineSeconds = 0 }, testLifecycleSchemaFailure},
		{"database API token", func(v *LifecycleChartValues) { v.Jobs[0].APIToken = true }, "only PKI Jobs"},
		{"render unowned namespace", func(v *LifecycleChartValues) { v.Jobs[0].Namespace = testDifferentIdentity }, "unowned namespace"},
		{"unowned mounted Secret", func(v *LifecycleChartValues) { v.Jobs[0].SecretNames = []string{testLifecycleUnknownName} }, "unowned imported Secret"},
		{"forged RBAC write", func(v *LifecycleChartValues) { v.Namespaces[0].WriteSecretNames = []string{lifecycleAuthoritySecret} }, testLifecycleOverwriteError},
		{"wrong revision suffix", func(v *LifecycleChartValues) { v.Jobs[0].Name = "union-db-r0" }, "content-addressed"},
		{"wrong release owner", func(v *LifecycleChartValues) { v.Jobs[0].Release = testDifferentIdentity }, "does not match"},
		{"database signer mount", func(v *LifecycleChartValues) { v.Jobs[0].SecretNames = []string{lifecycleAuthoritySecret} }, testLifecycleSigningError},
		{"duplicate Job identity", func(v *LifecycleChartValues) { v.Jobs = append(v.Jobs, v.Jobs[0]) }, "duplicate lifecycle Job"},
		{"fixture database command", func(v *LifecycleChartValues) {
			v.Jobs[0].Args = []string{testLifecycleLegacyProduct, testLifecycleLegacyService, testLifecycleInitCommand}
		}, testLifecycleDatabaseOperationError},
		{"wrong database operation", func(v *LifecycleChartValues) { v.Jobs[0].Args[1] = "--operation=replicate" }, testLifecycleDatabaseOperationError},
		{"missing protected database manifest", func(v *LifecycleChartValues) {
			for index := range v.Secrets {
				if v.Secrets[index].Name == "postgres-bootstrap" {
					delete(v.Secrets[index].Files, "lifecycle.json")
				}
			}
		}, "manifest source is missing"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSPGSelector)
			values, err := BuildLifecycleChartValues(model, inventory, input)
			require.NoError(t, err)
			test.edit(&values)
			raw, err := marshalCanonicalJSON(values)
			require.NoError(t, err)
			file := filepath.Join(t.TempDir(), "lifecycle.json")
			require.NoError(t, os.WriteFile(file, raw, 0o600))

			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()

			output, err := exec.CommandContext(ctx, helm, cryptoutilSharedMagic.SkeletonTemplateServiceName, "cu-lifecycle",
				filepath.Join("..", "..", "..", canonicalChartSourceRoot, "cryptoutil-lifecycle"), "--values", file).CombinedOutput()
			require.Error(t, err, string(output))
			require.Contains(t, string(output), test.want)
		})
	}
}

func TestHelmLifecycleRestrictedRBACAndExecutionFields(t *testing.T) {
	t.Parallel()

	helm := requirePinnedHelm(t)
	model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSPGSelector)
	values, err := BuildLifecycleChartValues(model, inventory, input)
	require.NoError(t, err)
	raw, err := marshalCanonicalJSON(values)
	require.NoError(t, err)
	file := filepath.Join(t.TempDir(), "lifecycle.json")
	require.NoError(t, os.WriteFile(file, raw, 0o600))

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	chart := filepath.Join("..", "..", "..", canonicalChartSourceRoot, "cryptoutil-lifecycle")
	output, err := exec.CommandContext(ctx, helm, "lint", chart, "--strict", "--kube-version", "1.37.1", "--values", file).CombinedOutput()
	require.NoError(t, err, string(output))
	output, err = exec.CommandContext(ctx, helm, cryptoutilSharedMagic.SkeletonTemplateServiceName, "cu-lifecycle", chart, "--values", file).CombinedOutput()
	require.NoError(t, err, string(output))
	decoder := yaml.NewDecoder(bytes.NewReader(output))

	for {
		var document map[string]any

		err := decoder.Decode(&document)
		if errors.Is(err, io.EOF) {
			break
		}

		require.NoError(t, err)

		if len(document) == 0 {
			continue
		}

		metadata := lifecycleValueAs[map[string]any](t, document["metadata"])
		namespace := lifecycleValueAs[string](t, metadata["namespace"])

		switch document[testLifecycleKindField] {
		case testLifecycleServiceAccountKind:
			require.Equal(t, false, document["automountServiceAccountToken"])
		case testLifecycleRoleKind:
			var boundary LifecycleNamespace

			for _, candidate := range values.Namespaces {
				if candidate.Name == namespace {
					boundary = candidate
				}
			}

			expected := make([]any, 0)
			if len(boundary.ReadSecretNames) > 0 {
				expected = append(expected, lifecycleExpectedRule(boundary.ReadSecretNames, []string{testLifecycleGetVerb}))
			}

			if len(boundary.WriteSecretNames) > 0 {
				expected = append(expected, lifecycleExpectedRule(boundary.WriteSecretNames, []string{testLifecycleGetVerb, "update", "patch"}))
			}

			require.Equal(t, expected, document["rules"])
		case testLifecycleRoleBindingKind:
			require.Equal(t, map[string]any{"apiGroup": "rbac.authorization.k8s.io", testLifecycleKindField: testLifecycleRoleKind, testLifecycleNameField: testLifecycleBootstrapResourceName}, document["roleRef"])
			require.Equal(t, []any{map[string]any{testLifecycleKindField: testLifecycleServiceAccountKind, testLifecycleNameField: testLifecycleBootstrapResourceName, "namespace": namespace}}, document["subjects"])
		case lifecycleJobKind:
			for _, job := range values.Jobs {
				if job.Namespace != namespace || job.Name != metadata[cryptoutilSharedMagic.ClaimName] {
					continue
				}

				spec := lifecycleValueAs[map[string]any](t, document["spec"])
				require.Equal(t, job.ActiveDeadlineSeconds, spec["activeDeadlineSeconds"])
				require.Equal(t, job.BackoffLimit, spec["backoffLimit"])
				require.Equal(t, 1, spec["parallelism"])
				require.Equal(t, 1, spec["completions"])
				template := lifecycleValueAs[map[string]any](t, spec[cryptoutilSharedMagic.SkeletonTemplateServiceName])
				pod := lifecycleValueAs[map[string]any](t, template["spec"])
				require.Equal(t, "Never", pod["restartPolicy"])
				require.Equal(t, testLifecycleBootstrapResourceName, pod["serviceAccountName"])
				containers := lifecycleValueAs[[]any](t, pod["containers"])
				require.Len(t, containers, 1)
				container := lifecycleValueAs[map[string]any](t, containers[0])
				require.Equal(t, job.Image, container["image"])
				require.Equal(t, lifecycleStringsToAny(job.Command), container["command"])
				require.Equal(t, lifecycleStringsToAny(job.Args), container["args"])
				require.Equal(t, map[string]any{
					"requests": map[string]any{testLifecycleCPUField: job.Resources.Requests.CPU, testLifecycleMemoryField: job.Resources.Requests.Memory},
					"limits":   map[string]any{testLifecycleCPUField: job.Resources.Limits.CPU, testLifecycleMemoryField: job.Resources.Limits.Memory},
				}, container[testLifecycleResourcesField])

				tokenCount := 0

				for _, volume := range lifecycleValueAs[[]any](t, pod["volumes"]) {
					entry := lifecycleValueAs[map[string]any](t, volume)
					if entry[cryptoutilSharedMagic.ClaimName] != testRuntimeBootstrapTokenVolume {
						continue
					}

					tokenCount++
					projected := lifecycleValueAs[map[string]any](t, entry["projected"])
					require.Equal(t, 288, projected["defaultMode"])
					sources := lifecycleValueAs[[]any](t, projected["sources"])
					require.Len(t, sources, 3)
					require.Equal(t, map[string]any{"audience": values.APIAudience, "expirationSeconds": 600, "path": cryptoutilSharedMagic.ParamToken},
						lifecycleValueAs[map[string]any](t, sources[0])["serviceAccountToken"])
					configSource := lifecycleValueAs[map[string]any](t, sources[1])["configMap"]
					require.Equal(t, "kube-root-ca.crt", lifecycleValueAs[map[string]any](t, configSource)[cryptoutilSharedMagic.ClaimName])
					require.Contains(t, lifecycleValueAs[map[string]any](t, sources[2]), "downwardAPI")
				}

				if job.APIToken {
					require.Equal(t, 1, tokenCount)
				} else {
					require.Zero(t, tokenCount)
				}
			}
		}
	}
}

func lifecycleExpectedRule(names, verbs []string) map[string]any {
	return map[string]any{
		"apiGroups": []any{""}, testLifecycleResourcesField: []any{testLifecycleSecretsResource},
		"resourceNames": lifecycleStringsToAny(names), "verbs": lifecycleStringsToAny(verbs),
	}
}

func lifecycleValueAs[T any](t *testing.T, value any) T {
	t.Helper()

	typed, ok := value.(T)
	require.True(t, ok, "unexpected rendered type %T", value)

	return typed
}

func lifecycleStringsToAny(values []string) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = value
	}

	return result
}
