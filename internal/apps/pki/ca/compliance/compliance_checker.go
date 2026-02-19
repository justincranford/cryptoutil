// Copyright (c) 2025 Justin Cranford

// Package compliance provides CA/Browser Forum compliance validation for certificates.
package compliance

import (
	"crypto/x509"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Framework represents a compliance framework or standard.
func (c *Checker) checkKeySize(cert *x509.Certificate) Requirement {
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

	return Requirement{
		ID:          "BR-6.1.5",
		Framework:   FrameworkCABFBaseline,
		Section:     "6.1.5",
		Title:       "Key Sizes",
		Description: "RSA keys MUST be at least 2048 bits",
		Severity:    SeverityCritical,
		Status:      status,
	}
}

func (c *Checker) checkAlgorithm(cert *x509.Certificate) Requirement {
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

	return Requirement{
		ID:          "BR-7.1.3",
		Framework:   FrameworkCABFBaseline,
		Section:     "7.1.3",
		Title:       "Signature Algorithm",
		Description: "MUST use SHA-256 or stronger",
		Severity:    SeverityCritical,
		Status:      status,
	}
}

func (c *Checker) evaluateSerialNumber5280(cert *x509.Certificate) Status {
	if cert.SerialNumber.Sign() > 0 {
		return StatusCompliant
	}

	return StatusNonCompliant
}

func (c *Checker) evaluateValidity5280(cert *x509.Certificate) Status {
	if cert.NotBefore.Before(cert.NotAfter) {
		return StatusCompliant
	}

	return StatusNonCompliant
}

func (c *Checker) evaluateExtensions5280(cert *x509.Certificate) Status {
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

func (c *Checker) evaluateBasicConstraints5280(cert *x509.Certificate) Status {
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

// GenerateReport generates a compliance report.
func GenerateReport(
	framework Framework,
	requirements []Requirement,
	period AuditPeriod,
	generatedBy string,
) *Report {
	report := &Report{
		ID:           fmt.Sprintf("CR-%d", time.Now().UTC().UnixNano()),
		Framework:    framework,
		GeneratedAt:  time.Now().UTC(),
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
		EffectiveDate: time.Now().UTC(),
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
		evidence.Timestamp = time.Now().UTC()
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
