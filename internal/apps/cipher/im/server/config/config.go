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

// CipherImServerSettings defines configuration settings for the Cipher-IM server.
type CipherImServerSettings struct {
	*cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings

	// Message encryption settings.
	MessageJWEAlgorithm string

	// Message validation constraints.
	MessageMinLength   int
	MessageMaxLength   int
	RecipientsMinCount int
	RecipientsMaxCount int
}

// Cipher-IM specific default values.
const (
	defaultMessageJWEAlgorithm = cryptoutilSharedMagic.CipherJWEAlgorithm
	defaultMessageMinLength    = cryptoutilSharedMagic.CipherMessageMinLength
	defaultMessageMaxLength    = cryptoutilSharedMagic.CipherMessageMaxLength
	defaultRecipientsMinCount  = cryptoutilSharedMagic.CipherRecipientsMinCount
	defaultRecipientsMaxCount  = cryptoutilSharedMagic.CipherRecipientsMaxCount
)

var allCipherIMServerRegisteredSettings []*cryptoutilAppsTemplateServiceConfig.Setting

// Cipher-IM specific Setting objects for parameter attributes.
var (
	messageJWEAlgorithm = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allCipherIMServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "message-jwe-algorithm",
		Shorthand:   "",
		Value:       defaultMessageJWEAlgorithm,
		Usage:       "JWE algorithm for message encryption (e.g., dir+A256GCM)",
		Description: "Message JWE Algorithm",
	})
	messageMinLength = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allCipherIMServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "message-min-length",
		Shorthand:   "",
		Value:       defaultMessageMinLength,
		Usage:       "minimum message length in characters",
		Description: "Message Min Length",
	})
	messageMaxLength = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allCipherIMServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "message-max-length",
		Shorthand:   "",
		Value:       defaultMessageMaxLength,
		Usage:       "maximum message length in characters",
		Description: "Message Max Length",
	})
	recipientsMinCount = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allCipherIMServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "recipients-min-count",
		Shorthand:   "",
		Value:       defaultRecipientsMinCount,
		Usage:       "minimum recipients per message",
		Description: "Recipients Min Count",
	})
	recipientsMaxCount = cryptoutilAppsTemplateServiceConfig.SetEnvAndRegisterSetting(allCipherIMServerRegisteredSettings, &cryptoutilAppsTemplateServiceConfig.Setting{
		Name:        "recipients-max-count",
		Shorthand:   "",
		Value:       defaultRecipientsMaxCount,
		Usage:       "maximum recipients per message",
		Description: "Recipients Max Count",
	})
)

// Parse parses command-line arguments and returns the Cipher-IM server settings.
func Parse(args []string, exitIfHelp bool) (*CipherImServerSettings, error) {
	// Parse base template settings first.
	baseSettings, err := cryptoutilAppsTemplateServiceConfig.Parse(args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	// Register cipher-im specific flags.
	pflag.StringP(messageJWEAlgorithm.Name, messageJWEAlgorithm.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsStringSetting(messageJWEAlgorithm), messageJWEAlgorithm.Description)
	pflag.IntP(messageMinLength.Name, messageMinLength.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(messageMinLength), messageMinLength.Description)
	pflag.IntP(messageMaxLength.Name, messageMaxLength.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(messageMaxLength), messageMaxLength.Description)
	pflag.IntP(recipientsMinCount.Name, recipientsMinCount.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(recipientsMinCount), recipientsMinCount.Description)
	pflag.IntP(recipientsMaxCount.Name, recipientsMaxCount.Shorthand, cryptoutilAppsTemplateServiceConfig.RegisterAsIntSetting(recipientsMaxCount), recipientsMaxCount.Description)

	// Parse flags.
	pflag.Parse()

	// Bind flags to viper.
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Create cipher-im settings.
	settings := &CipherImServerSettings{
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
	// See internal/apps/cipher/im/service/realm_service.go for realm management.
	// Realms are stored in database and loaded at runtime via migration 0005_add_realms.up.sql.

	// Override template defaults with cipher-im specific values.
	// NOTE: Only override public port - private admin port (9090) is universal across all services.
	settings.BindPublicPort = cryptoutilSharedMagic.CipherServicePort
	settings.OTLPService = cryptoutilSharedMagic.OTLPServiceCipherIM

	// Validate cipher-im specific settings.
	if err := validateCipherImSettings(settings); err != nil {
		return nil, fmt.Errorf("cipher-im settings validation failed: %w", err)
	}

	// Log cipher-im specific settings.
	logCipherImSettings(settings)

	return settings, nil
}

// validateCipherImSettings validates cipher-im specific configuration.
func validateCipherImSettings(s *CipherImServerSettings) error {
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

// logCipherImSettings logs cipher-im specific configuration to stderr.
func logCipherImSettings(s *CipherImServerSettings) {
	fmt.Fprintf(os.Stderr, "Cipher-IM Server Settings:\n")
	fmt.Fprintf(os.Stderr, "  Public Server: %s\n", s.PublicBaseURL())
	fmt.Fprintf(os.Stderr, "  Private Server: %s\n", s.PrivateBaseURL())
	fmt.Fprintf(os.Stderr, "  OTLP Service: %s\n", s.OTLPService)
	fmt.Fprintf(os.Stderr, "  Browser Realms: %s\n", strings.Join(s.BrowserRealms, ", "))
	fmt.Fprintf(os.Stderr, "  Service Realms: %s\n", strings.Join(s.ServiceRealms, ", "))
	fmt.Fprintf(os.Stderr, "  Message JWE Algorithm: %s\n", s.MessageJWEAlgorithm)
	fmt.Fprintf(os.Stderr, "  Message Length: %d - %d\n", s.MessageMinLength, s.MessageMaxLength)
	fmt.Fprintf(os.Stderr, "  Recipients Count: %d - %d\n", s.RecipientsMinCount, s.RecipientsMaxCount)
}
