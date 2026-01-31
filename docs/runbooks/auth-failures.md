# Runbook: High Authentication Failures

## Alert: HighAuthFailures

**Severity**: Warning
**Alert Expression**: `rate(auth_failures_total{job=~"cryptoutil.*"}[5m]) > 1`
**Duration**: 5 minutes

## Description

Authentication failure rate exceeds 1 per second (60/minute).

## Impact

- Potential brute force attack
- Legitimate users may be locked out
- Security incident indicator

## Investigation Steps

### 1. Identify Source IPs

```bash
# Check Grafana for IP breakdown
# Query: topk(10, sum by (source_ip) (rate(auth_failures_total[5m])))
```

### 2. Review Auth Logs

```bash
docker compose logs <service-name> --tail=500 | grep -i "auth\|login\|401\|403"
```

### 3. Check Failure Reasons

```bash
# Check Grafana for failure reasons
# Query: sum by (reason) (rate(auth_failures_total[5m]))
```

### 4. Verify Legitimate Traffic

```bash
# Check if failures correlate with legitimate services
# Compare with successful auth rates
# Query: rate(auth_successes_total[5m]) / rate(auth_failures_total[5m])
```

## Resolution Steps

### Brute Force Attack

1. Enable rate limiting per IP
2. Block offending IPs:
   ```bash
   # iptables example
   iptables -A INPUT -s <attacker_ip> -j DROP
   ```
3. Enable CAPTCHA for login
4. Report to security team

### Credential Stuffing

1. Enable MFA for affected accounts
2. Force password reset
3. Check for compromised credentials

### Configuration Issue

1. Verify service credentials
2. Check API key rotation status
3. Validate mTLS certificates

### Legitimate Traffic Spike

1. Review application changes
2. Check for client misconfiguration
3. Communicate with API consumers

## Escalation

- **If > 5 failures/sec**: Page on-call engineer immediately
- **If attack confirmed**: Activate security incident response

## Post-Incident

1. Implement IP reputation scoring
2. Add account lockout policies
3. Enable security alerting
4. Review authentication logs regularly
5. Consider implementing fail2ban
