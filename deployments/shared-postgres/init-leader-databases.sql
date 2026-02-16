-- PostgreSQL Leader Initialization Script
-- Creates 27 logical databases (9 services × 3 deployment types)
-- OLTP Read-Write databases for microservice isolation

-- Suite-level deployment databases (9)
CREATE DATABASE "suitedeployment-pki-ca" OWNER cryptoutil_admin;
CREATE DATABASE "suitedeployment-jose-ja" OWNER cryptoutil_admin;
CREATE DATABASE "suitedeployment-cipher-im" OWNER cryptoutil_admin;
CREATE DATABASE "suitedeployment-sm-kms" OWNER cryptoutil_admin;
CREATE DATABASE "suitedeployment-identity-authz" OWNER cryptoutil_admin;
CREATE DATABASE "suitedeployment-identity-idp" OWNER cryptoutil_admin;
CREATE DATABASE "suitedeployment-identity-rs" OWNER cryptoutil_admin;
CREATE DATABASE "suitedeployment-identity-rp" OWNER cryptoutil_admin;
CREATE DATABASE "suitedeployment-identity-spa" OWNER cryptoutil_admin;

-- Product-level deployment databases (9)
CREATE DATABASE "productdeployment-pki-ca" OWNER cryptoutil_admin;
CREATE DATABASE "productdeployment-jose-ja" OWNER cryptoutil_admin;
CREATE DATABASE "productdeployment-cipher-im" OWNER cryptoutil_admin;
CREATE DATABASE "productdeployment-sm-kms" OWNER cryptoutil_admin;
CREATE DATABASE "productdeployment-identity-authz" OWNER cryptoutil_admin;
CREATE DATABASE "productdeployment-identity-idp" OWNER cryptoutil_admin;
CREATE DATABASE "productdeployment-identity-rs" OWNER cryptoutil_admin;
CREATE DATABASE "productdeployment-identity-rp" OWNER cryptoutil_admin;
CREATE DATABASE "productdeployment-identity-spa" OWNER cryptoutil_admin;

-- Service-level deployment databases (9)
CREATE DATABASE "servicedeployment-pki-ca" OWNER cryptoutil_admin;
CREATE DATABASE "servicedeployment-jose-ja" OWNER cryptoutil_admin;
CREATE DATABASE "servicedeployment-cipher-im" OWNER cryptoutil_admin;
CREATE DATABASE "servicedeployment-sm-kms" OWNER cryptoutil_admin;
CREATE DATABASE "servicedeployment-identity-authz" OWNER cryptoutil_admin;
CREATE DATABASE "servicedeployment-identity-idp" OWNER cryptoutil_admin;
CREATE DATABASE "servicedeployment-identity-rs" OWNER cryptoutil_admin;
CREATE DATABASE "servicedeployment-identity-rp" OWNER cryptoutil_admin;
CREATE DATABASE "servicedeployment-identity-spa" OWNER cryptoutil_admin;

-- Enable logical replication for all databases
\c "suitedeployment-pki-ca"
ALTER SYSTEM SET wal_level = 'logical';
CREATE PUBLICATION suite_pki_ca_pub FOR ALL TABLES;

\c "suitedeployment-jose-ja"
CREATE PUBLICATION suite_jose_ja_pub FOR ALL TABLES;

\c "suitedeployment-cipher-im"
CREATE PUBLICATION suite_cipher_im_pub FOR ALL TABLES;

\c "suitedeployment-sm-kms"
CREATE PUBLICATION suite_sm_kms_pub FOR ALL TABLES;

\c "suitedeployment-identity-authz"
CREATE PUBLICATION suite_identity_authz_pub FOR ALL TABLES;

\c "suitedeployment-identity-idp"
CREATE PUBLICATION suite_identity_idp_pub FOR ALL TABLES;

\c "suitedeployment-identity-rs"
CREATE PUBLICATION suite_identity_rs_pub FOR ALL TABLES;

\c "suitedeployment-identity-rp"
CREATE PUBLICATION suite_identity_rp_pub FOR ALL TABLES;

\c "suitedeployment-identity-spa"
CREATE PUBLICATION suite_identity_spa_pub FOR ALL TABLES;

\c "productdeployment-pki-ca"
CREATE PUBLICATION product_pki_ca_pub FOR ALL TABLES;

\c "productdeployment-jose-ja"
CREATE PUBLICATION product_jose_ja_pub FOR ALL TABLES;

\c "productdeployment-cipher-im"
CREATE PUBLICATION product_cipher_im_pub FOR ALL TABLES;

\c "productdeployment-sm-kms"
CREATE PUBLICATION product_sm_kms_pub FOR ALL TABLES;

\c "productdeployment-identity-authz"
CREATE PUBLICATION product_identity_authz_pub FOR ALL TABLES;

\c "productdeployment-identity-idp"
CREATE PUBLICATION product_identity_idp_pub FOR ALL TABLES;

\c "productdeployment-identity-rs"
CREATE PUBLICATION product_identity_rs_pub FOR ALL TABLES;

\c "productdeployment-identity-rp"
CREATE PUBLICATION product_identity_rp_pub FOR ALL TABLES;

\c "productdeployment-identity-spa"
CREATE PUBLICATION product_identity_spa_pub FOR ALL TABLES;

\c "servicedeployment-pki-ca"
CREATE PUBLICATION service_pki_ca_pub FOR ALL TABLES;

\c "servicedeployment-jose-ja"
CREATE PUBLICATION service_jose_ja_pub FOR ALL TABLES;

\c "servicedeployment-cipher-im"
CREATE PUBLICATION service_cipher_im_pub FOR ALL TABLES;

\c "servicedeployment-sm-kms"
CREATE PUBLICATION service_sm_kms_pub FOR ALL TABLES;

\c "servicedeployment-identity-authz"
CREATE PUBLICATION service_identity_authz_pub FOR ALL TABLES;

\c "servicedeployment-identity-idp"
CREATE PUBLICATION service_identity_idp_pub FOR ALL TABLES;

\c "servicedeployment-identity-rs"
CREATE PUBLICATION service_identity_rs_pub FOR ALL TABLES;

\c "servicedeployment-identity-rp"
CREATE PUBLICATION service_identity_rp_pub FOR ALL TABLES;

\c "servicedeployment-identity-spa"
CREATE PUBLICATION service_identity_spa_pub FOR ALL TABLES;
