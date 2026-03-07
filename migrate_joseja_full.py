path = "internal/apps/jose/ja/server/server_integration_test.go"

with open(path, "r", encoding="utf-8") as f:
    content = f.read()

# 1. Replace import block: add testserver
old_imports = 'import (\n\t"context"\n\tcryptoutilSharedMagic "cryptoutil/internal/shared/magic"\n\t"fmt"\n\thttp "net/http"\n\t"testing"\n\t"time"\n\n\t"github.com/stretchr/testify/require"\n\n\tcryptoutilAppsJoseJaServerConfig "cryptoutil/internal/apps/jose/ja/server/config"\n)'
new_imports = 'import (\n\t"context"\n\tcryptoutilSharedMagic "cryptoutil/internal/shared/magic"\n\t"fmt"\n\thttp "net/http"\n\t"testing"\n\t"time"\n\n\t"github.com/stretchr/testify/require"\n\n\tcryptoutilAppsJoseJaServerConfig "cryptoutil/internal/apps/jose/ja/server/config"\n\tcryptoutilTestingTestserver "cryptoutil/internal/apps/template/service/testing/testserver"\n)'

if old_imports in content:
    content = content.replace(old_imports, new_imports)
    print("Replaced import block OK")
else:
    print("ERROR: import block not found")
    print("Expected:", repr(old_imports[:80]))

# 2. Replace Lifecycle test
old_lifecycle = 'func TestJoseJAServer_Lifecycle(t *testing.T) {\n\tt.Parallel()\n\t// Verify admin endpoints accessible.\n\trequire.NotEmpty(t, testAdminBaseURL, "admin base URL should not be empty")\n\n\t// Test /admin/api/v1/livez endpoint.\n\tlivezReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/admin/api/v1/livez", testAdminBaseURL), nil)\n\trequire.NoError(t, err, "livez request creation should succeed")\n\tlivezResp, err := testHTTPClient.Do(livezReq)\n\trequire.NoError(t, err, "livez request should succeed")\n\trequire.Equal(t, http.StatusOK, livezResp.StatusCode, "livez should return 200 OK")\n\trequire.NoError(t, livezResp.Body.Close())\n\n\t// Test /admin/api/v1/readyz endpoint.\n\treadyzReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/admin/api/v1/readyz", testAdminBaseURL), nil)\n\trequire.NoError(t, err, "readyz request creation should succeed")\n\treadyzResp, err := testHTTPClient.Do(readyzReq)\n\trequire.NoError(t, err, "readyz request should succeed")\n\trequire.Equal(t, http.StatusOK, readyzResp.StatusCode, "readyz should return 200 OK")\n\trequire.NoError(t, readyzResp.Body.Close())\n\n\t// Verify public endpoints accessible.\n\trequire.NotEmpty(t, testPublicBaseURL, "public base URL should not be empty")\n\t// Note: Cannot test actual routes without authentication setup\n\t// This integration test validates server lifecycle only\n}\n'

new_lifecycle = 'func TestJoseJAServer_Lifecycle(t *testing.T) {\n\tt.Parallel()\n\t// Verify admin endpoints accessible.\n\trequire.NotEmpty(t, testAdminBaseURL, "admin base URL should not be empty")\n\n\t// Test /admin/api/v1/livez endpoint.\n\tlivezResp, err := testHealthClient.Livez()\n\trequire.NoError(t, err, "livez request should succeed")\n\trequire.Equal(t, http.StatusOK, livezResp.StatusCode, "livez should return 200 OK")\n\trequire.NoError(t, livezResp.Body.Close())\n\n\t// Test /admin/api/v1/readyz endpoint.\n\treadyzResp, err := testHealthClient.Readyz()\n\trequire.NoError(t, err, "readyz request should succeed")\n\trequire.Equal(t, http.StatusOK, readyzResp.StatusCode, "readyz should return 200 OK")\n\trequire.NoError(t, readyzResp.Body.Close())\n\n\t// Verify public endpoints accessible.\n\trequire.NotEmpty(t, testPublicBaseURL, "public base URL should not be empty")\n\t// Note: Cannot test actual routes without authentication setup\n\t// This integration test validates server lifecycle only\n}\n'

if old_lifecycle in content:
    content = content.replace(old_lifecycle, new_lifecycle)
    print("Replaced Lifecycle test OK")
else:
    print("ERROR: Lifecycle test not found")

with open(path, "w", encoding="utf-8", newline="\n") as f:
    f.write(content)
print("File written OK")
print("Final length:", len(content))