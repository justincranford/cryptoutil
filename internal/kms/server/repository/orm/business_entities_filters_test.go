// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestGetElasticKeysFilters tests the GetElasticKeysFilters struct.
func TestGetElasticKeysFilters(t *testing.T) {
	t.Run("Create filters with all fields", func(t *testing.T) {
		ekID1 := googleUuid.New()
		ekID2 := googleUuid.New()
		versioningAllowed := true
		importAllowed := false
		exportAllowed := true

		filters := GetElasticKeysFilters{
			ElasticKeyID:      []googleUuid.UUID{ekID1, ekID2},
			Name:              []string{"key-1", "key-2"},
			Algorithm:         []string{"A256GCM/A256KW", "A192GCM/A192KW"},
			VersioningAllowed: &versioningAllowed,
			ImportAllowed:     &importAllowed,
			ExportAllowed:     &exportAllowed,
			Sort:              []string{"name", "-created_at"},
			PageNumber:        1,
			PageSize:          50,
		}

		require.Len(t, filters.ElasticKeyID, 2, "Should have 2 elastic key IDs")
		require.Len(t, filters.Name, 2, "Should have 2 names")
		require.Len(t, filters.Algorithm, 2, "Should have 2 algorithms")
		require.NotNil(t, filters.VersioningAllowed, "VersioningAllowed should not be nil")
		require.NotNil(t, filters.ImportAllowed, "ImportAllowed should not be nil")
		require.NotNil(t, filters.ExportAllowed, "ExportAllowed should not be nil")
		require.True(t, *filters.VersioningAllowed, "VersioningAllowed should be true")
		require.False(t, *filters.ImportAllowed, "ImportAllowed should be false")
		require.True(t, *filters.ExportAllowed, "ExportAllowed should be true")
		require.Len(t, filters.Sort, 2, "Should have 2 sort fields")
		require.Equal(t, 1, filters.PageNumber, "PageNumber should be 1")
		require.Equal(t, 50, filters.PageSize, "PageSize should be 50")
	})

	t.Run("Create filters with minimal fields", func(t *testing.T) {
		filters := GetElasticKeysFilters{
			PageNumber: 0,
			PageSize:   10,
		}

		require.Nil(t, filters.ElasticKeyID, "ElasticKeyID should be nil")
		require.Nil(t, filters.Name, "Name should be nil")
		require.Nil(t, filters.Algorithm, "Algorithm should be nil")
		require.Nil(t, filters.VersioningAllowed, "VersioningAllowed should be nil")
		require.Nil(t, filters.ImportAllowed, "ImportAllowed should be nil")
		require.Nil(t, filters.ExportAllowed, "ExportAllowed should be nil")
		require.Nil(t, filters.Sort, "Sort should be nil")
		require.Equal(t, 0, filters.PageNumber, "PageNumber should be 0")
		require.Equal(t, 10, filters.PageSize, "PageSize should be 10")
	})

	t.Run("Create filters with nil boolean pointers", func(t *testing.T) {
		filters := GetElasticKeysFilters{
			Name:              []string{"test-key"},
			VersioningAllowed: nil,
			ImportAllowed:     nil,
			ExportAllowed:     nil,
			PageNumber:        0,
			PageSize:          20,
		}

		require.Len(t, filters.Name, 1, "Should have 1 name")
		require.Nil(t, filters.VersioningAllowed, "VersioningAllowed should be nil")
		require.Nil(t, filters.ImportAllowed, "ImportAllowed should be nil")
		require.Nil(t, filters.ExportAllowed, "ExportAllowed should be nil")
	})
}

// TestGetElasticKeyMaterialKeysFilters tests the GetElasticKeyMaterialKeysFilters struct.
func TestGetElasticKeyMaterialKeysFilters(t *testing.T) {
	t.Run("Create filters with all fields", func(t *testing.T) {
		ekID1 := googleUuid.New()
		ekID2 := googleUuid.New()
		minDate := time.Now().UTC().Add(-24 * time.Hour)
		maxDate := time.Now().UTC()

		filters := GetElasticKeyMaterialKeysFilters{
			ElasticKeyID:        []googleUuid.UUID{ekID1, ekID2},
			MinimumGenerateDate: &minDate,
			MaximumGenerateDate: &maxDate,
			Sort:                []string{"generate_date", "-material_key_id"},
			PageNumber:          2,
			PageSize:            100,
		}

		require.Len(t, filters.ElasticKeyID, 2, "Should have 2 elastic key IDs")
		require.NotNil(t, filters.MinimumGenerateDate, "MinimumGenerateDate should not be nil")
		require.NotNil(t, filters.MaximumGenerateDate, "MaximumGenerateDate should not be nil")
		require.True(t, filters.MinimumGenerateDate.Before(*filters.MaximumGenerateDate), "MinDate should be before MaxDate")
		require.Len(t, filters.Sort, 2, "Should have 2 sort fields")
		require.Equal(t, 2, filters.PageNumber, "PageNumber should be 2")
		require.Equal(t, 100, filters.PageSize, "PageSize should be 100")
	})

	t.Run("Create filters with minimal fields", func(t *testing.T) {
		filters := GetElasticKeyMaterialKeysFilters{
			PageNumber: 0,
			PageSize:   25,
		}

		require.Nil(t, filters.ElasticKeyID, "ElasticKeyID should be nil")
		require.Nil(t, filters.MinimumGenerateDate, "MinimumGenerateDate should be nil")
		require.Nil(t, filters.MaximumGenerateDate, "MaximumGenerateDate should be nil")
		require.Nil(t, filters.Sort, "Sort should be nil")
		require.Equal(t, 0, filters.PageNumber, "PageNumber should be 0")
		require.Equal(t, 25, filters.PageSize, "PageSize should be 25")
	})

	t.Run("Create filters with single elastic key ID", func(t *testing.T) {
		ekID := googleUuid.New()

		filters := GetElasticKeyMaterialKeysFilters{
			ElasticKeyID: []googleUuid.UUID{ekID},
			PageNumber:   0,
			PageSize:     10,
		}

		require.Len(t, filters.ElasticKeyID, 1, "Should have 1 elastic key ID")
		require.Equal(t, ekID, filters.ElasticKeyID[0], "Elastic Key ID should match")
	})
}

// TestGetMaterialKeysFilters tests the GetMaterialKeysFilters struct.
func TestGetMaterialKeysFilters(t *testing.T) {
	t.Run("Create filters with all fields", func(t *testing.T) {
		ekID1 := googleUuid.New()
		ekID2 := googleUuid.New()
		mkID1 := googleUuid.New()
		mkID2 := googleUuid.New()
		mkID3 := googleUuid.New()
		minDate := time.Now().UTC().Add(-7 * 24 * time.Hour)
		maxDate := time.Now().UTC()

		filters := GetMaterialKeysFilters{
			ElasticKeyID:        []googleUuid.UUID{ekID1, ekID2},
			MaterialKeyID:       []googleUuid.UUID{mkID1, mkID2, mkID3},
			MinimumGenerateDate: &minDate,
			MaximumGenerateDate: &maxDate,
			Sort:                []string{"-generate_date", "elastic_key_id"},
			PageNumber:          5,
			PageSize:            200,
		}

		require.Len(t, filters.ElasticKeyID, 2, "Should have 2 elastic key IDs")
		require.Len(t, filters.MaterialKeyID, 3, "Should have 3 material key IDs")
		require.NotNil(t, filters.MinimumGenerateDate, "MinimumGenerateDate should not be nil")
		require.NotNil(t, filters.MaximumGenerateDate, "MaximumGenerateDate should not be nil")
		require.True(t, filters.MinimumGenerateDate.Before(*filters.MaximumGenerateDate), "MinDate should be before MaxDate")
		require.Len(t, filters.Sort, 2, "Should have 2 sort fields")
		require.Equal(t, 5, filters.PageNumber, "PageNumber should be 5")
		require.Equal(t, 200, filters.PageSize, "PageSize should be 200")
	})

	t.Run("Create filters with minimal fields", func(t *testing.T) {
		filters := GetMaterialKeysFilters{
			PageNumber: 0,
			PageSize:   50,
		}

		require.Nil(t, filters.ElasticKeyID, "ElasticKeyID should be nil")
		require.Nil(t, filters.MaterialKeyID, "MaterialKeyID should be nil")
		require.Nil(t, filters.MinimumGenerateDate, "MinimumGenerateDate should be nil")
		require.Nil(t, filters.MaximumGenerateDate, "MaximumGenerateDate should be nil")
		require.Nil(t, filters.Sort, "Sort should be nil")
		require.Equal(t, 0, filters.PageNumber, "PageNumber should be 0")
		require.Equal(t, 50, filters.PageSize, "PageSize should be 50")
	})

	t.Run("Create filters with material key IDs only", func(t *testing.T) {
		mkID1 := googleUuid.New()
		mkID2 := googleUuid.New()

		filters := GetMaterialKeysFilters{
			MaterialKeyID: []googleUuid.UUID{mkID1, mkID2},
			PageNumber:    0,
			PageSize:      10,
		}

		require.Len(t, filters.MaterialKeyID, 2, "Should have 2 material key IDs")
		require.Nil(t, filters.ElasticKeyID, "ElasticKeyID should be nil")
		require.Nil(t, filters.MinimumGenerateDate, "MinimumGenerateDate should be nil")
		require.Nil(t, filters.MaximumGenerateDate, "MaximumGenerateDate should be nil")
	})

	t.Run("Create filters with date range only", func(t *testing.T) {
		minDate := time.Now().UTC().Add(-30 * 24 * time.Hour)
		maxDate := time.Now().UTC()

		filters := GetMaterialKeysFilters{
			MinimumGenerateDate: &minDate,
			MaximumGenerateDate: &maxDate,
			PageNumber:          0,
			PageSize:            100,
		}

		require.NotNil(t, filters.MinimumGenerateDate, "MinimumGenerateDate should not be nil")
		require.NotNil(t, filters.MaximumGenerateDate, "MaximumGenerateDate should not be nil")
		require.Nil(t, filters.ElasticKeyID, "ElasticKeyID should be nil")
		require.Nil(t, filters.MaterialKeyID, "MaterialKeyID should be nil")
		require.True(t, filters.MinimumGenerateDate.Before(*filters.MaximumGenerateDate), "MinDate should be before MaxDate")
	})
}
