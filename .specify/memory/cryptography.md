# Cryptography Specifications

**Version**: 1.0.0
**Last Updated**: 2025-12-24
**Referenced By**: `.github/instructions/02-07.cryptography.instructions.md`

## FIPS 140-3 Compliance - MANDATORY

**CRITICAL: FIPS 140-3 mode is ALWAYS enabled by default and MUST NEVER be disabled**

All cryptographic operations MUST use FIPS-approved algorithms ONLY.

### Approved Algorithms

**Asymmetric Cryptography**:
- RSA ≥ 2048 bits (2048, 3072, 4096)
- ECDSA (NIST curves: P-256, P-384, P-521)
- ECDH (NIST curves: P-256, P-384, P-521)
- EdDSA (Ed25519, Ed448)

**Symmetric Cryptography**:
- AES ≥ 128 bits (128, 192, 256)
- AES-GCM (authenticated encryption)
- AES-CBC (with HMAC for authentication)

**Digest Functions**:
- SHA-256, SHA-384, SHA-512
- HMAC-SHA256, HMAC-SHA384, HMAC-SHA512

**Key Derivation**:
- PBKDF2-HMAC-SHA256, PBKDF2-HMAC-SHA384, PBKDF2-HMAC-SHA512
- HKDF-SHA256, HKDF-SHA384, HKDF-SHA512

### BANNED Algorithms - NEVER USE

❌ **bcrypt** - NOT FIPS-approved, use PBKDF2-HMAC-SHA256 instead
❌ **scrypt** - NOT FIPS-approved, use PBKDF2-HMAC-SHA256 instead
❌ **Argon2** - NOT FIPS-approved, use PBKDF2-HMAC-SHA256 instead
❌ **MD5** - NOT FIPS-approved, use SHA-256 or SHA-512 instead
❌ **SHA-1** - NOT FIPS-approved, use SHA-256 or SHA-512 instead
❌ **RSA < 2048 bits** - Insufficient key length
❌ **DES, 3DES** - Deprecated ciphers

## Algorithm Agility - MANDATORY

**All cryptographic operations MUST support configurable algorithms with FIPS-approved defaults**

### Configurable Algorithm Support

**Key Generation**:
- RSA: 2048, 3072, 4096 bit keys
- ECDSA: P-256, P-384, P-521 curves
- EdDSA: Ed25519, Ed448

**Encryption**:
- AES-GCM: 128, 192, 256 bit keys
- AES-HS (AES-HMAC): 256, 384, 512 bit keys

**Digests**:
- SHA-256, SHA-384, SHA-512
- HMAC-SHA256, HMAC-SHA384, HMAC-SHA512

**Key Derivation**:
- PBKDF2-HMAC-SHA256, PBKDF2-HMAC-SHA384, PBKDF2-HMAC-SHA512
- HKDF-SHA256, HKDF-SHA384, HKDF-SHA512

### Configuration-Driven Selection

**Pattern**: Use config structs with Algorithm and KeySize fields, switch on algorithm type:

```go
type CryptoConfig struct {
    Algorithm AlgorithmType  // RSA, ECDSA, EdDSA, AES, etc.
    KeySize   int            // 2048, 3072, 4096, 128, 192, 256, etc.
}

func GenerateKey(config CryptoConfig) (Key, error) {
    switch config.Algorithm {
    case AlgRSA:
        return rsa.GenerateKey(rand.Reader, config.KeySize)
    case AlgECDSA:
        return ecdsa.GenerateKey(getCurve(config.KeySize), rand.Reader)
    case AlgEdDSA:
        return ed25519.GenerateKey(rand.Reader)
    default:
        return nil, fmt.Errorf("unsupported algorithm: %s", config.Algorithm)
    }
}
```

## Unseal Key Management - MANDATORY

### Unseal Key Derivation - CRITICAL

**ALWAYS derive unseal keys deterministically for interoperability**

All cryptoutil instances using the same set of shared unseal secrets MUST derive the same unseal JWKs, including KIDs and key materials, for cryptographic interoperability between instances.

### Deterministic Derivation Pattern

**Use HKDF with master secret, salt, and purpose-specific info**:

```go
func DeriveUnsealKey(masterSecret, salt, info []byte) ([]byte, error) {
    // Extract: PRK = HKDF-Extract(salt, masterSecret)
    prk := hkdf.Extract(sha256.New, masterSecret, salt)

    // Expand: OKM = HKDF-Expand(PRK, info, length)
    reader := hkdf.Expand(sha256.New, prk, info)
    key := make([]byte, 32)  // 256-bit key
    if _, err := io.ReadFull(reader, key); err != nil {
        return nil, fmt.Errorf("failed to expand key: %w", err)
    }
    return key, nil
}
```

**Why Deterministic**: Same master secret + salt + info = same unseal key across all instances

## Elastic Key Rotation

**Key versioning pattern**: Elastic Keys are key rings with active Material Key for encrypting||signing, and historical Material Keys for decrypting||verifying

### Rotation Workflow

**Encryption**:
- Always use active key
- Embed key ID with ciphertext||signature

**Decryption**:
- Use key matching embedded key ID
- Deterministically identify historical key for decrypting||verifying

**Rotation**:
- Generate new Material Key
- Identify it as the active key ID
- Keep all old keys for decrypting||verifying

### Implementation Pattern

```go
type ElasticKeyRing struct {
    ActiveKeyID   string
    ActiveKey     Key
    HistoricalKeys map[string]Key  // keyID → key
}

func (kr *ElasticKeyRing) Encrypt(plaintext []byte) (Ciphertext, error) {
    // Always use active key, embed key ID
    ct, err := kr.ActiveKey.Encrypt(plaintext)
    return Ciphertext{KeyID: kr.ActiveKeyID, Data: ct}, err
}

func (kr *ElasticKeyRing) Decrypt(ciphertext Ciphertext) ([]byte, error) {
    // Use key matching embedded key ID
    key := kr.HistoricalKeys[ciphertext.KeyID]
    if key == nil {
        return nil, fmt.Errorf("key %s not found", ciphertext.KeyID)
    }
    return key.Decrypt(ciphertext.Data)
}

func (kr *ElasticKeyRing) Rotate() error {
    // Generate new Material Key
    newKey, err := GenerateKey(config)
    if err != nil {
        return err
    }

    // Move current active key to historical keys
    kr.HistoricalKeys[kr.ActiveKeyID] = kr.ActiveKey

    // Set new key as active
    kr.ActiveKeyID = generateKeyID()
    kr.ActiveKey = newKey

    return nil
}
```

## Secure Random Generation - MANDATORY

**ALWAYS use crypto/rand, NEVER use math/rand**

### Required Pattern

```go
import crand "crypto/rand"

// Generate random bytes
func generateRandomBytes(length int) ([]byte, error) {
    bytes := make([]byte, length)
    if _, err := crand.Read(bytes); err != nil {
        return nil, fmt.Errorf("failed to generate random bytes: %w", err)
    }
    return bytes, nil
}

// Use for: tokens, nonces, salts, IVs, session IDs
```

**Why**: `crypto/rand` uses OS-provided CSPRNG (cryptographically secure pseudo-random number generator). `math/rand` is predictable and NOT suitable for cryptographic operations.

## Cryptographic Libraries - MANDATORY

### Preferred Standard Library Packages

**Key Generation**:
- `crypto/rand` - Random number generation
- `crypto/rsa` - RSA key generation
- `crypto/ecdsa` - ECDSA key generation
- `crypto/ed25519` - EdDSA key generation

**Symmetric Encryption**:
- `crypto/aes` - AES cipher
- `crypto/cipher` - Block cipher modes (GCM, CBC)

**Hashing**:
- `crypto/sha256` - SHA-256 digest
- `crypto/sha512` - SHA-384, SHA-512 digest
- `crypto/hmac` - HMAC construction

**TLS**:
- `crypto/tls` - TLS connections

**Key Derivation**:
- `golang.org/x/crypto/pbkdf2` - PBKDF2 implementation
- `golang.org/x/crypto/hkdf` - HKDF implementation

### Third-Party Library Policy

**Avoid third-party crypto libraries unless necessary** (requires security review)

**Acceptable Use Cases**:
- Implementing protocols not in standard library (e.g., JWT, JOSE)
- Performance-critical operations with audited implementations
- Compliance with specific standards (e.g., FIPS modules)

**Review Requirements**:
- Security audit history
- Active maintenance and CVE response
- Community adoption and testing
- FIPS 140-3 certification (if applicable)

## Key Takeaways

1. **FIPS 140-3 ALWAYS enabled** - No exceptions, ONLY approved algorithms
2. **Password hashing: PBKDF2-HMAC-SHA256** - See `hashes.md` for hash registry patterns
3. **Algorithm agility: Configurable with FIPS defaults** - Support multiple key sizes
4. **Deterministic key derivation: HKDF** - For instance interoperability
5. **crypto/rand ALWAYS** - Never math/rand for cryptographic operations
6. **Standard library preferred** - Avoid third-party crypto unless necessary
7. **Elastic keys**: Active key for encrypt, historical keys for decrypt, rotation preserves old keys
8. **Certificate validation** - See `pki.md` for TLS certificate requirements
