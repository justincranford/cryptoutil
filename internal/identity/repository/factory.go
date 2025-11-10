package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityORM "cryptoutil/internal/identity/repository/orm"
)

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

// UserRepository returns the user repository.
func (f *RepositoryFactory) UserRepository() UserRepository {
	return f.userRepo
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
		return fn(ctx)
	}); err != nil {
		return cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseTransaction,
			fmt.Errorf("transaction failed: %w", err),
		)
	}

	return nil
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

// AutoMigrate runs database migrations for all domain models.
func (f *RepositoryFactory) AutoMigrate(ctx context.Context) error {
	// For SQLite, disable foreign key constraints during migration to avoid circular dependency errors.
	_ = f.db.WithContext(ctx).Exec("PRAGMA foreign_keys = OFF").Error

	// Migrate one model at a time with error logging to identify which model fails.
	models := []any{
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.Token{},
		&cryptoutilIdentityDomain.Session{},
		&cryptoutilIdentityDomain.ClientProfile{},
		&cryptoutilIdentityDomain.AuthFlow{},
		&cryptoutilIdentityDomain.AuthProfile{},
		&cryptoutilIdentityDomain.MFAFactor{},
	}

	for _, model := range models {
		if err := f.db.WithContext(ctx).AutoMigrate(model); err != nil {
			return cryptoutilIdentityAppErr.WrapError(
				cryptoutilIdentityAppErr.ErrDatabaseQuery,
				fmt.Errorf("auto-migration failed for %T: %w", model, err),
			)
		}
	}

	// Re-enable foreign key constraints for SQLite.
	_ = f.db.WithContext(ctx).Exec("PRAGMA foreign_keys = ON").Error

	return nil
}
