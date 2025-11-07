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
		UserID:      &userID,
		Scopes:      scopes,
		ExpiresAt:   expiresAt,
		Revoked:     false,
		IssuedAt:    time.Now(),
	}
}
