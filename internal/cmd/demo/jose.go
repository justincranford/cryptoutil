// Copyright (c) 2025 Justin Cranford

package demo

// TODO: This file needs updating to use new jose-ja server.
// The old cryptoutilJoseServer.NewServer() call needs to be replaced with:
// cryptoutilAppsJoseJaServer.NewFromConfig(ctx, cryptoutilAppsJoseJaServerConfig.NewTestConfig(...))
// See internal/apps/jose/ja/server/testmain_test.go for reference.

import (
	"context"
	"fmt"
	"os"
)

// runJOSEDemo is temporarily stubbed out during Phase 7 migration.
// Full implementation will be restored in a future phase.
func runJOSEDemo(ctx context.Context, config *Config) int {
	fmt.Fprintln(os.Stderr, "JOSE demo temporarily disabled during jose-ja migration (Phase 7)")
	fmt.Fprintln(os.Stderr, "Will be restored after internal/jose deletion (Task 7.4)")
	return 0 // Don't fail the demo command
}
