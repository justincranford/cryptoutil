# Architecture Fitness Functions

## What Are Fitness Functions?

Term from 'Building Evolutionary Architectures' (Ford, Parsons, Kua, 2022).
A fitness function is an automated assessment of a specific architectural goal.

They occupy the gap between unit tests (is this function correct?),
linters (is this style correct?), and architecture compliance
(do these 10 services follow agreed patterns?).

Examples from real companies:
- Netflix: 'No service may call another directly; must use API gateway'
- Amazon: 'All services must have independent deployment pipelines'
- Google: 'No package may import another at greater than N dependency hops'

## Why cryptoutil Needs Them

10 services, 1 developer, constraints written in ARCHITECTURE.md.
ARCHITECTURE.md is a document. Documents drift. Code is ground truth.

Without fitness functions, the following happen silently:
- A service imports from another service's internal package (coupling)
- A test uses a hardcoded port instead of port 0 (Windows CI failures)
- A handler registered without required middleware chain
- Migration number conflicts between template (1001-1004) and domain (2001+)
- A service uses bcrypt instead of PBKDF2 (FIPS violation)
- An admin endpoint bound to 0.0.0.0 instead of 127.0.0.1

Fitness functions catch these at commit-time, not code-review-time.

## Category 1: Dependency Isolation

### Rule: Service packages must not import other service packages

```go
// Generalize the existing go-check-identity-imports to all services
// For each internal/apps/{product}/{service}/ package,
// verify it does not import any other service package.
// shared packages (internal/shared/) are allowed
// template packages (internal/apps/template/) are allowed
func CheckImportIsolation(cfg Config) []Violation { ... }
```n
### Rule: Domain packages must not import server or client packages

```go
// internal/apps/{product}/{service}/domain/ must not import
// /server/, /client/, api/model, api/server, api/client
// This enforces clean hexagonal architecture
```n
## Category 2: Service Structure Completeness

### Rule: Every service must have the required directory layers

```go
var requiredLayers = []string{
    "domain", "repository", "server", "server/apis",
}
var requiredTestLayers = []string{
    "integration", "e2e",
}
func CheckServiceStructure(serviceRoot string) []Violation { ... }
```n
## Category 3: API Contract Integrity

### Rule: Server dirs must contain an OpenAPI spec

```go
var requiredAPIFiles = []string{
    "openapi_spec_paths.yaml",
    "openapi_spec_components.yaml",
    "openapi-gen_config_server.yaml",
}
```n
### Rule: Generated code must be up-to-date

```bash
# Run oapi-codegen, check git diff
# Non-empty diff = fitness function fails
# Message: 'Generated code is out of date. Run make generate.'
```n

## Category 4: Security Constraints

### Rule: No banned cryptographic algorithms

```go
var bannedCryptoImports = []string{
    "golang.org/x/crypto/bcrypt",
    "golang.org/x/crypto/scrypt",
    "golang.org/x/crypto/argon2",
    "crypto/md5",
    "crypto/sha1",
}
var bannedPatterns = []string{
    "InsecureSkipVerify: true",
    "math/rand",  // must use crypto/rand
}
```n
### Rule: Admin endpoints must bind to 127.0.0.1

```go
// AST walk for BindPrivateAddress field assignments
// Fail if any value other than '127.0.0.1' found
```n
### Rule: Tests must use port 0, not hardcoded ports

```go
// AST walk in _test.go files
// Find ServerSettings{} literals
// Verify BindPublicPort == 0 and BindPrivatePort == 0
// Hardcoded port = Windows CI failure waiting to happen
```n
## Category 5: Test Quality

### Rule: All test functions must call t.Parallel()

```go
// AST walk for func Test* declarations
// Verify first statement is t.Parallel()
// Verify subtests (t.Run lambdas) call t.Parallel()
// Exempt: TestMain
```n
### Rule: No hardcoded UUIDs in tests

```go
var hardcodedUUIDPattern = regexp.MustCompile(
    `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
)
// Scan _test.go files for UUID literals
// Recommend: googleUuid.NewV7()
```n
## Category 6: Framework Version Consistency

### Rule: All services must use the same ServerBuilder version

```go
// Once ServiceManifest is adopted:
// Scan all service manifest files for FrameworkVersion field
// Fail if any service is behind the current version
// Report: upgrade command per service
```n

## Implementation: Running Fitness Functions

### Integration in cicd command suite

```bash
go run ./cmd/cicd fitness-check
go run ./cmd/cicd fitness-check --service pki-ca
go run ./cmd/cicd fitness-check --category security
```n
### Fitness function registry (self-documenting)

```go
type FitnessFunction struct {
    Name        string
    Category    FitnessCategory
    Description string
    Check       func(root string) []Violation
    Severity    Severity // Error or Warning
}

var registeredFunctions = []FitnessFunction{
    {Name: "ImportIsolation", Category: CategoryDependency,
     Desc: "Services must not import other service packages",
     Check: checkImportIsolation, Severity: SeverityError},
    {Name: "BannedCrypto", Category: CategorySecurity,
     Desc: "No bcrypt/scrypt/argon2/MD5/SHA1",
     Check: checkBannedCrypto, Severity: SeverityError},
    // ... all other functions
}
```n
### GitHub Actions integration

```yaml
# .github/workflows/ci-fitness.yml
name: Architecture Fitness Functions
on: [push, pull_request]
jobs:
  fitness:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v6
        with: {go-version: '1.26.1', cache: true}
      - name: Run architecture fitness functions
        run: go run ./cmd/cicd fitness-check
```n
## Effort and Adoption Strategy

Phase 1 (1 week): Import isolation + banned crypto checks
  These exist partially today (go-check-identity-imports). Generalize.

Phase 2 (1 week): Test quality checks (t.Parallel, port 0, no hardcoded UUIDs)
  Highest impact on Windows CI reliability.

Phase 3 (1 week): Service structure completeness + API contract integrity
  Requires skeleton structure to be defined first (see 05-skeleton-scaffolding.md).

Phase 4 (2 weeks): Security constraints (127.0.0.1, TLS config, auth patterns)
  Most precise rules require AST inspection, not just grep.

Total: ~5 weeks for a comprehensive fitness function suite.
Return: automated enforcement of every architectural constraint in ARCHITECTURE.md.
