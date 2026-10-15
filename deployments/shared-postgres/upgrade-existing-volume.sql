-- Safe, repeatable upgrade for an already-initialized PostgreSQL leader volume.
-- Run with psql connected to the "postgres" database as the PostgreSQL admin user.
-- This script creates only missing PRODUCT/SUITE PS-ID databases and publications.
-- It never drops, truncates, or changes data in existing databases, including legacy *_database sources.
\set ON_ERROR_STOP on

-- CREATE DATABASE cannot run inside a transaction. \gexec executes only missing databases.
SELECT FORMAT('CREATE DATABASE %I', db_name)
FROM (
    VALUES
    ('productdeployment-sm-kms'),
    ('productdeployment-identity-authz'),
    ('productdeployment-identity-idp'),
    ('productdeployment-identity-rs'),
    ('productdeployment-identity-rp'),
    ('productdeployment-identity-spa'),
    ('productdeployment-skeleton-template'),
    ('suitedeployment-sm-kms'),
    ('suitedeployment-identity-authz'),
    ('suitedeployment-identity-idp'),
    ('suitedeployment-identity-rs'),
    ('suitedeployment-identity-rp'),
    ('suitedeployment-identity-spa'),
    ('suitedeployment-skeleton-template')
) AS requested_databases (db_name)
WHERE
    NOT EXISTS (
        SELECT 1
        FROM pg_database AS existing_database
        WHERE existing_database.datname = requested_databases.db_name
    )
\gexec

\connect "productdeployment-sm-kms"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_sm_kms_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_sm_kms_pub'
    )
\gexec

\connect "productdeployment-identity-authz"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_identity_authz_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_identity_authz_pub'
    )
\gexec

\connect "productdeployment-identity-idp"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_identity_idp_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_identity_idp_pub'
    )
\gexec

\connect "productdeployment-identity-rs"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_identity_rs_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_identity_rs_pub'
    )
\gexec

\connect "productdeployment-identity-rp"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_identity_rp_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_identity_rp_pub'
    )
\gexec

\connect "productdeployment-identity-spa"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_identity_spa_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_identity_spa_pub'
    )
\gexec

\connect "productdeployment-skeleton-template"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_skeleton_template_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_skeleton_template_pub'
    )
\gexec

\connect "suitedeployment-sm-kms"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_sm_kms_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_sm_kms_pub'
    )
\gexec

\connect "suitedeployment-identity-authz"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_identity_authz_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_identity_authz_pub'
    )
\gexec

\connect "suitedeployment-identity-idp"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_identity_idp_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_identity_idp_pub'
    )
\gexec

\connect "suitedeployment-identity-rs"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_identity_rs_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_identity_rs_pub'
    )
\gexec

\connect "suitedeployment-identity-rp"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_identity_rp_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_identity_rp_pub'
    )
\gexec

\connect "suitedeployment-identity-spa"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_identity_spa_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_identity_spa_pub'
    )
\gexec

\connect "suitedeployment-skeleton-template"
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_skeleton_template_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_skeleton_template_pub'
    )
\gexec
