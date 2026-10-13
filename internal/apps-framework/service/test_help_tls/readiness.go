// Copyright (c) 2025-2026 Justin Cranford.

package test_help_tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	http "net/http"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
)

type (
	httpsHealthEndpointFn func() (string, *x509.CertPool)
	httpsHealthProbeFn    func(context.Context, *http.Client, string) (int, error)
)

// WaitForHTTPSHealthEndpoint waits until the endpoint responds with HTTP 200 using the supplied trust pool.
func WaitForHTTPSHealthEndpoint(ctx context.Context, endpointFn func() (string, *x509.CertPool), timeout time.Duration) error {
	return waitForHTTPSHealthEndpoint(ctx, endpointFn, timeout, cryptoutilSharedMagic.TestServerWaitTickerInterval, requestHTTPSHealthEndpoint)
}

func waitForHTTPSHealthEndpoint(
	ctx context.Context,
	endpointFn httpsHealthEndpointFn,
	timeout time.Duration,
	interval time.Duration,
	probe httpsHealthProbeFn,
) error {
	switch {
	case ctx == nil:
		return fmt.Errorf("context cannot be nil")
	case endpointFn == nil:
		return fmt.Errorf("HTTPS health endpoint source cannot be nil")
	case probe == nil:
		return fmt.Errorf("HTTPS health probe cannot be nil")
	case timeout <= 0:
		return fmt.Errorf("HTTPS health timeout must be positive")
	case interval <= 0:
		return fmt.Errorf("HTTPS health interval must be positive")
	}

	readinessCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var lastProbeErr error

	pollErr := cryptoutilSharedUtilPoll.Until(readinessCtx, timeout, interval, func(probeCtx context.Context) (bool, error) {
		endpointURL, rootCAs := endpointFn()
		if endpointURL == "" {
			lastProbeErr = fmt.Errorf("HTTPS health endpoint URL is empty")

			return false, nil
		}

		if rootCAs == nil {
			lastProbeErr = fmt.Errorf("HTTPS health root CA pool is not ready")

			return false, nil
		}

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS13,
					RootCAs:    rootCAs,
				},
				DisableKeepAlives: true,
			},
			Timeout: timeout,
		}
		defer client.CloseIdleConnections()

		statusCode, err := probe(probeCtx, client, endpointURL)
		if err != nil {
			lastProbeErr = fmt.Errorf("HTTPS health request failed: %w", err)

			return false, nil
		}

		if statusCode != http.StatusOK {
			lastProbeErr = fmt.Errorf("HTTPS health endpoint returned HTTP %d", statusCode)

			return false, nil
		}

		lastProbeErr = nil

		return true, nil
	})
	if pollErr != nil {
		if lastProbeErr != nil {
			return fmt.Errorf("HTTPS health readiness failed: %w; last probe: %w", pollErr, lastProbeErr)
		}

		return fmt.Errorf("HTTPS health readiness failed: %w", pollErr)
	}

	return nil
}

func requestHTTPSHealthEndpoint(ctx context.Context, client *http.Client, endpointURL string) (int, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL, nil)
	if err != nil {
		return 0, fmt.Errorf("create HTTPS health request: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("send HTTPS health request: %w", err)
	}

	if err := response.Body.Close(); err != nil {
		return 0, fmt.Errorf("close HTTPS health response: %w", err)
	}

	return response.StatusCode, nil
}
