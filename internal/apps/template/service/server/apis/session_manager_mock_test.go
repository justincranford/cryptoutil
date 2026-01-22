//go:build !integration

package apis

import (
	"context"
	"errors"

	googleUuid "github.com/google/uuid"

	cryptoutilRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// mockSessionManager implements SessionManagerService interface for unit testing.
type mockSessionManager struct {
	// Configurable function fields for different test scenarios
	issueBrowserFunc    func(ctx context.Context, userID string, tenantID, realmID googleUuid.UUID) (string, error)
	issueServiceFunc    func(ctx context.Context, clientID string, tenantID, realmID googleUuid.UUID) (string, error)
	validateBrowserFunc func(ctx context.Context, token string) (*cryptoutilRepository.BrowserSession, error)
	validateServiceFunc func(ctx context.Context, token string) (*cryptoutilRepository.ServiceSession, error)
}

// IssueBrowserSessionWithTenant issues a browser session token.
func (m *mockSessionManager) IssueBrowserSessionWithTenant(ctx context.Context, userID string, tenantID, realmID googleUuid.UUID) (string, error) {
	if m.issueBrowserFunc != nil {
		return m.issueBrowserFunc(ctx, userID, tenantID, realmID)
	}
	return "mock-browser-token-" + userID, nil
}

// IssueServiceSessionWithTenant issues a service session token.
func (m *mockSessionManager) IssueServiceSessionWithTenant(ctx context.Context, clientID string, tenantID, realmID googleUuid.UUID) (string, error) {
	if m.issueServiceFunc != nil {
		return m.issueServiceFunc(ctx, clientID, tenantID, realmID)
	}
	return "mock-service-token-" + clientID, nil
}

// ValidateBrowserSession validates a browser session token.
func (m *mockSessionManager) ValidateBrowserSession(ctx context.Context, token string) (*cryptoutilRepository.BrowserSession, error) {
	if m.validateBrowserFunc != nil {
		return m.validateBrowserFunc(ctx, token)
	}
	userID := "mock-user-from-token"
	return &cryptoutilRepository.BrowserSession{
		UserID: &userID,
		Session: cryptoutilRepository.Session{
			TenantID: googleUuid.New(),
			RealmID:  googleUuid.New(),
		},
	}, nil
}

// ValidateServiceSession validates a service session token.
func (m *mockSessionManager) ValidateServiceSession(ctx context.Context, token string) (*cryptoutilRepository.ServiceSession, error) {
	if m.validateServiceFunc != nil {
		return m.validateServiceFunc(ctx, token)
	}
	clientID := "mock-client-from-token"
	return &cryptoutilRepository.ServiceSession{
		ClientID: &clientID,
		Session: cryptoutilRepository.Session{
			TenantID: googleUuid.New(),
			RealmID:  googleUuid.New(),
		},
	}, nil
}

// Helper constructors for common test scenarios.

// newMockSessionManagerSuccess creates a mock that returns successful responses.
func newMockSessionManagerSuccess() *mockSessionManager {
	return &mockSessionManager{} // Uses default success behavior
}

// newMockSessionManagerError creates a mock that returns errors for all operations.
func newMockSessionManagerError(err error) *mockSessionManager {
	return &mockSessionManager{
		issueBrowserFunc: func(context.Context, string, googleUuid.UUID, googleUuid.UUID) (string, error) {
			return "", err
		},
		issueServiceFunc: func(context.Context, string, googleUuid.UUID, googleUuid.UUID) (string, error) {
			return "", err
		},
		validateBrowserFunc: func(context.Context, string) (*cryptoutilRepository.BrowserSession, error) {
			return nil, err
		},
		validateServiceFunc: func(context.Context, string) (*cryptoutilRepository.ServiceSession, error) {
			return nil, err
		},
	}
}

// errMockSessionError is a common error for testing error paths.
var errMockSessionError = errors.New("mock session error")
