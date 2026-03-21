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

*(No notes yet — phase not started.)*

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
