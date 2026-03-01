#!/usr/bin/env python3
"""Tests for autoapprove command wrapper."""

import os
import sys
import tempfile
from pathlib import Path
from unittest.mock import patch

# Add the parent directory to path for imports.
sys.path.insert(0, str(Path(__file__).parent.parent))

from autoapprove import (
    extract_host_from_address,
    extract_host_from_url,
    is_allowed_host,
    is_loopback_address,
    validate_network_args,
)


class TestIsLoopbackAddress:
    """Tests for is_loopback_address function."""

    def test_ipv4_loopback(self) -> None:
        """Test that 127.0.0.1 is recognized as loopback."""
        assert is_loopback_address("127.0.0.1") is True

    def test_ipv4_loopback_range(self) -> None:
        """Test that 127.x.x.x addresses are recognized as loopback."""
        assert is_loopback_address("127.0.0.2") is True
        assert is_loopback_address("127.255.255.255") is True

    def test_ipv6_loopback(self) -> None:
        """Test that ::1 is recognized as loopback."""
        assert is_loopback_address("::1") is True

    def test_non_loopback_ipv4(self) -> None:
        """Test that non-loopback IPv4 addresses are rejected."""
        assert is_loopback_address("192.168.1.1") is False
        assert is_loopback_address("10.0.0.1") is False
        assert is_loopback_address("8.8.8.8") is False

    def test_non_loopback_ipv6(self) -> None:
        """Test that non-loopback IPv6 addresses are rejected."""
        assert is_loopback_address("2001:db8::1") is False
        assert is_loopback_address("fe80::1") is False

    def test_invalid_address(self) -> None:
        """Test that invalid addresses return False."""
        assert is_loopback_address("not-an-ip") is False
        assert is_loopback_address("") is False


class TestIsAllowedHost:
    """Tests for is_allowed_host function."""

    def test_localhost(self) -> None:
        """Test that localhost is allowed."""
        assert is_allowed_host("localhost") is True
        assert is_allowed_host("LOCALHOST") is True  # Case insensitive.

    def test_localhost_variants(self) -> None:
        """Test that localhost variants are allowed."""
        assert is_allowed_host("localhost.localdomain") is True
        assert is_allowed_host("ip6-localhost") is True
        assert is_allowed_host("ip6-loopback") is True

    def test_loopback_ips(self) -> None:
        """Test that loopback IPs are allowed."""
        assert is_allowed_host("127.0.0.1") is True
        assert is_allowed_host("::1") is True
        assert is_allowed_host("[::1]") is True

    def test_external_hosts(self) -> None:
        """Test that external hosts are rejected."""
        assert is_allowed_host("example.com") is False
        assert is_allowed_host("google.com") is False
        assert is_allowed_host("192.168.1.1") is False


class TestExtractHostFromUrl:
    """Tests for extract_host_from_url function."""

    def test_http_url(self) -> None:
        """Test extraction from HTTP URL."""
        assert extract_host_from_url("http://localhost:8080/api") == "localhost"
        assert extract_host_from_url("http://127.0.0.1:8080/api") == "127.0.0.1"

    def test_https_url(self) -> None:
        """Test extraction from HTTPS URL."""
        assert extract_host_from_url("https://localhost:443/api") == "localhost"

    def test_url_without_port(self) -> None:
        """Test extraction from URL without port."""
        assert extract_host_from_url("http://localhost/api") == "localhost"

    def test_ipv6_url(self) -> None:
        """Test extraction from IPv6 URL."""
        assert extract_host_from_url("http://[::1]:8080/api") == "::1"

    def test_invalid_url(self) -> None:
        """Test extraction from invalid URL."""
        result = extract_host_from_url("not-a-url")
        assert result is None or result == ""


class TestExtractHostFromAddress:
    """Tests for extract_host_from_address function."""

    def test_host_port(self) -> None:
        """Test extraction from host:port format."""
        assert extract_host_from_address("localhost:8080") == "localhost"
        assert extract_host_from_address("127.0.0.1:8080") == "127.0.0.1"

    def test_ipv6_with_brackets(self) -> None:
        """Test extraction from IPv6 with brackets."""
        assert extract_host_from_address("[::1]:8080") == "::1"
        assert extract_host_from_address("[::1]") == "::1"

    def test_plain_host(self) -> None:
        """Test extraction from plain host."""
        assert extract_host_from_address("localhost") == "localhost"


class TestValidateNetworkArgs:
    """Tests for validate_network_args function."""

    def test_allowed_localhost_url(self) -> None:
        """Test that localhost URLs are allowed."""
        is_valid, error = validate_network_args(["curl", "http://localhost:8080/api"])
        assert is_valid is True
        assert error == ""

    def test_allowed_loopback_url(self) -> None:
        """Test that loopback URLs are allowed."""
        is_valid, error = validate_network_args(["curl", "http://127.0.0.1:8080/api"])
        assert is_valid is True
        assert error == ""

    def test_allowed_ipv6_loopback_url(self) -> None:
        """Test that IPv6 loopback URLs are allowed."""
        is_valid, error = validate_network_args(["curl", "http://[::1]:8080/api"])
        assert is_valid is True
        assert error == ""

    def test_blocked_external_url(self) -> None:
        """Test that external URLs are blocked."""
        is_valid, error = validate_network_args(["curl", "http://example.com/api"])
        assert is_valid is False
        assert "non-loopback" in error.lower()

    def test_blocked_external_ip(self) -> None:
        """Test that external IPs are blocked."""
        is_valid, error = validate_network_args(["curl", "http://192.168.1.1:8080/api"])
        assert is_valid is False
        assert "non-loopback" in error.lower()

    def test_flag_with_url_value(self) -> None:
        """Test that flag=value format is validated."""
        is_valid, error = validate_network_args(["curl", "--url=http://example.com"])
        assert is_valid is False
        assert "non-loopback" in error.lower()

    def test_non_network_args(self) -> None:
        """Test that non-network args are allowed."""
        is_valid, error = validate_network_args(["go", "test", "./...", "-v"])
        assert is_valid is True
        assert error == ""

    def test_docker_commands(self) -> None:
        """Test that docker commands without network args are allowed."""
        is_valid, error = validate_network_args(["docker", "ps", "-a"])
        assert is_valid is True
        assert error == ""

    def test_mixed_args_with_loopback(self) -> None:
        """Test mixed args with loopback address."""
        is_valid, error = validate_network_args(["curl", "-s", "-X", "POST", "http://localhost:8080/api", "-d", '{"key":"value"}'])
        assert is_valid is True
        assert error == ""


class TestIntegration:
    """Integration tests for autoapprove script."""

    def test_simple_command_execution(self) -> None:
        """Test that a simple command can be executed."""
        with tempfile.TemporaryDirectory() as tmpdir:
            original_cwd = os.getcwd()
            try:
                os.chdir(tmpdir)
                # Import main after changing directory.
                from autoapprove import create_report_directory, run_command

                report_dir = create_report_directory("echo")
                exit_code = run_command(["echo", "hello"], report_dir)
                assert exit_code == 0
                assert (report_dir / "STDOUT.log").exists()
                assert (report_dir / "STDERR.log").exists()
                assert (report_dir / "result.log").exists()
                stdout_content = (report_dir / "STDOUT.log").read_text()
                assert "hello" in stdout_content
            finally:
                os.chdir(original_cwd)

    def test_command_not_found(self) -> None:
        """Test handling of non-existent command."""
        with tempfile.TemporaryDirectory() as tmpdir:
            original_cwd = os.getcwd()
            try:
                os.chdir(tmpdir)
                from autoapprove import create_report_directory, run_command

                report_dir = create_report_directory("nonexistent")
                exit_code = run_command(["nonexistent_command_xyz"], report_dir)
                assert exit_code == 127
                stderr_content = (report_dir / "STDERR.log").read_text()
                assert "not found" in stderr_content.lower()
            finally:
                os.chdir(original_cwd)

    def test_result_log_content(self) -> None:
        """Test that result.log contains expected fields."""
        with tempfile.TemporaryDirectory() as tmpdir:
            original_cwd = os.getcwd()
            try:
                os.chdir(tmpdir)
                from autoapprove import create_report_directory, run_command

                report_dir = create_report_directory("echo")
                run_command(["echo", "test"], report_dir)
                result_content = (report_dir / "result.log").read_text()
                assert "command:" in result_content
                assert "working_directory:" in result_content
                assert "exit_code:" in result_content
                assert "start_time:" in result_content
                assert "end_time:" in result_content
                assert "duration_seconds:" in result_content
                assert "cpu_user_time_seconds:" in result_content
                assert "cpu_system_time_seconds:" in result_content
                assert "max_memory_kb:" in result_content
            finally:
                os.chdir(original_cwd)


class TestMainFunction:
    """Tests for main function."""

    def test_main_no_args(self) -> None:
        """Test main function with no arguments."""
        with patch("sys.argv", ["autoapprove"]):
            from autoapprove import main

            exit_code = main()
            assert exit_code == 1

    def test_main_blocked_url(self) -> None:
        """Test main function with blocked URL."""
        with patch("sys.argv", ["autoapprove", "curl", "http://example.com"]):
            from autoapprove import main

            exit_code = main()
            assert exit_code == 1

    def test_main_skip_validation(self) -> None:
        """Test main function with skip-validation flag."""
        with tempfile.TemporaryDirectory() as tmpdir:
            original_cwd = os.getcwd()
            try:
                os.chdir(tmpdir)
                # Mock stdin to return that it's a tty (so no stdin read).
                mock_stdin = type("MockStdin", (), {"isatty": lambda self: True})()
                with patch("sys.argv", ["autoapprove", "--skip-validation", "echo", "test"]), patch("sys.stdin", mock_stdin):
                    from autoapprove import main

                    exit_code = main()
                    assert exit_code == 0
            finally:
                os.chdir(original_cwd)

    def test_main_with_stdin(self) -> None:
        """Test main function with stdin input."""
        with tempfile.TemporaryDirectory() as tmpdir:
            original_cwd = os.getcwd()
            try:
                os.chdir(tmpdir)
                # Create a mock stdin that returns bytes.
                mock_stdin = type("MockStdin", (), {"isatty": lambda self: False, "buffer": type("Buffer", (), {"read": lambda self: b"test input"})()})()
                with patch("sys.argv", ["autoapprove", "cat"]), patch("sys.stdin", mock_stdin):
                    from autoapprove import main

                    exit_code = main()
                    assert exit_code == 0
            finally:
                os.chdir(original_cwd)


class TestValidateNetworkArgsExtended:
    """Extended tests for validate_network_args function."""

    def test_ipv4_address_with_port(self) -> None:
        """Test validation of IPv4 address with port."""
        is_valid, error = validate_network_args(["curl", "127.0.0.1:8080"])
        assert is_valid is True

    def test_external_ipv4_with_port(self) -> None:
        """Test that external IPv4 with port is blocked."""
        is_valid, error = validate_network_args(["curl", "192.168.1.1:8080"])
        assert is_valid is False

    def test_ipv6_address_with_port(self) -> None:
        """Test validation of IPv6 address with port."""
        is_valid, error = validate_network_args(["curl", "[::1]:8080"])
        assert is_valid is True

    def test_external_ipv6_url(self) -> None:
        """Test that external IPv6 URL is blocked."""
        is_valid, error = validate_network_args(["curl", "http://[2001:db8::1]:8080/"])
        assert is_valid is False


class TestResourceUsage:
    """Tests for resource usage tracking."""

    def test_get_resource_usage(self) -> None:
        """Test that resource usage returns expected keys."""
        from autoapprove import get_resource_usage

        usage = get_resource_usage()
        assert "user_time_seconds" in usage
        assert "system_time_seconds" in usage
        assert "max_memory_kb" in usage


class TestTimestamp:
    """Tests for timestamp generation."""

    def test_get_timestamp_format(self) -> None:
        """Test that timestamp is in expected format."""
        from autoapprove import get_timestamp

        timestamp = get_timestamp()
        # Should match pattern: YYYY-MM-DDTHH-MM-SS.mmm
        assert len(timestamp) == 23
        assert timestamp[4] == "-"
        assert timestamp[7] == "-"
        assert timestamp[10] == "T"
        assert timestamp[13] == "-"
        assert timestamp[16] == "-"
        assert timestamp[19] == "."


class TestRunCommandEdgeCases:
    """Edge case tests for run_command function."""

    def test_command_with_stdin(self) -> None:
        """Test command execution with stdin data."""
        with tempfile.TemporaryDirectory() as tmpdir:
            original_cwd = os.getcwd()
            try:
                os.chdir(tmpdir)
                from autoapprove import create_report_directory, run_command

                report_dir = create_report_directory("cat")
                stdin_data = b"hello from stdin"
                exit_code = run_command(["cat"], report_dir, stdin_data)
                assert exit_code == 0
                stdout_content = (report_dir / "STDOUT.log").read_bytes()
                assert stdout_content == stdin_data
                stdin_log_content = (report_dir / "STDIN.log").read_bytes()
                assert stdin_log_content == stdin_data
            finally:
                os.chdir(original_cwd)

    def test_command_stderr_output(self) -> None:
        """Test that stderr is captured correctly."""
        with tempfile.TemporaryDirectory() as tmpdir:
            original_cwd = os.getcwd()
            try:
                os.chdir(tmpdir)
                from autoapprove import create_report_directory, run_command

                report_dir = create_report_directory("bash")
                # Use bash to write to stderr.
                exit_code = run_command(["bash", "-c", "echo error >&2"], report_dir)
                assert exit_code == 0
                stderr_content = (report_dir / "STDERR.log").read_text()
                assert "error" in stderr_content
            finally:
                os.chdir(original_cwd)
