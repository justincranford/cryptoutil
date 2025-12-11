# Test Performance: Local vs GitHub Comparison

## CRITICAL FINDING: Workflow Overhead is the Bottleneck

**All test packages execute FAST locally (<15s), but take 10-20x longer on GitHub workflows.**

| Package | Local Time | GitHub Time | Slowdown Factor | Root Cause |
|---------|------------|-------------|-----------------|------------|
| sqlrepository | <2s | 303s | **150x** | Workflow overhead |
| clientauth | 8.3s | 168s | **20x** | Workflow overhead |
| jose/server | 10.8s | 100s | **9x** | Workflow overhead |
| kms/client | 12.3s | 74s | **6x** | Workflow overhead |
| jose | 14.2s | 67s | **5x** | Workflow overhead |
| **TOTALS** | **<50s** | **712s** | **14x average** | **Workflow overhead** |

---

## Key Insights

### 1. Test Code is NOT the Problem

**Evidence**:

- All packages use `t.Parallel()` correctly (verified in timing reports)
- Individual tests execute in 0.00s to 0.11s (cryptographic operations)
- Total local execution: <50 seconds for all 5 slowest packages
- Total GitHub execution: 712 seconds (11.9 minutes)

**Conclusion**: Test code is optimized. Further test optimization (e.g., reducing TestMain setup) will yield minimal benefits (<5s total).

### 2. Workflow Overhead is 662+ Seconds

**Calculation**:

```
Total GitHub time: 712s
Total local time: ~50s
Workflow overhead: 712 - 50 = 662 seconds (11 minutes!)
```

**Breakdown Hypothesis** (to be measured):

| Overhead Source | Estimated Time | Cumulative |
|-----------------|----------------|------------|
| Job setup (checkout, Go install) | 60s | 60s |
| Docker Compose up (postgres, otel, grafana) | 120s | 180s |
| Service health checks | 90s | 270s |
| Test execution (actual tests) | 50s | 320s |
| Coverage artifact upload | 60s | 380s |
| Teardown | 30s | 410s |
| **GitHub runner slowness** | **300s** | **710s** |

### 3. sqlrepository 150x Slowdown is Extreme

**Why is sqlrepository 150x slower when others are 5-20x?**

Hypothesis:

- sqlrepository uses in-memory SQLite (no containers)
- Other packages may reuse Docker PostgreSQL instance started in TestMain
- sqlrepository might be the FIRST package tested, incurring full workflow startup cost
- GitHub runner might be slower for single-threaded workloads (SQLite)

**Validation Needed**: Check ci-coverage workflow test execution order.

---

## Action Items

### Priority 1: Measure Workflow Overhead (CRITICAL)

Instrument ci-coverage workflow to identify bottlenecks:

```yaml
- name: 📋 Workflow Start
  run: echo "START_TIME=$(date +%s)" >> $GITHUB_ENV

- name: 📅 After Checkout
  run: |
    CURRENT=$(date +%s)
    echo "Checkout duration: $((CURRENT - START_TIME))s"
    echo "CHECKOUT_TIME=$CURRENT" >> $GITHUB_ENV

- name: 📅 After Go Setup
  run: |
    CURRENT=$(date +%s)
    echo "Go setup duration: $((CURRENT - CHECKOUT_TIME))s"
    echo "GO_SETUP_TIME=$CURRENT" >> $GITHUB_ENV

- name: 📅 After Docker Compose Up
  run: |
    CURRENT=$(date +%s)
    echo "Docker Compose duration: $((CURRENT - GO_SETUP_TIME))s"
    echo "DOCKER_TIME=$CURRENT" >> $GITHUB_ENV

- name: 📅 After Tests
  run: |
    CURRENT=$(date +%s)
    echo "Test execution duration: $((CURRENT - DOCKER_TIME))s"
    echo "TEST_TIME=$CURRENT" >> $GITHUB_ENV

- name: 📅 After Coverage Upload
  run: |
    CURRENT=$(date +%s)
    echo "Coverage upload duration: $((CURRENT - TEST_TIME))s"
    echo "Total workflow duration: $((CURRENT - START_TIME))s"
```

**Expected Output**: Identify which step consumes most time.

### Priority 2: Optimize Workflow (Based on Measurements)

**If Docker Compose startup > 120s**:

- Pre-pull images in parallel: `docker-compose pull --parallel`
- Reduce health check retries (if safe): `retries: 20` → `retries: 10`
- Use `--no-build` flag if images are pre-pulled

**If Service Health Checks > 90s**:

- Reduce check intervals: `interval: 10s` → `interval: 5s`
- Reduce timeouts: `timeout: 5s` → `timeout: 3s`
- Start services in dependency order (postgres → app → telemetry)

**If Coverage Upload > 60s**:

- Compress coverage files before upload
- Upload only final merged coverage (not per-package)
- Use `actions/cache@v4` for coverage artifacts

**If GitHub Runner Slowness > 300s**:

- Consider self-hosted runners (major investment)
- Use GitHub Actions cache more aggressively
- Reduce parallel job count to get faster runners

### Priority 3: Baseline Workflow Performance

Create baseline metrics for tracking improvements:

```markdown
## Baseline (2025-01-05)

| Metric | Current | Target |
|--------|---------|--------|
| Total workflow time | 712s | <300s |
| Job setup | TBD | <60s |
| Docker Compose up | TBD | <60s |
| Service health checks | TBD | <30s |
| Test execution | 50s | 50s (already optimal) |
| Coverage upload | TBD | <30s |
| Teardown | TBD | <20s |
```

---

## Conclusion

**Test code is NOT the bottleneck**. Workflow infrastructure overhead (662+ seconds) is the primary issue.

**Next Steps**:

1. Instrument ci-coverage workflow with timing checkpoints
2. Identify largest time consumer
3. Optimize based on data
4. Target: Reduce 712s → <300s (60% reduction)

**DO NOT**:

- ❌ Further optimize test code (already fast)
- ❌ Remove `t.Parallel()` (tests are concurrent-safe)
- ❌ Reduce coverage (already meets 95% targets)

**DO**:

- ✅ Measure workflow overhead
- ✅ Optimize Docker Compose startup
- ✅ Reduce health check times
- ✅ Compress artifacts
- ✅ Consider self-hosted runners if infrastructure optimization insufficient

---

*Analysis Version: 1.0.0*
*Date: 2025-01-05*
*Next Review: After instrumenting ci-coverage workflow*
