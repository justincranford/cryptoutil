# Hardware Credential Administrator Guide

## Overview

This guide provides comprehensive operational guidance for managing hardware-based authentication (smart cards, FIDO keys, TPMs) in production environments. Administrators are responsible for day-0 provisioning, lifecycle management, break-glass recovery, and troubleshooting hardware credential failures.

---

## Prerequisites

- **Access Control**: Administrators must have elevated privileges to manage hardware credential lifecycles
- **CLI Installation**: Install `hardware-cred` CLI utility on administrative workstations
- **Database Access**: Ensure connectivity to identity database for credential operations
- **Break-Glass Credentials**: Maintain offline emergency access credentials (password-based)

---

## Day-0 Provisioning

### Initial Setup

**Goal**: Configure hardware credential infrastructure before user enrollment

1. **Install Required Dependencies**:

```bash
# Install hardware-cred CLI
go install cryptoutil/cmd/identity/hardware-cred@latest

# Verify installation
hardware-cred help
```

2. **Configure Database Connection**:

Edit configuration file (e.g., `/etc/identity/config.yml`):

```yaml
database:
  dsn: "postgres://user:password@localhost:5432/identity_db"

hardware:
  max_pin_retries: 3
  auth_timeout: 30s
  device_poll_interval: 1s
```

3. **Provision Administrator Hardware Credentials**:

```bash
# Enroll admin smart card
hardware-cred enroll \
  -user-id <ADMIN_UUID> \
  -device-name "Admin Smart Card" \
  -credential-type smart_card

# Enroll admin FIDO key
hardware-cred enroll \
  -user-id <ADMIN_UUID> \
  -device-name "Admin YubiKey 5C" \
  -credential-type security_key
```

4. **Verify Enrollment**:

```bash
# List admin credentials
hardware-cred list -user-id <ADMIN_UUID>
```

---

## User Enrollment Workflows

### Self-Service Enrollment

**Goal**: Enable end users to enroll hardware credentials via CLI or web interface

**CLI Enrollment**:

```bash
# User enrolls FIDO key
hardware-cred enroll \
  -user-id <USER_UUID> \
  -device-name "YubiKey 5 NFC" \
  -credential-type security_key
```

**Web Enrollment** (integration with WebAuthn):

1. User navigates to `/account/security/webauthn/register`
2. User clicks "Enroll New Device"
3. Browser prompts for platform authenticator (Windows Hello, TouchID) or external FIDO key
4. Server validates WebAuthn credential and stores in database
5. Audit event logged: `CREDENTIAL_ENROLLED`

### Bulk Enrollment (Corporate Devices)

**Goal**: Provision smart cards for large user populations

**CSV-Based Bulk Enrollment**:

```csv
user_id,device_name,credential_type
01930de8-c123-7890-abcd-ef1234567890,Employee Smart Card 001,smart_card
01930de8-c456-7890-abcd-ef1234567891,Employee Smart Card 002,smart_card
```

```bash
# Bulk enroll from CSV
while IFS=, read -r user_id device_name cred_type; do
  hardware-cred enroll \
    -user-id "$user_id" \
    -device-name "$device_name" \
    -credential-type "$cred_type"
done < enrollment.csv
```

---

## Lifecycle Management

### Credential Renewal/Rotation

**Goal**: Rotate cryptographic key material periodically (e.g., annually)

```bash
# Renew credential (rotates keys, preserves device name)
hardware-cred renew \
  -credential-id <CREDENTIAL_ID> \
  -device-name "YubiKey 5C (Renewed 2025)"
```

**Automated Renewal** (cron job example):

```bash
# Renew credentials expiring in next 30 days
0 2 * * * /usr/local/bin/renew-expiring-credentials.sh
```

### Credential Revocation

**Goal**: Disable lost/stolen hardware devices

```bash
# Revoke credential immediately
hardware-cred revoke -credential-id <CREDENTIAL_ID>
```

**Audit Trail Verification**:

```bash
# Check audit logs for revocation event
grep "CREDENTIAL_REVOKED" /var/log/identity/audit.log | grep <USER_ID>
```

### Inventory Tracking

**Goal**: Maintain visibility into all enrolled hardware credentials

```bash
# Generate inventory report
hardware-cred inventory > hardware-inventory-$(date +%Y%m%d).txt
```

**Inventory Report Contents**:
- Total credentials by type (smart_card, security_key, passkey)
- Credentials by user
- Enrollment dates and last usage timestamps
- Devices approaching renewal deadline

---

## Break-Glass Recovery Procedures

### Scenario 1: User Lost Hardware Device

**Goal**: Restore user access without hardware credential

**Steps**:

1. **Verify User Identity** (offline process: in-person verification, multi-factor verification via phone)

2. **Revoke Lost Credential**:

```bash
# Revoke lost/stolen device
hardware-cred revoke -credential-id <LOST_CREDENTIAL_ID>
```

3. **Provision Temporary Password**:

```bash
# Issue temporary password (password authentication must be enabled as fallback)
identity-cli user reset-password -user-id <USER_UUID> --temporary
```

4. **Re-Enroll New Hardware Device**:

```bash
# After user obtains replacement device
hardware-cred enroll \
  -user-id <USER_UUID> \
  -device-name "Replacement YubiKey" \
  -credential-type security_key
```

5. **Audit Logging**:

```bash
# Verify break-glass event logged
grep "BREAK_GLASS_RECOVERY" /var/log/identity/audit.log | grep <USER_UUID>
```

### Scenario 2: Hardware Device Malfunction

**Goal**: Troubleshoot device not responding during authentication

**Symptoms**:
- User reports "Device unresponsive" error
- Authentication times out (30+ seconds)
- Device not detected by system

**Troubleshooting Steps**:

1. **Verify Device Connectivity**:

```bash
# Check USB device detection (Linux)
lsusb | grep -i yubi

# Check smart card reader detection (Windows)
certutil -scinfo
```

2. **Test Device with CLI**:

```bash
# Attempt enrollment to verify device functionality
hardware-cred enroll \
  -user-id <TEST_USER_UUID> \
  -device-name "Device Test" \
  -credential-type security_key
```

3. **Check Error Logs**:

```bash
# Review hardware authentication errors
tail -f /var/log/identity/hardware-auth.log | grep ERROR
```

4. **Common Resolutions**:
   - **Device Locked (PIN retries exhausted)**: Reset device PIN via manufacturer tool (e.g., `ykman piv reset` for YubiKey)
   - **Driver Issues**: Update smart card reader drivers or USB controller firmware
   - **Certificate Expiration**: Renew device certificates (smart cards)

### Scenario 3: Administrator Lockout

**Goal**: Recover administrator access when all hardware credentials fail

**Prerequisites**:
- Offline emergency access account (password-based)
- Physical access to server or secure console

**Steps**:

1. **Use Break-Glass Account**:

```bash
# Login with emergency credentials
identity-cli login --username breakglass-admin --password <SECURE_PASSWORD>
```

2. **Provision New Admin Hardware Credential**:

```bash
# Enroll replacement admin device
hardware-cred enroll \
  -user-id <ADMIN_UUID> \
  -device-name "Emergency Admin YubiKey" \
  -credential-type security_key
```

3. **Rotate Break-Glass Password**:

```bash
# Change break-glass password after use
identity-cli user reset-password -user-id <BREAKGLASS_UUID>
```

4. **Audit Break-Glass Usage**:

```bash
# Log break-glass account activation
grep "BREAK_GLASS_LOGIN" /var/log/identity/audit.log
```

---

## PIN Management

### PIN Reset Procedures

**Goal**: Reset device PIN when user forgets or exhausts retry limit

**Smart Card PIN Reset**:

```bash
# Requires administrative smart card management tool
pkcs15-init --erase-card --reader 0
pkcs15-init --create-pkcs15 --pin 123456 --puk 12345678 --reader 0
```

**FIDO Key PIN Reset** (YubiKey example):

```bash
# Reset YubiKey PIN (erases all PIV credentials)
ykman piv reset

# Set new PIN
ykman piv change-pin
```

**Security Considerations**:
- PIN resets erase all stored credentials
- User must re-enroll device after PIN reset
- Audit log must record PIN reset events

### PIN Policy Configuration

**Best Practices**:
- Minimum PIN length: 8 characters
- Maximum retry attempts: 3
- PIN complexity: Require alphanumeric + special characters
- PIN expiration: 90 days for smart cards

---

## Device Replacement Workflows

### Replacing Lost/Stolen Devices

**Goal**: Securely transition user to replacement hardware

**Workflow**:

1. User reports lost device to helpdesk
2. Administrator revokes lost credential: `hardware-cred revoke -credential-id <OLD_CRED_ID>`
3. Administrator verifies user identity (in-person or video call)
4. User obtains replacement device (issued by IT or purchased)
5. User enrolls replacement device: `hardware-cred enroll -user-id <USER_UUID> ...`
6. Audit trail captures revocation + re-enrollment events

### Replacing Malfunctioning Devices

**Goal**: Swap faulty device with minimal downtime

**Workflow**:

1. User enrolls backup device before revoking primary: `hardware-cred enroll -user-id <USER_UUID> -device-name "Backup YubiKey"`
2. User verifies backup device works (test authentication)
3. Administrator revokes malfunctioning device: `hardware-cred revoke -credential-id <FAULTY_CRED_ID>`
4. User continues using backup device as primary

**Recommendation**: Encourage users to enroll multiple devices (primary + backup)

---

## Troubleshooting Guide

### Common Errors

#### Error: "Device removed during authentication"

**Cause**: User physically removed device before authentication completed

**Resolution**:
- Educate user to leave device inserted until authentication success/failure message appears
- Increase authentication timeout if device requires extended processing time

---

#### Error: "PIN retry limit exhausted"

**Cause**: User entered incorrect PIN 3+ times

**Resolution**:
1. Reset device PIN (see PIN Reset Procedures above)
2. Re-enroll device after PIN reset
3. Provide user training on PIN management

---

#### Error: "Hardware device locked"

**Cause**: Device locked due to security policy (e.g., too many failed PIN attempts)

**Resolution**:
- Smart cards: Use PUK (PIN Unblocking Key) to unlock
- FIDO keys: Perform factory reset (erases credentials)

---

#### Error: "Authentication timeout"

**Cause**: Device not responding within configured timeout (default 30s)

**Resolution**:
1. Check device connectivity (USB connection, smart card reader)
2. Restart smart card reader service:
   ```bash
   # Linux
   systemctl restart pcscd

   # Windows
   Restart-Service -Name "Smart Card"
   ```
3. Increase timeout in configuration if device legitimately requires longer processing

---

### Diagnostic Commands

#### Check Hardware Device Status

```bash
# List USB devices (Linux)
lsusb -v | grep -A 10 -i yubi

# Check smart card readers (Windows PowerShell)
Get-PnpDevice | Where-Object {$_.FriendlyName -like "*Smart Card*"}

# Verify FIDO device info (YubiKey)
ykman info
```

#### Test Device Cryptographic Operations

```bash
# Test FIDO key signature operation
ykman piv keys generate 9a /tmp/test-pubkey.pem

# Verify smart card certificate
pkcs11-tool --module opensc-pkcs11.so --list-objects --type cert
```

#### Audit Log Analysis

```bash
# Find failed hardware authentication attempts
grep "hardware authentication error" /var/log/identity/audit.log | tail -20

# Count credentials by type
grep "CREDENTIAL_ENROLLED" /var/log/identity/audit.log | \
  awk '{print $NF}' | sort | uniq -c
```

---

## Compliance and Audit

### Required Audit Events

**Lifecycle Events**:
- `CREDENTIAL_ENROLLED`: User enrolls new hardware credential
- `CREDENTIAL_RENEWED`: User rotates credential keys
- `CREDENTIAL_REVOKED`: Administrator or user revokes credential
- `CREDENTIAL_USAGE`: User authenticates with hardware credential (captured by authentication log)

**Break-Glass Events**:
- `BREAK_GLASS_LOGIN`: Administrator uses emergency access account
- `BREAK_GLASS_RECOVERY`: User recovers access without hardware credential
- `PIN_RESET`: Device PIN reset performed

### Compliance Reporting

**Generate Hardware Credential Compliance Report**:

```bash
# Monthly report: all hardware events
grep -E "CREDENTIAL_|BREAK_GLASS_|PIN_RESET" \
  /var/log/identity/audit.log | \
  awk '{print $1, $2, $5, $NF}' > \
  compliance-report-$(date +%Y%m).txt
```

**Required Retention**:
- Audit logs: 7 years (regulatory requirement for financial institutions)
- Credential metadata: Duration of user account lifecycle + 1 year

---

## Best Practices

### Enrollment

- Require multiple device enrollment per user (primary + backup)
- Enforce device naming conventions for inventory tracking
- Use corporate-issued devices for high-privilege accounts

### Lifecycle

- Rotate credentials annually
- Revoke credentials immediately upon user offboarding
- Maintain inventory reports for compliance audits

### Break-Glass

- Test break-glass procedures quarterly
- Rotate break-glass credentials after each use
- Limit break-glass account access to 2-3 administrators

### Security

- Enable PIN policies with complexity requirements
- Use hardware-backed keys where possible (TPMs for certificates)
- Disable password authentication for administrators (hardware-only)

---

## References

- Hardware Credential CLI Documentation: `hardware-cred help`
- WebAuthn Integration Guide: `docs/webauthn/browser-compatibility.md`
- Task 11 MFA Chain Configuration: `docs/02-identityV2/task-11-mfa-stabilization-COMPLETE.md`
- Task 13 Adaptive Policies: `docs/02-identityV2/task-13-adaptive-auth-COMPLETE.md`
