package barriercache

import (
	"context"
	"cryptoutil/internal/crypto/jose"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"

	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.Service
)

func TestMain(m *testing.M) {
	var err error

	testTelemetryService, err = cryptoutilTelemetry.NewService(testCtx, "barriercache_test", false, false)
	if err != nil {
		slog.Error("failed to initailize telemetry", "error", err)
		os.Exit(-1)
	}
	defer testTelemetryService.Shutdown()

	os.Exit(m.Run())
}

// Happy Path

func TestJWKCache_HappyPath_GetLatest(t *testing.T) {
	jwk := mockJWKKey()
	mockLoadLatestFunc := func() (joseJwk.Key, error) {
		return jwk, nil
	}

	jwkCache, err := NewJWKCache(testTelemetryService, 10, mockLoadLatestFunc, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	aesJwk, err := jwkCache.GetLatest()
	require.NoError(t, err)
	require.NotNil(t, aesJwk)
	require.Equal(t, jwk, aesJwk)
}

func TestJWKCache_HappyPath_Get(t *testing.T) {
	jwk := mockJWKKey()
	mockLoadFunc := func(kid googleUuid.UUID) (joseJwk.Key, error) {
		return jwk, nil
	}

	kid, err := jose.GetKidUuid(jwk)
	require.NoError(t, err)

	jwkCache, err := NewJWKCache(testTelemetryService, 10, nil, mockLoadFunc, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	cachedJwk, err := jwkCache.Get(kid)

	require.NoError(t, err)
	require.NotNil(t, cachedJwk)
	require.Equal(t, jwk, cachedJwk)
}

func TestJWKCache_HappyPath_Put(t *testing.T) {
	jwk := mockJWKKey()
	kek := mockJWKKey()
	mockStoreFunc := func(jwk joseJwk.Key, kek joseJwk.Key) error {
		return nil
	}

	jwkCache, err := NewJWKCache(testTelemetryService, 10, nil, nil, mockStoreFunc, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	err = jwkCache.Put(jwk, kek)

	require.NoError(t, err)
}

// Sad Path

func TestJWKCache_SadPath_CacheSize(t *testing.T) {
	jwkCache, err := NewJWKCache(testTelemetryService, 0, nil, nil, nil, nil)
	require.Error(t, err)
	require.Nil(t, jwkCache)
	require.Equal(t, "failed to create LRU cache: must provide a positive size", err.Error())
}

func TestJWKCache_SadPath_GetLatest_Function(t *testing.T) {
	dbErr := fmt.Errorf("database error")
	mockLoadLatestFunc := func() (joseJwk.Key, error) {
		return nil, dbErr
	}

	jwkCache, err := NewJWKCache(testTelemetryService, 10, mockLoadLatestFunc, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	latestJwk, err := jwkCache.GetLatest()
	require.Error(t, err)
	require.Nil(t, latestJwk)
	require.True(t, errors.Is(err, dbErr))
}

func TestJWKCache_SadPath_Get_Function(t *testing.T) {
	mockLoadFunc := func(kid googleUuid.UUID) (joseJwk.Key, error) {
		return nil, fmt.Errorf("database error")
	}

	jwkCache, err := NewJWKCache(testTelemetryService, 10, nil, mockLoadFunc, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	kid := googleUuid.Must(googleUuid.NewV7())
	cachedJwk, err := jwkCache.Get(kid)
	require.Nil(t, cachedJwk)
	require.Error(t, err)
}

func TestJWKCache_Put_SadPath(t *testing.T) {
	jwk := mockJWKKey()
	kek := mockJWKKey()

	mockStoreFunc := func(jwk joseJwk.Key, kek joseJwk.Key) error {
		return nil
	}

	jwkCache, err := NewJWKCache(testTelemetryService, 10, nil, nil, mockStoreFunc, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	err = jwkCache.Put(jwk, kek)
	require.NoError(t, err)
}
