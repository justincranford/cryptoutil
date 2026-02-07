# Runbook: Certificate Expiring Soon

## Alert: CertificateExpiringSoon

**Severity**: Warning
**Alert Expression**: `(tls_certificate_not_after - time()) / 86400 < 30`
**Duration**: 0 (immediate)

## Description

TLS certificate will expire within 30 days.

## Impact

- Service disruption when certificate expires
- Trust errors for clients
- Potential security warnings

## Investigation Steps

### 1. Identify Expiring Certificates

```bash
# Check certificate expiry
openssl s_client -connect localhost:8080 -servername localhost 2>/dev/null | \
  openssl x509 -noout -dates

# Check all certificates in a directory
for cert in /path/to/certs/*.pem; do
  echo "=== $cert ==="
  openssl x509 -in "$cert" -noout -dates
done
```

### 2. Verify Certificate Chain

```bash
openssl s_client -connect localhost:8080 -servername localhost 2>/dev/null | \
  openssl x509 -noout -text | grep -A2 "Validity"
```

### 3. Check Certificate Source

```bash
# Internal CA
# Check CA service status

# External CA (Let's Encrypt)
certbot certificates

# HashiCorp Vault
vault read pki/cert/<serial>
```

## Resolution Steps

### Auto-Renewal (Let's Encrypt / ACME)

1. Check certbot timer:
   ```bash
   systemctl status certbot.timer
   ```
2. Force renewal:
   ```bash
   certbot renew --force-renewal
   ```
3. Restart services to pick up new cert

### Internal CA

1. Request new certificate:
   ```bash
   # Using cryptoutil CA
   curl -X POST https://pki-ca:8050/api/v1/certificates \
     -d '{"csr": "...", "ttl": "365d"}'
   ```
2. Deploy to services
3. Restart services

### Manual Certificate

1. Generate new CSR:
   ```bash
   openssl req -new -key server.key -out server.csr
   ```
2. Submit to CA
3. Install new certificate
4. Restart services

### HashiCorp Vault PKI

1. Check PKI role configuration
2. Issue new certificate:
   ```bash
   vault write pki/issue/cryptoutil common_name=server.example.com
   ```
3. Deploy and restart

## Escalation

- **30 days out**: Create ticket for renewal
- **14 days out**: Assign to engineer
- **7 days out**: Page on-call engineer
- **1 day out**: Emergency escalation

## Post-Incident

1. Implement automated certificate renewal
2. Add certificate monitoring to all services
3. Create certificate inventory
4. Document renewal procedures
5. Consider using short-lived certificates
