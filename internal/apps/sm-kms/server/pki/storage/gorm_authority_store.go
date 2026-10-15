// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type GORMAuthorityStore struct {
	db *gorm.DB
}

func NewGORMAuthorityStore(db *gorm.DB) (*GORMAuthorityStore, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}

	return &GORMAuthorityStore{db: db}, nil
}

func (s *GORMAuthorityStore) Get(ctx context.Context, id string) (*StoredAuthority, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}

	var authority StoredAuthority

	err := s.db.WithContext(ctx).Where("id = ?", id).First(&authority).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAuthorityNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get PKI authority: %w", err)
	}

	return &authority, nil
}

func (s *GORMAuthorityStore) GetAny(ctx context.Context) (*StoredAuthority, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}

	var authorities []StoredAuthority
	if err := s.db.WithContext(ctx).Order("id ASC").Limit(2).Find(&authorities).Error; err != nil {
		return nil, fmt.Errorf("failed to list persisted PKI authorities: %w", err)
	}

	switch len(authorities) {
	case 0:
		return nil, ErrAuthorityNotFound
	case 1:
		return &authorities[0], nil
	default:
		return nil, ErrMultipleAuthorities
	}
}

func (s *GORMAuthorityStore) List(ctx context.Context) ([]*StoredAuthority, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}

	var authorities []*StoredAuthority
	if err := s.db.WithContext(ctx).Order("id ASC").Find(&authorities).Error; err != nil {
		return nil, fmt.Errorf("failed to list persisted PKI authorities: %w", err)
	}

	return authorities, nil
}

func (s *GORMAuthorityStore) Create(ctx context.Context, authority *StoredAuthority) error {
	if ctx == nil {
		return fmt.Errorf("context is required")
	}

	if authority == nil {
		return fmt.Errorf("PKI authority is required")
	}

	if err := s.db.WithContext(ctx).Create(authority).Error; err != nil {
		return fmt.Errorf("failed to create PKI authority: %w", err)
	}

	return nil
}

var _ AuthorityStore = (*GORMAuthorityStore)(nil)
