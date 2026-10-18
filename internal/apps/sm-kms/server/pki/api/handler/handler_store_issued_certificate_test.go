// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	"errors"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestStoreIssuedCertificate_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		nilIssued     bool
		mutate        func(*Handler, *cryptoutilCAServiceIssuer.IssuedCertificate)
		wantErrorText string
	}{
		{
			name:          "nil issued certificate",
			nilIssued:     true,
			wantErrorText: testInvalidCertificate,
		},
		{
			name: "nil x509 certificate",
			mutate: func(_ *Handler, issued *cryptoutilCAServiceIssuer.IssuedCertificate) {
				issued.Certificate = nil
			},
			wantErrorText: testInvalidCertificate,
		},
		{
			name: testIssuerNotConfigured,
			mutate: func(handler *Handler, _ *cryptoutilCAServiceIssuer.IssuedCertificate) {
				handler.issuer = nil
			},
			wantErrorText: testIssuingAuthorityIdentityNotConfigured,
		},
		{
			name: "authority ID has surrounding whitespace",
			mutate: func(handler *Handler, _ *cryptoutilCAServiceIssuer.IssuedCertificate) {
				handler.issuer.GetCAConfig().AuthorityID += testWhitespace
			},
			wantErrorText: testIssuingAuthorityIdentityNotConfigured,
		},
		{
			name: "authority certificate is not configured",
			mutate: func(handler *Handler, _ *cryptoutilCAServiceIssuer.IssuedCertificate) {
				handler.issuer.GetCAConfig().Certificate = nil
			},
			wantErrorText: "issuing authority certificate is not configured",
		},
		{
			name: "issued certificate has a different issuer",
			mutate: func(_ *Handler, issued *cryptoutilCAServiceIssuer.IssuedCertificate) {
				issued.Certificate.RawIssuer = []byte("untrusted issuer")
			},
			wantErrorText: testIssuerDoesNotMatch,
		},
		{
			name: "issued certificate has a different authority key ID",
			mutate: func(_ *Handler, issued *cryptoutilCAServiceIssuer.IssuedCertificate) {
				issued.Certificate.AuthorityKeyId = []byte("untrusted key")
			},
			wantErrorText: testAuthorityKeyIdentifierMismatch,
		},
		{
			name: "issued certificate signature is invalid",
			mutate: func(_ *Handler, issued *cryptoutilCAServiceIssuer.IssuedCertificate) {
				issued.Certificate.Signature[0] ^= 0xff
			},
			wantErrorText: testSignatureMismatch,
		},
		{
			name: "issued certificate serial number does not match",
			mutate: func(_ *Handler, issued *cryptoutilCAServiceIssuer.IssuedCertificate) {
				issued.SerialNumber = cryptoutilSharedMagic.DevSetupTestInvalidName
			},
			wantErrorText: testSerialNumberMismatch,
		},
		{
			name: "storage rejects certificate",
			mutate: func(handler *Handler, _ *cryptoutilCAServiceIssuer.IssuedCertificate) {
				handler.storage = &errorStore{
					inner:    cryptoutilCAStorage.NewMemoryStore(),
					storeErr: errors.New("injected store failure"),
				}
			},
			wantErrorText: "injected store failure",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			setup := createTestIssuer(t)
			store := cryptoutilCAStorage.NewMemoryStore()
			handler := &Handler{issuer: setup.Issuer, storage: store}

			issued := createIssuedCertificateForHandler(t, handler)
			if tc.mutate != nil {
				tc.mutate(handler, issued)
			}

			if tc.nilIssued {
				issued = nil
			}

			_, err := handler.storeIssuedCertificate(
				context.Background(),
				issued,
				cryptoutilSharedMagic.PKIProfileIDTLSServer,
				googleUuid.Must(googleUuid.NewV7()),
			)
			require.ErrorContains(t, err, tc.wantErrorText)
		})
	}
}

func createIssuedCertificateForHandler(
	t *testing.T,
	handler *Handler,
) *cryptoutilCAServiceIssuer.IssuedCertificate {
	t.Helper()

	csr, err := handler.parseCSR(createValidCSRPEM(t))
	require.NoError(t, err)
	issueRequest, err := handler.buildIssueRequest(csr, nil, &cryptoutilApiCaServer.EnrollmentRequest{
		ValidityDays: ptrInt(cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days),
	})
	require.NoError(t, err)
	issued, _, err := handler.issuer.Issue(issueRequest)
	require.NoError(t, err)

	return issued
}

func ptrInt(value int) *int {
	return &value
}
