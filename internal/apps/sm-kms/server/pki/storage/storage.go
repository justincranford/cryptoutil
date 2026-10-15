// Copyright (c) 2025-2026 Justin Cranford.
// Package storage provides certificate storage abstractions and implementations.
package storage

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Common errors.
var (
	ErrCertificateNotFound  = errors.New("certificate not found")
	ErrCertificateExists    = errors.New("certificate already exists")
	ErrCertificateAmbiguous = errors.New("certificate lookup is ambiguous")
	ErrCertificateRevoked   = errors.New("certificate already revoked")
	ErrInvalidCertificate   = errors.New("invalid certificate")
	ErrStorageFull          = errors.New("storage capacity exceeded")
	ErrAuthorityNotFound    = errors.New("PKI authority not found")
	ErrMultipleAuthorities  = errors.New("multiple PKI authorities are persisted")
	ErrInvalidIssuerKeyID   = errors.New("issuer key identifier is required")
)

// CertificateStatus represents the status of a stored certificate.
type CertificateStatus string

// Certificate status constants.
const (
	StatusActive    CertificateStatus = "active"
	StatusRevoked   CertificateStatus = "revoked"
	StatusExpired   CertificateStatus = "expired"
	StatusSuspended CertificateStatus = "suspended"
)

// RevocationReason represents the reason for revocation.
type RevocationReason int

// Revocation reason constants (RFC 5280).
const (
	ReasonUnspecified          RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonUnspecified
	ReasonKeyCompromise        RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonKeyCompromise
	ReasonCACompromise         RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonCACompromise
	ReasonAffiliationChanged   RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonAffiliationChanged
	ReasonSuperseded           RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonSuperseded
	ReasonCessationOfOperation RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonCessationOfOperation
	ReasonCertificateHold      RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonCertificateHold
	ReasonRemoveFromCRL        RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonRemoveFromCRL
	ReasonPrivilegeWithdrawn   RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonPrivilegeWithdrawn
	ReasonAACompromise         RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonAACompromise
)

// StoredCertificate represents a certificate in storage.
type StoredCertificate struct {
	ID                googleUuid.UUID   `json:"id" gorm:"type:text;primaryKey"`
	TenantID          googleUuid.UUID   `json:"tenant_id" gorm:"type:text;not null;index:idx_pki_certificates_tenant_status,priority:1;uniqueIndex:idx_pki_certificates_tenant_serial,priority:1;uniqueIndex:idx_pki_certificates_tenant_request_unique,priority:1,where:request_id <> '00000000-0000-0000-0000-000000000000'"`
	RequestID         googleUuid.UUID   `json:"request_id" gorm:"type:text;not null;uniqueIndex:idx_pki_certificates_tenant_request_unique,priority:2,where:request_id <> '00000000-0000-0000-0000-000000000000'"`
	SerialNumber      string            `json:"serial_number" gorm:"not null;uniqueIndex:idx_pki_certificates_tenant_serial,priority:2"`
	SubjectDN         string            `json:"subject_dn" gorm:"not null"`
	IssuerDN          string            `json:"issuer_dn" gorm:"not null"`
	IssuerAuthorityID string            `json:"issuer_authority_id" gorm:"not null;default:''"`
	NotBefore         time.Time         `json:"not_before" gorm:"not null"`
	NotAfter          time.Time         `json:"not_after" gorm:"not null"`
	Status            CertificateStatus `json:"status" gorm:"not null;index:idx_pki_certificates_tenant_status,priority:2"`
	ProfileID         string            `json:"profile_id" gorm:"not null"`
	RequesterID       string            `json:"requester_id" gorm:"not null"`
	CertificatePEM    string            `json:"certificate_pem" gorm:"not null"`
	CertificateDER    []byte            `json:"certificate_der"`
	PublicKeyHash     string            `json:"public_key_hash" gorm:"not null"`
	SubjectKeyID      string            `json:"subject_key_id" gorm:"not null"`
	AuthorityKeyID    string            `json:"authority_key_id" gorm:"not null"`
	RevocationTime    *time.Time        `json:"revocation_time,omitempty"`
	RevocationReason  *RevocationReason `json:"revocation_reason,omitempty"`
	Metadata          map[string]string `json:"metadata" gorm:"serializer:json;not null"`
	CreatedAt         time.Time         `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt         time.Time         `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

func (StoredCertificate) TableName() string {
	return cryptoutilSharedMagic.PKICertificatesTable
}

type StoredAuthority struct {
	ID                  string    `gorm:"type:text;primaryKey"`
	CertificateDER      []byte    `gorm:"not null"`
	PrivateKeyEncrypted []byte    `gorm:"not null"`
	CreatedAt           time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt           time.Time `gorm:"not null;autoUpdateTime"`
}

// StoredCRLSequence stores the last allocated issuer-scoped CRL number.
type StoredCRLSequence struct {
	IssuerKeyID string    `gorm:"type:text;primaryKey"`
	Number      int64     `gorm:"type:bigint;not null"`
	CreatedAt   time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"not null;autoUpdateTime"`
}

func (StoredCRLSequence) TableName() string {
	return cryptoutilSharedMagic.PKICRLSequencesTable
}

func (StoredAuthority) TableName() string {
	return cryptoutilSharedMagic.PKIAuthoritiesTable
}

type AuthorityStore interface {
	Get(ctx context.Context, id string) (*StoredAuthority, error)
	GetAny(ctx context.Context) (*StoredAuthority, error)
	List(ctx context.Context) ([]*StoredAuthority, error)
	Create(ctx context.Context, authority *StoredAuthority) error
}

// ListFilter defines filtering options for listing certificates.
type ListFilter struct {
	Status       *CertificateStatus
	ProfileID    *string
	RequesterID  *string
	IssuerDN     *string
	NotAfterFrom *time.Time
	NotAfterTo   *time.Time
	Limit        int
	Offset       int
}

// Store defines the certificate storage interface.
type Store interface {
	// Store stores a new certificate.
	Store(ctx context.Context, cert *StoredCertificate) error

	// Get retrieves a certificate by ID.
	Get(ctx context.Context, id googleUuid.UUID) (*StoredCertificate, error)

	// GetBySerialNumber retrieves a certificate by serial number.
	GetBySerialNumber(ctx context.Context, serialNumber string) (*StoredCertificate, error)

	// GetBySerialNumberForIssuer retrieves issuer-scoped status without tenant context.
	GetBySerialNumberForIssuer(ctx context.Context, serialNumber, issuerKeyID string) (*StoredCertificate, error)

	// GetByRequestID retrieves a certificate by enrollment request ID.
	GetByRequestID(ctx context.Context, requestID googleUuid.UUID) (*StoredCertificate, error)

	// List returns certificates matching the filter.
	List(ctx context.Context, filter *ListFilter) ([]*StoredCertificate, int, error)

	// Update updates an existing certificate.
	Update(ctx context.Context, cert *StoredCertificate) error

	// Delete removes a certificate from storage.
	Delete(ctx context.Context, id googleUuid.UUID) error

	// Revoke marks a certificate as revoked.
	Revoke(ctx context.Context, id googleUuid.UUID, reason RevocationReason) error

	// GetRevoked returns all revoked certificates for CRL generation.
	GetRevoked(ctx context.Context, issuerDN string) ([]*StoredCertificate, error)

	// GetRevokedByIssuerKeyID returns all tenant revocations for a single certificate issuer.
	GetRevokedByIssuerKeyID(ctx context.Context, issuerKeyID string) ([]*StoredCertificate, error)

	// AllocateCRLNumber atomically and durably allocates the next issuer-scoped CRL number.
	AllocateCRLNumber(ctx context.Context, issuerKeyID string) (*big.Int, error)

	// CountByStatus returns certificate counts by status.
	CountByStatus(ctx context.Context) (map[CertificateStatus]int64, error)

	// Close closes the storage connection.
	Close() error
}

// MemoryStore implements an in-memory certificate store.
type MemoryStore struct {
	certificates map[googleUuid.UUID]*StoredCertificate
	bySerial     map[string]googleUuid.UUID
	byRequestID  map[googleUuid.UUID]googleUuid.UUID
	crlSequences map[string]*big.Int
	mu           sync.RWMutex
}

// NewMemoryStore creates a new in-memory certificate store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		certificates: make(map[googleUuid.UUID]*StoredCertificate),
		bySerial:     make(map[string]googleUuid.UUID),
		byRequestID:  make(map[googleUuid.UUID]googleUuid.UUID),
		crlSequences: make(map[string]*big.Int),
	}
}

// Store stores a new certificate.
func (s *MemoryStore) Store(_ context.Context, cert *StoredCertificate) error {
	if cert == nil {
		return ErrInvalidCertificate
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.certificates[cert.ID]; exists {
		return ErrCertificateExists
	}

	if _, exists := s.bySerial[cert.SerialNumber]; exists {
		return fmt.Errorf("serial number already exists: %s", cert.SerialNumber)
	}

	if cert.RequestID != googleUuid.Nil {
		if _, exists := s.byRequestID[cert.RequestID]; exists {
			return ErrCertificateExists
		}
	}

	now := time.Now().UTC()
	cert.CreatedAt = now
	cert.UpdatedAt = now

	s.certificates[cert.ID] = cert

	s.bySerial[cert.SerialNumber] = cert.ID
	if cert.RequestID != googleUuid.Nil {
		s.byRequestID[cert.RequestID] = cert.ID
	}

	return nil
}

// Get retrieves a certificate by ID.
func (s *MemoryStore) Get(_ context.Context, id googleUuid.UUID) (*StoredCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cert, exists := s.certificates[id]
	if !exists {
		return nil, ErrCertificateNotFound
	}

	return cert, nil
}

// GetBySerialNumber retrieves a certificate by serial number.
func (s *MemoryStore) GetBySerialNumber(_ context.Context, serialNumber string) (*StoredCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, exists := s.bySerial[serialNumber]
	if !exists {
		return nil, ErrCertificateNotFound
	}

	cert, exists := s.certificates[id]
	if !exists {
		return nil, ErrCertificateNotFound
	}

	return cert, nil
}

func (s *MemoryStore) GetBySerialNumberForIssuer(_ context.Context, serialNumber, issuerKeyID string) (*StoredCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	issuerKeyID = strings.ToLower(strings.TrimSpace(issuerKeyID))
	if issuerKeyID == "" {
		return nil, ErrInvalidIssuerKeyID
	}

	var found *StoredCertificate

	for _, certificate := range s.certificates {
		if !strings.EqualFold(certificate.SerialNumber, serialNumber) ||
			!strings.EqualFold(certificate.AuthorityKeyID, issuerKeyID) {
			continue
		}

		if found != nil {
			return nil, ErrCertificateAmbiguous
		}

		found = certificate
	}

	if found == nil {
		return nil, ErrCertificateNotFound
	}

	return found, nil
}

func (s *MemoryStore) GetByRequestID(_ context.Context, requestID googleUuid.UUID) (*StoredCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, exists := s.byRequestID[requestID]
	if !exists {
		return nil, ErrCertificateNotFound
	}

	cert, exists := s.certificates[id]
	if !exists {
		return nil, ErrCertificateNotFound
	}

	return cert, nil
}

// List returns certificates matching the filter.
func (s *MemoryStore) List(_ context.Context, filter *ListFilter) ([]*StoredCertificate, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*StoredCertificate

	for _, cert := range s.certificates {
		if s.matchesFilter(cert, filter) {
			results = append(results, cert)
		}
	}

	total := len(results)

	// Apply pagination.
	if filter != nil && filter.Offset > 0 {
		if filter.Offset >= len(results) {
			return []*StoredCertificate{}, total, nil
		}

		results = results[filter.Offset:]
	}

	if filter != nil && filter.Limit > 0 && filter.Limit < len(results) {
		results = results[:filter.Limit]
	}

	return results, total, nil
}

// matchesFilter checks if a certificate matches the filter criteria.
func (s *MemoryStore) matchesFilter(cert *StoredCertificate, filter *ListFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Status != nil && cert.Status != *filter.Status {
		return false
	}

	if filter.ProfileID != nil && cert.ProfileID != *filter.ProfileID {
		return false
	}

	if filter.RequesterID != nil && cert.RequesterID != *filter.RequesterID {
		return false
	}

	if filter.IssuerDN != nil && cert.IssuerDN != *filter.IssuerDN {
		return false
	}

	if filter.NotAfterFrom != nil && cert.NotAfter.Before(*filter.NotAfterFrom) {
		return false
	}

	if filter.NotAfterTo != nil && cert.NotAfter.After(*filter.NotAfterTo) {
		return false
	}

	return true
}

// Update updates an existing certificate.
func (s *MemoryStore) Update(_ context.Context, cert *StoredCertificate) error {
	if cert == nil {
		return ErrInvalidCertificate
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.certificates[cert.ID]
	if !exists {
		return ErrCertificateNotFound
	}

	if existing.Status == StatusRevoked {
		cert.Status = existing.Status
		cert.RevocationTime = existing.RevocationTime
		cert.RevocationReason = existing.RevocationReason
	}

	cert.UpdatedAt = time.Now().UTC()
	s.certificates[cert.ID] = cert

	return nil
}

// Delete removes a certificate from storage.
func (s *MemoryStore) Delete(_ context.Context, id googleUuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cert, exists := s.certificates[id]
	if !exists {
		return ErrCertificateNotFound
	}

	delete(s.bySerial, cert.SerialNumber)

	if cert.RequestID != googleUuid.Nil {
		delete(s.byRequestID, cert.RequestID)
	}

	delete(s.certificates, id)

	return nil
}

// Revoke marks a certificate as revoked.
func (s *MemoryStore) Revoke(_ context.Context, id googleUuid.UUID, reason RevocationReason) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cert, exists := s.certificates[id]
	if !exists {
		return ErrCertificateNotFound
	}

	if cert.Status == StatusRevoked {
		return ErrCertificateRevoked
	}

	now := time.Now().UTC()
	cert.Status = StatusRevoked
	cert.RevocationTime = &now
	cert.RevocationReason = &reason
	cert.UpdatedAt = now

	return nil
}

// CountByStatus returns certificate counts by status.
func (s *MemoryStore) CountByStatus(_ context.Context) (map[CertificateStatus]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := make(map[CertificateStatus]int64)

	for _, cert := range s.certificates {
		counts[cert.Status]++
	}

	return counts, nil
}

// Close closes the storage connection.
func (s *MemoryStore) Close() error {
	return nil
}

// NewStoredCertificateFromX509 creates a StoredCertificate from an x509.Certificate.
func NewStoredCertificateFromX509(cert *x509.Certificate, profileID, requesterID string) (*StoredCertificate, error) {
	if cert == nil {
		return nil, ErrInvalidCertificate
	}

	id, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	return &StoredCertificate{
		ID:             id,
		SerialNumber:   cert.SerialNumber.Text(cryptoutilSharedMagic.HexBase),
		SubjectDN:      cert.Subject.String(),
		IssuerDN:       cert.Issuer.String(),
		NotBefore:      cert.NotBefore,
		NotAfter:       cert.NotAfter,
		Status:         StatusActive,
		ProfileID:      profileID,
		RequesterID:    requesterID,
		CertificatePEM: string(pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: cert.Raw})),
		CertificateDER: cert.Raw,
		SubjectKeyID:   fmt.Sprintf("%x", cert.SubjectKeyId),
		AuthorityKeyID: fmt.Sprintf("%x", cert.AuthorityKeyId),
		Metadata:       make(map[string]string),
	}, nil
}
