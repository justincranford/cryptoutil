// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"context"

	googleUuid "github.com/google/uuid"
)

// CARepository defines the persistence interface for Certificate Authorities.
type CARepository interface {
	// CreateCA creates a new Certificate Authority.
	CreateCA(ctx context.Context, ca *CertificateAuthority) error
	// GetCAByID retrieves a CA by its ID.
	GetCAByID(ctx context.Context, id googleUuid.UUID) (*CertificateAuthority, error)
	// GetCAByName retrieves a CA by its unique name.
	GetCAByName(ctx context.Context, name string) (*CertificateAuthority, error)
	// ListCAs retrieves all CAs with optional filtering.
	ListCAs(ctx context.Context, caType *CAType, status *CAStatus) ([]*CertificateAuthority, error)
	// UpdateCA updates an existing CA.
	UpdateCA(ctx context.Context, ca *CertificateAuthority) error
	// DeleteCA soft-deletes a CA.
	DeleteCA(ctx context.Context, id googleUuid.UUID) error
}

// CertificateRepository defines the persistence interface for issued certificates.
type CertificateRepository interface {
	// CreateCertificate stores a newly issued certificate.
	CreateCertificate(ctx context.Context, cert *Certificate) error
	// GetCertificateByID retrieves a certificate by its ID.
	GetCertificateByID(ctx context.Context, id googleUuid.UUID) (*Certificate, error)
	// GetCertificateBySerial retrieves a certificate by its serial number (hex string).
	GetCertificateBySerial(ctx context.Context, serialHex string) (*Certificate, error)
	// ListCertificates retrieves certificates with optional filtering.
	ListCertificates(ctx context.Context, issuerID *googleUuid.UUID, profileID *googleUuid.UUID, limit, offset int) ([]*Certificate, error)
	// ListExpiringSoon retrieves certificates expiring within the given duration.
	ListExpiringSoon(ctx context.Context, daysUntilExpiry int) ([]*Certificate, error)
	// RevokeCertificate marks a certificate as revoked.
	RevokeCertificate(ctx context.Context, id googleUuid.UUID, reason int) error
	// ListRevokedByIssuer retrieves revoked certificates for CRL generation.
	ListRevokedByIssuer(ctx context.Context, issuerID googleUuid.UUID) ([]*RevocationEntry, error)
}

// ProfileRepository defines the persistence interface for certificate profiles.
type ProfileRepository interface {
	// CreateProfile creates a new certificate profile.
	CreateProfile(ctx context.Context, profile *CertificateProfile) error
	// GetProfileByID retrieves a profile by its ID.
	GetProfileByID(ctx context.Context, id googleUuid.UUID) (*CertificateProfile, error)
	// GetProfileByName retrieves a profile by its unique name.
	GetProfileByName(ctx context.Context, name string) (*CertificateProfile, error)
	// ListProfiles retrieves all active profiles.
	ListProfiles(ctx context.Context, activeOnly bool) ([]*CertificateProfile, error)
	// UpdateProfile updates an existing profile.
	UpdateProfile(ctx context.Context, profile *CertificateProfile) error
	// DeleteProfile soft-deletes a profile.
	DeleteProfile(ctx context.Context, id googleUuid.UUID) error
}

// RequestRepository defines the persistence interface for certificate requests.
type RequestRepository interface {
	// CreateRequest stores a new certificate signing request.
	CreateRequest(ctx context.Context, req *CertificateRequest) error
	// GetRequestByID retrieves a request by its ID.
	GetRequestByID(ctx context.Context, id googleUuid.UUID) (*CertificateRequest, error)
	// ListPendingRequests retrieves pending requests for a given issuer.
	ListPendingRequests(ctx context.Context, issuerID googleUuid.UUID) ([]*CertificateRequest, error)
	// ApproveRequest marks a request as approved.
	ApproveRequest(ctx context.Context, id googleUuid.UUID, approvedBy string) error
	// RejectRequest marks a request as rejected.
	RejectRequest(ctx context.Context, id googleUuid.UUID, rejectedBy, reason string) error
	// MarkIssued marks a request as issued with the certificate ID.
	MarkIssued(ctx context.Context, id, certificateID googleUuid.UUID) error
}
