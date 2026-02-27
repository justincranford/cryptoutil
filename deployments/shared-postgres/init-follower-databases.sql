-- PostgreSQL Follower Initialization Script
-- Creates grouped databases for OLAP analytics with schemas
-- Read-only replica with logical replication from leader

-- Suite-level: 1 database with 10 schemas (all services across all products)
CREATE DATABASE cryptoutil OWNER cryptoutil_admin;
\c cryptoutil
CREATE SCHEMA pki_ca;
CREATE SCHEMA jose_ja;
CREATE SCHEMA sm_im;
CREATE SCHEMA sm_kms;
CREATE SCHEMA identity_authz;
CREATE SCHEMA identity_idp;
CREATE SCHEMA identity_rs;
CREATE SCHEMA identity_rp;
CREATE SCHEMA identity_spa;
CREATE SCHEMA skeleton_template;

-- Product-level: 5 product databases with schemas grouped by product
CREATE DATABASE pki OWNER cryptoutil_admin;
\c pki
CREATE SCHEMA ca;

CREATE DATABASE jose OWNER cryptoutil_admin;
\c jose
CREATE SCHEMA ja;

CREATE DATABASE sm OWNER cryptoutil_admin;
\c sm
CREATE SCHEMA im;
CREATE SCHEMA kms;

CREATE DATABASE identity OWNER cryptoutil_admin;
\c identity
CREATE SCHEMA authz;
CREATE SCHEMA idp;
CREATE SCHEMA rs;
CREATE SCHEMA rp;
CREATE SCHEMA spa;

CREATE DATABASE skeleton OWNER cryptoutil_admin;
\c skeleton
CREATE SCHEMA template;

-- Service-level: 10 databases with 1:1 schema mapping
CREATE DATABASE "pki-ca" OWNER cryptoutil_admin;
\c "pki-ca"
CREATE SCHEMA ca;

CREATE DATABASE "jose-ja" OWNER cryptoutil_admin;
\c "jose-ja"
CREATE SCHEMA ja;

CREATE DATABASE "sm-im" OWNER cryptoutil_admin;
\c "sm-im"
CREATE SCHEMA im;

CREATE DATABASE "sm-kms" OWNER cryptoutil_admin;
\c "sm-kms"
CREATE SCHEMA kms;

CREATE DATABASE "identity-authz" OWNER cryptoutil_admin;
\c "identity-authz"
CREATE SCHEMA authz;

CREATE DATABASE "identity-idp" OWNER cryptoutil_admin;
\c "identity-idp"
CREATE SCHEMA idp;

CREATE DATABASE "identity-rs" OWNER cryptoutil_admin;
\c "identity-rs"
CREATE SCHEMA rs;

CREATE DATABASE "identity-rp" OWNER cryptoutil_admin;
\c "identity-rp"
CREATE SCHEMA rp;

CREATE DATABASE "identity-spa" OWNER cryptoutil_admin;
\c "identity-spa"
CREATE SCHEMA spa;

CREATE DATABASE "skeleton-template" OWNER cryptoutil_admin;
\c "skeleton-template"
CREATE SCHEMA template;

-- Note: Logical replication subscriptions will be created by setup-logical-replication.sh
-- after tables are created in the leader databases
