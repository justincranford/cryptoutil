// Copyright (c) 2025-2026 Justin Cranford.
//
//

package apis

import (
	"context"
	"fmt"
	"sync"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
	mu               sync.RWMutex
	requests         map[googleUuid.UUID]*AuthorizationRequest
	codeIdx          map[string]googleUuid.UUID // Index by authorization code.
	requestCodeIndex map[googleUuid.UUID]string // Tracks each request's indexed code.
}

// NewInMemoryAuthorizationRequestStore creates a new in-memory authorization request store.
func NewInMemoryAuthorizationRequestStore() *InMemoryAuthorizationRequestStore {
	store := &InMemoryAuthorizationRequestStore{
		requests:         make(map[googleUuid.UUID]*AuthorizationRequest),
		codeIdx:          make(map[string]googleUuid.UUID),
		requestCodeIndex: make(map[googleUuid.UUID]string),
	}

	// Start cleanup goroutine.
	go store.cleanup()

	return store
}

// Store stores an authorization request.
func (s *InMemoryAuthorizationRequestStore) Store(_ context.Context, request *AuthorizationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.removeCodeIndex(request.RequestID)
	s.requests[request.RequestID] = request
	s.addCodeIndex(request)

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

	if request.Code != code {
		return nil, fmt.Errorf("authorization code not found")
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

	s.removeCodeIndex(request.RequestID)
	s.requests[request.RequestID] = request
	s.addCodeIndex(request)

	return nil
}

// Delete deletes an authorization request.
func (s *InMemoryAuthorizationRequestStore) Delete(_ context.Context, requestID googleUuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.removeCodeIndex(requestID)
	delete(s.requests, requestID)

	return nil
}

func (s *InMemoryAuthorizationRequestStore) addCodeIndex(request *AuthorizationRequest) {
	if request.Code != "" {
		s.codeIdx[request.Code] = request.RequestID
		s.requestCodeIndex[request.RequestID] = request.Code
	}
}

func (s *InMemoryAuthorizationRequestStore) removeCodeIndex(requestID googleUuid.UUID) {
	code, ok := s.requestCodeIndex[requestID]
	if ok && s.codeIdx[code] == requestID {
		delete(s.codeIdx, code)
	}

	delete(s.requestCodeIndex, requestID)
}

// cleanup removes expired authorization requests.
func (s *InMemoryAuthorizationRequestStore) cleanup() {
	ticker := time.NewTicker(cryptoutilSharedMagic.ChallengeCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupExpired(time.Now().UTC())
	}
}

func (s *InMemoryAuthorizationRequestStore) cleanupExpired(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, request := range s.requests {
		if now.After(request.ExpiresAt) {
			s.removeCodeIndex(id)
			delete(s.requests, id)
		}
	}
}
