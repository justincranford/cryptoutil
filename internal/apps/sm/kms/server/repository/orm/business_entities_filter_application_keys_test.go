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

func TestApplyKeyFilters(t *testing.T) {
	t.Parallel()
	require.NotNil(t, testOrmRepository)

	t.Run("Filter by single MaterialKeyID", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetMaterialKeysFilters{
			MaterialKeyID: []googleUuid.UUID{googleUuid.New()},
			PageSize:      cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by multiple MaterialKeyIDs", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetMaterialKeysFilters{
			MaterialKeyID: []googleUuid.UUID{googleUuid.New(), googleUuid.New(), googleUuid.New()},
			PageSize:      cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by empty MaterialKeyID slice", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetMaterialKeysFilters{
			MaterialKeyID: []googleUuid.UUID{},
			PageSize:      cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by single ElasticKeyID", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetMaterialKeysFilters{
			ElasticKeyID: []googleUuid.UUID{googleUuid.New()},
			PageSize:     cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by MinimumGenerateDate", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		minDate := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
		filters := &GetMaterialKeysFilters{
			MinimumGenerateDate: timePtr(minDate),
			PageSize:            cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by MaximumGenerateDate", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		maxDate := time.Now().UTC()
		filters := &GetMaterialKeysFilters{
			MaximumGenerateDate: timePtr(maxDate),
			PageSize:            cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by date range", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		minDate := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
		maxDate := time.Now().UTC()
		filters := &GetMaterialKeysFilters{
			MinimumGenerateDate: timePtr(minDate),
			MaximumGenerateDate: timePtr(maxDate),
			PageSize:            cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by combined MaterialKeyID and ElasticKeyID", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetMaterialKeysFilters{
			MaterialKeyID: []googleUuid.UUID{googleUuid.New()},
			ElasticKeyID:  []googleUuid.UUID{googleUuid.New()},
			PageSize:      cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by combined filters and date range", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		minDate := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
		maxDate := time.Now().UTC()
		filters := &GetMaterialKeysFilters{
			MaterialKeyID:       []googleUuid.UUID{googleUuid.New()},
			ElasticKeyID:        []googleUuid.UUID{googleUuid.New()},
			MinimumGenerateDate: timePtr(minDate),
			MaximumGenerateDate: timePtr(maxDate),
			PageSize:            cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by Sort ascending", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetMaterialKeysFilters{
			Sort:     []string{"material_key_generate_date ASC"},
			PageSize: cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by all fields comprehensive", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		minDate := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
		maxDate := time.Now().UTC()
		filters := &GetMaterialKeysFilters{
			MaterialKeyID:       []googleUuid.UUID{googleUuid.New()},
			ElasticKeyID:        []googleUuid.UUID{googleUuid.New()},
			MinimumGenerateDate: timePtr(minDate),
			MaximumGenerateDate: timePtr(maxDate),
			Sort:                []string{"material_key_generate_date DESC"},
			PageSize:            cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("No filters (minimal struct)", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetMaterialKeysFilters{
			PageSize: cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Nil filters", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filteredQuery := applyKeyFilters(query, nil)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Pagination with PageNumber", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetMaterialKeysFilters{
			PageNumber: 3,
			PageSize:   cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyKeyFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})
}

func TestApplyGetElasticKeyKeysFilters(t *testing.T) {
	t.Parallel()
	require.NotNil(t, testOrmRepository)

	t.Run("Filter by single ElasticKeyID", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetElasticKeyMaterialKeysFilters{
			ElasticKeyID: []googleUuid.UUID{googleUuid.New()},
			PageSize:     cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by multiple ElasticKeyIDs", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetElasticKeyMaterialKeysFilters{
			ElasticKeyID: []googleUuid.UUID{googleUuid.New(), googleUuid.New()},
			PageSize:     cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by MinimumGenerateDate", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		minDate := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
		filters := &GetElasticKeyMaterialKeysFilters{
			MinimumGenerateDate: timePtr(minDate),
			PageSize:            cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by MaximumGenerateDate", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		maxDate := time.Now().UTC()
		filters := &GetElasticKeyMaterialKeysFilters{
			MaximumGenerateDate: timePtr(maxDate),
			PageSize:            cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by date range", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		minDate := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
		maxDate := time.Now().UTC()
		filters := &GetElasticKeyMaterialKeysFilters{
			MinimumGenerateDate: timePtr(minDate),
			MaximumGenerateDate: timePtr(maxDate),
			PageSize:            cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by combined ElasticKeyID and date range", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		minDate := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
		maxDate := time.Now().UTC()
		filters := &GetElasticKeyMaterialKeysFilters{
			ElasticKeyID:        []googleUuid.UUID{googleUuid.New()},
			MinimumGenerateDate: timePtr(minDate),
			MaximumGenerateDate: timePtr(maxDate),
			PageSize:            cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by Sort ascending", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetElasticKeyMaterialKeysFilters{
			Sort:     []string{"material_key_generate_date ASC"},
			PageSize: cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Filter by all fields comprehensive", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		minDate := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
		maxDate := time.Now().UTC()
		filters := &GetElasticKeyMaterialKeysFilters{
			ElasticKeyID:        []googleUuid.UUID{googleUuid.New()},
			MinimumGenerateDate: timePtr(minDate),
			MaximumGenerateDate: timePtr(maxDate),
			Sort:                []string{"material_key_generate_date DESC"},
			PageSize:            cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("No filters (minimal struct)", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetElasticKeyMaterialKeysFilters{
			PageSize: cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Nil filters", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filteredQuery := applyGetElasticKeyKeysFilters(query, nil)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Empty ElasticKeyID slice", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetElasticKeyMaterialKeysFilters{
			ElasticKeyID: []googleUuid.UUID{},
			PageSize:     cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})

	t.Run("Pagination with PageNumber", func(t *testing.T) {
		query := testOrmRepository.gormDB.Model(&MaterialKey{})
		filters := &GetElasticKeyMaterialKeysFilters{
			PageNumber: 1,
			PageSize:   cryptoutilSharedMagic.DefaultPageSize,
		}
		filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
		require.NotNil(t, filteredQuery)
		require.IsType(t, &gorm.DB{}, filteredQuery)
	})
}
