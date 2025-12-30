// Copyright (c) 2025 Justin Cranford

package im

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"cryptoutil/internal/learn/repository"
	"cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestIM_ServerSubcommand_Startup(t *testing.T) {
	t.Skip("Server subcommand requires signal handling (SIGINT/SIGTERM) which blocks test execution")

	// Create unique in-memory database.
	uniqueDSN := fmt.Sprintf("file:%s?mode=memory&cache=shared", googleUuid.NewString())

	sqlDB, err := sql.Open("sqlite", uniqueDSN)
	require.NoError(t, err)

	defer sqlDB.Close() //nolint:errcheck // Test cleanup, error not critical

	// Apply migrations using embedded migration files.
	err = repository.ApplyMigrations(sqlDB, repository.DatabaseTypeSQLite)
	require.NoError(t, err)

	// Note: IM([]string{"server"}) creates its own database connection internally.
	// We just need to ensure migrations are available for the test.

	// Start server in background goroutine.
	serverDone := make(chan struct{})

	go func() {
		defer close(serverDone)

		// Capture output to prevent server logs polluting test output.
		output := captureOutput(t, func() {
			// Note: This will start server and block until shutdown signal.
			// We'll cancel the context to trigger shutdown.
			exitCode := IM([]string{"server"})
			require.Equal(t, 0, exitCode, "Server should exit cleanly")
		})

		// Verify server startup messages.
		require.Contains(t, output, "Starting learn-im service")
		require.Contains(t, output, fmt.Sprintf("Public Server: https://127.0.0.1:%d", magic.DefaultPublicPortLearnIM))
		require.Contains(t, output, fmt.Sprintf("Admin Server:  https://127.0.0.1:%d", magic.DefaultPrivatePortLearnIM))
		require.Empty(t, output, "Server should not output errors")
	}()

	// Wait a bit for server to start.
	time.Sleep(500 * time.Millisecond)

	// Verify server is accessible via health check.
	captureOutput(t, func() {
		exitCode := IM([]string{
			"health",
			"--url", fmt.Sprintf("https://127.0.0.1:%d/service/api/v1", magic.DefaultPublicPortLearnIM),
		})
		require.Equal(t, 0, exitCode, "Health check should succeed")
	})

	// Cancel goroutine shutdown simulation (server needs SIGINT/SIGTERM).
	// For this test, we'll just verify server startup, not shutdown behavior.

	// Wait for server to shutdown.
	select {
	case <-serverDone:
		// Success - server shutdown cleanly.
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shutdown within timeout")
	}
}

func TestIM_ClientSubcommand_NotImplemented(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"client"})
		require.Equal(t, 1, exitCode, "Client subcommand should return error (not implemented)")
	})

	require.Empty(t, output)
	require.Contains(t, output, "Client subcommand not yet implemented")
	require.Contains(t, output, "This will provide CLI tools for interacting with the IM service")
}

func TestIM_ClientSubcommand_Help(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"client", "--help"})
		require.Equal(t, 0, exitCode, "Help should return success")
	})

	require.Empty(t, output)
	require.Contains(t, output, "Usage: learn im client [options]")
	require.Contains(t, output, "Run client operations for instant messaging service")
}

func TestIM_InitSubcommand_NotImplemented(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"init"})
		require.Equal(t, 1, exitCode, "Init subcommand should return error (not implemented)")
	})

	require.Empty(t, output)
	require.Contains(t, output, "Init subcommand not yet implemented")
	require.Contains(t, output, "This will initialize database schema and configuration")
}

func TestIM_InitSubcommand_Help(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"init", "--help"})
		require.Equal(t, 0, exitCode, "Help should return success")
	})

	require.Empty(t, output)
	require.Contains(t, output, "Usage: learn im init [options]")
	require.Contains(t, output, "Initialize database schema and configuration")
}
