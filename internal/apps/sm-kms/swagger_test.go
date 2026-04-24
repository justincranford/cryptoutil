// Copyright (c) 2025 Justin Cranford
//

package kms_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerTestutil "cryptoutil/internal/apps/framework/service/server/testutil"
	cryptoutilkms "cryptoutil/internal/apps/sm-kms"
)

// TestServeOpenAPISpec_Success validates successful OpenAPI spec serving.
func TestServeOpenAPISpec_Success(t *testing.T) {
	t.Parallel()

	handler, err := cryptoutilkms.ServeOpenAPISpec()
	require.NoError(t, err)
	require.NotNil(t, handler)

	cryptoutilAppsFrameworkServiceServerTestutil.AssertOpenAPISpecHandler(t, handler)
}

// TestServeOpenAPISpec_HandlerInvocation validates handler can be invoked multiple times.
func TestServeOpenAPISpec_HandlerInvocation(t *testing.T) {
	t.Parallel()

	handler, err := cryptoutilkms.ServeOpenAPISpec()
	require.NoError(t, err)

	for range 3 {
		cryptoutilAppsFrameworkServiceServerTestutil.AssertOpenAPISpecHandler(t, handler)
	}
}
