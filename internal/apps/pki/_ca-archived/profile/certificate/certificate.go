// Copyright (c) 2025 Justin Cranford

// Package certificate provides certificate profile rendering for X.509 certificates.
// It implements YAML-driven certificate policy rendering with validation.
package certificate

import (
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Profile represents a certificate profile configuration.
type Profile struct {
	// Name is the unique identifier for this profile.
	Name string `yaml:"name"`

	// Description provides human-readable context.
	Description string `yaml:"description"`

	// Type indicates the certificate type.
	Type ProfileType `yaml:"type"`

	// Validity defines certificate lifetime constraints.
	Validity ValidityConfig `yaml:"validity"`

	// KeyUsage defines the key usage extensions.
	KeyUsage KeyUsageConfig `yaml:"key_usage"`

	// ExtendedKeyUsage defines extended key usage extensions.
	ExtendedKeyUsage ExtKeyUsageConfig `yaml:"extended_key_usage"`

	// BasicConstraints defines CA-related constraints.
	BasicConstraints BasicConstraintsConfig `yaml:"basic_constraints"`

	// Policies defines certificate policies.
	Policies []PolicyConfig `yaml:"policies"`

	// Extensions defines additional X.509 extensions.
	Extensions ExtensionConfig `yaml:"extensions"`
}

// ProfileType indicates the type of certificate.
type ProfileType string

// Certificate profile types.
const (
	ProfileTypeRoot         ProfileType = "root"
	ProfileTypeIntermediate ProfileType = "intermediate"
	ProfileTypeIssuing      ProfileType = "issuing"
	ProfileTypeTLSServer    ProfileType = "tls-server"
	ProfileTypeTLSClient    ProfileType = "tls-client"
	ProfileTypeCodeSigning  ProfileType = "code-signing"
	ProfileTypeSMIME        ProfileType = "smime"
	ProfileTypeOCSP         ProfileType = "ocsp"
	ProfileTypeTSA          ProfileType = "tsa"
)

// ValidityConfig defines certificate validity period.
type ValidityConfig struct {
	// Duration is the certificate lifetime.
	Duration string `yaml:"duration"`

	// MaxDuration is the maximum allowed lifetime.
	MaxDuration string `yaml:"max_duration"`

	// AllowCustom indicates if requester can specify duration.
	AllowCustom bool `yaml:"allow_custom"`

	// BackdateBuffer allows slight backdating for clock skew.
	BackdateBuffer string `yaml:"backdate_buffer"`
}

// KeyUsageConfig defines X.509 key usage flags.
type KeyUsageConfig struct {
	// DigitalSignature enables digital signature use.
	DigitalSignature bool `yaml:"digital_signature"`

	// ContentCommitment enables non-repudiation.
	ContentCommitment bool `yaml:"content_commitment"`

	// KeyEncipherment enables key encipherment.
	KeyEncipherment bool `yaml:"key_encipherment"`

	// DataEncipherment enables data encipherment.
	DataEncipherment bool `yaml:"data_encipherment"`

	// KeyAgreement enables key agreement.
	KeyAgreement bool `yaml:"key_agreement"`

	// CertSign enables certificate signing (CA only).
	CertSign bool `yaml:"cert_sign"`

	// CRLSign enables CRL signing (CA only).
	CRLSign bool `yaml:"crl_sign"`

	// EncipherOnly restricts key agreement to encipherment.
	EncipherOnly bool `yaml:"encipher_only"`

	// DecipherOnly restricts key agreement to decipherment.
	DecipherOnly bool `yaml:"decipher_only"`
}

// ExtKeyUsageConfig defines extended key usage.
type ExtKeyUsageConfig struct {
	// ServerAuth enables TLS server authentication.
	ServerAuth bool `yaml:"server_auth"`

	// ClientAuth enables TLS client authentication.
	ClientAuth bool `yaml:"client_auth"`

	// CodeSigning enables code signing.
	CodeSigning bool `yaml:"code_signing"`

	// EmailProtection enables S/MIME.
	EmailProtection bool `yaml:"email_protection"`

	// TimeStamping enables time stamping.
	TimeStamping bool `yaml:"time_stamping"`

	// OCSPSigning enables OCSP response signing.
	OCSPSigning bool `yaml:"ocsp_signing"`

	// CustomOIDs lists additional EKU OIDs.
	CustomOIDs []string `yaml:"custom_oids"`
}

// BasicConstraintsConfig defines basic constraints extension.
type BasicConstraintsConfig struct {
	// IsCA indicates this is a CA certificate.
	IsCA bool `yaml:"is_ca"`

	// PathLenConstraint limits CA path length.
	PathLenConstraint *int `yaml:"path_len_constraint"`
}

// PolicyConfig defines a certificate policy.
type PolicyConfig struct {
	// OID is the policy object identifier.
	OID string `yaml:"oid"`

	// CPS is the Certificate Practice Statement URL.
	CPS string `yaml:"cps"`

	// UserNotice provides a text notice.
	UserNotice string `yaml:"user_notice"`
}

// ExtensionConfig defines additional extensions.
type ExtensionConfig struct {
	// OCSP defines OCSP responder URLs.
	OCSP []string `yaml:"ocsp"`

	// CRLDistributionPoints defines CRL URLs.
	CRLDistributionPoints []string `yaml:"crl_distribution_points"`

	// IssuingCertificateURL defines AIA URLs.
	IssuingCertificateURL []string `yaml:"issuing_certificate_url"`

	// SubjectKeyID controls SKID generation.
	SubjectKeyID SubjectKeyIDConfig `yaml:"subject_key_id"`

	// AuthorityKeyID controls AKID usage.
	AuthorityKeyID AuthorityKeyIDConfig `yaml:"authority_key_id"`
}

// SubjectKeyIDConfig defines subject key identifier generation.
type SubjectKeyIDConfig struct {
	// Method specifies SKID generation method.
	Method string `yaml:"method"` // "hash" or "serial"
}

// AuthorityKeyIDConfig defines authority key identifier usage.
type AuthorityKeyIDConfig struct {
	// Include specifies what to include in AKID.
	Include []string `yaml:"include"` // "key_id", "issuer", "serial"
}

// LoadProfile loads a certificate profile from a YAML file.
func LoadProfile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate profile: %w", err)
	}

	return ParseProfile(data)
}

// ParseProfile parses a certificate profile from YAML data.
func ParseProfile(data []byte) (*Profile, error) {
	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse certificate profile YAML: %w", err)
	}

	if err := profile.Validate(); err != nil {
		return nil, fmt.Errorf("invalid certificate profile: %w", err)
	}

	return &profile, nil
}

// Validate validates the profile configuration.
func (p *Profile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if p.Type == "" {
		return fmt.Errorf("profile type is required")
	}

	if !isValidProfileType(p.Type) {
		return fmt.Errorf("invalid profile type: %s", p.Type)
	}

	// Validate duration if specified.
	if p.Validity.Duration != "" {
		if _, err := time.ParseDuration(p.Validity.Duration); err != nil {
			return fmt.Errorf("invalid validity duration: %w", err)
		}
	}

	if p.Validity.MaxDuration != "" {
		if _, err := time.ParseDuration(p.Validity.MaxDuration); err != nil {
			return fmt.Errorf("invalid max validity duration: %w", err)
		}
	}

	if p.Validity.BackdateBuffer != "" {
		if _, err := time.ParseDuration(p.Validity.BackdateBuffer); err != nil {
			return fmt.Errorf("invalid backdate buffer: %w", err)
		}
	}

	// Validate CA-specific constraints.
	if p.BasicConstraints.IsCA {
		if !p.KeyUsage.CertSign {
			return fmt.Errorf("CA profile must have cert_sign key usage")
		}
	}

	return nil
}

func isValidProfileType(t ProfileType) bool {
	switch t {
	case ProfileTypeRoot, ProfileTypeIntermediate, ProfileTypeIssuing,
		ProfileTypeTLSServer, ProfileTypeTLSClient, ProfileTypeCodeSigning,
		ProfileTypeSMIME, ProfileTypeOCSP, ProfileTypeTSA:
		return true
	default:
		return false
	}
}

// ToX509KeyUsage converts the key usage config to x509.KeyUsage.
func (k *KeyUsageConfig) ToX509KeyUsage() x509.KeyUsage {
	var usage x509.KeyUsage

	if k.DigitalSignature {
		usage |= x509.KeyUsageDigitalSignature
	}

	if k.ContentCommitment {
		usage |= x509.KeyUsageContentCommitment
	}

	if k.KeyEncipherment {
		usage |= x509.KeyUsageKeyEncipherment
	}

	if k.DataEncipherment {
		usage |= x509.KeyUsageDataEncipherment
	}

	if k.KeyAgreement {
		usage |= x509.KeyUsageKeyAgreement
	}

	if k.CertSign {
		usage |= x509.KeyUsageCertSign
	}

	if k.CRLSign {
		usage |= x509.KeyUsageCRLSign
	}

	if k.EncipherOnly {
		usage |= x509.KeyUsageEncipherOnly
	}

	if k.DecipherOnly {
		usage |= x509.KeyUsageDecipherOnly
	}

	return usage
}

// ToX509ExtKeyUsage converts the ext key usage config to []x509.ExtKeyUsage.
func (e *ExtKeyUsageConfig) ToX509ExtKeyUsage() []x509.ExtKeyUsage {
	var usages []x509.ExtKeyUsage

	if e.ServerAuth {
		usages = append(usages, x509.ExtKeyUsageServerAuth)
	}

	if e.ClientAuth {
		usages = append(usages, x509.ExtKeyUsageClientAuth)
	}

	if e.CodeSigning {
		usages = append(usages, x509.ExtKeyUsageCodeSigning)
	}

	if e.EmailProtection {
		usages = append(usages, x509.ExtKeyUsageEmailProtection)
	}

	if e.TimeStamping {
		usages = append(usages, x509.ExtKeyUsageTimeStamping)
	}

	if e.OCSPSigning {
		usages = append(usages, x509.ExtKeyUsageOCSPSigning)
	}

	return usages
}

// GetDuration returns the validity duration.
func (v *ValidityConfig) GetDuration() (time.Duration, error) {
	if v.Duration == "" {
		return 0, fmt.Errorf("duration not specified")
	}

	d, err := time.ParseDuration(v.Duration)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %w", err)
	}

	return d, nil
}

// GetMaxDuration returns the maximum validity duration.
func (v *ValidityConfig) GetMaxDuration() (time.Duration, error) {
	if v.MaxDuration == "" {
		return 0, fmt.Errorf("max duration not specified")
	}

	d, err := time.ParseDuration(v.MaxDuration)
	if err != nil {
		return 0, fmt.Errorf("invalid max duration: %w", err)
	}

	return d, nil
}

// GetBackdateBuffer returns the backdate buffer duration.
func (v *ValidityConfig) GetBackdateBuffer() (time.Duration, error) {
	if v.BackdateBuffer == "" {
		return 0, nil // No backdating by default.
	}

	d, err := time.ParseDuration(v.BackdateBuffer)
	if err != nil {
		return 0, fmt.Errorf("invalid backdate buffer: %w", err)
	}

	return d, nil
}

// ValidateDuration checks if a requested duration is valid.
func (v *ValidityConfig) ValidateDuration(requested time.Duration) error {
	if !v.AllowCustom {
		defaultDuration, err := v.GetDuration()
		if err != nil {
			return err
		}

		if requested != defaultDuration {
			return fmt.Errorf("custom duration not allowed, must use profile default")
		}
	}

	maxDuration, err := v.GetMaxDuration()
	if err == nil && requested > maxDuration {
		return fmt.Errorf("requested duration %v exceeds maximum %v", requested, maxDuration)
	}

	return nil
}
