// Package magic provides commonly used magic numbers and values as named constants.
// This file contains security-related constants.
package magic

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
