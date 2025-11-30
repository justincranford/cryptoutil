# Service Groups Taxonomy

**Last Updated**: 2025-11-21
**Purpose**: Define comprehensive service group taxonomy for repository restructuring
**Scope**: 3 mandatory groups + 10 additional repo-driven groups + 30 adjacent-market groups = 43 groups total

---

## Executive Summary

This taxonomy defines a hierarchical organization of service groups within the cryptoutil repository, establishing clear boundaries between domains while enabling strategic expansion into adjacent markets. The structure follows industry-standard patterns for security-focused products and aligns with enterprise PKI, identity management, and zero-trust architectures.

---

## Part 1: Mandatory Service Groups (3)

These groups represent core functionality already implemented or in active development.

### Group 1: KMS (Key Management Service)

**Package Path**: `internal/kms/`, `cmd/kms/`, `api/kms/`
**Current Location**: `internal/server/`, `internal/client/`, `cmd/cryptoutil/`, `api/server/`, `api/client/`
**Status**: ✅ Implemented, needs reorganization

**Scope**:

- Hierarchical key management (unseal → root → intermediate → content keys)
- FIPS 140-3 compliant cryptographic operations
- JWE/JWS support for encrypted key storage
- Multi-tier barrier system with configurable unseal modes
- Key rotation, versioning, and lifecycle management
- Elastic keys (tenant-specific) and material keys (key material management)

**Key Features**:

- Dual-context API: browser (`/browser/api/v1/*`) and service (`/service/api/v1/*`)
- PostgreSQL and SQLite backend support
- OpenTelemetry instrumentation (traces, metrics, logs)
- IP allowlisting and per-IP rate limiting
- CSRF protection and comprehensive security headers

**Dependencies**:

- Depends On: `internal/common/magic`, `internal/common/crypto`, `internal/common/pool`, `internal/common/telemetry`
- Used By: Identity, CA, Secrets (future)

**Rationale**: Core product offering; requires dedicated namespace for clarity and future HSM integration planning.

---

### Group 2: Identity (OAuth 2.1 / OIDC)

**Package Path**: `internal/identity/`, `cmd/identity/`, `api/identity/`
**Current Location**: `internal/identity/`, `cmd/identity/`, `api/identity/`
**Status**: ✅ Partially implemented (authorization server, identity provider, resource server)

**Scope**:

- OAuth 2.1 Authorization Server (AuthZ) with PKCE enforcement
- OpenID Connect Identity Provider (IdP) with UserInfo endpoint
- Resource Server (RS) with token validation
- SPA Relying Party (SPA-RP) for demonstration
- Multi-factor authentication (MFA) orchestration
- Session management with encrypted tokens
- Client authentication (client_secret_post, client_secret_basic, private_key_jwt, mTLS)

**Key Features**:

- Adaptive authentication with risk-based step-up
- Proof Key for Code Exchange (PKCE) mandatory for authorization code flow
- GORM-based persistence (PostgreSQL, SQLite)
- OAuth 2.1 and OIDC compliance (RFC 6749, RFC 6750, RFC 7636, RFC 8414)
- JWT-based access and refresh tokens

**Dependencies**:

- Depends On: `internal/common/magic` ONLY (domain isolation enforced via lint rules)
- Does NOT depend on: KMS server, client, API, crypto utilities (uses stdlib)
- Used By: SPA applications, microservices requiring authentication

**Rationale**: Identity domain must remain isolated from KMS to enable independent deployment. Critical for zero-trust architectures and multi-tenancy.

---

### Group 3: CA (Certificate Authority)

**Package Path**: `internal/ca/`, `cmd/ca/`, `api/ca/`
**Current Location**: (Planned - see `docs/05-ca/README.md`)
**Status**: 📋 Planned (20 tasks documented)

**Scope**:

- Root, Intermediate, and Issuing CA provisioning
- 20+ certificate profile library (TLS server/client, S/MIME, code signing, document signing, VPN, IoT, SAML, JWT, OCSP, RA, TSA, CT log, ACME, SCEP, EST, CMP, enterprise custom)
- Certificate lifecycle management (issuance, renewal, revocation)
- CRL and OCSP responder services
- Time-stamping authority (TSA) and Registration Authority (RA) workflows
- ACME/SCEP/EST/CMP protocol support
- Certificate transparency log integration

**Key Features**:

- CA/Browser Forum Baseline Requirements compliance
- RFC 5280 strict enforcement
- YAML-driven configuration for crypto, subject, and certificate profiles
- Multi-backend persistence (PostgreSQL, SQLite)
- Observability and audit logging for compliance

**Dependencies**:

- Depends On: `internal/common/magic`, `internal/common/crypto` (key generation)
- May Use: KMS for key storage (optional HSM integration)
- Used By: Identity (mTLS certificates), microservices (service mesh), IoT devices

**Rationale**: Enterprise PKI is a natural extension of KMS capabilities. CA/Browser Forum compliance opens path to public CA services.

---

## Part 2: Additional Repo-Driven Groups (10)

These groups represent strategic enhancements to existing functionality or closely related capabilities.

### Group 4: Secrets (Secrets Management)

**Package Path**: `internal/secrets/`, `cmd/secrets/`, `api/secrets/`
**Status**: 🔮 Future

**Scope**:

- Centralized secrets storage and retrieval
- Dynamic secrets generation (database credentials, API keys)
- Secrets rotation and expiration policies
- Audit logging for secrets access
- Integration with Kubernetes secrets, Vault, AWS Secrets Manager

**Dependencies**:

- Depends On: KMS (encryption at rest)
- Used By: All services requiring secrets management

**Rationale**: Natural extension of KMS; enables secure configuration management and service credential rotation.

---

### Group 5: Vault (Multi-Tenant Vault)

**Package Path**: `internal/vault/`, `cmd/vault/`, `api/vault/`
**Status**: 🔮 Future

**Scope**:

- Multi-tenant secure storage (documents, files, credentials)
- Versioning and access control per vault
- Encryption at rest using KMS
- Compliance with data residency requirements

**Dependencies**:

- Depends On: KMS (encryption), Identity (authentication/authorization)
- Used By: SaaS applications, enterprise customers

**Rationale**: Builds on KMS and Identity to provide HashiCorp Vault-like functionality for multi-tenant environments.

---

### Group 6: PKI (Public Key Infrastructure Utilities)

**Package Path**: `internal/pki/`, `cmd/pki/`, `api/pki/`
**Status**: 🔮 Future

**Scope**:

- PKCS#7, PKCS#12, PEM/DER utilities
- CSR generation and validation
- Certificate chain building and verification
- CRL/OCSP client utilities
- S/MIME and CMS message support

**Dependencies**:

- Depends On: CA (certificate operations), `internal/common/crypto`
- Used By: Identity (mTLS), email encryption, document signing

**Rationale**: PKI utilities complement CA; provides tooling for certificate ecosystem operations.

---

### Group 7: Automation (Workflow & Orchestration)

**Package Path**: `internal/automation/`, `cmd/automation/`, `api/automation/`
**Status**: 🔮 Future

**Scope**:

- Certificate lifecycle automation (renewal, revocation)
- Key rotation automation
- Policy enforcement automation
- Integration with CI/CD pipelines (GitHub Actions, GitLab CI, Jenkins)

**Dependencies**:

- Depends On: KMS (key operations), CA (certificate operations)
- Used By: DevOps teams, infrastructure automation

**Rationale**: Reduces operational burden; critical for zero-touch deployment and compliance.

---

### Group 8: Gateway (API Gateway / Proxy)

**Package Path**: `internal/gateway/`, `cmd/gateway/`, `api/gateway/`
**Status**: 🔮 Future

**Scope**:

- Reverse proxy with TLS termination
- Rate limiting and quota management
- Request routing and load balancing
- Authentication/authorization integration (OAuth 2.1, mTLS)
- API key management

**Dependencies**:

- Depends On: Identity (authentication), CA (TLS certificates)
- Used By: Microservices, API consumers

**Rationale**: Provides unified entry point for all cryptoutil services; enables service mesh integration.

---

### Group 9: Monitoring (Advanced Observability)

**Package Path**: `internal/monitoring/`, `cmd/monitoring/`, `api/monitoring/`
**Status**: 🔮 Future (basic observability already exists)

**Scope**:

- Custom Grafana dashboards for KMS, Identity, CA
- Alert rules for security events (failed unseal, unauthorized access)
- Performance monitoring (key generation latency, certificate issuance throughput)
- Audit log aggregation and analysis
- Anomaly detection for cryptographic operations

**Dependencies**:

- Depends On: OpenTelemetry infrastructure (already exists)
- Used By: Operations teams, security teams

**Rationale**: Extends existing observability with domain-specific insights; critical for production operations.

---

### Group 10: Backup (Backup & Disaster Recovery)

**Package Path**: `internal/backup/`, `cmd/backup/`, `api/backup/`
**Status**: 🔮 Future

**Scope**:

- Automated backup of KMS keys, CA hierarchies, identity data
- Encrypted backup storage (local, S3, GCS, Azure Blob)
- Point-in-time recovery (PITR) for databases
- Disaster recovery drills and validation
- Cross-region replication

**Dependencies**:

- Depends On: KMS (encryption), all service databases
- Used By: Operations teams, compliance teams

**Rationale**: Business continuity requirement; critical for enterprise deployments and compliance (GDPR, SOC 2).

---

### Group 11: Audit (Audit Logging & Compliance)

**Package Path**: `internal/audit/`, `cmd/audit/`, `api/audit/`
**Status**: 🔮 Future (basic audit logging exists)

**Scope**:

- Centralized audit log collection (KMS, Identity, CA operations)
- Tamper-proof audit trails (signed logs, blockchain anchoring)
- Compliance reporting (SOC 2, ISO 27001, PCI-DSS)
- SIEM integration (Splunk, ELK, Azure Sentinel)
- Forensic analysis tools

**Dependencies**:

- Depends On: All services (audit log sources)
- Used By: Compliance teams, security teams, auditors

**Rationale**: Regulatory requirement for many industries; enables trust and accountability.

---

### Group 12: Compliance (Policy & Governance)

**Package Path**: `internal/compliance/`, `cmd/compliance/`, `api/compliance/`
**Status**: 🔮 Future

**Scope**:

- Policy definition and enforcement (e.g., key rotation every 90 days)
- Compliance checks (FIPS 140-3, CA/Browser Forum)
- Automated remediation workflows
- Compliance dashboard and reporting
- Integration with GRC platforms

**Dependencies**:

- Depends On: Audit (compliance data), all services (policy enforcement)
- Used By: Compliance officers, security teams

**Rationale**: Enables scalable governance; required for enterprise and regulated industries.

---

### Group 13: SDK (Client SDKs & Libraries)

**Package Path**: `pkg/sdk/`, `cmd/sdk/`
**Status**: 🔮 Future

**Scope**:

- Go SDK for KMS, Identity, CA APIs
- Python, JavaScript/TypeScript, Java, C# SDKs
- Code generation from OpenAPI specs
- Example applications and tutorials
- SDK versioning and compatibility guarantees

**Dependencies**:

- Depends On: All service APIs (OpenAPI specs)
- Used By: Application developers, integrators

**Rationale**: Lowers integration barrier; expands cryptoutil ecosystem adoption.

---

## Part 3: Adjacent-Market Groups (30)

These groups represent strategic expansion opportunities into related markets and use cases.

### Category: Hardware & Infrastructure (5 groups)

#### Group 14: HSM (Hardware Security Module Integration)

**Scope**: PKCS#11, KMIP, CloudHSM, Azure Key Vault, GCP Cloud KMS integration for key storage
**Market**: Enterprise PKI, financial services, government
**Dependencies**: KMS (key operations)
**Rationale**: Required for FIPS 140-2 Level 3+ compliance; enterprise requirement

---

#### Group 15: PQC (Post-Quantum Cryptography)

**Scope**: NIST PQC algorithms (CRYSTALS-Kyber, CRYSTALS-Dilithium), hybrid key exchange, migration tooling
**Market**: Future-proofing, government contracts, long-term security
**Dependencies**: KMS (key generation), CA (PQC certificates)
**Rationale**: Quantum computing threat; proactive positioning for quantum-safe cryptography

---

#### Group 16: IoT (IoT Device Management)

**Scope**: Lightweight crypto for constrained devices, device provisioning, OTA key updates, TPM integration
**Market**: Industrial IoT, smart home, connected devices
**Dependencies**: CA (device certificates), KMS (key provisioning)
**Rationale**: Massive IoT market; secure device lifecycle management is critical

---

#### Group 17: Edge (Edge Computing Security)

**Scope**: Edge KMS nodes, offline operation, eventual consistency, edge-to-cloud sync
**Market**: CDN, 5G MEC, autonomous vehicles, retail
**Dependencies**: KMS (distributed keys), Vault (edge secrets)
**Rationale**: Edge computing growth; local crypto operations with centralized policy

---

#### Group 18: Mobile (Mobile SDK & App Security)

**Scope**: iOS/Android SDKs, biometric authentication, secure enclave integration, mobile app hardening
**Market**: Consumer apps, enterprise mobile, BYOD
**Dependencies**: Identity (mobile auth), KMS (mobile key storage)
**Rationale**: Mobile-first world; secure mobile authentication and data protection

---

### Category: Cloud & Multi-Cloud (5 groups)

#### Group 19: Cloud-Native (Kubernetes & Service Mesh)

**Scope**: Helm charts, operators, service mesh integration (Istio, Linkerd), Kubernetes secrets management
**Market**: Cloud-native enterprises, DevOps teams
**Dependencies**: All services (containerized deployments)
**Rationale**: Cloud-native is the dominant deployment model; native integration required

---

#### Group 20: Multi-Cloud (Multi-Cloud Key Management)

**Scope**: Unified KMS across AWS, Azure, GCP, bring-your-own-key (BYOK), cloud-agnostic APIs
**Market**: Enterprises avoiding vendor lock-in, hybrid cloud
**Dependencies**: KMS (multi-cloud orchestration)
**Rationale**: Multi-cloud adoption growing; competitive differentiator

---

#### Group 21: Serverless (Serverless Security)

**Scope**: Lambda/Function-as-a-Service integration, ephemeral key management, function authentication
**Market**: Serverless applications, event-driven architectures
**Dependencies**: KMS (ephemeral keys), Identity (function authentication)
**Rationale**: Serverless growth; unique security challenges in stateless environments

---

#### Group 22: Container (Container Security)

**Scope**: Image signing, runtime encryption, Notary integration, container registry authentication
**Market**: CI/CD pipelines, container platforms
**Dependencies**: CA (image signing certificates), PKI (container identity)
**Rationale**: Container security is critical; supply chain attacks increasing

---

#### Group 23: SaaS (Multi-Tenant SaaS Platform)

**Scope**: Tenant isolation, per-tenant encryption, API gateway, billing integration, usage metering
**Market**: SaaS providers, ISVs
**Dependencies**: Vault (tenant data), Identity (tenant authentication), KMS (tenant keys)
**Rationale**: SaaS is dominant software delivery model; secure multi-tenancy is core requirement

---

### Category: Security & Zero-Trust (5 groups)

#### Group 24: Zero-Trust (Zero-Trust Architecture)

**Scope**: Continuous authentication, micro-segmentation, policy engine, trust scoring
**Market**: Enterprises adopting zero-trust, remote work security
**Dependencies**: Identity (authentication), Gateway (policy enforcement)
**Rationale**: Zero-trust is security best practice; enables least-privilege access

---

#### Group 25: SIEM (Security Information & Event Management)

**Scope**: Log aggregation, threat detection, incident response, SOAR integration
**Market**: Security operations centers (SOCs), MSSPs
**Dependencies**: Audit (security logs), Monitoring (metrics)
**Rationale**: SIEM integration is enterprise requirement; enables proactive threat hunting

---

#### Group 26: Threat-Intel (Threat Intelligence Integration)

**Scope**: IoC feeds, malware analysis, vulnerability scanning, threat hunting
**Market**: Security teams, incident responders
**Dependencies**: SIEM (threat correlation), Audit (forensic analysis)
**Rationale**: Proactive security posture; reduces time-to-detect/time-to-respond

---

#### Group 27: DLP (Data Loss Prevention)

**Scope**: Sensitive data discovery, encryption enforcement, data exfiltration prevention
**Market**: Regulated industries (finance, healthcare, government)
**Dependencies**: KMS (data encryption), Audit (data access tracking)
**Rationale**: Regulatory compliance requirement (GDPR, HIPAA, PCI-DSS)

---

#### Group 28: Privacy (Privacy Engineering)

**Scope**: GDPR compliance, data anonymization, consent management, right-to-be-forgotten
**Market**: EU/UK businesses, privacy-conscious organizations
**Dependencies**: Audit (data access logs), Vault (encrypted user data)
**Rationale**: Privacy regulations increasing globally; privacy-by-design is competitive advantage

---

### Category: DevSecOps & Integration (5 groups)

#### Group 29: CI-CD (CI/CD Security Integration)

**Scope**: Pipeline security scanning, secret detection, signing artifacts, SBOM generation
**Market**: DevOps teams, software supply chain security
**Dependencies**: CA (code signing), Secrets (pipeline credentials)
**Rationale**: Shift-left security; DevSecOps adoption growing

---

#### Group 30: SBOM (Software Bill of Materials)

**Scope**: Dependency tracking, vulnerability scanning, license compliance, supply chain security
**Market**: Software vendors, regulated industries
**Dependencies**: CI-CD (artifact metadata), Audit (compliance reporting)
**Rationale**: Executive Order 14028 (US), supply chain attacks increasing

---

#### Group 31: Code-Signing (Code & Artifact Signing)

**Scope**: Binary signing, container image signing, firmware signing, Sigstore integration
**Market**: Software vendors, IoT manufacturers, automotive
**Dependencies**: CA (code signing certificates), PKI (signing infrastructure)
**Rationale**: Software supply chain security; trusted software distribution

---

#### Group 32: GitOps (GitOps Security)

**Scope**: Git-based secrets management, encrypted GitOps workflows, Flux/ArgoCD integration
**Market**: Cloud-native teams, Kubernetes operators
**Dependencies**: Secrets (encrypted configs), CI-CD (deployment automation)
**Rationale**: GitOps adoption growing; secure GitOps is critical for production

---

#### Group 33: API-Security (API Security & Testing)

**Scope**: API security testing, OWASP API Top 10, rate limiting, API firewall
**Market**: API-first companies, microservices architectures
**Dependencies**: Gateway (API protection), Monitoring (API analytics)
**Rationale**: APIs are primary attack surface; API security market growing rapidly

---

### Category: Emerging Technologies (5 groups)

#### Group 34: Blockchain (Blockchain Integration)

**Scope**: Private blockchain KMS, smart contract signing, DeFi key management, wallet security
**Market**: Cryptocurrency exchanges, DeFi platforms, NFT marketplaces
**Dependencies**: KMS (wallet keys), CA (blockchain certificates)
**Rationale**: Blockchain adoption growing; secure key management is critical pain point

---

#### Group 35: AI-ML (AI/ML Model Security)

**Scope**: Model encryption, federated learning, secure multi-party computation, model signing
**Market**: AI/ML platforms, healthcare AI, autonomous systems
**Dependencies**: KMS (model encryption keys), Code-Signing (model provenance)
**Rationale**: AI model theft and poisoning attacks; secure AI is emerging requirement

---

#### Group 36: Quantum-Safe (Quantum-Safe Transition)

**Scope**: Hybrid cryptography (classical + PQC), crypto-agility, migration tooling
**Market**: Government, defense, critical infrastructure
**Dependencies**: PQC (algorithms), KMS (hybrid keys)
**Rationale**: NIST PQC standardization complete; migration timelines starting

---

#### Group 37: Confidential (Confidential Computing)

**Scope**: Intel SGX, AMD SEV, ARM TrustZone integration, encrypted in-use data
**Market**: Financial services, healthcare, government
**Dependencies**: KMS (enclave keys), HSM (secure provisioning)
**Rationale**: Data-in-use protection is final frontier; confidential computing adoption accelerating

---

#### Group 38: Homomorphic (Homomorphic Encryption)

**Scope**: Fully homomorphic encryption (FHE), secure computation on encrypted data
**Market**: Healthcare analytics, financial modeling, privacy-preserving ML
**Dependencies**: KMS (FHE key management), AI-ML (encrypted ML)
**Rationale**: FHE performance improving; enables privacy-preserving analytics

---

### Category: Industry-Specific (5 groups)

#### Group 39: Financial (Financial Services Security)

**Scope**: PCI-DSS compliance, payment tokenization, fraud detection, SWIFT integration
**Market**: Banks, payment processors, fintech
**Dependencies**: KMS (payment keys), Compliance (PCI-DSS), Audit (transaction logs)
**Rationale**: Highly regulated industry; specific compliance requirements

---

#### Group 40: Healthcare (Healthcare Security & HIPAA)

**Scope**: HIPAA compliance, PHI encryption, patient consent management, medical device security
**Market**: Hospitals, health tech, pharma
**Dependencies**: KMS (PHI encryption), Privacy (consent), Compliance (HIPAA)
**Rationale**: Highly sensitive data; strict regulatory requirements

---

#### Group 41: Government (Government & Defense)

**Scope**: FedRAMP, FIPS 140-3, classified data handling, cross-domain solutions
**Market**: Federal agencies, defense contractors
**Dependencies**: HSM (FIPS compliance), PQC (future-proofing), Audit (classified logs)
**Rationale**: Stringent security requirements; long procurement cycles but stable market

---

#### Group 42: Energy (Energy & Critical Infrastructure)

**Scope**: ICS/SCADA security, grid security, smart meter key management
**Market**: Utilities, oil & gas, renewables
**Dependencies**: IoT (device management), Edge (field operations)
**Rationale**: Critical infrastructure protection; increasing cyber threats

---

#### Group 43: Retail (Retail & E-Commerce)

**Scope**: PCI-DSS for payments, customer data protection, loyalty program security
**Market**: Retailers, e-commerce platforms
**Dependencies**: Financial (payment security), Privacy (customer data)
**Rationale**: Large attack surface; frequent data breaches drive security investment

---

## Implementation Roadmap

### Phase 1: Consolidation (Tasks 1-9)

- Reorganize existing KMS, Identity, CA packages
- Establish import policies and documentation standards
- Update tooling and workflows

### Phase 2: Core Expansion (Tasks 10-13)

- Implement Secrets, Vault, PKI groups
- Deliver unified CLI across core services
- Enhance observability for new groups

### Phase 3: Infrastructure & Automation (Tasks 14-16)

- Build Automation, Gateway, Monitoring groups
- Integrate with CI/CD pipelines
- Establish backup and disaster recovery

### Phase 4: Governance & Compliance (Tasks 17-20)

- Implement Audit and Compliance groups
- Deliver SDKs for Go, Python, JavaScript
- Achieve SOC 2, ISO 27001 certifications

### Phase 5: Strategic Expansion (Tasks 21+)

- Prioritize adjacent-market groups based on customer demand
- Focus on HSM, PQC, IoT for enterprise adoption
- Expand cloud-native and multi-cloud capabilities

---

## Review Checklist

### Stakeholder Validation

- [ ] Engineering leadership approves taxonomy structure
- [ ] Product management aligns roadmap with service groups
- [ ] Security team validates compliance and audit requirements
- [ ] Operations team confirms observability and monitoring needs
- [ ] Sales/marketing validates adjacent-market priorities

### Requirements Alignment

- [ ] KMS requirements satisfied (hierarchical keys, FIPS 140-3)
- [ ] Identity requirements satisfied (OAuth 2.1, OIDC, domain isolation)
- [ ] CA requirements satisfied (20+ profiles, CA/Browser Forum compliance)
- [ ] Additional groups support strategic growth initiatives
- [ ] Adjacent-market groups address customer pain points

### Technical Validation

- [ ] Import boundaries defined and enforceable via lint rules
- [ ] Package naming conventions consistent with Go standards
- [ ] Dependencies mapped and coupling risks identified
- [ ] Migration paths defined for existing code
- [ ] Testing strategies defined for each group

### Documentation Standards

- [ ] Each group has clear scope and rationale
- [ ] Dependencies and consumers documented
- [ ] Roadmap aligns with business priorities
- [ ] Compliance and regulatory requirements captured
- [ ] Success metrics defined for each group

---

## Cross-References

- **KMS Details**: See `README.md`, `internal/server/`, `internal/client/`
- **Identity Details**: See `docs/03-identityV2/`, `internal/identity/`
- **CA Details**: See `docs/05-ca/README.md`
- **Refactor Plan**: See `docs/01-refactor/README.md` (Tasks 1-20)
- **Dependency Audit**: See `docs/03-identityV2/dependency-graph.md`

---

## Glossary

- **Service Group**: A cohesive set of packages, commands, and APIs addressing a specific functional domain
- **Mandatory Group**: Core functionality (KMS, Identity, CA)
- **Repo-Driven Group**: Strategic enhancements to existing capabilities (Secrets, Vault, PKI, etc.)
- **Adjacent-Market Group**: Expansion opportunities into related markets (HSM, PQC, IoT, etc.)
- **Domain Isolation**: Enforced separation via lint rules (e.g., Identity cannot import KMS)
- **FIPS 140-3**: US government cryptographic module validation standard
- **CA/Browser Forum**: Industry consortium defining TLS certificate standards
- **Zero-Trust**: Security model requiring continuous authentication and authorization
- **PQC**: Post-Quantum Cryptography (algorithms resistant to quantum computing attacks)
