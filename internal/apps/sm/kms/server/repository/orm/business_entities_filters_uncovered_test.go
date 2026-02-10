//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestApplyGetElasticKeysFilters_NilFilters tests nil filter handling.
func TestApplyGetElasticKeysFilters_NilFilters(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&ElasticKey{})
	filteredQuery := applyGetElasticKeysFilters(query, nil)
	require.NotNil(t, filteredQuery)
	require.Equal(t, query, filteredQuery, "Nil filters should return unmodified query")
}

// TestApplyKeyFilters_NilFilters tests nil filter handling.
func TestApplyKeyFilters_NilFilters(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&MaterialKey{})
	filteredQuery := applyKeyFilters(query, nil)
	require.NotNil(t, filteredQuery)
	require.Equal(t, query, filteredQuery, "Nil filters should return unmodified query")
}

// TestApplyGetElasticKeyKeysFilters_NilFilters tests nil filter handling.
func TestApplyGetElasticKeyKeysFilters_NilFilters(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&MaterialKey{})
	filteredQuery := applyGetElasticKeyKeysFilters(query, nil)
	require.NotNil(t, filteredQuery)
	require.Equal(t, query, filteredQuery, "Nil filters should return unmodified query")
}

// TestApplyKeyFilters_WithDates tests date filtering for material keys.
func TestApplyKeyFilters_WithDates(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&MaterialKey{})

	// Test with MinimumGenerateDate.
	minDate := time.Now().UTC().Add(-24 * time.Hour)
	filters := &GetMaterialKeysFilters{
		MinimumGenerateDate: &minDate,
	}
	filteredQuery := applyKeyFilters(query, filters)
	require.NotNil(t, filteredQuery)

	// Test with MaximumGenerateDate.
	maxDate := time.Now().UTC()
	filters2 := &GetMaterialKeysFilters{
		MaximumGenerateDate: &maxDate,
	}
	filteredQuery2 := applyKeyFilters(query, filters2)
	require.NotNil(t, filteredQuery2)

	// Test with both dates.
	filters3 := &GetMaterialKeysFilters{
		MinimumGenerateDate: &minDate,
		MaximumGenerateDate: &maxDate,
	}
	filteredQuery3 := applyKeyFilters(query, filters3)
	require.NotNil(t, filteredQuery3)
}

// TestApplyGetElasticKeyKeysFilters_WithDates tests date filtering for elastic key material keys.
func TestApplyGetElasticKeyKeysFilters_WithDates(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&MaterialKey{})

	// Test with MinimumGenerateDate.
	minDate := time.Now().UTC().Add(-24 * time.Hour)
	filters := &GetElasticKeyMaterialKeysFilters{
		MinimumGenerateDate: &minDate,
	}
	filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
	require.NotNil(t, filteredQuery)

	// Test with MaximumGenerateDate.
	maxDate := time.Now().UTC()
	filters2 := &GetElasticKeyMaterialKeysFilters{
		MaximumGenerateDate: &maxDate,
	}
	filteredQuery2 := applyGetElasticKeyKeysFilters(query, filters2)
	require.NotNil(t, filteredQuery2)

	// Test with both dates.
	filters3 := &GetElasticKeyMaterialKeysFilters{
		MinimumGenerateDate: &minDate,
		MaximumGenerateDate: &maxDate,
	}
	filteredQuery3 := applyGetElasticKeyKeysFilters(query, filters3)
	require.NotNil(t, filteredQuery3)
}

// TestApplyKeyFilters_WithSort tests sorting for material keys.
func TestApplyKeyFilters_WithSort(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&MaterialKey{})

	// Test with single sort field.
	filters := &GetMaterialKeysFilters{
		Sort: []string{"material_key_id DESC"},
	}
	filteredQuery := applyKeyFilters(query, filters)
	require.NotNil(t, filteredQuery)

	// Test with multiple sort fields.
	filters2 := &GetMaterialKeysFilters{
		Sort: []string{"elastic_key_id ASC", "material_key_id DESC"},
	}
	filteredQuery2 := applyKeyFilters(query, filters2)
	require.NotNil(t, filteredQuery2)
}

// TestApplyGetElasticKeyKeysFilters_WithSort tests sorting for elastic key material keys.
func TestApplyGetElasticKeyKeysFilters_WithSort(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&MaterialKey{})

	// Test with single sort field.
	filters := &GetElasticKeyMaterialKeysFilters{
		Sort: []string{"material_key_id DESC"},
	}
	filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
	require.NotNil(t, filteredQuery)

	// Test with multiple sort fields.
	filters2 := &GetElasticKeyMaterialKeysFilters{
		Sort: []string{"elastic_key_id ASC", "material_key_id DESC"},
	}
	filteredQuery2 := applyGetElasticKeyKeysFilters(query, filters2)
	require.NotNil(t, filteredQuery2)
}

// TestApplyGetElasticKeysFilters_WithSort tests sorting for elastic keys.
func TestApplyGetElasticKeysFilters_WithSort(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&ElasticKey{})

	// Test with single sort field.
	filters := &GetElasticKeysFilters{
		Sort: []string{"elastic_key_name DESC"},
	}
	filteredQuery := applyGetElasticKeysFilters(query, filters)
	require.NotNil(t, filteredQuery)

	// Test with multiple sort fields.
	filters2 := &GetElasticKeysFilters{
		Sort: []string{"elastic_key_name ASC", "elastic_key_id DESC"},
	}
	filteredQuery2 := applyGetElasticKeysFilters(query, filters2)
	require.NotNil(t, filteredQuery2)
}

// TestApplyKeyFilters_WithPagination tests pagination for material keys.
func TestApplyKeyFilters_WithPagination(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&MaterialKey{})

	// Test with pagination (PageSize > 0).
	filters := &GetMaterialKeysFilters{
		PageNumber: 0,
		PageSize:   10,
	}
	filteredQuery := applyKeyFilters(query, filters)
	require.NotNil(t, filteredQuery)

	// Test with PageSize=0 (no pagination applied).
	filters2 := &GetMaterialKeysFilters{
		PageNumber: 0,
		PageSize:   0,
	}
	filteredQuery2 := applyKeyFilters(query, filters2)
	require.NotNil(t, filteredQuery2)
}

// TestApplyGetElasticKeyKeysFilters_WithPagination tests pagination for elastic key material keys.
func TestApplyGetElasticKeyKeysFilters_WithPagination(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&MaterialKey{})

	// Test with pagination (PageSize > 0).
	filters := &GetElasticKeyMaterialKeysFilters{
		PageNumber: 0,
		PageSize:   10,
	}
	filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
	require.NotNil(t, filteredQuery)

	// Test with PageSize=0 (no pagination applied).
	filters2 := &GetElasticKeyMaterialKeysFilters{
		PageNumber: 0,
		PageSize:   0,
	}
	filteredQuery2 := applyGetElasticKeyKeysFilters(query, filters2)
	require.NotNil(t, filteredQuery2)
}

// TestApplyKeyFilters_WithMaterialKeyIDs tests filtering by material key IDs.
func TestApplyKeyFilters_WithMaterialKeyIDs(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&MaterialKey{})

	// Test with single material key ID.
	materialKeyID := googleUuid.New()
	filters := &GetMaterialKeysFilters{
		MaterialKeyID: []googleUuid.UUID{materialKeyID},
	}
	filteredQuery := applyKeyFilters(query, filters)
	require.NotNil(t, filteredQuery)

	// Test with multiple material key IDs.
	filters2 := &GetMaterialKeysFilters{
		MaterialKeyID: []googleUuid.UUID{googleUuid.New(), googleUuid.New(), googleUuid.New()},
	}
	filteredQuery2 := applyKeyFilters(query, filters2)
	require.NotNil(t, filteredQuery2)
}

// TestApplyKeyFilters_WithElasticKeyIDs tests filtering by elastic key IDs.
func TestApplyKeyFilters_WithElasticKeyIDs(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&MaterialKey{})

	// Test with single elastic key ID.
	elasticKeyID := googleUuid.New()
	filters := &GetMaterialKeysFilters{
		ElasticKeyID: []googleUuid.UUID{elasticKeyID},
	}
	filteredQuery := applyKeyFilters(query, filters)
	require.NotNil(t, filteredQuery)

	// Test with multiple elastic key IDs.
	filters2 := &GetMaterialKeysFilters{
		ElasticKeyID: []googleUuid.UUID{googleUuid.New(), googleUuid.New(), googleUuid.New()},
	}
	filteredQuery2 := applyKeyFilters(query, filters2)
	require.NotNil(t, filteredQuery2)
}

// TestApplyGetElasticKeyKeysFilters_WithElasticKeyIDs tests filtering by elastic key IDs.
func TestApplyGetElasticKeyKeysFilters_WithElasticKeyIDs(t *testing.T) {
	t.Parallel()
	query := testOrmRepository.gormDB.Model(&MaterialKey{})

	// Test with single elastic key ID.
	elasticKeyID := googleUuid.New()
	filters := &GetElasticKeyMaterialKeysFilters{
		ElasticKeyID: []googleUuid.UUID{elasticKeyID},
	}
	filteredQuery := applyGetElasticKeyKeysFilters(query, filters)
	require.NotNil(t, filteredQuery)

	// Test with multiple elastic key IDs.
	filters2 := &GetElasticKeyMaterialKeysFilters{
		ElasticKeyID: []googleUuid.UUID{googleUuid.New(), googleUuid.New(), googleUuid.New()},
	}
	filteredQuery2 := applyGetElasticKeyKeysFilters(query, filters2)
	require.NotNil(t, filteredQuery2)
}
