// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package magic consolidates all project-wide named constants.
package magic

// Cat 3: PS-ID HTTPS server entity cert Common Names.
//
// Format: "public-https-server-entity-{PS-ID}-{variant}"
//
// Used in E2E tests to assert the correct CN on the public HTTPS server cert
// presented during TLS handshake. Must be in the magic package to satisfy the
// literal-use linter.
//
// Note: sm-kms Cat 3 CNs are defined in magic_otel_e2e.go (AppSMKMS*ServerCertCN).
const (
	// --- pki-ca ---.

	// AppPKICASQLite1ServerCertCN is the Cat 3 server cert CN for pki-ca-app-sqlite-1.
	AppPKICASQLite1ServerCertCN = "public-https-server-entity-pki-ca-sqlite-1"

	// AppPKICASQLite2ServerCertCN is the Cat 3 server cert CN for pki-ca-app-sqlite-2.
	AppPKICASQLite2ServerCertCN = "public-https-server-entity-pki-ca-sqlite-2"

	// AppPKICAPostgres1ServerCertCN is the Cat 3 server cert CN for pki-ca-app-postgres-1.
	AppPKICAPostgres1ServerCertCN = "public-https-server-entity-pki-ca-postgres-1"

	// AppPKICAPostgres2ServerCertCN is the Cat 3 server cert CN for pki-ca-app-postgres-2.
	AppPKICAPostgres2ServerCertCN = "public-https-server-entity-pki-ca-postgres-2"

	// --- identity-authz ---.

	// AppIdentityAuthzSQLite1ServerCertCN is the Cat 3 server cert CN for
	// identity-authz-app-sqlite-1.
	AppIdentityAuthzSQLite1ServerCertCN = "public-https-server-entity-identity-authz-sqlite-1"

	// AppIdentityAuthzSQLite2ServerCertCN is the Cat 3 server cert CN for
	// identity-authz-app-sqlite-2.
	AppIdentityAuthzSQLite2ServerCertCN = "public-https-server-entity-identity-authz-sqlite-2"

	// AppIdentityAuthzPostgres1ServerCertCN is the Cat 3 server cert CN for
	// identity-authz-app-postgres-1.
	AppIdentityAuthzPostgres1ServerCertCN = "public-https-server-entity-identity-authz-postgres-1"

	// AppIdentityAuthzPostgres2ServerCertCN is the Cat 3 server cert CN for
	// identity-authz-app-postgres-2.
	AppIdentityAuthzPostgres2ServerCertCN = "public-https-server-entity-identity-authz-postgres-2"

	// --- identity-idp ---.

	// AppIdentityIDPSQLite1ServerCertCN is the Cat 3 server cert CN for
	// identity-idp-app-sqlite-1.
	AppIdentityIDPSQLite1ServerCertCN = "public-https-server-entity-identity-idp-sqlite-1"

	// AppIdentityIDPSQLite2ServerCertCN is the Cat 3 server cert CN for
	// identity-idp-app-sqlite-2.
	AppIdentityIDPSQLite2ServerCertCN = "public-https-server-entity-identity-idp-sqlite-2"

	// AppIdentityIDPPostgres1ServerCertCN is the Cat 3 server cert CN for
	// identity-idp-app-postgres-1.
	AppIdentityIDPPostgres1ServerCertCN = "public-https-server-entity-identity-idp-postgres-1"

	// AppIdentityIDPPostgres2ServerCertCN is the Cat 3 server cert CN for
	// identity-idp-app-postgres-2.
	AppIdentityIDPPostgres2ServerCertCN = "public-https-server-entity-identity-idp-postgres-2"

	// --- identity-rs ---.

	// AppIdentityRSSQLite1ServerCertCN is the Cat 3 server cert CN for
	// identity-rs-app-sqlite-1.
	AppIdentityRSSQLite1ServerCertCN = "public-https-server-entity-identity-rs-sqlite-1"

	// AppIdentityRSSQLite2ServerCertCN is the Cat 3 server cert CN for
	// identity-rs-app-sqlite-2.
	AppIdentityRSSQLite2ServerCertCN = "public-https-server-entity-identity-rs-sqlite-2"

	// AppIdentityRSPostgres1ServerCertCN is the Cat 3 server cert CN for
	// identity-rs-app-postgres-1.
	AppIdentityRSPostgres1ServerCertCN = "public-https-server-entity-identity-rs-postgres-1"

	// AppIdentityRSPostgres2ServerCertCN is the Cat 3 server cert CN for
	// identity-rs-app-postgres-2.
	AppIdentityRSPostgres2ServerCertCN = "public-https-server-entity-identity-rs-postgres-2"

	// --- identity-rp ---.

	// AppIdentityRPSQLite1ServerCertCN is the Cat 3 server cert CN for
	// identity-rp-app-sqlite-1.
	AppIdentityRPSQLite1ServerCertCN = "public-https-server-entity-identity-rp-sqlite-1"

	// AppIdentityRPSQLite2ServerCertCN is the Cat 3 server cert CN for
	// identity-rp-app-sqlite-2.
	AppIdentityRPSQLite2ServerCertCN = "public-https-server-entity-identity-rp-sqlite-2"

	// AppIdentityRPPostgres1ServerCertCN is the Cat 3 server cert CN for
	// identity-rp-app-postgres-1.
	AppIdentityRPPostgres1ServerCertCN = "public-https-server-entity-identity-rp-postgres-1"

	// AppIdentityRPPostgres2ServerCertCN is the Cat 3 server cert CN for
	// identity-rp-app-postgres-2.
	AppIdentityRPPostgres2ServerCertCN = "public-https-server-entity-identity-rp-postgres-2"

	// --- identity-spa ---.

	// AppIdentitySPASQLite1ServerCertCN is the Cat 3 server cert CN for
	// identity-spa-app-sqlite-1.
	AppIdentitySPASQLite1ServerCertCN = "public-https-server-entity-identity-spa-sqlite-1"

	// AppIdentitySPASQLite2ServerCertCN is the Cat 3 server cert CN for
	// identity-spa-app-sqlite-2.
	AppIdentitySPASQLite2ServerCertCN = "public-https-server-entity-identity-spa-sqlite-2"

	// AppIdentitySPAPostgres1ServerCertCN is the Cat 3 server cert CN for
	// identity-spa-app-postgres-1.
	AppIdentitySPAPostgres1ServerCertCN = "public-https-server-entity-identity-spa-postgres-1"

	// AppIdentitySPAPostgres2ServerCertCN is the Cat 3 server cert CN for
	// identity-spa-app-postgres-2.
	AppIdentitySPAPostgres2ServerCertCN = "public-https-server-entity-identity-spa-postgres-2"

	// --- skeleton-template ---.

	// AppSkeletonTemplateSQLite1ServerCertCN is the Cat 3 server cert CN for
	// skeleton-template-app-sqlite-1.
	AppSkeletonTemplateSQLite1ServerCertCN = "public-https-server-entity-skeleton-template-sqlite-1"

	// AppSkeletonTemplateSQLite2ServerCertCN is the Cat 3 server cert CN for
	// skeleton-template-app-sqlite-2.
	AppSkeletonTemplateSQLite2ServerCertCN = "public-https-server-entity-skeleton-template-sqlite-2"

	// AppSkeletonTemplatePostgres1ServerCertCN is the Cat 3 server cert CN for
	// skeleton-template-app-postgres-1.
	AppSkeletonTemplatePostgres1ServerCertCN = "public-https-server-entity-skeleton-template-postgres-1"

	// AppSkeletonTemplatePostgres2ServerCertCN is the Cat 3 server cert CN for
	// skeleton-template-app-postgres-2.
	AppSkeletonTemplatePostgres2ServerCertCN = "public-https-server-entity-skeleton-template-postgres-2"
)
