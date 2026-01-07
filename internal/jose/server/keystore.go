// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"fmt"
	"sync"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// StoredKey represents a JWK with metadata stored in the key store.
type StoredKey struct {
	KID        googleUuid.UUID `json:"kid"`
	PrivateJWK joseJwk.Key     `json:"-"`          // Private key (not exposed in API).
	PublicJWK  joseJwk.Key     `json:"public_jwk"` // Public key (exposed in JWKS).
	KeyType    string          `json:"kty"`        // Key type (RSA, EC, OKP, oct).
	Algorithm  string          `json:"alg"`        // Algorithm hint.
	Use        string          `json:"use"`        // Key use (sig, enc).
	CreatedAt  int64           `json:"created_at"` // Unix timestamp.
}

// KeyStore manages JWKs in memory.
type KeyStore struct {
	keys map[string]*StoredKey // Map of KID string to stored key.
	mu   sync.RWMutex
}

// NewKeyStore creates a new in-memory key store.
func NewKeyStore() *KeyStore {
	return &KeyStore{
		keys: make(map[string]*StoredKey),
	}
}

// Store adds a key to the store.
func (ks *KeyStore) Store(key *StoredKey) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()

	kidStr := key.KID.String()
	if _, exists := ks.keys[kidStr]; exists {
		return fmt.Errorf("key with KID %s already exists", kidStr)
	}

	ks.keys[kidStr] = key

	return nil
}

// Get retrieves a key by KID.
func (ks *KeyStore) Get(kid string) (*StoredKey, bool) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	key, exists := ks.keys[kid]

	return key, exists
}

// Delete removes a key by KID.
func (ks *KeyStore) Delete(kid string) bool {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if _, exists := ks.keys[kid]; !exists {
		return false
	}

	delete(ks.keys, kid)

	return true
}

// List returns all stored keys (metadata only, not private keys).
func (ks *KeyStore) List() []*StoredKey {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	keys := make([]*StoredKey, 0, len(ks.keys))
	for _, key := range ks.keys {
		keys = append(keys, key)
	}

	return keys
}

// GetJWKS returns a JWK Set containing all public keys.
func (ks *KeyStore) GetJWKS() joseJwk.Set {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	jwks := joseJwk.NewSet()

	for _, key := range ks.keys {
		if key.PublicJWK != nil {
			_ = jwks.AddKey(key.PublicJWK)
		}
	}

	return jwks
}

// Count returns the number of keys in the store.
func (ks *KeyStore) Count() int {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	return len(ks.keys)
}
