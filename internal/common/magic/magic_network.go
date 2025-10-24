// Package magic provides commonly used magic numbers and values as named constants.
// This file contains network-related constants.
package magic

import (
	"strconv"
	"time"
)

// Standard protocols.
const (
	// ProtocolHTTPS - HTTPS protocol string.
	ProtocolHTTPS = "https"
	// ProtocolHTTP - HTTP protocol string.
	ProtocolHTTP = "http"
)

// Standard ports.
const (
	// PortHTTPS - Standard HTTPS port.
	PortHTTPS uint16 = 443
	// PortHTTP - Standard HTTP port.
	PortHTTP uint16 = 80
)

// Loopback addresses.
const (
	// IPv4Loopback - IPv4 loopback address.
	IPv4Loopback = "127.0.0.1"
	// IPv6Loopback - IPv6 loopback address.
	IPv6Loopback = "::1"
	// IPv4MappedIPv6Loopback - IPv4 mapped IPv6 loopback address.
	IPv4MappedIPv6Loopback = "::ffff:127.0.0.1"
)

// Any addresses.
const (
	// IPv4AnyAddress - IPv4 any address.
	IPv4AnyAddress = "0.0.0.0"
	// IPv6AnyAddress - IPv6 any address.
	IPv6AnyAddress = "::"
)

const (
	// HostnameLocalhost - Localhost hostname.
	HostnameLocalhost = "localhost"
)

const (
	// LoopbackCIDRv4 - Localhost IPv4 CIDR.
	LoopbackCIDRv4 = "127.0.0.0/8"
	// LinkLocalCIDRv4 - Link local IPv4 CIDR.
	LinkLocalCIDRv4 = "169.254.0.0/16"
	// PrivateLANClassACIDRv4 - Private LAN class A IPv4 CIDR.
	PrivateLANClassACIDRv4 = "10.0.0.0/8"
	// PrivateLANClassBCIDRv4 - Private LAN class B IPv4 CIDR.
	PrivateLANClassBCIDRv4 = "172.16.0.0/12"
	// PrivateLANClassCCIDRv4 - Private LAN class C IPv4 CIDR.
	PrivateLANClassCCIDRv4 = "192.168.0.0/16"
	// LocalhostCIDRv6 - Localhost IPv6 CIDR.
	LocalhostCIDRv6 = "::1/128"
	// LinkLocalCIDRv6 - Link local IPv6 CIDR.
	LinkLocalCIDRv6 = "fe80::/10"
	// PrivateLANCIDRv6 - Private LAN IPv6 CIDR.
	PrivateLANCIDRv6 = "fc00::/7"
)

const (
	// IPv6LoopbackURL - IPv6 loopback URL with brackets.
	IPv6LoopbackURL = "[::1]"
	// IPv4MappedIPv6LoopbackURL - IPv4 mapped IPv6 loopback URL with brackets.
	IPv4MappedIPv6LoopbackURL = "[::ffff:127.0.0.1]"
)

// Network ports.
const (
	// DefaultPublicProtocolCryptoutil - Default public bind protocol.
	DefaultPublicProtocolCryptoutil = ProtocolHTTPS
	// DefaultPublicAddressCryptoutil - Default public bind address.
	DefaultPublicAddressCryptoutil = IPv4Loopback // Use 127.0.0.1 to avoid Docker localhost=[::1] issues
	// DefaultPublicPortCryptoutil - Default browser/server API port.
	DefaultPublicPortCryptoutil uint16 = 8080

	// DefaultPrivateProtocolCryptoutil - Default private bind protocol.
	DefaultPrivateProtocolCryptoutil = ProtocolHTTPS
	// DefaultPrivateAddressCryptoutil - Default private bind address.
	DefaultPrivateAddressCryptoutil = IPv4Loopback // Use 127.0.0.1 to avoid Docker localhost=[::1] issues
	// DefaultPrivatePortCryptoutil - Default admin API port.
	DefaultPrivatePortCryptoutil uint16 = 9090

	// DefaultPublicPortCryptoutilCompose0 - Port for cryptoutil SQLite instance.
	DefaultPublicPortCryptoutilCompose0 uint16 = 8080
	// DefaultPublicPortCryptoutilCompose1 - Port for cryptoutil PostgreSQL instance 1.
	DefaultPublicPortCryptoutilCompose1 uint16 = 8081
	// DefaultPublicPortCryptoutilCompose2 - Port for cryptoutil PostgreSQL instance 2.
	DefaultPublicPortCryptoutilCompose2 uint16 = 8082

	// DefaultPublicPortPostgres - Default PostgreSQL port.
	DefaultPublicPortPostgres uint16 = 5432
)

// Telemetry ports. See https://opentelemetry.io/docs/collector/configuration/.
const (
	// DefaultPublicPortInternalMetrics - Default OpenTelemetry collector internal metrics port (Prometheus).
	DefaultPublicPortInternalMetrics uint16 = 8888
	// PortOtelCollectorReceivedMetrics - Default OpenTelemetry collector received metrics port (Prometheus).
	PortOtelCollectorReceivedMetrics uint16 = 8889
	// DefaultPublicPortOtelCollectorHealth - Default OpenTelemetry collector health check port.
	DefaultPublicPortOtelCollectorHealth uint16 = 13133
	// DefaultPublicPortOtelCollectorPprof - Default OpenTelemetry collector pprof port.
	DefaultPublicPortOtelCollectorPprof uint16 = 1777
	// DefaultPublicPortOtelCollectorZPages - Default OpenTelemetry collector zPages port.
	DefaultPublicPortOtelCollectorZPages uint16 = 55679

	// DefaultPublicPortOtelCollectorGRPC - Default OpenTelemetry OTLP gRPC port.
	DefaultPublicPortOtelCollectorGRPC uint16 = 4317
	// DefaultPublicPortOtelCollectorHTTP - Default OpenTelemetry OTLP HTTP port.
	DefaultPublicPortOtelCollectorHTTP uint16 = 4318

	// DefaultPublicPortGrafana - Default Grafana port.
	DefaultPublicPortGrafana uint16 = 3000
	// PortGrafanaOTLPGRPC - Grafana OTEL LGTM OTLP gRPC receiver port (receives from OTEL collector).
	PortGrafanaOTLPGRPC uint16 = 14317
	// PortGrafanaOTLPHTTP - Grafana OTEL LGTM OTLP HTTP receiver port (receives from OTEL collector).
	PortGrafanaOTLPHTTP uint16 = 14318
	// DefaultPublicPortPrometheus - Default Prometheus port.
	DefaultPublicPortPrometheus uint16 = 9090
)

// Network URLs and prefixes.
const (
	// URLPrefixLocalhostHTTPS - HTTPS URL prefix for localhost.
	URLPrefixLocalhostHTTPS = "https://127.0.0.1:"
	// URLPrefixLocalhostHTTP - HTTP URL prefix for localhost.
	URLPrefixLocalhostHTTP = "http://127.0.0.1:"
)

const (
	// MaxIPRateLimit - Maximum allowed IP rate limit.
	MaxIPRateLimit uint16 = 10000

	// DefaultPublicBrowserAPIIPRateLimit - Default browser IP rate limit (100 requests/second).
	DefaultPublicBrowserAPIIPRateLimit uint16 = 100
	// DefaultPublicBrowserAPIContextPath - Default public browser API context path.
	DefaultPublicBrowserAPIContextPath = "/browser/api/v1"

	// DefaultPublicServiceAPIIPRateLimit - Default service IP rate limit (25 requests/second).
	DefaultPublicServiceAPIIPRateLimit uint16 = 25
	// DefaultPublicServiceAPIContextPath - Default public service API context path.
	DefaultPublicServiceAPIContextPath = "/service/api/v1"

	// PrivateAdminLivezRequestPath - Livez endpoint path.
	PrivateAdminLivezRequestPath = "/livez"
	// PrivateAdminReadyzRequestPath - Readyz endpoint path.
	PrivateAdminReadyzRequestPath = "/readyz"
	// PrivateAdminShutdownRequestPath - Shutdown endpoint path.
	PrivateAdminShutdownRequestPath = "/shutdown"
)

// DefaultIPFilterAllowedCIDRs - Default allowed CIDR ranges.
var DefaultIPFilterAllowedCIDRs = []string{
	LoopbackCIDRv4,
	LinkLocalCIDRv4,
	PrivateLANClassACIDRv4,
	PrivateLANClassBCIDRv4,
	PrivateLANClassCCIDRv4,
	LocalhostCIDRv6,
	LinkLocalCIDRv6,
	PrivateLANCIDRv6,
}

const (
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

	// DefaultCORSMaxAge - Default CORS max age in seconds.
	DefaultCORSMaxAge uint16 = 3600
)

var (
	// DefaultCORSAllowedMethods - Default CORS allowed methods.
	DefaultCORSAllowedMethods = []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"}
	// DefaultCORSAllowedHeaders - Default CORS allowed headers.
	DefaultCORSAllowedHeaders = []string{
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
	// DefaultCORSAllowedOrigins - Default CORS allowed origins.
	DefaultCORSAllowedOrigins = []string{
		"http://" + HostnameLocalhost + ":" + strconv.Itoa(int(DefaultPublicPortCryptoutil)),
		"http://" + IPv4Loopback + ":" + strconv.Itoa(int(DefaultPublicPortCryptoutil)),
		"http://" + IPv6LoopbackURL + ":" + strconv.Itoa(int(DefaultPublicPortCryptoutil)),
		"http://" + IPv4MappedIPv6LoopbackURL + ":" + strconv.Itoa(int(DefaultPublicPortCryptoutil)),
		"https://" + HostnameLocalhost + ":" + strconv.Itoa(int(DefaultPublicPortCryptoutil)),
		"https://" + IPv4Loopback + ":" + strconv.Itoa(int(DefaultPublicPortCryptoutil)),
		"https://" + IPv6LoopbackURL + ":" + strconv.Itoa(int(DefaultPublicPortCryptoutil)),
		"https://" + IPv4MappedIPv6LoopbackURL + ":" + strconv.Itoa(int(DefaultPublicPortCryptoutil)),
	}
)

var (
	// DefaultTLSPublicDNSNames - Default TLS public DNS names.
	DefaultTLSPublicDNSNames = []string{HostnameLocalhost}
	// DefaultTLSPublicIPAddresses - Default TLS public IP addresses.
	DefaultTLSPublicIPAddresses = []string{IPv4Loopback, IPv6Loopback, IPv4MappedIPv6Loopback}
	// DefaultTLSPrivateDNSNames - Default TLS private DNS names.
	DefaultTLSPrivateDNSNames = []string{HostnameLocalhost}
	// DefaultTLSPrivateIPAddresses - Default TLS private IP addresses.
	DefaultTLSPrivateIPAddresses = []string{IPv4Loopback, IPv6Loopback, IPv4MappedIPv6Loopback}
	// DefaultIPFilterAllowedIPs - Default allowed IP addresses.
	DefaultIPFilterAllowedIPs = []string{IPv4Loopback, IPv6Loopback, IPv4MappedIPv6Loopback}
)

const (
	// DefaultHTTPRequestBodyLimit - Default request body limit in bytes (2MB).
	DefaultHTTPRequestBodyLimit = 2 << 20
)
