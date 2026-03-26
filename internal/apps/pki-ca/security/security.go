// Copyright (c) 2025 Justin Cranford

// Package security provides certificate security validation and compliance checking.
package security

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	rsa "crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ThreatCategory represents a category in the STRIDE threat model.
type ThreatCategory string

// STRIDE threat categories.
const (
	ThreatSpoofing            ThreatCategory = "spoofing"
	ThreatTampering           ThreatCategory = "tampering"
	ThreatRepudiation         ThreatCategory = "repudiation"
	ThreatInformationDisclose ThreatCategory = "information_disclosure"
	ThreatDenialOfService     ThreatCategory = "denial_of_service"
	ThreatElevationPrivilege  ThreatCategory = "elevation_of_privilege"
)

// Severity levels for threats and vulnerabilities.
type Severity string

// Severity constants.
const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Security configuration defaults.
const (
	defaultMinRSAKeySize       = 2048
	defaultMinECKeySize        = 256
	defaultMaxCertValidityDays = 398
)

// Threat represents a security threat in the threat model.
type Threat struct {
	ID          string         `json:"id"`
	Category    ThreatCategory `json:"category"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Asset       string         `json:"asset"`
	Severity    Severity       `json:"severity"`
	Likelihood  string         `json:"likelihood"`
	Impact      string         `json:"impact"`
	Mitigations []string       `json:"mitigations"`
	Status      string         `json:"status"`
}

// ThreatModel represents a complete STRIDE threat model for CA operations.
type ThreatModel struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Assets      []Asset   `json:"assets"`
	Threats     []Threat  `json:"threats"`
	Controls    []Control `json:"controls"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Asset represents a system asset in the threat model.
type Asset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Sensitivity string `json:"sensitivity"`
}

// Control represents a security control mitigating threats.
type Control struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Mitigates   []string `json:"mitigates"`
	Status      string   `json:"status"`
}

// Config defines security hardening configuration.
type Config struct {
	// MinRSAKeySize is the minimum RSA key size in bits.
	MinRSAKeySize int `yaml:"min_rsa_key_size" json:"min_rsa_key_size"`

	// MinECKeySize is the minimum EC key size in bits.
	MinECKeySize int `yaml:"min_ec_key_size" json:"min_ec_key_size"`

	// AllowedSignatureAlgorithms lists allowed signature algorithms.
	AllowedSignatureAlgorithms []x509.SignatureAlgorithm `yaml:"allowed_signature_algorithms" json:"allowed_signature_algorithms"`

	// MaxCertValidityDays is the maximum certificate validity period.
	MaxCertValidityDays int `yaml:"max_cert_validity_days" json:"max_cert_validity_days"`

	// RequireKeyUsage enforces key usage extension.
	RequireKeyUsage bool `yaml:"require_key_usage" json:"require_key_usage"`

	// RequireBasicConstraints enforces basic constraints extension.
	RequireBasicConstraints bool `yaml:"require_basic_constraints" json:"require_basic_constraints"`

	// RequireSAN enforces Subject Alternative Name extension.
	RequireSAN bool `yaml:"require_san" json:"require_san"`

	// DisallowWeakAlgorithms prevents use of weak algorithms.
	DisallowWeakAlgorithms bool `yaml:"disallow_weak_algorithms" json:"disallow_weak_algorithms"`

	// EnforcePathLengthConstraints enforces path length in CA certificates.
	EnforcePathLengthConstraints bool `yaml:"enforce_path_length_constraints" json:"enforce_path_length_constraints"`

	// AuditLoggingEnabled enables security audit logging.
	AuditLoggingEnabled bool `yaml:"audit_logging_enabled" json:"audit_logging_enabled"`
}

// Vulnerability represents a security vulnerability finding.
type Vulnerability struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    Severity  `json:"severity"`
	Category    string    `json:"category"`
	Location    string    `json:"location"`
	Remediation string    `json:"remediation"`
	FoundAt     time.Time `json:"found_at"`
}

// ValidationResult contains results of security validation.
type ValidationResult struct {
	Valid           bool            `json:"valid"`
	Errors          []string        `json:"errors"`
	Warnings        []string        `json:"warnings"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	CheckedAt       time.Time       `json:"checked_at"`
}

// Validator validates certificates and keys against security policies.
type Validator struct {
	config *Config
	mu     sync.RWMutex
}

// NewValidator creates a new security validator.
func NewValidator(config *Config) *Validator {
	if config == nil {
		config = DefaultConfig()
	}

	return &Validator{
		config: config,
	}
}

// DefaultConfig returns a secure default configuration.
func DefaultConfig() *Config {
	return &Config{
		MinRSAKeySize: defaultMinRSAKeySize,
		MinECKeySize:  defaultMinECKeySize,
		AllowedSignatureAlgorithms: []x509.SignatureAlgorithm{
			x509.SHA256WithRSA,
			x509.SHA384WithRSA,
			x509.SHA512WithRSA,
			x509.ECDSAWithSHA256,
			x509.ECDSAWithSHA384,
			x509.ECDSAWithSHA512,
			x509.PureEd25519,
		},
		MaxCertValidityDays:          defaultMaxCertValidityDays,
		RequireKeyUsage:              true,
		RequireBasicConstraints:      true,
		RequireSAN:                   true,
		DisallowWeakAlgorithms:       true,
		EnforcePathLengthConstraints: true,
		AuditLoggingEnabled:          true,
	}
}

// ValidateCertificate validates a certificate against security policies.
func (v *Validator) ValidateCertificate(_ context.Context, cert *x509.Certificate) (*ValidationResult, error) {
	if cert == nil {
		return nil, errors.New("certificate cannot be nil")
	}

	v.mu.RLock()
	config := v.config
	v.mu.RUnlock()

	result := &ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
		CheckedAt: time.Now().UTC(),
	}

	// Validate key size.
	if err := v.validateKeySize(cert, result); err != nil {
		return nil, fmt.Errorf("failed to validate key size: %w", err)
	}

	// Validate signature algorithm.
	v.validateSignatureAlgorithm(cert, config, result)

	// Validate validity period.
	v.validateValidityPeriod(cert, config, result)

	// Validate extensions.
	v.validateExtensions(cert, config, result)

	// Check for weak algorithms.
	if config.DisallowWeakAlgorithms {
		v.checkWeakAlgorithms(cert, result)
	}

	// Validate path length constraints for CA certificates.
	if cert.IsCA && config.EnforcePathLengthConstraints {
		v.validatePathLength(cert, result)
	}

	return result, nil
}

// validateKeySize validates the certificate's key size.
func (v *Validator) validateKeySize(cert *x509.Certificate, result *ValidationResult) error {
	v.mu.RLock()
	config := v.config
	v.mu.RUnlock()

	switch key := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		keySize := key.N.BitLen()
		if keySize < config.MinRSAKeySize {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("RSA key size %d bits is below minimum %d bits", keySize, config.MinRSAKeySize))
		}
	case *ecdsa.PublicKey:
		keySize := key.Curve.Params().BitSize
		if keySize < config.MinECKeySize {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("EC key size %d bits is below minimum %d bits", keySize, config.MinECKeySize))
		}
	case ed25519.PublicKey:
		// Ed25519 keys are always 256 bits, which is acceptable.
	default:
		result.Warnings = append(result.Warnings, "unknown public key type")
	}

	return nil
}

// validateSignatureAlgorithm validates the certificate's signature algorithm.
func (v *Validator) validateSignatureAlgorithm(cert *x509.Certificate, config *Config, result *ValidationResult) {
	allowed := false

	for _, alg := range config.AllowedSignatureAlgorithms {
		if cert.SignatureAlgorithm == alg {
			allowed = true

			break
		}
	}

	if !allowed {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("signature algorithm %s is not in allowed list", cert.SignatureAlgorithm))
	}
}

// validateValidityPeriod validates the certificate's validity period.
func (v *Validator) validateValidityPeriod(cert *x509.Certificate, config *Config, result *ValidationResult) {
	validityDays := int(cert.NotAfter.Sub(cert.NotBefore).Hours() / cryptoutilSharedMagic.HoursPerDay)

	if validityDays > config.MaxCertValidityDays {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("validity period %d days exceeds maximum %d days", validityDays, config.MaxCertValidityDays))
	}

	// Check for expired or not yet valid certificates.
	now := time.Now().UTC()
	if now.Before(cert.NotBefore) {
		result.Warnings = append(result.Warnings, "certificate is not yet valid")
	}

	if now.After(cert.NotAfter) {
		result.Valid = false
		result.Errors = append(result.Errors, "certificate has expired")
	}
}

// validateExtensions validates required certificate extensions.
func (v *Validator) validateExtensions(cert *x509.Certificate, config *Config, result *ValidationResult) {
	// Check key usage.
	if config.RequireKeyUsage && cert.KeyUsage == 0 {
		result.Warnings = append(result.Warnings, "certificate does not have key usage extension")
	}

	// Check basic constraints for CA certificates.
	if config.RequireBasicConstraints && cert.IsCA && !cert.BasicConstraintsValid {
		result.Valid = false
		result.Errors = append(result.Errors, "CA certificate missing valid basic constraints")
	}

	// Check SAN for non-CA certificates.
	if config.RequireSAN && !cert.IsCA {
		hasSAN := len(cert.DNSNames) > 0 || len(cert.EmailAddresses) > 0 ||
			len(cert.IPAddresses) > 0 || len(cert.URIs) > 0

		if !hasSAN {
			result.Warnings = append(result.Warnings, "certificate does not have Subject Alternative Name")
		}
	}
}

// checkWeakAlgorithms checks for use of weak cryptographic algorithms.
func (v *Validator) checkWeakAlgorithms(cert *x509.Certificate, result *ValidationResult) {
	weakAlgorithms := map[x509.SignatureAlgorithm]bool{
		x509.MD2WithRSA:  true,
		x509.MD5WithRSA:  true,
		x509.SHA1WithRSA: true,
		x509.DSAWithSHA1: true,
	}

	if weakAlgorithms[cert.SignatureAlgorithm] {
		result.Valid = false
		result.Vulnerabilities = append(result.Vulnerabilities, Vulnerability{
			ID:          "WEAK-ALG-001",
			Title:       "Weak Signature Algorithm",
			Description: fmt.Sprintf("Certificate uses weak signature algorithm: %s", cert.SignatureAlgorithm),
			Severity:    SeverityHigh,
			Category:    "cryptography",
			Location:    cert.Subject.CommonName,
			Remediation: "Re-issue certificate with SHA-256 or stronger signature algorithm",
			FoundAt:     time.Now().UTC(),
		})
	}
}

// validatePathLength validates path length constraints for CA certificates.
func (v *Validator) validatePathLength(cert *x509.Certificate, result *ValidationResult) {
	if cert.MaxPathLen == 0 && !cert.MaxPathLenZero {
		result.Warnings = append(result.Warnings, "CA certificate has no path length constraint")
	}
}

// ValidatePrivateKey validates a private key against security policies.
func (v *Validator) ValidatePrivateKey(_ context.Context, key any) (*ValidationResult, error) {
	if key == nil {
		return nil, errors.New("private key cannot be nil")
	}

	v.mu.RLock()
	config := v.config
	v.mu.RUnlock()

	result := &ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
		CheckedAt: time.Now().UTC(),
	}

	switch k := key.(type) {
	case *rsa.PrivateKey:
		keySize := k.N.BitLen()
		if keySize < config.MinRSAKeySize {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("RSA key size %d bits is below minimum %d bits", keySize, config.MinRSAKeySize))
		}
	case *ecdsa.PrivateKey:
		keySize := k.Curve.Params().BitSize
		if keySize < config.MinECKeySize {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("EC key size %d bits is below minimum %d bits", keySize, config.MinECKeySize))
		}
	case ed25519.PrivateKey:
		// Ed25519 keys are always acceptable.
	default:
		result.Warnings = append(result.Warnings, "unknown private key type")
	}

	return result, nil
}

// ValidateCSR validates a certificate signing request against security policies.
func (v *Validator) ValidateCSR(_ context.Context, csr *x509.CertificateRequest) (*ValidationResult, error) {
	if csr == nil {
		return nil, errors.New("CSR cannot be nil")
	}

	v.mu.RLock()
	config := v.config
	v.mu.RUnlock()

	result := &ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
		CheckedAt: time.Now().UTC(),
	}

	// Validate public key in CSR.
	switch key := csr.PublicKey.(type) {
	case *rsa.PublicKey:
		keySize := key.N.BitLen()
		if keySize < config.MinRSAKeySize {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("CSR RSA key size %d bits is below minimum %d bits", keySize, config.MinRSAKeySize))
		}
	case *ecdsa.PublicKey:
		keySize := key.Curve.Params().BitSize
		if keySize < config.MinECKeySize {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("CSR EC key size %d bits is below minimum %d bits", keySize, config.MinECKeySize))
		}
	case ed25519.PublicKey:
		// Ed25519 keys are always acceptable.
	default:
		result.Warnings = append(result.Warnings, "CSR has unknown public key type")
	}

	// Validate signature algorithm.
	allowed := false

	for _, alg := range config.AllowedSignatureAlgorithms {
		if csr.SignatureAlgorithm == alg {
			allowed = true

			break
		}
	}

	if !allowed {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("CSR signature algorithm %s is not in allowed list", csr.SignatureAlgorithm))
	}

	// Check for SAN.
	if config.RequireSAN {
		hasSAN := len(csr.DNSNames) > 0 || len(csr.EmailAddresses) > 0 ||
			len(csr.IPAddresses) > 0 || len(csr.URIs) > 0

		if !hasSAN {
			result.Warnings = append(result.Warnings, "CSR does not contain Subject Alternative Name")
		}
	}

	return result, nil
}

// ThreatModelBuilder helps construct threat models.
type ThreatModelBuilder struct {
	model *ThreatModel
}

// NewThreatModelBuilder creates a new threat model builder.
