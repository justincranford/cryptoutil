// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentitySpaServer "cryptoutil/internal/apps/identity-spa/server"
)

// TestServeOpenAPISpec_Success validates successful OpenAPI spec serving.
func TestServeOpenAPISpec_Success(t *testing.T) {
	t.Parallel()

	handler, err := cryptoutilIdentitySpaServer.ServeOpenAPISpec()
	require.NoError(t, err)
	require.NotNil(t, handler)
}
