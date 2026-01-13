// Copyright (c) 2025 Justin Cranford

package config

import (
	"fmt"
	"os"
	"strings"

	cryptoutilTemplateConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type CipherImServerSettings struct {
	*cryptoutilTemplateConfig.ServiceTemplateServerSettings

	// Message encryption settings.
	MessageJWEAlgorithm string

	// Message validation constraints.
	MessageMinLength   int
	MessageMaxLength   int
	RecipientsMinCount int
	RecipientsMaxCount int
}

func Parse(args []string, exitIfHelp bool) (*CipherImServerSettings, error) {
	// Parse base template settings first.
	baseSettings, err := cryptoutilTemplateConfig.Parse(args, exitIfHelp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template settings: %w", err)
	}

	// Register cipher-im specific flags.
	pflag.StringP("message-jwe-algorithm", "j", cryptoutilSharedMagic.CipherJWEAlgorithm, "JWE algorithm for message encryption (e.g., dir+A256GCM)")
	pflag.IntP("message-min-length", "m", cryptoutilSharedMagic.CipherMessageMinLength, "Minimum message length in characters")
	pflag.IntP("message-max-length", "M", cryptoutilSharedMagic.CipherMessageMaxLength, "Maximum message length in characters")
	pflag.IntP("recipients-min-count", "r", cryptoutilSharedMagic.CipherRecipientsMinCount, "Minimum recipients per message")
	pflag.IntP("recipients-max-count", "R", cryptoutilSharedMagic.CipherRecipientsMaxCount, "Maximum recipients per message")

	// Parse flags.
	pflag.Parse()

	// Bind flags to viper.
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Create cipher-im settings.
	settings := &CipherImServerSettings{
		ServiceTemplateServerSettings: baseSettings,
		MessageJWEAlgorithm:           viper.GetString("message-jwe-algorithm"),
		MessageMinLength:              viper.GetInt("message-min-length"),
		MessageMaxLength:              viper.GetInt("message-max-length"),
		RecipientsMinCount:            viper.GetInt("recipients-min-count"),
		RecipientsMaxCount:            viper.GetInt("recipients-max-count"),
	}

	// Initialize cipher-im specific realms (6 non-federated authn methods).
	settings.Realms = []string{
		"jwe-session-cookie",
		"jws-session-cookie",
		"opaque-session-cookie",
		"basic-username-password",
		"bearer-api-token",
		"https-client-cert",
	}

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
	fmt.Fprintf(os.Stderr, "  Realms: %s\n", strings.Join(s.Realms, ", "))
	fmt.Fprintf(os.Stderr, "  Message JWE Algorithm: %s\n", s.MessageJWEAlgorithm)
	fmt.Fprintf(os.Stderr, "  Message Length: %d - %d\n", s.MessageMinLength, s.MessageMaxLength)
	fmt.Fprintf(os.Stderr, "  Recipients Count: %d - %d\n", s.RecipientsMinCount, s.RecipientsMaxCount)
}
