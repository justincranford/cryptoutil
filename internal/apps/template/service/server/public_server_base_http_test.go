// Copyright (c) 2025 Justin Cranford
//

//go:build !integration

package server

import (
	"crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	http "net/http"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
)

func createTestHTTPClient(t *testing.T, tlsMaterial *cryptoutilAppsTemplateServiceConfig.TLSMaterial) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    tlsMaterial.RootCAPool,
				MinVersion: tls.VersionTLS12,
			},
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}
}
