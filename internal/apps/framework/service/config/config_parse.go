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

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// getTLSPEMBytes safely retrieves PEM bytes from a viper instance for BytesBase64 flags.
// Returns nil if the value is not set or cannot be converted to []byte.
func getTLSPEMBytes(v *viper.Viper, key string) []byte {
	val := v.Get(key)
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
// ParseWithFlagSet parses command parameters into ServiceFrameworkServerSettings using a custom FlagSet.
// This function enables benchmark testing by accepting a fresh FlagSet for each iteration,
// avoiding pflag's "flag redefined" panics when the same flags are registered multiple times.
//
// Parameters:
//   - fs: Custom FlagSet to register flags on (use pflag.NewFlagSet() for benchmarks, pflag.CommandLine for production)
//   - commandParameters: Command line arguments (first element is subcommand, rest are flags)
//   - exitIfHelp: If true, os.Exit(0) when --help flag is set
//
// Returns:
//   - *ServiceFrameworkServerSettings: Parsed configuration settings
//   - error: Validation or parsing errors
func ParseWithFlagSet(fs *pflag.FlagSet, commandParameters []string, exitIfHelp bool) (*ServiceFrameworkServerSettings, error) {
	if len(commandParameters) == 0 {
		return nil, fmt.Errorf("missing subcommand: use \"start\", \"stop\", \"init\", \"live\", or \"ready\"")
	}

	subCommand := commandParameters[0]
	if _, ok := subcommands[subCommand]; !ok {
		return nil, fmt.Errorf("invalid subcommand: use \"start\", \"stop\", \"init\", \"live\", or \"ready\"")
	}

	subCommandParameters := commandParameters[1:]

	// Create a viper instance per call to prevent global state contamination between concurrent callers.
	v := viper.New()

	// Enable environment variable support with CRYPTOUTIL_ prefix BEFORE parsing flags.
	v.SetEnvPrefix("CRYPTOUTIL")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// Explicitly bind boolean environment variables (viper.AutomaticEnv may not handle booleans correctly)
	// Note: v.BindEnv errors are logged but don't prevent startup as they are extremely rare
	for _, setting := range allServiceFrameworkServerRegisteredSettings {
		if _, ok := setting.Value.(bool); ok {
			if err := v.BindEnv(setting.Name, setting.Env); err != nil {
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
	fs.StringP(browserSessionCookie.Name, browserSessionCookie.Shorthand, RegisterAsStringSetting(&browserSessionCookie), browserSessionCookie.Usage)
	fs.StringP(browserSessionAlgorithm.Name, browserSessionAlgorithm.Shorthand, RegisterAsStringSetting(&browserSessionAlgorithm), browserSessionAlgorithm.Usage)
	fs.StringP(browserSessionJWSAlgorithm.Name, browserSessionJWSAlgorithm.Shorthand, RegisterAsStringSetting(&browserSessionJWSAlgorithm), browserSessionJWSAlgorithm.Usage)
	fs.StringP(browserSessionJWEAlgorithm.Name, browserSessionJWEAlgorithm.Shorthand, RegisterAsStringSetting(&browserSessionJWEAlgorithm), browserSessionJWEAlgorithm.Usage)
	fs.DurationP(browserSessionExpiration.Name, browserSessionExpiration.Shorthand, RegisterAsDurationSetting(&browserSessionExpiration), browserSessionExpiration.Usage)
	fs.StringP(serviceSessionAlgorithm.Name, serviceSessionAlgorithm.Shorthand, RegisterAsStringSetting(&serviceSessionAlgorithm), serviceSessionAlgorithm.Usage)
	fs.StringP(serviceSessionJWSAlgorithm.Name, serviceSessionJWSAlgorithm.Shorthand, RegisterAsStringSetting(&serviceSessionJWSAlgorithm), serviceSessionJWSAlgorithm.Usage)
	fs.StringP(serviceSessionJWEAlgorithm.Name, serviceSessionJWEAlgorithm.Shorthand, RegisterAsStringSetting(&serviceSessionJWEAlgorithm), serviceSessionJWEAlgorithm.Usage)
	fs.DurationP(serviceSessionExpiration.Name, serviceSessionExpiration.Shorthand, RegisterAsDurationSetting(&serviceSessionExpiration), serviceSessionExpiration.Usage)
	fs.DurationP(sessionIdleTimeout.Name, sessionIdleTimeout.Shorthand, RegisterAsDurationSetting(&sessionIdleTimeout), sessionIdleTimeout.Usage)
	fs.DurationP(sessionCleanupInterval.Name, sessionCleanupInterval.Shorthand, RegisterAsDurationSetting(&sessionCleanupInterval), sessionCleanupInterval.Usage)
	fs.StringP(databaseSSLMode.Name, databaseSSLMode.Shorthand, RegisterAsStringSetting(&databaseSSLMode), databaseSSLMode.Usage)
	fs.StringP(databaseSSLCert.Name, databaseSSLCert.Shorthand, RegisterAsStringSetting(&databaseSSLCert), databaseSSLCert.Usage)
	fs.StringP(databaseSSLKey.Name, databaseSSLKey.Shorthand, RegisterAsStringSetting(&databaseSSLKey), databaseSSLKey.Usage)
	fs.StringP(databaseSSLRootCert.Name, databaseSSLRootCert.Shorthand, RegisterAsStringSetting(&databaseSSLRootCert), databaseSSLRootCert.Usage)

	err := fs.Parse(subCommandParameters)
	if err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	err = v.BindPFlags(fs)
	if err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Enable environment variable support for all configuration settings.
	// Environment variables use CRYPTOUTIL_ prefix with underscores instead of hyphens.
	// Example: CRYPTOUTIL_DATABASE_URL overrides --database-url flag.
	// Precedence: flags > env vars > config files > defaults
	v.AutomaticEnv()
	v.SetEnvPrefix("CRYPTOUTIL")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	configFiles := v.GetStringSlice(configFile.Name)
	if len(configFiles) > 0 {
		// Set the first config file
		if info, err := os.Stat(configFiles[0]); err == nil && !info.IsDir() {
			v.SetConfigFile(configFiles[0])

			if err := v.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("error reading config file %s: %w", configFiles[0], err)
			}
		}
		// Merge additional config files
		for i := 1; i < len(configFiles); i++ {
			if info, err := os.Stat(configFiles[i]); err == nil && !info.IsDir() {
				v.SetConfigFile(configFiles[i])

				if err := v.MergeInConfig(); err != nil {
					return nil, fmt.Errorf("error merging config file %s: %w", configFiles[i], err)
				}
			}
		}
	}

	// Apply configuration profile if specified
	profileName := v.GetString(profile.Name)
	if profileName != "" {
		if profileConfig, exists := profiles[profileName]; exists {
			// Apply profile settings (these can be overridden by config files or command line flags)
			for key, value := range profileConfig {
				if !v.IsSet(key) {
					v.Set(key, value)
				}
			}
		} else {
			return nil, fmt.Errorf("unknown configuration profile: %s (available: local, stg, prod, test)", profileName)
		}
	}

	// Parse TLS mode and PEM fields
	tlsPublicModeStr := v.GetString(tlsPublicMode.Name)
	if tlsPublicModeStr == "" {
		tlsPublicModeStr = string(defaultTLSPublicMode)
	}

	tlsPrivateModeStr := v.GetString(tlsPrivateMode.Name)
	if tlsPrivateModeStr == "" {
		tlsPrivateModeStr = string(defaultTLSPrivateMode)
	}

	s := &ServiceFrameworkServerSettings{
		TLSPublicMode:               TLSMode(tlsPublicModeStr),
		TLSPrivateMode:              TLSMode(tlsPrivateModeStr),
		TLSStaticCertPEM:            getTLSPEMBytes(v, tlsStaticCertPEM.Name),
		TLSStaticKeyPEM:             getTLSPEMBytes(v, tlsStaticKeyPEM.Name),
		TLSMixedCACertPEM:           getTLSPEMBytes(v, tlsMixedCACertPEM.Name),
		TLSMixedCAKeyPEM:            getTLSPEMBytes(v, tlsMixedCAKeyPEM.Name),
		SubCommand:                  subCommand,
		Help:                        v.GetBool(help.Name),
		ConfigFile:                  v.GetStringSlice(configFile.Name),
		LogLevel:                    v.GetString(logLevel.Name),
		VerboseMode:                 v.GetBool(verboseMode.Name),
		DevMode:                     v.GetBool(devMode.Name),
		DryRun:                      v.GetBool(dryRun.Name),
		Profile:                     v.GetString(profile.Name),
		BindPublicProtocol:          v.GetString(bindPublicProtocol.Name),
		BindPublicAddress:           v.GetString(bindPublicAddress.Name),
		BindPublicPort:              v.GetUint16(bindPublicPort.Name),
		TLSPublicDNSNames:           v.GetStringSlice(tlsPublicDNSNames.Name),
		TLSPublicIPAddresses:        v.GetStringSlice(tlsPublicIPAddresses.Name),
		TLSPrivateDNSNames:          v.GetStringSlice(tlsPrivateDNSNames.Name),
		TLSPrivateIPAddresses:       v.GetStringSlice(tlsPrivateIPAddresses.Name),
		BindPrivateProtocol:         v.GetString(bindPrivateProtocol.Name),
		BindPrivateAddress:          v.GetString(bindPrivateAddress.Name),
		BindPrivatePort:             v.GetUint16(bindPrivatePort.Name),
		PublicBrowserAPIContextPath: v.GetString(publicBrowserAPIContextPath.Name),
		PublicServiceAPIContextPath: v.GetString(publicServiceAPIContextPath.Name),
		PrivateAdminAPIContextPath:  v.GetString(privateAdminAPIContextPath.Name),
		CORSAllowedOrigins:          v.GetStringSlice(corsAllowedOrigins.Name),
		CORSAllowedMethods:          v.GetStringSlice(corsAllowedMethods.Name),
		CORSAllowedHeaders:          v.GetStringSlice(corsAllowedHeaders.Name),
		CORSMaxAge:                  v.GetUint16(corsMaxAge.Name),
		RequestBodyLimit:            v.GetInt(requestBodyLimit.Name),
		CSRFTokenName:               v.GetString(csrfTokenName.Name),
		CSRFTokenSameSite:           v.GetString(csrfTokenSameSite.Name),
		CSRFTokenMaxAge:             v.GetDuration(csrfTokenMaxAge.Name),
		CSRFTokenCookieSecure:       v.GetBool(csrfTokenCookieSecure.Name),
		CSRFTokenCookieHTTPOnly:     v.GetBool(csrfTokenCookieHTTPOnly.Name),
		CSRFTokenCookieSessionOnly:  v.GetBool(csrfTokenCookieSessionOnly.Name),
		CSRFTokenSingleUseToken:     v.GetBool(csrfTokenSingleUseToken.Name),
		BrowserIPRateLimit:          v.GetUint16(browserIPRateLimit.Name),
		ServiceIPRateLimit:          v.GetUint16(serviceIPRateLimit.Name),
		AllowedIPs:                  v.GetStringSlice(allowedIps.Name),
		AllowedCIDRs:                v.GetStringSlice(allowedCidrs.Name),
		DatabaseContainer:           v.GetString(databaseContainer.Name),
		DatabaseURL:                 v.GetString(databaseURL.Name),
		DatabaseInitTotalTimeout:    v.GetDuration(databaseInitTotalTimeout.Name),
		DatabaseInitRetryWait:       v.GetDuration(databaseInitRetryWait.Name),
		ServerShutdownTimeout:       v.GetDuration(serverShutdownTimeout.Name),
		OTLPEnabled:                 v.GetBool(otlpEnabled.Name),
		OTLPConsole:                 v.GetBool(otlpConsole.Name),
		OTLPService:                 v.GetString(otlpService.Name),
		OTLPInstance:                v.GetString(otlpInstance.Name),
		OTLPVersion:                 v.GetString(otlpVersion.Name),
		OTLPEnvironment:             v.GetString(otlpEnvironment.Name),
		OTLPHostname:                v.GetString(otlpHostname.Name),
		OTLPEndpoint:                v.GetString(otlpEndpoint.Name),
		UnsealMode:                  v.GetString(unsealMode.Name),
		UnsealFiles:                 v.GetStringSlice(unsealFiles.Name),
		BrowserRealms:               v.GetStringSlice(browserRealms.Name),
		ServiceRealms:               v.GetStringSlice(serviceRealms.Name),
		BrowserSessionCookie:        v.GetString(browserSessionCookie.Name),
		BrowserSessionAlgorithm:     v.GetString(browserSessionAlgorithm.Name),
		BrowserSessionJWSAlgorithm:  v.GetString(browserSessionJWSAlgorithm.Name),
		BrowserSessionJWEAlgorithm:  v.GetString(browserSessionJWEAlgorithm.Name),
		BrowserSessionExpiration:    v.GetDuration(browserSessionExpiration.Name),
		ServiceSessionAlgorithm:     v.GetString(serviceSessionAlgorithm.Name),
		ServiceSessionJWSAlgorithm:  v.GetString(serviceSessionJWSAlgorithm.Name),
		ServiceSessionJWEAlgorithm:  v.GetString(serviceSessionJWEAlgorithm.Name),
		ServiceSessionExpiration:    v.GetDuration(serviceSessionExpiration.Name),
		SessionIdleTimeout:          v.GetDuration(sessionIdleTimeout.Name),
		SessionCleanupInterval:      v.GetDuration(sessionCleanupInterval.Name),
		DatabaseSSLMode:             v.GetString(databaseSSLMode.Name),
		DatabaseSSLCert:             v.GetString(databaseSSLCert.Name),
		DatabaseSSLKey:              v.GetString(databaseSSLKey.Name),
		DatabaseSSLRootCert:         v.GetString(databaseSSLRootCert.Name),
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
		fmt.Println("  -d, --local                         run in development mode; enables in-memory SQLite")
		fmt.Println("  -h, --help                          print help")
		fmt.Println("  -y, --config strings                path to config file (can be specified multiple times)")
		fmt.Println("  -Y, --dry-run                       validate configuration and exit without starting server")
		fmt.Println("  -P, --profile strings                configuration profile: local, stg, prod, test")
		fmt.Println()
		fmt.Println("DATABASE SETTINGS:")
		fmt.Println("  -u, --database-url string           database URL (default " + formatDefault(cryptoutilSharedMagic.DefaultDatabaseURL) + ")")
		fmt.Println("  -D, --database-container string     database container mode (default " + formatDefault(cryptoutilSharedMagic.DefaultDatabaseContainerDisabled) + ")")
		fmt.Println("  -Z, --database-init-total-timeout duration database init total timeout (default " + formatDefault(cryptoutilSharedMagic.DefaultDatabaseInitTotalTimeout) + ")")
		fmt.Println("  -W, --database-init-retry-wait duration database init retry wait (default " + formatDefault(cryptoutilSharedMagic.DefaultDataInitRetryWait) + ")")
		fmt.Println()
		fmt.Println("SERVER SETTINGS:")
		fmt.Println("  -a, --bind-public-address string    bind public address (default " + formatDefault(cryptoutilSharedMagic.DefaultPublicAddressCryptoutil) + ")")
		fmt.Println("  -p, --bind-public-port uint16       bind public port (default " + formatDefault(cryptoutilSharedMagic.DefaultPublicPortCryptoutil) + ")")
		fmt.Println("  -t, --bind-public-protocol string   bind public protocol (http or https) (default " + formatDefault(cryptoutilSharedMagic.DefaultPublicProtocolCryptoutil) + ")")
		fmt.Println("  -A, --bind-private-address string   bind private address (default " + formatDefault(cryptoutilSharedMagic.DefaultPrivateAddressCryptoutil) + ")")
		fmt.Println("  -P, --bind-private-port uint16      bind private port (default " + formatDefault(cryptoutilSharedMagic.DefaultPrivatePortCryptoutil) + ")")
		fmt.Println("  -T, --bind-private-protocol string  bind private protocol (http or https) (default " + formatDefault(cryptoutilSharedMagic.DefaultPrivateProtocolCryptoutil) + ")")
		fmt.Println("  -c, --browser-api-context-path string  context path for Public Browser API (default " + formatDefault(cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath) + ")")
		fmt.Println("  -b, --service-api-context-path string  context path for Public Service API (default " + formatDefault(cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath) + ")")
		fmt.Println()
		fmt.Println("NETWORK SECURITY SETTINGS:")
		fmt.Println("  -I, --allowed-ips strings           comma-separated list of allowed IPs (default " + formatDefault(defaultAllowedIps) + ")")
		fmt.Println("  -C, --allowed-cidrs strings         comma-separated list of allowed CIDRs (default " + formatDefault(defaultAllowedCIDRs) + ")")
		fmt.Println("  -e, --browser-rate-limit uint16     rate limit for browser API requests per second (default " + formatDefault(cryptoutilSharedMagic.DefaultPublicBrowserAPIIPRateLimit) + ")")
		fmt.Println("  -w, --service-rate-limit uint16     rate limit for service API requests per second (default " + formatDefault(cryptoutilSharedMagic.DefaultPublicServiceAPIIPRateLimit) + ")")
		fmt.Println("  -L, --request-body-limit int        Maximum request body size in bytes (default " + formatDefault(cryptoutilSharedMagic.DefaultHTTPRequestBodyLimit) + ")")
		fmt.Println()
		fmt.Println("SWAGGER UI SETTINGS:")
		fmt.Println("      --swagger-ui-username string    username for Swagger UI basic authentication")
		fmt.Println("      --swagger-ui-password string    password for Swagger UI basic authentication")
		fmt.Println()
		fmt.Println("BROWSER CORS SECURITY SETTINGS:")
		fmt.Println("  -o, --cors-origins strings          CORS allowed origins")
		fmt.Println("  -m, --cors-methods strings          CORS allowed methods (default " + formatDefault(defaultCORSAllowedMethods) + ")")
		fmt.Println("  -H, --cors-headers strings          CORS allowed headers (default " + formatDefault(defaultCORSAllowedHeaders) + ")")
		fmt.Println("  -x, --cors-max-age uint16           CORS max age in seconds (default " + formatDefault(cryptoutilSharedMagic.DefaultCORSMaxAge) + ")")
		fmt.Println()
		fmt.Println("BROWSER CSRF SECURITY SETTINGS:")
		fmt.Println("  -N, --csrf-token-name string        CSRF token name (default " + formatDefault(cryptoutilSharedMagic.DefaultCSRFTokenName) + ")")
		fmt.Println("  -S, --csrf-token-same-site string   CSRF token SameSite attribute (default " + formatDefault(cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict) + ")")
		fmt.Println("  -M, --csrf-token-max-age duration   CSRF token max age (expiration) (default " + formatDefault(cryptoutilSharedMagic.DefaultCSRFTokenMaxAge) + ")")
		fmt.Println("  -R, --csrf-token-cookie-secure      CSRF token cookie Secure attribute (default " + formatDefault(cryptoutilSharedMagic.DefaultCSRFTokenCookieSecure) + ")")
		fmt.Println("  -J, --csrf-token-cookie-http-only   CSRF token cookie HttpOnly attribute (default " + formatDefault(cryptoutilSharedMagic.DefaultCSRFTokenCookieHTTPOnly) + ")")
		fmt.Println("  -E, --csrf-token-cookie-session-only CSRF token cookie SessionOnly attribute (default " + formatDefault(cryptoutilSharedMagic.DefaultCSRFTokenCookieSessionOnly) + ")")
		fmt.Println("  -G, --csrf-token-single-use-token   CSRF token SingleUse attribute (default " + formatDefault(cryptoutilSharedMagic.DefaultCSRFTokenSingleUseToken) + ")")
		fmt.Println()
		fmt.Println("TLS SECURITY SETTINGS:")
		fmt.Println("  -n, --tls-public-dns-names strings  TLS public DNS names (default " + formatDefault(defaultTLSPublicDNSNames) + ")")
		fmt.Println("  -i, --tls-public-ip-addresses strings TLS public IP addresses (default " + formatDefault(defaultTLSPublicIPAddresses) + ")")
		fmt.Println("  -j, --tls-private-dns-names strings TLS private DNS names (default " + formatDefault(defaultTLSPrivateDNSNames) + ")")
		fmt.Println("  -k, --tls-private-ip-addresses strings TLS private IP addresses (default " + formatDefault(defaultTLSPrivateIPAddresses) + ")")
		fmt.Println()
		fmt.Println("BARRIER ENCRYPTION SECURITY SETTINGS:")
		fmt.Println("  -U, --unseal-mode string            unseal mode: N, M-of-N, sysinfo (default " + formatDefault(cryptoutilSharedMagic.DefaultUnsealModeSysInfo) + ")")
		fmt.Println("  -F, --unseal-files strings          unseal files")
		fmt.Println()
		fmt.Println("OBSERVABILITY SETTINGS:")
		fmt.Println("  -l, --log-level string              log level: ALL, TRACE, DEBUG, CONFIG, INFO, NOTICE, WARN, ERROR, FATAL, OFF (default " + formatDefault(cryptoutilSharedMagic.DefaultLogLevelInfo) + ")")
		fmt.Println("  -v, --verbose                       verbose modifier for log level")
		fmt.Println("  -z, --otlp                          enable OTLP export")
		fmt.Println("  -q, --otlp-console                  enable OTLP logging to console (STDOUT)")
		fmt.Println("  -s, --otlp-service string           OTLP service (default " + formatDefault(cryptoutilSharedMagic.DefaultOTLPServiceDefault) + ")")
		fmt.Println("  -B, --otlp-version string           OTLP version (default " + formatDefault(cryptoutilSharedMagic.DefaultOTLPVersionDefault) + ")")
		fmt.Println("  -I, --otlp-instance string          OTLP instance id (default " + formatDefault(defaultOTLPInstance) + ")")
		fmt.Println("  -K, --otlp-environment string       OTLP environment (default " + formatDefault(cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault) + ")")
		fmt.Println("  -O, --otlp-hostname string          OTLP hostname (default " + formatDefault(cryptoutilSharedMagic.DefaultOTLPHostnameDefault) + ")")
		fmt.Println("  -Q, --otlp-endpoint string          OTLP endpoint (default " + formatDefault(cryptoutilSharedMagic.DefaultOTLPEndpointDefault) + ")")
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
func Parse(commandParameters []string, exitIfHelp bool) (*ServiceFrameworkServerSettings, error) {
	return ParseWithFlagSet(pflag.CommandLine, commandParameters, exitIfHelp)
}
