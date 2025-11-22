# Incident Response Runbook

## Overview

This runbook defines procedures for responding to security incidents affecting OTP/magic link authentication services, including compromised tokens, provider outages, abuse detection, and recovery validation.

## Incident Classification

### Severity Levels

| Severity | Impact | Response Time | Examples |
|----------|--------|---------------|----------|
| **P0 - Critical** | Complete service outage or active security breach | < 15 minutes | Database compromise, mass token leakage, DDoS attack |
| **P1 - High** | Partial service degradation or suspected compromise | < 1 hour | Single provider outage, elevated rate limit violations, suspicious token validation patterns |
| **P2 - Medium** | Limited impact or potential security issue | < 4 hours | Intermittent provider failures, moderate abuse patterns, configuration drift |
| **P3 - Low** | Minimal impact or informational security event | < 24 hours | Low-level abuse attempts, audit log anomalies, monitoring gaps |

## Incident Types and Response Procedures

### 1. Compromised Token Discovery

**Indicators:**
- Token values found in logs, traces, or monitoring dashboards
- Unauthorized access using valid OTP/magic link tokens
- Mass token validation attempts from single IP
- Tokens shared across multiple users or sessions

**Immediate Actions (< 15 minutes):**

1. **Isolate Affected Users**:
   ```bash
   # Revoke all active challenges for compromised users
   docker compose exec postgres psql -U USR -d DB -c \
     "DELETE FROM auth_challenges WHERE user_id IN ('user1', 'user2', ...);"

   # Force password reset for affected accounts
   docker compose exec postgres psql -U USR -d DB -c \
     "UPDATE users SET force_password_reset = true WHERE sub IN ('user1', 'user2', ...);"
   ```

2. **Block Suspicious IPs**:
   ```bash
   # Add IP to blocklist (if IP-based attack)
   # Update config file or use runtime configuration
   curl -k https://127.0.0.1:9090/admin/api/v1/blocklist/ip \
     -X POST \
     -H "Content-Type: application/json" \
     -d '{"ip": "192.168.1.100", "reason": "Mass token validation attempts"}'
   ```

3. **Rotate Cryptographic Keys** (see Token Rotation Runbook):
   ```bash
   # Immediate key rotation
   ./scripts/emergency-key-rotation.sh

   # Verify new keys loaded
   docker compose logs cryptoutil-sqlite | grep -i "unseal.*success"
   ```

**Investigation (< 1 hour):**

1. **Audit Log Analysis**:
   ```bash
   # Export audit logs for forensic analysis
   docker compose logs cryptoutil-sqlite --since 24h | \
     grep "identity.auth.validation_attempt" > audit-$(date +%Y%m%d-%H%M%S).log

   # Look for anomalies
   grep "outcome=success" audit-*.log | awk '{print $NF}' | sort | uniq -c | sort -rn
   ```

2. **Database Forensics**:
   ```sql
   -- Check for mass token generation from single user
   SELECT user_id, COUNT(*) as challenge_count, MIN(created_at), MAX(created_at)
   FROM auth_challenges
   GROUP BY user_id
   HAVING COUNT(*) > 10
   ORDER BY challenge_count DESC;

   -- Check for validation attempts on expired challenges
   SELECT challenge_id, user_id, validation_attempts, last_attempt_at
   FROM challenge_validation_attempts
   WHERE challenge_expired = true
   ORDER BY validation_attempts DESC;
   ```

3. **OpenTelemetry Trace Analysis**:
   ```bash
   # Check Grafana Tempo for suspicious traces
   # Look for token values in trace spans (CRITICAL: should never appear)
   # Example query in Grafana: {service.name="cryptoutil"} | json | token != ""
   ```

**Recovery (< 4 hours):**

1. **Validate No Token Leakage**:
   ```bash
   # Grep logs for potential token patterns (6-digit numeric, hex tokens)
   docker compose logs cryptoutil-sqlite | grep -E '\b[0-9]{6}\b|\b[a-f0-9]{32,}\b'

   # Check OpenTelemetry spans for token attributes
   # Should find ZERO occurrences of plaintext tokens
   ```

2. **Notify Affected Users**:
   ```bash
   # Send email notification
   cat <<EOF > incident-notification.txt
   Subject: Security Incident Notification

   Dear User,

   We recently detected unauthorized access attempts on your account.
   As a precaution, we have:
   - Invalidated all active authentication tokens
   - Reset your password (temporary password sent separately)
   - Enabled additional security monitoring

   Please:
   1. Change your password immediately
   2. Review recent account activity
   3. Enable multi-factor authentication

   If you have questions, contact security@example.com

   Security Team
   EOF

   # Send via email provider (pseudocode)
   ./scripts/send-bulk-notification.sh --users=affected_users.txt --template=incident-notification.txt
   ```

3. **Post-Incident Report**:
   - Timeline of detection, response, and recovery
   - Root cause analysis (token storage, logging, access controls)
   - Remediation steps taken (key rotation, user notifications, code fixes)
   - Lessons learned and preventive measures

### 2. SMS/Email Provider Outage

**Indicators:**
- High rate of delivery failures (SendSMS/SendEmail errors)
- Provider API returning 5xx errors or timeouts
- Monitoring alerts for provider uptime degradation

**Immediate Actions (< 15 minutes):**

1. **Verify Outage Scope**:
   ```bash
   # Check recent delivery failures
   docker compose logs cryptoutil-sqlite | grep -i "failed to send" | tail -50

   # Check provider status page
   curl https://status.twilio.com/api/v2/status.json
   curl https://status.sendgrid.com/api/v2/status.json
   ```

2. **Switch to Backup Provider** (if configured):
   ```yaml
   # Update config file (configs/identity/production.yml)
   delivery:
     sms:
       primary_provider: "twilio"
       backup_provider: "plivo"
       failover_threshold: 5  # Switch after 5 consecutive failures
     email:
       primary_provider: "sendgrid"
       backup_provider: "ses"
       failover_threshold: 3
   ```

3. **Enable Graceful Degradation**:
   ```yaml
   # Allow email-only auth if SMS unavailable
   auth:
     sms_otp:
       enabled: false  # Disable SMS temporarily
     magic_link:
       enabled: true   # Keep email working
   ```

**Investigation (< 1 hour):**

1. **Provider Status Monitoring**:
   ```bash
   # Poll provider status API every 30 seconds
   while true; do
     curl -s https://status.twilio.com/api/v2/status.json | jq '.status.indicator'
     sleep 30
   done
   ```

2. **Rate Limiting Impact Analysis**:
   ```bash
   # Check if outage triggered rate limiting false positives
   docker compose logs cryptoutil-sqlite | grep "identity.ratelimit.exceeded"

   # Temporarily increase rate limits if users retrying due to delivery failures
   # Update config or use runtime configuration API
   ```

**Recovery (< 4 hours):**

1. **Re-enable Primary Provider**:
   ```yaml
   # Restore primary provider when status confirmed healthy
   delivery:
     sms:
       primary_provider: "twilio"
       enabled: true
   ```

2. **Resend Failed Notifications** (if applicable):
   ```sql
   -- Identify users with failed delivery attempts during outage
   SELECT user_id, phone_number, email, failed_attempts, last_attempt_at
   FROM delivery_failures
   WHERE last_attempt_at BETWEEN '2025-01-15 10:00:00' AND '2025-01-15 12:00:00';

   -- Trigger manual resend (via admin API or script)
   ```

3. **Validate Service Recovery**:
   ```bash
   # Test SMS delivery
   curl -k https://127.0.0.1:8080/browser/api/v1/identity/auth/otp/initiate \
     -X POST -H "Content-Type: application/json" \
     -d '{"user_id": "test-user"}'

   # Test email delivery
   curl -k https://127.0.0.1:8080/browser/api/v1/identity/auth/magic-link/initiate \
     -X POST -H "Content-Type: application/json" \
     -d '{"user_id": "test-user"}'
   ```

### 3. Abuse Detection (Rate Limiting Violations)

**Indicators:**
- Sustained high rate of `identity.ratelimit.exceeded` metrics
- Single IP generating excessive token requests
- Single user generating excessive challenges
- Validation attempts on non-existent challenges (enumeration attack)

**Immediate Actions (< 30 minutes):**

1. **Identify Abusive Patterns**:
   ```bash
   # Top IPs by rate limit violations (last hour)
   docker compose logs cryptoutil-sqlite --since 1h | \
     grep "identity.ratelimit.exceeded" | \
     awk '{print $NF}' | sort | uniq -c | sort -rn | head -20

   # Top users by rate limit violations
   docker compose logs cryptoutil-sqlite --since 1h | \
     grep "identity.ratelimit.exceeded" | \
     grep -oP 'user_id=\K[^ ]+' | sort | uniq -c | sort -rn | head -20
   ```

2. **Temporary IP Blocking** (if automated attack):
   ```bash
   # Block abusive IPs at load balancer or firewall level
   # Example: iptables rule
   iptables -A INPUT -s 192.168.1.100 -j DROP

   # Or update allowed-ips config to exclude abusive IPs
   ```

3. **Increase Rate Limiting Strictness** (temporary):
   ```yaml
   # Reduce rate limits during attack
   rate_limiting:
     per_user:
       window: 5m
       max_attempts: 3  # Reduced from 10
     per_ip:
       window: 5m
       max_attempts: 10  # Reduced from 50
   ```

**Investigation (< 2 hours):**

1. **Attack Pattern Analysis**:
   ```bash
   # Extract attack timeline
   docker compose logs cryptoutil-sqlite --since 24h | \
     grep "identity.ratelimit.exceeded" | \
     awk '{print $1, $2}' | uniq -c

   # Identify target users (if credential stuffing)
   docker compose logs cryptoutil-sqlite --since 24h | \
     grep "invalid_otp" | \
     grep -oP 'user_id=\K[^ ]+' | sort | uniq -c | sort -rn | head -50
   ```

2. **CAPTCHA Integration** (if persistent abuse):
   ```yaml
   # Enable CAPTCHA for high-risk operations
   auth:
     captcha:
       enabled: true
       provider: "hcaptcha"
       site_key_file: "/run/secrets/hcaptcha_site_key.secret"
       secret_key_file: "/run/secrets/hcaptcha_secret_key.secret"
       threshold: 3  # Require CAPTCHA after 3 failed attempts
   ```

**Recovery (< 4 hours):**

1. **Restore Normal Rate Limits**:
   ```yaml
   # Return to standard rate limits once attack subsides
   rate_limiting:
     per_user:
       window: 5m
       max_attempts: 10
     per_ip:
       window: 5m
       max_attempts: 50
   ```

2. **Unblock Legitimate Users** (if false positives):
   ```sql
   -- Clear rate limit state for legitimate users
   DELETE FROM rate_limit_attempts WHERE user_id IN ('user1', 'user2', ...);
   ```

3. **Update Abuse Detection Rules**:
   - Add IP reputation checks (block known VPN/proxy IPs)
   - Implement device fingerprinting (detect automated scripts)
   - Enable behavioral analysis (unusual geographic patterns)

### 4. Database Connectivity Issues

**Indicators:**
- High rate of `failed to store challenge` errors
- Database connection pool exhaustion
- Slow query performance (>1s for token operations)
- Database health check failures

**Immediate Actions (< 15 minutes):**

1. **Check Database Health**:
   ```bash
   # PostgreSQL health
   docker compose exec postgres pg_isready -U USR -d DB

   # Check connection count
   docker compose exec postgres psql -U USR -d DB -c \
     "SELECT count(*) FROM pg_stat_activity WHERE datname='DB';"

   # SQLite in-memory (check container health)
   docker compose exec cryptoutil-sqlite wget --no-check-certificate -q -O - https://127.0.0.1:9090/readyz
   ```

2. **Increase Connection Pool** (temporary):
   ```yaml
   # Update database config (configs/identity/production.yml)
   database:
     max_open_conns: 50  # Increased from 25
     max_idle_conns: 10  # Increased from 5
     conn_max_lifetime: 5m
   ```

3. **Graceful Degradation**:
   ```yaml
   # Enable in-memory fallback (if PostgreSQL unavailable)
   database:
     primary: "postgres"
     fallback: "sqlite"
     fallback_mode: "readonly"  # Allow reads, block writes
   ```

**Investigation (< 1 hour):**

1. **Query Performance Analysis**:
   ```sql
   -- Check slow queries (PostgreSQL)
   SELECT query, mean_exec_time, calls
   FROM pg_stat_statements
   ORDER BY mean_exec_time DESC
   LIMIT 10;

   -- Check lock contention
   SELECT blocked_locks.pid AS blocked_pid,
          blocking_locks.pid AS blocking_pid,
          blocked_activity.query AS blocked_query,
          blocking_activity.query AS blocking_query
   FROM pg_catalog.pg_locks blocked_locks
   JOIN pg_catalog.pg_locks blocking_locks
     ON blocking_locks.locktype = blocked_locks.locktype;
   ```

2. **Connection Pool Monitoring**:
   ```bash
   # Check Prometheus metrics for connection pool usage
   curl -s http://127.0.0.1:8888/metrics | grep db_connections
   ```

**Recovery (< 4 hours):**

1. **Database Failover** (if primary unavailable):
   ```bash
   # Promote PostgreSQL replica to primary
   docker compose exec postgres-replica pg_ctl promote

   # Update connection string to point to new primary
   ```

2. **Validate Data Consistency**:
   ```sql
   -- Check for orphaned challenges (no corresponding user)
   SELECT ac.id, ac.user_id
   FROM auth_challenges ac
   LEFT JOIN users u ON ac.user_id = u.sub
   WHERE u.sub IS NULL;

   -- Check for expired challenges not cleaned up
   SELECT COUNT(*) FROM auth_challenges WHERE expires_at < NOW();
   ```

3. **Restore Connection Pool Defaults**:
   ```yaml
   # Return to normal connection pool settings
   database:
     max_open_conns: 25
     max_idle_conns: 5
   ```

## Escalation Paths

### Internal Escalation

| Role | Contact | Responsibility | Escalation Trigger |
|------|---------|----------------|-------------------|
| **On-Call Engineer** | oncall@example.com | Initial response, triage | Incident detected |
| **Security Team Lead** | security-lead@example.com | Security incident coordination | P0/P1 security incidents |
| **Database Administrator** | dba@example.com | Database performance/recovery | Database connectivity issues |
| **Engineering Manager** | eng-manager@example.com | Resource allocation, external comms | P0 incidents >1 hour |
| **VP Engineering** | vp-eng@example.com | Executive decision-making | P0 incidents >4 hours, data breach |

### External Escalation

| Situation | Contact | Timeline | Actions |
|-----------|---------|----------|---------|
| **Data Breach** | Legal team, CISO | Immediate | Regulatory notification (GDPR 72 hours), customer notification |
| **Provider Outage** | Provider support | < 30 minutes | Open critical support ticket, request ETA |
| **DDoS Attack** | ISP, CDN provider | < 15 minutes | Enable DDoS mitigation, traffic filtering |

### Communication Templates

**Internal Incident Notification (Slack/Teams):**
```
🚨 INCIDENT ALERT - P{SEVERITY}
Title: {Brief incident description}
Detected: {Timestamp}
Impact: {User-facing impact}
Status: {Investigating / Mitigating / Resolved}
Owner: @{oncall-engineer}
War Room: {Zoom/Teams link}

Updates: {Thread below}
```

**External Customer Notification:**
```
Subject: Service Incident Notification - {Date}

Dear Customers,

We experienced a temporary service disruption affecting authentication services between {start_time} and {end_time} UTC.

Impact:
- {Percentage}% of users experienced delays or failures during SMS OTP/Magic Link authentication
- No user data was compromised
- All services fully restored as of {resolution_time}

Root Cause:
{Brief non-technical explanation}

Resolution:
{Steps taken to resolve and prevent recurrence}

We apologize for any inconvenience. If you have questions, contact support@example.com

Operations Team
```

## Recovery Validation Checklist

After incident resolution, validate ALL items before closing incident:

### Functional Validation

- [ ] SMS OTP generation working (test with 3+ users)
- [ ] Email magic link generation working (test with 3+ users)
- [ ] Token validation succeeding for valid OTPs/links
- [ ] Token validation failing for invalid/expired tokens
- [ ] Rate limiting functioning correctly (per-user and per-IP)
- [ ] Audit logging capturing all events (generation, validation, invalidation)

### Security Validation

- [ ] No plaintext tokens in logs/traces/monitoring dashboards
- [ ] Bcrypt hashing functioning (verify hash format in database)
- [ ] Key rotation successful (new keys loaded, old keys archived)
- [ ] IP blocklists updated (abusive IPs blocked at firewall/load balancer)
- [ ] User notifications sent to affected accounts

### Performance Validation

- [ ] Token generation latency <500ms (p95)
- [ ] Token validation latency <200ms (p95)
- [ ] Database connection pool healthy (no exhaustion warnings)
- [ ] Provider API response times normal (<1s for SMS/email)
- [ ] OpenTelemetry metrics reporting correctly

### Monitoring Validation

- [ ] Prometheus alerts firing correctly (test with synthetic incident)
- [ ] Grafana dashboards showing accurate metrics
- [ ] Audit log exports successful (verify log integrity)
- [ ] Rate limiting metrics updating in real-time

## Post-Incident Activities

### Incident Report Template

```markdown
# Incident Report: {Title}

## Summary
- **Incident ID**: INC-{YYYY-MM-DD}-{NN}
- **Severity**: P{0-3}
- **Detected**: {Timestamp}
- **Resolved**: {Timestamp}
- **Duration**: {Hours/Minutes}
- **Impact**: {User count, service degradation percentage}

## Timeline
- {HH:MM} - Incident detected (monitoring alert / user report)
- {HH:MM} - On-call engineer notified
- {HH:MM} - Incident confirmed, response initiated
- {HH:MM} - Root cause identified
- {HH:MM} - Mitigation applied
- {HH:MM} - Service restored, monitoring validation
- {HH:MM} - Incident resolved

## Root Cause
{Technical explanation of what caused the incident}

## Resolution
{Steps taken to resolve the incident}

## Preventive Measures
1. {Action item 1}
2. {Action item 2}
3. {Action item 3}

## Lessons Learned
- **What went well**: {Positive aspects of response}
- **What could be improved**: {Areas for improvement}
- **Action items**: {Follow-up tasks with owners and deadlines}
```

### Preventive Actions

After EVERY incident, implement at least one preventive measure:

1. **Code Changes**:
   - Improve error handling for provider failures
   - Add graceful degradation for database outages
   - Implement circuit breakers for external dependencies

2. **Monitoring Enhancements**:
   - Add new Prometheus alerts for detected failure modes
   - Create Grafana dashboards for incident-specific metrics
   - Implement automated runbook execution (PagerDuty/Opsgenie)

3. **Process Improvements**:
   - Update incident response runbook with new scenarios
   - Conduct tabletop exercises for similar incidents
   - Document escalation paths and contact updates

4. **Infrastructure Changes**:
   - Add redundancy for single points of failure
   - Implement automated failover for critical services
   - Increase monitoring coverage for blind spots

## References

- Token rotation runbook: `docs/02-identityV2/token-rotation-runbook.md`
- Token hashing implementation: `internal/identity/idp/userauth/token_hashing.go`
- Audit logging: `internal/identity/idp/userauth/audit.go`
- Rate limiting: `internal/identity/idp/userauth/rate_limiter.go`
- SMS OTP authenticator: `internal/identity/idp/userauth/sms_otp.go`
- Magic link authenticator: `internal/identity/idp/userauth/magic_link.go`

## Change History

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2025-01-XX | 1.0 | Initial incident response runbook | Identity Team |
