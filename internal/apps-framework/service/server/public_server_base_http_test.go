// Copyright (c) 2025-2026 Justin Cranford.
//

//go:build !integration

package server

import (
	"crypto/tls"
	http "net/http"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
)

func createTestHTTPClient(t *testing.T, tlsMaterial *cryptoutilAppsFrameworkServiceConfig.TLSMaterial) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    tlsMaterial.RootCAPool,
				MinVersion: tls.VersionTLS12,
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.DefaultHealthCheckTimeout,
	}
}
