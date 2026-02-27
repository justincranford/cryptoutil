#!/bin/bash
# Setup Logical Replication from Leader to Follower
# Creates subscriptions for each schema in follower databases

set -e

export PGPASSWORD=$(cat /run/secrets/postgres_password.secret)
LEADER_HOST="postgres-leader"
LEADER_PORT="5432"
LEADER_USER="cryptoutil_admin"

echo "Waiting for leader to be ready..."
until pg_isready -h "$LEADER_HOST" -p "$LEADER_PORT" -U "$LEADER_USER" -d postgres; do
  echo "Leader not ready, retrying..."
  sleep 5
done

echo "Setting up logical replication subscriptions..."

# Suite-level subscriptions (cryptoutil database, 9 schemas)
psql -h localhost -U cryptoutil_admin -d cryptoutil <<EOF
CREATE SUBSCRIPTION suite_pki_ca_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=suitedeployment-pki-ca'
  PUBLICATION suite_pki_ca_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_pki_ca_slot');

CREATE SUBSCRIPTION suite_jose_ja_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=suitedeployment-jose-ja'
  PUBLICATION suite_jose_ja_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_jose_ja_slot');

CREATE SUBSCRIPTION suite_sm_im_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=suitedeployment-sm-im'
  PUBLICATION suite_sm_im_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_sm_im_slot');

CREATE SUBSCRIPTION suite_sm_kms_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=suitedeployment-sm-kms'
  PUBLICATION suite_sm_kms_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_sm_kms_slot');

CREATE SUBSCRIPTION suite_identity_authz_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=suitedeployment-identity-authz'
  PUBLICATION suite_identity_authz_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_identity_authz_slot');

CREATE SUBSCRIPTION suite_identity_idp_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=suitedeployment-identity-idp'
  PUBLICATION suite_identity_idp_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_identity_idp_slot');

CREATE SUBSCRIPTION suite_identity_rs_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=suitedeployment-identity-rs'
  PUBLICATION suite_identity_rs_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_identity_rs_slot');

CREATE SUBSCRIPTION suite_identity_rp_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=suitedeployment-identity-rp'
  PUBLICATION suite_identity_rp_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_identity_rp_slot');

CREATE SUBSCRIPTION suite_identity_spa_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=suitedeployment-identity-spa'
  PUBLICATION suite_identity_spa_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'suite_identity_spa_slot');
EOF

# Product-level subscriptions (5 product databases)
psql -h localhost -U cryptoutil_admin -d pki <<EOF
CREATE SUBSCRIPTION product_pki_ca_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=productdeployment-pki-ca'
  PUBLICATION product_pki_ca_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_pki_ca_slot');
EOF

psql -h localhost -U cryptoutil_admin -d jose <<EOF
CREATE SUBSCRIPTION product_jose_ja_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=productdeployment-jose-ja'
  PUBLICATION product_jose_ja_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_jose_ja_slot');
EOF

psql -h localhost -U cryptoutil_admin -d sm <<EOF
CREATE SUBSCRIPTION product_sm_im_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=productdeployment-sm-im'
  PUBLICATION product_sm_im_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_sm_im_slot');
EOF

psql -h localhost -U cryptoutil_admin -d sm <<EOF
CREATE SUBSCRIPTION product_sm_kms_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=productdeployment-sm-kms'
  PUBLICATION product_sm_kms_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_sm_kms_slot');
EOF

psql -h localhost -U cryptoutil_admin -d identity <<EOF
CREATE SUBSCRIPTION product_identity_authz_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=productdeployment-identity-authz'
  PUBLICATION product_identity_authz_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_identity_authz_slot');

CREATE SUBSCRIPTION product_identity_idp_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=productdeployment-identity-idp'
  PUBLICATION product_identity_idp_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_identity_idp_slot');

CREATE SUBSCRIPTION product_identity_rs_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=productdeployment-identity-rs'
  PUBLICATION product_identity_rs_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_identity_rs_slot');

CREATE SUBSCRIPTION product_identity_rp_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=productdeployment-identity-rp'
  PUBLICATION product_identity_rp_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_identity_rp_slot');

CREATE SUBSCRIPTION product_identity_spa_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=productdeployment-identity-spa'
  PUBLICATION product_identity_spa_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'product_identity_spa_slot');
EOF

# Service-level subscriptions (9 service databases)
# Note: Simplified example - expand similarly for all 9 services
psql -h localhost -U cryptoutil_admin -d "pki-ca" <<EOF
CREATE SUBSCRIPTION service_pki_ca_sub
  CONNECTION 'host=$LEADER_HOST port=$LEADER_PORT user=$LEADER_USER password=$(cat /run/secrets/postgres_password.secret) dbname=servicedeployment-pki-ca'
  PUBLICATION service_pki_ca_pub
  WITH (copy_data = true, create_slot = true, enabled = true, slot_name = 'service_pki_ca_slot');
EOF

# Add remaining service subscriptions (jose-ja, sm-im, sm-kms, identity-*, etc.)

echo "Logical replication setup complete"
