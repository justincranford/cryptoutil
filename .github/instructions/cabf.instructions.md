---
description: "Instructions for CA/Browser Forum Baseline Requirements compliance"
applyTo: "**/*.go"
---
# CA/Browser Forum Baseline Requirements Instructions

- Adhere to the latest CA/Browser Forum Baseline Requirements for TLS Server Certificates when generating any certificates (CAs and end entities)
- Follow all certificate profile requirements specified in Section 7 of the Baseline Requirements
- Implement proper certificate serial number generation as specified in Section 7.1 (minimum 64 bits of CSPRNG output, non-sequential, greater than zero, less than 2^159)
- Use only approved cryptographic algorithms and key sizes as specified in Section 6.1.5 and 6.1.6
- Follow proper certificate validity period requirements (maximum 398 days for subscriber certificates as of 2020-09-01)
- Implement required certificate extensions and profiles as specified in Section 7.1.2
- Ensure proper subject and issuer name encoding as specified in Section 7.1.4
- Use approved signature algorithms as specified in Section 7.1.3.2
- Follow CRL and OCSP profile requirements as specified in Sections 7.2 and 7.3
- Implement proper audit logging for all certificate lifecycle events as specified in Section 5.4.1
- Ensure compliance with validation requirements in Section 3.2.2 for domain and organization validation
