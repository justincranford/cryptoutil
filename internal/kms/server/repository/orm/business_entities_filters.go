// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// GetElasticKeysFilters contains filter criteria for querying elastic keys.
type GetElasticKeysFilters struct {
	ElasticKeyID      []googleUuid.UUID `validate:"optional,min=1"`
	Name              []string          `validate:"optional,min=1"`
	Algorithm         []string          `validate:"optional,min=1"`
	VersioningAllowed *bool             `validate:"optional"`
	ImportAllowed     *bool             `validate:"optional"`
	ExportAllowed     *bool             `validate:"optional"`
	Sort              []string          `validate:"optional,min=1"`
	PageNumber        int               `validate:"min=0"`
	PageSize          int               `validate:"min=1"`
}

// GetElasticKeyMaterialKeysFilters contains filter criteria for querying material keys within an elastic key.
type GetElasticKeyMaterialKeysFilters struct {
	ElasticKeyID        []googleUuid.UUID `validate:"optional,min=1"`
	MinimumGenerateDate *time.Time        `validate:"optional"`
	MaximumGenerateDate *time.Time        `validate:"optional"`
	Sort                []string          `validate:"optional,min=1"`
	PageNumber          int               `validate:"min=0"`
	PageSize            int               `validate:"min=1"`
}

// GetMaterialKeysFilters contains filter criteria for querying material keys across all elastic keys.
type GetMaterialKeysFilters struct {
	ElasticKeyID        []googleUuid.UUID `validate:"optional,min=1"`
	MaterialKeyID       []googleUuid.UUID `validate:"optional,min=1"`
	MinimumGenerateDate *time.Time        `validate:"optional"`
	MaximumGenerateDate *time.Time        `validate:"optional"`
	Sort                []string          `validate:"optional,min=1"`
	PageNumber          int               `validate:"min=0"`
	PageSize            int               `validate:"min=1"`
}
