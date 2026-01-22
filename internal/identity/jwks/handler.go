// Copyright (c) 2025 Justin Cranford

// Package jwks provides JSON Web Key Set (JWKS) endpoint handlers.
package jwks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilAppErr "cryptoutil/internal/shared/apperr"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// Handler provides JWKS endpoint for exposing public signing keys.
type Handler struct {
	logger  *slog.Logger
	keyRepo cryptoutilIdentityRepository.KeyRepository
}

// NewHandler creates a new JWKS handler instance.
func NewHandler(logger *slog.Logger, keyRepo cryptoutilIdentityRepository.KeyRepository) (*Handler, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil: %w", cryptoutilAppErr.ErrCantBeNil)
	}

	if keyRepo == nil {
		return nil, fmt.Errorf("keyRepo cannot be nil: %w", cryptoutilAppErr.ErrCantBeNil)
	}

	return &Handler{
		logger:  logger,
		keyRepo: keyRepo,
	}, nil
}

// ServeHTTP handles JWKS requests (GET /.well-known/jwks.json).
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only allow GET requests.
	if r.Method != http.MethodGet {
		h.logger.WarnContext(ctx, "JWKS endpoint only supports GET", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	// Get public signing keys from repository.
	jwkSet, err := h.getPublicSigningKeys(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get public signing keys", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	// Marshal JWKS to JSON.
	jwksBytes, err := json.Marshal(jwkSet)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to marshal JWKS", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	// Write response.
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour.
	w.WriteHeader(http.StatusOK)

	if _, writeErr := w.Write(jwksBytes); writeErr != nil {
		h.logger.ErrorContext(ctx, "Failed to write JWKS response", "error", writeErr)
	}
}

// getPublicSigningKeys retrieves all active public signing keys.
func (h *Handler) getPublicSigningKeys(ctx context.Context) (joseJwk.Set, error) {
	// Get all active public signing keys.
	keys, err := h.keyRepo.FindByUsage(ctx, cryptoutilIdentityMagic.KeyUsageSigning, true)
	if err != nil {
		return nil, fmt.Errorf("failed to find active signing keys: %w", err)
	}

	// Create JWK Set.
	jwkSet := joseJwk.NewSet()

	// Add each public key to the set.
	for _, key := range keys {
		if key.PublicKey == "" {
			continue // Skip symmetric keys (no public key).
		}

		// Parse public key JWK.
		publicJWK, parseErr := joseJwk.ParseKey([]byte(key.PublicKey))
		if parseErr != nil {
			h.logger.WarnContext(ctx, "Skipping invalid public key", "key_id", key.ID, "error", parseErr)

			continue
		}

		// Add to set.
		if addErr := jwkSet.AddKey(publicJWK); addErr != nil {
			h.logger.WarnContext(ctx, "Failed to add key to JWKS", "key_id", key.ID, "error", addErr)

			continue
		}
	}

	return jwkSet, nil
}
