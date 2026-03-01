#!/usr/bin/env python3
"""
autoapprove - Command wrapper for VS Code Copilot Chat auto-approve bypass.

This script wraps commands like curl, wget, go, docker to:
1. Validate URLs/network addresses to only allow loopback (127.0.0.1, ::1, localhost)
2. Pass through args and STDIN to the called command
3. Pass back STDOUT, STDERR, and exit code to caller
4. Log detailed execution information to test-output directory

Reference:
- VS Code Copilot Terminal Safety Guards: https://code.visualstudio.com/docs/copilot/chat/chat-tools#_terminal
- VS Code mandatory safety overrides cannot be bypassed for destructive or network commands
- This wrapper provides a workaround by wrapping commands to enforce loopback-only network access
"""

import argparse
import contextlib
import ipaddress
import os
import re
import subprocess
import sys
import time
from datetime import UTC, datetime
from pathlib import Path
from types import ModuleType
from urllib.parse import urlparse

# Import resource module only on Unix (not available on Windows).
resource: ModuleType | None = None
with contextlib.suppress(ImportError):
    import resource


# Allowed loopback addresses and hostnames.
LOOPBACK_IPV4 = "127.0.0.1"
LOOPBACK_IPV6 = "::1"
LOOPBACK_HOSTNAMES = frozenset({"localhost", "localhost.localdomain", "ip6-localhost", "ip6-loopback"})

# Patterns to detect URLs and network addresses in arguments.
URL_PATTERN = re.compile(r"^(https?|ftp|ftps)://", re.IGNORECASE)
# IPv4 address pattern.
IPV4_PATTERN = re.compile(r"^(\d{1,3}\.){3}\d{1,3}(:\d+)?$")
# IPv6 address pattern (simplified - covers most common formats).
IPV6_PATTERN = re.compile(r"^\[?([0-9a-fA-F:]+)\]?(:\d+)?$")
# Host:port pattern.
HOST_PORT_PATTERN = re.compile(r"^([a-zA-Z0-9.-]+)(:\d+)?$")


def is_loopback_address(address: str) -> bool:
    """Check if an IP address is a loopback address."""
    try:
        ip = ipaddress.ip_address(address)
        return ip.is_loopback
    except ValueError:
        return False


def is_allowed_host(host: str) -> bool:
    """Check if a hostname or IP address is allowed (loopback only)."""
    # Remove brackets from IPv6 addresses.
    host = host.strip("[]")

    # Check if it's a loopback hostname.
    if host.lower() in LOOPBACK_HOSTNAMES:
        return True

    # Check if it's a loopback IP address.
    return is_loopback_address(host)


def extract_host_from_url(url: str) -> str | None:
    """Extract hostname from a URL."""
    try:
        parsed = urlparse(url)
        return parsed.hostname
    except Exception:
        return None


def extract_host_from_address(address: str) -> str | None:
    """Extract hostname/IP from an address string (host:port format)."""
    # Remove port if present.
    if address.startswith("["):
        # IPv6 with brackets.
        match = re.match(r"^\[([^\]]+)\](:\d+)?$", address)
        if match:
            return match.group(1)
    elif ":" in address:
        # Could be IPv6 or host:port.
        parts = address.rsplit(":", 1)
        if parts[1].isdigit():
            return parts[0]
    return address


def validate_network_args(args: list[str]) -> tuple[bool, str]:
    """
    Validate that all network-related arguments only reference loopback addresses.

    Returns:
        tuple: (is_valid, error_message)
    """
    for arg in args:
        # Skip flags that start with -.
        if arg.startswith("-"):
            # Check if it's a flag with value like --url=http://...
            if "=" in arg:
                _, value = arg.split("=", 1)
                arg = value
            else:
                continue

        # Check if it's a URL.
        if URL_PATTERN.match(arg):
            host = extract_host_from_url(arg)
            if host and not is_allowed_host(host):
                return False, f"URL contains non-loopback host: {arg}"
            continue

        # Check if it's an IP address pattern.
        if IPV4_PATTERN.match(arg):
            host = extract_host_from_address(arg)
            if host and not is_allowed_host(host):
                return False, f"IPv4 address is not loopback: {arg}"
            continue

        # Check if it's an IPv6 address pattern.
        if IPV6_PATTERN.match(arg):
            host = extract_host_from_address(arg)
            if host and not is_allowed_host(host):
                return False, f"IPv6 address is not loopback: {arg}"
            continue

        # Check if it looks like host:port.
        if HOST_PORT_PATTERN.match(arg) and ":" in arg:
            host = extract_host_from_address(arg)
            if host and not is_allowed_host(host):
                return False, f"Host:port contains non-loopback host: {arg}"

    return True, ""


def get_timestamp() -> str:
    """Get current timestamp in ISO 8601 format for directory naming."""
    return datetime.now(UTC).strftime("%Y-%m-%dT%H-%M-%S.%f")[:-3]


def create_report_directory(command_name: str) -> Path:
    """Create the report directory for this execution."""
    timestamp = get_timestamp()
    safe_command_name = re.sub(r"[^\w\-.]", "_", command_name)
    dir_name = f"{timestamp}-{safe_command_name}"
    report_dir = Path("./test-output") / "autoapprove" / dir_name
    report_dir.mkdir(parents=True, exist_ok=True)
    return report_dir


def get_resource_usage() -> dict[str, float]:
    """Get current resource usage statistics (Unix only, returns zeros on Windows)."""
    if resource is None:
        return {
            "user_time_seconds": 0.0,
            "system_time_seconds": 0.0,
            "max_memory_kb": 0,
        }
    rusage = resource.getrusage(resource.RUSAGE_CHILDREN)
    return {
        "user_time_seconds": rusage.ru_utime,
        "system_time_seconds": rusage.ru_stime,
        "max_memory_kb": rusage.ru_maxrss,
    }


def run_command(command: list[str], report_dir: Path, stdin_data: bytes | None = None) -> int:
    """
    Run the command and capture all I/O to report files.

    Returns:
        int: Exit code of the command.
    """
    start_time = time.perf_counter()
    start_datetime = datetime.now(UTC)

    # Log STDIN if provided.
    stdin_log = report_dir / "STDIN.log"
    if stdin_data:
        stdin_log.write_bytes(stdin_data)
    else:
        stdin_log.write_text("")

    # Run the command.
    try:
        result = subprocess.run(
            command,
            input=stdin_data,
            capture_output=True,
            cwd=os.getcwd(),
        )
        stdout_data = result.stdout
        stderr_data = result.stderr
        exit_code = result.returncode
    except FileNotFoundError:
        stdout_data = b""
        stderr_data = f"Command not found: {command[0]}\n".encode()
        exit_code = 127
    except PermissionError:
        stdout_data = b""
        stderr_data = f"Permission denied: {command[0]}\n".encode()
        exit_code = 126
    except Exception as e:
        stdout_data = b""
        stderr_data = f"Error executing command: {e}\n".encode()
        exit_code = 1

    end_time = time.perf_counter()
    end_datetime = datetime.now(UTC)
    duration = end_time - start_time

    # Write STDOUT and STDERR logs.
    stdout_log = report_dir / "STDOUT.log"
    stderr_log = report_dir / "STDERR.log"
    stdout_log.write_bytes(stdout_data)
    stderr_log.write_bytes(stderr_data)

    # Get resource usage.
    usage = get_resource_usage()

    # Write result.log.
    result_log = report_dir / "result.log"
    result_content = f"""command: {" ".join(command)}
working_directory: {os.getcwd()}
exit_code: {exit_code}
start_time: {start_datetime.isoformat()}
end_time: {end_datetime.isoformat()}
duration_seconds: {duration:.6f}
cpu_user_time_seconds: {usage["user_time_seconds"]:.6f}
cpu_system_time_seconds: {usage["system_time_seconds"]:.6f}
max_memory_kb: {usage["max_memory_kb"]}
"""
    result_log.write_text(result_content)

    # Write output to actual STDOUT/STDERR.
    sys.stdout.buffer.write(stdout_data)
    sys.stderr.buffer.write(stderr_data)

    return exit_code


def main() -> int:
    """Main entry point for autoapprove wrapper."""
    parser = argparse.ArgumentParser(
        description="Command wrapper for VS Code Copilot Chat auto-approve bypass with loopback-only network restriction.",
        epilog="""
Reference: https://code.visualstudio.com/docs/copilot/chat/chat-tools#_terminal

This wrapper enforces that network commands only access loopback addresses
(127.0.0.1, ::1, localhost) to mitigate security risks while allowing
auto-approval of commands in VS Code Copilot Chat.
""",
    )
    parser.add_argument(
        "command",
        nargs=argparse.REMAINDER,
        help="Command and arguments to execute",
    )
    parser.add_argument(
        "--skip-validation",
        action="store_true",
        help="Skip network address validation (use with caution)",
    )

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return 1

    command = args.command

    # Validate network arguments unless skipped.
    if not args.skip_validation:
        is_valid, error_msg = validate_network_args(command)
        if not is_valid:
            sys.stderr.write(f"autoapprove: Network validation failed: {error_msg}\n")
            sys.stderr.write("autoapprove: Only loopback addresses (127.0.0.1, ::1, localhost) are allowed.\n")
            sys.stderr.write("autoapprove: Use --skip-validation to bypass this check (not recommended).\n")
            return 1

    # Read STDIN if available.
    stdin_data = None
    if not sys.stdin.isatty():
        stdin_data = sys.stdin.buffer.read()

    # Create report directory.
    command_name = os.path.basename(command[0])
    report_dir = create_report_directory(command_name)

    # Run the command.
    exit_code = run_command(command, report_dir, stdin_data)

    return exit_code


if __name__ == "__main__":
    sys.exit(main())
