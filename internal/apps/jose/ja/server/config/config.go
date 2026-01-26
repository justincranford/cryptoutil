// Copyright (c) 2025 Justin Cranford

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

// JoseJAServerSettings contains jose-ja specific configuration.
type JoseJAServerSettings struct {
	*cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings

	// Elastic Key settings.
	DefaultMaxMaterials int

	// Audit settings.
	AuditEnabled      bool
	AuditSamplingRate int
}

// Jose-JA specific default values.
const (
	defaultMaxMaterials      = cryptoutilSharedMagic.JoseJADefaultMaxMaterials
	defaultAuditEnabled      = cryptoutilSharedMagic.JoseJAAuditDefaultEnabled
	defaultAuditSamplingRate = cryptoutilSharedMagic.JoseJAAuditDefaultSamplingRate
)

var allJoseJAServerRegisteredSettings []*cryptoutilAppsTemplateServiceConfig.Setting

// Jose-JA specific Setting objects for parameter attributes.
var (
	maxMaterialsSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allJoseJAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "max-materials",
		Shorthand:   "",
		Value:       defaultMaxMaterials,
		Usage:       "default maximum material keys per elastic key",
		Description: "Max Materials Per Elastic Key",
	})
	auditEnabledSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allJoseJAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "audit-enabled",
		Shorthand:   "",
		Value:       defaultAuditEnabled,
		Usage:       "enable audit logging for JWK operations",
		Description: "Audit Enabled",
	})
	auditSamplingRateSetting = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allJoseJAServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "audit-sampling-rate",
		Shorthand:   "",
		Value:       defaultAuditSamplingRate,
		Usage:       "audit sampling rate (0-100, percentage of operations to log)",
		Description: "Audit Sampling Rate",
	})
)

// ParseWithFlagSet parses command line arguments using provided FlagSet and returns jose-ja settings.
// This enables test isolation by allowing each test to use its own FlagSet.
func ParseWithFlagSet(fs *pflag.FlagSet, args []string, exitIfHelp bool) (*JoseJAServerSettings, error) {
	// Register jose-ja specific flags on the provided FlagSet BEFORE parsing.
	// This must happen before calling template ParseWithFlagSet since it will call fs.Parse().
	fs.IntP(maxMaterialsSetting.Name, maxMaterialsSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(maxMaterialsSetting), maxMaterialsSetting.Description)
	fs.BoolP(auditEnabledSetting.Name, auditEnabledSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsBoolSetting(auditEnabledSetting), auditEnabledSetting.Description)
	fs.IntP(auditSamplingRateSetting.Name, auditSamplingRateSetting.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(auditSamplingRateSetting), auditSamplingRateSetting.Description)

	// Parse base template settings using the same FlagSet.
	// This will register template flags and call fs.Parse() + viper.BindPFlags().
	baseSettings, err := cryptoutilAppsTemplateServiceConfig.ParseWithFlagSet(fs, args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	// Create jose-ja settings using values from viper (bound by template ParseWithFlagSet).
	settings := &JoseJAServerSettings{
		ServiceTemplateServerSettings: baseSettings,
		DefaultMaxMaterials:           viper.GetInt(maxMaterialsSetting.Name),
		AuditEnabled:                  viper.GetBool(auditEnabledSetting.Name),
		AuditSamplingRate:             viper.GetInt(auditSamplingRateSetting.Name),
	}

	// Override template defaults with jose-ja specific values.
	// NOTE: Only override public port - private admin port (9090) is universal across all services.
	settings.BindPublicPort = cryptoutilSharedMagic.JoseJAServicePort
	settings.OTLPService = cryptoutilSharedMagic.OTLPServiceJoseJA

	// Validate jose-ja specific settings.
	if err := validateJoseJASettings(settings); err != nil {
		return nil, fmt.Errorf("jose-ja settings validation failed: %w", err)
	}

	// Log jose-ja specific settings.
	logJoseJASettings(settings)

	return settings, nil
}

// Parse parses command line arguments and returns jose-ja settings.
// Uses global pflag.CommandLine for backward compatibility.
func Parse(args []string, exitIfHelp bool) (*JoseJAServerSettings, error) {
	return ParseWithFlagSet(pflag.CommandLine, args, exitIfHelp)
}

// validateJoseJASettings validates jose-ja specific configuration.
func validateJoseJASettings(s *JoseJAServerSettings) error {
	var validationErrors []string

	// Validate max materials.
	if s.DefaultMaxMaterials < cryptoutilSharedMagic.JoseJAMinMaterials {
		validationErrors = append(validationErrors, fmt.Sprintf("max-materials must be >= %d, got %d", cryptoutilSharedMagic.JoseJAMinMaterials, s.DefaultMaxMaterials))
	}

	if s.DefaultMaxMaterials > cryptoutilSharedMagic.JoseJAMaxMaterials {
		validationErrors = append(validationErrors, fmt.Sprintf("max-materials must be <= %d, got %d", cryptoutilSharedMagic.JoseJAMaxMaterials, s.DefaultMaxMaterials))
	}

	// Validate audit sampling rate.
	if s.AuditSamplingRate < cryptoutilSharedMagic.JoseJAAuditMinSamplingRate {
		validationErrors = append(validationErrors, fmt.Sprintf("audit-sampling-rate must be >= %d, got %d", cryptoutilSharedMagic.JoseJAAuditMinSamplingRate, s.AuditSamplingRate))
	}

	if s.AuditSamplingRate > cryptoutilSharedMagic.JoseJAAuditMaxSamplingRate {
		validationErrors = append(validationErrors, fmt.Sprintf("audit-sampling-rate must be <= %d, got %d", cryptoutilSharedMagic.JoseJAAuditMaxSamplingRate, s.AuditSamplingRate))
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// logJoseJASettings logs jose-ja specific configuration to stderr.
func logJoseJASettings(s *JoseJAServerSettings) {
	fmt.Fprintf(os.Stderr, "Jose-JA Server Settings:\n")
	fmt.Fprintf(os.Stderr, "  Public Server: %s\n", s.PublicBaseURL())
	fmt.Fprintf(os.Stderr, "  Private Server: %s\n", s.PrivateBaseURL())
	fmt.Fprintf(os.Stderr, "  OTLP Service: %s\n", s.OTLPService)
	fmt.Fprintf(os.Stderr, "  Browser Realms: %s\n", strings.Join(s.BrowserRealms, ", "))
	fmt.Fprintf(os.Stderr, "  Service Realms: %s\n", strings.Join(s.ServiceRealms, ", "))
	fmt.Fprintf(os.Stderr, "  Default Max Materials: %d\n", s.DefaultMaxMaterials)
	fmt.Fprintf(os.Stderr, "  Audit Enabled: %t\n", s.AuditEnabled)
	fmt.Fprintf(os.Stderr, "  Audit Sampling Rate: %d%%\n", s.AuditSamplingRate)
}
