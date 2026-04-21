# GAP-0.5: Usage String Deduplication Deferred to V16

**Status**: Deferred
**Task**: 0.5 — Refactor Duplicate usage.go Files
**Gap**: 2.1 (HIGH) from gaps.md

## Root Cause

4 service pairs have nearly identical `usage.go` files at both the PS-ID level and the product
sub-package level:

| PS-ID file | Product sub-package file |
|-----------|--------------------------|
| `internal/apps/sm-kms/kms_usage.go` | `internal/apps/sm/kms/kms_usage.go` |
| `internal/apps/sm-im/im_usage.go` | `internal/apps/sm/im/im_usage.go` |
| `internal/apps/jose-ja/ja_usage.go` | `internal/apps/jose/ja/ja_usage.go` |
| `internal/apps/pki-ca/ca_usage.go` | `internal/apps/pki/ca/ca_usage.go` |

## Why Deferred

The usage strings are declared as Go `const` blocks. Extracting them to a shared package requires
changing them to `var` initialized by a function call, touching 7+ files across 4 product trees.
The mechanical change is straightforward but the risk surface (linter reactions to `const→var`,
test compilation across all product packages) exceeded the 2h budget cap stated in the plan.

## Recommended V16 Approach

1. Create `internal/apps/framework/service/usage/` package with:
   ```go
   func BuildUsageMain(productCmd, serviceCmd, serviceName, configFile string) string
   func BuildUsageServer(productCmd, serviceCmd, configFile string) string
   // etc.
   ```
2. Replace all 8 `const` blocks with `var` blocks calling the shared functions.
3. Run `go build ./... && golangci-lint run ./...` to verify no regressions.
4. Delete this file when complete.

## Impact Assessment

- **Risk**: LOW — usage strings are only surfaced in CLI help output, not in runtime logic.
- **Coverage impact**: NONE — `internal/shared/magic/` coverage exclusion does not apply here,
  but usage strings have no testable branch logic.
- **Deferral cost**: Minimal — the duplication is cosmetic text, not shared logic.
