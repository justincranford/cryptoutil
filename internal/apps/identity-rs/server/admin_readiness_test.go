// Copyright (c) 2025-2026 Justin Cranford.

package server

import (
	"context"
	"crypto/x509"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceTestHelpTls "cryptoutil/internal/apps-framework/service/test_help_tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func waitForAdminPort(t *testing.T, server *AdminServer, timeout time.Duration) int {
	t.Helper()

	var port int

	err := cryptoutilAppsFrameworkServiceTestHelpTls.WaitForHTTPSHealthEndpoint(context.Background(), func() (string, *x509.CertPool) {
		port = server.ActualPort()
		if port <= 0 {
			return "", nil
		}

		return cryptoutilSharedMagic.URLPrefixLocalhostHTTPS + strconv.Itoa(port) + cryptoutilSharedMagic.DefaultAdminLivezEndpoint, server.TLSRootCAPool()
	}, timeout)
	require.NoError(t, err)

	return port
}
