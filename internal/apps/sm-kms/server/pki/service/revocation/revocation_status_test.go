// Copyright (c) 2025-2026 Justin Cranford.
package revocation

import (
	"crypto/x509"
	"errors"
	"math/big"
	"testing"
	"time"

	"golang.org/x/crypto/ocsp"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestOCSPService_RespondToRequestWithStatus(t *testing.T) {
	t.Parallel()

	issuer, issuerKey := createTestCA(t)
	otherIssuer, otherIssuerKey := createTestCA(t)
	certificate := createTestCertificate(t, issuer, issuerKey)
	otherCertificate := createTestCertificate(t, otherIssuer, otherIssuerKey)
	provider := cryptoutilCACrypto.NewSoftwareProvider()
	crlService, err := NewCRLService(&CRLConfig{
		Issuer:     issuer,
		PrivateKey: issuerKey,
		Provider:   provider,
		Validity:   time.Hour,
	})
	require.NoError(t, err)
	service, err := NewOCSPService(&OCSPConfig{
		Issuer:       issuer,
		Responder:    issuer,
		ResponderKey: issuerKey,
		Provider:     provider,
		Validity:     time.Hour,
	}, crlService)
	require.NoError(t, err)

	request, err := ocsp.CreateRequest(certificate, issuer, nil)
	require.NoError(t, err)
	otherIssuerRequest, err := ocsp.CreateRequest(otherCertificate, otherIssuer, nil)
	require.NoError(t, err)

	revocationTime := time.Now().UTC()

	tests := []struct {
		name           string
		request        []byte
		lookup         func(*big.Int) (*x509.Certificate, *RevokedCertificate, error)
		wantStatus     int
		wantErr        string
		wantRevocation bool
	}{
		{
			name:    "nil status lookup",
			request: request,
			wantErr: "certificate status lookup is required",
		},
		{
			name:    "invalid request",
			request: []byte{0xff},
			lookup:  func(*big.Int) (*x509.Certificate, *RevokedCertificate, error) { return nil, nil, nil },
			wantErr: "failed to parse OCSP request",
		},
		{
			name:    "request for another issuer",
			request: otherIssuerRequest,
			lookup:  func(*big.Int) (*x509.Certificate, *RevokedCertificate, error) { return certificate, nil, nil },
			wantErr: "does not match the selected authority",
		},
		{
			name:    "status lookup failure",
			request: request,
			lookup: func(*big.Int) (*x509.Certificate, *RevokedCertificate, error) {
				return nil, nil, errors.New("injected lookup failure")
			},
			wantErr: "injected lookup failure",
		},
		{
			name:       "unknown certificate",
			request:    request,
			lookup:     func(*big.Int) (*x509.Certificate, *RevokedCertificate, error) { return nil, nil, nil },
			wantStatus: ocsp.Unknown,
		},
		{
			name:       "good certificate",
			request:    request,
			lookup:     func(*big.Int) (*x509.Certificate, *RevokedCertificate, error) { return certificate, nil, nil },
			wantStatus: ocsp.Good,
		},
		{
			name:    "revoked certificate",
			request: request,
			lookup: func(*big.Int) (*x509.Certificate, *RevokedCertificate, error) {
				return certificate, &RevokedCertificate{
					SerialNumber:   certificate.SerialNumber,
					RevocationTime: revocationTime,
					Reason:         ReasonKeyCompromise,
				}, nil
			},
			wantStatus:     ocsp.Revoked,
			wantRevocation: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			response, responseErr := service.RespondToRequestWithStatus(tc.request, tc.lookup)
			if tc.wantErr != "" {
				require.ErrorContains(t, responseErr, tc.wantErr)
				require.Empty(t, response)

				return
			}

			require.NoError(t, responseErr)

			parsed, parseErr := ocsp.ParseResponse(response, issuer)
			require.NoError(t, parseErr)
			require.Equal(t, tc.wantStatus, parsed.Status)

			if tc.wantRevocation {
				require.Equal(t, int(ReasonKeyCompromise), parsed.RevocationReason)
				require.Equal(t, revocationTime.Unix(), parsed.RevokedAt.Unix())
			}
		})
	}
}

func TestOCSPService_CreateResponseWithRevocation_NilCertificate(t *testing.T) {
	t.Parallel()

	issuer, issuerKey := createTestCA(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()
	crlService, err := NewCRLService(&CRLConfig{
		Issuer:     issuer,
		PrivateKey: issuerKey,
		Provider:   provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	})
	require.NoError(t, err)
	service, err := NewOCSPService(&OCSPConfig{
		Issuer:       issuer,
		Responder:    issuer,
		ResponderKey: issuerKey,
		Provider:     provider,
		Validity:     time.Hour,
	}, crlService)
	require.NoError(t, err)

	_, err = service.CreateResponseWithRevocation(nil, nil)
	require.ErrorContains(t, err, "certificate is required")
}
