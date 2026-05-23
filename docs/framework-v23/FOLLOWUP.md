# Framework V23 — Remaining Followup Work

Only unresolved work is listed here.

---

## 1. sm-im E2E Registration Flow Returns 500 Instead of 201

**Status**: Open blocker (runtime behavior)

**Failing command**:
```
go test -tags e2e ./internal/apps/sm-im/e2e/...
```

**Latest observed failure (2026-05-23)**:
```
--- FAIL: TestE2E_RegistrationFlowWithTenantCreation (0.00s)
    --- FAIL: TestE2E_RegistrationFlowWithTenantCreation/sm-im-app-postgresql-1_service (0.16s)
        Error: Not equal:
            expected: 201
            actual  : 500
        Messages: Registration with create_tenant=true should return 201 Created
FAIL    cryptoutil/internal/apps/sm-im/e2e      126.160s
```

**Important update**:
- The prior `cryptoutil-postgres-leader exited (1)` startup blocker is no longer the active failure.
- In the latest run, `cryptoutil-postgres-leader` reached healthy and the stack advanced to test execution.

**Likely investigation surface**:
- `internal/apps/sm-im/e2e/e2e_registration_test.go` (failing assertion around line 80)
- Registration/create-tenant handler path in sm-im service runtime
- Application logs for `sm-im-app-postgresql-1` during the failing request

**Required evidence for closure**:
1. Capture request/response and service log evidence explaining the 500 root cause.
2. Implement corrective change in code/config as indicated by root cause.
3. Re-run and pass:
   ```
   go test -tags e2e ./internal/apps/sm-im/e2e/...
   ```
