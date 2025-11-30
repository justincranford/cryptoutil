# golangci-lint v2 Migration - Summary

## Migration Complete ✅

Successfully migrated `.golangci.yml` from mixed v1/v2 configuration to clean golangci-lint v2.6.2 schema.

## Strategy

**Incremental 12-step migration** with verification after each step:

1. Reset to minimal v2 config (header + version only)
2. Add configuration in order of execution speed (fastest → slowest)
3. Commit with pre-commit verification after each step
4. Document problems/changes in `migrate-v2-problems.md`

## Key v2 API Changes Encountered

### Removed Properties (v1 → v2)

**Run Section:**

- `skip-dirs` - removed (directory exclusions now automatic for generated code)
- `skip-files` - removed (file pattern exclusions now automatic)
- `build-cache` - removed (always enabled in v2)
- `modules-download-mode` - removed

**Output Section:**

- `formats.text` structure → `formats` map with format names as keys
- `sort-results: true` → `sort-order: [linter, severity, file]`

**Linters Section:**

- `gosimple` - merged into `staticcheck`
- `stylecheck` - merged into `staticcheck`
- `gofmt` - now a formatter, not a linter
- `gofumpt` - now a formatter, not a linter
- `goimports` - now a formatter, not a linter
- `wsl` - deprecated, replaced by `wsl_v5`

**Settings Section:**

- `goconst.ignore-tests` - removed
- `misspell.ignore-words` - removed
- `wrapcheck.ignoreSigs` - removed (API changed)

**Issues/Exclusions Section:**

- `exclude-dirs` - removed
- `exclude-files` - removed
- `exclude-rules` - removed
- `exclude-generated` - removed
- `exclude-dirs-use-default` - removed

**Severity Section:**

- `default-severity` → `default`

### New/Changed Properties (v2)

**Output:**

- `formats` is now a map: `formats: { tab: { path: stdout } }`
- `sort-order` replaces `sort-results`

**Linters:**

- `wsl_v5` replaces deprecated `wsl`
- `depguard` now requires explicit `rules` configuration

**Issues:**

- Simplified to just `max-issues-per-linter` and `max-same-issues`
- Most exclusions now automatic via generated code detection

## Migration Steps Detail

| Step | Description | Commit | Status |
|------|-------------|--------|--------|
| 0 | Strategy docs | 3a054d63 | ✅ |
| 0 | Minimal config reset | d64fa850 | ✅ |
| 1 | Core configuration | ce424756 | ✅ |
| 2 | Output configuration | 64f23999 | ✅ |
| 3 | Fast essential linters | 70c42917 | ✅ |
| 4 | Fast code quality linters | c4115f29 | ✅ |
| 5 | Security & error handling | 32956208 | ✅ |
| 6 | Testing quality linters | 6a51c398 | ✅ |
| 7 | Performance & style linters | e05cc85b | ✅ |
| 8 | Maintainability & headers | a5e7e06a | ✅ |
| 9 | Linter-specific settings | 609143ae | ✅ |
| 10 | Import alias configuration | 43bdce6e | ✅ |
| 11-12 | Issues & severity | c14988a8 | ✅ |
| Final | Fix wsl_v5 & depguard | 3d846909 | ✅ |

## Final Configuration

**Linters Enabled (22):**

- Essential: errcheck, govet, ineffassign, staticcheck, unused
- Code Quality: revive, godot, copyloopvar, goconst, importas
- Security: gosec, noctx, wrapcheck, errorlint
- Testing: thelper, tparallel, testpackage, gomodguard, gomoddirectives
- Performance: prealloc, bodyclose, mnd, wsl_v5, nlreturn
- Maintainability: goheader, depguard

**Linters Disabled (3):**

- dupl (code duplication)
- gocyclo (cyclomatic complexity)
- godox (TODO/FIXME tracking)

**Key Settings:**

- 60+ import aliases (cryptoutil packages + dependencies)
- gosec excludes: G204, G301, G302, G304, G402
- mnd ignored: '2', math.*, len, make
- depguard: block github.com/pkg/errors
- Severity: error (default), warning (revive, godot), info (misspell)

## Testing Results

**Final Configuration Validates:**

```bash
golangci-lint run --timeout=10m
```

- Configuration loads successfully
- No schema validation errors
- Linters execute without deprecation warnings (except context-specific)
- Finds expected issues in codebase

## Documentation Updates Needed

After migration complete, update these files:

1. **`.github/instructions/01-06.linting.instructions.md`**
   - Update v2 API changes
   - Document removed properties
   - Update wsl → wsl_v5
   - Add depguard rules configuration

2. **`docs/pre-commit-hooks.md`**
   - Update golangci-lint configuration section
   - Document v2-specific settings
   - Update formatter vs linter distinction

3. **`.golangci.yml` inline comments**
   - Add v2 migration notes
   - Document why certain v1 properties removed
   - Reference this migration summary

## Lessons Learned

### What Worked Well

1. **Incremental approach** - Catch issues early, easier to debug
2. **Fast-to-slow ordering** - Get basic functionality first
3. **Commit after each step** - Clear history, easy rollback
4. **Problem documentation** - Track API changes for future reference

### Challenges Encountered

1. **Incomplete documentation** - v2 schema not fully documented
2. **Trial and error** - Some settings required testing to discover v2 syntax
3. **Deprecation warnings** - wsl → wsl_v5 not obvious from error messages
4. **Missing exclusion patterns** - v2 simplified exclusions, need to understand new auto-detection

### Recommendations

1. **Always test incrementally** - Don't migrate everything at once
2. **Document API changes** - Future migrations will benefit
3. **Use schema validation** - VS Code JSON schema catches many issues
4. **Keep backup config** - Reference for complex settings migration

## Next Steps

1. ✅ Migration complete
2. ⬜ Update instruction files (`.github/instructions/01-06.linting.instructions.md`)
3. ⬜ Update pre-commit hooks documentation (`docs/pre-commit-hooks.md`)
4. ⬜ Test full CI/CD pipeline with new config
5. ⬜ Monitor for any unexpected linter behavior changes
6. ⬜ Clean up: Delete `.golangci.yml.backup` after validation period

## References

- [golangci-lint v2 Release Notes](https://github.com/golangci/golangci-lint/releases/tag/v2.0.0)
- [golangci-lint Configuration Docs](https://golangci-lint.run/docs/configuration/file/)
- [JSON Schema](https://json.schemastore.org/golangci-lint.json)
- [Reference Config](https://github.com/golangci/golangci-lint/blob/HEAD/.golangci.reference.yml)
