// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
)

// CredentialType represents the type of WebAuthn credential.
type CredentialType string

// CredentialTypePasskey is the credential type for passkey-based authentication.
const (
	CredentialTypePasskey CredentialType = "passkey"
)

// Credential represents a WebAuthn/FIDO2 credential.
type Credential struct {
	ID              string
	UserID          string
	Type            CredentialType
	PublicKey       []byte
	AttestationType string
	AAGUID          []byte
	SignCount       uint32
	CreatedAt       time.Time
	LastUsedAt      time.Time
	Metadata        map[string]any
}

// CredentialStore manages WebAuthn/FIDO2 credentials.
type CredentialStore interface {
	StoreCredential(ctx context.Context, credential *Credential) error
	GetCredential(ctx context.Context, credentialID string) (*Credential, error)
	GetUserCredentials(ctx context.Context, userID string) ([]*Credential, error)
	DeleteCredential(ctx context.Context, credentialID string) error
}

// WebAuthnCredential represents a WebAuthn credential in the database.
type WebAuthnCredential struct {
	// Primary key.
	ID googleUuid.UUID `gorm:"type:text;primaryKey"`

	// Foreign key to User.
	UserID googleUuid.UUID `gorm:"type:text;not null;index"`

	// WebAuthn credential ID (base64 URL-encoded).
	CredentialID string `gorm:"uniqueIndex;not null"`

	// Public key (DER-encoded).
	PublicKey []byte `gorm:"not null"`

	// Attestation type (none, indirect, direct).
	AttestationType string `gorm:"not null"`

	// AAGUID (Authenticator Attestation GUID).
	AAGUID []byte

	// Sign counter for replay attack prevention.
	SignCount uint32 `gorm:"not null;default:0"`

	// User-friendly device name.
	DeviceName string

	// Timestamps.
	CreatedAt  time.Time
	LastUsedAt time.Time
}

// TableName specifies the custom table name.
func (WebAuthnCredential) TableName() string {
	return "webauthn_credentials"
}

// WebAuthnCredentialRepository implements CredentialStore interface with GORM.
type WebAuthnCredentialRepository struct {
	db *gorm.DB
}

// NewWebAuthnCredentialRepository creates a new WebAuthn credential repository.
func NewWebAuthnCredentialRepository(db *gorm.DB) (*WebAuthnCredentialRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}

	return &WebAuthnCredentialRepository{db: db}, nil
}

// StoreCredential stores a WebAuthn credential.
func (r *WebAuthnCredentialRepository) StoreCredential(ctx context.Context, credential *Credential) error {
	if credential == nil {
		return fmt.Errorf("credential cannot be nil")
	}

	// Parse user ID.
	userID, err := googleUuid.Parse(credential.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Check if credential ID already exists (update case).
	var existing WebAuthnCredential

	result := getDB(ctx, r.db).WithContext(ctx).Where("credential_id = ?", credential.ID).First(&existing)

	if result.Error == nil {
		// Update existing credential (sign counter, last used).
		existing.SignCount = credential.SignCount
		existing.LastUsedAt = credential.LastUsedAt

		if err := getDB(ctx, r.db).WithContext(ctx).Save(&existing).Error; err != nil {
			return cryptoutilIdentityAppErr.WrapError(
				cryptoutilIdentityAppErr.ErrDatabaseQuery,
				fmt.Errorf("failed to update credential: %w", err),
			)
		}

		return nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to check existing credential: %w", result.Error),
		)
	}

	// Create new credential.
	id, err := googleUuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate credential UUID: %w", err)
	}

	dbCred := &WebAuthnCredential{
		ID:              id,
		UserID:          userID,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		AAGUID:          credential.AAGUID,
		SignCount:       credential.SignCount,
		DeviceName:      getDeviceName(credential),
		CreatedAt:       credential.CreatedAt,
		LastUsedAt:      credential.LastUsedAt,
	}

	if err := getDB(ctx, r.db).WithContext(ctx).Create(dbCred).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to create credential: %w", err),
		)
	}

	return nil
}

// GetCredential retrieves a WebAuthn credential by credential ID.
func (r *WebAuthnCredentialRepository) GetCredential(ctx context.Context, credentialID string) (*Credential, error) {
	var dbCred WebAuthnCredential

	if err := getDB(ctx, r.db).WithContext(ctx).Where("credential_id = ?", credentialID).First(&dbCred).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrCredentialNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to get credential: %w", err),
		)
	}

	return toCredential(&dbCred), nil
}

// GetUserCredentials retrieves all WebAuthn credentials for a user.
func (r *WebAuthnCredentialRepository) GetUserCredentials(ctx context.Context, userID string) ([]*Credential, error) {
	// Parse user ID.
	uid, err := googleUuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var dbCreds []WebAuthnCredential

	if err := getDB(ctx, r.db).WithContext(ctx).Where("user_id = ?", uid).Order("created_at DESC").Find(&dbCreds).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to get user credentials: %w", err),
		)
	}

	creds := make([]*Credential, 0, len(dbCreds))

	for i := range dbCreds {
		creds = append(creds, toCredential(&dbCreds[i]))
	}

	return creds, nil
}

// DeleteCredential deletes a WebAuthn credential (revocation).
func (r *WebAuthnCredentialRepository) DeleteCredential(ctx context.Context, credentialID string) error {
	result := getDB(ctx, r.db).WithContext(ctx).Where("credential_id = ?", credentialID).Delete(&WebAuthnCredential{})

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to delete credential: %w", result.Error),
		)
	}

	if result.RowsAffected == 0 {
		return cryptoutilIdentityAppErr.ErrCredentialNotFound
	}

	return nil
}

// toCredential converts database WebAuthnCredential to Credential.
func toCredential(dbCred *WebAuthnCredential) *Credential {
	return &Credential{
		ID:              dbCred.CredentialID,
		UserID:          dbCred.UserID.String(),
		Type:            CredentialTypePasskey,
		PublicKey:       dbCred.PublicKey,
		AttestationType: dbCred.AttestationType,
		AAGUID:          dbCred.AAGUID,
		SignCount:       dbCred.SignCount,
		CreatedAt:       dbCred.CreatedAt,
		LastUsedAt:      dbCred.LastUsedAt,
		Metadata: map[string]any{
			"device_name": dbCred.DeviceName,
		},
	}
}

// getDeviceName extracts device name from credential metadata.
func getDeviceName(cred *Credential) string {
	if cred.Metadata != nil {
		if name, ok := cred.Metadata["device_name"].(string); ok {
			return name
		}
	}

	return "Unknown Device"
}
