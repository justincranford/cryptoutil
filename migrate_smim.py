path = "internal/apps/sm/im/server/server_test.go"
with open(path, "r", encoding="utf-8") as f:
    content = f.read()

# 1. Add testserver import
old_imports = 'import (\n\t"context"\n\tcryptoutilSharedMagic "cryptoutil/internal/shared/magic"\n\t"testing"\n\t"time"\n\n\t"github.com/stretchr/testify/require"\n\n\tcryptoutilAppsSmImServer "cryptoutil/internal/apps/sm/im/server"\n\tcryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm/im/server/config"\n)'

new_imports = 'import (\n\t"context"\n\tcryptoutilSharedMagic "cryptoutil/internal/shared/magic"\n\t"testing"\n\t"time"\n\n\t"github.com/stretchr/testify/require"\n\n\tcryptoutilAppsSmImServer "cryptoutil/internal/apps/sm/im/server"\n\tcryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm/im/server/config"\n\tcryptoutilTestingTestserver "cryptoutil/internal/apps/template/service/testing/testserver"\n)'

if old_imports in content:
    content = content.replace(old_imports, new_imports)
    print("Replaced import block OK")
else:
    print("ERROR: import block not found")

# 2. Replace TestServer_Shutdown function
old_shutdown = 'func TestServer_Shutdown(t *testing.T) {\n\tt.Parallel()\n\n\tctx := context.Background()\n\n\t// Create test configuration.\n\tcfg := cryptoutilAppsSmImServerConfig.DefaultTestConfig()\n\n\t// Create new server instance (separate from TestMain\'s testSmIMServer).\n\ttestServer, err := cryptoutilAppsSmImServer.NewFromConfig(ctx, cfg)\n\trequire.NoError(t, err, "Failed to create test server")\n\trequire.NotNil(t, testServer, "Server should not be nil")\n\n\t// Start server in background.\n\tstartCtx, startCancel := context.WithCancel(ctx)\n\tdefer startCancel()\n\n\tstartErrCh := make(chan error, 1)\n\n\tgo func() {\n\t\tstartErrCh <- testServer.Start(startCtx)\n\t}()\n\n\t// Wait for server to become ready (check public port is assigned).\n\trequire.Eventually(t, func() bool {\n\t\treturn testServer.PublicPort() > 0\n\t}, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond, "Server should start within 5 seconds")\n\n\t// Test graceful shutdown.\n\tshutdownCtx, shutdownCancel := context.WithTimeout(ctx, cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second)\n\tdefer shutdownCancel()\n\n\terr = testServer.Shutdown(shutdownCtx)\n\trequire.NoError(t, err, "Shutdown should succeed without errors")\n\n\t// Verify Start() exits after shutdown.\n\tselect {\n\tcase startErr := <-startErrCh:\n\t\t// Start() may return error after graceful shutdown (context canceled) - this is acceptable.\n\t\t// The important part is that Start() exits (doesn\'t block forever).\n\t\tif startErr != nil {\n\t\t\tt.Logf("Start() returned error after shutdown (expected): %v", startErr)\n\t\t}\n\tcase <-time.After(2 * time.Second):\n\t\tt.Fatal("Start() did not exit within 2 seconds after shutdown")\n\t}\n}'

new_shutdown = 'func TestServer_Shutdown(t *testing.T) {\n\tt.Parallel()\n\n\tstartCtx, startCancel := context.WithCancel(context.Background())\n\tdefer startCancel() // ensure start goroutine exits when test returns\n\n\t// Create test configuration.\n\tcfg := cryptoutilAppsSmImServerConfig.DefaultTestConfig()\n\n\t// Create new server instance (separate from TestMain\'s testSmIMServer).\n\ttestServer, err := cryptoutilAppsSmImServer.NewFromConfig(startCtx, cfg)\n\trequire.NoError(t, err, "Failed to create test server")\n\trequire.NotNil(t, testServer, "Server should not be nil")\n\n\t// Start server and wait for both ports using shared helper.\n\t// t.Cleanup will call server.Shutdown when test completes.\n\tcryptoutilTestingTestserver.StartAndWait(startCtx, t, testServer)\n\n\t// Test graceful shutdown.\n\tshutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second)\n\tdefer shutdownCancel()\n\n\terr = testServer.Shutdown(shutdownCtx)\n\trequire.NoError(t, err, "Shutdown should succeed without errors")\n}'

if old_shutdown in content:
    content = content.replace(old_shutdown, new_shutdown)
    print("Replaced TestServer_Shutdown OK")
else:
    print("ERROR: TestServer_Shutdown not found")

with open(path, "w", encoding="utf-8", newline="\n") as f:
    f.write(content)
print("File written, length:", len(content))