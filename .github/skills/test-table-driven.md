# test-table-driven

Generate table-driven Go tests conforming to cryptoutil project standards.

## Purpose

Use this skill when writing or reviewing Go tests. Ensures correct patterns:
`t.Parallel()`, `UUIDv7` test data, `require` over `assert`, proper subtests.

## Key Rules

- `t.Parallel()` MANDATORY on parent and ALL subtests
- Use `googleUuid.NewV7()` for test data IDs (thread-safe, unique, no conflicts)
- `require` package (fail-fast) over `assert` (continue-on-failure)
- Table-driven for ALL multi-case tests (happy path AND sad path)
- TestMain for heavyweight resources (DB, servers, containers) — one per package
- Fiber `app.Test()` for ALL HTTP handler tests (no real network listeners)

## Template

```go
func TestXxx_Description(t *testing.T) {
t.Parallel()
tests := []struct {
name    string
input   TypeA
want    TypeB
wantErr string
}{
{name: "happy path basic", input: ..., want: ...},
{name: "error case missing field", input: ..., wantErr: "missing X"},
}
for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()
got, err := FunctionUnderTest(tc.input)
if tc.wantErr != "" {
require.ErrorContains(t, err, tc.wantErr)
return
}
require.NoError(t, err)
require.Equal(t, tc.want, got)
})
}
}
```

## TestMain Pattern (heavyweight resources)

```go
var (
testDB     *gorm.DB
testServer *Server
)

func TestMain(m *testing.M) {
ctx := context.Background()
container, _ := postgres.RunContainer(ctx,
postgres.WithDatabase(fmt.Sprintf("test_%s", googleUuid.NewV7())),
postgres.WithUsername(fmt.Sprintf("user_%s", googleUuid.NewV7())),
)
defer container.Terminate(ctx)
// initialize testDB, testServer...
os.Exit(m.Run())
}
```

## References

See [ARCHITECTURE.md Section 10.2 Unit Testing Strategy](../../docs/ARCHITECTURE.md#102-unit-testing-strategy) for full testing requirements.
