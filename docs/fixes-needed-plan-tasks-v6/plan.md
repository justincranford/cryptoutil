# Implementation Plan - Service Template & CICD Fixes

**Status**: Planning
**Created**: 2026-01-31
**Last Updated**: 2026-01-31

## Overview

This plan consolidates incomplete work from v6 analysis. It covers service-template fixes, CICD enforcement improvements, and test conformance issues.

## Technical Context

- **Language**: Go 1.25.5
- **Framework**: Fiber v2, GORM
- **Database**: PostgreSQL OR SQLite with GORM
- **Testing**: Table-driven tests, TestMain pattern, app.Test() for handlers

## Phases

### Phase 1: Critical Fixes (TODOs and Security) - 4-6h

- Complete registration handler TODOs (password hashing, validation)
- Add admin middleware to registration routes
- Implement realm lookup for multi-tenant deployments

### Phase 2: Test Architecture Refactoring - 6-8h

- Refactor listener tests to use app.Test() instead of real HTTPS listeners
- Consolidate standalone tests to table-driven pattern
- Ensure t.Parallel() in all tests

### Phase 3: Coverage Improvements - 4-6h

- Repository package: 84.8% → 95%
- Application package: 89.8% → 95%
- Businesslogic package: 87.4% → 95%
- Config packages: 86.9%/87.1% → 95%
- Other below-target packages

### Phase 4: Code Cleanup - 2-3h

- Verify and remove dead code
- Fix config bug acknowledged in tests

### Phase 5: CICD Enforcement Improvements - 6-8h (NEW)

HIGH Priority linters to add:
- Docker secrets pattern enforcement
- testify require over assert
- t.Parallel() enforcement
- crypto/rand over math/rand
- No inline env vars in compose
- No InsecureSkipVerify

### Phase 6: Deployment and Workflow - 2-3h

- Create template deployment (deployments/template/compose.yml)
- Document template testing strategy

### Phase 7: Copilot Instructions Updates - 2-3h

- Update 02-02.service-template.instructions.md
- Update 03-02.testing.instructions.md
- Update 03-08.server-builder.instructions.md

## Technical Decisions

### Decision 1: Test Architecture Pattern
- **Chosen**: app.Test() for all handler tests, TestMain for heavyweight resources
- **Rationale**: Prevents Windows Firewall prompts, faster execution, no port binding
- **Alternatives**: Real HTTPS listeners (rejected - triggers firewall, slower)

### Decision 2: Table-Driven Tests
- **Chosen**: Single table-driven test per error category
- **Rationale**: Reduced duplication, easier to add cases, faster execution
- **Alternatives**: Standalone functions (rejected - violates 03-02.testing.instructions.md)

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Test refactoring breaks existing tests | Medium | Medium | Run tests after each file refactored |
| Coverage improvement takes longer than estimated | Medium | Low | Prioritize critical gaps first |
| CICD linters have false positives | Low | Medium | Test extensively before enabling |

## Quality Gates

- ✅ All tests pass (`go test ./...`)
- ✅ Coverage ≥95% production, ≥98% infrastructure
- ✅ Linting clean (`golangci-lint run`)
- ✅ No new TODOs without tracking
- ✅ No standalone tests (table-driven only)
- ✅ No real HTTPS listeners in unit tests
- ✅ All copilot instructions accurate

## Success Criteria

- [ ] Phase 1-7 complete
- [ ] All quality gates pass
- [ ] CI/CD green
- [ ] Documentation updated

## References

- Analysis docs archived in [archive/](./archive/)
- Copilot instructions in `.github/instructions/`
