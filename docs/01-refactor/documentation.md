# Documentation Restructuring Plan

## Overview

This document defines the reorganization of documentation to align with the service group refactoring, ensuring clear navigation, consistent structure, and accurate cross-references.

**Cross-references:**

- [Group Directory Blueprint](./blueprint.md) - Defines target package locations
- [Service Group Taxonomy](./service-groups.md) - Defines 43 service groups
- [CLI Strategy Framework](./cli-strategy.md) - CLI command structure

---

## Current Documentation Structure

### Root README.md

**Current Sections:**

1. Notice (ASUS RMA status)
2. Introduction & Key Features
3. Cryptographic Standards
4. API Architecture (Browser vs Service contexts)
5. Context Paths Hierarchy
6. Security Features
7. Observability & Monitoring
8. Telemetry Architecture
9. Port Architecture
10. Production Ready features
11. Quick Start
12. Development Setup
13. Testing Strategy
14. Documentation Organization (points to docs/README.md)

**Refactor Impact:**

- **Medium** - CLI commands change (server → kms server)
- Add service group navigation section
- Update Docker Compose service names (optional)

---

### docs/ Directory Structure

**Current Organization:**

```
docs/
├── 01-refactor/             # NEW - Current refactoring planning
├── 02-mixed/                # Active cross-cutting tasks
├── 03-identityV2/           # Archive - superseded
├── 04-identity/             # Current identity implementation
├── 05-ca/                   # CA planning
├── archive/                 # Historical documentation
├── DEV-SETUP.md             # Development environment setup
├── PLANS-INDEX.md           # Documentation index (2025-11-21)
├── pre-commit-hooks.md      # Pre-commit hook documentation
└── README.md                # Project deep dive
```

**Problems:**

- Numbered directories (01-05) are ad-hoc, not aligned with service groups
- Mixed active/archive content (03-identityV2 superseded)
- No clear service group navigation
- PLANS-INDEX.md needs updating for refactor docs

---

## Proposed Documentation Structure Post-Refactor

### Root README.md Updates

**New Section: Service Groups** (after Introduction)

```markdown
## Service Groups

cryptoutil is organized into three primary service groups:

### 🔐 KMS (Key Management System)
**Purpose:** Secure key generation, storage, and cryptographic operations
**CLI:** `cryptoutil kms <subcommand>`
**Documentation:** [docs/kms/README.md](./docs/kms/README.md)

**Key Features:**
- Hierarchical key management (unseal → root → intermediate → content keys)
- Elastic and Material key operations
- JWE/JWS encryption and signing
- Barrier system with M-of-N secret sharing

**Subcommands:**
- `cryptoutil kms server start` - Start KMS server
- `cryptoutil kms client status` - Check KMS server status
- `cryptoutil kms keygen` - Generate cryptographic keys
- `cryptoutil kms barrier unseal` - Unseal the barrier

---

### 🔑 Identity (OAuth2/OIDC)
**Purpose:** Authentication, authorization, and identity federation
**CLI:** `cryptoutil identity <subcommand>`
**Documentation:** [docs/identity/README.md](./docs/identity/README.md)

**Key Features:**
- OAuth 2.0 Authorization Server (AuthZ)
- OpenID Connect Identity Provider (IdP)
- Resource Server (RS)
- Single Page Application Relying Party (SPA-RP)

**Subcommands:**
- `cryptoutil identity authz start` - Start OAuth2 Authorization Server
- `cryptoutil identity idp start` - Start OpenID Connect IdP
- `cryptoutil identity rs start` - Start Resource Server
- `cryptoutil identity spa-rp start` - Start SPA Relying Party

---

### 🏛️ CA (Certificate Authority)
**Purpose:** X.509 certificate issuance, revocation, and PKI operations
**CLI:** `cryptoutil ca <subcommand>`
**Documentation:** [docs/ca/README.md](./docs/ca/README.md)

**Key Features:**
- Certificate issuance and renewal
- Certificate revocation and CRL management
- OCSP responder
- CA/Browser Forum Baseline Requirements compliance

**Subcommands:**
- `cryptoutil ca server start` - Start CA server
- `cryptoutil ca client issue` - Issue certificate
- `cryptoutil ca client revoke` - Revoke certificate
- `cryptoutil ca client crl` - Manage CRL
```

**Update Quick Start section:**

```markdown
## Quick Start

### Choose Your Service Group

#### 🔐 KMS Server
```sh
# Development (SQLite)
cryptoutil kms server start --dev

# Production (PostgreSQL)
docker compose -f ./deployments/compose/compose.yml up -d
```

#### 🔑 Identity Services

```sh
# Start OAuth2 Authorization Server
cryptoutil identity authz start --config=configs/identity/development.yml

# Start OpenID Connect IdP
cryptoutil identity idp start --config=configs/identity/development.yml
```

#### 🏛️ CA Server (Planned)

```sh
# Start CA server
cryptoutil ca server start --config=configs/ca/development.yml
```

```

---

### docs/ Reorganization

**Proposed Structure:**
```

docs/
├── README.md                # Project deep dive (updated with service group TOC)
├── PLANS-INDEX.md           # Documentation index (updated for refactor)
├── DEV-SETUP.md             # Development environment setup (unchanged)
├── pre-commit-hooks.md      # Pre-commit hook documentation (unchanged)
│
├── refactor/                # Refactoring planning (renamed from 01-refactor)
│   ├── README.md            # Refactor overview and progress tracking
│   ├── service-groups.md    # Service group taxonomy (43 groups)
│   ├── dependency-analysis.md
│   ├── blueprint.md
│   ├── import-aliases.md
│   ├── cli-strategy.md
│   ├── shared-utilities.md
│   ├── pipeline-impact.md
│   ├── tooling.md
│   └── documentation.md     # This file
│
├── kms/                     # KMS service group documentation
│   ├── README.md            # KMS overview and getting started
│   ├── architecture.md      # Barrier system, key hierarchy, JWE/JWS
│   ├── api-reference.md     # Browser/Service API endpoints
│   ├── deployment.md        # Docker Compose, Kubernetes, production config
│   ├── security.md          # IP allowlisting, rate limiting, unseal modes
│   ├── migration.md         # Migration from legacy 'cryptoutil server' commands
│   └── troubleshooting.md   # Common issues and solutions
│
├── identity/                # Identity service group documentation
│   ├── README.md            # Identity overview and OAuth2/OIDC concepts
│   ├── authz.md             # OAuth2 Authorization Server
│   ├── idp.md               # OpenID Connect IdP
│   ├── rs.md                # Resource Server
│   ├── spa-rp.md            # Single Page Application Relying Party
│   ├── flows.md             # OAuth2 flows (authorization code, client credentials, etc.)
│   ├── deployment.md        # Docker Compose, configuration
│   └── troubleshooting.md   # Common issues and solutions
│
├── ca/                      # CA service group documentation (placeholder)
│   ├── README.md            # CA overview and PKI concepts
│   ├── architecture.md      # CA hierarchy, certificate profiles
│   ├── api-reference.md     # Certificate issuance, revocation, CRL, OCSP endpoints
│   ├── deployment.md        # CA deployment and configuration
│   ├── compliance.md        # CA/Browser Forum Baseline Requirements
│   └── troubleshooting.md   # Common issues and solutions
│
├── mixed/                   # Cross-cutting concerns (renamed from 02-mixed)
│   ├── README.md            # Overview of cross-cutting tasks
│   ├── todos-development.md
│   ├── todos-infrastructure.md
│   ├── todos-observability.md
│   ├── todos-quality.md
│   ├── todos-security.md
│   └── todos-testing.md
│
└── archive/                 # Historical documentation
    ├── identityV2/          # Renamed from 03-identityV2
    ├── identity-legacy/     # Renamed from 04-identity (after migration)
    ├── ca-planning/         # Renamed from 05-ca (after implementation)
    ├── cicd-refactoring-nov2025/
    └── codecov-nov2025/

```

**Key Changes:**
1. Remove numeric prefixes (`01-`, `02-`, etc.) for consistency
2. Create service group directories (`kms/`, `identity/`, `ca/`)
3. Consolidate refactoring docs under `refactor/`
4. Move legacy docs to `archive/` with descriptive names
5. Update PLANS-INDEX.md for new structure

---

## Service Group Documentation Templates

### Template Structure (kms/README.md, identity/README.md, ca/README.md)

```markdown
# <Service Group Name> Documentation

## Overview
Brief introduction to the service group and its purpose.

## Quick Start
Minimal commands to get started:
- Development setup
- Docker Compose deployment
- Basic API usage

## Architecture
High-level architecture diagram and component overview.

## Key Features
- Feature 1 with description
- Feature 2 with description
- Feature 3 with description

## CLI Commands
Table of all CLI commands for this service group:
| Command | Description | Example |
|---------|-------------|---------|
| ... | ... | ... |

## API Endpoints
Link to detailed API reference or embedded table:
| Endpoint | Method | Description |
|----------|--------|-------------|
| ... | ... | ... |

## Configuration
Configuration file structure and common settings.

## Deployment
- Docker Compose
- Kubernetes
- Production considerations

## Security
Service-specific security features and best practices.

## Troubleshooting
Common issues and solutions.

## Further Reading
- [Architecture Deep Dive](./architecture.md)
- [API Reference](./api-reference.md)
- [Deployment Guide](./deployment.md)
- [Security Guide](./security.md)
```

---

## Cross-Reference Updates

### Update All Documentation Cross-References

**Search Patterns:**

```
docs/01-refactor/  → docs/refactor/
docs/02-mixed/     → docs/mixed/
docs/03-identityV2/ → docs/archive/identityV2/
docs/04-identity/  → docs/identity/
docs/05-ca/        → docs/ca/
```

**Files Requiring Updates:**

- Root README.md
- docs/README.md
- docs/PLANS-INDEX.md
- All files in `docs/refactor/`
- All files in `docs/mixed/`
- All files in `docs/identity/`
- GitHub Copilot instructions (.github/copilot-instructions.md)

**Tool for automated updates:**

```sh
# Find all cross-references (dry run)
grep -r "docs/01-refactor/" docs/ .github/
grep -r "docs/02-mixed/" docs/ .github/

# Replace with sed (after validation)
find docs/ .github/ -type f -name "*.md" -exec sed -i 's|docs/01-refactor/|docs/refactor/|g' {} +
find docs/ .github/ -type f -name "*.md" -exec sed -i 's|docs/02-mixed/|docs/mixed/|g' {} +
```

---

## PLANS-INDEX.md Updates

**New Structure:**

```markdown
# Cryptoutil Documentation Plans Index

**Last Updated**: 2025-11-<date>
**Purpose**: Organized index of all documentation plan directories

---

## Documentation Organization

### Service Groups

#### KMS (Key Management System)
**Location**: `docs/kms/`
**Purpose**: KMS architecture, API reference, deployment, security
**Key Files**:
- `README.md` - KMS overview and getting started
- `architecture.md` - Barrier system, key hierarchy
- `api-reference.md` - Browser/Service API endpoints
- `deployment.md` - Docker Compose, Kubernetes
- `security.md` - IP allowlisting, rate limiting, unseal modes

#### Identity (OAuth2/OIDC)
**Location**: `docs/identity/`
**Purpose**: Identity services, OAuth2, OIDC implementation
**Key Files**:
- `README.md` - Identity overview and OAuth2/OIDC concepts
- `authz.md` - OAuth2 Authorization Server
- `idp.md` - OpenID Connect IdP
- `flows.md` - OAuth2 flows (authorization code, client credentials)

#### CA (Certificate Authority)
**Location**: `docs/ca/`
**Purpose**: PKI operations, certificate issuance/revocation (planned)
**Key Files**:
- `README.md` - CA overview and PKI concepts
- `architecture.md` - CA hierarchy, certificate profiles (planned)

---

### Project Planning

#### Refactoring (Active)
**Location**: `docs/refactor/`
**Last Updated**: 2025-11-<date>
**Purpose**: Service group refactoring planning and migration
**Key Files**:
- `service-groups.md` - 43 service group taxonomy
- `dependency-analysis.md` - Package coupling analysis
- `blueprint.md` - Target directory structure
- `import-aliases.md` - Import alias migration (85 → 115 aliases)
- `cli-strategy.md` - CLI framework patterns
- `shared-utilities.md` - Utility extraction plan
- `pipeline-impact.md` - CI/CD workflow updates
- `tooling.md` - VS Code configuration updates
- `documentation.md` - Documentation restructuring

#### Cross-Cutting Concerns
**Location**: `docs/mixed/`
**Last Updated**: 2025-11-21
**Purpose**: Ongoing maintenance tasks across all service groups
**Key Files**:
- `todos-development.md` - Development workflow (12-factor compliance ✅ COMPLETE)
- `todos-infrastructure.md` - Infrastructure & deployment
- `todos-observability.md` - Observability & monitoring
- `todos-quality.md` - Code quality & testing
- `todos-security.md` - Security hardening
- `todos-testing.md` - Testing infrastructure (⚠️ GORM AutoMigrate blocker)

---

### Historical Documentation

**Location**: `docs/archive/`
**Purpose**: Superseded or legacy documentation for reference
**Directories**:
- `identityV2/` - Second iteration of identity implementation (superseded)
- `identity-legacy/` - Original identity implementation (migration reference)
- `ca-planning/` - Early CA planning documents (superseded by docs/ca/)
- `cicd-refactoring-nov2025/` - CI/CD refactoring archive
- `codecov-nov2025/` - Code coverage investigation archive
```

---

## Migration Checklist

### Phase 1: Directory Reorganization

- [ ] Rename `docs/01-refactor/` → `docs/refactor/`
- [ ] Rename `docs/02-mixed/` → `docs/mixed/`
- [ ] Move `docs/03-identityV2/` → `docs/archive/identityV2/`
- [ ] Create `docs/kms/` directory
- [ ] Create `docs/identity/` directory (keep 04-identity temporarily)
- [ ] Create `docs/ca/` directory (keep 05-ca temporarily)

### Phase 2: Service Group Documentation

- [ ] Create `docs/kms/README.md` (KMS overview)
- [ ] Create `docs/kms/architecture.md` (barrier system, key hierarchy)
- [ ] Create `docs/kms/api-reference.md` (Browser/Service API endpoints)
- [ ] Create `docs/kms/deployment.md` (Docker Compose, Kubernetes)
- [ ] Create `docs/kms/security.md` (IP allowlisting, rate limiting)
- [ ] Create `docs/kms/migration.md` (legacy command migration)
- [ ] Create `docs/identity/README.md` (Identity overview)
- [ ] Create `docs/identity/authz.md` (OAuth2 Authorization Server)
- [ ] Create `docs/identity/idp.md` (OpenID Connect IdP)
- [ ] Create `docs/identity/flows.md` (OAuth2 flows)
- [ ] Create `docs/ca/README.md` (CA overview placeholder)

### Phase 3: Cross-Reference Updates

- [ ] Update root README.md with service group navigation
- [ ] Update docs/README.md with new TOC
- [ ] Update docs/PLANS-INDEX.md with new structure
- [ ] Find and replace all `docs/01-refactor/` → `docs/refactor/`
- [ ] Find and replace all `docs/02-mixed/` → `docs/mixed/`
- [ ] Find and replace all `docs/03-identityV2/` → `docs/archive/identityV2/`
- [ ] Update .github/copilot-instructions.md with new paths

### Phase 4: Archive Legacy Docs

- [ ] Move `docs/04-identity/` → `docs/archive/identity-legacy/` (after migration complete)
- [ ] Move `docs/05-ca/` → `docs/archive/ca-planning/` (after CA implementation complete)
- [ ] Document migration in `docs/archive/README.md`

### Phase 5: Validation

- [ ] Verify all Markdown links work
- [ ] Check Markdown link validation: `markdown.validate.fileLinks.enabled: "warning"`
- [ ] Verify GitHub renders all files correctly
- [ ] Test VS Code Markdown preview for all service group docs
- [ ] Ensure PLANS-INDEX.md is accurate and up-to-date

---

## Documentation Style Guide

### File Naming Conventions

- **Use kebab-case**: `api-reference.md`, `deployment-guide.md`
- **Descriptive names**: `kms-barrier-system.md` instead of `kms-bs.md`
- **Avoid abbreviations**: `troubleshooting.md` not `ts.md`

### Markdown Standards

- **Headers**: Use ATX-style (`#`, `##`, `###`)
- **Lists**: Use `-` for unordered, `1.` for ordered
- **Code blocks**: Always specify language (```go,```sh, ```yaml)
- **Links**: Use relative paths for internal docs, absolute for external
- **Cross-references**: Always use relative paths from current file
- **Tables**: Use pipe-separated markdown tables with header separators

### Content Structure

- **Overview section**: 1-2 paragraphs summarizing the document
- **Quick start**: Minimal example to get started
- **Deep dive**: Detailed explanations with examples
- **Cross-references**: Link to related documents at the end
- **Notes section**: Additional context or caveats

---

## Cross-References

- **Group Directory Blueprint:** [docs/refactor/blueprint.md](./blueprint.md)
- **Service Group Taxonomy:** [docs/refactor/service-groups.md](./service-groups.md)
- **CLI Strategy Framework:** [docs/refactor/cli-strategy.md](./cli-strategy.md)
- **Import Alias Policy:** [docs/refactor/import-aliases.md](./import-aliases.md)

---

## Notes

- **Directory numbering removed** - Simplifies navigation and aligns with service groups
- **Service group docs created** - Provides clear entry points for KMS, Identity, CA
- **Archive strategy** - Preserves legacy docs while cleaning up main documentation tree
- **Cross-reference automation** - Use find/replace with sed for bulk updates
- **Markdown validation** - Leverage VS Code settings for link checking
- **PLANS-INDEX.md is critical** - Single source of truth for documentation navigation
- **Template structure ensures consistency** - All service group docs follow same pattern
