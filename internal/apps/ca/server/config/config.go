// Copyright (c) 2025 Justin Cranford

// Package config provides pki-ca server configuration settings.
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

// CAServerSettings contains pki-ca specific configuration.
type CAServerSettings struct {
	*cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings

	// CA-specific settings.
	CAConfigPath    string // Path to CA definition YAML.
	ProfilesPath    string // Path to certificate profiles directory.
	EnableEST       bool   // Enable EST (Enrollment over Secure Transport) endpoints.
	EnableOCSP      bool   // Enable OCSP responder.
	EnableCRL       bool   // Enable CRL distribution point.
	EnableTimestamp bool   // Enable time-stamping service.
}

// CA-specific default values.
const (
	defaultCAConfigPath    = ""
	defaultProfilesPath    = ""
	defaultEnableEST       = true
	defaultEnableOCSP      = true
	defaultEnableCRL       = true
	defaultEnableTimestamp = false
)

var allCAServerRegisteredSettings []*cryptoutilAppsTemplateServiceConfig.Setting //nolint:gochecknoglobals

// CA-specific Setting objects for parameter attributes.
var (
	caConfigPathSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allCAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "ca-config",
		Shorthand:   "",
		Value:       defaultCAConfigPath,
		Usage:       "path to CA definition YAML file",
		Description: "CA Config Path",
	})
	profilesPathSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allCAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "profiles-path",
		Shorthand:   "",
		Value:       defaultProfilesPath,
		Usage:       "path to certificate profiles directory",
		Description: "Profiles Path",
	})
	enableESTSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allCAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "enable-est",
		Shorthand:   "",
		Value:       defaultEnableEST,
		Usage:       "enable EST (Enrollment over Secure Transport) endpoints",
		Description: "Enable EST",
	})
	enableOCSPSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allCAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "enable-ocsp",
		Shorthand:   "",
		Value:       defaultEnableOCSP,
		Usage:       "enable OCSP responder",
		Description: "Enable OCSP",
	})
	enableCRLSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allCAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "enable-crl",
		Shorthand:   "",
		Value:       defaultEnableCRL,
		Usage:       "enable CRL distribution point",
		Description: "Enable CRL",
	})
	enableTimestampSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allCAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "enable-timestamp",
		Shorthand:   "",
		Value:       defaultEnableTimestamp,
		Usage:       "enable time-stamping service",
		Description: "Enable Timestamp",
	})
)

// Parse parses command line arguments and returns pki-ca settings.
func Parse(args []string, exitIfHelp bool) (*CAServerSettings, error) {
	// Parse base template settings first.
	baseSettings, err := cryptoutilAppsTemplateServiceConfig.Parse(args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	// Register pki-ca specific flags.
	pflag.StringP(caConfigPathSetting.Name, caConfigPathSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsStringSetting(caConfigPathSetting), caConfigPathSetting.Description)
	pflag.StringP(profilesPathSetting.Name, profilesPathSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsStringSetting(profilesPathSetting), profilesPathSetting.Description)
	pflag.BoolP(enableESTSetting.Name, enableESTSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsBoolSetting(enableESTSetting), enableESTSetting.Description)
	pflag.BoolP(enableOCSPSetting.Name, enableOCSPSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsBoolSetting(enableOCSPSetting), enableOCSPSetting.Description)
	pflag.BoolP(enableCRLSetting.Name, enableCRLSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsBoolSetting(enableCRLSetting), enableCRLSetting.Description)
	pflag.BoolP(enableTimestampSetting.Name, enableTimestampSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsBoolSetting(enableTimestampSetting), enableTimestampSetting.Description)

	// Parse flags.
	pflag.Parse()

	// Bind flags to viper.
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Create pki-ca settings.
	settings := &CAServerSettings{
		ServiceTemplateServerSettings: baseSettings,
		CAConfigPath:                  viper.GetString(caConfigPathSetting.Name),
		ProfilesPath:                  viper.GetString(profilesPathSetting.Name),
		EnableEST:                     viper.GetBool(enableESTSetting.Name),
		EnableOCSP:                    viper.GetBool(enableOCSPSetting.Name),
		EnableCRL:                     viper.GetBool(enableCRLSetting.Name),
		EnableTimestamp:               viper.GetBool(enableTimestampSetting.Name),
	}

	// Override template defaults with pki-ca specific values.
	// NOTE: Only override public port - private admin port (9090) is universal across all services.
	settings.BindPublicPort = cryptoutilSharedMagic.PKICAServicePort
	settings.OTLPService = cryptoutilSharedMagic.OTLPServicePKICA

	// Validate pki-ca specific settings.
	if err := validateCASettings(settings); err != nil {
		return nil, fmt.Errorf("pki-ca settings validation failed: %w", err)
	}

	// Log pki-ca specific settings.
	logCASettings(settings)

	return settings, nil
}

// validateCASettings validates pki-ca specific configuration.
func validateCASettings(s *CAServerSettings) error {
	var validationErrors []string

	// Validate CA config path if specified.
	if s.CAConfigPath != "" {
		if _, err := os.Stat(s.CAConfigPath); os.IsNotExist(err) {
			validationErrors = append(validationErrors, fmt.Sprintf("ca-config file does not exist: %s", s.CAConfigPath))
		}
	}

	// Validate profiles path if specified.
	if s.ProfilesPath != "" {
		if info, err := os.Stat(s.ProfilesPath); os.IsNotExist(err) {
			validationErrors = append(validationErrors, fmt.Sprintf("profiles-path does not exist: %s", s.ProfilesPath))
		} else if err == nil && !info.IsDir() {
			validationErrors = append(validationErrors, fmt.Sprintf("profiles-path is not a directory: %s", s.ProfilesPath))
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// logCASettings logs pki-ca specific configuration to stderr.
func logCASettings(s *CAServerSettings) {
	fmt.Fprintf(os.Stderr, "PKI-CA Server Settings:\n")
	fmt.Fprintf(os.Stderr, "  Public Server: %s\n", s.PublicBaseURL())
	fmt.Fprintf(os.Stderr, "  Private Server: %s\n", s.PrivateBaseURL())
	fmt.Fprintf(os.Stderr, "  OTLP Service: %s\n", s.OTLPService)
	fmt.Fprintf(os.Stderr, "  Browser Realms: %s\n", strings.Join(s.BrowserRealms, ", "))
	fmt.Fprintf(os.Stderr, "  Service Realms: %s\n", strings.Join(s.ServiceRealms, ", "))
	fmt.Fprintf(os.Stderr, "  CA Config Path: %s\n", s.CAConfigPath)
	fmt.Fprintf(os.Stderr, "  Profiles Path: %s\n", s.ProfilesPath)
	fmt.Fprintf(os.Stderr, "  Enable EST: %t\n", s.EnableEST)
	fmt.Fprintf(os.Stderr, "  Enable OCSP: %t\n", s.EnableOCSP)
	fmt.Fprintf(os.Stderr, "  Enable CRL: %t\n", s.EnableCRL)
	fmt.Fprintf(os.Stderr, "  Enable Timestamp: %t\n", s.EnableTimestamp)
}

// NewTestConfig creates a CAServerSettings instance for testing without calling Parse().
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
// Returns directly populated CAServerSettings matching Parse() behavior.
func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *CAServerSettings {
	// Get base template config.
	baseConfig := cryptoutilAppsTemplateServiceConfig.NewTestConfig(bindAddr, bindPort, devMode)

	// Override template defaults with pki-ca specific values.
	baseConfig.BindPublicPort = bindPort
	baseConfig.OTLPService = cryptoutilSharedMagic.OTLPServicePKICA

	return &CAServerSettings{
		ServiceTemplateServerSettings: baseConfig,
		CAConfigPath:                  "",
		ProfilesPath:                  "",
		EnableEST:                     defaultEnableEST,
		EnableOCSP:                    defaultEnableOCSP,
		EnableCRL:                     defaultEnableCRL,
		EnableTimestamp:               defaultEnableTimestamp,
	}
}

// DefaultTestConfig creates a default test configuration suitable for most unit tests.
// Uses loopback address, dynamic port allocation, and dev mode.
func DefaultTestConfig() *CAServerSettings {
	return NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}
