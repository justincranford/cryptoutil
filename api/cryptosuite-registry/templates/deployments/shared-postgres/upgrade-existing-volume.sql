-- Safe, repeatable upgrade for an already-initialized PostgreSQL leader volume.
-- Run with psql connected to the "postgres" database as the tier's application database
-- role, with administrative privileges. The role must match the user in postgres-url secrets.
-- This script creates missing PRODUCT/SUITE PS-ID databases and publications, normalizes database
-- and publication owners and database grants, and validates publication definitions. It never
-- changes rows or publication definitions and never modifies legacy *_database sources.
\set ON_ERROR_STOP on

-- CREATE DATABASE cannot run inside a transaction. \gexec executes only missing databases.
SELECT FORMAT('CREATE DATABASE %I OWNER %I', db_name, CURRENT_USER)
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

-- Existing databases may have been created by another tier's database role.
-- Set the active application role as owner and grant database privileges explicitly.
SELECT FORMAT('ALTER DATABASE %I OWNER TO %I', db_name, CURRENT_USER)
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
\gexec

SELECT FORMAT('GRANT ALL PRIVILEGES ON DATABASE %I TO %I', db_name, CURRENT_USER)
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
\gexec

\connect "productdeployment-sm-kms"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_sm_kms_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_sm_kms_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'product_sm_kms_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'product_sm_kms_pub',
        'publication product_sm_kms_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "productdeployment-identity-authz"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_identity_authz_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_identity_authz_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'product_identity_authz_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'product_identity_authz_pub',
        'publication product_identity_authz_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "productdeployment-identity-idp"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_identity_idp_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_identity_idp_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'product_identity_idp_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'product_identity_idp_pub',
        'publication product_identity_idp_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "productdeployment-identity-rs"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_identity_rs_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_identity_rs_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'product_identity_rs_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'product_identity_rs_pub',
        'publication product_identity_rs_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "productdeployment-identity-rp"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_identity_rp_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_identity_rp_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'product_identity_rp_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'product_identity_rp_pub',
        'publication product_identity_rp_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "productdeployment-identity-spa"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_identity_spa_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_identity_spa_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'product_identity_spa_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'product_identity_spa_pub',
        'publication product_identity_spa_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "productdeployment-skeleton-template"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'product_skeleton_template_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'product_skeleton_template_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'product_skeleton_template_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'product_skeleton_template_pub',
        'publication product_skeleton_template_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "suitedeployment-sm-kms"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_sm_kms_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_sm_kms_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'suite_sm_kms_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'suite_sm_kms_pub',
        'publication suite_sm_kms_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "suitedeployment-identity-authz"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_identity_authz_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_identity_authz_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'suite_identity_authz_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'suite_identity_authz_pub',
        'publication suite_identity_authz_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "suitedeployment-identity-idp"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_identity_idp_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_identity_idp_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'suite_identity_idp_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'suite_identity_idp_pub',
        'publication suite_identity_idp_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "suitedeployment-identity-rs"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_identity_rs_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_identity_rs_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'suite_identity_rs_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'suite_identity_rs_pub',
        'publication suite_identity_rs_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "suitedeployment-identity-rp"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_identity_rp_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_identity_rp_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'suite_identity_rp_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'suite_identity_rp_pub',
        'publication suite_identity_rp_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "suitedeployment-identity-spa"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_identity_spa_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_identity_spa_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'suite_identity_spa_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'suite_identity_spa_pub',
        'publication suite_identity_spa_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec

\connect "suitedeployment-skeleton-template"
GRANT ALL PRIVILEGES ON SCHEMA public TO current_user;
SELECT FORMAT('CREATE PUBLICATION %I FOR ALL TABLES', 'suite_skeleton_template_pub')
WHERE
    NOT EXISTS (
        SELECT 1 FROM pg_publication
        WHERE pubname = 'suite_skeleton_template_pub'
    )
\gexec
SELECT FORMAT('ALTER PUBLICATION %I OWNER TO %I', 'suite_skeleton_template_pub', CURRENT_USER)
\gexec
SELECT
    FORMAT(
        'DO $validation$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = %L'
        || ' AND pubowner = CURRENT_USER::regrole AND puballtables AND pubinsert'
        || ' AND pubupdate AND pubdelete AND pubtruncate AND NOT pubviaroot)'
        || ' THEN RAISE EXCEPTION %L; END IF; END $validation$',
        'suite_skeleton_template_pub',
        'publication suite_skeleton_template_pub must publish all tables with all supported DML operations and be owned by the active application role'
    )
\gexec
