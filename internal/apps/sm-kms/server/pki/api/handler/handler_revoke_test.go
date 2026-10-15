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
	"math/big"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
)

// TestRevokeCertificate_ErrorPaths tests various error conditions in RevokeCertificate.
func TestRevokeCertificate_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupStorage   func() cryptoutilCAStorage.Store
		serialNumber   string
		requestBody    string
		expectedStatus int
	}{
		{
			name: testCaseNameInvalidJSONBody,
			setupStorage: func() cryptoutilCAStorage.Store {
				return cryptoutilCAStorage.NewMemoryStore()
			},
			serialNumber:   testCertificateSerialNumber,
			requestBody:    `{invalid json}`,
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "certificate_not_found",
			setupStorage: func() cryptoutilCAStorage.Store {
				return cryptoutilCAStorage.NewMemoryStore()
			},
			serialNumber:   "NONEXISTENT",
			requestBody:    testRequestBodyKeyCompromise,
			expectedStatus: fiber.StatusNotFound,
		},
		{
			name: "already_revoked",
			setupStorage: func() cryptoutilCAStorage.Store {
				storage := cryptoutilCAStorage.NewMemoryStore()
				cert := &cryptoutilCAStorage.StoredCertificate{
					ID:             googleUuid.New(),
					RequestID:      googleUuid.Must(googleUuid.NewV7()),
					SerialNumber:   testSerialAlreadyRevoked,
					SubjectDN:      testSubjectDNExample,
					IssuerDN:       testIssuerDNTestCA,
					NotBefore:      time.Now().UTC().Add(-time.Hour),
					NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
					Status:         cryptoutilCAStorage.StatusRevoked,
					ProfileID:      cryptoutilSharedMagic.PKIProfileIDTLSServer,
					CertificatePEM: testCertificatePEMInvalid,
				}
				_ = storage.Store(context.Background(), cert)

				return storage
			},
			serialNumber:   testSerialAlreadyRevoked,
			requestBody:    testRequestBodyKeyCompromise,
			expectedStatus: fiber.StatusConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			storage := tc.setupStorage()
			handler := &Handler{storage: storage}

			app.Post("/certificates/:serialNumber/revoke", func(c *fiber.Ctx) error {
				serial := c.Params("serialNumber")

				return handler.RevokeCertificate(c, serial)
			})

			req := httptest.NewRequest(http.MethodPost, "/certificates/"+tc.serialNumber+cryptoutilSharedMagic.PathRevoke, bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", testContentTypeJSON)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}

// TestRevokeCertificate_AllReasons tests all revocation reasons are mapped correctly.
func TestRevokeCertificate_AllReasons(t *testing.T) {
	t.Parallel()

	testIssuer := createTestIssuer(t).Issuer
	caConfig := testIssuer.GetCAConfig()

	reasons := []struct {
		name           string
		reason         cryptoutilApiCaServer.RevocationReason
		expectedStatus int
	}{
		{name: testRevocationCaseKeyCompromise, reason: cryptoutilApiCaServer.KeyCompromise, expectedStatus: fiber.StatusOK},
		{name: testRevocationCaseCACompromise, reason: cryptoutilApiCaServer.CACompromise, expectedStatus: fiber.StatusOK},
		{name: testRevocationCaseAffiliationChanged, reason: cryptoutilApiCaServer.AffiliationChanged, expectedStatus: fiber.StatusOK},
		{name: cryptoutilSharedMagic.RFC5280RevocationReasonNameSuperseded, reason: cryptoutilApiCaServer.Superseded, expectedStatus: fiber.StatusOK},
		{name: testRevocationCaseCessationOfOperation, reason: cryptoutilApiCaServer.CessationOfOperation, expectedStatus: fiber.StatusOK},
		{name: testRevocationCaseCertificateHold, reason: cryptoutilApiCaServer.CertificateHold, expectedStatus: fiber.StatusOK},
		{name: testRevocationCaseRemoveFromCRL, reason: cryptoutilApiCaServer.RemoveFromCRL, expectedStatus: fiber.StatusUnprocessableEntity},
		{name: testRevocationCasePrivilegeWithdrawn, reason: cryptoutilApiCaServer.PrivilegeWithdrawn, expectedStatus: fiber.StatusOK},
		{name: testRevocationCaseAACompromise, reason: cryptoutilApiCaServer.AaCompromise, expectedStatus: fiber.StatusOK},
	}

	for _, tc := range reasons {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			storage := cryptoutilCAStorage.NewMemoryStore()

			// Create unique certificate for each test case.
			cert := &cryptoutilCAStorage.StoredCertificate{
				ID:                googleUuid.New(),
				RequestID:         googleUuid.Must(googleUuid.NewV7()),
				SerialNumber:      googleUuid.NewString(),
				SubjectDN:         testSubjectDNExample,
				IssuerDN:          caConfig.Certificate.Subject.String(),
				IssuerAuthorityID: caConfig.AuthorityID,
				AuthorityKeyID:    fmt.Sprintf("%x", caConfig.Certificate.SubjectKeyId),
				NotBefore:         time.Now().UTC().Add(-time.Hour),
				NotAfter:          time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
				Status:            cryptoutilCAStorage.StatusActive,
				ProfileID:         cryptoutilSharedMagic.PKIProfileIDTLSServer,
				CertificatePEM:    testCertificatePEMInvalid,
			}
			err := storage.Store(context.Background(), cert)
			require.NoError(t, err)

			handler := &Handler{issuer: testIssuer, storage: storage}

			app.Post("/certificates/:serialNumber/revoke", func(c *fiber.Ctx) error {
				return handler.RevokeCertificate(c, c.Params("serialNumber"))
			})

			requestBody := `{"reason": "` + tc.name + `"}`
			req := httptest.NewRequest(http.MethodPost, "/certificates/"+cert.SerialNumber+cryptoutilSharedMagic.PathRevoke, bytes.NewBufferString(requestBody))
			req.Header.Set("Content-Type", testContentTypeJSON)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			err = resp.Body.Close()
			require.NoError(t, err)

			stored, err := storage.GetBySerialNumber(context.Background(), cert.SerialNumber)
			require.NoError(t, err)

			if tc.expectedStatus == fiber.StatusOK {
				require.Equal(t, cryptoutilCAStorage.StatusRevoked, stored.Status)
			} else {
				require.Equal(t, cryptoutilCAStorage.StatusActive, stored.Status)
			}
		})
	}
}

func TestRevokeCertificate_RejectsIssuerMetadataMismatch(t *testing.T) {
	t.Parallel()

	testIssuer := createTestIssuer(t).Issuer
	caConfig := testIssuer.GetCAConfig()
	tests := []struct {
		name   string
		mutate func(*cryptoutilCAStorage.StoredCertificate)
	}{
		{
			name: "authority ID",
			mutate: func(certificate *cryptoutilCAStorage.StoredCertificate) {
				certificate.IssuerAuthorityID = "different-authority"
			},
		},
		{
			name: "issuer DN",
			mutate: func(certificate *cryptoutilCAStorage.StoredCertificate) {
				certificate.IssuerDN = "CN=Different Issuer"
			},
		},
		{
			name: "authority key ID",
			mutate: func(certificate *cryptoutilCAStorage.StoredCertificate) {
				certificate.AuthorityKeyID = "00"
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			storage := cryptoutilCAStorage.NewMemoryStore()
			certificate := &cryptoutilCAStorage.StoredCertificate{
				ID:                googleUuid.Must(googleUuid.NewV7()),
				RequestID:         googleUuid.Must(googleUuid.NewV7()),
				SerialNumber:      googleUuid.NewString(),
				SubjectDN:         testSubjectDNExample,
				IssuerDN:          caConfig.Certificate.Subject.String(),
				IssuerAuthorityID: caConfig.AuthorityID,
				AuthorityKeyID:    fmt.Sprintf("%x", caConfig.Certificate.SubjectKeyId),
				NotBefore:         time.Now().UTC().Add(-time.Hour),
				NotAfter:          time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
				Status:            cryptoutilCAStorage.StatusActive,
				ProfileID:         cryptoutilSharedMagic.PKIProfileIDTLSServer,
				CertificatePEM:    testCertificatePEMInvalid,
			}
			tc.mutate(certificate)
			require.NoError(t, storage.Store(context.Background(), certificate))

			handler := &Handler{issuer: testIssuer, storage: storage}
			app := fiber.New()
			app.Post("/certificates/:serialNumber/revoke", func(c *fiber.Ctx) error {
				return handler.RevokeCertificate(c, c.Params("serialNumber"))
			})

			req := httptest.NewRequest(
				http.MethodPost,
				"/certificates/"+certificate.SerialNumber+cryptoutilSharedMagic.PathRevoke,
				bytes.NewBufferString(testRequestBodyKeyCompromise),
			)
			req.Header.Set("Content-Type", testContentTypeJSON)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
			require.NoError(t, resp.Body.Close())

			stored, err := storage.GetBySerialNumber(context.Background(), certificate.SerialNumber)
			require.NoError(t, err)
			require.Equal(t, cryptoutilCAStorage.StatusActive, stored.Status)
		})
	}
}

func TestRevokeCertificate_LegacyIssuerAuthorityID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mutate         func(*cryptoutilCAStorage.StoredCertificate)
		expectedStatus int
		expectedState  cryptoutilCAStorage.CertificateStatus
	}{
		{
			name:           "legacy signed certificate can be revoked",
			expectedStatus: fiber.StatusOK,
			expectedState:  cryptoutilCAStorage.StatusRevoked,
		},
		{
			name: "legacy invalid certificate is rejected",
			mutate: func(certificate *cryptoutilCAStorage.StoredCertificate) {
				certificate.CertificatePEM = testCertificatePEMInvalid
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedState:  cryptoutilCAStorage.StatusActive,
		},
		{
			name: "legacy serial mismatch is rejected",
			mutate: func(certificate *cryptoutilCAStorage.StoredCertificate) {
				certificate.SerialNumber = "1"
			},
			expectedStatus: fiber.StatusInternalServerError,
			expectedState:  cryptoutilCAStorage.StatusActive,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			issuer := createTestIssuer(t).Issuer

			certificate := newLegacyStoredCertificate(t, issuer)
			if tc.mutate != nil {
				tc.mutate(certificate)
			}

			storage := cryptoutilCAStorage.NewMemoryStore()
			require.NoError(t, storage.Store(context.Background(), certificate))

			handler := &Handler{issuer: issuer, storage: storage}
			app := fiber.New()
			app.Post("/certificates/:serialNumber/revoke", func(c *fiber.Ctx) error {
				return handler.RevokeCertificate(c, c.Params("serialNumber"))
			})

			req := httptest.NewRequest(
				http.MethodPost,
				"/certificates/"+certificate.SerialNumber+cryptoutilSharedMagic.PathRevoke,
				bytes.NewBufferString(testRequestBodyKeyCompromise),
			)
			req.Header.Set("Content-Type", testContentTypeJSON)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			require.NoError(t, resp.Body.Close())

			stored, err := storage.GetBySerialNumber(context.Background(), certificate.SerialNumber)
			require.NoError(t, err)
			require.Equal(t, tc.expectedState, stored.Status)
		})
	}
}

func newLegacyStoredCertificate(t *testing.T, issuer *cryptoutilCAServiceIssuer.Issuer) *cryptoutilCAStorage.StoredCertificate {
	t.Helper()

	caConfig := issuer.GetCAConfig()
	issuerSigner, ok := caConfig.PrivateKey.(crypto.Signer)
	require.True(t, ok)

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	serialUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	now := time.Now().UTC()
	leafTemplate := &x509.Certificate{
		SerialNumber: new(big.Int).SetBytes(serialUUID[:]),
		Subject:      pkix.Name{CommonName: "legacy-revocation.example.com"},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
	}

	leafDER, err := x509.CreateCertificate(crand.Reader, leafTemplate, caConfig.Certificate, &leafKey.PublicKey, issuerSigner)
	require.NoError(t, err)
	leafCertificate, err := x509.ParseCertificate(leafDER)
	require.NoError(t, err)

	certificateID, err := googleUuid.NewV7()
	require.NoError(t, err)

	return &cryptoutilCAStorage.StoredCertificate{
		ID:             certificateID,
		RequestID:      googleUuid.Must(googleUuid.NewV7()),
		SerialNumber:   leafCertificate.SerialNumber.Text(cryptoutilSharedMagic.HexBase),
		SubjectDN:      leafCertificate.Subject.String(),
		IssuerDN:       leafCertificate.Issuer.String(),
		AuthorityKeyID: fmt.Sprintf("%x", leafCertificate.AuthorityKeyId),
		NotBefore:      leafCertificate.NotBefore,
		NotAfter:       leafCertificate.NotAfter,
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      cryptoutilSharedMagic.PKIProfileIDTLSServer,
		CertificatePEM: string(pem.EncodeToMemory(&pem.Block{
			Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: leafDER,
		})),
	}
}
