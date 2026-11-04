// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	http "net/http"
	"net/url"
	"strconv"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const publicHealthCheckRetryInterval = 25 * time.Millisecond

// CheckResponsive performs a local, certificate-verified request through the public
// listener. It does not call readyz or any remote dependency.
func (s *PublicServerBase) CheckResponsive(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	s.mu.RLock()
	port := s.actualPort
	tlsMaterial := s.tlsMaterial
	s.mu.RUnlock()

	if port <= 0 {
		return fmt.Errorf("public listener is not bound")
	}

	if tlsMaterial == nil || tlsMaterial.RootCAPool == nil || tlsMaterial.Config == nil {
		return fmt.Errorf("public listener TLS trust is not configured")
	}

	serverName, err := publicProbeServerName(tlsMaterial.Config)
	if err != nil {
		return err
	}

	portString := strconv.Itoa(port)
	probeURL := (&url.URL{
		Scheme: cryptoutilSharedMagic.ProtocolHTTPS,
		Host:   net.JoinHostPort(serverName, portString),
		Path:   cryptoutilSharedMagic.IdentityE2EHealthEndpoint,
	}).String()
	dialer := &net.Dialer{}

	transport := &http.Transport{
		Proxy: nil,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			RootCAs:    tlsMaterial.RootCAPool,
			ServerName: serverName,
		},
		DialContext: func(dialContext context.Context, network, _ string) (net.Conn, error) {
			// The target name is certificate-verified; the socket itself always loops back.
			return dialer.DialContext(dialContext, network, net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, portString))
		},
		DisableKeepAlives: true,
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{Transport: transport}

	var lastErr error

	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("local public listener observation: %w (last check: %w)", err, lastErr)
		}

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
		if err != nil {
			return fmt.Errorf("create local public health request: %w", err)
		}

		response, requestErr := client.Do(request)
		if requestErr == nil {
			if _, err := io.Copy(io.Discard, response.Body); err != nil {
				lastErr = fmt.Errorf("read local public health response: %w", err)
			} else if response.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("local public listener returned HTTP %d", response.StatusCode)
			} else {
				_ = response.Body.Close()

				return nil
			}

			_ = response.Body.Close()
		} else {
			lastErr = fmt.Errorf("local public listener request failed: %w", requestErr)
		}

		timer := time.NewTimer(publicHealthCheckRetryInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
		case <-timer.C:
		}
	}
}

func publicProbeServerName(config *tls.Config) (string, error) {
	if len(config.Certificates) == 0 || len(config.Certificates[0].Certificate) == 0 {
		return "", fmt.Errorf("public listener certificate is unavailable")
	}

	leaf := config.Certificates[0].Leaf
	if leaf == nil {
		var err error

		leaf, err = x509.ParseCertificate(config.Certificates[0].Certificate[0])
		if err != nil {
			return "", fmt.Errorf("parse public listener certificate: %w", err)
		}
	}

	if len(leaf.DNSNames) > 0 && leaf.DNSNames[0] != "" {
		return leaf.DNSNames[0], nil
	}

	if len(leaf.IPAddresses) > 0 {
		return leaf.IPAddresses[0].String(), nil
	}

	return "", fmt.Errorf("public listener certificate has no DNS or IP subject alternative name")
}
