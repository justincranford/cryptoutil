# Lessons Learned - Summary

Quick-reference lessons from quality enforcement and refactoring sessions.

| # | Lesson | Category |
|---|--------|----------|
| 1 | Use per-function t.Parallel() enforcement, not package-level | Testing |
| 2 | magic_usage: only literal-use blocks are errors; const-redefine are informational | Linting |
| 3 | Replace hardcoded URL strings with URLPrefixLocalhostHTTPS magic constant | Code Quality |
| 4 | NEVER write Go files via shell heredocs - tabs get stripped | Tooling |
| 5 | Use Python writes for Go files requiring exact whitespace (tabs) | Tooling |
| 6 | Extract duplicated mergedFS logic into shared MergedMigrationsFS utility | Architecture |
| 7 | Always run go build after every file write to verify correctness | Workflow |
| 8 | All magic values go in internal/shared/magic/, NEVER inline | Linting |

See [details.md](details.md) for full explanations.
