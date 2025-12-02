# Grooming Session 03: KMS Key Management Deep Dive

## Overview

- **Focus Area**: Key Management Service, ElasticKey/MaterialKey operations, key hierarchy
- **Related Spec Section**: Spec P3: KMS, OpenAPI spec components/paths
- **Prerequisites**: Sessions 01-02 completed, understanding of cryptographic key management

---

## Questions

### Q1: What is an ElasticKey in the cryptoutil KMS?

A) A single cryptographic key
B) A policy container for versioned key material
C) A hardware security module key
D) An ephemeral session key

**Answer**: B
**Explanation**: ElasticKey is a policy container that holds configuration (algorithm, versioning policy, import policy) for multiple MaterialKey versions.

---

### Q2: What is the relationship between ElasticKey and MaterialKey?

A) One-to-one (each ElasticKey has one MaterialKey)
B) One-to-many (each ElasticKey can have multiple MaterialKeys)
C) Many-to-one (multiple ElasticKeys share one MaterialKey)
D) Many-to-many (ElasticKeys and MaterialKeys are independent)

**Answer**: B
**Explanation**: One ElasticKey can have multiple MaterialKey versions, enabling key rotation while maintaining the same policy.

---

### Q3: What HTTP method is used to create a new ElasticKey?

A) GET
B) PUT
C) POST
D) PATCH

**Answer**: C
**Explanation**: POST to `/api/v1/elastic-keys` creates a new ElasticKey with the specified policy.

---

### Q4: What is the purpose of the `versioning_allowed` flag on ElasticKey?

A) Enables automatic key rotation
B) Controls whether new MaterialKey versions can be created
C) Enables version history viewing
D) Controls encryption algorithm version

**Answer**: B
**Explanation**: When versioning_allowed=true, new MaterialKey versions can be created under the ElasticKey for rotation.

---

### Q5: What is the purpose of the `import_allowed` flag on ElasticKey?

A) Enables importing keys from other KMS systems
B) Enables exporting keys to other systems
C) Controls whether external key material can be imported
D) Enables key migration between databases

**Answer**: C
**Explanation**: import_allowed controls whether external key material can be imported as a new MaterialKey version.

---

### Q6: What endpoint is used to list MaterialKeys for a specific ElasticKey?

A) `/api/v1/material-keys`
B) `/api/v1/elastic-keys/{id}/material-keys`
C) `/api/v1/elastic-keys/{id}/versions`
D) `/api/v1/keys/{id}/materials`

**Answer**: B
**Explanation**: MaterialKeys are listed under their parent ElasticKey at `/api/v1/elastic-keys/{id}/material-keys`.

---

### Q7: How is a MaterialKey revoked?

A) DELETE request to `/api/v1/material-keys/{id}`
B) PUT request to `/api/v1/material-keys/{id}` with status=revoked
C) POST request to `/api/v1/material-keys/{id}/revoke`
D) PATCH request to `/api/v1/material-keys/{id}/status`

**Answer**: C
**Explanation**: POST to `/api/v1/material-keys/{id}/revoke` revokes a specific key version.

---

### Q8: What is the highest level in the cryptoutil key hierarchy?

A) Root keys
B) Unseal secrets
C) ElasticKeys
D) Content keys

**Answer**: B
**Explanation**: Unseal secrets (from file:///run/secrets/*) are the highest level, from which Root keys are derived.

---

### Q9: What is the purpose of Intermediate keys in the hierarchy?

A) Encrypt user data directly
B) Provide per-tenant isolation
C) Manage key rotation
D) Store audit logs

**Answer**: B
**Explanation**: Intermediate keys provide per-tenant isolation between Root keys and the ElasticKey/MaterialKey layer.

---

### Q10: Which status value indicates an ElasticKey is no longer usable?

A) inactive
B) suspended
C) deleted
D) revoked

**Answer**: C
**Explanation**: ElasticKey statuses are: active, suspended, deleted. Deleted keys are soft-deleted and no longer usable.

---

### Q11: What query parameter filters ElasticKeys by algorithm?

A) `alg`
B) `algorithm`
C) `key_algorithm`
D) `encryption_algorithm`

**Answer**: B
**Explanation**: The `algorithm` query parameter filters ElasticKeys by their cryptographic algorithm.

---

### Q12: How are multiple sort fields specified in API requests?

A) Comma-separated in single parameter
B) Repeated `sort` parameter for each field
C) JSON array in request body
D) Semicolon-separated in single parameter

**Answer**: B
**Explanation**: Multiple sort fields use repeated `sort` parameter: `?sort=name:asc&sort=created_at:desc`.

---

### Q13: What is the format for sort parameter values?

A) `fieldName_direction`
B) `fieldName:direction`
C) `direction(fieldName)`
D) `fieldName.direction`

**Answer**: B
**Explanation**: Sort format is `fieldName:direction` (e.g., `name:asc`, `created_at:desc`).

---

### Q14: What query parameter controls pagination page number?

A) `offset`
B) `page`
C) `pageNumber`
D) `skip`

**Answer**: B
**Explanation**: The `page` parameter controls which page of results to return, combined with `size` for page size.

---

### Q15: What happens when an ElasticKey is soft-deleted?

A) All data is immediately removed
B) Key is marked deleted but data remains
C) Only metadata is deleted
D) Key is moved to archive storage

**Answer**: B
**Explanation**: Soft delete marks the key status as "deleted" but preserves data for audit and potential recovery.

---

### Q16: Which provider types are supported for ElasticKeys?

A) HSM only
B) Software only
C) Configurable provider abstraction
D) Cloud KMS only

**Answer**: C
**Explanation**: ElasticKeys use a configurable provider abstraction that can support software, HSM, or cloud backends.

---

### Q17: What is the endpoint to update ElasticKey metadata?

A) POST `/api/v1/elastic-keys/{id}`
B) PUT `/api/v1/elastic-keys/{id}`
C) PATCH `/api/v1/elastic-keys/{id}`
D) UPDATE `/api/v1/elastic-keys/{id}`

**Answer**: B
**Explanation**: PUT to `/api/v1/elastic-keys/{id}` updates the ElasticKey metadata (name, status, etc.).

---

### Q18: What filtering is available for MaterialKey expiration dates?

A) Only exact date matching
B) Date range filtering (min/max)
C) Only filtering by year
D) No date filtering available

**Answer**: B
**Explanation**: MaterialKeys support min_expiration_date and max_expiration_date parameters for date range filtering.

---

### Q19: What is the relationship between key rotation and MaterialKeys?

A) Key rotation deletes old MaterialKeys
B) Key rotation creates new MaterialKey version under same ElasticKey
C) Key rotation requires new ElasticKey
D) Key rotation is not supported

**Answer**: B
**Explanation**: Key rotation creates a new MaterialKey version under the same ElasticKey, preserving the policy while using new key material.

---

### Q20: Which HTTP status indicates successful ElasticKey creation?

A) 200 OK
B) 201 Created
C) 204 No Content
D) 202 Accepted

**Answer**: B
**Explanation**: 201 Created indicates successful resource creation, with the new ElasticKey in the response body.

---

### Q21: What query parameter filters ElasticKeys by name?

A) `elastic_key_name`
B) `key_name`
C) `name`
D) `title`

**Answer**: C
**Explanation**: The `name` parameter filters ElasticKeys by their assigned name.

---

### Q22: How are multiple IDs specified when filtering ElasticKeys?

A) Comma-separated in single parameter
B) Repeated `elastic_key_id` parameter
C) JSON array in parameter
D) Pipe-separated in single parameter

**Answer**: B
**Explanation**: Multiple IDs use repeated parameter: `?elastic_key_id=uuid1&elastic_key_id=uuid2`.

---

### Q23: What is the purpose of the `provider` field on ElasticKey?

A) Specifies the cloud provider
B) Identifies the cryptographic provider/backend
C) Specifies the authentication provider
D) Identifies the key owner

**Answer**: B
**Explanation**: The provider field identifies which cryptographic backend (software, HSM, cloud) manages the key.

---

### Q24: What date fields are tracked for MaterialKeys?

A) Only creation date
B) Creation and expiration dates
C) Import date, expiration date, revocation date
D) Only revocation date

**Answer**: C
**Explanation**: MaterialKeys track import_date, expiration_date, and revocation_date for lifecycle management.

---

### Q25: What API operation imports external key material?

A) PUT `/api/v1/elastic-keys/{id}`
B) POST `/api/v1/elastic-keys/{id}/import`
C) POST `/api/v1/material-keys/import`
D) PUT `/api/v1/material-keys/{id}/import`

**Answer**: B
**Explanation**: POST to `/api/v1/elastic-keys/{id}/import` imports external key material as a new MaterialKey version.

---

### Q26: What must be true for key import to succeed?

A) ElasticKey must be in active status
B) ElasticKey must have import_allowed=true
C) No other MaterialKeys can exist
D) Key must be in PEM format

**Answer**: B
**Explanation**: The ElasticKey's import_allowed flag must be true to accept imported key material.

---

### Q27: How does the KMS handle concurrent key operations?

A) Locks the entire database
B) Uses connection pooling and transactions
C) Queues all operations sequentially
D) Rejects concurrent requests

**Answer**: B
**Explanation**: KMS uses proper database connection pooling and transactions to handle concurrent operations safely.

---

### Q28: What is the recommended approach for key rotation testing?

A) Test in production
B) Test key rotation scenarios in integration tests
C) Skip rotation testing
D) Only manual testing

**Answer**: B
**Explanation**: The plan includes "Test key rotation scenarios" as a HIGH priority integration testing task.

---

### Q29: What query parameter controls the number of results per page?

A) `limit`
B) `count`
C) `size`
D) `pageSize`

**Answer**: C
**Explanation**: The `size` parameter controls the number of results per page in paginated responses.

---

### Q30: What is the demo command to run KMS demonstration?

A) `go run ./cmd/cryptoutil demo`
B) `go run ./cmd/demo kms`
C) `./cryptoutil demo kms`
D) `go run ./cmd/kms demo`

**Answer**: B
**Explanation**: `go run ./cmd/demo kms` runs the KMS demonstration.

---

### Q31: Which ElasticKey fields can be updated after creation?

A) Algorithm and provider
B) Name and status
C) All fields except ID
D) No fields can be updated

**Answer**: B
**Explanation**: Metadata fields like name and status can be updated. Algorithm and provider are typically immutable after creation.

---

### Q32: What is multi-tenant isolation in the KMS context?

A) Running multiple KMS instances
B) Using Intermediate keys for tenant separation
C) Using different databases per tenant
D) Network-level isolation only

**Answer**: B
**Explanation**: Multi-tenant isolation is achieved through Intermediate keys that separate different tenants' key hierarchies.

---

### Q33: What sort directions are supported?

A) asc only
B) desc only
C) asc and desc
D) ascending, descending, natural

**Answer**: C
**Explanation**: Sort directions are `asc` (ascending) and `desc` (descending).

---

### Q34: What is the purpose of the Swagger documentation endpoint?

A) Authenticate API requests
B) Provide OpenAPI specification for API exploration
C) Generate client code
D) Validate request schemas

**Answer**: B
**Explanation**: `/ui/swagger/doc.json` provides the OpenAPI specification for API documentation and exploration.

---

### Q35: What happens to MaterialKeys when an ElasticKey is deleted?

A) MaterialKeys are immediately deleted
B) MaterialKeys remain accessible
C) MaterialKeys are soft-deleted with the ElasticKey
D) MaterialKeys are moved to another ElasticKey

**Answer**: C
**Explanation**: MaterialKeys follow their parent ElasticKey's lifecycle - soft deletion cascades logically.

---

### Q36: What filtering is available for MaterialKey revocation dates?

A) Only exact date matching
B) Range filtering with min/max revocation date
C) No revocation date filtering
D) Only filtering for revoked vs non-revoked

**Answer**: B
**Explanation**: MaterialKeys support min_revocation_date and max_revocation_date for filtering by revocation date range.

---

### Q37: What is the primary key type for KMS entities?

A) Auto-increment integers
B) UUIDv4
C) UUIDv7
D) String identifiers

**Answer**: C
**Explanation**: UUIDv7 is used for primary keys throughout cryptoutil for time-ordered, unique identifiers.

---

### Q38: How are ElasticKey listing results sorted by default?

A) By name ascending
B) By creation date descending
C) By ID ascending
D) No default sort (unordered)

**Answer**: D
**Explanation**: Without explicit sort parameters, results may be unordered. Explicit sorting is recommended.

---

### Q39: What is the integration demo target for KMS with Identity?

A) KMS authenticates Identity users
B) KMS authenticated by Identity (OAuth2)
C) Bidirectional authentication
D) No integration planned

**Answer**: B
**Explanation**: The integration demo shows "KMS authenticated by Identity" - KMS as an OAuth2 resource server validated by Identity.

---

### Q40: What must be configured for KMS to validate Identity tokens?

A) Shared database access
B) JWKS endpoint and token validation middleware
C) Direct password sharing
D) Network-level security only

**Answer**: B
**Explanation**: KMS must fetch JWKS from Identity and implement token validation middleware to verify JWT access tokens.

---

### Q41: What is the purpose of per-IP rate limiting in KMS?

A) Track usage statistics
B) Prevent abuse and DDoS attacks
C) Manage licensing
D) Enable caching

**Answer**: B
**Explanation**: Per-IP rate limiting prevents API abuse and provides protection against denial-of-service attacks.

---

### Q42: What ElasticKey status allows normal operations?

A) active
B) enabled
C) valid
D) operational

**Answer**: A
**Explanation**: ElasticKey status "active" allows normal key operations. "suspended" and "deleted" restrict operations.

---

### Q43: How does KMS handle audit logging?

A) Logs to separate audit database
B) Integrated with telemetry/observability stack
C) No audit logging implemented
D) Logs to filesystem only

**Answer**: B
**Explanation**: KMS audit logging integrates with the telemetry/observability infrastructure (OpenTelemetry, Grafana).

---

### Q44: What is the demo success criterion for KMS?

A) `go build` passes
B) All 7 demo steps complete
C) 80% test coverage
D) Zero linting errors

**Answer**: B
**Explanation**: Success requires `go run ./cmd/demo all` completing 7/7 steps including KMS, Identity, and Integration.

---

### Q45: What database configuration is required for KMS with SQLite?

A) MaxOpenConns=1 for all cases
B) MaxOpenConns=1 (KMS uses database/sql)
C) MaxOpenConns=5 for all cases
D) No connection limits

**Answer**: B
**Explanation**: KMS uses database/sql (not GORM transactions) so MaxOpenConns=1 is appropriate. GORM-based Identity needs MaxOpenConns=5.

---

### Q46: What is the correct content type for KMS API requests?

A) text/plain
B) application/xml
C) application/json
D) application/x-www-form-urlencoded

**Answer**: C
**Explanation**: KMS API uses application/json for request and response bodies.

---

### Q47: What is the purpose of key versioning?

A) Track API version changes
B) Enable key rotation without changing policies
C) Version control for configurations
D) Audit log versioning

**Answer**: B
**Explanation**: Key versioning (MaterialKey versions) enables rotation while maintaining the same ElasticKey policy.

---

### Q48: How are KMS errors returned to clients?

A) Plain text messages
B) HTTP status codes only
C) Structured JSON error responses
D) XML error documents

**Answer**: C
**Explanation**: KMS returns structured JSON error responses with error codes and descriptions.

---

### Q49: What is required for key import validation?

A) Only format validation
B) Algorithm compatibility, format, and policy checks
C) No validation required
D) Only size validation

**Answer**: B
**Explanation**: Key import validates algorithm compatibility with ElasticKey policy, proper format, and other policy constraints.

---

### Q50: What is the KMS service public API port?

A) 8080 (SQLite), 8081/8082 (PostgreSQL)
B) 9090 for all backends
C) 8090 for all backends
D) 443 for all backends

**Answer**: A
**Explanation**: KMS uses 8080 for SQLite backend, 8081/8082 for PostgreSQL backends. 9090 is the admin API port.

---

## Session Summary

**Topics Covered**:

- ElasticKey and MaterialKey concepts
- Key hierarchy (Unseal → Root → Intermediate → ElasticKey → MaterialKey)
- CRUD operations for keys
- Filtering, sorting, and pagination
- Key rotation and versioning
- Import/export capabilities
- Multi-tenant isolation
- Integration with Identity for authentication

**Next Session**: GROOMING-SESSION-04 - Certificate Authority Planning
