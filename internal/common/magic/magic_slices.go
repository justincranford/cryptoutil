// Package magic provides commonly used magic numbers and values as named constants.
// This file contains slice constants.
package magic

// Slice constants.
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
