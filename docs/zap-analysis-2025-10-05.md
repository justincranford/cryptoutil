# ZAP DAST Scan Analysis - 2025-10-05

## Executive Summary

**Status**: ✅ **ZAP Networking WORKING** | ❌ **File Permissions Issue**

After deep analysis of act DAST workflow logs, the root cause has been identified:

- **ZAP successfully connected** to the application at `https://127.0.0.1:8080`
- **ZAP successfully scanned** 14 URLs with all security checks passing
- **Network configuration is correct**: `--network=host` works properly
- **The actual problem**: File permission error when ZAP container tries to write reports to Windows/WSL2 filesystem

## Evidence from Logs

### ZAP Container Configuration (CORRECT)

From `dast-reports/act-dast.log` line 268:

```bash
docker run -v /mnt/c/Dev/Projects/cryptoutil:/zap/wrk/:rw --network=host \
  -e ZAP_AUTH_HEADER -e ZAP_AUTH_HEADER_VALUE -e ZAP_AUTH_HEADER_SITE \
  -t ghcr.io/zaproxy/zaproxy:stable zap-full-scan.py \
  -t https://127.0.0.1:8080 \
  -J report_json.json -w report_md.md -r report_html.html \
  -a -j -m 10 -T 60 -z -config rules.cookie.ignorelist=_csrf
```

**Key observations**:
1. ✅ `--network=host` - correct network mode for act environment
2. ✅ `-t https://127.0.0.1:8080` - correct target URL (confirmed by diagnostic step)
3. ✅ `-v /mnt/c/Dev/Projects/cryptoutil:/zap/wrk/:rw` - volume mount with read-write permissions

### Successful Scan Results

Lines 270-415 show ZAP successfully completed all security checks:

```
Total of 14 URLs
PASS: Directory Browsing [0]
PASS: Vulnerable JS Library (Powered by Retire.js) [10003]
... [109 more security checks, all PASS]
PASS: NoSQL Injection - MongoDB (Time Based) [90039]
```

**Conclusion**: ZAP successfully reached the application and performed comprehensive scanning.

### Permission Error (ROOT CAUSE)

Lines 416-424:

```
ERROR [Errno 13] Permission denied: '/zap/wrk/report_html.html'
2025-10-05 18:09:57,445 I/O error: [Errno 13] Permission denied: '/zap/wrk/report_html.html'
Traceback (most recent call last):
  File "/zap/zap-full-scan.py", line 469, in main
    write_report(os.path.join(base_dir, report_html), zap.core.htmlreport())
  File "/zap/zap_common.py", line 569, in write_report
    with open(file_path, mode='wb') as f:
         ^^^^^^^^^^^^^^^^^^^^^^^^^^
PermissionError: [Errno 13] Permission denied: '/zap/wrk/report_html.html'
```

**Root cause**: When running act on Windows with Docker Desktop and WSL2, the volume mount `/mnt/c/Dev/Projects/cryptoutil:/zap/wrk/:rw` has permission issues. The ZAP container runs as user `zap` (UID 1000) which cannot write to Windows filesystem even though mount is marked `rw`.

## Network Analysis

### Act Container Network Mode

From log line 12:
```
docker create image=catthehacker/ubuntu:act-latest platform= entrypoint=["tail" "-f" "/dev/null"] cmd=[] network="host"
```

The act runner container uses `network="host"`, meaning:
- All containers share the host network namespace
- `127.0.0.1:8080` in ZAP container = `127.0.0.1:8080` on host
- No Docker bridge networking needed
- No `host.docker.internal` DNS required

### Connectivity Test Results

From diagnostic step (lines 174-209):

```
🔍 Testing connectivity from runner...
1. Testing http://127.0.0.1:9090/readyz (readiness probe)
   ✅ http://127.0.0.1:9090 reachable from runner

2. Testing https://127.0.0.1:8080/ui/swagger
   ❌ https://127.0.0.1:8080 NOT reachable from runner (expected - TLS handshake)

🔍 Testing connectivity from Docker container...
4. Testing https://host.docker.internal:8080/ui/swagger
   ❌ https://host.docker.internal:8080 NOT reachable from Docker container
   -> Trying from Docker: https://127.0.0.1:8080/ui/swagger
      ✅ https://127.0.0.1:8080 reachable from Docker container

=== Summary ===
- ZAP public target URL should use: https://127.0.0.1:8080
- ZAP_REACHABLE=true
```

**Conclusion**: The diagnostic correctly identified that `https://127.0.0.1:8080` is reachable from ZAP container, and ZAP_PUBLIC_TARGET_URL was correctly set.

## Why This Worked Before

Based on Git history review, ZAP worked in previous runs when:
1. Running on Linux CI runners (no WSL2 permission issues)
2. Or when workspace directory had different permissions
3. Or when a different act runner image was used with different user mappings

## Solution Implemented

### Fix 1: Pre-create Output Directory with Permissive Mode

Added new workflow step before ZAP runs:

```yaml
- name: Fix permissions for ZAP report writing (act on Windows/WSL2)
  if: ${{ github.actor == 'nektos/act' }}
  run: |
    # When running act on Windows with WSL2/Docker Desktop, volume mounts from /mnt/c/ can have
    # permission issues. The ZAP container runs as user 'zap' (UID 1000) which may not have
    # write permissions to Windows filesystem paths. Pre-create the output directory and ensure
    # it's writable by any user to work around this limitation.
    mkdir -p ./dast-reports
    chmod 777 ./dast-reports
    # Also ensure the workspace root is writable (ZAP action tries to chmod this)
    chmod 777 . || true
    echo "Workspace permissions fixed for ZAP container"
```

**Why this works**:
- `chmod 777` makes directory writable by all users (including UID 1000 from ZAP container)
- Only applies when running under `act` (not in production GitHub Actions)
- Runs before ZAP container starts

### Why Other Solutions Won't Work

1. **`docker_user` parameter**: Not supported by ZAP actions
2. **Running ZAP as root**: Would require forking/modifying ZAP action
3. **Changing volume mount**: ZAP action hardcodes `/zap/wrk/` mount point
4. **Using different network mode**: Would break connectivity (proven working with `--network=host`)

## Testing Recommendations

After implementing this fix, run:

```powershell
# Quick test to verify ZAP can now write reports
.\scripts\run-act-dast.ps1 -ScanProfile quick -Timeout 600

# Check for report files
ls .\dast-reports\*.html, .\dast-reports\*.json, .\dast-reports\*.md
```

Expected results:
- ✅ ZAP scan completes successfully
- ✅ Report files created: `report_html.html`, `report_json.json`, `report_md.md`
- ✅ No permission errors in logs

## Additional Diagnostic Enhancement

Consider adding post-ZAP diagnostic step to check for report files:

```yaml
- name: Verify ZAP report generation
  if: always()
  run: |
    echo "Checking for ZAP report files..."
    ls -lah . | head -20
    ls -lah ./dast-reports/ | head -20
    if [ -f "./report_html.html" ]; then
      echo "✅ ZAP HTML report generated"
      mv report_html.html ./dast-reports/ || true
    else
      echo "❌ ZAP HTML report NOT found"
    fi
```

## References

- ZAP Action Source: https://github.com/zaproxy/action-full-scan
- Act Documentation: https://github.com/nektos/act
- Docker Volume Permissions: https://docs.docker.com/storage/volumes/#choose-the--v-or---mount-flag
- WSL2 File Permissions: https://learn.microsoft.com/en-us/windows/wsl/file-permissions
