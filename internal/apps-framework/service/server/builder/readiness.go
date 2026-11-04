// Copyright (c) 2025-2026 Justin Cranford.
package builder

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	http "net/http"
	"net/url"
	"os"
	"time"

	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsFrameworkServiceServerReadiness "cryptoutil/internal/apps-framework/service/server/readiness"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const collectorHealthResponseMaxBytes = 4096

func newReadinessMonitor(
	db *gorm.DB,
	publicServer *cryptoutilAppsFrameworkServiceServer.PublicServerBase,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	additionalChecks []cryptoutilAppsFrameworkServiceServerReadiness.Check,
) (*cryptoutilAppsFrameworkServiceServerReadiness.Monitor, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	if publicServer == nil {
		return nil, fmt.Errorf("public server is required")
	}

	if settings == nil {
		return nil, fmt.Errorf("service settings are required")
	}

	checks := []cryptoutilAppsFrameworkServiceServerReadiness.Check{
		{
			Name:       cryptoutilSharedMagic.RealmStorageTypeDatabase,
			ReasonCode: "database_unavailable",
			Run:        databaseReadinessCheck(db),
		},
		{
			Name:       "public_listener",
			ReasonCode: "public_listener_unresponsive",
			Run:        publicServer.CheckResponsive,
		},
	}

	if settings.OTLPEnabled {
		collectorCheck, err := newCollectorHealthCheck(
			settings.OTLPHealthEndpoint,
			settings.OTLPTLSCAFile,
			settings.OTLPTLSCertFile,
			settings.OTLPTLSKeyFile,
		)
		if err != nil {
			return nil, fmt.Errorf("configure collector health check: %w", err)
		}

		checks = append(checks, cryptoutilAppsFrameworkServiceServerReadiness.Check{
			Name:       "collector_transport",
			ReasonCode: "collector_unavailable",
			Run:        collectorCheck,
		})
	}

	checks = append(checks, additionalChecks...)

	monitor, err := cryptoutilAppsFrameworkServiceServerReadiness.New(cryptoutilAppsFrameworkServiceServerReadiness.Config{
		Checks: checks,
	})
	if err != nil {
		return nil, fmt.Errorf("create readiness monitor: %w", err)
	}

	return monitor, nil
}

func newCollectorHealthCheck(endpoint, caPath, certPath, keyPath string) (cryptoutilAppsFrameworkServiceServerReadiness.CheckFunc, error) {
	return newCollectorHealthCheckWithTimeout(endpoint, caPath, certPath, keyPath, 2*time.Second)
}

// The fixture transport budget is independent of the production observation
// deadline; production keeps its two-second client and 500ms monitor bounds.
func newCollectorHealthCheckWithTimeout(endpoint, caPath, certPath, keyPath string, clientTimeout time.Duration) (cryptoutilAppsFrameworkServiceServerReadiness.CheckFunc, error) {
	healthURL, err := url.Parse(endpoint)
	if err != nil || healthURL.Scheme != cryptoutilSharedMagic.ProtocolHTTPS || healthURL.Host == "" ||
		healthURL.Path == "" || healthURL.Path == "/" || healthURL.User != nil ||
		healthURL.RawQuery != "" || healthURL.ForceQuery || healthURL.Fragment != "" {
		return nil, fmt.Errorf("collector health URL must be an HTTPS URL with host and path and no credentials, query, or fragment")
	}

	caPEM, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("read collector health CA certificate")
	}

	rootCAPool := x509.NewCertPool()
	if !rootCAPool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("parse collector health CA certificate")
	}

	clientCertificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("load collector health client certificate")
	}

	transport := &http.Transport{
		Proxy: nil,
		TLSClientConfig: &tls.Config{
			MinVersion:   tls.VersionTLS13,
			RootCAs:      rootCAPool,
			Certificates: []tls.Certificate{clientCertificate},
		},
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   clientTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return func(ctx context.Context) error {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL.String(), nil)
		if err != nil {
			return fmt.Errorf("create collector health request")
		}

		response, err := client.Do(request)
		if err != nil {
			return fmt.Errorf("collector health endpoint request failed")
		}
		defer func() { _ = response.Body.Close() }()

		if _, err := io.Copy(io.Discard, io.LimitReader(response.Body, collectorHealthResponseMaxBytes)); err != nil {
			return fmt.Errorf("read collector health response")
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("collector health endpoint returned HTTP %d", response.StatusCode)
		}

		return nil
	}, nil
}

func databaseReadinessCheck(db *gorm.DB) cryptoutilAppsFrameworkServiceServerReadiness.CheckFunc {
	return func(ctx context.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("get application database pool: %w", err)
		}

		if err := sqlDB.PingContext(ctx); err != nil {
			return fmt.Errorf("ping application database: %w", err)
		}

		rows, err := sqlDB.QueryContext(ctx, "SELECT 1 FROM barrier_root_keys LIMIT 0")
		if err != nil {
			return fmt.Errorf("verify framework schema: %w", err)
		}

		rows.Next()

		if err := rows.Err(); err != nil {
			_ = rows.Close()

			return fmt.Errorf("read framework schema observation: %w", err)
		}

		if err := rows.Close(); err != nil {
			return fmt.Errorf("close framework schema observation: %w", err)
		}

		return nil
	}
}
