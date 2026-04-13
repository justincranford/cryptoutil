#!/bin/bash
# ARCHITECTURE EXCEPTION: Docker container init script.
# This script runs inside the PostgreSQL Docker container via docker-entrypoint-initdb.d/.
# Shell is required by the PostgreSQL image mechanism (only .sh and .sql files are supported).
# See docs/ENG-HANDBOOK.md Section 14.9 (Scripting Language Policy) for the exception rule.
# Minimize all logic here; prefer .sql files for pure SQL operations.
#
# Setup Logical Replication from Leader to Follower
# Creates subscriptions for each schema in follower databases

set -e

PGPASSWORD=$(cat /run/secrets/postgres-password.secret)  # SC2155: assign first, then export
export PGPASSWORD
LEADER_HOST="postgres-leader"
LEADER_PORT="5432"
LEADER_USER="$(cat /run/secrets/postgres-username.secret)"

echo "Waiting for leader to be ready..."
until pg_isready -h "$LEADER_HOST" -p "$LEADER_PORT" -U "$LEADER_USER" -d postgres; do
    echo "Leader not ready, retrying..."
    sleep 5
done

echo "Setting up logical replication subscriptions..."

# Suite-level subscriptions (cryptoutil database, 10 schemas)
psql -h localhost -U "$LEADER_USER" -d cryptoutil <<EOF
CREATE SUBSCRIPTION suite_pki_ca_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=suitedeployment-pki-ca'
    PUBLICATION suite_pki_ca_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_pki_ca_slot');

CREATE SUBSCRIPTION suite_jose_ja_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=suitedeployment-jose-ja'
    PUBLICATION suite_jose_ja_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_jose_ja_slot');

CREATE SUBSCRIPTION suite_sm_im_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=suitedeployment-sm-im'
    PUBLICATION suite_sm_im_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_sm_im_slot');

CREATE SUBSCRIPTION suite_sm_kms_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=suitedeployment-sm-kms'
    PUBLICATION suite_sm_kms_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_sm_kms_slot');

CREATE SUBSCRIPTION suite_identity_authz_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=suitedeployment-identity-authz'
    PUBLICATION suite_identity_authz_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_identity_authz_slot');

CREATE SUBSCRIPTION suite_identity_idp_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=suitedeployment-identity-idp'
    PUBLICATION suite_identity_idp_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_identity_idp_slot');

CREATE SUBSCRIPTION suite_identity_rs_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=suitedeployment-identity-rs'
    PUBLICATION suite_identity_rs_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_identity_rs_slot');

CREATE SUBSCRIPTION suite_identity_rp_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=suitedeployment-identity-rp'
    PUBLICATION suite_identity_rp_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_identity_rp_slot');

CREATE SUBSCRIPTION suite_identity_spa_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=suitedeployment-identity-spa'
    PUBLICATION suite_identity_spa_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_identity_spa_slot');

CREATE SUBSCRIPTION suite_skeleton_template_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=suitedeployment-skeleton-template'
    PUBLICATION suite_skeleton_template_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_skeleton_template_slot');
EOF

# Product-level subscriptions (5 product databases)
psql -h localhost -U "$LEADER_USER" -d pki <<EOF
CREATE SUBSCRIPTION product_pki_ca_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=productdeployment-pki-ca'
    PUBLICATION product_pki_ca_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_pki_ca_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d jose <<EOF
CREATE SUBSCRIPTION product_jose_ja_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=productdeployment-jose-ja'
    PUBLICATION product_jose_ja_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_jose_ja_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d sm <<EOF
CREATE SUBSCRIPTION product_sm_im_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=productdeployment-sm-im'
    PUBLICATION product_sm_im_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_sm_im_slot');

CREATE SUBSCRIPTION product_sm_kms_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=productdeployment-sm-kms'
    PUBLICATION product_sm_kms_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_sm_kms_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d identity <<EOF
CREATE SUBSCRIPTION product_identity_authz_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=productdeployment-identity-authz'
    PUBLICATION product_identity_authz_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_identity_authz_slot');

CREATE SUBSCRIPTION product_identity_idp_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=productdeployment-identity-idp'
    PUBLICATION product_identity_idp_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_identity_idp_slot');

CREATE SUBSCRIPTION product_identity_rs_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=productdeployment-identity-rs'
    PUBLICATION product_identity_rs_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_identity_rs_slot');

CREATE SUBSCRIPTION product_identity_rp_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=productdeployment-identity-rp'
    PUBLICATION product_identity_rp_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_identity_rp_slot');

CREATE SUBSCRIPTION product_identity_spa_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=productdeployment-identity-spa'
    PUBLICATION product_identity_spa_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_identity_spa_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d skeleton <<EOF
CREATE SUBSCRIPTION product_skeleton_template_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=productdeployment-skeleton-template'
    PUBLICATION product_skeleton_template_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_skeleton_template_slot');
EOF

# Service-level subscriptions (10 service databases)
psql -h localhost -U "$LEADER_USER" -d "pki-ca" <<EOF
CREATE SUBSCRIPTION service_pki_ca_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=servicedeployment-pki-ca'
    PUBLICATION service_pki_ca_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'service_pki_ca_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d "jose-ja" <<EOF
CREATE SUBSCRIPTION service_jose_ja_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=servicedeployment-jose-ja'
    PUBLICATION service_jose_ja_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'service_jose_ja_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d "sm-im" <<EOF
CREATE SUBSCRIPTION service_sm_im_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=servicedeployment-sm-im'
    PUBLICATION service_sm_im_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'service_sm_im_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d "sm-kms" <<EOF
CREATE SUBSCRIPTION service_sm_kms_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=servicedeployment-sm-kms'
    PUBLICATION service_sm_kms_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'service_sm_kms_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d "identity-authz" <<EOF
CREATE SUBSCRIPTION service_identity_authz_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=servicedeployment-identity-authz'
    PUBLICATION service_identity_authz_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'service_identity_authz_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d "identity-idp" <<EOF
CREATE SUBSCRIPTION service_identity_idp_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=servicedeployment-identity-idp'
    PUBLICATION service_identity_idp_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'service_identity_idp_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d "identity-rs" <<EOF
CREATE SUBSCRIPTION service_identity_rs_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=servicedeployment-identity-rs'
    PUBLICATION service_identity_rs_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'service_identity_rs_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d "identity-rp" <<EOF
CREATE SUBSCRIPTION service_identity_rp_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=servicedeployment-identity-rp'
    PUBLICATION service_identity_rp_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'service_identity_rp_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d "identity-spa" <<EOF
CREATE SUBSCRIPTION service_identity_spa_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=servicedeployment-identity-spa'
    PUBLICATION service_identity_spa_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'service_identity_spa_slot');
EOF

psql -h localhost -U "$LEADER_USER" -d "skeleton-template" <<EOF
CREATE SUBSCRIPTION service_skeleton_template_sub
    CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres-password.secret) dbname=servicedeployment-skeleton-template'
    PUBLICATION service_skeleton_template_pub
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'service_skeleton_template_slot');
EOF

echo "Logical replication setup complete"
