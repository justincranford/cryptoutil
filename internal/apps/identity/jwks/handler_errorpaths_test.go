// Copyright (c) 2025 Justin Cranford

package jwks

import (
	"errors"
	"fmt"
	"log/slog"
	http "net/http"
	"net/http/httptest"
	"testing"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// failingResponseWriter wraps httptest.ResponseRecorder but fails on Write.
type failingResponseWriter struct {
	*httptest.ResponseRecorder
}

func (fw *failingResponseWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("simulated write failure")
}

func TestHandler_ServeHTTP_JSONMarshalError(t *testing.T) {
	t.Parallel()

	logger := slog.Default()
	keyRepo := &MockKeyRepository{}
	keyRepo.On("FindByUsage", mock.Anything, cryptoutilIdentityMagic.KeyUsageSigning, true).
		Return([]*cryptoutilIdentityDomain.Key{}, nil)

	handler, err := NewHandler(logger, keyRepo)
	require.NoError(t, err)

	// Override the instance-level marshal function to simulate failure.
	handler.marshalJSONFn = func(_ any) ([]byte, error) {
		return nil, errors.New("simulated marshal failure")
	}

	req := httptest.NewRequest(http.MethodGet, cryptoutilIdentityMagic.PathJWKS, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	closeErr := resp.Body.Close()
	require.NoError(t, closeErr)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestHandler_ServeHTTP_WriteError(t *testing.T) {
	t.Parallel()

	logger := slog.Default()
	keyRepo := &MockKeyRepository{}
	keyRepo.On("FindByUsage", mock.Anything, cryptoutilIdentityMagic.KeyUsageSigning, true).
		Return([]*cryptoutilIdentityDomain.Key{}, nil)

	handler, err := NewHandler(logger, keyRepo)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, cryptoutilIdentityMagic.PathJWKS, nil)
	w := &failingResponseWriter{ResponseRecorder: httptest.NewRecorder()}

	handler.ServeHTTP(w, req)

	// Write fails, but handler should not panic.
	require.Equal(t, http.StatusOK, w.Code)
}
