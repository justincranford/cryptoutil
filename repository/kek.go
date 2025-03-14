package repository

import (
	"time"

	"gorm.io/gorm"
)

type Kek struct {
	ID               string    `gorm:"primaryKey;type:text"`
	Name             string    `gorm:"uniqueIndex:idx_kek_name_id;not null;check:length(name) BETWEEN 3 AND 50"`
	Description      *string   `gorm:"type:text"`
	Algorithm        string    `gorm:"not null;check:length(algorithm) BETWEEN 5 AND 20"`
	CreateDate       time.Time `gorm:"not null;check:create_date GLOB '[0-9]*T[0-9]*Z'"`
	Status           string    `gorm:"not null;check:status IN ('pending_import', 'active', 'disabled', 'expired')"`
	EnableVersioning bool      `gorm:"not null"`
	Imported         bool      `gorm:"not null"`
	Provider         string    `gorm:"not null;check:provider IN ('AWS', 'GCP', 'Azure', 'Internal')"`
	Exportable       bool      `gorm:"not null"`
	KeyMaterial      *string   `gorm:"type:text"`
}

type KekVersion struct {
	Version    string     `gorm:"primaryKey;type:text;check:length(version) BETWEEN 5 AND 100"`
	KekID      string     `gorm:"not null;index:idx_kek_version_version_kek_id"`
	CreateDate time.Time  `gorm:"not null;check:create_date GLOB '[0-9]*T[0-9]*Z'"`
	UpdateDate *time.Time `gorm:"check:update_date GLOB '[0-9]*T[0-9]*Z'"`
	DeleteDate *time.Time `gorm:"check:delete_date GLOB '[0-9]*T[0-9]*Z'"`
	ImportDate *time.Time `gorm:"check:import_date GLOB '[0-9]*T[0-9]*Z'"`

	Kek Kek `gorm:"foreignKey:KekID;constraint:OnDelete:CASCADE"`
}

type KekRepository struct {
	DB *gorm.DB
}

func NewKekRepository(db *gorm.DB) *KekRepository {
	return &KekRepository{DB: db}
}

func (r *KekRepository) CreateKek(kek *Kek) error {
	return r.DB.Create(kek).Error
}

func (r *KekRepository) GetKekByID(id string) (*Kek, error) {
	var kek Kek
	if err := r.DB.First(&kek, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &kek, nil
}

func (r *KekRepository) ListKeks() ([]Kek, error) {
	var keks []Kek
	if err := r.DB.Find(&keks).Error; err != nil {
		return nil, err
	}
	return keks, nil
}

func (r *KekRepository) UpdateKek(kek *Kek) error {
	return r.DB.Save(kek).Error
}

func (r *KekRepository) DeleteKek(id string) error {
	return r.DB.Delete(&Kek{}, "id = ?", id).Error
}

func (r *KekRepository) CreateKekVersion(version *KekVersion) error {
	return r.DB.Create(version).Error
}

func (r *KekRepository) GetKekVersion(kekID, version string) (*KekVersion, error) {
	var kekVersion KekVersion
	if err := r.DB.First(&kekVersion, "kek_id = ? AND version = ?", kekID, version).Error; err != nil {
		return nil, err
	}
	return &kekVersion, nil
}

func (r *KekRepository) ListKekVersions(kekID string) ([]KekVersion, error) {
	var versions []KekVersion
	if err := r.DB.Where("kek_id = ?", kekID).Find(&versions).Error; err != nil {
		return nil, err
	}
	return versions, nil
}

func (r *KekRepository) UpdateKekVersion(version *KekVersion) error {
	return r.DB.Save(version).Error
}

func (r *KekRepository) DeleteKekVersion(kekID, version string) error {
	return r.DB.Delete(&KekVersion{}, "kek_id = ? AND version = ?", kekID, version).Error
}
