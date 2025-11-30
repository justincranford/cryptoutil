# Client Secret Rotation Runbook

## Overview

OAuth 2.1 client secret rotation procedures for security and compliance.

## When to Rotate

### Scheduled Rotation

- **Frequency**: Every 90 days (compliance requirement)
- **Planning**: Schedule during maintenance windows
- **Notification**: Alert client owners 7 days in advance

### Security Incident

- **Immediate**: Secret compromise suspected/confirmed
- **Priority**: Critical (rotate within 1 hour)
- **Follow-up**: Incident post-mortem required

### Employee Departure

- **Timing**: When team member with secret access leaves
- **Scope**: Rotate all secrets accessible by departing employee
- **Verification**: Confirm old secrets invalidated

### System Migration

- **Timing**: When moving to new secret management system
- **Planning**: Coordinate with client applications
- **Testing**: Verify new secrets work before invalidating old ones

## How to Rotate (API)

### Prerequisites

- Client authentication credentials (current secret)
- Authorization to rotate (must be client owner or admin)
- Secure storage ready for new secret

### API Call

**Endpoint**: `POST /oauth2/v1/clients/{client_id}/rotate-secret`

**Authentication**: HTTP Basic Auth with current credentials

**Request**:

```bash
curl -X POST https://authz-server/oauth2/v1/clients/example-client/rotate-secret \
  -u "example-client:current-secret" \
  -H "Content-Type: application/json"
```

**Response** (200 OK):

```json
{
  "client_id": "example-client",
  "client_secret": "new-secret-here",
  "rotated_at": "2025-01-15T10:30:00Z",
  "message": "Client secret rotated successfully. Store this secret securely - it will not be shown again."
}
```

### Store New Secret Securely

**CRITICAL**: The new secret is ONLY shown once in the response. Store it immediately in your secret management system.

**Recommended Storage**:

- HashiCorp Vault
- AWS Secrets Manager
- Azure Key Vault
- Kubernetes Secrets (encrypted at rest)

**DO NOT**:

- Log the secret to files
- Store in version control
- Send via email/chat
- Store in environment variables (production)

### Update Client Application

1. **Deploy new secret** to application secret store
2. **Restart application** to load new secret
3. **Verify authentication** works with new secret
4. **Monitor logs** for authentication errors
5. **Remove old secret** from application configuration

## Rollback Procedure

### When to Rollback

- New secret not working after rotation
- Client application cannot authenticate
- Database issues during rotation
- Accidental rotation (wrong client)

### Rollback Steps

1. **Query rotation history**:

```sql
SELECT * FROM client_secret_history
WHERE client_id = 'example-client'
ORDER BY rotated_at DESC
LIMIT 2;
```

1. **Extract old secret hash** from most recent history record:

```sql
SELECT old_secret_hash FROM client_secret_history
WHERE client_id = 'example-client'
ORDER BY rotated_at DESC
LIMIT 1;
```

1. **Manual database update** (requires database admin access):

```sql
UPDATE clients
SET client_secret = '<old_secret_hash_from_step_2>'
WHERE client_id = 'example-client';
```

1. **Verify rollback**:

```bash
curl -X POST https://authz-server/oauth2/v1/token \
  -u "example-client:old-secret" \
  -d "grant_type=client_credentials"
```

1. **Notify operations team** of rollback for audit trail

### Rollback Limitations

- **Cannot recover plaintext secret** - only hash available in history
- **Requires database access** - API does not support rollback
- **Manual process** - no automated rollback mechanism
- **Time-sensitive** - perform within 1 hour of rotation

## Monitoring

### Application Logs

- Rotation events logged with client_id, rotated_by, reason
- Timestamp and operation result (success/failure)
- Example: `Client secret rotated: client_id=example-client, rotated_by=admin@example.com, reason=Scheduled 90-day rotation`

### Database Audit Trail

- All rotations tracked in `client_secret_history` table
- Columns: client_id, old_secret_hash, new_secret_hash, rotated_at, rotated_by, rotation_reason
- Retention: Permanent (compliance requirement)

### Alerts

**Alert on**:

- Failed rotation attempts (potential security incident)
- Multiple rotation attempts in short time (suspicious activity)
- Rotation without authenticated client (unauthorized access)

**Alert thresholds**:

- 3 failed attempts within 5 minutes: CRITICAL
- 5 rotations in 1 hour for same client: WARNING
- Rotation outside maintenance window: INFO

## Troubleshooting

### Error: "Client authentication failed"

**Cause**: Current secret is incorrect or expired

**Resolution**:

1. Verify current secret is correct
2. Check Authorization header format: `Basic base64(client_id:client_secret)`
3. Confirm client exists and is active in database
4. Check application logs for authentication errors

### Error: "Client can only rotate its own secret"

**Cause**: Authenticated client ID doesn't match target client ID

**Resolution**:

1. Verify authenticated client_id matches URL parameter
2. Use correct credentials for target client
3. If admin rotation needed, use admin client credentials (future enhancement)

### Error: "Failed to rotate secret"

**Cause**: Database transaction failed

**Resolution**:

1. Check database connectivity
2. Review application logs for details
3. Verify `client_secret_history` table exists (migration 0003)
4. Check database permissions for INSERT/UPDATE operations
5. Retry operation after resolving database issues

### Error: "Client not found"

**Cause**: Client ID doesn't exist in database

**Resolution**:

1. Verify client_id spelling/format
2. Check client was created successfully
3. Confirm client not deleted (soft delete check)

## Best Practices

### Secret Generation

- **Entropy**: 256 bits (32 bytes)
- **Encoding**: Base64 URL-safe encoding
- **Length**: 43 characters (after base64 encoding)
- **Algorithm**: PBKDF2-HMAC-SHA256 for hashing

### Rotation Schedule

- **Regular**: Every 90 days (minimum)
- **High-security clients**: Every 30 days
- **Low-risk clients**: Every 180 days (maximum)

### Communication

- **Advance notice**: 7 days before scheduled rotation
- **Notification channels**: Email, Slack, ticketing system
- **Documentation**: Update client documentation after rotation
- **Confirmation**: Request acknowledgment from client owners

### Testing

- **Pre-rotation**: Test current secret works
- **Post-rotation**: Test new secret works
- **Monitoring**: Watch authentication logs for 1 hour after rotation
- **Rollback plan**: Prepared before rotation begins

## Compliance

### Audit Requirements

- **Retention**: All rotation history retained permanently
- **Tracking**: Who rotated, when, why (mandatory fields)
- **Reports**: Quarterly rotation compliance reports
- **Review**: Annual audit of rotation procedures

### Security Standards

- **NIST 800-63B**: Client secret management guidelines
- **OAuth 2.1**: Client authentication best practices
- **PCI DSS**: Cryptographic key rotation requirements (if applicable)

## Related Documentation

- [Identity Server README](../../README.md) - Architecture overview
- [Database Migrations](../internal/identity/repository/migrations/README.md) - Schema changes
- [API Documentation](../api/identity/openapi_spec.yaml) - Endpoint specifications
- [Security Guidelines](../docs/02-identityV2/security-guidelines.md) - Security best practices
