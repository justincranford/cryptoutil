# Command Line Patterns for Cryptoutil Services

## Overview

This document describes three aspirational patterns for organizing command-line executables in the cryptoutil project. These patterns provide structured and uniform approaches to command-line interfaces for cryptographic services.

## Patterns

### 1. Suite Pattern

**Description**: A single unified `cryptoutil` executable that provides subcommands for all products and services.

**Call Flow**:

```
cmd/cryptoutil/main.go
  → internal/cmd/cryptoutil/cryptoutil.go
    → internal/cmd/PRODUCT/PRODUCT.go
      → PRODUCT(args) function
        → SERVICE(args) function
          → server|client|init|etc subcommand functions
```

**Examples**:

- `cryptoutil identity authz server --config configs/cryptoutil/config.yml --config configs/identity/config.yml --config configs/identity/authz/config.yml`
  → `cryptoutil identity authz server` → `internal/cmd/identity/identity.go` → `authz(args)` → `server(args)`

- `cryptoutil identity authz client --config configs/cryptoutil/config.yml --config configs/identity/config.yml --config configs/identity/authz/config.yml`
  → `cryptoutil identity authz client` → `internal/cmd/identity/identity.go` → `authz(args)` → `client(args)`

- `cryptoutil identity authz init --config configs/cryptoutil/config.yml --config configs/identity/config.yml --config configs/identity/authz/config.yml`
  → `cryptoutil identity authz init` → `internal/cmd/identity/identity.go` → `authz(args)` → `init(args)`

- `cryptoutil identity authz health --config configs/cryptoutil/config.yml --config configs/identity/config.yml --config configs/identity/authz/config.yml`
  → `cryptoutil identity authz health` → `internal/cmd/identity/identity.go` → `authz(args)` → `health(args)`

- `cryptoutil identity authz livez --config configs/cryptoutil/config.yml --config configs/identity/config.yml --config configs/identity/authz/config.yml`
  → `cryptoutil identity authz livez` → `internal/cmd/identity/identity.go` → `authz(args)` → `livez(args)`

- `cryptoutil identity authz readyz --config configs/cryptoutil/config.yml --config configs/identity/config.yml --config configs/identity/authz/config.yml`
  → `cryptoutil identity authz readyz` → `internal/cmd/identity/identity.go` → `authz(args)` → `readyz(args)`

- `cryptoutil identity shutdown stop --config configs/cryptoutil/config.yml --config configs/identity/config.yml --config configs/identity/authz/config.yml`
  → `cryptoutil identity shutdown stop` → `internal/cmd/identity/identity.go` → `authz(args)` → `shutdown(args)`

- `cryptoutil jose ja client --config configs/cryptoutil/config.yml --config configs/jose/config.yml --config configs/jose/ja/config.yml`
  → `cryptoutil jose ja client` → `internal/cmd/jose/jose.go` → `ja(args)` → `client(args)`

- `cryptoutil jose ja server --config configs/cryptoutil/config.yml --config configs/jose/config.yml --config configs/jose/ja/config.yml`
  → `cryptoutil jose ja server` → `internal/cmd/jose/jose.go` → `ja(args)` → `server(args)`

- `cryptoutil jose ja init --config configs/cryptoutil/config.yml --config configs/jose/config.yml --config configs/jose/ja/config.yml`
  → `cryptoutil jose ja init` → `internal/cmd/jose/jose.go` → `ja(args)` → `init(args)`

- `cryptoutil jose ja health --config configs/cryptoutil/config.yml --config configs/jose/config.yml --config configs/jose/ja/config.yml`
  → `cryptoutil jose ja health` → `internal/cmd/jose/jose.go` → `ja(args)` → `health(args)`

- `cryptoutil jose ja livez --config configs/cryptoutil/config.yml --config configs/jose/config.yml --config configs/jose/ja/config.yml`
  → `cryptoutil jose ja livez` → `internal/cmd/jose/jose.go` → `ja(args)` → `livez(args)`

- `cryptoutil jose ja readyz --config configs/cryptoutil/config.yml --config configs/jose/config.yml --config configs/jose/ja/config.yml`
  → `cryptoutil jose ja readyz` → `internal/cmd/jose/jose.go` → `ja(args)` → `readyz(args)`

- `cryptoutil jose shutdown stop --config configs/cryptoutil/config.yml --config configs/jose/config.yml --config configs/jose/ja/config.yml`
  → `cryptoutil jose shutdown stop` → `internal/cmd/jose/jose.go` → `ja(args)` → `shutdown(args)`

- `cryptoutil pki ca server --config configs/cryptoutil/config.yml --config configs/pki/config.yml --config configs/pki/ca/config.yml`
  → `cryptoutil pki ca server` → `internal/cmd/pki/pki.go` → `ca(args)` → `server(args)`

- `cryptoutil pki ca init --config configs/cryptoutil/config.yml --config configs/pki/config.yml --config configs/pki/ca/config.yml`
  → `cryptoutil pki ca init` → `internal/cmd/pki/pki.go` → `ca(args)` → `init(args)`

- `cryptoutil pki ca health --config configs/cryptoutil/config.yml --config configs/pki/config.yml --config configs/pki/ca/config.yml`
  → `cryptoutil pki ca health` → `internal/cmd/pki/pki.go` → `ca(args)` → `health(args)`

- `cryptoutil pki ca livez --config configs/cryptoutil/config.yml --config configs/pki/config.yml --config configs/pki/ca/config.yml`
  → `cryptoutil pki ca livez` → `internal/cmd/pki/pki.go` → `ca(args)` → `livez(args)`

- `cryptoutil pki ca readyz --config configs/cryptoutil/config.yml --config configs/pki/config.yml --config configs/pki/ca/config.yml`
  → `cryptoutil pki ca readyz` → `internal/cmd/pki/pki.go` → `ca(args)` → `readyz(args)`

- `cryptoutil pki shutdown stop --config configs/cryptoutil/config.yml --config configs/pki/config.yml --config configs/pki/ca/config.yml`
  → `cryptoutil pki shutdown stop` → `internal/cmd/pki/pki.go` → `ca(args)` → `shutdown(args)`

- `cryptoutil sm kms server --config configs/cryptoutil/config.yml --config configs/sm/config.yml --config configs/sm/kms/config.yml`
  → `cryptoutil sm kms server` → `internal/cmd/sm/sm.go` → `kms(args)` → `server(args)`

- `cryptoutil sm kms client --config configs/cryptoutil/config.yml --config configs/sm/config.yml --config configs/sm/kms/config.yml`
  → `cryptoutil sm kms client` → `internal/cmd/sm/sm.go` → `kms(args)` → `client(args)`

- `cryptoutil sm kms init --config configs/cryptoutil/config.yml --config configs/sm/config.yml --config configs/sm/kms/config.yml`
  → `cryptoutil sm kms init` → `internal/cmd/sm/sm.go` → `kms(args)` → `init(args)`

- `cryptoutil sm kms health --config configs/cryptoutil/config.yml --config configs/sm/config.yml --config configs/sm/kms/config.yml`
  → `cryptoutil sm kms health` → `internal/cmd/sm/sm.go` → `kms(args)` → `health(args)`

- `cryptoutil sm kms livez --config configs/cryptoutil/config.yml --config configs/sm/config.yml --config configs/sm/kms/config.yml`
  → `cryptoutil sm kms livez` → `internal/cmd/sm/sm.go` → `kms(args)` → `livez(args)`

- `cryptoutil sm kms readyz --config configs/cryptoutil/config.yml --config configs/sm/config.yml --config configs/sm/kms/config.yml`
  → `cryptoutil sm kms readyz` → `internal/cmd/sm/sm.go` → `kms(args)` → `readyz(args)`

- `cryptoutil sm shutdown stop --config configs/cryptoutil/config.yml --config configs/sm/config.yml --config configs/sm/kms/config.yml`
  → `cryptoutil sm shutdown stop` → `internal/cmd/sm/sm.go` → `kms(args)` → `shutdown(args)`

- `cryptoutil cipher im server --config configs/cryptoutil/config.yml --config configs/cipher/config.yml --config configs/cipher/im/config.yml`
  → `cryptoutil cipher im server` → `internal/cmd/cipher/cipher.go` → `im(args)` → `server(args)`

- `cryptoutil cipher im client --config configs/cryptoutil/config.yml --config configs/cipher/config.yml --config configs/cipher/im/config.yml`
  → `cryptoutil cipher im client` → `internal/cmd/cipher/cipher.go` → `im(args)` → `client(args)`

- `cryptoutil cipher im init --config configs/cryptoutil/config.yml --config configs/cipher/config.yml --config configs/cipher/im/config.yml`
  → `cryptoutil cipher im init` → `internal/cmd/cipher/cipher.go` → `im(args)` → `init(args)`

- `cryptoutil cipher im health --config configs/cryptoutil/config.yml --config configs/cipher/config.yml --config configs/cipher/im/config.yml`
  → `cryptoutil cipher im health` → `internal/cmd/cipher/cipher.go` → `im(args)` → `health(args)`

- `cryptoutil cipher im livez --config configs/cryptoutil/config.yml --config configs/cipher/config.yml --config configs/cipher/im/config.yml`
  → `cryptoutil cipher im livez` → `internal/cmd/cipher/cipher.go` → `im(args)` → `livez(args)`

- `cryptoutil cipher im readyz --config configs/cryptoutil/config.yml --config configs/cipher/config.yml --config configs/cipher/im/config.yml`
  → `cryptoutil cipher im readyz` → `internal/cmd/cipher/cipher.go` → `im(args)` → `readyz(args)`

- `cryptoutil cipher shutdown stop --config configs/cryptoutil/config.yml --config configs/cipher/config.yml --config configs/cipher/im/config.yml`
  → `cryptoutil cipher shutdown stop` → `internal/cmd/cipher/cipher.go` → `im(args)` → `shutdown(args)`

---

### 2. Product Pattern

**Description**: Separate executables for each product, with subcommands for services within that product.

**Call Flow**:

```
cmd/PRODUCT/main.go
  → internal/cmd/PRODUCT/PRODUCT.go
    → PRODUCT(args) function
      → SERVICE(args) function
        → server|client|init|etc subcommand functions
```

**Examples**:

- `identity authz server --config configs/cryptoutil/config.yml --config configs/identity/config.yml --config configs/identity/authz/config.yml`
  → `identity authz server` → `internal/cmd/identity/identity.go` → `authz(args)` → `server(args)`

- `jose ja client --config configs/cryptoutil/config.yml --config configs/jose/config.yml --config configs/jose/ja/config.yml`
  → `jose ja client` → `internal/cmd/jose/jose.go` → `ja(args)` → `client(args)`

- `pki ca init --config configs/cryptoutil/config.yml --config configs/pki/config.yml --config configs/pki/ca/config.yml`
  → `pki ca init` → `internal/cmd/pki/pki.go` → `ca(args)` → `init(args)`

- `sm kms server --config configs/cryptoutil/config.yml --config configs/sm/config.yml --config configs/sm/kms/config.yml`
  → `sm kms server` → `internal/cmd/sm/sm.go` → `kms(args)` → `server(args)`

- `cipher im server --config configs/cryptoutil/config.yml --config configs/cipher/config.yml --config configs/cipher/im/config.yml`
  → `cipher im server` → `internal/cmd/cipher/cipher.go` → `im(args)` → `server(args)`

- `cipher im client --config configs/cryptoutil/config.yml --config configs/cipher/config.yml --config configs/cipher/im/config.yml`
  → `cipher im client` → `internal/cmd/cipher/cipher.go` → `im(args)` → `client(args)`

---

### 3. Product-Service Pattern

**Description**: Separate executables for each product-service combination, providing focused functionality.

**Call Flow**:

```
cmd/PRODUCT-SERVICE/main.go
  → internal/cmd/PRODUCT/PRODUCT.go
    → SERVICE(args) function
      → server|client|init|etc subcommand functions
```

**Examples**:

- `identity-authz server --config configs/cryptoutil/config.yml --config configs/identity/config.yml --config configs/identity/authz/config.yml`
  → `identity-authz server` → `internal/cmd/identity/identity.go` → `authz(args)` → `server(args)`

- `jose-ja client --config configs/cryptoutil/config.yml --config configs/jose/config.yml --config configs/jose/ja/config.yml`
  → `jose-ja client` → `internal/cmd/jose/jose.go` → `ja(args)` → `client(args)`

- `pki-ca init --config configs/cryptoutil/config.yml --config configs/pki/config.yml --config configs/pki/ca/config.yml`
  → `pki-ca init` → `internal/cmd/pki/pki.go` → `ca(args)` → `init(args)`

- `sm-kms server --config configs/cryptoutil/config.yml --config configs/sm/config.yml --config configs/sm/kms/config.yml`
  → `sm-kms server` → `internal/cmd/sm/sm.go` → `kms(args)` → `server(args)`

- `cipher-im server --config configs/cryptoutil/config.yml --config configs/cipher/config.yml --config configs/cipher/im/config.yml`
  → `cipher-im server` → `internal/cmd/cipher/cipher.go` → `im(args)` → `server(args)`

- `cipher-im client --config configs/cryptoutil/config.yml --config configs/cipher/config.yml --config configs/cipher/im/config.yml`
  → `cipher-im client` → `internal/cmd/cipher/cipher.go` → `im(args)` → `client(args)`

## Comparison

| Aspect | Suite Pattern | Product Pattern | Product-Service Pattern |
|--------|---------------|-----------------|------------------------|
| **Executables** | 1 (`cryptoutil`) | N (per product) | M (per service) |
| **Command Length** | Longer | Medium | Shortest |
| **Discoverability** | Single entry point | Product-focused | Service-focused |
| **Deployment** | Monolithic | Product isolation | Service isolation |
| **Complexity** | High (routing) | Medium | Low |
| **Maintenance** | Complex | Moderate | Simple |

## Recommended Pattern

**Product-Service Pattern** is recommended for the following reasons:

1. **Service Isolation**: Each service has its own executable, enabling independent deployment and scaling
2. **Clear Ownership**: Each binary corresponds to a specific service, making responsibilities clear
3. **Operational Simplicity**: Easier to manage, monitor, and troubleshoot individual services
4. **Container-First**: Aligns with microservices architecture and container deployment patterns
5. **Future-Proof**: Supports independent evolution of each service

## Implementation Notes

- All patterns support the same subcommand structure: `server`, `client`, `init`, etc.
- Configuration files follow the pattern: `configs/PRODUCT/SERVICE/config.yml`
- Internal command structure uses `internal/cmd/PRODUCT/PRODUCT.go` with `SERVICE(args)` functions
- Each service supports standard operations: server startup, client operations, initialization

## Top-Level Commands

These commands are available regardless of the chosen pattern and provide orchestration capabilities:

### cryptoutil-compose

**Purpose**: Top-level orchestration command for managing suite, product, or service startup/shutdown/cleanup operations. Orchestrates underlying services based on the selected pattern (configuration must be provided). Optional for production deployments. Required for demonstrations and end-to-end testing.

**Call Flow**:

```
cmd/cryptoutil-compose/main.go
  → internal/cmd/cryptoutil-compose/cryptoutil-compose.go
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
  → internal/cmd/cryptoutil-demo/cryptoutil-demo.go
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
  → internal/cmd/cryptoutil-e2e/cryptoutil-e2e.go
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
