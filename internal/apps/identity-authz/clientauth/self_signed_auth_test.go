// Copyright (c) 2025-2026 Justin Cranford.
package clientauth

import (
	"context"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestSelfSignedAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	caCert, caKey := createTestCAForAuth(t)
	clientCert := createTestClientCertForAuth(t, caCert, caKey)
	clientPEM := encodeCertToPEM(clientCert)

	fingerprint := computeSHA256Fingerprint(clientCert)

	tests := []struct {
		name       string
		client     *cryptoutilIdentityDomain.Client
		clientID   string
		credential string
		wantErr    bool
	}{
		{
			name: "valid self-signed certificate",
			client: &cryptoutilIdentityDomain.Client{
				ID:                      mustNewUUID(),
				ClientID:                testSelfSignedClient,
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth,
				CertificateSubject:      cryptoutilSharedMagic.TestClientName,
				CertificateFingerprint:  fingerprint,
				Enabled:                 boolPtr(true),
			},
			clientID:   testSelfSignedClient,
			credential: string(clientPEM),
			wantErr:    false,
		},
		{
			name: testMissingCertificate,
			client: &cryptoutilIdentityDomain.Client{
				ID:                      mustNewUUID(),
				ClientID:                testSelfSignedClient,
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth,
				Enabled:                 boolPtr(true),
			},
			clientID:   testSelfSignedClient,
			credential: "",
			wantErr:    true,
		},
		{
			name: testSubjectMismatch,
			client: &cryptoutilIdentityDomain.Client{
				ID:                      mustNewUUID(),
				ClientID:                testSelfSignedClient,
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth,
				CertificateSubject:      testWrongSubject,
				Enabled:                 boolPtr(true),
			},
			clientID:   testSelfSignedClient,
			credential: string(clientPEM),
			wantErr:    true,
		},
		{
			name: testFingerprintMismatch,
			client: &cryptoutilIdentityDomain.Client{
				ID:                      mustNewUUID(),
				ClientID:                testSelfSignedClient,
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth,
				CertificateFingerprint:  "0000000000000000000000000000000000000000000000000000000000000000",
				Enabled:                 boolPtr(true),
			},
			clientID:   testSelfSignedClient,
			credential: string(clientPEM),
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockClientRepository{clients: []*cryptoutilIdentityDomain.Client{tc.client}}
			trustedCAs := x509.NewCertPool()
			trustedCAs.AddCert(caCert)
			validator := NewCACertificateValidator(trustedCAs, nil)
			validator.SetValidationOptions(false, false)
			authenticator := NewSelfSignedAuthenticator(mockRepo, validator)

			client, err := authenticator.Authenticate(ctx, tc.clientID, tc.credential)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, client)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, client)
			require.Equal(t, tc.clientID, client.ClientID)
		})
	}
}

func TestSelfSignedAuthenticator_Method_Custom(t *testing.T) {
	t.Parallel()

	mockRepo := &mockClientRepository{}
	trustedCAs := x509.NewCertPool()
	validator := NewCACertificateValidator(trustedCAs, nil)
	authenticator := NewSelfSignedAuthenticator(mockRepo, validator)

	method := authenticator.Method()
	require.Equal(t, string(cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth), method)
}
