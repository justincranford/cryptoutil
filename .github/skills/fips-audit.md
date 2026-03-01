# fips-audit

Detect FIPS 140-3 violations in Go code and provide fix guidance.

## Purpose

Use to audit cryptographic usage for FIPS 140-3 compliance. Goes beyond
the `cicd lint-go` non-fips-algorithms checker by analyzing usage patterns,
key sizes, and algorithm configurations.

## FIPS 140-3 Approved Algorithms

| Category | Approved | Banned |
|----------|---------|--------|
| Asymmetric | RSA ≥2048, ECDSA P-256/384/521, EdDSA Ed25519/448 | RSA <2048 |
| Symmetric | AES ≥128 (GCM, CBC+HMAC) | DES, 3DES, RC4 |
| Hash | SHA-256/384/512, HMAC-SHA256/384/512 | MD5, SHA-1 |
| KDF | PBKDF2-HMAC-SHA256/384/512, HKDF-SHA256/384/512 | bcrypt, scrypt, Argon2 |
| Random | crypto/rand | math/rand |

## Common Violations

```go
// ❌ VIOLATION: weak hash
import "crypto/md5"
hash := md5.Sum(data)

// ✅ FIX: use SHA-256
import "crypto/sha256"
hash := sha256.Sum256(data)

// ❌ VIOLATION: math/rand instead of crypto/rand
import "math/rand"
n := rand.Int()

// ✅ FIX: crypto/rand
import crand "crypto/rand"
var buf [8]byte
crand.Read(buf[:])

// ❌ VIOLATION: bcrypt (not FIPS compliant)
import "golang.org/x/crypto/bcrypt"

// ✅ FIX: PBKDF2 with SHA-256
import "golang.org/x/crypto/pbkdf2"
key := pbkdf2.Key(password, salt, 600000, 32, sha256.New)

// ❌ VIOLATION: RSA key size too small
rsa.GenerateKey(rand, 1024)

// ✅ FIX: RSA ≥2048
rsa.GenerateKey(rand, 2048) // minimum; prefer 3072 or 4096
```

## Audit Checklist

```bash
# Find math/rand usage (should be crypto/rand)
grep -rn ""math/rand"" --include="*.go" .

# Find MD5/SHA1 usage
grep -rn "crypto/md5\|crypto/sha1" --include="*.go" .

# Find bcrypt/scrypt/argon2
grep -rn "golang.org/x/crypto/bcrypt\|golang.org/x/crypto/scrypt\|golang.org/x/crypto/argon2" --include="*.go" .

# Find DES/RC4/3DES
grep -rn ""crypto/des"\|"crypto/rc4"" --include="*.go" .

# Find weak RSA key sizes
grep -rn "GenerateKey.*1024\|GenerateKey.*512" --include="*.go" .
```

## References

See [ARCHITECTURE.md Section 6.1 FIPS 140-3 Compliance Strategy](../../docs/ARCHITECTURE.md#61-fips-140-3-compliance-strategy) for full requirements.
See [ARCHITECTURE.md Section 6.4 Cryptographic Architecture](../../docs/ARCHITECTURE.md#64-cryptographic-architecture) for approved implementations.
