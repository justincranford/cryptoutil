// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityRpServer "cryptoutil/internal/apps/identity-rp/server"
)

// TestServeOpenAPISpec_Success validates successful OpenAPI spec serving.
func TestServeOpenAPISpec_Success(t *testing.T) {
	t.Parallel()

	handler, err := cryptoutilIdentityRpServer.ServeOpenAPISpec()
	require.NoError(t, err)
	require.NotNil(t, handler)
}
