# Phase 0 Implementation Guide - Slow Test Optimization

**Objective**: Optimize all 11 slow test packages from ~600s total to <200s total  
**Strategy**: Shared test infrastructure via TestMain + aggressive parallelization  
**Evidence**: Test execution time measurements before/after

---

## Implementation Pattern (Apply to ALL 11 Packages)

### Step 1: Add TestMain for Shared Infrastructure

```go
package packagename

import (
	"context"
	"os"
	"testing"
	
	googleUuid "github.com/google/uuid"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

var (
	testRepoFactory *cryptoutilIdentityRepository.RepositoryFactory
	testCtx         context.Context
)

func TestMain(m *testing.M) {
	testCtx = context.Background()
	
	// Create shared in-memory SQLite database ONCE for entire package
	dsn := "file::memory:?cache=shared"
	
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:            "sqlite",
		DSN:             dsn,
		MaxOpenConns:    5,  // Allow concurrent test access
		MaxIdleConns:    5,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
		AutoMigrate:     true,
	}
	
	var err error
	testRepoFactory, err = cryptoutilIdentityRepository.NewRepositoryFactory(testCtx, dbConfig)
	if err != nil {
		panic(err)
	}
	
	err = testRepoFactory.AutoMigrate(testCtx)
	if err != nil {
		panic(err)
	}
	
	// Run all tests
	exitCode := m.Run()
	
	// Cleanup
	_ = testRepoFactory.Close()
	
	os.Exit(exitCode)
}
```

### Step 2: Update Test Functions to Use Shared Infrastructure

**BEFORE (Slow - creates DB per test)**:
```go
func TestSomething(t *testing.T) {
	t.Parallel()
	
	repoFactory, ctx := setupTestRepository(t)  // ❌ Creates new DB + migrations
	defer repoFactory.Close()
	
	// Test code using repoFactory
}
```

**AFTER (Fast - uses shared DB with unique data)**:
```go
func TestSomething(t *testing.T) {
	t.Parallel()
	
	// Use global testRepoFactory from TestMain
	// Create unique test data with UUIDv7
	clientID := googleUuid.NewV7()
	client := &cryptoutilIdentityDomain.Client{
		ID:   clientID,
		Name: "test-client-" + clientID.String(),
		// ... other fields
	}
	
	repo := testRepoFactory.ClientRepository()
	err := repo.Create(testCtx, client)
	require.NoError(t, err)
	
	// Test code - data is orthogonal (unique UUIDs = no conflicts)
	
	// Optional: cleanup if needed (usually not necessary for in-memory DB)
	// _ = repo.Delete(testCtx, clientID)
}
```

### Step 3: Verify Parallelization

Ensure ALL test functions have `t.Parallel()` immediately after function start:

```bash
# Check for missing t.Parallel()
grep -r "func Test" internal/identity/authz/clientauth/*_test.go | \
  while read line; do
    file=$(echo $line | cut -d: -f1)
    func=$(echo $line | cut -d: -f2)
    if ! grep -A 2 "$func" "$file" | grep -q "t.Parallel()"; then
      echo "MISSING t.Parallel(): $file - $func"
    fi
  done
```

---

## Package-Specific Optimizations

### P0.1: clientauth (168s → <30s)

**Current Bottleneck**: `integration_test.go` creates new DB + migrations per test  
**Solution**: TestMain with shared repository factory  
**Files**:
- `internal/identity/authz/clientauth/integration_test.go` - Add TestMain, update all tests
- All `*_test.go` files - Verify t.Parallel() present

**Validation**:
```bash
go test ./internal/identity/authz/clientauth -v  # Should show <30s
```

### P0.2: jose/server (94s → <20s)

**Current Bottleneck**: Fiber app startup/shutdown per test  
**Solution**: TestMain with single shared Fiber app on dynamic port  
**Files**:
- `internal/jose/server/*_test.go` - Add TestMain with shared server
- Use same port allocation pattern as `internal/server/application/application_test.go`

**Validation**:
```bash
go test ./internal/jose/server -v  # Should show <20s
```

### P0.3: kms/client (74s → <20s)

**Current Bottleneck**: KMS server startup per test  
**Solution**: TestMain with single shared KMS server instance  
**CRITICAL**: MUST use real KMS server (NOT MOCKS)  
**Files**:
- `internal/kms/client/*_test.go` - Add TestMain with real KMS server

**Implementation**:
```go
var (
	testKMSServer *kmsApp.Server
	testServerURL string
)

func TestMain(m *testing.M) {
	// Start KMS server ONCE with in-memory SQLite
	config := &kmsConfig.Config{
		BindAddress: "127.0.0.1",
		Port:        0,  // Dynamic port allocation
		Database: &kmsConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  "file::memory:?cache=shared",
		},
	}
	
	var err error
	testKMSServer, err = kmsApp.NewServer(config)
	if err != nil {
		panic(err)
	}
	
	go testKMSServer.Start()
	
	// Get actual port
	testServerURL = fmt.Sprintf("https://127.0.0.1:%d", testKMSServer.ActualPort())
	
	exitCode := m.Run()
	
	_ = testKMSServer.Shutdown()
	os.Exit(exitCode)
}
```

**Validation**:
```bash
go test ./internal/kms/client -v  # Should show <20s
```

### P0.4: jose (67s → <15s)

**Current Bottleneck**: Crypto operations + low coverage (48.8%)  
**Solution**: 
1. Increase coverage to 95% FIRST (add missing tests)
2. Then apply parallelization optimizations  
**Files**:
- `internal/jose/*_test.go` - Add tests for uncovered code
- Focus on JWK generation, JWS/JWE edge cases

**Validation**:
```bash
go test ./internal/jose -cover  # Should show ≥95%
go test ./internal/jose -v      # Should show <15s
```

### P0.5: kms/server/application (28s → <10s)

**Current Bottleneck**: Server startup/shutdown per test  
**Solution**: TestMain with shared server instance  
**Files**:
- `internal/kms/server/application/*_test.go`

**Validation**:
```bash
go test ./internal/kms/server/application -v  # Should show <10s
```

### P0.6: identity/authz (19s → <10s)

**Current Bottleneck**: Database operations (already has t.Parallel())  
**Solution**: Review test data isolation, optimize transaction patterns  
**Files**:
- `internal/identity/authz/*_test.go`

### P0.7: identity/idp (15s → <10s)

**Current Bottleneck**: Low coverage (54.9%) + database setup  
**Solution**: 
1. Increase coverage to 80%+ FIRST
2. Use TestMain for shared DB
**Files**:
- `internal/identity/idp/*_test.go`

### P0.8: identity/test/unit (18s → <10s)

**Current Bottleneck**: Infrastructure test setup  
**Solution**: Review and optimize test patterns  
**Files**:
- `internal/identity/test/unit/*_test.go`

### P0.9: identity/test/integration (16s → <10s)

**Current Bottleneck**: Docker container startup  
**Solution**: Optimize container lifecycle management  
**Files**:
- `internal/identity/test/integration/*_test.go`

### P0.10: infra/realm (14s → <10s)

**Current Bottleneck**: Configuration loading  
**Solution**: Apply t.Parallel(), reduce redundant config loading  
**Files**:
- `internal/infra/realm/*_test.go`

### P0.11: kms/server/barrier (13s → <10s)

**Current Bottleneck**: Crypto operations  
**Solution**: Parallelize crypto tests, reduce key generation redundancy  
**Files**:
- `internal/kms/server/barrier/*_test.go`

---

## Validation Strategy

### Before Starting

```bash
# Baseline current test times
go test ./... -v 2>&1 | grep "^ok" | sort -k3 -n -r | head -20 > baseline_times.txt
```

### After Each Package Optimization

```bash
# Verify specific package
go test ./internal/path/to/package -v

# Verify coverage maintained/improved
go test ./internal/path/to/package -cover

# Update PROGRESS.md checklist
```

### After Phase 0 Complete

```bash
# Verify all tests pass
go test ./... -shuffle=on

# Verify total time <200s
go test ./... -v 2>&1 | grep "^ok" | awk '{sum+=$3} END {print "Total:", sum, "seconds"}'

# Compare to baseline
go test ./... -v 2>&1 | grep "^ok" | sort -k3 -n -r | head -20 > optimized_times.txt
diff baseline_times.txt optimized_times.txt
```

---

## Success Criteria

✅ **DONE** when:
- All 11 packages optimized to targets
- Total test suite time <200s (currently ~600s)
- All tests pass with `-shuffle=on`
- Coverage maintained or improved
- No test failures
- All changes committed and pushed

---

## Common Pitfalls

❌ **DON'T**:
- Use mocks for happy path tests (use real dependencies via TestMain)
- Create new database per test (use shared DB with unique UUIDs)
- Skip t.Parallel() (required for concurrent execution)
- Batch commits (commit after each package optimization)

✅ **DO**:
- Use TestMain for shared infrastructure
- Use UUIDv7 for data isolation
- Verify tests pass AND are faster
- Update PROGRESS.md after each task
- Commit incrementally with evidence in commit message

---

## Implementation Order

1. P0.1 (clientauth) - Biggest impact (168s → <30s = 138s saved)
2. P0.2 (jose/server) - Second biggest (94s → <20s = 74s saved)
3. P0.3 (kms/client) - Third biggest (74s → <20s = 54s saved)
4. P0.4 (jose) - Coverage dependency (must reach 95% first)
5. P0.5 (kms/server/app) - Smaller but important
6. P0.6-P0.11 - Secondary packages in parallel

**Total Expected Savings**: ~400+ seconds (67% reduction)

---

**Next**: After Phase 0 complete, proceed to Phase 1 (CI/CD Workflow Fixes)
