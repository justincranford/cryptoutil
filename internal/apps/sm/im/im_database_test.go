package im

import (
	"testing"

	cryptoutilAppsSmImRepository "cryptoutil/internal/apps/sm/im/repository"
	cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
)

const (
	// Updated to check for base template tables + sm-im specific tables:
	// Template tables (1001-1004): browser_session_jwks, service_session_jwks, browser_sessions, service_sessions,
	//                               barrier_root_keys, barrier_intermediate_keys, barrier_content_keys,
	//                               template_realms, tenants, users, clients, unverified_users, unverified_clients,
	//                               roles, user_roles, client_roles
	// SM-IM tables (2001+): messages, messages_recipient_jwks.
	countTablesQueryPostgres = `
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name IN (
			'browser_session_jwks', 'service_session_jwks', 'browser_sessions', 'service_sessions',
			'barrier_root_keys', 'barrier_intermediate_keys', 'barrier_content_keys',
			'template_realms', 'tenants', 'users', 'clients', 'unverified_users', 'unverified_clients',
			'roles', 'user_roles', 'client_roles',
			'messages', 'messages_recipient_jwks'
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
			'messages', 'messages_recipient_jwks'
		)
	`
	// Template tables (16): browser_session_jwks, service_session_jwks, browser_sessions, service_sessions,
	//                       barrier_root_keys, barrier_intermediate_keys, barrier_content_keys,
	//                       template_realms, tenants, users, clients, unverified_users, unverified_clients,
	//                       roles, user_roles, client_roles
	// SM-IM tables (2): messages, messages_recipient_jwks
	// Total: 18 tables.
	expectedTableCount = 18
)

// TestInitDatabase_HappyPaths tests successful database initialization for PostgreSQL and SQLite.
//
// ENVIRONMENTAL NOTE: The PostgreSQL_Container subtest requires Docker Desktop to be running.
// On Windows, testcontainers-go may panic with "rootless Docker is not supported on Windows"
// if Docker Desktop is not running or is configured in rootless mode. This is an environmental
// limitation, not a code issue. The SQLite subtest will still pass and provides sufficient coverage.
func TestInitDatabase_HappyPaths(t *testing.T) {
	t.Parallel()
	// Use merged filesystem to get all migrations (1001-1999 template + 2001+ sm-im).
	cryptoutilAppsTemplateServiceServerTestutil.HelpTestInitDatabaseHappyPaths(
		t,
		cryptoutilAppsSmImRepository.GetMergedMigrationsFS(),
		expectedTableCount,
		countTablesQueryPostgres,
		countTablesQuerySQLite,
	)
}

// TestInitDatabase_SadPaths tests error handling for database initialization failures.
func TestInitDatabase_SadPaths(t *testing.T) {
	t.Parallel()
	// Use merged filesystem to get all migrations (1001-1006).
	cryptoutilAppsTemplateServiceServerTestutil.HelpTestInitDatabaseSadPaths(t, cryptoutilAppsSmImRepository.GetMergedMigrationsFS())
}
