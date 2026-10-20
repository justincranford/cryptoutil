#!/bin/bash
# This script runs after the follower's final PostgreSQL server starts.
# Creating subscriptions during docker-entrypoint-initdb.d runs would launch table-sync workers
# on its temporary server; stopping that server can discard their transient publisher slots.
#
# Prepare subscriber schemas, reconcile retained client-secret columns, and create subscriptions.

set -euo pipefail

PGPASSWORD="$(cat /run/secrets/postgres-password.secret)"
export PGPASSWORD

LEADER_HOST="postgres-leader"
LEADER_PORT="5432"
LEADER_USER="$(cat /run/secrets/postgres-username.secret)"
CERT_SCOPE="${POSTGRES_CERT_SCOPE:?POSTGRES_CERT_SCOPE must identify the active certificate directory}"
REPLICATION_TIER="${POSTGRES_REPLICATION_TIER:?POSTGRES_REPLICATION_TIER must be service, product, or suite}"
REPLICATION_PSIDS="${POSTGRES_REPLICATION_PSIDS:?POSTGRES_REPLICATION_PSIDS must list active PS-IDs}"
MIGRATION_WAIT_INTERVAL_SECONDS=5
MIGRATION_WAIT_ATTEMPTS=180

CERT_ROOT="/mnt/ps-certs-src/${CERT_SCOPE}"
CLIENT_CERT="${CERT_ROOT}/postgres-tls-client-entity-follower-replication/postgres-tls-client-entity-follower-replication.crt"
SOURCE_CLIENT_KEY="${CERT_ROOT}/postgres-tls-client-entity-follower-replication/postgres-tls-client-entity-follower-replication.key"
SERVER_CA="${CERT_ROOT}/postgres-tls-server-issuing-ca/truststore/postgres-tls-server-issuing-ca.crt"
REPLICATION_KEY="${PGDATA:?PGDATA must be set}/postgres-tls-client-entity-follower-replication.key"

for cert_file in "$CLIENT_CERT" "$SOURCE_CLIENT_KEY" "$SERVER_CA"; do
    if [[ ! -r "$cert_file" ]]; then
        echo "Required replication TLS material is not readable: $cert_file" >&2
        exit 1
    fi
done

install -m 0600 "$SOURCE_CLIENT_KEY" "$REPLICATION_KEY"
CLIENT_KEY="$REPLICATION_KEY"

export PGSSLMODE="verify-full"
export PGSSLCERT="${CLIENT_CERT}"
export PGSSLKEY="${CLIENT_KEY}"
export PGSSLROOTCERT="${SERVER_CA}"

echo "Waiting for leader to be ready..."
until pg_isready -h "$LEADER_HOST" -p "$LEADER_PORT" -U "$LEADER_USER" -d postgres; do
    echo "Leader not ready, retrying..."
    sleep 5
done

case "$REPLICATION_TIER" in
    service|product|suite)
        ;;
    *)
        echo "Unsupported POSTGRES_REPLICATION_TIER: $REPLICATION_TIER" >&2
        exit 1
        ;;
esac

IFS=',' read -r -a psids <<< "$REPLICATION_PSIDS"
if [[ ${#psids[@]} -eq 0 ]]; then
    echo "POSTGRES_REPLICATION_PSIDS must contain at least one PS-ID" >&2
    exit 1
fi

echo "Preparing $REPLICATION_TIER logical replication subscriptions..."

for psid in "${psids[@]}"; do
    if [[ ! "$psid" =~ ^[a-z0-9]+(-[a-z0-9]+)*$ ]]; then
        echo "Invalid PS-ID in POSTGRES_REPLICATION_PSIDS: $psid" >&2
        exit 1
    fi

    psid_identifier="${psid//-/_}"
    source_database="${REPLICATION_TIER}deployment-${psid}"
    publication="${REPLICATION_TIER}_${psid_identifier}_pub"
    subscription="${REPLICATION_TIER}_${psid_identifier}_sub"
    slot="${REPLICATION_TIER}_${psid_identifier}_slot"

    # Existing follower volumes bypass docker-entrypoint-initdb.d; create only missing targets.
    echo "Ensuring follower database ${source_database} exists..."
    psql \
        -v ON_ERROR_STOP=1 \
        -v database="$source_database" \
        -U "$LEADER_USER" \
        -d postgres <<'SQL'
SELECT FORMAT('CREATE DATABASE %I OWNER %I', :'database', CURRENT_USER)
WHERE NOT EXISTS (
    SELECT 1 FROM pg_database WHERE datname = :'database'
)
\gexec
SQL

    source_conn="host=${LEADER_HOST} port=${LEADER_PORT} user=${LEADER_USER} password=${PGPASSWORD} dbname=${source_database} sslmode=verify-full sslcert=${CLIENT_CERT} sslkey=${CLIENT_KEY} sslrootcert=${SERVER_CA}"

    echo "Waiting for ${source_database} client-secret hash schema..."
    migration_ready=false
    for ((attempt = 1; attempt <= MIGRATION_WAIT_ATTEMPTS; attempt++)); do
        schema_ready="$(
            psql "$source_conn" \
                -At \
                -v ON_ERROR_STOP=1 \
                -c "SELECT
                    to_regclass('public.clients') IS NOT NULL
                    AND to_regclass('public.unverified_clients') IS NOT NULL
                    AND (
                        SELECT count(*) = 2
                        FROM information_schema.columns
                        WHERE table_schema = 'public'
                            AND table_name IN ('clients', 'unverified_clients')
                            AND column_name = 'client_secret_hash'
                    )
                    AND NOT EXISTS (
                        SELECT 1
                        FROM information_schema.columns
                        WHERE table_schema = 'public'
                            AND table_name IN ('clients', 'unverified_clients')
                            AND column_name = 'client_secret'
                    );" \
                2>/dev/null || true
        )"
        if [[ "$schema_ready" == "t" ]]; then
            migration_ready=true
            break
        fi

        sleep "$MIGRATION_WAIT_INTERVAL_SECONDS"
    done
    if [[ "$migration_ready" != "true" ]]; then
        echo "Timed out waiting for ${source_database} client-secret hash schema" >&2
        exit 1
    fi

    subscriber_schema_ready="$(
        psql \
            -At \
            -v ON_ERROR_STOP=1 \
            -c "SELECT to_regclass('public.clients') IS NOT NULL AND to_regclass('public.unverified_clients') IS NOT NULL;" \
            -U "$LEADER_USER" \
            -d "$source_database"
    )"
    if [[ "$subscriber_schema_ready" != "t" ]]; then
        echo "Copying schema for ${source_database} to the matching follower database..."
        pg_dump \
            --schema-only \
            --no-owner \
            --no-privileges \
            --dbname="$source_conn" |
            psql -v ON_ERROR_STOP=1 -U "$LEADER_USER" -d "$source_database"
    else
        echo "Subscriber schema for ${source_database} is already present."
    fi

    echo "Reconciling follower client-secret hash columns..."
    psql \
        -v ON_ERROR_STOP=1 \
        -U "$LEADER_USER" \
        -d "$source_database" <<'SQL'
DO $$
DECLARE
    target_table text;
    has_legacy_column boolean;
    has_hash_column boolean;
BEGIN
    FOREACH target_table IN ARRAY ARRAY['clients', 'unverified_clients'] LOOP
        IF to_regclass(format('public.%I', target_table)) IS NULL THEN
            RAISE EXCEPTION 'Required follower table public.% is missing', target_table;
        END IF;

        SELECT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public'
                AND table_name = target_table
                AND column_name = 'client_secret'
        ) INTO has_legacy_column;

        SELECT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = 'public'
                AND table_name = target_table
                AND column_name = 'client_secret_hash'
        ) INTO has_hash_column;

        IF has_legacy_column AND NOT has_hash_column THEN
            EXECUTE format(
                'ALTER TABLE public.%I RENAME COLUMN client_secret TO client_secret_hash',
                target_table
            );
        ELSIF NOT has_legacy_column AND has_hash_column THEN
            CONTINUE;
        ELSE
            RAISE EXCEPTION
                'Follower table public.% has an unexpected client-secret schema (legacy column: %, hash column: %)',
                target_table,
                has_legacy_column,
                has_hash_column;
        END IF;
    END LOOP;
END;
$$;
SQL

    subscription_state="$(
        psql \
            -At \
            -v ON_ERROR_STOP=1 \
            -U "$LEADER_USER" \
            -d "$source_database" \
            -c "SELECT COALESCE((SELECT subenabled::text FROM pg_subscription WHERE subname = '${subscription}'), 'missing');"
    )"
    if [[ "$subscription_state" == "true" ]]; then
        echo "${subscription} is already enabled; leaving it in place."
        continue
    fi
    if [[ "$subscription_state" != "missing" ]]; then
        echo "${subscription} exists but is not enabled; refusing to report replication setup complete." >&2
        exit 1
    fi

    echo "Creating ${subscription}..."
    psql \
        -v ON_ERROR_STOP=1 \
        -v subscription="$subscription" \
        -v source_conn="$source_conn" \
        -v publication="$publication" \
        -v slot="$slot" \
        -U "$LEADER_USER" \
        -d "$source_database" <<'SQL'
CREATE SUBSCRIPTION :"subscription"
    CONNECTION :'source_conn'
    PUBLICATION :"publication"
    WITH (copy_data = true, create_slot = true, enabled = true, slot_name = :'slot');
SQL
done

echo "Logical replication schema preparation and subscription setup complete"
