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
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// SessionRepositoryGORM implements the SessionRepository interface using GORM.
type SessionRepositoryGORM struct {
	db *gorm.DB
}

// NewSessionRepository creates a new SessionRepositoryGORM.
func NewSessionRepository(db *gorm.DB) *SessionRepositoryGORM {
	return &SessionRepositoryGORM{db: db}
}

// Create creates a new session.
func (r *SessionRepositoryGORM) Create(ctx context.Context, session *cryptoutilIdentityDomain.Session) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(session).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create session: %w", err))
	}

	return nil
}

// GetByID retrieves a session by ID.
func (r *SessionRepositoryGORM) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.Session, error) {
	var session cryptoutilIdentityDomain.Session
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrSessionNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get session by ID: %w", err))
	}

	return &session, nil
}

// GetBySessionID retrieves a session by session_id.
func (r *SessionRepositoryGORM) GetBySessionID(ctx context.Context, sessionID string) (*cryptoutilIdentityDomain.Session, error) {
	var session cryptoutilIdentityDomain.Session
	if err := getDB(ctx, r.db).WithContext(ctx).Where("session_id = ? AND deleted_at IS NULL", sessionID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrSessionNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get session by session_id: %w", err))
	}

	return &session, nil
}

// Update updates an existing session.
func (r *SessionRepositoryGORM) Update(ctx context.Context, session *cryptoutilIdentityDomain.Session) error {
	session.UpdatedAt = time.Now().UTC()
	if err := getDB(ctx, r.db).WithContext(ctx).Save(session).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to update session: %w", err))
	}

	return nil
}

// Delete deletes a session by ID (soft delete).
func (r *SessionRepositoryGORM) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).Delete(&cryptoutilIdentityDomain.Session{}).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete session: %w", err))
	}

	return nil
}

// TerminateByID terminates a session by ID.
func (r *SessionRepositoryGORM) TerminateByID(ctx context.Context, id googleUuid.UUID) error {
	result := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.Session{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"active": false,
		})

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to terminate session by ID: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return cryptoutilIdentityAppErr.ErrSessionNotFound
	}

	return nil
}

// TerminateBySessionID terminates a session by session_id.
func (r *SessionRepositoryGORM) TerminateBySessionID(ctx context.Context, sessionID string) error {
	result := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.Session{}).
		Where("session_id = ? AND deleted_at IS NULL", sessionID).
		Updates(map[string]any{
			"active": false,
		})

	if result.Error != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to terminate session by session_id: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return cryptoutilIdentityAppErr.ErrSessionNotFound
	}

	return nil
}

// DeleteExpired deletes expired sessions (hard delete).
func (r *SessionRepositoryGORM) DeleteExpired(ctx context.Context) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Unscoped().
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&cryptoutilIdentityDomain.Session{}).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete expired sessions: %w", err))
	}

	return nil
}

// DeleteExpiredBefore deletes all sessions expired before the given time (hard delete).
func (r *SessionRepositoryGORM) DeleteExpiredBefore(ctx context.Context, beforeTime time.Time) (int, error) {
	result := getDB(ctx, r.db).WithContext(ctx).Unscoped().
		Where("expires_at < ?", beforeTime).
		Delete(&cryptoutilIdentityDomain.Session{})

	if result.Error != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete expired sessions before %s: %w", beforeTime, result.Error))
	}

	return int(result.RowsAffected), nil
}

// List lists sessions with pagination.
func (r *SessionRepositoryGORM) List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.Session, error) {
	var sessions []*cryptoutilIdentityDomain.Session
	if err := getDB(ctx, r.db).WithContext(ctx).Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&sessions).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to list sessions: %w", err))
	}

	return sessions, nil
}

// Count returns the total number of sessions.
func (r *SessionRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.Session{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to count sessions: %w", err))
	}

	return count, nil
}
