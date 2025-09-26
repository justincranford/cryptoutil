# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| main    | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities responsibly by following these steps:

### Preferred Method: GitHub Security Advisory

1. Go to the [Security tab](https://github.com/justincranford/cryptoutil/security) of this repository
2. Click "Report a vulnerability"
3. Fill out the security advisory form with:
   - Detailed description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact assessment
   - Suggested fix (if known)

### Alternative Method: Email

If you prefer email communication, send details to:
- **Email**: [security contact email]
- **Subject**: `[SECURITY] cryptoutil vulnerability report`

### What to Include

Please include as much of the following information as possible:

- **Vulnerability Type**: (e.g., cryptographic weakness, authentication bypass, etc.)
- **Component**: Affected module/package/function
- **Severity**: Your assessment of the impact
- **Reproduction Steps**: Clear steps to reproduce the issue
- **Proof of Concept**: Code or commands that demonstrate the vulnerability
- **Impact**: Description of what an attacker could accomplish
- **Mitigation**: Any workarounds or temporary fixes
- **Environment**: Go version, OS, deployment context where applicable

## Security Response Process

### Initial Response
- **Acknowledgment**: Within 24 hours of report receipt
- **Initial Assessment**: Within 72 hours
- **Regular Updates**: Every 7 days until resolution

### Investigation Process
1. **Triage**: Confirm and classify the vulnerability
2. **Impact Assessment**: Determine severity using CVSS 3.1 framework
3. **Fix Development**: Create and test security patches
4. **Coordinated Disclosure**: Plan release timeline with reporter

### Disclosure Timeline
- **Critical/High Severity**: 7-14 days
- **Medium Severity**: 30 days
- **Low Severity**: 90 days

We follow responsible disclosure practices and will:
- Work with you to understand and validate the issue
- Acknowledge your contribution in our security advisory (unless you prefer anonymity)
- Provide credit in release notes and changelog

## Security Features

### Cryptographic Security
- **FIPS 140-3 Compliance**: Only approved algorithms and key sizes
- **Key Management**: Multi-layered barrier system with hierarchical keys
- **Secure Defaults**: Conservative security configurations by default

### Application Security
- **Input Validation**: All API inputs validated and sanitized
- **Authentication**: Multi-factor barrier system (unseal + root + intermediate keys)
- **Authorization**: Role-based access with IP allowlisting
- **Rate Limiting**: DoS protection with per-IP rate limits
- **CSRF Protection**: Cross-site request forgery protection enabled
- **Security Headers**: Comprehensive HTTP security headers

### Infrastructure Security
- **Container Security**: Non-root user execution, minimal attack surface
- **TLS Configuration**: TLS 1.2+ minimum with proper certificate validation
- **Secret Management**: Secure handling and storage of sensitive data
- **Audit Logging**: Comprehensive security event logging

## Security Best Practices

### For Users
- Keep cryptoutil updated to the latest version
- Use strong, unique unseal keys and root tokens
- Enable IP allowlisting in production environments
- Monitor security advisories and apply patches promptly
- Follow principle of least privilege for API access
- Implement network segmentation and access controls

### For Developers
- Follow secure coding practices outlined in our development guidelines
- Run security scans before committing code
- Use pre-commit hooks to catch security issues early
- Validate all inputs and handle errors securely
- Never commit secrets or cryptographic material

## Vulnerability Assessment

We use the following severity classification based on CVSS 3.1:

### Critical (9.0-10.0)
- Remote code execution without authentication
- Complete compromise of cryptographic material
- Bypass of all security controls

### High (7.0-8.9)
- Privilege escalation to administrative access
- Exposure of sensitive cryptographic keys
- Authentication bypass

### Medium (4.0-6.9)
- Information disclosure of non-critical data
- Limited privilege escalation
- DoS attacks requiring authentication

### Low (0.1-3.9)
- Minor information disclosure
- Issues requiring local access
- Configuration weaknesses

## Security Tools and Scanning

Our CI/CD pipeline includes automated security scanning:

- **SAST**: Static Application Security Testing with CodeQL and Gosec
- **DAST**: Dependency scanning with Trivy and Nancy
- **Container Scanning**: Image vulnerability assessment
- **SBOM Generation**: Software Bill of Materials for supply chain security
- **License Compliance**: Automated license violation detection

## Security Contact

For non-critical security questions or general security guidance:
- **GitHub Discussions**: Use the Security category
- **Issues**: Tag with `security` label for public discussions

## Hall of Fame

We recognize security researchers who help improve cryptoutil security:

<!-- Security researchers who have responsibly disclosed vulnerabilities will be listed here -->

*No security vulnerabilities have been reported yet.*

## Legal

This security policy is subject to our standard terms of service and applicable law. We commit to:

- Not pursuing legal action against security researchers acting in good faith
- Providing safe harbor for vulnerability research conducted responsibly
- Acknowledging and crediting researchers (unless anonymity is requested)

---

**Last Updated**: September 26, 2025  
**Version**: 1.0
