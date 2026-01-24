// Copyright (c) 2025 Justin Cranford

// Package config provides configuration for the JOSE Authority Server.
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

// JoseServerSettings contains JOSE Authority Server configuration.
type JoseServerSettings struct {
	*cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings

	// Key management settings.
	MaxMaterialsPerElasticKey int

	// Audit settings.
	AuditEnabled      bool
	AuditSamplingRate float64
}

// JOSE-specific default values.
const (
	defaultMaxMaterialsPerElasticKey = 1000
	defaultAuditEnabled              = true
	defaultAuditSamplingRate         = 0.01 // 1% sampling rate.
)

var allJoseServerRegisteredSettings []*cryptoutilAppsTemplateServiceConfig.Setting

// JOSE-specific Setting objects for parameter attributes.
var (
	maxMaterialsPerElasticKey = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allJoseServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "max-materials-per-elastic-key",
		Shorthand:   "",
		Value:       defaultMaxMaterialsPerElasticKey,
		Usage:       "maximum number of material keys per elastic key",
		Description: "Max Materials per Elastic Key",
	})
	auditEnabled = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allJoseServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "audit-enabled",
		Shorthand:   "",
		Value:       defaultAuditEnabled,
		Usage:       "enable audit logging for JOSE operations",
		Description: "Audit Enabled",
	})
	auditSamplingRate = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allJoseServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "audit-sampling-rate",
		Shorthand:   "",
		Value:       defaultAuditSamplingRate,
		Usage:       "sampling rate for audit logging (0.0-1.0)",
		Description: "Audit Sampling Rate",
	})
)

// Parse parses command-line arguments and config files to produce JoseServerSettings.
// It layers: defaults < config file < environment variables < command-line flags.
func Parse(args []string, exitIfHelp bool) (*JoseServerSettings, error) {
	// Parse base template settings first.
	baseSettings, err := cryptoutilAppsTemplateServiceConfig.Parse(args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	// Register JOSE-specific flags.
	pflag.IntP(maxMaterialsPerElasticKey.Name, maxMaterialsPerElasticKey.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(maxMaterialsPerElasticKey), maxMaterialsPerElasticKey.Description)
	pflag.BoolP(auditEnabled.Name, auditEnabled.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsBoolSetting(auditEnabled), auditEnabled.Description)
	pflag.Float64(auditSamplingRate.Name, defaultAuditSamplingRate, auditSamplingRate.Description)

	// Parse flags.
	pflag.Parse()

	// Bind flags to viper.
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Create JOSE settings.
	settings := &JoseServerSettings{
		ServiceTemplateServerSettings: baseSettings,
		MaxMaterialsPerElasticKey:     viper.GetInt(maxMaterialsPerElasticKey.Name),
		AuditEnabled:                  viper.GetBool(auditEnabled.Name),
		AuditSamplingRate:             viper.GetFloat64(auditSamplingRate.Name),
	}

	// Override template defaults with JOSE-specific values.
	// NOTE: Only override public port - private admin port (9090) is universal across all services.
	settings.BindPublicPort = cryptoutilSharedMagic.JoseJAServicePort
	settings.OTLPService = cryptoutilSharedMagic.OTLPServiceJoseJA

	// Validate JOSE-specific settings.
	if err := validateJoseSettings(settings); err != nil {
		return nil, fmt.Errorf("jose settings validation failed: %w", err)
	}

	// Log JOSE-specific settings.
	logJoseSettings(settings)

	return settings, nil
}

// validateJoseSettings validates JOSE-specific configuration.
func validateJoseSettings(s *JoseServerSettings) error {
	var validationErrors []string

	// Validate max materials per elastic key.
	if s.MaxMaterialsPerElasticKey < 1 {
		validationErrors = append(validationErrors, fmt.Sprintf("max-materials-per-elastic-key must be >= 1, got %d", s.MaxMaterialsPerElasticKey))
	}

	// Validate audit sampling rate.
	if s.AuditSamplingRate < 0.0 || s.AuditSamplingRate > 1.0 {
		validationErrors = append(validationErrors, fmt.Sprintf("audit-sampling-rate must be between 0.0 and 1.0, got %f", s.AuditSamplingRate))
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// logJoseSettings logs JOSE-specific configuration to stderr.
func logJoseSettings(s *JoseServerSettings) {
	fmt.Fprintf(os.Stderr, "JOSE Server Settings:\n")
	fmt.Fprintf(os.Stderr, "  Public Server: %s\n", s.PublicBaseURL())
	fmt.Fprintf(os.Stderr, "  Private Server: %s\n", s.PrivateBaseURL())
	fmt.Fprintf(os.Stderr, "  OTLP Service: %s\n", s.OTLPService)
	fmt.Fprintf(os.Stderr, "  Browser Realms: %s\n", strings.Join(s.BrowserRealms, ", "))
	fmt.Fprintf(os.Stderr, "  Service Realms: %s\n", strings.Join(s.ServiceRealms, ", "))
	fmt.Fprintf(os.Stderr, "  Max Materials Per Elastic Key: %d\n", s.MaxMaterialsPerElasticKey)
	fmt.Fprintf(os.Stderr, "  Audit Enabled: %t\n", s.AuditEnabled)
	fmt.Fprintf(os.Stderr, "  Audit Sampling Rate: %.2f\n", s.AuditSamplingRate)
}

// NewDevSettings creates development settings with sensible defaults.
func NewDevSettings() *JoseServerSettings {
	return &JoseServerSettings{
		ServiceTemplateServerSettings: cryptoutilAppsTemplateServiceConfig.NewForJOSEServer(
			cryptoutilSharedMagic.IPv4Loopback,
			cryptoutilSharedMagic.DefaultPublicPortJOSEServer,
			true, // dev mode.
		),
		MaxMaterialsPerElasticKey: defaultMaxMaterialsPerElasticKey,
		AuditEnabled:              defaultAuditEnabled,
		AuditSamplingRate:         defaultAuditSamplingRate,
	}
}

// NewTestSettings creates settings for testing without calling Parse().
// This bypasses pflag's global FlagSet to allow multiple config creations in tests.
// Use this in tests instead of NewDevSettings to avoid "flag redefined" panics.
func NewTestSettings() *JoseServerSettings {
	return &JoseServerSettings{
		ServiceTemplateServerSettings: cryptoutilAppsTemplateServiceConfig.NewTestConfig(
			cryptoutilSharedMagic.IPv4Loopback,
			0, // Dynamic port allocation for tests.
			true,
		),
		MaxMaterialsPerElasticKey: defaultMaxMaterialsPerElasticKey,
		AuditEnabled:              defaultAuditEnabled,
		AuditSamplingRate:         defaultAuditSamplingRate,
	}
}
