package im

import (
	"testing"

	cipherIMRepository "cryptoutil/internal/apps/cipher/im/repository"
	cryptoutilTemplateServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
)

const (
	// Updated to check for base template tables + cipher-im specific tables:
	// Template tables (1001-1004): browser_session_jwks, service_session_jwks, browser_sessions, service_sessions, 
	//                               barrier_root_keys, barrier_intermediate_keys, barrier_content_keys,
	//                               template_realms, tenants, users, clients, unverified_users, unverified_clients,
	//                               roles, user_roles, client_roles
	// Cipher-IM tables (1005-1006): messages, messages_recipient_jwks, cipher_im_realms
	countTablesQueryPostgres = `
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name IN (
			'browser_session_jwks', 'service_session_jwks', 'browser_sessions', 'service_sessions',
			'barrier_root_keys', 'barrier_intermediate_keys', 'barrier_content_keys',
			'template_realms', 'tenants', 'users', 'clients', 'unverified_users', 'unverified_clients',
			'roles', 'user_roles', 'client_roles',
			'messages', 'messages_recipient_jwks', 'cipher_im_realms'
		)
	`
	countTablesQuerySQLite = `
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' 
		AND name IN (
			'browser_session_jwks', 'service_session_jwks', 'browser_sessions', 'service_sessions',
			'barrier_root_keys', 'barrier_intermediate_keys', 'barrier_content_keys',
			'template_realms', 'tenants', 'users', 'clients', 'unverified_users', 'unverified_clients',
			'roles', 'user_roles', 'client_roles',
			'messages', 'messages_recipient_jwks', 'cipher_im_realms'
		)
	`
	// Template tables (16): browser_session_jwks, service_session_jwks, browser_sessions, service_sessions,
	//                       barrier_root_keys, barrier_intermediate_keys, barrier_content_keys,
	//                       template_realms, tenants, users, clients, unverified_users, unverified_clients,
	//                       roles, user_roles, client_roles
	// Cipher-IM tables (3): messages, messages_recipient_jwks, cipher_im_realms
	// Total: 19 tables
	expectedTableCount = 19
)

// TestInitDatabase_HappyPaths tests successful database initialization for PostgreSQL and SQLite.
func TestInitDatabase_HappyPaths(t *testing.T) {
	// Use merged filesystem to get all migrations (1001-1006).
	cryptoutilTemplateServerTestutil.HelpTest_InitDatabase_HappyPaths(
		t,
		cipherIMRepository.GetMergedMigrationsFS(),
		expectedTableCount,
		countTablesQueryPostgres,
		countTablesQuerySQLite,
	)
}

// TestInitDatabase_SadPaths tests error handling for database initialization failures.
func TestInitDatabase_SadPaths(t *testing.T) {
	// Use merged filesystem to get all migrations (1001-1006).
	cryptoutilTemplateServerTestutil.HelpTest_InitDatabase_SadPaths(t, cipherIMRepository.GetMergedMigrationsFS())
}

