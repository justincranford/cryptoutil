package cryptoutil;

import static cryptoutil.GatlingHttpUtil.createServiceApiProtocol;
import static io.gatling.javaapi.core.CoreDsl.global;
import static io.gatling.javaapi.core.CoreDsl.rampUsers;
import static io.gatling.javaapi.core.CoreDsl.scenario;
import static io.gatling.javaapi.http.HttpDsl.http;
import static io.gatling.javaapi.http.HttpDsl.status;

import java.util.ArrayList;
import java.util.List;

import io.gatling.javaapi.core.PopulationBuilder;
import io.gatling.javaapi.core.Simulation;
import io.gatling.javaapi.http.HttpProtocolBuilder;

/**
 * Skeleton Product Load Testing Simulation.
 * Tests skeleton-template service: health endpoints and basic API validation.
 * The skeleton-template is the reference implementation for new services.
 *
 * <p>Example Commands:
 * <pre>
 * # Quick validation (health check)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.SkeletonProductSimulation -Dprofile=quick
 *
 * # Standard load test (health + browser endpoints)
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.SkeletonProductSimulation -Dprofile=standard -Dvirtualclients=10
 *
 * # Stress test
 * .\mvnw.cmd gatling:test -Dgatling.simulationClass=cryptoutil.SkeletonProductSimulation -Dprofile=stress -Dvirtualclients=50 -DdurationSeconds=120
 * </pre>
 */
public class SkeletonProductSimulation extends Simulation {

    private static final String PROFILE = System.getProperty("profile", "quick");
    private static final String SKELETON_PORT = System.getProperty("skeletonPort", "8900");
    private static final int VIRTUAL_CLIENTS = Integer.parseInt(System.getProperty("virtualclients", "1"));
    private static final int DURATION_SECONDS = Integer.parseInt(System.getProperty("durationSeconds", "60"));

    private static final String SKELETON_SERVICE_URL = "https://127.0.0.1:" + SKELETON_PORT + "/service/api/v1";
    private static final String SKELETON_BROWSER_URL = "https://127.0.0.1:" + SKELETON_PORT + "/browser/api/v1";

    {
        List<PopulationBuilder> populations = new ArrayList<>();
        HttpProtocolBuilder serviceProtocol = createServiceApiProtocol(SKELETON_SERVICE_URL);

        switch (PROFILE) {
            case "standard":
                // Service + browser health checks
                populations.add(
                    scenario("Skeleton Service Health")
                        .exec(http("Service Health Check")
                            .get("/health")
                            .check(status().in(200, 503)))
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(serviceProtocol)
                );
                populations.add(
                    scenario("Skeleton Browser Health")
                        .exec(http("Browser Health Check")
                            .get(SKELETON_BROWSER_URL + "/health")
                            .check(status().in(200, 503)))
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(serviceProtocol)
                );
                break;

            case "stress":
                // High-concurrency health checks
                populations.add(
                    scenario("Skeleton Stress - Service")
                        .exec(http("Service Health Check")
                            .get("/health")
                            .check(status().in(200, 503)))
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS * 2).during(DURATION_SECONDS))
                        .protocols(serviceProtocol)
                );
                populations.add(
                    scenario("Skeleton Stress - Browser")
                        .exec(http("Browser Health Check")
                            .get(SKELETON_BROWSER_URL + "/health")
                            .check(status().in(200, 503)))
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS * 2).during(DURATION_SECONDS))
                        .protocols(serviceProtocol)
                );
                break;

            case "quick":
            default:
                // Basic health check only
                populations.add(
                    scenario("Skeleton Quick Validation")
                        .exec(http("Service Health Check")
                            .get("/health")
                            .check(status().in(200, 503)))
                        .injectOpen(rampUsers(VIRTUAL_CLIENTS).during(DURATION_SECONDS))
                        .protocols(serviceProtocol)
                );
                break;
        }

        setUp(populations.toArray(new PopulationBuilder[0]))
            .assertions(
                global().responseTime().max().lt(2000),
                global().successfulRequests().percent().gte(95.0)
            );
    }
}
