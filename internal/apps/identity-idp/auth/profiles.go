// Copyright (c) 2025 Justin Cranford
//
//

package auth

import (
	"context"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// AuthenticationProfile defines the interface for authentication methods.
type AuthenticationProfile interface {
	// Name returns the authentication profile name.
	Name() string

	// Authenticate performs user authentication and returns the authenticated user.
	Authenticate(ctx context.Context, credentials map[string]string) (*cryptoutilIdentityDomain.User, error)

	// RequiresMFA indicates whether this profile requires multi-factor authentication.
	RequiresMFA() bool

	// ValidateCredentials validates the credential format without performing authentication.
	ValidateCredentials(credentials map[string]string) error
}

// ProfileRegistry manages authentication profiles.
type ProfileRegistry struct {
	profiles map[string]AuthenticationProfile
}

// NewProfileRegistry creates a new authentication profile registry.
func NewProfileRegistry() *ProfileRegistry {
	return &ProfileRegistry{
		profiles: make(map[string]AuthenticationProfile),
	}
}

// Register registers an authentication profile.
func (r *ProfileRegistry) Register(profile AuthenticationProfile) {
	r.profiles[profile.Name()] = profile
}

// Get retrieves an authentication profile by name.
func (r *ProfileRegistry) Get(name string) (AuthenticationProfile, bool) {
	profile, ok := r.profiles[name]

	return profile, ok
}

// List returns all registered authentication profile names.
func (r *ProfileRegistry) List() []string {
	names := make([]string, 0, len(r.profiles))
	for name := range r.profiles {
		names = append(names, name)
	}

	return names
}
