# autoapprove

Command wrapper for VS Code Copilot Chat auto-approve bypass with loopback-only network restriction.

## Purpose

VS Code Copilot Chat has [mandatory safety overrides](https://code.visualstudio.com/docs/copilot/chat/chat-tools#_terminal) that cannot be bypassed for certain commands deemed potentially unsafe. This includes network commands that may connect to external hosts.

This wrapper provides a workaround by:
1. Enforcing that all network-related arguments only reference loopback addresses (127.0.0.1, ::1, localhost)
2. Logging detailed execution information for audit trails
3. Passing through all I/O transparently to the wrapped command

## Reference Documentation

- [VS Code Copilot Chat Tools - Terminal](https://code.visualstudio.com/docs/copilot/chat/chat-tools#_terminal)
- [VS Code Copilot Settings Reference](https://code.visualstudio.com/docs/copilot/reference/copilot-settings)
- [VS Code Agent Mode Blog Post](https://code.visualstudio.com/blogs/2025/04/07/agentMode)
- [GitHub Issue #265775 - Auto-approve network commands](https://github.com/microsoft/vscode/issues/265775)
- [GitHub Issue #266651 - Terminal command safety](https://github.com/microsoft/vscode/issues/266651)

## Installation

```bash
# From the repository root
cd scripts/autoapprove
pip install -e .
```

## Usage

```bash
# Wrap a curl command (only allows loopback)
autoapprove curl -s http://127.0.0.1:8080/api/health

# Wrap a wget command
autoapprove wget -q -O - http://localhost:8080/api/status

# Wrap a go test command
autoapprove go test ./...

# Wrap a docker command
autoapprove docker ps

# Skip validation (not recommended, use with caution)
autoapprove --skip-validation curl http://example.com
```

## Network Validation

The wrapper validates all arguments that appear to be URLs or network addresses:

**Allowed:**
- `127.0.0.1` (IPv4 loopback)
- `::1` (IPv6 loopback)
- `localhost`
- `localhost.localdomain`
- `ip6-localhost`
- `ip6-loopback`

**Blocked:**
- Any other IP address
- Any external hostname

## Output Logging

Each command execution creates a timestamped directory under `./test-reports/`:

```
./test-reports/autoapprove.2025-01-15T10-30-45.123.curl/
├── STDIN.log      # Input sent to command
├── STDOUT.log     # Standard output from command
├── STDERR.log     # Standard error from command
└── result.log     # Execution metadata (timing, exit code, resources)
```

### result.log Format

```
command: curl -s http://127.0.0.1:8080/api/health
working_directory: /home/user/project
exit_code: 0
start_time: 2025-01-15T10:30:45.123456+00:00
end_time: 2025-01-15T10:30:45.234567+00:00
duration_seconds: 0.111111
cpu_user_time_seconds: 0.005000
cpu_system_time_seconds: 0.002000
max_memory_kb: 12288
```

## VS Code Configuration

Add `autoapprove` to your VS Code settings for auto-approve:

```json
{
  "chat.tools.terminal.autoApprove": {
    "autoapprove ": true
  }
}
```

## License

AGPL-3.0 - See repository LICENSE file.
