// Copyright (c) 2025 Justin Cranford
//
//

package repository_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// TestBrowserSessionJWK_TableName tests BrowserSessionJWK table name.
func TestBrowserSessionJWK_TableName(t *testing.T) {
	t.Parallel()

	jwk := cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK{}
	require.Equal(t, "browser_session_jwks", jwk.TableName())
}

// TestServiceSessionJWK_TableName tests ServiceSessionJWK table name.
func TestServiceSessionJWK_TableName(t *testing.T) {
	t.Parallel()

	jwk := cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK{}
	require.Equal(t, "service_session_jwks", jwk.TableName())
}

// TestBrowserSession_TableName tests BrowserSession table name.
func TestBrowserSession_TableName(t *testing.T) {
	t.Parallel()

	session := cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}
	require.Equal(t, "browser_sessions", session.TableName())
}

// TestServiceSession_TableName tests ServiceSession table name.
func TestServiceSession_TableName(t *testing.T) {
	t.Parallel()

	session := cryptoutilAppsTemplateServiceServerRepository.ServiceSession{}
	require.Equal(t, "service_sessions", session.TableName())
}

// TestSession_IsExpired tests IsExpired method.
func TestSession_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		expiration time.Time
		expected   bool
	}{
		{"Expired session", time.Now().UTC().Add(-1 * time.Hour), true},
		{"Not expired session", time.Now().UTC().Add(1 * time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			browserSession := &cryptoutilAppsTemplateServiceServerRepository.BrowserSession{
				Session: cryptoutilAppsTemplateServiceServerRepository.Session{Expiration: tt.expiration},
			}
			require.Equal(t, tt.expected, browserSession.IsExpired())

			serviceSession := &cryptoutilAppsTemplateServiceServerRepository.ServiceSession{
				Session: cryptoutilAppsTemplateServiceServerRepository.Session{Expiration: tt.expiration},
			}
			require.Equal(t, tt.expected, serviceSession.IsExpired())
		})
	}
}

// TestSession_UpdateLastActivity tests UpdateLastActivity method.
func TestSession_UpdateLastActivity(t *testing.T) {
	t.Parallel()

	browserSession := &cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}
	require.Nil(t, browserSession.LastActivity)

	browserSession.UpdateLastActivity()
	require.NotNil(t, browserSession.LastActivity)
	require.True(t, time.Since(*browserSession.LastActivity) < time.Second)

	serviceSession := &cryptoutilAppsTemplateServiceServerRepository.ServiceSession{}
	require.Nil(t, serviceSession.LastActivity)

	serviceSession.UpdateLastActivity()
	require.NotNil(t, serviceSession.LastActivity)
	require.True(t, time.Since(*serviceSession.LastActivity) < time.Second)
}
