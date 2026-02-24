// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"fmt"
	"time"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ClientClaims represents the JWT claims for client authentication.
type ClientClaims struct {
	Issuer    string    // iss: Client ID
	Subject   string    // sub: Client ID
	Audience  []string  // aud: Token endpoint URL
	ExpiresAt time.Time // exp: Expiration time
	IssuedAt  time.Time // iat: Issued at time
	JWTID     string    // jti: Unique token identifier
}

// JWTValidator defines the interface for JWT validation.
type JWTValidator interface {
	// ValidateJWT validates a JWT string and returns the parsed token.
	ValidateJWT(ctx context.Context, jwtString string, client *cryptoutilIdentityDomain.Client) (joseJwt.Token, error)

	// ExtractClaims extracts client claims from a validated token.
	ExtractClaims(ctx context.Context, token joseJwt.Token) (*ClientClaims, error)
}

// PrivateKeyJWTValidator validates JWTs signed with a client's private key.
type PrivateKeyJWTValidator struct {
	expectedAudience string
	jtiRepo          cryptoutilIdentityRepository.JTIReplayCacheRepository
}

// NewPrivateKeyJWTValidator creates a new private key JWT validator.
func NewPrivateKeyJWTValidator(tokenEndpointURL string, jtiRepo cryptoutilIdentityRepository.JTIReplayCacheRepository) *PrivateKeyJWTValidator {
	return &PrivateKeyJWTValidator{
		expectedAudience: tokenEndpointURL,
		jtiRepo:          jtiRepo,
	}
}

// ValidateJWT validates a JWT signed with the client's private key.
func (v *PrivateKeyJWTValidator) ValidateJWT(ctx context.Context, jwtString string, client *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
	// Parse client's public key set from JWKs field.
	if client.JWKs == "" {
		return nil, fmt.Errorf("client has no JWK set configured")
	}

	publicKeySet, err := joseJwk.Parse([]byte(client.JWKs))
	if err != nil {
		return nil, fmt.Errorf("failed to parse client JWK set: %w", err)
	}

	// Parse and validate JWT with public key verification.
	// JWX v3 handles signature verification during parsing when WithKeySet is provided.
	token, err := joseJwt.Parse(
		[]byte(jwtString),
		joseJwt.WithKeySet(publicKeySet),
		joseJwt.WithAcceptableSkew(time.Minute), // Allow 1 minute clock skew.
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse and verify JWT: %w", err)
	}

	// Validate standard claims.
	if err := v.validateClaims(ctx, token, client); err != nil {
		return nil, err
	}

	return token, nil
}

// validateClaims validates JWT standard claims.
func (v *PrivateKeyJWTValidator) validateClaims(ctx context.Context, token joseJwt.Token, client *cryptoutilIdentityDomain.Client) error {
	// Extract issuer (returns value and boolean indicating presence).
	iss, hasIssuer := token.Issuer()
	if !hasIssuer || iss != client.ClientID {
		return fmt.Errorf("invalid issuer: expected %s, got %s", client.ClientID, iss)
	}

	// Extract subject.
	sub, hasSubject := token.Subject()
	if !hasSubject || sub != client.ClientID {
		return fmt.Errorf("invalid subject: expected %s, got %s", client.ClientID, sub)
	}

	// Extract and validate audience.
	aud, hasAudience := token.Audience()
	if !hasAudience {
		return fmt.Errorf("missing audience claim")
	}

	audienceValid := false

	for _, a := range aud {
		if a == v.expectedAudience {
			audienceValid = true

			break
		}
	}

	if !audienceValid {
		return fmt.Errorf("invalid audience: expected %s in %v", v.expectedAudience, aud)
	}

	// Extract and validate expiration.
	exp, hasExp := token.Expiration()
	if !hasExp {
		return fmt.Errorf("missing expiration claim")
	}

	if time.Now().UTC().After(exp) {
		return fmt.Errorf("JWT expired at %v", exp)
	}

	// Extract and validate issued at.
	iat, hasIat := token.IssuedAt()
	if !hasIat {
		return fmt.Errorf("missing issued at claim")
	}

	if time.Now().UTC().Before(iat) {
		return fmt.Errorf("JWT issued in the future at %v", iat)
	}

	// Validate assertion lifetime (RFC 7523 Section 3).
	assertionLifetime := exp.Sub(iat)
	if assertionLifetime > cryptoutilSharedMagic.JWTAssertionMaxLifetime {
		return fmt.Errorf("JWT assertion lifetime %v exceeds maximum %v", assertionLifetime, cryptoutilSharedMagic.JWTAssertionMaxLifetime)
	}

	// Validate JTI (JWT ID) claim for replay protection.
	jti, hasJTI := token.JwtID()
	if !hasJTI || jti == "" {
		return fmt.Errorf("missing jti (JWT ID) claim")
	}

	// Store JTI in cache to prevent replay attacks.
	if v.jtiRepo != nil {
		if err := v.jtiRepo.Store(ctx, jti, client.ID, exp); err != nil {
			return fmt.Errorf("JTI replay detected: %w", err)
		}
	}

	return nil
}

// ExtractClaims extracts client claims from a validated token.
func (v *PrivateKeyJWTValidator) ExtractClaims(_ context.Context, token joseJwt.Token) (*ClientClaims, error) {
	iss, _ := token.Issuer()
	sub, _ := token.Subject()
	aud, _ := token.Audience()
	exp, _ := token.Expiration()
	iat, _ := token.IssuedAt()
	jti, _ := token.JwtID()

	return &ClientClaims{
		Issuer:    iss,
		Subject:   sub,
		Audience:  aud,
		ExpiresAt: exp,
		IssuedAt:  iat,
		JWTID:     jti,
	}, nil
}

// ClientSecretJWTValidator validates JWTs signed with a client's secret.
type ClientSecretJWTValidator struct {
	expectedAudience string
	jtiRepo          cryptoutilIdentityRepository.JTIReplayCacheRepository
}

// NewClientSecretJWTValidator creates a new client secret JWT validator.
func NewClientSecretJWTValidator(tokenEndpointURL string, jtiRepo cryptoutilIdentityRepository.JTIReplayCacheRepository) *ClientSecretJWTValidator {
	return &ClientSecretJWTValidator{
		expectedAudience: tokenEndpointURL,
		jtiRepo:          jtiRepo,
	}
}

// ValidateJWT validates a JWT signed with the client's secret (HMAC).
func (v *ClientSecretJWTValidator) ValidateJWT(ctx context.Context, jwtString string, client *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
	// Check client secret is configured.
	if client.ClientSecret == "" {
		return nil, fmt.Errorf("client has no secret configured")
	}

	// Create HMAC key from client secret using joseJwk.Import.
	keyData := []byte(client.ClientSecret)

	key, err := joseJwk.Import(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to import symmetric key: %w", err)
	}

	if err := key.Set(joseJwk.KeyIDKey, "test-hmac-key"); err != nil {
		return nil, fmt.Errorf("failed to set key ID: %w", err)
	}

	if err := key.Set(joseJwk.AlgorithmKey, joseJwa.HS256()); err != nil {
		return nil, fmt.Errorf("failed to set key algorithm: %w", err)
	}

	// Create key set with single key.
	keySet := joseJwk.NewSet()
	if err := keySet.AddKey(key); err != nil {
		return nil, fmt.Errorf("failed to add key to set: %w", err)
	}

	// Parse and validate JWT with HMAC signature verification.
	token, err := joseJwt.Parse(
		[]byte(jwtString),
		joseJwt.WithKeySet(keySet),
		joseJwt.WithAcceptableSkew(time.Minute), // Allow 1 minute clock skew.
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse and verify JWT: %w", err)
	}

	// Validate standard claims.
	if err := v.validateClaims(ctx, token, client); err != nil {
		return nil, err
	}

	return token, nil
}

// validateClaims validates JWT standard claims.
func (v *ClientSecretJWTValidator) validateClaims(ctx context.Context, token joseJwt.Token, client *cryptoutilIdentityDomain.Client) error {
	// Extract issuer.
	iss, hasIssuer := token.Issuer()
	if !hasIssuer || iss != client.ClientID {
		return fmt.Errorf("invalid issuer: expected %s, got %s", client.ClientID, iss)
	}

	// Extract subject.
	sub, hasSubject := token.Subject()
	if !hasSubject || sub != client.ClientID {
		return fmt.Errorf("invalid subject: expected %s, got %s", client.ClientID, sub)
	}

	// Extract and validate audience.
	aud, hasAudience := token.Audience()
	if !hasAudience {
		return fmt.Errorf("missing audience claim")
	}

	audienceValid := false

	for _, a := range aud {
		if a == v.expectedAudience {
			audienceValid = true

			break
		}
	}

	if !audienceValid {
		return fmt.Errorf("invalid audience: expected %s in %v", v.expectedAudience, aud)
	}

	// Extract and validate expiration.
	exp, hasExp := token.Expiration()
	if !hasExp {
		return fmt.Errorf("missing expiration claim")
	}

	if time.Now().UTC().After(exp) {
		return fmt.Errorf("JWT expired at %v", exp)
	}

	// Extract and validate issued at.
	iat, hasIat := token.IssuedAt()
	if !hasIat {
		return fmt.Errorf("missing issued at claim")
	}

	if time.Now().UTC().Before(iat) {
		return fmt.Errorf("JWT issued in the future at %v", iat)
	}

	// Validate assertion lifetime (RFC 7523 Section 3): exp - iat should not exceed maximum.
	assertionLifetime := exp.Sub(iat)
	if assertionLifetime > cryptoutilSharedMagic.JWTAssertionMaxLifetime {
		return fmt.Errorf("JWT assertion lifetime %v exceeds maximum %v", assertionLifetime, cryptoutilSharedMagic.JWTAssertionMaxLifetime)
	}

	// Extract and validate jti (JWT ID) for replay protection.
	jti, hasJTI := token.JwtID()
	if !hasJTI || jti == "" {
		return fmt.Errorf("missing jti (JWT ID) claim")
	}

	// Check JTI replay cache.
	if v.jtiRepo != nil {
		// Store JTI with expiration time from token. If already exists, this is a replay attack.
		if err := v.jtiRepo.Store(ctx, jti, client.ID, exp); err != nil {
			return fmt.Errorf("JTI replay detected: %w", err)
		}
	}

	return nil
}

// ExtractClaims extracts client claims from a validated token.
func (v *ClientSecretJWTValidator) ExtractClaims(_ context.Context, token joseJwt.Token) (*ClientClaims, error) {
	iss, _ := token.Issuer()
	sub, _ := token.Subject()
	aud, _ := token.Audience()
	exp, _ := token.Expiration()
	iat, _ := token.IssuedAt()
	jti, _ := token.JwtID()

	return &ClientClaims{
		Issuer:    iss,
		Subject:   sub,
		Audience:  aud,
		ExpiresAt: exp,
		IssuedAt:  iat,
		JWTID:     jti,
	}, nil
}
