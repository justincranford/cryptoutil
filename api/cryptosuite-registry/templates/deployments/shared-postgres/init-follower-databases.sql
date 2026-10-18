-- PostgreSQL Follower Initialization Script
-- Creates one subscriber database for each published leader database.
-- Native logical replication requires matching relation names and schemas.

-- Suite-level deployment databases (7).
CREATE DATABASE "suitedeployment-sm-kms";
CREATE DATABASE "suitedeployment-identity-authz";
CREATE DATABASE "suitedeployment-identity-idp";
CREATE DATABASE "suitedeployment-identity-rs";
CREATE DATABASE "suitedeployment-identity-rp";
CREATE DATABASE "suitedeployment-identity-spa";
CREATE DATABASE "suitedeployment-skeleton-template";

-- Product-level deployment databases (7).
CREATE DATABASE "productdeployment-sm-kms";
CREATE DATABASE "productdeployment-identity-authz";
CREATE DATABASE "productdeployment-identity-idp";
CREATE DATABASE "productdeployment-identity-rs";
CREATE DATABASE "productdeployment-identity-rp";
CREATE DATABASE "productdeployment-identity-spa";
CREATE DATABASE "productdeployment-skeleton-template";

-- Service-level deployment databases (7).
CREATE DATABASE "servicedeployment-sm-kms";
CREATE DATABASE "servicedeployment-identity-authz";
CREATE DATABASE "servicedeployment-identity-idp";
CREATE DATABASE "servicedeployment-identity-rs";
CREATE DATABASE "servicedeployment-identity-rp";
CREATE DATABASE "servicedeployment-identity-spa";
CREATE DATABASE "servicedeployment-skeleton-template";
