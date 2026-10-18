// Copyright (c) 2025-2026 Justin Cranford.
package clientauth

import (
	"context"
	"crypto/x509"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type certificateAuthenticator interface {
	Authenticate(context.Context, string, string) (*cryptoutilIdentityDomain.Client, error)
}

type certificateAuthenticatorFactory func(
	cryptoutilIdentityRepository.ClientRepository,
	CertificateValidator,
) certificateAuthenticator

type rejectingCertificateValidator struct {
	err error
}

func (v rejectingCertificateValidator) ValidateCertificate(*x509.Certificate, [][]byte) error {
	return v.err
}

func (rejectingCertificateValidator) IsRevoked(*big.Int) bool {
	return false
}

func TestCertificateAuthenticators_ErrorPathsAndOptionalPins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	caCert, caKey := createTestCAForAuth(t)
	clientCert := createTestClientCertForAuth(t, caCert, caKey)
	credential := string(encodeCertToPEM(clientCert))
	roots := x509.NewCertPool()
	roots.AddCert(caCert)
	validValidator := NewCACertificateValidator(roots, nil)
	validValidator.SetValidationOptions(false, false)

	repositoryErr := errors.New("repository unavailable")
	validatorErr := errors.New("certificate rejected")

	authenticators := []struct {
		name   string
		method cryptoutilIdentityDomain.ClientAuthMethod
		new    certificateAuthenticatorFactory
	}{
		{
			name:   "self-signed",
			method: cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth,
			new: func(repo cryptoutilIdentityRepository.ClientRepository, validator CertificateValidator) certificateAuthenticator {
				return NewSelfSignedAuthenticator(repo, validator)
			},
		},
		{
			name:   "tls client",
			method: cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth,
			new: func(repo cryptoutilIdentityRepository.ClientRepository, validator CertificateValidator) certificateAuthenticator {
				return NewTLSClientAuthenticator(repo, validator)
			},
		},
	}

	for _, authenticatorCase := range authenticators {
		t.Run(authenticatorCase.name, func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name            string
				credential      string
				repositoryError error
				validatorError  error
				clientMethod    cryptoutilIdentityDomain.ClientAuthMethod
				wantErr         string
				wantSuccess     bool
			}{
				{name: "malformed certificate", credential: "not a certificate", wantErr: "failed to parse"},
				{name: "validator rejects certificate", credential: credential, validatorError: validatorErr, wantErr: "certificate validation failed"},
				{name: "repository lookup fails", credential: credential, repositoryError: repositoryErr, wantErr: clientauthTestClientLookupError},
				{name: "authentication method mismatch", credential: credential, clientMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic, wantErr: "Invalid client authentication"},
				{name: "certificate accepted without optional pins", credential: credential, wantSuccess: true},
			}

			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					clientMethod := tc.clientMethod
					if clientMethod == "" {
						clientMethod = authenticatorCase.method
					}

					client := &cryptoutilIdentityDomain.Client{
						ClientID:                cryptoutilSharedMagic.TestClientID,
						TokenEndpointAuthMethod: clientMethod,
						Enabled:                 boolPtr(true),
					}
					repository := &mockClientRepo{clients: map[string]*cryptoutilIdentityDomain.Client{
						client.ClientID: client,
					}}

					var clientRepo cryptoutilIdentityRepository.ClientRepository = repository
					if tc.repositoryError != nil {
						clientRepo = &lookupErrorClientRepository{
							mockClientRepo: *repository,
							err:            tc.repositoryError,
						}
					}

					validator := CertificateValidator(validValidator)
					if tc.validatorError != nil {
						validator = rejectingCertificateValidator{err: tc.validatorError}
					}

					result, err := authenticatorCase.new(clientRepo, validator).Authenticate(ctx, client.ClientID, tc.credential)
					if tc.wantErr != "" {
						require.ErrorContains(t, err, tc.wantErr)
						require.Nil(t, result)

						return
					}

					require.NoError(t, err)
					require.True(t, tc.wantSuccess)
					require.Same(t, client, result)
				})
			}
		})
	}
}
