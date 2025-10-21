// Package magic provides commonly used magic numbers and values as named constants.
// This file contains network-related constants.
package magic

// Network ports.
const (
	// PortHTTPS - Standard HTTPS port.
	PortHTTPS = 443
	// PortDefaultBrowserAPI - Default browser/server API port.
	PortDefaultBrowserAPI = 8080
	// PortCryptoutilPostgres1 - Port for cryptoutil postgres instance 1.
	PortCryptoutilPostgres1 = 8081
	// PortCryptoutilPostgres2 - Port for cryptoutil postgres instance 2.
	PortCryptoutilPostgres2 = 8082
	// PortDefaultAdminAPI - Default admin API port.
	PortDefaultAdminAPI = 9090
	// PortGrafana - Default Grafana port.
	PortGrafana = 3000
	// PortOtelCollector - Default OpenTelemetry collector port.
	PortOtelCollector = 8888
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
	RateLimitBrowserIP = 1000
	// RateLimitServiceIP - Default service IP rate limit.
	RateLimitServiceIP = 500
)
