// Copyright (c) 2025 Justin Cranford

package revocation

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ocsp"

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
)

func TestNewOCSPService(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	responderCert, responderKey := createTestCA(t) // Using same for simplicity.
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	crlConfig := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}
	crlSvc, _ := NewCRLService(crlConfig)

	tests := []struct {
		name       string
		config     *OCSPConfig
		crlService *CRLService
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "nil-config",
			config:     nil,
			crlService: crlSvc,
			wantErr:    true,
			errMsg:     "config is required",
		},
		{
			name: "nil-issuer",
			config: &OCSPConfig{
				Issuer:       nil,
				Responder:    responderCert,
				ResponderKey: responderKey,
				Provider:     provider,
			},
			crlService: crlSvc,
			wantErr:    true,
			errMsg:     "issuer certificate is required",
		},
		{
			name: "nil-responder",
			config: &OCSPConfig{
				Issuer:       caCert,
				Responder:    nil,
				ResponderKey: responderKey,
				Provider:     provider,
			},
			crlService: crlSvc,
			wantErr:    true,
			errMsg:     "responder certificate is required",
		},
		{
			name: "nil-responder-key",
			config: &OCSPConfig{
				Issuer:       caCert,
				Responder:    responderCert,
				ResponderKey: nil,
				Provider:     provider,
			},
			crlService: crlSvc,
			wantErr:    true,
			errMsg:     "responder private key is required",
		},
		{
			name: "nil-provider",
			config: &OCSPConfig{
				Issuer:       caCert,
				Responder:    responderCert,
				ResponderKey: responderKey,
				Provider:     nil,
			},
			crlService: crlSvc,
			wantErr:    true,
			errMsg:     "crypto provider is required",
		},
		{
			name: "nil-crl-service",
			config: &OCSPConfig{
				Issuer:       caCert,
				Responder:    responderCert,
				ResponderKey: responderKey,
				Provider:     provider,
				Validity:     time.Hour,
			},
			crlService: nil,
			wantErr:    true,
			errMsg:     "CRL service is required",
		},
		{
			name: "valid-config",
			config: &OCSPConfig{
				Issuer:       caCert,
				Responder:    responderCert,
				ResponderKey: responderKey,
				Provider:     provider,
				Validity:     time.Hour,
			},
			crlService: crlSvc,
			wantErr:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, err := NewOCSPService(tc.config, tc.crlService)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				require.Nil(t, svc)
			} else {
				require.NoError(t, err)
				require.NotNil(t, svc)
			}
		})
	}
}

func TestOCSPService_CreateResponse(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	crlConfig := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}
	crlSvc, err := NewCRLService(crlConfig)
	require.NoError(t, err)

	ocspConfig := &OCSPConfig{
		Issuer:       caCert,
		Responder:    caCert, // Self-signed for test.
		ResponderKey: caKey,
		Provider:     provider,
		Validity:     time.Hour,
	}
	ocspSvc, err := NewOCSPService(ocspConfig, crlSvc)
	require.NoError(t, err)

	// Create a test certificate.
	testCert := createTestCertificate(t, caCert, caKey)

	// Test good response.
	response, err := ocspSvc.CreateResponse(testCert)
	require.NoError(t, err)
	require.NotEmpty(t, response)

	// Parse response.
	parsed, err := ocsp.ParseResponse(response, caCert)
	require.NoError(t, err)
	require.Equal(t, ocsp.Good, parsed.Status)

	// Revoke the certificate.
	err = crlSvc.Revoke(testCert.SerialNumber, ReasonKeyCompromise)
	require.NoError(t, err)

	// Test revoked response.
	response, err = ocspSvc.CreateResponse(testCert)
	require.NoError(t, err)

	parsed, err = ocsp.ParseResponse(response, caCert)
	require.NoError(t, err)
	require.Equal(t, ocsp.Revoked, parsed.Status)
	require.Equal(t, int(ReasonKeyCompromise), parsed.RevocationReason)
}

func TestOCSPService_CreateResponse_NilCert(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	crlConfig := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}
	crlSvc, err := NewCRLService(crlConfig)
	require.NoError(t, err)

	ocspConfig := &OCSPConfig{
		Issuer:       caCert,
		Responder:    caCert,
		ResponderKey: caKey,
		Provider:     provider,
		Validity:     time.Hour,
	}
	ocspSvc, err := NewOCSPService(ocspConfig, crlSvc)
	require.NoError(t, err)

	_, err = ocspSvc.CreateResponse(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "certificate is required")
}

func TestOCSPService_ParseRequest(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	crlConfig := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}
	crlSvc, err := NewCRLService(crlConfig)
	require.NoError(t, err)

	ocspConfig := &OCSPConfig{
		Issuer:       caCert,
		Responder:    caCert,
		ResponderKey: caKey,
		Provider:     provider,
		Validity:     time.Hour,
	}
	ocspSvc, err := NewOCSPService(ocspConfig, crlSvc)
	require.NoError(t, err)

	// Test empty request.
	_, err = ocspSvc.ParseRequest(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty OCSP request")

	// Test invalid request.
	_, err = ocspSvc.ParseRequest([]byte("invalid"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse")
}

func TestRevokedCertificate_ToEntry(t *testing.T) {
	t.Parallel()

	revTime := time.Now().UTC()
	rc := &RevokedCertificate{
		SerialNumber:   big.NewInt(12345),
		RevocationTime: revTime,
		Reason:         ReasonKeyCompromise,
	}

	entry := rc.ToEntry("CN=Test CA")
	require.Equal(t, "3039", entry.SerialNumber)
	require.Equal(t, revTime, entry.RevocationTime)
	require.Equal(t, "keyCompromise", entry.Reason)
	require.Equal(t, int(ReasonKeyCompromise), entry.ReasonCode)
	require.Equal(t, "CN=Test CA", entry.IssuerDN)
}

// Helper functions.

func createTestCA(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test CA",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageOCSPSigning},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert, privateKey
}

func createTestCertificate(t *testing.T, issuer *x509.Certificate, issuerKey *ecdsa.PrivateKey) *x509.Certificate {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	serial, err := crand.Int(crand.Reader, big.NewInt(1000000))
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
		},
		NotBefore:   time.Now().UTC(),
		NotAfter:    time.Now().UTC().Add(cryptoutilSharedMagic.StrictCertificateMaxAgeDays * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{"test.example.com"},
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, issuer, &privateKey.PublicKey, issuerKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert
}
