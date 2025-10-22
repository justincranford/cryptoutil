// Package magic provides commonly used magic numbers and values as named constants.
// This file contains network-related constants.
package magic

import "time"

// Network ports.
const (
	// PortHTTPS - Standard HTTPS port.
	PortHTTPS uint16 = 443

	// DefaultPortGrafana - Default Grafana port.
	DefaultPortGrafana uint16 = 3000

	// DefaultPortOtelCollectorMetrics - Default OpenTelemetry collector internal metrics port (Prometheus).
	DefaultPortOtelCollectorMetrics uint16 = 8888
	// PortOtelCollectorHealth - Default OpenTelemetry collector health port.
	PortOtelCollectorHealth uint16 = 8889

	// PortOtelGRPC - Default OpenTelemetry OTLP gRPC port.
	PortOtelGRPC uint16 = 4317
	// PortOtelHTTP - Default OpenTelemetry OTLP HTTP port.
	PortOtelHTTP uint16 = 4318

	// PortPostgres - Default PostgreSQL port.
	PortPostgres uint16 = 5432

	// DefaultPortDefaultBrowserAPI - Default browser/server API port.
	DefaultPortDefaultBrowserAPI uint16 = 8080
	// DefaultPortCryptoutilPostgres1 - Port for cryptoutil postgres instance 1.
	DefaultPortCryptoutilPostgres1 uint16 = 8081
	// DefaultPortCryptoutilPostgres2 - Port for cryptoutil postgres instance 2.
	DefaultPortCryptoutilPostgres2 uint16 = 8082
	// DefaultPortAdminAPI - Default admin API port.
	DefaultPortAdminAPI uint16 = 9090
)

// Network URLs and prefixes.
const (
	// URLPrefixLocalhostHTTPS - HTTPS URL prefix for localhost.
	URLPrefixLocalhostHTTPS = "https://127.0.0.1:"
	// URLPrefixLocalhostHTTP - HTTP URL prefix for localhost.
	URLPrefixLocalhostHTTP = "http://127.0.0.1:"
)

const (
	// DefaultServerMaxRequestBodySize - Maximum request body size for test server (1MB).
	DefaultServerMaxRequestBodySize = 1 << 20
)

// Rate limiting defaults.
const (
	// DefaultRateLimitBrowserIP - Default browser IP rate limit (100 requests/second).
	DefaultRateLimitBrowserIP uint16 = 100
	// DefaultRateLimitServiceIP - Default service IP rate limit (25 requests/second).
	DefaultRateLimitServiceIP uint16 = 25
	// MaxRateLimitIP - Maximum allowed IP rate limit.
	MaxRateLimitIP uint16 = 10000
)

const (
	// DefaultPublicBrowserAPIContextPath - Default public browser API context path.
	DefaultPublicBrowserAPIContextPath = "/browser/api/v1"
	// DefaultPublicServiceAPIContextPath - Default public service API context path.
	DefaultPublicServiceAPIContextPath = "/service/api/v1"
	// StringLivezPath - Livez endpoint path.
	StringLivezPath = "/livez"
	// StringReadyzPath - Readyz endpoint path.
	StringReadyzPath = "/readyz"
	// StringShutdownPath - Shutdown endpoint path.
	StringShutdownPath = "/shutdown"
	// StringProtocolHTTPS - HTTPS protocol string.
	StringProtocolHTTPS = "https"
	// StringProtocolHTTP - HTTP protocol string.
	StringProtocolHTTP = "http"
	// StringLocalhost - Localhost hostname.
	StringLocalhost = "localhost"
	// StringIPv4Loopback - IPv4 loopback address.
	StringIPv4Loopback = "127.0.0.1"
	// StringIPv6Loopback - IPv6 loopback address.
	StringIPv6Loopback = "::1"
	// StringIPv4MappedIPv6Loopback - IPv4 mapped IPv6 loopback address.
	StringIPv4MappedIPv6Loopback = "::ffff:127.0.0.1"
	// StringIPv6LoopbackURL - IPv6 loopback URL with brackets.
	StringIPv6LoopbackURL = "[::1]"
	// StringIPv4MappedIPv6LoopbackURL - IPv4 mapped IPv6 loopback URL with brackets.
	StringIPv4MappedIPv6LoopbackURL = "[::ffff:127.0.0.1]"
	// StringLocalhostCIDRv4 - Localhost IPv4 CIDR.
	StringLocalhostCIDRv4 = "127.0.0.0/8"
	// StringLinkLocalCIDRv4 - Link local IPv4 CIDR.
	StringLinkLocalCIDRv4 = "169.254.0.0/16"
	// StringPrivateLANClassACIDRv4 - Private LAN class A IPv4 CIDR.
	StringPrivateLANClassACIDRv4 = "10.0.0.0/8"
	// StringPrivateLANClassBCIDRv4 - Private LAN class B IPv4 CIDR.
	StringPrivateLANClassBCIDRv4 = "172.16.0.0/12"
	// StringPrivateLANClassCCIDRv4 - Private LAN class C IPv4 CIDR.
	StringPrivateLANClassCCIDRv4 = "192.168.0.0/16"
	// StringLocalhostCIDRv6 - Localhost IPv6 CIDR.
	StringLocalhostCIDRv6 = "::1/128"
	// StringLinkLocalCIDRv6 - Link local IPv6 CIDR.
	StringLinkLocalCIDRv6 = "fe80::/10"
	// StringPrivateLANCIDRv6 - Private LAN IPv6 CIDR.
	StringPrivateLANCIDRv6 = "fc00::/7"
)

const (
	// DefaultCORSMaxAge - Default CORS max age in seconds.
	DefaultCORSMaxAge uint16 = 3600
	// DefaultHTTPRequestBodyLimit - Default request body limit in bytes (2MB).
	DefaultHTTPRequestBodyLimit = 2 << 20
	// DefaultCSRFTokenName - Default CSRF token name.
	DefaultCSRFTokenName = "_csrf"
	// DefaultCSRFTokenSameSiteStrict - Strict SameSite attribute.
	DefaultCSRFTokenSameSiteStrict = "Strict"
	// DefaultCSRFTokenMaxAge - CSRF token maximum age (1 hour).
	DefaultCSRFTokenMaxAge = 1 * time.Hour

	// DefaultCSRFTokenCookieSecure - Default CSRF token cookie secure flag.
	DefaultCSRFTokenCookieSecure = true
	// DefaultCSRFTokenCookieHTTPOnly - Default CSRF token cookie HTTPOnly flag.
	DefaultCSRFTokenCookieHTTPOnly = false
	// DefaultCSRFTokenCookieSessionOnly - Default CSRF token cookie session only flag.
	DefaultCSRFTokenCookieSessionOnly = true
	// DefaultCSRFTokenSingleUseToken - Default CSRF token single use flag.
	DefaultCSRFTokenSingleUseToken = false
)

var (
	// SliceDefaultAllowedIPs - Default allowed IP addresses.
	SliceDefaultAllowedIPs = []string{StringIPv4Loopback, StringIPv6Loopback, StringIPv4MappedIPv6Loopback}
	// SliceDefaultTLSPublicDNSNames - Default TLS public DNS names.
	SliceDefaultTLSPublicDNSNames = []string{StringLocalhost}
	// SliceDefaultTLSPublicIPAddresses - Default TLS public IP addresses.
	SliceDefaultTLSPublicIPAddresses = []string{StringIPv4Loopback, StringIPv6Loopback, StringIPv4MappedIPv6Loopback}
	// SliceDefaultTLSPrivateDNSNames - Default TLS private DNS names.
	SliceDefaultTLSPrivateDNSNames = []string{StringLocalhost}
	// SliceDefaultTLSPrivateIPAddresses - Default TLS private IP addresses.
	SliceDefaultTLSPrivateIPAddresses = []string{StringIPv4Loopback, StringIPv6Loopback, StringIPv4MappedIPv6Loopback}
	// SliceDefaultAllowedCIDRs - Default allowed CIDR ranges.
	SliceDefaultAllowedCIDRs = []string{
		StringLocalhostCIDRv4,
		StringLinkLocalCIDRv4,
		StringPrivateLANClassACIDRv4,
		StringPrivateLANClassBCIDRv4,
		StringPrivateLANClassCCIDRv4,
		StringLocalhostCIDRv6,
		StringLinkLocalCIDRv6,
		StringPrivateLANCIDRv6,
	}
	// SliceDefaultCORSAllowedMethods - Default CORS allowed methods.
	SliceDefaultCORSAllowedMethods = []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"}
	// SliceDefaultCORSAllowedHeaders - Default CORS allowed headers.
	SliceDefaultCORSAllowedHeaders = []string{
		"Content-Type",
		"Authorization",
		"Accept",
		"Origin",
		"X-Requested-With",
		"Cache-Control",
		"Pragma",
		"Expires",
		"_csrf",
	}
	// SliceDefaultConfigFiles - Default config files slice.
	SliceDefaultConfigFiles = []string{}
	// SliceDefaultUnsealFiles - Default unseal files slice.
	SliceDefaultUnsealFiles = []string{}
	// SliceDefaultCORSAllowedOrigins - Default CORS allowed origins.
	SliceDefaultCORSAllowedOrigins = []string{
		StringProtocolHTTP + "://" + StringLocalhost + ":" + "8080",
		StringProtocolHTTP + "://" + StringIPv4Loopback + ":" + "8080",
		StringProtocolHTTP + "://" + StringIPv6LoopbackURL + ":" + "8080",
		StringProtocolHTTP + "://" + StringIPv4MappedIPv6LoopbackURL + ":" + "8080",
		StringProtocolHTTPS + "://" + StringLocalhost + ":" + "8080",
		StringProtocolHTTPS + "://" + StringIPv4Loopback + ":" + "8080",
		StringProtocolHTTPS + "://" + StringIPv6LoopbackURL + ":" + "8080",
		StringProtocolHTTPS + "://" + StringIPv4MappedIPv6LoopbackURL + ":" + "8080",
	}
)
