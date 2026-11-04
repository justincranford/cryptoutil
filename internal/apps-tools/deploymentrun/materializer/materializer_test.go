// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"io"
	http "net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type materializerResponse struct {
	status int
	body   []byte
	err    error
}

type requestTrace struct {
	method        string
	url           string
	authorization string
	hasDeadline   bool
}

type sequenceDoer struct {
	responses    []materializerResponse
	requests     []requestTrace
	cancelOnCall int
	cancel       context.CancelFunc
}

func (doer *sequenceDoer) Do(request *http.Request) (*http.Response, error) {
	_, hasDeadline := request.Context().Deadline()

	doer.requests = append(doer.requests, requestTrace{
		method: request.Method, url: request.URL.String(),
		authorization: request.Header.Get("Authorization"), hasDeadline: hasDeadline,
	})
	if len(doer.requests) > len(doer.responses) {
		return nil, errors.New("unexpected additional Secret request")
	}

	response := doer.responses[len(doer.requests)-1]
	if doer.cancel != nil && len(doer.requests) == doer.cancelOnCall {
		doer.cancel()
	}

	if response.err != nil {
		return nil, response.err
	}

	return &http.Response{
		StatusCode: response.status, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(response.body)),
		ContentLength: -1,
	}, nil
}

type materializerHarness struct {
	fixture       *testPKIFixture
	configYAML    []byte
	inputs        RuntimeInputs
	trust         TrustInputs
	projected     map[string][]byte
	readPaths     []string
	readErrorPath string
	readError     error
	output        string
	fileSystem    FileSystem
	doer          *sequenceDoer
}

func newMaterializerHarness(t *testing.T, fixture *testPKIFixture) *materializerHarness {
	t.Helper()

	current := cloneTestPKIFixture(fixture)
	harness := &materializerHarness{
		fixture:    current,
		configYAML: []byte(validInstanceYAML),
		inputs: RuntimeInputs{
			PodName: "identity-authz-pg-1", Namespace: testNamespace,
			APIEndpoint: testAPIEndpoint, APIServerName: testAPIServerName,
			ExpectedAdminDNS:       testExpectedAdministrativeDNS(t, current.config, current.service),
			ExpectedPublicDNSNames: testExpectedPublicDNSNames(current.config, current.service, testNamespace),
		},
		trust: current.trust,
		projected: map[string][]byte{
			cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath: current.apiCAPEM, cryptoutilSharedMagic.DeploymentMaterializerProjectedTokenPath: []byte("projected-token-value"),
		},
		readError:  errors.New("injected projected file read failure"),
		output:     filepath.Join(t.TempDir(), "pki-output"),
		fileSystem: operatingSystemFileSystem{},
	}
	harness.doer = &sequenceDoer{responses: validMaterializerResponses(t, current)}

	return harness
}

func (harness *materializerHarness) materializer() *Materializer {
	return New(Dependencies{
		HTTPDoer: harness.doer,
		ReadFile: func(name string) ([]byte, error) {
			harness.readPaths = append(harness.readPaths, name)
			if name == harness.readErrorPath {
				return nil, harness.readError
			}

			data, exists := harness.projected[name]
			if !exists {
				return nil, errors.New("projected file not found")
			}

			return append([]byte(nil), data...), nil
		},
		FileSystem:      harness.fileSystem,
		Now:             func() time.Time { return harness.fixture.now },
		OutputDirectory: harness.output,
	})
}

func validMaterializerResponses(t *testing.T, fixture *testPKIFixture) []materializerResponse {
	t.Helper()

	objects := []apiSecret{
		materializerSecret(fixture.config, fixture.service, fixture.service.TLSLeafSecretName, roleTLSBundle, fixture.tlsBundle),
		materializerSecret(fixture.config, fixture.service, fixture.service.HealthClientSecretName, roleHealth, fixture.health),
		materializerSecret(fixture.config, fixture.service, fixture.service.OTLPClientSecretName, roleOTLP, fixture.otlp),
	}

	responses := make([]materializerResponse, 0, len(objects))
	for _, object := range objects {
		body, err := json.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}

		responses = append(responses, materializerResponse{status: http.StatusOK, body: body})
	}

	return responses
}

func materializerSecret(
	config InstanceConfig,
	service InstanceService,
	name string,
	role string,
	data map[string][]byte,
) apiSecret {
	return apiSecret{
		APIVersion: "v1", Kind: cryptoutilSharedMagic.DeploymentMaterializerKubernetesSecretKind, Type: cryptoutilSharedMagic.DeploymentMaterializerSecretTypeOpaque,
		Metadata: apiSecretMetadata{
			Name: name, Namespace: testNamespace,
			Labels: map[string]string{
				labelWorkload: config.PSID + "-" + config.Pool, labelPSID: config.PSID, labelPool: config.Pool,
				labelRun: config.Run, labelRelease: config.Release, labelRole: role,
			},
		},
		Data: cloneByteMap(data),
	}
}

func TestMaterializePublishesOnlyCurrentPodRoleFiles(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	harness := newMaterializerHarness(t, fixture)
	err := harness.materializer().Materialize(context.Background(), harness.configYAML, harness.inputs, harness.trust)
	require.NoError(t, err)
	require.Len(t, harness.doer.requests, 3)
	require.Equal(t, []string{
		testAPIEndpoint + "/api/v1/namespaces/namespace-17/secrets/identity-authz-pg-tls-1",
		testAPIEndpoint + "/api/v1/namespaces/namespace-17/secrets/identity-authz-pg-health-1",
		testAPIEndpoint + "/api/v1/namespaces/namespace-17/secrets/identity-authz-pg-otlp-1",
	}, []string{
		harness.doer.requests[0].url, harness.doer.requests[1].url, harness.doer.requests[2].url,
	})

	for _, request := range harness.doer.requests {
		require.Equal(t, http.MethodGet, request.method)
		require.Equal(t, "Bearer projected-token-value", request.authorization)
		require.True(t, request.hasDeadline)
	}

	require.Equal(t, []string{cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath, cryptoutilSharedMagic.DeploymentMaterializerProjectedTokenPath}, harness.readPaths)

	expectedFiles := []string{
		cryptoutilSharedMagic.DeploymentMaterializerPublicServerCertificateFile, cryptoutilSharedMagic.DeploymentMaterializerPublicServerPrivateKeyFile,
		cryptoutilSharedMagic.DeploymentMaterializerAdminServerCertificateFile, cryptoutilSharedMagic.DeploymentMaterializerAdminServerPrivateKeyFile,
		cryptoutilSharedMagic.DeploymentMaterializerHealthClientCertificateFile, cryptoutilSharedMagic.DeploymentMaterializerHealthClientPrivateKeyFile,
		cryptoutilSharedMagic.DeploymentMaterializerOTLPClientCertificateFile, cryptoutilSharedMagic.DeploymentMaterializerOTLPClientPrivateKeyFile,
		cryptoutilSharedMagic.DeploymentMaterializerPostgreSQLClientCertificateFile, cryptoutilSharedMagic.DeploymentMaterializerPostgreSQLClientPrivateKeyFile,
	}
	for _, relativeName := range expectedFiles {
		name := filepath.Join(harness.output, filepath.FromSlash(relativeName))
		content, readErr := os.ReadFile(name)
		require.NoError(t, readErr)
		require.NotEmpty(t, content)

		info, statErr := os.Stat(name)
		require.NoError(t, statErr)
		requirePlatformFileMode(t, os.FileMode(privateFileMode), info.Mode().Perm())
	}

	require.NoFileExists(t, filepath.Join(harness.output, "ca.crt"))
	require.NoFileExists(t, filepath.Join(harness.output, "root.key"))
	require.NotContains(t, expectedFiles, "root.key")
	require.DirExists(t, filepath.Join(harness.output, cryptoutilSharedMagic.DeploymentMaterializerSecretRoleHealth))
	require.DirExists(t, filepath.Join(harness.output, cryptoutilSharedMagic.DeploymentMaterializerSecretRoleOTLP))
	require.Equal(t, expectedHealthClientDirectory, cryptoutilSharedMagic.DeploymentMaterializerHealthClientDirectory)
	require.Equal(t, expectedOTLPClientDirectory, cryptoutilSharedMagic.DeploymentMaterializerOTLPClientDirectory)
}

func TestMaterializeFailsClosedBeforePublishing(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	tests := []struct {
		name       string
		mutate     func(*testing.T, *materializerHarness, context.Context) context.Context
		wantErr    string
		wantReads  int
		wantCalls  int
		keepOutput bool
	}{
		{
			name: "unknown YAML field",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.configYAML = append(harness.configYAML, []byte("unrecognized: true\n")...)

				return ctx
			},
			wantErr: "field unrecognized not found",
		},
		{
			name: "pod ordinal not found",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.inputs.PodName = "identity-authz-pg-9"

				return ctx
			},
			wantErr: "exactly one service entry",
		},
		{
			name: "invalid namespace",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.inputs.Namespace = testInvalidNamespaceValue

				return ctx
			},
			wantErr: "Downward API pod identity is invalid",
		},
		{
			name: "invalid expected admin DNS",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.inputs.ExpectedAdminDNS = "../localhost"

				return ctx
			},
			wantErr: "expected administrative DNS name is invalid",
		},
		{
			name: "expected admin DNS differs from derived pod identity",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.inputs.ExpectedAdminDNS = testWrongDNSName

				return ctx
			},
			wantErr: "expected administrative DNS name differs from the derived pod identity",
		},
		{
			name: "expected public SAN list omits derived pod FQDN",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.inputs.ExpectedPublicDNSNames = harness.inputs.ExpectedPublicDNSNames[:1]

				return ctx
			},
			wantErr: testMissingDerivedServiceOrPodFQDNError,
		},
		{
			name: "invalid controller supplied public SNI",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.inputs.ExpectedPublicDNSNames = append(harness.inputs.ExpectedPublicDNSNames, "bad/name")

				return ctx
			},
			wantErr: "expected public DNS identity is invalid",
		},
		{
			name: "unsafe output directory is rejected before API reads",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.output = "relative/output"

				return ctx
			},
			wantErr: "clean absolute path",
		},
		{
			name: "missing trusted health anchors",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.trust.HealthClientPEM = nil

				return ctx
			},
			wantErr: "trusted CA anchors for health-client",
		},
		{
			name: "canceled context",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				canceled, cancel := context.WithCancel(ctx)
				cancel()

				return canceled
			},
			wantErr: "canceled before initialization",
		},
		{
			name: "reused output rejected before Secret reads",
			mutate: func(t *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				require.NoError(t, os.Mkdir(harness.output, privateDirectoryMode))

				return ctx
			},
			wantErr:    "already exists",
			keepOutput: true,
		},
		{
			name: "API CA file read failure",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.readErrorPath = cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath

				return ctx
			},
			wantErr:   "read projected Kubernetes API CA bundle: injected projected file read failure",
			wantReads: 1,
		},
		{
			name: "malformed API CA",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.projected[cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath] = []byte("not PEM")

				return ctx
			},
			wantErr:   testParseProjectedKubernetesAPICABundleError,
			wantReads: 1,
		},
		{
			name: "oversized API CA",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.projected[cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath] = bytes.Repeat([]byte("x"), cryptoutilSharedMagic.DeploymentMaterializerMaximumAPICABundleBytes+1)

				return ctx
			},
			wantErr:   "empty or oversized",
			wantReads: 1,
		},
		{
			name: "service token file read failure",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.readErrorPath = cryptoutilSharedMagic.DeploymentMaterializerProjectedTokenPath

				return ctx
			},
			wantErr:   "read projected service-account token: injected projected file read failure",
			wantReads: 2,
		},
		{
			name: "malformed service token",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.projected[cryptoutilSharedMagic.DeploymentMaterializerProjectedTokenPath] = []byte("invalid token")

				return ctx
			},
			wantErr:   "empty or malformed",
			wantReads: 2,
		},
		{
			name: "API forbids access",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.doer.responses[0].status = http.StatusForbidden

				return ctx
			},
			wantErr:   "HTTP 403",
			wantReads: 2,
			wantCalls: 1,
		},
		{
			name: "metadata identity mismatch",
			mutate: func(t *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.mutateSecretResponse(t, 0, func(secret *apiSecret) {
					secret.Metadata.Namespace = "other-namespace"
				})

				return ctx
			},
			wantErr:   testMetadataIdentityError,
			wantReads: 2,
			wantCalls: 1,
		},
		{
			name: "Secret transport failure",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.doer.responses[0].err = errors.New("injected transport detail")

				return ctx
			},
			wantErr:   "transport failed",
			wantReads: 2,
			wantCalls: 1,
		},
		{
			name: "certificate failure publishes nothing",
			mutate: func(t *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.mutateSecretResponse(t, 1, func(secret *apiSecret) {
					delete(secret.Data, cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey)
				})

				return ctx
			},
			wantErr:   "health client Secret data is invalid",
			wantReads: 2,
			wantCalls: testExpectedSecretRequestCount,
		},
		{
			name: "output failure cleans staging",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				harness.fileSystem = &faultFileSystem{
					base: operatingSystemFileSystem{}, outputDirectory: harness.output, failOperation: testFileOperationOpen,
				}

				return ctx
			},
			wantErr:   "create materializer output file",
			wantReads: 2,
			wantCalls: 3,
		},
		{
			name: "cancellation after Secret response",
			mutate: func(_ *testing.T, harness *materializerHarness, ctx context.Context) context.Context {
				canceled, cancel := context.WithCancel(ctx)
				harness.doer.cancel = cancel
				harness.doer.cancelOnCall = 1

				return canceled
			},
			wantErr:   "canceled",
			wantReads: 2,
			wantCalls: 1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			harness := newMaterializerHarness(t, fixture)

			ctx := context.Background()
			if testCase.mutate != nil {
				ctx = testCase.mutate(t, harness, ctx)
			}

			err := harness.materializer().Materialize(ctx, harness.configYAML, harness.inputs, harness.trust)
			require.ErrorContains(t, err, testCase.wantErr)
			require.Len(t, harness.readPaths, testCase.wantReads)
			require.Len(t, harness.doer.requests, testCase.wantCalls)

			if harness.readErrorPath != "" {
				require.ErrorIs(t, err, harness.readError)
			}

			if !testCase.keepOutput {
				require.NoDirExists(t, harness.output)
			}
		})
	}
}

func (harness *materializerHarness) mutateSecretResponse(
	t *testing.T,
	index int,
	mutate func(*apiSecret),
) {
	t.Helper()

	var secret apiSecret

	err := json.Unmarshal(harness.doer.responses[index].body, &secret)
	require.NoError(t, err)
	mutate(&secret)
	harness.doer.responses[index].body, err = json.Marshal(secret)
	require.NoError(t, err)
}
