# Task 08: Token Service Key Rotation Operations Guide

## Overview

This guide covers operational procedures for managing cryptographic key rotation in the Identity V2 token service, including normal rotation, emergency rotation, and rollback procedures.

## Key Rotation Policies

### Default Policy (Balanced Security)

```yaml
rotation_interval: 720h  # 30 days
grace_period: 168h       # 7 days
max_active_keys: 3
auto_rotation: false     # Manual rotation recommended
```

**Use Cases:**

- Standard production deployments
- Services with moderate security requirements
- Environments where manual rotation oversight is preferred

**Characteristics:**

- 30-day rotation cycle balances security and operational overhead
- 7-day grace period ensures smooth transitions
- Up to 3 active keys supported for extended grace periods

---

### Strict Policy (Maximum Security)

```yaml
rotation_interval: 168h  # 7 days
grace_period: 24h        # 1 day
max_active_keys: 2
auto_rotation: true      # Automated rotation enabled
```

**Use Cases:**

- High-security production environments
- Financial services, healthcare, government
- Compliance-driven deployments (PCI-DSS, HIPAA)

**Characteristics:**

- Weekly rotation minimizes key exposure window
- 1-day grace period enforces rapid key turnover
- Automated rotation reduces manual errors
- Strict 2-key limit minimizes attack surface

**CRITICAL**: Monitor automated rotations closely in first deployment cycles

---

### Development Policy (Relaxed)

```yaml
rotation_interval: 8760h  # 365 days
grace_period: 720h        # 30 days
max_active_keys: 5
auto_rotation: false
```

**Use Cases:**

- Development and testing environments
- CI/CD pipelines
- Local development workstations

**Characteristics:**

- Annual rotation reduces operational friction
- 30-day grace period accommodates long-running tests
- 5 active keys support extended testing scenarios

**WARNING**: Never use development policy in production

---

## Normal Key Rotation Procedures

### Manual Rotation (Recommended for Production)

**Step 1: Pre-Rotation Checklist**

```bash
# Verify current key status
curl -X GET https://localhost:8443/admin/keys/status \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Check token issuance rate (low-traffic window recommended)
curl -X GET https://localhost:8443/admin/metrics/token-issuance \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Verify monitoring and alerting operational
curl -X GET https://localhost:8443/admin/health/telemetry \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Step 2: Trigger Rotation**

```bash
# Rotate signing keys
curl -X POST https://localhost:8443/admin/keys/rotate/signing \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"algorithm": "RS256"}'

# Rotate encryption keys
curl -X POST https://localhost:8443/admin/keys/rotate/encryption \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Step 3: Verify Rotation**

```bash
# Check new active key
curl -X GET https://localhost:8443/admin/keys/active \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Verify old keys still valid for verification/decryption
curl -X GET https://localhost:8443/admin/keys/all \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Test token issuance with new key
curl -X POST https://localhost:8443/oauth2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&client_id=test&client_secret=secret"
```

**Step 4: Monitor Grace Period**

```bash
# Monitor token validation failures (should remain low)
curl -X GET https://localhost:8443/admin/metrics/token-validation-errors \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Track key usage distribution
curl -X GET https://localhost:8443/admin/keys/usage-stats \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Step 5: Wait for Grace Period Expiration**

- Default policy: Wait 7 days
- Strict policy: Wait 1 day
- Monitor key usage throughout grace period
- Old keys automatically pruned after expiration

---

### Automated Rotation (Strict Policy)

**Enable Automated Rotation:**

```go
policy := StrictKeyRotationPolicy()
policy.AutoRotationEnabled = true

generator := &ProductionKeyGenerator{...}
callback := func(keyID string) {
    log.Info("Key rotated", "key_id", keyID)
    // Send alerts, update monitoring
}

manager, _ := NewKeyRotationManager(policy, generator, callback)

// Start auto-rotation background process
ctx := context.Background()
go manager.StartAutoRotation(ctx, "RS256")
```

**Monitoring Automated Rotation:**

```bash
# Check rotation schedule
curl -X GET https://localhost:8443/admin/keys/rotation-schedule \
  -H "Authorization: Bearer ADMIN_TOKEN"

# View rotation history
curl -X GET https://localhost:8443/admin/keys/rotation-history \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Configure alerts for rotation events
curl -X POST https://localhost:8443/admin/alerts/configure \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "key_rotation",
    "channels": ["email", "slack", "pagerduty"],
    "severity": "info"
  }'
```

---

## Emergency Key Rotation

### When to Perform Emergency Rotation

**IMMEDIATE rotation required if:**

- Key compromise suspected or confirmed
- Insider threat detected
- Security breach involving token systems
- Compliance violation requiring key invalidation

**Step 1: Disable Current Keys**

```bash
# Immediately mark active key as invalid
curl -X POST https://localhost:8443/admin/keys/invalidate \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"key_id": "COMPROMISED_KEY_ID", "reason": "Security incident #12345"}'
```

**Step 2: Force Immediate Rotation**

```bash
# Rotate without waiting for grace period
curl -X POST https://localhost:8443/admin/keys/rotate/signing \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "algorithm": "RS256",
    "emergency": true,
    "skip_grace_period": false
  }'

curl -X POST https://localhost:8443/admin/keys/rotate/encryption \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"emergency": true, "skip_grace_period": false}'
```

**Step 3: Revoke All Tokens Signed with Compromised Key**

```bash
# Revoke tokens by key ID
curl -X POST https://localhost:8443/admin/tokens/revoke-by-key \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"key_id": "COMPROMISED_KEY_ID"}'
```

**Step 4: Force Client Re-Authentication**

```bash
# Invalidate all active sessions
curl -X POST https://localhost:8443/admin/sessions/invalidate-all \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Emergency key rotation"}'
```

**Step 5: Communicate to Stakeholders**

- Notify security team immediately
- Alert affected clients/users
- Document incident timeline
- Update security runbooks

---

## Rollback Procedures

### Scenario: New Key Causing Validation Failures

**Step 1: Identify Issue**

```bash
# Check token validation error rates
curl -X GET https://localhost:8443/admin/metrics/validation-errors \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Identify problematic key
curl -X GET https://localhost:8443/admin/keys/error-stats \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Step 2: Reactivate Previous Key**

```bash
# Mark previous key as active again
curl -X POST https://localhost:8443/admin/keys/reactivate \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"key_id": "PREVIOUS_WORKING_KEY_ID"}'
```

**Step 3: Deactivate Problematic Key**

```bash
# Mark new key as inactive (keep for verification only)
curl -X POST https://localhost:8443/admin/keys/deactivate \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"key_id": "PROBLEMATIC_KEY_ID"}'
```

**Step 4: Investigate Root Cause**

- Check key generation logs
- Verify algorithm compatibility
- Test key material integrity
- Review cryptographic library versions

**Step 5: Retry Rotation**

```bash
# After fixing issue, retry rotation
curl -X POST https://localhost:8443/admin/keys/rotate/signing \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"algorithm": "RS256", "retry": true}'
```

---

## Key Rotation Best Practices

### Planning and Scheduling

1. **Choose Low-Traffic Windows**
   - Rotate during maintenance windows
   - Avoid peak business hours
   - Coordinate with deployment schedules

2. **Test in Staging First**
   - Perform rotation in staging environment
   - Monitor for 24-48 hours
   - Validate all token operations

3. **Document Rotation Schedule**
   - Maintain rotation calendar
   - Set up automated reminders
   - Track rotation history

### Monitoring and Alerting

**Key Metrics to Monitor:**

- Token issuance rate by key ID
- Token validation success/failure rates
- Key age and expiration times
- Grace period utilization
- Number of active keys

**Alert Thresholds:**

```yaml
alerts:
  - name: key_expiring_soon
    condition: key.expires_in < 24h
    severity: warning

  - name: key_rotation_failed
    condition: rotation.status == "failed"
    severity: critical

  - name: validation_errors_spike
    condition: validation_errors > baseline * 2
    severity: critical

  - name: max_keys_reached
    condition: active_keys >= max_active_keys
    severity: warning
```

### Security Considerations

1. **Key Generation**
   - Use cryptographically secure random number generators
   - Generate keys on secure hardware (HSM recommended)
   - Never reuse key material
   - Minimum key sizes: RSA 2048-bit, ECDSA P-256

2. **Key Storage**
   - Store keys encrypted at rest
   - Use file-based secrets (never environment variables)
   - Restrict filesystem permissions (0600)
   - Implement key access logging

3. **Key Distribution**
   - Never transmit keys over insecure channels
   - Use secure key distribution mechanisms
   - Rotate distribution credentials regularly

### Compliance and Audit

**Audit Log Requirements:**

- All key rotation events (who, when, why)
- Key invalidation events
- Emergency rotation triggers
- Rollback procedures executed

**Compliance Documentation:**

```bash
# Generate rotation audit report
curl -X GET https://localhost:8443/admin/audit/key-rotation \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Accept: application/pdf" \
  --output rotation-audit-$(date +%Y%m%d).pdf
```

---

## Troubleshooting

### Common Issues

**Issue: "No active signing key available"**

```bash
# Cause: Rotation failed or key expired without replacement
# Solution: Force immediate rotation
curl -X POST https://localhost:8443/admin/keys/rotate/signing \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"algorithm": "RS256", "force": true}'
```

**Issue: Token validation failures after rotation**

```bash
# Cause: Clock skew or grace period too short
# Solution: Extend grace period temporarily
curl -X PATCH https://localhost:8443/admin/keys/policy \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"grace_period": "336h"}'  # 14 days
```

**Issue: Key rotation taking too long**

```bash
# Cause: Large number of tokens to re-sign
# Solution: Monitor progress and wait for completion
curl -X GET https://localhost:8443/admin/keys/rotation-progress \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

---

## CLI Commands Reference

### Key Status and Information

```bash
# List all keys
curl -X GET https://localhost:8443/admin/keys

# Get active signing key
curl -X GET https://localhost:8443/admin/keys/signing/active

# Get active encryption key
curl -X GET https://localhost:8443/admin/keys/encryption/active

# Get key by ID
curl -X GET https://localhost:8443/admin/keys/{key_id}

# Get key rotation policy
curl -X GET https://localhost:8443/admin/keys/policy
```

### Key Rotation Operations

```bash
# Rotate signing key
curl -X POST https://localhost:8443/admin/keys/rotate/signing \
  -d '{"algorithm": "RS256"}'

# Rotate encryption key
curl -X POST https://localhost:8443/admin/keys/rotate/encryption

# Batch rotation (signing + encryption)
curl -X POST https://localhost:8443/admin/keys/rotate/all \
  -d '{"signing_algorithm": "RS256"}'
```

### Policy Management

```bash
# Update rotation policy
curl -X PATCH https://localhost:8443/admin/keys/policy \
  -d '{
    "rotation_interval": "168h",
    "grace_period": "24h",
    "max_active_keys": 2,
    "auto_rotation": true
  }'

# Apply preset policy
curl -X POST https://localhost:8443/admin/keys/policy/apply \
  -d '{"preset": "strict"}'  # or "default", "development"
```

---

## References

- [RFC 7517: JSON Web Key (JWK)](https://datatracker.ietf.org/doc/html/rfc7517)
- [RFC 7518: JSON Web Algorithms (JWA)](https://datatracker.ietf.org/doc/html/rfc7518)
- [NIST SP 800-57: Key Management Recommendations](https://csrc.nist.gov/publications/detail/sp/800-57-part-1/rev-5/final)
- [OWASP Key Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Key_Management_Cheat_Sheet.html)
