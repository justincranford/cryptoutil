// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"crypto"
	crand "crypto/rand"
	"crypto/x509"
	"errors"
	"io"
	"testing"

	cryptoutilCAHandler "cryptoutil/internal/apps/sm-kms/server/pki/api/handler"
	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"

	"github.com/stretchr/testify/require"
)

const pkiComponentInjectedFailure = "injected factory failure"

func pkiComponentFactoriesForTest() pkiComponentFactories {
	provider := cryptoutilCACrypto.NewSoftwareProvider()

	return pkiComponentFactories{
		provider:             provider,
		storage:              cryptoutilCAStorage.NewMemoryStore(),
		randomReader:         crand.Reader,
		createSelfSignedCAFn: createSelfSignedCA,
		newIssuerFn:          cryptoutilCAServiceIssuer.NewIssuer,
		newCRLServiceFn:      cryptoutilCAServiceRevocation.NewCRLService,
		newOCSPServiceFn:     cryptoutilCAServiceRevocation.NewOCSPService,
		newHandlerFn:         cryptoutilCAHandler.NewHandler,
	}
}

func TestNewPKIComponents_FactoryFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*pkiComponentFactories)
		wantErr   string
	}{
		{
			name: "successful construction",
		},
		{
			name: "CA creation failure",
			configure: func(factories *pkiComponentFactories) {
				factories.createSelfSignedCAFn = func(_ cryptoutilCACrypto.Provider, _ io.Reader) (*x509.Certificate, crypto.PrivateKey, error) {
					return nil, nil, errors.New(pkiComponentInjectedFailure)
				}
			},
			wantErr: "failed to create CA",
		},
		{
			name: "issuer creation failure",
			configure: func(factories *pkiComponentFactories) {
				factories.newIssuerFn = func(_ cryptoutilCACrypto.Provider, _ *cryptoutilCAServiceIssuer.IssuingCAConfig) (*cryptoutilCAServiceIssuer.Issuer, error) {
					return nil, errors.New(pkiComponentInjectedFailure)
				}
			},
			wantErr: "failed to create issuer",
		},
		{
			name: "non-signer CA private key",
			configure: func(factories *pkiComponentFactories) {
				factories.createSelfSignedCAFn = func(provider cryptoutilCACrypto.Provider, randomReader io.Reader) (*x509.Certificate, crypto.PrivateKey, error) {
					certificate, _, err := createSelfSignedCA(provider, randomReader)
					if err != nil {
						return nil, nil, err
					}

					return certificate, struct{}{}, nil
				}
				factories.newIssuerFn = func(_ cryptoutilCACrypto.Provider, _ *cryptoutilCAServiceIssuer.IssuingCAConfig) (*cryptoutilCAServiceIssuer.Issuer, error) {
					return nil, nil
				}
			},
			wantErr: "CA private key is not a signer",
		},
		{
			name: "CRL service creation failure",
			configure: func(factories *pkiComponentFactories) {
				factories.newCRLServiceFn = func(_ *cryptoutilCAServiceRevocation.CRLConfig) (*cryptoutilCAServiceRevocation.CRLService, error) {
					return nil, errors.New(pkiComponentInjectedFailure)
				}
			},
			wantErr: "failed to create CRL service",
		},
		{
			name: "OCSP service creation failure",
			configure: func(factories *pkiComponentFactories) {
				factories.newOCSPServiceFn = func(_ *cryptoutilCAServiceRevocation.OCSPConfig, _ *cryptoutilCAServiceRevocation.CRLService) (*cryptoutilCAServiceRevocation.OCSPService, error) {
					return nil, errors.New(pkiComponentInjectedFailure)
				}
			},
			wantErr: "failed to create OCSP service",
		},
		{
			name: "handler creation failure",
			configure: func(factories *pkiComponentFactories) {
				factories.newHandlerFn = func(_ *cryptoutilCAServiceIssuer.Issuer, _ cryptoutilCAStorage.Store, _ map[string]*cryptoutilCAHandler.ProfileConfig) (*cryptoutilCAHandler.Handler, error) {
					return nil, errors.New(pkiComponentInjectedFailure)
				}
			},
			wantErr: "failed to create handler",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			factories := pkiComponentFactoriesForTest()
			if tc.configure != nil {
				tc.configure(&factories)
			}

			components, err := newPKIComponents(factories)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, components)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, components)
			require.NotNil(t, components.issuer)
			require.NotNil(t, components.storage)
			require.NotNil(t, components.crlService)
			require.NotNil(t, components.ocspService)
			require.NotNil(t, components.handler)
		})
	}
}
