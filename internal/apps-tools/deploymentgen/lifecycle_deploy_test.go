// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	db "cryptoutil/internal/apps-tools/deploymentrun/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testLifecycleCoverageError = "cover every"
	testLifecyclePinnedError   = "pinned suite"
	testLifecycleUnknownName   = cryptoutilSharedMagic.STRATEGY_UNKNOWN
	testLifecycleRelativeError = "relative file"
)

func lifecycleDeployFixture(t *testing.T, selector string, configure ...func(*Model)) (Model, Inventory, LifecycleInput) {
	t.Helper()

	model, scopes := fixture(t)

	for _, apply := range configure {
		apply(&model)
	}

	inventory, err := build(model, scopes, selector)
	require.NoError(t, err)
	identities, err := BuildLifecycleInventory(inventory)
	require.NoError(t, err)

	input := LifecycleInput{Version: 1, Secrets: []LifecycleSecretSource{}, Jobs: []LifecycleJobInput{}}

	for _, identity := range identities.Identities {
		switch identity.Kind {
		case testLifecycleSecretKind:
			files := map[string]string{"value": identity.Namespace + "/" + identity.Name + cryptoutilSharedMagic.CICDTemplateSecretFileSuffix}
			for _, key := range []string{cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey, "ca.crt", "username", "password"} {
				files[key] = identity.Namespace + "/" + identity.Name + "/" + key
			}

			files[db.InputName] = identity.Namespace + "/" + identity.Name + "/" + db.InputName

			input.Secrets = append(input.Secrets, LifecycleSecretSource{
				Namespace: identity.Namespace, Name: identity.Name, Type: cryptoutilSharedMagic.KubernetesSecretTypeOpaque,
				Files: files,
			})
		case lifecycleJobKind:
			name := strings.TrimSuffix(identity.Name, "-"+model.Revision)
			args, _, _ := db.JobArguments(name)

			if strings.HasPrefix(name, lifecyclePKICommand) || name == lifecycleInfraPKIName {
				args = []string{cryptoutilSharedMagic.PSIDPKIInit, "--domain=cryptoutil", "--output-dir=/tmp/pki"}
			}

			input.Jobs = append(input.Jobs, LifecycleJobInput{
				Namespace: identity.Namespace, Name: name,
				Image:   "registry.example.invalid/cryptoutil@sha256:" + strings.Repeat("a", 64),
				Command: []string{cryptoutilSharedMagic.KubernetesSuiteImageBinary}, Args: args,
				Resources:             chartMaterializerFixture().Resources,
				ActiveDeadlineSeconds: 300, BackoffLimit: 1,
				SecretNames: []string{}, WriteSecretNames: []string{}, ConfigMapNames: []string{},
			})
		}
	}

	for index := range input.Jobs {
		job := &input.Jobs[index]
		if job.Name != lifecycleInfraPKIName && !strings.HasPrefix(job.Name, lifecyclePKICommand) {
			_, manifestSecret, err := db.JobArguments(job.Name)
			require.NoError(t, err)

			job.SecretNames = []string{manifestSecret}
			if strings.HasSuffix(job.Name, "-migrate") {
				job.SecretNames = []string{strings.TrimSuffix(job.Name, "-migrate") + "-migration-db"}
			}

			continue
		}

		for _, secret := range input.Secrets {
			if secret.Namespace == job.Namespace && strings.HasSuffix(secret.Name, "-tls") {
				job.WriteSecretNames = append(job.WriteSecretNames, secret.Name)
			}
		}

		if job.Name == lifecycleInfraPKIName {
			job.SecretNames = []string{lifecycleAuthoritySecret}
		}
	}

	model.Revision, err = LifecycleRevision(input)
	require.NoError(t, err)
	inventory, err = build(model, scopes, selector)
	require.NoError(t, err)

	return model, inventory, input
}

func TestLifecycleDeployContract(t *testing.T) {
	t.Parallel()

	for _, selector := range []string{testServiceSMKMSPGSelector, testServiceSMKMSMemSelector, ""} {
		t.Run("selector="+selector, func(t *testing.T) {
			t.Parallel()

			model, inventory, input := lifecycleDeployFixture(t, selector)
			values, err := BuildLifecycleChartValues(model, inventory, input)
			require.NoError(t, err)
			require.Len(t, values.Jobs, len(input.Jobs))
			require.Len(t, values.Secrets, len(input.Secrets))
			require.Len(t, values.Namespaces, len(inventory.Releases)+1)

			for _, job := range values.Jobs {
				require.True(t, strings.HasSuffix(job.Name, "-"+model.Revision))
				require.Equal(t, strings.HasPrefix(job.Name, "pki-init-") || strings.HasPrefix(job.Name, "infra-pki-"), job.APIToken)
			}

			second, err := BuildLifecycleChartValues(model, inventory, input)
			require.NoError(t, err)
			require.Equal(t, values, second)
		})
	}
}

func TestLifecycleDeployRejectsMissingAndUnsafeInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*LifecycleInput)
		want string
	}{
		{"missing secrets", func(v *LifecycleInput) { v.Secrets = nil }, testLifecycleCoverageError},
		{"missing jobs", func(v *LifecycleInput) { v.Jobs = nil }, testLifecycleCoverageError},
		{"unpinned image", func(v *LifecycleInput) { v.Jobs[0].Image = "cryptoutil:latest" }, testLifecyclePinnedError},
		{"missing command", func(v *LifecycleInput) { v.Jobs[0].Command = nil }, testLifecyclePinnedError},
		{"missing args", func(v *LifecycleInput) { v.Jobs[0].Args = nil }, testLifecyclePinnedError},
		{"missing deadline", func(v *LifecycleInput) { v.Jobs[0].ActiveDeadlineSeconds = 0 }, "bounded execution"},
		{"missing resources", func(v *LifecycleInput) { v.Jobs[0].Resources = ResourceRequirements{} }, testLifecycleResourcesField},
		{"unknown secret", func(v *LifecycleInput) { v.Secrets[0].Name = testLifecycleUnknownName }, testLifecycleUnknownObject},
		{"duplicate secret", func(v *LifecycleInput) { v.Secrets[1] = v.Secrets[0] }, testLifecycleUnknownObject},
		{"traversal", func(v *LifecycleInput) { v.Secrets[0].Files["value"] = "../secret" }, testLifecycleRelativeError},
		{"absolute path", func(v *LifecycleInput) { v.Secrets[0].Files["value"] = "C:/secret" }, testLifecycleRelativeError},
		{"inline material", func(v *LifecycleInput) { v.Secrets[0].Type = "Inline" }, "supported type"},
		{"TLS missing key", func(v *LifecycleInput) {
			v.Secrets[0].Type = "kubernetes.io/tls"
			delete(v.Secrets[0].Files, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey)
		}, "tls.crt and tls.key"},
		{"unowned mount", func(v *LifecycleInput) { v.Jobs[0].SecretNames = []string{testLifecycleUnknownName} }, "unowned Secret"},
		{"unowned config", func(v *LifecycleInput) { v.Jobs[0].ConfigMapNames = []string{testLifecycleUnknownName} }, "unowned ConfigMap"},
		{"duplicate mount", func(v *LifecycleInput) {
			v.Jobs[0].SecretNames = []string{lifecyclePostgresBootstrap, lifecyclePostgresBootstrap}
		}, "duplicate resource"},
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

func TestLifecycleRevisionAndStrictDecode(t *testing.T) {
	t.Parallel()

	model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
	raw, err := json.Marshal(input)
	require.NoError(t, err)
	decoded, err := DecodeLifecycleInput(raw)
	require.NoError(t, err)
	require.Equal(t, input, decoded)

	_, err = DecodeLifecycleInput([]byte(strings.Replace(string(raw), `"version":`, `"Version":`, 1)))
	require.ErrorContains(t, err, "case-mismatched")

	input.Jobs[0].ActiveDeadlineSeconds++
	_, err = BuildLifecycleChartValues(model, inventory, input)
	require.ErrorContains(t, err, "content-addressed")
}
