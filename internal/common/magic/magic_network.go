// Package magic provides commonly used magic numbers and values as named constants.
// This file contains network-related constants.
package magic

// Network ports.
const (
	// PortHTTPS - Standard HTTPS port.
	PortHTTPS uint16 = 443

	// PortGrafana - Default Grafana port.
	PortGrafana uint16 = 3000

	// PortOtelCollectorMetrics - Default OpenTelemetry collector internal metrics port (Prometheus).
	PortOtelCollectorMetrics uint16 = 8888
	// PortOtelCollectorHealth - Default OpenTelemetry collector health port.
	PortOtelCollectorHealth uint16 = 8889

	// PortOtelGRPC - Default OpenTelemetry OTLP gRPC port.
	PortOtelGRPC uint16 = 4317
	// PortOtelHTTP - Default OpenTelemetry OTLP HTTP port.
	PortOtelHTTP uint16 = 4318

	// PortPostgres - Default PostgreSQL port.
	PortPostgres uint16 = 5432

	// PortDefaultBrowserAPI - Default browser/server API port.
	PortDefaultBrowserAPI uint16 = 8080
	// PortCryptoutilPostgres1 - Port for cryptoutil postgres instance 1.
	PortCryptoutilPostgres1 uint16 = 8081
	// PortCryptoutilPostgres2 - Port for cryptoutil postgres instance 2.
	PortCryptoutilPostgres2 uint16 = 8082
	// PortDefaultAdminAPI - Default admin API port.
	PortDefaultAdminAPI uint16 = 9090
)

// Network URLs and prefixes.
const (
	// URLPrefixLocalhostHTTPS - HTTPS URL prefix for localhost.
	URLPrefixLocalhostHTTPS = "https://127.0.0.1:"
	// URLPrefixLocalhostHTTP - HTTP URL prefix for localhost.
	URLPrefixLocalhostHTTP = "http://127.0.0.1:"
)

// Rate limiting defaults.
const (
	// RateLimitBrowserIP - Default browser IP rate limit.
	RateLimitBrowserIP uint16 = 1000
	// RateLimitServiceIP - Default service IP rate limit.
	RateLimitServiceIP uint16 = 500
	// RateLimitBrowserIPDefault - Default browser IP rate limit (100 requests/second).
	RateLimitBrowserIPDefault uint16 = 100
	// RateLimitServiceIPDefault - Default service IP rate limit (25 requests/second).
	RateLimitServiceIPDefault uint16 = 25
	// RateLimitMaxIP - Maximum allowed IP rate limit.
	RateLimitMaxIP uint16 = 10000
)
