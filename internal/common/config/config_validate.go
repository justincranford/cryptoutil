// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"fmt"
	"strings"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// validateConfiguration performs comprehensive validation of the configuration
// and returns detailed error messages with suggestions for fixes.
func validateConfiguration(s *Settings) error {
	var errors []string

	// Validate port ranges.
	if s.BindPublicPort < 1 {
		errors = append(errors, fmt.Sprintf("invalid public port %d: must be between 1 and 65535", s.BindPublicPort))
	}

	if s.BindPrivatePort < 1 {
		errors = append(errors, fmt.Sprintf("invalid private port %d: must be between 1 and 65535", s.BindPrivatePort))
	}

	if s.BindPublicPort == s.BindPrivatePort {
		errors = append(errors, fmt.Sprintf("public port (%d) and private port (%d) cannot be the same", s.BindPublicPort, s.BindPrivatePort))
	}

	// Validate protocols.
	if s.BindPublicProtocol != cryptoutilMagic.ProtocolHTTP && s.BindPublicProtocol != cryptoutilMagic.ProtocolHTTPS {
		errors = append(errors, fmt.Sprintf("invalid public protocol '%s': must be '%s' or '%s'", s.BindPublicProtocol, cryptoutilMagic.ProtocolHTTP, cryptoutilMagic.ProtocolHTTPS))
	}

	if s.BindPrivateProtocol != cryptoutilMagic.ProtocolHTTP && s.BindPrivateProtocol != cryptoutilMagic.ProtocolHTTPS {
		errors = append(errors, fmt.Sprintf("invalid private protocol '%s': must be '%s' or '%s'", s.BindPrivateProtocol, cryptoutilMagic.ProtocolHTTP, cryptoutilMagic.ProtocolHTTPS))
	}

	// Validate HTTPS requirements.
	if s.BindPublicProtocol == cryptoutilMagic.ProtocolHTTPS && len(s.TLSPublicDNSNames) == 0 && len(s.TLSPublicIPAddresses) == 0 {
		errors = append(errors, "HTTPS public protocol requires TLS DNS names or IP addresses to be configured")
	}

	if s.BindPrivateProtocol == "https" && len(s.TLSPrivateDNSNames) == 0 && len(s.TLSPrivateIPAddresses) == 0 {
		errors = append(errors, "HTTPS private protocol requires TLS DNS names or IP addresses to be configured")
	}

	// Validate database URL format.
	if s.DatabaseURL != "" && !strings.Contains(s.DatabaseURL, "://") {
		errors = append(errors, fmt.Sprintf("invalid database URL format '%s': must contain '://' (e.g., 'postgres://user:pass@host:port/db')", s.DatabaseURL))
	}

	// Validate CORS origins format.
	for _, origin := range s.CORSAllowedOrigins {
		if !strings.Contains(origin, "://") {
			errors = append(errors, fmt.Sprintf("invalid CORS origin format '%s': must contain '://' (e.g., 'https://example.com')", origin))
		}
	}

	// Validate log level.
	validLogLevels := []string{"ALL", "TRACE", "DEBUG", "CONFIG", "INFO", "NOTICE", "WARN", "WARNING", "ERROR", "FATAL", "OFF"}
	logLevelValid := false

	for _, level := range validLogLevels {
		if strings.EqualFold(s.LogLevel, level) {
			logLevelValid = true

			break
		}
	}

	if !logLevelValid {
		errors = append(errors, fmt.Sprintf("invalid log level '%s': must be one of %v", s.LogLevel, validLogLevels))
	}

	// Validate rate limits.
	if s.BrowserIPRateLimit == 0 {
		errors = append(errors, "browser rate limit cannot be 0 (would block all browser requests)")
	} else if s.BrowserIPRateLimit > cryptoutilMagic.MaxIPRateLimit {
		errors = append(errors, fmt.Sprintf("browser rate limit %d is very high (>%d), may impact performance", s.BrowserIPRateLimit, cryptoutilMagic.MaxIPRateLimit))
	}

	if s.ServiceIPRateLimit == 0 {
		errors = append(errors, "service rate limit cannot be 0 (would block all service requests)")
	} else if s.ServiceIPRateLimit > cryptoutilMagic.MaxIPRateLimit {
		errors = append(errors, fmt.Sprintf("service rate limit %d is very high (>%d), may impact performance", s.ServiceIPRateLimit, cryptoutilMagic.MaxIPRateLimit))
	}

	// Validate OTLP endpoint format.
	if s.OTLPEnabled && s.OTLPEndpoint != "" {
		if !strings.HasPrefix(s.OTLPEndpoint, "grpc://") && !strings.HasPrefix(s.OTLPEndpoint, "http://") && !strings.HasPrefix(s.OTLPEndpoint, "https://") {
			errors = append(errors, fmt.Sprintf("invalid OTLP endpoint format '%s': must start with 'grpc://', 'http://', or 'https://'", s.OTLPEndpoint))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s\n\nSuggestions:\n- Use --dry-run to validate configuration without starting\n- Check configuration file syntax\n- Use --profile flag for common deployment scenarios\n- See --help for detailed option descriptions", strings.Join(errors, "\n"))
	}

	return nil
}
