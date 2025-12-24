# Cryptography Specifications

**Referenced By**: `.github/instructions/02-07.cryptography.instructions.md`

## FIPS 140-3 Compliance - MANDATORY

**CRITICAL: FIPS 140-3 mode is ALWAYS enabled by default and MUST NEVER be disabled**

All cryptographic operations MUST use FIPS-approved algorithms ONLY.

### Approved Algorithms

| Category | Algorithms |
|----------|------------|
| **Asymmetric** | RSA ≥2048, ECDSA/ECDH (P-256/384/521), EdDSA (Ed25519/448) |
| **Symmetric** | AES ≥128 (GCM, CBC+HMAC) |
| **Digest** | SHA-256/384/512, HMAC-SHA256/384/512 |
| **KDF** | PBKDF2-HMAC-SHA256/384/512, HKDF-SHA256/384/512 |

### BANNED Algorithms

❌ **Password Hashing**: bcrypt, scrypt, Argon2 (use PBKDF2-HMAC-SHA256)
❌ **Digests**: MD5, SHA-1 (use SHA-256/512)
❌ **Other**: RSA <2048, DES, 3DES

## Algorithm Agility - MANDATORY

**Pattern**: Config-driven selection with Algorithm + KeySize fields, switch on type. Support RSA 2048/3072/4096, ECDSA P-256/384/521, EdDSA Ed25519/448, AES 128/192/256, SHA-256/384/512.

## Key Management - MANDATORY

### Unseal Key Derivation

**ALWAYS use HKDF** for deterministic derivation: Same master secret + salt + info = same unseal key across instances

### Elastic Key Rotation

**Pattern**: Key ring with active key (encrypt/sign) + historical keys (decrypt/verify). Embed key ID with ciphertext. Rotation: generate new key, move current to historical, keep all old keys.

## Secure Random Generation - MANDATORY

**ALWAYS use crypto/rand** (CSPRNG), NEVER math/rand (predictable)

## Cryptographic Libraries - MANDATORY

**Preferred**: `crypto/*` (rand, rsa, ecdsa, ed25519, aes, cipher, sha256, sha512, hmac, tls), `golang.org/x/crypto/*` (pbkdf2, hkdf)
**Third-Party**: Avoid unless necessary (JWT, JOSE, FIPS modules) - requires security review

## Key Takeaways

1. **FIPS 140-3 ALWAYS enabled** - No exceptions, ONLY approved algorithms
2. **Password hashing: PBKDF2-HMAC-SHA256** - See `hashes.md` for hash registry patterns
3. **Algorithm agility: Configurable with FIPS defaults** - Support multiple key sizes
4. **Deterministic key derivation: HKDF** - For instance interoperability
5. **crypto/rand ALWAYS** - Never math/rand for cryptographic operations
6. **Standard library preferred** - Avoid third-party crypto unless necessary
7. **Elastic keys**: Active key for encrypt, historical keys for decrypt, rotation preserves old keys
8. **Certificate validation** - See `pki.md` for TLS certificate requirements
