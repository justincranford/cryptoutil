---
description: "Generate custom quality checklists that validate requirements completeness"
---

# /speckit.checklist

Generate custom quality checklists that validate requirements completeness, clarity, and consistency.

## User Input

```
$ARGUMENTS
```

You MUST consider the user input before proceeding (if not empty).

## Outline

1. **Analyze context**:
   - Read specification and plan documents
   - Identify quality dimensions relevant to the feature
   - Consider project constitution requirements

2. **Generate checklists** for relevant areas:
   - **UX Checklist**: User experience validation
   - **Security Checklist**: Security requirements coverage
   - **Test Checklist**: Test coverage validation
   - **API Checklist**: API design consistency
   - **Performance Checklist**: Performance requirements
   - **Accessibility Checklist**: Accessibility requirements

3. **Checklist format**:

```markdown
# [Area] Checklist

## [Category 1]

- [ ] [Specific, verifiable item]
- [ ] [Specific, verifiable item]

## [Category 2]

- [ ] [Specific, verifiable item]
```

**Output**: Write checklists to `FEATURE_DIR/checklists/[area].md`.

## cryptoutil-Specific Checklists

### Security Checklist

- [ ] All cryptographic operations use FIPS 140-3 approved algorithms
- [ ] Secrets never in environment variables (use Docker/K8s secrets)
- [ ] Input validation on all API endpoints
- [ ] Rate limiting implemented
- [ ] Audit logging for security events
- [ ] Token validation includes expiration and revocation checks
- [ ] CSRF protection enabled
- [ ] CORS properly configured

### Code Quality Checklist

- [ ] `go build ./...` succeeds
- [ ] `golangci-lint run` passes with no errors
- [ ] Test coverage ≥90% for new code
- [ ] No `//nolint:` directives added
- [ ] File size ≤300 lines (or justified)
- [ ] Table-driven tests with `t.Parallel()`
- [ ] Conventional commit messages used

### API Checklist

- [ ] OpenAPI spec updated for new endpoints
- [ ] Request/response validation
- [ ] Proper HTTP status codes
- [ ] Error responses follow standard format
- [ ] Pagination for list endpoints
- [ ] Rate limiting headers included

### Docker Compose Checklist

- [ ] Health checks defined for all services
- [ ] Secrets mounted via `/run/secrets/`
- [ ] Named volumes for persistent data
- [ ] Resource limits configured
- [ ] Network isolation between services
- [ ] Service naming follows `product-service-instance` pattern
