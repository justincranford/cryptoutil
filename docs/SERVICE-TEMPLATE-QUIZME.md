# SERVICE-TEMPLATE Implementation Blockers - QUIZME

## Critical Unknown: Barrier Layer Encryption Service Location

**Question**: Where should the barrier layer encryption service be implemented?

A) `internal/shared/crypto/barrier/` - Shared infrastructure package accessible to all services
B) `internal/learn/server/barrier/` - Learn-specific implementation to be extracted later
C) `internal/kms/barrier/` - KMS service owns encryption, other services call KMS APIs
D) `pkg/barrier/` - Public library for external consumption
E) Write-in: _____________________

**Rationale**: Instructions state `internal/learn/server/server.go` provides services to ALL products, but barrier encryption is cryptographic infrastructure that may belong in KMS (sm-kms service).

E It must be shared infrastructure; use internal\shared\barrier; i thought it was already extracted for use in internal/template/server
---

## Critical Unknown: JwkGenService Ownership

**Question**: Should JwkGenService be owned by learn-im or jose-ja service?

A) `internal/learn/server/jwkgen/` - learn-im owns JWK generation for all services
B) `internal/jose/server/jwkgen/` - jose-ja service owns JWK generation (aligned with JOSE product)
C) `internal/shared/jwk/` - Shared infrastructure package
D) Dual implementation - learn-im has basic version, jose-ja has full JOSE compliance
E) Write-in: _____________________

E it already exists in internal\shared\crypto\jose; it is shared infrastructure; use that

**Rationale**: JOSE product (jose-ja service) exists for JWK/JWKS/JWE/JWS operations. Duplicating JWK generation in learn-im creates architectural ambiguity.

---

## Critical Unknown: Service Federation Pattern

**Question**: How should learn-im consume barrier encryption and JWK services?

A) Direct Go package imports - `internal/learn/server` imports `internal/kms/barrier` and `internal/jose/jwkgen`
B) REST API calls - learn-im calls sm-kms and jose-ja HTTP APIs with service-to-service auth
C) gRPC - High-performance RPC for internal service communication
D) Embedded services - learn-im embeds barrier and JWK services in-process
E) Write-in: _____________________

**Rationale**: Federation pattern impacts architecture significantly. Direct imports violate microservices isolation. REST/gRPC require network configuration and failure handling.

D use shared infrastructure packages directly; avoid network calls for performance and simplicity

---

## Critical Unknown: Database Schema Migration Strategy

**Question**: How should the 4 new tables (`users_jwks`, `users_messages_jwks`, `messages_jwks`, `messages`) be created?

A) Embedded SQL migrations in `internal/learn/migrations/` (golang-migrate pattern)
B) GORM AutoMigrate during service startup
C) Manual SQL scripts in `deployments/compose/learn/migrations/`
D) Code-first models with GORM tags, migrations generated later
E) Write-in: _____________________

**Rationale**: Existing services use embedded migrations (`//go:embed migrations/*.sql`). Consistency required.

E use same design as KMS

---

## Critical Unknown: JWE Message Format

**Question**: What JWE serialization format should `messages` table use?

A) JWE Compact Serialization - Single string `eyJ...` (URL-safe, recommended for storage)
B) JWE JSON Serialization - Full JSON object with protected/unprotected headers
C) Custom binary format - Optimized for database storage
D) Base64-encoded protobuf - Cross-language compatibility
E) Write-in: _____________________

**Rationale**: JWE RFC 7516 supports multiple serialization formats. Database storage should use compact serialization (single TEXT column).

A if possible

---

## Critical Unknown: Key Rotation Strategy

**Question**: How should JWKs in database tables be rotated?

A) Elastic Key pattern - Active key for encryption, historical keys for decryption (key ID embedded)
B) Time-based rotation - Rotate all keys every N days, re-encrypt all messages
C) Manual rotation - Admin API triggers rotation on-demand
D) No rotation - Keys are immutable once created
E) Write-in: _____________________

**Rationale**: See `02-07.cryptography.instructions.md` Elastic Key Rotation section. Messages encrypted with old keys must remain decryptable.

C

---

## Critical Unknown: AESGCMKW vs Direct AESGCM

**Question**: Why use AESGCMKW (AES-GCM Key Wrap) for `users_messages_jwks` and `messages_jwks` instead of direct AESGCM?

A) Key wrapping provides additional security layer - Content Encryption Key (CEK) wrapped with Key Encryption Key (KEK)
B) Performance optimization - AESGCMKW is faster for small messages
C) JWE standard requirement - AESGCMKW required for multi-recipient scenarios
D) Mistake - Should use direct AESGCM for all tables
E) Write-in: _____________________

**Rationale**: Understanding crypto protocol selection is critical. AESGCMKW is JWE key management algorithm, AESGCM is content encryption algorithm.

D

---

## Critical Unknown: ECDH P-256 Usage

**Question**: How is ECDH P-256 used with AESGCM256 in `users_jwks` table?

A) ECDH key agreement derives shared secret, used as AES-256 key for content encryption
B) ECDH is the encryption algorithm, AESGCM is the authentication algorithm
C) ECDH P-256 signs the JWK, AESGCM encrypts the message
D) Hybrid encryption - ECDH for key exchange, AES-GCM for bulk encryption
E) Write-in: _____________________

**Rationale**: ECDH is key agreement protocol, not encryption. Clarify how it integrates with AESGCM.

A JWK requires two headers `alg` and `enc` for both key agreement and content encryption algorithms

---

## Critical Unknown: Version Flag Removal Impact

**Question**: Removing version flag from learn-im parameters - what replaces it for build tracking?

A) Git commit hash embedded via `-ldflags "-X main.version=$(git rev-parse HEAD)"`
B) Build timestamp embedded via `-ldflags "-X main.buildDate=$(date)"`
C) Semantic versioning from `go.mod` or separate VERSION file
D) No version tracking needed - Docker image tags provide versioning
E) Write-in: _____________________

**Rationale**: Version tracking is mandatory for support, debugging, and compliance. Removal must have replacement strategy.

D

---

## Critical Unknown: Package Import Cycles

**Question**: If `internal/learn/server` provides services to ALL products, how to avoid import cycles?

A) Interface-based dependency injection - Services define interfaces, implementations injected at runtime
B) Service mesh - All services communicate via network APIs, no direct imports
C) Shared interfaces package - `internal/shared/interfaces/` defines contracts
D) Restrict to one-way dependency - Only learn-im provides services, never consumes from other domains
E) Write-in: _____________________

**Rationale**: Go forbids import cycles. If `internal/kms` imports `internal/learn/server`, and `internal/learn/server` imports `internal/kms`, build fails.

E don't create cycles; use shared infrastructure only one way
