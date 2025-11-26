# R10: Requirements Validation Automation

**Task ID**: R10
**Priority**: 📋 MEDIUM
**Effort**: 1 day (8 hours)
**Status**: ⏳ NOT STARTED (0%)
**Dependencies**: R07 (Repository Tests), R08 (OpenAPI Sync)

---

## Objective

Create automated requirements traceability tooling to map functional requirements from acceptance criteria to actual test implementations, ensuring complete validation coverage for production readiness.

---

## Context

**Current State**:
- Acceptance criteria defined in MASTER-PLAN.md for each task (R01-R11)
- Tests exist but no formal mapping to requirements
- No automated validation of requirements coverage
- Manual verification time-consuming and error-prone

**Target State**:
- Requirements extracted from acceptance criteria into machine-readable format
- Automated mapping from requirements to test files/functions
- Coverage report showing which requirements validated by which tests
- CI/CD integration to block deployment if requirements uncovered

---

## Deliverables

### D10.1: Requirements Extraction (2 hours)

**Extract requirements from MASTER-PLAN.md acceptance criteria:**

File: `docs/02-identityV2/requirements.yml`

```yaml
# Auto-generated from MASTER-PLAN.md acceptance criteria

requirements:
  R01-01:
    task: R01
    id: "R01-01"
    description: "/oauth2/v1/authorize stores request and redirects to login"
    category: "authorization_flow"
    priority: "CRITICAL"
    acceptance_criteria: "GET /oauth2/v1/authorize creates AuthorizationRequest in database and returns 302 redirect to /oidc/v1/login?request_id={uuid}"

  R01-02:
    task: R01
    id: "R01-02"
    description: "User login associates real user ID with authorization request"
    category: "authorization_flow"
    priority: "CRITICAL"
    acceptance_criteria: "POST /oidc/v1/login updates AuthorizationRequest.UserID with authenticated user's ID"

  R01-03:
    task: R01
    id: "R01-03"
    description: "Consent approval generates authorization code with user context"
    category: "authorization_flow"
    priority: "CRITICAL"
    acceptance_criteria: "POST /oidc/v1/consent generates authorization_code tied to UserID, not random UUID"

  R01-04:
    task: R01
    id: "R01-04"
    description: "/oauth2/v1/token exchanges code for tokens with real user ID"
    category: "token_exchange"
    priority: "CRITICAL"
    acceptance_criteria: "POST /oauth2/v1/token returns access_token with sub claim = UserID from AuthorizationRequest"

  R01-05:
    task: R01
    id: "R01-05"
    description: "Authorization code single-use enforced"
    category: "security"
    priority: "CRITICAL"
    acceptance_criteria: "Second POST /oauth2/v1/token with same code returns 400 invalid_grant error"

  R01-06:
    task: R01
    id: "R01-06"
    description: "Integration test validates end-to-end flow"
    category: "testing"
    priority: "HIGH"
    acceptance_criteria: "TestOAuth2AuthorizationCodeFlow passes without mocks, validates full flow from authorize to token"

  # R02-R11 requirements follow same pattern...
```

### D10.2: Requirements-Test Mapping Tool (4 hours)

**Create cicd utility**: `internal/cmd/cicd/identity_requirements_check/`

```go
// identity_requirements_check.go

package main

import (
    "context"
    "fmt"
    "os"

    "gopkg.in/yaml.v3"
)

type Requirement struct {
    Task              string `yaml:"task"`
    ID                string `yaml:"id"`
    Description       string `yaml:"description"`
    Category          string `yaml:"category"`
    Priority          string `yaml:"priority"`
    AcceptanceCriteria string `yaml:"acceptance_criteria"`
    TestFiles         []string `yaml:"test_files,omitempty"`
    TestFunctions     []string `yaml:"test_functions,omitempty"`
    Validated         bool     `yaml:"validated"`
}

type RequirementsDoc struct {
    Requirements map[string]Requirement `yaml:"requirements"`
}

func main() {
    ctx := context.Background()

    // Load requirements.yml
    reqDoc, err := loadRequirements("docs/02-identityV2/requirements.yml")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to load requirements: %v\n", err)
        os.Exit(1)
    }

    // Scan test files for requirement references
    testMappings, err := scanTestFiles(ctx, "./internal/identity")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to scan test files: %v\n", err)
        os.Exit(1)
    }

    // Map requirements to tests
    coverage := mapRequirementsToTests(reqDoc, testMappings)

    // Generate coverage report
    report := generateCoverageReport(coverage)
    fmt.Println(report)

    // Exit with error if uncovered requirements exist
    if hasUncoveredCriticalRequirements(coverage) {
        os.Exit(1)
    }
}

func scanTestFiles(ctx context.Context, rootPath string) (map[string][]string, error) {
    // Scan *_test.go files for comments like:
    // // Validates: R01-01, R01-02
    // func TestOAuth2AuthorizationFlow(t *testing.T) { ... }

    mappings := make(map[string][]string)

    // TODO: Implement test file scanning
    // Walk directory tree
    // Parse Go test files
    // Extract requirement IDs from comments
    // Map requirement ID → test function name

    return mappings, nil
}

func generateCoverageReport(coverage map[string]Requirement) string {
    // Generate markdown report:
    // # Requirements Coverage Report
    //
    // ## Summary
    // - Total Requirements: X
    // - Validated: Y (Z%)
    // - Uncovered CRITICAL: A
    // - Uncovered HIGH: B
    //
    // ## Coverage by Category
    // ### authorization_flow: 6/6 (100%)
    // ### token_exchange: 3/4 (75%)
    //
    // ## Uncovered Requirements
    // | ID | Priority | Description | Test Needed |
    // |----|----|----|----|
    // | R08-03 | CRITICAL | Swagger UI reflects real API | Manual validation |

    return "# Requirements Coverage Report\n\n..."
}
```

**Test annotation pattern**:

```go
// TestOAuth2AuthorizationCodeFlow validates the complete OAuth 2.1 authorization code flow.
//
// Validates requirements:
// - R01-01: /oauth2/v1/authorize stores request and redirects
// - R01-02: User login associates real user ID
// - R01-03: Consent generates code with user context
// - R01-04: Token exchange includes real user ID
// - R01-05: Authorization code single-use enforcement
// - R01-06: End-to-end integration test passes
func TestOAuth2AuthorizationCodeFlow(t *testing.T) {
    t.Parallel()
    // ... test implementation
}
```

### D10.3: Coverage Report Generation (1 hour)

**Output format**: `docs/02-identityV2/REQUIREMENTS-COVERAGE.md`

```markdown
# Identity V2 Requirements Coverage Report

**Generated**: 2025-11-23 16:30:00 UTC
**Total Requirements**: 65
**Validated**: 58 (89%)
**Uncovered CRITICAL**: 2
**Uncovered HIGH**: 3
**Uncovered MEDIUM**: 2

## Summary by Task

| Task | Requirements | Validated | Coverage |
|------|--------------|-----------|----------|
| R01 | 6 | 6 | 100% ✅ |
| R02 | 7 | 7 | 100% ✅ |
| R03 | 5 | 5 | 100% ✅ |
| R04 | 6 | 5 | 83% ⚠️ |
| R05 | 6 | 6 | 100% ✅ |
| R06 | 4 | 4 | 100% ✅ |
| R07 | 5 | 5 | 100% ✅ |
| R08 | 6 | 4 | 67% ⚠️ |
| R09 | 4 | 4 | 100% ✅ |
| R10 | 4 | 0 | 0% ❌ |
| R11 | 12 | 12 | 100% ✅ |

## Uncovered Requirements

### CRITICAL Priority

| ID | Task | Description | Recommended Test |
|----|------|-------------|------------------|
| R08-03 | R08 | Swagger UI reflects real API | Manual UI validation test |
| R08-04 | R08 | No placeholder/TODO endpoints in specs | Automated YAML linting |

### HIGH Priority

| ID | Task | Description | Recommended Test |
|----|------|-------------|------------------|
| R04-05 | R04 | Security tests validate attack prevention | Add penetration test suite |
| R08-02 | R08 | Client libraries functional | Add client SDK integration tests |

## Coverage by Category

### authorization_flow: 6/6 (100%) ✅
### token_exchange: 4/4 (100%) ✅
### security: 8/10 (80%) ⚠️
### testing: 12/12 (100%) ✅
### documentation: 4/6 (67%) ⚠️

## Validation Details

### R01-01: /oauth2/v1/authorize stores request and redirects
- **Status**: ✅ VALIDATED
- **Test**: `TestOAuth2AuthorizationCodeFlow` in `internal/identity/integration/integration_test.go:280`
- **Evidence**: Line 305 asserts 302 redirect with request_id parameter

### R01-02: User login associates real user ID
- **Status**: ✅ VALIDATED
- **Test**: `TestOAuth2AuthorizationCodeFlow` in `internal/identity/integration/integration_test.go:280`
- **Evidence**: Line 320 verifies UserID populated in AuthorizationRequest after login

[... continues for all requirements ...]

---

## Recommendations

1. **Immediate**: Add manual validation test for R08-03 (Swagger UI)
2. **High Priority**: Implement penetration test suite for R04-05
3. **Medium Priority**: Create client SDK integration tests for R08-02
4. **Low Priority**: Automate YAML linting for R08-04

---

**Report Generation Command**: `go run ./cmd/cicd identity-requirements-check`
**CI/CD Integration**: Add to `.github/workflows/ci-identity.yml` as quality gate
```

### D10.4: CI/CD Integration (1 hour)

**Add to `.github/workflows/ci-identity.yml`:**

```yaml
- name: Requirements Validation
  run: |
    go run ./cmd/cicd identity-requirements-check > docs/02-identityV2/REQUIREMENTS-COVERAGE.md

    # Fail workflow if critical requirements uncovered
    if grep -q "Uncovered CRITICAL: [1-9]" docs/02-identityV2/REQUIREMENTS-COVERAGE.md; then
      echo "::error::Critical requirements not validated by tests"
      exit 1
    fi
```

---

## Acceptance Criteria

- [ ] `requirements.yml` extracted from MASTER-PLAN.md with all R01-R11 acceptance criteria
- [ ] `identity-requirements-check` tool implemented and functional
- [ ] Test files annotated with `// Validates requirements: R01-01, R01-02` comments
- [ ] Coverage report generated showing requirements-to-test mappings
- [ ] CI/CD workflow fails if critical requirements uncovered
- [ ] Documentation explains how to add new requirements and annotate tests

---

## Implementation Notes

**Parsing Strategy**:
- Extract acceptance criteria from MASTER-PLAN.md using regex patterns
- Generate unique requirement IDs: `{TASK}-{SEQ}` (e.g., R01-01, R01-02)
- Store in YAML for easy editing and version control

**Test Annotation Strategy**:
- Use structured comments in test function godoc
- Pattern: `// Validates requirements: R##-##, R##-##`
- Parser scans for this pattern in *_test.go files

**Coverage Calculation**:
- Requirement VALIDATED if referenced by at least one test
- Report shows percentage by task and by priority level
- Highlight gaps for manual review

**CI/CD Integration**:
- Run on every PR to identity package
- Block merge if critical requirements uncovered
- Generate coverage report artifact for download

---

## Dependencies

**Requires Completion**:
- R07: Repository integration tests (provides test coverage baseline)
- R08: OpenAPI sync (provides endpoint validation tests)

**Provides Input To**:
- R11: Final verification (uses coverage report for readiness decision)

---

## Success Metrics

- [ ] 100% of CRITICAL requirements mapped to tests
- [ ] ≥90% of HIGH requirements mapped to tests
- [ ] ≥80% of MEDIUM requirements mapped to tests
- [ ] Coverage report generated in <10 seconds
- [ ] Zero false positives in requirement validation
- [ ] CI/CD integration catches gaps before merge

---

## Open Questions

1. **Manual vs Automated Tests**: How to track manual validation requirements (e.g., UI smoke tests)?
2. **Requirement Granularity**: Split complex acceptance criteria into multiple requirements?
3. **Test Multiplicity**: Allow one test to validate multiple requirements, or enforce 1:1 mapping?
4. **Historical Tracking**: Track coverage over time to show progress?

**Decisions**:
- Allow one test to validate multiple requirements (more realistic)
- Tag manual tests with `// Validates (manual): R##-##` for tracking
- Prioritize automation over 1:1 mapping
- Save coverage reports to git for historical trending

---

## Implementation Order

1. Extract R01-R11 requirements to `requirements.yml` (manual, 1 hour)
2. Implement test file scanner (2 hours)
3. Implement requirements mapper (1 hour)
4. Generate coverage report (1 hour)
5. Integrate with CI/CD (30 min)
6. Annotate existing tests with requirement IDs (2 hours)
7. Verify coverage report accuracy (30 min)

**Total Estimated Effort**: 8 hours
**Actual Effort**: TBD (track in post-mortem)
