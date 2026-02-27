// Copyright (c) 2025 Justin Cranford

package revocation

import (
	"crypto/x509"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ocsp"

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
)

func TestGenerateCRLPEM_Success(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	config := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	svc, err := NewCRLService(config)
	require.NoError(t, err)

	pemData, err := svc.GenerateCRLPEM()
	require.NoError(t, err)
	require.Contains(t, string(pemData), "BEGIN X509 CRL")
}

func TestOCSPService_RespondToRequest_CertFound(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	responderCert, responderKey := caCert, caKey
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	crlConfig := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	crlService, err := NewCRLService(crlConfig)
	require.NoError(t, err)

	ocspConfig := &OCSPConfig{
		Issuer:       caCert,
		Responder:    responderCert,
		ResponderKey: responderKey,
		Provider:     provider,
		Validity:     cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	ocspService, err := NewOCSPService(ocspConfig, crlService)
	require.NoError(t, err)

	// Create a test certificate.
	testCert := createTestCertificate(t, caCert, caKey)

	// Create an OCSP request for the test certificate.
	requestBytes, err := ocsp.CreateRequest(testCert, caCert, nil)
	require.NoError(t, err)

	// Respond with cert found.
	lookup := func(_ *big.Int) *x509.Certificate {
		return testCert
	}

	respBytes, err := ocspService.RespondToRequest(requestBytes, lookup)
	require.NoError(t, err)
	require.NotEmpty(t, respBytes)

	// Parse response and verify status.
	resp, err := ocsp.ParseResponse(respBytes, caCert)
	require.NoError(t, err)
	require.Equal(t, ocsp.Good, resp.Status)
}

func TestOCSPService_RespondToRequest_CertNotFound(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	responderCert, responderKey := caCert, caKey
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	crlConfig := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	crlService, err := NewCRLService(crlConfig)
	require.NoError(t, err)

	ocspConfig := &OCSPConfig{
		Issuer:       caCert,
		Responder:    responderCert,
		ResponderKey: responderKey,
		Provider:     provider,
		Validity:     cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	ocspService, err := NewOCSPService(ocspConfig, crlService)
	require.NoError(t, err)

	// Create a test certificate.
	testCert := createTestCertificate(t, caCert, caKey)

	// Create an OCSP request.
	requestBytes, err := ocsp.CreateRequest(testCert, caCert, nil)
	require.NoError(t, err)

	// Respond with cert not found (unknown).
	lookup := func(_ *big.Int) *x509.Certificate {
		return nil
	}

	respBytes, err := ocspService.RespondToRequest(requestBytes, lookup)
	require.NoError(t, err)
	require.NotEmpty(t, respBytes)

	// Parse and verify unknown status.
	resp, err := ocsp.ParseResponse(respBytes, caCert)
	require.NoError(t, err)
	require.Equal(t, ocsp.Unknown, resp.Status)
}

func TestOCSPService_ParseRequest_Success(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	responderCert, responderKey := caCert, caKey
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	crlConfig := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	crlService, err := NewCRLService(crlConfig)
	require.NoError(t, err)

	ocspConfig := &OCSPConfig{
		Issuer:       caCert,
		Responder:    responderCert,
		ResponderKey: responderKey,
		Provider:     provider,
		Validity:     cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	ocspService, err := NewOCSPService(ocspConfig, crlService)
	require.NoError(t, err)

	testCert := createTestCertificate(t, caCert, caKey)

	requestBytes, err := ocsp.CreateRequest(testCert, caCert, nil)
	require.NoError(t, err)

	req, err := ocspService.ParseRequest(requestBytes)
	require.NoError(t, err)
	require.NotNil(t, req)
	require.NotNil(t, req.SerialNumber)
}

func TestOCSPService_RespondToRequest_InvalidRequest(t *testing.T) {
	t.Parallel()

	caCert, caKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	crlConfig := &CRLConfig{
		Issuer:     caCert,
		PrivateKey: caKey,
		Provider:   provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	crlService, err := NewCRLService(crlConfig)
	require.NoError(t, err)

	ocspConfig := &OCSPConfig{
		Issuer:       caCert,
		Responder:    caCert,
		ResponderKey: caKey,
		Provider:     provider,
		Validity:     cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	ocspService, err := NewOCSPService(ocspConfig, crlService)
	require.NoError(t, err)

	// Invalid request bytes.
	lookup := func(_ *big.Int) *x509.Certificate {
		return nil
	}

	_, err = ocspService.RespondToRequest([]byte{0xFF, 0xFF}, lookup)
	require.Error(t, err)
}
