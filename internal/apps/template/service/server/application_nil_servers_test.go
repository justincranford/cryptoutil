// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

//go:build !integration

package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestApplication_PortsBeforeInit tests port/URL getters when servers not initialized.
// Targets application.go:205-207, 220-222, 236-238, 251-253 (nil server checks).
func TestApplication_PortsBeforeInit(t *testing.T) {
	t.Parallel()

	// Create application without initializing servers
	app := &Application{
		publicServer: nil,
		adminServer:  nil,
	}

	// All getters should return zero values when servers not initialized
	require.Equal(t, 0, app.PublicPort(), "PublicPort should return 0 when server nil")
	require.Equal(t, 0, app.AdminPort(), "AdminPort should return 0 when server nil")
	require.Equal(t, "", app.PublicBaseURL(), "PublicBaseURL should return empty string when server nil")
	require.Equal(t, "", app.AdminBaseURL(), "AdminBaseURL should return empty string when server nil")
}
