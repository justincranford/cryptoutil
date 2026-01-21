// Copyright (c) 2025 Justin Cranford

package compliance

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// ComplianceFramework represents a compliance framework or standard.
type ComplianceFramework string

// Supported compliance frameworks.
const (
	FrameworkCABFBaseline ComplianceFramework = "cabf_baseline"
	FrameworkWebTrust     ComplianceFramework = "webtrust"
	FrameworkRFC5280      ComplianceFramework = "rfc5280"
	FrameworkNIST80057    ComplianceFramework = "nist_sp_800_57"
)

// ComplianceStatus represents the status of a compliance check.
type ComplianceStatus string

// Compliance status values.
const (
	StatusCompliant     ComplianceStatus = "compliant"
	StatusNonCompliant  ComplianceStatus = "non_compliant"
	StatusPartial       ComplianceStatus = "partial"
	StatusNotApplicable ComplianceStatus = "not_applicable"
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

// ComplianceRequirement represents a specific compliance requirement.
type ComplianceRequirement struct {
	ID          string              `json:"id"`
	Framework   ComplianceFramework `json:"framework"`
	Section     string              `json:"section"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Severity    Severity            `json:"severity"`
	Status      ComplianceStatus    `json:"status"`
	Evidence    []string            `json:"evidence,omitempty"`
	Notes       string              `json:"notes,omitempty"`
}

// ComplianceReport represents a compliance audit report.
type ComplianceReport struct {
	ID           string                  `json:"id"`
	Framework    ComplianceFramework     `json:"framework"`
	GeneratedAt  time.Time               `json:"generated_at"`
	GeneratedBy  string                  `json:"generated_by"`
	Period       AuditPeriod             `json:"period"`
	Requirements []ComplianceRequirement `json:"requirements"`
	Summary      ComplianceSummary       `json:"summary"`
}

// AuditPeriod represents the time period covered by an audit.
type AuditPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// ComplianceSummary summarizes compliance findings.
type ComplianceSummary struct {
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

// ComplianceChecker validates compliance with various frameworks.
type ComplianceChecker struct {
	framework ComplianceFramework
}

// NewComplianceChecker creates a new compliance checker.
func NewComplianceChecker(framework ComplianceFramework) *ComplianceChecker {
	return &ComplianceChecker{
		framework: framework,
	}
}

// CheckCertificate checks a certificate for compliance.
func (c *ComplianceChecker) CheckCertificate(_ context.Context, cert *x509.Certificate) ([]ComplianceRequirement, error) {
	if cert == nil {
		return nil, errors.New("certificate cannot be nil")
	}

	var requirements []ComplianceRequirement

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
func (c *ComplianceChecker) checkCABFBaseline(cert *x509.Certificate) []ComplianceRequirement {
	requirements := make([]ComplianceRequirement, 0)

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
func (c *ComplianceChecker) checkRFC5280(cert *x509.Certificate) []ComplianceRequirement {
	requirements := make([]ComplianceRequirement, 0)

	// RFC 5280 Section 4.1.2.2 - Serial Number.
	requirements = append(requirements, ComplianceRequirement{
		ID:          "RFC5280-4.1.2.2",
		Framework:   FrameworkRFC5280,
		Section:     "4.1.2.2",
		Title:       "Serial Number",
		Description: "Serial number MUST be a positive integer",
		Severity:    SeverityHigh,
		Status:      c.evaluateSerialNumber5280(cert),
	})

	// RFC 5280 Section 4.1.2.5 - Validity.
	requirements = append(requirements, ComplianceRequirement{
		ID:          "RFC5280-4.1.2.5",
		Framework:   FrameworkRFC5280,
		Section:     "4.1.2.5",
		Title:       "Validity",
		Description: "Validity period must be well-formed",
		Severity:    SeverityMedium,
		Status:      c.evaluateValidity5280(cert),
	})

	// RFC 5280 Section 4.2 - Certificate Extensions.
	requirements = append(requirements, ComplianceRequirement{
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
		requirements = append(requirements, ComplianceRequirement{
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

func (c *ComplianceChecker) checkSubject(cert *x509.Certificate) ComplianceRequirement {
	status := StatusCompliant

	if cert.Subject.CommonName == "" && len(cert.DNSNames) == 0 {
		status = StatusNonCompliant
	}

	return ComplianceRequirement{
		ID:          "BR-7.1.2.1",
		Framework:   FrameworkCABFBaseline,
		Section:     "7.1.2.1",
		Title:       "Subject Information",
		Description: "Subject MUST contain either CN or SAN",
		Severity:    SeverityHigh,
		Status:      status,
	}
}

func (c *ComplianceChecker) checkSerialNumber(cert *x509.Certificate) ComplianceRequirement {
	status := StatusCompliant

	// BR requires at least 64 bits of entropy.
	serialBits := cert.SerialNumber.BitLen()

	const minSerialBits = 64
	if serialBits < minSerialBits {
		status = StatusNonCompliant
	}

	return ComplianceRequirement{
		ID:          "BR-7.1.2.2",
		Framework:   FrameworkCABFBaseline,
		Section:     "7.1.2.2",
		Title:       "Serial Number",
		Description: "Serial number MUST contain at least 64 bits of entropy",
		Severity:    SeverityHigh,
		Status:      status,
	}
}

func (c *ComplianceChecker) checkValidityPeriod(cert *x509.Certificate) ComplianceRequirement {
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

	return ComplianceRequirement{
		ID:          "BR-7.1.2.3",
		Framework:   FrameworkCABFBaseline,
		Section:     "7.1.2.3",
		Title:       "Validity Period",
		Description: "Subscriber certificate validity MUST NOT exceed 398 days",
		Severity:    SeverityHigh,
		Status:      status,
	}
}

func (c *ComplianceChecker) checkExtensions(cert *x509.Certificate) []ComplianceRequirement {
	requirements := make([]ComplianceRequirement, 0)

	// Key Usage.
	keyUsageStatus := StatusCompliant
	if cert.KeyUsage == 0 {
		keyUsageStatus = StatusNonCompliant
	}

	requirements = append(requirements, ComplianceRequirement{
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

		requirements = append(requirements, ComplianceRequirement{
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

	requirements = append(requirements, ComplianceRequirement{
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

func (c *ComplianceChecker) checkKeySize(cert *x509.Certificate) ComplianceRequirement {
	status := StatusCompliant

	// Check RSA key size using type assertion.
	type rsaPublicKey interface {
		Size() int
	}

	if rsaKey, ok := cert.PublicKey.(rsaPublicKey); ok {
		keyBits := rsaKey.Size() * bitsPerByte
		if keyBits < minRSAKeyBits {
			status = StatusNonCompliant
		}
	}

	return ComplianceRequirement{
		ID:          "BR-6.1.5",
		Framework:   FrameworkCABFBaseline,
		Section:     "6.1.5",
		Title:       "Key Sizes",
		Description: "RSA keys MUST be at least 2048 bits",
		Severity:    SeverityCritical,
		Status:      status,
	}
}

func (c *ComplianceChecker) checkAlgorithm(cert *x509.Certificate) ComplianceRequirement {
	status := StatusCompliant

	// Check for weak algorithms.
	weakAlgorithms := map[x509.SignatureAlgorithm]bool{
		x509.MD2WithRSA:    true,
		x509.MD5WithRSA:    true,
		x509.SHA1WithRSA:   true,
		x509.DSAWithSHA1:   true,
		x509.DSAWithSHA256: true,
		x509.ECDSAWithSHA1: true,
	}

	if weakAlgorithms[cert.SignatureAlgorithm] {
		status = StatusNonCompliant
	}

	return ComplianceRequirement{
		ID:          "BR-7.1.3",
		Framework:   FrameworkCABFBaseline,
		Section:     "7.1.3",
		Title:       "Signature Algorithm",
		Description: "MUST use SHA-256 or stronger",
		Severity:    SeverityCritical,
		Status:      status,
	}
}

func (c *ComplianceChecker) evaluateSerialNumber5280(cert *x509.Certificate) ComplianceStatus {
	if cert.SerialNumber.Sign() > 0 {
		return StatusCompliant
	}

	return StatusNonCompliant
}

func (c *ComplianceChecker) evaluateValidity5280(cert *x509.Certificate) ComplianceStatus {
	if cert.NotBefore.Before(cert.NotAfter) {
		return StatusCompliant
	}

	return StatusNonCompliant
}

func (c *ComplianceChecker) evaluateExtensions5280(cert *x509.Certificate) ComplianceStatus {
	// Check for critical extensions that we don't understand.
	for _, ext := range cert.Extensions {
		if ext.Critical {
			// All critical extensions should be recognized.
			if !isKnownExtension(ext.Id.String()) {
				return StatusNonCompliant
			}
		}
	}

	return StatusCompliant
}

func (c *ComplianceChecker) evaluateBasicConstraints5280(cert *x509.Certificate) ComplianceStatus {
	if cert.BasicConstraintsValid && cert.IsCA {
		return StatusCompliant
	}

	return StatusNonCompliant
}

// isKnownExtension checks if an OID is a known extension.
func isKnownExtension(oid string) bool {
	knownOIDs := map[string]bool{
		"2.5.29.14":         true, // Subject Key Identifier.
		"2.5.29.15":         true, // Key Usage.
		"2.5.29.17":         true, // Subject Alternative Name.
		"2.5.29.19":         true, // Basic Constraints.
		"2.5.29.31":         true, // CRL Distribution Points.
		"2.5.29.32":         true, // Certificate Policies.
		"2.5.29.35":         true, // Authority Key Identifier.
		"2.5.29.37":         true, // Extended Key Usage.
		"1.3.6.1.5.5.7.1.1": true, // Authority Information Access.
	}

	return knownOIDs[oid]
}

// GenerateComplianceReport generates a compliance report.
func GenerateComplianceReport(
	framework ComplianceFramework,
	requirements []ComplianceRequirement,
	period AuditPeriod,
	generatedBy string,
) *ComplianceReport {
	report := &ComplianceReport{
		ID:           fmt.Sprintf("CR-%d", time.Now().UnixNano()),
		Framework:    framework,
		GeneratedAt:  time.Now(),
		GeneratedBy:  generatedBy,
		Period:       period,
		Requirements: requirements,
	}

	// Calculate summary.
	for _, req := range requirements {
		report.Summary.TotalRequirements++

		switch req.Status {
		case StatusCompliant:
			report.Summary.Compliant++
		case StatusNonCompliant:
			report.Summary.NonCompliant++

			switch req.Severity {
			case SeverityCritical:
				report.Summary.CriticalFindings++
			case SeverityHigh:
				report.Summary.HighFindings++
			case SeverityMedium:
				report.Summary.MediumFindings++
			case SeverityLow:
				report.Summary.LowFindings++
			case SeverityInfo:
				// Info severity findings are not counted separately.
			}
		case StatusPartial:
			report.Summary.Partial++
		case StatusNotApplicable:
			report.Summary.NotApplicable++
		}
	}

	return report
}

// AuditTrail represents a complete audit trail for a period.
type AuditTrail struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Events    []*AuditEvent `json:"events"`
	Summary   AuditSummary  `json:"summary"`
}

// AuditSummary summarizes audit events.
type AuditSummary struct {
	TotalEvents           int            `json:"total_events"`
	EventsByType          map[string]int `json:"events_by_type"`
	CertificatesIssued    int            `json:"certificates_issued"`
	CertificatesRevoked   int            `json:"certificates_revoked"`
	FailedAuthentications int            `json:"failed_authentications"`
}

// AuditTrailBuilder builds an audit trail.
type AuditTrailBuilder struct {
	trail *AuditTrail
}

// NewAuditTrailBuilder creates a new audit trail builder.
func NewAuditTrailBuilder(startTime, endTime time.Time) *AuditTrailBuilder {
	return &AuditTrailBuilder{
		trail: &AuditTrail{
			StartTime: startTime,
			EndTime:   endTime,
			Events:    make([]*AuditEvent, 0),
			Summary: AuditSummary{
				EventsByType: make(map[string]int),
			},
		},
	}
}

// AddEvent adds an event to the audit trail.
func (b *AuditTrailBuilder) AddEvent(event *AuditEvent) *AuditTrailBuilder {
	if event == nil {
		return b
	}

	b.trail.Events = append(b.trail.Events, event)
	b.trail.Summary.TotalEvents++
	b.trail.Summary.EventsByType[string(event.EventType)]++

	// Update specific counters.
	switch event.EventType {
	case EventCertificateIssued:
		b.trail.Summary.CertificatesIssued++
	case EventCertificateRevoked:
		b.trail.Summary.CertificatesRevoked++
	case EventAuthenticationFailed:
		b.trail.Summary.FailedAuthentications++
	case EventCertificateRenewed, EventKeyGenerated, EventKeyDestroyed,
		EventCSRReceived, EventCSRApproved, EventCSRRejected,
		EventCRLGenerated, EventOCSPResponseIssued, EventConfigChanged,
		EventAccessGranted, EventAccessRevoked, EventSystemStartup, EventSystemShutdown:
		// These events are tracked in EventsByType map but don't have dedicated counters.
	}

	return b
}

// Build returns the constructed audit trail.
func (b *AuditTrailBuilder) Build() *AuditTrail {
	return b.trail
}

// PolicyDocument represents a CA policy document.
type PolicyDocument struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	Version       string          `json:"version"`
	EffectiveDate time.Time       `json:"effective_date"`
	Sections      []PolicySection `json:"sections"`
}

// PolicySection represents a section of a policy document.
type PolicySection struct {
	Number  string `json:"number"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// CreateCPSTemplate creates a Certificate Practice Statement template.
func CreateCPSTemplate() *PolicyDocument {
	return &PolicyDocument{
		ID:            "CPS-001",
		Title:         "Certificate Practice Statement",
		Version:       "1.0",
		EffectiveDate: time.Now(),
		Sections: []PolicySection{
			{Number: "1", Title: "Introduction", Content: "This CPS describes the practices..."},
			{Number: "1.1", Title: "Overview", Content: "The Certificate Authority..."},
			{Number: "2", Title: "Publication and Repository", Content: "Certificates and CRLs are published..."},
			{Number: "3", Title: "Identification and Authentication", Content: "Subscriber identification..."},
			{Number: "4", Title: "Certificate Life-Cycle", Content: "Certificate application, issuance..."},
			{Number: "5", Title: "Management and Operations", Content: "Physical controls, procedural..."},
			{Number: "6", Title: "Technical Security Controls", Content: "Key generation, activation..."},
			{Number: "7", Title: "Certificate Profiles", Content: "Certificate content, extensions..."},
			{Number: "8", Title: "Compliance Audit", Content: "Audit frequency, scope..."},
			{Number: "9", Title: "Other Business and Legal", Content: "Fees, liability, disputes..."},
		},
	}
}

// EvidenceCollector collects evidence for compliance audits.
type EvidenceCollector struct {
	evidence []Evidence
	mu       sync.Mutex
}

// Evidence represents compliance evidence.
type Evidence struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Source      string    `json:"source"`
	Data        string    `json:"data,omitempty"`
}

// NewEvidenceCollector creates a new evidence collector.
func NewEvidenceCollector() *EvidenceCollector {
	return &EvidenceCollector{
		evidence: make([]Evidence, 0),
	}
}

// Collect adds evidence to the collector.
func (c *EvidenceCollector) Collect(evidence Evidence) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if evidence.Timestamp.IsZero() {
		evidence.Timestamp = time.Now()
	}

	c.evidence = append(c.evidence, evidence)
}

// GetEvidence returns all collected evidence.
func (c *EvidenceCollector) GetEvidence() []Evidence {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]Evidence, len(c.evidence))
	copy(result, c.evidence)

	return result
}

// GetEvidenceByType returns evidence filtered by type.
func (c *EvidenceCollector) GetEvidenceByType(evidenceType string) []Evidence {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]Evidence, 0)

	for _, e := range c.evidence {
		if strings.EqualFold(e.Type, evidenceType) {
			result = append(result, e)
		}
	}

	return result
}
