// Copyright (c) 2025 Justin Cranford
//

// Package realms provides generic authentication and authorization services.
// This package is part of the service template infrastructure and is designed
// to be reusable across all cryptoutil services (sm-im, jose-ja, identity, ca).
//
// Domain-Agnostic Design:
// - UserModel interface abstracts different User entities (sm.User, jose.User, identity.User)
// - UserRepository interface abstracts different persistence layers (GORM, mock, etc.)
// - Services implement these interfaces for their specific domain models
//
// Key Features:
// - JWT-based authentication with HMAC-SHA256 signing
// - User registration with PBKDF2-HMAC-SHA256 password hashing (FIPS-approved, OWASP 2023)
// - User login with credential validation
// - JWT middleware for protecting authenticated routes
//
// Usage Pattern (Example for sm-im):
//
//	// 1. Domain model implements UserModel interface
//	type User struct {
//	    ID           googleUuid.UUID
//	    Username     string
//	    PasswordHash string
//	    CreatedAt    time.Time
//	}
//
//	func (u *User) GetID() googleUuid.UUID          { return u.ID }
//	func (u *User) GetUsername() string             { return u.Username }
//	func (u *User) GetPasswordHash() string         { return u.PasswordHash }
//	func (u *User) SetID(id googleUuid.UUID)        { u.ID = id }
//	func (u *User) SetUsername(username string)     { u.Username = username }
//	func (u *User) SetPasswordHash(hash string)     { u.PasswordHash = hash }
//
//	// 2. Repository implements UserRepository interface
//	type UserRepository struct { db *gorm.DB }
//
//	func (r *UserRepository) Create(ctx context.Context, user realms.UserModel) error {
//	    return r.db.WithContext(ctx).Create(user).Error
//	}
//
//	func (r *UserRepository) FindByUsername(ctx context.Context, username string) (realms.UserModel, error) {
//	    var u User
//	    err := r.db.WithContext(ctx).First(&u, "username = ?", username).Error
//	    return &u, err
//	}
//
//	func (r *UserRepository) FindByID(ctx context.Context, id googleUuid.UUID) (realms.UserModel, error) {
//	    var u User
//	    err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
//	    return &u, err
//	}
//
//	// 3. Wire into server
//	authnHandler := realms.NewAuthnHandler(realms.AuthnConfig{
//	    UserRepo:      userRepo,
//	    JWTSecret:     jwtSecret,
//	    JWTExpiration: 15 * time.Minute,
//	    UserModelFactory: func() realms.UserModel { return &User{} },
//	})
//
//	app.Post("/service/api/v1/users/register", authnHandler.HandleRegisterUser())
//	app.Post("/service/api/v1/users/login", authnHandler.HandleLoginUser())
//	app.Get("/service/api/v1/protected", realms.JWTMiddleware(jwtSecret), protectedHandler)
package realms

import (
	"context"

	googleUuid "github.com/google/uuid"
)

// UserModel abstracts domain User entities across different cryptoutil services.
//
// Services implement this interface for their specific User struct to enable
// generic authentication handlers. The interface requires only the fields
// necessary for authentication (ID, Username, PasswordHash).
//
// Domain models MAY extend with additional fields (Email, Role, CreatedAt, etc.)
// without affecting the authentication logic.
//
// Example Implementation:
//
//	type User struct {
//	    ID           googleUuid.UUID `gorm:"type:text;primaryKey"`
//	    Username     string          `gorm:"type:text;uniqueIndex;not null"`
//	    PasswordHash string          `gorm:"type:text;not null"`
//	    CreatedAt    time.Time       `gorm:"autoCreateTime"`
//	}
//
//	func (u *User) GetID() googleUuid.UUID          { return u.ID }
//	func (u *User) GetUsername() string             { return u.Username }
//	func (u *User) GetPasswordHash() string         { return u.PasswordHash }
//	func (u *User) SetID(id googleUuid.UUID)        { u.ID = id }
//	func (u *User) SetUsername(username string)     { u.Username = username }
//	func (u *User) SetPasswordHash(hash string)     { u.PasswordHash = hash }
//
// Compile-Time Interface Check:
//
//	var _ realms.UserModel = (*User)(nil)
type UserModel interface {
	// GetID returns the user's unique identifier (UUIDv7).
	// Used for JWT token generation and user lookup.
	GetID() googleUuid.UUID

	// GetUsername returns the user's username.
	// Used for login validation and duplicate username checks.
	GetUsername() string

	// GetPasswordHash returns the PBKDF2-hashed password (versioned format).
	// Used for password verification during login.
	GetPasswordHash() string

	// SetID sets the user's unique identifier (UUIDv7).
	// Called during user registration to assign a new ID.
	SetID(googleUuid.UUID)

	// SetUsername sets the user's username.
	// Called during user registration to store the provided username.
	SetUsername(string)

	// SetPasswordHash sets the PBKDF2-hashed password (versioned format).
	// Called during user registration after hashing the plaintext password.
	SetPasswordHash(string)
}

// UserRepository abstracts user persistence across different database implementations.
//
// Services implement this interface for their specific database layer (GORM, mock, etc.)
// to enable generic authentication handlers. The interface requires only CRUD operations
// necessary for authentication (Create, FindByUsername, FindByID).
//
// Repository implementations SHOULD support transactions via context.Context pattern:
//
//	type txKey struct{}
//
//	func WithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
//	    return context.WithValue(ctx, txKey{}, tx)
//	}
//
//	func getDB(ctx context.Context, baseDB *gorm.DB) *gorm.DB {
//	    if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
//	        return tx
//	    }
//	    return baseDB
//	}
//
// Example Implementation:
//
//	type UserRepository struct {
//	    db *gorm.DB
//	}
//
//	func (r *UserRepository) Create(ctx context.Context, user realms.UserModel) error {
//	    return getDB(ctx, r.db).WithContext(ctx).Create(user).Error
//	}
//
//	func (r *UserRepository) FindByUsername(ctx context.Context, username string) (realms.UserModel, error) {
//	    var u User
//	    err := getDB(ctx, r.db).WithContext(ctx).First(&u, "username = ?", username).Error
//	    return &u, err
//	}
//
//	func (r *UserRepository) FindByID(ctx context.Context, id googleUuid.UUID) (realms.UserModel, error) {
//	    var u User
//	    err := getDB(ctx, r.db).WithContext(ctx).First(&u, "id = ?", id).Error
//	    return &u, err
//	}
//
// Compile-Time Interface Check:
//
//	var _ realms.UserRepository = (*UserRepository)(nil)
type UserRepository interface {
	// Create persists a new user to the database.
	// Returns error if username already exists (duplicate key violation).
	//
	// Example:
	//   user := &User{ID: googleUuid.NewV7(), Username: "alice", PasswordHash: hash}
	//   err := repo.Create(ctx, user)
	Create(ctx context.Context, user UserModel) error

	// FindByUsername retrieves a user by username.
	// Returns error if user not found (gorm.ErrRecordNotFound).
	//
	// Example:
	//   user, err := repo.FindByUsername(ctx, "alice")
	FindByUsername(ctx context.Context, username string) (UserModel, error)

	// FindByID retrieves a user by unique identifier.
	// Returns error if user not found (gorm.ErrRecordNotFound).
	//
	// Example:
	//   user, err := repo.FindByID(ctx, userID)
	FindByID(ctx context.Context, id googleUuid.UUID) (UserModel, error)
}
