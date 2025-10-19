# Performance Testing Infrastructure

This directory contains scripts and workflows for comprehensive performance testing of the cryptoutil application using k6 load testing framework.

## Overview

The performance testing infrastructure provides:

- **Automated Load Testing**: k6-based performance tests with configurable profiles
- **Performance Analysis**: Trend analysis and regression detection
- **Visual Dashboards**: HTML dashboards with charts and metrics
- **CI/CD Integration**: GitHub Actions workflow for automated performance testing
- **GitHub Pages**: Published performance dashboards and reports

## Quick Start

### Running Performance Tests

The performance testing infrastructure is now implemented in Go for better cross-platform compatibility:

```bash
# Quick performance test (30 seconds, 1 VU)
go run ./scripts/run-performance-tests -profile quick

# Full performance test (2 minutes, 5 VUs)
go run ./scripts/run-performance-tests -profile full -base-url http://localhost:8080

# Deep performance test (5 minutes, 10 VUs)
go run ./scripts/run-performance-tests -profile deep

# Dry run to see what would be executed
go run ./scripts/run-performance-tests -profile quick -dry-run -verbose

# Custom output directory
go run ./scripts/run-performance-tests -profile full -output ./my-results
```

### Script Features

- **Cross-platform**: Works on Windows, Linux, and macOS
- **Configurable profiles**: Quick, full, and deep testing profiles
- **k6 integration**: Uses k6 for load testing with custom metrics
- **Result storage**: Saves results in JSON format for analysis
- **Dry run support**: Preview commands without executing tests

## Test Profiles

| Profile | Duration | Virtual Users | Purpose |
|---------|----------|---------------|---------|
| `quick` | 30s | 1 | Fast feedback, CI/CD |
| `full` | 2m | 5 | Balanced testing |
| `deep` | 5m | 10 | Load testing, limits |

## Scripts

### `run-performance-tests` (Go)
Generates and executes k6 test scripts with configurable profiles.

**Parameters**:
- `-profile`: Test profile (`quick`, `full`, `deep`) (default: `quick`)
- `-base-url`: Application base URL (default: `http://localhost:8080`)
- `-output`: Results directory (default: `./performance-results`)
- `-vus`: Override number of virtual users
- `-duration`: Override test duration
- `-dry-run`: Show commands without executing
- `-verbose`: Enable verbose output

**Example**:
```bash
# Run full profile against local app
go run ./scripts/run-performance-tests -profile full -base-url "http://localhost:8080"

# Run deep test with custom output directory
go run ./scripts/run-performance-tests -profile deep -output "./my-results"
```

### `analyze-performance-results` (Go)
Analyzes k6 JSON results, generates trend reports, and creates HTML dashboards.

**Parameters**:
- `-results-dir`: Directory containing k6 results (default: `./performance-results`)
- `-output-dir`: Directory for reports (default: `./performance-reports`)

**Outputs**:
- `performance-history.json`: Historical performance data
- `performance-dashboard.html`: Interactive dashboard with charts
- `performance-summary-YYYY-MM-DD_HH-mm-ss.md`: Markdown summary report

## CI/CD Integration

### GitHub Actions Workflow

The `.github/workflows/performance.yml` workflow provides:

- **Scheduled Testing**: Daily performance tests at 2 AM UTC
- **On-Demand Testing**: Manual trigger with profile selection
- **Change Detection**: Tests run on pushes to main/develop branches
- **Regression Detection**: Automatic issue creation for performance regressions
- **Dashboard Publishing**: GitHub Pages deployment of performance dashboards

### Workflow Triggers

- **Manual**: Via GitHub Actions UI with profile selection
- **Scheduled**: Daily at 2 AM UTC
- **Push**: On changes to application code or performance scripts

### Performance Metrics

The workflow tracks and reports:

- Average response time
- 95th/99th percentile response times
- Error rate percentage
- Requests per second (throughput)
- Performance status (stable/significant_change)

## Performance Dashboards

### GitHub Pages Integration

Performance dashboards are automatically published to GitHub Pages at:
```
https://{username}.github.io/{repository}/performance/{run_number}/performance-dashboard.html
```

### Dashboard Features

- **Real-time Metrics**: Current performance indicators
- **Trend Analysis**: Historical performance trends with percentage changes
- **Interactive Charts**: Response time, error rate, and throughput over time
- **Status Indicators**: Visual health status (healthy/warning/critical)
- **Threshold Compliance**: Pass/fail indicators for performance thresholds

### Thresholds

| Metric | Threshold | Status |
|--------|-----------|--------|
| P95 Response Time | < 500ms | ✅ Pass |
| Error Rate | < 10% | ✅ Pass |
| Significant Change | > 10% (response) or > 5% (error) | ⚠️ Review |

## Test Scenarios

The performance tests cover key cryptoutil operations:

### Cryptographic Operations
- **Key Generation**: RSA, ECDSA, EdDSA key pair generation
- **Encryption/Decryption**: AES encryption and decryption operations
- **Digital Signatures**: Signing and verification operations

### API Endpoints
- **Health Checks**: Service availability and responsiveness
- **Key Management**: Key creation, retrieval, and rotation
- **Crypto Operations**: Encryption, decryption, signing APIs

### Load Patterns
- **Ramp-up**: Gradual increase in virtual users
- **Steady State**: Sustained load for performance measurement
- **Spike Testing**: Sudden load increases (deep profile only)

## Configuration

### k6 Test Configuration

Test scenarios are defined in the Go script with configurable parameters:

```javascript
// Example k6 test structure (generated by Go script)
export let options = {
  vus: 5,
  duration: '2m',
  stages: [
    { duration: '30s', target: 5 },   // Ramp up
    { duration: '1m30s', target: 5 }, // Steady state
  ],
  thresholds: {
    http_req_duration: ['p(95)<750'],  // 95th percentile < 750ms
    http_req_failed: ['rate<0.05'],    // Error rate < 5%
  },
};
```

### Custom Thresholds

Modify thresholds in the Go script profiles:

```go
// Custom thresholds for your environment
fullProfile := TestProfile{
    Thresholds: map[string]map[string]interface{}{
        "http_req_duration": {"p(95)<1000": true},  // Adjust for your requirements
        "http_req_failed":   {"rate<0.02": true},   // Stricter error rate
    },
}
```

## Troubleshooting

### Common Issues

1. **k6 Not Found**:
   ```bash
   # Install k6
   choco install k6  # Windows
   # or
   brew install k6   # macOS
   # or
   # Download from https://k6.io/docs/get-started/installation/
   ```

2. **Application Not Running**:
   ```bash
   # Start cryptoutil for testing
   go run ./cmd/cryptoutil server --config configs/test/config.yml
   ```

3. **Go Not Available**:
   ```bash
   # Install Go
   # Download from https://golang.org/dl/
   ```

4. **Dashboard Not Loading**:
   - Ensure `performance-reports/performance-dashboard.html` exists
   - Check browser console for JavaScript errors
   - Verify Chart.js library is accessible

### Performance Regression Investigation

When a performance regression is detected:

1. **Review Dashboard**: Check the published GitHub Pages dashboard
2. **Compare Results**: Look at historical trends in `performance-history.json`
3. **Analyze Changes**: Review recent code changes that might affect performance
4. **Run Locally**: Reproduce the issue with local performance tests
5. **Profile Application**: Use Go profiling tools to identify bottlenecks

## Contributing

### Adding New Test Scenarios

1. **Modify Test Script**: Update `scripts/run-performance-tests/main.go` with new scenarios
2. **Add Metrics**: Include relevant custom metrics for your operations
3. **Update Thresholds**: Adjust performance thresholds as needed
4. **Test Locally**: Validate new scenarios work correctly
5. **Update Documentation**: Document new scenarios and expected performance

### Customizing Dashboards

The HTML dashboard can be customized by modifying `scripts/analyze-performance-results/main.go`:

- **Chart Configuration**: Modify Chart.js options for different visualizations
- **Metrics Display**: Add or remove metrics from the dashboard
- **Styling**: Update CSS for different visual themes
- **Thresholds**: Adjust status indicators and compliance checks

## Performance Best Practices

### Test Environment
- **Isolated Environment**: Run tests in dedicated environments
- **Consistent Hardware**: Use similar hardware for baseline comparisons
- **Network Conditions**: Account for network latency in distributed setups

### Result Interpretation
- **Statistical Significance**: Consider variance in performance measurements
- **Baseline Establishment**: Establish performance baselines over multiple runs
- **Trend Analysis**: Focus on trends rather than individual measurements
- **Context Awareness**: Consider external factors (system load, network issues)

### Continuous Monitoring
- **Regular Testing**: Run performance tests regularly, not just before releases
- **Automated Alerts**: Set up notifications for performance regressions
- **Historical Tracking**: Maintain performance history for trend analysis
- **Documentation**: Keep performance expectations and thresholds documented
