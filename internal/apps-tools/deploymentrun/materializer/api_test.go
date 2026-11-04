// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	"context"
	"crypto/tls"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	http "net/http"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type testDoer func(*http.Request) (*http.Response, error)

func (doer testDoer) Do(request *http.Request) (*http.Response, error) {
	return doer(request)
}

type testReadCloser struct {
	reader   io.Reader
	readErr  error
	closeErr error
	closed   bool
	cancel   context.CancelFunc
}

func (body *testReadCloser) Read(buffer []byte) (int, error) {
	if body.cancel != nil {
		body.cancel()
	}

	if body.readErr != nil {
		return 0, body.readErr
	}

	return body.reader.Read(buffer)
}

func (body *testReadCloser) Close() error {
	body.closed = true

	return body.closeErr
}

func TestNewAPIHTTPClient(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	unapprovedCAPEM := newTestUnapprovedCAPEM(t, fixture)

	tests := []struct {
		name       string
		caPEM      []byte
		serverName string
		wantErr    string
	}{
		{name: "explicit cluster CA with TLS 1.3", caPEM: fixture.apiCAPEM, serverName: testAPIServerName},
		{name: "malformed cluster CA", caPEM: []byte("not PEM"), serverName: testAPIServerName, wantErr: testParseProjectedKubernetesAPICABundleError},
		{name: "leaf is not an API CA", caPEM: fixture.tlsBundle[cryptoutilSharedMagic.DeploymentMaterializerSecretDataPublicCertificate], serverName: testAPIServerName, wantErr: "non-CA"},
		{name: "unapproved CA key", caPEM: unapprovedCAPEM, serverName: testAPIServerName, wantErr: "not FIPS-approved"},
		{
			name:       "invalid TLS server name",
			caPEM:      fixture.apiCAPEM,
			serverName: testInvalidAPIServerName,
			wantErr:    "invalid Kubernetes API TLS server name",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client, err := newAPIHTTPClient(testCase.caPEM, testCase.serverName)
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)

			transport, ok := client.Transport.(*http.Transport)
			require.True(t, ok)
			require.Nil(t, transport.Proxy)
			require.Equal(t, uint16(tls.VersionTLS13), transport.TLSClientConfig.MinVersion)
			require.Equal(t, uint16(tls.VersionTLS13), transport.TLSClientConfig.MaxVersion)
			require.False(t, transport.TLSClientConfig.InsecureSkipVerify)
			require.NotNil(t, transport.TLSClientConfig.RootCAs)
			require.Equal(t, testAPIServerName, transport.TLSClientConfig.ServerName)

			redirectErr := client.CheckRedirect(&http.Request{}, nil)
			require.ErrorIs(t, redirectErr, http.ErrUseLastResponse)
		})
	}
}

func TestFetchSecret(t *testing.T) {
	t.Parallel()

	fixture := newTestPKIFixture(t)
	tests := []struct {
		name       string
		status     int
		body       string
		contentLen int64
		mutate     func(*apiSecret)
		wantErr    string
	}{
		{name: "exact named Secret and request contract", status: http.StatusOK},
		{name: "unauthorized status", status: http.StatusUnauthorized, wantErr: "HTTP 401"},
		{name: "forbidden status", status: http.StatusForbidden, wantErr: testForbiddenHTTPStatus},
		{name: "missing Secret status", status: http.StatusNotFound, wantErr: "HTTP 404"},
		{name: "redirect status is not followed", status: http.StatusFound, wantErr: "HTTP 302"},
		{name: "server error status", status: http.StatusInternalServerError, wantErr: "HTTP 500"},
		{name: "malformed JSON", status: http.StatusOK, body: "{invalid", wantErr: testDecodeKubernetesSecretError},
		{name: "malformed base64", status: http.StatusOK, body: `{"apiVersion":"v1","kind":"Secret","metadata":{"name":"health-1","namespace":"namespace-17","labels":{}},"type":"` + cryptoutilSharedMagic.DeploymentMaterializerSecretTypeOpaque + `","data":{"tls.crt":"%%%="}}`, wantErr: testDecodeKubernetesSecretError},
		{name: "metadata name mismatch", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Metadata.Name = testMismatchedIdentityValue }, wantErr: testMetadataIdentityError},
		{name: "namespace mismatch", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Metadata.Namespace = testMismatchedIdentityValue }, wantErr: testMetadataIdentityError},
		{name: "wrong workload role", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Metadata.Labels[labelRole] = roleTLSBundle }, wantErr: testIdentityLabelError},
		{name: "wrong workload name", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Metadata.Labels[labelWorkload] = testMismatchedIdentityValue }, wantErr: testIdentityLabelError},
		{name: "wrong PSID label", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Metadata.Labels[labelPSID] = testMismatchedIdentityValue }, wantErr: testIdentityLabelError},
		{name: "wrong pool label", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Metadata.Labels[labelPool] = testMismatchedIdentityValue }, wantErr: testIdentityLabelError},
		{name: "wrong run label", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Metadata.Labels[labelRun] = testMismatchedIdentityValue }, wantErr: testIdentityLabelError},
		{name: "wrong release label", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Metadata.Labels[labelRelease] = testMismatchedIdentityValue }, wantErr: testIdentityLabelError},
		{name: "missing required label", status: http.StatusOK, mutate: func(secret *apiSecret) { delete(secret.Metadata.Labels, labelPool) }, wantErr: testIdentityLabelError},
		{name: testMalformedLabelError, status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Metadata.Labels["bad/key/shape"] = testLabelValue }, wantErr: testMalformedLabelError},
		{name: "malformed label value", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Metadata.Labels["other.example/key"] = "invalid value" }, wantErr: testMalformedLabelError},
		{name: "empty label key", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Metadata.Labels[""] = testLabelValue }, wantErr: testMalformedLabelError},
		{name: "valid additional label", status: http.StatusOK, mutate: func(secret *apiSecret) {
			secret.Metadata.Labels["publisher"] = cryptoutilSharedMagic.DefaultOTLPServiceDefault
		}},
		{name: "wrong API version", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.APIVersion = "v2" }, wantErr: testOpaqueV1SecretError},
		{name: "wrong object kind", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Kind = "ConfigMap" }, wantErr: testOpaqueV1SecretError},
		{name: "wrong API object type", status: http.StatusOK, mutate: func(secret *apiSecret) { secret.Type = "kubernetes.io/tls" }, wantErr: testOpaqueV1SecretError},
		{name: "stringData is rejected", status: http.StatusOK, mutate: func(secret *apiSecret) {
			secret.StringData = map[string]string{cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey: "not-returned-by-api"}
		}, wantErr: "stringData"},
		{name: "declared body too large", status: http.StatusOK, contentLen: cryptoutilSharedMagic.DeploymentMaterializerMaximumSecretBodyBytes + 1, wantErr: testSizeLimitError},
		{name: "stream body too large", status: http.StatusOK, body: strings.Repeat("x", int(cryptoutilSharedMagic.DeploymentMaterializerMaximumSecretBodyBytes+1)), wantErr: testSizeLimitError},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			config := fixture.config
			service := fixture.service

			var requestSeen *http.Request

			responseBody := testCase.body
			if responseBody == "" && testCase.status == http.StatusOK {
				secret := validAPISecret(config, service)
				if testCase.mutate != nil {
					testCase.mutate(&secret)
				}

				encoded, err := json.Marshal(secret)
				require.NoError(t, err)

				responseBody = string(encoded)
			}

			doer := testDoer(func(request *http.Request) (*http.Response, error) {
				requestSeen = request
				body := &testReadCloser{reader: strings.NewReader(responseBody)}

				response := &http.Response{
					StatusCode: testCase.status, Body: body, Header: make(http.Header), ContentLength: -1,
				}
				if testCase.contentLen > 0 {
					response.ContentLength = testCase.contentLen
				}

				return response, nil
			})

			data, err := fetchSecretAtTestEndpoint(
				context.Background(), doer, "test-bearer-token", testNamespace, testHealthSecretName,
				config, service, roleHealth,
			)
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)
				require.NotContains(t, err.Error(), "test-bearer-token")

				return
			}

			require.NoError(t, err)
			require.Equal(t, []byte("certificate"), data[cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate])
			require.Equal(t, "identity-authz-pg", validAPISecret(config, service).Metadata.Labels[labelWorkload])
			require.Equal(t, config.PSID+"-"+config.Pool, validAPISecret(config, service).Metadata.Labels[labelWorkload])
			require.NotNil(t, requestSeen)
			require.Equal(t, http.MethodGet, requestSeen.Method)
			require.Equal(t, cryptoutilSharedMagic.ProtocolHTTPS, requestSeen.URL.Scheme)
			require.Equal(t, "192.0.2.10:443", requestSeen.URL.Host)
			require.Equal(t, "/api/v1/namespaces/namespace-17/secrets/"+testHealthSecretName, requestSeen.URL.Path)
			require.Equal(t, "Bearer test-bearer-token", requestSeen.Header.Get("Authorization"))
			_, hasDeadline := requestSeen.Context().Deadline()
			require.True(t, hasDeadline)
		})
	}

	t.Run("transport error is sanitized", func(t *testing.T) {
		t.Parallel()

		transportCause := &net.OpError{Op: "dial", Net: "tcp", Err: io.ErrUnexpectedEOF}
		_, err := fetchSecretAtTestEndpoint(
			context.Background(),
			testDoer(func(*http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("request included test-bearer-token: %w", transportCause)
			}),
			"test-bearer-token", testNamespace, testHealthSecretName, fixture.config, fixture.service, roleHealth,
		)
		require.ErrorContains(t, err, "transport failed")
		require.NotContains(t, err.Error(), "test-bearer-token")
		require.ErrorContains(t, err, "unexpected EOF")
		require.ErrorIs(t, err, transportCause)

		var networkError *net.OpError
		require.ErrorAs(t, err, &networkError)
	})
	t.Run("canceled before request", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		called := false
		_, err := fetchSecretAtTestEndpoint(ctx, testDoer(func(*http.Request) (*http.Response, error) {
			called = true

			return nil, nil
		}), cryptoutilSharedMagic.ParamToken, testNamespace, testHealthSecretName, fixture.config, fixture.service, roleHealth)
		require.ErrorContains(t, err, testCanceledError)
		require.False(t, called)
	})
	t.Run("nil response and body", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			response *http.Response
		}{
			{name: "nil response"},
			{name: "nil body", response: &http.Response{StatusCode: http.StatusOK}},
		}
		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				_, err := fetchSecretAtTestEndpoint(
					context.Background(),
					testDoer(func(*http.Request) (*http.Response, error) { return testCase.response, nil }),
					cryptoutilSharedMagic.ParamToken, testNamespace, testHealthSecretName, fixture.config, fixture.service, roleHealth,
				)
				require.ErrorContains(t, err, "empty response")
			})
		}
	})
	t.Run("body read and close failures", func(t *testing.T) {
		t.Parallel()

		readFailure := errors.New("read failed")
		closeFailure := errors.New("close failed")

		tests := []struct {
			name   string
			body   *testReadCloser
			causes []error
		}{
			{
				name:   "read error",
				body:   &testReadCloser{reader: bytes.NewReader(nil), readErr: readFailure},
				causes: []error{readFailure},
			},
			{
				name:   "close error",
				body:   &testReadCloser{reader: strings.NewReader("{}"), closeErr: closeFailure},
				causes: []error{closeFailure},
			},
			{
				name: "read and close errors",
				body: &testReadCloser{
					reader: bytes.NewReader(nil), readErr: readFailure, closeErr: closeFailure,
				},
				causes: []error{readFailure, closeFailure},
			},
		}
		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				_, err := readBoundedBody(testCase.body, 8)
				require.Error(t, err)
				require.True(t, testCase.body.closed)

				for _, cause := range testCase.causes {
					require.ErrorIs(t, err, cause)
				}

				if strings.Contains(testCase.name, cryptoutilSharedMagic.ScopeRead) {
					require.ErrorContains(t, err, "read failed")
				}

				if strings.Contains(testCase.name, testFileOperationClose) {
					require.ErrorContains(t, err, "close failed")
				}
			})
			t.Run("request cancellation and response close errors", func(t *testing.T) {
				t.Parallel()

				ctx, cancel := context.WithCancel(context.Background())
				_, err := fetchSecretAtTestEndpoint(ctx, testDoer(func(*http.Request) (*http.Response, error) {
					cancel()

					return nil, errors.New("injected transport error")
				}), cryptoutilSharedMagic.ParamToken, testNamespace, testHealthSecretName, fixture.config, fixture.service, roleHealth)
				require.ErrorContains(t, err, testCanceledError)

				bodyContext, bodyCancel := context.WithCancel(context.Background())
				cancelingBody := &testReadCloser{
					reader: strings.NewReader("{}"), readErr: errors.New("injected read failure"), cancel: bodyCancel,
				}
				_, readErr := fetchSecretAtTestEndpoint(bodyContext, testDoer(func(*http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Body: cancelingBody}, nil
				}), cryptoutilSharedMagic.ParamToken, testNamespace, testHealthSecretName, fixture.config, fixture.service, roleHealth)
				require.ErrorContains(t, readErr, testCanceledError)
				require.True(t, cancelingBody.closed)

				errorResponseBody := &testReadCloser{
					reader: strings.NewReader("discarded"), closeErr: errors.New("injected transport response close failure"),
				}
				_, transportErr := fetchSecretAtTestEndpoint(context.Background(), testDoer(func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusInternalServerError, Body: errorResponseBody,
					}, errors.New("injected transport detail")
				}), cryptoutilSharedMagic.ParamToken, testNamespace, testHealthSecretName, fixture.config, fixture.service, roleHealth)
				require.ErrorContains(t, transportErr, "transport failed")
				require.ErrorContains(t, transportErr, "transport response close failure")
				require.True(t, errorResponseBody.closed)

				for _, testCase := range []struct {
					name       string
					status     int
					contentLen int64
					wantErr    string
				}{
					{name: "status response", status: http.StatusForbidden, wantErr: testForbiddenHTTPStatus},
					{name: "oversized response", status: http.StatusOK, contentLen: cryptoutilSharedMagic.DeploymentMaterializerMaximumSecretBodyBytes + 1, wantErr: testSizeLimitError},
				} {
					t.Run(testCase.name, func(t *testing.T) {
						t.Parallel()

						body := &testReadCloser{reader: strings.NewReader("{}"), closeErr: errors.New("injected close error")}
						_, fetchErr := fetchSecretAtTestEndpoint(context.Background(), testDoer(func(*http.Request) (*http.Response, error) {
							return &http.Response{
								StatusCode: testCase.status, Body: body, ContentLength: testCase.contentLen,
							}, nil
						}), cryptoutilSharedMagic.ParamToken, testNamespace, testHealthSecretName, fixture.config, fixture.service, roleHealth)
						require.ErrorContains(t, fetchErr, testCase.wantErr)
						require.ErrorContains(t, fetchErr, "injected close error")
						require.True(t, body.closed)
					})
				}
			})
		}
	})
}

func validAPISecret(config InstanceConfig, service InstanceService) apiSecret {
	return apiSecret{
		APIVersion: cryptoutilSharedMagic.DeploymentMaterializerKubernetesAPIVersion, Kind: cryptoutilSharedMagic.DeploymentMaterializerKubernetesSecretKind, Type: cryptoutilSharedMagic.DeploymentMaterializerSecretTypeOpaque,
		Metadata: apiSecretMetadata{
			Name: testHealthSecretName, Namespace: testNamespace,
			Labels: map[string]string{
				labelWorkload: config.PSID + "-" + config.Pool, labelPSID: config.PSID, labelPool: config.Pool,
				labelRun: config.Run, labelRelease: config.Release, labelRole: roleHealth,
			},
		},
		Data: map[string][]byte{cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate: []byte("certificate"), cryptoutilSharedMagic.DeploymentMaterializerSecretDataPrivateKey: []byte("key")},
	}
}
