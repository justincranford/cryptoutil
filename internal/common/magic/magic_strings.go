// Package magic provides commonly used magic numbers and values as named constants.
// This file contains string constants.
package magic

// String constants.
const (
	// StringUTCFormat - UTC time format string.
	StringUTCFormat = "2006-01-02T15:04:05Z"
	// StringLivezPath - Livez endpoint path.
	StringLivezPath = "/livez"
	// StringReadyzPath - Readyz endpoint path.
	StringReadyzPath = "/readyz"
	// StringShutdownPath - Shutdown endpoint path.
	StringShutdownPath = "/shutdown"
	// StringError - Error string constant.
	StringError = "error"
	// StringStatus - Status string constant.
	StringStatus = "status"
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
	// StringStatusOK - OK status string.
	StringStatusOK = "ok"
	// StringStatusDegraded - Degraded status string.
	StringStatusDegraded = "degraded"
	// StringProviderInternal - Internal provider string.
	StringProviderInternal = "Internal"
	// StringUUIDRegexPattern - UUID regex pattern for validation.
	StringUUIDRegexPattern = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	// StringPEMTypePKCS8PrivateKey - PKCS8 private key PEM type.
	StringPEMTypePKCS8PrivateKey = "PRIVATE KEY" // pragma: allowlist secret
	// StringPEMTypePKIXPublicKey - PKIX public key PEM type.
	StringPEMTypePKIXPublicKey = "PUBLIC KEY"
	// StringPEMTypeRSAPrivateKey - RSA private key PEM type.
	StringPEMTypeRSAPrivateKey = "RSA PRIVATE KEY" // pragma: allowlist secret
	// StringPEMTypeRSAPublicKey - RSA public key PEM type.
	StringPEMTypeRSAPublicKey = "RSA PUBLIC KEY"
	// StringPEMTypeECPrivateKey - EC private key PEM type.
	StringPEMTypeECPrivateKey = "EC PRIVATE KEY" // pragma: allowlist secret
	// StringPEMTypeCertificate - Certificate PEM type.
	StringPEMTypeCertificate = "CERTIFICATE"
	// StringPEMTypeCSR - Certificate signing request PEM type.
	StringPEMTypeCSR = "CERTIFICATE REQUEST"
	// StringPEMTypeSecretKey - Secret key PEM type.
	StringPEMTypeSecretKey = "SECRET KEY" // pragma: allowlist secret
	// StringLogLevelInfo - Default log level INFO.
	StringLogLevelInfo = "INFO"
	// StringDatabaseContainerDisabled - Disabled database container mode.
	StringDatabaseContainerDisabled = "disabled"
	// StringUnsealModeSysinfo - Sysinfo unseal mode.
	StringUnsealModeSysinfo = "sysinfo"
	// StringCSRFTokenName - Default CSRF token name.
	StringCSRFTokenName = "_csrf"
	// StringCSRFTokenSameSiteStrict - Strict SameSite attribute.
	StringCSRFTokenSameSiteStrict = "Strict"
	// StringEmpty - Empty string.
	StringEmpty = ""
	// StringPublicBrowserAPIContextPath - Default public browser API context path.
	StringPublicBrowserAPIContextPath = "/browser/api/v1"
	// StringPublicServiceAPIContextPath - Default public service API context path.
	StringPublicServiceAPIContextPath = "/service/api/v1"
	// StringOTLPServiceDefault - Default OTLP service name.
	StringOTLPServiceDefault = "cryptoutil"
	// StringOTLPVersionDefault - Default OTLP version.
	StringOTLPVersionDefault = "0.0.1"
	// StringOTLPEnvironmentDefault - Default OTLP environment.
	StringOTLPEnvironmentDefault = "dev"
	// StringOTLPHostnameDefault - Default OTLP hostname.
	StringOTLPHostnameDefault = "localhost"
	// StringOTLPEndpointDefault - Default OTLP endpoint.
	StringOTLPEndpointDefault = "grpc://127.0.0.1:4317"
	// StringDatabaseURLDefault - Default database URL with placeholder credentials.
	StringDatabaseURLDefault = "postgres://USR:PWD@localhost:5432/DB?sslmode=disable" // pragma: allowlist secret
)
