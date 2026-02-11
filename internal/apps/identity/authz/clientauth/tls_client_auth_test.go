// Copyright (c) 2025 Justin Cranford

package clientauth

import (
	"context"
	sha256 "crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestTLSClientAuthenticator_Authenticate_Cert(t *testing.T) {
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
			name: "valid certificate",
			client: &cryptoutilIdentityDomain.Client{
				ID:                      mustNewUUID(),
				ClientID:                "tls-client",
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth,
				CertificateSubject:      "Test Client",
				CertificateFingerprint:  fingerprint,
				Enabled:                 boolPtr(true),
			},
			clientID:   "tls-client",
			credential: string(clientPEM),
			wantErr:    false,
		},
		{
			name: "missing certificate",
			client: &cryptoutilIdentityDomain.Client{
				ID:                      mustNewUUID(),
				ClientID:                "tls-client",
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth,
				Enabled:                 boolPtr(true),
			},
			clientID:   "tls-client",
			credential: "",
			wantErr:    true,
		},
		{
			name: "subject mismatch",
			client: &cryptoutilIdentityDomain.Client{
				ID:                      mustNewUUID(),
				ClientID:                "tls-client",
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth,
				CertificateSubject:      "Wrong Subject",
				Enabled:                 boolPtr(true),
			},
			clientID:   "tls-client",
			credential: string(clientPEM),
			wantErr:    true,
		},
		{
			name: "fingerprint mismatch",
			client: &cryptoutilIdentityDomain.Client{
				ID:                      mustNewUUID(),
				ClientID:                "tls-client",
				TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth,
				CertificateFingerprint:  "0000000000000000000000000000000000000000000000000000000000000000",
				Enabled:                 boolPtr(true),
			},
			clientID:   "tls-client",
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
			authenticator := NewTLSClientAuthenticator(mockRepo, validator)

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

func TestTLSClientAuthenticator_Method_Cert(t *testing.T) {
	t.Parallel()

	mockRepo := &mockClientRepository{}
	trustedCAs := x509.NewCertPool()
	validator := NewCACertificateValidator(trustedCAs, nil)
	authenticator := NewTLSClientAuthenticator(mockRepo, validator)

	method := authenticator.Method()
	require.Equal(t, string(cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth), method)
}

func encodeCertToPEM(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
}

func computeSHA256Fingerprint(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.Raw)

	return hex.EncodeToString(hash[:])
}
