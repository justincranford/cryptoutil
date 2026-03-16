// Copyright (c) 2025 Justin Cranford

package handler

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/ca/server"
	cryptoutilCAStorage "cryptoutil/internal/apps/pki/ca/storage"
)

// errorStore is a Store that returns errors for specific operations.
type errorStore struct {
	inner     cryptoutilCAStorage.Store
	listErr   error
	getErr    error
	revokeErr error
	storeErr  error
}

func (s *errorStore) Store(ctx context.Context, cert *cryptoutilCAStorage.StoredCertificate) error {
	if s.storeErr != nil {
		return s.storeErr
	}

	return s.inner.Store(ctx, cert)
}

func (s *errorStore) Get(ctx context.Context, id googleUuid.UUID) (*cryptoutilCAStorage.StoredCertificate, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}

	return s.inner.Get(ctx, id)
}

func (s *errorStore) GetBySerialNumber(ctx context.Context, serialNumber string) (*cryptoutilCAStorage.StoredCertificate, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}

	return s.inner.GetBySerialNumber(ctx, serialNumber)
}

func (s *errorStore) List(ctx context.Context, filter *cryptoutilCAStorage.ListFilter) ([]*cryptoutilCAStorage.StoredCertificate, int, error) {
	if s.listErr != nil {
		return nil, 0, s.listErr
	}

	return s.inner.List(ctx, filter)
}

func (s *errorStore) Update(ctx context.Context, cert *cryptoutilCAStorage.StoredCertificate) error {
	return s.inner.Update(ctx, cert)
}

func (s *errorStore) Delete(ctx context.Context, id googleUuid.UUID) error {
	return s.inner.Delete(ctx, id)
}

func (s *errorStore) Revoke(ctx context.Context, id googleUuid.UUID, reason cryptoutilCAStorage.RevocationReason) error {
	if s.revokeErr != nil {
		return s.revokeErr
	}

	return s.inner.Revoke(ctx, id, reason)
}

func (s *errorStore) GetRevoked(ctx context.Context, issuerDN string) ([]*cryptoutilCAStorage.StoredCertificate, error) {
	return s.inner.GetRevoked(ctx, issuerDN)
}

func (s *errorStore) CountByStatus(ctx context.Context) (map[cryptoutilCAStorage.CertificateStatus]int64, error) {
	return s.inner.CountByStatus(ctx)
}

func (s *errorStore) Close() error {
	return s.inner.Close()
}

// TestListCertificates_StorageError tests ListCertificates when storage returns an error.
func TestListCertificates_StorageError(t *testing.T) {
	t.Parallel()

	store := &errorStore{
		inner:   cryptoutilCAStorage.NewMemoryStore(),
		listErr: errors.New("database connection failed"),
	}
	handler := &Handler{storage: store}

	app := fiber.New()
	app.Get("/certificates", func(c *fiber.Ctx) error {
		params := cryptoutilApiCaServer.ListCertificatesParams{}

		return handler.ListCertificates(c, params)
	})

	req := httptest.NewRequest(http.MethodGet, "/certificates", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestGetCertificate_EmptySerial tests GetCertificate with empty serial number.
func TestGetCertificate_EmptySerial(t *testing.T) {
	t.Parallel()

	handler := &Handler{storage: cryptoutilCAStorage.NewMemoryStore()}

	app := fiber.New()
	app.Get("/certificates/empty", func(c *fiber.Ctx) error {
		return handler.GetCertificate(c, "")
	})

	req := httptest.NewRequest(http.MethodGet, "/certificates/empty", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestGetCertificate_InternalStorageError tests GetCertificate when storage returns a non-NotFound error.
func TestGetCertificate_InternalStorageError(t *testing.T) {
	t.Parallel()

	store := &errorStore{
		inner:  cryptoutilCAStorage.NewMemoryStore(),
		getErr: errors.New("internal database error"),
	}
	handler := &Handler{storage: store}

	app := fiber.New()
	app.Get("/certificates/:sn", func(c *fiber.Ctx) error {
		return handler.GetCertificate(c, "SERIAL999")
	})

	req := httptest.NewRequest(http.MethodGet, "/certificates/SERIAL999", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestGetCertificateChain_EmptySerial tests GetCertificateChain with empty serial number.
func TestGetCertificateChain_EmptySerial(t *testing.T) {
	t.Parallel()

	handler := &Handler{storage: cryptoutilCAStorage.NewMemoryStore()}

	app := fiber.New()
	app.Get("/certificates/empty/chain", func(c *fiber.Ctx) error {
		return handler.GetCertificateChain(c, "")
	})

	req := httptest.NewRequest(http.MethodGet, "/certificates/empty/chain", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestGetCertificateChain_NotFoundSerial tests GetCertificateChain when certificate not found.
func TestGetCertificateChain_NotFoundSerial(t *testing.T) {
	t.Parallel()

	handler := &Handler{storage: cryptoutilCAStorage.NewMemoryStore()}

	app := fiber.New()
	app.Get("/certificates/:sn/chain", func(c *fiber.Ctx) error {
		return handler.GetCertificateChain(c, "NONEXISTENT")
	})

	req := httptest.NewRequest(http.MethodGet, "/certificates/NONEXISTENT/chain", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestGetCertificateChain_InternalStorageError tests GetCertificateChain when storage returns a non-NotFound error.
func TestGetCertificateChain_InternalStorageError(t *testing.T) {
	t.Parallel()

	store := &errorStore{
		inner:  cryptoutilCAStorage.NewMemoryStore(),
		getErr: errors.New("internal database error"),
	}
	handler := &Handler{storage: store}

	app := fiber.New()
	app.Get("/certificates/:sn/chain", func(c *fiber.Ctx) error {
		return handler.GetCertificateChain(c, "SERIAL999")
	})

	req := httptest.NewRequest(http.MethodGet, "/certificates/SERIAL999/chain", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestRevokeCertificate_InternalStorageError tests RevokeCertificate when GetBySerialNumber returns non-NotFound error.
func TestRevokeCertificate_InternalStorageError(t *testing.T) {
	t.Parallel()

	store := &errorStore{
		inner:  cryptoutilCAStorage.NewMemoryStore(),
		getErr: errors.New("internal database error"),
	}
	handler := &Handler{storage: store}

	app := fiber.New()
	app.Post("/certificates/:sn/revoke", func(c *fiber.Ctx) error {
		return handler.RevokeCertificate(c, "SERIAL999")
	})

	req := httptest.NewRequest(http.MethodPost, "/certificates/SERIAL999/revoke",
		bytes.NewBufferString(`{"reason":"key_compromise"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestRevokeCertificate_RevokeStorageError tests RevokeCertificate when Revoke returns an error.
func TestRevokeCertificate_RevokeStorageError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	inner := cryptoutilCAStorage.NewMemoryStore()
	cert := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   "REVOKE_ERR_SN",
		SubjectDN:      "CN=test.example.com",
		IssuerDN:       "CN=Test CA",
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
	}
	require.NoError(t, inner.Store(ctx, cert))

	store := &errorStore{
		inner:     inner,
		revokeErr: errors.New("revocation storage failure"),
	}
	handler := &Handler{storage: store}

	app := fiber.New()
	app.Post("/certificates/:sn/revoke", func(c *fiber.Ctx) error {
		return handler.RevokeCertificate(c, "REVOKE_ERR_SN")
	})

	req := httptest.NewRequest(http.MethodPost, "/certificates/REVOKE_ERR_SN/revoke",
		bytes.NewBufferString(`{"reason":"key_compromise"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestSubmitEnrollment_ErrorPaths tests various SubmitEnrollment error conditions.
func TestSubmitEnrollment_ErrorPaths(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	profiles := map[string]*ProfileConfig{
		"tls-server": {
			ID:       "tls-server",
			Name:     "TLS Server",
			Category: "tls",
		},
	}

	tests := []struct {
		name           string
		body           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "invalid_json_body",
			body:           `{invalid json`,
			contentType:    "application/json",
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name:           "missing_csr",
			body:           `{"csr":"","profile":"tls-server"}`,
			contentType:    "application/json",
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name:           "missing_profile",
			body:           `{"csr":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0K","profile":""}`,
			contentType:    "application/json",
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name:           "invalid_csr",
			body:           `{"csr":"not-a-valid-csr","profile":"tls-server"}`,
			contentType:    "application/json",
			expectedStatus: fiber.StatusUnprocessableEntity,
		},
		{
			name:           "unknown_profile_invalid_csr",
			body:           `{"csr":"not-a-valid-csr","profile":"nonexistent-profile"}`,
			contentType:    "application/json",
			expectedStatus: fiber.StatusUnprocessableEntity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				issuer:            testSetup.Issuer,
				storage:           cryptoutilCAStorage.NewMemoryStore(),
				profiles:          profiles,
				enrollmentTracker: newEnrollmentTracker(cryptoutilSharedMagic.JoseJAMaxMaterials),
			}

			app := fiber.New()
			app.Post("/enroll", func(c *fiber.Ctx) error {
				return handler.SubmitEnrollment(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/enroll",
				bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", tc.contentType)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			require.NoError(t, resp.Body.Close())
		})
	}
}
