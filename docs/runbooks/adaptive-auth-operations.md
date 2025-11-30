# Adaptive Authentication Operations Runbook

**Version**: 1.0
**Last Updated**: 2025-01-28
**Owner**: Identity Platform Team
**Scope**: Task 13 - Adaptive Authentication Engine

---

## Table of Contents

1. [Overview](#overview)
2. [Policy Tuning Procedures](#policy-tuning-procedures)
3. [False Positive Investigation](#false-positive-investigation)
4. [Risk Factor Debugging](#risk-factor-debugging)
5. [Policy Rollback](#policy-rollback)
6. [Incident Response](#incident-response)
7. [Monitoring Playbook](#monitoring-playbook)
8. [Common Issues](#common-issues)
9. [Escalation Procedures](#escalation-procedures)

---

## Overview

### Purpose

This runbook provides operational procedures for managing the Adaptive Authentication Engine (Task 13). It covers policy tuning, troubleshooting, incident response, and common operational scenarios.

### System Components

- **BehavioralRiskEngine**: Calculates risk scores based on authentication context
- **StepUpAuthenticator**: Evaluates step-up authentication requirements
- **PolicyLoader**: Loads and hot-reloads YAML policy files
- **Telemetry**: OpenTelemetry metrics and tracing
- **Simulation CLI**: Policy testing tool (`adaptive-sim`)

### Policy Files

- `configs/identity/policies/risk_scoring.yml`: Risk factor weights and thresholds
- `configs/identity/policies/step_up.yml`: Step-up authentication rules
- `configs/identity/policies/adaptive_auth.yml`: Global adaptive auth configuration

### Key Metrics

- `identity.risk.score`: Risk score distribution (0.0-1.0)
- `identity.stepup.triggered.total`: Step-up authentication requests
- `identity.operations.blocked.total`: Blocked operations
- `identity.policy.evaluation.duration`: Policy evaluation latency

---

## Policy Tuning Procedures

### 1. Simulation Phase

**Before deploying policy changes to production, ALWAYS simulate against historical data.**

#### Step 1: Collect Historical Authentication Logs

```bash
# Export authentication logs from production (last 7 days)
# Format: JSON array with timestamp, user_id, operation, location, network, device, auth_level, success, metadata

# Example query (adjust for your logging infrastructure):
kubectl logs -n identity deployment/identity-idp --since=168h | grep "auth_attempt" > historical_logs.json
```

#### Step 2: Create Test Policy Files

```bash
# Copy production policies to test directory
mkdir -p configs/identity/policies/test
cp configs/identity/policies/risk_scoring.yml configs/identity/policies/test/risk_scoring_test.yml
cp configs/identity/policies/step_up.yml configs/identity/policies/test/step_up_test.yml
cp configs/identity/policies/adaptive_auth.yml configs/identity/policies/test/adaptive_auth_test.yml

# Edit test policies with proposed changes
vim configs/identity/policies/test/risk_scoring_test.yml
```

#### Step 3: Run Simulation

```bash
# Run simulation CLI against historical logs
go run ./cmd/cicd adaptive-sim \
  --risk-scoring configs/identity/policies/test/risk_scoring_test.yml \
  --step-up configs/identity/policies/test/step_up_test.yml \
  --adaptive configs/identity/policies/test/adaptive_auth_test.yml \
  --logs historical_logs.json \
  --output simulation_results/

# Review simulation report
cat simulation_results/simulation_report.json | jq .
```

#### Step 4: Analyze Simulation Results

**Key Metrics to Review:**

- **Step-up rate**: Should be <15% (from step_up.yml `step_up_rate_threshold`)
- **Blocked rate**: Should be <5% (from step_up.yml `blocked_operation_rate_threshold`)
- **Risk distribution**: Majority should be low/medium risk
- **False positive rate**: Check user feedback and support tickets

**Acceptable Ranges:**

| Metric | Good | Warning | Critical |
|--------|------|---------|----------|
| Step-up rate | <10% | 10-15% | >15% |
| Blocked rate | <2% | 2-5% | >5% |
| Critical risk rate | <1% | 1-5% | >5% |
| Avg confidence | >0.7 | 0.3-0.7 | <0.3 |

**Example Analysis:**

```bash
# Extract key metrics from simulation report
jq '{
  step_up_rate: .metrics.step_up_rate,
  blocked_rate: .metrics.blocked_rate,
  risk_distribution: .metrics.risk_distribution,
  recommendations: .recommendations
}' simulation_results/simulation_report.json
```

### 2. Staging Deployment

**After successful simulation, deploy to staging environment.**

#### Step 1: Deploy to Staging

```bash
# Copy test policies to staging config directory
kubectl create configmap identity-policies-staging \
  --from-file=configs/identity/policies/test/ \
  -n identity-staging \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart identity-idp pods to load new policies
kubectl rollout restart deployment/identity-idp -n identity-staging
```

#### Step 2: Monitor Staging Metrics (24-48 hours)

```bash
# Check Grafana dashboard: "Adaptive Authentication - Risk & Step-Up Monitoring"
# Staging URL: https://grafana.staging.example.com/d/adaptive-auth

# Key panels to watch:
# - Risk Score Distribution: Should match simulation
# - Step-Up Trigger Rate: Should be <15%
# - Blocked vs Allowed Operations: Blocked rate <5%
# - Policy Evaluation Latency: p95 <500ms
```

#### Step 3: Staging Validation

**Checklist:**

- [ ] Step-up rate within acceptable range (<15%)
- [ ] Blocked operation rate within acceptable range (<5%)
- [ ] No increase in support tickets for authentication issues
- [ ] Policy evaluation latency acceptable (p95 <500ms)
- [ ] No policy load errors in logs
- [ ] Confidence scores healthy (avg >0.3)

### 3. Production Deployment

**After successful staging validation, deploy to production.**

#### Step 1: Production Deployment (Blue-Green)

```bash
# Create new configmap with production policies
kubectl create configmap identity-policies-v2 \
  --from-file=configs/identity/policies/test/ \
  -n identity-production \
  --dry-run=client -o yaml | kubectl apply -f -

# Update deployment to use new configmap
kubectl set env deployment/identity-idp -n identity-production \
  POLICY_CONFIG_VERSION=v2

# Gradual rollout (10% → 50% → 100%)
kubectl patch deployment identity-idp -n identity-production \
  -p '{"spec":{"strategy":{"rollingUpdate":{"maxSurge":"10%","maxUnavailable":"0%"}}}}'

kubectl rollout restart deployment/identity-idp -n identity-production
```

#### Step 2: Monitor Production Metrics (First 2 Hours)

**Critical monitoring period - watch for anomalies.**

```bash
# Open Grafana dashboard
# Production URL: https://grafana.prod.example.com/d/adaptive-auth

# Set up alert channels for immediate notification
# Slack channel: #identity-alerts
# PagerDuty: identity-oncall
```

**Watch for alerts:**

- HighStepUpRate (>15%)
- HighBlockedOperationRate (>5%)
- LowConfidenceScores (<0.3)
- PolicyLoadFailures

#### Step 3: Production Validation (24 Hours)

**Extended monitoring period.**

**Checklist:**

- [ ] Step-up rate stable and within range
- [ ] Blocked operation rate stable and within range
- [ ] No increase in support tickets
- [ ] User feedback positive or neutral
- [ ] Confidence scores healthy
- [ ] No performance degradation

### 4. Policy Iteration

**Continuous improvement based on production data.**

```bash
# Weekly policy review
# 1. Export last 7 days of production auth logs
# 2. Run simulation with current policies
# 3. Compare simulation vs actual metrics
# 4. Identify drift or anomalies
# 5. Propose policy adjustments
# 6. Repeat simulation → staging → production cycle
```

---

## False Positive Investigation

### Scenario: User Reports "Unable to Access Account - Step-Up Required"

**Step 1: Gather Context**

```bash
# Get user's recent authentication attempts
kubectl logs -n identity deployment/identity-idp | grep "user_id: USER123" | tail -n 50

# Key fields to extract:
# - timestamp
# - operation
# - risk_score
# - risk_level
# - required_level
# - decision (allow/step-up/block)
# - risk_factors (which factors contributed)
```

**Step 2: Analyze Risk Factors**

Example log entry:

```json
{
  "timestamp": "2025-01-28T14:30:00Z",
  "user_id": "USER123",
  "operation": "transfer_funds",
  "risk_score": 0.65,
  "risk_level": "high",
  "required_level": "strong_mfa",
  "decision": "step_up_required",
  "risk_factors": {
    "new_location": 0.25,
    "vpn_detected": 0.20,
    "unusual_time": 0.15,
    "device_age_days": 0.05
  },
  "confidence": 0.45
}
```

**Step 3: Determine Root Cause**

| Risk Factor | Weight | Likely Cause | Action |
|-------------|--------|--------------|--------|
| new_location (0.25) | High | User traveling | Legitimate, reduce weight |
| vpn_detected (0.20) | High | User using company VPN | Legitimate, add trusted VPN exception |
| unusual_time (0.15) | Medium | User in different timezone | Legitimate, adjust time windows |
| device_age_days (0.05) | Low | Recently replaced device | Legitimate, reduce weight |

**Step 4: Policy Adjustment Options**

**Option 1: Reduce VPN Weight for Corporate VPN Ranges**

```yaml
# configs/identity/policies/risk_scoring.yml
risk_factors:
  vpn_detected:
    weight: 0.20
    trusted_vpn_cidrs:
      - 10.0.0.0/8      # Corporate VPN range
      - 172.16.0.0/12   # Corporate VPN range
    trusted_vpn_weight: 0.05  # Reduced weight for trusted VPNs
```

**Option 2: Expand Time Windows**

```yaml
# configs/identity/policies/risk_scoring.yml
risk_factors:
  unusual_time:
    weight: 0.15
    typical_hours_start: "06:00"  # Expand from 08:00
    typical_hours_end: "23:00"    # Expand from 22:00
```

**Option 3: Adjust Step-Up Threshold for High-Risk Operations**

```yaml
# configs/identity/policies/step_up.yml
operations:
  transfer_funds:
    risk_levels:
      low:
        required_level: basic
      medium:
        required_level: mfa
      high:
        required_level: mfa  # Change from strong_mfa to mfa
      critical:
        required_level: strong_mfa
```

**Step 5: Test Adjustment**

```bash
# Run simulation with adjusted policy
go run ./cmd/cicd adaptive-sim \
  --risk-scoring configs/identity/policies/test/risk_scoring_adjusted.yml \
  --step-up configs/identity/policies/test/step_up_adjusted.yml \
  --logs production_logs_last_7d.json \
  --output simulation_results_adjusted/

# Compare before/after
diff simulation_results/simulation_report.json simulation_results_adjusted/simulation_report.json
```

---

## Risk Factor Debugging

### Scenario: High Risk Scores for Legitimate Users

**Step 1: Identify Problematic Risk Factors**

```bash
# Aggregate risk factor contributions across all high-risk assessments
kubectl logs -n identity deployment/identity-idp | \
  grep '"risk_level":"high"' | \
  jq -r '.risk_factors | to_entries | .[] | "\(.key): \(.value)"' | \
  sort | uniq -c | sort -rn

# Example output:
# 1523 vpn_detected: 0.20
# 1234 new_location: 0.25
# 892 unusual_time: 0.15
# 456 proxy_detected: 0.18
```

**Step 2: Drill Down by Risk Factor**

**VPN Detected (Most Common)**

```bash
# Get VPN IPs causing high risk scores
kubectl logs -n identity deployment/identity-idp | \
  grep '"vpn_detected"' | \
  jq -r '.metadata.ip_address' | \
  sort | uniq -c | sort -rn | head -n 20

# Check if these are legitimate corporate VPNs
# Cross-reference with IT network team's VPN CIDR ranges
```

**New Location (Second Most Common)**

```bash
# Get location patterns
kubectl logs -n identity deployment/identity-idp | \
  grep '"new_location"' | \
  jq -r '"\(.metadata.country_code)/\(.metadata.city)"' | \
  sort | uniq -c | sort -rn | head -n 20

# Example output:
# 523 GB/London
# 412 US/New York
# 301 DE/Berlin

# Cross-reference with user base geography
# Are these legitimate business locations?
```

**Step 3: Baseline Data Quality Check**

```bash
# Check baseline data coverage
kubectl exec -n identity deployment/identity-idp -- \
  psql -c "SELECT
    user_id,
    baseline_age_days,
    event_count,
    known_locations_count,
    known_devices_count
  FROM user_behavioral_baselines
  WHERE event_count < 10
  ORDER BY event_count ASC
  LIMIT 20;"

# Low event counts → low confidence → higher risk scores
```

**Step 4: Adjust Risk Factor Weights**

**Guidelines:**

- If >10% of legitimate users affected by a factor, reduce weight by 0.05
- If >20% affected, reduce weight by 0.10
- If factor is noise (e.g., corporate VPNs), add trusted exceptions
- Always simulate before deploying

**Example Adjustment:**

```yaml
# Original weights
risk_factors:
  vpn_detected:
    weight: 0.20
  new_location:
    weight: 0.25
  proxy_detected:
    weight: 0.18
  unusual_time:
    weight: 0.15

# Adjusted weights (after finding 15% of users on corporate VPNs)
risk_factors:
  vpn_detected:
    weight: 0.10  # Reduced from 0.20
    trusted_vpn_cidrs:
      - 10.0.0.0/8
  new_location:
    weight: 0.20  # Reduced from 0.25 (business travel common)
  proxy_detected:
    weight: 0.18  # Unchanged
  unusual_time:
    weight: 0.15  # Unchanged
```

---

## Policy Rollback

### Emergency Rollback Procedure

**When to Rollback:**

- Critical alerts firing (HighBlockedOperationRate, CriticalStepUpRate)
- >10% increase in support tickets
- User-facing service degradation
- Policy load failures

**Step 1: Immediate Rollback (Kubernetes)**

```bash
# Rollback to previous configmap version
kubectl rollout undo deployment/identity-idp -n identity-production

# Verify rollback
kubectl rollout status deployment/identity-idp -n identity-production

# Confirm metrics return to normal (check Grafana)
```

**Step 2: Rollback Verification**

```bash
# Check step-up rate (should decrease)
kubectl logs -n identity deployment/identity-idp | \
  grep "step_up_triggered" | \
  tail -n 100 | \
  jq -r .operation | \
  sort | uniq -c

# Check blocked operation rate (should decrease)
kubectl logs -n identity deployment/identity-idp | \
  grep "operation_blocked" | \
  tail -n 100 | \
  jq -r .operation | \
  sort | uniq -c
```

**Step 3: Post-Rollback Analysis**

```bash
# Extract problematic policy changes
git diff configs/identity/policies/risk_scoring.yml
git diff configs/identity/policies/step_up.yml

# Run simulation with rolled-back policies
go run ./cmd/cicd adaptive-sim \
  --risk-scoring configs/identity/policies/risk_scoring.yml \
  --step-up configs/identity/policies/step_up.yml \
  --logs production_logs_incident.json \
  --output simulation_results_rollback/

# Identify what went wrong
jq '.recommendations' simulation_results_rollback/simulation_report.json
```

**Step 4: Fix and Re-Deploy**

```bash
# Fix policy issues
vim configs/identity/policies/risk_scoring.yml

# Re-run simulation
go run ./cmd/cicd adaptive-sim ...

# Deploy to staging first
kubectl apply -f configs/identity/policies/ -n identity-staging

# Monitor staging for 24h before re-deploying to production
```

---

## Incident Response

### Incident Type 1: High Step-Up Rate (>30%)

**Symptom:** Prometheus alert `CriticalStepUpRate` firing

**Step 1: Assess Impact**

```bash
# Check current step-up rate
kubectl logs -n identity deployment/identity-idp | \
  grep "step_up_triggered" | \
  tail -n 1000 | \
  jq -r .operation | \
  sort | uniq -c

# Check affected operations
# If all operations affected → policy issue
# If specific operation → operation-specific issue
```

**Step 2: Identify Root Cause**

| Scenario | Indicators | Root Cause |
|----------|-----------|------------|
| All operations affected | Sudden policy change, recent deployment | Policy too strict |
| Specific operation (e.g., transfer_funds) | High-risk operation, increased usage | Policy correctly applying step-up |
| Time-based pattern (e.g., night hours) | Unusual_time risk factor high | Time window too narrow |
| Location-based pattern | New_location risk factor high | User base traveling |

**Step 3: Immediate Mitigation**

**Option 1: Temporary Policy Adjustment**

```yaml
# Emergency hotfix: Reduce step-up threshold for high-risk operations
operations:
  transfer_funds:
    risk_levels:
      high:
        required_level: mfa  # Temporarily reduce from strong_mfa
```

**Option 2: Emergency Rollback**

```bash
# Rollback to previous policy version
kubectl rollout undo deployment/identity-idp -n identity-production
```

**Step 4: Long-Term Fix**

```bash
# Analyze last 7 days of production data
# Run simulation with adjusted policies
# Deploy to staging → production
```

### Incident Type 2: High Blocked Operation Rate (>10%)

**Symptom:** Prometheus alert `CriticalBlockedOperationRate` firing

**Step 1: Assess Severity**

```bash
# Check blocked operation types
kubectl logs -n identity deployment/identity-idp | \
  grep "operation_blocked" | \
  tail -n 1000 | \
  jq -r '{operation: .operation, risk_level: .risk_level}' | \
  sort | uniq -c

# Example output:
# 523 {"operation":"transfer_funds","risk_level":"critical"}
# 201 {"operation":"update_profile","risk_level":"critical"}
```

**Step 2: Determine Attack vs False Positive**

| Indicator | Attack | False Positive |
|-----------|--------|----------------|
| Risk level | 90%+ critical | Mixed (high/critical) |
| Source IPs | Concentrated (few IPs) | Distributed (many IPs) |
| User accounts | New accounts | Established accounts |
| Time pattern | Sudden spike | Gradual increase |
| Geographic | High-risk countries | Normal business locations |

**Step 3: Response Action**

**If Attack:**

```bash
# Enable rate limiting
kubectl patch configmap identity-config -n identity-production \
  --type merge \
  -p '{"data":{"rate_limit_enabled":"true"}}'

# Add IP blocklists
kubectl create configmap identity-ip-blocklist \
  --from-file=blocklist.txt \
  -n identity-production
```

**If False Positive:**

```bash
# Emergency policy adjustment
# Reduce critical risk threshold or adjust risk factor weights
kubectl apply -f configs/identity/policies/emergency_fix.yml
```

### Incident Type 3: Policy Load Failures

**Symptom:** Prometheus alert `PolicyLoadFailures` firing

**Step 1: Check Policy File Syntax**

```bash
# Validate YAML syntax
yamllint configs/identity/policies/risk_scoring.yml
yamllint configs/identity/policies/step_up.yml

# Check for parsing errors in logs
kubectl logs -n identity deployment/identity-idp | grep "policy_load_error"
```

**Step 2: Common Policy Errors**

| Error | Cause | Fix |
|-------|-------|-----|
| `unknown field` | Extra fields in YAML | Remove invalid fields |
| `cannot unmarshal` | Type mismatch (e.g., string vs float) | Fix data types |
| `file not found` | Incorrect file path | Fix configmap mount paths |
| `permission denied` | File permissions | Fix file mode (should be 0644) |

**Step 3: Emergency Fallback**

```bash
# Use default embedded policies
kubectl set env deployment/identity-idp -n identity-production \
  USE_DEFAULT_POLICIES=true

# Restart pods
kubectl rollout restart deployment/identity-idp -n identity-production
```

---

## Monitoring Playbook

### Daily Health Check

**Run every morning (automated via cron).**

```bash
# Check key metrics from Grafana API
curl -H "Authorization: Bearer $GRAFANA_API_KEY" \
  "https://grafana.prod.example.com/api/datasources/proxy/1/api/v1/query?query=identity_stepup_triggered_total" | \
  jq '.data.result'

# Check for active alerts
curl -H "Authorization: Bearer $GRAFANA_API_KEY" \
  "https://grafana.prod.example.com/api/alerts" | \
  jq '.[] | select(.state == "alerting")'
```

**Checklist:**

- [ ] Step-up rate <15%
- [ ] Blocked operation rate <5%
- [ ] No active critical alerts
- [ ] Policy evaluation latency p95 <500ms
- [ ] Confidence scores avg >0.3
- [ ] No policy load errors in last 24h

### Weekly Review

**Run every Monday (manual review).**

```bash
# Export last 7 days of auth logs
kubectl logs -n identity deployment/identity-idp --since=168h > logs_last_7d.json

# Run simulation with current production policies
go run ./cmd/cicd adaptive-sim \
  --risk-scoring configs/identity/policies/risk_scoring.yml \
  --step-up configs/identity/policies/step_up.yml \
  --logs logs_last_7d.json \
  --output weekly_review/

# Review recommendations
cat weekly_review/simulation_report.json | jq '.recommendations'
```

**Checklist:**

- [ ] Simulation metrics match production metrics (±5%)
- [ ] No trending increase in step-up/blocked rates
- [ ] User support tickets <10 auth-related per week
- [ ] Risk factor weights still appropriate
- [ ] Baseline data quality healthy (avg confidence >0.3)

### Monthly Policy Optimization

**Run first week of month (quarterly for production).**

```bash
# Export last 30 days of auth logs
kubectl logs -n identity deployment/identity-idp --since=720h > logs_last_30d.json

# Run comprehensive simulation
go run ./cmd/cicd adaptive-sim \
  --risk-scoring configs/identity/policies/risk_scoring.yml \
  --step-up configs/identity/policies/step_up.yml \
  --logs logs_last_30d.json \
  --output monthly_review/

# Analyze trends
jq '{
  step_up_rate_trend: .metrics.step_up_rate,
  blocked_rate_trend: .metrics.blocked_rate,
  risk_distribution_trend: .metrics.risk_distribution
}' monthly_review/simulation_report.json
```

**Optimization Goals:**

- Reduce false positive rate by 10% quarter-over-quarter
- Maintain step-up rate <10%
- Maintain blocked rate <2%
- Improve user experience (fewer support tickets)

---

## Common Issues

### Issue 1: VPN False Positives

**Symptom:** Users on corporate VPNs receiving unnecessary step-up prompts

**Diagnosis:**

```bash
# Identify VPN CIDR ranges causing false positives
kubectl logs -n identity deployment/identity-idp | \
  grep '"vpn_detected":true' | \
  jq -r '.metadata.ip_address' | \
  sort | uniq -c | sort -rn
```

**Solution:**

```yaml
# Add trusted VPN exceptions
risk_factors:
  vpn_detected:
    weight: 0.20
    trusted_vpn_cidrs:
      - 10.0.0.0/8      # Corporate VPN range
      - 172.16.0.0/12   # Corporate VPN range
      - 192.168.0.0/16  # Corporate VPN range
    trusted_vpn_weight: 0.05  # Reduced weight for trusted VPNs
```

### Issue 2: Travel False Positives

**Symptom:** Users traveling for business receiving excessive step-up prompts

**Diagnosis:**

```bash
# Identify common business travel locations
kubectl logs -n identity deployment/identity-idp | \
  grep '"new_location":true' | \
  jq -r '"\(.metadata.country_code)/\(.metadata.city)"' | \
  sort | uniq -c | sort -rn
```

**Solution:**

```yaml
# Reduce new_location weight
risk_factors:
  new_location:
    weight: 0.20  # Reduced from 0.25

  # Or add trusted location exceptions
  trusted_locations:
    - country_code: GB
      city: London
    - country_code: US
      city: New York
    - country_code: DE
      city: Berlin
```

### Issue 3: New Device False Positives

**Symptom:** Users with new devices (phone upgrades, laptop replacements) receiving step-up prompts

**Diagnosis:**

```bash
# Check device age distribution
kubectl logs -n identity deployment/identity-idp | \
  grep '"device_age_days"' | \
  jq -r '.risk_factors.device_age_days' | \
  sort -n | uniq -c
```

**Solution:**

```yaml
# Reduce device_age_days weight for devices <30 days old
risk_factors:
  device_age_days:
    weight: 0.10
    new_device_threshold_days: 30
    new_device_weight: 0.05  # Reduced weight for new devices

# Or implement device trust transfer
# When user registers new device from trusted location with MFA
```

### Issue 4: Low Confidence Scores for New Users

**Symptom:** New users receiving high risk scores due to insufficient baseline data

**Diagnosis:**

```bash
# Check baseline data for new users
kubectl exec -n identity deployment/identity-idp -- \
  psql -c "SELECT user_id, baseline_age_days, event_count
           FROM user_behavioral_baselines
           WHERE event_count < 10
           ORDER BY event_count ASC
           LIMIT 20;"
```

**Solution:**

```yaml
# Reduce risk score contribution for users with low baseline data
risk_scoring:
  confidence_thresholds:
    low_confidence_threshold: 0.3  # Users with confidence <0.3
    low_confidence_score_reduction: 0.15  # Reduce risk score by 0.15

# Or require minimum baseline before adaptive policies apply
adaptive_auth:
  min_baseline_events: 5  # Don't apply adaptive policies until 5 events
  min_baseline_age_days: 7  # Don't apply adaptive policies until 7 days
```

---

## Escalation Procedures

### Level 1: Operations Team (First Response)

**Responsibilities:**

- Monitor Grafana dashboards and Prometheus alerts
- Execute runbook procedures for common issues
- Rollback policies if critical alerts fire
- Collect logs and metrics for escalation

**Contact:**

- Slack: #identity-ops
- PagerDuty: identity-ops-oncall

### Level 2: Identity Platform Team (Engineering)

**Responsibilities:**

- Investigate complex policy issues
- Debug risk factor calculation bugs
- Optimize policy evaluation performance
- Implement policy adjustments and hotfixes

**Contact:**

- Slack: #identity-platform
- PagerDuty: identity-platform-oncall

### Level 3: Security Team (Critical Incidents)

**Responsibilities:**

- Analyze attack patterns
- Approve emergency policy changes
- Coordinate with SOC for security incidents
- Review compliance implications

**Contact:**

- Slack: #security-incident
- PagerDuty: security-oncall

---

## Appendices

### A. Useful Commands

```bash
# Check current policy versions
kubectl get configmap identity-policies -n identity-production -o yaml

# View recent policy reload events
kubectl logs -n identity deployment/identity-idp | grep "policy_reload"

# Export authentication metrics
kubectl logs -n identity deployment/identity-idp | \
  grep "auth_attempt" | \
  jq -r '[.timestamp, .user_id, .operation, .risk_score, .decision] | @csv' > metrics.csv

# Run simulation with custom time range
go run ./cmd/cicd adaptive-sim \
  --logs <(kubectl logs -n identity deployment/identity-idp --since=24h) \
  --output /tmp/simulation/
```

### B. Policy File Templates

See `configs/identity/policies/` for reference templates:

- `risk_scoring.yml`: Risk factor weights and thresholds
- `step_up.yml`: Step-up authentication rules
- `adaptive_auth.yml`: Global adaptive auth configuration

### C. Support Resources

- **Documentation**: `docs/02-identityV2/task-13-adaptive-engine-COMPLETE.md`
- **Code Repository**: `internal/identity/idp/userauth/`
- **Monitoring Dashboard**: `configs/observability/grafana/adaptive-auth-dashboard.json`
- **Alert Rules**: `configs/observability/prometheus/adaptive-auth-alerts.yml`

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-01-28 | Identity Platform Team | Initial release |
