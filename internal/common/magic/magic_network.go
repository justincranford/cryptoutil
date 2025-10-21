// Package magic provides commonly used magic numbers and values as named constants.
// This file contains network-related constants.
package magic

// Network ports.
const (
	// PortHTTPS - Standard HTTPS port.
	PortHTTPS = 443
	// PortDefaultBrowserAPI - Default browser/server API port.
	PortDefaultBrowserAPI = 8080
	// PortDefaultAdminAPI - Default admin API port.
	PortDefaultAdminAPI = 9090
)

// Rate limiting defaults.
const (
	// RateLimitBrowserIP - Default browser IP rate limit.
	RateLimitBrowserIP = 1000
	// RateLimitServiceIP - Default service IP rate limit.
	RateLimitServiceIP = 500
)
