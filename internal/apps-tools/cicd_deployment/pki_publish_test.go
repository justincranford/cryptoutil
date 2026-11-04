// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"bytes"
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	json "encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
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

const (
	testPKICertificatePEM   = cryptoutilSharedMagic.StringPEMTypeCertificate
	testPKITrustSecret      = "trust-ca"
	testPKIOwnerLabel       = "cu.io/owner"
	testPKIPrepareFailure   = "prepare"
	testPKINamespaceFailure = "namespace"
	testPKISecretFailure    = "secret"
	testPKINameField        = cryptoutilSharedMagic.ClaimName
	testPKINamespaceKind    = "Namespace"
	testPKIReleaseLabel     = cryptoutilSharedMagic.KubernetesLabelRelease
	testPKIKindField        = "kind"
	testPKIRunLabel         = cryptoutilSharedMagic.KubernetesLabelRun
	testPKILabelsField      = "labels"
	testPKIMetadataField    = "metadata"
	testPKIGetCommand       = "get"
	testPKIImportOwner      = "external-secret-import"
	testPKIRun              = "review"
	testPKIClientFailure    = "client"
)

func publicationFixture(t *testing.T) ([]gen.RuntimeChartValues, []gen.LifecycleSecret, PKIPublicationInput) {
	t.Helper()

	now := time.Now().UTC()
	key, err := ecdsa.GenerateKey(elliptic.P384(), crand.Reader)
	require.NoError(t, err)

	rootTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1), NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	rootDER, err := x509.CreateCertificate(crand.Reader, rootTemplate, rootTemplate, &key.PublicKey, key)
	require.NoError(t, err)
	root, err := x509.ParseCertificate(rootDER)
	require.NoError(t, err)

	rootPEM := pem.EncodeToMemory(&pem.Block{Type: testPKICertificatePEM, Bytes: rootDER})
	namespace, pod, service := "cu-review-s-sm-kms-mem", "sm-kms-mem-0", "sm-kms-mem-i01"
	serviceDNS := service + "." + namespace + ".svc.cluster.local"
	workloadDNS := "sm-kms-mem." + namespace + ".svc.cluster.local"
	podDNS := pod + ".sm-kms-mem-pods." + namespace + ".svc.cluster.local"
	adminDNS := podDNS
	leaf := func(dns []string, loopback bool, usage x509.ExtKeyUsage) ([]byte, []byte) {
		leafKey, err := ecdsa.GenerateKey(elliptic.P384(), crand.Reader)
		require.NoError(t, err)

		template := &x509.Certificate{
			SerialNumber: big.NewInt(2), NotBefore: now.Add(-time.Minute), NotAfter: now.Add(time.Hour),
			BasicConstraintsValid: true, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{usage}, DNSNames: dns,
		}
		if loopback {
			template.IPAddresses = []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)}
		}

		der, err := x509.CreateCertificate(crand.Reader, template, root, &leafKey.PublicKey, key)
		require.NoError(t, err)
		keyDER, err := x509.MarshalPKCS8PrivateKey(leafKey)
		require.NoError(t, err)

		return append(pem.EncodeToMemory(&pem.Block{Type: testPKICertificatePEM, Bytes: der}), rootPEM...),
			pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, Bytes: keyDER})
	}
	publicCRT, publicKey := leaf([]string{serviceDNS, workloadDNS, podDNS}, false, x509.ExtKeyUsageServerAuth)
	adminCRT, adminKey := leaf([]string{adminDNS}, true, x509.ExtKeyUsageServerAuth)
	clientCRT, clientKey := leaf(nil, false, x509.ExtKeyUsageClientAuth)
	names := []string{pod + "-tls", pod + "-health-client", pod + "-otlp-client"}
	data := []map[string][]byte{
		{
			pki.SecretDataPublicCertificate: publicCRT, pki.SecretDataPublicPrivateKey: publicKey,
			pki.SecretDataAdminCertificate: adminCRT, pki.SecretDataAdminPrivateKey: adminKey,
		},
		{pki.SecretDataCertificate: clientCRT, pki.SecretDataPrivateKey: clientKey},
		{pki.SecretDataCertificate: clientCRT, pki.SecretDataPrivateKey: clientKey},
	}

	imports := make([]gen.LifecycleSecret, 0, len(names)+1)
	for index, name := range names {
		imports = append(imports, gen.LifecycleSecret{
			APIVersion: "v1", Kind: testPKISecretKind, Type: testPKIOpaque,
			Metadata: gen.LifecycleMetadata{Name: name, Namespace: namespace, Labels: map[string]string{
				testPKIRunLabel: testPKIRun, testPKIOwnerLabel: testPKIImportOwner,
			}}, Data: data[index],
		})
	}

	imports = append(imports, gen.LifecycleSecret{Metadata: gen.LifecycleMetadata{
		Name: testPKITrustSecret, Namespace: namespace,
		Labels: map[string]string{testPKIRunLabel: testPKIRun, testPKIOwnerLabel: testPKIImportOwner},
	}, Data: map[string][]byte{testPKICAKey: rootPEM}})
	values := []gen.RuntimeChartValues{{
		Namespace: namespace, Run: testPKIRun, ReleaseKey: "s-sm-kms-mem", HelmRelease: namespace, ClusterDomain: "cluster.local",
		Workloads: []gen.RuntimeChartWorkload{{
			PSID: testPKIServiceID, Pool: cryptoutilSharedMagic.DeploymentPoolMem, TrustSecretName: testPKITrustSecret,
			FrameworkConfigMapName: cryptoutilSharedMagic.FrameworkProductName, FrameworkConfigFile: testPKIFrameworkFile, DomainConfigMapName: "domain", DomainConfigFile: testPKIDomainFile,
			InstanceServices: []gen.InstanceService{{
				Name: service, PodName: pod, TLSLeafSecretName: names[0],
				HealthClientSecretName: names[1], OTLPClientSecretName: names[2],
			}},
		}},
	}}

	trust := make(map[string]PKITrustReference)
	for _, role := range []string{pkiPublicTrustRole, pkiAdminTrustRole, pki.SecretRoleHealthClient, pki.SecretRoleOTLPClient} {
		trust[role] = PKITrustReference{Secret: testPKITrustSecret, Key: testPKICAKey}
	}

	return values, imports, PKIPublicationInput{Pods: []PKIPublicationPod{{
		Namespace: namespace, PodName: pod,
		ExpectedAdminDNS: adminDNS, ExpectedPublicDNSNames: []string{serviceDNS, workloadDNS, podDNS}, Trust: trust,
	}}}
}

func TestPKIPublicationBridge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func([]gen.RuntimeChartValues, []gen.LifecycleSecret, *PKIPublicationInput)
		want string
	}{
		{name: "verified memory publication"},
		{name: "duplicate pod", edit: func(_ []gen.RuntimeChartValues, _ []gen.LifecycleSecret, m *PKIPublicationInput) {
			m.Pods = append(m.Pods, m.Pods[0])
		}, want: "duplicate"},
		{name: "missing pod", edit: func(_ []gen.RuntimeChartValues, _ []gen.LifecycleSecret, m *PKIPublicationInput) { m.Pods = nil }, want: "every reserved"},
		{name: "unowned pod", edit: func(_ []gen.RuntimeChartValues, _ []gen.LifecycleSecret, m *PKIPublicationInput) {
			m.Pods = append(m.Pods, PKIPublicationPod{PodName: "forged"})
		}, want: "unowned"},
		{name: "unknown trust", edit: func(_ []gen.RuntimeChartValues, _ []gen.LifecycleSecret, m *PKIPublicationInput) {
			m.Pods[0].Trust["forged"] = PKITrustReference{}
		}, want: cryptoutilSharedMagic.STRATEGY_UNKNOWN},
		{name: "missing trust reference", edit: func(_ []gen.RuntimeChartValues, _ []gen.LifecycleSecret, m *PKIPublicationInput) {
			m.Pods[0].Trust["public-server"] = PKITrustReference{Secret: "absent"}
		}, want: "unavailable"},
		{name: "no anchors", edit: func(_ []gen.RuntimeChartValues, _ []gen.LifecycleSecret, m *PKIPublicationInput) {
			m.Pods[0].Trust = nil
		}, want: "CA anchors"},
		{name: "forged cert", edit: func(_ []gen.RuntimeChartValues, s []gen.LifecycleSecret, _ *PKIPublicationInput) {
			s[0].Data[pki.SecretDataPublicCertificate] = []byte("forged")
		}, want: "certificate validation"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			values, imports, manifest := publicationFixture(t)
			if test.edit != nil {
				test.edit(values, imports, &manifest)
			}

			got, err := buildPKIPublication(values, imports, manifest, time.Now().UTC())
			if test.want != "" {
				require.ErrorContains(t, err, test.want)
				require.Nil(t, got)

				return
			}

			require.NoError(t, err)
			require.Len(t, got, 3)
			require.Equal(t, pki.SecretRoleTLSBundle, got[0].Metadata.Labels[pki.SecretLabelRole])
		})
	}
}

func TestDecodePKIPublication(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"", "[", `{"extra":1}`, `{"pods":[]} {}`, `{"pods":[],"pods":[]}`, `{"pods":[{"trust":{"x":{},"x":{}}}]}`} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			_, err := decodePKIPublication([]byte(raw))
			require.Error(t, err)
		})
	}

	_, err := decodePKIPublication(bytes.Repeat([]byte(" "), 1024*1024+1))
	require.Error(t, err)
	_, err = decodePKIPublication([]byte(`{"pods":[]}`))
	require.NoError(t, err)
}

func TestPKIPublishCLI(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	require.Zero(t, Main([]string{pkiPublishCommand, testPKIHelpFlag}, nil, &stdout, &stderr))
	require.Contains(t, stderr.String(), "publication-input")
	args, _ := executionCLIArgs(t)
	args = append(args, testExecutionPublicationFlag, "")
	require.Equal(t, 2, pkiPublishMain(args, &stdout, &stderr))

	_, err := publishPKIRequest(t.Context(), gen.BootstrapPreflightInput{}, nil, filepath.Join(t.TempDir(), "absent"), execution.KubectlConfig{}, execution.Policy{})
	require.Error(t, err)
}

type rejectedPublicationRunner struct{ calls *int }

func (r rejectedPublicationRunner) Run(context.Context, string, []string, []byte) ([]byte, error) {
	*r.calls++

	return nil, errors.New(testExecutionPrivate)
}

func TestPKIPublicationRejectsBeforeTransport(t *testing.T) {
	t.Parallel()

	calls := 0
	_, err := publishPKIWithRunner(t.Context(), gen.BootstrapPreflightInput{Lifecycle: &gen.LifecycleInput{}},
		nil, PKIPublicationInput{}, execution.KubectlConfig{}, execution.Policy{MaxDuration: time.Minute, OperationTimeout: time.Second}, rejectedPublicationRunner{&calls})
	require.Error(t, err)
	require.Zero(t, calls)

	values, imports, manifest := publicationFixture(t)
	_, err = publishPKIWithRunner(t.Context(), gen.BootstrapPreflightInput{Lifecycle: &gen.LifecycleInput{Version: 1}, RuntimeValues: values},
		imports, manifest, execution.KubectlConfig{}, execution.Policy{MaxDuration: time.Minute, OperationTimeout: time.Second}, rejectedPublicationRunner{&calls})
	require.ErrorContains(t, err, "publication graph")
	require.Zero(t, calls)

	raw, err := json.Marshal(PKIPublicationInput{})
	require.NoError(t, err)
	require.NotContains(t, string(raw), testExecutionPrivate)
}

type publicationRunner struct {
	t         *testing.T
	input     gen.BootstrapPreflightInput
	helm      string
	documents map[string]map[string]any
	releases  map[string]string
	creates   int
	fail      string
}

func (r *publicationRunner) Run(_ context.Context, executable string, args []string, input []byte) ([]byte, error) {
	if executable == r.helm {
		if r.fail == testPKIPrepareFailure {
			return nil, errors.New(testExecutionPrivate)
		}

		var values gen.LifecycleChartValues
		require.NoError(r.t, json.Unmarshal(input, &values))
		r.releases = make(map[string]string)

		var docs []string

		add := func(api, kind, name, namespace, release, revision string, spec any) {
			meta := map[string]any{
				testPKINameField: name, testPKINamespaceFailure: namespace,
				testPKILabelsField: map[string]string{testPKIRunLabel: values.Run, testPKIReleaseLabel: release},
			}
			if revision != "" {
				meta["annotations"] = map[string]string{"cu.io/revision": revision}
			}

			doc := map[string]any{"apiVersion": api, testPKIKindField: kind, testPKIMetadataField: meta}
			if spec != nil {
				doc["spec"] = spec
			}

			raw, err := json.Marshal(doc)
			require.NoError(r.t, err)

			docs = append(docs, string(raw))
		}

		for _, boundary := range values.Namespaces {
			r.releases[boundary.Name] = boundary.Release
			for _, kind := range []string{"ServiceAccount", "Role", "RoleBinding"} {
				add("v1", kind, "bootstrap", boundary.Name, boundary.Release, "", nil)
			}
		}

		for _, job := range values.Jobs {
			add("batch/v1", testPKIJobKind, job.Name, job.Namespace, job.Release, values.Revision, map[string]any{
				"activeDeadlineSeconds": job.ActiveDeadlineSeconds, "backoffLimit": job.BackoffLimit, "completions": 1, "parallelism": 1,
				cryptoutilSharedMagic.SkeletonTemplateServiceName: map[string]any{"spec": map[string]any{
					"serviceAccountName": "bootstrap", "automountServiceAccountToken": false,
					"restartPolicy": "Never", "containers": []any{map[string]any{
						"image": job.Image, "command": job.Command,
						"args": job.Args, "resources": job.Resources,
					}},
				}},
			})
		}

		return []byte(strings.Join(docs, "\n---\n")), nil
	}

	if args[4] == testPKIGetCommand && args[5] == testPKINamespaceKind {
		if r.fail == testPKINamespaceFailure {
			return nil, errors.New(testExecutionPrivate)
		}

		return json.Marshal(map[string]any{testPKIKindField: testPKINamespaceKind, testPKIMetadataField: map[string]any{
			testPKINameField: args[6], "uid": "namespace-uid",
			testPKILabelsField: map[string]string{testPKIRunLabel: r.input.Model.Run, testPKIReleaseLabel: r.releases[args[6]]},
		}, cryptoutilSharedMagic.StringStatus: map[string]string{"phase": "Active"}})
	}

	if args[4] == testPKIGetCommand {
		// Kubectl uses --namespace as a flag, not a shell-interpolated command.
		key := args[len(args)-1] + "/" + args[6]
		if doc := r.documents[key]; doc != nil {
			return json.Marshal(doc)
		}

		return nil, nil
	}

	require.Equal(r.t, "create", args[4])

	if r.fail == testPKISecretFailure {
		return nil, errors.New(testExecutionPrivate)
	}

	var desired map[string]any
	require.NoError(r.t, json.Unmarshal(input, &desired))
	meta, ok := desired[testPKIMetadataField].(map[string]any)
	require.True(r.t, ok)

	meta["uid"] = "secret-uid"
	namespace, ok := meta[testPKINamespaceFailure].(string)
	require.True(r.t, ok)
	name, ok := meta[testPKINameField].(string)
	require.True(r.t, ok)
	r.documents[namespace+"/"+name] = desired
	r.creates++

	return json.Marshal(desired)
}

func TestPKIPublishGuardedTransport(t *testing.T) {
	t.Parallel()

	for _, fail := range []string{"", testPKIPrepareFailure, testPKINamespaceFailure, testPKISecretFailure, testPKIClientFailure} {
		t.Run("boundary-"+fail, func(t *testing.T) {
			t.Parallel()

			values, secrets, manifest := publicationFixture(t)
			raw, err := os.ReadFile(filepath.Join("..", "..", "..", "api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "k8s-model.json"))
			require.NoError(t, err)
			model, err := gen.DecodeModel(raw)
			require.NoError(t, err)

			model.Replicas.Min, model.Replicas.Max = 1, 1
			inventory, err := gen.Build(model, "s:sm-kms:mem")
			require.NoError(t, err)
			identities, err := gen.BuildLifecycleInventory(inventory)
			require.NoError(t, err)

			input := gen.LifecycleInput{Version: 1}

			for _, id := range identities.Identities {
				if id.Kind == testPKISecretKind {
					files := map[string]string{testPKITLSCertKey: testExecutionMaterialFile, testPKITLSPrivateKey: testExecutionMaterialFile, testPKICAKey: testExecutionMaterialFile, testPKIPasswordKey: testExecutionMaterialFile, testPKIUsernameKey: testExecutionMaterialFile}

					for _, secret := range secrets {
						if id.Name == secret.Metadata.Name && id.Namespace == secret.Metadata.Namespace {
							files = make(map[string]string, len(secret.Data))
							for key := range secret.Data {
								files[key] = testExecutionMaterialFile
							}
						}
					}

					input.Secrets = append(input.Secrets, gen.LifecycleSecretSource{Namespace: id.Namespace, Name: id.Name, Type: testPKIOpaque, Files: files})
				}

				if id.Kind == testPKIJobKind {
					input.Jobs = append(input.Jobs, gen.LifecycleJobInput{
						Namespace: id.Namespace, Name: strings.TrimSuffix(id.Name, "-"+model.Revision),
						Image: "registry.example.invalid/suite@sha256:" + strings.Repeat("a", 64), Command: []string{testPKISuiteCommand},
						Args: []string{testPKIInitCommand, testPKIDomainArgument, testPKIOutputArgument}, ActiveDeadlineSeconds: 60,
						Resources: gen.ResourceRequirements{Requests: gen.ResourceSet{CPU: testPKIRequestCPU, Memory: testPKIRequestMemory}, Limits: gen.ResourceSet{CPU: testPKILimitCPU, Memory: testPKILimitMemory}},
					})
				}
			}

			model.Revision, err = gen.LifecycleRevision(input)
			require.NoError(t, err)
			inventory, err = gen.Build(model, "s:sm-kms:mem")
			require.NoError(t, err)

			preflight := gen.BootstrapPreflightInput{Model: model, Inventory: inventory, Lifecycle: &input, RuntimeValues: values}
			root := t.TempDir()

			config := execution.KubectlConfig{
				RepositoryRoot: root, Helm: filepath.Join(root, "helm.exe"), Kubectl: filepath.Join(root, "kubectl.exe"),
				Kubeconfig: filepath.Join(root, "kubeconfig"), Context: testExecutionContext, PollInterval: time.Second,
			}
			if fail == testPKIClientFailure {
				config.Context = ""
			}

			runner := &publicationRunner{t: t, input: preflight, helm: config.Helm, documents: make(map[string]map[string]any), fail: fail}
			policy := execution.Policy{MaxDuration: time.Minute, OperationTimeout: time.Second}

			result, err := publishPKIWithRunner(t.Context(), preflight, secrets, manifest, config, policy, runner)
			if fail != "" {
				require.Error(t, err)
				require.NotContains(t, lifecycleExecutionFailureCategory(err), testExecutionPrivate)

				return
			}

			require.NoError(t, err)
			require.Equal(t, 3, result.Secrets)
			require.Equal(t, 3, runner.creates)
			_, err = publishPKIWithRunner(t.Context(), preflight, secrets, manifest, config, policy, runner)
			require.NoError(t, err)
			require.Equal(t, 3, runner.creates, "existing exact labelled data is verify-only")
		})
	}
}

func TestPKIPublishCommandPipeline(t *testing.T) {
	t.Parallel()

	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			t.Parallel()
			args, _ := executionCLIArgs(t)
			args = append(args, "-publication-input", testPKIPublicationFile)
			called := false
			process := lifecycleExecutionIO{
				read: func(string) ([]byte, error) { return []byte(`{"version":1,"secrets":null,"jobs":null}`), nil },
				load: func(string, string, string, func(string) ([]byte, error)) (gen.BootstrapPreflightInput, error) {
					return gen.BootstrapPreflightInput{}, nil
				},
				admit: func(gen.BootstrapPreflightInput) (gen.CapacityEstimate, error) { return gen.CapacityEstimate{}, nil },
				importSecrets: func(gen.Model, gen.Inventory, gen.LifecycleInput, string) ([]gen.LifecycleSecret, error) {
					return nil, nil
				},
				publish: func(ctx context.Context, input gen.BootstrapPreflightInput, _ []gen.LifecycleSecret, path string,
					_ execution.KubectlConfig, _ execution.Policy,
				) (execution.Result, error) {
					called = true

					require.NotNil(t, input.Lifecycle)
					require.Equal(t, testPKIPublicationFile, path)
					require.NoError(t, ctx.Err())

					if fail {
						return execution.Result{}, errors.New(testExecutionPrivate)
					}

					return execution.Result{Secrets: 3}, nil
				},
			}

			var stdout, stderr bytes.Buffer

			code := lifecycleCommandWithIO(args, &stdout, &stderr, process, pkiPublishCommand)

			require.True(t, called, stderr.String())
			require.NotContains(t, stdout.String()+stderr.String(), testExecutionPrivate)

			if fail {
				require.Equal(t, 1, code)
			} else {
				require.Zero(t, code)
				require.Contains(t, stdout.String(), "secrets=3 jobs=0")
			}
		})
	}
}

func TestPKIPublishFileAndPolicyErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, data string
		directory  bool
	}{
		{name: "read directory", directory: true},
		{name: "invalid manifest", data: "{"},
		{name: "valid manifest no lifecycle", data: `{"pods":[]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			path := t.TempDir()
			if !test.directory {
				path = filepath.Join(path, testPKIPublicationFile)
				require.NoError(t, os.WriteFile(path, []byte(test.data), 0o600))
			}

			_, err := publishPKIRequest(t.Context(), gen.BootstrapPreflightInput{}, nil, path, execution.KubectlConfig{}, execution.Policy{})
			require.Error(t, err)
		})
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := publishPKIWithRunner(ctx, gen.BootstrapPreflightInput{Lifecycle: &gen.LifecycleInput{}}, nil, PKIPublicationInput{},
		execution.KubectlConfig{}, execution.Policy{MaxDuration: time.Minute, OperationTimeout: time.Second}, rejectedPublicationRunner{})
	require.ErrorIs(t, err, context.Canceled)
}
