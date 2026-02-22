// Copyright (c) 2025 Justin Cranford
//
//

// Package testutil provides test utilities and factories for common entities.
package testutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Default test timeout for integration tests (configurable via TestTimeoutOverride).
const (
	DefaultIntegrationTimeout = cryptoutilSharedMagic.TestIntegrationTimeout
)

// TestTimeoutOverride allows tests to configure a custom timeout.
var TestTimeoutOverride time.Duration

// IntegrationTimeout returns the configured integration test timeout.
func IntegrationTimeout() time.Duration {
	if TestTimeoutOverride > 0 {
		return TestTimeoutOverride
	}

	return DefaultIntegrationTimeout
}

// IntegrationContext returns a context with the integration timeout.
func IntegrationContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), IntegrationTimeout())
	t.Cleanup(cancel)

	return ctx
}

// WriteTempFile is a helper function for creating temporary test files.
func WriteTempFile(t *testing.T, tempDir, filename, content string) string {
	t.Helper()

	filePath := filepath.Join(tempDir, filename)
	WriteTestFile(t, filePath, content)

	return filePath
}

// WriteTestFile is a helper function for creating test files with content.
func WriteTestFile(t *testing.T, filePath, content string) {
	t.Helper()

	err := os.WriteFile(filePath, []byte(content), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
}

// ReadTestFile is a helper function for reading test files with content.
func ReadTestFile(t *testing.T, filePath string) []byte {
	t.Helper()

	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	return content
}

// TestID returns a UUIDv7 for test data isolation with optional prefix.
func TestID(prefix string) string {
	id := googleUuid.Must(googleUuid.NewV7())

	if prefix == "" {
		return id.String()
	}

	return prefix + "-" + id.String()
}

// TestUserFactory creates test user data.
type TestUserFactory struct {
	IDPrefix string
}

// TestUser represents test user data.
type TestUser struct {
	ID       string
	Username string
	Email    string
	Password string
	Enabled  bool
}

// NewTestUserFactory creates a new test user factory.
func NewTestUserFactory(prefix string) *TestUserFactory {
	return &TestUserFactory{IDPrefix: prefix}
}

// Create creates a new test user with unique ID.
func (f *TestUserFactory) Create(username string) *TestUser {
	id := TestID(f.IDPrefix)
	// Use last 12 chars of UUID for better uniqueness in short strings.
	suffix := id[len(id)-12:]

	return &TestUser{
		ID:       id,
		Username: username + "-" + suffix,
		Email:    username + "-" + suffix + "@test.example.com",
		Password: "TestPassword123!",
		Enabled:  true,
	}
}

// TestClientFactory creates test OAuth client data.
type TestClientFactory struct {
	IDPrefix string
}

// TestClient represents test OAuth client data.
type TestClient struct {
	ID           string
	ClientID     string
	ClientSecret string
	Name         string
	RedirectURIs []string
	Scopes       []string
	Public       bool
}

// NewTestClientFactory creates a new test client factory.
func NewTestClientFactory(prefix string) *TestClientFactory {
	return &TestClientFactory{IDPrefix: prefix}
}

// CreateConfidential creates a confidential OAuth client.
func (f *TestClientFactory) CreateConfidential(name string) *TestClient {
	id := TestID(f.IDPrefix)
	suffix := id[len(id)-12:]

	return &TestClient{
		ID:           id,
		ClientID:     "client-" + suffix,
		ClientSecret: "secret-" + googleUuid.NewString()[:16],
		Name:         name,
		RedirectURIs: []string{"https://localhost/callback"},
		Scopes:       []string{"openid", "profile", "email"},
		Public:       false,
	}
}

// CreatePublic creates a public OAuth client.
func (f *TestClientFactory) CreatePublic(name string) *TestClient {
	id := TestID(f.IDPrefix)
	suffix := id[len(id)-12:]

	return &TestClient{
		ID:           id,
		ClientID:     "public-" + suffix,
		ClientSecret: "",
		Name:         name,
		RedirectURIs: []string{"https://localhost/callback"},
		Scopes:       []string{"openid", "profile"},
		Public:       true,
	}
}

// TestTenantFactory creates test tenant data.
type TestTenantFactory struct {
	IDPrefix string
}

// TestTenant represents test tenant data.
type TestTenant struct {
	ID          string
	Name        string
	Description string
	RealmID     string
	Enabled     bool
}

// NewTestTenantFactory creates a new test tenant factory.
func NewTestTenantFactory(prefix string) *TestTenantFactory {
	return &TestTenantFactory{IDPrefix: prefix}
}

// Create creates a new test tenant with unique ID.
func (f *TestTenantFactory) Create(name string) *TestTenant {
	// Use UUIDv4 for tenant IDs per Session 3 Q10.
	id := googleUuid.NewString()
	suffix := id[len(id)-12:]

	return &TestTenant{
		ID:          id,
		Name:        name + "-" + suffix,
		Description: "Test tenant: " + name,
		RealmID:     "default",
		Enabled:     true,
	}
}
