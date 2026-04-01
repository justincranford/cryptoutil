Audit Go code for FIPS 140-3 compliance violations and provide fixes.

**Full Copilot original**: [.github/skills/fips-audit/SKILL.md](.github/skills/fips-audit/SKILL.md)

## Approved Algorithms

| Algorithm | Approved | Notes |
|-----------|----------|-------|
| RSA | ≥2048 bit only | Key generation + signing |
| ECDSA | P-256, P-384, P-521 | NIST curves only |
| ECDH | P-256, P-384, P-521 | Key exchange |
| EdDSA | Ed25519 only | Signing |
| AES | ≥128 bit | GCM mode preferred |
| SHA | SHA-256, SHA-384, SHA-512 | SHA-1 BANNED |
| HMAC | With approved hash | HMAC-SHA256+ |
| PBKDF2 | ≥600k iterations | For low-entropy inputs (passwords) |
| HKDF | With SHA-256+ | For high-entropy key derivation |

## BANNED (Zero Tolerance)

```
RSA < 2048 bits
DES, 3DES, RC4, RC2, Blowfish
MD5, SHA-1
bcrypt, scrypt, Argon2 (not FIPS-approved KDFs)
math/rand (not cryptographic)
crypto/elliptic with non-NIST curves
```

## Common Violations and Fixes

```go
// VIOLATION: math/rand
import "math/rand"
n := rand.Intn(100)

// FIX: crypto/rand
import "crypto/rand"
import "math/big"
n, _ := rand.Int(rand.Reader, big.NewInt(100))
```

```go
// VIOLATION: MD5 hash
import "crypto/md5"
h := md5.Sum(data)

// FIX: SHA-256
import "crypto/sha256"
h := sha256.Sum256(data)
```

```go
// VIOLATION: bcrypt
import "golang.org/x/crypto/bcrypt"
hash, _ := bcrypt.GenerateFromPassword(password, 14)

// FIX: PBKDF2 with SHA-256 (≥600k iterations)
import "golang.org/x/crypto/pbkdf2"
hash := pbkdf2.Key(password, salt, 600000, 32, sha256.New)
```

## Audit Steps

1. Search for banned imports: `grep -r "crypto/md5\|crypto/des\|crypto/rc4\|math/rand\|bcrypt\|scrypt\|argon2" ./internal`
2. Check RSA key sizes: all `rsa.GenerateKey` calls must use ≥2048
3. Check elliptic curves: all `elliptic.*` must use P-256/P-384/P-521
4. Check KDF iterations: all PBKDF2 calls must use ≥600000 iterations
5. Verify pepper is applied before all hash operations
