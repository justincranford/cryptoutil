// Copyright (c) 2025 Justin Cranford
//

package contract

import (
"fmt"
http "net/http"
"strings"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"

cryptoutilTestingHealthclient "cryptoutil/internal/apps/template/service/testing/healthclient"
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// RunResponseFormatContracts verifies response format consistency across all endpoints.
// Tests 7 contracts:
//  1. livez response Content-Type is application/json
//  2. readyz response Content-Type is application/json
//  3. browser health response Content-Type is application/json
//  4. service health response Content-Type is application/json
//  5. non-existent service path returns HTTP 404
//  6. non-existent browser path returns HTTP 404
//  7. non-existent admin path returns HTTP 404
func RunResponseFormatContracts(t *testing.T, server ServiceServer) {
t.Helper()

hc := cryptoutilTestingHealthclient.NewHealthClient(server.PublicBaseURL(), server.AdminBaseURL())

contentTypeTests := []struct {
name  string
fetch func() (*http.Response, error)
}{
{"livez_has_json_content_type", hc.Livez},
{"readyz_has_json_content_type", hc.Readyz},
{"browser_health_has_json_content_type", hc.BrowserHealth},
{"service_health_has_json_content_type", hc.ServiceHealth},
}

for _, tc := range contentTypeTests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

resp, err := tc.fetch()
require.NoError(t, err, "health endpoint request should succeed")

defer func() { require.NoError(t, resp.Body.Close()) }()

contentType := resp.Header.Get("Content-Type")
assert.True(t, strings.Contains(contentType, "application/json"),
"Content-Type should contain application/json, got: %s", contentType)
})
}

client := newTLSHTTPClient(t)

nonExistentTests := []struct {
name string
url  string
}{
{
name: "nonexistent_service_path_returns_404",
url: fmt.Sprintf("%s%s/nonexistent-contract-test-path",
server.PublicBaseURL(),
cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
),
},
{
name: "nonexistent_browser_path_returns_404",
url: fmt.Sprintf("%s%s/nonexistent-contract-test-path",
server.PublicBaseURL(),
cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
),
},
{
name: "nonexistent_admin_path_returns_404",
url: fmt.Sprintf("%s%s/nonexistent-contract-test-path",
server.AdminBaseURL(),
cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath,
),
},
}

for _, tc := range nonExistentTests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

req := newContractRequest(t, tc.url)

resp, err := client.Do(req)
require.NoError(t, err, "GET request to non-existent path should not fail (expects 404)")

defer func() { require.NoError(t, resp.Body.Close()) }()

assert.Equal(t, http.StatusNotFound, resp.StatusCode, "non-existent path must return 404")
})
}
}