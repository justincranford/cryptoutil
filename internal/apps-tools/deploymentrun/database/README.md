# Executable database lifecycle

The existing `cmd/cryptoutil` suite dispatches `db-lifecycle` to this package.
There is no separate database subcommand binary:

```text
/app/cryptoutil db-lifecycle --operation=init --input=/run/cryptoutil/secrets/postgres-bootstrap/lifecycle.json
/app/cryptoutil db-lifecycle --operation=migrate --input=/run/cryptoutil/secrets/identity-idp-migration-db/lifecycle.json
/app/cryptoutil db-lifecycle --operation=replicate --input=/run/cryptoutil/secrets/postgres-follower-bootstrap/lifecycle.json
```

The migration Secret name changes with the selected PS-ID. The canonical chart
and typed graph require the exact operation and manifest mount for each Job;
fixture commands are rejected. Host execution decodes protected manifests and
checks the independent selected catalog, identities, and every referenced
mounted key **before** Kubernetes execution.

## Protected input

The manifest is strict JSON, bounded to 1 MiB; unknown/duplicate properties,
trailing input, unsupported operations/services, extra catalog entries, and
missing material reject execution. It contains file references, never inline
credentials. Example shape (replace every placeholder explicitly):

```json
{
    "backend": "postgres",
    "targets": [
        {
            "psid": "identity-idp",
            "database": "GRAPH_DATABASE",
            "role": "GRAPH_OWNER_ROLE",
            "publication": "GRAPH_PUBLICATION",
            "subscription": "GRAPH_SUBSCRIPTION",
            "slot": "GRAPH_SLOT",
            "leader": {
                "dsnFile": "/run/cryptoutil/secrets/OWNED-SECRET/leader-dsn",
                "usernameFile": "/run/cryptoutil/secrets/OWNED-SECRET/leader-user",
                "passwordFile": "/run/cryptoutil/secrets/OWNED-SECRET/leader-password",
                "caFile": "/run/cryptoutil/secrets/OWNED-SECRET/leader-ca",
                "certificateFile": "/run/cryptoutil/secrets/OWNED-SECRET/leader-cert",
                "keyFile": "/run/cryptoutil/secrets/OWNED-SECRET/leader-key"
            },
            "follower": {
                "dsnFile": "/run/cryptoutil/secrets/OWNED-SECRET/follower-dsn",
                "usernameFile": "/run/cryptoutil/secrets/OWNED-SECRET/follower-user",
                "passwordFile": "/run/cryptoutil/secrets/OWNED-SECRET/follower-password",
                "caFile": "/run/cryptoutil/secrets/OWNED-SECRET/follower-ca",
                "certificateFile": "/run/cryptoutil/secrets/OWNED-SECRET/follower-cert",
                "keyFile": "/run/cryptoutil/secrets/OWNED-SECRET/follower-key"
            },
            "subscriberTLS": {
                "ca": "/ABSOLUTE-SUBSCRIBER-SERVER-PATH/ca",
                "certificate": "/ABSOLUTE-SUBSCRIBER-SERVER-PATH/cert",
                "key": "/ABSOLUTE-SUBSCRIBER-SERVER-PATH/key"
            }
        }
    ]
}
```

This illustrative replication shape is intentionally not runnable operator
configuration: graph names, lowercase role identifiers, actual Secret mounts,
trust material, and image pins must be supplied and certified.

For `init`, supply the complete selected union and both `leaderAdmin` and
`followerAdmin` connections using the same six-file connection shape. Supply
both target connections and graph publication/subscription/slot names, but omit
`subscriberTLS`. For `migrate`, supply exactly one selected service's target and
leader connection; omit follower/admin/subscriber transport. For `replicate`,
supply the complete selected union and both target connections plus
`subscriberTLS`; omit administrator connections.

PostgreSQL DSN files contain only `postgres://HOST:PORT/DATABASE`, without query,
fragment, user, or password. The database must match the graph at each barrier.
Username files identify explicit roles. Password files preserve spaces; one
terminal LF/CRLF record delimiter is removed, and embedded CR/LF/NUL are rejected.
Every Job-side file is under `/run/cryptoutil/secrets/SECRET/KEY` and must exist in
an explicitly allowed read-only Secret projection.

TLS requires complete approved PEM, a matching valid client keypair and explicit
server trust roots. The adapter enforces TLS 1.3, hostname verification, and no
plaintext fallback. It rejects recognized nonempty ambient PostgreSQL settings
before pgx can read service files. No environment-based configuration is used or
modified. Connection/statement/lock timeouts are 10s/60s/10s; the suite operation
has a ten-minute overall bound, additionally bounded by its owning Job policy.
Errors and CLI results do not print DSNs, passwords, SQL, or certificate bytes.

`subscriberTLS` paths are accessible to the **PostgreSQL subscriber server**,
not the Job filesystem. Its native replication transport requires the same
explicit hostname-verified TLS 1.3 policy. Native PostgreSQL inherently persists
connection credentials in its protected subscription catalog; this is not
ephemeral in-memory-only credential transport. Catalog access/storage must be
protected by the operator.

## Production effects and ordering

1. `init` checks preprovisioned owner roles and verifies-or-creates owned leader
   and follower databases. It creates/verifies an owned `FOR ALL TABLES`
   publication in each explicitly selected leader database, with every
   DML/TRUNCATE action enabled and partition-root remapping disabled.
2. `migrate` applies the actual seven services' merged embedded migration
   filesystems through the production framework/golang-migrate drivers.
3. `replicate` requires the leader publication and clean migrated schema; it
   does **not** rerun leader migrations. A fresh follower receives those same
   production migrations within one owner-controlled serializable transaction.
   Only tables created in that initially empty public schema have their
   transaction-owned metadata/seed rows cleared, using quoted identifiers,
   bounded inventory, and `TRUNCATE ... RESTART IDENTITY` without `CASCADE`.
   Otherwise the transaction rolls back. Existing schemas are verify-only:
   never migrated, truncated, or overwritten. A staged empty schema can resume.
4. Initial native replication copy requires every public follower table empty.
   Existing subscriptions must exactly match owner, database, publication, slot,
   enabled state and transport. Every subscription lookup/readiness query binds
   the cluster-wide catalog to the current database. No ALTER/DROP/replacement
   recovery is attempted.
5. Wait for the worker and all table-copy states ready, then compare the final
   follower schema and exactly one clean migration cursor against the leader.
   This is bootstrap/table-copy readiness, not a live WAL catch-up or
   continuously synchronized application acceptance proof.

Independent follower migrations followed by `copy_data=true` would duplicate
the default tenant and migration metadata primary keys. The transaction-owned
schema-only bootstrap avoids that conflict without deleting existing data.
The legacy shared-postgres shell's client-table assumptions, tier identifiers,
and `PGPASSWORD` export are not reused.

The framework's scoped SQL view releases its leased connection without closing
the caller's database. Transactional migration driver transactions become
savepoints; only the owner commits/rolls back. PostgreSQL transaction-scoped
advisory locks last until that boundary. Cancellation discards only a borrowed
PostgreSQL physical connection if needed to release a session lock, never the
owner pool or the shared in-memory SQLite fixture.

SQLite supports explicit `file:` migration DSNs and forbids PostgreSQL
credentials, shared initialization, and logical replication. Shared Kubernetes
database Jobs require PostgreSQL; SQLite runtime pods retain their own owning
startup migrations. A SQLite library test is not proof of migration of a live
runtime pod's PVC.

## Operator prerequisites and acceptance boundary

The operator must preprovision database owner/login/replication roles and the
operation-specific privileges: database creation/ownership, publication creation
(including PostgreSQL's `FOR ALL TABLES` requirements), migrations/schema access,
subscription creation, replication/login and table-copy access. Publication
initialization and replication verification must use its explicit owning role.
The adapter does not create roles, grant privileges, alter credentials, install
server certificates, tune WAL, or replace conflicting objects.

Also required: owned admitted cluster/namespaces, actual digest-pinned suite
images containing this route, current lifecycle-aware quotas, protected
manifests/material, correct leader/follower identities, subscriber-visible TLS
files, server authentication policy and logical replication capacity. Host PKI
prepublication and guarded Job completion remain separate boundaries; PKI Job
output publication/attestation is not implemented here.

Tests exercise ordered injected SQL/Kubernetes boundaries, every seven-service
merged filesystem on shared SQLite, transaction rollback, identity/collision
guards, and canonical native Helm contracts. No PostgreSQL container or cluster
was launched. Real PostgreSQL migrations/copy/privilege/concurrency behavior,
live image/cluster/client acceptance and rollback remain runtime E2E
prerequisites, not accepted Phase2 results.
