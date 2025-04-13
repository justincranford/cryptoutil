package barrierrepository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

var (
	testCtx              = context.Background()
	testTelemetryService *cryptoutilTelemetry.TelemetryService
)

func TestMain(m *testing.M) {
	testTelemetryService = cryptoutilTelemetry.RequireNewForTest(testCtx, "barrierrepository_test", false, false)
	defer testTelemetryService.Shutdown()
	os.Exit(m.Run())
}

// Happy Path

func TestJWKCache_HappyPath_GetLatest(t *testing.T) {
	var mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction
	jwk, _, _ := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgA256GCMKW)
	mockLoadLatestFunc := func(mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction) (joseJwk.Key, error) {
		return jwk, nil
	}

	jwkCache, err := NewBarrierRepository("TestJWKCache_HappyPath_GetLatest", testTelemetryService, 10, mockLoadLatestFunc, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	aesJwk, err := jwkCache.GetLatest(mockSqlTransaction)
	require.NoError(t, err)
	require.NotNil(t, aesJwk)
	require.Equal(t, jwk, aesJwk)
}

func TestJWKCache_HappyPath_Get(t *testing.T) {
	var mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction
	jwk, _, _ := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgA256GCMKW)
	mockLoadFunc := func(mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction, kid googleUuid.UUID) (joseJwk.Key, error) {
		return jwk, nil
	}

	kid, err := cryptoutilJose.ExtractKidUuid(jwk)
	require.NoError(t, err)

	jwkCache, err := NewBarrierRepository("TestJWKCache_HappyPath_Get", testTelemetryService, 10, nil, mockLoadFunc, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	cachedJwk, err := jwkCache.Get(mockSqlTransaction, kid)

	require.NoError(t, err)
	require.NotNil(t, cachedJwk)
	require.Equal(t, jwk, cachedJwk)
}

func TestJWKCache_HappyPath_Put(t *testing.T) {
	var mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction
	jwk, _, _ := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgA256GCMKW)
	mockStoreFunc := func(mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction, jwk joseJwk.Key) error {
		return nil
	}

	jwkCache, err := NewBarrierRepository("TestJWKCache_HappyPath_Put", testTelemetryService, 10, nil, nil, mockStoreFunc, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	err = jwkCache.Put(mockSqlTransaction, jwk)

	require.NoError(t, err)
}

// Sad Path

func TestJWKCache_SadPath_CacheSize(t *testing.T) {
	jwkCache, err := NewBarrierRepository("TestJWKCache_SadPath_CacheSize", testTelemetryService, 0, nil, nil, nil, nil)
	require.Error(t, err)
	require.Nil(t, jwkCache)
	require.Equal(t, "failed to create LRU cache: must provide a positive size", err.Error())
}

func TestJWKCache_SadPath_GetLatest_Function(t *testing.T) {
	var mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction
	dbErr := fmt.Errorf("database error")
	mockLoadLatestFunc := func(mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction) (joseJwk.Key, error) {
		return nil, dbErr
	}

	jwkCache, err := NewBarrierRepository("TestJWKCache_SadPath_GetLatest_Function", testTelemetryService, 10, mockLoadLatestFunc, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	latestJwk, err := jwkCache.GetLatest(mockSqlTransaction)
	require.Error(t, err)
	require.Nil(t, latestJwk)
	require.True(t, errors.Is(err, dbErr))
}

func TestJWKCache_SadPath_Get_Function(t *testing.T) {
	var mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction
	mockLoadFunc := func(mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction, kid googleUuid.UUID) (joseJwk.Key, error) {
		return nil, fmt.Errorf("database error")
	}

	jwkCache, err := NewBarrierRepository("TestJWKCache_SadPath_Get_Function", testTelemetryService, 10, nil, mockLoadFunc, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	kid := googleUuid.Must(googleUuid.NewV7())
	cachedJwk, err := jwkCache.Get(mockSqlTransaction, kid)
	require.Nil(t, cachedJwk)
	require.Error(t, err)
}

func TestJWKCache_Put_SadPath(t *testing.T) {
	var mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction
	jwk, _, _ := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgA256GCMKW)

	mockStoreFunc := func(mockSqlTransaction *cryptoutilOrmRepository.OrmTransaction, jwk joseJwk.Key) error {
		return nil
	}

	jwkCache, err := NewBarrierRepository("TestJWKCache_Put_SadPath", testTelemetryService, 10, nil, nil, mockStoreFunc, nil)
	require.NoError(t, err)
	require.NotNil(t, jwkCache)

	err = jwkCache.Put(mockSqlTransaction, jwk)
	require.NoError(t, err)
}
