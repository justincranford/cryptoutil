// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	sha256 "crypto/sha256"
	"encoding/hex"
	json "encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	pki "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// This fixture exercises production orchestration with an injected renderer.
// Its synthetic images/material and minimal renderer are not runtime certification.
func completeVerificationRequest(t *testing.T) (gen.BootstrapPreflightInput, execution.Request, PKIPublicationInput) {
	t.Helper()

	values, leaves, manifest := publicationFixture(t)
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "k8s-model.json"))
	require.NoError(t, err)
	model, err := gen.DecodeModel(raw)
	require.NoError(t, err)

	model.Replicas.Min, model.Replicas.Max = 1, 1
	inventory, err := gen.Build(model, "s:sm-kms:mem")
	require.NoError(t, err)
	identities, err := gen.BuildLifecycleInventory(inventory)
	require.NoError(t, err)

	request := execution.Request{Model: model, Inventory: inventory, Input: gen.LifecycleInput{Version: 1}}
	roots := map[string]int{}

	for _, id := range identities.Identities {
		if id.Kind != testPKISecretKind {
			continue
		}

		data := map[string][]byte{
			testPKICAKey:              []byte("synthetic non-certificate material"),
			pki.SecretDataCertificate: []byte("synthetic certificate placeholder"),
			pki.SecretDataPrivateKey:  []byte("synthetic key placeholder"),
			testPKIUsernameKey:        []byte("synthetic role"),
			testPKIPasswordKey:        []byte("synthetic credential"),
		}
		isLeaf := false

		for _, leaf := range leaves[:3] {
			if id.Name == leaf.Metadata.Name && id.Namespace == leaf.Metadata.Namespace {
				data = leaf.Data
				isLeaf = true
			}
		}

		if _, exists := roots[id.Namespace]; !exists && !isLeaf {
			roots[id.Namespace] = len(request.Secrets)
			data = map[string][]byte{
				testPKICAKey:              leaves[3].Data[testPKICAKey],
				pki.SecretDataCertificate: leaves[1].Data[pki.SecretDataCertificate],
				pki.SecretDataPrivateKey:  leaves[1].Data[pki.SecretDataPrivateKey],
				testPKIUsernameKey:        []byte("synthetic role"),
				testPKIPasswordKey:        []byte("synthetic credential"),
			}
		}

		request.Secrets = append(request.Secrets, gen.LifecycleSecret{
			APIVersion: testVerifyAPIVersion, Kind: testPKISecretKind, Type: testPKIOpaque,
			Metadata: gen.LifecycleMetadata{Name: id.Name, Namespace: id.Namespace, Labels: map[string]string{
				testPKIRunLabel: model.Run, testPKIOwnerLabel: testPKIImportOwner,
			}}, Data: data,
		})
	}

	for _, id := range identities.Identities {
		if id.Kind != testPKIJobKind {
			continue
		}

		rootIndex, exists := roots[id.Namespace]
		require.True(t, exists)

		root := &request.Secrets[rootIndex]
		file := func(name, key string) string { return "/run/cryptoutil/secrets/" + name + "/" + key }
		trust := file(root.Metadata.Name, testPKICAKey)
		plan := pki.VerificationInput{Version: 1, Files: map[string]string{}, Certificates: []pki.VerificationCertificate{{
			Certificate: file(root.Metadata.Name, pki.SecretDataCertificate),
			Key:         file(root.Metadata.Name, pki.SecretDataPrivateKey), Trust: trust, Usage: testPKIClientFailure,
		}}}
		names := []string{root.Metadata.Name}

		if id.Namespace == values[0].Namespace {
			values[0].Workloads[0].TrustSecretName = root.Metadata.Name
			for role, ref := range manifest.Pods[0].Trust {
				ref.Secret = root.Metadata.Name
				manifest.Pods[0].Trust[role] = ref
			}

			for index, leaf := range leaves[:3] {
				names = append(names, leaf.Metadata.Name)
				if index == 0 {
					plan.Certificates = append(plan.Certificates,
						pki.VerificationCertificate{Certificate: file(leaf.Metadata.Name, pki.SecretDataPublicCertificate), Key: file(leaf.Metadata.Name, pki.SecretDataPublicPrivateKey), Trust: trust, Usage: testVerifyServerRole, DNSNames: manifest.Pods[0].ExpectedPublicDNSNames},
						pki.VerificationCertificate{Certificate: file(leaf.Metadata.Name, pki.SecretDataAdminCertificate), Key: file(leaf.Metadata.Name, pki.SecretDataAdminPrivateKey), Trust: trust, Usage: testVerifyServerRole, DNSNames: []string{manifest.Pods[0].ExpectedAdminDNS}},
					)
				} else {
					plan.Certificates = append(plan.Certificates, pki.VerificationCertificate{Certificate: file(leaf.Metadata.Name, pki.SecretDataCertificate), Key: file(leaf.Metadata.Name, pki.SecretDataPrivateKey), Trust: trust, Usage: testPKIClientFailure})
				}
			}
		}

		for _, secret := range request.Secrets {
			if secret.Metadata.Namespace != id.Namespace || !containsPKIName(names, secret.Metadata.Name) {
				continue
			}

			for key, data := range secret.Data {
				digest := sha256.Sum256(data)
				plan.Files[file(secret.Metadata.Name, key)] = hex.EncodeToString(digest[:])
			}
		}

		planBytes, err := json.Marshal(plan)
		require.NoError(t, err)

		root.Data[pki.VerificationInputName] = planBytes
		request.Input.Jobs = append(request.Input.Jobs, gen.LifecycleJobInput{
			Namespace: id.Namespace, Name: strings.TrimSuffix(id.Name, "-"+model.Revision),
			Image: "registry.example.invalid/suite@sha256:" + strings.Repeat("a", 64), Command: []string{testPKISuiteCommand},
			Args:        []string{pki.VerifyCommand, "--input=" + file(root.Metadata.Name, pki.VerificationInputName)},
			SecretNames: names, WriteSecretNames: []string{}, ActiveDeadlineSeconds: 60,
			Resources: gen.ResourceRequirements{Requests: gen.ResourceSet{CPU: testPKIRequestCPU, Memory: testPKIRequestMemory}, Limits: gen.ResourceSet{CPU: testPKILimitCPU, Memory: testPKILimitMemory}},
		})
	}

	for _, secret := range request.Secrets {
		source := gen.LifecycleSecretSource{Namespace: secret.Metadata.Namespace, Name: secret.Metadata.Name, Type: secret.Type, Files: map[string]string{}}
		for key := range secret.Data {
			source.Files[key] = testExecutionMaterialFile
		}

		request.Input.Secrets = append(request.Input.Secrets, source)
	}

	request.Model.Revision, err = gen.LifecycleRevision(request.Input)
	require.NoError(t, err)
	request.Inventory, err = gen.Build(request.Model, "s:sm-kms:mem")
	require.NoError(t, err)
	_, err = gen.BuildLifecycleChartValues(request.Model, request.Inventory, request.Input)
	require.NoError(t, err)

	input := gen.BootstrapPreflightInput{Model: request.Model, Inventory: request.Inventory, Lifecycle: &request.Input, RuntimeValues: values}

	return input, request, manifest
}

type completionVerificationRunner struct {
	*publicationRunner
	lastJob     string
	jobReads    int
	secretReads int
	jobCreates  int
	corrupt     bool
	resources   map[string]map[string]any
}

func (r *completionVerificationRunner) Run(ctx context.Context, executable string, args []string, input []byte) ([]byte, error) {
	if executable != r.helm && args[4] == testVerificationCreate {
		var desired map[string]any
		require.NoError(r.t, json.Unmarshal(input, &desired))

		switch desired[testPKIKindField] {
		case testPKIJobKind:
			if r.lastJob != "" {
				require.GreaterOrEqual(r.t, r.jobReads, 2, "next Job must follow live postComplete Job verification")
				require.Positive(r.t, r.secretReads, "next Job must follow live postComplete material verification")
			}

			r.jobCreates++
			r.jobReads, r.secretReads = 0, 0
		case testPKISecretKind:
			require.Empty(r.t, r.lastJob, "all Secret creation must precede Jobs")
		}
	}

	var (
		raw []byte
		err error
	)

	if executable != r.helm && args[4] == testPKIGetCommand && args[5] != testPKINamespaceKind {
		doc := r.resources[args[5]+"/"+args[len(args)-1]+"/"+args[6]]
		if doc != nil {
			raw, err = json.Marshal(doc)
		}
	} else {
		raw, err = r.publicationRunner.Run(ctx, executable, args, input)
		if err == nil && executable != r.helm && args[4] == testVerificationCreate {
			var doc map[string]any
			require.NoError(r.t, json.Unmarshal(raw, &doc))
			meta, ok := doc[testPKIMetadataField].(map[string]any)
			require.True(r.t, ok)
			kind, ok := doc[testPKIKindField].(string)
			require.True(r.t, ok)
			namespace, ok := meta[testPKINamespaceFailure].(string)
			require.True(r.t, ok)
			name, ok := meta[testPKINameField].(string)
			require.True(r.t, ok)

			r.resources[kind+"/"+namespace+"/"+name] = doc
		}
	}

	if err != nil || executable == r.helm || args[4] != testPKIGetCommand || len(raw) == 0 {
		return raw, err
	}

	var actual map[string]any
	require.NoError(r.t, json.Unmarshal(raw, &actual))

	if actual[testPKIKindField] == testPKIJobKind {
		r.lastJob = args[6]
		r.jobReads++
		actual[testVerificationStatus] = map[string]any{"conditions": []any{map[string]any{"type": "Complete", testVerificationStatus: "True"}}}

		if r.corrupt && r.jobReads == 2 {
			for _, doc := range r.resources {
				meta, ok := doc[testPKIMetadataField].(map[string]any)
				if !ok || meta[testPKINamespaceFailure] != args[len(args)-1] || doc[testPKIKindField] != testPKISecretKind {
					continue
				}

				data, ok := doc["data"].(map[string]any)
				require.True(r.t, ok)

				for key := range data {
					data[key] = testVerifyEncodedMaterial
				}
			}
		}

		return json.Marshal(actual)
	}

	if actual[testPKIKindField] == testPKISecretKind && r.jobReads >= 2 {
		r.secretReads++
	}

	return raw, nil
}

const testVerifyEncodedMaterial = "Zm9yZ2Vk"
const (
	testVerificationCreate = "create"
	testVerificationStatus = cryptoutilSharedMagic.StringStatus
)

func TestExecuteProductionPKIVerifierAfterCompleteBeforeDependentJob(t *testing.T) {
	t.Parallel()

	for _, corrupt := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordered read-only completion", true: "Complete cannot hide replaced material"}[corrupt], func(t *testing.T) {
			t.Parallel()

			input, request, manifest := completeVerificationRequest(t)
			request, err := preparePKIVerification(input, request, manifest)
			require.NoError(t, err)
			root := t.TempDir()
			config := execution.KubectlConfig{RepositoryRoot: root, Helm: filepath.Join(root, "helm.exe"), Kubectl: filepath.Join(root, "kubectl.exe"), Kubeconfig: filepath.Join(root, "kubeconfig"), Context: testExecutionContext, PollInterval: time.Millisecond}
			runner := &completionVerificationRunner{publicationRunner: &publicationRunner{t: t, input: input, helm: config.Helm, documents: map[string]map[string]any{}}, corrupt: corrupt, resources: map[string]map[string]any{}}
			client, err := execution.NewKubectlClient(config, runner)
			require.NoError(t, err)

			result, err := execution.Execute(t.Context(), request, client, execution.Policy{MaxDuration: time.Minute, OperationTimeout: 10 * time.Second})
			if corrupt {
				require.ErrorIs(t, err, execution.ErrCollision)
				require.Zero(t, result.Jobs)
				require.Equal(t, 1, runner.jobCreates)

				return
			}

			require.NoError(t, err)
			require.Equal(t, 2, result.Jobs)
			require.Equal(t, len(request.Secrets), result.Secrets)
			require.Equal(t, 2, runner.jobCreates)
			require.Equal(t, 2, runner.jobReads)
			require.Positive(t, runner.secretReads)
		})
	}
}
