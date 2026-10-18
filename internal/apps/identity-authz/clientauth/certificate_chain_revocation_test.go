// Copyright (c) 2025-2026 Justin Cranford.
package clientauth

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ocsp"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type revocationCheckerStub struct {
	err error
}

func (r revocationCheckerStub) CheckRevocation(context.Context, *x509.Certificate, *x509.Certificate) error {
	return r.err
}

func TestCACertificateValidator_TrustedChainsAndRevocation(t *testing.T) {
	t.Parallel()

	caCert, clientCert, _ := createTestCertificatePair(t, false, false)
	validClientCert := *clientCert
	validClientCert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	trustedRoots := x509.NewCertPool()
	trustedRoots.AddCert(caCert)

	revocationErr := errors.New("revocation service unavailable")

	tests := []struct {
		name          string
		rawCerts      [][]byte
		checker       RevocationChecker
		wantErrSubstr string
	}{
		{
			name: "trusted certificate",
			rawCerts: [][]byte{
				validClientCert.Raw,
			},
		},
		{
			name: "trusted certificate with parsed issuer chain",
			rawCerts: [][]byte{
				validClientCert.Raw,
				caCert.Raw,
			},
			checker: revocationCheckerStub{},
		},
		{
			name: "revocation failure",
			rawCerts: [][]byte{
				validClientCert.Raw,
				caCert.Raw,
			},
			checker:       revocationCheckerStub{err: revocationErr},
			wantErrSubstr: "certificate revocation check failed",
		},
		{
			name: "malformed intermediate is ignored",
			rawCerts: [][]byte{
				validClientCert.Raw,
				[]byte("not a certificate"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator := NewCACertificateValidator(trustedRoots, tc.checker)

			err := validator.ValidateCertificate(&validClientCert, tc.rawCerts)
			if tc.wantErrSubstr != "" {
				require.ErrorContains(t, err, tc.wantErrSubstr)

				return
			}

			require.NoError(t, err)
		})
	}

	notYetValid := validClientCert
	notYetValid.NotBefore = time.Now().UTC().Add(time.Hour)
	validator := NewCACertificateValidator(trustedRoots, nil)
	require.ErrorContains(t, validator.ValidateCertificate(&notYetValid, nil), "not yet valid")
}

func TestCertificateParser_ParsePEMCertificateChain_ParseFailure(t *testing.T) {
	t.Parallel()

	parser := &CertificateParser{}

	_, err := parser.ParsePEMCertificateChain(pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: []byte("malformed DER"),
	}))
	require.ErrorContains(t, err, "failed to parse certificate in chain")
}

func TestCertificateParser_ParsePEMCertificateChain_ValidAndEmpty(t *testing.T) {
	t.Parallel()

	caCert, _, _ := createTestCertificatePair(t, false, false)
	validPEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: caCert.Raw})
	validPEM = append(validPEM, []byte("trailing non-PEM data")...)

	tests := []struct {
		name        string
		pemData     []byte
		wantCount   int
		wantErrText string
	}{
		{
			name:      "valid certificate followed by non-PEM data",
			pemData:   validPEM,
			wantCount: 1,
		},
		{
			name:        "no certificates",
			pemData:     []byte("no PEM blocks"),
			wantErrText: "no certificates found in PEM data",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			certs, err := (&CertificateParser{}).ParsePEMCertificateChain(tc.pemData)
			if tc.wantErrText != "" {
				require.ErrorContains(t, err, tc.wantErrText)

				return
			}

			require.NoError(t, err)
			require.Len(t, certs, tc.wantCount)
			require.Equal(t, caCert.Raw, certs[0].Raw)
		})
	}
}

func TestCertificatesEqual_NilAndDifferentCertificates(t *testing.T) {
	t.Parallel()

	require.False(t, certificatesEqual(nil, &x509.Certificate{}))
	require.False(t, certificatesEqual(&x509.Certificate{}, nil))
	require.False(t, certificatesEqual(&x509.Certificate{Raw: []byte("first")}, &x509.Certificate{Raw: []byte("second")}))
}

func TestOCSPRevocationChecker_SignedResponses(t *testing.T) {
	t.Parallel()

	caCert, clientCert, caPrivateKey := createTestCertificatePair(t, false, false)
	tests := []struct {
		name          string
		status        int
		wantErrSubstr string
	}{
		{name: "good response", status: ocsp.Good},
		{name: "revoked response", status: ocsp.Revoked, wantErrSubstr: "certificate has been revoked (OCSP)"},
		{name: "unknown response", status: ocsp.Unknown, wantErrSubstr: "OCSP check failed for all servers"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			now := time.Now().UTC()
			responseBytes, err := ocsp.CreateResponse(caCert, caCert, ocsp.Response{
				Status:       tc.status,
				SerialNumber: clientCert.SerialNumber,
				ThisUpdate:   now,
				NextUpdate:   now.Add(time.Hour),
				RevokedAt:    now,
			}, caPrivateKey)
			require.NoError(t, err)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseBytes)
			}))
			t.Cleanup(server.Close)

			clientCertificate := *clientCert
			clientCertificate.OCSPServer = []string{server.URL}
			checker := NewOCSPRevocationChecker(cryptoutilSharedMagic.DefaultRevocationTimeout)

			err = checker.CheckRevocation(context.Background(), &clientCertificate, caCert)
			if tc.wantErrSubstr != "" {
				require.ErrorContains(t, err, tc.wantErrSubstr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCRLCache_FetchRequestAndTransportErrors(t *testing.T) {
	t.Parallel()

	cache := NewCRLCache(time.Hour)

	_, err := cache.GetCRL(context.Background(), "://invalid-url")
	require.ErrorContains(t, err, "failed to create CRL request")

	server := httptest.NewServer(http.NotFoundHandler())
	serverURL := server.URL
	server.Close()

	_, err = cache.GetCRL(context.Background(), serverURL)
	require.ErrorContains(t, err, "failed to download CRL")
}

func TestCombinedRevocationChecker_OCSPGoodSkipsCRL(t *testing.T) {
	t.Parallel()

	caCert, clientCert, caPrivateKey := createTestCertificatePair(t, true, false)
	now := time.Now().UTC()
	responseBytes, err := ocsp.CreateResponse(caCert, caCert, ocsp.Response{
		Status:       ocsp.Good,
		SerialNumber: clientCert.SerialNumber,
		ThisUpdate:   now,
		NextUpdate:   now.Add(time.Hour),
	}, caPrivateKey)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseBytes)
	}))
	t.Cleanup(server.Close)

	clientCertificate := *clientCert
	clientCertificate.OCSPServer = []string{server.URL}
	checker := NewCombinedRevocationChecker(time.Second, time.Second, time.Hour)
	require.NoError(t, checker.CheckRevocation(context.Background(), &clientCertificate, caCert))

	require.Equal(t, big.NewInt(42), clientCertificate.SerialNumber)
}
