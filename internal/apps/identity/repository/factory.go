// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityORM "cryptoutil/internal/apps/identity/repository/orm"
)

// contextKey is the type for context keys to avoid collisions.
type contextKey string

// txKey is the context key for storing the transaction DB.
const txKey contextKey = "gorm_tx"

// RepositoryFactory creates and manages repository instances.
type RepositoryFactory struct {
	db                 *gorm.DB
	dbType             string
	userRepo           UserRepository
	clientRepo         ClientRepository
	tokenRepo          TokenRepository
	sessionRepo        SessionRepository
	clientProfileRepo  ClientProfileRepository
	authFlowRepo       AuthFlowRepository
	authProfileRepo    AuthProfileRepository
	mfaFactorRepo      MFAFactorRepository
	authzReqRepo       AuthorizationRequestRepository
	consentRepo        ConsentDecisionRepository
	keyRepo            KeyRepository
	deviceAuthRepo     DeviceAuthorizationRepository
	parRepo            PushedAuthorizationRequestRepository
	recoveryCodeRepo   RecoveryCodeRepository
	emailOTPRepo       EmailOTPRepository
	jtiReplayCacheRepo JTIReplayCacheRepository
}

// NewRepositoryFactory creates a new repository factory with database initialization.
func NewRepositoryFactory(ctx context.Context, cfg *cryptoutilIdentityConfig.DatabaseConfig) (*RepositoryFactory, error) {
	db, err := initializeDatabase(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &RepositoryFactory{
		db:                 db,
		dbType:             cfg.Type,
		userRepo:           cryptoutilIdentityORM.NewUserRepository(db),
		clientRepo:         cryptoutilIdentityORM.NewClientRepository(db),
		tokenRepo:          cryptoutilIdentityORM.NewTokenRepository(db),
		sessionRepo:        cryptoutilIdentityORM.NewSessionRepository(db),
		clientProfileRepo:  cryptoutilIdentityORM.NewClientProfileRepository(db),
		authFlowRepo:       cryptoutilIdentityORM.NewAuthFlowRepository(db),
		authProfileRepo:    cryptoutilIdentityORM.NewAuthProfileRepository(db),
		mfaFactorRepo:      cryptoutilIdentityORM.NewMFAFactorRepository(db),
		authzReqRepo:       cryptoutilIdentityORM.NewAuthorizationRequestRepository(db),
		consentRepo:        cryptoutilIdentityORM.NewConsentDecisionRepository(db),
		keyRepo:            cryptoutilIdentityORM.NewKeyRepository(db),
		deviceAuthRepo:     cryptoutilIdentityORM.NewDeviceAuthorizationRepository(db),
		parRepo:            cryptoutilIdentityORM.NewPushedAuthorizationRequestRepository(db),
		recoveryCodeRepo:   cryptoutilIdentityORM.NewRecoveryCodeRepository(db),
		emailOTPRepo:       cryptoutilIdentityORM.NewEmailOTPRepository(db),
		jtiReplayCacheRepo: NewJTIReplayCacheRepository(db),
	}, nil
}

// User returns the user repository.
func (f *RepositoryFactory) User() UserRepository {
	return f.userRepo
}

// UserRepository returns the user repository (alias for User for backwards compatibility).
func (f *RepositoryFactory) UserRepository() UserRepository {
	return f.userRepo
}

// UserWithContext returns the user repository with transaction support from context.
func (f *RepositoryFactory) UserWithContext(ctx context.Context) UserRepository {
	return cryptoutilIdentityORM.NewUserRepository(f.getDB(ctx))
}

// ClientRepository returns the client repository.
func (f *RepositoryFactory) ClientRepository() ClientRepository {
	return f.clientRepo
}

// TokenRepository returns the token repository.
func (f *RepositoryFactory) TokenRepository() TokenRepository {
	return f.tokenRepo
}

// SessionRepository returns the session repository.
func (f *RepositoryFactory) SessionRepository() SessionRepository {
	return f.sessionRepo
}

// ClientProfileRepository returns the client profile repository.
func (f *RepositoryFactory) ClientProfileRepository() ClientProfileRepository {
	return f.clientProfileRepo
}

// AuthFlowRepository returns the authorization flow repository.
func (f *RepositoryFactory) AuthFlowRepository() AuthFlowRepository {
	return f.authFlowRepo
}

// AuthProfileRepository returns the authentication profile repository.
func (f *RepositoryFactory) AuthProfileRepository() AuthProfileRepository {
	return f.authProfileRepo
}

// MFAFactorRepository returns the MFA factor repository.
func (f *RepositoryFactory) MFAFactorRepository() MFAFactorRepository {
	return f.mfaFactorRepo
}

// AuthorizationRequestRepository returns the authorization request repository.
func (f *RepositoryFactory) AuthorizationRequestRepository() AuthorizationRequestRepository {
	return f.authzReqRepo
}

// ConsentDecisionRepository returns the consent decision repository.
func (f *RepositoryFactory) ConsentDecisionRepository() ConsentDecisionRepository {
	return f.consentRepo
}

// KeyRepository returns the cryptographic key repository.
func (f *RepositoryFactory) KeyRepository() KeyRepository {
	return f.keyRepo
}

// DeviceAuthorizationRepository returns the device authorization repository.
func (f *RepositoryFactory) DeviceAuthorizationRepository() DeviceAuthorizationRepository {
	return f.deviceAuthRepo
}

// PushedAuthorizationRequestRepository returns the pushed authorization request repository.
func (f *RepositoryFactory) PushedAuthorizationRequestRepository() PushedAuthorizationRequestRepository {
	return f.parRepo
}

// RecoveryCodeRepository returns the recovery code repository.
func (f *RepositoryFactory) RecoveryCodeRepository() RecoveryCodeRepository {
	return f.recoveryCodeRepo
}

// EmailOTPRepository returns the email OTP repository.
func (f *RepositoryFactory) EmailOTPRepository() EmailOTPRepository {
	return f.emailOTPRepo
}

// JTIReplayCacheRepository returns the JTI replay cache repository.
func (f *RepositoryFactory) JTIReplayCacheRepository() JTIReplayCacheRepository {
	return f.jtiReplayCacheRepo
}

// DB returns the underlying database connection for transaction management.
func (f *RepositoryFactory) DB() *gorm.DB {
	return f.db
}

// Transaction executes a function within a database transaction.
func (f *RepositoryFactory) Transaction(ctx context.Context, fn func(context.Context) error) error {
	err := f.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Enable debug mode to see actual SQL statements during transaction.
		tx = tx.Debug()
		txCtx := context.WithValue(ctx, txKey, tx)

		return fn(txCtx)
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

// getDB returns the transaction DB from context if present, otherwise returns the base DB.
func (f *RepositoryFactory) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}

	return f.db
}

// Close closes the database connection.
func (f *RepositoryFactory) Close() error {
	sqlDB, err := f.db.DB()
	if err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseConnection,
			fmt.Errorf("failed to get database instance: %w", err),
		)
	}

	if err := sqlDB.Close(); err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseConnection,
			fmt.Errorf("failed to close database: %w", err),
		)
	}

	return nil
}

// NewRepositoryFactoryForTesting creates a minimal RepositoryFactory for unit tests,
// injecting only the repositories needed for the code under test.
// The db field is left nil; do not call Close() or AutoMigrate() on the result.
func NewRepositoryFactoryForTesting(userRepo UserRepository, clientRepo ClientRepository) *RepositoryFactory {
	return &RepositoryFactory{
		userRepo:   userRepo,
		clientRepo: clientRepo,
	}
}

// AutoMigrate runs database migrations using golang-migrate with embedded SQL files.
func (f *RepositoryFactory) AutoMigrate(_ context.Context) error {
	// Get underlying *sql.DB from GORM DB instance.
	sqlDB, err := f.db.DB()
	if err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseConnection,
			fmt.Errorf("failed to get sql.DB: %w", err),
		)
	}

	// Apply migrations using golang-migrate with database type.
	if err := Migrate(sqlDB, f.dbType); err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("database migration failed: %w", err),
		)
	}

	return nil
}
