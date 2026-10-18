// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	"crypto"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
)

const (
	strictHandlerTestPresent                   = "present"
	strictHandlerTestResultKey                 = "result"
	strictHandlerTestReaderError               = "reader error"
	strictHandlerTestBadRequest                = "bad_request"
	strictHandlerTestInvalidCSR                = "invalid_csr"
	strictHandlerTestEmptyBody                 = "empty body"
	strictHandlerTestServiceUnavailable        = "service_unavailable"
	strictHandlerTestStorageError              = "storage error"
	strictHandlerTestIssuerAbsent              = "issuer absent"
	strictHandlerTestEnrollmentPath            = "/pki/enrollments"
	strictHandlerTestOtherEnrollmentPath       = "/other/enrollments"
	strictHandlerTestOtherPKIPath              = "/pki/other"
	strictHandlerTestListCAs                   = "list CAs"
	strictHandlerTestGetCA                     = "get CA"
	strictHandlerTestSubmitEnrollment          = "submit enrollment"
	strictHandlerTestGetEnrollmentStatus       = "get enrollment status"
	strictHandlerTestListProfiles              = "list profiles"
	strictHandlerTestGetProfile                = "get profile"
	strictHandlerTestMissingBody               = "missing body"
	strictHandlerTestCertificateNotFound       = "certificate not found"
	strictHandlerTestUnknownProfile            = "unknown profile"
	strictHandlerTestIssuerMismatch            = "issuer mismatch"
	strictHandlerTestSerialX                   = "x"
	strictHandlerTestCertificatePEM            = "certificate-pem"
	strictHandlerTestMissingRevocationMetadata = "missing revocation metadata"
	strictHandlerTestSerialNotHex              = "not-hex"
)

func newStrictServerFixture(t *testing.T) (*testIssuerSetup, *Handler, *StrictServer) {
	t.Helper()

	issuerSetup := createTestIssuer(t)
	handler := &Handler{
		issuer:  issuerSetup.Issuer,
		storage: cryptoutilCAStorage.NewMemoryStore(),
		profiles: map[string]*ProfileConfig{
			cryptoutilSharedMagic.PKIProfileIDTLSServer: {
				ID:          cryptoutilSharedMagic.PKIProfileIDTLSServer,
				Name:        testProfileNameTLSServer,
				Description: testProfileDescriptionTLSServer,
				Category:    cryptoutilSharedMagic.PKIProfileCategoryTLS,
			},
		},
	}

	return issuerSetup, handler, NewStrictServer(handler)
}

func strictTestStoredCertificate(t *testing.T, serialNumber string) *cryptoutilCAStorage.StoredCertificate {
	t.Helper()

	id, err := googleUuid.NewV7()
	require.NoError(t, err)
	requestID, err := googleUuid.NewV7()
	require.NoError(t, err)

	now := time.Now().UTC()

	return &cryptoutilCAStorage.StoredCertificate{
		ID:             id,
		RequestID:      requestID,
		SerialNumber:   serialNumber,
		SubjectDN:      "CN=Strict Test Certificate",
		IssuerDN:       "CN=Strict Test Issuer",
		NotBefore:      now.Add(-time.Hour),
		NotAfter:       now.Add(time.Hour),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      cryptoutilSharedMagic.PKIProfileIDTLSServer,
		CertificatePEM: strictHandlerTestCertificatePEM,
	}
}

func newStrictCRLService(t *testing.T, issuerSetup *testIssuerSetup) *cryptoutilCAServiceRevocation.CRLService {
	t.Helper()

	caConfig := issuerSetup.Issuer.GetCAConfig()
	signer, ok := caConfig.PrivateKey.(crypto.Signer)
	require.True(t, ok)

	service, err := cryptoutilCAServiceRevocation.NewCRLService(&cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:     caConfig.Certificate,
		PrivateKey: signer,
		Provider:   issuerSetup.Provider,
		Validity:   time.Hour,
	})
	require.NoError(t, err)

	return service
}

func assertStrictResult[T any](t *testing.T, result T, err error, statusCode int) strictResponse {
	t.Helper()
	require.NoError(t, err)

	response, ok := any(result).(strictResponse)
	require.True(t, ok, "strict response type = %T", result)
	require.Equal(t, statusCode, response.statusCode)

	return response
}

func assertStrictCall(t *testing.T, statusCode int, call func() (any, error)) strictResponse {
	t.Helper()

	result, err := call()

	return assertStrictResult(t, result, err, statusCode)
}

func strictResponseBody[T any](t *testing.T, response strictResponse) T {
	t.Helper()

	body, ok := response.body.(T)
	require.True(t, ok, "strict response body type = %T", response.body)

	return body
}

func strictTestContext() context.Context {
	return context.Background()
}
