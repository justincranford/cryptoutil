// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// getTLSPEMBytes safely retrieves PEM bytes from viper for BytesBase64 flags.
// Returns nil if the value is not set or cannot be converted to []byte.
func getTLSPEMBytes(key string) []byte {
	val := viper.Get(key)
	if val == nil {
		return nil
	}

	// BytesBase64P flags are stored in viper as strings (base64-encoded)
	// We need to decode them manually
	if str, ok := val.(string); ok {
		if str == "" {
			return nil
		}

		bytes, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return nil
		}

		return bytes
	}

	// Fallback: if already []byte (e.g., from config file), use as-is
	if bytes, ok := val.([]byte); ok {
		return bytes
	}

	return nil
}

// Parse parses command line parameters and returns application settings.
// ParseWithFlagSet parses command parameters into ServiceTemplateServerSettings using a custom FlagSet.
// This function enables benchmark testing by accepting a fresh FlagSet for each iteration,
// avoiding pflag's "flag redefined" panics when the same flags are registered multiple times.
//
// Parameters:
//   - fs: Custom FlagSet to register flags on (use pflag.NewFlagSet() for benchmarks, pflag.CommandLine for production)
//   - commandParameters: Command line arguments (first element is subcommand, rest are flags)
//   - exitIfHelp: If true, os.Exit(0) when --help flag is set
//
// Returns:
//   - *ServiceTemplateServerSettings: Parsed configuration settings
//   - error: Validation or parsing errors
func ParseWithFlagSet(fs *pflag.FlagSet, commandParameters []string, exitIfHelp bool) (*ServiceTemplateServerSettings, error) {
	if len(commandParameters) == 0 {
		return nil, fmt.Errorf("missing subcommand: use \"start\", \"stop\", \"init\", \"live\", or \"ready\"")
	}

	subCommand := commandParameters[0]
	if _, ok := subcommands[subCommand]; !ok {
		return nil, fmt.Errorf("invalid subcommand: use \"start\", \"stop\", \"init\", \"live\", or \"ready\"")
	}

	subCommandParameters := commandParameters[1:]

	// Lock viperMutex to prevent concurrent map writes when tests run in parallel.
	// viper uses global maps for environment variable bindings and other state, so we must serialize access.
	viperMutex.Lock()
	defer viperMutex.Unlock()

	// Enable environment variable support with CRYPTOUTIL_ prefix BEFORE parsing flags
	viper.SetEnvPrefix("CRYPTOUTIL")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Explicitly bind boolean environment variables (viper.AutomaticEnv may not handle booleans correctly)
	// Note: viper.BindEnv errors are logged but don't prevent startup as they are extremely rare
	for _, setting := range allServiceTemplateServerRegisteredSettings {
		if _, ok := setting.Value.(bool); ok {
			if err := viper.BindEnv(setting.Name, setting.Env); err != nil {
				fmt.Printf("Warning: failed to bind environment variable %s: %v\n", setting.Env, err)
			}
		}
	}

	// Register flags on custom FlagSet (fs parameter instead of global pflag.CommandLine)
	fs.BoolP(help.Name, help.Shorthand, RegisterAsBoolSetting(&help), help.Usage)
	fs.StringSliceP(configFile.Name, configFile.Shorthand, RegisterAsStringSliceSetting(&configFile), configFile.Usage)
	fs.StringP(logLevel.Name, logLevel.Shorthand, RegisterAsStringSetting(&logLevel), logLevel.Usage)
	fs.BoolP(verboseMode.Name, verboseMode.Shorthand, RegisterAsBoolSetting(&verboseMode), verboseMode.Usage)
	fs.BoolP(devMode.Name, devMode.Shorthand, RegisterAsBoolSetting(&devMode), devMode.Usage)
	fs.BoolP(demoMode.Name, demoMode.Shorthand, RegisterAsBoolSetting(&demoMode), demoMode.Usage)
	fs.BoolP(dryRun.Name, dryRun.Shorthand, RegisterAsBoolSetting(&dryRun), dryRun.Usage)
	fs.StringP(profile.Name, profile.Shorthand, RegisterAsStringSetting(&profile), profile.Usage)
	fs.StringP(bindPublicProtocol.Name, bindPublicProtocol.Shorthand, RegisterAsStringSetting(&bindPublicProtocol), bindPublicProtocol.Usage)
	fs.StringP(bindPublicAddress.Name, bindPublicAddress.Shorthand, RegisterAsStringSetting(&bindPublicAddress), bindPublicAddress.Usage)
	fs.Uint16P(bindPublicPort.Name, bindPublicPort.Shorthand, RegisterAsUint16Setting(&bindPublicPort), bindPublicPort.Usage)
	fs.StringSliceP(tlsPublicDNSNames.Name, tlsPublicDNSNames.Shorthand, RegisterAsStringSliceSetting(&tlsPublicDNSNames), tlsPublicDNSNames.Usage)
	fs.StringSliceP(tlsPublicIPAddresses.Name, tlsPublicIPAddresses.Shorthand, RegisterAsStringSliceSetting(&tlsPublicIPAddresses), tlsPublicIPAddresses.Usage)
	fs.StringSliceP(tlsPrivateDNSNames.Name, tlsPrivateDNSNames.Shorthand, RegisterAsStringSliceSetting(&tlsPrivateDNSNames), tlsPrivateDNSNames.Usage)
	fs.StringSliceP(tlsPrivateIPAddresses.Name, tlsPrivateIPAddresses.Shorthand, RegisterAsStringSliceSetting(&tlsPrivateIPAddresses), tlsPrivateIPAddresses.Usage)
	fs.StringP(tlsPublicMode.Name, tlsPublicMode.Shorthand, string(defaultTLSPublicMode), tlsPublicMode.Usage)
	fs.StringP(tlsPrivateMode.Name, tlsPrivateMode.Shorthand, string(defaultTLSPrivateMode), tlsPrivateMode.Usage)
	fs.BytesBase64P(tlsStaticCertPEM.Name, tlsStaticCertPEM.Shorthand, []byte(nil), tlsStaticCertPEM.Usage)
	fs.BytesBase64P(tlsStaticKeyPEM.Name, tlsStaticKeyPEM.Shorthand, []byte(nil), tlsStaticKeyPEM.Usage)
	fs.BytesBase64P(tlsMixedCACertPEM.Name, tlsMixedCACertPEM.Shorthand, []byte(nil), tlsMixedCACertPEM.Usage)
	fs.BytesBase64P(tlsMixedCAKeyPEM.Name, tlsMixedCAKeyPEM.Shorthand, []byte(nil), tlsMixedCAKeyPEM.Usage)
	fs.StringP(bindPrivateProtocol.Name, bindPrivateProtocol.Shorthand, RegisterAsStringSetting(&bindPrivateProtocol), bindPrivateProtocol.Usage)
	fs.StringP(bindPrivateAddress.Name, bindPrivateAddress.Shorthand, RegisterAsStringSetting(&bindPrivateAddress), bindPrivateAddress.Usage)
	fs.Uint16P(bindPrivatePort.Name, bindPrivatePort.Shorthand, RegisterAsUint16Setting(&bindPrivatePort), bindPrivatePort.Usage)
	fs.StringP(publicBrowserAPIContextPath.Name, publicBrowserAPIContextPath.Shorthand, RegisterAsStringSetting(&publicBrowserAPIContextPath), publicBrowserAPIContextPath.Usage)
	fs.StringP(publicServiceAPIContextPath.Name, publicServiceAPIContextPath.Shorthand, RegisterAsStringSetting(&publicServiceAPIContextPath), publicServiceAPIContextPath.Usage)
	fs.StringP(privateAdminAPIContextPath.Name, privateAdminAPIContextPath.Shorthand, RegisterAsStringSetting(&privateAdminAPIContextPath), privateAdminAPIContextPath.Usage)
	fs.StringSliceP(corsAllowedOrigins.Name, corsAllowedOrigins.Shorthand, RegisterAsStringSliceSetting(&corsAllowedOrigins), corsAllowedOrigins.Usage)
	fs.StringSliceP(corsAllowedMethods.Name, corsAllowedMethods.Shorthand, RegisterAsStringSliceSetting(&corsAllowedMethods), corsAllowedMethods.Usage)
	fs.StringSliceP(corsAllowedHeaders.Name, corsAllowedHeaders.Shorthand, RegisterAsStringSliceSetting(&corsAllowedHeaders), corsAllowedHeaders.Usage)
	fs.Uint16P(corsMaxAge.Name, corsMaxAge.Shorthand, RegisterAsUint16Setting(&corsMaxAge), corsMaxAge.Usage)
	fs.StringP(csrfTokenName.Name, csrfTokenName.Shorthand, RegisterAsStringSetting(&csrfTokenName), csrfTokenName.Usage)
	fs.StringP(csrfTokenSameSite.Name, csrfTokenSameSite.Shorthand, RegisterAsStringSetting(&csrfTokenSameSite), csrfTokenSameSite.Usage)
	fs.DurationP(csrfTokenMaxAge.Name, csrfTokenMaxAge.Shorthand, RegisterAsDurationSetting(&csrfTokenMaxAge), csrfTokenMaxAge.Usage)
	fs.BoolP(csrfTokenCookieSecure.Name, csrfTokenCookieSecure.Shorthand, RegisterAsBoolSetting(&csrfTokenCookieSecure), csrfTokenCookieSecure.Usage)
	fs.BoolP(csrfTokenCookieHTTPOnly.Name, csrfTokenCookieHTTPOnly.Shorthand, RegisterAsBoolSetting(&csrfTokenCookieHTTPOnly), csrfTokenCookieHTTPOnly.Usage)
	fs.BoolP(csrfTokenCookieSessionOnly.Name, csrfTokenCookieSessionOnly.Shorthand, RegisterAsBoolSetting(&csrfTokenCookieSessionOnly), csrfTokenCookieSessionOnly.Usage)
	fs.BoolP(csrfTokenSingleUseToken.Name, csrfTokenSingleUseToken.Shorthand, RegisterAsBoolSetting(&csrfTokenSingleUseToken), csrfTokenSingleUseToken.Usage)
	fs.Uint16P(browserIPRateLimit.Name, browserIPRateLimit.Shorthand, RegisterAsUint16Setting(&browserIPRateLimit), browserIPRateLimit.Usage)
	fs.Uint16P(serviceIPRateLimit.Name, serviceIPRateLimit.Shorthand, RegisterAsUint16Setting(&serviceIPRateLimit), serviceIPRateLimit.Usage)
	fs.StringSliceP(allowedIps.Name, allowedIps.Shorthand, RegisterAsStringSliceSetting(&allowedIps), allowedIps.Usage)
	fs.StringSliceP(allowedCidrs.Name, allowedCidrs.Shorthand, RegisterAsStringSliceSetting(&allowedCidrs), allowedCidrs.Usage)
	fs.IntP(requestBodyLimit.Name, requestBodyLimit.Shorthand, RegisterAsIntSetting(&requestBodyLimit), requestBodyLimit.Usage)
	fs.StringP(databaseContainer.Name, databaseContainer.Shorthand, RegisterAsStringSetting(&databaseContainer), databaseContainer.Usage)
	fs.StringP(databaseURL.Name, databaseURL.Shorthand, RegisterAsStringSetting(&databaseURL), databaseURL.Usage)
	fs.DurationP(databaseInitTotalTimeout.Name, databaseInitTotalTimeout.Shorthand, RegisterAsDurationSetting(&databaseInitTotalTimeout), databaseInitTotalTimeout.Usage)
	fs.DurationP(databaseInitRetryWait.Name, databaseInitRetryWait.Shorthand, RegisterAsDurationSetting(&databaseInitRetryWait), databaseInitRetryWait.Usage)
	fs.DurationP(serverShutdownTimeout.Name, serverShutdownTimeout.Shorthand, RegisterAsDurationSetting(&serverShutdownTimeout), serverShutdownTimeout.Usage)
	fs.BoolP(otlpEnabled.Name, otlpEnabled.Shorthand, RegisterAsBoolSetting(&otlpEnabled), otlpEnabled.Usage)
	fs.BoolP(otlpConsole.Name, otlpConsole.Shorthand, RegisterAsBoolSetting(&otlpConsole), otlpConsole.Usage)
	fs.StringP(otlpService.Name, otlpService.Shorthand, RegisterAsStringSetting(&otlpService), otlpService.Usage)
	fs.StringP(otlpVersion.Name, otlpVersion.Shorthand, RegisterAsStringSetting(&otlpVersion), otlpVersion.Usage)
	fs.StringP(otlpEnvironment.Name, otlpEnvironment.Shorthand, RegisterAsStringSetting(&otlpEnvironment), otlpEnvironment.Usage)
	fs.StringP(otlpHostname.Name, otlpHostname.Shorthand, RegisterAsStringSetting(&otlpHostname), otlpHostname.Usage)
	fs.StringP(otlpEndpoint.Name, otlpEndpoint.Shorthand, RegisterAsStringSetting(&otlpEndpoint), otlpEndpoint.Usage)
	fs.StringP(otlpInstance.Name, otlpInstance.Shorthand, RegisterAsStringSetting(&otlpInstance), otlpInstance.Usage)
	fs.StringP(unsealMode.Name, unsealMode.Shorthand, RegisterAsStringSetting(&unsealMode), unsealMode.Usage)
	fs.StringArrayP(unsealFiles.Name, unsealFiles.Shorthand, RegisterAsStringArraySetting(&unsealFiles), unsealFiles.Usage)
	fs.StringSliceP(browserRealms.Name, browserRealms.Shorthand, RegisterAsStringSliceSetting(&browserRealms), browserRealms.Usage)
	fs.StringSliceP(serviceRealms.Name, serviceRealms.Shorthand, RegisterAsStringSliceSetting(&serviceRealms), serviceRealms.Usage)

	err := fs.Parse(subCommandParameters)
	if err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	err = viper.BindPFlags(fs)
	if err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Enable environment variable support for all configuration settings.
	// Environment variables use CRYPTOUTIL_ prefix with underscores instead of hyphens.
	// Example: CRYPTOUTIL_DATABASE_URL overrides --database-url flag.
	// Precedence: flags > env vars > config files > defaults
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CRYPTOUTIL")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	configFiles := viper.GetStringSlice(configFile.Name)
	if len(configFiles) > 0 {
		// Set the first config file
		if info, err := os.Stat(configFiles[0]); err == nil && !info.IsDir() {
			viper.SetConfigFile(configFiles[0])

			if err := viper.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("error reading config file %s: %w", configFiles[0], err)
			}
		}
		// Merge additional config files
		for i := 1; i < len(configFiles); i++ {
			if info, err := os.Stat(configFiles[i]); err == nil && !info.IsDir() {
				viper.SetConfigFile(configFiles[i])

				if err := viper.MergeInConfig(); err != nil {
					return nil, fmt.Errorf("error merging config file %s: %w", configFiles[i], err)
				}
			}
		}
	}

	// Apply configuration profile if specified
	profileName := viper.GetString(profile.Name)
	if profileName != "" {
		if profileConfig, exists := profiles[profileName]; exists {
			// Apply profile settings (these can be overridden by config files or command line flags)
			for key, value := range profileConfig {
				if !viper.IsSet(key) {
					viper.Set(key, value)
				}
			}
		} else {
			return nil, fmt.Errorf("unknown configuration profile: %s (available: dev, stg, prod, test)", profileName)
		}
	}

	// Parse TLS mode and PEM fields
	tlsPublicModeStr := viper.GetString(tlsPublicMode.Name)
	if tlsPublicModeStr == "" {
		tlsPublicModeStr = string(defaultTLSPublicMode)
	}

	tlsPrivateModeStr := viper.GetString(tlsPrivateMode.Name)
	if tlsPrivateModeStr == "" {
		tlsPrivateModeStr = string(defaultTLSPrivateMode)
	}

	s := &ServiceTemplateServerSettings{
		TLSPublicMode:               TLSMode(tlsPublicModeStr),
		TLSPrivateMode:              TLSMode(tlsPrivateModeStr),
		TLSStaticCertPEM:            getTLSPEMBytes(tlsStaticCertPEM.Name),
		TLSStaticKeyPEM:             getTLSPEMBytes(tlsStaticKeyPEM.Name),
		TLSMixedCACertPEM:           getTLSPEMBytes(tlsMixedCACertPEM.Name),
		TLSMixedCAKeyPEM:            getTLSPEMBytes(tlsMixedCAKeyPEM.Name),
		SubCommand:                  subCommand,
		Help:                        viper.GetBool(help.Name),
		ConfigFile:                  viper.GetStringSlice(configFile.Name),
		LogLevel:                    viper.GetString(logLevel.Name),
		VerboseMode:                 viper.GetBool(verboseMode.Name),
		DevMode:                     viper.GetBool(devMode.Name),
		DemoMode:                    viper.GetBool(demoMode.Name),
		ResetDemoMode:               viper.GetBool(resetDemoMode.Name),
		DryRun:                      viper.GetBool(dryRun.Name),
		Profile:                     viper.GetString(profile.Name),
		BindPublicProtocol:          viper.GetString(bindPublicProtocol.Name),
		BindPublicAddress:           viper.GetString(bindPublicAddress.Name),
		BindPublicPort:              viper.GetUint16(bindPublicPort.Name),
		TLSPublicDNSNames:           viper.GetStringSlice(tlsPublicDNSNames.Name),
		TLSPublicIPAddresses:        viper.GetStringSlice(tlsPublicIPAddresses.Name),
		TLSPrivateDNSNames:          viper.GetStringSlice(tlsPrivateDNSNames.Name),
		TLSPrivateIPAddresses:       viper.GetStringSlice(tlsPrivateIPAddresses.Name),
		BindPrivateProtocol:         viper.GetString(bindPrivateProtocol.Name),
		BindPrivateAddress:          viper.GetString(bindPrivateAddress.Name),
		BindPrivatePort:             viper.GetUint16(bindPrivatePort.Name),
		PublicBrowserAPIContextPath: viper.GetString(publicBrowserAPIContextPath.Name),
		PublicServiceAPIContextPath: viper.GetString(publicServiceAPIContextPath.Name),
		PrivateAdminAPIContextPath:  viper.GetString(privateAdminAPIContextPath.Name),
		CORSAllowedOrigins:          viper.GetStringSlice(corsAllowedOrigins.Name),
		CORSAllowedMethods:          viper.GetStringSlice(corsAllowedMethods.Name),
		CORSAllowedHeaders:          viper.GetStringSlice(corsAllowedHeaders.Name),
		CORSMaxAge:                  viper.GetUint16(corsMaxAge.Name),
		RequestBodyLimit:            viper.GetInt(requestBodyLimit.Name),
		CSRFTokenName:               viper.GetString(csrfTokenName.Name),
		CSRFTokenSameSite:           viper.GetString(csrfTokenSameSite.Name),
		CSRFTokenMaxAge:             viper.GetDuration(csrfTokenMaxAge.Name),
		CSRFTokenCookieSecure:       viper.GetBool(csrfTokenCookieSecure.Name),
		CSRFTokenCookieHTTPOnly:     viper.GetBool(csrfTokenCookieHTTPOnly.Name),
		CSRFTokenCookieSessionOnly:  viper.GetBool(csrfTokenCookieSessionOnly.Name),
		CSRFTokenSingleUseToken:     viper.GetBool(csrfTokenSingleUseToken.Name),
		BrowserIPRateLimit:          viper.GetUint16(browserIPRateLimit.Name),
		ServiceIPRateLimit:          viper.GetUint16(serviceIPRateLimit.Name),
		AllowedIPs:                  viper.GetStringSlice(allowedIps.Name),
		AllowedCIDRs:                viper.GetStringSlice(allowedCidrs.Name),
		DatabaseContainer:           viper.GetString(databaseContainer.Name),
		DatabaseURL:                 viper.GetString(databaseURL.Name),
		DatabaseInitTotalTimeout:    viper.GetDuration(databaseInitTotalTimeout.Name),
		DatabaseInitRetryWait:       viper.GetDuration(databaseInitRetryWait.Name),
		ServerShutdownTimeout:       viper.GetDuration(serverShutdownTimeout.Name),
		OTLPEnabled:                 viper.GetBool(otlpEnabled.Name),
		OTLPConsole:                 viper.GetBool(otlpConsole.Name),
		OTLPService:                 viper.GetString(otlpService.Name),
		OTLPInstance:                viper.GetString(otlpInstance.Name),
		OTLPVersion:                 viper.GetString(otlpVersion.Name),
		OTLPEnvironment:             viper.GetString(otlpEnvironment.Name),
		OTLPHostname:                viper.GetString(otlpHostname.Name),
		OTLPEndpoint:                viper.GetString(otlpEndpoint.Name),
		UnsealMode:                  viper.GetString(unsealMode.Name),
		UnsealFiles:                 viper.GetStringSlice(unsealFiles.Name),
		BrowserRealms:               viper.GetStringSlice(browserRealms.Name),
		ServiceRealms:               viper.GetStringSlice(serviceRealms.Name),
		BrowserSessionCookie:        viper.GetString(browserSessionCookie.Name),
		BrowserSessionAlgorithm:     viper.GetString(browserSessionAlgorithm.Name),
		BrowserSessionJWSAlgorithm:  viper.GetString(browserSessionJWSAlgorithm.Name),
		BrowserSessionJWEAlgorithm:  viper.GetString(browserSessionJWEAlgorithm.Name),
		BrowserSessionExpiration:    viper.GetDuration(browserSessionExpiration.Name),
		ServiceSessionAlgorithm:     viper.GetString(serviceSessionAlgorithm.Name),
		ServiceSessionJWSAlgorithm:  viper.GetString(serviceSessionJWSAlgorithm.Name),
		ServiceSessionJWEAlgorithm:  viper.GetString(serviceSessionJWEAlgorithm.Name),
		ServiceSessionExpiration:    viper.GetDuration(serviceSessionExpiration.Name),
		SessionIdleTimeout:          viper.GetDuration(sessionIdleTimeout.Name),
		SessionCleanupInterval:      viper.GetDuration(sessionCleanupInterval.Name),
	}

	// Resolve file:// URLs for sensitive settings from Docker secrets or Kubernetes secrets.
	// This allows configuration to reference secret files rather than embedding sensitive values directly.
	s.DatabaseURL = resolveFileURL(s.DatabaseURL)

	logSettings(s)

	if s.Help {
		fs.SetOutput(os.Stdout)
		fmt.Println("cryptoutil - Cryptographic utility server")
		fmt.Println()
		fmt.Println("USAGE:")
		fmt.Println("  cryptoutil [subcommand] [flags]")
		fmt.Println()
		fmt.Println("SUBCOMMANDS:")
		fmt.Println("  start    Start the server")
		fmt.Println("  stop     Stop the server")
		fmt.Println("  init     Initialize the server")
		fmt.Println("  live     Check server liveness")
		fmt.Println("  ready    Check server readiness")
		fmt.Println()
		fmt.Println("CONFIGURATION SETTINGS:")
		fmt.Println("  -d, --dev                           run in development mode; enables in-memory SQLite")
		fmt.Println("  -h, --help                          print help")
		fmt.Println("  -y, --config strings                path to config file (can be specified multiple times)")
		fmt.Println("  -Y, --dry-run                       validate configuration and exit without starting server")
		fmt.Println("  -P, --profile strings                configuration profile: dev, stg, prod, test")
		fmt.Println()
		fmt.Println("DATABASE SETTINGS:")
		fmt.Println("  -u, --database-url string           database URL (default " + formatDefault(defaultDatabaseURL) + ")")
		fmt.Println("  -D, --database-container string     database container mode (default " + formatDefault(defaultDatabaseContainer) + ")")
		fmt.Println("  -Z, --database-init-total-timeout duration database init total timeout (default " + formatDefault(defaultDatabaseInitTotalTimeout) + ")")
		fmt.Println("  -W, --database-init-retry-wait duration database init retry wait (default " + formatDefault(defaultDatabaseInitRetryWait) + ")")
		fmt.Println()
		fmt.Println("SERVER SETTINGS:")
		fmt.Println("  -a, --bind-public-address string    bind public address (default " + formatDefault(defaultBindPublicAddress) + ")")
		fmt.Println("  -p, --bind-public-port uint16       bind public port (default " + formatDefault(defaultBindPublicPort) + ")")
		fmt.Println("  -t, --bind-public-protocol string   bind public protocol (http or https) (default " + formatDefault(defaultBindPublicProtocol) + ")")
		fmt.Println("  -A, --bind-private-address string   bind private address (default " + formatDefault(defaultBindPrivateAddress) + ")")
		fmt.Println("  -P, --bind-private-port uint16      bind private port (default " + formatDefault(defaultBindPrivatePort) + ")")
		fmt.Println("  -T, --bind-private-protocol string  bind private protocol (http or https) (default " + formatDefault(defaultBindPrivateProtocol) + ")")
		fmt.Println("  -c, --browser-api-context-path string  context path for Public Browser API (default " + formatDefault(defaultPublicBrowserAPIContextPath) + ")")
		fmt.Println("  -b, --service-api-context-path string  context path for Public Service API (default " + formatDefault(defaultPublicServiceAPIContextPath) + ")")
		fmt.Println()
		fmt.Println("NETWORK SECURITY SETTINGS:")
		fmt.Println("  -I, --allowed-ips strings           comma-separated list of allowed IPs (default " + formatDefault(defaultAllowedIps) + ")")
		fmt.Println("  -C, --allowed-cidrs strings         comma-separated list of allowed CIDRs (default " + formatDefault(defaultAllowedCIDRs) + ")")
		fmt.Println("  -e, --browser-rate-limit uint16     rate limit for browser API requests per second (default " + formatDefault(defaultBrowserIPRateLimit) + ")")
		fmt.Println("  -w, --service-rate-limit uint16     rate limit for service API requests per second (default " + formatDefault(defaultServiceIPRateLimit) + ")")
		fmt.Println("  -L, --request-body-limit int        Maximum request body size in bytes (default " + formatDefault(defaultRequestBodyLimit) + ")")
		fmt.Println()
		fmt.Println("SWAGGER UI SETTINGS:")
		fmt.Println("      --swagger-ui-username string    username for Swagger UI basic authentication")
		fmt.Println("      --swagger-ui-password string    password for Swagger UI basic authentication")
		fmt.Println()
		fmt.Println("BROWSER CORS SECURITY SETTINGS:")
		fmt.Println("  -o, --cors-origins strings          CORS allowed origins")
		fmt.Println("  -m, --cors-methods strings          CORS allowed methods (default " + formatDefault(defaultCORSAllowedMethods) + ")")
		fmt.Println("  -H, --cors-headers strings          CORS allowed headers (default " + formatDefault(defaultCORSAllowedHeaders) + ")")
		fmt.Println("  -x, --cors-max-age uint16           CORS max age in seconds (default " + formatDefault(defaultCORSMaxAge) + ")")
		fmt.Println()
		fmt.Println("BROWSER CSRF SECURITY SETTINGS:")
		fmt.Println("  -N, --csrf-token-name string        CSRF token name (default " + formatDefault(defaultCSRFTokenName) + ")")
		fmt.Println("  -S, --csrf-token-same-site string   CSRF token SameSite attribute (default " + formatDefault(defaultCSRFTokenSameSite) + ")")
		fmt.Println("  -M, --csrf-token-max-age duration   CSRF token max age (expiration) (default " + formatDefault(defaultCSRFTokenMaxAge) + ")")
		fmt.Println("  -R, --csrf-token-cookie-secure      CSRF token cookie Secure attribute (default " + formatDefault(defaultCSRFTokenCookieSecure) + ")")
		fmt.Println("  -J, --csrf-token-cookie-http-only   CSRF token cookie HttpOnly attribute (default " + formatDefault(defaultCSRFTokenCookieHTTPOnly) + ")")
		fmt.Println("  -E, --csrf-token-cookie-session-only CSRF token cookie SessionOnly attribute (default " + formatDefault(defaultCSRFTokenCookieSessionOnly) + ")")
		fmt.Println("  -G, --csrf-token-single-use-token   CSRF token SingleUse attribute (default " + formatDefault(defaultCSRFTokenSingleUseToken) + ")")
		fmt.Println()
		fmt.Println("TLS SECURITY SETTINGS:")
		fmt.Println("  -n, --tls-public-dns-names strings  TLS public DNS names (default " + formatDefault(defaultTLSPublicDNSNames) + ")")
		fmt.Println("  -i, --tls-public-ip-addresses strings TLS public IP addresses (default " + formatDefault(defaultTLSPublicIPAddresses) + ")")
		fmt.Println("  -j, --tls-private-dns-names strings TLS private DNS names (default " + formatDefault(defaultTLSPrivateDNSNames) + ")")
		fmt.Println("  -k, --tls-private-ip-addresses strings TLS private IP addresses (default " + formatDefault(defaultTLSPrivateIPAddresses) + ")")
		fmt.Println()
		fmt.Println("BARRIER ENCRYPTION SECURITY SETTINGS:")
		fmt.Println("  -U, --unseal-mode string            unseal mode: N, M-of-N, sysinfo (default " + formatDefault(defaultUnsealMode) + ")")
		fmt.Println("  -F, --unseal-files strings          unseal files")
		fmt.Println()
		fmt.Println("OBSERVABILITY SETTINGS:")
		fmt.Println("  -l, --log-level string              log level: ALL, TRACE, DEBUG, CONFIG, INFO, NOTICE, WARN, ERROR, FATAL, OFF (default " + formatDefault(defaultLogLevel) + ")")
		fmt.Println("  -v, --verbose                       verbose modifier for log level")
		fmt.Println("  -z, --otlp                          enable OTLP export")
		fmt.Println("  -q, --otlp-console                  enable OTLP logging to console (STDOUT)")
		fmt.Println("  -s, --otlp-service string           OTLP service (default " + formatDefault(defaultOTLPService) + ")")
		fmt.Println("  -B, --otlp-version string           OTLP version (default " + formatDefault(defaultOTLPVersion) + ")")
		fmt.Println("  -I, --otlp-instance string          OTLP instance id (default " + formatDefault(defaultOTLPInstance) + ")")
		fmt.Println("  -K, --otlp-environment string       OTLP environment (default " + formatDefault(defaultOTLPEnvironment) + ")")
		fmt.Println("  -O, --otlp-hostname string          OTLP hostname (default " + formatDefault(defaultOTLPHostname) + ")")
		fmt.Println("  -Q, --otlp-endpoint string          OTLP endpoint (default " + formatDefault(defaultOTLPEndpoint) + ")")
		fmt.Println()
		fmt.Println("ENVIRONMENT VARIABLES:")
		fmt.Println("  All flags can be set via environment variables using the CRYPTOUTIL_ prefix.")
		fmt.Println("  Examples: CRYPTOUTIL_LOG_LEVEL=DEBUG, CRYPTOUTIL_DATABASE_URL=...")
		fmt.Println()
		fmt.Println("Quickstart Examples:")
		fmt.Println("  kms cryptoutil server start --d                              Start server with in-memory SQLite")
		fmt.Println("  kms cryptoutil server stop  --d                              Stop server")
		fmt.Println("  kms cryptoutil server start --D required                     Start server with PostgreSQL container")
		fmt.Println("  kms cryptoutil server start --y global.yml --y preprod.yml   Start server with settings in YAML config files")
		fmt.Println("  kms cryptoutil server start --Y --y config.yml               Validate configuration without starting")
		fmt.Println("  kms cryptoutil server stop                                   Stop server")

		if exitIfHelp {
			os.Exit(0)
		}
	}

	// Validate configuration before returning
	if err := validateConfiguration(s); err != nil {
		return nil, err
	}

	return s, nil
}

// Parse parses command parameters using the global pflag.CommandLine FlagSet.
// This is the standard entry point for production use maintaining backward compatibility.
// For benchmark testing, use ParseWithFlagSet with a fresh FlagSet to avoid "flag redefined" panics.
func Parse(commandParameters []string, exitIfHelp bool) (*ServiceTemplateServerSettings, error) {
	return ParseWithFlagSet(pflag.CommandLine, commandParameters, exitIfHelp)
}
