// Copyright (c) 2025-2026 Justin Cranford.
//go:build e2e

package e2e_test

import (
	"context"
	"crypto/tls"
	http "net/http"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestE2E_TLSChainValidation verifies that the public HTTPS endpoint presents a certificate
// chain signed by the pki-init CA. Happy path uses sharedHTTPClientWithCA (CA-validated);
// sad path uses a client with wrong/no CA and expects a TLS error.
func TestE2E_TLSChainValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		publicURL string
		client    *http.Client
		wantOK    bool
	}{
		{
			name:      "sqlite-1-correct-CA",
			publicURL: sqlitePublicURL,
			client:    sharedHTTPClientWithCA,
			wantOK:    true,
		},
		{
			name:      "postgresql-1-correct-CA",
			publicURL: postgres1PublicURL,
			client:    sharedHTTPClientWithCA,
			wantOK:    true,
		},
		{
			name:      "postgresql-2-correct-CA",
			publicURL: postgres2PublicURL,
			client:    sharedHTTPClientWithCA,
			wantOK:    true,
		},
		{
			name:      "sqlite-1-wrong-CA",
			publicURL: sqlitePublicURL,
			client: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig:   &tls.Config{MinVersion: tls.VersionTLS13},
					DisableKeepAlives: true,
				},
				Timeout: cryptoutilSharedMagic.E2EHTTPClientTimeout,
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, tt.publicURL+cryptoutilSharedMagic.IME2EHealthEndpoint, nil)
			require.NoError(t, err)

			resp, doErr := tt.client.Do(req)

			if tt.wantOK {
				require.NoError(t, doErr, "CA-validated TLS should succeed for %s", tt.name)
				require.NoError(t, resp.Body.Close())
				require.Equal(t, http.StatusOK, resp.StatusCode, "%s should return 200 OK", tt.name)
			} else {
				require.Error(t, doErr, "TLS with wrong CA should fail for %s", tt.name)
				require.ErrorContains(t, doErr, "certificate", "error should be a TLS certificate error for %s", tt.name)
			}
		})
	}
}
