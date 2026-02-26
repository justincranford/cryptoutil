// Copyright (c) 2025 Justin Cranford
//
//

// Package config provides configuration types and loading for the CA subsystem.
package config

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// CA type constants.
const (
	CATypeRoot         = "root"
	CATypeIntermediate = "intermediate"
	CATypeIssuing      = "issuing"
)

// CAConfig represents the configuration for a Certificate Authority.
type CAConfig struct {
	CA CADefinition `yaml:"ca"`
}

// CADefinition defines a Certificate Authority's properties.
type CADefinition struct {
	Name          string             `yaml:"name"`
	Description   string             `yaml:"description"`
	Type          string             `yaml:"type"`   // root, intermediate, issuing
	Parent        string             `yaml:"parent"` // Parent CA name (empty for root)
	Subject       SubjectConfig      `yaml:"subject"`
	Key           KeyConfig          `yaml:"key"`
	Validity      ValidityConfig     `yaml:"validity"`
	MaxPathLength int                `yaml:"max_path_length"`
	Distribution  DistributionConfig `yaml:"distribution"`
}

// SubjectConfig defines the subject distinguished name components.
type SubjectConfig struct {
	CommonName         string `yaml:"common_name"`
	Organization       string `yaml:"organization"`
	OrganizationalUnit string `yaml:"organizational_unit"`
	Country            string `yaml:"country"`
	State              string `yaml:"state"`
	Locality           string `yaml:"locality"`
}

// KeyConfig defines key generation parameters.
type KeyConfig struct {
	Algorithm   string `yaml:"algorithm"`     // RSA, ECDSA, EdDSA
	CurveOrSize string `yaml:"curve_or_size"` // P-256, P-384, 2048, 4096, etc.
}

// ValidityConfig defines certificate validity period.
type ValidityConfig struct {
	Days                   int `yaml:"days"`
	NotBeforeOffsetMinutes int `yaml:"not_before_offset_minutes"`
}

// DistributionConfig defines CRL and OCSP distribution points.
type DistributionConfig struct {
	CRLURL  string `yaml:"crl_url"`
	OCSPURL string `yaml:"ocsp_url"`
	AIAURL  string `yaml:"aia_url"`
}

// ProfileConfig represents a certificate profile configuration.
type ProfileConfig struct {
	Profile ProfileDefinition `yaml:"profile"`
}

// ProfileDefinition defines certificate issuance policy.
type ProfileDefinition struct {
	Name             string                 `yaml:"name"`
	Description      string                 `yaml:"description"`
	Validity         ProfileValidityConfig  `yaml:"validity"`
	Key              ProfileKeyConfig       `yaml:"key"`
	KeyUsage         []string               `yaml:"key_usage"`
	ExtendedKeyUsage ExtendedKeyUsageConfig `yaml:"extended_key_usage"`
	Subject          SubjectConstraints     `yaml:"subject"`
	SAN              SANConstraints         `yaml:"san"`
	Extensions       ExtensionConstraints   `yaml:"extensions"`
	BasicConstraints BasicConstraintsConfig `yaml:"basic_constraints"`
	Signature        SignatureConfig        `yaml:"signature"`
}

// ProfileValidityConfig defines validity constraints for a profile.
type ProfileValidityConfig struct {
	MaxDays     int `yaml:"max_days"`
	MinDays     int `yaml:"min_days"`
	DefaultDays int `yaml:"default_days"`
}

// ProfileKeyConfig defines key constraints for a profile.
type ProfileKeyConfig struct {
	AllowedAlgorithms  []AlgorithmConstraint `yaml:"allowed_algorithms"`
	DefaultAlgorithm   string                `yaml:"default_algorithm"`
	DefaultCurveOrSize string                `yaml:"default_curve_or_size"`
}

// AlgorithmConstraint defines constraints for a specific algorithm.
type AlgorithmConstraint struct {
	Algorithm     string   `yaml:"algorithm"`
	MinSize       int      `yaml:"min_size,omitempty"`
	MaxSize       int      `yaml:"max_size,omitempty"`
	AllowedCurves []string `yaml:"allowed_curves,omitempty"`
}

// ExtendedKeyUsageConfig defines extended key usage constraints.
type ExtendedKeyUsageConfig struct {
	Required []string `yaml:"required"`
	Optional []string `yaml:"optional"`
}

// SubjectConstraints defines subject field constraints.
type SubjectConstraints struct {
	RequireCommonName       bool `yaml:"require_common_name"`
	AllowOrganization       bool `yaml:"allow_organization"`
	AllowOrganizationalUnit bool `yaml:"allow_organizational_unit"`
	AllowCountry            bool `yaml:"allow_country"`
	AllowState              bool `yaml:"allow_state"`
	AllowLocality           bool `yaml:"allow_locality"`
	MaxCNLength             int  `yaml:"max_cn_length"`
}

// SANConstraints defines Subject Alternative Name constraints.
type SANConstraints struct {
	AllowDNSNames       bool     `yaml:"allow_dns_names"`
	AllowIPAddresses    bool     `yaml:"allow_ip_addresses"`
	AllowEmailAddresses bool     `yaml:"allow_email_addresses"`
	AllowURIs           bool     `yaml:"allow_uris"`
	RequireAtLeastOne   bool     `yaml:"require_at_least_one"`
	MaxEntries          int      `yaml:"max_entries"`
	DNSPatterns         []string `yaml:"dns_patterns,omitempty"`
	AllowedIPRanges     []string `yaml:"allowed_ip_ranges,omitempty"`
}

// ExtensionConstraints defines certificate extension constraints.
type ExtensionConstraints struct {
	Required []string `yaml:"required"`
	Optional []string `yaml:"optional"`
}

// BasicConstraintsConfig defines basicConstraints extension settings.
type BasicConstraintsConfig struct {
	IsCA       bool `yaml:"is_ca"`
	PathLength int  `yaml:"path_length"`
}

// SignatureConfig defines signature algorithm preferences.
type SignatureConfig struct {
	Preferred []string `yaml:"preferred"`
	Forbidden []string `yaml:"forbidden"`
}

// LoadCAConfig loads a CA configuration from a YAML file.
func LoadCAConfig(path string) (*CAConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA config file %s: %w", path, err)
	}

	var config CAConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse CA config file %s: %w", path, err)
	}

	if err := validateCAConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid CA config in %s: %w", path, err)
	}

	return &config, nil
}

// LoadProfileConfig loads a certificate profile configuration from a YAML file.
func LoadProfileConfig(path string) (*ProfileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile config file %s: %w", path, err)
	}

	var config ProfileConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse profile config file %s: %w", path, err)
	}

	if err := validateProfileConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid profile config in %s: %w", path, err)
	}

	return &config, nil
}

func validateCAConfig(config *CAConfig) error {
	if config.CA.Name == "" {
		return fmt.Errorf("CA name is required")
	}

	switch config.CA.Type {
	case CATypeRoot, CATypeIntermediate, CATypeIssuing:
		// Valid.
	default:
		return fmt.Errorf("invalid CA type: %s (must be root, intermediate, or issuing)", config.CA.Type)
	}

	if config.CA.Type != CATypeRoot && config.CA.Parent == "" {
		return fmt.Errorf("parent CA is required for %s CA", config.CA.Type)
	}

	if config.CA.Subject.CommonName == "" {
		return fmt.Errorf("subject common name is required")
	}

	if err := validateKeyConfig(&config.CA.Key); err != nil {
		return fmt.Errorf("invalid key config: %w", err)
	}

	if config.CA.Validity.Days <= 0 {
		return fmt.Errorf("validity days must be positive")
	}

	return nil
}

func validateKeyConfig(key *KeyConfig) error {
	switch key.Algorithm {
	case cryptoutilSharedMagic.KeyTypeRSA:
		// CurveOrSize should be a number (2048, 3072, 4096).
	case "ECDSA":
		switch key.CurveOrSize {
		case "P-256", "P-384", "P-521":
			// Valid.
		default:
			return fmt.Errorf("invalid ECDSA curve: %s", key.CurveOrSize)
		}
	case cryptoutilSharedMagic.JoseAlgEdDSA:
		switch key.CurveOrSize {
		case cryptoutilSharedMagic.EdCurveEd25519, cryptoutilSharedMagic.EdCurveEd448:
			// Valid.
		default:
			return fmt.Errorf("invalid EdDSA curve: %s", key.CurveOrSize)
		}
	default:
		return fmt.Errorf("invalid algorithm: %s (must be RSA, ECDSA, or EdDSA)", key.Algorithm)
	}

	return nil
}

func validateProfileConfig(config *ProfileConfig) error {
	if config.Profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if config.Profile.Validity.MaxDays <= 0 {
		return fmt.Errorf("max validity days must be positive")
	}

	if config.Profile.Validity.DefaultDays > config.Profile.Validity.MaxDays {
		return fmt.Errorf("default validity days cannot exceed max validity days")
	}

	if len(config.Profile.Key.AllowedAlgorithms) == 0 {
		return fmt.Errorf("at least one allowed algorithm is required")
	}

	return nil
}
