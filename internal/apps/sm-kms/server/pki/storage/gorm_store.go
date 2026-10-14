// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilSharedDatabase "cryptoutil/internal/shared/database"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type GORMStore struct {
	db *gorm.DB
}

func NewGORMStore(db *gorm.DB) (*GORMStore, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GORMStore{db: db}, nil
}

func (s *GORMStore) Store(ctx context.Context, cert *StoredCertificate) error {
	if cert == nil {
		return ErrInvalidCertificate
	}

	tenantID, err := requireTenantID(ctx)
	if err != nil {
		return err
	}

	if cert.ID == googleUuid.Nil {
		cert.ID, err = googleUuid.NewV7()
		if err != nil {
			return fmt.Errorf("failed to generate certificate ID: %w", err)
		}
	}

	var existing StoredCertificate

	err = s.db.WithContext(ctx).Where("tenant_id = ? AND serial_number = ?", tenantID, cert.SerialNumber).Take(&existing).Error
	switch {
	case err == nil:
		return ErrCertificateExists
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return fmt.Errorf("failed to check certificate serial number: %w", err)
	}

	cert.TenantID = tenantID
	if err := s.db.WithContext(ctx).Create(cert).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrCertificateExists
		}

		return fmt.Errorf("failed to store certificate: %w", err)
	}

	return nil
}

func (s *GORMStore) Get(ctx context.Context, id googleUuid.UUID) (*StoredCertificate, error) {
	tenantID, err := requireTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var cert StoredCertificate

	err = s.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&cert).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCertificateNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return &cert, nil
}

func (s *GORMStore) GetBySerialNumber(ctx context.Context, serialNumber string) (*StoredCertificate, error) {
	tenantID, err := requireTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var cert StoredCertificate

	err = s.db.WithContext(ctx).Where("tenant_id = ? AND serial_number = ?", tenantID, serialNumber).First(&cert).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCertificateNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get certificate by serial number: %w", err)
	}

	return &cert, nil
}

func (s *GORMStore) GetByRequestID(ctx context.Context, requestID googleUuid.UUID) (*StoredCertificate, error) {
	if requestID == googleUuid.Nil {
		return nil, ErrCertificateNotFound
	}

	tenantID, err := requireTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var cert StoredCertificate

	err = s.db.WithContext(ctx).Where("tenant_id = ? AND request_id = ?", tenantID, requestID).First(&cert).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCertificateNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get certificate by request ID: %w", err)
	}

	return &cert, nil
}

func (s *GORMStore) List(ctx context.Context, filter *ListFilter) ([]*StoredCertificate, int, error) {
	tenantID, err := requireTenantID(ctx)
	if err != nil {
		return nil, 0, err
	}

	query := applyGORMListFilter(s.db.WithContext(ctx).Model(&StoredCertificate{}).Where("tenant_id = ?", tenantID), filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count certificates: %w", err)
	}

	if filter != nil && filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if filter != nil && filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	var certificates []*StoredCertificate
	if err := query.Order("created_at DESC").Find(&certificates).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list certificates: %w", err)
	}

	return certificates, int(total), nil
}

func applyGORMListFilter(query *gorm.DB, filter *ListFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.ProfileID != nil {
		query = query.Where("profile_id = ?", *filter.ProfileID)
	}

	if filter.RequesterID != nil {
		query = query.Where("requester_id = ?", *filter.RequesterID)
	}

	if filter.IssuerDN != nil {
		query = query.Where("issuer_dn = ?", *filter.IssuerDN)
	}

	if filter.NotAfterFrom != nil {
		query = query.Where("not_after >= ?", *filter.NotAfterFrom)
	}

	if filter.NotAfterTo != nil {
		query = query.Where("not_after <= ?", *filter.NotAfterTo)
	}

	return query
}

func (s *GORMStore) Update(ctx context.Context, cert *StoredCertificate) error {
	if cert == nil {
		return ErrInvalidCertificate
	}

	tenantID, err := requireTenantID(ctx)
	if err != nil {
		return err
	}

	cert.TenantID = tenantID
	cert.UpdatedAt = time.Now().UTC()

	result := s.db.WithContext(ctx).Model(&StoredCertificate{}).
		Where("tenant_id = ? AND id = ?", tenantID, cert.ID).
		Select("*").Updates(cert)
	if result.Error != nil {
		return fmt.Errorf("failed to update certificate: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrCertificateNotFound
	}

	return nil
}

func (s *GORMStore) Delete(ctx context.Context, id googleUuid.UUID) error {
	tenantID, err := requireTenantID(ctx)
	if err != nil {
		return err
	}

	result := s.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&StoredCertificate{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete certificate: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrCertificateNotFound
	}

	return nil
}

func (s *GORMStore) Revoke(ctx context.Context, id googleUuid.UUID, reason RevocationReason) error {
	tenantID, err := requireTenantID(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	result := s.db.WithContext(ctx).Model(&StoredCertificate{}).
		Where("tenant_id = ? AND id = ? AND status <> ?", tenantID, id, StatusRevoked).
		Updates(map[string]any{
			cryptoutilSharedMagic.StringStatus:   StatusRevoked,
			"revocation_time":                    now,
			"revocation_reason":                  reason,
			cryptoutilSharedMagic.ClaimUpdatedAt: now,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to revoke certificate: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		return nil
	}

	cert, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	if cert.Status == StatusRevoked {
		return ErrCertificateRevoked
	}

	return fmt.Errorf("failed to revoke certificate %s", id)
}

func (s *GORMStore) GetRevoked(ctx context.Context, issuerDN string) ([]*StoredCertificate, error) {
	tenantID, err := requireTenantID(ctx)
	if err != nil {
		return nil, err
	}

	query := s.db.WithContext(ctx).Where("tenant_id = ? AND status = ?", tenantID, StatusRevoked)
	if issuerDN != "" {
		query = query.Where("issuer_dn = ?", issuerDN)
	}

	var certificates []*StoredCertificate
	if err := query.Order("revocation_time ASC").Find(&certificates).Error; err != nil {
		return nil, fmt.Errorf("failed to get revoked certificates: %w", err)
	}

	return certificates, nil
}

func (s *GORMStore) CountByStatus(ctx context.Context) (map[CertificateStatus]int64, error) {
	tenantID, err := requireTenantID(ctx)
	if err != nil {
		return nil, err
	}

	type statusCount struct {
		Status CertificateStatus
		Count  int64
	}

	var countsByStatus []statusCount

	err = s.db.WithContext(ctx).Model(&StoredCertificate{}).
		Select(cryptoutilSharedMagic.StringStatus+", count(*) AS count").
		Where("tenant_id = ?", tenantID).
		Group(cryptoutilSharedMagic.StringStatus).
		Scan(&countsByStatus).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count certificates by status: %w", err)
	}

	counts := make(map[CertificateStatus]int64, len(countsByStatus))
	for _, count := range countsByStatus {
		counts[count.Status] = count.Count
	}

	return counts, nil
}

func (s *GORMStore) Close() error {
	return nil
}

func requireTenantID(ctx context.Context) (googleUuid.UUID, error) {
	if ctx == nil {
		return googleUuid.Nil, fmt.Errorf("context is required")
	}

	tenantContext, err := cryptoutilSharedDatabase.RequireTenantContext(ctx)
	if err != nil {
		return googleUuid.Nil, fmt.Errorf("tenant context is required: %w", err)
	}

	return tenantContext.TenantID, nil
}
