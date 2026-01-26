// Copyright (c) 2025 Justin Cranford
//
//

package issuer

import (
	"context"
	"crypto"
	ecdsa "crypto/ecdsa"
	crand "crypto/rand"
	rsa "crypto/rsa"
	sha256 "crypto/sha256"
	"encoding/base64"
	json "encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// JWSIssuer issues JWS (signed) JWT tokens using versioned signing keys.
type JWSIssuer struct {
	issuer           string
	keyRotationMgr   *KeyRotationManager
	defaultAlgorithm string
	accessTokenTTL   time.Duration
	idTokenTTL       time.Duration
	legacySigningKey any    // Deprecated: for backward compatibility.
	legacySigningAlg string // Deprecated: for backward compatibility.
}

// NewJWSIssuer creates a new JWS issuer with the specified key rotation manager.
func NewJWSIssuer(
	issuer string,
	keyRotationMgr *KeyRotationManager,
	defaultAlgorithm string,
	accessTokenTTL time.Duration,
	idTokenTTL time.Duration,
) (*JWSIssuer, error) {
	// Validate issuer.
	if issuer == "" {
		return nil, cryptoutilIdentityAppErr.ErrInvalidConfiguration
	}

	// Validate signing algorithm.
	if defaultAlgorithm == "" {
		return nil, cryptoutilIdentityAppErr.ErrInvalidConfiguration
	}

	// Key rotation manager is optional for backward compatibility.
	// If nil, must use NewJWSIssuerLegacy instead.
	if keyRotationMgr == nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("key rotation manager is required; use NewJWSIssuerLegacy for backward compatibility"),
		)
	}

	return &JWSIssuer{
		issuer:           issuer,
		keyRotationMgr:   keyRotationMgr,
		defaultAlgorithm: defaultAlgorithm,
		accessTokenTTL:   accessTokenTTL,
		idTokenTTL:       idTokenTTL,
	}, nil
}

// NewJWSIssuerLegacy creates a new JWS issuer with a single signing key (deprecated).
func NewJWSIssuerLegacy(
	issuer string,
	signingKey any,
	signingAlg string,
	accessTokenTTL time.Duration,
	idTokenTTL time.Duration,
) (*JWSIssuer, error) {
	// Validate issuer.
	if issuer == "" {
		return nil, cryptoutilIdentityAppErr.ErrInvalidConfiguration
	}

	// Validate signing algorithm.
	if signingAlg == "" {
		return nil, cryptoutilIdentityAppErr.ErrInvalidConfiguration
	}

	// Validate signing key.
	if signingKey == nil {
		return nil, cryptoutilIdentityAppErr.ErrInvalidConfiguration
	}

	return &JWSIssuer{
		issuer:           issuer,
		legacySigningKey: signingKey,
		legacySigningAlg: signingAlg,
		accessTokenTTL:   accessTokenTTL,
		idTokenTTL:       idTokenTTL,
	}, nil
}

// IssueAccessToken issues a JWS access token with the specified claims.
func (i *JWSIssuer) IssueAccessToken(_ context.Context, claims map[string]any) (string, error) {
	// Create token claims.
	tokenClaims := make(map[string]any)
	tokenClaims[cryptoutilIdentityMagic.ClaimIss] = i.issuer
	tokenClaims[cryptoutilIdentityMagic.ClaimIat] = time.Now().UTC().Unix()
	tokenClaims[cryptoutilIdentityMagic.ClaimExp] = time.Now().UTC().Add(i.accessTokenTTL).Unix()
	tokenClaims[cryptoutilIdentityMagic.ClaimJti] = googleUuid.NewString()

	// Add standard claims.
	if sub, ok := claims[cryptoutilIdentityMagic.ClaimSub].(string); ok {
		tokenClaims[cryptoutilIdentityMagic.ClaimSub] = sub
	}

	if aud, ok := claims[cryptoutilIdentityMagic.ClaimAud]; ok {
		tokenClaims[cryptoutilIdentityMagic.ClaimAud] = aud
	}

	if scope, ok := claims[cryptoutilIdentityMagic.ParamScope].(string); ok {
		tokenClaims[cryptoutilIdentityMagic.ParamScope] = scope
	}

	// Add custom claims.
	for key, value := range claims {
		if !isStandardClaim(key) {
			tokenClaims[key] = value
		}
	}

	// Build JWS token (simplified format: base64(header).base64(claims).signature).
	return i.buildJWS(tokenClaims)
}

// IssueIDToken issues a JWS ID token with OIDC claims.
func (i *JWSIssuer) IssueIDToken(_ context.Context, claims map[string]any) (string, error) {
	// Validate required OIDC claims.
	if _, ok := claims[cryptoutilIdentityMagic.ClaimSub].(string); !ok {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("missing required claim: %s", cryptoutilIdentityMagic.ClaimSub),
		)
	}

	if _, ok := claims[cryptoutilIdentityMagic.ClaimAud]; !ok {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("missing required claim: %s", cryptoutilIdentityMagic.ClaimAud),
		)
	}

	// Create token claims.
	tokenClaims := make(map[string]any)
	tokenClaims[cryptoutilIdentityMagic.ClaimIss] = i.issuer
	tokenClaims[cryptoutilIdentityMagic.ClaimIat] = time.Now().UTC().Unix()
	tokenClaims[cryptoutilIdentityMagic.ClaimExp] = time.Now().UTC().Add(i.idTokenTTL).Unix()
	tokenClaims[cryptoutilIdentityMagic.ClaimJti] = googleUuid.NewString()

	// Add all claims (including OIDC profile/email/address/phone claims).
	for key, value := range claims {
		if !isStandardClaim(key) || key == cryptoutilIdentityMagic.ClaimSub || key == cryptoutilIdentityMagic.ClaimAud {
			tokenClaims[key] = value
		}
	}

	// Build JWS token.
	return i.buildJWS(tokenClaims)
}

// ValidateToken validates a JWS token and returns its claims.
func (i *JWSIssuer) ValidateToken(_ context.Context, tokenString string) (map[string]any, error) {
	// Parse JWT parts (header.claims.signature).
	parts := strings.Split(tokenString, ".")
	if len(parts) != cryptoutilIdentityMagic.JWSPartCount {
		return nil, cryptoutilIdentityAppErr.ErrInvalidToken
	}

	// Decode header.
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("failed to decode header: %w", err),
		)
	}

	// Parse header JSON.
	var header map[string]any
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("failed to parse header: %w", err),
		)
	}

	// Get algorithm and key ID from header.
	alg, _ := header["alg"].(string) //nolint:errcheck // Type assertion ok ignored
	kid, _ := header["kid"].(string) //nolint:errcheck // Type assertion ok ignored

	// Decode claims.
	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("failed to decode claims: %w", err),
		)
	}

	// Parse claims JSON.
	var claims map[string]any
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("failed to parse claims: %w", err),
		)
	}

	// Decode signature.
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("failed to decode signature: %w", err),
		)
	}

	// Verify signature using the appropriate key.
	signingInput := parts[0] + "." + parts[1]

	if err := i.verifySignature(signingInput, signature, alg, kid); err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenValidationFailed,
			fmt.Errorf("signature verification failed: %w", err),
		)
	}

	// Validate expiration.
	if exp, ok := claims[cryptoutilIdentityMagic.ClaimExp].(float64); ok {
		if time.Now().UTC().Unix() > int64(exp) {
			return nil, cryptoutilIdentityAppErr.ErrTokenExpired
		}
	}

	return claims, nil
}

// verifySignature verifies the JWT signature using the key with the given key ID.
func (i *JWSIssuer) verifySignature(signingInput string, signature []byte, algorithm, keyID string) error {
	// Get the public key for verification.
	var publicKey any

	if i.legacySigningKey != nil {
		// Legacy mode: extract public key from private key.
		switch key := i.legacySigningKey.(type) {
		case *rsa.PrivateKey:
			publicKey = &key.PublicKey
		case *ecdsa.PrivateKey:
			publicKey = &key.PublicKey
		default:
			publicKey = i.legacySigningKey
		}
	} else if i.keyRotationMgr != nil {
		// Try to find the key by key ID.
		var signingKey *SigningKey

		if keyID != "" {
			key, err := i.keyRotationMgr.GetSigningKeyByID(keyID)
			if err == nil {
				signingKey = key
			}
		}

		// If no key found by ID, try all valid verification keys.
		if signingKey == nil {
			keys := i.keyRotationMgr.GetAllValidVerificationKeys()
			if len(keys) == 0 {
				return fmt.Errorf("no valid verification keys available")
			}
			// Use the first available key for verification.
			signingKey = keys[0]
		}

		// Extract public key from signing key.
		switch key := signingKey.Key.(type) {
		case *rsa.PrivateKey:
			publicKey = &key.PublicKey
		case *ecdsa.PrivateKey:
			publicKey = &key.PublicKey
		default:
			publicKey = signingKey.Key
		}
	} else {
		return fmt.Errorf("no signing key available for verification")
	}

	// Verify the signature.
	return verifyJWTSignature(signingInput, signature, algorithm, publicKey)
}

// verifyJWTSignature verifies the JWT signature with the given algorithm and public key.
func verifyJWTSignature(signingInput string, signature []byte, algorithm string, publicKey any) error {
	// Hash the signing input.
	hash := sha256.Sum256([]byte(signingInput))

	// Verify based on algorithm.
	switch algorithm {
	case cryptoutilIdentityMagic.AlgorithmRS256:
		rsaPubKey, ok := publicKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("expected RSA public key for %s algorithm", algorithm)
		}

		if err := rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hash[:], signature); err != nil {
			return fmt.Errorf("RSA signature verification failed: %w", err)
		}

		return nil

	case cryptoutilIdentityMagic.AlgorithmES256:
		ecPubKey, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return fmt.Errorf("expected ECDSA public key for %s algorithm", algorithm)
		}

		// ECDSA signature for ES256 is r || s, each 32 bytes.
		const es256ComponentSize = 32

		if len(signature) != es256ComponentSize*2 {
			return fmt.Errorf("invalid ECDSA signature length: %d", len(signature))
		}

		r := new(big.Int).SetBytes(signature[:es256ComponentSize])
		s := new(big.Int).SetBytes(signature[es256ComponentSize:])

		if !ecdsa.Verify(ecPubKey, hash[:], r, s) {
			return fmt.Errorf("ECDSA signature verification failed")
		}

		return nil

	default:
		return fmt.Errorf("unsupported verification algorithm: %s", algorithm)
	}
}

// buildJWS builds a JWS token using the active signing key.
func (i *JWSIssuer) buildJWS(claims map[string]any) (string, error) {
	var signingAlg string

	var keyID string

	var signingKey any

	// Get active signing key (or use legacy key).
	if i.legacySigningKey != nil {
		// Legacy mode: use single signing key.
		signingAlg = i.legacySigningAlg
		keyID = "" // No key ID in legacy mode.
		signingKey = i.legacySigningKey
	} else if i.keyRotationMgr != nil {
		activeKey, err := i.keyRotationMgr.GetActiveSigningKey()
		if err != nil {
			return "", cryptoutilIdentityAppErr.WrapError(
				cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
				fmt.Errorf("failed to get active signing key: %w", err),
			)
		}

		signingAlg = activeKey.Algorithm
		keyID = activeKey.KeyID
		signingKey = activeKey.Key
	} else {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("no signing key available: neither legacy key nor key rotation manager configured"),
		)
	}

	// Create JWS header.
	header := map[string]any{
		"alg": signingAlg,
		"typ": "JWT",
	}

	// Add key ID if available.
	if keyID != "" {
		header["kid"] = keyID
	}

	// Encode header.
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to marshal header: %w", err),
		)
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)

	// Encode claims.
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to marshal claims: %w", err),
		)
	}

	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsBytes)

	// Create signing input.
	signingInput := headerEncoded + "." + claimsEncoded

	// Sign the token with the actual private key.
	signature, err := signJWT(signingInput, signingAlg, signingKey)
	if err != nil {
		return "", cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrTokenIssuanceFailed,
			fmt.Errorf("failed to sign token: %w", err),
		)
	}

	return signingInput + "." + signature, nil
}

// signJWT signs the JWT signing input with the given algorithm and private key.
func signJWT(signingInput, algorithm string, privateKey any) (string, error) {
	// Hash the signing input.
	hash := sha256.Sum256([]byte(signingInput))

	// Sign based on algorithm.
	switch algorithm {
	case cryptoutilIdentityMagic.AlgorithmRS256:
		rsaKey, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("expected RSA private key for %s algorithm", algorithm)
		}

		signature, err := rsa.SignPKCS1v15(crand.Reader, rsaKey, crypto.SHA256, hash[:])
		if err != nil {
			return "", fmt.Errorf("RSA signing failed: %w", err)
		}

		return base64.RawURLEncoding.EncodeToString(signature), nil

	case cryptoutilIdentityMagic.AlgorithmES256:
		ecKey, ok := privateKey.(*ecdsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("expected ECDSA private key for %s algorithm", algorithm)
		}

		r, s, err := ecdsa.Sign(crand.Reader, ecKey, hash[:])
		if err != nil {
			return "", fmt.Errorf("ECDSA signing failed: %w", err)
		}

		// ECDSA signature for ES256 is r || s, each 32 bytes.
		const es256ComponentSize = 32

		signature := make([]byte, es256ComponentSize*2)
		rBytes := r.Bytes()
		sBytes := s.Bytes()

		copy(signature[es256ComponentSize-len(rBytes):es256ComponentSize], rBytes)
		copy(signature[es256ComponentSize*2-len(sBytes):], sBytes)

		return base64.RawURLEncoding.EncodeToString(signature), nil

	default:
		return "", fmt.Errorf("unsupported signing algorithm: %s", algorithm)
	}
}

// isStandardClaim checks if a claim is a standard JWT/OIDC claim.
func isStandardClaim(claim string) bool {
	standardClaims := []string{
		cryptoutilIdentityMagic.ClaimIss,
		cryptoutilIdentityMagic.ClaimSub,
		cryptoutilIdentityMagic.ClaimAud,
		cryptoutilIdentityMagic.ClaimExp,
		cryptoutilIdentityMagic.ClaimNbf,
		cryptoutilIdentityMagic.ClaimIat,
		cryptoutilIdentityMagic.ClaimJti,
	}

	for _, std := range standardClaims {
		if claim == std {
			return true
		}
	}

	return false
}
