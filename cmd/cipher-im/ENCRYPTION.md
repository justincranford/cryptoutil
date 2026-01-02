# cipher-im Encryption Architecture

Technical details of cryptographic operations in cipher-im service.

## Overview

cipher-im uses **hybrid encryption** combining RSA-OAEP (asymmetric) for
content encryption and Ed25519 (elliptic curve) for digital signatures.

**EDUCATIONAL USE ONLY**: This implementation stores private keys server-side
for demonstration purposes. Production systems MUST store private keys
client-side only (user devices, hardware security modules, secure enclaves).

## Key Generation

### User Registration Flow

When user registers, two keypairs are generated:

1. **RSA-4096 Keypair** (Encryption):
   - Algorithm: RSA-OAEP with SHA-256
   - Key size: 4096 bits
   - Purpose: Encrypt/decrypt message content
   - Public key: Shared with all users
   - Private key: ⚠️ Stored server-side (demo only)

2. **Ed25519 Keypair** (Signing):
   - Algorithm: Edwards-curve Digital Signature Algorithm
   - Curve: Curve25519
   - Purpose: Sign/verify message integrity
   - Public key: Shared with all users
   - Private key: ⚠️ Stored server-side (demo only)

### Key Storage Format

Keys stored as PEM-encoded text in database:

```sql
-- RSA public key (encryption)
-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA...
-----END PUBLIC KEY-----

-- Ed25519 public key (signing)
-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEA...
-----END PUBLIC KEY-----
```

Private keys follow same PEM format with different headers.

## Message Encryption Flow

### Step 1: Plaintext Preparation

```
Original Message: "Hello, World! This is a secret."
Encoding: UTF-8 bytes
```

### Step 2: Per-Receiver RSA-OAEP Encryption

For EACH receiver independently:

```go
// Pseudocode
receiverPublicKey := GetRSAPublicKey(receiverID)
ciphertext := RSA_OAEP_Encrypt(
    publicKey: receiverPublicKey,
    plaintext: messageContent,
    hash: SHA256,
    label: ""  // Optional authenticated data
)
```

**Output**: Base64-encoded ciphertext unique per receiver

**Properties**:

- Each receiver gets different ciphertext (non-deterministic)
- Ciphertext size: 512 bytes (4096-bit RSA key)
- Max plaintext: ~446 bytes (4096 bits - padding overhead)
- For longer messages: split into chunks or use hybrid scheme

### Step 3: Generate Random Nonce

```go
nonce := GenerateRandomBytes(32)  // 256-bit nonce
nonceBase64 := Base64Encode(nonce)
```

**Purpose**: Prevent replay attacks, ensure unique signatures

### Step 4: Create Digital Signature

```go
// Pseudocode
senderPrivateKey := GetEd25519PrivateKey(senderID)
dataToSign := ciphertext + nonce
signature := Ed25519_Sign(
    privateKey: senderPrivateKey,
    message: dataToSign
)
```

**Output**: Base64-encoded 64-byte signature

**Properties**:

- Signature proves sender identity
- Signature proves ciphertext integrity
- Signature includes nonce (freshness proof)

### Step 5: Store Per-Receiver Records

```sql
-- One record per receiver
INSERT INTO message_receivers (
    message_id,
    receiver_id,
    encrypted_content,  -- Base64(RSA-OAEP ciphertext)
    nonce,              -- Base64(random 32 bytes)
    signature           -- Base64(Ed25519 signature)
) VALUES (...);
```

**Multi-Receiver Example**:

```
Message to [Alice, Bob, Charlie]:
- 1 message record (metadata)
- 3 message_receivers records (one per recipient)
- Each has different encrypted_content (same plaintext)
- All share same signature (proves sender + integrity)
```

## Message Decryption Flow

### Step 1: Retrieve Receiver-Specific Record

```sql
SELECT encrypted_content, nonce, signature
FROM message_receivers
WHERE message_id = ? AND receiver_id = ?
```

### Step 2: RSA-OAEP Decryption

```go
// Pseudocode
receiverPrivateKey := GetRSAPrivateKey(receiverID)
plaintext := RSA_OAEP_Decrypt(
    privateKey: receiverPrivateKey,
    ciphertext: Base64Decode(encrypted_content),
    hash: SHA256,
    label: ""
)
```

**Output**: UTF-8 encoded plaintext message

### Step 3: Signature Verification

```go
// Pseudocode
senderPublicKey := GetEd25519PublicKey(senderID)
dataToVerify := Base64Decode(encrypted_content) + Base64Decode(nonce)
isValid := Ed25519_Verify(
    publicKey: senderPublicKey,
    message: dataToVerify,
    signature: Base64Decode(signature)
)
```

**Result**: `true` if signature valid, `false` otherwise

**Security Properties**:

- ✅ Proves sender identity (only sender has private key)
- ✅ Proves message integrity (any tampering invalidates signature)
- ✅ Proves message freshness (nonce prevents replay)

## Cryptographic Algorithms

### RSA-OAEP (Encryption)

- **Standard**: PKCS#1 v2.2 (RFC 8017)
- **Key Size**: 4096 bits
- **Hash Function**: SHA-256
- **Padding**: Optimal Asymmetric Encryption Padding (OAEP)
- **Security Level**: ~152 bits (classical), ~128 bits (quantum-resistant)

**Implementation**:

```go
import "crypto/rsa"
import "crypto/rand"
import "crypto/sha256"

func EncryptRSA(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
    return rsa.EncryptOAEP(
        sha256.New(),
        rand.Reader,
        publicKey,
        plaintext,
        nil,  // label
    )
}
```

### Ed25519 (Signing)

- **Standard**: RFC 8032
- **Curve**: Curve25519 (Edwards form)
- **Signature Size**: 64 bytes
- **Public Key Size**: 32 bytes
- **Security Level**: ~128 bits (classical), ~64 bits (quantum-resistant)

**Implementation**:

```go
import "crypto/ed25519"

func SignEd25519(privateKey ed25519.PrivateKey, message []byte) []byte {
    return ed25519.Sign(privateKey, message)
}
```

## Security Analysis

### Threat Model

**Attacker Capabilities**:

- ✅ Can observe all database records (encrypted content, signatures, nonces)
- ✅ Can observe all network traffic (HTTPS protects in transit)
- ✅ Can attempt cryptanalysis (brute force, chosen plaintext, etc.)

**Attacker Constraints**:

- ❌ Cannot access private keys (⚠️ except in this demo - server-side storage)
- ❌ Cannot break RSA-OAEP or Ed25519 (computationally infeasible)
- ❌ Cannot modify database without detection (signature verification fails)

### Security Properties

**Confidentiality** (RSA-OAEP):

- ✅ Only receiver can decrypt (requires private key)
- ✅ Non-deterministic (same plaintext → different ciphertext)
- ✅ Semantic security (ciphertext reveals nothing about plaintext)

**Integrity** (Ed25519):

- ✅ Signature proves message not modified
- ✅ Signature binds sender identity to content
- ✅ Nonce prevents replay attacks

**Authentication** (Ed25519):

- ✅ Only sender could create valid signature
- ✅ Public key cryptography (no shared secrets)

### Known Limitations

**⚠️ CRITICAL LIMITATIONS - EDUCATIONAL DEMO ONLY**:

1. **Server-Side Private Keys**:
   - Private keys stored in database
   - Server compromise = all messages compromised
   - Production MUST use client-side key storage

2. **No Forward Secrecy**:
   - Past messages compromised if private key leaked
   - Use ephemeral keys (ECDHE, Signal Protocol) for production

3. **No Message Deletion Guarantee**:
   - Soft delete (receiver's copy only)
   - Other receivers retain copies
   - Production MUST support sender-initiated expiry

4. **No Key Rotation**:
   - Keys generated once at registration
   - No mechanism to rotate compromised keys
   - Production MUST support key updates

5. **No Quantum Resistance**:
   - RSA-4096 vulnerable to Shor's algorithm
   - Ed25519 partially vulnerable (64-bit security)
   - Future-proof: Use post-quantum algorithms (CRYSTALS-Kyber, CRYSTALS-Dilithium)

## Comparison with Production Systems

### Signal Protocol (Recommended)

**Features**:

- ✅ End-to-End Encryption (E2EE)
- ✅ Forward Secrecy (ephemeral keys)
- ✅ Deniability (sender repudiation)
- ✅ Client-side private keys only
- ✅ Double Ratchet Algorithm

**vs cipher-im**:

- cipher-im: Server-side keys, no forward secrecy
- Signal: Client-side only, ephemeral keys per message

### PGP/GPG (Email Encryption)

**Features**:

- ✅ Client-side private keys
- ✅ Web of trust / X.509 PKI
- ✅ RSA or ECC support
- ❌ No forward secrecy (long-lived keys)

**vs cipher-im**:

- Similar RSA-OAEP encryption
- cipher-im: Server-side keys (demo only)
- PGP: Client-side keys (production)

### TLS 1.3 (Transport Security)

**Features**:

- ✅ Perfect Forward Secrecy (ephemeral ECDHE)
- ✅ Certificate-based authentication
- ✅ Transport-layer security
- ❌ No end-to-end encryption (server sees plaintext)

**vs cipher-im**:

- TLS: In-transit security only
- cipher-im: At-rest encryption (database records encrypted)
- Combined: TLS + cipher-im = defense in depth

## Code Examples

### Complete Encryption/Decryption Example

```go
package main

import (
    "crypto/ed25519"
    "crypto/rand"
    "crypto/rsa"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
)

func main() {
    // 1. Generate keys
    senderPrivateEd, senderPublicEd, _ := ed25519.GenerateKey(rand.Reader)
    receiverPrivateRSA, _ := rsa.GenerateKey(rand.Reader, 4096)
    receiverPublicRSA := &receiverPrivateRSA.PublicKey

    // 2. Encrypt message
    plaintext := []byte("Hello, World! This is a secret.")
    ciphertext, _ := rsa.EncryptOAEP(
        sha256.New(),
        rand.Reader,
        receiverPublicRSA,
        plaintext,
        nil,
    )

    // 3. Generate nonce
    nonce := make([]byte, 32)
    rand.Read(nonce)

    // 4. Sign ciphertext + nonce
    dataToSign := append(ciphertext, nonce...)
    signature := ed25519.Sign(senderPrivateEd, dataToSign)

    // 5. Encode for storage
    encryptedContentB64 := base64.StdEncoding.EncodeToString(ciphertext)
    nonceB64 := base64.StdEncoding.EncodeToString(nonce)
    signatureB64 := base64.StdEncoding.EncodeToString(signature)

    fmt.Println("Encrypted:", encryptedContentB64)
    fmt.Println("Nonce:", nonceB64)
    fmt.Println("Signature:", signatureB64)

    // 6. Decrypt message
    ciphertextDecoded, _ := base64.StdEncoding.DecodeString(encryptedContentB64)
    decrypted, _ := rsa.DecryptOAEP(
        sha256.New(),
        rand.Reader,
        receiverPrivateRSA,
        ciphertextDecoded,
        nil,
    )

    // 7. Verify signature
    nonceDecoded, _ := base64.StdEncoding.DecodeString(nonceB64)
    signatureDecoded, _ := base64.StdEncoding.DecodeString(signatureB64)
    dataToVerify := append(ciphertextDecoded, nonceDecoded...)
    isValid := ed25519.Verify(senderPublicEd, dataToVerify, signatureDecoded)

    fmt.Println("Decrypted:", string(decrypted))
    fmt.Println("Signature valid:", isValid)
}
```

## Best Practices for Production

1. **Client-Side Key Storage ONLY**:
   - Store private keys on user devices
   - Use secure enclaves (iOS Keychain, Android KeyStore)
   - Never transmit private keys to server

2. **Forward Secrecy**:
   - Use ephemeral keys (ECDHE, Signal Protocol)
   - Rotate keys per message or session
   - Delete old keys after use

3. **Key Rotation**:
   - Support key updates without message loss
   - Implement key expiry and renewal
   - Provide key revocation mechanism

4. **Quantum Resistance**:
   - Migrate to post-quantum algorithms (NIST standardized)
   - Hybrid schemes: Classical + PQC (defense in depth)

5. **Metadata Protection**:
   - Minimize metadata leakage (sender, receiver, timestamps)
   - Use onion routing (Tor) for network anonymity
   - Implement sealed sender (Signal)

## References

- **RSA-OAEP**: RFC 8017 (PKCS#1 v2.2)
- **Ed25519**: RFC 8032
- **Signal Protocol**: <https://signal.org/docs/>
- **NIST PQC**: <https://csrc.nist.gov/projects/post-quantum-cryptography>
- **FIPS 140-3**: Federal cryptography standards

## See Also

- [README.md](README.md): Quick start and deployment
- [API.md](API.md): Complete API reference
- [TUTORIAL.md](TUTORIAL.md): Step-by-step user guide
