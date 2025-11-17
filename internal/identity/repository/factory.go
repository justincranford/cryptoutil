package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityORM "cryptoutil/internal/identity/repository/orm"
)

// contextKey is the type for context keys to avoid collisions.
type contextKey string

// txKey is the context key for storing the transaction DB.
const txKey contextKey = "gorm_tx"

// RepositoryFactory creates and manages repository instances.
type RepositoryFactory struct {
	db                *gorm.DB
	userRepo          UserRepository
	clientRepo        ClientRepository
	tokenRepo         TokenRepository
	sessionRepo       SessionRepository
	clientProfileRepo ClientProfileRepository
	authFlowRepo      AuthFlowRepository
	authProfileRepo   AuthProfileRepository
	mfaFactorRepo     MFAFactorRepository
}

// NewRepositoryFactory creates a new repository factory with database initialization.
func NewRepositoryFactory(ctx context.Context, cfg *cryptoutilIdentityConfig.DatabaseConfig) (*RepositoryFactory, error) {
	db, err := initializeDatabase(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &RepositoryFactory{
		db:                db,
		userRepo:          cryptoutilIdentityORM.NewUserRepository(db),
		clientRepo:        cryptoutilIdentityORM.NewClientRepository(db),
		tokenRepo:         cryptoutilIdentityORM.NewTokenRepository(db),
		sessionRepo:       cryptoutilIdentityORM.NewSessionRepository(db),
		clientProfileRepo: cryptoutilIdentityORM.NewClientProfileRepository(db),
		authFlowRepo:      cryptoutilIdentityORM.NewAuthFlowRepository(db),
		authProfileRepo:   cryptoutilIdentityORM.NewAuthProfileRepository(db),
		mfaFactorRepo:     cryptoutilIdentityORM.NewMFAFactorRepository(db),
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

// DB returns the underlying database connection for transaction management.
func (f *RepositoryFactory) DB() *gorm.DB {
	return f.db
}

// Transaction executes the given function within a database transaction.
func (f *RepositoryFactory) Transaction(ctx context.Context, fn func(context.Context) error) error {
	if err := f.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Store transaction DB in context so repositories can use it.
		txCtx := context.WithValue(ctx, txKey, tx)

		return fn(txCtx)
	}); err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseTransaction,
			fmt.Errorf("transaction failed: %w", err),
		)
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

	// Apply migrations using golang-migrate.
	if err := Migrate(sqlDB); err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("database migration failed: %w", err),
		)
	}

	return nil
}
