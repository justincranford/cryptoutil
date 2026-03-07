// Copyright (c) 2025 Justin Cranford
//

package contract

import (
"fmt"
http "net/http"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"

cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// RunServerContracts verifies server infrastructure contracts.
// Tests 6 contracts:
//  1. Public port is dynamically allocated (> 0)
//  2. Admin port is dynamically allocated (> 0)
//  3. Public and admin ports are different (server isolation)
//  4. Public base URL uses HTTPS scheme
//  5. Admin base URL uses HTTPS scheme
//  6. Admin path is NOT accessible from the public port (server isolation)
func RunServerContracts(t *testing.T, server ServiceServer) {
t.Helper()

t.Run("public_port_is_positive", func(t *testing.T) {
t.Parallel()

assert.Greater(t, server.PublicPort(), 0, "public port should be dynamically allocated (> 0)")
})

t.Run("admin_port_is_positive", func(t *testing.T) {
t.Parallel()

assert.Greater(t, server.AdminPort(), 0, "admin port should be dynamically allocated (> 0)")
})

t.Run("public_and_admin_ports_differ", func(t *testing.T) {
t.Parallel()

assert.NotEqual(t, server.PublicPort(), server.AdminPort(), "public and admin ports must be different for server isolation")
})

t.Run("public_base_url_uses_https", func(t *testing.T) {
t.Parallel()

assert.Contains(t, server.PublicBaseURL(), "https://", "public server must use HTTPS")
})

t.Run("admin_base_url_uses_https", func(t *testing.T) {
t.Parallel()

assert.Contains(t, server.AdminBaseURL(), "https://", "admin server must use HTTPS")
})

t.Run("admin_path_not_accessible_from_public_port", func(t *testing.T) {
t.Parallel()

client := newTLSHTTPClient(t)
url := fmt.Sprintf(
"%s%s%s",
server.PublicBaseURL(),
cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath,
cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
)

req := newContractRequest(t, url)

resp, err := client.Do(req)
require.NoError(t, err, "GET request to admin-path-on-public-port should not fail (expects 404)")

defer func() { require.NoError(t, resp.Body.Close()) }()

assert.Equal(t, http.StatusNotFound, resp.StatusCode, "admin livez path must return 404 on public server (server isolation)")
})
}