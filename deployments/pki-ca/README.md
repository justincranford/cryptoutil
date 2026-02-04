# CA Infrastructure Deployment

This directory contains deployment configurations for the Certificate Authority infrastructure.

## Contents

- `compose/` - Docker Compose manifests for local development and testing
- `kubernetes/` - Kubernetes manifests for production deployment

## Quick Start

### Docker Compose (Development)

```bash
# Start the CA infrastructure
docker compose -f deployments/ca/compose/compose.yml up -d

# View logs
docker compose -f deployments/ca/compose/compose.yml logs -f

# Stop the infrastructure
docker compose -f deployments/ca/compose/compose.yml down
```

### Kubernetes (Production)

```bash
# Create namespace
kubectl create namespace ca-system

# Apply configurations
kubectl apply -f deployments/ca/kubernetes/

# Check status
kubectl get pods -n ca-system
```

## Architecture

The CA infrastructure consists of:

1. **Root CA** - Offline root certificate authority (manual operation only)
2. **Intermediate CA** - Online intermediate CA for signing
3. **Issuing CA** - Online issuing CA for end-entity certificates
4. **OCSP Responder** - Online Certificate Status Protocol responder
5. **CRL Server** - Certificate Revocation List distribution point
6. **Database** - Certificate database (PostgreSQL)

## Security Considerations

- Root CA should remain offline except during key ceremonies
- HSM integration recommended for production deployments
- Network segmentation between CA tiers
- Audit logging to external SIEM
- Regular backup of CA databases and keys
