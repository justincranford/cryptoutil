# E2E Test Design Documentation

## 🏗️ **Logical Architecture**

The e2e test suite follows a **layered, component-based architecture** designed for comprehensive end-to-end testing of the cryptoutil application stack.

### **Core Design Principles**

- **Separation of Concerns**: Each component has a single responsibility
- **Test Lifecycle Management**: Proper setup/teardown with resource cleanup
- **Dual Output Logging**: Console + timestamped file logging for debugging
- **Step-by-Step Tracking**: Detailed execution monitoring with timing and status
- **Infrastructure as Code**: Docker Compose-based service orchestration

### **Test Execution Flow**

```
1. Suite Setup → 2. Infrastructure Setup → 3. Service Health Checks →
4. API Testing → 5. Telemetry Verification → 6. Suite Teardown
```

## 📁 **File Structure & Responsibilities**

### **`e2e_test.go`** - Test Entry Points

- **Purpose**: Main test entry points and quick demonstration suite
- **Contains**:
  - `TestE2E()` - Full end-to-end test suite
  - `TestSummaryReportOnly()` - Quick summary report demo
  - `SummaryTestSuite` - Lightweight test for report validation
- **Design**: Minimal orchestration, delegates to `E2ETestSuite`

### **`test_suite.go`** - Core Test Orchestration

- **Purpose**: Main test suite with step tracking and summary reporting
- **Key Components**:
  - `E2ETestSuite` - Main test suite struct
  - `TestStep` & `TestSummary` - Execution tracking structures
  - Individual test methods (`TestInfrastructureHealth`, `TestCryptoutilSQLite`, etc.)
- **Features**:
  - Step-by-step execution tracking with timing
  - Comprehensive summary reports
  - Panic recovery with proper error reporting
  - Dual logging (console + files)

### **`fixtures.go`** - Test Infrastructure Setup

- **Purpose**: Shared test infrastructure and utilities
- **Key Components**:
  - `TestFixture` - Central fixture managing all test resources
  - Service URL management
  - API client initialization
  - Log file management
- **Responsibilities**:
  - Resource lifecycle (setup/teardown)
  - Service URL configuration
  - Cross-platform path handling

### **`infrastructure.go`** - Docker Service Management

- **Purpose**: Docker Compose operations and service health monitoring
- **Key Components**:
  - `InfrastructureManager` - Docker orchestration
  - Service health checking
  - Port reachability verification
- **Features**:
  - Clean environment assurance
  - Service startup/shutdown
  - Health status monitoring
  - HTTP endpoint verification

### **`assertions.go`** - Service Verification Logic

- **Purpose**: Common assertions for service testing
- **Key Components**:
  - `ServiceAssertions` - Assertion helper methods
  - Health check utilities
  - Telemetry flow verification
- **Features**:
  - HTTP readiness checks
  - Docker service health validation
  - Cryptographic operation testing
  - Telemetry data flow verification

### **`e2e-reports/`** - Generated Test Artifacts

- **Purpose**: Timestamped log files for test execution records
- **Contents**:
  - `e2e-test-YYYY-MM-DD_HH-MM-SS.log` - Detailed execution logs
- **Features**:
  - Automatic cleanup (added to `.gitignore`)
  - Structured logging with timestamps
  - Console + file dual output

## 🔄 **Data Flow Architecture**

### **Test Execution Pipeline**

```
TestRunner → TestSuite → TestFixture → InfrastructureManager
                                      ↓
                               ServiceAssertions → API Clients
                                      ↓
                               Log Files + Console Output
```

### **Logging Flow**

```
Test Methods → logStep() → TestSummary.Steps[] → completeStep()
     ↓              ↓                    ↓              ↓
Console Output  Log File            Timing Tracking   Status Updates
```

### **Infrastructure Flow**

```
Setup() → InfrastructureManager.StartServices() → WaitForServicesReady()
     ↓                                                ↓
Clean Environment                              Health Checks + HTTP Ready
     ↓                                                ↓
Teardown() ← StopServices() ← Context Cancellation
```

## 🎯 **Key Design Patterns**

### **1. Test Suite Pattern**

- Uses `testify/suite` for structured test organization
- `SetupSuite`/`TearDownSuite` for lifecycle management
- `SetupTest`/`TearDownTest` for per-test setup

### **2. Fixture Pattern**

- Central `TestFixture` manages all shared resources
- Resource initialization and cleanup
- Cross-cutting concerns (logging, context management)

### **3. Manager Pattern**

- `InfrastructureManager` handles Docker operations
- `ServiceAssertions` provides reusable verification logic
- Single responsibility per manager

### **4. Step Tracking Pattern**

- `TestStep` captures individual operations
- `TestSummary` aggregates execution metrics
- Timing and status monitoring for debugging

### **5. Dual Logging Pattern**

- Console output for real-time visibility
- File logging for permanent records
- Structured format with timestamps and elapsed time

## 🔧 **Configuration & Constants**

### **Magic Values**

- All timeouts, ports, URLs defined in `cryptoutilMagic` package
- Centralized configuration prevents hard-coded values
- Cross-platform compatibility

### **Service Configuration**

- Multiple cryptoutil instances (SQLite, PostgreSQL x2)
- Supporting services (Grafana, OTEL collector, PostgreSQL)
- Port-based service isolation

## 📊 **Execution Monitoring**

### **Step Tracking**

- Each test operation creates a `TestStep`
- Captures start time, description, duration, status
- Aggregated into `TestSummary` for reporting

### **Summary Reports**

- Execution date and total duration
- Step-by-step breakdown with timings
- Success rate calculation
- Status indicators (✅ PASS, ❌ FAIL, ⏭️ SKIP)

### **Health Monitoring**

- Docker service health checks
- HTTP endpoint readiness verification
- Telemetry flow validation
- Port accessibility testing

## 🛡️ **Error Handling & Resilience**

### **Panic Recovery**

- Each test method wrapped in `recover()` blocks
- Proper error reporting through step completion
- Test suite continues despite individual failures

### **Timeout Management**

- Configurable timeouts for different operations
- Context-based cancellation
- Graceful degradation on failures

### **Resource Cleanup**

- Guaranteed teardown execution
- Docker service cleanup
- Log file closure
- Context cancellation

## 🚀 **Usage Examples**

### **Full E2E Test Run**

```bash
go test -tags e2e ./internal/cmd/e2e -run TestE2E -v -timeout 30s
```

### **Quick Summary Demo**

```bash
go test -tags e2e ./internal/cmd/e2e -run TestSummaryReportOnly -v
```

### **Log File Location**

```
internal/cmd/e2e/e2e-reports/e2e-test-YYYY-MM-DD_HH-MM-SS.log
```

## 📋 **Sample Summary Report Output**

```
🎯 E2E TEST EXECUTION SUMMARY REPORT
================================================================================

📅 Execution Date: 2025-10-24 01:24:46
⏱️  Total Duration: 151ms
📊 Total Steps: 5
✅ Passed: 5
❌ Failed: 0
⏭️  Skipped: 0
📈 Success Rate: 100.0%

--------------------------------------------------------------------------------
📋 DETAILED STEP BREAKDOWN
--------------------------------------------------------------------------------
 1. ✅ Summary Test Setup             1ms  Setting up summary test suite
 2. ✅ Quick Demo Test                 0s  Demonstrating summary tracking functionality
 3. ✅ Sub-operation 1              100ms  Performing first sub-operation
 4. ✅ Sub-operation 2               51ms  Performing second sub-operation
 5. ✅ Summary Test Cleanup            0s  Cleaning up summary test suite

================================================================================
🎉 EXECUTION STATUS: FULL SUCCESS
================================================================================
```

This design provides a robust, maintainable, and observable e2e testing framework that can comprehensively validate the cryptoutil application stack while providing detailed execution tracking and debugging capabilities.
