// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	"testing"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

const jwkCount = 2

func TestNewUnsealKeysServiceSimple_HappyPath(t *testing.T) {
	unsealKeys, _, err := cryptoutilJose.GenerateJWEJWKsForTest(t, jwkCount, &cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	unsealKeysService, err := NewUnsealKeysServiceSimple(unsealKeys)
	require.NoError(t, err)
	require.NotNil(t, unsealKeysService)
}

func TestNewUnsealKeysServiceSimple_SadPath_EmptyInput(t *testing.T) {
	unsealKeysService, err := NewUnsealKeysServiceSimple([]joseJwk.Key{})
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.EqualError(t, err, "unsealJWKs can't be empty", "Unexpected error message")
}

func TestNewUnsealKeysServiceSimple_SadPath_NilInput(t *testing.T) {
	unsealKeysService, err := NewUnsealKeysServiceSimple(nil)
	require.Error(t, err)
	require.Nil(t, unsealKeysService)
	require.EqualError(t, err, "unsealJWKs can't be nil", "Unexpected error message")
}
