# Secrets Directory

This directory contains Docker secrets for the CA infrastructure.

**WARNING**: Do not commit actual secrets to version control.

## Required Secrets

Create the following secret files before running Docker Compose:

### db_password.secret

Contains the PostgreSQL database password.

```bash
# Generate a secure password
openssl rand -base64 32 > db_password.secret
```

### ca_key_password.secret

Contains the password for CA private keys.

```bash
# Generate a secure password
openssl rand -base64 32 > ca_key_password.secret
```

### ocsp_key_password.secret

Contains the password for OCSP responder private key.

```bash
# Generate a secure password
openssl rand -base64 32 > ocsp_key_password.secret
```

## Security Notes

- All secret files should have restrictive permissions (600)
- Use different passwords for each secret
- Rotate secrets regularly
- Consider using a secrets manager for production
