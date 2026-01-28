# Docker Secrets Pattern - Comprehensive Guide

## Overview

Docker secrets provide a secure mechanism for managing sensitive data (passwords, API keys, certificates, tokens) in containerized applications. **ALL cryptoutil services MUST use Docker secrets for credentials** - inline environment variables are FORBIDDEN.

### Why Docker Secrets are MANDATORY

- **Security**: Credentials never stored in version control or container images
- **Isolation**: Secrets only accessible to authorized services (not visible in `docker inspect`)
- **Auditability**: Secret changes tracked separately from application code
- **Rotation**: Update secrets without modifying compose.yml or rebuilding images
- **Compliance**: Follows Docker security best practices and industry standards

---

## Pattern Syntax

### Top-Level Secrets Section

Define all secrets in the `secrets:` section at the top level of `compose.yml`:

```yaml
secrets:
  postgres_user:
    file: ./secrets/postgres_user.secret
  postgres_password:
    file: ./secrets/postgres_password.secret
  postgres_db:
    file: ./secrets/postgres_db.secret
  unseal_1of5:
    file: ./secrets/unseal_1of5.secret
  api_key:
    file: ./secrets/api_key.secret
```

**Rules**:
- Each secret references a file in the `./secrets/` directory (relative to compose.yml)
- Secret files contain ONLY the credential value (no newlines, no comments)
- File permissions MUST be `440` (r--r-----): `chmod 440 ./secrets/*.secret`

### Service-Level Secrets Mounting

Services declare which secrets they need:

```yaml
services:
  myapp:
    image: myapp:latest
    secrets:
      - postgres_user
      - postgres_password
      - postgres_db
      - api_key
```

**At runtime**: Docker mounts secrets as files in `/run/secrets/` directory inside the container.

---

## PostgreSQL Implementation Patterns

### Pattern 1: Official PostgreSQL Image

PostgreSQL official image supports `*_FILE` environment variables for secrets:

```yaml
services:
  postgres:
    image: postgres:18
    secrets:
      - postgres_user
      - postgres_password
      - postgres_db
    environment:
      POSTGRES_USER_FILE: /run/secrets/postgres_user
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
      POSTGRES_DB_FILE: /run/secrets/postgres_db

secrets:
  postgres_user:
    file: ./secrets/postgres_user.secret
  postgres_password:
    file: ./secrets/postgres_password.secret
  postgres_db:
    file: ./secrets/postgres_db.secret
```

### Pattern 2A: Application Database URL - File Reference

When application supports reading database URL from a file:

```yaml
services:
  myapp:
    image: myapp:latest
    secrets:
      - postgres_url
    command:
      - --database-url
      - file:///run/secrets/postgres_url

secrets:
  postgres_url:
    file: ./secrets/postgres_url.secret
```

**Secret file content** (`./secrets/postgres_url.secret`):
```
postgres://myuser:mypassword@postgres:5432/mydb?sslmode=disable
```

**Used by**: KMS, Cipher-IM

### Pattern 2B: Application Database URL - Command Interpolation

When application needs inline URL string (doesn't support file reading):

```yaml
services:
  myapp:
    image: myapp:latest
    secrets:
      - postgres_user
      - postgres_password
      - postgres_db
    command:
      - --db-url
      - postgres://$(cat /run/secrets/postgres_user):$(cat /run/secrets/postgres_password)@postgres:5432/$(cat /run/secrets/postgres_db)?sslmode=disable
```

**Used by**: Identity services

**Note**: Both patterns 2A and 2B are **equally secure** - choose based on application capabilities.

---

## Unseal Keys Pattern (KMS)

For multi-key thresholds (e.g., 5-of-5 unseal keys):

```yaml
services:
  kms:
    image: cryptoutil-kms:latest
    secrets:
      - postgres_url
      - unseal_1of5
      - unseal_2of5
      - unseal_3of5
      - unseal_4of5
      - unseal_5of5
    command:
      - --database-url
      - file:///run/secrets/postgres_url
      - --unseal-key
      - file:///run/secrets/unseal_1of5
      - --unseal-key
      - file:///run/secrets/unseal_2of5
      - --unseal-key
      - file:///run/secrets/unseal_3of5
      - --unseal-key
      - file:///run/secrets/unseal_4of5
      - --unseal-key
      - file:///run/secrets/unseal_5of5

secrets:
  postgres_url:
    file: ./secrets/postgres_url.secret
  unseal_1of5:
    file: ./secrets/unseal_1of5.secret
  unseal_2of5:
    file: ./secrets/unseal_2of5.secret
  unseal_3of5:
    file: ./secrets/unseal_3of5.secret
  unseal_4of5:
    file: ./secrets/unseal_4of5.secret
  unseal_5of5:
    file: ./secrets/unseal_5of5.secret
```

**CRITICAL**: NEVER modify unseal secrets (breaks HKDF deterministic key derivation for instance interoperability).

---

## Migration Steps: Inline to Docker Secrets

### Step 1: Create Secrets Directory

```bash
cd deployments/myservice
mkdir -p secrets
```

### Step 2: Create Secret Files

```bash
# PostgreSQL credentials (3 files)
echo -n "myuser" > secrets/postgres_user.secret
echo -n "mypassword" > secrets/postgres_password.secret
echo -n "mydb" > secrets/postgres_db.secret

# OR single database URL file
echo -n "postgres://myuser:mypassword@postgres:5432/mydb?sslmode=disable" > secrets/postgres_url.secret

# Set permissions
chmod 440 secrets/*.secret
```

**CRITICAL**: Use `echo -n` (no trailing newline) or secret values may include unwanted newlines.

### Step 3: Add Top-Level Secrets Section

In `compose.yml`:

```yaml
secrets:
  postgres_user:
    file: ./secrets/postgres_user.secret
  postgres_password:
    file: ./secrets/postgres_password.secret
  postgres_db:
    file: ./secrets/postgres_db.secret
```

### Step 4: Update Service Configuration

**BEFORE** (inline environment variables - FORBIDDEN):

```yaml
services:
  postgres:
    image: postgres:18
    environment:
      POSTGRES_USER: myuser
      POSTGRES_PASSWORD: mypassword  # ❌ SECURITY VIOLATION
      POSTGRES_DB: mydb
```

**AFTER** (Docker secrets - CORRECT):

```yaml
services:
  postgres:
    image: postgres:18
    secrets:
      - postgres_user
      - postgres_password
      - postgres_db
    environment:
      POSTGRES_USER_FILE: /run/secrets/postgres_user
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
      POSTGRES_DB_FILE: /run/secrets/postgres_db

secrets:
  postgres_user:
    file: ./secrets/postgres_user.secret
  postgres_password:
    file: ./secrets/postgres_password.secret
  postgres_db:
    file: ./secrets/postgres_db.secret
```

### Step 5: Validate

**Syntax validation**:
```bash
docker compose -f compose.yml config > /dev/null
```
Expected: `✅ Valid` (no errors)

**Inline credentials detection**:
```bash
grep -E "PASSWORD|SECRET|TOKEN|PASSPHRASE|PRIVATE_KEY" compose.yml \
  | grep -v "# " \
  | grep -v "secrets:" \
  | grep -v "_FILE:" \
  | grep -v "run/secrets"
```
Expected: Empty output (no matches)

---

## Complete Examples

### Example 1: PostgreSQL with Official Image

```yaml
services:
  postgres:
    image: postgres:18
    secrets:
      - postgres_user
      - postgres_password
      - postgres_db
    environment:
      POSTGRES_USER_FILE: /run/secrets/postgres_user
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
      POSTGRES_DB_FILE: /run/secrets/postgres_db
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $(cat /run/secrets/postgres_user)"]
      interval: 10s
      timeout: 5s
      retries: 5

secrets:
  postgres_user:
    file: ./secrets/postgres_user.secret
  postgres_password:
    file: ./secrets/postgres_password.secret
  postgres_db:
    file: ./secrets/postgres_db.secret
```

### Example 2: Application with File Reference

```yaml
services:
  myapp:
    image: myapp:latest
    secrets:
      - database_url
      - api_key
    command:
      - --database-url
      - file:///run/secrets/database_url
      - --api-key
      - file:///run/secrets/api_key

secrets:
  database_url:
    file: ./secrets/database_url.secret
  api_key:
    file: ./secrets/api_key.secret
```

### Example 3: Application with Command Interpolation

```yaml
services:
  myapp:
    image: myapp:latest
    secrets:
      - postgres_user
      - postgres_password
      - postgres_db
      - api_key
    command:
      - --db-url
      - postgres://$(cat /run/secrets/postgres_user):$(cat /run/secrets/postgres_password)@postgres:5432/$(cat /run/secrets/postgres_db)?sslmode=disable
      - --api-key
      - $(cat /run/secrets/api_key)

secrets:
  postgres_user:
    file: ./secrets/postgres_user.secret
  postgres_password:
    file: ./secrets/postgres_password.secret
  postgres_db:
    file: ./secrets/postgres_db.secret
  api_key:
    file: ./secrets/api_key.secret
```

---

## Troubleshooting

### Issue: "secret not found" error

**Symptom**: Service fails to start with "secret 'X' not found"

**Cause**: Secret not defined in top-level `secrets:` section or file path incorrect

**Fix**:
1. Verify secret defined in top-level `secrets:` section
2. Check file path is relative to compose.yml: `file: ./secrets/secret_name.secret`
3. Verify file exists: `ls -la ./secrets/secret_name.secret`

### Issue: Incorrect credentials read from secrets

**Symptom**: Application receives credentials with trailing newlines or extra characters

**Cause**: Secret file created with `echo` instead of `echo -n`

**Fix**:
```bash
# WRONG - includes newline
echo "mypassword" > secrets/postgres_password.secret

# CORRECT - no newline
echo -n "mypassword" > secrets/postgres_password.secret
```

### Issue: Permission denied reading secrets

**Symptom**: Application cannot read `/run/secrets/secret_name`

**Cause**: Secret file permissions too restrictive or ownership incorrect

**Fix**:
```bash
chmod 440 ./secrets/*.secret
```

### Issue: Secrets visible in `docker inspect`

**Symptom**: Credentials appear in container environment

**Cause**: Using environment variables instead of Docker secrets

**Fix**: Convert to Docker secrets pattern (see Migration Steps above)

### Issue: YAML syntax error in secrets section

**Symptom**: `docker compose config` fails with parse error

**Cause**: Incorrect YAML indentation or structure

**Fix**:
```yaml
# WRONG - incorrect indentation
secrets:
postgres_user:
  file: ./secrets/postgres_user.secret

# CORRECT - proper indentation
secrets:
  postgres_user:
    file: ./secrets/postgres_user.secret
```

---

## Validation Checklist

Before deploying or committing compose files:

- [ ] All credentials moved to Docker secrets (NO inline environment variables)
- [ ] Top-level `secrets:` section defines all secrets with `file:` paths
- [ ] Services declare required secrets in `secrets:` list
- [ ] Secret files exist in `./secrets/` directory (relative to compose.yml)
- [ ] Secret files have 440 permissions: `chmod 440 ./secrets/*.secret`
- [ ] Secret files created with `echo -n` (no trailing newlines)
- [ ] Syntax validated: `docker compose -f compose.yml config > /dev/null` succeeds
- [ ] Inline credentials scan clean (grep command returns no matches)
- [ ] `.gitignore` includes `secrets/*.secret` (prevent accidental commits)

---

## References

### Official Documentation

- **Docker Secrets**: https://docs.docker.com/engine/swarm/secrets/
- **Docker Compose Secrets**: https://docs.docker.com/compose/use-secrets/
- **PostgreSQL Docker Image**: https://hub.docker.com/_/postgres (see "Environment Variables" section for `*_FILE` patterns)

### Cryptoutil Documentation

- **Copilot Docker Instructions**: `.github/instructions/04-02.docker.instructions.md` (comprehensive Docker patterns)
- **Copilot Security Instructions**: `.github/instructions/03-06.security.instructions.md` (secret management priority order)

### Reference Implementations

All cryptoutil services use Docker secrets - see these compose files for complete examples:

- **deployments/ca/compose.yml**: PostgreSQL with Docker secrets
- **deployments/kms/compose.yml**: PostgreSQL + unseal keys (5-of-5 threshold)
- **deployments/cipher/compose.yml**: PostgreSQL with Docker secrets (fixed in Phase 7 Task 7.1)
- **deployments/identity/compose.advanced.yml**: PostgreSQL with command interpolation pattern
- **deployments/identity/compose.e2e.yml**: PostgreSQL with Docker secrets (E2E testing variant)
- **deployments/jose/compose.yml**: SQLite pattern (no credentials needed)
- **deployments/identity/compose.simple.yml**: SQLite pattern (demo/development variant)
