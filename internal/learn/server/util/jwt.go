// Copyright (c) 2025 Justin Cranford

// Package util provides utility functions for the learn-im server.
package util

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	googleUuid "github.com/google/uuid"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	// ContextKeyUserID is the context key for storing user ID from JWT.
	ContextKeyUserID = "user_id"
)

// Claims represents JWT claims for learn-im authentication.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token for the given user.
func GenerateJWT(userID googleUuid.UUID, username, secret string) (string, time.Time, error) {
	expirationTime := time.Now().Add(cryptoutilSharedMagic.LearnJWTExpiration)
	claims := &Claims{
		UserID:   userID.String(),
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cryptoutilSharedMagic.LearnJWTIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, expirationTime, nil
}
