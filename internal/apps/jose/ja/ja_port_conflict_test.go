// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package ja

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	cryptoutilSharedTestutil "cryptoutil/internal/shared/testutil"
)

// TestJA_ServerStartPortConflict verifies the server error path through errChan when
// the public port is already occupied and srv.Start(ctx) fails to bind.
// Sequential: uses viper global state via ParseWithFlagSet.
func TestJA_ServerStartPortConflict(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	// Occupy a TCP port so the server's public listener fails to bind.
	var lc net.ListenConfig

	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	defer func() { require.NoError(t, ln.Close()) }()

	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok, "expected *net.TCPAddr")

	occupiedPort := tcpAddr.Port

	var stdout, stderr cryptoutilSharedTestutil.SafeBuffer

	exitCode := jaServerStart(
		[]string{
			"--profile=test",
			fmt.Sprintf("--bind-public-port=%d", occupiedPort),
			"--bind-private-port=0",
		},
		&stdout, &stderr,
	)

	require.Equal(t, 1, exitCode, "server should return exit code 1 when port is occupied")
	require.Contains(t, stderr.String(), "Server error", "stderr should contain server error message")
}
