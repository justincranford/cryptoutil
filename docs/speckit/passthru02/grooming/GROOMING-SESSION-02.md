# Grooming Session 02: JOSE Authority & CA Server APIs

## Overview

- **Focus Area**: JOSE Authority standalone service and CA Server REST API implementation
- **Related Spec Section**: P1 JOSE Authority API, P4 CA Server REST API
- **Prerequisites**: Understanding of JOSE standards (RFC 7515-7520), PKI concepts, REST API design

## Questions

### Q1: What is the primary architectural goal for the JOSE Authority service in Iteration 2?

A) Migrate existing JOSE primitives from internal/common/crypto/jose/ to cmd/jose-server/
B) Create standalone JOSE service at internal/jose/ with REST API while keeping embedded primitives
C) Replace all embedded JOSE usage with external API calls
D) Combine JOSE Authority with Identity Server for better integration

**Answer**: B
**Explanation**: The JOSE Authority serves dual purposes: embedded library for P2/P3/P4 AND standalone service. This maintains existing embedded usage while exposing REST APIs.

---

### Q2: Which JOSE Authority endpoint should handle JWK generation requests?

A) `/v1/jose/keys` POST
B) `/jose/v1/keys` POST
C) `/api/v1/jwk/generate` POST
D) `/jose/keys/generate` POST

**Answer**: B
**Explanation**: Following API versioning best practices, the correct pattern is `/jose/v1/keys` POST for JWK generation.

---

### Q3: What authentication method should the CA Server REST API primarily use?

A) OAuth 2.1 Bearer tokens only
B) API keys with IP allowlisting
C) mTLS client certificate authentication
D) HTTP Basic Auth

**Answer**: C
**Explanation**: mTLS is the primary authentication for CA operations, providing strong certificate-based authentication appropriate for PKI services.

---

### Q4: Which RFC standard governs the EST protocol endpoints being added to the CA Server?

A) RFC 5280 (X.509)
B) RFC 7030 (EST)
C) RFC 6960 (OCSP)
D) RFC 3161 (TSP)

**Answer**: B
**Explanation**: RFC 7030 defines Enrollment over Secure Transport (EST) protocol for certificate enrollment operations.

---

### Q5: What is the correct endpoint pattern for CA certificate issuance from CSR?

A) `/ca/v1/enroll` POST
B) `/ca/v1/certificate` POST
C) `/ca/v1/issue` POST
D) `/ca/v1/csr/submit` POST

**Answer**: B
**Explanation**: The spec defines `/ca/v1/certificate` POST for certificate issuance from CSR, following RESTful resource naming.

---

### Q6: How many JOSE Authority API endpoints must be implemented in Iteration 2?

A) 6 endpoints (keys, sign, verify, encrypt, decrypt, jwt)
B) 8 endpoints (keys, jwks, sign, verify, encrypt, decrypt, jwt issue/validate)
C) 10 endpoints (all JOSE operations plus administrative endpoints)
D) 12 endpoints (JOSE operations plus health/metrics endpoints)

**Answer**: C
**Explanation**: The spec lists 10 JOSE Authority endpoints: keys (POST/GET/LIST), jwks, sign, verify, encrypt, decrypt, jwt (issue/validate).

---

### Q7: What is the primary purpose of the `/ca/v1/est/cacerts` endpoint?

A) Download the current Certificate Revocation List
B) Submit a Certificate Signing Request
C) Get CA certificates per EST protocol
D) Revoke an existing certificate

**Answer**: C
**Explanation**: The EST cacerts endpoint returns CA certificates, allowing clients to obtain trust anchors per RFC 7030.

---

### Q8: Which database backends must both JOSE Authority and CA Server support?

A) PostgreSQL only (production focus)
B) SQLite only (development simplicity)
C) PostgreSQL and SQLite (cross-database compatibility)
D) MongoDB and PostgreSQL (document vs relational)

**Answer**: C
**Explanation**: Constitution requires both PostgreSQL (prod/dev) and SQLite (dev/test) support for all products.

---

### Q9: What is the expected Docker Compose integration for JOSE Authority?

A) Single service with SQLite only
B) Three services: jose-sqlite, jose-postgres-1, jose-postgres-2
C) Two services: jose-dev (SQLite) and jose-prod (PostgreSQL)
D) Four services: one per supported algorithm type

**Answer**: B
**Explanation**: Following the established pattern from KMS, JOSE needs three service variants for testing and demonstration.

---

### Q10: Which FIPS 140-3 signing algorithms must the JOSE Authority support?

A) RS256, RS384, RS512 (RSA only)
B) ES256, ES384, ES512 (ECDSA only)
C) PS256, PS384, PS512, RS256, RS384, RS512, ES256, ES384, ES512, EdDSA
D) All RSA and ECDSA variants plus experimental algorithms

**Answer**: C
**Explanation**: The spec lists all approved FIPS algorithms: PSS variants, PKCS#1 v1.5 variants, ECDSA variants, and EdDSA.

---

### Q11: What is the correct approach for JOSE Authority key storage?

A) Always store private keys in database for persistence
B) Generate ephemeral keys per request (no storage)
C) Optional database storage with in-memory fallback
D) File-based key storage only

**Answer**: C
**Explanation**: JOSE Authority supports optional DB storage for key persistence while allowing ephemeral operation for stateless deployments.

---

### Q12: How should the CA Server handle OCSP responder functionality?

A) Implement full OCSP responder in `/ca/v1/ocsp` endpoint
B) Delegate to external OCSP service
C) OCSP not supported in Iteration 2
D) OCSP via CRL download only

**Answer**: A
**Explanation**: The CA Server includes `/ca/v1/ocsp` POST endpoint for RFC 6960 OCSP responder functionality.

---

### Q13: What is the primary difference between JOSE embedded library and JOSE Authority service?

A) Different algorithms supported
B) Different key formats used
C) Embedded has no external dependencies, Authority has REST API
D) Authority service is more secure

**Answer**: C
**Explanation**: The embedded library provides internal crypto primitives, while the Authority service exposes the same capabilities via REST API.

---

### Q14: Which EST endpoint requires PKCS#7/CMS encoding support?

A) `/ca/v1/est/cacerts`
B) `/ca/v1/est/simpleenroll`
C) `/ca/v1/est/serverkeygen`
D) All EST endpoints

**Answer**: C
**Explanation**: EST serverkeygen endpoint requires PKCS#7/CMS encoding for server-side key generation responses per RFC 7030.

---

### Q15: What is the correct HTTP method for JWT validation in JOSE Authority?

A) GET with JWT in query parameter
B) POST with JWT in request body
C) PUT with JWT in Authorization header
D) HEAD with JWT in custom header

**Answer**: B
**Explanation**: The `/jose/v1/jwt/validate` endpoint uses POST method with JWT payload in request body for validation.

---

### Q16: How should CA Server handle certificate serial number generation?

A) Incremental integers starting from 1
B) UUIDv4 converted to integer
C) ≥64 bits CSPRNG, non-sequential, >0, <2^159
D) Hash of certificate content

**Answer**: C
**Explanation**: CA/Browser Forum requirements mandate cryptographically secure, non-sequential serial numbers within specified bit ranges.

---

### Q17: What authentication methods should CA Server REST API support beyond mTLS?

A) mTLS only (maximum security)
B) mTLS + OAuth 2.1 JWT Bearer + API keys
C) mTLS + Basic Auth
D) API keys only

**Answer**: B
**Explanation**: The spec lists three authentication methods: mTLS (primary), JWT Bearer (delegated), and API keys (automated systems).

---

### Q18: Which JOSE Authority endpoint provides public key discovery?

A) `/jose/v1/keys` GET
B) `/jose/v1/jwks` GET
C) `/jose/v1/public` GET
D) `/.well-known/jwks.json` GET

**Answer**: B
**Explanation**: The `/jose/v1/jwks` GET endpoint provides JSON Web Key Set for public key discovery, distinct from individual key management.

---

### Q19: What is the correct approach for CA Server certificate profile validation?

A) Use hardcoded certificate templates
B) Support 25+ configurable certificate profiles
C) Single universal certificate profile
D) Profile validation disabled for flexibility

**Answer**: B
**Explanation**: The CA implementation includes 25+ predefined certificate profiles with configurable certificate profile engine.

---

### Q20: How should JOSE Authority handle algorithm agility requirements?

A) Fixed algorithms per endpoint
B) Algorithm negotiation via HTTP headers
C) Configurable algorithms with FIPS-approved defaults
D) Client-specified algorithms without validation

**Answer**: C
**Explanation**: Constitution requires algorithm agility with FIPS-approved defaults for all cryptographic operations.

---

### Q21: What is the purpose of the `/ca/v1/profiles` GET endpoint?

A) Download CA certificate profiles
B) List available certificate profiles
C) Create new certificate profiles
D) Validate certificate against profiles

**Answer**: B
**Explanation**: The endpoint lists available certificate profiles that can be used for certificate issuance requests.

---

### Q22: Which Docker services should be included in CA Server deployment?

A) ca-server only
B) ca-server + postgresql
C) ca-sqlite + ca-postgres-1 + ca-postgres-2
D) ca-server + postgresql + redis + nginx

**Answer**: C
**Explanation**: Following the established pattern, CA Server needs three service variants: SQLite and two PostgreSQL instances.

---

### Q23: What is the correct content type for CSR submission to EST simpleenroll?

A) application/json only
B) application/pkcs10 only
C) application/pkcs10, application/base64, application/pem
D) multipart/form-data

**Answer**: C
**Explanation**: EST protocol supports multiple CSR encodings: DER (pkcs10), Base64, and PEM formats for flexibility.

---

### Q24: How should JOSE Authority handle concurrent key generation requests?

A) Sequential processing to avoid conflicts
B) Use keygen package pools for concurrent operations
C) Lock-based synchronization
D) Refuse concurrent requests

**Answer**: B
**Explanation**: The project uses keygen package pools for efficient concurrent cryptographic key generation operations.

---

### Q25: What is the primary testing strategy for both JOSE Authority and CA Server?

A) Unit tests only for speed
B) Integration tests with external services
C) Table-driven tests with t.Parallel(), TestMain for dependencies
D) Manual testing only

**Answer**: C
**Explanation**: Testing instructions mandate table-driven tests with t.Parallel() and TestMain pattern for shared test dependencies.

---

### Q26: Which endpoint should handle CA certificate chain retrieval?

A) `/ca/v1/ca` GET
B) `/ca/v1/ca/{ca_id}` GET
C) `/ca/v1/chain/{ca_id}` GET
D) `/ca/v1/certificates/chain` GET

**Answer**: B
**Explanation**: The `/ca/v1/ca/{ca_id}` GET endpoint returns CA details and certificate chain per the API specification.

---

### Q27: What is the correct approach for JOSE Authority error handling?

A) Return HTTP 500 for all errors
B) Use application error mapping (toAppErr) with proper HTTP status codes
C) Return HTTP 200 with error in JSON payload
D) Use HTTP status codes only without error details

**Answer**: B
**Explanation**: Database instructions require proper error mapping from internal errors to application HTTP errors with appropriate status codes.

---

### Q28: How should CA Server handle CRL generation requests?

A) Generate CRL on every request
B) Cache CRL with configurable expiration
C) Return pre-generated static CRL
D) CRL not supported in Iteration 2

**Answer**: B
**Explanation**: CRL generation should balance freshness requirements with performance via caching mechanisms with appropriate expiration.

---

### Q29: What is the required coverage target for JOSE Authority and CA Server code?

A) 80% minimum coverage
B) 90% production code coverage
C) 95% production, 100% infrastructure, 100% utility
D) 100% coverage for all code

**Answer**: C
**Explanation**: Quality requirements specify ≥95% production, ≥100% infrastructure, ≥100% utility coverage targets.

---

### Q30: Which observability features must both services include?

A) Logging only
B) Metrics only via Prometheus
C) OpenTelemetry instrumentation (OTLP traces, metrics, logs)
D) Health endpoints only

**Answer**: C
**Explanation**: Infrastructure requirements mandate OpenTelemetry instrumentation with OTLP export for comprehensive observability.

---

### Q31: What is the correct approach for CA Server certificate revocation?

A) Immediate certificate deletion
B) Update certificate status + CRL regeneration + OCSP update
C) Mark certificate as revoked without CRL update
D) Revocation not supported

**Answer**: B
**Explanation**: Proper certificate revocation requires status update, CRL regeneration, and OCSP responder notification for complete lifecycle management.

---

### Q32: How should JOSE Authority handle JWE (encryption) operations?

A) Support content encryption algorithms only
B) Support key wrapping + content encryption with FIPS algorithms
C) Encryption not supported for security reasons
D) Use external encryption services

**Answer**: B
**Explanation**: JOSE Authority must support both key wrapping (RSA-OAEP, AES-KW, ECDH-ES) and content encryption (AES-GCM, AES-CBC) per FIPS requirements.

---

### Q33: What is the primary configuration source for both services?

A) Environment variables only
B) Command line arguments only
C) YAML files with CLI override support
D) Database-stored configuration

**Answer**: C
**Explanation**: Configuration instructions specify YAML files as primary source with CLI flag override capability, avoiding environment variables for secrets.

---

### Q34: Which networking requirements apply to both JOSE Authority and CA Server?

A) HTTP only for development simplicity
B) HTTPS with TLS 1.3+ minimum, HTTP/2 support
C) HTTPS with any TLS version for compatibility
D) Mixed HTTP/HTTPS based on endpoint sensitivity

**Answer**: B
**Explanation**: Networking requirements mandate HTTPS with TLS 1.3+ minimum and HTTP/2 support for all production services.

---

### Q35: How should both services handle rate limiting?

A) No rate limiting for maximum throughput
B) Global rate limiting across all endpoints
C) Per-IP rate limiting for abuse prevention
D) Per-user rate limiting only

**Answer**: C
**Explanation**: Infrastructure components include rate limiting per IP for abuse prevention while allowing legitimate high-volume usage.

---

### Q36: What is the correct approach for health endpoint implementation?

A) Single `/health` endpoint for all checks
B) Separate `/livez` and `/readyz` endpoints on admin port 9090
C) Health checks via main API endpoints only
D) External health monitoring service

**Answer**: B
**Explanation**: The spec defines private admin API on port 9090 with `/livez` (liveness) and `/readyz` (readiness) endpoints for health monitoring.

---

### Q37: Which container image requirements apply to both services?

A) Alpine Linux base images only
B) Multi-stage builds with static linking, CGO disabled
C) Ubuntu base images for compatibility
D) Scratch images for minimal size

**Answer**: B
**Explanation**: Docker instructions specify multi-stage builds with static linking and CGO disabled for maximum portability and security.

---

### Q38: What is the correct approach for testing cryptographic operations?

A) Unit tests only for speed
B) Unit tests + integration tests + fuzz tests + benchmarks
C) Manual testing only for security validation
D) External penetration testing only

**Answer**: B
**Explanation**: Testing instructions mandate comprehensive testing including unit, integration, fuzz (for parsers/validators), and benchmark tests for crypto operations.

---

### Q39: How should both services handle graceful shutdown?

A) Immediate process termination
B) Complete current requests + close listeners + database cleanup
C) Wait for all background tasks indefinitely
D) Graceful shutdown not required

**Answer**: B
**Explanation**: Architecture requirements include graceful shutdown with request completion, listener cleanup, and proper resource disposal.

---

### Q40: What is the primary difference between EST and SCEP protocols for CA operations?

A) EST is newer and more secure than SCEP
B) SCEP supports more certificate types
C) EST uses HTTPS transport, SCEP uses proprietary transport
D) No functional differences

**Answer**: A
**Explanation**: EST (RFC 7030) is the modern replacement for SCEP, using HTTPS transport with improved security characteristics.

---

### Q41: Which JOSE algorithms require ECDH key agreement support?

A) All signing algorithms
B) ECDH-ES key agreement algorithms only
C) ECDH-ES, ECDH-ES+A128KW, ECDH-ES+A192KW, ECDH-ES+A256KW
D) All encryption algorithms

**Answer**: C
**Explanation**: The spec lists specific ECDH variants that require key agreement support for JOSE key management operations.

---

### Q42: What is the correct mutation testing target for both services?

A) 60% gremlins score minimum
B) 70% gremlins score minimum
C) ≥80% gremlins score per package
D) 90% gremlins score minimum

**Answer**: C
**Explanation**: Quality requirements specify ≥80% gremlins score per package for mutation testing to ensure test quality.

---

### Q43: How should CA Server handle time-stamping service integration?

A) External TSA service required
B) Built-in RFC 3161 TSA functionality via `/ca/v1/tsa/timestamp`
C) Time-stamping not supported
D) Database timestamp fields only

**Answer**: B
**Explanation**: The CA implementation includes RFC 3161 Time-Stamp Authority functionality as an integrated service endpoint.

---

### Q44: What is the correct approach for JOSE Authority JWT claims validation?

A) Accept all claims without validation
B) Validate signature only, ignore claims
C) Validate signature + standard claims (exp, iat, nbf, iss, aud, sub)
D) Custom claims validation only

**Answer**: C
**Explanation**: Proper JWT validation requires both signature verification and standard claims validation for security compliance.

---

### Q45: Which database transaction patterns should both services implement?

A) Auto-commit for simplicity
B) Explicit transactions for write operations, read-only for queries
C) Always use transactions for ACID guarantees
D) No transactions for maximum performance

**Answer**: C
**Explanation**: Database instructions emphasize proper transaction usage, error mapping, and ACID guarantees for data integrity.

---

### Q46: What is the correct approach for API versioning in both services?

A) No versioning for simplicity
B) Version in URL path: `/v1/`, `/v2/`
C) Version in HTTP headers only
D) Version in query parameters

**Answer**: B
**Explanation**: The API specifications consistently use path-based versioning (e.g., `/jose/v1/`, `/ca/v1/`) for clear version management.

---

### Q47: How should both services handle CORS configuration?

A) CORS disabled for security
B) CORS enabled for all origins (*)
C) Configurable CORS with security defaults
D) CORS configuration not required

**Answer**: C
**Explanation**: Networking requirements include CORS protection with configurable origins for secure cross-origin API access.

---

### Q48: What is the primary logging strategy for both services?

A) Console output only
B) File-based logging only
C) Structured logging with OpenTelemetry integration
D) Syslog integration only

**Answer**: C
**Explanation**: Observability requirements mandate structured logging with OpenTelemetry instrumentation for centralized log aggregation.

---

### Q49: Which deployment orchestration should both services support?

A) Docker Compose only
B) Kubernetes only
C) Docker Compose + Docker Swarm + Kubernetes manifests
D) Bare metal deployment only

**Answer**: A
**Explanation**: Current implementation focuses on Docker Compose deployments with service mesh capabilities for production readiness.

---

### Q50: What is the correct approach for handling configuration validation in both services?

A) Runtime validation during operation
B) Startup validation with fail-fast behavior
C) No validation for flexibility
D) Validation only in development mode

**Answer**: B
**Explanation**: Configuration requirements mandate validation on startup with fail-fast behavior to prevent runtime configuration errors.

---

## Answer Summary

| Q# | Answer | Q# | Answer | Q# | Answer | Q# | Answer | Q# | Answer |
|----|--------|----|--------|----|--------|----|--------|----|--------|
| 1  | B      | 11 | C      | 21 | B      | 31 | B      | 41 | C      |
| 2  | B      | 12 | A      | 22 | C      | 32 | B      | 42 | C      |
| 3  | C      | 13 | C      | 23 | C      | 33 | C      | 43 | B      |
| 4  | B      | 14 | C      | 24 | B      | 34 | B      | 44 | C      |
| 5  | B      | 15 | B      | 25 | C      | 35 | C      | 45 | C      |
| 6  | C      | 16 | C      | 26 | B      | 36 | B      | 46 | B      |
| 7  | C      | 17 | B      | 27 | B      | 37 | B      | 47 | C      |
| 8  | C      | 18 | B      | 28 | B      | 38 | B      | 48 | C      |
| 9  | B      | 19 | B      | 29 | C      | 39 | B      | 49 | A      |
| 10 | C      | 20 | C      | 30 | C      | 40 | A      | 50 | B      |

---

*Grooming Session Version: 2.0.0*
*Generated for: cryptoutil Iteration 2*
*Focus: JOSE Authority standalone service and CA Server REST API*
