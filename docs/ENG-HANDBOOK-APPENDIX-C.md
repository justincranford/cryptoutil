
---

## Appendix C: Compliance Matrix

### C.1 FIPS 140-3 Compliance

**Status**: ALWAYS enabled, NEVER disabled

**Approved Algorithms**: RSA ≥2048, ECDSA (P-256/384/521), ECDH, EdDSA (25519/448), AES ≥128 (GCM, CBC+HMAC), SHA-256/384/512, HMAC-SHA256/384/512, PBKDF2, HKDF

**BANNED**: bcrypt, scrypt, Argon2, MD5, SHA-1, RSA <2048, DES, 3DES

### C.2 PKI Standards Compliance

**CA/Browser Forum Baseline Requirements**:

- Serial ≥64 bits CSPRNG, >0, <2^159
- Validity ≤398 days (subscriber), 5-10 years (intermediate), 20-25 years (root)
- Extensions: Key Usage (critical), EKU, SAN, AKI, SKI, CRL, OCSP
- Audit logging: 7-year retention

**Validation**: DV (<30 days), OV/EV (<13 months), CAA DNS checks

### C.3 OAuth 2.1 / OIDC 1.0 Compliance

**OAuth 2.1**: Authorization Code + PKCE, Client Credentials, Token Exchange
**OIDC 1.0**: ID Token (JWS), UserInfo endpoint, Discovery (.well-known/openid-configuration)

**Security**: HTTPS required, state parameter, nonce in ID tokens, consent tracking

### C.4 Security Standards Compliance

**OWASP**: Password Storage Cheat Sheet (peppering, PBKDF2), Top 10
**NIST**: FIPS 140-3 (crypto), SP 800-63 (digital identity)
**Zero Trust**: No caching authz, mTLS, least privilege, audit logging (90-day retention)
