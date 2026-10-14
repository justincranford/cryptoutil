-- PostgreSQL User Initialization Script
-- Grants all privileges to the admin user (created by POSTGRES_USER_FILE) on all 32 databases.
-- Per-PS-ID service users are registered at application startup, not here.
-- The admin user already exists when this script runs (PostgreSQL entrypoint creates it first).

-- Grant ownership of all databases to the admin user.
-- The admin user already exists (created by POSTGRES_USER_FILE in PostgreSQL entrypoint).
-- These GRANT statements ensure the admin user owns each database for full DDL/DML access.

-- Suite-level databases
GRANT ALL PRIVILEGES ON DATABASE "suitedeployment-sm-kms" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "suitedeployment-identity-authz" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "suitedeployment-identity-idp" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "suitedeployment-identity-rs" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "suitedeployment-identity-rp" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "suitedeployment-identity-spa" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "suitedeployment-skeleton-template" TO current_user;

-- Product-level databases
GRANT ALL PRIVILEGES ON DATABASE "productdeployment-sm-kms" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "productdeployment-identity-authz" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "productdeployment-identity-idp" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "productdeployment-identity-rs" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "productdeployment-identity-rp" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "productdeployment-identity-spa" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "productdeployment-skeleton-template" TO current_user;

-- Service-level databases
GRANT ALL PRIVILEGES ON DATABASE "servicedeployment-sm-kms" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "servicedeployment-identity-authz" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "servicedeployment-identity-idp" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "servicedeployment-identity-rs" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "servicedeployment-identity-rp" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "servicedeployment-identity-spa" TO current_user;
GRANT ALL PRIVILEGES ON DATABASE "servicedeployment-skeleton-template" TO current_user;

-- PS-ID application databases
GRANT ALL PRIVILEGES ON DATABASE sm_kms_database TO current_user;
GRANT ALL PRIVILEGES ON DATABASE identity_authz_database TO current_user;
GRANT ALL PRIVILEGES ON DATABASE identity_idp_database TO current_user;
GRANT ALL PRIVILEGES ON DATABASE identity_rp_database TO current_user;
GRANT ALL PRIVILEGES ON DATABASE identity_rs_database TO current_user;
GRANT ALL PRIVILEGES ON DATABASE identity_spa_database TO current_user;
GRANT ALL PRIVILEGES ON DATABASE skeleton_template_database TO current_user;

-- Product- and suite-tier application databases
GRANT ALL PRIVILEGES ON DATABASE sm_database TO current_user;
GRANT ALL PRIVILEGES ON DATABASE identity_database TO current_user;
GRANT ALL PRIVILEGES ON DATABASE skeleton_database TO current_user;
GRANT ALL PRIVILEGES ON DATABASE cryptoutil_database TO current_user;
