// Copyright (c) 2025 Justin Cranford

// Package config provides identity-spa server configuration settings.
package config

import (
	"fmt"
	"os"
	"strings"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// IdentitySPAServerSettings contains identity-spa specific configuration.
type IdentitySPAServerSettings struct {
	*cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings

	// StaticFilesPath is the path to the static files directory.
	// Default: "./static" (relative to working directory).
	StaticFilesPath string

	// IndexFile is the default file to serve for SPA routing.
	// Default: "index.html".
	IndexFile string

	// RPOrigin is the Relying Party (BFF) origin for API proxying configuration.
	// Example: "https://localhost:18300".
	RPOrigin string

	// CacheControlMaxAge is the max-age for Cache-Control header in seconds.
	// Default: 3600 (1 hour) for production, 0 for development.
	CacheControlMaxAge int

	// EnableGzip enables gzip compression for static files.
	// Default: true.
	EnableGzip bool

	// EnableBrotli enables brotli compression for static files.
	// Default: true.
	EnableBrotli bool

	// CSPDirectives is the Content-Security-Policy header value.
	CSPDirectives string
}

// Default values for identity-spa settings.
const (
	defaultStaticFilesPath = "./static"
	defaultIndexFile       = "index.html"
	defaultRPOrigin        = ""   // Optional, can be empty.
	defaultCacheMaxAge     = 3600 // 1 hour for production.
	defaultCacheMaxAgeDev  = 0    // No caching for development.
	defaultEnableGzip      = true // Enable gzip by default.
	defaultEnableBrotli    = true // Enable brotli by default.
	defaultCSPDirectives   = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'"
)

var allIdentitySPAServerRegisteredSettings []*cryptoutilAppsTemplateServiceConfig.Setting //nolint:gochecknoglobals

// Identity-SPA specific Setting objects for parameter attributes.
var (
	staticFilesPathSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allIdentitySPAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "static-files-path",
		Shorthand:   "",
		Value:       defaultStaticFilesPath,
		Usage:       "path to the static files directory",
		Description: "Static Files Path",
	})
	indexFileSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allIdentitySPAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "index-file",
		Shorthand:   "",
		Value:       defaultIndexFile,
		Usage:       "default file to serve for SPA routing",
		Description: "Index File",
	})
	rpOriginSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allIdentitySPAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "rp-origin",
		Shorthand:   "",
		Value:       defaultRPOrigin,
		Usage:       "origin of the Relying Party (BFF) for API proxying",
		Description: "RP Origin",
	})
	cacheControlMaxAgeSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allIdentitySPAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "cache-control-max-age",
		Shorthand:   "",
		Value:       defaultCacheMaxAge,
		Usage:       "max-age for Cache-Control header in seconds",
		Description: "Cache Control Max Age",
	})
	enableGzipSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allIdentitySPAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "enable-gzip",
		Shorthand:   "",
		Value:       defaultEnableGzip,
		Usage:       "enable gzip compression for static files",
		Description: "Enable Gzip",
	})
	enableBrotliSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allIdentitySPAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "enable-brotli",
		Shorthand:   "",
		Value:       defaultEnableBrotli,
		Usage:       "enable brotli compression for static files",
		Description: "Enable Brotli",
	})
	cspDirectivesSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allIdentitySPAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "csp-directives",
		Shorthand:   "",
		Value:       defaultCSPDirectives,
		Usage:       "Content-Security-Policy header value",
		Description: "CSP Directives",
	})
)

// Parse parses command line arguments and returns identity-spa settings.
func Parse(args []string, exitIfHelp bool) (*IdentitySPAServerSettings, error) {
	// Parse base template settings first.
	baseSettings, err := cryptoutilAppsTemplateServiceConfig.Parse(args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	// Register identity-spa specific flags.
	pflag.StringP(staticFilesPathSetting.Name, staticFilesPathSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsStringSetting(staticFilesPathSetting), staticFilesPathSetting.Description)
	pflag.StringP(indexFileSetting.Name, indexFileSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsStringSetting(indexFileSetting), indexFileSetting.Description)
	pflag.StringP(rpOriginSetting.Name, rpOriginSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsStringSetting(rpOriginSetting), rpOriginSetting.Description)
	pflag.IntP(cacheControlMaxAgeSetting.Name, cacheControlMaxAgeSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(cacheControlMaxAgeSetting), cacheControlMaxAgeSetting.Description)
	pflag.BoolP(enableGzipSetting.Name, enableGzipSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsBoolSetting(enableGzipSetting), enableGzipSetting.Description)
	pflag.BoolP(enableBrotliSetting.Name, enableBrotliSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsBoolSetting(enableBrotliSetting), enableBrotliSetting.Description)
	pflag.StringP(cspDirectivesSetting.Name, cspDirectivesSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsStringSetting(cspDirectivesSetting), cspDirectivesSetting.Description)

	// Parse flags.
	pflag.Parse()

	// Bind flags to viper.
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Create identity-spa settings.
	settings := &IdentitySPAServerSettings{
		ServiceTemplateServerSettings: baseSettings,
		StaticFilesPath:               viper.GetString(staticFilesPathSetting.Name),
		IndexFile:                     viper.GetString(indexFileSetting.Name),
		RPOrigin:                      viper.GetString(rpOriginSetting.Name),
		CacheControlMaxAge:            viper.GetInt(cacheControlMaxAgeSetting.Name),
		EnableGzip:                    viper.GetBool(enableGzipSetting.Name),
		EnableBrotli:                  viper.GetBool(enableBrotliSetting.Name),
		CSPDirectives:                 viper.GetString(cspDirectivesSetting.Name),
	}

	// Override template defaults with identity-spa specific values.
	// NOTE: Only override public port - private admin port (9090) is universal across all services.
	settings.BindPublicPort = cryptoutilSharedMagic.IdentitySPAServicePort
	settings.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentitySPA

	// Adjust cache control for dev mode.
	if settings.DevMode {
		settings.CacheControlMaxAge = defaultCacheMaxAgeDev
	}

	// Validate identity-spa specific settings.
	if err := validateIdentitySPASettings(settings); err != nil {
		return nil, fmt.Errorf("identity-spa settings validation failed: %w", err)
	}

	// Log identity-spa specific settings.
	logIdentitySPASettings(settings)

	return settings, nil
}

// validateIdentitySPASettings validates identity-spa specific configuration.
func validateIdentitySPASettings(s *IdentitySPAServerSettings) error {
	var validationErrors []string

	// StaticFilesPath is required.
	if s.StaticFilesPath == "" {
		validationErrors = append(validationErrors, "static-files-path is required")
	}

	// IndexFile is required.
	if s.IndexFile == "" {
		validationErrors = append(validationErrors, "index-file is required")
	}

	// Validate RP origin format if specified.
	if s.RPOrigin != "" && !strings.HasPrefix(s.RPOrigin, "http://") && !strings.HasPrefix(s.RPOrigin, "https://") {
		validationErrors = append(validationErrors, "rp-origin must start with http:// or https://")
	}

	// CacheControlMaxAge cannot be negative.
	if s.CacheControlMaxAge < 0 {
		validationErrors = append(validationErrors, "cache-control-max-age cannot be negative")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// logIdentitySPASettings logs identity-spa specific configuration to stderr.
func logIdentitySPASettings(s *IdentitySPAServerSettings) {
	fmt.Fprintf(os.Stderr, "Identity-SPA Server Settings:\n")
	fmt.Fprintf(os.Stderr, "  Public Server: %s\n", s.PublicBaseURL())
	fmt.Fprintf(os.Stderr, "  Private Server: %s\n", s.PrivateBaseURL())
	fmt.Fprintf(os.Stderr, "  OTLP Service: %s\n", s.OTLPService)
	fmt.Fprintf(os.Stderr, "  Static Files Path: %s\n", s.StaticFilesPath)
	fmt.Fprintf(os.Stderr, "  Index File: %s\n", s.IndexFile)
	fmt.Fprintf(os.Stderr, "  RP Origin: %s\n", s.RPOrigin)
	fmt.Fprintf(os.Stderr, "  Cache Control Max Age: %d\n", s.CacheControlMaxAge)
	fmt.Fprintf(os.Stderr, "  Enable Gzip: %t\n", s.EnableGzip)
	fmt.Fprintf(os.Stderr, "  Enable Brotli: %t\n", s.EnableBrotli)
}

// NewTestConfig creates an IdentitySPAServerSettings instance for testing without calling Parse().
// This bypasses pflag's global FlagSet to allow multiple config creations in tests.
//
// Use this in tests instead of Parse() to avoid "flag redefined" panics
// when creating multiple server instances.
//
// Parameters:
//   - bindAddr: public bind address (typically cryptoutilSharedMagic.IPv4Loopback).
//   - bindPort: public bind port (use 0 for dynamic allocation).
//   - devMode: enable development mode (in-memory SQLite, relaxed security).
//
// Returns directly populated IdentitySPAServerSettings matching Parse() behavior.
func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *IdentitySPAServerSettings {
	// Get base template config.
	baseConfig := cryptoutilAppsTemplateServiceConfig.NewTestConfig(bindAddr, bindPort, devMode)

	// Override template defaults with identity-spa specific values.
	baseConfig.BindPublicPort = bindPort
	baseConfig.OTLPService = cryptoutilSharedMagic.OTLPServiceIdentitySPA

	// Determine cache max age based on mode.
	cacheMaxAge := defaultCacheMaxAge
	if devMode {
		cacheMaxAge = defaultCacheMaxAgeDev
	}

	return &IdentitySPAServerSettings{
		ServiceTemplateServerSettings: baseConfig,
		StaticFilesPath:               defaultStaticFilesPath,
		IndexFile:                     defaultIndexFile,
		RPOrigin:                      defaultRPOrigin,
		CacheControlMaxAge:            cacheMaxAge,
		EnableGzip:                    false, // Disable compression for easier test assertions.
		EnableBrotli:                  false,
		CSPDirectives:                 defaultCSPDirectives,
	}
}

// DefaultTestConfig creates a default test configuration suitable for most unit tests.
// Uses loopback address, dynamic port allocation, and dev mode.
func DefaultTestConfig() *IdentitySPAServerSettings {
	return NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}
