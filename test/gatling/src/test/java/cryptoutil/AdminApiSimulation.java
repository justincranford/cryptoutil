package cryptoutil;

import static io.gatling.javaapi.core.CoreDsl.*;
import static io.gatling.javaapi.http.HttpDsl.*;

import io.gatling.javaapi.core.*;
import io.gatling.javaapi.http.*;

public class AdminApiSimulation extends Simulation {

  // Load configuration from system properties
  private static final String baseUrl = System.getProperty("adminApiBaseUrl", "https://localhost:9090");
  private static final int virtualadminclients = Integer.getInteger("virtualadminclients", 1);

  // Define HTTP configuration for admin API (private interface)
  private static final HttpProtocolBuilder httpProtocol = http
      .baseUrl(baseUrl)
      .acceptHeader("application/json")
      .userAgentHeader("Gatling-Cryptoutil-Admin-API-Test/1.0");
      // SSL certificate validation disabled via JVM system properties

  // Liveness probe chain
  private static final ChainBuilder livenessChain = exec(http("Liveness Check")
      .get("/livez")
      .check(status().is(200))
      .check(jsonPath("$.status").is("ok")));

  // Readiness probe chain
  private static final ChainBuilder readinessChain = exec(http("Readiness Check")
      .get("/readyz")
      .check(status().is(200))
      .check(jsonPath("$.status").is("ok")));

  // Liveness probe scenario
  private static final ScenarioBuilder livenessScenario = scenario("Admin API - Liveness Probe")
      .exec(livenessChain);

  // Readiness probe scenario
  private static final ScenarioBuilder readinessScenario = scenario("Admin API - Readiness Probe")
      .exec(readinessChain);

  // Combined health check scenario
  private static final ScenarioBuilder healthCheckScenario = scenario("Admin API - Health Checks")
      .exec(livenessChain)
      .pause(1)  // 1 second pause
      .exec(readinessChain);

  // High-frequency health monitoring scenario
  private static final ScenarioBuilder monitoringScenario = scenario("Admin API - Continuous Monitoring")
      .forever()
      .on(
          exec(livenessChain)
          .pause(1)  // Check every second
          .exec(readinessChain)
          .pause(1)
      );

  // Performance assertions for admin endpoints (should be very fast)
  private static final Assertion assertions = global()
      .responseTime().max().lt(100);  // max response time < 100ms (health checks should be fast)

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
