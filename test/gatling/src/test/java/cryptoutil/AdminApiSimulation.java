package cryptoutil;

import static io.gatling.javaapi.core.CoreDsl.*;
import static io.gatling.javaapi.http.HttpDsl.*;

import io.gatling.javaapi.core.*;
import io.gatling.javaapi.http.*;

public class AdminApiSimulation extends Simulation {

  // Load configuration from system properties
  private static final String baseUrl = System.getProperty("adminApiBaseUrl", "http://localhost:9090");
  private static final int virtualadminclients = Integer.getInteger("virtualadminclients", 1);

  // Define HTTP configuration for admin API (private interface)
  private static final HttpProtocolBuilder httpProtocol = http
      .baseUrl(baseUrl)
      .acceptHeader("application/json")
      .userAgentHeader("Gatling-Cryptoutil-Admin-API-Test/1.0");

  // Liveness probe scenario
  private static final ScenarioBuilder livenessScenario = scenario("Admin API - Liveness Probe")
      .exec(http("Liveness Check")
          .get("/livez")
          .check(status().is(200))
          .check(jsonPath("$.status").is("ok")));

  // Readiness probe scenario
  private static final ScenarioBuilder readinessScenario = scenario("Admin API - Readiness Probe")
      .exec(http("Readiness Check")
          .get("/readyz")
          .check(status().is(200))
          .check(jsonPath("$.status").is("ok")));

  // Combined health check scenario
  private static final ScenarioBuilder healthCheckScenario = scenario("Admin API - Health Checks")
      .exec(livenessScenario)
      .pause(0.5)
      .exec(readinessScenario);

  // High-frequency health monitoring scenario
  private static final ScenarioBuilder monitoringScenario = scenario("Admin API - Continuous Monitoring")
      .forever()
      .on(
          exec(livenessScenario)
          .pause(1)  // Check every second
          .exec(readinessScenario)
          .pause(1)
      );

  // Performance assertions for admin endpoints (should be very fast)
  private static final Assertion assertions = global()
      .responseTime().percentile(95).lt(100)  // 95th percentile < 100ms (health checks should be fast)
      .and(global().failedRequests().percentile(99).lt(0.1)); // < 0.1% failure rate (near-zero tolerance)

  // Injection profile and test setup
  {
    setUp(
        livenessScenario.injectOpen(rampUsers(virtualadminclients).during(5)),
        readinessScenario.injectOpen(rampUsers(virtualadminclients).during(5)),
        healthCheckScenario.injectOpen(rampUsers(virtualadminclients).during(10))
        // Note: monitoringScenario is commented out by default as it runs forever
        // Uncomment for continuous monitoring tests:
        // monitoringScenario.injectOpen(atOnceUsers(virtualadminclients)).throttle(reachRps(10).in(10), holdFor(ofSeconds(60)))
    )
    .assertions(assertions)
    .protocols(httpProtocol);
  }
}
