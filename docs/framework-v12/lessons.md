# Lessons - Framework V12: PostgreSQL mTLS + Private PS-ID App mTLS Trust

**Created**: 2025-06-26
**Last Updated**: 2026-04-16

---

## Phase 0: pki-init Patch — Cat 9 infra + Cat 14 postgres-only

- Cat 14 certs are postgres-only: use `PKIInitPostgresAppInstanceSuffixes()` not `PKIInitAdminInstanceSuffixes()`
- The SAME-AS-DIR-NAME convention: all files in a cert dir share the dir name (no secondary naming)

---

## Phase 1: PostgreSQL Server TLS — Leader + Follower

- PostgreSQL `ssl_cert_file`/`ssl_key_file`/`ssl_ca_file` point to Cat 11 keystore certs
- `ssl_ca_file` (truststore) is required for mTLS client cert verification

---

## Phase 2: PostgreSQL Replication Server TLS

- Follower's `primary_conninfo` requires `sslmode=verify-full sslrootcert=<Cat 10>` for server TLS

---

## Phase 3: Verify PostgreSQL Standalone

- Deferred (requires Docker). Validation done in Phase 9.

---

## Phase 4: PostgreSQL Client mTLS — HBA + GORM Config

**Critical patterns**:
1. **D2**: SSL params in YAML config files, NOT in postgres-url.secret DSN. Bare DSN: no `?sslmode=disable`
2. **stripQueryParam**: Call before appending new sslmode to prevent duplicate params. pgx first-value-wins but double params indicate a code smell.
3. **allowedInstanceKeys**: Add `server-admin-tls-*` keys here, not `requiredCommonKeys`
4. **secret-schemas.yaml**: Update `value_pattern` when postgres-url.secret format changes — secret-content pre-commit hook validates against it
5. **template-compliance**: All canonical templates must be updated when live deployment files change — template-compliance pre-commit hook validates template alignment

---

## Phase 5: PostgreSQL Replication Client mTLS

- Cat 13 certs used for replication client (not Cat 14 — Cat 14 is for app clients only)
- `pg_hba.conf` replication rule needs `clientcert=verify-full` alongside app rule

---

## Phase 6: Verify PostgreSQL Full Stack

- Deferred (requires Docker).

---

## Phase 7: Deployment Templates for PostgreSQL TLS

- **D5**: Full named volume strategy — `{suite}-certs:/certs:ro`. Never mount individual cert directories.
- Deployment templates must be updated in sync with live deployment files to pass template-compliance

---

## Phase 8: Deployment Linting

- lint-deployments covers 54 validators (not 8). Always run to verify template compliance.
- After any deployment file change, run lint-deployments before committing.

---

## Phase 9: Deployment Verification — PostgreSQL TLS

- Deferred (requires Docker).

---

## Phase 10: Private PS-ID App mTLS Trust

**Critical patterns**:
1. Admin mTLS config: `server-admin-tls-cert-file`, `server-admin-tls-key-file`, `server-admin-tls-ca-file`
2. When all 3 set: `tls.RequireAndVerifyClientCert` mode; when cert+key only: server TLS; when none: Auto
3. `applyAdminMTLS()` function in `admin.go` called from `newAdminHTTPServerInternal` with `osReadFileFn` seam for testing
4. **lint-go-test**: MUST use `require.*` not `assert.*` in tests — pre-commit hook enforces this

---

## Phase 11: Knowledge Propagation

- ENG-HANDBOOK.md `replace_string_in_file` must include the `---\n\n## N. Section` header in newString when that text was in oldString — otherwise the heading is silently deleted causing `#N-section-name` anchor breakage
- `lint-docs` catches broken anchors — always run after ENG-HANDBOOK.md edits
