// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "time"

// Security header policy constants.
const (
	// HSTSMaxAge - HSTS max-age for production (1 year with preload).
	HSTSMaxAge = "max-age=31536000; includeSubDomains; preload"
	// HSTSMaxAgeDev - HSTS max-age for development (1 day).
	HSTSMaxAgeDev = "max-age=86400; includeSubDomains"
	// ReferrerPolicy - Referrer policy for security.
	ReferrerPolicy = "strict-origin-when-cross-origin"
	// PermissionsPolicy - Permissions policy to disable device access.
	PermissionsPolicy = "camera=(), microphone=(), geolocation=(), payment=(), usb=(), accelerometer=(), gyroscope=(), magnetometer=()"
	// CrossOriginOpenerPolicy - Cross-origin opener policy.
	CrossOriginOpenerPolicy = "same-origin"
	// CrossOriginEmbedderPolicy - Cross-origin embedder policy.
	CrossOriginEmbedderPolicy = "require-corp"
	// CrossOriginResourcePolicy - Cross-origin resource policy.
	CrossOriginResourcePolicy = "same-origin"
	// XPermittedCrossDomainPolicies - X-Permitted-Cross-Domain-Policies header.
	XPermittedCrossDomainPolicies = "none"
	// ContentTypeOptions - X-Content-Type-Options header.
	ContentTypeOptions = "nosniff"
	// ClearSiteDataLogout - Clear-Site-Data header for logout.
	ClearSiteDataLogout = "\"cache\", \"cookies\", \"storage\""
)

// Token introspection constants.
const (
	// IntrospectionCacheTTLMinutes is the default cache TTL for introspection results.
	IntrospectionCacheTTLMinutes = 5

	// IntrospectionMaxBatchSize is the default max batch size for introspection.
	IntrospectionMaxBatchSize = 10

	// IntrospectionHTTPTimeoutSeconds is the default HTTP timeout for introspection.
	IntrospectionHTTPTimeoutSeconds = 10

	// JWKSCacheTTLMinutes is the default cache TTL for JWKS keys.
	JWKSCacheTTLMinutes = 5
)

// Token introspection durations.
var (
	// IntrospectionCacheTTL is the default cache TTL as a duration.
	IntrospectionCacheTTL = IntrospectionCacheTTLMinutes * time.Minute

	// IntrospectionHTTPTimeout is the default HTTP timeout as a duration.
	IntrospectionHTTPTimeout = IntrospectionHTTPTimeoutSeconds * time.Second

	// JWKSCacheTTL is the default cache TTL for JWKS keys.
	JWKSCacheTTL = JWKSCacheTTLMinutes * time.Minute
)
