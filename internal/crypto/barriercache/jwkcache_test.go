package barriercache

import (
	"errors"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func TestJWKCache_HappyPath_GetLatest(t *testing.T) {
	kid, mockAesJwk := mockJWKKey()
	mockLoadLatestFunc := func() (*Entry, error) {
		return &Entry{Key: kid, Value: mockAesJwk}, nil
	}

	jwkCache, _ := NewJWKCache(10, mockLoadLatestFunc, nil, nil)
	aesJwk, err := jwkCache.GetLatest()

	require.NoError(t, err)
	require.NotNil(t, aesJwk)
	require.Equal(t, mockAesJwk, aesJwk)
}

func TestJWKCache_HappyPath_Get(t *testing.T) {
	kid, mockAesJwk := mockJWKKey()
	mockLoadFunc := func(kid googleUuid.UUID) (joseJwk.Key, error) {
		return mockAesJwk, nil
	}

	jwkCache, _ := NewJWKCache(10, nil, mockLoadFunc, nil)
	aesJwk, err := jwkCache.Get(kid)

	require.NoError(t, err)
	require.NotNil(t, aesJwk)
	require.Equal(t, mockAesJwk, aesJwk)
}

func TestJWKCache_HappyPath_Put(t *testing.T) {
	mockStoreFunc := func(kid googleUuid.UUID, jwk joseJwk.Key, parentUuid googleUuid.UUID) error {
		return nil
	}

	jwkCache, _ := NewJWKCache(10, nil, nil, mockStoreFunc)
	kid, aesJwk := mockJWKKey()
	err := jwkCache.Put(kid, aesJwk, googleUuid.Nil)

	require.NoError(t, err)
}

func TestJWKCache_SadPath_CacheSize(t *testing.T) {
	_, err := NewJWKCache(0, nil, nil, nil)

	require.Error(t, err)
	require.Equal(t, "failed to create LRU cache: must provide a positive size", err.Error())
}

func TestJWKCache_SadPath_GetLatest_Function(t *testing.T) {
	dbErr := fmt.Errorf("database error")
	mockLoadLatestFunc := func() (*Entry, error) {
		return nil, dbErr
	}

	jwkCache, _ := NewJWKCache(10, mockLoadLatestFunc, nil, nil)
	_, err := jwkCache.GetLatest()

	require.Error(t, err)
	require.True(t, errors.Is(err, dbErr))
}

func TestJWKCache_SadPath_Get_Function(t *testing.T) {
	mockLoadFunc := func(kid googleUuid.UUID) (joseJwk.Key, error) {
		return nil, fmt.Errorf("database error")
	}

	jwkCache, _ := NewJWKCache(10, nil, mockLoadFunc, nil)
	kid := googleUuid.Must(googleUuid.NewV7())
	_, err := jwkCache.Get(kid)

	require.Error(t, err)
}

func TestJWKCache_Put_SadPath(t *testing.T) {
	mockStoreFunc := func(kid googleUuid.UUID, jwk joseJwk.Key, parentUuid googleUuid.UUID) error {
		return fmt.Errorf("database error")
	}

	jwkCache, _ := NewJWKCache(10, nil, nil, mockStoreFunc)
	kid, aesJwk := mockJWKKey()
	err := jwkCache.Put(kid, aesJwk, googleUuid.Nil)

	require.Error(t, err)
}
