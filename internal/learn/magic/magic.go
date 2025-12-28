// Copyright (c) 2025 Justin Cranford

// Package magic contains magic values for the learn-im service.
package magic

import "time"

// User registration and authentication constants.
const (
	// MinUsernameLength is the minimum acceptable username length.
	MinUsernameLength = 3

	// MaxUsernameLength is the maximum acceptable username length.
	MaxUsernameLength = 50

	// MinPasswordLength is the minimum acceptable password length.
	MinPasswordLength = 8
)

// JWT token configuration.
const (
	// JWTIssuer is the issuer claim for JWT tokens.
	JWTIssuer = "learn-im"

	// JWTExpiration is the default JWT token expiration time.
	JWTExpiration = 24 * time.Hour
)
