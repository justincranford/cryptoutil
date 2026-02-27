// Copyright (c) 2025 Justin Cranford

// Package storage provides certificate storage abstractions and implementations.
package storage

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"
	"time"

	googleUuid "github.com/google/uuid"
)

// Common errors.
var (
	ErrCertificateNotFound = errors.New("certificate not found")
	ErrCertificateExists   = errors.New("certificate already exists")
	ErrInvalidCertificate  = errors.New("invalid certificate")
	ErrStorageFull         = errors.New("storage capacity exceeded")
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
	ReasonUnspecified          RevocationReason = 0
	ReasonKeyCompromise        RevocationReason = 1
	ReasonCACompromise         RevocationReason = 2
	ReasonAffiliationChanged   RevocationReason = 3
	ReasonSuperseded           RevocationReason = 4
	ReasonCessationOfOperation RevocationReason = 5
	ReasonCertificateHold      RevocationReason = 6
	ReasonRemoveFromCRL        RevocationReason = 8
	ReasonPrivilegeWithdrawn   RevocationReason = 9
	ReasonAACompromise         RevocationReason = 10
)

// StoredCertificate represents a certificate in storage.
type StoredCertificate struct {
	ID               googleUuid.UUID   `json:"id"`
	SerialNumber     string            `json:"serial_number"`
	SubjectDN        string            `json:"subject_dn"`
	IssuerDN         string            `json:"issuer_dn"`
	NotBefore        time.Time         `json:"not_before"`
	NotAfter         time.Time         `json:"not_after"`
	Status           CertificateStatus `json:"status"`
	ProfileID        string            `json:"profile_id"`
	RequesterID      string            `json:"requester_id"`
	CertificatePEM   string            `json:"certificate_pem"`
	CertificateDER   []byte            `json:"certificate_der"`
	PublicKeyHash    string            `json:"public_key_hash"`
	SubjectKeyID     string            `json:"subject_key_id"`
	AuthorityKeyID   string            `json:"authority_key_id"`
	RevocationTime   *time.Time        `json:"revocation_time,omitempty"`
	RevocationReason *RevocationReason `json:"revocation_reason,omitempty"`
	Metadata         map[string]string `json:"metadata"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
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

	// CountByStatus returns certificate counts by status.
	CountByStatus(ctx context.Context) (map[CertificateStatus]int64, error)

	// Close closes the storage connection.
	Close() error
}

// MemoryStore implements an in-memory certificate store.
type MemoryStore struct {
	certificates map[googleUuid.UUID]*StoredCertificate
	bySerial     map[string]googleUuid.UUID
	mu           sync.RWMutex
}

// NewMemoryStore creates a new in-memory certificate store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		certificates: make(map[googleUuid.UUID]*StoredCertificate),
		bySerial:     make(map[string]googleUuid.UUID),
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

	now := time.Now().UTC()
	cert.CreatedAt = now
	cert.UpdatedAt = now

	s.certificates[cert.ID] = cert
	s.bySerial[cert.SerialNumber] = cert.ID

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

	if _, exists := s.certificates[cert.ID]; !exists {
		return ErrCertificateNotFound
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
		return errors.New("certificate already revoked")
	}

	now := time.Now().UTC()
	cert.Status = StatusRevoked
	cert.RevocationTime = &now
	cert.RevocationReason = &reason
	cert.UpdatedAt = now

	return nil
}

// GetRevoked returns all revoked certificates for CRL generation.
func (s *MemoryStore) GetRevoked(_ context.Context, issuerDN string) ([]*StoredCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*StoredCertificate

	for _, cert := range s.certificates {
		if cert.Status == StatusRevoked && (issuerDN == "" || cert.IssuerDN == issuerDN) {
			results = append(results, cert)
		}
	}

	return results, nil
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
		SerialNumber:   cert.SerialNumber.String(),
		SubjectDN:      cert.Subject.String(),
		IssuerDN:       cert.Issuer.String(),
		NotBefore:      cert.NotBefore,
		NotAfter:       cert.NotAfter,
		Status:         StatusActive,
		ProfileID:      profileID,
		RequesterID:    requesterID,
		CertificateDER: cert.Raw,
		SubjectKeyID:   fmt.Sprintf("%x", cert.SubjectKeyId),
		AuthorityKeyID: fmt.Sprintf("%x", cert.AuthorityKeyId),
		Metadata:       make(map[string]string),
	}, nil
}
