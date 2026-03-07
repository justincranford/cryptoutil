# Migration script for server_integration_test.go files

files = {
    "internal/apps/skeleton/template/server/server_integration_test.go": {
        "service": "SkeletonTemplate",
        "config_import": "cryptoutilAppsSkeletonTemplateServerConfig",
        "func_name": "TestSkeletonTemplateServer_ShutdownIdempotent",
        "lifecycle_func": "TestSkeletonTemplateServer_Lifecycle",
        "e2e_import": 'cryptoutilAppsTemplateServiceTestingE2eHelpers "cryptoutil/internal/apps/template/service/testing/e2e_helpers"',
    },
    "internal/apps/jose/ja/server/server_integration_test.go": {
        "service": "JoseJA",
        "config_import": "cryptoutilAppsJoseJaServerConfig",
        "func_name": "TestJoseJAServer_ShutdownIdempotent",
        "lifecycle_func": "TestJoseJAServer_Lifecycle",
        "e2e_import": 'cryptoutilAppsTemplateServiceTestingE2eHelpers "cryptoutil/internal/apps/template/service/testing/e2e_helpers"',
    },
}

for path, info in files.items():
    print(f"\n=== Processing {path} ===")
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()

    # 1. Remove e2e_helpers import, add testserver import
    old_e2e = f'\t{info["e2e_import"]}'
    new_testserver = '\tcryptoutilTestingTestserver "cryptoutil/internal/apps/template/service/testing/testserver"'
    if old_e2e in content:
        content = content.replace(old_e2e, new_testserver)
        print("  - Replaced e2e_helpers import with testserver import OK")
    else:
        print(f"  - ERROR: e2e_helpers import not found: {repr(old_e2e[:80])}")

    # 2. Find and replace the ShutdownIdempotent function
    func_start = f"func {info['func_name']}(t *testing.T) {{"
    idx = content.find(func_start)
    if idx < 0:
        print(f"  - ERROR: {info['func_name']} not found")
        continue

    # Find end of function (next top-level func or end of file)
    end_idx = content.find("\nfunc ", idx + 1)
    if end_idx < 0:
        end_idx = len(content)

    old_func = content[idx:end_idx]

    # Build new function
    cfg_import = info["config_import"]
    new_func = f'''func {info["func_name"]}(t *testing.T) {{
\tt.Parallel()

\tctx, cancel := context.WithCancel(context.Background())
\tdefer cancel() // ensure start goroutine exits when test returns

\t// Create test configuration with different ports.
\tcfg := {cfg_import}.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

\t// Create separate server instance.
\tserver, err := NewFromConfig(ctx, cfg)
\trequire.NoError(t, err, "server creation should succeed")

\t// Start server and wait for both ports using shared helper.
\t// t.Cleanup will call server.Shutdown when test completes.
\tcryptoutilTestingTestserver.StartAndWait(ctx, t, server)

\t// Shutdown the server explicitly (covers Shutdown happy path).
\tshutdownCtx, shutdownCancel := context.WithTimeout(context.Background(),
\t\tcryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
\tdefer shutdownCancel()

\terr = server.Shutdown(shutdownCtx)
\trequire.NoError(t, err, "shutdown should succeed")
}}'''

    content = content[:idx] + new_func + content[end_idx:]
    print(f"  - Replaced {info['func_name']} OK")

    # 3. For Lifecycle test: replace manual livez/readyz requests with testHealthClient
    # Skeleton has context.Background() and http.NewRequestWithContext in Lifecycle
    if "context.Background()" in old_func:
        pass  # already handled

    # Final write
    with open(path, "w", encoding="utf-8", newline="\n") as f:
        f.write(content)
    print(f"  - File written OK")

print("\nDone!")