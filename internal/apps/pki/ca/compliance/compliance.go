// Copyright (c) 2025 Justin Cranford

// Package compliance provides CA/Browser Forum compliance validation for certificates.
package compliance

import (
	"context"
	"crypto/x509"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Framework represents a compliance framework or standard.
type Framework string

// Supported compliance frameworks.
const (
	FrameworkCABFBaseline Framework = "cabf_baseline"
	FrameworkWebTrust     Framework = "webtrust"
	FrameworkRFC5280      Framework = "rfc5280"
	FrameworkNIST80057    Framework = "nist_sp_800_57"
)

// Status represents the status of a compliance check.
type Status string

// Compliance status values.
const (
	StatusCompliant     Status = "compliant"
	StatusNonCompliant  Status = "non_compliant"
	StatusPartial       Status = "partial"
	StatusNotApplicable Status = "not_applicable"
)

// Severity levels for compliance findings.
type Severity string

// Severity constants.
const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// AuditEventType represents types of audit events.
type AuditEventType string

// Audit event types.
const (
	EventCertificateIssued    AuditEventType = "certificate_issued"
	EventCertificateRevoked   AuditEventType = "certificate_revoked"
	EventCertificateRenewed   AuditEventType = "certificate_renewed"
	EventKeyGenerated         AuditEventType = "key_generated"
	EventKeyDestroyed         AuditEventType = "key_destroyed"
	EventCSRReceived          AuditEventType = "csr_received"
	EventCSRApproved          AuditEventType = "csr_approved"
	EventCSRRejected          AuditEventType = "csr_rejected"
	EventCRLGenerated         AuditEventType = "crl_generated"
	EventOCSPResponseIssued   AuditEventType = "ocsp_response_issued"
	EventConfigChanged        AuditEventType = "config_changed"
	EventAccessGranted        AuditEventType = "access_granted"
	EventAccessRevoked        AuditEventType = "access_revoked"
	EventAuthenticationFailed AuditEventType = "authentication_failed"
	EventSystemStartup        AuditEventType = "system_startup"
	EventSystemShutdown       AuditEventType = "system_shutdown"
)

// AuditEvent represents an auditable event in the CA system.
type AuditEvent struct {
	ID            string            `json:"id"`
	Timestamp     time.Time         `json:"timestamp"`
	EventType     AuditEventType    `json:"event_type"`
	Actor         string            `json:"actor"`
	ActorRole     string            `json:"actor_role,omitempty"`
	Resource      string            `json:"resource"`
	ResourceType  string            `json:"resource_type,omitempty"`
	Action        string            `json:"action"`
	Outcome       string            `json:"outcome"`
	Details       map[string]string `json:"details,omitempty"`
	ClientIP      string            `json:"client_ip,omitempty"`
	SessionID     string            `json:"session_id,omitempty"`
	CorrelationID string            `json:"correlation_id,omitempty"`
}

// Requirement represents a specific compliance requirement.
type Requirement struct {
	ID          string    `json:"id"`
	Framework   Framework `json:"framework"`
	Section     string    `json:"section"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    Severity  `json:"severity"`
	Status      Status    `json:"status"`
	Evidence    []string  `json:"evidence,omitempty"`
	Notes       string    `json:"notes,omitempty"`
}

// Report represents a compliance audit report.
type Report struct {
	ID           string        `json:"id"`
	Framework    Framework     `json:"framework"`
	GeneratedAt  time.Time     `json:"generated_at"`
	GeneratedBy  string        `json:"generated_by"`
	Period       AuditPeriod   `json:"period"`
	Requirements []Requirement `json:"requirements"`
	Summary      Summary       `json:"summary"`
}

// AuditPeriod represents the time period covered by an audit.
type AuditPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// Summary summarizes compliance findings.
type Summary struct {
	TotalRequirements int `json:"total_requirements"`
	Compliant         int `json:"compliant"`
	NonCompliant      int `json:"non_compliant"`
	Partial           int `json:"partial"`
	NotApplicable     int `json:"not_applicable"`
	CriticalFindings  int `json:"critical_findings"`
	HighFindings      int `json:"high_findings"`
	MediumFindings    int `json:"medium_findings"`
	LowFindings       int `json:"low_findings"`
}

// AuditLogger logs audit events to configured outputs.
type AuditLogger struct {
	writers []io.Writer
	mu      sync.Mutex
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger(writers ...io.Writer) *AuditLogger {
	return &AuditLogger{
		writers: writers,
	}
}

// Log logs an audit event.
func (l *AuditLogger) Log(_ context.Context, event *AuditEvent) error {
	if event == nil {
		return errors.New("audit event cannot be nil")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Ensure timestamp is set.
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Serialize event to JSON.
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	data = append(data, '\n')

	// Write to all configured writers.
	for _, w := range l.writers {
		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("failed to write audit event: %w", err)
		}
	}

	return nil
}

// AddWriter adds a writer to the audit logger.
func (l *AuditLogger) AddWriter(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.writers = append(l.writers, w)
}

// Checker validates compliance with various frameworks.
type Checker struct {
	framework Framework
}

// NewChecker creates a new compliance checker.
func NewChecker(framework Framework) *Checker {
	return &Checker{
		framework: framework,
	}
}

// CheckCertificate checks a certificate for compliance.
func (c *Checker) CheckCertificate(_ context.Context, cert *x509.Certificate) ([]Requirement, error) {
	if cert == nil {
		return nil, errors.New("certificate cannot be nil")
	}

	var requirements []Requirement

	switch c.framework {
	case FrameworkCABFBaseline:
		requirements = c.checkCABFBaseline(cert)
	case FrameworkRFC5280:
		requirements = c.checkRFC5280(cert)
	case FrameworkWebTrust, FrameworkNIST80057:
		// WebTrust and NIST frameworks fallback to CABF baseline checks.
		requirements = c.checkCABFBaseline(cert)
	}

	return requirements, nil
}

// checkCABFBaseline checks CA/Browser Forum Baseline Requirements.
func (c *Checker) checkCABFBaseline(cert *x509.Certificate) []Requirement {
	requirements := make([]Requirement, 0)

	// BR Section 7.1.2.1 - Subject
	requirements = append(requirements, c.checkSubject(cert))

	// BR Section 7.1.2.2 - Serial Number.
	requirements = append(requirements, c.checkSerialNumber(cert))

	// BR Section 7.1.2.3 - Validity Period.
	requirements = append(requirements, c.checkValidityPeriod(cert))

	// BR Section 7.1.2.4 - Extensions.
	requirements = append(requirements, c.checkExtensions(cert)...)

	// BR Section 6.1.5 - Key Sizes.
	requirements = append(requirements, c.checkKeySize(cert))

	// BR Section 7.1.3 - Algorithm Object Identifiers.
	requirements = append(requirements, c.checkAlgorithm(cert))

	return requirements
}

// checkRFC5280 checks RFC 5280 compliance.
func (c *Checker) checkRFC5280(cert *x509.Certificate) []Requirement {
	requirements := make([]Requirement, 0)

	// RFC 5280 Section 4.1.2.2 - Serial Number.
	requirements = append(requirements, Requirement{
		ID:          "RFC5280-4.1.2.2",
		Framework:   FrameworkRFC5280,
		Section:     "4.1.2.2",
		Title:       "Serial Number",
		Description: "Serial number MUST be a positive integer",
		Severity:    SeverityHigh,
		Status:      c.evaluateSerialNumber5280(cert),
	})

	// RFC 5280 Section 4.1.2.5 - Validity.
	requirements = append(requirements, Requirement{
		ID:          "RFC5280-4.1.2.5",
		Framework:   FrameworkRFC5280,
		Section:     "4.1.2.5",
		Title:       "Validity",
		Description: "Validity period must be well-formed",
		Severity:    SeverityMedium,
		Status:      c.evaluateValidity5280(cert),
	})

	// RFC 5280 Section 4.2 - Certificate Extensions.
	requirements = append(requirements, Requirement{
		ID:          "RFC5280-4.2",
		Framework:   FrameworkRFC5280,
		Section:     "4.2",
		Title:       "Certificate Extensions",
		Description: "Extensions MUST be handled correctly",
		Severity:    SeverityMedium,
		Status:      c.evaluateExtensions5280(cert),
	})

	// RFC 5280 Section 4.2.1.9 - Basic Constraints.
	if cert.IsCA {
		requirements = append(requirements, Requirement{
			ID:          "RFC5280-4.2.1.9",
			Framework:   FrameworkRFC5280,
			Section:     "4.2.1.9",
			Title:       "Basic Constraints",
			Description: "CA certificates MUST have Basic Constraints",
			Severity:    SeverityCritical,
			Status:      c.evaluateBasicConstraints5280(cert),
		})
	}

	return requirements
}

func (c *Checker) checkSubject(cert *x509.Certificate) Requirement {
	status := StatusCompliant

	if cert.Subject.CommonName == "" && len(cert.DNSNames) == 0 {
		status = StatusNonCompliant
	}

	return Requirement{
		ID:          "BR-7.1.2.1",
		Framework:   FrameworkCABFBaseline,
		Section:     "7.1.2.1",
		Title:       "Subject Information",
		Description: "Subject MUST contain either CN or SAN",
		Severity:    SeverityHigh,
		Status:      status,
	}
}

func (c *Checker) checkSerialNumber(cert *x509.Certificate) Requirement {
	status := StatusCompliant

	// BR requires at least 64 bits of entropy.
	serialBits := cert.SerialNumber.BitLen()

	if serialBits < cryptoutilSharedMagic.MinSerialNumberBits {
		status = StatusNonCompliant
	}

	return Requirement{
		ID:          "BR-7.1.2.2",
		Framework:   FrameworkCABFBaseline,
		Section:     "7.1.2.2",
		Title:       "Serial Number",
		Description: "Serial number MUST contain at least 64 bits of entropy",
		Severity:    SeverityHigh,
		Status:      status,
	}
}

func (c *Checker) checkValidityPeriod(cert *x509.Certificate) Requirement {
	status := StatusCompliant

	// BR limits subscriber certificates to 398 days.
	const (
		maxValidityDays = 398
		hoursPerDay     = 24
	)

	validityDays := int(cert.NotAfter.Sub(cert.NotBefore).Hours() / hoursPerDay)

	if !cert.IsCA && validityDays > maxValidityDays {
		status = StatusNonCompliant
	}

	return Requirement{
		ID:          "BR-7.1.2.3",
		Framework:   FrameworkCABFBaseline,
		Section:     "7.1.2.3",
		Title:       "Validity Period",
		Description: "Subscriber certificate validity MUST NOT exceed 398 days",
		Severity:    SeverityHigh,
		Status:      status,
	}
}

func (c *Checker) checkExtensions(cert *x509.Certificate) []Requirement {
	requirements := make([]Requirement, 0)

	// Key Usage.
	keyUsageStatus := StatusCompliant
	if cert.KeyUsage == 0 {
		keyUsageStatus = StatusNonCompliant
	}

	requirements = append(requirements, Requirement{
		ID:          "BR-7.1.2.4-KU",
		Framework:   FrameworkCABFBaseline,
		Section:     "7.1.2.4",
		Title:       "Key Usage Extension",
		Description: "Key Usage extension SHOULD be present",
		Severity:    SeverityMedium,
		Status:      keyUsageStatus,
	})

	// Basic Constraints for CA certificates.
	if cert.IsCA {
		bcStatus := StatusCompliant
		if !cert.BasicConstraintsValid {
			bcStatus = StatusNonCompliant
		}

		requirements = append(requirements, Requirement{
			ID:          "BR-7.1.2.4-BC",
			Framework:   FrameworkCABFBaseline,
			Section:     "7.1.2.4",
			Title:       "Basic Constraints Extension",
			Description: "CA certificates MUST have Basic Constraints",
			Severity:    SeverityCritical,
			Status:      bcStatus,
		})
	}

	// Subject Alternative Name.
	sanStatus := StatusCompliant

	hasSAN := len(cert.DNSNames) > 0 || len(cert.EmailAddresses) > 0 ||
		len(cert.IPAddresses) > 0 || len(cert.URIs) > 0

	if !cert.IsCA && !hasSAN {
		sanStatus = StatusNonCompliant
	}

	requirements = append(requirements, Requirement{
		ID:          "BR-7.1.2.4-SAN",
		Framework:   FrameworkCABFBaseline,
		Section:     "7.1.2.4",
		Title:       "Subject Alternative Name Extension",
		Description: "Subscriber certificates MUST have SAN",
		Severity:    SeverityHigh,
		Status:      sanStatus,
	})

	return requirements
}

const (
	minRSAKeyBits = 2048
	minECKeyBits  = 256
	bitsPerByte   = 8
)
