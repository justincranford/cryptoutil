---
description: "Instructions for specialized domains: CA/Browser Forum, project layout, PRs, documentation"
applyTo: "**"
---
# Specialized Domain Instructions

## CA/Browser Forum Baseline Requirements

- Adhere to latest CA/Browser Forum Baseline Requirements for TLS Server Certificates
- Follow certificate profile requirements in Section 7
- Implement proper certificate serial number generation (Section 7.1): minimum 64 bits CSPRNG, non-sequential, >0, <2^159
- Use only approved cryptographic algorithms and key sizes (Section 6.1.5, 6.1.6)
- Follow validity period requirements (max 398 days for subscriber certificates post-2020-09-01)
- Implement required certificate extensions (Section 7.1.2)
- Ensure proper subject/issuer name encoding (Section 7.1.4)
- Use approved signature algorithms (Section 7.1.3.2)
- Follow CRL and OCSP profile requirements (Sections 7.2, 7.3)
- Implement proper audit logging for certificate lifecycle events (Section 5.4.1)
- Ensure compliance with validation requirements (Section 3.2.2)

## Go Project Layout

- Follow [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- Use `/cmd`, `/internal`, `/pkg`, `/api`, `/configs`, `/scripts`, `/docs`, `/deployments`
- Keep main apps minimal, logic in `/internal` or `/pkg`
- Test directories may contain non-Go code (e.g., Java Gatling in `/test/gatling/`)
- Avoid: logic in `/cmd`, `/src` at root, deep nesting

## Pull Request Descriptions

### Title
- Use conventional commit format: `type(scope): description`
- Keep under 72 characters
- Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore

### Sections
- **📋 What**: Clear description of what PR does (present tense)
- **🎯 Why**: Business/technical rationale, impact on users/system
- **🔧 How**: High-level implementation approach, key decisions
- **✅ Testing**: How tested, coverage impact, manual steps
- **🔍 Breaking Changes**: Migration guidance, API changes
- **📚 Documentation**: README changes, migration guides

### Code Review Checklist
- **🔒 Security**: No sensitive data exposure, proper validation, secure defaults
- **🧪 Quality**: Tests added/updated, linting passes, docs updated
- **🚀 Performance**: No regressions, memory leaks addressed
- **🔧 Operations**: Logging appropriate, monitoring/metrics added

### PR Size Guidelines
- **🟢 Small (<200 lines)**: Single focused change, low risk
- **🟡 Medium (200-500 lines)**: Multiple related changes, moderate risk
- **🔴 Large (500+ lines)**: Complex feature, high risk, consider splitting
- **📦 Epic**: Break down into smaller, independently deployable PRs

## Documentation Organization

- Keep docs in 2 files: `README.md` (main), `docs/README.md` (deep dive)
- **ALWAYS add to existing README.md** instead of creating new markdown files
- **NEVER create separate documentation files** for scripts or tools
- See README for content distribution and organization
