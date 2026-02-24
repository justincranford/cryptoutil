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

// SmIMServerSettings defines configuration settings for the SM-IM server.
type SmIMServerSettings struct {
	*cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings

	// Message encryption settings.
	MessageJWEAlgorithm string

	// Message validation constraints.
	MessageMinLength   int
	MessageMaxLength   int
	RecipientsMinCount int
	RecipientsMaxCount int
}

// SM-IM specific default values.
const (
	defaultMessageJWEAlgorithm = cryptoutilSharedMagic.IMJWEAlgorithm
	defaultMessageMinLength    = cryptoutilSharedMagic.IMMessageMinLength
	defaultMessageMaxLength    = cryptoutilSharedMagic.IMMessageMaxLength
	defaultRecipientsMinCount  = cryptoutilSharedMagic.IMRecipientsMinCount
	defaultRecipientsMaxCount  = cryptoutilSharedMagic.IMRecipientsMaxCount
)

var allSmIMServerRegisteredSettings []*cryptoutilAppsTemplateServiceConfig.Setting

// SM-IM specific Setting objects for parameter attributes.
var (
	messageJWEAlgorithm = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allSmIMServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "message-jwe-algorithm",
		Shorthand:   "",
		Value:       defaultMessageJWEAlgorithm,
		Usage:       "JWE algorithm for message encryption (e.g., dir+A256GCM)",
		Description: "Message JWE Algorithm",
	})
	messageMinLength = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allSmIMServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "message-min-length",
		Shorthand:   "",
		Value:       defaultMessageMinLength,
		Usage:       "minimum message length in characters",
		Description: "Message Min Length",
	})
	messageMaxLength = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allSmIMServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "message-max-length",
		Shorthand:   "",
		Value:       defaultMessageMaxLength,
		Usage:       "maximum message length in characters",
		Description: "Message Max Length",
	})
	recipientsMinCount = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allSmIMServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "recipients-min-count",
		Shorthand:   "",
		Value:       defaultRecipientsMinCount,
		Usage:       "minimum recipients per message",
		Description: "Recipients Min Count",
	})
	recipientsMaxCount = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allSmIMServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "recipients-max-count",
		Shorthand:   "",
		Value:       defaultRecipientsMaxCount,
		Usage:       "maximum recipients per message",
		Description: "Recipients Max Count",
	})
)

// ParseWithFlagSet parses command line arguments using provided FlagSet and returns sm-im settings.
// This enables test isolation by allowing each test to use its own FlagSet.
func ParseWithFlagSet(fs *pflag.FlagSet, args []string, exitIfHelp bool) (*SmIMServerSettings, error) {
	// Register sm-im specific flags on the provided FlagSet BEFORE parsing.
	// This must happen before calling template ParseWithFlagSet since it will call fs.Parse().
	fs.StringP(messageJWEAlgorithm.Name, messageJWEAlgorithm.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsStringSetting(messageJWEAlgorithm), messageJWEAlgorithm.Description)
	fs.IntP(messageMinLength.Name, messageMinLength.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(messageMinLength), messageMinLength.Description)
	fs.IntP(messageMaxLength.Name, messageMaxLength.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(messageMaxLength), messageMaxLength.Description)
	fs.IntP(recipientsMinCount.Name, recipientsMinCount.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(recipientsMinCount), recipientsMinCount.Description)
	fs.IntP(recipientsMaxCount.Name, recipientsMaxCount.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(recipientsMaxCount), recipientsMaxCount.Description)

	// Parse base template settings using the same FlagSet.
	// This will register template flags and call fs.Parse() + viper.BindPFlags().
	baseSettings, err := cryptoutilAppsTemplateServiceConfig.ParseWithFlagSet(fs, args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	// Create sm-im settings using values from viper (bound by template ParseWithFlagSet).
	settings := &SmIMServerSettings{
		ServiceTemplateServerSettings: baseSettings,
		MessageJWEAlgorithm:           viper.GetString(messageJWEAlgorithm.Name),
		MessageMinLength:              viper.GetInt(messageMinLength.Name),
		MessageMaxLength:              viper.GetInt(messageMaxLength.Name),
		RecipientsMinCount:            viper.GetInt(recipientsMinCount.Name),
		RecipientsMaxCount:            viper.GetInt(recipientsMaxCount.Name),
	}

	// NOTE: BrowserRealms and ServiceRealms are inherited from template configuration.
	// Template uses cryptoutilSharedMagic.DefaultBrowserRealms (6 browser realms) and
	// cryptoutilSharedMagic.DefaultServiceRealms (6 service realms) as defaults.
	// See internal/shared/magic/magic_identity.go for complete realm definitions.
	// See internal/apps/template/service/server/repository/tenant_domain.go for TenantRealm types.
	// See internal/apps/sm/im/service/realm_service.go for realm management.
	// Realms are stored in database and loaded at runtime via migration 0005_add_realms.up.sql.

	// Override template defaults with sm-im specific values.
	// Only override public port if user didn't explicitly specify one via CLI flag.
	// Private admin port (9090) is universal across all services.
	if !fs.Changed("bind-public-port") {
		settings.BindPublicPort = cryptoutilSharedMagic.IMServicePort
	}

	settings.OTLPService = cryptoutilSharedMagic.OTLPServiceSMIM

	// Validate sm-im specific settings.
	if err := validateSmIMSettings(settings); err != nil {
		return nil, fmt.Errorf("sm-im settings validation failed: %w", err)
	}

	// Log sm-im specific settings.
	logSmIMSettings(settings)

	return settings, nil
}

// Parse parses command-line arguments and returns the SM-IM server settings.
func Parse(args []string, exitIfHelp bool) (*SmIMServerSettings, error) {
	return ParseWithFlagSet(pflag.CommandLine, args, exitIfHelp)
}

// validateSmIMSettings validates sm-im specific configuration.
func validateSmIMSettings(s *SmIMServerSettings) error {
	var validationErrors []string

	// Validate message JWE algorithm.
	if s.MessageJWEAlgorithm == "" {
		validationErrors = append(validationErrors, "message-jwe-algorithm cannot be empty")
	}

	// Validate message length constraints.
	if s.MessageMinLength < 1 {
		validationErrors = append(validationErrors, fmt.Sprintf("message-min-length must be >= 1, got %d", s.MessageMinLength))
	}

	if s.MessageMaxLength < s.MessageMinLength {
		validationErrors = append(validationErrors, fmt.Sprintf("message-max-length (%d) must be >= message-min-length (%d)", s.MessageMaxLength, s.MessageMinLength))
	}

	// Validate recipients count constraints.
	if s.RecipientsMinCount < 1 {
		validationErrors = append(validationErrors, fmt.Sprintf("recipients-min-count must be >= 1, got %d", s.RecipientsMinCount))
	}

	if s.RecipientsMaxCount < s.RecipientsMinCount {
		validationErrors = append(validationErrors, fmt.Sprintf("recipients-max-count (%d) must be >= recipients-min-count (%d)", s.RecipientsMaxCount, s.RecipientsMinCount))
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// logSmIMSettings logs sm-im specific configuration to stderr.
func logSmIMSettings(s *SmIMServerSettings) {
	fmt.Fprintf(os.Stderr, "SM-IM Server Settings:\n")
	fmt.Fprintf(os.Stderr, "  Public Server: %s\n", s.PublicBaseURL())
	fmt.Fprintf(os.Stderr, "  Private Server: %s\n", s.PrivateBaseURL())
	fmt.Fprintf(os.Stderr, "  OTLP Service: %s\n", s.OTLPService)
	fmt.Fprintf(os.Stderr, "  Browser Realms: %s\n", strings.Join(s.BrowserRealms, ", "))
	fmt.Fprintf(os.Stderr, "  Service Realms: %s\n", strings.Join(s.ServiceRealms, ", "))
	fmt.Fprintf(os.Stderr, "  Message JWE Algorithm: %s\n", s.MessageJWEAlgorithm)
	fmt.Fprintf(os.Stderr, "  Message Length: %d - %d\n", s.MessageMinLength, s.MessageMaxLength)
	fmt.Fprintf(os.Stderr, "  Recipients Count: %d - %d\n", s.RecipientsMinCount, s.RecipientsMaxCount)
}
