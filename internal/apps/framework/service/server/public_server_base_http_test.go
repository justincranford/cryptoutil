// Copyright (c) 2025 Justin Cranford
//

//go:build !integration

package server

import (
	"crypto/tls"
	http "net/http"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
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
		Timeout: cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Second,
	}
}
