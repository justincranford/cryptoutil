// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	"context"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ocsp"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestGetCRLUsesPersistedRevocations(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	caConfig := testSetup.Issuer.GetCAConfig()
	signer, ok := caConfig.PrivateKey.(crypto.Signer)
	require.True(t, ok)

	crlService, err := cryptoutilCAServiceRevocation.NewCRLService(&cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:     caConfig.Certificate,
		PrivateKey: signer,
		Provider:   testSetup.Provider,
		Validity:   time.Hour,
	})
	require.NoError(t, err)

	storage := cryptoutilCAStorage.NewMemoryStore()
	serialUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	serialNumber := new(big.Int).SetBytes(serialUUID[:])
	reason := cryptoutilCAStorage.ReasonKeyCompromise
	revokedAt := time.Now().UTC().Add(-time.Minute)
	certificateID, err := googleUuid.NewV7()
	require.NoError(t, err)
	require.NoError(t, storage.Store(context.Background(), &cryptoutilCAStorage.StoredCertificate{
		ID:               certificateID,
		RequestID:        googleUuid.Must(googleUuid.NewV7()),
		SerialNumber:     serialNumber.Text(cryptoutilSharedMagic.HexBase),
		IssuerDN:         caConfig.Certificate.Subject.String(),
		AuthorityKeyID:   fmt.Sprintf("%x", caConfig.Certificate.SubjectKeyId),
		Status:           cryptoutilCAStorage.StatusRevoked,
		RevocationTime:   &revokedAt,
		RevocationReason: &reason,
	}))

	staleSerialUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	staleSerialNumber := new(big.Int).SetBytes(staleSerialUUID[:])
	require.NoError(t, crlService.Revoke(staleSerialNumber, cryptoutilCAServiceRevocation.ReasonSuperseded))

	handler := &Handler{
		issuer:     testSetup.Issuer,
		storage:    storage,
		crlService: crlService,
	}
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/ca/:caId/crl", func(c *fiber.Ctx) error {
		return handler.GetCRL(c, c.Params("caId"), cryptoutilApiCaServer.GetCRLParams{})
	})

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/ca/"+caConfig.Name+"/crl", nil), -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, response.StatusCode)

	crlDER, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())

	crl, err := x509.ParseRevocationList(crlDER)
	require.NoError(t, err)
	require.Len(t, crl.RevokedCertificateEntries, 1)
	require.Equal(t, serialNumber, crl.RevokedCertificateEntries[0].SerialNumber)
	require.Equal(t, int(cryptoutilCAServiceRevocation.ReasonKeyCompromise), crl.RevokedCertificateEntries[0].ReasonCode)
}

func TestHandleOCSPUsesPersistedCertificateStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     cryptoutilCAStorage.CertificateStatus
		wantStatus int
		wantReason int
	}{
		{
			name:       "active certificate",
			status:     cryptoutilCAStorage.StatusActive,
			wantStatus: ocsp.Good,
		},
		{
			name:       "revoked certificate",
			status:     cryptoutilCAStorage.StatusRevoked,
			wantStatus: ocsp.Revoked,
			wantReason: int(cryptoutilCAServiceRevocation.ReasonKeyCompromise),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler, leafCert, responderIssuerCert, requestBody := newPersistedOCSPTestHandler(t, tc.status)
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Post("/ocsp", handler.HandleOCSP)

			request := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader(requestBody))
			request.Header.Set("Content-Type", "application/ocsp-request")
			response, err := app.Test(request, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, response.StatusCode)
			require.Equal(t, cryptoutilSharedMagic.PKIContentTypeOCSPResponse, response.Header.Get("Content-Type"))

			responseBody, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())

			parsed, err := ocsp.ParseResponse(responseBody, responderIssuerCert)
			require.NoError(t, err)
			require.Equal(t, leafCert.SerialNumber, parsed.SerialNumber)
			require.Equal(t, tc.wantStatus, parsed.Status)

			if tc.wantStatus == ocsp.Revoked {
				require.Equal(t, tc.wantReason, parsed.RevocationReason)
			}
		})
	}
}

func newPersistedOCSPTestHandler(
	t *testing.T,
	status cryptoutilCAStorage.CertificateStatus,
) (*Handler, *x509.Certificate, *x509.Certificate, []byte) {
	t.Helper()

	testSetup := createTestIssuer(t)
	caConfig := testSetup.Issuer.GetCAConfig()
	signer, ok := caConfig.PrivateKey.(crypto.Signer)
	require.True(t, ok)

	crlService, err := cryptoutilCAServiceRevocation.NewCRLService(&cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:     caConfig.Certificate,
		PrivateKey: signer,
		Provider:   testSetup.Provider,
		Validity:   time.Hour,
	})
	require.NoError(t, err)
	ocspService, err := cryptoutilCAServiceRevocation.NewOCSPService(&cryptoutilCAServiceRevocation.OCSPConfig{
		Issuer:       caConfig.Certificate,
		Responder:    caConfig.Certificate,
		ResponderKey: signer,
		Provider:     testSetup.Provider,
		Validity:     time.Hour,
	}, crlService)
	require.NoError(t, err)

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)
	serialUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	serialNumber := new(big.Int).SetBytes(serialUUID[:])
	now := time.Now().UTC()
	leafTemplate := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{CommonName: "ocsp-persisted.example.com"},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
	}

	leafDER, err := x509.CreateCertificate(crand.Reader, leafTemplate, caConfig.Certificate, &leafKey.PublicKey, signer)
	require.NoError(t, err)
	leafCert, err := x509.ParseCertificate(leafDER)
	require.NoError(t, err)

	certificateID, err := googleUuid.NewV7()
	require.NoError(t, err)

	storedCert := &cryptoutilCAStorage.StoredCertificate{
		ID:             certificateID,
		RequestID:      googleUuid.Must(googleUuid.NewV7()),
		SerialNumber:   leafCert.SerialNumber.Text(cryptoutilSharedMagic.HexBase),
		SubjectDN:      leafCert.Subject.String(),
		IssuerDN:       leafCert.Issuer.String(),
		AuthorityKeyID: fmt.Sprintf("%x", leafCert.AuthorityKeyId),
		NotBefore:      leafCert.NotBefore,
		NotAfter:       leafCert.NotAfter,
		Status:         status,
		CertificatePEM: string(pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: leafDER})),
	}
	if status == cryptoutilCAStorage.StatusRevoked {
		revokedAt := now.Add(-time.Minute)
		reason := cryptoutilCAStorage.ReasonKeyCompromise
		storedCert.RevocationTime = &revokedAt
		storedCert.RevocationReason = &reason
	}

	storage := cryptoutilCAStorage.NewMemoryStore()
	require.NoError(t, storage.Store(context.Background(), storedCert))

	requestBody, err := ocsp.CreateRequest(leafCert, caConfig.Certificate, nil)
	require.NoError(t, err)

	return &Handler{
		issuer:      testSetup.Issuer,
		storage:     storage,
		ocspService: ocspService,
	}, leafCert, testSetup.RootIssuer.GetCAConfig().Certificate, requestBody
}
