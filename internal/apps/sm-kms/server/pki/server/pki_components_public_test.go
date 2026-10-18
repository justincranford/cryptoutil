// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	crand "crypto/rand"
	"crypto/x509"
	"errors"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilAppsFrameworkTestDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilCAHandler "cryptoutil/internal/apps/sm-kms/server/pki/api/handler"
	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	pkiPublicTestHistoricalAuthorityID = "retired-test-authority"
	pkiPublicTestAuthorityPath         = "/pki/cas/"
	pkiPublicTestOCSPPath              = "/pki/ocsp"
)

func TestPKIComponentConstructorsAndRegistrationValidation(t *testing.T) {
	t.Parallel()

	t.Run("default in-memory constructor", func(t *testing.T) {
		t.Parallel()

		components, err := NewInMemoryPKIComponents()
		require.NoError(t, err)
		require.NotNil(t, components)
	})

	t.Run("nil components", func(t *testing.T) {
		t.Parallel()

		var components *PKIComponents
		require.ErrorContains(t, components.RegisterPublicRevocationEndpoints(fiber.New()), "PKI handler is required")
	})

	t.Run("nil application", func(t *testing.T) {
		t.Parallel()

		components, err := NewInMemoryPKIComponents()
		require.NoError(t, err)
		require.ErrorContains(t, components.RegisterPublicRevocationEndpoints(nil), "fiber application is required")
	})

	t.Run("nil handler", func(t *testing.T) {
		t.Parallel()

		components, err := NewInMemoryPKIComponents()
		require.NoError(t, err)

		components.handler = nil
		require.ErrorContains(t, components.RegisterPublicRevocationEndpoints(fiber.New()), "PKI handler is required")
	})

	t.Run("constructor argument validation", func(t *testing.T) {
		t.Parallel()

		_, err := NewPKIComponents(nilPKITestContext(), nil, nil)
		require.ErrorContains(t, err, "context is required")

		_, err = NewPKIComponents(context.Background(), nil, nil)
		require.ErrorContains(t, err, "database is required")

		db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
		_, err = NewPKIComponentsWithAuthorityConfig(
			context.Background(),
			nil,
			nil,
			DefaultPKIAuthorityConfig(),
		)
		require.ErrorContains(t, err, "database is required")

		_, err = NewPKIComponentsWithAuthorityConfig(
			context.Background(),
			db,
			nil,
			DefaultPKIAuthorityConfig(),
		)
		require.ErrorContains(t, err, "barrier service is required")

		invalidConfig := DefaultPKIAuthorityConfig()
		invalidConfig.ID = ""
		_, err = NewPKIComponentsWithAuthorityConfig(
			context.Background(),
			db,
			&cryptoutilAppsFrameworkServiceServerBarrier.Service{},
			invalidConfig,
		)
		require.ErrorContains(t, err, "invalid PKI authority configuration")
	})
}

func nilPKITestContext() context.Context {
	return nil
}

func TestPKIComponentsRegisterPublicRevocationEndpoints(t *testing.T) {
	t.Parallel()

	components, err := NewInMemoryPKIComponents()
	require.NoError(t, err)

	app := fiber.New()
	require.NoError(t, components.RegisterPublicRevocationEndpoints(app))

	activeAuthorityID := components.issuer.GetCAConfig().Name
	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
	}{
		{
			name:       "unknown CA CRL",
			method:     http.MethodGet,
			target:     pkiPublicTestAuthorityPath + "unknown/crl",
			wantStatus: fiber.StatusNotFound,
		},
		{
			name:       "active CA DER CRL",
			method:     http.MethodGet,
			target:     pkiPublicTestAuthorityPath + activeAuthorityID + "/crl",
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "active CA PEM CRL",
			method:     http.MethodGet,
			target:     pkiPublicTestAuthorityPath + activeAuthorityID + "/crl?format=pem",
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "OCSP endpoint rejects an empty body",
			method:     http.MethodPost,
			target:     pkiPublicTestOCSPPath,
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "unknown CA OCSP",
			method:     http.MethodPost,
			target:     pkiPublicTestOCSPPath + "/unknown",
			wantStatus: fiber.StatusNotFound,
		},
		{
			name:       "active CA OCSP rejects an empty body",
			method:     http.MethodPost,
			target:     pkiPublicTestOCSPPath + "/" + activeAuthorityID,
			wantStatus: fiber.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(tc.method, tc.target, nil)
			response, requestErr := app.Test(request, -1)
			require.NoError(t, requestErr)
			require.Equal(t, tc.wantStatus, response.StatusCode)
			require.NoError(t, response.Body.Close())
		})
	}
}

func TestNewPKIComponentsWithAuthorityConfigClosedDatabase(t *testing.T) {
	t.Parallel()

	db := cryptoutilAppsFrameworkTestDB.NewInMemorySQLiteDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	_, err = NewPKIComponentsWithAuthorityConfig(
		context.Background(),
		db,
		&cryptoutilAppsFrameworkServiceServerBarrier.Service{},
		DefaultPKIAuthorityConfig(),
	)
	require.ErrorContains(t, err, "failed to create PKI certificate store")
}

func TestNewPKIComponents_RestoresHistoricalAuthorities(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	certificate, privateKey, err := createSelfSignedCA(provider, crand.Reader)
	require.NoError(t, err)
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	factories := pkiComponentFactoriesForTest()
	factories.authorityID = pkiComponentStableAuthorityID
	factories.ctx = context.Background()
	factories.listAuthoritiesFn = func(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error) {
		return []*cryptoutilCAStorage.StoredAuthority{{
			ID:                        pkiPublicTestHistoricalAuthorityID,
			CertificateDER:            certificate.Raw,
			PrivateKeyEncrypted:       privateKeyDER,
			PrivateKeyEnvelopeVersion: cryptoutilSharedMagic.PKIPrivateKeyEnvelopeVersionV1,
		}}, nil
	}
	factories.decryptAuthorityKeyFn = func(_ context.Context, ciphertext []byte) ([]byte, error) {
		return append([]byte(nil), ciphertext...), nil
	}

	components, err := newPKIComponents(factories)
	require.NoError(t, err)
	require.Contains(t, components.publicRevocationEndpoints, pkiComponentStableAuthorityID)
	require.Contains(t, components.publicRevocationEndpoints, components.issuer.GetCAConfig().Name)
	require.Contains(t, components.publicRevocationEndpoints, pkiPublicTestHistoricalAuthorityID)

	app := fiber.New()
	require.NoError(t, components.RegisterPublicRevocationEndpoints(app))

	request := httptest.NewRequest(
		http.MethodGet,
		pkiPublicTestAuthorityPath+pkiPublicTestHistoricalAuthorityID+"/crl",
		nil,
	)
	response, err := app.Test(request, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, response.StatusCode)
	require.NoError(t, response.Body.Close())
}

func TestNewPKIComponents_HistoricalAuthorityFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		listFn  func(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error)
		wantErr string
	}{
		{
			name: "list authorities error",
			listFn: func(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error) {
				return nil, errors.New("injected authority list failure")
			},
			wantErr: pkiTestFailedListPersistedPKIAuthorities,
		},
		{
			name: "nil authority in list",
			listFn: func(context.Context) ([]*cryptoutilCAStorage.StoredAuthority, error) {
				return []*cryptoutilCAStorage.StoredAuthority{nil}, nil
			},
			wantErr: pkiTestAuthorityNilEntry,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			factories := pkiComponentFactoriesForTest()
			factories.authorityID = pkiComponentStableAuthorityID
			factories.listAuthoritiesFn = tc.listFn

			components, err := newPKIComponents(factories)
			require.ErrorContains(t, err, tc.wantErr)
			require.Nil(t, components)
		})
	}
}

func TestAddPublicRevocationEndpoint(t *testing.T) {
	t.Parallel()

	handler := &cryptoutilCAHandler.Handler{}
	otherHandler := &cryptoutilCAHandler.Handler{}
	endpoint := publicRevocationEndpoint{handler: handler, issuerName: pkiPublicTestHistoricalAuthorityID}
	endpoints := make(map[string]publicRevocationEndpoint)

	require.ErrorContains(t, addPublicRevocationEndpoint(endpoints, "", endpoint), "authority ID is required")
	require.NoError(t, addPublicRevocationEndpoint(endpoints, pkiPublicTestHistoricalAuthorityID, endpoint))
	require.NoError(t, addPublicRevocationEndpoint(endpoints, pkiPublicTestHistoricalAuthorityID, endpoint))
	require.ErrorContains(
		t,
		addPublicRevocationEndpoint(
			endpoints,
			pkiPublicTestHistoricalAuthorityID,
			publicRevocationEndpoint{handler: otherHandler, issuerName: pkiPublicTestHistoricalAuthorityID},
		),
		"duplicate public revocation endpoint authority ID",
	)
}
