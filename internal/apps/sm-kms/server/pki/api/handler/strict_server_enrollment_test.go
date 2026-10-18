// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"encoding/hex"
	"errors"
	http "net/http"
	"testing"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestStrictServer_SubmitEnrollmentValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       *cryptoutilApiCaServer.EnrollmentRequest
		wantStatus int
	}{
		{name: strictHandlerTestMissingBody, body: nil, wantStatus: http.StatusBadRequest},
		{name: "missing CSR", body: &cryptoutilApiCaServer.EnrollmentRequest{Profile: cryptoutilSharedMagic.PKIProfileIDTLSServer}, wantStatus: http.StatusBadRequest},
		{name: "missing profile", body: &cryptoutilApiCaServer.EnrollmentRequest{CSR: cryptoutilSharedMagic.DevSetupTestInvalidName}, wantStatus: http.StatusBadRequest},
		{name: testCaseNameInvalidCSR, body: &cryptoutilApiCaServer.EnrollmentRequest{CSR: cryptoutilSharedMagic.DevSetupTestInvalidName, Profile: cryptoutilSharedMagic.PKIProfileIDTLSServer}, wantStatus: http.StatusUnprocessableEntity},
		{name: strictHandlerTestUnknownProfile, body: &cryptoutilApiCaServer.EnrollmentRequest{Profile: cryptoutilSharedMagic.STRATEGY_UNKNOWN, CSR: createValidCSRPEM(t)}, wantStatus: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				storage:  cryptoutilCAStorage.NewMemoryStore(),
				profiles: map[string]*ProfileConfig{},
			}
			if tc.name != strictHandlerTestUnknownProfile {
				handler.profiles[cryptoutilSharedMagic.PKIProfileIDTLSServer] = &ProfileConfig{
					ID: cryptoutilSharedMagic.PKIProfileIDTLSServer,
				}
			}

			result := callStrictSubmit(t, NewStrictServer(handler), tc.body, tc.wantStatus)
			require.NotEmpty(t, strictResponseBody[cryptoutilApiCaServer.ErrorResponse](t, result).Error)
		})
	}
}

func TestStrictServer_SubmitEnrollmentSuccessAndStorageError(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		_, _, strict := newStrictServerFixture(t)
		body := validStrictEnrollmentRequest(t)
		result := callStrictSubmit(t, strict, body, http.StatusCreated)
		response := strictResponseBody[*cryptoutilApiCaServer.EnrollmentResponse](t, result)
		require.NotNil(t, response)
		require.Equal(t, cryptoutilApiCaServer.Issued, response.Status)
		require.NotEmpty(t, response.Certificate.CertificatePEM)
		require.NotEmpty(t, response.Certificate.SerialNumber)

		storedResult := assertStrictCall(t, http.StatusOK, func() (any, error) {
			return strict.GetEnrollmentStatus(strictTestContext(), cryptoutilApiCaServer.GetEnrollmentStatusRequestObject{
				RequestID: response.RequestID,
			})
		})
		stored := strictResponseBody[cryptoutilApiCaServer.EnrollmentStatusResponse](t, storedResult)
		require.Equal(t, response.RequestID, stored.RequestID)
		require.Equal(t, response.Certificate.SerialNumber, stored.Certificate.SerialNumber)
	})

	t.Run("storage error", func(t *testing.T) {
		t.Parallel()

		setup := createTestIssuer(t)
		store := &errorStore{
			inner:    cryptoutilCAStorage.NewMemoryStore(),
			storeErr: errors.New("injected enrollment storage error"),
		}
		handler := &Handler{
			issuer:  setup.Issuer,
			storage: store,
			profiles: map[string]*ProfileConfig{
				cryptoutilSharedMagic.PKIProfileIDTLSServer: {
					ID: cryptoutilSharedMagic.PKIProfileIDTLSServer,
				},
			},
		}

		result := callStrictSubmit(t, NewStrictServer(handler), validStrictEnrollmentRequest(t), http.StatusInternalServerError)
		body := strictResponseBody[cryptoutilApiCaServer.ErrorResponse](t, result)
		require.Equal(t, "storage_error", body.Error)
		require.Contains(t, *body.Message, "injected enrollment storage error")
	})

	t.Run("issuance error", func(t *testing.T) {
		t.Parallel()

		_, _, strict := newStrictServerFixture(t)
		body := validStrictEnrollmentRequest(t)
		invalidValidity := -1
		body.ValidityDays = &invalidValidity

		result := callStrictSubmit(t, strict, body, http.StatusInternalServerError)
		require.Equal(t, "issuance_error", strictResponseBody[cryptoutilApiCaServer.ErrorResponse](t, result).Error)
	})
}

func TestStrictServer_RevokeCertificateResults(t *testing.T) {
	t.Parallel()

	setup := createTestIssuer(t)
	ctx := strictTestContext()
	keyCompromise := cryptoutilApiCaServer.RevocationRequest{Reason: cryptoutilApiCaServer.KeyCompromise}
	removeFromCRL := cryptoutilApiCaServer.RevocationRequest{Reason: cryptoutilApiCaServer.RemoveFromCRL}

	tests := []struct {
		name       string
		serial     string
		body       *cryptoutilApiCaServer.RevocationRequest
		status     cryptoutilCAStorage.CertificateStatus
		issuerID   string
		getErr     error
		revokeErr  error
		wantStatus int
	}{
		{name: "missing serial", body: &keyCompromise, wantStatus: http.StatusBadRequest},
		{name: strictHandlerTestMissingBody, serial: "501", wantStatus: http.StatusBadRequest},
		{name: "unsupported reason", serial: "501", body: &removeFromCRL, wantStatus: http.StatusUnprocessableEntity},
		{name: strictHandlerTestCertificateNotFound, serial: strictServerTestMissing, body: &keyCompromise, wantStatus: http.StatusNotFound},
		{name: "storage lookup error", serial: "501", body: &keyCompromise, getErr: errors.New("injected lookup failure"), wantStatus: http.StatusInternalServerError},
		{name: "already revoked", serial: "501", body: &keyCompromise, status: cryptoutilCAStorage.StatusRevoked, wantStatus: http.StatusConflict},
		{name: strictHandlerTestIssuerMismatch, serial: "501", body: &keyCompromise, issuerID: testDifferentAuthorityID, wantStatus: http.StatusInternalServerError},
		{name: "revoke error", serial: "501", body: &keyCompromise, revokeErr: errors.New("injected revoke failure"), wantStatus: http.StatusInternalServerError},
		{name: "success", serial: "501", body: &keyCompromise, wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := cryptoutilCAStorage.NewMemoryStore()
			wrappedStore := &errorStore{inner: store, getErr: tc.getErr, revokeErr: tc.revokeErr}
			handler := &Handler{issuer: setup.Issuer, storage: wrappedStore}

			if tc.serial == "501" {
				stored := strictStoredCertificateForIssuer(t, setup, tc.serial)
				stored.Status = tc.status

				if tc.issuerID != "" {
					stored.IssuerAuthorityID = tc.issuerID
				}

				require.NoError(t, store.Store(ctx, stored))
			}

			result := callStrictRevoke(t, NewStrictServer(handler), tc.serial, tc.body, tc.wantStatus)
			if tc.wantStatus == http.StatusOK {
				body := strictResponseBody[cryptoutilApiCaServer.RevocationResponse](t, result)
				require.Equal(t, cryptoutilApiCaServer.RevocationResponseStatusRevoked, body.Status)
				require.Equal(t, cryptoutilCAStorage.StatusRevoked, func() cryptoutilCAStorage.CertificateStatus {
					cert, err := store.GetBySerialNumber(ctx, tc.serial)
					require.NoError(t, err)

					return cert.Status
				}())
			}
		})
	}
}

func callStrictSubmit(t *testing.T, strict *StrictServer, body *cryptoutilApiCaServer.EnrollmentRequest, statusCode int) strictResponse {
	t.Helper()

	return assertStrictCall(t, statusCode, func() (any, error) {
		return strict.SubmitEnrollment(strictTestContext(), cryptoutilApiCaServer.SubmitEnrollmentRequestObject{Body: body})
	})
}

func callStrictRevoke(t *testing.T, strict *StrictServer, serial string, body *cryptoutilApiCaServer.RevocationRequest, statusCode int) strictResponse {
	t.Helper()

	return assertStrictCall(t, statusCode, func() (any, error) {
		return strict.RevokeCertificate(strictTestContext(), cryptoutilApiCaServer.RevokeCertificateRequestObject{
			SerialNumber: serial,
			Body:         body,
		})
	})
}

func validStrictEnrollmentRequest(t *testing.T) *cryptoutilApiCaServer.EnrollmentRequest {
	t.Helper()

	return &cryptoutilApiCaServer.EnrollmentRequest{
		CSR:     createValidCSRPEM(t),
		Profile: cryptoutilSharedMagic.PKIProfileIDTLSServer,
	}
}

func strictStoredCertificateForIssuer(t *testing.T, setup *testIssuerSetup, serial string) *cryptoutilCAStorage.StoredCertificate {
	t.Helper()

	cert := strictTestStoredCertificate(t, serial)
	caConfig := setup.Issuer.GetCAConfig()
	cert.IssuerDN = caConfig.Certificate.Subject.String()
	cert.IssuerAuthorityID = caConfig.AuthorityID
	cert.AuthorityKeyID = hex.EncodeToString(caConfig.Certificate.SubjectKeyId)

	return cert
}
