// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func logSettings(s *ServiceTemplateServerSettings) {
	if s.VerboseMode {
		log.Info("Sub Command: ", s.SubCommand)

		// Create a map to get values by setting name
		valueMap := map[string]any{
			help.Name:                        s.Help,
			configFile.Name:                  s.ConfigFile,
			logLevel.Name:                    s.LogLevel,
			verboseMode.Name:                 s.VerboseMode,
			devMode.Name:                     s.DevMode,
			dryRun.Name:                      s.DryRun,
			profile.Name:                     s.Profile,
			bindPublicProtocol.Name:          s.BindPublicProtocol,
			bindPublicAddress.Name:           s.BindPublicAddress,
			bindPublicPort.Name:              s.BindPublicPort,
			tlsPublicDNSNames.Name:           s.TLSPublicDNSNames,
			tlsPublicIPAddresses.Name:        s.TLSPublicIPAddresses,
			tlsPrivateDNSNames.Name:          s.TLSPrivateDNSNames,
			tlsPrivateIPAddresses.Name:       s.TLSPrivateIPAddresses,
			bindPrivateProtocol.Name:         s.BindPrivateProtocol,
			bindPrivateAddress.Name:          s.BindPrivateAddress,
			bindPrivatePort.Name:             s.BindPrivatePort,
			publicBrowserAPIContextPath.Name: s.PublicBrowserAPIContextPath,
			publicServiceAPIContextPath.Name: s.PublicServiceAPIContextPath,
			privateAdminAPIContextPath.Name:  s.PrivateAdminAPIContextPath,
			corsAllowedOrigins.Name:          s.CORSAllowedOrigins,
			corsAllowedMethods.Name:          s.CORSAllowedMethods,
			corsAllowedHeaders.Name:          s.CORSAllowedHeaders,
			corsMaxAge.Name:                  s.CORSMaxAge,
			requestBodyLimit.Name:            s.RequestBodyLimit,
			csrfTokenName.Name:               s.CSRFTokenName,
			csrfTokenSameSite.Name:           s.CSRFTokenSameSite,
			csrfTokenMaxAge.Name:             s.CSRFTokenMaxAge,
			csrfTokenCookieSecure.Name:       s.CSRFTokenCookieSecure,
			csrfTokenCookieHTTPOnly.Name:     s.CSRFTokenCookieHTTPOnly,
			csrfTokenCookieSessionOnly.Name:  s.CSRFTokenCookieSessionOnly,
			csrfTokenSingleUseToken.Name:     s.CSRFTokenSingleUseToken,
			browserIPRateLimit.Name:          s.BrowserIPRateLimit,
			serviceIPRateLimit.Name:          s.ServiceIPRateLimit,
			allowedIps.Name:                  s.AllowedIPs,
			allowedCidrs.Name:                s.AllowedCIDRs,
			databaseContainer.Name:           s.DatabaseContainer,
			databaseURL.Name:                 s.DatabaseURL,
			databaseInitTotalTimeout.Name:    s.DatabaseInitTotalTimeout,
			databaseInitRetryWait.Name:       s.DatabaseInitRetryWait,
			otlpEnabled.Name:                 s.OTLPEnabled,
			otlpConsole.Name:                 s.OTLPConsole,
			otlpService.Name:                 s.OTLPService,
			otlpVersion.Name:                 s.OTLPVersion,
			otlpEnvironment.Name:             s.OTLPEnvironment,
			otlpHostname.Name:                s.OTLPHostname,
			otlpEndpoint.Name:                s.OTLPEndpoint,
			unsealMode.Name:                  s.UnsealMode,
			unsealFiles.Name:                 s.UnsealFiles,
			browserRealms.Name:               s.BrowserRealms,
			serviceRealms.Name:               s.ServiceRealms,
			browserSessionCookie.Name:        s.BrowserSessionCookie,
		}

		// Iterate through all registered settings and log them
		for _, setting := range allServeiceTemplateServerRegisteredSettings {
			value := valueMap[setting.Name]
			if setting.Redacted && (!s.DevMode || !s.VerboseMode) {
				value = "REDACTED"
			}

			log.Info(setting.Description+" (-"+setting.Shorthand+"): ", value)
		}

		analysis := analyzeSettings(allServeiceTemplateServerRegisteredSettings)

		var usedShorthands []string

		var unusedShorthands []string

		// Check all letters (lowercase and uppercase) and digits
		allPossibleShorthands := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		for _, r := range allPossibleShorthands {
			possibleShorthand := string(r)
			if _, ok := analysis.SettingsByShorthands[possibleShorthand]; ok {
				usedShorthands = append(usedShorthands, possibleShorthand)
			} else {
				unusedShorthands = append(unusedShorthands, possibleShorthand)
			}
		}

		sort.Strings(usedShorthands)
		sort.Strings(unusedShorthands)
		log.Info("Shorthands, Used:   ", len(usedShorthands), ", Values: ", usedShorthands)
		log.Info("Shorthands, Unused: ", len(unusedShorthands), ", Values: ", unusedShorthands)
	}
}

func resetFlags() {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	viper.Reset()
}

// SetEnvAndRegisterSetting sets the environment variable name and registers the setting.
func SetEnvAndRegisterSetting(_ []*Setting, setting *Setting) *Setting {
	setting.Env = "CRYPTOUTIL_" + strings.ToUpper(strings.ReplaceAll(setting.Name, "-", "_"))

	allServeiceTemplateServerRegisteredSettings = append(allServeiceTemplateServerRegisteredSettings, setting)

	return setting
}

// RegisterAsBoolSetting extracts a bool value from a Setting with type assertion.
func RegisterAsBoolSetting(s *Setting) bool {
	if v, ok := s.Value.(bool); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not bool", s.Name))
}

// RegisterAsStringSetting extracts a string value from a Setting with type assertion.
func RegisterAsStringSetting(s *Setting) string {
	if v, ok := s.Value.(string); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not string", s.Name))
}

// RegisterAsUint16Setting extracts a uint16 value from a Setting with type assertion.
func RegisterAsUint16Setting(s *Setting) uint16 {
	if v, ok := s.Value.(uint16); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not uint16", s.Name))
}

// RegisterAsStringSliceSetting extracts a string slice value from a Setting with type assertion.
func RegisterAsStringSliceSetting(s *Setting) []string {
	if v, ok := s.Value.([]string); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not []string", s.Name))
}

// RegisterAsStringArraySetting extracts a string array value from a Setting with type assertion.
func RegisterAsStringArraySetting(s *Setting) []string {
	if v, ok := s.Value.([]string); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not []string for array", s.Name))
}

// RegisterAsDurationSetting extracts a time.Duration value from a Setting with type assertion.
func RegisterAsDurationSetting(s *Setting) time.Duration {
	if v, ok := s.Value.(time.Duration); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not time.Duration", s.Name))
}

// RegisterAsIntSetting extracts an int value from a Setting with type assertion.
func RegisterAsIntSetting(s *Setting) int {
	if v, ok := s.Value.(int); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not int", s.Name))
}

func formatDefault(value any) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case []string:
		if len(v) == 0 {
			return "[]"
		}

		return fmt.Sprintf("[%s]", strings.Join(v, ","))
	case bool:
		return fmt.Sprintf("%t", v)
	case uint16:
		return fmt.Sprintf("%d", v)
	case time.Duration:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func analyzeSettings(settings []*Setting) analysisResult {
	result := analysisResult{
		SettingsByNames:      make(map[string][]*Setting),
		SettingsByShorthands: make(map[string][]*Setting),
	}
	for _, setting := range settings {
		result.SettingsByNames[setting.Name] = append(result.SettingsByNames[setting.Name], setting)
		// Only track non-empty shorthands
		if setting.Shorthand != "" {
			result.SettingsByShorthands[setting.Shorthand] = append(result.SettingsByShorthands[setting.Shorthand], setting)
		}
	}

	for _, setting := range settings {
		if len(result.SettingsByNames[setting.Name]) > 1 {
			result.DuplicateNames = append(result.DuplicateNames, setting.Name)
		}

		if setting.Shorthand != "" && len(result.SettingsByShorthands[setting.Shorthand]) > 1 {
			result.DuplicateShorthands = append(result.DuplicateShorthands, setting.Shorthand)
		}
	}

	return result
}

// validateConfiguration performs comprehensive validation of the configuration
// and returns detailed error messages with suggestions for fixes.
func validateConfiguration(s *ServiceTemplateServerSettings) error {
	var errors []string

	// Validate bind addresses (CRITICAL: blank address produces ":port" which binds to 0.0.0.0 triggering Windows Firewall).
	if s.BindPublicAddress == "" {
		errors = append(errors, "bind public address cannot be blank (would bind to 0.0.0.0 triggering Windows Firewall): use '127.0.0.1' for localhost or explicit IP")
	}

	if s.BindPrivateAddress == "" {
		errors = append(errors, "bind private address cannot be blank (would bind to 0.0.0.0 triggering Windows Firewall): use '127.0.0.1' for localhost or explicit IP")
	}

	// CRITICAL: In test/dev environments, reject 0.0.0.0 to prevent Windows Firewall prompts.
	// Production containers may use 0.0.0.0 for external access (isolated network namespace).
	if s.DevMode && s.BindPublicAddress == cryptoutilSharedMagic.IPv4AnyAddress {
		errors = append(errors, "CRITICAL: bind public address cannot be 0.0.0.0 in test/dev mode (triggers Windows Firewall prompts): use '127.0.0.1' for localhost")
	}

	if s.DevMode && s.BindPrivateAddress == cryptoutilSharedMagic.IPv4AnyAddress {
		errors = append(errors, "CRITICAL: bind private address cannot be 0.0.0.0 in test/dev mode (triggers Windows Firewall prompts): use '127.0.0.1' for localhost")
	}

	// Validate port ranges (port 0 is valid - OS assigns dynamic port).
	if s.BindPublicPort > cryptoutilSharedMagic.MaxPortNumber {
		errors = append(errors, fmt.Sprintf("invalid public port %d: must be between 0 and 65535", s.BindPublicPort))
	}

	if s.BindPrivatePort > cryptoutilSharedMagic.MaxPortNumber {
		errors = append(errors, fmt.Sprintf("invalid private port %d: must be between 0 and 65535", s.BindPrivatePort))
	}

	// Ports cannot be the same unless both are 0 (OS assigns different dynamic ports).
	if s.BindPublicPort == s.BindPrivatePort && s.BindPublicPort != 0 {
		errors = append(errors, fmt.Sprintf("public port (%d) and private port (%d) cannot be the same", s.BindPublicPort, s.BindPrivatePort))
	}

	// Validate protocols
	if s.BindPublicProtocol != cryptoutilSharedMagic.ProtocolHTTP && s.BindPublicProtocol != cryptoutilSharedMagic.ProtocolHTTPS {
		errors = append(errors, fmt.Sprintf("invalid public protocol '%s': must be '%s' or '%s'", s.BindPublicProtocol, cryptoutilSharedMagic.ProtocolHTTP, cryptoutilSharedMagic.ProtocolHTTPS))
	}

	if s.BindPrivateProtocol != cryptoutilSharedMagic.ProtocolHTTP && s.BindPrivateProtocol != cryptoutilSharedMagic.ProtocolHTTPS {
		errors = append(errors, fmt.Sprintf("invalid private protocol '%s': must be '%s' or '%s'", s.BindPrivateProtocol, cryptoutilSharedMagic.ProtocolHTTP, cryptoutilSharedMagic.ProtocolHTTPS))
	}

	// Validate HTTPS requirements
	if s.BindPublicProtocol == cryptoutilSharedMagic.ProtocolHTTPS && len(s.TLSPublicDNSNames) == 0 && len(s.TLSPublicIPAddresses) == 0 {
		errors = append(errors, "HTTPS public protocol requires TLS DNS names or IP addresses to be configured")
	}

	if s.BindPrivateProtocol == "https" && len(s.TLSPrivateDNSNames) == 0 && len(s.TLSPrivateIPAddresses) == 0 {
		errors = append(errors, "HTTPS private protocol requires TLS DNS names or IP addresses to be configured")
	}

	// Validate database URL format
	// Allow special SQLite formats: ":memory:", "file::memory:?cache=shared"
	// Standard formats must contain "://" (e.g., "postgres://...", "file://...")
	if s.DatabaseURL != "" &&
		s.DatabaseURL != ":memory:" &&
		s.DatabaseURL != "file::memory:?cache=shared" &&
		!strings.Contains(s.DatabaseURL, "://") {
		errors = append(errors, fmt.Sprintf("invalid database URL format '%s': must contain '://' (e.g., 'postgres://user:pass@host:port/db') or use SQLite special formats (':memory:', 'file::memory:?cache=shared')", s.DatabaseURL))
	}

	// Validate CORS origins format
	for _, origin := range s.CORSAllowedOrigins {
		if !strings.Contains(origin, "://") {
			errors = append(errors, fmt.Sprintf("invalid CORS origin format '%s': must contain '://' (e.g., 'https://example.com')", origin))
		}
	}

	// Validate log level
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

	// Validate rate limits
	if s.BrowserIPRateLimit == 0 {
		errors = append(errors, "browser rate limit cannot be 0 (would block all browser requests)")
	} else if s.BrowserIPRateLimit > cryptoutilSharedMagic.MaxIPRateLimit {
		errors = append(errors, fmt.Sprintf("browser rate limit %d is very high (>%d), may impact performance", s.BrowserIPRateLimit, cryptoutilSharedMagic.MaxIPRateLimit))
	}

	if s.ServiceIPRateLimit == 0 {
		errors = append(errors, "service rate limit cannot be 0 (would block all service requests)")
	} else if s.ServiceIPRateLimit > cryptoutilSharedMagic.MaxIPRateLimit {
		errors = append(errors, fmt.Sprintf("service rate limit %d is very high (>%d), may impact performance", s.ServiceIPRateLimit, cryptoutilSharedMagic.MaxIPRateLimit))
	}

	// Validate OTLP endpoint format
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

// resolveFileURL reads the content of a file if the value starts with "file://".
// This pattern is used for Docker secrets and Kubernetes secrets mounted as files.
// Example: "file:///run/secrets/database_url" reads the secret file content.
func resolveFileURL(value string) string {
	if !strings.HasPrefix(value, "file://") {
		return value
	}

	filePath := strings.TrimPrefix(value, "file://")

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Warnf("Failed to read file URL %s: %v (using value as-is)", value, err)

		return value
	}

	return strings.TrimSpace(string(content))
}

// NewForJOSEServer creates settings suitable for the JOSE Authority Server.
func NewForJOSEServer(bindAddr string, bindPort uint16, devMode bool) *ServiceTemplateServerSettings {
	// Build args for Parse()
	args := []string{
		"start", // Subcommand required
		"--bind-public-address", bindAddr,
		"--bind-public-port", fmt.Sprintf("%d", bindPort),
		"--otlp-service", "jose-ja",
	}

	if devMode {
		args = append(args, "--dev")
	}

	settings, err := Parse(args, false)
	if err != nil {
		// Should not fail with valid default args
		panic(fmt.Sprintf("NewForJOSEServer failed to parse args: %v", err))
	}

	return settings
}

// NewForCAServer creates settings suitable for the CA Server.
func NewForCAServer(bindAddr string, bindPort uint16, devMode bool) *ServiceTemplateServerSettings {
	// Build args for Parse()
	args := []string{
		"start", // Subcommand required
		"--bind-public-address", bindAddr,
		"--bind-public-port", fmt.Sprintf("%d", bindPort),
		"--otlp-service", "pki-ca",
	}

	if devMode {
		args = append(args, "--dev")
	}

	settings, err := Parse(args, false)
	if err != nil {
		// Should not fail with valid default args
		panic(fmt.Sprintf("NewForCAServer failed to parse args: %v", err))
	}

	return settings
}
