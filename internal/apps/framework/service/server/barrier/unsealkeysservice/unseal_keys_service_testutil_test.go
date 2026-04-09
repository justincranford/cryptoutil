// Copyright (c) 2025 Justin Cranford

package unsealkeysservice

import (
	"testing"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	"github.com/stretchr/testify/require"
)

func TestRequireNewSimpleForTest(t *testing.T) {
	t.Parallel()

	unsealKeys, _, err := cryptoutilSharedCryptoJose.GenerateJWEJWKsForTest(t, 2, &cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	svc := RequireNewSimpleForTest(unsealKeys)
	require.NotNil(t, svc)
}
