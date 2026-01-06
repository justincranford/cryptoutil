# Command Line Patterns for Cryptoutil Services

## Overview

This document describes aspirational patterns for organizing cryptoutil's command-line executables for suite, products, and services.

## Patterns

### 1. Suite Pattern

**Description**: A single unified `cryptoutil` executable that routes to internal cryptoutil suite, then product, then service, then subcommand.

**Call Flow**:

```
cmd/cryptoutil/main.go PRODUCT SERVICE SUBCOMMAND
  → internal/app/cryptoutil/cryptoutil.go PRODUCT SERVICE SUBCOMMAND
    → internal/app/PRODUCT/PRODUCT.go SERVICE SUBCOMMAND
      → internal/app/PRODUCT/SERVICE/SERVICE.go SUBCOMMAND
```

**Examples**:

- `cmd/cryptoutil/main.go cipher im server` → `internal/app/cryptoutil/cryptoutil.go cipher im server` → `internal/app/cipher/cipher.go im server` → `internal/app/cipher/im/im.go server`
- `cmd/cryptoutil/main.go jose ja client` → `internal/app/cryptoutil/cryptoutil.go jose ja client` → `internal/app/jose/jose.go ja client` → `internal/app/jose/ja/ja.go client`
- `cmd/cryptoutil/main.go pki ca health` → `internal/app/cryptoutil/cryptoutil.go pki ca health` → `internal/app/pki/pki.go ca health` → `internal/app/pki/ca/ca.go health`
- `cmd/cryptoutil/main.go identity authz livez` → `internal/app/cryptoutil/cryptoutil.go identity authz livez` → `internal/app/identity/identity.go authz livez` → `internal/app/identity/authz/authz.go livez`
- `cmd/cryptoutil/main.go identity idp readyz` → `internal/app/cryptoutil/cryptoutil.go identity idp readyz` → `internal/app/identity/identity.go idp readyz` → `internal/app/identity/idp/idp.go readyz`
- `cmd/cryptoutil/main.go sm kms shutdown` → `internal/app/cryptoutil/cryptoutil.go sm kms shutdown` → `internal/app/sm/sm.go kms shutdown` → `internal/app/sm/kms/kms.go shutdown`

Cryptoutil suite supports these subcommands: health, readyz, livez, shutdown, init.
Every product supports these subcommands: health, readyz, livez, shutdown, init.
Every service supports these subcommands: health, readyz, livez, shutdown, init, server, client.

---

### 2. Product Pattern

**Description**: Separate executables per product that routes to product entry point, then service entry point, then subcommand.

**Call Flow**:

```
cmd/PRODUCT/main.go
  → internal/app/PRODUCT/PRODUCT.go SERVICE SUBCOMMAND
    → internal/app/PRODUCT/SERVICE/SERVICE.go SUBCOMMAND
```

**Examples**:

- `cmd/cipher/main.go im server` → `internal/app/cipher/cipher.go im server` → `internal/app/cipher/im/im.go server`
- `cmd/jose/main.go ja client` → `internal/app/jose/ja.go ja client` → `internal/app/jose/ja/ja.go client`
- `cmd/pki/main.go ca health` → `internal/app/pki/pki.go ca health` → `internal/app/pki/ca/ca.go health`
- `cmd/identity/main.go authz livez` → `internal/app/identity/identity.go authz livez` → `internal/app/identity/authz/authz.go livez`
- `cmd/identity/main.go idp readyz` → `internal/app/identity/identity.go readyz` → `internal/app/identity/authz/authz.go readyz`
- `cmd/sm/main.go kms shutdown` → `internal/app/sm/sm.go kms shutdown` → `internal/app/sm/kms/kms.go shutdown`

---

### 3. Product-Service Pattern

**Description**: Separate executables for each product-service combination that routes to service entry point, then subcommand.

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
- `cmd/identity-idp/main.go readyz` → `internal/app/identity/authz/authz.go readyz`
- `cmd/sm-kms/main.go shutdown` → `internal/app/sm/kms/kms.go shutdown`

## Top-Level Commands

These commands are available regardless of the chosen pattern and provide orchestration capabilities:

### cryptoutil-compose

**Purpose**: Top-level orchestration command for managing suite, product, or service startup/shutdown/cleanup operations. Orchestrates underlying services based on the selected pattern (configuration must be provided). Optional for production deployments. Required for demonstrations and end-to-end testing.

**Call Flow**:

```
cmd/cryptoutil-compose/main.go
  → internal/app/cryptoutil-compose/cryptoutil-compose.go
    → compose(args) function
      → up|down|clean|status subcommand functions
```

**Examples**:

- `cryptoutil-compose all up` - Start all products, including all of their services
- `cryptoutil-compose all down` - Stop all products, including all of their services
- `cryptoutil-compose all status` - Show status for all products, including all of their services
- `cryptoutil-compose all clean` - Remove all containers and volumes for all products, including all of their services
- `cryptoutil-compose identity up` - Start all services for identity product
- `cryptoutil-compose identity-authz up` - Start identity-authz service only
- `cryptoutil-compose identity-idp up` - Start identity-idp service only
- `cryptoutil-compose identity-rp up` - Start identity-rp service only
- `cryptoutil-compose identity-rs up` - Start identity-rs service only
- `cryptoutil-compose identity-spa up` - Start identity-spa service only
- `cryptoutil-compose jose up` - Start all services for cipher product
- `cryptoutil-compose jose-ja up` - Start cipher-im service only
- `cryptoutil-compose pki up` - Start all services for cipher product
- `cryptoutil-compose pki-ca up` - Start cipher-im service only
- `cryptoutil-compose sm up` - Start all services for cipher product
- `cryptoutil-compose sm-kms up` - Start cipher-im service only
- `cryptoutil-compose cipher up` - Start all services for cipher product
- `cryptoutil-compose cipher-im up` - Start cipher-im service only

### cryptoutil-demo

**Purpose**: Interactive demonstration command that orchestrates cryptoutil-compose. It uses opinionated demonstration configuration data for manual demonstration or manual developer testing.

**Call Flow**:

```
cmd/cryptoutil-demo/main.go
  → internal/app/cryptoutil-demo/cryptoutil-demo.go
    → demo(args) function
      → start|test|stop subcommand functions
```

**Examples**:

- `cryptoutil-demo all up` - Start all products, including all of their services, with demo configs
- `cryptoutil-demo all down` - Stop all products, including all of their services, with demo configs
- `cryptoutil-demo all status` - Show status for all products, including all of their services, with demo configs
- `cryptoutil-demo all clean` - Remove all containers and volumes for all products, including all of their services, with demo configs
- `cryptoutil-demo identity up` - Start all services for identity product, with demo configs
- `cryptoutil-demo identity-authz up` - Start identity-authz service only, with demo configs
- `cryptoutil-demo identity-idp up` - Start identity-idp service only, with demo configs
- `cryptoutil-demo identity-rp up` - Start identity-rp service only, with demo configs
- `cryptoutil-demo identity-rs up` - Start identity-rs service only, with demo configs
- `cryptoutil-demo identity-spa up` - Start identity-spa service only, with demo configs
- `cryptoutil-demo jose up` - Start all services for cipher product, with demo configs
- `cryptoutil-demo jose-ja up` - Start cipher-im service only, with demo configs
- `cryptoutil-demo pki up` - Start all services for cipher product, with demo configs
- `cryptoutil-demo pki-ca up` - Start cipher-im service only, with demo configs
- `cryptoutil-demo sm up` - Start all services for cipher product, with demo configs
- `cryptoutil-demo sm-kms up` - Start cipher-im service only, with demo configs
- `cryptoutil-demo cipher up` - Start all services for cipher product, with demo configs
- `cryptoutil-demo cipher-im up` - Start cipher-im service only, with demo configs

### cryptoutil-e2e

**Purpose**: End-to-end testing command that orchestrates cryptoutil-compose. It uses opinionated test configuration data for comprehensive, automated testing.

**Call Flow**:

```
cmd/cryptoutil-e2e/main.go
  → internal/app/cryptoutil-e2e/cryptoutil-e2e.go
    → e2e(args) function
      → run|report|cleanup subcommand functions
```

**Examples**:

- `cryptoutil-e2e all up` - Start all products, including all of their services, with e2e configs
- `cryptoutil-e2e all down` - Stop all products, including all of their services, with e2e configs
- `cryptoutil-e2e all full` - Run E2E functionality tests for all products, including all of their services, with e2e configs
- `cryptoutil-e2e all load` - Run E2E load tests for all products, including all of their services, with e2e configs
- `cryptoutil-e2e all bench` - Run E2E benchmark tests for all products, including all of their services, with e2e configs
- `cryptoutil-e2e all smoke` - Run E2E smoke tests for all products, including all of their services, with e2e configs
- `cryptoutil-e2e all status` - Show status for all products, including all of their services, with e2e configs
- `cryptoutil-e2e all clean` - Remove all containers and volumes for all products, including all of their services, with e2e configs
- `cryptoutil-e2e identity up` - Start all services for identity product, with e2e configs
- `cryptoutil-e2e identity-authz up` - Start identity-authz service only, with e2e configs
- `cryptoutil-e2e identity-idp up` - Start identity-idp service only, with e2e configs
- `cryptoutil-e2e identity-rp up` - Start identity-rp service only, with e2e configs
- `cryptoutil-e2e identity-rs up` - Start identity-rs service only, with e2e configs
- `cryptoutil-e2e identity-spa up` - Start identity-spa service only, with e2e configs
- `cryptoutil-e2e jose up` - Start all services for cipher product, with e2e configs
- `cryptoutil-e2e jose-ja up` - Start cipher-im service only, with e2e configs
- `cryptoutil-e2e pki up` - Start all services for cipher product, with e2e configs
- `cryptoutil-e2e pki-ca up` - Start cipher-im service only, with e2e configs
- `cryptoutil-e2e sm up` - Start all services for cipher product, with e2e configs
- `cryptoutil-e2e sm-kms up` - Start cipher-im service only, with e2e configs
- `cryptoutil-e2e cipher up` - Start all services for cipher product, with e2e configs
- `cryptoutil-e2e cipher-im up` - Start cipher-im service only, with e2e configs
