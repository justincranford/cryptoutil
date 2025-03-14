package repository

import (
	"cryptoutil/database"
	"cryptoutil/uuid"
	"time"
)

type KEK struct {
	ID                  string    `gorm:"primaryKey;type:text"`
	Name                string    `gorm:"uniqueIndex:idx_kek_name_id;not null;check:length(name) BETWEEN 3 AND 50"`
	Description         *string   `gorm:"type:text"`
	Algorithm           string    `gorm:"not null;check:length(algorithm) BETWEEN 5 AND 20"`
	Status              string    `gorm:"not null;check:status IN ('active', 'disabled', 'pending_import', 'expired')"`
	Provider            string    `gorm:"not null;check:provider IN ('Internal', 'AWS', 'GCP', 'Azure')"`
	IsVersioningEnabled bool      `gorm:"not null"`
	IsImported          bool      `gorm:"not null"`
	IsExportable        bool      `gorm:"not null"`
	CreateDate          time.Time `gorm:"not null;check:create_date GLOB '[0-9]*T[0-9]*Z'"`
}

type KEKVersion struct {
	KEKID       string     `gorm:"not null;index:idx_kek_version_version_kek_id"`
	Version     string     `gorm:"primaryKey;type:text;check:length(version) BETWEEN 5 AND 100"`
	KeyMaterial *string    `gorm:"type:text"`
	UpdateDate  *time.Time `gorm:"check:update_date GLOB '[0-9]*T[0-9]*Z'"`
	DeleteDate  *time.Time `gorm:"check:delete_date GLOB '[0-9]*T[0-9]*Z'"`
	ImportDate  *time.Time `gorm:"check:import_date GLOB '[0-9]*T[0-9]*Z'"`
	CreateDate  time.Time  `gorm:"not null;check:create_date GLOB '[0-9]*T[0-9]*Z'"`

	KEK KEK `gorm:"foreignKey:KEKID;constraint:OnDelete:CASCADE"`
}

type CreateKEK struct {
	Name                string
	Description         *string
	Algorithm           string
	Status              string
	Provider            string
	IsVersioningEnabled bool
	IsImported          bool
	IsExportable        bool
}

type CreateKEKVersion struct {
	KEKID       string
	Version     string
	KeyMaterial *string
	ImportDate  *time.Time
}

type KEKRepository struct {
	databaseService *database.Service
}

func NewKEKRepository(databaseService *database.Service) *KEKRepository {
	return &KEKRepository{databaseService: databaseService}
}

func (r *KEKRepository) CreateKEK(input *CreateKEK) (*KEK, error) {
	newKEK := &KEK{
		ID:                  uuid.V7(),
		Name:                input.Name,
		Description:         input.Description,
		Algorithm:           input.Algorithm,
		Status:              input.Status,
		Provider:            input.Provider,
		IsVersioningEnabled: input.IsVersioningEnabled,
		IsImported:          input.IsImported,
		IsExportable:        input.IsExportable,
		CreateDate:          time.Now(),
	}

	if err := r.databaseService.GormDB().Create(newKEK).Error; err != nil {
		return nil, err
	}
	return newKEK, nil
}

func (r *KEKRepository) ReadKEKs() ([]KEK, error) {
	var keks []KEK
	if err := r.databaseService.GormDB().Find(&keks).Error; err != nil {
		return nil, err
	}
	return keks, nil
}

func (r *KEKRepository) UpdateKEK(kek *KEK) (*KEK, error) {
	if err := r.databaseService.GormDB().Save(kek).Error; err != nil {
		return nil, err
	}
	var updatedKEK KEK
	if err := r.databaseService.GormDB().First(&updatedKEK, "id = ?", kek.ID).Error; err != nil {
		return nil, err
	}
	return &updatedKEK, nil
}

func (r *KEKRepository) DeleteKEK(id string) (*KEK, error) {
	var deletedKEK KEK
	if err := r.databaseService.GormDB().Where("id = ?", id).First(&deletedKEK).Error; err != nil {
		return nil, err
	}
	if err := r.databaseService.GormDB().Delete(&deletedKEK).Error; err != nil {
		return nil, err
	}
	return &deletedKEK, nil
}

func (r *KEKRepository) CreateKEKVersion(input *CreateKEKVersion) (*KEKVersion, error) {
	newKEKVersion := &KEKVersion{
		KEKID:       input.KEKID,
		Version:     input.Version,
		KeyMaterial: input.KeyMaterial,
		ImportDate:  input.ImportDate,
		CreateDate:  time.Now(),
	}

	if err := r.databaseService.GormDB().Create(newKEKVersion).Error; err != nil {
		return nil, err
	}
	return newKEKVersion, nil
}

func (r *KEKRepository) ReadKEKVersions(kekID string) ([]KEKVersion, error) {
	var kekVersions []KEKVersion
	if err := r.databaseService.GormDB().Where("kek_id = ?", kekID).Find(&kekVersions).Error; err != nil {
		return nil, err
	}
	return kekVersions, nil
}

func (r *KEKRepository) UpdateKEKVersion(kekVersion *KEKVersion) (*KEKVersion, error) {
	if err := r.databaseService.GormDB().Save(kekVersion).Error; err != nil {
		return nil, err
	}

	// Re-fetch the updated record to reflect changes from DB triggers
	var updatedKEKVersion KEKVersion
	if err := r.databaseService.GormDB().First(&updatedKEKVersion, "kek_id = ? AND version = ?", kekVersion.KEKID, kekVersion.Version).Error; err != nil {
		return nil, err
	}
	return &updatedKEKVersion, nil
}

func (r *KEKRepository) DeleteKEKVersion(kekID, version string) (*KEKVersion, error) {
	var kekVersion KEKVersion
	if err := r.databaseService.GormDB().Where("kek_id = ? AND version = ?", kekID, version).First(&kekVersion).Error; err != nil {
		return nil, err
	}
	if err := r.databaseService.GormDB().Delete(&kekVersion).Error; err != nil {
		return nil, err
	}
	return &kekVersion, nil
}
