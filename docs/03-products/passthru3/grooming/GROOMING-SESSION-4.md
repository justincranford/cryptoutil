# Grooming Session 4: Deployment & Final Verification

**Purpose**: Define deployment validation and closure criteria BEFORE implementation
**Date**: 2025-12-01

---

## Topic 1: Docker Compose Deployment Validation

### Q1.1: What Docker Compose files need validation?

**Files to Validate**:

| File | Profile | Purpose |
|------|---------|---------|
| `deployments/identity/compose.demo.yml` | demo | Identity server demo |
| `deployments/kms/compose.demo.yml` | demo | KMS server demo |
| `deployments/telemetry/compose.yml` | (default) | Observability stack |

### Q1.2: Validation sequence

**Step-by-Step Validation**:

```bash
# 1. Ensure clean state
docker compose -f deployments/identity/compose.demo.yml --profile demo down -v
docker compose -f deployments/kms/compose.demo.yml --profile demo down -v
docker compose -f deployments/telemetry/compose.yml down -v

# 2. Validate configs (no errors in output)
docker compose -f deployments/identity/compose.demo.yml --profile demo config > /dev/null
docker compose -f deployments/kms/compose.demo.yml --profile demo config > /dev/null
docker compose -f deployments/telemetry/compose.yml config > /dev/null

# 3. Start telemetry first (other services depend on it)
docker compose -f deployments/telemetry/compose.yml up -d

# 4. Wait for telemetry ready
sleep 10
docker compose -f deployments/telemetry/compose.yml ps  # All should be "running"

# 5. Start Identity
docker compose -f deployments/identity/compose.demo.yml --profile demo up -d

# 6. Wait for Identity ready
sleep 10
curl -k https://localhost:8082/.well-known/openid-configuration  # Should return JSON

# 7. Start KMS
docker compose -f deployments/kms/compose.demo.yml --profile demo up -d

# 8. Wait for KMS ready
sleep 10
curl -k https://localhost:8080/ui/swagger/doc.json  # Should return JSON

# 9. Verify all services
docker compose -f deployments/identity/compose.demo.yml --profile demo ps
docker compose -f deployments/kms/compose.demo.yml --profile demo ps
docker compose -f deployments/telemetry/compose.yml ps

# 10. Cleanup
docker compose -f deployments/kms/compose.demo.yml --profile demo down -v
docker compose -f deployments/identity/compose.demo.yml --profile demo down -v
docker compose -f deployments/telemetry/compose.yml down -v
```

---

## Topic 2: End-to-End Verification

### Q2.1: Complete verification checklist

**Code Level**:

- [ ] `go build ./...` succeeds
- [ ] `golangci-lint run ./...` returns 0 errors
- [ ] `grep -c "TODO" internal/cmd/demo/integration.go` returns 0
- [ ] `go test ./internal/cmd/demo/...` passes

**Demo Level**:

- [ ] `go run ./cmd/demo kms` shows 4/4 PASSED
- [ ] `go run ./cmd/demo identity` shows 5/5 PASSED
- [ ] `go run ./cmd/demo all` shows 7/7 PASSED

**Docker Level**:

- [ ] Identity compose config validates
- [ ] KMS compose config validates
- [ ] Identity compose up/down works
- [ ] KMS compose up/down works
- [ ] Services can communicate over telemetry-network

### Q2.2: Evidence capture commands

```bash
# Capture all evidence in single session

# 1. Build evidence
go build ./... 2>&1 | tee evidence_build.txt

# 2. Lint evidence
golangci-lint run ./... 2>&1 | tee evidence_lint.txt

# 3. Test evidence
go test -v ./internal/cmd/demo/... 2>&1 | tee evidence_test.txt

# 4. KMS demo evidence
go run ./cmd/demo kms 2>&1 | tee evidence_demo_kms.txt

# 5. Identity demo evidence
go run ./cmd/demo identity 2>&1 | tee evidence_demo_identity.txt

# 6. Integration demo evidence
go run ./cmd/demo all 2>&1 | tee evidence_demo_integration.txt

# 7. TODO count evidence
echo "TODO count in integration.go: $(grep -c 'TODO' internal/cmd/demo/integration.go 2>/dev/null || echo 0)" | tee evidence_todos.txt

# 8. Docker compose evidence
docker compose -f deployments/identity/compose.demo.yml --profile demo config > evidence_compose_identity.txt 2>&1
docker compose -f deployments/kms/compose.demo.yml --profile demo config > evidence_compose_kms.txt 2>&1
```

---

## Topic 3: Known Issues & Workarounds

### Q3.1: What known issues exist?

| Issue | Impact | Workaround |
|-------|--------|------------|
| Integration tests timeout (600s) | Medium | Use shorter demo-specific timeouts |
| Port 55679 conflict on Windows | Fixed | Changed to 15679 in telemetry compose |
| Self-signed cert warnings | Expected | Use -k flag or InsecureSkipVerify |
| OTEL trace submission failures | Low | Non-blocking, logs warning |

### Q3.2: What issues are NOT blockers?

**Not Blockers**:

- MFA/OTP not implemented (Phase 3 work)
- WebAuthn passkey not implemented (Phase 3 work)
- Notification webhooks stub (Phase 3 work)
- Integration test suite timeout (separate from demo)

---

## Topic 4: Closure Criteria

### Q4.1: What must be true for passthru3 to close?

**Hard Requirements** (ALL must be true):

| # | Criterion | Verification |
|---|-----------|--------------|
| 1 | No TODOs in integration.go | `grep -c "TODO" integration.go` = 0 |
| 2 | All demos pass | kms: 4/4, identity: 5/5, integration: 7/7 |
| 3 | Lint passes | `golangci-lint run ./internal/cmd/demo/...` = 0 errors |
| 4 | Build passes | `go build ./...` exits 0 |
| 5 | Docker configs valid | All compose config commands succeed |
| 6 | Evidence documented | EVIDENCE.md complete |
| 7 | REQUIREMENTS-CHECKLIST complete | All R1-R6 verified |

### Q4.2: Sign-off requirements

**Required Sign-offs**:

1. **Code Complete**: All integration.go TODOs resolved
2. **Quality Complete**: Lint and build pass
3. **Demo Complete**: All demo commands work
4. **Evidence Complete**: EVIDENCE.md has all captured outputs
5. **Documentation Complete**: All grooming and tracking docs updated

---

## Topic 5: Post-Passthru3 Work

### Q5.1: What is explicitly OUT of scope?

**Out of Scope for Passthru3**:

| Item | Why | Future Pass |
|------|-----|-------------|
| MFA/OTP implementation | Phase 3 work | passthru4 |
| WebAuthn passkey | Phase 3 work | passthru4 |
| Email/Webhook notifications | Phase 3 work | passthru4 |
| Production deployment | Separate concern | deployment pass |
| Load testing | Separate concern | performance pass |
| Security hardening | Separate concern | security pass |

### Q5.2: Definition of "productized"

For passthru3, "productized" means:

1. **Demo works end-to-end** without manual intervention
2. **Code is clean** (no TODOs, lint passes)
3. **Documentation is complete** (all grooming sessions documented)
4. **Evidence is captured** (all verification commands run and output saved)
5. **Can be demonstrated** to stakeholders at any time

It does NOT mean:

- Production-ready (that's future work)
- Feature-complete (MFA, WebAuthn are future)
- Fully tested (just demo-level testing)

---

## Sign-Off

**All deployment and closure criteria in this document are LOCKED**

- [ ] Reviewed and approved
- [ ] No open questions remain
- [ ] Ready for implementation

**Date**: ____________
**Approved By**: ____________
