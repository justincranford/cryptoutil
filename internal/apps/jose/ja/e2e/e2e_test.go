// Copyright (c) 2025 Justin Cranford

//go:build e2e

package e2e_test

import (
"context"
http "net/http"
"testing"

cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

"github.com/stretchr/testify/require"
)


// TestE2E_HealthChecks validates /health endpoint for all jose-ja instances.
func TestE2E_HealthChecks(t *testing.T) {
t.Parallel()

tests := []struct {
name      string
publicURL string
}{
{sqliteContainer, sqlitePublicURL},
{postgres1Container, postgres1PublicURL},
{postgres2Container, postgres2PublicURL},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
t.Parallel()

ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
defer cancel()

healthURL := tt.publicURL + cryptoutilSharedMagic.JoseJAE2EHealthEndpoint

req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
require.NoError(t, err, "Creating health check request should succeed")

healthResp, err := sharedHTTPClient.Do(req)
require.NoError(t, err, "Health check should succeed for %s", tt.name)
require.NoError(t, healthResp.Body.Close())
require.Equal(t, http.StatusOK, healthResp.StatusCode,
"%s should return 200 OK for /health", tt.name)
})
}
}
