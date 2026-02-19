//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

//nolint:gochecknoglobals // Test helper function.
func boolPtr(b bool) *bool {
	return &b
}

//nolint:gochecknoglobals // Test helper function.
func timePtr(t time.Time) *time.Time {
	return &t
}

func TestApplyGetElasticKeysFilters(t *testing.T) {
	t.Parallel()
	require.NotNil(t, testOrmRepository)

	t.Run("Filter by single ElasticKeyID", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			ElasticKeyID: []googleUuid.UUID{googleUuid.New()},
			PageSize:     cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by multiple ElasticKeyIDs", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			ElasticKeyID: []googleUuid.UUID{googleUuid.New(), googleUuid.New(), googleUuid.New()},
			PageSize:     cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by empty ElasticKeyID slice", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			ElasticKeyID: []googleUuid.UUID{},
			PageSize:     cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by single Name", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			Name:     []string{"test-key"},
			PageSize: cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by multiple Names", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			Name:     []string{"key1", "key2", "key3"},
			PageSize: cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by single Algorithm", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			Algorithm: []string{"RSA-2048"},
			PageSize:  cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by multiple Algorithms", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			Algorithm: []string{"RSA-2048", "RSA-4096", "ECDSA-P256"},
			PageSize:  cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by VersioningAllowed true", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			VersioningAllowed: boolPtr(true),
			PageSize:          cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by VersioningAllowed false", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			VersioningAllowed: boolPtr(false),
			PageSize:          cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by ImportAllowed true", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			ImportAllowed: boolPtr(true),
			PageSize:      cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by ImportAllowed false", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			ImportAllowed: boolPtr(false),
			PageSize:      cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by ExportAllowed true", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			ExportAllowed: boolPtr(true),
			PageSize:      cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by ExportAllowed false", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			ExportAllowed: boolPtr(false),
			PageSize:      cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by Sort ascending", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			Sort:     []string{"elastic_key_name ASC"},
			PageSize: cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by Sort descending", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			Sort:     []string{"elastic_key_name DESC"},
			PageSize: cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by multiple Sort fields", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			Sort:     []string{"elastic_key_algorithm ASC", "elastic_key_name DESC"},
			PageSize: cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by combined fields", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			ElasticKeyID: []googleUuid.UUID{googleUuid.New()},
			Name:         []string{"test"},
			Algorithm:    []string{"RSA-2048"},
			PageSize:     cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by all filter types together", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			ElasticKeyID:      []googleUuid.UUID{googleUuid.New()},
			Name:              []string{"test"},
			Algorithm:         []string{"RSA-2048"},
			VersioningAllowed: boolPtr(true),
			ImportAllowed:     boolPtr(false),
			ExportAllowed:     boolPtr(true),
			Sort:              []string{"elastic_key_name ASC"},
			PageSize:          cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("No filters (minimal struct)", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			PageSize: cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Nil filters", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filteredQuery := applyGetElasticKeysFilters(query, nil)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Pagination with PageNumber", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			PageNumber: 2,
			PageSize:   cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Custom PageSize", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&ElasticKey{})
		filters := &GetElasticKeysFilters{
			PageSize: 50,
		}
		filteredQuery := applyGetElasticKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})
}
