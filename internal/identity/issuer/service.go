// Copyright (c) 2025 Justin Cranford
//
//

package issuer

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// TokenService manages token issuance, validation, and introspection.
type TokenService struct {
	jwsIssuer  *JWSIssuer
	jweIssuer  *JWEIssuer
	uuidIssuer *UUIDIssuer
	config     *cryptoutilIdentityConfig.TokenConfig
}

// NewTokenService creates a new token service with the specified configuration.
func NewTokenService(
	jwsIssuer *JWSIssuer,
	jweIssuer *JWEIssuer,
	uuidIssuer *UUIDIssuer,
	config *cryptoutilIdentityConfig.TokenConfig,
) *TokenService {
	return &TokenService{
		jwsIssuer:  jwsIssuer,
		jweIssuer:  jweIssuer,
		uuidIssuer: uuidIssuer,
		config:     config,
	}
}

// IssueAccessToken issues an access token with the configured format.
func (s *TokenService) IssueAccessToken(ctx context.Context, claims map[string]any) (string, error) {
	switch s.config.AccessTokenFormat {
	case cryptoutilIdentityMagic.TokenFormatJWS:
		return s.jwsIssuer.IssueAccessToken(ctx, claims)
	case cryptoutilIdentityMagic.TokenFormatJWE:
		jws, err := s.jwsIssuer.IssueAccessToken(ctx, claims)
		if err != nil {
			return "", err
		}

		return s.jweIssuer.EncryptToken(ctx, jws)
	case cryptoutilIdentityMagic.TokenFormatUUID:
		return s.uuidIssuer.IssueToken(ctx)
	default:
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("unsupported access token format: %s", s.config.AccessTokenFormat),
		)
	}
}

// IssueIDToken issues an OIDC ID token (always JWS).
func (s *TokenService) IssueIDToken(ctx context.Context, claims map[string]any) (string, error) {
	return s.jwsIssuer.IssueIDToken(ctx, claims)
}

// IssueRefreshToken issues a refresh token (always opaque UUID).
func (s *TokenService) IssueRefreshToken(ctx context.Context) (string, error) {
	return s.uuidIssuer.IssueToken(ctx)
}

// ValidateAccessToken validates an access token and returns its claims.
func (s *TokenService) ValidateAccessToken(ctx context.Context, token string) (map[string]any, error) {
	switch s.config.AccessTokenFormat {
	case cryptoutilIdentityMagic.TokenFormatJWS:
		return s.jwsIssuer.ValidateToken(ctx, token)
	case cryptoutilIdentityMagic.TokenFormatJWE:
		jws, err := s.jweIssuer.DecryptToken(ctx, token)
		if err != nil {
			return nil, err
		}

		return s.jwsIssuer.ValidateToken(ctx, jws)
	case cryptoutilIdentityMagic.TokenFormatUUID:
		if err := s.uuidIssuer.ValidateToken(ctx, token); err != nil {
			return nil, err
		}

		return map[string]any{}, nil
	default:
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("unsupported access token format: %s", s.config.AccessTokenFormat),
		)
	}
}

// ValidateIDToken validates an OIDC ID token and returns its claims.
func (s *TokenService) ValidateIDToken(ctx context.Context, token string) (map[string]any, error) {
	return s.jwsIssuer.ValidateToken(ctx, token)
}

// IsTokenActive checks if a token is currently active (not expired, not before valid).
func (s *TokenService) IsTokenActive(claims map[string]any) bool {
	now := time.Now().UTC().Unix()

	// Check expiration time (exp claim).
	if exp, ok := claims[cryptoutilIdentityMagic.ClaimExp].(float64); ok {
		if int64(exp) < now {
			return false
		}
	}

	// Check not before time (nbf claim).
	if nbf, ok := claims[cryptoutilIdentityMagic.ClaimNbf].(float64); ok {
		if int64(nbf) > now {
			return false
		}
	}

	return true
}

// IntrospectToken introspects a token and returns metadata.
func (s *TokenService) IntrospectToken(ctx context.Context, token string) (*TokenMetadata, error) {
	// Validate token.
	claims, err := s.ValidateAccessToken(ctx, token)
	if err != nil {
		return &TokenMetadata{
			Active: false,
		}, nil
	}

	// Extract metadata.
	metadata := &TokenMetadata{
		Active: true,
		Claims: claims,
	}

	// Add expiration time.
	if exp, ok := claims["exp"].(float64); ok {
		expiresAt := time.Unix(int64(exp), 0)
		metadata.ExpiresAt = &expiresAt
	}

	return metadata, nil
}

// TokenMetadata represents token introspection metadata.
type TokenMetadata struct {
	Active    bool           `json:"active"`
	TokenType string         `json:"token_type,omitempty"`
	Claims    map[string]any `json:"claims,omitempty"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
}

// IssueUserInfoJWT issues a signed JWT containing userinfo claims.
// This fulfills the OAuth 2.1 requirement for JWT-signed userinfo responses.
// The JWT includes iss, aud, iat, and the userinfo claims (sub, profile, email, etc.).
func (s *TokenService) IssueUserInfoJWT(ctx context.Context, clientID string, claims map[string]any) (string, error) {
	// Ensure required claims are present.
	if _, ok := claims[cryptoutilIdentityMagic.ClaimSub].(string); !ok {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("missing required claim: %s", cryptoutilIdentityMagic.ClaimSub),
		)
	}

	// Add audience claim (client_id that requested the userinfo).
	claims[cryptoutilIdentityMagic.ClaimAud] = clientID

	// Issue as ID token (same signing mechanism).
	return s.jwsIssuer.IssueIDToken(ctx, claims)
}

// GetPublicKeys returns the public keys for JWT signature verification.
func (s *TokenService) GetPublicKeys() []map[string]any {
	if s.jwsIssuer == nil || s.jwsIssuer.keyRotationMgr == nil {
		return []map[string]any{}
	}

	return s.jwsIssuer.keyRotationMgr.GetPublicKeys()
}

// BuildTokenDomain builds a domain Token entity from issuance parameters.
func BuildTokenDomain(
	tokenType cryptoutilIdentityDomain.TokenType,
	tokenFormat cryptoutilIdentityDomain.TokenFormat,
	tokenValue string,
	clientID, userID googleUuid.UUID,
	scopes []string,
	expiresAt time.Time,
) *cryptoutilIdentityDomain.Token {
	return &cryptoutilIdentityDomain.Token{
		TokenType:   tokenType,
		TokenFormat: tokenFormat,
		TokenValue:  tokenValue,
		ClientID:    clientID,
		UserID:      cryptoutilIdentityDomain.NewNullableUUID(&userID),
		Scopes:      scopes,
		ExpiresAt:   expiresAt,
		Revoked:     false,
		IssuedAt:    time.Now().UTC(),
	}
}
