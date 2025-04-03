package orm

import (
	"time"

	googleUuid "github.com/google/uuid"
)

type GetKeyPoolsFilters struct {
	ID                []googleUuid.UUID `validate:"optional,min=1"`
	Name              []string          `validate:"optional,min=1"`
	Algorithm         []string          `validate:"optional,min=1"`
	VersioningAllowed *bool             `validate:"optional"`
	ImportAllowed     *bool             `validate:"optional"`
	ExportAllowed     *bool             `validate:"optional"`
	Sort              []string          `validate:"optional,min=1"`
	PageNumber        int               `validate:"min=0"`
	PageSize          int               `validate:"min=1"`
}

type GetKeyPoolKeysFilters struct {
	ID                  []googleUuid.UUID `validate:"optional,min=1"`
	MinimumGenerateDate *time.Time        `validate:"optional"`
	MaximumGenerateDate *time.Time        `validate:"optional"`
	Sort                []string          `validate:"optional,min=1"`
	PageNumber          int               `validate:"min=0"`
	PageSize            int               `validate:"min=1"`
}

type GetKeysFilters struct {
	Pool                []googleUuid.UUID `validate:"optional,min=1"`
	ID                  []googleUuid.UUID `validate:"optional,min=1"`
	MinimumGenerateDate *time.Time        `validate:"optional"`
	MaximumGenerateDate *time.Time        `validate:"optional"`
	Sort                []string          `validate:"optional,min=1"`
	PageNumber          int               `validate:"min=0"`
	PageSize            int               `validate:"min=1"`
}
