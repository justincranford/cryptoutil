# Hash Registry and Password Hashing Specifications

**Version**: 1.0.0  
**Last Updated**: 2025-12-24  
**Referenced By**: `.github/instructions/02-08.hashes.instructions.md`

## Overview

Use PBKDF2 for low-entropy inputs like PII (e.g. username, email, IP address) and secrets (e.g. passwords).

Use HKDF for high-entropy inputs like blobs (e.g. secret configuration) and secrets (e.g. API keys).

### Entropy Thresholds

**Low-entropy**: Anything below 128-bits entropy
**High-entropy**: Anything with 128-bits entropy or higher

**Practical Definition**: 128-bits entropy means the input MUST have a 256-bit (or higher) search space to brute force it; entropy bit count is half the search space.

## Pepper Requirements - MANDATORY

**MANDATORY: All inputs must be peppered before input into hash functions**

**Reference**: OWASP Password Storage Cheat Sheet - <https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#peppering>

### Pepper Storage Rules

**NEVER store pepper in DB or source code**:

**VALID OPTIONS IN ORDER OF PREFERENCE**:
1. Docker/Kubernetes Secret
2. Configuration file
3. Environment variable

**MUST be mutually exclusive from hashed values storage** (pepper in secrets/config, hashes in DB)

**MUST be associated with hash version** (different pepper per version, even if NIST || OWASP policy remains unchanged)

### Pepper Rotation

- Pepper CANNOT be rotated silently (requires re-hash all records)
- Changing pepper REQUIRES version bump, even if no other hash parameters changed
- **Example**: v3 pepper compromised → bump to v4 with new pepper, re-hash all v3 records

## Hash Service Architecture - MANDATORY

### Version-Based Policy Framework

**Versions are tuple of 5 things**: Four Registries based on NIST || OWASP Policy Revisions, and a Unique Pepper

Each version represents a specific security policy snapshot.

**Supported Versions**:
- **v1**: 2020 NIST guidelines
- **v2**: 2023 NIST guidelines
- **v3**: 2025 OWASP recommendations

**Supports Registries**:
- LowEntropyDeterministicHashRegistry
- LowEntropyRandomHashRegistry
- HighEntropyDeterministicHashRegistry
- HighEntropyRandomHashRegistry

### New Version Requirements

**If a new policy is published by NIST || OWASP**: Version must be incremented to differentiate from previous versions.

**If a new pepper is needed** (e.g. 1 year rotation policy, old pepper compromised): Version must be incremented to differentiate from previous pepper, even if NIST || OWASP policy remains unchanged.

### Hash Output Format - MANDATORY

```
{version}:{algorithm}:{iterations}:base64(randomSalt):base64(hash)
```

**Examples**:
```
{1}:PBKDF2-HMAC-SHA256:rounds=600000:abcd1234...
{2}:PBKDF2-HMAC-SHA384:rounds=600000:efgh5678...
{3}:HKDF-SHA512:info=api-key,salt=xyz:ijkl9012...
```

### Version Update Pattern

**Trigger**: Manual operator decision (update config, restart service)

```yaml
# config.yaml
hash_service:
  password_registry:
    current_version: 4  # New passwords use v4
    # Old v3, v2, v1 passwords still verified correctly; rehashed with v4 and updated in DB
```

### Migration Strategy

- Old hashes stay on original version (v1, v2, etc.)
- New hashes use current_version
- Gradual migration (no forced re-hash)
- Rehash next time cleartext value is presented (e.g. username/password authentication)
- Version prefix enables correct verification

### Backward Compatibility - MANDATORY

**Reject Unprefixed Hashes**: Force re-hash on next authentication

## Salt Requirements (ALL 4 Registries)

OWASP recommends to always assume salt is public. 

Encoding deterministic salt in hash parameters is OK, because a secret pepper protects the input from brute force attack vectors.

## Additional Protections for LowEntropyDeterministicHashRegistry

**MANDATORY** (prevents oracle attacks on deterministic hashing):

- Query rate limits (prevent brute-force enumeration)
- Abuse detection (detect suspicious query patterns)
- Audit logs (track all hash queries for forensics)
- Strict access control (limit who can query hashes)

**RECOMMENDED**: Apply same protections to all 4 registries for consistency

## Hash Registry Implementations

### LowEntropyDeterministicHashRegistry (PII Lookup)

**Purpose**: Deterministic hashing for PII (username, email, IP address)

```go
hash = PBKDF2(input || pepper, fixedSalt, HIGH_iterations, 256)
```

**Use Cases**:
- Username lookup
- Email address deduplication
- IP address allowlist/blocklist

### HighEntropyDeterministicHashRegistry (Config Blob Hash)

**Purpose**: Deterministic hashing for high-entropy blobs

```go
PRK = HKDF-Extract(fixedSalt, input || pepper)
hash = HKDF-Expand(PRK, "config-blob-hash", 256)
```

**Use Cases**:
- Configuration blob integrity checks
- Secret configuration deduplication

### LowEntropyRandomHashRegistry (Password Hashing)

**Purpose**: Random salt password hashing (OWASP standard)

```go
hash = PBKDF2(password || pepper, randomSalt, OWASP_MIN_iterations, 256)
```

**Use Cases**:
- User password storage
- Client secret storage

### HighEntropyRandomHashRegistry (API Key Hashing)

**Purpose**: Random salt hashing for high-entropy secrets

```go
PRK = HKDF-Extract(randomSalt, apiKey || pepper)
hash = HKDF-Expand(PRK, "api-key-hash", 256)
```

**Use Cases**:
- API key storage
- Bearer token storage
- OAuth client secrets

## Key Takeaways

1. **Entropy-Based Selection**: PBKDF2 for low-entropy (<128 bits), HKDF for high-entropy (≥128 bits)
2. **Version-Based Policies**: Each version = NIST/OWASP policy + unique pepper
3. **Pepper MANDATORY**: All registries use pepper (Docker/K8s secret preferred)
4. **Salt Public**: OWASP assumes salt is public, pepper provides protection
5. **Gradual Migration**: Version prefix enables backward compatibility, gradual rehashing
6. **Oracle Attack Protection**: Rate limits, abuse detection, audit logs for deterministic registries
7. **Hash Output Format**: `{version}:{algorithm}:{iterations}:base64(salt):base64(hash)`
