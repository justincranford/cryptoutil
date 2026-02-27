// Copyright (c) 2025 Justin Cranford
//
//

package hash

import (
	"fmt"
	"sync"

	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"
)

// ParameterSetRegistry manages versioned PBKDF2 parameter sets.
type ParameterSetRegistry struct {
	mu             sync.RWMutex
	parameterSets  map[string]*cryptoutilSharedCryptoDigests.PBKDF2Params
	defaultVersion string
}

// NewParameterSetRegistry creates a new registry with pre-registered parameter sets.
func NewParameterSetRegistry() *ParameterSetRegistry {
	registry := &ParameterSetRegistry{
		parameterSets:  make(map[string]*cryptoutilSharedCryptoDigests.PBKDF2Params),
		defaultVersion: "1",
	}

	// Register all standard parameter sets.
	registry.registerParameterSet(PBKDF2ParameterSetV1())
	registry.registerParameterSet(PBKDF2ParameterSetV2())
	registry.registerParameterSet(PBKDF2ParameterSetV3())

	return registry
}

// registerParameterSet registers a parameter set (internal use).
func (r *ParameterSetRegistry) registerParameterSet(params *cryptoutilSharedCryptoDigests.PBKDF2Params) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.parameterSets[params.Version] = params
}

// GetParameterSet retrieves a parameter set by version string.
func (r *ParameterSetRegistry) GetParameterSet(version string) (*cryptoutilSharedCryptoDigests.PBKDF2Params, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	params, exists := r.parameterSets[version]
	if !exists {
		return nil, fmt.Errorf("parameter set version %q not found in registry", version)
	}

	return params, nil
}

// GetDefaultParameterSet returns the default parameter set.
func (r *ParameterSetRegistry) GetDefaultParameterSet() *cryptoutilSharedCryptoDigests.PBKDF2Params {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Default version always exists (registered in constructor).
	params, exists := r.parameterSets[r.defaultVersion]
	if !exists {
		// Should never happen - indicates programming error.
		panic(fmt.Sprintf("default parameter set version %q not registered", r.defaultVersion))
	}

	return params
}

// ListVersions returns all registered parameter set versions.
func (r *ParameterSetRegistry) ListVersions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions := make([]string, 0, len(r.parameterSets))
	for version := range r.parameterSets {
		versions = append(versions, version)
	}

	return versions
}

// GetDefaultVersion returns the default version string.
func (r *ParameterSetRegistry) GetDefaultVersion() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.defaultVersion
}

// Global registry instance.
var globalRegistry = NewParameterSetRegistry()

// GetGlobalRegistry returns the global parameter set registry.
func GetGlobalRegistry() *ParameterSetRegistry {
	return globalRegistry
}
