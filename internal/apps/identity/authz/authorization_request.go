// Copyright (c) 2025 Justin Cranford
//
//

package authz

import (
	"context"
	"fmt"
	"sync"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// AuthorizationRequest represents a pending OAuth 2.1 authorization request.
type AuthorizationRequest struct {
	// Request identifier.
	RequestID googleUuid.UUID

	// Client information.
	ClientID    string
	RedirectURI string

	// Request parameters.
	ResponseType string
	Scope        string
	State        string

	// PKCE parameters.
	CodeChallenge       string
	CodeChallengeMethod string

	// User information (populated after authentication).
	UserID *googleUuid.UUID

	// Authorization code (generated after consent).
	Code string

	// Request metadata.
	CreatedAt time.Time
	ExpiresAt time.Time

	// Consent status.
	ConsentGranted bool
}

// AuthorizationRequestStore manages authorization requests.
type AuthorizationRequestStore interface {
	// Store stores an authorization request.
	Store(ctx context.Context, request *AuthorizationRequest) error

	// GetByRequestID retrieves an authorization request by ID.
	GetByRequestID(ctx context.Context, requestID googleUuid.UUID) (*AuthorizationRequest, error)

	// GetByCode retrieves an authorization request by authorization code.
	GetByCode(ctx context.Context, code string) (*AuthorizationRequest, error)

	// Update updates an authorization request.
	Update(ctx context.Context, request *AuthorizationRequest) error

	// Delete deletes an authorization request.
	Delete(ctx context.Context, requestID googleUuid.UUID) error
}

// InMemoryAuthorizationRequestStore implements AuthorizationRequestStore using in-memory storage.
type InMemoryAuthorizationRequestStore struct {
	mu       sync.RWMutex
	requests map[googleUuid.UUID]*AuthorizationRequest
	codeIdx  map[string]googleUuid.UUID // Index by authorization code.
}

// NewInMemoryAuthorizationRequestStore creates a new in-memory authorization request store.
func NewInMemoryAuthorizationRequestStore() *InMemoryAuthorizationRequestStore {
	store := &InMemoryAuthorizationRequestStore{
		requests: make(map[googleUuid.UUID]*AuthorizationRequest),
		codeIdx:  make(map[string]googleUuid.UUID),
	}

	// Start cleanup goroutine.
	go store.cleanup()

	return store
}

// Store stores an authorization request.
func (s *InMemoryAuthorizationRequestStore) Store(_ context.Context, request *AuthorizationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests[request.RequestID] = request

	// If code is set, update code index.
	if request.Code != "" {
		s.codeIdx[request.Code] = request.RequestID
	}

	return nil
}

// GetByRequestID retrieves an authorization request by ID.
func (s *InMemoryAuthorizationRequestStore) GetByRequestID(_ context.Context, requestID googleUuid.UUID) (*AuthorizationRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	request, ok := s.requests[requestID]
	if !ok {
		return nil, fmt.Errorf("authorization request not found")
	}

	if time.Now().UTC().After(request.ExpiresAt) {
		return nil, fmt.Errorf("authorization request expired")
	}

	return request, nil
}

// GetByCode retrieves an authorization request by authorization code.
func (s *InMemoryAuthorizationRequestStore) GetByCode(_ context.Context, code string) (*AuthorizationRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requestID, ok := s.codeIdx[code]
	if !ok {
		return nil, fmt.Errorf("authorization code not found")
	}

	request, ok := s.requests[requestID]
	if !ok {
		return nil, fmt.Errorf("authorization request not found")
	}

	if time.Now().UTC().After(request.ExpiresAt) {
		return nil, fmt.Errorf("authorization code expired")
	}

	return request, nil
}

// Update updates an authorization request.
func (s *InMemoryAuthorizationRequestStore) Update(_ context.Context, request *AuthorizationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.requests[request.RequestID]; !ok {
		return fmt.Errorf("authorization request not found")
	}

	s.requests[request.RequestID] = request

	// Update code index if code changed.
	if request.Code != "" {
		s.codeIdx[request.Code] = request.RequestID
	}

	return nil
}

// Delete deletes an authorization request.
func (s *InMemoryAuthorizationRequestStore) Delete(_ context.Context, requestID googleUuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	request, ok := s.requests[requestID]
	if ok && request.Code != "" {
		delete(s.codeIdx, request.Code)
	}

	delete(s.requests, requestID)

	return nil
}

// cleanup removes expired authorization requests.
func (s *InMemoryAuthorizationRequestStore) cleanup() {
	ticker := time.NewTicker(cryptoutilIdentityMagic.ChallengeCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()

		now := time.Now().UTC()

		for id, request := range s.requests {
			if now.After(request.ExpiresAt) {
				if request.Code != "" {
					delete(s.codeIdx, request.Code)
				}

				delete(s.requests, id)
			}
		}

		s.mu.Unlock()
	}
}
