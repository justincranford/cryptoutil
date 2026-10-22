// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"
	"time"

	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/sm-kms/server/pki/profile/certificate"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuildESTIssueRequest_ErrorsAndDefaultValidity(t *testing.T) {
	t.Parallel()

	request := &x509.CertificateRequest{}
	defaultValidity := time.Duration(cryptoutilSharedMagic.DaysPerYear) *
		cryptoutilSharedMagic.HoursPerDay * time.Hour

	tests := []struct {
		name         string
		csr          *x509.CertificateRequest
		profile      *ProfileConfig
		wantValidity time.Duration
		wantErr      string
	}{
		{name: "nil CSR", profile: &ProfileConfig{}, wantErr: "CSR is required"},
		{name: "nil profile", csr: request, wantErr: "EST certificate profile is required"},
		{
			name:         "default validity without certificate profile",
			csr:          request,
			profile:      &ProfileConfig{},
			wantValidity: defaultValidity,
		},
		{
			name: "invalid configured duration",
			csr:  request,
			profile: &ProfileConfig{CertificateProfile: &cryptoutilCAProfileCertificate.Profile{
				Validity: cryptoutilCAProfileCertificate.ValidityConfig{Duration: cryptoutilSharedMagic.DevSetupTestInvalidName, MaxDuration: testESTProfileValidity48Hours},
			}},
			wantErr: testESTInvalidProfileValidity,
		},
		{
			name: "configured duration exceeds maximum",
			csr:  request,
			profile: &ProfileConfig{CertificateProfile: &cryptoutilCAProfileCertificate.Profile{
				Validity: cryptoutilCAProfileCertificate.ValidityConfig{Duration: testESTProfileValidity24Hours, MaxDuration: "12h"},
			}},
			wantErr: testESTInvalidProfileValidity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			issueRequest, err := (&Handler{}).buildESTIssueRequest(tc.csr, tc.profile)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, issueRequest)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantValidity, issueRequest.ValidityDuration)
		})
	}
}

func TestAuthorizeESTReenrollmentPreconditions(t *testing.T) {
	t.Parallel()

	setup := createTestIssuer(t)
	clientCert := &x509.Certificate{}
	tests := []struct {
		name    string
		handler *Handler
		cert    *x509.Certificate
		wantErr string
	}{
		{name: testHandlerNilClientCertificate, handler: &Handler{}, wantErr: errESTClientCertificateRequired.Error()},
		{
			name:    "missing dependencies",
			handler: &Handler{},
			cert:    clientCert,
			wantErr: testESTReenrollmentDependenciesMissing,
		},
		{
			name: "missing storage",
			handler: &Handler{
				issuer: setup.Issuer,
			},
			cert:    clientCert,
			wantErr: testESTReenrollmentDependenciesMissing,
		},
		{
			name: "missing issuer configuration",
			handler: &Handler{
				issuer:  &cryptoutilCAServiceIssuer.Issuer{},
				storage: cryptoutilCAStorage.NewMemoryStore(),
			},
			cert:    clientCert,
			wantErr: "EST issuing authority is not configured",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.handler.authorizeESTReenrollment(context.Background(), tc.cert)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestAuthorizeESTReenrollmentCertificatePolicy(t *testing.T) {
	t.Parallel()

	setup := createTestIssuer(t)
	caCertificate := setup.Issuer.GetCAConfig().Certificate
	clientCert := newESTReenrollmentTestCertificate(t, caCertificate, testIssuerSigner(t, setup))

	tests := []struct {
		name         string
		mutateCert   func(*x509.Certificate)
		storageState string
		wantErr      string
	}{
		{
			name: strictHandlerTestIssuerMismatch,
			mutateCert: func(cert *x509.Certificate) {
				cert.RawIssuer = []byte("different issuer")
			},
			wantErr: errESTReenrollmentDenied.Error(),
		},
		{
			name: testESTAuthorityKeyIdentifierMismatch,
			mutateCert: func(cert *x509.Certificate) {
				cert.AuthorityKeyId = []byte("different authority")
			},
			wantErr: errESTReenrollmentDenied.Error(),
		},
		{
			name: "certificate signature mismatch",
			mutateCert: func(cert *x509.Certificate) {
				cert.Signature = []byte("invalid signature")
			},
			wantErr: errESTReenrollmentDenied.Error(),
		},
		{name: testESTCertificateNotStored, wantErr: errESTReenrollmentDenied.Error()},
		{
			name:         strictHandlerTestStorageError,
			storageState: cryptoutilSharedMagic.StringError,
			wantErr:      "failed to look up EST re-enrollment certificate",
		},
		{
			name:         testESTRevokedCertificate,
			storageState: testESTStorageStateRevoked,
			wantErr:      errESTReenrollmentDenied.Error(),
		},
		{
			name:         "stored certificate DER mismatch",
			storageState: testESTStorageStateMismatchedDER,
			wantErr:      errESTReenrollmentDenied.Error(),
		},
		{name: "active certificate matches", storageState: testESTStorageStateActive},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			certCopy := *clientCert
			if tc.mutateCert != nil {
				tc.mutateCert(&certCopy)
			}

			var store cryptoutilCAStorage.Store = cryptoutilCAStorage.NewMemoryStore()

			switch tc.storageState {
			case cryptoutilSharedMagic.StringError:
				store = errorESTReenrollmentStore{
					MemoryStore: cryptoutilCAStorage.NewMemoryStore(),
					err:         errors.New("injected storage error"),
				}
			case testESTStorageStateRevoked, testESTStorageStateMismatchedDER, testESTStorageStateActive:
				der := certCopy.Raw
				if tc.storageState == testESTStorageStateMismatchedDER {
					der = []byte("different DER")
				}

				status := cryptoutilCAStorage.StatusActive
				if tc.storageState == testESTStorageStateRevoked {
					status = cryptoutilCAStorage.StatusRevoked
				}

				storeESTReenrollmentCertificate(t, store, caCertificate, &certCopy, status, der)
			}

			err := (&Handler{issuer: setup.Issuer, storage: store}).authorizeESTReenrollment(context.Background(), &certCopy)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func newESTReenrollmentTestCertificate(
	t *testing.T,
	caCertificate *x509.Certificate,
	caSigner crypto.Signer,
) *x509.Certificate {
	t.Helper()

	certificatePEM := createOCSPTestCertificatePEM(t, caCertificate, caSigner, big.NewInt(42))
	block, _ := pem.Decode([]byte(certificatePEM))
	require.NotNil(t, block)

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	return cert
}

func storeESTReenrollmentCertificate(
	t *testing.T,
	store cryptoutilCAStorage.Store,
	caCertificate *x509.Certificate,
	cert *x509.Certificate,
	status cryptoutilCAStorage.CertificateStatus,
	certificateDER []byte,
) {
	t.Helper()

	id := googleUuid.Must(googleUuid.NewV7())
	requestID := googleUuid.Must(googleUuid.NewV7())
	require.NoError(t, store.Store(context.Background(), &cryptoutilCAStorage.StoredCertificate{
		ID:             id,
		RequestID:      requestID,
		SerialNumber:   cert.SerialNumber.Text(cryptoutilSharedMagic.HexBase),
		Status:         status,
		CertificateDER: certificateDER,
		AuthorityKeyID: fmtIssuerKeyID(caCertificate),
	}))
}

type errorESTReenrollmentStore struct {
	*cryptoutilCAStorage.MemoryStore
	err error
}

func (s errorESTReenrollmentStore) GetBySerialNumberForIssuer(
	context.Context,
	string,
	string,
) (*cryptoutilCAStorage.StoredCertificate, error) {
	return nil, s.err
}
