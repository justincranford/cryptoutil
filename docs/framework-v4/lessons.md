# Framework v4 - Lessons Learned

---

## Phase 1: Fix Legacy sm-kms-pg- Naming and Add OTLP Service Name Check

**Completed**: commit `dc5970d47`

### Lessons

1. **PowerShell heredoc strips leading whitespace**: `@'...'@` removes leading spaces/tabs from each line. Go files written via PowerShell heredoc have zero indentation and fail `gofumpt`. Fix: always run `gofmt -w` + `gofumpt -w` immediately after writing any Go file via PowerShell.

2. **Magic literal checker covers more than expected**: The `lint-go` `literal-use` linter enforces magic constants for service name strings (`"sm-im"`, `"sm-kms"`, `"jose-ja"`), file permission octals (`0o600`, `0o755`), and path segment strings (`"im"`, `"kms"`). Tests must use `cryptoutilSharedMagic.*` constants for all known service names and permissions — even in table-driven test data fields.

3. **backtick strings in `wantErrContain` are exempt**: Composite strings like `` `got "sm-kms-pg-1", want "sm-kms-postgres-1"` `` inside backtick raw string literals are NOT flagged by the literal checker. Only standalone string assignments trigger the check.

4. **Multi_replace_string_in_file fails on mixed whitespace**: The tool uses exact string matching. If the file uses tabs and the replacement string uses spaces, it will fail silently. When in doubt, read the file first with `Get-Content -Raw` to see exact bytes.

5. **`configs/orphaned/` exclusion is mandatory**: The `orphaned/` directory contains legacy configs with intentionally incorrect naming. Any new OTLP name check MUST explicitly skip this directory or it will produce false positives on every run.

---

## Phase 2: Registry-Driven Foundation and Entity Registry Check

**Completed**: commit `7bae5aee3`

### Lessons

1. **`lint-go` literal-use flags test case `name:` fields too**: The `magic-usage` linter flags string literals in table-driven test struct fields — including `name: "sm-im"` used only as a `t.Run` label. These must use the corresponding magic constant (e.g., `name: cryptoutilSharedMagic.OTLPServiceSMIM`) even though the semantic meaning is just "test display name". Rationale: the linter enforces consistency mechanically, not semantically.

2. **Count literals need semantic constants**: Literal integers like `5` and `10` in `require.Len()` assertions are flagged. The linter suggested `JWECompactParts = 5` (wrong semantics) and `PercentageBasis10 = 10` (also wrong). The correct fix was to add `SuiteProductCount = 5` to `magic_cicd.go` alongside the existing `SuiteServiceCount = 10`. This is the right abstraction: the count is meaningful domain knowledge, not just any arbitrary integer.

3. **Display name strings need title-case constants**: `"Skeleton"` (title-case display name for the Skeleton product) is a magic literal requiring `SkeletonProductNameTitleCase`. Scan existing magic constants before creating new ones — `SkeletonProductNameTitleCase` already existed.

4. **Suite ID `"cryptoutil"` uses `DefaultOTLPServiceDefault`**: The suite name matches the default OTLP service name. Reuse `DefaultOTLPServiceDefault = "cryptoutil"` even in non-OTLP contexts where it refers to the suite identity.

5. **`configs/{PRODUCT}/{SERVICE}/` structure already existed**: All 10 `configs/pki/ca/`, `configs/jose/ja/`, etc. directories were created in the prior session and were already git-tracked. The entity-registry-completeness check confirmed all 10 pass on the real workspace immediately. No remediation was needed.

6. **Magic file names (e.g., `"magic_sm.go"`) are NOT flagged by literal-use**: The checker does not have entries for the magic file names themselves. Only service ID strings (`"sm-im"`) and product names (`"pki"`) are registered in the checker's allowlist. When writing `TestCheckInDir_MissingMagicFile`, the `magicFile:` strings like `"magic_sm.go"` can remain as-is.

7. **Registry defensive copies prevent mutation in tests**: `AllProducts()`, `AllProductServices()`, `AllSuites()` each use `copy()` to return independent slices. The `TestAllReturnsIndependentCopies` test validates this invariant. This is important because callers in fitness checks iterate the registry without expecting surprises from prior test runs.

---

## Phase 3: Banned Name Detection

*(No notes yet — phase not started.)*

---

## Phase 4: Deployment Directory Completeness

*(No notes yet — phase not started.)*

---

## Phase 5: Compose File Header and Service Name Validation

*(No notes yet — phase not started.)*

---

## Phase 6: Magic Constants Cross-Reference Validation

*(No notes yet — phase not started.)*

---

## Phase 7: Standalone Config File Presence and Naming

*(No notes yet — phase not started.)*

---

## Phase 8: Migration Comment Header Validation

*(No notes yet — phase not started.)*

---

## Phase 9: ARCHITECTURE.md Updates and CICD Tool Catalog

*(No notes yet — phase not started.)*

---

## Phase 10: Knowledge Propagation

*(No notes yet — phase not started.)*
