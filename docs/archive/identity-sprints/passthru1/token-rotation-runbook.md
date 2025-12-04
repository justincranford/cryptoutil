# Token Rotation Runbook

## Overview

This runbook documents procedures for rotating cryptographic keys used for signing and validating OTP/magic link tokens, emergency token invalidation, and monitoring token lifecycle metrics.

## Token Types and Storage

### OTP Tokens (SMS)

- **Generation**: 6-digit numeric codes via crypto/rand
- **Storage**: Bcrypt-hashed (cost 12) in challenge store
- **Lifetime**: 5 minutes (configurable via `DefaultOTPLifetime`)
- **Verification**: Constant-time comparison via `VerifyToken(plaintext, hash)`

### Magic Link Tokens (Email)

- **Generation**: 32-byte secure random tokens (hex-encoded)
- **Storage**: Bcrypt-hashed (cost 12) in challenge store
- **Lifetime**: 15 minutes (configurable via `DefaultMagicLinkLifetime`)
- **Verification**: Constant-time comparison via `VerifyToken(plaintext, hash)`

### Token Hashing Details

- **Algorithm**: bcrypt with cost 12 (2^12 = 4096 iterations)
- **Salt**: Random per token (bcrypt handles automatically)
- **Hash Output**: ~60 bytes (base64-encoded salt + hash)
- **CRITICAL**: Tokens NEVER stored in plaintext (only hashes in database)

## Key Rotation Procedures

### When to Rotate Keys

1. **Scheduled Rotation**: Every 90 days (best practice for defense in depth)
2. **Security Incident**: Immediately if key compromise suspected
3. **Compliance Requirement**: As mandated by security policies or regulations
4. **Personnel Changes**: When staff with key access leave organization

### Rotation Steps

#### Step 1: Generate New Keys

**For development/testing (SQLite in-memory):**

```bash
# Generate new unseal secrets (5-of-5 shares)
go run ./cmd/cryptoutil keygen unseal \
  --shares=5 \
  --threshold=3 \
  --output-dir=./deployments/compose/cryptoutil/

# Verify file permissions (600 for secrets)
ls -la ./deployments/compose/cryptoutil/*.secret
```

**For production (PostgreSQL with Docker secrets):**

```bash
# Generate new unseal secrets on secure host
go run ./cmd/cryptoutil keygen unseal \
  --shares=5 \
  --threshold=3 \
  --output-dir=/secure/secrets/new/

# Verify file permissions
chmod 600 /secure/secrets/new/*.secret
```

#### Step 2: Deploy New Keys (Rolling Update)

**Development/testing:**

```bash
# Stop services
docker compose -f deployments/compose/compose.yml down -v

# Replace secret files
cp /secure/secrets/new/*.secret deployments/compose/cryptoutil/

# Restart services (keys loaded on startup)
docker compose -f deployments/compose/compose.yml up -d
```

**Production (zero-downtime rolling update):**

```bash
# Deploy new keys to Docker secrets volume (Kubernetes ConfigMap for K8s)
docker secret create cryptoutil_unseal_1of5_v2.secret /secure/secrets/new/cryptoutil_unseal_1of5.secret
docker secret create cryptoutil_unseal_2of5_v2.secret /secure/secrets/new/cryptoutil_unseal_2of5.secret
docker secret create cryptoutil_unseal_3of5_v2.secret /secure/secrets/new/cryptoutil_unseal_3of5.secret
docker secret create cryptoutil_unseal_4of5_v2.secret /secure/secrets/new/cryptoutil_unseal_4of5.secret
docker secret create cryptoutil_unseal_5of5_v2.secret /secure/secrets/new/cryptoutil_unseal_5of5.secret

# Update service to use new secrets (rolling update)
docker service update --secret-rm cryptoutil_unseal_1of5.secret --secret-add cryptoutil_unseal_1of5_v2.secret cryptoutil
docker service update --secret-rm cryptoutil_unseal_2of5.secret --secret-add cryptoutil_unseal_2of5_v2.secret cryptoutil
docker service update --secret-rm cryptoutil_unseal_3of5.secret --secret-add cryptoutil_unseal_3of5_v2.secret cryptoutil
docker service update --secret-rm cryptoutil_unseal_4of5.secret --secret-add cryptoutil_unseal_4of5_v2.secret cryptoutil
docker service update --secret-rm cryptoutil_unseal_5of5.secret --secret-add cryptoutil_unseal_5of5_v2.secret cryptoutil
```

#### Step 3: Verify Key Rotation

**Check service health:**

```bash
# Verify all instances healthy
docker compose ps

# Check logs for key loading confirmation
docker compose logs cryptoutil-sqlite | grep -i "unseal"
docker compose logs cryptoutil-postgres-1 | grep -i "unseal"
docker compose logs cryptoutil-postgres-2 | grep -i "unseal"

# Test token generation (should use new keys)
curl -k https://127.0.0.1:8080/browser/api/v1/identity/auth/otp/initiate \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test-user"}'
```

**Verify old tokens invalidated:**

```bash
# Attempt to verify old OTP (should fail)
curl -k https://127.0.0.1:8080/browser/api/v1/identity/auth/otp/verify \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"challenge_id": "old-challenge-uuid", "otp": "123456"}'

# Expected: HTTP 401 Unauthorized or "challenge not found"
```

#### Step 4: Secure Old Keys

```bash
# Archive old keys securely (encrypted backup for audit trail)
tar -czf cryptoutil-unseal-keys-$(date +%Y%m%d).tar.gz \
  deployments/compose/cryptoutil/*.secret

# Encrypt archive
gpg --encrypt --recipient security@example.com \
  cryptoutil-unseal-keys-$(date +%Y%m%d).tar.gz

# Move to secure backup location
mv cryptoutil-unseal-keys-$(date +%Y%m%d).tar.gz.gpg \
  /secure/key-archive/

# Securely delete unencrypted keys
shred -u -n 3 deployments/compose/cryptoutil/*.secret
```

## Emergency Token Invalidation

### Scenario: Suspected Token Compromise

**Immediate Actions:**

1. **Invalidate All Active Challenges**:

   ```bash
   # Connect to database
   docker compose exec postgres psql -U USR -d DB

   # Delete all active challenges (SQLite uses same table structure)
   DELETE FROM auth_challenges WHERE expires_at > NOW();

   # Verify deletion
   SELECT COUNT(*) FROM auth_challenges;
   ```

2. **Rotate Keys Immediately** (follow Step 1-4 above)

3. **Notify Users**:
   - Send email to all active users about security incident
   - Recommend password changes and review recent account activity
   - Provide incident timeline and resolution steps

4. **Audit Logging Review**:

   ```bash
   # Check audit logs for suspicious token validation attempts
   docker compose logs cryptoutil-sqlite | grep "identity.auth.validation_attempt"

   # Export audit logs for investigation
   docker compose logs cryptoutil-sqlite --since 24h > audit-$(date +%Y%m%d).log
   ```

### Scenario: SMS/Email Provider Compromise

**Immediate Actions:**

1. **Disable Compromised Provider**:

   ```yaml
   # Update config file (configs/identity/production.yml)
   delivery:
     sms:
       enabled: false  # Disable SMS until provider secured
     email:
       enabled: true   # Keep email working (or vice versa)
   ```

2. **Switch to Backup Provider** (if available):

   ```yaml
   # Configure alternate provider
   delivery:
     sms:
       provider: "twilio-backup"  # Or alternative SMS service
       credentials_file: "/run/secrets/twilio_backup.secret"
   ```

3. **Monitor for Abuse**:
   - Check rate limiting metrics for unusual patterns
   - Review IP-based rate limiting (identify automated attacks)
   - Verify no mass token generation from single IP

## Monitoring Token Lifecycle Metrics

### OpenTelemetry Metrics

**Token Generation Events:**

```prometheus
# Total token generation count (by method: sms_otp, magic_link)
identity_auth_token_generation_total{method="sms_otp"} 1500
identity_auth_token_generation_total{method="magic_link"} 850

# Rate limiting: total attempts and exceeded counts
identity_ratelimit_attempts_total{scope="user"} 2350
identity_ratelimit_exceeded_total{scope="user"} 45
identity_ratelimit_attempts_total{scope="ip"} 5200
identity_ratelimit_exceeded_total{scope="ip"} 120
```

**Token Validation Events:**

```prometheus
# Validation attempt outcomes (success, invalid_otp, expired)
identity_auth_validation_attempt_total{outcome="success"} 1200
identity_auth_validation_attempt_total{outcome="invalid_otp"} 85
identity_auth_validation_attempt_total{outcome="expired"} 65

# Latency histogram (token generation)
identity_auth_token_generation_duration_seconds_bucket{method="sms_otp",le="0.1"} 1450
identity_auth_token_generation_duration_seconds_bucket{method="sms_otp",le="0.5"} 1500
```

**Token Invalidation Events:**

```prometheus
# Manual token invalidations (admin action or user logout)
identity_auth_token_invalidation_total{reason="user_logout"} 320
identity_auth_token_invalidation_total{reason="admin_action"} 12
identity_auth_token_invalidation_total{reason="expired_cleanup"} 580
```

### Prometheus Alerting Rules

**High Rate Limit Exceeded Rate:**

```yaml
groups:
  - name: identity_auth_alerts
    interval: 30s
    rules:
      - alert: HighRateLimitExceededRate
        expr: |
          rate(identity_ratelimit_exceeded_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High rate limit exceeded rate detected"
          description: "Rate limit exceeded {{ $value }} times/sec (threshold: 10/sec)"
```

**Unusual Token Validation Failure Rate:**

```yaml
      - alert: HighTokenValidationFailureRate
        expr: |
          (
            rate(identity_auth_validation_attempt_total{outcome="invalid_otp"}[5m]) +
            rate(identity_auth_validation_attempt_total{outcome="expired"}[5m])
          ) /
          rate(identity_auth_validation_attempt_total[5m]) > 0.3
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "High token validation failure rate (>30%)"
          description: "{{ $value | humanizePercentage }} of token validations failing"
```

**Token Generation Latency Spike:**

```yaml
      - alert: TokenGenerationLatencyHigh
        expr: |
          histogram_quantile(0.95,
            rate(identity_auth_token_generation_duration_seconds_bucket[5m])
          ) > 1.0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Token generation latency high (p95 > 1s)"
          description: "95th percentile latency: {{ $value }}s"
```

### Grafana Dashboard Queries

**Token Generation Rate (by method):**

```promql
rate(identity_auth_token_generation_total[5m])
```

**Validation Success Rate:**

```promql
rate(identity_auth_validation_attempt_total{outcome="success"}[5m]) /
rate(identity_auth_validation_attempt_total[5m])
```

**Rate Limit Exceeded Percentage (per-user):**

```promql
rate(identity_ratelimit_exceeded_total{scope="user"}[5m]) /
rate(identity_ratelimit_attempts_total{scope="user"}[5m])
```

**Token Expiration Before Use (wasted tokens):**

```promql
rate(identity_auth_validation_attempt_total{outcome="expired"}[5m]) /
rate(identity_auth_token_generation_total[5m])
```

## Rollback Procedures

### If Key Rotation Fails

**Immediate rollback:**

```bash
# Restore old keys from backup
gpg --decrypt /secure/key-archive/cryptoutil-unseal-keys-YYYYMMDD.tar.gz.gpg | tar -xzf -

# Copy old keys back
cp cryptoutil-unseal-keys-YYYYMMDD/*.secret deployments/compose/cryptoutil/

# Restart services
docker compose -f deployments/compose/compose.yml down -v
docker compose -f deployments/compose/compose.yml up -d

# Verify service health
docker compose ps
docker compose logs cryptoutil-sqlite | grep -i "unseal"
```

### If Emergency Invalidation Too Broad

**Restore specific challenges (if backup available):**

```sql
-- Example: Restore valid challenges from backup table
INSERT INTO auth_challenges (id, user_id, method, expires_at, metadata, token_hash)
SELECT id, user_id, method, expires_at, metadata, token_hash
FROM auth_challenges_backup
WHERE expires_at > NOW() AND user_id NOT IN (
  SELECT user_id FROM compromised_users
);
```

## References

- Token hashing implementation: `internal/identity/idp/userauth/token_hashing.go`
- SMS OTP authenticator: `internal/identity/idp/userauth/sms_otp.go`
- Magic link authenticator: `internal/identity/idp/userauth/magic_link.go`
- Audit logging: `internal/identity/idp/userauth/audit.go`
- Rate limiting: `internal/identity/idp/userauth/rate_limiter.go`

## Change History

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2025-01-XX | 1.0 | Initial token rotation runbook | Identity Team |
