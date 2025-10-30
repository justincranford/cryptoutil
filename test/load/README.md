# Gatling Performance Tests for Cryptoutil

This directory contains Gatling performance tests for the cryptoutil API using Maven.

## About This Project

A simple showcase of a Maven project using the Gatling plugin for Maven. Refer to the plugin documentation [on the Gatling website](https://docs.gatling.io/reference/integrations/build-tools/maven-plugin/) for usage.

This project is written in Java, others are available for [Kotlin](https://github.com/gatling/gatling-maven-plugin-demo-kotlin) and [Scala](https://github.com/gatling/gatling-maven-plugin-demo-scala).

It includes:

* [Maven Wrapper](https://maven.apache.org/wrapper/), so that you can immediately run Maven with `./mvnw` without having to install it on your computer
* minimal `pom.xml`
* latest version of `io.gatling:gatling-maven-plugin` applied
* sample [Simulation](https://docs.gatling.io/reference/glossary/#simulation) class, demonstrating sufficient Gatling functionality
* proper source file layout

## Overview

These performance tests use [Gatling](https://gatling.io/) to load test the cryptoutil REST API endpoints across three different API contexts. The tests simulate multiple virtual clients/users making requests to the service-to-service, browser, and admin APIs to measure their performance under load.

## API Contexts Tested

cryptoutil exposes three distinct API contexts with different security and performance characteristics:

### 1. Service API (`/service/api/v1/*`)
- **Purpose**: Machine-to-machine service communication
- **Security**: Core security only (no browser-specific middleware)
- **Base URL**: `https://localhost:8080/service/api/v1/`
- **Parameter**: `virtualclients`
- **Endpoints**: Key management, encryption/decryption operations

### 2. Browser API (`/browser/api/v1/*`)
- **Purpose**: Browser/web client communication
- **Security**: Full browser security (CORS, CSRF, CSP, security headers)
- **Base URL**: `https://localhost:8080/browser/api/v1/`
- **Parameter**: `virtualusers`
- **Endpoints**: Same operations as service API but with CSRF protection

### 3. Admin API (Port 9090)
- **Purpose**: Health checks and system monitoring
- **Security**: Private interface for Kubernetes probes
- **Base URL**: `http://localhost:9090/`
- **Parameter**: `virtualadminclients`
- **Endpoints**: `/livez` (liveness), `/readyz` (readiness)

## Prerequisites

- Java 21 or higher
- Maven (or use the included Maven Wrapper)

## JDK Configuration

This project is configured to use Java 21. If you have Java 21 installed at a custom location, you can configure it in several ways:

### Option 1: Project-specific JVM Configuration (Recommended)

The project includes a `.mvn/jvm.config` file that specifies the JDK path. Update this file to point to your Java 21 installation:

```properties
--java-home
C:\Dev\jdk21
```

### Option 2: Environment Variable

Set the `JAVA_HOME` environment variable to point to your Java 21 installation:

```powershell
$env:JAVA_HOME = "C:\Dev\jdk21"
```

### Option 3: Maven Toolchains

Create a `~/.m2/toolchains.xml` file:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<toolchains>
  <toolchain>
    <type>jdk</type>
    <provides>
      <version>21</version>
      <vendor>oracle</vendor>
    </provides>
    <configuration>
      <jdkHome>C:\Dev\jdk21</jdkHome>
    </configuration>
  </toolchain>
</toolchains>
```

Then update the pom.xml to use toolchains:

```xml
<properties>
  <maven.compiler.release>21</maven.compiler.release>
</properties>

<build>
  <plugins>
    <plugin>
      <groupId>org.apache.maven.plugins</groupId>
      <artifactId>maven-toolchains-plugin</artifactId>
      <version>3.2.0</version>
      <executions>
        <execution>
          <goals>
            <goal>toolchain</goal>
          </goals>
        </execution>
      </executions>
      <configuration>
        <toolchains>
          <jdk>
            <version>21</version>
          </jdk>
        </toolchains>
      </configuration>
    </plugin>
    <!-- other plugins -->
  </plugins>
</build>
```

## Quick Start

### Run Tests with Maven Wrapper (Recommended)

#### Run All Simulations
```bash
./mvnw gatling:test
```

#### Run Specific API Simulations
```bash
# Service API tests
./mvnw gatling:test -Dgatling.simulationClass=cryptoutil.ServiceApiSimulation -Dvirtualclients=10

# Browser API tests
./mvnw gatling:test -Dgatling.simulationClass=cryptoutil.BrowserApiSimulation -Dvirtualusers=5

# Admin API tests
./mvnw gatling:test -Dgatling.simulationClass=cryptoutil.AdminApiSimulation -Dvirtualadminclients=2

mvnw.cmd gatling:test
```

## Test Configuration

### System Properties

#### Service API Configuration
- `baseurl`: Base URL for service API (default: `https://localhost:8080/service/api/v1/`)
- `virtualclients`: Number of virtual clients for service API tests (default: `1`)

#### Browser API Configuration
- `browserApiBaseUrl`: Base URL for browser API (default: `https://localhost:8080/browser/api/v1/`)
- `virtualusers`: Number of virtual users for browser API tests (default: `1`)

#### Admin API Configuration
- `adminApiBaseUrl`: Base URL for admin API (default: `http://localhost:9090/`)
- `virtualadminclients`: Number of virtual admin clients for health check tests (default: `1`)

### Gatling Maven Plugin Properties

- `gatling.simulationClass`: Fully qualified class name of the simulation to run (default: runs all simulations)
- `gatling.runDescription`: Description for the test run
- `gatling.outputDirectory`: Directory for test results (default: `target/gatling`)
- `gatling.resultsFolder`: Subfolder name for results (default: timestamp-based)
- `gatling.noReports`: Skip HTML report generation (default: `false`)
- `gatling.reportsOnly`: Only generate reports from existing results (default: `false`)

### Example Usage

#### Service API Testing
```bash
# Test service API with 5 virtual clients
./mvnw gatling:test -Dgatling.simulationClass=cryptoutil.ServiceApiSimulation -Dvirtualclients=5

# Test service API against staging environment
./mvnw gatling:test \
  -Dgatling.simulationClass=cryptoutil.ServiceApiSimulation \
  -Dbaseurl=https://staging-api.cryptoutil.com/service/api/v1/ \
  -Dvirtualclients=20
```

#### Browser API Testing
```bash
# Test browser API with 3 virtual users
./mvnw gatling:test -Dgatling.simulationClass=cryptoutil.BrowserApiSimulation -Dvirtualusers=3

# Test browser API against staging environment
./mvnw gatling:test \
  -Dgatling.simulationClass=cryptoutil.BrowserApiSimulation \
  -DbrowserApiBaseUrl=https://staging-api.cryptoutil.com/browser/api/v1/ \
  -Dvirtualusers=10
```

#### Admin API Testing
```bash
# Test admin API health checks with 2 virtual admin clients
./mvnw gatling:test -Dgatling.simulationClass=cryptoutil.AdminApiSimulation -Dvirtualadminclients=2

# Test admin API against different port
./mvnw gatling:test \
  -Dgatling.simulationClass=cryptoutil.AdminApiSimulation \
  -DadminApiBaseUrl=http://localhost:9091/ \
  -Dvirtualadminclients=5
```

#### Combined Testing
```bash
# Run all simulations
./mvnw gatling:test

# Run specific simulation with custom description
./mvnw gatling:test \
  -Dgatling.simulationClass=cryptoutil.ServiceApiSimulation \
  -Dgatling.runDescription="Production Load Test" \
  -Dbaseurl=https://api.cryptoutil.com/service/api/v1/ \
  -Dvirtualclients=50

# Generate reports only (no new test execution)
./mvnw gatling:test -Dgatling.reportsOnly=target/gatling/results/my-test-run

# Skip report generation for CI environments
./mvnw gatling:test -Dgatling.noReports=true
```

## Project Structure

```
src/test/java/
├── cryptoutil/
│   ├── ServiceApiSimulation.java    # Service API performance tests (/service/api/v1/*)
│   ├── BrowserApiSimulation.java    # Browser API performance tests (/browser/api/v1/*)
│   └── AdminApiSimulation.java      # Admin API health check tests (port 9090)
```

## Creating Custom Simulations

1. Create a new Java class extending `Simulation`
2. Define HTTP protocol configuration with appropriate base URL
3. Create scenarios with request chains for your API endpoints
4. Set up injection profiles with the appropriate virtual client/user parameter
5. Add assertions for performance thresholds

Example for a custom service API simulation:

```java
public class CustomServiceSimulation extends Simulation {
  private static final String baseUrl = System.getProperty("baseurl", "https://localhost:8080/service/api/v1");
  private static final int virtualclients = Integer.getInteger("virtualclients", 1);

  private static final HttpProtocolBuilder httpProtocol = http
      .baseUrl(baseUrl)
      .acceptHeader("application/json")
      .contentTypeHeader("application/json");

  private static final ScenarioBuilder customScenario = scenario("Custom Service API Test")
      .exec(http("Custom API Call")
          .post("/your-endpoint")
          .body(StringBody("{\"param\":\"value\"}"))
          .check(status().is(200)));

  {
    setUp(customScenario.injectOpen(rampUsers(virtualclients).during(30)))
         .protocols(httpProtocol);
  }
}
```

## Reports

Gatling generates HTML reports in `target/gatling/results/` after each test run. Open `index.html` to view detailed performance metrics.

## Integration with CI/CD

### GitHub Actions Workflow

The project includes a dedicated `ci-load.yml` workflow for automated load testing:

```bash
# Run locally using the workflow runner
go run ./cmd/workflow -workflows=load -inputs="load_profile=quick"
go run ./cmd/workflow -workflows=load -inputs="load_profile=standard"
go run ./cmd/workflow -workflows=load -inputs="load_profile=stress"
```

**Workflow Features:**
- Automatic Docker Compose service orchestration
- Real-time infrastructure monitoring (CPU, memory, network, disk)
- Service health verification before tests
- Comprehensive artifact collection (Gatling reports, metrics, logs)
- Three load profiles: quick (10 clients/30s), standard (50 clients/120s), stress (200 clients/300s)

**GitHub Actions Integration:**

Add to your CI/CD pipeline:

#### Service API Testing
```yaml
- name: Run Service API Performance Tests
  run: |
    cd test/load
    ./mvnw gatling:test \
      -Dgatling.simulationClass=cryptoutil.ServiceApiSimulation \
      -Dbaseurl=${{ secrets.SERVICE_API_BASE_URL }} \
      -Dvirtualclients=${{ vars.SERVICE_API_LOAD_TEST_USERS || 10 }}
```

#### Browser API Testing
```yaml
- name: Run Browser API Performance Tests
  run: |
    cd test/load
    ./mvnw gatling:test \
      -Dgatling.simulationClass=cryptoutil.BrowserApiSimulation \
      -DbrowserApiBaseUrl=${{ secrets.BROWSER_API_BASE_URL }} \
      -Dvirtualusers=${{ vars.BROWSER_API_LOAD_TEST_USERS || 5 }}
```

#### Admin API Health Check Testing
```yaml
- name: Run Admin API Health Check Tests
  run: |
    cd test/load
    ./mvnw gatling:test \
      -Dgatling.simulationClass=cryptoutil.AdminApiSimulation \
      -DadminApiBaseUrl=${{ secrets.ADMIN_API_BASE_URL }} \
      -Dvirtualadminclients=${{ vars.ADMIN_API_LOAD_TEST_USERS || 2 }}
```

#### Matrix Testing (All APIs)
```yaml
- name: Run All API Performance Tests
  strategy:
    matrix:
      include:
        - simulation: cryptoutil.ServiceApiSimulation
          base-url: ${{ secrets.SERVICE_API_BASE_URL }}
          virtual-clients: ${{ vars.SERVICE_API_LOAD_TEST_USERS || 10 }}
          api-type: service
        - simulation: cryptoutil.BrowserApiSimulation
          base-url: ${{ secrets.BROWSER_API_BASE_URL }}
          virtual-clients: ${{ vars.BROWSER_API_LOAD_TEST_USERS || 5 }}
          api-type: browser
        - simulation: cryptoutil.AdminApiSimulation
          base-url: ${{ secrets.ADMIN_API_BASE_URL }}
          virtual-clients: ${{ vars.ADMIN_API_LOAD_TEST_USERS || 2 }}
          api-type: admin
  run: |
    cd test/load
    ./mvnw gatling:test \
      -Dgatling.simulationClass=${{ matrix.simulation }} \
      -D${{ matrix.api-type }}ApiBaseUrl=${{ matrix.base-url }} \
      -Dvirtual${{ matrix.api-type }}clients=${{ matrix.virtual-clients }}
```

## Available Maven Goals

The Gatling Maven plugin provides several goals:

- `gatling:test`: Run Gatling simulations (default goal)
- `gatling:recorder`: Start the Gatling recorder for creating new simulations
- `gatling:enterpriseDeploy`: Deploy to Gatling Enterprise Cloud
- `gatling:enterpriseStart`: Start simulations on Gatling Enterprise Cloud

### Advanced Usage

```bash
# Start the Gatling recorder to create new simulations
./mvnw gatling:recorder

# Deploy to Gatling Enterprise (requires configuration)
./mvnw gatling:enterpriseDeploy -Dgatling.enterprise.url=https://cloud.gatling.io
```

## Performance Thresholds

Configure performance assertions in your simulations based on API type:

### Service API (Machine-to-Machine)
```java
private static final Assertion serviceAssertions = global()
  .responseTime().percentile(95).lt(1000)  // 95th percentile < 1s
  .and(global().failedRequests().percentile(99).lt(1)); // < 1% failure rate
```

### Browser API (User-to-Machine)
```java
private static final Assertion browserAssertions = global()
  .responseTime().percentile(95).lt(1500)  // 95th percentile < 1.5s (CSRF overhead)
  .and(global().failedRequests().percentile(99).lt(1)); // < 1% failure rate
```

### Admin API (Health Checks)
```java
private static final Assertion adminAssertions = global()
  .responseTime().percentile(95).lt(100)   // 95th percentile < 100ms (very fast)
  .and(global().failedRequests().percentile(99).lt(0.1)); // < 0.1% failure rate
```

## Troubleshooting

### Common Issues

1. **JAVA_HOME not set**: Ensure Java 21+ is installed and JAVA_HOME is configured
2. **Port conflicts**: Ensure cryptoutil is running on the expected ports (8080 for public APIs, 9090 for admin)
3. **SSL/TLS issues**: Use `https://` for public APIs (ports 8080), `http://` for admin API (port 9090)
4. **CSRF token errors**: Browser API requires CSRF tokens; ensure `/csrf-token` endpoint is accessible
5. **Network timeouts**: Adjust HTTP timeouts in protocol configuration for slower browser API calls
6. **API context confusion**: Use correct base URLs for each API type:
   - Service API: `https://localhost:8080/service/api/v1/`
   - Browser API: `https://localhost:8080/browser/api/v1/`
   - Admin API: `http://localhost:9090/`

### Debug Mode

Enable debug logging and additional output:

```bash
# Enable console and file data writers
./mvnw gatling:test -Dgatling.data.writers=console,file

# Run with verbose Maven output
./mvnw gatling:test -X

# Run with custom logback configuration
./mvnw gatling:test -Dlogback.configurationFile=logback-debug.xml
```

### Performance Tuning

```bash
# Adjust JVM settings for high load tests
./mvnw gatling:test -Dgatling.simulationClass=cryptoutil.CryptoSimulation \
  -DjvmArgs="-Xmx2g -Xms1g"

# Run with custom thread pools
./mvnw gatling:test -Dgatling.elFileBodiesCacheMaxCapacity=100 \
  -Dgatling.rawFileBodiesCacheMaxCapacity=100
```

## Resources

- [Gatling Documentation](https://docs.gatling.io/)
- [Gatling Maven Plugin](https://docs.gatling.io/reference/integrations/build-tools/maven-plugin/)
- [Cryptoutil API Documentation](../../docs/README.md)
