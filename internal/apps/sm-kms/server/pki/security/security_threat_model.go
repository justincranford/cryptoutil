// Copyright (c) 2025-2026 Justin Cranford.
// Package security provides certificate security validation and compliance checking.
package security

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ThreatCategory represents a category in the STRIDE threat model.
func NewThreatModelBuilder(name, version string) *ThreatModelBuilder {
	return &ThreatModelBuilder{
		model: &ThreatModel{
			Name:      name,
			Version:   version,
			Assets:    make([]Asset, 0),
			Threats:   make([]Threat, 0),
			Controls:  make([]Control, 0),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}
}

// WithDescription sets the threat model description.
func (b *ThreatModelBuilder) WithDescription(desc string) *ThreatModelBuilder {
	b.model.Description = desc

	return b
}

// AddAsset adds an asset to the threat model.
func (b *ThreatModelBuilder) AddAsset(asset Asset) *ThreatModelBuilder {
	b.model.Assets = append(b.model.Assets, asset)
	b.model.UpdatedAt = time.Now().UTC()

	return b
}

// AddThreat adds a threat to the threat model.
func (b *ThreatModelBuilder) AddThreat(threat Threat) *ThreatModelBuilder {
	b.model.Threats = append(b.model.Threats, threat)
	b.model.UpdatedAt = time.Now().UTC()

	return b
}

// AddControl adds a security control to the threat model.
func (b *ThreatModelBuilder) AddControl(control Control) *ThreatModelBuilder {
	b.model.Controls = append(b.model.Controls, control)
	b.model.UpdatedAt = time.Now().UTC()

	return b
}

// Build returns the constructed threat model.
func (b *ThreatModelBuilder) Build() *ThreatModel {
	return b.model
}

// CAThreatModel creates a predefined threat model for CA operations.
func CAThreatModel() *ThreatModel {
	builder := NewThreatModelBuilder("CA Security Threat Model", cryptoutilSharedMagic.ServiceVersion)

	builder.WithDescription("STRIDE-based threat model for Certificate Authority operations")

	// Add assets.
	builder.AddAsset(Asset{
		ID:          cryptoutilSharedMagic.PKIAssetIDRootCAKey,
		Name:        "Root CA Private Key",
		Description: "Private key for the root certificate authority",
		Type:        cryptoutilSharedMagic.PKIThreatTypeCryptographicKey,
		Sensitivity: cryptoutilSharedMagic.ThreatLevelCritical,
	})

	builder.AddAsset(Asset{
		ID:          cryptoutilSharedMagic.PKIAssetIDIntermediateCAKey,
		Name:        "Intermediate CA Private Keys",
		Description: "Private keys for intermediate certificate authorities",
		Type:        cryptoutilSharedMagic.PKIThreatTypeCryptographicKey,
		Sensitivity: cryptoutilSharedMagic.ThreatLevelHigh,
	})

	builder.AddAsset(Asset{
		ID:          cryptoutilSharedMagic.PKIAssetIDCertificateDatabase,
		Name:        "Certificate Database",
		Description: "Database storing issued certificates and revocation status",
		Type:        cryptoutilSharedMagic.RealmStorageTypeDatabase,
		Sensitivity: cryptoutilSharedMagic.ThreatLevelHigh,
	})

	builder.AddAsset(Asset{
		ID:          cryptoutilSharedMagic.PKIAssetIDAuditLogs,
		Name:        "Audit Logs",
		Description: "Logs of all CA operations for compliance",
		Type:        "log_data",
		Sensitivity: cryptoutilSharedMagic.ThreatLevelMedium,
	})

	// Add STRIDE threats.
	builder.AddThreat(Threat{
		ID:          cryptoutilSharedMagic.PKIThreatIDSpoofing,
		Category:    ThreatSpoofing,
		Title:       "Unauthorized Certificate Issuance",
		Description: "Attacker issues certificates without proper authorization",
		Asset:       cryptoutilSharedMagic.PKIAssetIDIntermediateCAKey,
		Severity:    SeverityCritical,
		Likelihood:  cryptoutilSharedMagic.ThreatLevelMedium,
		Impact:      cryptoutilSharedMagic.ThreatLevelCritical,
		Mitigations: []string{"Multi-party approval", "HSM-protected keys", cryptoutilSharedMagic.PKIThreatMitigationAuditLogging},
		Status:      cryptoutilSharedMagic.PKIThreatStatusMitigated,
	})

	builder.AddThreat(Threat{
		ID:          cryptoutilSharedMagic.PKIThreatIDTampering,
		Category:    ThreatTampering,
		Title:       "Certificate Database Tampering",
		Description: "Attacker modifies certificate records in the database",
		Asset:       cryptoutilSharedMagic.PKIAssetIDCertificateDatabase,
		Severity:    SeverityHigh,
		Likelihood:  cryptoutilSharedMagic.ThreatLevelLow,
		Impact:      cryptoutilSharedMagic.ThreatLevelHigh,
		Mitigations: []string{"Database integrity checks", cryptoutilSharedMagic.PKIThreatMitigationAccessControls, cryptoutilSharedMagic.PKIThreatMitigationAuditLogging},
		Status:      cryptoutilSharedMagic.PKIThreatStatusMitigated,
	})

	builder.AddThreat(Threat{
		ID:          cryptoutilSharedMagic.PKIThreatIDRepudiation,
		Category:    ThreatRepudiation,
		Title:       "Denied Certificate Operations",
		Description: "Operator denies performing certificate operations",
		Asset:       cryptoutilSharedMagic.PKIAssetIDAuditLogs,
		Severity:    SeverityMedium,
		Likelihood:  cryptoutilSharedMagic.ThreatLevelMedium,
		Impact:      cryptoutilSharedMagic.ThreatLevelMedium,
		Mitigations: []string{"Immutable audit logs", "Digital signatures on logs", "Log forwarding"},
		Status:      cryptoutilSharedMagic.PKIThreatStatusMitigated,
	})

	builder.AddThreat(Threat{
		ID:          cryptoutilSharedMagic.PKIThreatIDInformationDisclosure,
		Category:    ThreatInformationDisclose,
		Title:       "Private Key Exposure",
		Description: "CA private keys are disclosed to unauthorized parties",
		Asset:       cryptoutilSharedMagic.PKIAssetIDRootCAKey,
		Severity:    SeverityCritical,
		Likelihood:  cryptoutilSharedMagic.ThreatLevelLow,
		Impact:      cryptoutilSharedMagic.ThreatLevelCritical,
		Mitigations: []string{"HSM storage", "Key ceremony procedures", cryptoutilSharedMagic.PKIThreatMitigationAccessControls},
		Status:      cryptoutilSharedMagic.PKIThreatStatusMitigated,
	})

	builder.AddThreat(Threat{
		ID:          cryptoutilSharedMagic.PKIThreatIDDenialOfService,
		Category:    ThreatDenialOfService,
		Title:       "CA Service Unavailability",
		Description: "CA services become unavailable due to attack or failure",
		Asset:       cryptoutilSharedMagic.PKIAssetIDCertificateDatabase,
		Severity:    SeverityMedium,
		Likelihood:  cryptoutilSharedMagic.ThreatLevelMedium,
		Impact:      cryptoutilSharedMagic.ThreatLevelMedium,
		Mitigations: []string{"High availability deployment", "Rate limiting", "DDoS protection"},
		Status:      cryptoutilSharedMagic.PKIThreatStatusMitigated,
	})

	builder.AddThreat(Threat{
		ID:          cryptoutilSharedMagic.PKIThreatIDElevationOfPrivilege,
		Category:    ThreatElevationPrivilege,
		Title:       "Privilege Escalation to CA Admin",
		Description: "Attacker gains CA administrator privileges",
		Asset:       cryptoutilSharedMagic.PKIAssetIDIntermediateCAKey,
		Severity:    SeverityCritical,
		Likelihood:  cryptoutilSharedMagic.ThreatLevelLow,
		Impact:      cryptoutilSharedMagic.ThreatLevelCritical,
		Mitigations: []string{"Role-based access control", "MFA", "Separation of duties"},
		Status:      cryptoutilSharedMagic.PKIThreatStatusMitigated,
	})

	// Add controls.
	builder.AddControl(Control{
		ID:          "CTRL-001",
		Name:        "HSM Key Storage",
		Description: "Store CA private keys in Hardware Security Modules",
		Type:        cryptoutilSharedMagic.PKIControlTypeTechnical,
		Mitigates:   []string{cryptoutilSharedMagic.PKIThreatIDInformationDisclosure, cryptoutilSharedMagic.PKIThreatIDSpoofing},
		Status:      cryptoutilSharedMagic.PKIThreatStatusImplemented,
	})

	builder.AddControl(Control{
		ID:          "CTRL-002",
		Name:        "Audit Logging",
		Description: "Log all CA operations with integrity protection",
		Type:        cryptoutilSharedMagic.PKIControlTypeTechnical,
		Mitigates:   []string{cryptoutilSharedMagic.PKIThreatIDRepudiation, cryptoutilSharedMagic.PKIThreatIDTampering},
		Status:      cryptoutilSharedMagic.PKIThreatStatusImplemented,
	})

	builder.AddControl(Control{
		ID:          "CTRL-003",
		Name:        "Multi-Party Approval",
		Description: "Require multiple approvers for sensitive operations",
		Type:        "procedural",
		Mitigates:   []string{cryptoutilSharedMagic.PKIThreatIDSpoofing, cryptoutilSharedMagic.PKIThreatIDElevationOfPrivilege},
		Status:      cryptoutilSharedMagic.PKIThreatStatusImplemented,
	})

	builder.AddControl(Control{
		ID:          "CTRL-004",
		Name:        "Rate Limiting",
		Description: "Limit request rates to prevent abuse",
		Type:        cryptoutilSharedMagic.PKIControlTypeTechnical,
		Mitigates:   []string{cryptoutilSharedMagic.PKIThreatIDDenialOfService},
		Status:      cryptoutilSharedMagic.PKIThreatStatusImplemented,
	})

	return builder.Build()
}

// Scanner performs security scans on CA components.
type Scanner struct {
	validator *Validator
}

// NewScanner creates a new security scanner.
func NewScanner(config *Config) *Scanner {
	return &Scanner{
		validator: NewValidator(config),
	}
}

// ScanCertificateChain validates an entire certificate chain.
func (s *Scanner) ScanCertificateChain(ctx context.Context, chain []*x509.Certificate) (*ValidationResult, error) {
	if len(chain) == 0 {
		return nil, errors.New("certificate chain cannot be empty")
	}

	combinedResult := &ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
		CheckedAt: time.Now().UTC(),
	}

	// Validate each certificate in the chain.
	for i, cert := range chain {
		result, err := s.validator.ValidateCertificate(ctx, cert)
		if err != nil {
			return nil, fmt.Errorf("failed to validate certificate %d: %w", i, err)
		}

		if !result.Valid {
			combinedResult.Valid = false
		}

		for _, e := range result.Errors {
			combinedResult.Errors = append(combinedResult.Errors, fmt.Sprintf("[cert %d] %s", i, e))
		}

		for _, w := range result.Warnings {
			combinedResult.Warnings = append(combinedResult.Warnings, fmt.Sprintf("[cert %d] %s", i, w))
		}

		combinedResult.Vulnerabilities = append(combinedResult.Vulnerabilities, result.Vulnerabilities...)
	}

	// Validate chain linkage.
	s.validateChainLinkage(chain, combinedResult)

	return combinedResult, nil
}

// validateChainLinkage validates that certificates in the chain are properly linked.
func (s *Scanner) validateChainLinkage(chain []*x509.Certificate, result *ValidationResult) {
	for i := 0; i < len(chain)-1; i++ {
		child := chain[i]
		parent := chain[i+1]

		// Verify that child was signed by parent.
		if err := child.CheckSignatureFrom(parent); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("certificate %d not signed by certificate %d: %v", i, i+1, err))
		}

		// Verify issuer/subject match.
		if child.Issuer.String() != parent.Subject.String() {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("certificate %d issuer does not match certificate %d subject", i, i+1))
		}
	}
}

// Report generates a comprehensive security report.
type Report struct {
	GeneratedAt time.Time          `json:"generated_at"`
	ThreatModel *ThreatModel       `json:"threat_model,omitempty"`
	Validations []ValidationResult `json:"validations,omitempty"`
	Summary     Summary            `json:"summary"`
}

// Summary summarizes security findings.
type Summary struct {
	TotalThreats         int `json:"total_threats"`
	MitigatedThreats     int `json:"mitigated_threats"`
	OpenThreats          int `json:"open_threats"`
	TotalVulnerabilities int `json:"total_vulnerabilities"`
	CriticalCount        int `json:"critical_count"`
	HighCount            int `json:"high_count"`
	MediumCount          int `json:"medium_count"`
	LowCount             int `json:"low_count"`
	InfoCount            int `json:"info_count"`
}

// GenerateReport creates a security report from threat model and validations.
func GenerateReport(threatModel *ThreatModel, validations []ValidationResult) *Report {
	report := &Report{
		GeneratedAt: time.Now().UTC(),
		ThreatModel: threatModel,
		Validations: validations,
	}

	// Calculate summary.
	if threatModel != nil {
		report.Summary.TotalThreats = len(threatModel.Threats)

		for _, threat := range threatModel.Threats {
			if threat.Status == cryptoutilSharedMagic.PKIThreatStatusMitigated {
				report.Summary.MitigatedThreats++
			} else {
				report.Summary.OpenThreats++
			}
		}
	}

	// Count vulnerabilities.
	for _, v := range validations {
		report.Summary.TotalVulnerabilities += len(v.Vulnerabilities)

		for _, vuln := range v.Vulnerabilities {
			switch vuln.Severity {
			case SeverityCritical:
				report.Summary.CriticalCount++
			case SeverityHigh:
				report.Summary.HighCount++
			case SeverityMedium:
				report.Summary.MediumCount++
			case SeverityLow:
				report.Summary.LowCount++
			case SeverityInfo:
				report.Summary.InfoCount++
			}
		}
	}

	return report
}
