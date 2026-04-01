Generate table-driven Go tests for the specified function or package.

**Full Copilot original**: [.github/skills/test-table-driven/SKILL.md](.github/skills/test-table-driven/SKILL.md)

## Rules

- `t.Parallel()` MANDATORY in every test function and every subtest
- Use `googleUuid.NewV7()` for all test data IDs (not hardcoded strings or UUIDs)
- Use `require` (not `assert`) — fail-fast on first failure
- Table-driven with named subtests via `t.Run(tc.name, ...)`
- HTTP handler tests: use `app.Test()` (Fiber in-memory, NO real servers)
- TestMain pattern for heavyweight resources (DB, servers) — init once per package
- `<15s` per package; parallel execution required

## Template

```go
func TestFunctionName(t *testing.T) {
    t.Parallel()

    testCases := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    InputType{ID: googleUuid.NewV7()},
            expected: ExpectedType{...},
        },
        {
            name:    "invalid input",
            input:   InputType{},
            wantErr: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            result, err := FunctionUnderTest(tc.input)

            if tc.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            require.Equal(t, tc.expected, result)
        })
    }
}
```

## TestMain Pattern (heavyweight resources)

```go
var (
    sharedDB     *gorm.DB
    sharedServer *application.Application
)

func TestMain(m *testing.M) {
    ctx := context.Background()
    var err error

    sharedDB, err = repository.InitSQLite(ctx, cryptoutilSharedMagic.SQLiteInMemoryDSN, migrationsFS)
    if err != nil {
        log.Fatalf("failed to init db: %v", err)
    }

    os.Exit(m.Run())
}
```
