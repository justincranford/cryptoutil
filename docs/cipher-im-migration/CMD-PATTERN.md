# Command Line Patterns for Cryptoutil Services

## Overview

This document describes aspirational patterns for organizing cryptoutil's command-line executables for suite, products, and services.

## Patterns

### 1. Product-Service Pattern

**Description**: Separate executables for each product-service combination that routes to service entry point, then subcommand.

Every service executable supports these SUBCOMMANDs: server, client, health, livez, readyz, shutdown, init, compose, demo, e2e.

**Call Flow**:

```
cmd/PRODUCT-SERVICE/main.go
  → internal/app/PRODUCT/SERVICE/SERVICE.go SUBCOMMAND
```

**Examples**:

- `cmd/cipher-im/main.go server` → `internal/app/cipher/im/im.go server`
- `cmd/jose-ja/main.go client` → `internal/app/jose/ja/ja.go client`
- `cmd/pki-ca/main.go health` → `internal/app/pki/ca/ca.go health`
- `cmd/identity-authz/main.go livez` → `internal/app/identity/authz/authz.go livez`
- `cmd/identity-idp/main.go readyz` → `internal/app/identity/idp/idp.go readyz`
- `cmd/sm-kms/main.go shutdown` → `internal/app/sm/kms/kms.go shutdown`

---

### 2. Product Pattern

**Description**: Separate executables per product that routes to product entry point, then service entry point, then subcommand.

Every product executable supports these product-level SUBCOMMANDs, that recurse to each of its services: health, readyz, livez, shutdown, init, compose, demo, e2e.

**Call Flow**:

```
cmd/PRODUCT/main.go
  → internal/app/PRODUCT/PRODUCT.go SERVICE SUBCOMMAND
    → [1-to-1] internal/app/PRODUCT/SERVICE/SERVICE.go SUBCOMMAND
  → internal/app/PRODUCT/PRODUCT.go SUBCOMMAND
    → [1-to-N] internal/app/PRODUCT/*/SERVICE.go SUBCOMMAND
```

**Examples (1-to-1)**:

- `cmd/cipher/main.go im server` → `internal/app/cipher/cipher.go im server` → `internal/app/cipher/im/im.go server`
- `cmd/jose/main.go ja client` → `internal/app/jose/jose.go ja client` → `internal/app/jose/ja/ja.go client`
- `cmd/pki/main.go ca health` → `internal/app/pki/pki.go ca health` → `internal/app/pki/ca/ca.go health`
- `cmd/identity/main.go authz livez` → `internal/app/identity/identity.go authz livez` → `internal/app/identity/authz/authz.go livez`
- `cmd/identity/main.go idp readyz` → `internal/app/identity/identity.go idp readyz` → `internal/app/identity/idp/idp.go readyz`
- `cmd/sm/main.go kms shutdown` → `internal/app/sm/sm.go kms shutdown` → `internal/app/sm/kms/kms.go shutdown`

**Examples (1-to-N)**:

- `cmd/cipher/main.go init` → `internal/app/cipher/cipher.go init` → `internal/app/cipher/*/*.go init`
- `cmd/jose/main.go compose` → `internal/app/jose/jose.go compose` → `internal/app/jose/*/*.go compose`
- `cmd/pki/main.go demo` → `internal/app/pki/pki.go demo` → `internal/app/pki/*/*.go demo`
- `cmd/identity/main.go e2e` → `internal/app/identity/identity.go main` → `internal/app/identity/*/*.go e2e`
- `cmd/sm/main.go kms health` → `internal/app/sm/sm.go health` → `internal/app/sm/*/*.go health`

---

### 3. Suite Pattern

**Description**: A single unified `cryptoutil` executable that routes to internal cryptoutil suite, then product, then service, then subcommand.

Cryptoutil suite executable supports these suite-level SUBCOMMANDs, that recurse to each of its products, and then their services: health, readyz, livez, shutdown, init, compose, demo, e2e.

**Call Flow**:

```
cmd/cryptoutil/main.go PRODUCT SERVICE SUBCOMMAND
  → internal/app/cryptoutil/cryptoutil.go PRODUCT SERVICE SUBCOMMAND
    → internal/app/PRODUCT/PRODUCT.go SERVICE SUBCOMMAND
      → [1-to-1] internal/app/PRODUCT/SERVICE/SERVICE.go SUBCOMMAND
    → internal/app/PRODUCT/PRODUCT.go SUBCOMMAND
      → [1-to-N] internal/app/PRODUCT/*/SERVICE.go SUBCOMMAND
```

**Examples (1-to-1)**:

- `cmd/cryptoutil/main.go cipher im server` → `internal/app/cryptoutil/cryptoutil.go cipher im server` → `internal/app/cipher/cipher.go im server` → `internal/app/cipher/im/im.go server`
- `cmd/cryptoutil/main.go jose ja client` → `internal/app/cryptoutil/cryptoutil.go jose ja client` → `internal/app/jose/jose.go ja client` → `internal/app/jose/ja/ja.go client`
- `cmd/cryptoutil/main.go pki ca health` → `internal/app/cryptoutil/cryptoutil.go pki ca health` → `internal/app/pki/pki.go ca health` → `internal/app/pki/ca/ca.go health`
- `cmd/cryptoutil/main.go identity authz livez` → `internal/app/cryptoutil/cryptoutil.go identity authz livez` → `internal/app/identity/identity.go authz livez` → `internal/app/identity/authz/authz.go livez`
- `cmd/cryptoutil/main.go identity idp readyz` → `internal/app/cryptoutil/cryptoutil.go identity idp readyz` → `internal/app/identity/identity.go idp readyz` → `internal/app/identity/idp/idp.go readyz`
- `cmd/cryptoutil/main.go sm kms shutdown` → `internal/app/cryptoutil/cryptoutil.go sm kms shutdown` → `internal/app/sm/sm.go kms shutdown` → `internal/app/sm/kms/kms.go shutdown`

**Examples (1-to-N)**:

- `cmd/cryptoutil/main.go init` → `internal/app/cryptoutil/cryptoutil.go init` → `internal/app/cipher/cipher.go init` → `internal/app/cipher/*/*.go init`
- `cmd/cryptoutil/main.go compose` → `internal/app/cryptoutil/cryptoutil.go compose` → `internal/app/jose/jose.go compose` → `internal/app/jose/*/*.go compose`
- `cmd/cryptoutil/main.go demo` → `internal/app/cryptoutil/cryptoutil.go demo` → `internal/app/pki/pki.go ca demo` → `internal/app/pki/*/*.go demo`
- `cmd/cryptoutil/main.go e2e` → `internal/app/cryptoutil/cryptoutil.go e2e` → `internal/app/identity/identity.go e2e` → `internal/app/identity/*/*.go e2e`
- `cmd/cryptoutil/main.go health` → `internal/app/cryptoutil/cryptoutil.go health` → `internal/app/sm/sm.go health` → `internal/app/sm/*/*.go health`

---

## Anti-Patterns

**Description**: No executables are allowed that correspond directly to subcommands.

**Service-Level Anti-Pattern Examples**:

- `cmd/cipher-im-server/main.go`: No allowed
- `cmd/jose-ja-client/main.go`: No allowed
- `cmd/pki-ca-health/main.go`: No allowed
- `cmd/identity-authz-livez/main.go`: No allowed
- `cmd/identity-idp-readyz/main.go`: No allowed
- `cmd/sm-kms-shutdown/main.go`: No allowed

**Product-Level Anti-Pattern Examples**:

- `cmd/cipher-server/main.go`: No allowed
- `cmd/jose-client/main.go`: No allowed
- `cmd/pki-health/main.go`: No allowed
- `cmd/identity-livez/main.go`: No allowed
- `cmd/sm-shutdown/main.go`: No allowed

**Suite-Level Anti-Pattern Examples**:

- `cmd/cryptoutil-health/main.go`: No allowed
- `cmd/cryptoutil-livez/main.go`: No allowed
- `cmd/cryptoutil-shutdown/main.go`: No allowed
- `cmd/cryptoutil-init/main.go`: No allowed
- `cmd/cryptoutil-compose/main.go`: No allowed
- `cmd/cryptoutil-demo/main.go`: No allowed
- `cmd/cryptoutil-e2e/main.go`: No allowed
- `cmd/cryptoutil-server/main.go`: No allowed
- `cmd/cryptoutil-client/main.go`: No allowed

---
