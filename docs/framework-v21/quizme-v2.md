# Quiz Me - Framework V21: Canonical PS-ID Recursive Structure (Round 2)

**Created**: 2026-04-30
**Purpose**: Close the remaining Q2 decision by selecting the canonical recursive directory structure that will be enforced for all 10 PS-IDs.

---

## Research Snapshot (Evidence-Based)

### Requested Focus Services (interpreting `jose-ca` as `jose-ja`)

- `sm-kms` currently has: `server/businesslogic`, `server/handler`, `server/repository`, `server/repository/migrations`, `server/repository/orm`
- `sm-im` currently has: `server/apis`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`
- `jose-ja` currently has: `server/apis`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`, `server/service`
- `skeleton-template` currently has: `server/apis`, `server/config`, `server/handler`, `server/model`, `server/repository`, `server/repository/migrations`

## Recursive Directory Inventory with Per-Directory CSV File Lists

### Group 1: sm-kms, sm-im, jose-ja, skeleton-template

| filename | allow? | sm-kms | sm-im | jose-ja | skeleton-template | comment |
|---|---|---|---|---|---|---|
| internal/apps/{PS-ID}/**PS_ID**.go | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical PS-ID entrypoint file; required across services. |
| internal/apps/{PS-ID}/__PS_ID___test.go | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical test file. |
| internal/apps/{PS-ID}/client/client.go | ✓ | ✓ | ✓ | ✓ | ✓ | Root client package is in required root set. |
| internal/apps/{PS-ID}/client/client_*.go | ✓ | ✓ | ✓ | ✓ | ✓ | Client implementation split files are allowed. |
| internal/apps/{PS-ID}/client/package_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | Package-level client tests are allowed. |
| internal/apps/{PS-ID}/client/*_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | Client tests are expected/allowed. |
| internal/apps/{PS-ID}/e2e/e2e_*.go | ✓ | ✓ | ✓ | ✓ | ✓ | E2E directory is part of canonical root set. |
| internal/apps/{PS-ID}/e2e/testmain_e2e_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | E2E TestMain pattern is allowed. |
| internal/apps/{PS-ID}/testing/testmain_helper.go | ✓ | ✓ | ✓ | ✓ | ✓ | testing is optional-but-approved root helper module. |
| internal/apps/{PS-ID}/testing/testmain_helper_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | testing helper tests are allowed with optional testing root. |
| internal/apps/{PS-ID}/domain/*.go |  | ✓ | ✓ | ✓ | ✓ | Root domain package is outside strict root policy; migrate behind canonical boundaries. |
| internal/apps/{PS-ID}/domain/*_test.go |  | ✓ | ✓ | ✓ | ✓ | Same as above; root domain tests should move with domain migration. |
| internal/apps/{PS-ID}/repository/*.go |  | ✓ | ✓ | ✓ | ✓ | Root repository package is non-canonical once server/repository is enforced. |
| internal/apps/{PS-ID}/repository/*_test.go |  | ✓ | ✓ | ✓ | ✓ | Non-canonical root repository tests should sunset with root repository. |
| internal/apps/{PS-ID}/repository/migrations.go |  | ✓ | ✓ | ✓ | ✓ | Migration registry should live under server/repository in canonical layout. |
| internal/apps/{PS-ID}/repository/migrations/*.up.sql |  | ✓ | ✓ | ✓ | ✓ | Canonical SQL path is server/repository/migrations, not root repository. |
| internal/apps/{PS-ID}/repository/migrations/*.down.sql |  | ✓ | ✓ | ✓ | ✓ | Canonical SQL path is server/repository/migrations, not root repository. |
| internal/apps/{PS-ID}/server/server.go | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical server composition file. |
| internal/apps/{PS-ID}/server/public_server.go | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical dual-listener public server wiring. |
| internal/apps/{PS-ID}/server/admin.go | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical admin endpoint wiring. |
| internal/apps/{PS-ID}/server/service.go | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical service bootstrap file at server root. |
| internal/apps/{PS-ID}/server/validator.go | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical server config/validation wiring file. |
| internal/apps/{PS-ID}/server/swagger.go | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical OpenAPI wiring file. |
| internal/apps/{PS-ID}/server/swagger_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical OpenAPI wiring test file. |
| internal/apps/{PS-ID}/server/*_lifecycle_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | Lifecycle test pattern is allowed in server root. |
| internal/apps/{PS-ID}/server/*_port_conflict_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | Port conflict tests remain mandatory. |
| internal/apps/{PS-ID}/server/*_integration_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | Server integration tests are allowed/required by quality gates. |
| internal/apps/{PS-ID}/server/testmain_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | Shared server TestMain is canonical. |
| internal/apps/{PS-ID}/server/apis/*.go | ✓ |  | ✓ | ✓ | ✓ | Required canonical directory; missing services are migration gaps. |
| internal/apps/{PS-ID}/server/apis/*_test.go | ✓ |  | ✓ | ✓ | ✓ | Required canonical directory; tests should follow apis migration. |
| internal/apps/{PS-ID}/server/config/config.go | ✓ |  | ✓ | ✓ | ✓ | Required canonical directory for unified config ownership. |
| internal/apps/{PS-ID}/server/config/config_test.go | ✓ |  | ✓ | ✓ | ✓ | Required canonical config test pattern. |
| internal/apps/{PS-ID}/server/config/config_test_helper.go | ✓ |  | ✓ | ✓ | ✓ | Allowed helper in canonical config package. |
| internal/apps/{PS-ID}/server/config/config_*_test.go | ✓ |  | ✓ | ✓ | ✓ | Allowed config variant tests in canonical package. |
| internal/apps/{PS-ID}/server/model/model.go | ✓ |  | ✓ | ✓ | ✓ | Required canonical model package. |
| internal/apps/{PS-ID}/server/model/models.go | ✓ |  | ✓ | ✓ | ✓ | Allowed split model declarations in canonical model package. |
| internal/apps/{PS-ID}/server/model/*_test.go | ✓ |  | ✓ | ✓ | ✓ | Required model test coverage in canonical location. |
| internal/apps/{PS-ID}/server/repository/*.go | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical repository package. |
| internal/apps/{PS-ID}/server/repository/*_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical repository tests. |
| internal/apps/{PS-ID}/server/repository/migrations.go | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical migration registry location. |
| internal/apps/{PS-ID}/server/repository/migrations/*.up.sql | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical SQL migration path. |
| internal/apps/{PS-ID}/server/repository/migrations/*.down.sql | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical SQL migration path. |
| internal/apps/{PS-ID}/server/businesslogic/*.go | ✓ | ✓ |  |  |  | Required canonical businesslogic package; missing services need migration. |
| internal/apps/{PS-ID}/server/businesslogic/*_bench_test.go | ✓ | ✓ |  |  |  | Allowed benchmark tests in canonical businesslogic package. |
| internal/apps/{PS-ID}/server/businesslogic/*_fuzz_test.go | ✓ | ✓ |  |  |  | Allowed fuzz tests in canonical businesslogic package. |
| internal/apps/{PS-ID}/server/businesslogic/*_property_test.go | ✓ | ✓ |  |  |  | Allowed property tests in canonical businesslogic package. |
| internal/apps/{PS-ID}/server/businesslogic/*_test.go | ✓ | ✓ |  |  |  | Required unit tests in canonical businesslogic package. |
| internal/apps/{PS-ID}/server/handler/*.go | ✓ | ✓ |  |  | ✓ | Transitional allowlist only; sunset after move to apis/businesslogic. |
| internal/apps/{PS-ID}/server/handler/*_test.go | ✓ | ✓ |  |  | ✓ | Transitional allowlist only; sunset with handler package. |
| internal/apps/{PS-ID}/server/service/*.go | ✓ |  |  | ✓ |  | Transitional allowlist only; sunset after consolidation. |
| internal/apps/{PS-ID}/server/service/*_test.go | ✓ |  |  | ✓ |  | Transitional allowlist only; sunset after consolidation. |
| internal/apps/{PS-ID}/server/repository/orm/*.go | ✓ | ✓ |  |  |  | Transitional allowlist only; sunset after repository unification. |
| internal/apps/{PS-ID}/server/repository/orm/*_test.go | ✓ | ✓ |  |  |  | Transitional allowlist only; sunset after repository unification. |

### Group 2: pki-ca, identity-*

| filename | allow? | pki-ca | identity-authz | identity-idp | identity-rs | identity-rp | identity-spa | comment |
|---|---|---|---|---|---|---|---|---|
| internal/apps/{PS-ID}/**PS_ID**.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical PS-ID entrypoint file; required across services. |
| internal/apps/{PS-ID}/__PS_ID___test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical test file. |
| internal/apps/{PS-ID}/*.TODO |  | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | TODO marker files are technical debt and should not be template/lint allowed. |
| internal/apps/{PS-ID}/client/client.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Root client package is in required root set. |
| internal/apps/{PS-ID}/client/package_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Package-level client tests are allowed. |
| internal/apps/{PS-ID}/client/client_*.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Client implementation split files are allowed. |
| internal/apps/{PS-ID}/client/*_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Client tests are expected/allowed. |
| internal/apps/{PS-ID}/e2e/*_e2e_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | E2E directory is part of canonical root set. |
| internal/apps/{PS-ID}/e2e/testmain_e2e_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | E2E TestMain pattern is allowed. |
| internal/apps/{identity-*}/unified/**PS_ID**.go | ✓ |  | ✓ | ✓ | ✓ | ✓ | ✓ | unified is explicitly optional-approved root module for identity services. |
| internal/apps/{identity-authz,identity-idp}/auth/*.go | ✓ |  | ✓ | ✓ |  |  |  | Authn/authz-specific module is explicitly optional-approved. |
| internal/apps/{identity-authz,identity-idp}/auth/*_test.go | ✓ |  | ✓ | ✓ |  |  |  | Authn/authz-specific tests remain allowed with optional module. |
| internal/apps/{identity-authz,identity-idp}/clientauth/*.go | ✓ |  | ✓ | ✓ |  |  |  | Authn/authz-specific module is explicitly optional-approved. |
| internal/apps/{identity-authz,identity-idp}/clientauth/*_test.go | ✓ |  | ✓ | ✓ |  |  |  | Authn/authz-specific tests remain allowed with optional module. |
| internal/apps/{identity-authz}/dpop/*.go | ✓ |  | ✓ |  |  |  |  | Identity-specific authz extension module is optional-approved. |
| internal/apps/{identity-authz}/dpop/*_test.go | ✓ |  | ✓ |  |  |  |  | Identity-specific authz extension tests are optional-approved. |
| internal/apps/{identity-authz}/pkce/*.go | ✓ |  | ✓ |  |  |  |  | Identity-specific authz extension module is optional-approved. |
| internal/apps/{identity-authz}/pkce/*_test.go | ✓ |  | ✓ |  |  |  |  | Identity-specific authz extension tests are optional-approved. |
| internal/apps/{identity-idp}/userauth/*.go | ✓ |  |  | ✓ |  |  |  | Identity-specific authn extension module is optional-approved. |
| internal/apps/{identity-idp}/userauth/*_test.go | ✓ |  |  | ✓ |  |  |  | Identity-specific authn extension tests are optional-approved. |
| internal/apps/{identity-idp}/userauth/mocks/*.go | ✓ |  |  | ✓ |  |  |  | Test/mock support under approved identity extension module is allowed. |
| internal/apps/{identity-idp}/userauth/mocks/*_test.go | ✓ |  |  | ✓ |  |  |  | Test/mock support tests under approved identity extension module are allowed. |
| internal/apps/pki-ca/api/*.go |  | ✓ |  |  |  |  |  | pki-ca root api package is legacy sprawl; should move behind canonical server boundaries. |
| internal/apps/pki-ca/api/handler/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root handler package should consolidate into server/apis/businesslogic. |
| internal/apps/pki-ca/api/handler/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root handler tests should move with consolidation. |
| internal/apps/pki-ca/bootstrap/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package; stage-2 consolidation target. |
| internal/apps/pki-ca/bootstrap/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package tests; stage-2 consolidation target. |
| internal/apps/pki-ca/cli/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package; stage-2 consolidation target. |
| internal/apps/pki-ca/cli/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package tests; stage-2 consolidation target. |
| internal/apps/pki-ca/compliance/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package; stage-2 consolidation target. |
| internal/apps/pki-ca/compliance/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package tests; stage-2 consolidation target. |
| internal/apps/pki-ca/config/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package; stage-2 consolidation target. |
| internal/apps/pki-ca/config/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package tests; stage-2 consolidation target. |
| internal/apps/pki-ca/crypto/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package; stage-2 consolidation target. |
| internal/apps/pki-ca/crypto/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package tests; stage-2 consolidation target. |
| internal/apps/pki-ca/domain/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package; stage-2 consolidation target. |
| internal/apps/pki-ca/domain/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package tests; stage-2 consolidation target. |
| internal/apps/pki-ca/domain-v2/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package; stage-2 consolidation target. |
| internal/apps/pki-ca/domain-v2/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package tests; stage-2 consolidation target. |
| internal/apps/pki-ca/intermediate/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package; stage-2 consolidation target. |
| internal/apps/pki-ca/intermediate/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package tests; stage-2 consolidation target. |
| internal/apps/pki-ca/observability/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package; stage-2 consolidation target. |
| internal/apps/pki-ca/observability/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package tests; stage-2 consolidation target. |
| internal/apps/pki-ca/security/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package; stage-2 consolidation target. |
| internal/apps/pki-ca/security/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package tests; stage-2 consolidation target. |
| internal/apps/pki-ca/storage/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package; stage-2 consolidation target. |
| internal/apps/pki-ca/storage/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca root package tests; stage-2 consolidation target. |
| internal/apps/pki-ca/profile/certificate/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca profile package should move behind canonical boundaries. |
| internal/apps/pki-ca/profile/certificate/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca profile tests should move with profile consolidation. |
| internal/apps/pki-ca/profile/subject/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca profile package should move behind canonical boundaries. |
| internal/apps/pki-ca/profile/subject/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca profile tests should move with profile consolidation. |
| internal/apps/pki-ca/service/{issuer,ra,revocation,timestamp}/*.go |  | ✓ |  |  |  |  |  | Legacy pki-ca service root packages should consolidate into canonical server structure. |
| internal/apps/pki-ca/service/{issuer,ra,revocation,timestamp}/*_bench_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca benchmark tests should move with service consolidation. |
| internal/apps/pki-ca/service/{issuer,ra,revocation,timestamp}/*_test.go |  | ✓ |  |  |  |  |  | Legacy pki-ca service tests should move with service consolidation. |
| internal/apps/pki-ca/repository-v2/migrations.go |  | ✓ |  |  |  |  |  | repository-v2 is legacy; migration registry should move to server/repository. |
| internal/apps/pki-ca/repository-v2/migrations_test.go |  | ✓ |  |  |  |  |  | repository-v2 tests should move with repository consolidation. |
| internal/apps/pki-ca/repository-v2/migrations/*.up.sql |  | ✓ |  |  |  |  |  | SQL canonical path is server/repository/migrations; repository-v2 path should sunset. |
| internal/apps/pki-ca/repository-v2/migrations/*.down.sql |  | ✓ |  |  |  |  |  | SQL canonical path is server/repository/migrations; repository-v2 path should sunset. |
| internal/apps/{PS-ID}/server/server.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical server composition file. |
| internal/apps/{PS-ID}/server/public_server.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical dual-listener public server wiring. |
| internal/apps/{PS-ID}/server/admin.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical admin endpoint wiring. |
| internal/apps/{PS-ID}/server/service.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical service bootstrap file at server root. |
| internal/apps/{PS-ID}/server/validator.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical server config/validation wiring file. |
| internal/apps/{PS-ID}/server/swagger.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical OpenAPI wiring file. |
| internal/apps/{PS-ID}/server/swagger_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical OpenAPI wiring test file. |
| internal/apps/{PS-ID}/server/*_lifecycle_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Lifecycle test pattern is allowed in server root. |
| internal/apps/{PS-ID}/server/*_port_conflict_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Port conflict tests remain mandatory. |
| internal/apps/{PS-ID}/server/*_integration_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Server integration tests are allowed/required by quality gates. |
| internal/apps/{PS-ID}/server/*_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Additional server tests are allowed. |
| internal/apps/{PS-ID}/server/testmain_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Shared server TestMain is canonical. |
| internal/apps/{PS-ID}/server/apis/*.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical directory for endpoint definitions. |
| internal/apps/{PS-ID}/server/apis/*_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical apis test pattern. |
| internal/apps/identity-idp/server/apis/templates/*.html | ✓ |  |  | ✓ |  |  |  | Transitional allowlist item (server/apis/templates); sunset after handler/template consolidation. |
| internal/apps/{PS-ID}/server/config/config.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical directory for unified config ownership. |
| internal/apps/{PS-ID}/server/config/config_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical config test pattern. |
| internal/apps/{PS-ID}/server/config/config_test_helper.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Allowed helper in canonical config package. |
| internal/apps/{PS-ID}/server/config/config_*_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Allowed config variant tests in canonical package. |
| internal/apps/{PS-ID}/server/model/model.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical model package. |
| internal/apps/{PS-ID}/server/model/*_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Required model test coverage in canonical location. |
| internal/apps/{PS-ID}/server/repository/migrations.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Canonical migration registry location. |
| internal/apps/{PS-ID}/server/repository/*.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical repository package. |
| internal/apps/{PS-ID}/server/repository/*_test.go | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical repository tests. |
| internal/apps/{identity-*}/server/repository/migrations/*.up.sql | ✓ |  | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical SQL path; pki-ca currently has migration-gap to close. |
| internal/apps/{identity-*}/server/repository/migrations/*.down.sql | ✓ |  | ✓ | ✓ | ✓ | ✓ | ✓ | Required canonical SQL path; pki-ca currently has migration-gap to close. |
| internal/apps/pki-ca/server/cmd/*.go | ✓ | ✓ |  |  |  |  |  | Transitional allowlist item (server/cmd); sunset after server package consolidation. |
| internal/apps/pki-ca/server/cmd/*_test.go | ✓ | ✓ |  |  |  |  |  | Transitional allowlist item (server/cmd); sunset after server package consolidation. |
| internal/apps/pki-ca/server/middleware/*.go | ✓ | ✓ |  |  |  |  |  | Transitional allowlist item (server/middleware); sunset after middleware integration. |
| internal/apps/pki-ca/server/middleware/*_test.go | ✓ | ✓ |  |  |  |  |  | Transitional allowlist item (server/middleware); sunset after middleware integration. |

### Cross-Table Consistency Analysis (Deep Pass)

1. Canonical server directories are uniformly marked allowed in both tables: server/apis, server/businesslogic, server/config, server/model, server/repository, server/repository/migrations.
2. Transitional allowlist directories are consistently marked allowed (with sunset comment) in both tables where they appear: server/handler, server/service, server/cmd, server/middleware, server/repository/orm, server/apis/templates.
3. Non-canonical root sprawl is consistently marked not allowed: Group 1 root domain/repository and Group 2 pki-ca root api/bootstrap/cli/compliance/config/crypto/domain/domain-v2/intermediate/observability/security/storage/profile/service/repository-v2.
4. Optional root modules are consistently treated as allowed only when explicitly policy-approved: testing (Group 1), unified/auth/clientauth/dpop/pkce/userauth/userauth/mocks (Group 2 identity services).
5. Debt marker files are consistently treated as not allowed by template/lint-fitness policy: *.TODO is disallowed even if currently present.
6. Presence-vs-policy gaps are called out without contradiction: rows can be allowed but currently missing in a service (migration gap), or disallowed but currently present (legacy debt).
7. No row contradicts the selected policy direction (transitional canonical server set + strict root policy + staged pki-ca consolidation).

### pki-ca SQL Migration Evidence

- Current migration SQL files are in:
  - `internal/apps/pki-ca/repository-v2/migrations/5001_ca_items.up.sql`
  - `internal/apps/pki-ca/repository-v2/migrations/5001_ca_items.down.sql`

---

## Question 1: Canonical `server/**` recursive structure to enforce for all 10 PS-IDs

**Question**: Which policy should V21 adopt as the target canonical recursive `server/**` structure across all 10 PS-IDs (with linter/template enforcement)?

**A)** Strict immediate canonical set:
- Required everywhere: `server/apis`, `server/businesslogic`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`
- Forbidden everywhere: `server/handler`, `server/service`, `server/cmd`, `server/middleware`, `server/repository/orm`, `server/apis/templates`
- One-shot migration for all 10 in V21

**B)** Transitional canonical set with sunset (recommended):
- Required everywhere: `server/apis`, `server/businesslogic`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`
- Temporary allowlist (must be retired by scheduled phases): `server/handler`, `server/service`, `server/cmd`, `server/middleware`, `server/repository/orm`, `server/apis/templates`
- Linter enforces required-now plus time-boxed deprecation plan

**C)** Minimal convergence:
- Require only: `server/apis`, `server/model`, `server/repository`
- Keep service-specific subdirectories indefinitely (no sunset)

**D)** Keep current mixed structure and only ensure required dirs exist (no consolidation mandate)

**E)**

**Answer**:

**Rationale**: This decision controls the all-10 migration scope, linter invariants, and how aggressively sprawl (especially pki-ca) is reduced.

---

## Question 2: pki-ca consolidation strategy under the selected canonical policy

**Question**: For pki-ca package/subdirectory sprawl, which execution strategy should tasks implement?

**A)** Full consolidation in V21:
- Move/merge pki-ca subdirectories to canonical targets immediately
- Migrate domain packages that sit outside canonical paths
- Remove legacy directories in same phase

**B)** Two-stage consolidation (recommended):
- Stage 1 (V21): establish canonical `server/**` directories, introduce wrappers/adapters, migrate SQL paths from `repository-v2/migrations` to `server/repository/migrations`
- Stage 2 (next phase): move domain-heavy packages (`bootstrap`, `compliance`, `intermediate`, `profile`, `service`, `storage`, etc.) behind canonical boundaries and remove legacy paths after compatibility gates pass

**C)** Structural-only for V21:
- Create canonical dirs and linter checks
- Keep pki-ca legacy package sprawl untouched

**D)** pki-ca-specific exception:
- Exempt pki-ca from canonical structure and keep bespoke layout

**E)**

**Answer**:

**Rationale**: Determines whether V21 includes concrete pki-ca sprawl reduction tasks versus deferring most consolidation work.

---

## Question 3: Root-level PS-ID directory policy for all 10 services

**Question**: Should V21 enforce a canonical root-level PS-ID directory policy in addition to `server/**` policy?

**A)** Yes, strict required-only root set for all 10 (recommended):
- Required: `client`, `e2e`, `server`
- Optional (explicitly approved only): `testing`, `unified`, authn/authz-specific modules
- All other root-level directories must be migrated or explicitly sunset

**B)** Yes, but service-class based policy:
- Identity services may keep additional authn/authz roots
- pki-ca may keep additional PKI roots
- SM/JOSE services follow strict root set

**C)** No root-level policy in V21; enforce only `server/**`

**D)** Keep current root-level sprawl and rely on naming conventions only

**E)**

**Answer**:

**Rationale**: This controls whether V21 includes all-10 root-level cleanup tasks or limits scope to `server/**` only.
