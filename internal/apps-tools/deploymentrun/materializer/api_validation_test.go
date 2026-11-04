// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	"context"
	"errors"
	http "net/http"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestFetchSecretRejectsInvalidRequestIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		namespace string
		secret    string
	}{
		{name: testInvalidNamespaceCaseName, namespace: testInvalidNamespaceValue, secret: testHealthSecretName},
		{name: testInvalidSecretNameCaseName, namespace: testNamespace, secret: testParentHealthSecretName},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			called := false
			_, err := fetchSecretAtTestEndpoint(
				context.Background(),
				testDoer(func(*http.Request) (*http.Response, error) {
					called = true

					return nil, nil
				}),
				cryptoutilSharedMagic.ParamToken, testCase.namespace, testCase.secret, InstanceConfig{}, InstanceService{}, roleHealth,
			)
			require.ErrorContains(t, err, testInvalidIdentityError)
			require.False(t, called)
		})
	}
}

func TestValidLabelKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{name: "bare name", key: "publisher", valid: true},
		{name: "qualified name", key: "example.com/publisher", valid: true},
		{name: "invalid suffix", key: "example.com/not-valid-", valid: false},
		{name: "invalid prefix", key: "bad_prefix.example/publisher", valid: false},
		{name: "multiple separators", key: "example.com/one/two", valid: false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, testCase.valid, validLabelKey(testCase.key))
		})
	}
}

func TestSecretURLAndTokenValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		endpoint  string
		namespace string
		secret    string
		wantErr   string
	}{
		{name: "fixed configured API IP endpoint", namespace: testNamespace, secret: testHealthSecretName},
		{
			name:      "invalid endpoint",
			endpoint:  testUntrustedKubernetesDNSAPIEndpoint,
			namespace: testNamespace, secret: testHealthSecretName,
			wantErr: testEndpointOrTLSServerNameError,
		},
		{name: testInvalidNamespaceCaseName, namespace: testInvalidNamespaceValue, secret: testHealthSecretName, wantErr: testInvalidIdentityError},
		{name: testInvalidSecretNameCaseName, namespace: testNamespace, secret: testParentHealthSecretName, wantErr: testInvalidIdentityError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			endpoint := testCase.endpoint
			if endpoint == "" {
				endpoint = testAPIEndpoint
			}

			requestURL, err := makeSecretURL(endpoint, testCase.namespace, testCase.secret)
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testAPIEndpoint+"/api/v1/namespaces/namespace-17/secrets/"+testHealthSecretName, requestURL)
		})
	}

	tokenCases := []struct {
		name    string
		input   []byte
		want    string
		wantErr string
	}{
		{name: "trim projected newline", input: []byte(" token-value \n"), want: "token-value"},
		{name: cryptoutilSharedMagic.EMPTY, input: []byte(" \n"), wantErr: testMalformedError},
		{name: "embedded whitespace", input: []byte("token value"), wantErr: testMalformedError},
		{name: "oversized", input: bytes.Repeat([]byte("x"), cryptoutilSharedMagic.DeploymentMaterializerMaximumTokenBytes+1), wantErr: testSizeLimitError},
	}
	for _, testCase := range tokenCases {
		t.Run("token "+testCase.name, func(t *testing.T) {
			t.Parallel()

			token, err := validateToken(testCase.input)
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.want, token)
		})
	}
}

func TestRedactErrorPreservesCause(t *testing.T) {
	t.Parallel()

	cause := errors.New(testNetworkFailureMessage)

	tests := []struct {
		name        string
		cause       error
		secret      string
		wantMessage string
		wantNil     bool
		sameCause   bool
	}{
		{name: "nil error", secret: cryptoutilSharedMagic.ParamToken, wantNil: true},
		{name: "empty secret", cause: cause, wantMessage: testNetworkFailureMessage, sameCause: true},
		{name: "cause does not contain token", cause: cause, secret: cryptoutilSharedMagic.ParamToken, wantMessage: testNetworkFailureMessage, sameCause: true},
		{
			name:        "token is redacted and cause remains unwrap-able",
			cause:       errors.New("transport included bearer-token"),
			secret:      "bearer-token",
			wantMessage: "transport included [REDACTED]",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			redacted := redactError(testCase.cause, testCase.secret)
			if testCase.wantNil {
				require.NoError(t, redacted)

				return
			}

			require.EqualError(t, redacted, testCase.wantMessage)

			if testCase.sameCause {
				require.Equal(t, testCase.cause, redacted)
			} else {
				require.ErrorIs(t, redacted, testCase.cause)
			}
		})
	}
}

func TestValidateAPIEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		endpoint   string
		serverName string
		wantErr    bool
	}{
		{name: "configured IPv4 service endpoint", endpoint: testAPIEndpoint, serverName: testAPIServerName},
		{name: "configured IPv4 endpoint without explicit SNI", endpoint: testAPIEndpoint},
		{name: "configured IPv6 service endpoint", endpoint: "https://[2001:db8::1]:443"},
		{name: "HTTP is rejected", endpoint: "http://192.0.2.10:443", wantErr: true},
		{name: "DNS endpoint is rejected", endpoint: testUntrustedKubernetesDNSAPIEndpoint, wantErr: true},
		{name: "wrong port is rejected", endpoint: "https://192.0.2.10:8443", wantErr: true},
		{name: "userinfo is rejected", endpoint: "https://user@192.0.2.10:443", wantErr: true},
		{name: "API path is rejected", endpoint: "https://192.0.2.10:443/api", wantErr: true},
		{name: "query is rejected", endpoint: "https://192.0.2.10:443?token=value", wantErr: true},
		{name: "empty query is rejected", endpoint: "https://192.0.2.10:443?", wantErr: true},
		{name: "fragment is rejected", endpoint: "https://192.0.2.10:443#secret", wantErr: true},
		{name: "malformed SNI is rejected", endpoint: testAPIEndpoint, serverName: testInvalidAPIServerName, wantErr: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			endpointURL, err := validateAPIEndpoint(testCase.endpoint, testCase.serverName)
			if testCase.wantErr {
				require.Error(t, err)
				require.Nil(t, endpointURL)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, endpointURL)
			require.Equal(t, cryptoutilSharedMagic.ProtocolHTTPS, endpointURL.Scheme)
			require.Equal(t, "443", endpointURL.Port())
		})
	}
}
