---
description: "Instructions for cryptographic operations"
applyTo: "**"
---
# Crypto Instructions

- Only use NIST FIPS 140-3 approved algorithms and key sizes (e.g., RSA ≥ 2048 bits, AES ≥ 128 bits, EC NIST curves, EdDSA)
- Use keygen package for generating RSA, ECDSA, ECDH, EdDSA, AES, HMAC, and UUIDv7 keys.
- Avoid fallback values for cryptographic operations.
- Use provided pool and keygen abstractions for for concurrent cryptographic key generation operations.
- Review cryptographic practices for compliance and security.
- Periodically audit cryptographic code for compliance.
